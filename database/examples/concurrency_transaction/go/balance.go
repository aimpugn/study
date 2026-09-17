package concurrency

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync"
)

// ErrAccountNotFound 오류는 출금하려는 계좌가 없다는 뜻입니다.
var ErrAccountNotFound = errors.New("account not found")

// ErrInsufficientBalance 오류는 잔액이 출금액보다 적다는 뜻입니다.
var ErrInsufficientBalance = errors.New("insufficient balance")

// ErrTooManyConflicts 오류는 WithdrawOptimistic이 maxAttempts번 모두, 읽은 뒤 다른 요청이 같은 행을 먼저 바꾸는 바람에
// 쓰지 못했다는 뜻입니다.
var ErrTooManyConflicts = errors.New("too many concurrent modifications")

// ============================================================================
// 깨지는 구현
// ============================================================================

// WithdrawReadModifyWrite 함수는 잔액을 읽고, 애플리케이션 메모리에서 빼고, 계산한 값을 덮어씁니다.
// 동시에 부르면 틀린 결과를 내는 안티패턴입니다.
//
// 읽기(SELECT)와 쓰기(UPDATE)가 서로 다른 SQL 문이라 둘 사이에 틈이 있습니다. 잔액 20,000원에서 5,000원 출금
// 두 건이 겹치면, 두 요청이 모두 20,000을 읽고 각자 15,000을 계산합니다. A가 15,000을 쓴 뒤 B도 자기 계산값
// 15,000을 씁니다. B의 UPDATE는 A가 뺀 5,000을 모르므로, 출금 두 건이 성공했는데 잔액은 한 번만 줄어듭니다.
// 이렇게 나중에 쓴 쪽이 먼저 쓴 변경을 지우는 현상을 갱신 손실(lost update)이라고 합니다.
// pause는 그 틈에서 멈추는 자리이며 운영 코드에서는 nil입니다.
//
// 성공하면 nil을, 계좌가 없으면 ErrAccountNotFound를, 읽은 잔액이 모자라면
// ErrInsufficientBalance를 돌려줍니다. 그 밖의 DB 오류는 감싸서 돌려줍니다.
func WithdrawReadModifyWrite(ctx context.Context, db *sql.DB, accountID string, amount int64, pause Pause) error {
	// DB의 잔액(accounts.balance)을 변수 balance에 복사합니다. 변수 balance는 읽은 순간의 값에 고정되고,
	// 이후 다른 요청이 accounts.balance를 바꿔도 따라 바뀌지 않습니다.
	var balance int64
	err := db.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = ?", accountID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAccountNotFound
	}
	if err != nil {
		return fmt.Errorf("read balance: %w", err)
	}

	// 잔액 복사본(balance)이 출금액(amount)보다 적은지 판단합니다. 그사이 다른 요청이 accounts.balance를
	// 줄였다면, 이 판단은 이미 낡은 값에 기댄 것입니다.
	if balance < amount {
		return ErrInsufficientBalance
	}

	// ------------------------------------------------------------------------
	// 주의: 위 SELECT와 아래 UPDATE 사이가 틈입니다. 여기서 다른 요청이 accounts.balance를 바꾸면 변수 balance는
	// 낡은 값이 되는데, 다른 요청이 이 요청과 같은 잔액 값을 읽고 이 줄까지 오는 것을 막는 장치가 없습니다.
	// ------------------------------------------------------------------------
	pause.here()

	// 잔액 복사본으로 계산한 값(balance-amount)을 accounts.balance에 그대로 덮어씁니다. 틈 사이에 다른 요청이
	// accounts.balance를 20,000에서 15,000으로 줄였어도, 이 UPDATE는 변수 balance(20,000)로 계산한 15,000을
	// 그대로 쓰므로 다른 요청의 차감이 반영되지 않습니다.
	_, err = db.ExecContext(
		ctx,
		"UPDATE accounts SET balance = ?, version = version + 1 WHERE id = ?",
		balance-amount,
		accountID,
	)

	if err != nil {
		return fmt.Errorf("write balance: %w", err)
	}
	return nil
}

// Server 타입은 요청을 처리하는 애플리케이션 인스턴스 한 대를 흉내 냅니다.
//
// 인스턴스마다 자기 뮤텍스를 갖고, 인스턴스끼리 공유하는 것은 DB뿐입니다. 한 프로세스에
// Server 값을 두 개 만들면 서버 두 대를 띄운 것과 공유 관계가 같아집니다. 다만 실제
// 프로세스 두 개를 띄워 확인한 것은 아닙니다.
type Server struct {
	mu sync.Mutex
}

// Withdraw 메서드는 이 인스턴스의 뮤텍스를 잡은 채 WithdrawReadModifyWrite를 실행합니다.
// 서버가 한 대일 때만 안전하고 여러 대가 되면 안전하지 않은 안티패턴입니다.
//
// 인자, 반환값, 오류는 WithdrawReadModifyWrite와 같습니다.
func (s *Server) Withdraw(ctx context.Context, db *sql.DB, accountID string, amount int64, pause Pause) error {
	// ------------------------------------------------------------------------
	// 주의: 이 뮤텍스는 같은 Server 값으로 들어온 요청끼리만 줄을 세웁니다.
	// 다른 Server 값의 요청은 자기 뮤텍스(s.mu)를 잡고 같은 accounts 행에 그대로 들어옵니다.
	// ------------------------------------------------------------------------
	s.mu.Lock()
	defer s.mu.Unlock()
	return WithdrawReadModifyWrite(ctx, db, accountID, amount, pause)
}

// WithdrawInPlainTx 함수는 WithdrawReadModifyWrite 함수와 같은 일을 트랜잭션 안에서 합니다.
// 트랜잭션만 열면 읽은 행이 잠긴다고 착각하기 쉬운 안티패턴입니다.
//
// 트랜잭션은 여러 변경을 한 묶음으로 커밋하거나 되돌릴 뿐, 일반 SELECT로 읽은 행을 다른
// 요청이 못 바꾸게 잠그지는 않습니다. 같은 코드라도 DB에 따라 증상이 다릅니다.
//
//   - SQLite: 트랜잭션은 읽기 트랜잭션(첫 SELECT 시점의 DB 상태, 즉 스냅숏을 기준으로 읽기만 함, 여러 연결이 동시에 가능)과
//     쓰기 트랜잭션(DB 전체에 하나뿐인 쓰기 락을 쥠)으로 나뉩니다. 기본 BEGIN(DEFERRED)은 첫 문장을
//     보고 종류를 정하므로, 첫 SELECT에서 읽기 트랜잭션이 됩니다. 같은 트랜잭션에서 UPDATE를 실행하려면
//     그 순간 쓰기 락을 얻어 쓰기 트랜잭션으로 전환해야 합니다(공식 문서의 upgrade). 이 전환은 다른
//     연결이 쓰기 락을 쥐고 있으면 `SQLITE_BUSY`로, 다른 연결이 이미 써서 커밋했으면 스냅숏이 낡았으므로
//     `SQLITE_BUSY_SNAPSHOT`으로 실패합니다. busy timeout을 설정해도 기다리지 않습니다(실행으로 확인함).
//     기다린 끝에 상대가 커밋하면 이 트랜잭션의 스냅숏이 낡아 어차피 전환할 수 없다는 점이 이유로 추론되지만,
//     상대가 롤백하는 경우까지 설명하지는 못합니다. 데이터는 지켜지지만 요청 하나가 오류로 끝납니다.
//   - PostgreSQL READ COMMITTED: PostgreSQL은 UPDATE할 때 행을 덮어쓰지 않고 고친 내용으로 새 행 버전을
//     만듭니다. 먼저 UPDATE한 트랜잭션이 그 행에 락을 걸므로, 뒤의 UPDATE는 앞 트랜잭션이 커밋할 때까지 기다리고, 커밋되면
//     앞 트랜잭션이 만든 최신 행 버전을 보고 WHERE id = ? 가 아직 맞는지만 다시 확인한 뒤 SET을 실행합니다.
//     이때 SET balance = ? 에 넘긴 값은 애플리케이션이 앞선 SELECT의 잔액 복사본으로 미리 계산한
//     숫자(balance-amount)라서, PostgreSQL은 그 숫자가 낡았는지 알 수 없고 그대로 씁니다. 그래서 오류 없이
//     갱신 손실이 납니다. 공식 문서의 규칙에서 따라 나온 결론이며, 이 예제를 PostgreSQL로 실행해 확인하지는 않았습니다.
//
// 다른 연결이 쓰기 락을 쥐고 있었거나 먼저 커밋해 UPDATE가 거부되면 IsBusy가 true인 오류를 돌려줍니다. 나머지 반환값은 WithdrawReadModifyWrite와 같습니다.
func WithdrawInPlainTx(ctx context.Context, db *sql.DB, accountID string, amount int64, pause Pause) error {
	// ------------------------------------------------------------------------
	// 주의: BeginTx는 SQLite에 기본 BEGIN(DEFERRED)을 보냅니다. 트랜잭션은 열렸지만
	// 이 줄에서는 어떤 락도 잡지 않습니다.
	// ------------------------------------------------------------------------
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	// 첫 SELECT에서 tx의 읽기 트랜잭션이 시작됩니다. tx는 이 시점의 DB 상태, 즉 스냅숏을 기준으로 읽습니다.
	var balance int64
	err = tx.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = ?", accountID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAccountNotFound
	}
	if err != nil {
		return fmt.Errorf("read balance: %w", err)
	}
	if balance < amount {
		return ErrInsufficientBalance
	}

	// ------------------------------------------------------------------------
	// 주의: 트랜잭션 안인데도 여기가 틈입니다. 다른 연결도 같은 accounts.balance를 읽고 이 줄에 올 수 있습니다.
	// ------------------------------------------------------------------------
	pause.here()

	// tx가 쓰기 락을 얻어 읽기 트랜잭션에서 쓰기 트랜잭션으로 전환해야 하는 순간입니다. 다른 연결이 먼저 DB에 쓰기 시작했거나
	// 이미 커밋했다면 SQLite는 여기서 쓰기를 거부합니다. SQLite에서는 틈이 이렇게 오류로 드러납니다.
	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = ?, version = version + 1 WHERE id = ?",
		balance-amount, accountID)
	if err != nil {
		return fmt.Errorf("write balance: %w", err)
	}
	return tx.Commit()
}

// ============================================================================
// 지키는 구현
// ============================================================================

// WithdrawConditional 함수는 잔액 확인과 차감을 UPDATE 한 문장에 담아 원자적으로 실행합니다.
//
// "잔액이 충분하면 뺀다"는 규칙을 WHERE로 옮기면 확인과 쓰기 사이에 애플리케이션 코드가 끼지
// 않습니다. 그래서 이 함수는 Pause를 받지 않습니다. 멈출 틈이 없습니다. 두 요청이 겹칠 때
// 뒤의 요청이 앞의 결과를 보고 판단하게 만드는 주체는 DB입니다.
//
//   - SQLite: 쓰기는 DB 전체에서 한 번에 하나라서, 뒤의 UPDATE는 앞 UPDATE가 자동 커밋된 뒤에 실행됩니다.
//     그래서 WHERE balance >= ? 와 SET balance = balance - ? 모두 앞 요청이 줄인 잔액을 보고 계산합니다.
//   - PostgreSQL READ COMMITTED: 뒤의 UPDATE는 앞 트랜잭션의 커밋을 기다린 뒤, 앞 트랜잭션이 만든 최신 행 버전의
//     잔액으로 WHERE balance >= ? 를 다시 확인하고 SET balance = balance - ? 도 그 최신 잔액에서 계산합니다.
//     SQL 밖에서 미리 계산한 값이 없으므로 낡은 값이 쓰이지 않습니다.
//     격리 수준을 REPEATABLE READ 이상으로 올리면 결과가 다릅니다. 이 수준의 트랜잭션은 처음 읽은 시점의
//     스냅숏만 봐야 하므로, 그 이후 앞 트랜잭션이 바꾼 행을 최신 버전으로 다시 확인하는 대신
//     could not serialize access 오류(SQL 표준 오류 코드 SQLSTATE 40001)로 실패합니다. 호출자는 트랜잭션
//     전체를 처음부터 다시 실행해야 합니다.
//   - MySQL InnoDB: UPDATE는 대상 행에 배타 락(다른 트랜잭션이 그 행을 고치거나 잠그지 못하게 하는 락)을
//     겁니다. 뒤의 UPDATE는 이 락이 풀릴 때까지 기다린 뒤, 스냅숏이 아니라 최신 커밋 상태의 잔액으로
//     WHERE와 SET을 계산합니다.
//
// 성공하면 nil을, 계좌가 없으면 ErrAccountNotFound를, 잔액이 모자라면 ErrInsufficientBalance를 돌려줍니다.
// 그 밖의 DB 오류는 감싸서 돌려줍니다.
func WithdrawConditional(ctx context.Context, db *sql.DB, accountID string, amount int64) error {
	// ------------------------------------------------------------------------
	// 핵심: SQL 안의 balance는 변수가 아니라 컬럼 accounts.balance입니다. SET balance = balance - ? 는
	// DB의 현재 잔액에서 출금액(amount)을 빼고, WHERE balance >= ? 는 같은 문장 안에서 현재 잔액이
	// 출금액 이상인지 확인합니다. 확인과 쓰기 사이에 틈이 없습니다.
	// ------------------------------------------------------------------------
	res, err := db.ExecContext(ctx,
		"UPDATE accounts SET balance = balance - ?, version = version + 1 WHERE id = ? AND balance >= ?",
		amount, accountID, amount)
	if err != nil {
		return fmt.Errorf("conditional update: %w", err)
	}

	// 바뀐 행 수(rows)가 판정 결과입니다. 1이면 출금됐고, 0이면 WHERE 조건(계좌가 있고 잔액이 출금액 이상)에
	// 맞는 행이 없었습니다.
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 1 {
		return nil
	}

	// 0행인 이유가 계좌(accountID) 부재인지 잔액 부족인지 구분하려고 한 번 더 SELECT합니다. 이 SELECT는
	// 돌려줄 오류를 고르는 데만 쓰고 돈을 움직이는 판단에는 쓰지 않으므로, 틈이 있어도 무해합니다.
	var exists int
	err = db.QueryRowContext(ctx, "SELECT 1 FROM accounts WHERE id = ?", accountID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAccountNotFound
	}
	if err != nil {
		return fmt.Errorf("check account: %w", err)
	}
	return ErrInsufficientBalance
}

// WithdrawLocked 함수는 쓰기 락을 먼저 잡고, 그 안에서 읽고 판단하고 씁니다(비관적 락).
//
// 락을 커밋까지 쥐고 있으므로, 락 안에서 읽은 잔액은 트랜잭션이 끝날 때까지 DB의 실제 값과 같습니다.
// UPDATE 한 문장으로 표현하기 어려운 규칙, 예를 들어 여러 테이블을 읽고 판단한 뒤 쓰는 일에
// 쓰는 방식입니다. PostgreSQL과 MySQL에서는 SELECT ... FOR UPDATE가 같은 역할을 행 단위로 합니다.
//
// pause는 락을 쥔 채 멈추는 자리입니다. 테스트는 여기서 멈춘 동안 다른 요청이 약속 지점에
// 오지 못한다는 사실로 락이 실제로 상대를 막았는지 확인합니다.
//
// 성공하면 nil을, 계좌가 없으면 ErrAccountNotFound를, 잔액이 모자라면 ErrInsufficientBalance를
// 돌려줍니다. busy timeout 안에 락을 못 잡으면 IsBusy가 true인 오류를 돌려줍니다.
func WithdrawLocked(ctx context.Context, db *sql.DB, accountID string, amount int64, pause Pause) error {
	return inImmediateTx(ctx, db, func(conn *sql.Conn) error {
		// --------------------------------------------------------------------
		// 핵심: 이 SELECT는 inImmediateTx가 BEGIN IMMEDIATE로 쓰기 락을 잡은 뒤에 실행됩니다. 커밋 전까지
		// 다른 요청은 accounts.balance를 바꿀 수 없으므로, 변수 balance에 복사한 잔액은 낡지 않습니다.
		// --------------------------------------------------------------------
		var balance int64
		err := conn.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = ?", accountID).Scan(&balance)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAccountNotFound
		}
		if err != nil {
			return fmt.Errorf("read balance: %w", err)
		}
		if balance < amount {
			return ErrInsufficientBalance
		}

		// WithdrawReadModifyWrite의 틈과 같은 위치지만, 여기서 멈춘 동안 다른 요청은 inImmediateTx의
		// BEGIN IMMEDIATE에서 기다립니다.
		pause.here()

		_, err = conn.ExecContext(ctx,
			"UPDATE accounts SET balance = ?, version = version + 1 WHERE id = ?",
			balance-amount, accountID)
		if err != nil {
			return fmt.Errorf("write balance: %w", err)
		}
		return nil
	})
}

// WithdrawOptimistic 함수는 락 없이 읽고, 쓰는 순간에 "읽은 뒤 아무도 안 바꿨는가"를 확인합니다(낙관적 락).
//
// 충돌이 드물면 락을 기다리지 않아 빠르고, 충돌이 잦으면 재시도가 늘어 느려집니다. 재시도 횟수에는
// 상한이 필요합니다. 여기서 충돌은 version 비교 UPDATE가 0행을 바꾼 경우입니다. 0행이 나오려면 다른 요청의
// UPDATE가 먼저 성공해 version을 올렸어야 하므로, 충돌이 날 때마다 적어도 한 요청의 출금은 끝납니다.
// 하지만 같은 요청이 매번 그 다른 요청에게 질 수는 있으므로 시도 횟수를 maxAttempts로 제한합니다.
//
// 성공하면 실행한 시도 횟수와 nil을 돌려줍니다. maxAttempts번 모두 충돌하면 ErrTooManyConflicts를,
// 계좌가 없거나 잔액이 모자라면 해당 오류를 돌려줍니다. 그 밖의 DB 오류는 감싸서 돌려줍니다.
func WithdrawOptimistic(ctx context.Context, db *sql.DB, accountID string, amount int64, maxAttempts int, pause Pause) (int, error) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// accounts.balance와 함께, 그 행이 지금까지 몇 번 바뀌었는지 세는 accounts.version을
		// 변수 balance와 version으로 복사합니다.
		var balance, version int64
		err := db.QueryRowContext(ctx,
			"SELECT balance, version FROM accounts WHERE id = ?", accountID).Scan(&balance, &version)
		if errors.Is(err, sql.ErrNoRows) {
			return attempt, ErrAccountNotFound
		}
		if err != nil {
			return attempt, fmt.Errorf("read balance: %w", err)
		}
		if balance < amount {
			return attempt, ErrInsufficientBalance
		}

		// WithdrawReadModifyWrite와 같은 틈이 여기에도 있습니다. 낙관적 락은 틈을 없애지 않고,
		// 틈 사이에 생긴 변경을 아래 UPDATE에서 검출합니다.
		pause.here()

		// ------------------------------------------------------------------------
		// 핵심: WHERE version = ? 는 DB의 accounts.version이 복사해 온 변수 version과 같을 때만 씁니다.
		// 틈 사이에 다른 쓰기가 있었다면 accounts.version이 이미 올라가 있어 0행이 바뀝니다.
		// ------------------------------------------------------------------------
		res, err := db.ExecContext(ctx,
			"UPDATE accounts SET balance = ?, version = version + 1 WHERE id = ? AND version = ?",
			balance-amount, accountID, version)
		if err != nil {
			return attempt, fmt.Errorf("compare and set: %w", err)
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return attempt, fmt.Errorf("rows affected: %w", err)
		}
		if rows == 1 {
			return attempt, nil
		}
		// 0행이면 변수 balance와 version이 낡았습니다. balance-amount로 계산한 값을 버리고 처음부터 다시 복사합니다.
	}
	return maxAttempts, ErrTooManyConflicts
}

// inImmediateTx 함수는 SQLite 쓰기 락을 먼저 잡는 트랜잭션 안에서 fn을 실행하고 커밋합니다.
//
// SQLite 공식 문서에 따르면 BEGIN IMMEDIATE가 성공한 뒤로는 커밋까지 SQLITE_BUSY가 나지 않습니다.
// SQLite에는 행 락이 없어 DB 전체를 잠급니다.
//
// fn이 오류를 돌려주면 롤백하고 fn의 오류를 그대로 돌려주므로, 호출자는 errors.Is로 ErrInsufficientBalance 같은
// 업무 오류(규칙에 따라 거절된 경우)를 DB 오류와 구분할 수 있습니다. 커밋이 실패해도 롤백하고 오류를 돌려줍니다. busy timeout 안에 락을 못 잡으면
// BEGIN의 오류를 돌려줍니다.
func inImmediateTx(ctx context.Context, db *sql.DB, fn func(conn *sql.Conn) error) error {
	// 트랜잭션의 모든 SQL이 같은 연결로 가야 하므로 풀에서 연결 하나(conn)를 꺼내 고정합니다.
	// database/sql의 BeginTx는 SQLite에 DEFERRED로 시작하므로, 이 연결에 BEGIN IMMEDIATE를 직접 보냅니다.
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("get connection: %w", err)
	}
	defer conn.Close()

	// ------------------------------------------------------------------------
	// 핵심: fn의 첫 SQL보다 먼저 DB 전체의 쓰기 락을 잡습니다. 다른 연결이 락을 쥐고 있으면 busy timeout까지
	// 여기서 기다립니다. 이 연결은 아직 아무것도 읽지 않아 스냅숏이 없으므로, 앞 연결이 커밋할 때까지 기다렸다가
	// 쓰기 락을 얻어도 낡은 스냅숏 문제가 없어 SQLite가 기다려 줍니다. 먼저 SELECT한 뒤 UPDATE하는
	// WithdrawInPlainTx가 곧바로 SQLITE_BUSY를 받는 것과 이 점이 다릅니다.
	// ------------------------------------------------------------------------
	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return fmt.Errorf("begin immediate: %w", err)
	}

	if err := fn(conn); err != nil {
		if rollbackErr := rollback(ctx, conn); rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	// 커밋하는 순간 락이 풀리고, BEGIN IMMEDIATE에서 기다리던 다음 연결이 락을 잡습니다.
	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		commitErr := fmt.Errorf("commit: %w", err)
		if rollbackErr := rollback(ctx, conn); rollbackErr != nil {
			return errors.Join(commitErr, rollbackErr)
		}
		return commitErr
	}
	return nil
}

// rollback 함수는 conn에 열린 트랜잭션을 되돌립니다.
//
// 성공하면 nil을, 롤백마저 실패하면 롤백 오류를 돌려줍니다. 롤백이 실패한 연결은 불량으로 표시해 database/sql의
// 연결 풀이 버리게 합니다. 그대로 풀에 돌아가면, 다음에 이 연결을 받은 호출자의 SQL이 롤백되지 않은 이전
// 트랜잭션 안에서 실행됩니다.
func rollback(ctx context.Context, conn *sql.Conn) error {
	// ctx가 이미 취소됐어도 롤백은 보내야 하므로 취소 신호를 떼어 낸 문맥으로 실행합니다.
	_, err := conn.ExecContext(context.WithoutCancel(ctx), "ROLLBACK")
	if err == nil {
		return nil
	}

	// driver.ErrBadConn을 돌려주면 database/sql이 이 연결을 닫고 풀에서 뺍니다.
	_ = conn.Raw(func(any) error { return driver.ErrBadConn })
	return fmt.Errorf("rollback: %w", err)
}
