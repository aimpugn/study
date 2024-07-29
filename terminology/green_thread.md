# Green Thread

- [Green Thread](#green-thread)
    - [Green Thread?](#green-thread-1)
    - [동작 방식](#동작-방식)
    - [이론적 배경](#이론적-배경)
    - [그린 쓰레드 예시](#그린-쓰레드-예시)
        - [Erlang 예시](#erlang-예시)
        - [Haskell 예시](#haskell-예시)
    - [기타](#기타)

## Green Thread?

- 그린 스레드는 OS의 스레드 관리 기능을 사용하지 않고, 가상 머신이나 런타임 라이브러리에서 직접 구현된 사용자 수준 스레드
- "그린"이라는 명칭의 기원은 명확하지 않지만, 흔히 이러한 스레드를 '가볍고, 환경 친화적'이라는 의미에서 "그린"으로 지칭한다.
- 경량 스레드와 그린 스레드는 기본적으로 유사한 개념을 내포한다
    - 둘 다 전통적인 OS 스레드보다 가볍고,
    - 런타임 레벨에서 관리되어 성능 최적화에 기여한다

## 동작 방식

자바 가상 머신(JVM) 또는 비슷한 런타임 환경에서 스케줄링과 실행을 관리한다.
이는 OS가 아닌 가상 머신 내에서 컨텍스트 스위칭이 이루어지므로, OS의 스레드 관리 메커니즘과는 독립적이다.

## 이론적 배경

Green thread의 개념은 OS에서 제공하는 스레딩 기능이 제한적이거나 성능이 저하되는 경우, 더 효율적인 스레딩 방식을 제공하기 위해 고안되었다.
특히 초기의 Java 가상 머신에서는 OS의 스레드 기능이 충분히 발달하지 않았기 때문에, green thread가 널리 사용되었었다.

하지만 green thread는 멀티코어 프로세서의 이점을 충분히 활용하지 못한다는 단점이 있다.
현대의 시스템에서는 OS의 네이티브 스레드가 더 효율적으로 멀티코어를 활용할 수 있으므로, 대부분의 현대 JVM은 OS의 네이티브 스레드를 사용한다.

## 그린 쓰레드 예시

그린 스레드는 운영 체제가 아닌 런타임 레벨에서 스케줄링되는 스레드를 말합니다. 대표적으로 자바 가상 머신(JVM) 초기 버전이나 Erlang, Haskell 같은 언어의 런타임에서 그린 스레드를 사용했습니다.

하지만 *현대의 대부분의 프로그래밍 언어와 환경은 운영 체제의 네이티브 스레드를 사용하는 방향으로 이동*했기 때문에, 순수 그린 스레드를 사용하는 예제를 찾기 어려울 수 있습니다.

그럼에도 Erlang이나 Haskell 같은 언어에서는 그린 스레드에 해당하는 개념을 볼 수 있습니다. Erlang에서는 경량 프로세스(lightweight process)라고 부르며, 이는 Erlang 런타임에서 관리합니다. Haskell에서는 "스레드"가 기본적으로 그린 스레드로 구현됩니다.

아래 예시들은 그린 스레드가 어떻게 사용될 수 있는지를 보여주지만, 실제로 이러한 스레드를 구현하고 관리하는 것은 프로그래밍 언어의 런타임 시스템에 의해 이루어집니다. 따라서 개발자는 일반적으로 이러한 세부 사항을 직접 관리할 필요가 없으며, 런타임 시스템이 제공하는 API를 사용하여 동시성을 구현합니다.

### Erlang 예시

Erlang에서 간단한 메시지 전송을 수행하는 경량 프로세스 생성 예시입니다.

```erlang
-module(hello).
-export([start/0, say_hello/0]).

start() ->
    Pid = spawn(hello, say_hello, []),
    Pid ! hello.

say_hello() ->
    receive
        hello ->
            io:format("Hello, Erlang!~n")
    end.
```

위 코드에서 `spawn/3` 함수는 새로운 Erlang 프로세스(경량 스레드)를 생성하고, 이 프로세스는 `say_hello/0` 함수를 실행합니다. 메시지를 받아 처리하는 간단한 예제입니다.

### Haskell 예시

Haskell에서 간단한 병렬 실행을 보여주는 예시입니다.

```haskell
import Control.Concurrent

main = do
    forkIO $ putStrLn "Hello, "
    forkIO $ putStrLn "Haskell!"
```

이 코드에서 `forkIO` 함수는 새로운 스레드를 생성해 주어진 작업을 병렬로 실행합니다. Haskell의 스레드는 런타임에 의해 스케줄링되는 그린 스레드입니다.

## 기타

- [Green Thread vs Native Thread](https://perfectacle.github.io/2019/03/10/green-thread-vs-native-thread/)
- [Green Thread vs Kernel Thread](https://velog.io/@choiish98/Green-Thread-vs-Kernel-Thread)
