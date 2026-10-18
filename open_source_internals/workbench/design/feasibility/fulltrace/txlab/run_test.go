package main

import "testing"

// TestReadyCountCountsEachRestart 는 처음 뜰 때와 크래시 복구 뒤에 접속을 받기 시작한 줄을 따로 세는지 봅니다. kill 뒤에는
// 이 수가 kill 앞보다 늘어야 서버가 다시 떴다고 봅니다.
func TestReadyCountCountsEachRestart(t *testing.T) {
	// Given: 처음 뜬 로그와, 그 뒤 백엔드 하나가 죽어 다시 뜬 로그(PostgreSQL 18 로그 문구)
	first := "2026-10-14 13:02:49.516 UTC [161] LOG:  database system is ready to accept connections\n"
	crashed := first +
		"2026-10-14 13:03:01.000 UTC [161] LOG:  client backend (PID 185) was terminated by signal 9: Killed\n" +
		"2026-10-14 13:03:01.010 UTC [161] LOG:  all server processes terminated; reinitializing\n" +
		"2026-10-14 13:03:01.300 UTC [161] LOG:  database system is ready to accept connections\n"

	// When, Then
	if n := readyCount([]byte(first)); n != 1 {
		t.Fatalf("처음 뜬 로그: %d; want 1", n)
	}
	if n := readyCount([]byte(crashed)); n != 2 {
		t.Fatalf("다시 뜬 로그: %d; want 2", n)
	}
}
