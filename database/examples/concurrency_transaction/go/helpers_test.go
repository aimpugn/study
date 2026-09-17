package concurrency

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// rendezvousTimeout 상수는 약속 지점에서 상대를 기다리는 최대 시간입니다.
//
// 락이 상대를 막는 테스트에서는 락을 쥔 요청이 약속 지점에서 이 시간만큼 기다린 뒤 진행하므로, 너무 길면 테스트가 느려지고
// 너무 짧으면 느린 기계에서 락이 없는데도 상대가 늦게 도착해 Missed가 참이 될 수 있습니다.
const rendezvousTimeout = 300 * time.Millisecond

// newTestDB 함수는 테스트마다 새 임시 디렉터리에 DB를 만들고, 테스트가 끝나면 닫습니다.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := OpenDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// seedAccount 함수는 balance원이 들어 있는 계좌를 만듭니다.
func seedAccount(t *testing.T, db *sql.DB, id string, balance int64) {
	t.Helper()
	if _, err := db.Exec("INSERT INTO accounts (id, balance) VALUES (?, ?)", id, balance); err != nil {
		t.Fatalf("seed account: %v", err)
	}
}

// balanceOf 함수는 계좌의 현재 잔액을 읽습니다.
func balanceOf(t *testing.T, db *sql.DB, id string) int64 {
	t.Helper()
	var balance int64
	if err := db.QueryRow("SELECT balance FROM accounts WHERE id = ?", id).Scan(&balance); err != nil {
		t.Fatalf("read balance: %v", err)
	}
	return balance
}

// outcomeCounts 타입은 동시 실행 결과를 성공, 기대한 업무 오류, 그 밖의 오류로 나눠 셉니다.
//
// 그 밖의 오류를 따로 세는 이유가 있습니다. 잔액 30,000원에서 10,000원 출금 다섯 건을 보내면, BEGIN IMMEDIATE로
// 락을 잡는 구현은 성공 3, 잔액 부족 2가 나옵니다. 락을 잡지 않는 일반 BEGIN 구현은 성공 3, database is locked
// (SQLITE_BUSY) 2가 나옵니다. 성공 수와 최종 잔액(0원)이 똑같으므로, 그 밖의 오류를 따로 세지 않으면 락 없는
// 구현도 테스트를 통과합니다.
type outcomeCounts struct {
	ok       int
	expected int
	other    []error
}

// countOutcomes 함수는 results를 세면서, expected와 errors.Is로 일치하는 오류만 기대한 오류로 칩니다.
func countOutcomes[T any](results []Result[T], expected error) outcomeCounts {
	var c outcomeCounts
	for _, r := range results {
		switch {
		case r.Err == nil:
			c.ok++
		case expected != nil && errors.Is(r.Err, expected):
			c.expected++
		default:
			c.other = append(c.other, r.Err)
		}
	}
	return c
}
