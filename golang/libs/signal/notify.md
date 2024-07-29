# Notify

## `Notify`?

`Notify` causes package signal to relay incoming signals to c. If no signals are provided, all incoming signals will be relayed to c. Otherwise, just the provided signals will.
Package signal will not block sending to c: the caller must ensure that c has sufficient buffer space to keep up with the expected signal rate. For a channel used for notification of just one signal value, a buffer of size 1 is sufficient.
It is allowed to call Notify multiple times with the same channel: each call expands the set of signals sent to that channel. The only way to remove signals from the set is to call Stop.
It is allowed to call Notify multiple times with different channels and the same signals: each channel receives copies of incoming signals independently.

## 동작 원리

Go 런타임과 운영 체제(OS)의 시그널 핸들링 메커니즘을 활용하기 때문, Go의 `signal.Notify` 함수를 사용하여 시스템 시그널을 처리하는 방식은 실제로 매우 효율적이다. 폴링(polling)처럼 지속적으로 시그널을 확인하기 위해 CPU 시간을 소모하지 않는다.

### 이벤트 기반 시그널 처리

- Go의 `signal.Notify`는 내부적으로 OS의 시그널 핸들링 기능을 사용한다.
- OS는 프로세스에 시그널이 전송될 때 이를 처리하는 방식을 제공하고, Go 런타임은 이 시그널을 받아 `signal.Notify`에 등록된 채널로 전달

### 운영 체제의 역할

- 운영 체제는 시그널을 감지하고 프로세스에 알리는 역할을 한다
- 프로세스가 특정 시그널을 받을 준비가 되어 있으면, OS는 해당 시그널을 프로세스에 전달한다
- 이 과정에서 추가적인 CPU 자원을 사용하는 것은 OS의 시그널 핸들러이며, Go 프로그램 자체는 이벤트가 발생할 때까지 대기 상태에 있는다

### 고루틴과 블로킹

- `signal.Notify`를 사용할 때, Go 런타임은 내부적으로 고루틴을 사용하여 이 시그널들을 관리한다
- 이 고루틴은 `<-quit`과 같은 코드에서 블로킹되어 있으며, 시그널이 발생하면 해당 고루틴은 블로킹에서 해제되고 실행을 이어간다

### 이벤트 대기

- Go 프로그램이 시그널을 기다리는 동안, 실제로는 CPU 시간을 소모하지 않는다
- 운영 체제가 시그널을 감지하고 적절한 프로세스에 전달하는 방식으로, 프로세스는 이벤트가 발생하기 전까지 대기 상태에 있는다
- 전통적인 폴링 방식보다 훨씬 효율적

## Example

```go
// Set up channel on which to send signal notifications. 
// We must use a buffered channel or risk missing the signal 
// if we're not ready to receive when the signal is sent. 
c := make(chan os.Signal, 1) 
signal.Notify(c, os.Interrupt)  

// Block until a signal is received. 
s := <-c 
fmt.Println("Got signal:", s)
```

```go
// allSignals example
// Set up channel on which to send signal notifications. 
// We must use a buffered channel or risk missing the signal 
// if we're not ready to receive when the signal is sent. 
c := make(chan os.Signal, 1)  

// Passing no signals to Notify means that 
// all signals will be sent to the channel. 
signal.Notify(c)  

// Block until any signal is received. 
s := <-c 
fmt.Println("Got signal:", s)
```

```go
go func() {
    defer wg.Done()
    quit := make(chan os.Signal, 1)
    // 지정된 종료 시그널들이 채널로 릴레이 되도록 설정
    // `signal.Notify`는 내부적으로 시그널을 캐치하는 데 사용되는 고루틴을 생성하고,
    // 이 고루틴은 별도의 스레드를 점유하지 않고, 운영체제로부터 시그널을 받으면 그 시그널을 quit 채널로 전송한다.
    // 이벤트 기반으로 동작하므로, 시그널이 발생하기 전까지는 리소스를 소모하지 않는다
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) 

    // 고루틴은 quit 채널에서 시그널을 기다린다. 이 부분은 **블로킹**되며, 시그널이 수신될 때까지 대기
    sig := <-quit // SIGINT 또는 SIGTERM 시그널이 발생할 때까지 이 라인에서 멈춰 있는다.

    // 
}()
```
