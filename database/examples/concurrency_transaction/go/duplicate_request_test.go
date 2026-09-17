package concurrency

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// holdingGateway 타입은 release가 닫힐 때까지 Approve를 붙잡아 두는 테스트용 PG입니다.
//
// 선점한 요청의 PG 호출이 끝나기 전에 나머지 요청이 모두 돌아오도록 강제해, 요청들이 실제로 겹쳤음을
// 타이밍 운 없이 보장합니다. release는 PENDING을 받고 돌아온 요청이 네 개가 되는 순간 테스트가 닫습니다.
// timeout 안에 닫히지 않으면 요청들이 겹치지 않은 것이므로, 그대로 진행해 pendingReturned != 4 검사에서 실패하게 둡니다.
type holdingGateway struct {
	*FakeGateway
	release chan struct{}
	timeout time.Duration
}

// Approve 메서드는 release가 닫히거나 timeout이 지날 때까지 기다린 뒤 FakeGateway.Approve를 부릅니다.
func (g *holdingGateway) Approve(ctx context.Context, key string, amount int64) (string, error) {
	select {
	case <-g.release:
	case <-time.After(g.timeout):
	}
	return g.FakeGateway.Approve(ctx, key, amount)
}

// 같은 키로 온 두 요청이 조회와 기록 사이에서 만나면 PG가 두 번 승인합니다.
func TestPayCheckThenAct_ChargesTwice(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	gateway := NewFakeGateway(0)
	req := PaymentRequest{Key: "order-1", AccountID: "acc", Amount: 30_000}

	meet := NewRendezvous(2, rendezvousTimeout)
	RunTogether(2, func(int) (Payment, error) {
		return PayCheckThenAct(ctx, db, gateway, req, meet.Wait)
	})

	var rows int
	if err := db.QueryRow("SELECT COUNT(*) FROM payments_unguarded WHERE idempotency_key = ?", req.Key).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	t.Logf("pg calls=%d rows=%d", gateway.Calls(), rows)
	if meet.Missed() || gateway.Calls() != 2 || rows != 2 {
		t.Fatalf("이중 승인이 재현되지 않았습니다: missed=%v calls=%d rows=%d", meet.Missed(), gateway.Calls(), rows)
	}
}

// 같은 키로 다섯 요청이 동시에 와도 PG 승인은 한 번이고, 모든 요청이 같은 결제 기록을 받습니다.
//
// 선점한 요청의 PG 호출은 선점에 실패한 네 요청이 모두 PENDING을 받아 돌아올 때까지 붙잡아 둡니다.
// 그래서 결과는 항상 SUCCEEDED 하나, PENDING 넷입니다. 네 요청이 선점한 요청의 결과 기록보다 늦게 왔다면
// SUCCEEDED를 받았을 것이므로, 넷 모두 PENDING을 받았다는 것은 선점한 요청이 PG 응답을 기다리는 동안 들어왔다는 뜻입니다.
// 호출자는 PENDING을 받으면 잠시 뒤 같은 키로 다시 요청해 최종 결과를 받습니다.
func TestPayOnce_SameKeyConcurrentChargesOnce(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	gateway := &holdingGateway{FakeGateway: NewFakeGateway(0), release: make(chan struct{}), timeout: time.Second}
	req := PaymentRequest{Key: "order-1", AccountID: "acc", Amount: 30_000}

	// PENDING을 받고 돌아온 요청이 네 개가 되면 선점한 요청의 PG 호출을 풀어 줍니다.
	var pendingReturned atomic.Int32
	results := RunTogether(5, func(int) (Payment, error) {
		p, err := PayOnce(ctx, db, gateway, req)
		if err == nil && p.Status == StatusPending && pendingReturned.Add(1) == 4 {
			close(gateway.release)
		}
		return p, err
	})

	// 모든 호출이 같은 결제 기록을 받아야 합니다. 상태는 받은 시점에 따라 PENDING이나 SUCCEEDED입니다.
	for _, r := range results {
		t.Logf("worker %d: status=%s err=%v", r.Worker, r.Value.Status, r.Err)
		if r.Err != nil {
			t.Fatalf("worker %d: unexpected error %v", r.Worker, r.Err)
		}
		if r.Value.Key != req.Key || r.Value.AccountID != req.AccountID || r.Value.Amount != req.Amount {
			t.Fatalf("worker %d: 다른 결제 기록을 받았습니다: %+v", r.Worker, r.Value)
		}
		if r.Value.Status != StatusPending && r.Value.Status != StatusSucceeded {
			t.Fatalf("worker %d: want PENDING or SUCCEEDED, got %s", r.Worker, r.Value.Status)
		}
	}
	if gateway.Calls() != 1 {
		t.Fatalf("PG 호출은 1번이어야 합니다: %d", gateway.Calls())
	}
	if pendingReturned.Load() != 4 {
		t.Fatalf("PENDING을 받은 요청이 4개여야 요청들이 겹쳤다고 볼 수 있습니다: %d", pendingReturned.Load())
	}

	// 모든 호출이 끝난 뒤의 재요청은 최종 결과를 그대로 돌려받습니다.
	replay, err := PayOnce(ctx, db, gateway, req)
	if err != nil || replay.Status != StatusSucceeded || gateway.Calls() != 1 {
		t.Fatalf("재요청: status=%s err=%v calls=%d", replay.Status, err, gateway.Calls())
	}
}

// 키가 같아도 금액이나 계좌가 다르면 기존 결과를 돌려주지 않고 거부합니다.
func TestPayOnce_SameKeyDifferentRequestIsRejected(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	gateway := NewFakeGateway(0)

	if _, err := PayOnce(ctx, db, gateway, PaymentRequest{Key: "order-1", AccountID: "acc", Amount: 30_000}); err != nil {
		t.Fatal(err)
	}

	different := []PaymentRequest{
		{Key: "order-1", AccountID: "acc", Amount: 99_999},
		{Key: "order-1", AccountID: "other-acc", Amount: 30_000},
	}
	for _, req := range different {
		_, err := PayOnce(ctx, db, gateway, req)
		if !errors.Is(err, ErrKeyReused) || gateway.Calls() != 1 {
			t.Fatalf("%+v: want ErrKeyReused and 1 call, got err=%v calls=%d", req, err, gateway.Calls())
		}
	}
}

// 호출자의 요청 기한이 PG 응답을 기다리는 도중에 끝나도, 결과는 PENDING이 아니라 UNKNOWN으로 기록됩니다.
//
// 결과를 기록할 때 이미 기한이 끝난 호출자의 ctx를 그대로 쓰면, DB 기록도 실패해 행이 PENDING으로 남습니다.
func TestPayOnce_CallerDeadlineDuringApprovalIsUnknown(t *testing.T) {
	db := newTestDB(t)
	gateway := NewFakeGateway(200 * time.Millisecond)
	req := PaymentRequest{Key: "order-1", AccountID: "acc", Amount: 30_000}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	p, err := PayOnce(ctx, db, gateway, req)
	if err != nil || p.Status != StatusUnknown {
		t.Fatalf("want UNKNOWN without error, got status=%s err=%v", p.Status, err)
	}
}

// PG가 승인해 놓고 응답을 잃어버린 경우입니다.
//
// PayOnce는 FAILED가 아니라 UNKNOWN을 기록하고, 재요청도 PG를 다시 부르지 않습니다.
// Reconcile이 PG에 조회해 SUCCEEDED로 확정한 뒤에야 재요청이 최종 결과를 받습니다.
func TestPayOnce_TimeoutAfterApprovalIsUnknownUntilReconciled(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	gateway := NewFakeGateway(0)
	req := PaymentRequest{Key: "order-1", AccountID: "acc", Amount: 30_000}
	gateway.SetOutcome(req.Key, OutcomeTimeoutAfterApproval)

	first, err := PayOnce(ctx, db, gateway, req)
	if err != nil || first.Status != StatusUnknown {
		t.Fatalf("첫 응답: want UNKNOWN, got status=%s err=%v", first.Status, err)
	}

	retry, err := PayOnce(ctx, db, gateway, req)
	if err != nil || retry.Status != StatusUnknown || gateway.Calls() != 1 {
		t.Fatalf("재요청: want UNKNOWN without new call, got status=%s err=%v calls=%d", retry.Status, err, gateway.Calls())
	}

	settled, err := Reconcile(ctx, db, gateway, req.Key, time.Hour)
	if err != nil || settled.Status != StatusSucceeded || settled.PGTransactionID == "" {
		t.Fatalf("대사: want SUCCEEDED with transaction id, got %+v err=%v", settled, err)
	}
	t.Logf("reconciled: status=%s pg_transaction_id=%s calls=%d", settled.Status, settled.PGTransactionID, gateway.Calls())
}

// PG가 승인하지 않은 채 응답만 잃어버린 경우, 대사는 FAILED로 확정합니다.
func TestReconcile_TimeoutBeforeApprovalBecomesFailed(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	gateway := NewFakeGateway(0)
	req := PaymentRequest{Key: "order-1", AccountID: "acc", Amount: 30_000}
	gateway.SetOutcome(req.Key, OutcomeTimeoutBeforeApproval)

	if p, err := PayOnce(ctx, db, gateway, req); err != nil || p.Status != StatusUnknown {
		t.Fatalf("want UNKNOWN, got status=%s err=%v", p.Status, err)
	}
	settled, err := Reconcile(ctx, db, gateway, req.Key, time.Hour)
	if err != nil || settled.Status != StatusFailed {
		t.Fatalf("want FAILED, got status=%s err=%v", settled.Status, err)
	}
}

// PG가 확정 거절하면 대사 없이 곧바로 FAILED입니다.
func TestPayOnce_DeclineIsFailedImmediately(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	gateway := NewFakeGateway(0)
	req := PaymentRequest{Key: "order-1", AccountID: "acc", Amount: 30_000}
	gateway.SetOutcome(req.Key, OutcomeDecline)

	p, err := PayOnce(ctx, db, gateway, req)
	if err != nil || p.Status != StatusFailed {
		t.Fatalf("want FAILED, got status=%s err=%v", p.Status, err)
	}
}

// 선점 후 PG 승인까지 받고 결과를 기록하기 전에 프로세스가 죽은 경우입니다.
//
// 행은 PENDING으로 남습니다. stuckAfter를 한 시간으로 주면 막 생긴 PENDING은 건드리지 않고, 0으로 주면 모든
// PENDING이 기준보다 오래된 것으로 취급돼 PG에 조회한 뒤 SUCCEEDED로 확정됩니다.
func TestReconcile_ClosesPaymentLeftPendingByCrash(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	gateway := NewFakeGateway(0)
	req := PaymentRequest{Key: "order-1", AccountID: "acc", Amount: 30_000}

	// PayOnce의 앞 두 단계(claim으로 선점, PG Approve)만 실행하고 결과 기록(finish)에 가지 못한 상황을 만듭니다.
	if claimed, err := claim(ctx, db, req); err != nil || !claimed {
		t.Fatalf("claim: claimed=%v err=%v", claimed, err)
	}
	if _, err := gateway.Approve(ctx, req.Key, req.Amount); err != nil {
		t.Fatal(err)
	}

	fresh, err := Reconcile(ctx, db, gateway, req.Key, time.Hour)
	if err != nil || fresh.Status != StatusPending {
		t.Fatalf("막 생긴 PENDING은 그대로여야 합니다: status=%s err=%v", fresh.Status, err)
	}

	settled, err := Reconcile(ctx, db, gateway, req.Key, 0)
	if err != nil || settled.Status != StatusSucceeded {
		t.Fatalf("오래된 PENDING은 SUCCEEDED로 확정돼야 합니다: status=%s err=%v", settled.Status, err)
	}
}

// stuckAfter를 PG 응답 시간보다 짧게 잡으면 생기는 사고를 재현합니다.
//
// Reconcile이 아직 PG 응답을 기다리는 결제를 조회해 FAILED로 확정하고, 곧이어 PG 승인이 도착합니다.
// 선점한 요청의 SUCCEEDED 기록 UPDATE는 ErrStateChanged로 거부되고, 돈은 나갔는데 결제는 FAILED로 남습니다.
// finish가 상태를 조건으로 거는 덕분에 사고가 조용히 덮이지 않고 오류로 드러납니다.
func TestReconcile_TooEarlyConflictsWithInFlightPayment(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	gateway := NewFakeGateway(0)
	req := PaymentRequest{Key: "order-1", AccountID: "acc", Amount: 30_000}

	if claimed, err := claim(ctx, db, req); err != nil || !claimed {
		t.Fatalf("claim: claimed=%v err=%v", claimed, err)
	}

	// PG 응답을 기다리는 도중에 너무 이른 대사가 끼어듭니다.
	early, err := Reconcile(ctx, db, gateway, req.Key, 0)
	if err != nil || early.Status != StatusFailed {
		t.Fatalf("early reconcile: status=%s err=%v", early.Status, err)
	}

	// 그 뒤에 승인이 도착하고, 선점한 요청이 결과를 기록하려 합니다.
	transactionID, err := gateway.Approve(ctx, req.Key, req.Amount)
	if err != nil {
		t.Fatal(err)
	}
	_, err = finish(ctx, db, req.Key, StatusPending, StatusSucceeded, transactionID)
	if !errors.Is(err, ErrStateChanged) {
		t.Fatalf("want ErrStateChanged, got %v", err)
	}
	t.Logf("PG는 %s로 승인했지만 결제는 FAILED로 남았고, 이 충돌은 ErrStateChanged로 드러났습니다", transactionID)
}
