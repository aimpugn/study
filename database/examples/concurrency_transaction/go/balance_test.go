package concurrency

import (
	"context"
	"errors"
	"testing"
)

// 잔액 20,000원에서 5,000원 출금 두 건이 겹치면, 둘 다 성공했는데 잔액은 15,000원이 됩니다.
//
// 두 건이 차례로 실행됐다면 10,000원이어야 합니다. 출금 5,000원 한 건은 성공으로 응답했는데 잔액에는 반영되지 않았습니다.
// 출금액을 잔액과 같게 잡으면 틀린 결과(0원)가 정상 결과처럼 보이므로 금액을 이렇게 고릅니다.
func TestWithdrawReadModifyWrite_LosesUpdate(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	seedAccount(t, db, "acc", 20_000)

	// 두 요청이 모두 잔액을 읽은 뒤에야 쓰도록 약속 지점에서 만나게 합니다.
	meet := NewRendezvous(2, rendezvousTimeout)
	results := RunTogether(2, func(int) (struct{}, error) {
		return struct{}{}, WithdrawReadModifyWrite(ctx, db, "acc", 5_000, meet.Wait)
	})

	counts := countOutcomes(results, nil)
	got := balanceOf(t, db, "acc")
	t.Logf("success=%d balance=%d (차례로 실행했다면 10000)", counts.ok, got)

	if meet.Missed() {
		t.Fatal("두 요청이 틈에서 만나지 못해 재현 조건이 성립하지 않았습니다")
	}
	if counts.ok != 2 || got != 15_000 {
		t.Fatalf("갱신 손실이 재현되지 않았습니다: success=%d balance=%d other=%v", counts.ok, got, counts.other)
	}
}

// 서버 안의 뮤텍스는 같은 서버로 온 요청만 줄 세웁니다.
//
// 같은 Server라면 두 번째 요청이 뮤텍스에 막혀 약속 지점에 오지 못하고(Missed), 결과가 맞습니다.
// Server가 둘이면 두 요청이 틈에서 만나고 갱신 손실이 그대로 납니다.
func TestServerWithdraw_LocalLockProtectsOnlyOneServer(t *testing.T) {
	ctx := context.Background()

	t.Run("같은 서버", func(t *testing.T) {
		db := newTestDB(t)
		seedAccount(t, db, "acc", 20_000)
		server := &Server{}

		meet := NewRendezvous(2, rendezvousTimeout)
		RunTogether(2, func(int) (struct{}, error) {
			return struct{}{}, server.Withdraw(ctx, db, "acc", 5_000, meet.Wait)
		})

		got := balanceOf(t, db, "acc")
		t.Logf("missed=%v balance=%d", meet.Missed(), got)
		if !meet.Missed() || got != 10_000 {
			t.Fatalf("같은 서버의 뮤텍스가 요청을 줄 세우지 못했습니다: missed=%v balance=%d", meet.Missed(), got)
		}
	})

	t.Run("서버 두 대", func(t *testing.T) {
		db := newTestDB(t)
		seedAccount(t, db, "acc", 20_000)
		servers := []*Server{{}, {}}

		meet := NewRendezvous(2, rendezvousTimeout)
		RunTogether(2, func(worker int) (struct{}, error) {
			return struct{}{}, servers[worker].Withdraw(ctx, db, "acc", 5_000, meet.Wait)
		})

		got := balanceOf(t, db, "acc")
		t.Logf("missed=%v balance=%d", meet.Missed(), got)
		if meet.Missed() || got != 15_000 {
			t.Fatalf("서버 두 대에서 갱신 손실이 재현되지 않았습니다: missed=%v balance=%d", meet.Missed(), got)
		}
	})
}

// 트랜잭션을 열었을 뿐 락을 잡지 않았으므로 두 요청은 틈에서 만납니다.
//
// SQLite는 먼저 쓴 쪽만 성공시키고 나머지를 SQLITE_BUSY 계열 오류로 거부합니다. 잔액은 맞지만
// 요청 하나가 오류로 끝나므로, 이 방식을 쓰려면 오류가 난 트랜잭션을 처음부터 재시도해야 합니다.
func TestWithdrawInPlainTx_TransactionAloneDoesNotLock(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	seedAccount(t, db, "acc", 20_000)

	meet := NewRendezvous(2, rendezvousTimeout)
	results := RunTogether(2, func(int) (struct{}, error) {
		return struct{}{}, WithdrawInPlainTx(ctx, db, "acc", 5_000, meet.Wait)
	})

	busy := 0
	for _, r := range results {
		if IsBusy(r.Err) {
			busy++
		}
		t.Logf("worker %d: err=%v", r.Worker, r.Err)
	}
	counts := countOutcomes(results, nil)
	got := balanceOf(t, db, "acc")

	if meet.Missed() {
		t.Fatal("두 트랜잭션이 틈에서 만나지 못했습니다. 트랜잭션이 락을 잡았다면 이렇게 됩니다")
	}
	if counts.ok != 1 || busy != 1 || got != 15_000 {
		t.Fatalf("기대: 성공 1, busy 1, 잔액 15000 / 실제: 성공 %d, busy %d, 잔액 %d", counts.ok, busy, got)
	}
}

// 조건부 UPDATE는 틈이 없으므로 약속 지점 없이 다섯 건을 한꺼번에 보내도 한 건만 성공합니다.
func TestWithdrawConditional_RejectsOverdraft(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	seedAccount(t, db, "acc", 10_000)

	results := RunTogether(5, func(int) (struct{}, error) {
		return struct{}{}, WithdrawConditional(ctx, db, "acc", 10_000)
	})

	counts := countOutcomes(results, ErrInsufficientBalance)
	got := balanceOf(t, db, "acc")
	t.Logf("success=%d insufficient=%d other=%v balance=%d", counts.ok, counts.expected, counts.other, got)
	if counts.ok != 1 || counts.expected != 4 || len(counts.other) != 0 || got != 0 {
		t.Fatalf("기대: 성공 1, 잔액 부족 4, 기타 오류 0, 잔액 0")
	}

	if err := WithdrawConditional(ctx, db, "missing", 1); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("없는 계좌: want ErrAccountNotFound, got %v", err)
	}
}

// 비관적 락은 두 가지로 확인합니다.
//
// 첫째, 락을 쥔 요청이 멈춘 동안 다른 요청은 BEGIN IMMEDIATE에서 막혀 약속 지점에 오지 못합니다(Missed).
// 둘째, 다섯 건을 보내면 성공과 잔액 부족만 나오고 그 밖의 오류는 0건입니다. 락을 잡지 않는 일반 BEGIN 구현도
// 성공 3, 잔액 0원이 나오지만 나머지 둘이 잔액 부족이 아니라 database is locked이므로(outcomeCounts 설명 참고),
// 결과 숫자와 그 밖의 오류 0건을 함께 봅니다.
func TestWithdrawLocked_SerializesRequests(t *testing.T) {
	ctx := context.Background()

	t.Run("락을 쥔 동안 상대가 들어오지 못한다", func(t *testing.T) {
		db := newTestDB(t)
		seedAccount(t, db, "acc", 20_000)

		meet := NewRendezvous(2, rendezvousTimeout)
		results := RunTogether(2, func(int) (struct{}, error) {
			return struct{}{}, WithdrawLocked(ctx, db, "acc", 5_000, meet.Wait)
		})

		counts := countOutcomes(results, nil)
		got := balanceOf(t, db, "acc")
		t.Logf("missed=%v success=%d balance=%d", meet.Missed(), counts.ok, got)
		if !meet.Missed() {
			t.Fatal("두 요청이 락 안에서 만났습니다. 락이 상대를 막지 못했습니다")
		}
		if counts.ok != 2 || got != 10_000 {
			t.Fatalf("기대: 성공 2, 잔액 10000 / 실제: 성공 %d, 잔액 %d, 기타 %v", counts.ok, got, counts.other)
		}
	})

	t.Run("다섯 건 중 세 건만 성공하고 기타 오류가 없다", func(t *testing.T) {
		db := newTestDB(t)
		seedAccount(t, db, "acc", 30_000)

		results := RunTogether(5, func(int) (struct{}, error) {
			return struct{}{}, WithdrawLocked(ctx, db, "acc", 10_000, nil)
		})

		counts := countOutcomes(results, ErrInsufficientBalance)
		got := balanceOf(t, db, "acc")
		t.Logf("success=%d insufficient=%d other=%v balance=%d", counts.ok, counts.expected, counts.other, got)
		if counts.ok != 3 || counts.expected != 2 || len(counts.other) != 0 || got != 0 {
			t.Fatalf("기대: 성공 3, 잔액 부족 2, 기타 오류 0, 잔액 0")
		}
	})
}

// 낙관적 락은 틈을 없애지 않고 쓰는 순간에 충돌을 검출합니다.
//
// 두 요청이 같은 version을 읽고 틈에서 만나면 한쪽의 UPDATE는 0행이 되고, 그 요청은 새로 읽어
// 한 번 더 시도합니다. 그래서 시도 횟수 합은 3이고 잔액은 차례로 실행한 결과와 같습니다.
func TestWithdrawOptimistic_RetriesAfterConflict(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	seedAccount(t, db, "acc", 20_000)

	meet := NewRendezvous(2, rendezvousTimeout)
	results := RunTogether(2, func(int) (int, error) {
		return WithdrawOptimistic(ctx, db, "acc", 5_000, 5, meet.Wait)
	})

	attempts := 0
	for _, r := range results {
		attempts += r.Value
		t.Logf("worker %d: attempts=%d err=%v", r.Worker, r.Value, r.Err)
	}
	counts := countOutcomes(results, nil)
	got := balanceOf(t, db, "acc")

	if meet.Missed() {
		t.Fatal("두 요청이 틈에서 만나지 못해 충돌이 일어나지 않았습니다")
	}
	if counts.ok != 2 || attempts != 3 || got != 10_000 {
		t.Fatalf("기대: 성공 2, 시도 합 3, 잔액 10000 / 실제: 성공 %d, 시도 %d, 잔액 %d", counts.ok, attempts, got)
	}
}
