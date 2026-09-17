package concurrency

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

// seedLimit 함수는 한도가 limit원인 고객을 만듭니다.
func seedLimit(t *testing.T, db *sql.DB, customerID string, limit int64) {
	t.Helper()
	if _, err := db.Exec("INSERT INTO daily_limits (customer_id, limit_amount) VALUES (?, ?)", customerID, limit); err != nil {
		t.Fatalf("seed limit: %v", err)
	}
}

// chargedTotal 함수는 고객의 결제 내역 합계를 읽습니다.
func chargedTotal(t *testing.T, db *sql.DB, customerID string) int64 {
	t.Helper()
	var total int64
	if err := db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM charges WHERE customer_id = ?", customerID).Scan(&total); err != nil {
		t.Fatalf("read total: %v", err)
	}
	return total
}

// 한도 10,000원인 고객에게 6,000원 결제 두 건이 겹치면, 각자 합계 0원을 보고 둘 다 추가해 12,000원이 됩니다.
//
// 두 요청은 같은 행을 건드리지 않았으므로 어떤 행도 덮어쓰이지 않았습니다. 그래도 규칙은 깨졌습니다.
func TestChargeReadThenInsert_WriteSkewExceedsLimit(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	seedLimit(t, db, "cust", 10_000)

	meet := NewRendezvous(2, rendezvousTimeout)
	results := RunTogether(2, func(int) (struct{}, error) {
		return struct{}{}, ChargeReadThenInsert(ctx, db, "cust", 6_000, meet.Wait)
	})

	counts := countOutcomes(results, ErrLimitExceeded)
	total := chargedTotal(t, db, "cust")
	t.Logf("success=%d total=%d (한도 10000)", counts.ok, total)
	if meet.Missed() || counts.ok != 2 || total != 12_000 {
		t.Fatalf("쓰기 스큐가 재현되지 않았습니다: missed=%v success=%d total=%d", meet.Missed(), counts.ok, total)
	}
}

// 한도 10,000원에 4,000원 다섯 건을 보내면 두 건만 성공하고, 사용액과 내역 합계가 같습니다.
func TestChargeReserved_KeepsLimit(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	seedLimit(t, db, "cust", 10_000)

	results := RunTogether(5, func(int) (struct{}, error) {
		return struct{}{}, ChargeReserved(ctx, db, "cust", 4_000)
	})

	counts := countOutcomes(results, ErrLimitExceeded)
	var used int64
	if err := db.QueryRow("SELECT used_amount FROM daily_limits WHERE customer_id = ?", "cust").Scan(&used); err != nil {
		t.Fatal(err)
	}
	total := chargedTotal(t, db, "cust")
	t.Logf("success=%d exceeded=%d other=%v used=%d total=%d", counts.ok, counts.expected, counts.other, used, total)
	if counts.ok != 2 || counts.expected != 3 || len(counts.other) != 0 || used != 8_000 || total != 8_000 {
		t.Fatalf("기대: 성공 2, 한도 초과 3, 기타 오류 0, 사용액과 합계 8000")
	}
}

// 결제 내역 INSERT가 실패하면 같은 트랜잭션의 사용액 증가도 되돌아가야 합니다.
//
// 테스트에서만 charges에 INSERT를 거부하는 트리거를 만들어 실패를 일으킵니다. UPDATE와 INSERT를
// 따로 자동 커밋하는 구현이라면 사용액만 4,000원 늘어난 채로 남아 이 테스트가 실패합니다.
func TestChargeReserved_RollsBackReservationWhenInsertFails(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	seedLimit(t, db, "cust", 10_000)
	if _, err := db.Exec("CREATE TRIGGER reject_charge BEFORE INSERT ON charges BEGIN SELECT RAISE(ABORT, 'injected failure'); END"); err != nil {
		t.Fatal(err)
	}

	if err := ChargeReserved(ctx, db, "cust", 4_000); err == nil {
		t.Fatal("INSERT 실패가 오류로 돌아와야 합니다")
	}

	var used int64
	if err := db.QueryRow("SELECT used_amount FROM daily_limits WHERE customer_id = ?", "cust").Scan(&used); err != nil {
		t.Fatal(err)
	}
	if used != 0 {
		t.Fatalf("사용액이 롤백되지 않았습니다: used=%d", used)
	}
}

// 한도가 등록되지 않은 고객은 한도 초과가 아니라 고객 없음 오류입니다.
func TestCharge_UnknownCustomer(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	if err := ChargeReserved(ctx, db, "missing", 1_000); !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("ChargeReserved: want ErrCustomerNotFound, got %v", err)
	}
	if err := ChargeReadThenInsert(ctx, db, "missing", 1_000, nil); !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("ChargeReadThenInsert: want ErrCustomerNotFound, got %v", err)
	}
}
