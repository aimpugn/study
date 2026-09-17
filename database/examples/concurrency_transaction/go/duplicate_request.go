package concurrency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ApproveTimeout 상수는 PayOnce가 PG 승인 응답을 기다리는 최대 시간입니다.
//
// PayOnce는 PG를 부르기 전에 PENDING 행을 넣고(선점), PG 응답을 최대 ApproveTimeout 기다린 뒤 결과 기록에
// 최대 recordTimeout을 씁니다. 그래서 정상 처리 중인 결제도 선점 뒤 최대 두 기한의 합만큼 PENDING으로 남을 수
// 있습니다. Reconcile이 "멈춘 결제"로 판단하는 기준 stuckAfter는 이 합에 여유를 더한 값보다 길어야 합니다.
const ApproveTimeout = 5 * time.Second

// recordTimeout 상수는 PG 응답을 받은 뒤 결과를 DB에 기록하는 데 쓰는 최대 시간입니다.
const recordTimeout = 5 * time.Second

// ErrKeyReused 오류는 이미 쓰인 멱등성 키로 내용이 다른 요청이 들어왔다는 뜻입니다.
var ErrKeyReused = errors.New("idempotency key reused with a different request")

// ErrStateChanged 오류는 이 호출이 결제 상태를 바꾸기 전에 다른 호출(대사 작업 Reconcile이나 같은 키를 처리한
// 다른 요청)이 결제 상태를 먼저 바꿨다는 뜻입니다.
var ErrStateChanged = errors.New("payment state changed by another actor")

// PaymentStatus 타입은 결제 한 건이 어디까지 진행됐는지 나타냅니다.
type PaymentStatus string

const (
	// StatusPending 값은 이 멱등성 키로 PG를 부를 권리를 한 요청이 가져갔고(선점), PG 결과는 아직 기록하지 못했다는 뜻입니다.
	StatusPending PaymentStatus = "PENDING"
	// StatusSucceeded 값은 PG 승인을 확인했다는 뜻입니다.
	StatusSucceeded PaymentStatus = "SUCCEEDED"
	// StatusFailed 값은 PG가 거절했거나, 승인되지 않았음을 조회로 확인했다는 뜻입니다.
	StatusFailed PaymentStatus = "FAILED"
	// StatusUnknown 값은 PG에 요청을 보냈지만 결과를 모른다는 뜻입니다. 실패가 아닙니다.
	StatusUnknown PaymentStatus = "UNKNOWN"
)

// PaymentRequest 타입은 클라이언트가 보낸 결제 요청입니다.
//
// Key는 클라이언트가 요청마다 만들어 재시도할 때도 그대로 보내는 멱등성 키입니다.
// 같은 Key로 두 번 온 요청은 "같은 결제를 다시 요청한 것"으로 봅니다.
type PaymentRequest struct {
	Key       string
	AccountID string
	Amount    int64
}

// Payment 타입은 DB에 기록된 결제 한 건입니다.
type Payment struct {
	Key             string
	AccountID       string
	Amount          int64
	Status          PaymentStatus
	PGTransactionID string
	UpdatedAt       time.Time
}

// ============================================================================
// 깨지는 구현
// ============================================================================

// PayCheckThenAct 함수는 같은 키의 결제가 있는지 조회하고(check), 없으면 PG를 부른 뒤 기록합니다(act).
// 동시에 부르면 PG를 여러 번 부르는 안티패턴입니다.
//
// 조회와 기록 사이에 PG 호출이 끼어 있어 틈이 수십에서 수백 밀리초나 됩니다. 두 번째 요청의 조회가
// 첫 요청의 조회와 기록 사이에 들어오면, 둘 다 "없음"을 보고 둘 다 PG를 부릅니다. payments_unguarded의
// idempotency_key에 UNIQUE를 걸어도 늦습니다. 두 번째 요청의 INSERT가 제약에 걸려 실패하는 시점에는 그 요청의
// PG 승인이 이미 끝났습니다.
//
// PG가 승인하고 기록까지 마치면 그 결제를 돌려줍니다. 이미 같은 키의 행이 있으면 PG를 부르지 않고, 저장된 행을
// 읽지 않은 채 요청 값(req)에 Status SUCCEEDED를 붙여 돌려줍니다.
// 조회, PG 승인, 기록 중 하나가 실패하면 그 오류를 감싸서 돌려줍니다.
func PayCheckThenAct(ctx context.Context, db *sql.DB, gateway Gateway, req PaymentRequest, pause Pause) (Payment, error) {
	// 같은 멱등성 키(req.Key)의 기록이 payments_unguarded에 몇 건 있는지 셉니다.
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM payments_unguarded WHERE idempotency_key = ?", req.Key).Scan(&count)
	if err != nil {
		return Payment{}, fmt.Errorf("check existing payment: %w", err)
	}
	if count > 0 {
		return Payment{Key: req.Key, AccountID: req.AccountID, Amount: req.Amount, Status: StatusSucceeded}, nil
	}

	// ------------------------------------------------------------------------
	// 주의: 여기가 틈입니다. 이 틈은 아래 PG 호출이 끝나고 INSERT할 때까지 이어집니다.
	// 그동안 들어온 같은 req.Key의 요청도 count == 0을 보고 PG를 부릅니다.
	// ------------------------------------------------------------------------
	pause.here()

	// 틈에 들어온 요청 수만큼 PG가 승인합니다.
	transactionID, err := gateway.Approve(ctx, req.Key, req.Amount)
	if err != nil {
		return Payment{}, fmt.Errorf("approve: %w", err)
	}

	// 결제 기록(payments_unguarded 행)은 PG 승인으로 돈이 움직인 뒤에야 남습니다.
	_, err = db.ExecContext(ctx,
		"INSERT INTO payments_unguarded (idempotency_key, account_id, amount) VALUES (?, ?, ?)",
		req.Key, req.AccountID, req.Amount)
	if err != nil {
		return Payment{}, fmt.Errorf("record payment: %w", err)
	}
	return Payment{Key: req.Key, AccountID: req.AccountID, Amount: req.Amount,
		Status: StatusSucceeded, PGTransactionID: transactionID}, nil
}

// ============================================================================
// 지키는 구현
// ============================================================================

// PayOnce 함수는 같은 멱등성 키로 몇 번이 동시에 불려도 PG를 한 번만 부르고, 모든 호출에 같은 결제 기록을 돌려줍니다.
//
// PG를 부르기 전에 DB에 PENDING 행을 INSERT해, 이 멱등성 키로 PG를 부를 권리를 먼저 가져갑니다(선점). payments.idempotency_key의
// UNIQUE 제약이 INSERT를 한 건만 성공시키므로 선점에 성공한 요청만 PG를 부르고, 실패한 요청은 이미
// 있는 기록을 읽어 돌려줍니다.
//
// PG 응답은 세 갈래로 기록합니다. 승인이면 SUCCEEDED, 거절이면 FAILED, 그 밖의 오류는 UNKNOWN입니다.
// ApproveTimeout 안에 응답이 없거나 호출자의 ctx가 끝나도 UNKNOWN입니다.
//
// 반환한 Payment.Status를 확인해야 합니다. 같은 키를 다른 요청이 아직 처리 중이면 PENDING이, PG 결과를 모르면
// UNKNOWN이 오류 없이 돌아옵니다. PENDING을 받은 호출자는 잠시 뒤 같은 키로 다시 요청하고, UNKNOWN은
// Reconcile이 확정할 때까지 기다립니다.
//
// 같은 키에 내용이 다른 요청이면 ErrKeyReused를 돌려줍니다. PG 응답을 기다리는 동안 Reconcile이 이 결제를
// FAILED로 먼저 확정했다면, 이 함수가 결과를 기록하려는 UPDATE는 0행이 되어 ErrStateChanged를 돌려줍니다. 이때
// PG가 승인했다면 돈은 나갔는데 결제는 FAILED로 남은 것이므로 호출자가 경보와 환불 대상으로 다뤄야 하며, 이 예제는
// 그 처리를 구현하지 않습니다. DB 기록 자체가 실패하면 그 오류를 돌려주고, 행은 PENDING으로 남았다가 stuckAfter가
// 지난 뒤 Reconcile이 PG에 조회해 확정합니다.
func PayOnce(ctx context.Context, db *sql.DB, gateway Gateway, req PaymentRequest) (Payment, error) {
	// ------------------------------------------------------------------------
	// 핵심: PG를 부르기 전에 req.Key로 PENDING 행을 INSERT해, 이 키로 PG를 부를 권리부터 가져갑니다(선점).
	// payments.idempotency_key의 UNIQUE 제약이 같은 req.Key의 INSERT를 한 건만 통과시키고,
	// 이 INSERT가 PG 호출보다 먼저 실행되므로 PG도 한 번만 불립니다.
	// ------------------------------------------------------------------------
	claimed, err := claim(ctx, db, req)
	if err != nil {
		return Payment{}, err
	}

	// 선점하지 못했다면(claimed == false) 다른 요청이 같은 req.Key를 이미 처리 중이거나 처리를 마쳤습니다.
	// PG를 부르지 않고 저장된 기록(existing)을 읽습니다.
	if !claimed {
		existing, err := loadPayment(ctx, db, req.Key)
		if err != nil {
			return Payment{}, err
		}
		// 키(req.Key)가 같아도 계좌(AccountID)나 금액(Amount)이 저장된 기록(existing)과 다르면, 재시도가 아니라
		// 키를 잘못 재사용한 요청입니다. existing을 그대로 돌려주면 호출자는 자기가 보낸 금액으로 결제됐다고 오해합니다.
		if existing.AccountID != req.AccountID || existing.Amount != req.Amount {
			return existing, ErrKeyReused
		}
		return existing, nil
	}

	// 선점에 성공한 요청만 PG를 부릅니다. PG 응답 대기 시간에 ApproveTimeout이라는 상한을 둬야, Reconcile이
	// "이만큼 지나도 PENDING이면 멈춘 결제다"라고 판단할 기준(stuckAfter)을 정할 수 있습니다.
	approveCtx, cancelApprove := context.WithTimeout(ctx, ApproveTimeout)
	transactionID, err := gateway.Approve(approveCtx, req.Key, req.Amount)
	cancelApprove()

	// ------------------------------------------------------------------------
	// 핵심: PG의 확정 거절(ErrDeclined)만 FAILED입니다. 타임아웃(ErrTimeout), 연결 끊김, 기한 초과처럼
	// 결과를 모르는 오류는 모두 UNKNOWN으로 기록합니다. PG가 승인해 놓고 응답만 잃었을 수 있기 때문입니다.
	// ------------------------------------------------------------------------
	status := StatusSucceeded
	switch {
	case err == nil:
	case errors.Is(err, ErrDeclined):
		status = StatusFailed
	default:
		status = StatusUnknown
	}

	// 호출자의 ctx가 끝났더라도 PG 응답에 따른 status는 payments에 기록해야 합니다. 끝난 ctx로 기록하면
	// UPDATE가 실패해 행이 UNKNOWN이 아니라 PENDING으로 남습니다. context.WithoutCancel(ctx)로 만든 문맥은 ctx의
	// 값은 이어받지만 ctx가 취소되거나 기한이 지나도 함께 끝나지 않으므로, 여기에 기록용 기한(recordTimeout)을 따로 줍니다.
	recordCtx, cancelRecord := context.WithTimeout(context.WithoutCancel(ctx), recordTimeout)
	defer cancelRecord()
	return finish(recordCtx, db, req.Key, StatusPending, status, transactionID)
}

// Reconcile 함수는 결과를 모르는 결제를 PG에 조회해 SUCCEEDED나 FAILED로 확정합니다(대사, reconciliation).
//
// 대상은 UNKNOWN 결제와, DB 시각 기준으로 stuckAfter보다 오래 PENDING으로 남은 결제입니다.
// 오래된 PENDING은 PG를 부른 프로세스가 결과를 기록하기 전에 죽었을 가능성이 큽니다.
//
// stuckAfter는 ApproveTimeout + recordTimeout에 여유를 더한 값보다 길어야 합니다(ApproveTimeout 설명 참고).
// 짧으면 다음 순서로 사고가 납니다. 1) 선점한 요청이 PG 응답을 기다리는 동안 2) Reconcile이 PG에 조회하고,
// 아직 승인이 없으니 FAILED로 확정합니다. 3) 그 뒤 승인 응답이 오면 선점한 요청의 SUCCEEDED 기록 UPDATE가
// ErrStateChanged로 거부됩니다.
//
// stuckAfter를 길게 잡아도 막지 못하는 순서가 있습니다. 1) 선점 직후 프로세스가 가비지 컬렉터 실행(GC 정지) 같은
// 이유로 stuckAfter보다 오래 멈추고 2) 그사이 Reconcile이 FAILED로 확정한 뒤 3) 프로세스가 깨어나 PG를 부르는
// 경우입니다. 시간 기준으로는 이 경우를 구분할 수 없으므로, 완전한 방어는 PG 쪽에 있습니다. PG가 같은 멱등성 키의
// 승인을 한 번만 처리하거나, 확정 뒤 도착한 승인을 취소하는 기능을 제공해야 합니다.
//
// 승인 기록이 없으면 FAILED로 확정합니다. 이 판단은 "조회에 없는 요청은 이후에도 승인되지 않는다"는
// PG의 보장을 전제로 합니다.
//
// 확정했거나 이미 확정돼 있던 결제를 돌려줍니다. 결제가 PENDING이고 선점한 지 stuckAfter가 아직 지나지 않았으면 조회하지 않고 현재 기록을
// 돌려줍니다. 결제 기록이 없으면 sql.ErrNoRows를 감싼 오류를, PG 조회가 실패하면 그 오류를 돌려줍니다.
func Reconcile(ctx context.Context, db *sql.DB, gateway Gateway, key string, stuckAfter time.Duration) (Payment, error) {
	current, err := loadPayment(ctx, db, key)
	if err != nil {
		return Payment{}, err
	}

	// 이미 확정된 결제(SUCCEEDED, FAILED)는 건드리지 않습니다.
	if current.Status == StatusSucceeded || current.Status == StatusFailed {
		return current, nil
	}

	// PENDING은 선점한 요청이 아직 PG 응답을 기다리는 중일 수 있습니다. PENDING 행의 updated_at_ms는 claim이
	// INSERT한 시각 그대로이므로(finish가 바꾸면 더는 PENDING이 아닙니다), 지금까지 지난 시간(age)은 곧 선점한 뒤
	// 지난 시간입니다. age를 DB 시각으로 재고, stuckAfter보다 짧으면 건드리지 않습니다. 서버마다 시계가 조금씩
	// 다르므로, 애플리케이션 시각으로 재면 서버에 따라 판단이 달라집니다.
	if current.Status == StatusPending {
		age, err := pendingAge(ctx, db, key)
		if err != nil {
			return Payment{}, err
		}
		if age < stuckAfter {
			return current, nil
		}
	}

	// PG에 key로 승인된 거래가 있는지 묻습니다. 이 답(approved)이 확정할 상태(next)의 근거입니다.
	transactionID, approved, err := gateway.Inquire(ctx, key)
	if err != nil {
		return Payment{}, fmt.Errorf("inquire: %w", err)
	}
	next := StatusFailed
	if approved {
		next = StatusSucceeded
	}

	// payments.status가 처음 읽은 상태(current.Status) 그대로일 때만 next로 확정합니다. 그사이 선점한 요청이
	// 결과를 기록했거나 다른 Reconcile 호출이 먼저 확정했다면(ErrStateChanged) 그 결과를 덮지 않고 다시 읽어 돌려줍니다.
	payment, err := finish(ctx, db, key, current.Status, next, transactionID)
	if errors.Is(err, ErrStateChanged) {
		return loadPayment(ctx, db, key)
	}
	return payment, err
}

// claim 함수는 req.Key로 PENDING 결제를 INSERT해, 이 키로 PG를 부를 권리를 먼저 가져갑니다(선점).
//
// 선점하면 true를, 같은 키의 행이 이미 있어 UNIQUE 제약에 막히면 false를 돌려줍니다. 둘 다 오류는 nil입니다.
// 그 밖의 DB 오류는 감싸서 돌려줍니다.
//
// UNIQUE 충돌은 DB가 판정하고, 판정 방식은 DB마다 다릅니다.
//
//   - SQLite: 쓰기를 한 번에 하나만 실행하므로 뒤의 INSERT가 앞의 커밋 결과를 보고 중복을 발견합니다.
//   - PostgreSQL: 같은 키의 행을 먼저 넣은 트랜잭션이 아직 커밋 전이면, 뒤의 INSERT가 그 트랜잭션이 끝날 때까지
//     기다린 뒤 판정합니다. 앞 트랜잭션이 롤백하면 충돌이 아니므로 뒤의 INSERT가 성공합니다.
//   - MySQL InnoDB: 중복 키를 만나면 그 인덱스 레코드에 공유 락(여러 트랜잭션이 함께 쥘 수 있지만 배타 락과는
//     함께 쥘 수 없는 락)을 요청합니다. 매뉴얼의 세 세션 예시는 이렇게 데드락에 빠집니다.
//     1) S1이 키를 INSERT해 그 레코드에 배타 락을 쥡니다. 2) S2와 S3이 같은 키를 INSERT하다 중복을 만나
//     각자 공유 락을 요청하며 기다립니다. 3) S1이 롤백하면 S2와 S3이 둘 다 공유 락을 얻습니다. 4) 둘 다 실제로
//     INSERT하려고 배타 락을 요청하지만 상대의 공유 락 때문에 서로를 기다립니다. 5) InnoDB가 한쪽을 오류 1213으로
//     끝냅니다. MySQL로 옮길 때는 호출자가 1213을 재시도해야 하며, 이 코드는 SQLite용이라 그 재시도를 넣지 않았습니다.
func claim(ctx context.Context, db *sql.DB, req PaymentRequest) (bool, error) {
	// 자동 커밋 INSERT 한 문장으로 끝냅니다. 긴 트랜잭션 안에서 선점하면, 같은 req.Key로 온 다른 요청이
	// 즉시 거절되지 않고 그 트랜잭션이 끝날 때까지 기다리게 됩니다.
	_, err := db.ExecContext(ctx,
		"INSERT INTO payments (idempotency_key, account_id, amount, status, updated_at_ms) VALUES (?, ?, ?, ?, "+dbNowMillis+")",
		req.Key, req.AccountID, req.Amount, StatusPending)

	// UNIQUE 위반은 오류가 아니라 "같은 req.Key를 다른 요청이 먼저 선점했다"는 정상 결과입니다.
	if isUniqueViolation(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim payment: %w", err)
	}
	return true, nil
}

// finish 함수는 key의 결제가 아직 from 상태일 때만 to 상태로 바꾸고, 바뀐 결제를 돌려줍니다.
//
// from 상태가 아니어서 바뀐 행이 없으면 ErrStateChanged를, DB 오류가 나면 감싼 오류를 돌려줍니다.
func finish(ctx context.Context, db *sql.DB, key string, from, to PaymentStatus, transactionID string) (Payment, error) {
	// ------------------------------------------------------------------------
	// 핵심: WHERE status = ? 는 payments.status가 아직 from일 때만 to로 바꿉니다. 이 조건이 없으면, Reconcile이
	// FAILED로 확정한 뒤 선점한 요청이 뒤늦게 부른 finish가 그 FAILED를 오류 없이 SUCCEEDED로 덮어씁니다.
	// ------------------------------------------------------------------------
	res, err := db.ExecContext(ctx,
		"UPDATE payments SET status = ?, pg_transaction_id = NULLIF(?, ''), updated_at_ms = "+dbNowMillis+
			" WHERE idempotency_key = ? AND status = ?",
		to, transactionID, key, from)
	if err != nil {
		return Payment{}, fmt.Errorf("record %s: %w", to, err)
	}

	// 바뀐 행(rows)이 0이면 payments.status가 이미 from이 아닙니다. 다른 호출이 먼저 바꿨습니다.
	rows, err := res.RowsAffected()
	if err != nil {
		return Payment{}, fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return Payment{}, ErrStateChanged
	}
	return loadPayment(ctx, db, key)
}

// loadPayment 함수는 key의 결제 기록을 읽습니다.
//
// 기록이 없으면 sql.ErrNoRows를 감싼 오류를, 그 밖의 DB 오류도 감싸서 돌려줍니다.
func loadPayment(ctx context.Context, db *sql.DB, key string) (Payment, error) {
	var p Payment
	var transactionID sql.NullString
	var updatedAtMillis int64
	err := db.QueryRowContext(ctx,
		"SELECT idempotency_key, account_id, amount, status, pg_transaction_id, updated_at_ms FROM payments WHERE idempotency_key = ?",
		key).Scan(&p.Key, &p.AccountID, &p.Amount, &p.Status, &transactionID, &updatedAtMillis)
	if err != nil {
		return Payment{}, fmt.Errorf("load payment %q: %w", key, err)
	}

	// PG 거래 번호(pg_transaction_id)는 승인 전에는 NULL이므로 sql.NullString으로 받고,
	// Payment.PGTransactionID에는 NULL 대신 빈 문자열을 넣습니다.
	p.PGTransactionID = transactionID.String
	p.UpdatedAt = time.UnixMilli(updatedAtMillis)
	return p, nil
}

// pendingAge 함수는 key의 결제가 마지막으로 바뀐 뒤 지난 시간을 DB 시각 기준으로 돌려줍니다.
//
// 기록이 없으면 sql.ErrNoRows를 감싼 오류를 돌려줍니다.
func pendingAge(ctx context.Context, db *sql.DB, key string) (time.Duration, error) {
	var ageMillis int64
	err := db.QueryRowContext(ctx,
		"SELECT "+dbNowMillis+" - updated_at_ms FROM payments WHERE idempotency_key = ?", key).Scan(&ageMillis)
	if err != nil {
		return 0, fmt.Errorf("payment age %q: %w", key, err)
	}
	return time.Duration(ageMillis) * time.Millisecond, nil
}
