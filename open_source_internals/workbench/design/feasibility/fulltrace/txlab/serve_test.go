package main

import (
	"reflect"
	"testing"
)

// TestIsreadyStartupFindsBothLineForms 는 pg_isready의 시작 메시지가 한 줄로 끝난 recvfrom에 있든, 다른 pid의 줄이 끼어
// unfinished·resumed 두 줄로 나뉘었든 그 백엔드를 찾는지 봅니다. 2026-10-14 실험 2의 2-for-update/b에서 시작 메시지가
// resumed 줄에만 있어 그 백엔드의 함수 기록 95 MB가 지워지지 않고 남았습니다.
func TestIsreadyStartupFindsBothLineForms(t *testing.T) {
	// Given: 같은 시작 메시지를 받은 두 꼴과, 같은 호출의 unfinished 줄
	whole := `440   1791982950.100000 recvfrom(9<UNIX-STREAM:[1->2,"/tmp/pgi/.s.PGSQL.5433"]>, "\0\0\0E\0\3\0\0user\0postgres\0database\0postgres\0application_name\0pg_isready\0\0", 8192, 0, NULL, NULL) = 69 <0.000010>`
	unfinished := `529   1791983052.548465 recvfrom(9<UNIX-STREAM:[61458->61457,"/tmp/pgi/.s.PGSQL.5433"]>,  <unfinished ...>`
	resumed := `529   1791983052.548644 <... recvfrom resumed>"\0\0\0E\0\3\0\0user\0postgres\0database\0postgres\0application_name\0pg_isready\0\0", 8192, 0, NULL, NULL) = 69 <0.000072>`
	psql := `539   1791983054.000000 <... recvfrom resumed>"\0\0\0?\0\3\0\0user\0postgres\0database\0postgres\0application_name\0psql\0\0", 8192, 0, NULL, NULL) = 63 <0.000010>`

	for _, c := range []struct {
		name string
		line string
		pid  int
		ok   bool
	}{
		{"한 줄", whole, 440, true},
		{"resumed 줄", resumed, 529, true},
		{"unfinished 줄에는 메시지가 없음", unfinished, 0, false},
		{"psql 접속은 남김", psql, 0, false},
	} {
		// When
		pid, ok := isreadyStartup(c.line)

		// Then
		if pid != c.pid || ok != c.ok {
			t.Errorf("%s: isreadyStartup = %d, %v; want %d, %v", c.name, pid, ok, c.pid, c.ok)
		}
	}
}

// TestSetupBatchesKeepsOneTransactionExceptAlterSystem 은 준비 SQL을 한 트랜잭션으로 묶되 ALTER SYSTEM만 따로 내는지 봅니다.
// 묶음이 깨지면 준비가 번호를 더 써서 세션의 번호가 실험 1·2와 달라지고, ALTER SYSTEM을 묶으면 준비가 실패합니다.
func TestSetupBatchesKeepsOneTransactionExceptAlterSystem(t *testing.T) {
	// Given: 실험 3-crash-sync-off의 준비 줄
	setup := []string{
		"CREATE TABLE acct(id int PRIMARY KEY, owner text, balance int);",
		"INSERT INTO acct VALUES (1, 'kim', 100);",
		"ALTER SYSTEM SET wal_writer_delay = '10s';",
	}

	// When
	got := setupBatches(setup)

	// Then
	want := []string{
		"CREATE EXTENSION pageinspect;\nCREATE TABLE acct(id int PRIMARY KEY, owner text, balance int);\nINSERT INTO acct VALUES (1, 'kim', 100);",
		"ALTER SYSTEM SET wal_writer_delay = '10s';",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("setupBatches = %q; want %q", got, want)
	}
}
