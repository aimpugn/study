package concurrency

import (
	"sync"
	"sync/atomic"
	"time"
)

// Pause 타입은 함수가 값을 확인한 뒤 그 결과로 쓰기 전까지의 구간에서 잠시 멈출 수 있게 끼워 둔 자리입니다.
// 이 패키지는 확인과 쓰기 사이의 구간을 "틈"이라고 부릅니다.
//
// 운영 코드는 nil을 넘기고, 테스트는 Rendezvous.Wait를 넘겨 두 요청이 같은 틈에
// 함께 머물도록 만듭니다. 이 인자를 받는 함수는 틈에 다른 요청이 끼어들 수 있는 함수이고,
// 받지 않는 함수는 확인과 쓰기를 SQL 한 문장 안에서 끝내 틈이 없는 함수입니다. 시그니처만 보고도 틈이 있는지 알 수 있게 하려는 설계입니다.
type Pause func()

// here 메서드는 p가 nil이 아니면 그 자리에서 p를 실행해 멈춥니다. nil이면 아무 일도 하지 않습니다.
func (p Pause) here() {
	if p != nil {
		p()
	}
}

// Rendezvous 타입은 정해진 수의 고루틴이 모두 도착할 때까지 도착한 고루틴을 붙잡아 두는 약속 지점입니다.
//
// 경쟁 조건(race condition)은 두 요청이 같은 순간에 도착해서 생기지 않습니다. 한 요청이
// 틈에 머무는 동안 다른 요청이 같은 틈에 들어오면 생깁니다. 이 겹침을 타이밍 운에 맡기면
// 재현이 들쭉날쭉하므로, 이 타입으로 겹침을 강제합니다.
//
// 기다림은 timeout이 지나면 끝나며, 이 성질로 락이 실제로 상대를 막았는지 확인합니다.
// 요청 A와 B가 같은 락을 쓰고 둘 다 락 안에서 Wait를 부르면 이렇게 진행됩니다.
//
//  1. A가 락을 잡고 Wait에 도착합니다(arrived=1). B는 락을 기다리느라 Wait에 오지 못합니다.
//  2. timeout이 지나 A가 혼자 풀려나고 missed가 true가 됩니다. A는 진행해 커밋하고 락을 놓습니다.
//  3. B가 락을 잡고 Wait에 도착합니다(arrived=2). 인원이 찼으므로 바로 통과합니다.
//
// 락이 없으면 B도 1번에서 Wait에 도착해 둘이 함께 풀려나므로 missed는 false입니다. 테스트는 성공 수나
// 최종 잔액과 별도로 Missed 값을 봐서, 락이 B를 막았는지 확인합니다.
type Rendezvous struct {
	parties int
	timeout time.Duration

	mu      sync.Mutex
	arrived int
	full    chan struct{}
	missed  atomic.Bool
}

// NewRendezvous 함수는 parties개가 모이면 모두 풀어 주고, 한 고루틴이 timeout보다 오래
// 기다리면 그 고루틴을 먼저 풀어 주는 약속 지점을 만듭니다.
func NewRendezvous(parties int, timeout time.Duration) *Rendezvous {
	return &Rendezvous{parties: parties, timeout: timeout, full: make(chan struct{})}
}

// Wait 메서드는 약속 지점에 도착했음을 기록하고, 모두 모이거나 timeout이 지날 때까지 기다립니다.
//
// full 채널은 한 번 닫히면 계속 닫혀 있으므로, 인원이 찬 뒤에 도착한 고루틴은 기다리지 않고 바로 돌아갑니다.
// 그래서 WithdrawOptimistic의 재시도 루프처럼 같은 요청이 Wait를 여러 번 불러도 첫 시도에서만 멈춥니다.
func (r *Rendezvous) Wait() {
	// 도착 수(arrived)를 올리고, 이번 도착으로 parties만큼 모였으면 full 채널을 닫아 기다리던 고루틴을 모두 깨웁니다.
	r.mu.Lock()
	r.arrived++
	if r.arrived == r.parties {
		close(r.full)
	}
	r.mu.Unlock()

	// full이 닫히기를 기다리되, timeout이 먼저 지나면 상대가 오지 못했다고 missed에 기록하고 진행합니다.
	select {
	case <-r.full:
	case <-time.After(r.timeout):
		r.missed.Store(true)
	}
}

// Missed 메서드는 누군가 인원이 차기 전에 timeout으로 풀려났는지 알려 줍니다.
func (r *Rendezvous) Missed() bool {
	return r.missed.Load()
}

// Result 타입은 RunTogether가 실행한 작업 하나의 결과입니다.
type Result[T any] struct {
	Worker int
	Value  T
	Err    error
}

// RunTogether 함수는 fn을 고루틴 n개로 실행하되, 모두 출발선에 세운 뒤 한꺼번에 출발시킵니다.
//
// 한꺼번에 출발시키면 실행이 겹칠 가능성이 커질 뿐, 같은 순간에 실행된다는 보장은 없습니다.
// 각 고루틴이 언제 CPU에서 도는지는 Go 스케줄러와 운영체제가 정합니다. 겹침이 반드시
// 필요한 테스트는 fn 안에서 Rendezvous로 멈춰 세웁니다.
//
// 모든 fn이 끝난 뒤, worker 번호 순서로 담긴 결과를 돌려줍니다.
func RunTogether[T any](n int, fn func(worker int) (T, error)) []Result[T] {
	// start 채널은 출발 신호입니다. 닫히는 순간 <-start에서 기다리던 고루틴이 모두 깨어납니다.
	start := make(chan struct{})
	var ready, done sync.WaitGroup
	ready.Add(n)
	done.Add(n)
	results := make([]Result[T], n)

	for worker := range n {
		go func() {
			defer done.Done()
			// 출발선에 도착했다고 ready에 먼저 알립니다. 알린 직후 start가 닫혀도, 닫힌 채널에서
			// 받는 동작은 즉시 끝나므로 출발 신호를 놓치지 않습니다.
			ready.Done()
			<-start

			value, err := fn(worker)
			// 고루틴마다 results에서 자기 칸(results[worker])에만 쓰므로 락이 필요 없습니다.
			results[worker] = Result[T]{Worker: worker, Value: value, Err: err}
		}()
	}

	// 모든 고루틴이 ready에 알린 뒤 start를 닫고, done으로 모두 끝날 때까지 기다립니다.
	ready.Wait()
	close(start)
	done.Wait()
	return results
}
