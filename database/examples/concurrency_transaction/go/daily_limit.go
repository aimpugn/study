package concurrency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrLimitExceeded 오류는 이번 결제를 더하면 고객의 한도를 넘는다는 뜻입니다.
var ErrLimitExceeded = errors.New("daily limit exceeded")

// ErrCustomerNotFound 오류는 한도가 등록된 고객이 없다는 뜻입니다.
var ErrCustomerNotFound = errors.New("customer not found")

// ChargeReadThenInsert 함수는 지금까지 결제한 합계를 읽고, 한도 안이면 결제 행을 추가합니다.
// 동시에 부르면 한도를 넘기는 안티패턴입니다.
//
// 이 실패는 갱신 손실과 모양이 다릅니다. 두 요청은 같은 행을 덮어쓰지 않고 charges에 각자 새 행을
// 추가합니다. 깨지는 것은 "고객의 charges.amount 합계 <= daily_limits.limit_amount"처럼 여러 행에 걸친
// 규칙입니다. 각 요청은 자기가 읽은 합계 기준으로는 옳게 판단했지만, 두 결정이 합쳐지면 규칙이 깨집니다.
// 이런 현상을 쓰기 스큐(write skew)라고 합니다.
//
// 성공하면 nil을, 고객이 없으면 ErrCustomerNotFound를, 한도를 넘으면 ErrLimitExceeded를 돌려줍니다.
// 그 밖의 DB 오류는 감싸서 돌려줍니다.
func ChargeReadThenInsert(ctx context.Context, db *sql.DB, customerID string, amount int64, pause Pause) error {
	// 한도(daily_limits.limit_amount)를 변수 limit에, 지금까지의 결제 합계(charges.amount의 합)를 변수 used에
	// 복사합니다. 두 변수 모두 읽은 순간의 값에 고정됩니다.
	var limit, used int64
	err := db.QueryRowContext(ctx,
		`SELECT l.limit_amount, COALESCE((SELECT SUM(amount) FROM charges WHERE customer_id = l.customer_id), 0)
		   FROM daily_limits l WHERE l.customer_id = ?`, customerID).Scan(&limit, &used)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCustomerNotFound
	}
	if err != nil {
		return fmt.Errorf("read limit: %w", err)
	}
	// 합계 복사본(used)에 이번 결제액(amount)을 더하면 한도 복사본(limit)을 넘는지 판단합니다.
	if used+amount > limit {
		return ErrLimitExceeded
	}

	// ------------------------------------------------------------------------
	// 주의: 여기가 틈입니다. 다른 요청도 같은 합계(used)를 읽고 한도 안이라고 판단해 이 줄에 올 수 있습니다.
	// PostgreSQL이나 MySQL에서 합계를 읽기 전에 이 고객의 기존 charges 행을 SELECT ... FOR UPDATE로 잠근다고
	// 해도 막히지 않습니다. 첫 결제라면 잠글 행이 0개라서 두 요청 모두 기다리지 않고, 이어서 각자 서로 다른
	// 새 행을 INSERT하므로 같은 행을 두고 기다리는 순간이 한 번도 오지 않습니다.
	// ------------------------------------------------------------------------
	pause.here()

	_, err = db.ExecContext(ctx, "INSERT INTO charges (customer_id, amount) VALUES (?, ?)", customerID, amount)
	if err != nil {
		return fmt.Errorf("insert charge: %w", err)
	}
	return nil
}

// ChargeReserved 함수는 고객 한도 행에서 사용액을 조건부로 늘린 뒤, 같은 트랜잭션에서 결제 행을 추가합니다.
//
// 여러 행에 흩어진 규칙을 한 행의 규칙으로 모으는 방법입니다. daily_limits.used_amount가 charges 합계를
// 대신하므로, 같은 고객의 결제는 모두 daily_limits의 그 한 행을 UPDATE합니다. 한도 10,000원에 6,000원 두 건이
// 겹치면 PostgreSQL과 MySQL에서 이렇게 진행됩니다.
//
//  1. A의 UPDATE가 WHERE 0+6000 <= 10000을 참으로 보고 used_amount를 6,000으로 바꾸며 그 행에 락을 겁니다.
//  2. B의 UPDATE는 같은 행을 고치려다 A의 락 때문에 기다립니다.
//  3. A가 charges 행을 INSERT하고 커밋합니다. 락이 풀립니다.
//  4. B의 UPDATE가 최신 used_amount 6,000으로 WHERE 6000+6000 <= 10000을 다시 계산해 거짓이 되고 0행을 바꿉니다.
//
// SQLite는 DB 전체에 쓰기 락이 하나뿐이라 이 트랜잭션이 한 번에 하나씩만 실행됩니다. 그래서 이 예제의 테스트가
// 통과한 이유가 "규칙을 daily_limits 한 행으로 모았기 때문"인지, "SQLite가 원래 모든 쓰기를 줄 세웠기 때문"인지
// 가릴 수 없습니다. 행마다 따로 락을 거는 PostgreSQL과 MySQL에서도 규칙이 지켜진다는 근거는 위 순서(공식 문서의
// 행 락 동작)이며, 실행해 확인한 것은 아닙니다.
//
// 성공하면 nil을, 고객이 없으면 ErrCustomerNotFound를, 한도를 넘으면 ErrLimitExceeded를 돌려줍니다.
// 결제 행 추가가 실패하면 사용액 증가도 함께 롤백하고 그 오류를 돌려줍니다.
func ChargeReserved(ctx context.Context, db *sql.DB, customerID string, amount int64) error {
	// BEGIN IMMEDIATE가 아닌 기본 BeginTx(SQLite에서는 DEFERRED)로, PostgreSQL이나 MySQL 코드와 같은 모양입니다.
	// 첫 문장이 SELECT가 아니라 UPDATE라서 이 트랜잭션은 스냅숏 없이 곧바로 쓰기 락을 요청합니다.
	// 읽기 트랜잭션에서 전환하는 경우가 아니므로 SQLite가 busy timeout까지 기다려 줍니다.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	// ------------------------------------------------------------------------
	// 핵심: 한도 확인(used_amount + ? <= limit_amount)과 사용액(used_amount) 증가를 UPDATE 한 문장으로 합니다.
	// 같은 고객(customer_id)의 모든 결제가 daily_limits의 이 한 행에서 차례로 판정받습니다.
	// ------------------------------------------------------------------------
	res, err := tx.ExecContext(ctx,
		"UPDATE daily_limits SET used_amount = used_amount + ? WHERE customer_id = ? AND used_amount + ? <= limit_amount",
		amount, customerID, amount)
	if err != nil {
		return fmt.Errorf("reserve limit: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	// 0행이면 daily_limits에 고객(customer_id)이 없거나, used_amount에 결제액(amount)을 더하면 limit_amount를
	// 넘는 경우입니다. 같은 트랜잭션(tx)에서 한 번 더 SELECT해 둘을 구분합니다.
	if rows == 0 {
		var exists int
		err := tx.QueryRowContext(ctx, "SELECT 1 FROM daily_limits WHERE customer_id = ?", customerID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCustomerNotFound
		}
		if err != nil {
			return fmt.Errorf("check customer: %w", err)
		}
		return ErrLimitExceeded
	}

	// 사용액(used_amount)을 확보한 뒤 결제 내역(charges 행)을 남깁니다. 이 INSERT가 실패하면 defer의
	// tx.Rollback이 위 UPDATE까지 되돌려, used_amount만 늘고 charges 행은 없는 상태가 남지 않습니다.
	if _, err := tx.ExecContext(ctx, "INSERT INTO charges (customer_id, amount) VALUES (?, ?)", customerID, amount); err != nil {
		return fmt.Errorf("insert charge: %w", err)
	}
	return tx.Commit()
}
