package main

import (
	"encoding/binary"
	"fmt"
	"strings"
	"sync"
)

// 이 파일은 앱과 서버 사이를 오가는 PostgreSQL 프로토콜 바이트를 메시지 단위로 자릅니다(실험 5, 2026-10-15).
// 중계기(relay.go)가 맥 쪽과 컨테이너 쪽에서 같은 바이트를 각각 자르고, 두 쪽의 같은 순번 메시지를 이어 두 시계의 차이를 잽니다.
//
// 메시지의 꼴(PostgreSQL 18 문서 「Message Formats」): 시작 단계의 메시지는 형식 바이트 없이 int32 길이와 int32 코드로 시작하고,
// 그 뒤의 메시지는 형식 바이트 하나와 int32 길이(길이 칸 포함, 형식 바이트 제외)로 시작합니다. SSLRequest·GSSENCRequest에
// 서버는 길이 없는 바이트 하나('N' 거절, 'S'·'G' 수락)로 답합니다.

// 시작 단계 메시지의 코드입니다(PostgreSQL 18 문서 「Message Formats」, pqcomm.h).
const (
	codeSSL      = 80877103
	codeGSSENC   = 80877104
	codeCancel   = 80877102
	codeProto30  = 3 << 16
	codeProto32  = 3<<16 | 2
	maxValueSize = 64
)

// wireMsg 는 중계기가 본 메시지 하나입니다. Dir은 F(앱 → 서버)와 B(서버 → 앱)이고, N은 그 방향 안의 순번(1부터)입니다.
// Type은 형식 바이트(Q, P, B …)이거나, 형식 바이트가 없는 시작 단계의 이름(ssl, gssenc, cancel, startup, sslreply)입니다.
// Len은 형식 바이트를 뺀 길이(길이 칸 포함)이고, 바이트 하나짜리 sslreply는 1입니다.
// Ns는 그 메시지의 끝 바이트를 읽은 시각(그쪽 시계의 CLOCK_REALTIME 나노초)입니다.
// 나머지 칸은 형식마다 뜻이 있는 것만 채웁니다. 값은 이 실험의 앱이 보낸 것이며, 열 값은 앞 64바이트까지만 남깁니다.
type wireMsg struct {
	Conn     int               `json:"conn"`
	Dir      string            `json:"dir"`
	N        int               `json:"n"`
	Type     string            `json:"type"`
	Len      int               `json:"len"`
	Ns       int64             `json:"ns"`
	Name     string            `json:"name,omitempty"`
	SQL      string            `json:"sql,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
	Tag      string            `json:"tag,omitempty"`
	Status   string            `json:"status,omitempty"`
	Pid      int               `json:"pid,omitempty"`
	Auth     *int              `json:"auth,omitempty"`
	Count    int               `json:"count,omitempty"`
	Values   []string          `json:"values,omitempty"`
	Severity string            `json:"severity,omitempty"`
	Code     string            `json:"code,omitempty"`
	Message  string            `json:"message,omitempty"`
	Where    string            `json:"where,omitempty"`
}

// wireConn 은 연결 하나의 두 방향 자르기 상태입니다. 두 방향을 서로 다른 고루틴이 먹이므로 잠금으로 묶습니다.
// replies는 서버가 바이트 하나로 답해야 할 요청(SSLRequest·GSSENCRequest)의 수입니다. 앱 쪽 요청을 먼저 자른 뒤 서버에 쓰므로,
// 서버의 답을 자를 때는 늘 그 요청이 이미 세어져 있습니다. opaque는 서버가 암호화를 받아들여 더 자를 수 없게 된 상태입니다.
type wireConn struct {
	mu      sync.Mutex
	id      int
	buf     [2][]byte
	n       [2]int
	typedF  bool
	replies int
	opaque  bool
}

// dirIndex 는 방향 이름을 buf·n의 자리로 바꿉니다.
func dirIndex(dir byte) int {
	if dir == 'B' {
		return 1
	}
	return 0
}

// feed 는 한 방향으로 읽은 바이트 조각을 더하고, 그 조각으로 끝난 메시지를 차례대로 돌려줍니다. 끝나지 않은 메시지는 다음 조각을
// 기다립니다. ns는 이 조각을 읽은 시각이고, 이 조각으로 끝난 메시지는 모두 이 시각을 받습니다.
//
// Returns: 이 조각으로 끝난 메시지들입니다. 암호화가 시작된 뒤에는 처음 한 번만 Type "opaque"를 돌려주고 그 뒤로는 자르지 않습니다.
func (c *wireConn) feed(dir byte, chunk []byte, ns int64) []wireMsg {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.opaque {
		return nil
	}
	d := dirIndex(dir)
	c.buf[d] = append(c.buf[d], chunk...)
	var out []wireMsg
	for {
		m, size, ok := c.next(dir)
		if !ok {
			break
		}
		c.buf[d] = c.buf[d][size:]
		c.n[d]++
		m.Conn, m.Dir, m.N, m.Ns = c.id, string(dir), c.n[d], ns
		out = append(out, m)
		if m.Type == "opaque" {
			c.opaque = true
			break
		}
	}
	return out
}

// next 는 한 방향의 버퍼 앞에서 메시지 하나를 자릅니다. 메시지가 아직 다 오지 않았으면 ok가 거짓입니다.
// 잠금은 feed가 잡고 있습니다.
func (c *wireConn) next(dir byte) (m wireMsg, size int, ok bool) {
	b := c.buf[dirIndex(dir)]
	if dir == 'B' && c.replies > 0 {
		if len(b) < 1 {
			return m, 0, false
		}
		c.replies--
		if b[0] != 'N' {
			// 'S'(TLS)나 'G'(GSSAPI)를 받으면 그 뒤 바이트는 암호문이라 자를 수 없습니다. 이 실습 서버는 ssl=off라 'N'만 옵니다.
			return wireMsg{Type: "opaque", Len: 1, Status: string(b[:1])}, 1, true
		}
		return wireMsg{Type: "sslreply", Len: 1, Status: "N"}, 1, true
	}
	if dir == 'F' && !c.typedF {
		if len(b) < 8 {
			return m, 0, false
		}
		l := int(binary.BigEndian.Uint32(b))
		if l < 8 {
			// 길이가 코드 칸보다 짧으면 프로토콜을 벗어난 바이트입니다. 그 뒤는 믿을 수 없으므로 자르기를 멈춥니다.
			return wireMsg{Type: "opaque", Len: l}, len(b), true
		}
		if len(b) < l {
			return m, 0, false
		}
		m = startupMsg(binary.BigEndian.Uint32(b[4:]), b[8:l])
		m.Len = l
		switch m.Type {
		case "ssl", "gssenc":
			c.replies++
		case "startup":
			c.typedF = true
		}
		return m, l, true
	}
	if len(b) < 5 {
		return m, 0, false
	}
	l := int(binary.BigEndian.Uint32(b[1:]))
	if l < 4 {
		return wireMsg{Type: "opaque", Len: l}, len(b), true
	}
	if len(b) < 1+l {
		return m, 0, false
	}
	m = typedMsg(dir, b[0], b[5:1+l])
	m.Len = l
	return m, 1 + l, true
}

// startupMsg 는 형식 바이트가 없는 시작 단계 메시지 하나를 풉니다. body는 코드 뒤의 바이트입니다.
func startupMsg(code uint32, body []byte) wireMsg {
	switch code {
	case codeSSL:
		return wireMsg{Type: "ssl"}
	case codeGSSENC:
		return wireMsg{Type: "gssenc"}
	case codeCancel:
		m := wireMsg{Type: "cancel"}
		if len(body) >= 4 {
			m.Pid = int(binary.BigEndian.Uint32(body))
		}
		return m
	}
	// StartupMessage는 "이름\0값\0" 쌍이 이어지고 빈 이름으로 끝납니다.
	m := wireMsg{Type: "startup", Params: map[string]string{}}
	if code != codeProto30 && code != codeProto32 {
		m.Params["protocol"] = fmt.Sprintf("%d.%d", code>>16, code&0xffff)
	} else {
		m.Params["protocol"] = fmt.Sprintf("3.%d", code&0xffff)
	}
	r := reader{b: body}
	for {
		k := r.cstr()
		if k == "" {
			break
		}
		m.Params[k] = r.cstr()
	}
	return m
}

// typedMsg 는 형식 바이트가 있는 메시지 하나를 풉니다. 같은 형식 바이트가 방향에 따라 다른 메시지이므로(F의 'D' Describe와
// B의 'D' DataRow 등) 방향을 함께 받습니다. 모르는 형식은 형식 바이트와 길이만 남깁니다.
func typedMsg(dir, t byte, body []byte) wireMsg {
	m := wireMsg{Type: string(t)}
	r := reader{b: body}
	if dir == 'F' {
		switch t {
		case 'Q':
			m.SQL = r.cstr()
		case 'P':
			m.Name, m.SQL = r.cstr(), r.cstr()
			m.Count = int(r.u16())
		case 'B':
			// Bind: 포털 이름, 문장 이름, 인자 형식 수와 형식들, 인자 수. 인자 값은 남기지 않습니다(이 실험의 문장은 인자가 없음).
			portal := r.cstr()
			m.Name = r.cstr()
			r.skip(2 * int(r.u16()))
			m.Count = int(r.u16())
			if portal != "" {
				m.Tag = portal
			}
		case 'E':
			m.Name = r.cstr()
			m.Count = int(r.u32())
		case 'D', 'C':
			m.Status = string(r.byte1())
			m.Name = r.cstr()
		}
		return m
	}
	switch t {
	case 'R':
		v := int(r.u32())
		m.Auth = &v
	case 'S':
		m.Params = map[string]string{r.cstr(): r.cstr()}
	case 'K':
		m.Pid = int(r.u32())
	case 'Z':
		m.Status = string(r.byte1())
	case 'C':
		m.Tag = r.cstr()
	case 'T':
		m.Count = int(r.u16())
		for i := 0; i < m.Count && !r.bad; i++ {
			m.Values = append(m.Values, r.cstr())
			r.skip(18) // 표 oid, 열 번호, 형식 oid, 크기, 형식 수정자, 형식 코드
		}
	case 'D':
		m.Count = int(r.u16())
		for i := 0; i < m.Count && !r.bad; i++ {
			n := int(int32(r.u32()))
			if n < 0 {
				m.Values = append(m.Values, "NULL")
				continue
			}
			v := r.bytes(n)
			if len(v) > maxValueSize {
				v = v[:maxValueSize]
			}
			m.Values = append(m.Values, string(v))
		}
	case 't':
		m.Count = int(r.u16())
	case 'E', 'N':
		// 오류·알림은 (필드 코드 바이트, 문자열) 쌍이 0 바이트로 끝납니다. V는 번역되지 않은 심각도, C는 SQLSTATE,
		// F·L·R은 오류를 낸 서버 소스의 파일, 줄, 함수입니다.
		f := map[byte]string{}
		for {
			k := r.byte1()
			if k == 0 || r.bad {
				break
			}
			f[k] = r.cstr()
		}
		m.Severity, m.Code, m.Message = f['V'], f['C'], f['M']
		if f['F'] != "" {
			m.Where = strings.TrimSpace(fmt.Sprintf("%s:%s %s", f['F'], f['L'], f['R']))
		}
	}
	return m
}

// reader 는 메시지 본문을 앞에서부터 읽습니다. 본문이 모자라면 bad를 켜고 빈 값을 돌려줘, 깨진 메시지가 중계를 멈추지 않게 합니다.
type reader struct {
	b   []byte
	bad bool
}

func (r *reader) cstr() string {
	i := 0
	for i < len(r.b) && r.b[i] != 0 {
		i++
	}
	if i == len(r.b) {
		r.bad = true
		s := string(r.b)
		r.b = nil
		return s
	}
	s := string(r.b[:i])
	r.b = r.b[i+1:]
	return s
}

func (r *reader) bytes(n int) []byte {
	if n > len(r.b) {
		r.bad = true
		n = len(r.b)
	}
	v := r.b[:n]
	r.b = r.b[n:]
	return v
}

func (r *reader) skip(n int) { r.bytes(n) }

func (r *reader) byte1() byte {
	v := r.bytes(1)
	if len(v) < 1 {
		return 0
	}
	return v[0]
}

func (r *reader) u16() uint16 {
	v := r.bytes(2)
	if len(v) < 2 {
		return 0
	}
	return binary.BigEndian.Uint16(v)
}

func (r *reader) u32() uint32 {
	v := r.bytes(4)
	if len(v) < 4 {
		return 0
	}
	return binary.BigEndian.Uint32(v)
}
