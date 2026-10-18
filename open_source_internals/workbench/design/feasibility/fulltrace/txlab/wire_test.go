package main

import (
	"encoding/binary"
	"reflect"
	"testing"
)

// 시험용 바이트를 만드는 도우미들입니다. 모양은 PostgreSQL 18 문서 「Message Formats」를 따릅니다.

func be32(v int) []byte  { b := make([]byte, 4); binary.BigEndian.PutUint32(b, uint32(v)); return b }
func be16(v int) []byte  { b := make([]byte, 2); binary.BigEndian.PutUint16(b, uint16(v)); return b }
func cz(s string) []byte { return append([]byte(s), 0) }

func cat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// typed 는 형식 바이트가 있는 메시지 하나를 만듭니다.
func typed(t byte, body ...[]byte) []byte {
	b := cat(body...)
	return cat([]byte{t}, be32(len(b)+4), b)
}

// untyped 는 시작 단계 메시지 하나를 만듭니다.
func untyped(code int, body ...[]byte) []byte {
	b := cat(body...)
	return cat(be32(len(b)+8), be32(code), b)
}

// feedAll 은 바이트를 size 크기 조각으로 나눠 먹이고 나온 메시지를 모읍니다. 시각은 조각 순번입니다.
func feedAll(c *wireConn, dir byte, b []byte, size int) []wireMsg {
	var out []wireMsg
	for i := 0; i < len(b); i += size {
		end := min(i+size, len(b))
		out = append(out, c.feed(dir, b[i:end], int64(i/size))...)
	}
	return out
}

// TestFeedCutsStartupTheSameWayAtAnyChunkSize 는 SSLRequest, 거절 바이트, StartupMessage, 인증과 BackendKeyData를 조각 크기와
// 상관없이 같은 메시지로 자르는지 봅니다. 중계기는 운영 체제가 주는 대로 조각을 받으므로, 메시지 경계가 조각 중간에 와도 같아야
// 맥 쪽과 컨테이너 쪽의 같은 순번 메시지를 이을 수 있습니다.
func TestFeedCutsStartupTheSameWayAtAnyChunkSize(t *testing.T) {
	// Given: pgjdbc가 sslmode=prefer로 여는 접속의 앞부분(앱 쪽)과 trust 인증 서버의 답(서버 쪽)
	front := cat(untyped(codeSSL), untyped(codeProto30, cz("user"), cz("postgres"), cz("database"), cz("postgres"), []byte{0}))
	back := cat([]byte{'N'}, typed('R', be32(0)), typed('S', cz("server_version"), cz("18.6")),
		typed('K', be32(4242), be32(99)), typed('Z', []byte{'I'}))

	for _, size := range []int{1, 3, 7, len(front) + len(back)} {
		// When: 같은 순서로(앱 쪽 요청을 먼저) 조각 크기만 바꿔 먹임
		c := &wireConn{id: 1}
		f := feedAll(c, 'F', front[:8], size)
		b := feedAll(c, 'B', back[:1], size)
		f = append(f, feedAll(c, 'F', front[8:], size)...)
		b = append(b, feedAll(c, 'B', back[1:], size)...)

		// Then
		var ft, bt []string
		for _, m := range f {
			ft = append(ft, m.Type)
		}
		for _, m := range b {
			bt = append(bt, m.Type)
		}
		if !reflect.DeepEqual(ft, []string{"ssl", "startup"}) || !reflect.DeepEqual(bt, []string{"sslreply", "R", "S", "K", "Z"}) {
			t.Fatalf("조각 %d: 앱 쪽 %v, 서버 쪽 %v", size, ft, bt)
		}
		if f[1].Params["user"] != "postgres" || f[1].Params["protocol"] != "3.0" {
			t.Errorf("조각 %d: 시작 메시지 인자 %v", size, f[1].Params)
		}
		if b[3].Pid != 4242 || b[4].Status != "I" || b[1].Auth == nil || *b[1].Auth != 0 {
			t.Errorf("조각 %d: K pid %d, Z %q, R %v", size, b[3].Pid, b[4].Status, b[1].Auth)
		}
		if b[4].N != 5 || f[1].N != 2 {
			t.Errorf("조각 %d: 순번 B %d, F %d; want 5, 2", size, b[4].N, f[1].N)
		}
	}
}

// TestFeedReadsExtendedQueryAndError 는 확장 질의 한 묶음(Parse, Bind, Describe, Execute, Sync)과 오류 응답의 필드를 푸는지 봅니다.
// 오류의 F·L·R 필드는 서버 소스의 자리라서, 화면이 오류를 그 소스 줄로 이을 때 씁니다.
func TestFeedReadsExtendedQueryAndError(t *testing.T) {
	// Given: 이름 없는 문장의 한 묶음과, 락 시간 초과 오류
	front := cat(
		typed('P', cz(""), cz("UPDATE acct SET balance = balance - 10 WHERE id = 1"), be16(0)),
		typed('B', cz(""), cz(""), be16(0), be16(0), be16(0)),
		typed('D', []byte{'P'}, cz("")),
		typed('E', cz(""), be32(1)),
		typed('S'),
	)
	back := typed('E', []byte{'S'}, cz("오류"), []byte{'V'}, cz("ERROR"), []byte{'C'}, cz("55P03"),
		[]byte{'M'}, cz("canceling statement due to lock timeout"), []byte{'F'}, cz("postgres.c"), []byte{'L'}, cz("3370"),
		[]byte{'R'}, cz("ProcessInterrupts"), []byte{0})
	c := &wireConn{id: 2, typedF: true}

	// When
	f := c.feed('F', front, 10)
	b := c.feed('B', back, 11)

	// Then
	var ft []string
	for _, m := range f {
		ft = append(ft, m.Type)
	}
	if !reflect.DeepEqual(ft, []string{"P", "B", "D", "E", "S"}) {
		t.Fatalf("앱 쪽 %v", ft)
	}
	if f[0].SQL != "UPDATE acct SET balance = balance - 10 WHERE id = 1" || f[2].Status != "P" || f[3].Count != 1 {
		t.Errorf("Parse %q, Describe %q, Execute 행 수 %d", f[0].SQL, f[2].Status, f[3].Count)
	}
	if len(b) != 1 || b[0].Severity != "ERROR" || b[0].Code != "55P03" || b[0].Where != "postgres.c:3370 ProcessInterrupts" {
		t.Fatalf("오류 %+v", b)
	}
	if b[0].Conn != 2 || b[0].Dir != "B" || b[0].Ns != 11 {
		t.Errorf("연결 %d, 방향 %s, 시각 %d", b[0].Conn, b[0].Dir, b[0].Ns)
	}
}

// TestFeedStopsAfterEncryptionIsAccepted 는 서버가 SSLRequest를 받아들이면('S') 그 뒤 바이트를 메시지로 자르지 않는지 봅니다.
// 암호문을 메시지로 잘못 자르면 엉뚱한 형식과 길이가 기록에 남습니다.
func TestFeedStopsAfterEncryptionIsAccepted(t *testing.T) {
	// Given: SSLRequest와 수락 바이트, 그 뒤의 암호문
	c := &wireConn{id: 3}
	c.feed('F', untyped(codeSSL), 1)

	// When
	b := c.feed('B', []byte{'S', 0x16, 0x03, 0x01, 0x00, 0x05}, 2)
	more := c.feed('F', []byte{0x16, 0x03, 0x01}, 3)

	// Then
	if len(b) != 1 || b[0].Type != "opaque" || b[0].Status != "S" || len(more) != 0 {
		t.Fatalf("수락 뒤: %+v, 그다음 %+v", b, more)
	}
}
