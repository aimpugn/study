package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

// TestBridgeLive 는 맥의 TCP 접속이 bridge와 컨테이너 안 relay를 거쳐 실제 기록 서버에 닿는지, 두 쪽이 같은 메시지를 같은 순번으로
// 적는지, 두 시계의 차이가 인과와 맞는 구간 안에 있는지 봅니다.
//
// 시계: 앱 → 서버 메시지는 맥이 먼저 읽고 컨테이너가 나중에 읽으므로 (컨테이너 시각 − 맥 시각) = 차이 + 지연(≥ 0)이고,
// 서버 → 앱 메시지는 컨테이너가 먼저이므로 (컨테이너 시각 − 맥 시각) = 차이 − 지연입니다. 그래서 두 시계의 차이는
// [서버 → 앱 값의 최댓값, 앱 → 서버 값의 최솟값] 안에 있어야 합니다. 이 구간이 비면 한쪽 기록의 시각이나 순번이 틀린 것입니다.
//
// 실습 컨테이너가 필요하므로 TXLAB_LIVE=1일 때만 돕니다. 그 전에 이 빌드를 linux/arm64로 /tmp/txlab에 넣고 기록 서버를 띄워 둡니다.
func TestBridgeLive(t *testing.T) {
	if os.Getenv("TXLAB_LIVE") == "" {
		t.Skip("TXLAB_LIVE=1일 때만 돕니다(pg-lab에 기록 서버와 이 빌드의 /tmp/txlab이 있어야 함)")
	}
	// Given: 맥의 듣는 자리와, 받은 접속을 컨테이너에 잇는 bridge
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	const id = 9
	var mu sync.Mutex
	var mac []wireMsg
	done := make(chan error, 1)
	go func() {
		tc, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		done <- bridge(tc, id, t.TempDir()+"/wire-mac.jsonl", func(m wireMsg) {
			mu.Lock()
			mac = append(mac, m)
			mu.Unlock()
		})
	}()

	// When: 날 클라이언트가 SSLRequest → 거절 바이트 → StartupMessage → 준비, SELECT 1 → 준비, Terminate
	c, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c.SetDeadline(time.Now().Add(30 * time.Second))
	r := bufio.NewReader(c)
	c.Write(untyped(codeSSL))
	if b, err := r.ReadByte(); err != nil || b != 'N' {
		t.Fatalf("SSLRequest의 답 %q, %v; want N", b, err)
	}
	c.Write(untyped(codeProto30, cz("user"), cz("postgres"), cz("database"), cz("postgres"),
		cz("application_name"), cz("txlab-bridge-test"), []byte{0}))
	readUntilReady(t, r)
	c.Write(typed('Q', cz("SELECT 1")))
	readUntilReady(t, r)
	c.Write(typed('X'))
	c.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("bridge: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("bridge가 30초 안에 끝나지 않았습니다")
	}

	// Then: 맥 쪽 메시지의 꼴
	types := func(ms []wireMsg, dir string) []string {
		var out []string
		for _, m := range ms {
			if m.Dir == dir {
				out = append(out, m.Type)
			}
		}
		return out
	}
	front := types(mac, "F")
	if want := []string{"ssl", "startup", "Q", "X"}; !equalStrings(front, want) {
		t.Fatalf("맥 쪽 앱 → 서버 %v; want %v", front, want)
	}
	pid := 0
	for _, m := range mac {
		if m.Type == "K" {
			pid = m.Pid
		}
	}
	if pid == 0 {
		t.Fatalf("BackendKeyData가 없습니다: %v", types(mac, "B"))
	}

	// Then: 컨테이너 쪽 기록이 같은 메시지를 같은 순번으로 적었고, 두 시계의 차이 구간이 비지 않음
	out, err := exec.Command("docker", "exec", container, "cat", labDir+"/out/wire-9.jsonl").Output()
	if err != nil {
		t.Fatal(err)
	}
	var box []wireMsg
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var m wireMsg
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		box = append(box, m)
	}
	if len(box) != len(mac) {
		t.Fatalf("컨테이너 쪽 %d개, 맥 쪽 %d개", len(box), len(mac))
	}
	key := func(m wireMsg) [2]any { return [2]any{m.Dir + m.Type, m.N} }
	byKey := map[[2]any]wireMsg{}
	for _, m := range box {
		byKey[key(m)] = m
	}
	lo, hi := int64(-1<<62), int64(1<<62)
	for _, m := range mac {
		b, ok := byKey[key(m)]
		if !ok || b.Len != m.Len {
			t.Fatalf("맥 쪽 %s%s #%d(%d바이트)가 컨테이너 쪽에 없거나 길이가 다릅니다", m.Dir, m.Type, m.N, m.Len)
		}
		d := b.Ns - m.Ns
		if m.Dir == "F" {
			hi = min(hi, d)
		} else {
			lo = max(lo, d)
		}
	}
	if lo > hi {
		t.Fatalf("시계 차이 구간이 비었습니다: [%d, %d] ns", lo, hi)
	}
	t.Logf("백엔드 pid %d, 메시지 %d개, 컨테이너 − 맥 시계 차이 ∈ [%.3f, %.3f] ms", pid, len(mac), float64(lo)/1e6, float64(hi)/1e6)
}

// readUntilReady 는 형식 있는 메시지를 ReadyForQuery('Z')까지 읽어 버립니다.
func readUntilReady(t *testing.T, r *bufio.Reader) {
	t.Helper()
	for {
		h := make([]byte, 5)
		if _, err := io.ReadFull(r, h); err != nil {
			t.Fatal(err)
		}
		body := make([]byte, int(binary.BigEndian.Uint32(h[1:]))-4)
		if _, err := io.ReadFull(r, body); err != nil {
			t.Fatal(err)
		}
		if h[0] == 'E' {
			t.Fatalf("서버 오류: %q", body)
		}
		if h[0] == 'Z' {
			return
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
