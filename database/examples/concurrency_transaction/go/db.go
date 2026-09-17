// Package concurrency 안의 코드는 동시에 들어온 요청이 결제 기록과 잔액을 망가뜨리는 방식과,
// 그 일을 DB가 막게 만드는 방법을 SQLite 위에서 재현합니다.
//
// 읽는 순서와 실행 방법은 같은 디렉터리의 README.md에 있습니다.
package concurrency

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

// schema 상수는 예제가 쓰는 테이블을 만듭니다.
//
// accounts.version은 이 행이 몇 번 바뀌었는지 세는 값입니다. 잔액을 읽는 요청은 version도 함께 읽어 두고,
// 쓸 때 DB의 version이 읽어 둔 값과 같은지 비교합니다. 같으면 그사이 아무도 이 행을 바꾸지 않은 것입니다.
// 이 방식을 낙관적 락이라고 하며 balance.go의 WithdrawOptimistic이 씁니다.
//
// 이 비교가 맞으려면 잔액을 바꾸는 모든 SQL이 version을 올려야 합니다. 어떤 SQL이 version을 그대로 둔 채
// 잔액만 바꾸면, 그 전에 version=0을 읽어 둔 요청은 쓰는 순간에도 version=0을 보고 바뀐 잔액을 모른 채
// 덮어씁니다. 그래서 이 패키지의 모든 잔액 쓰기는 version = version + 1을 함께 실행합니다.
const schema = `
CREATE TABLE IF NOT EXISTS accounts (
    id      TEXT PRIMARY KEY,
    balance INTEGER NOT NULL CHECK (balance >= 0),
    version INTEGER NOT NULL DEFAULT 0
);

-- 멱등성 키(idempotency_key)는 클라이언트가 결제 요청마다 만들고, 재시도할 때도 그대로 보내는 식별자입니다.
-- 같은 키로 몇 번이 와도 결제가 한 건만 남도록 UNIQUE를 겁니다.
-- 키를 PRIMARY KEY로 두면 중복일 때 SQLite가 SQLITE_CONSTRAINT_PRIMARYKEY(1555)를 돌려주는데,
-- 중복을 판별하는 isUniqueViolation은 SQLITE_CONSTRAINT_UNIQUE(2067)만 검사하므로 UNIQUE 컬럼으로 따로 둡니다.
CREATE TABLE IF NOT EXISTS payments (
    id                INTEGER PRIMARY KEY,
    idempotency_key   TEXT    NOT NULL UNIQUE,
    account_id        TEXT    NOT NULL,
    amount            INTEGER NOT NULL,
    status            TEXT    NOT NULL,
    pg_transaction_id TEXT,
    updated_at_ms     INTEGER NOT NULL
);

-- 깨지는 구현인 PayCheckThenAct가 쓰는 테이블입니다. payments와 같지만 UNIQUE가 없어서,
-- 같은 키의 행이 두 개 생기는 모습을 확인할 수 있습니다.
CREATE TABLE IF NOT EXISTS payments_unguarded (
    id              INTEGER PRIMARY KEY,
    idempotency_key TEXT    NOT NULL,
    account_id      TEXT    NOT NULL,
    amount          INTEGER NOT NULL
);

-- 고객별 결제 내역입니다. 지켜야 할 규칙은 "한 고객의 charges.amount 합계가 daily_limits.limit_amount 이하"이며,
-- 이 합계는 charges의 여러 행을 더해야 나옵니다.
CREATE TABLE IF NOT EXISTS charges (
    id          INTEGER PRIMARY KEY,
    customer_id TEXT    NOT NULL,
    amount      INTEGER NOT NULL
);

-- 고객별 한도와 지금까지 쓴 금액을 한 행에 모읍니다.
CREATE TABLE IF NOT EXISTS daily_limits (
    customer_id  TEXT    PRIMARY KEY,
    limit_amount INTEGER NOT NULL,
    used_amount  INTEGER NOT NULL DEFAULT 0 CHECK (used_amount <= limit_amount)
);
`

// OpenDB 함수는 path에 SQLite 파일을 만들고 예제 테이블을 준비합니다.
//
// 반환한 *sql.DB는 호출자가 닫아야 합니다. 파일을 열거나 스키마를 만들다 실패하면
// 오류를 돌려주고, 열어 둔 연결은 닫습니다.
func OpenDB(path string) (*sql.DB, error) {
	// WAL(write-ahead log) 모드로 엽니다. 이 모드의 SQLite는 여러 연결이 동시에 읽게 두지만,
	// 쓰기는 DB 전체에서 한 번에 하나만 허용합니다.
	//
	// busy timeout은 쓰기 락을 못 잡은 연결이 곧바로 실패하지 않고 5초까지 기다리게 합니다.
	//
	// 예외가 하나 있습니다. SQLite 트랜잭션은 읽기 트랜잭션(첫 SELECT 시점의 DB 상태, 즉 스냅숏을 기준으로
	// 읽기만 함)과 쓰기 트랜잭션(DB에 하나뿐인 쓰기 락을 쥠)으로 나뉩니다. SELECT로 시작한 트랜잭션이
	// UPDATE를 실행하려면 쓰기 락을 얻어 쓰기 트랜잭션으로 전환해야 하는데, 이때 막히면 SQLite는 기다리지 않고
	// 곧바로 SQLITE_BUSY를 돌려줍니다(실행으로 확인함). 기다린 끝에 상대가 커밋하면 이 트랜잭션의 스냅숏에 없는
	// 값이 DB에 생겨 어차피 전환할 수 없다는 점이 이유로 추론되지만, 상대가 롤백하는 경우까지 설명하지는
	// 못합니다. 자세한 순서는 balance.go의 WithdrawInPlainTx에 있습니다.
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// 테이블이 없을 때만 만듭니다. 같은 파일을 다시 열어도 기존 데이터는 그대로입니다.
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return db, nil
}

// IsBusy 함수는 err가 SQLite의 쓰기 락 경합 오류인지 알려 줍니다.
//
// SQLite 오류에는 기본 코드와, 원인을 더 자세히 나눈 확장 코드가 있습니다. SQLITE_BUSY_SNAPSHOT(517)의
// 기본 코드는 SQLITE_BUSY(5)이므로, 기본 코드(sqliteErr.Code)만 비교해도 두 오류를 함께 참으로 봅니다.
// 두 오류 모두 다른 연결이 쓰기 락을 쥐고 있었거나, 다른 연결의 커밋 때문에 이 트랜잭션의 스냅숏이 낡아
// 쓰지 못했다는 뜻이며, 트랜잭션을 처음부터 다시 실행하는 것이 대응 방법입니다.
func IsBusy(err error) bool {
	var sqliteErr sqlite3.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrBusy
}

// isUniqueViolation 함수는 err가 UNIQUE 제약 위반(SQLITE_CONSTRAINT_UNIQUE, 2067)인지 알려 줍니다.
//
// 오류 문장 대신 확장 코드로 판별합니다. 문장은 SQLite 버전에 따라 바뀔 수 있지만 코드는 바뀌지 않습니다.
func isUniqueViolation(err error) bool {
	var sqliteErr sqlite3.Error
	return errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
}

// dbNowMillis 상수는 DB의 현재 시각을 밀리초 단위 유닉스 시간으로 계산하는 SQL 식입니다.
//
// 결제 상태가 바뀐 시각을 기록할 때와, 결제가 처리 중(PENDING) 상태로 얼마나 오래 남았는지 잴 때 이 식을 씁니다.
// 여러 서버가 "이 결제가 멈췄는가"를 판단할 때 각자 자기 시계를 쓰면 판단이 서로 달라지므로, 하나뿐인
// DB 시계를 기준으로 삼습니다.
//
// SQLite의 CURRENT_TIMESTAMP는 초 단위라 쓰지 않습니다. julianday('now')는 현재 시각을 율리우스일(날짜 단위의
// 연속된 수)로 돌려줍니다. 여기서 1970-01-01 00:00 UTC의 율리우스일인 2440587.5를 빼면 유닉스 시간(일 단위)이
// 되고, 하루의 밀리초 수 86400000을 곱하면 밀리초가 됩니다.
const dbNowMillis = "CAST((julianday('now') - 2440587.5) * 86400000 AS INTEGER)"
