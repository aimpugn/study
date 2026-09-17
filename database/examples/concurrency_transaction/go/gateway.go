package concurrency

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrDeclined 오류는 외부 결제 대행사(PG, Payment Gateway)가 요청을 받아 거절했다는 확정 응답입니다.
// 돈은 움직이지 않았습니다.
var ErrDeclined = errors.New("payment declined by gateway")

// ErrTimeout 오류는 PG의 응답을 받지 못했다는 뜻입니다. PG에서 승인이 됐는지는 알 수 없습니다.
var ErrTimeout = errors.New("gateway response timed out")

// Gateway 인터페이스는 PG와 주고받는 요청을 정의합니다.
type Gateway interface {
	// Approve는 멱등성 키(key)로 식별되는 결제의 승인을 요청하고, 승인되면 PG가 발급한 거래 번호를 돌려줍니다.
	// 거절이면 ErrDeclined를, 응답을 못 받았으면 ErrTimeout을 돌려줍니다.
	Approve(ctx context.Context, key string, amount int64) (string, error)

	// Inquire는 이 멱등성 키(key)로 승인된 거래가 있는지 조회합니다. 있으면 거래 번호와 true를 돌려줍니다.
	// 조회 자체가 실패하면 오류를 돌려줍니다.
	Inquire(ctx context.Context, key string) (string, bool, error)
}

// Outcome 타입은 FakeGateway가 Approve 호출에 어떻게 반응할지 정합니다.
type Outcome int

const (
	// OutcomeApprove 값을 받은 가짜 PG는 승인하고 정상 응답합니다. 아무것도 정하지 않은 키의 기본값입니다.
	OutcomeApprove Outcome = iota
	// OutcomeDecline 값을 받은 가짜 PG는 승인하지 않고 ErrDeclined로 응답합니다.
	OutcomeDecline
	// OutcomeTimeoutAfterApproval 값을 받은 가짜 PG는 승인해 놓고, 응답 대신 ErrTimeout을 돌려줍니다.
	OutcomeTimeoutAfterApproval
	// OutcomeTimeoutBeforeApproval 값을 받은 가짜 PG는 승인하지 않은 채 ErrTimeout을 돌려줍니다.
	OutcomeTimeoutBeforeApproval
)

// FakeGateway 타입은 테스트용 PG입니다.
//
// 실제 PG 중에는 같은 키의 중복 승인을 걸러 주는 곳도 있지만, 이 가짜 PG는 걸러 주지
// 않습니다. 애플리케이션이 PG의 중복 방어에 기대지 않고도 한 번만 부르는지 확인하기 위해서입니다.
type FakeGateway struct {
	delay time.Duration

	mu       sync.Mutex
	outcomes map[string]Outcome
	approved map[string]string
	calls    int
}

// NewFakeGateway 함수는 Approve마다 delay만큼 지연되는 가짜 PG를 만듭니다.
func NewFakeGateway(delay time.Duration) *FakeGateway {
	return &FakeGateway{
		delay:    delay,
		outcomes: map[string]Outcome{},
		approved: map[string]string{},
	}
}

// SetOutcome 메서드는 key로 들어오는 Approve에 대한 반응을 정합니다.
func (g *FakeGateway) SetOutcome(key string, outcome Outcome) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.outcomes[key] = outcome
}

// Approve 메서드는 delay만큼 기다린 뒤 SetOutcome으로 정한 대로 승인하거나 거절하거나 응답을 잃어버립니다.
//
// 기다리는 동안 ctx가 끝나면 PG에 도달하지 못한 것으로 보고 ctx.Err()를 돌려주며, 이때는 Calls를 늘리지
// 않습니다. PG에 도달한 호출만 Calls가 하나 늘고, 승인하면 approved 맵에 key별 새 거래 번호를 기록합니다.
// 같은 key로 두 번 승인하면 approved의 거래 번호가 두 번째 것으로 바뀌어 Inquire에는 마지막 번호만 보이고,
// Calls는 2가 됩니다. 실제 PG라면 돈이 두 번 나간 이중 승인입니다.
//
// 승인이면 거래 번호와 nil을, 거절이면 ErrDeclined를, 응답을 잃어버리면 ErrTimeout을 돌려줍니다.
func (g *FakeGateway) Approve(ctx context.Context, key string, amount int64) (string, error) {
	// delay만큼 기다려 네트워크 왕복 시간을 흉내 냅니다. 이 시간 동안 호출자는 승인 결과를 모릅니다.
	select {
	case <-time.After(g.delay):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls++
	outcome := g.outcomes[key]

	if outcome == OutcomeDecline {
		return "", ErrDeclined
	}
	if outcome == OutcomeTimeoutBeforeApproval {
		return "", ErrTimeout
	}

	// 여기부터는 PG 쪽에서 승인이 끝나 approved에 기록됩니다. outcome에 따라 거래 번호를 돌려주느냐,
	// ErrTimeout으로 응답을 잃어버리느냐만 다릅니다.
	transactionID := fmt.Sprintf("pg-%s-%d", key, g.calls)
	g.approved[key] = transactionID
	if outcome == OutcomeTimeoutAfterApproval {
		return "", ErrTimeout
	}
	return transactionID, nil
}

// Inquire 메서드는 key로 승인된 거래가 있으면 그 거래 번호와 true를 돌려줍니다.
func (g *FakeGateway) Inquire(ctx context.Context, key string) (string, bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	transactionID, ok := g.approved[key]
	return transactionID, ok, nil
}

// Calls 메서드는 지금까지 Approve가 실행된 횟수를 돌려줍니다.
func (g *FakeGateway) Calls() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls
}
