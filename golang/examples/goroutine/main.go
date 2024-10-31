package main

import (
    "fmt"
    "strconv"
    "strings"
    "sync"
    "time"
)

// completeAndDone 함수는 complete 채널 통해 어떤 로직을 완료했음을 수신하면,
// 리턴한 done 채널로 후속 처리를 할 수 있도록 알리는 별도의 고루틴을 생성하는 함수입니다.
func completeAndDone() (chan bool, chan bool) {
    complete := make(chan bool)
    done := make(chan bool)
    go func() {
        <-complete
        done <- true
    }()

    return complete, done
}

// goroutineSimple 함수는 간단하게 별도의 고루틴에서 문자열을 출력하는 예제입니다.
func goroutineSimple() chan bool {
    complete, done := completeAndDone()

    printMessage := func(message string) {
        for i := 0; i < 5; i++ {
            fmt.Println(message, i)
            time.Sleep(1 * time.Millisecond) // 1초 대기
        }
        complete <- true
    }

    go printMessage("Hello from Goroutine!") // 고루틴 실행
    println("Hello from Main!")              // 메인 고루틴에서 실행

    return done
}

// goroutineUnbufferedChan 함수는 비버퍼 채널을 통해서 메시지를 송신하고 수신하는 예제입니다.
// 고루틴은 송신자와 수신자가 동시에 준비되어 있어야 하므로, 메인 고루틴에서 동시에 실행될 수 없습니다.
// 자세한 사항은 deadlockByUnbufferedChannel 함수를 참고합니다.
//
// 따라서 송신자 또는 수신자 둘 중 적어도 하나는 별도의 고루틴에서 실행되어야 합니다.
func goroutineUnbufferedChan() chan bool {
    complete, done := completeAndDone()

    sendMessage := func(ch chan string) {
        ch <- "Hello via chan!" // 채널을 통해 메시지 전송
    }

    // 비버퍼 채널을 생성합니다.
    // 송신자가 데이터를 전송하면 수신자가 받을 때까지 블로킹됩니다.
    ch := make(chan string)
    go sendMessage(ch) // 고루틴 실행
    fmt.Println(<-ch)  // Hello via chan!

    message2 := "Initialized"
    go func() {
        message2 = <-ch // 수신자가 채널을 대기합니다.
    }()

    // 수신자 고루틴이 준비될 때까지 송신자는 대기합니다.
    sendMessage(ch)
    fmt.Println("When receiver is in goroutine, message is", message2) // When receiver is in goroutine, message is Hello via chan!

    complete <- true
    close(ch)
    return done
}

// goroutineBufferedChan 함수는 버퍼 채널 통해 메시지를 송신하고 수신하는 예제입니다.
// 버퍼 채널의 경우에는 송신 후 즉시 다음 코드로 넘어가므로, 별도의 고루틴으로 실행할 필요가 없습니다.
func goroutineBufferedChan() chan bool {
    complete, done := completeAndDone()

    sendMessage := func(ch chan string) {
        // 고루틴이 채널에 두 개의 메시지를 전송할 때,
        // 채널의 버퍼에 저장되고 메인 고루틴에서 이를 수신합니다.
        ch <- "Hello from Goroutine!"
        ch <- "Another message from Goroutine!"
    }

    // 버퍼 크기 2인 채널을 생성합니다.
    ch := make(chan string, 2)
    go sendMessage(ch)

    // 두 개의 메시지를 채널에서 수신
    fmt.Println(<-ch)
    fmt.Println(<-ch)
    complete <- true

    return done
}

// goroutineWaitGroup 함수는 sync.WaitGroup 사용한 고루틴 제어 예제입니다.
func goroutineWaitGroup() chan bool {
    complete, done := completeAndDone()

    printMessage := func(message string, wg *sync.WaitGroup) {
        // 고루틴이 종료될 때마다 wg.Done()을 호출하여 WaitGroup에 알립니다.
        defer wg.Done()
        for i := 0; i < 3; i++ {
            fmt.Println(message, i)
            time.Sleep(1 * time.Second) // 메인 고루틴이 종료되지 않도록 대기
        }
    }

    // 두 개의 고루틴이 모두 완료될 때까지 메인 고루틴이 기다립니다.
    var wg sync.WaitGroup // WaitGroup 생성
    // 두 개의 고루틴을 기다립니다.
    wg.Add(2)

    go printMessage("First Goroutine", &wg)
    go printMessage("Second Goroutine", &wg)

    wg.Wait() // 두 개의 고루틴이 끝날 때까지 대기
    complete <- true

    return done
}

// goroutineChanSelect 함수는 select 문을 사용하여 메시지를 수신하는 예제입니다.
func goroutineChanSelect() chan bool {
    complete, done := completeAndDone()

    sendToChannel1 := func(ch chan string) {
        time.Sleep(1 * time.Second)
        ch <- "Message from Channel 1"
    }

    sendToChannel2 := func(ch chan string) {
        time.Sleep(2 * time.Second)
        ch <- "Message from Channel 2"
    }

    ch1 := make(chan string)
    ch2 := make(chan string)

    go sendToChannel1(ch1)
    go sendToChannel2(ch2)

    for i := 0; i < 2; i++ {
        // select문은 여러 채널에서 데이터를 비동기적으로 수신할 수 있습니다.
        // select 구문은 채널에서 데이터가 들어오기를 기다리면서 비동기적으로 동작하기 때문에, 해당 채널에 데이터가 없다면
        // - 해당 고루틴은 블로킹되고
        // - 고루틴은 스케줄러에 의해 비활성화되어 CPU를 소비하지 않습니다
        // 즉, 채널에 데이터가 없을 때는 CPU를 사용하지 않고 대기 상태로 들어갑니다.
        // Go의 스케줄러는 채널에서 이벤트가 발생할 때 고루틴을 다시 활성화하여 데이터를 처리하게 하며, 이벤트가 발생하지 않으면 CPU를 차지하지 않게 관리합니다.
        select {
        case msg1 := <-ch1:
            fmt.Println(msg1)
        case msg2 := <-ch2:
            fmt.Println(msg2)
        }
    }
    complete <- true

    return done
}

// goroutineChanUntilClose  함수는 채널을 통해 데이터를 송신하고, 채널이 닫힐 때까지 데이터를 수신하는 예제입니다.
func goroutineChanUntilClose() chan bool {
    complete, done := completeAndDone()

    sendMessages := func(ch chan string) {
        ch <- "Message 1"
        ch <- "Message 2"
        // close(ch)는 채널을 닫습니다.
        // 채널이 닫히면 더 이상 송신할 수 없지만, 수신은 계속 가능합니다.
        close(ch)
    }

    ch := make(chan string)
    go sendMessages(ch)

    for msg := range ch { // 채널이 닫힐 때까지 데이터를 수신합니다.
        fmt.Println(msg)
    }

    fmt.Println("Channel closed")
    complete <- true

    return done
}

// goroutineMultiplexingAndTimeout 함수는 select 문을 사용하여 여러 채널을 동시에 대기하는 멀티플렉싱 예제입니다.
// 멀티플렉싱은 하나의 통로(여기서는 select 문)를 통해 여러 입력을 동시에 처리할 수 있는 기술입니다
//
// 이 예제에서는 select 문을 통해 여러 채널을 동시에 대기하고, 준비된 case 채널의 값을 선택적으로 처리합니다.
func goroutineMultiplexingAndTimeout() chan bool {
    complete, done := completeAndDone()

    ch := make(chan string)

    go func() {
        time.Sleep(3 * time.Second)
        select {
        case ch <- "Message from Goroutine":
            // 메시지 전송 성공
        default:
            // 수신할 고루틴이 없어서 블록될 경우
            return
        }
    }()

    now := time.Now()
    fmt.Println("Before select", now.Format(time.DateTime)) // Before select 2024-10-31 17:24:03
    // select와 time.After를 사용하여 타임아웃을 처리할 수 있습니다.
    // select문에서 타임아웃이 발생하면 타임아웃 메시지를 출력하고,
    // 그렇지 않으면 고루틴에서 보낸 메시지를 출력합니다.
    select {
    // select는 다음과 같이 동작합니다:
    // 1. 모든 case를 평가하여 즉시 실행 가능한 case가 있는지 확인
    // 2. 실행 가능한 case가 없으면 모든 채널에 대해 대기(블록킹)
    // 3. 어떤 채널이든 준비되면 해당 case 실행
    //
    // 아래 경우 select에서 블록킹되다가
    //'ch 채널에서 메시지를 받기' 또는 '2초가 지나서 time.After 채널에서 값이 수신'
    // 둘 중 하나가 발생할 때 재개됩니다.
    //
    // Reference:
    // - https://github.com/golang/go/blob/6d39245514c675cdea5c7fd7f778e97bf0728dd5/src/runtime/select.go#L107-L528
    case msg := <-ch:
        fmt.Println(msg)
    case <-time.After(2 * time.Second):
        fmt.Println("Timeout!")
        complete <- true
    }
    fmt.Println("After select", time.Since(now)) // After select 2.00148275s

    return done
}

// goroutineBufferedChannelAsSemaphore 함수는 버퍼링된 채널을 세마포어처럼 사용하여
// 동시에 실행될 수 있는 고루틴 수를 제한하는 예제입니다.
func goroutineBufferedChannelAsSemaphore() chan bool {
    complete, done := completeAndDone()

    // 동시에 실행될 수 있는 고루틴의 수를 제한합니다.
    maxGoroutineNumber := 3
    // Semaphore:
    // - 동시성 제어를 위해 고안된 동기화 메커니즘으로, 주로 공유 자원에 대한 접근을 제한하여 경합을 방지하는 데 사용
    // - 세마포어의 기본 개념은 공유 자원의 사용 가능 수량을 추적하는 카운터
    // - 자원이 사용될 때 감소하고, 해제되면 증가
    semaphore := make(chan bool, maxGoroutineNumber) // 고루틴 수를 maxGoroutineNumber 수만큼 제한합니다.
    loopMax := 10

    for i := 0; i < loopMax; i++ {
        // 각 고루틴이 시작될 때 semaphore 채널에 true를 보냅니다.
        // 사용 가능한 세마포어 수를 감소시킵니다.
        semaphore <- true
        go func(num int) {
            // defer는 고루틴 내부에서 에러 등의 이유로 비정상적으로 종료될 때도 보장됩니다.
            defer func() {
                // semaphore에서 값을 꺼내어 세마포어를 해제하고,
                // 다른 대기 중인 고루틴이 실행될 수 있도록 합니다.
                <-semaphore
            }()
            fmt.Println("[" + strconv.Itoa(num) + "] message, and sleep 2 sec")
            time.Sleep(2 * time.Second)
            if num == loopMax-1 {
                complete <- true
            }
        }(i)
    }

    return done
}

func goroutineCommunicateAfterChannelClosed() chan bool {
    complete, done := completeAndDone()

    // 버퍼 채널을 생성합니다.
    ch := make(chan string, 2)

    ch <- "first"
    ch <- "second"
    // ch <- "third" 버퍼가 2인데 메시지를 더 송신하려고 하면 버퍼가 빌 때까지 무한 대기 상태가 됩니다.
    //               이 경우 메인 고루틴에서 발생하므로 데드락 상태가 됩니다.
    //               fatal error: all goroutines are asleep - deadlock!

    close(ch)

    fmt.Println(1, <-ch)
    fmt.Println(2, <-ch)
    // Output:
    //  1 first
    //  2 second

    complete <- true
    return done
}

/*
데드락(Deadlock)은 두 개 이상의 고루틴 또는 프로세스가 자원 점유와 대기를 통해 서로 간의 작업을 방해하면서 무한 대기 상태에 빠지는 현상입니다.
Go에서는 주로 뮤텍스(Mutex), 세마포어, 채널 등을 사용할 때 발생할 수 있습니다.

데드락이 발생하기 위해서는 다음 네 가지 조건이 모두 성립해야 합니다:
- 상호 배제(Mutual Exclusion): 자원은 한 번에 하나의 고루틴만 사용할 수 있습니다. 즉, 자원을 배타적으로 사용해야 합니다.
- 점유와 대기(Hold and Wait): 고루틴이 이미 자원을 점유하고 있으면서, 추가 자원을 기다리는 상태에 있어야 합니다.
- 비선점(No Preemption): 다른 고루틴이 점유 중인 자원을 강제로 빼앗을 수 없어야 합니다.
- 순환 대기(Circular Wait): 자원을 점유한 고루틴들이 서로의 자원을 기다리는 순환 대기 상태가 발생해야 합니다.
*/

// deadlockMutualExclusionByGoroutineWithFatalError 함수는 상호 배제(Mutual Exclusion)로 인한 데드락 예제입니다.
//
// Goroutine 1이 mu 뮤텍스를 잠금 상태로 유지하고 있는 동안 해제를 하지 않고 끝나게 되어, Goroutine 2는 무한 대기 상태에 빠지며 데드락이 발생합니다.
// 이는 상호 배제(Mutual Exclusion)와 점유와 대기(Hold and Wait)를 충족하여 데드락이 발생합니다.
//
// Output:
//   Goroutine 1: Locked
//   fatal error: all goroutines are asleep - deadlock!
func deadlockMutualExclusionByGoroutineWithFatalError() chan bool {
    complete, done := completeAndDone()
    var mu sync.Mutex
    go func() {
        mu.Lock() // 자원 잠금
        fmt.Println("Goroutine 1: Locked")
        // 어떤 작업 수행
        time.Sleep(2 * time.Second)
        // 자원 해제
        // 만약 여기서 Goroutine 1이 자원을 해제하지 않는다면, Goroutine 2는 영원히 대기 상태에 빠지며 데드락이 발생합니다.
        // fatal error: all goroutines are asleep - deadlock!
        // mu.Unlock()
        done <- true
    }()

    go func() {
        time.Sleep(1 * time.Second) // 잠시 대기
        mu.Lock()                   // 다른 고루틴이 해제하기 전까지 대기
        fmt.Println("Goroutine 2: Locked")
        mu.Unlock() // 자원 해제
        done <- true
    }()

    <-done
    <-done
    fmt.Println("Completed")
    complete <- true

    return done
}

// deadlockHoldAndWaitWithFatalError 함수는 자원을 점유한 채 다른 점유된 자원을 기다리면서(Hold and Wait) 발생하는 데드락 예제입니다.
//
// 두 개의 고루틴이 각각 다른 자원을 점유한 상태에서 상대방의 자원을 기다리며 데드락이 발생합니다.
// 상호 배제(Mutual Exclusion), 점유와 대기(Hold and Wait), 순환 대기(Circular Wait) 조건을 충족하여 데드락이 발생합니다.
//
// Output:
//
//      deadlockHoldAndWaitWithFatalError 함수 내부에서 실행되는 두 개의 고루틴 39와 40이 각각 락을 획득하려고 하지만,
//      두 고루틴 모두 다른 자원이 해제되기를 기다리며 서로 대기 상태에 빠져 데드락이 발생합니다.
//
//      ========== deadlockHoldAndWaitWithFatalError Start ==========
//      Goroutine 2: Locked mu2
//      Goroutine 1: Locked mu1
//      fatal error: all goroutines are asleep - deadlock!
//
//      ... 생략 ...
//
//      goroutine 39 [sync.Mutex.Lock]:
//      sync.runtime_SemacquireMutex
//      sync.(*Mutex).lockSlow
//      sync.(*Mutex).Lock
//      main.deadlockHoldAndWaitWithFatalError.func1()
//
//      goroutine 40 [sync.Mutex.Lock]:
//      sync.runtime_SemacquireMutex
//      sync.(*Mutex).lockSlow
//      sync.(*Mutex).Lock
//      main.deadlockHoldAndWaitWithFatalError.func2()
func deadlockHoldAndWaitWithFatalError() chan bool {
    var mu1, mu2 sync.Mutex
    complete, done := completeAndDone()
    goroutine1Complete := make(chan bool)
    goroutine2Complete := make(chan bool)

    go func() {
        mu1.Lock() // 첫 번째 자원 점유
        fmt.Println("Goroutine 1: Locked mu1")
        time.Sleep(1 * time.Second)

        mu2.Lock() // 두 번째 자원을 요청하며 대기
        fmt.Println("Goroutine 1: Locked mu2")
        mu2.Unlock()
        mu1.Unlock()
        goroutine1Complete <- true
    }()

    go func() {
        mu2.Lock() // 두 번째 자원 점유
        fmt.Println("Goroutine 2: Locked mu2")
        time.Sleep(1 * time.Second)

        mu1.Lock() // 첫 번째 자원을 요청하며 대기
        fmt.Println("Goroutine 2: Locked mu1")
        mu1.Unlock()
        mu2.Unlock()
        goroutine2Complete <- true
    }()

    <-goroutine1Complete
    <-goroutine2Complete
    <-complete
    fmt.Println("Completed")

    return done
}

// deadlockCircularWaitWithFatalError 함수는 고루틴들이 순환 대기(Circular Wait) 상태에 빠져 발생하는 데드락 예제입니다.
//
// 세 개의 고루틴이 각각 자원을 점유한 상태에서 다른 자원을 요청하며 순환 구조가 형성되어 데드락이 발생합니다.
// 상호 배제(Mutual Exclusion), 점유와 대기(Hold and Wait), 순환 대기(Circular Wait)를 충족하여 데드락이 발생합니다.
//
// Output:
//
//      deadlockCircularWaitWithFatalError 함수 내부에서 실행되는 세 개의 고루틴 35, 36, 37이 각각 락을 획득하려고 하지만,
//      각 고루틴이 서로의 자원을 기다리며 순환 대기 상태가 발생하고, 이로 인해 데드락이 발생합니다.
//
//      ========== deadlockCircularWaitWithFatalError Start ==========
//      Goroutine 3: Locked mu3
//      Goroutine 2: Locked mu2
//      Goroutine 1: Locked mu1
//      fatal error: all goroutines are asleep - deadlock!
//
//      ... 생략 ...
//
//      goroutine 35 [sync.Mutex.Lock]:
//      sync.runtime_SemacquireMutex(0x14000108f58?, 0xac?, 0x30181ff894bc1?)
//      sync.(*Mutex).lockSlow(0x14000104028)
//      sync.(*Mutex).Lock(...)
//      main.deadlockCircularWaitWithFatalError.func1()
//
//      goroutine 36 [sync.Mutex.Lock]:
//      sync.runtime_SemacquireMutex(0x14000133758?, 0x8c?, 0x30181ff889e17?)
//      sync.(*Mutex).lockSlow(0x14000104030)
//      sync.(*Mutex).Lock(...)
//      main.deadlockCircularWaitWithFatalError.func2()
//
//      goroutine 37 [sync.Mutex.Lock]:
//      sync.runtime_SemacquireMutex(0x14000133f58?, 0x6c?, 0x30181ff886a7e?)
//      sync.(*Mutex).lockSlow(0x14000104020)
//      sync.(*Mutex).Lock(...)
//      main.deadlockCircularWaitWithFatalError.func3()
func deadlockCircularWaitWithFatalError() chan bool {
    var mu1, mu2, mu3 sync.Mutex
    complete, done := completeAndDone()
    goroutine1Complete := make(chan bool)
    goroutine2Complete := make(chan bool)
    goroutine3Complete := make(chan bool)

    go func() {
        mu1.Lock() // 첫 번째 자원 점유
        fmt.Println("Goroutine 1: Locked mu1")
        time.Sleep(1 * time.Second)

        mu2.Lock() // 두 번째 자원 요청하며 대기
        fmt.Println("Goroutine 1: Locked mu2")
        mu2.Unlock()
        mu1.Unlock()
        goroutine1Complete <- true
    }()

    go func() {
        mu2.Lock() // 두 번째 자원 점유
        fmt.Println("Goroutine 2: Locked mu2")
        time.Sleep(1 * time.Second)

        mu3.Lock() // 세 번째 자원 요청하며 대기
        fmt.Println("Goroutine 2: Locked mu3")
        mu3.Unlock()
        mu2.Unlock()
        goroutine2Complete <- true
    }()

    go func() {
        mu3.Lock() // 세 번째 자원 점유
        fmt.Println("Goroutine 3: Locked mu3")
        time.Sleep(1 * time.Second)

        mu1.Lock() // 첫 번째 자원 요청하며 대기
        fmt.Println("Goroutine 3: Locked mu1")
        mu1.Unlock()
        mu3.Unlock()
        goroutine3Complete <- true
    }()

    <-goroutine1Complete
    <-goroutine2Complete
    <-goroutine3Complete
    <-complete
    fmt.Println("Completed")

    return done
}

// deadlockByUnbufferedChannel 함수는 비버퍼 채널 사용 시 수신자가 없을 때 발생하는 데드락 예제입니다.
//
// 비버퍼 채널은 송신자와 수신자가 동시에 준비되어 있어야만 데이터가 전달됩니다.
// 즉, 비버퍼 채널에서는 송신자가 데이터를 보내려 할 때 수신자가 준비되지 않았다면 송신자는 대기 상태에 빠집니다.
//
// 송신자가 채널로 `ch <- value`로 데이터를 보내려 할 때,
// 수신자가 `<-ch`로 데이터를 받을 준비가 되어 있지 않으므로 송신자는 대기 상태에 빠지고,
// 프로그램 전체가 데드락 상태가 됩니다.
//
// 정상 동작하는 경우는 goroutineUnbufferedChan 함수를 참고합니다.
//
// Output:
//
//      fatal error: all goroutines are asleep - deadlock!
func deadlockByUnbufferedChannel() chan bool {
    complete, done := completeAndDone()
    ch := make(chan int) // 비버퍼 채널 생성

    // 수신자가 없기 때문에,
    ch <- 42 // 여기서 영원히 대기 상태에 빠집니다.
    fmt.Println("Sent 42")

    value := <-ch // 수신 대기하는 것처럼 보이지만, 애초에 위에서 대기 상태에 빠지므로 의미가 없으며 Fatal error가 발생합니다.
    complete <- true
    fmt.Println("Complete", value)

    return done
}

func runner(f func() chan bool, name string) {
    logBanner := strings.Repeat("=", 10)

    fmt.Println(fmt.Sprintf("%s %s Start %s", logBanner, name, logBanner))
    done := f()
    <-done
    close(done)
    fmt.Println(fmt.Sprintf("%s %s End %s\n", logBanner, name, logBanner))
}

func main() {
    runner(goroutineSimple, "goroutineSimple")
    runner(goroutineUnbufferedChan, "goroutineUnbufferedChan")
    runner(goroutineBufferedChan, "goroutineBufferedChan")
    runner(goroutineWaitGroup, "goroutineWaitGroup")
    runner(goroutineChanSelect, "goroutineChanSelect")
    runner(goroutineChanUntilClose, "goroutineChanUntilClose")
    runner(goroutineMultiplexingAndTimeout, "goroutineMultiplexingAndTimeout")
    runner(goroutineBufferedChannelAsSemaphore, "goroutineBufferedChannelAsSemaphore")
    runner(goroutineCommunicateAfterChannelClosed, "goroutineCommunicateAfterChannelClosed")

    // Deadlock 예제
    // runner(deadlockMutualExclusionByGoroutineWithFatalError, "deadlockMutualExclusionByGoroutineWithFatalError")
    // runner(deadlockHoldAndWaitWithFatalError, "deadlockHoldAndWaitWithFatalError")
    // runner(deadlockCircularWaitWithFatalError, "deadlockCircularWaitWithFatalError")
    // runner(deadlockByUnbufferedChannel, "deadlockByUnbufferedChannel")

    fmt.Println("Done")
}
