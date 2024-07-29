# Go Memory Model

- [Go Memory Model](#go-memory-model)
    - [The Go Memory Model](#the-go-memory-model)
        - [Introduction](#introduction)
        - [조언](#조언)
        - [비공식 개요](#비공식-개요)
        - [Memory Model](#memory-model)
        - [메모리 모델](#메모리-모델)
    - [데이터 경합을 포함하는 프로그램에 대한 구현 제한 사항](#데이터-경합을-포함하는-프로그램에-대한-구현-제한-사항)
    - [동기화](#동기화)
        - [Initialization](#initialization)
        - [고루틴 생성](#고루틴-생성)
        - [고루틴 파괴](#고루틴-파괴)
        - [채널 통신](#채널-통신)
    - [`append` 함수를 사용하여 슬라이스에 요소를 추가하는 두 방식](#append-함수를-사용하여-슬라이스에-요소를-추가하는-두-방식)
        - [변수에 결과를 할당한 후 `append`하는 방식](#변수에-결과를-할당한-후-append하는-방식)
        - [직접 `append`하는 방식](#직접-append하는-방식)
        - [결론](#결론)
    - [전역 선언과 함수내 선언](#전역-선언과-함수내-선언)
        - [전역 선언](#전역-선언)
        - [함수내 선언](#함수내-선언)
        - [결론](#결론-1)

## [The Go Memory Model](https://go.dev/ref/mem)

### Introduction

> The Go memory model **specifies the conditions** under which *reads of a variable in one goroutine* can be guaranteed to observe values produced by *writes to the same variable in a different goroutine*.
>
> Go 메모리 모델은 *한 고루틴(A)에서 어떤 변수(a)를 읽는 행위*가 *다른 고루틴(B)에서 동일한 변수(a)에 쓰는 행위*로 생성된 값을 정확하게 인지할 수 있도록 보장하는 조건들을 명시한다.

`A` 고루틴에서 `a`라는 변수를 읽을 때, `B` 고루틴에서 `a` 변수에 `"test"`라는 값을 쓴다면, 이때 `A` 고루틴은 `B` 고루틴에 의해 `a` 변수에 쓰여진 `"test"` 값을 정확히 인지할 수 있어야 한다. 이는 Go 메모리 모델에 의해 보장된다.

이 메모리 모델의 핵심은 고루틴 간의 데이터 공유와 동시성 작업에서 데이터 일관성과 안전성을 유지하는 것이다.
이를 위해 Go는 *채널*, *뮤텍스*, *웨이트그룹* 등 다양한 동기화 메커니즘을 제공하여 이러한 조건들을 충족시킬 수 있게 돕는다.

### 조언

> Programs that modify data being simultaneously accessed by multiple goroutines must serialize such access.
> To serialize access, protect the data with channel operations or other synchronization primitives such as those in the `sync` and `sync/atomic` packages.
>If you must read the rest of this document to understand the behavior of your program, you are being too clever. Don't be clever.

여러 고루틴이 동시에 액세스하는 데이터를 수정하는 프로그램은 이러한 *액세스를 직렬화*해야 한다.
액세스를 직렬화하기 위해, 데이터를 채널 연산 또는 `sync`와 `sync/atomic` 패키지의 다른 동기화 기본 요소(primitives)로 보호한다.
만약 당신의 프로그램의 동작을 이해하기 위해 이 문서의 나머지를 반드시 읽어야 한다면, 너무 교묘하게 접근하고 있는 것이다. 교묘해지지 말 것.

### 비공식 개요

Go는 나머지 언어 부분과 마찬가지로 메모리 모델을 접근하며, 의미를 단순하고 이해하기 쉽고 유용하게 유지하는 것을 목표로 합니다.
이 섹션은 접근 방식에 대한 일반적인 개요를 제공하며, 대부분의 프로그래머에게 충분할 것이다. 메모리 모델은 다음 섹션에서 더 공식적으로 명시된다.

데이터 경합은 *어떤 메모리 위치에 대한 쓰기가 동시에 동일한 메모리 위치에 다른 읽기 또는 쓰기가 발생하는 것*으로 정의되며, 관련된 모든 접근이 `sync/atomic` 패키지에서 제공하는 원자적 데이터 접근인 경우에는 예외다.
이미 언급했듯이, 개발자는 데이터 레이스를 피하기 위해 적절한 동기화를 사용하도록 강력히 권장된다.
데이터 경합이 없는 경우, *Go 프로그램은 모든 고루틴이 단일 프로세서에 다중화된 것처럼 동작*한다.
이 속성은 때때로 `DRF-SC`로 언급된다: 데이터 경합이 없는 프로그램은 순차적으로 일관된 방식으로 실행된다

개발자는 데이터 경합이 없는 Go 프로그램을 작성해야 하지만, 데이터 경합에 대해 Go 구현이 할 수 있는 것에는 한계가 있다.
Go 구현은 데이터 경합에 대해 이를 보고하고 프로그램을 종료하는 것으로 항상 반응할 것이다.
그렇지 않으면, *단일 워드 크기* 또는 *서브 워드 크기*의 메모리 위치에 대한 각 읽기는 해당 위치에 (아마도 동시에 실행되는 다른 고루틴에 의해) *실제로 쓰였고 (다른 값으로) 아직 덮어쓰기 되지 않은 값*을 관찰해야 한다.

> **실제로 쓰였고 아직 덮어쓰기 되지 않은 값?**
>
> 메모리의 특정 위치에 대한 데이터 경합(data race) 상황에서, 한 스레드(또는 고루틴)에 의해 메모리 위치에 쓰여진 값이 다른 스레드에 의해 아직 덮어쓰여지지 않았을 때를 의미한다.
>
> 고루틴 A와 B가 있고, 두 고루틴이 동일한 메모리 위치에 접근한다고 가정해보자.
> - A 고루틴이 메모리 위치 X에 값을 쓰기(write)
> - B 고루틴이 거의 동시에 X를 읽기(read) 시도
> 여기서 Go의 메모리 모델은 B 고루틴이 X 위치에서 "실제로 쓰여진 최신 값"을 읽도록 보장한다.
> 즉, A 고루틴에 의해 쓰여진 값이 B 고루틴에 의해 "관찰(observed)"되어야 하며, 이 값은 아직 다른 값으로 덮어쓰기 되지 않은 상태여야 한다.
>
> - A 고루틴이 변수 a에 10을 쓰기
> - B 고루틴이 a를 읽기
> B는 A가 쓴 최신 값인 10을 읽게 된다.
> 만약 이후에 A 또는 다른 고루틴이 a에 20을 쓰게 되면, 그 쓰기 연산이 발생하기 전 B 고루틴의 읽기 연산은 여전히 10이라는 값을 관찰한다.
> "아직 덮어쓰기 되지 않았다"는 것은 새로운 쓰기 연산이 발생하기 전의 상태, 즉 가장 최근에 쓰여진 값이 아직 다른 값으로 변경되지 않았음을 의미한다.

프로그램의 경합이 완전히 정의되지 않고 컴파일러가 아무 것이나 할 수 있는 C/C++와 다르게, 이러한 구현 제약은 Go를 Java 또는 JavaScript와 유사하게 만들어 대부분의 경합은 제한된 수의 결과를 갖는다.
Go의 접근 방식은 프로그램을 더 신뢰성 있고 디버깅하기 쉽게 만드는 것을 목표로 하면서도, *데이터 경합이 오류*라는 점과 도구가 이를 진단하고 보고할 수 있다는 점을 강조한다.

> **`single-word`-sized 또는 `sub-word`-sized?**
>
> **단일 워드 크기**는 CPU가 한 번에 처리할 수 있는 데이터의 최대 크기를 의미하며, 프로세서의 레지스터 크기 및 데이터 버스 크기와 일치한다.
> 반면, **서브 워드 크기**는 이보다 작은 데이터 단위를 말하며, 더 작은 크기의 데이터를 효율적으로 처리할 수 있다.
> 이 두 개념은 컴퓨터 아키텍처와 메모리 관리에 있어서 중요한 기본 요소입니다.
>
> **단일 워드 크기(`single-word`-sized) 메모리?**
>
> - 단일 워드 크기는 컴퓨터 프로세서가 한 번의 연산이나 기계어 명령을 통해 저장 장치로부터 옮겨 놓을 수 있는 데이터의 단위를 의미한다. 이는 동시에 컴퓨터 프로세서가 한 번의 작업으로 처리할 수 있는 데이터의 최대 크기와도 동일하다.
> - 워드 크기는 프로세서의 레지스터 크기와 데이터 버스의 너비를 반영하며, CPU의 처리 능력과 시스템의 데이터 처리 효율성을 결정짓는 중요한 요소다.
> - 32비트 시스템에서는 4바이트(32비트)가 단일 워드 크기가 되며, 64비트 시스템에서는 8바이트(64비트)가 단일 워드 크기가 된다.
> - 워드 크기 = 레지스터의 크기(비트 수) = 메모리의 입/출력 단위 = 버스의 크기(선 갯수)
>
> **서브 워드 크기(`sub-word`-sized) 메모리?**
>
> - word integer 32bit(4byte)/64bit(8byte) size보다 작은 데이터
> - 단일 워드 크기보다 작은 데이터 단위. 이는 프로세서가 한 번에 처리할 수 있는 최대 데이터 크기보다 작은 데이터 조각을 의미한다.
> - 작은 데이터 단위의 처리에 사용되며, 효율적인 메모리 사용과 빠른 데이터 접근을 가능하게 한다. 주로 작은 크기의 데이터를 처리할 때 사용된다.
> - 32비트 시스템에서 2바이트(16비트) 또는 1바이트(8비트) 데이터는 서브 워드 크기에 해당한다.
>

### Memory Model

A *read-write* data race on memory location x consists of a read-like memory operation `r` on x and a write-like memory operation `w` on `x`, at least one of which is non-synchronizing, which are unordered by happens before (that is, neither `r` happens before `w` nor `w` happens before r).

A *write-write* data race on memory location x consists of two write-like memory operations `w` and `w'` on `x`, at least one of which is non-synchronizing, which are unordered by happens before.

Note that if there are no read-write or write-write data races on memory location `x`, then any read `r` on x has only one possible W(r): the single `w` that immediately precedes it in the happens before order.

More generally, it can be shown that any Go program that is data-race-free, meaning it has no program executions with read-write or write-write data races, can only have outcomes explained by some sequentially consistent interleaving of the goroutine executions. (The proof is the same as Section 7 of Boehm and Adve's paper cited above.) This property is called DRF-SC.

The intent of the formal definition is to match the DRF-SC guarantee provided to race-free programs by other languages, including C, C++, Java, JavaScript, Rust, and Swift.

Certain Go language operations such as goroutine creation and memory allocation act as synchronization operations. The effect of these operations on the synchronized-before partial order is documented in the “Synchronization” section below. Individual packages are responsible for providing similar documentation for their own operations.

-------------------------------------------------------------------------------------------

### 메모리 모델

다음 Go의 메모리 모델에 대한 공식 정의는 Hans-J. Boehm과 Sarita V. Adve가 PLDI 2008에서 발표한 “[C++ 동시성 메모리 모델의 기초](https://www.hpl.hp.com/techreports/2008/HPL-2008-56.pdf)”에 제시된 접근 방식을 밀접하게 따른다.
*데이터 경합이 없는 프로그램의 정의*와 *경합이 없는 없는 프로그램에 대한 순차적 일관성 보장*은 위 논문과 동일하다.

메모리 모델은 프로그램 실행에 대한 요구사항(requirements)을 설명한다. 프로그램 실행은 고루틴 실행들로 구성되며, 고루틴 실행들은 메모리 연산으로 구성된다.

*메모리 연산*은 네 가지 세부 사항으로 모델링된다:

- 일반 데이터 읽기, 일반 데이터 쓰기, 또는 동기화 작업(원자적 데이터 접근, 뮤텍스 연산 또는 채널 연산)인지 나타내는 그 종류(its kind)
- 프로그램 내의 위치
- 접근되는 메모리 위치 또는 변수
- 연산에 의해 읽거나 쓰여진 값들

> Some memory operations are *read-like*, including read, atomic read, mutex lock, and channel receive.
> Other memory operations are *write-like*, including write, atomic write, mutex unlock, channel send, and channel close. Some, such as atomic compare-and-swap, are both read-like and write-like.

*read-like* 메모리 연산
- 읽기(read)
- 원자적 읽기(atomic read)
- 뮤텍스 잠금(mutex lock)
- 채널 수신(channel receive)

*write-like* 메모리 연산
- 쓰기(read)
- 원자적 쓰기(atomic write)
- 뮤텍스 잠금 해제(mutex unlock)
- 채널 송신(channel send)
- 채널 닫기(channel close)

*read-like* 이면서 *write-like* 메모리 연산
- 원자적 비교-교환(atomic compare-and-swap)

*고루틴 실행*은 단일 고루틴이 실행한 메모리 연산들의 집합으로 모델링된다.

**요구사항 1**: 메모리에서 읽고 메모리에 쓰여지는 값들이 주어졌을 때, 각 고루틴 내의 메모리 연산은 해당 고루틴의 올바른 순차적 실행과 일치해야 한다.

> *각 고루틴 내의 메모리 연산이 순차적 실행과 일치해야 한다*는 말은, 프로그램이 동시성을 다루는 과정에서도 고루틴 내에서 발생하는 메모리 연산(데이터 읽기, 쓰기 등)은 개발자가 의도한 순서대로 실행되어야 함을 의미한다.
> 고루틴이 메모리에서 값을 읽고 쓰는 행위가 고루틴의 코드가 순차적으로 실행됐을 때의 결과와 일치해야 한다.

고루틴의 올바른 순차적 실행은 *순서 전(sequenced before)* 관계와 일치되어야 한다.

> 고루틴 내의 코드 실행 순서가 메모리 연산의 순서를 결정하는 근거가 되며, 이를 통해 고루틴의 실행이 예측 가능하고 일관된 방식으로 이루어질 수 있게 한다. 즉, *코드 내에서 작성된 명령어의 순서*가 *프로그램 실행 시 메모리에 적용되는 순서*와 일치해야 한다는 것을 의미한다.

*순서 전(sequenced before)* 관계는 [Go 언어 명세](https://go.dev/ref/spec)에 정의된 Go의 제어 흐름 구조에 대한 부분적인 순서 요구사항(partial order requirements)과 [표현식의 평가 순서](https://go.dev/ref/spec#Order_of_evaluation)로 정의된다.

> **순서 전(sequenced before)의 관계와 동기화 전(synchronized before)의 관계**
>
> 이 용어들은 프로그램의 연산 간에 정의된 순서 관계를 나타내기 위해 사용된다.
>
> - *Sequenced before 관계*: 특정 연산이 다른 연산보다 순서상 먼저 실행되어야 함을 의미한다. 예를 들어, 한 고루틴 내에서 변수에 값을 할당한 후 그 값을 다른 연산에서 사용하는 경우, 값의 할당은 사용하기 전에 "sequenced before" 관계를 통해 실행되어야 한다. 이를 통해 메모리 접근과 수정이 예측 가능한 순서로 발생하도록 보장한다.
> - *Synchronized before 관계*: 동기화 메커니즘(예: 채널, 뮤텍스)을 사용하여 서로 다른 고루틴 간의 연산 순서를 조정한다. 이 관계는 동기화 작업을 통해 한 고루틴의 연산 결과가 다른 고루틴에서 관찰될 수 있도록 순서를 정의한다. 예를 들어, 한 고루틴이 채널을 통해 메시지를 전송하고, 다른 고루틴이 그 메시지를 수신하는 경우, 메시지의 전송은 수신 작업 전("before")에 동기화되어야 한다.
>
> ```go
> var a string
> var c = make(chan int)
> 
> func f() {
>     a = "hello world" // 1. `a`에 "hello world"를 할당한다. `c <- 0` 전에 실행된다.
>     c <- 0            // 2. `c` 채널에 0을 보낸다. 이는 `a`의 할당 후에 실행되며,
>                       //    `main` 함수에서 `c`에서 값을 수신하기 전에 실행된다.
> }
>
> func main() {
>     go f()
>     <-c               // 3. `f` 함수 내에서 `c <- 0` 후에 `c` 채널에서 값을 수신한다.
>     print(a)          // 4. `c`에서의 수신 연산 후 `a`의 값을 출력한다.
> }
> ```
>
> - `a = "hello world"`와 `c <- 0`은 `f` 함수 내에서 순차적으로 실행된다("sequenced before").
> - `main` 함수에서는 `c` 채널에서의 수신 연산이 `f` 함수 내의 `c <- 0` 연산과 동기화되어 있다("synchronized before").

Go *프로그램 실행*은 고루틴 실행들의 집합과, 각 *read-like* 연산이 어떤 *write-like* 연산에서 읽는지를 지정하는 매핑 `W`로 모델링된다. (동일한 프로그램의 다수 실행은 다른 프로그램 실행을 가질 수 있다.)

> **매핑 `W`?**
>
> ```go
> var data int
> var done = make(chan bool)
> 
> func writeData() {
>     data = 1 // `data` 변수에 쓰기 연산
>     close(done) // `done` 채널을 통해 쓰기 연산이 완료되었음을 `readData` 함수에 알림
> }
> 
> func readData() {
>     <-done // `done` 채널 수신
>     fmt.Println(data) // `data` 변수 읽기 연산
> }
> 
> func main() {
>     go writeData()
>     readData()
> }
> ```
>
> `W` 매핑은 `readData` 함수 내의 `data` 변수 읽기 연산이 `writeData` 함수 내의 `data` 변수 쓰기 연산에서 값을 읽어야 한다는 것을 지정한다.
> `done` 채널을 통한 동기화는 `W` 매핑이 올바르게 설정되어, `readData` 함수가 항상 최신의 `data` 값을 읽을 수 있도록 보장한다.
> 이러한 `W` 매핑을 통해 메모리 연산 간의 순서와 일관성이 유지되며, 모든 고루틴이 동일한 메모리 상태를 기반으로 작업을 수행할 수 있도록 한다.

**요구사항 2**: 주어진 프로그램 실행에 대해, 동기화 작업(고루틴 간의 데이터 접근을 조정하기 위해 사용되는 작업. 예: 채널을 통한 데이터 전송, 뮤텍스 잠금/잠금 해제 등)에 한정될 때, (특정 *read-like* 연산이 어떤 *write-like* 연산으로부터 데이터를 읽는지 지정하는) `W` 매핑은 그 동기화 작업들의 암시적인 전체 순서에 의해 설명될 수 있어야 한다. 이 암시적인 전체 순서는 작업들의 실행 순서와 해당 작업들에 의해 읽거나 쓰여진 값들과 일관되어야 한다.

> **암시적 전체 순서(implicit total order)?**
>
> 프로그램의 동기화 메커니즘을 통해 정의되는 연산들의 순서를 의미한다.
> 이 순서는 프로그램 코드에 직접적으로 명시되어 있지 않지만, 동기화 작업을 통해 간접적으로 정의되며, 프로그램의 실행 과정에서 연산들이 수행되는 실제 순서를 반영한다.
>
> ```go
> package main
> 
> import (
>     "fmt"
>     "sync"
> )
> 
> var data int
> // `sync.WaitGroup`을 사용하는 동기화 메커니즘이 "암시적 전체 순서"를 생성하는 역할을 한다.
> // 코드에 직접적으로 명시되어 있지 않지만, `wg.Add(1)`, `wg.Done()`, `wg.Wait()` 등을 통해 
> // `writer` 고루틴과 `reader` 함수 사이의 순서가 정해진다.
> // 이러한 순서를 "암시적 전체 순서"라고 한다.
> var wg sync.WaitGroup
> 
> func writer() {
>     data = 42 // write operation
>     wg.Done() // 작업이 완료되었음을 알림
> }
> 
> func reader() {
>     // `writer` 고루틴의 쓰기 작업이 완료될 때까지 `reader` 함수를 대기.
>     // 이는 `writer` 고루틴의 쓰기 작업과 `reader` 함수의 읽기 작업 사이에 
>     // "synchronized before" 관계를 생성.
>     // `reader` 함수는 항상 `writer` 고루틴에 의해 업데이트된 최신 `data` 값을 읽게 된다
>     wg.Wait()
>     fmt.Println(data) // read operation
> }
> 
> func main() {
>     wg.Add(1) // `writer` 고루틴이 시작되기 전에 호출되어, 작업을 시작하기 위한 준비
>     go writer()
>     reader()
> }
> ```

*동기화 전(synchronized before)*의 관계는 `W`로부터 파생된 동기화 메모리 연산의 부분적인 순서(partial order)다. 동기화 *read-like* 메모리 연산 `r`이 동기화 *write-like* 메모리 연산 `w`를 관찰한다면(즉, `W(r) = w`인 경우), 그러면 `w`는 `r` 전에 동기화된다.

> *"동기화 전" 관계가 `W`로부터 파생된다*는 것은 프로그램의 실행에서 특정 동기화 메모리 연산들 간의 순서가 `W` 매핑에 의해 결정된다는 것을 의미한다. 즉, 프로그램의 동기화 메커니즘에 의해 정의된 메모리 접근 순서가 프로그램의 올바른 실행을 보장하는 데 중요한 역할을 한다.
>
> `W(r) = w` 관계: `w` 연산이 완료된 후에야 `r` 연산이 해당 값을 읽을 수 있는 관계
>
> 예를 들어,
> - 스레드 A가 변수 `x`에 쓰기 작업을 수행(`w`)
> - 다른 스레드 B가 나중에 변수 `x`에서 읽기 작업을 수행(`r`)
>
> 스레드 A의 쓰기 작업(`w`)이 스레드 B의 읽기 작업(`r`)보다 먼저 "동기화 전" 관계에 있으려면,
> 스레드 B의 읽기 작업(`r`)이 스레드 A의 쓰기 작업(`w`)을 "관찰"해야 한다(즉, `W(r) = w`).
>
> 이는 스레드 B가 스레드 A의 작업 결과를 "보고" 있음을 의미하며, 스레드 A의 작업(`w`)은 스레드 B의 작업(`r`) 전에 일어났다고 볼 수 있다.

비공식적으로, 동기화 전(synchronized before)의 관계는 이전 단락에서 언급된 암시적 전체 순서의 부분 집합이며, `W`가 직접 관찰하는 정보에 한정된다.

발생 전(happens before) 관계는 순서 전(sequenced before) 관계와 동기화 이전(synchronized before) 관계의 합집합에 대한 전이적 폐쇄로 정의된다.

> **발생 전(happens before)**
>
> 프로그램 내에서 두 연산 간의 순서를 정의하는 개념이다.
> 특히, 한 연산이 다른 연산보다 먼저 "발생한다"는 것은 프로그램의 실행 순서에서 첫 번째 연산이 두 번째 연산보다 앞서 실행되어야 함을 의미한다. 이는 메모리 접근이나 고루틴 간의 상호 작용을 정확하게 조율하기 위해 중요하다.
>
> **전이적 폐쇄(transitive closure)?**
>
> 어떤 관계에 대해, 그 관계가 간접적으로라도 적용될 수 있는 모든 경우를 포함시켜 확장한 것을 의미한다.
>
> 예를 들어,
> - 스레드 A 내에서
>     - 먼저 변수 `y`에 쓰기 작업을 수행하고(`w1`)
>     - 그 다음 `x`에 쓰기 작업을 수행(`w2`)
> - 스레드 B가 `x`에서의 쓰기 작업을 "관찰"(`W(r1) = w2`)
>
> `w1`은 `w2`보다 순서 전(sequenced before) 관계에 있다.
> `W(r1) = w2`에 따라 `w2`는 `r1`보다 앞서게 된다(synchronized-before 관계).
> 결과적으로 `w1`은 `w2`보다 앞서고, `w2`가 `r1`보다 앞서므로, `w1`도 `r1`보다 먼저 발생한다(happens-before 관계).

**요구사항 3**: 메모리 위치 `x`에서 일반적인 (동기화하지 않는) 데이터 읽기 `r`에 대해, `W(r)`은 `r`에게 *보이는* 쓰기 `w`여야 한다. 여기서 보이는 것은 다음 두 가지 모두를 만족할 때를 의미한다:

1. 쓰기 `w`는 읽기 `r` 전에 발생한다
2. 읽기 `r` 전에 일어나는 (*x*에 대한) 다른 쓰기 `w'` 전에 쓰기 `w`는 발생하지 않는다.

메모리 위치 `x`에서의 *read-write* 데이터 경합은 `x`에서 *read-like* 메모리 연산 `r`과 *write-like* 메모리 연산 `w`로 구성된다. 이 중 적어도 하나는 비동기화(non-synchronizing)이며, 이들은 *발생 전(happens before)*에 의해 정렬되지 않는다(즉, *read-like* 메모리 연산 `r`이 *write-like* 메모리 연산 `w` 전에 일어나지 않으며 *write-like* 메모리 연산 `w`가 *read-like* 메모리 연산 `r` 전에 일어나지 않는다).

> **정렬되지 않는다?**
>
> 두 연산 사이에 명확한 실행 순서가 프로그램에 의해 정의되지 않았음을 의미한다.
> 이러한 상황에서는 한 연산이 다른 연산보다 먼저 일어난다는 것을 보장할 수 없다.
> 이는 동시에 발생하는 것처럼 보일 수 있으며, 이 경우 데이터 경쟁이 발생할 수 있다.

메모리 위치 `x`에서의 *write-write* 데이터 경합은 `x`에서 두 개의 *write-like* 메모리 연산 `w`와 `w'`로 구성된다. 이 중 적어도 하나는 비동기화(non-synchronizing)이며, 이들은 *발생 전(happens before)*에 의해 정렬되지 않는다.

메모리 위치 `x`에서 *read-write* 또는 *write-write* 데이터 경합이 없다면, 메모리 위치 `x`에서의 모든 읽기 `r`은 하나의 가능한 `W(r)`만을 갖는다: *발생 전(happens before)* 순서에서 바로 앞서는 단일한 쓰기 `w`다.

> **발생 전 순서에서 바로 앞서는 쓰기?**
>
> 프로그램의 실행 과정에서 메모리 위치 `x`에 대한 쓰기 연산 `w`가 있고, 그 뒤에 읽기 연산 `r`이 이루어진다는 것을 의미한다.
> 즉, `w`와 `r` 사이에 다른 쓰기 연산이 존재하지 않으며, 프로그램의 실행 순서(*발생 전(happens before)* 순서)에 따라 `w` 연산이 `r` 연산 바로 전에 일어난다는 것을 나타낸다.

보다 일반적으로, *read-write* 또는 *write-write* 데이터 경합이 없는 모든 Go 프로그램은 고루틴 실행들의 어떤 순차적으로 일관된(consistent) 인터리빙(interleaving)에 의해서 설명되는 결과들만 가질 수 있다. (증명은 위에 인용된 Boehm과 Adve의 논문의 7장과 동일합니다.) 이 속성을 `DRF-SC`라고 한다.

> **고루틴 실행들의 어떤 순차적으로 일관된 인터리빙(Interleaving)?**
>
> 프로그램 내 여러 고루틴에서 수행되는 연산들이 서로 교차하면서도, 마치 한 번에 하나씩 순차적으로 실행되는 것처럼 결과를 낳는 실행 순서를 의미한다. 즉, 고루틴들 사이에서 발생하는 여러 연산들이 실제로는 동시에 병렬로 실행되더라도, 그 프로그램의 실행 결과는 마치 고루틴들이 순차적으로 실행된 것처럼 일관된 결과를 나타내야 한다는 것을 의미한다.
>
> ```go
> package main
> 
> import (
>     "fmt"
>     "sync"
> )
> 
> var (
>     a int
>     wg sync.WaitGroup
> )
> 
> func set(value int) {
>     a = value
>     wg.Done()
> }
> 
> func main() {
>     wg.Add(2)
>     go set(1) // 고루틴 1에서 변수 'a'에 1을 할당
>     go set(2) // 고루틴 2에서 변수 'a'에 2를 할당
> 
>     wg.Wait() // 모든 고루틴의 작업이 완료될 때까지 메인 고루틴이 대기하도록 한다
>     fmt.Println(a) // `a`의 최종 값은 두 고루틴 중 마지막으로 실행된 고루틴에 의해 할당된 값이 된다
> }
> ```
>
> 위 코드는 동기화 메커니즘(`sync.WaitGroup`)을 사용했기 때문에, 프로그램의 결과는 고루틴 실행들의 순차적으로 일관된 인터리빙에 의해 설명될 수 있다.

공식 정의의 의도는 C, C++, Java, JavaScript, Rust, Swift를 포함한 다른 언어들이 경합 없는 프로그램에 제공하는 `DRF-SC` 보장과 일치하는 것이다.

고루틴 생성과 메모리 할당과 같은 특정 Go 언어 연산은 동기화 작업으로 작동한다. 이러한 연산들이 *동기화 전(synchronized before)*의 부분 순서에 미치는 영향은 아래 "동기화" 섹션에서 문서화되어 있다. 개별 패키지는 자체 연산에 대해 유사한 문서화를 제공할 책임이 있다.

## 데이터 경합을 포함하는 프로그램에 대한 구현 제한 사항

앞 섹션에서는 데이터 경쟁이 없는 프로그램 실행에 대한 형식적 정의를 제공했다.
이 섹션에서는 경합을 포함하는 프로그램에 대해 구현이 반드시 제공해야 하는 의미론(semantics)을 비공식적으로 설명한다.

데이터 경쟁을 감지한 경우, 어떤 구현이든 경쟁을 보고하고 프로그램의 실행을 중단할 수 있다. `ThreadSanitizer`(“go build -race”로 접근)를 사용하는 구현은 정확히 이렇게 한다.

*배열, 구조체 또는 복소수(complex number)의 읽기*는 개별 하위 값(배열 요소, 구조체 필드 또는 실수/허수 구성 요소)의 읽기로 구현될 수 있으며, 순서는 임의로 정해진다.
마찬가지로, *배열, 구조체 또는 복소수(complex number)의 쓰기*는 개별 하위 값의 쓰기로 구현될 수 있으며, 순서는 임의로 정해진다.

machine word보다 크지 않은 값을 저장하는 메모리 위치 `x`의 읽기 `r`은
- 읽기 `r`이 쓰기 `w` 전에 발생하지 않으며
- 쓰기 `w`가 쓰기 `w'` 전에 발생하고 `w'`가 `r` 전에 발생하는 쓰기 `w'`가 없는
- 어떤 쓰기 `w`를 관찰해야 한다.

즉, 각 읽기는 선행하거나 동시에 이루어진 쓰기에 의해 작성된 값을 관찰해야 한다.

또한, 인과적이지 않고(acausal) 근거가 없는(out of thin air) 쓰기의 관찰은 허용되지 않는다.

단일 기계어 단어(single machine word)보다 큰 메모리 위치에 대한 읽기는 워드 크기의 메모리 공간과 같은 동일한 의미 체계(semantics)를 만족시키도록 권장되지만, 필수는 아니다. 이는 읽기가 허용된 단일 쓰기 `w`를 관찰해야 함을 의미한다.
성능상의 이유로, 구현체는 더 큰 연산을 순서가 지정되지 않은 개별 기계어 단어 크기의 연산들의 집합으로 처리할 수 있다.
이는 멀티워드 데이터 구조체에 대한 경합 상태가 단일 쓰기에 발생하지 않는 일관성 없는 값을 초래할 수 있음을 의미한다.
Go의 대부분의 구현에서 인터페이스 값, 맵, 슬라이스, 문자열과 같이 내부적인 (포인터, 길이) 또는 (포인터, 타입) 쌍의 일관성에 의존하는 값들의 경우, 이러한 경쟁 상태는 임의의 메모리 손상으로 이어질 수 있다.

> **워드 크기의 메모리 공간과 같은 동일한 의미 체계(semantics)를 만족시키도록 권장?**
>
> "동일한 의미 체계"란 무엇일까?
> - 프로그래밍 언어나 시스템 수준에서 "의미 체계": 변수 할당이나 함수 호출 등 코드가 실제로 어떻게 실행되는지, 즉 프로그램의 동작과 그 결과를 결정하는 규칙이나 원칙을 의미한다.
> - 메모리 접근과 관련된 "의미 체계": 하드웨어 아키텍처와 운영 체제가 제공하는 기본적인 데이터 접근 방식과 관련된 규칙을 다룬다.
>
> 컴퓨터 시스템에서 메모리는 데이터와 명령어를 저장하는 기본적인 수단이다.
> 메모리는 비트(bit)의 배열로 구성되며, 이러한 비트들은 더 큰 단위인 바이트(byte)와 워드(word)로 구분된다.
> "단일 기계어 단어(single machine word)" 크기의 메모리 위치에 대한 읽기와 쓰기 작업의 "의미 체계"는 *특정 아키텍처에서 정의한 기본 데이터 단위로의 접근*을 의미한다.
> 예를 들어, 64비트 아키텍처에서 단일 기계어 단어는 64비트(또는 8바이트)가 된다.
>
> 이러한 단위로 메모리에 접근할 때의 "의미 체계"는 다음과 같다.
> - 각 접근이 원자성(atomicity)을 가져야 한다. 즉, 한 번의 작업으로 완전히 수행되는 것이 보장되어야 하고, 다른 작업에 의해 중간 상태에서 관찰되거나 방해받지 않아야 한다.
> - 읽기 작업은 해당 메모리 위치에 기록된 최신 값을 반영해야 한다.
>
> 따라서 단일 기계어 단어보다 큰 메모리 위치에 대한 읽기가 "동일한 의미 체계를 만족시키도록 권장"된다는 것은, 이러한 큰 단위의 메모리 접근도 가능한 한 원자적이고, 일관된 최신의 쓰기 작업에 의해 기록된 값을 반영해야 함을 의미한다. 그러나 성능상의 이유로, 구현체는 이러한 큰 단위의 메모리 접근을 여러 개의 작은 단위로 나누어 처리할 수 있으며, 이는 단일 원자적 작업으로 처리될 때와는 다른 결과를 초래할 수 있다.

잘못된 동기화의 예는 아래 “잘못된 동기화” 섹션에 제공되어 있습니다.

구현의 제한에 대한 예는 아래 “잘못된 컴파일” 섹션에 제공되어 있습니다.

## 동기화

### Initialization

프로그램 초기화는 단일 고루틴에서 실행되지만, 해당 고루틴은 동시에 실행되는 다른 고루틴을 생성할 수 있다.

*package p가 package q를 임포트 한다면, q의 init 함수가 완료된 후에 p의 init 함수가 시작된다.*

*모든 init 함수의 완료는 `main.main` 함수가 시작되기 전에 동기화된다.*

> - 프로그램 초기화 과정에서 전역 변수와 패키지 수준에서 정의된 `init` 함수들이 호출되며, 이는 `main.main` 함수 실행 전에 이루어진다.
> - Go에서는 프로그램 초기화가 단일 고루틴에서 순차적으로 진행되지만, `init` 함수 내에서 다른 고루틴을 생성하여 병렬로 작업을 수행할 수 있다. 이는 Go의 동시성 모델을 활용한 효율적인 초기화 작업을 가능하게 한다.
> - 패키지 임포트 관계에 따라 `init` 함수의 호출 순서가 결정된다. p 패키지를 임포트하는 패키지 q의 `q.init` 함수는 임포트된 패키지 `p.init` 함수가 모두 완료된 후에 실행된다. 이는 초기화의 종속성을 관리하고 순환 종속성을 방지하는 데 중요하다.
> - 모든 `init` 함수의 실행이 완료된 후, 최종적으로 `main.main` 함수가 호출되어 프로그램의 주 실행 흐름이 시작된다. 이 단계에서 모든 전역 변수와 패키지 초기화가 이미 완료된 상태여야 하며, 이는 프로그램 실행의 안정성을 보장한다.
>
> ```go
> import (
>     "fmt"
> 
>     "github.com/path/to/sub"
>     "github.com/path/to/sub2"
> )
> 
> func main() {
>     sub.ExportedTestFuncOfSub()
>     sub2.ExportedTestFuncOfSub2()
> 
>     fmt.Println("run main")
> }
> // --- sub, sub2 패키지에 정의해둔 `init` 함수들이 자동으로 실행된다 ---
> // sub.init() called
> // sub2.init() called
> // ExportedTestFuncOfSub
> // ExportedTestFuncOfSub2: 4
> // run main
> ```

### 고루틴 생성

*새로운 고루틴을 시작하는 `go` 문장은 고루틴 실행이 시작되기 전에 동기화된다.*

> **`go` 문장은 무엇과 동기화 된다는 건가?**
>
> 고루틴이 실행에 앞서 `go` 문장 자체의 처리가 완료된다는 것을 의미한다. 즉, 아래와 같은 *고루틴을 시작하기 위한 준비 작업*이 모두 완료하여 고루틴이 실행될 때 필요한 모든 조건이 충족되었음을 보장한다.
>
> - **고루틴을 위한 스택 공간 할당**: 각 고루틴은 독립적인 실행 스택을 가지며, 이는 고루틴이 실행될 때 필요한 메모리 공간을 제공한다.
> - **고루틴 실행을 위한 내부 데이터 구조의 초기화**: 고루틴의 관리와 스케줄링을 위해 필요한 내부 데이터 구조가 초기화된다.
> - **`go` 문장에 의해 호출된 함수 또는 클로저에 대한 참조 설정**: 고루틴이 실행할 실제 코드를 결정한다.
> - **준비가 완료된 고루틴은 Go 런타임의 스케줄러에 등록**: 이를 통해 고루틴이 적절한 시점에 실행될 수 있도록 한다.
>
> 이러한 동기화 메커니즘은 고루틴이 예측 가능한 방식으로 실행되도록 도와주며, 동시에 여러 고루틴이 안전하게 실행될 수 있는 환경을 제공한다.

```go
var a string

func f() {
    print(a)
}

func hello() {
    a = "hello, world"
    go f()
}
```

`hello` 함수를 호출하면 미래 어느 순간에 "hello, world"를 출력할 것이다.(아마도 `hello` 함수가 리턴된 후)

### 고루틴 파괴

고루틴의 종료가 프로그램의 어떤 이벤트 전에 동기화되는 것은 보장되지 않는다.

```go
var a string

func hello() {
    go func() { a = "hello" }()
    print(a)
}
```

`a` 변수에 대한 할당 후 동기화 이벤트가 없으므로, 다른 고루틴의 관찰이 보장되지 않는다. 실제로 공격적인 컴파일러는 `go` 문장 전체를 삭제할 수도 있다.

> **공격적인 컴파일러?**
>
> 별도의 컴파일러가 있는 게 아니라, Go 컴파일러가 실제로 컴파일시에 최적화 위해 삭제하는 경우가 있다. 자세한 내용은 [링크](https://stackoverflow.com/a/47245780) 참고.
>
> 참고 명령어:
> - `go build -gcflags="-N -l -S -m=2"`
> - `go tool compile -S main.go`

고루틴의 결과를 다른 고루틴이 관찰해야 하는 경우, 상대적인 순서를 설정하기 위해 잠금 또는 채널 통신과 같은 동기화 메커니즘을 사용한다.

### 채널 통신

## `append` 함수를 사용하여 슬라이스에 요소를 추가하는 두 방식

Go 언어에서 `append` 함수를 사용하여 슬라이스에 요소를 추가하는 두 방식을 비교하면, 메모리 사용 방식과 관리에 있어서 약간의 차이가 있습니다. 다만, 이 차이는 상황과 사용 패턴에 따라 다를 수 있으며, 어떤 방식이 '더 나은' 방식이라고 일반화하기는 어렵습니다.

### 변수에 결과를 할당한 후 `append`하는 방식

   ```go
   aaaa := pa.getAAAA()
   bbbb = append(qualifiedTransactions, aaaa...)
   ```

이 방식에서는 `pa.getAAAA()` 함수의 결과를 먼저 변수 `aaaa`에 저장한다.
이는 `getAAAA` 함수가 반환하는 슬라이스에 대한 참조를 유지하고, 이후에 `append` 함수를 호출할 때 이 참조를 사용한다.
이 방식의 장점은 `getAAAA` 함수의 결과를 다른 곳에서 재사용할 수 있다는 것이다.
단점은 추가적인 변수를 선언해야 하며, 이 변수가 차지하는 메모리 공간이 있다.

### 직접 `append`하는 방식

   ```go
   bbbb = append(qualifiedTransactions, pa.getAAAA()...)
   ```

이 방식에서는 `pa.getAAAA()` 함수의 결과를 바로 `append` 함수에 전달합니다. 이는 추가적인 변수 할당 없이 직접적으로 함수의 반환값을 사용합니다. 이 방식의 장점은 코드가 더 간결하고, 추가적인 변수에 대한 메모리 할당이 필요 없다는 것입니다. 그러나 `getAAAA` 함수의 결과를 다른 곳에서 재사용할 수 없습니다.

`pa.getAAAA()` 함수가 리턴하는 데이터는 함수 호출 시점에 힙(heap) 메모리에 할당된다.
힙 메모리는 동적 메모리 할당에 사용되며, 함수 호출이 종료되어도 그 데이터는 힙에 남아있게 된다.
이 데이터는 다른 변수에 할당되지 않더라도 힙에 존재하며, 가비지 컬렉터에 의해 관리된다.

### 결론

메모리 관점에서 봤을 때, 두 번째 방식은 추가적인 변수 할당이 필요 없기 때문에 약간 더 효율적일 수 있다.
하지만, 이 차이는 대부분의 경우 매우 미미하며, 실제 메모리 사용량에 큰 영향을 미치지 않습니다.

중요한 것은 코드의 가독성과 유지보수성입니다. 함수의 결과를 재사용할 필요가 없다면 두 번째 방식이 더 간결하고 명확할 수 있습니다. 반면, 함수의 결과를 다시 사용해야 한다면 첫 번째 방식이 더 나을 수 있습니다. 메모리 사용보다는 코드의 목적과 컨텍스트에 따라 적절한 방식을 선택하는 것이 중요합니다.

## 전역 선언과 함수내 선언

### 전역 선언

```go
var currencyScalingFactors = map[int]*big.Int{
    0: new(big.Int).Exp(big.NewInt(10), big.NewInt(0), nil), // 10^0
    1: new(big.Int).Exp(big.NewInt(10), big.NewInt(1), nil), // 10^1
    2: new(big.Int).Exp(big.NewInt(10), big.NewInt(2), nil), // 10^2
    3: new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil), // 10^3
    4: new(big.Int).Exp(big.NewInt(10), big.NewInt(4), nil), // 10^4
}
```

1. **메모리 사용**: 전역 변수로 `currencyScalingFactors`를 선언하면, 프로그램이 실행될 때 단 한 번만 메모리에 할당되고, 프로그램이 종료될 때까지 메모리에 유지됩니다. 이는 해당 변수가 프로그램 전반에 걸쳐 빈번하게 사용될 경우 메모리 사용량을 줄이고 성능을 향상시킬 수 있습니다.

2. **초기화 오버헤드**: 전역 변수는 프로그램 시작 시 초기화되므로, 시작 시간에 약간의 오버헤드가 발생할 수 있습니다. 하지만, 이는 일반적으로 미미한 수준입니다.

3. **공유와 재사용성**: 전역 변수는 프로그램의 다양한 부분에서 공유되므로, 코드 중복을 줄이고 재사용성을 높일 수 있습니다.

### 함수내 선언

```go
func Test(key int) *big.Int{
  return map[int]*big.Int{
      0: new(big.Int).Exp(big.NewInt(10), big.NewInt(0), nil), // 10^0
      1: new(big.Int).Exp(big.NewInt(10), big.NewInt(1), nil), // 10^1
      2: new(big.Int).Exp(big.NewInt(10), big.NewInt(2), nil), // 10^2
      3: new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil), // 10^3
      4: new(big.Int).Exp(big.NewInt(10), big.NewInt(4), nil), // 10^4
    }[key]
}
```

1. **메모리 사용**: 함수 내에서 맵을 선언하고 반환하면, 해당 함수가 호출될 때마다 맵이 새로 생성되고, 함수 호출이 끝나면 가비지 컬렉터에 의해 수집될 수 있습니다. 이 방식은 메모리 사용이 더 동적이지만, 맵을 생성하는 데 드는 비용이 반복적으로 발생합니다.

2. **캡슐화와 독립성**: 함수 내에서만 사용되는 데이터는 해당 함수 내부에 캡슐화하는 것이 좋습니다. 이는 코드의 독립성을 높이고, 전역 상태에 의존하지 않는 순수 함수를 만드는 데 도움이 됩니다.

3. **스레드 안전성**: 멀티 스레드 환경에서 전역 변수는 스레드 간에 공유되므로, 동시성 문제가 발생할 수 있습니다. 함수 내에서 맵을 생성하면, 각 함수 호출이 독립적인 맵 인스턴스를 갖게 되어 스레드 안전성이 자연스럽게 보장됩니다.

### 결론

- **성능과 재사용성이 중요한 경우**에는 전역 변수를 사용하는 것이 좋습니다. 특히, 생성 비용이 높거나 메모리 사용량을 최소화하고 싶은 경우에 유리합니다.

- **코드의 독립성, 캡슐화, 스레드 안전성을 우선시하는 경우**에는 함수 내에서 필요한 데이터를 생성하는 것이 바람직합니다. 특히, 해당 데이터가 특정 함수에서만 사용되고, 함수 호출 사이에 상태를 공유할 필요가 없는 경우에 적합합니다.

각 방식의 장단점을 고려하여, 프로젝트의 요구 사항과 컨텍스트에 가장 잘 맞는 방식을 선택하는 것이 중요합니다.
