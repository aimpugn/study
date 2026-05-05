# JVM Java Kotlin Runtime

> 원문 보존형 이동본입니다. 이 파일의 source chunk 본문은 원본 `intervie*.md`에서 그대로 복사되었고, 기술적 보강과 딥 리라이트는 다음 단계에서 수행합니다.

## Source Chunks

<!-- source-chunk: sha256=b6ad5f1248b2d1f9a43415eabf943d73258e8d96b61aad6169a0265e44c0142b topic=jvm-java-kotlin-runtime sources=interview_questions.md:3562-3673, interviews.md:3562-3673 -->

> Source: `interview_questions.md:3562-3673`
> Duplicate source aliases: `interview_questions.md:3562-3673, interviews.md:3562-3673`

## kotlin의 `when` 표현식에서의 exhaustiveness 체크

Kotlin에서 `when` 표현식의 exhaustiveness 체크는 컴파일러가 모든 가능한 경우를 처리하도록 보장합니다.
처리되지 않은 케이스로 인한 런타임 오류를 방지합니다.
이 기능은 주로 sealed 클래스나 enum과 함께 사용되지만, 다른 상황에서도 exhaustiveness를 달성할 수 있습니다.

- Sealed 클래스:

    sealed 클래스는 상속 가능한 클래스를 제한하여, 동일한 파일 내에서만 서브클래스를 정의할 수 있습니다.
    `when` 표현식에서 sealed 클래스를 사용하면 컴파일러는 모든 가능한 서브클래스를 알고 있으므로 exhaustiveness를 체크할 수 있습니다.

    ```kotlin
    sealed class PaymentResult
    class Success : PaymentResult()
    class Failure : PaymentResult()

    fun handleResult(result: PaymentResult) = when (result) {
        is Success -> // 성공 처리
        is Failure -> // 실패 처리
        // 'else' 필요 없음; 컴파일러가 모든 서브클래스를 알고 있음
    }
    ```

- Enums:

    enum은 고정된 상수 집합을 가지므로, `when` 표현식에서 enum을 사용할 때 컴파일러가 모든 케이스를 알고 있어 exhaustiveness를 체크할 수 있습니다.

    ```kotlin
    enum class PaymentMethod { CreditCard, PayPal, BankTransfer }

    fun processPayment(method: PaymentMethod) = when (method) {
        PaymentMethod.CreditCard -> // 신용카드 처리
        PaymentMethod.PayPal -> // 페이팔 처리
        PaymentMethod.BankTransfer -> // 은행 이체 처리
        // 'else' 필요 없음
    }
    ```

Sealed 클래스나 enum을 사용하지 않더라도 컴파일러가 exhaustiveness 체크를 할 수 있는 몇 가지 경우가 있습니다.

- Boolean 조건을 사용하는 경우

    `when`을 주제(subject) 없이 표현식으로 사용하고 모든 가능한 boolean 조건을 다루면,
    컴파일러가 이를 exhaustiveness로 간주할 수 있습니다.

    ```kotlin
    fun categorizeNumber(number: Int) = when {
        number > 0 -> "양수"
        number == 0 -> "영"
        number < 0 -> "음수"
        // 모든 케이스를 다룸
    }
    ```

- 알려진 타입으로 스마트 캐스팅하는 경우

    컴파일러가 스마트 캐스팅을 통해 모든 가능한 타입이 처리되었음을 알 수 있다면, exhaustiveness 체크가 가능합니다.

    ```kotlin
    interface PaymentGateway
    class PayPalGateway : PaymentGateway
    class CreditCardGateway : PaymentGateway
    class BankTransferGateway : PaymentGateway

    fun processGateway(gateway: PaymentGateway) = when (gateway) {
        is PayPalGateway -> // 페이팔 처리
        is CreditCardGateway -> // 신용카드 처리
        is BankTransferGateway -> // 은행 이체 처리
        else -> throw IllegalArgumentException("알 수 없는 게이트웨이")
    }
    ```

    이 경우, `PaymentGateway`가 열린 인터페이스이기 때문에 컴파일러는 exhaustiveness를 보장하지 않으므로 `else`가 필요합니다.

- Companion Object나 Object 선언을 사용하는 경우

    각 케이스를 나타내는 object 선언을 사용하고, 이를 `when` 표현식에서 사용할 때 모든 object를 다루면 exhaustiveness를 달성할 수 있습니다.

    ```kotlin
    object PayPalGateway : PaymentGateway
    object CreditCardGateway : PaymentGateway
    object BankTransferGateway : PaymentGateway

    fun processGateway(gateway: PaymentGateway) = when (gateway) {
        PayPalGateway -> // 페이팔 처리
        CreditCardGateway -> // 신용카드 처리
        BankTransferGateway -> // 은행 이체 처리
        // 'PaymentGateway'가 sealed인 경우 컴파일러가 exhaustiveness를 체크
    }
    ```

    하지만 `PaymentGateway`가 sealed 클래스나 인터페이스가 아니라면 컴파일러는 모든 구현체를 알지 못하므로 exhaustiveness를 보장하지 않습니다.

- Sealed 인터페이스 (Kotlin 1.5 이상)

    Kotlin 1.5부터 sealed 인터페이스를 지원합니다.
    인터페이스를 sealed로 정의하면, 컴파일러가 모든 구현체를 알고 exhaustiveness를 체크할 수 있습니다.

    ```kotlin
    sealed interface PaymentGateway
    object PayPalGateway : PaymentGateway
    object CreditCardGateway : PaymentGateway
    object BankTransferGateway : PaymentGateway

    fun processGateway(gateway: PaymentGateway) = when (gateway) {
        PayPalGateway -> // 페이팔 처리
        CreditCardGateway -> // 신용카드 처리
        BankTransferGateway -> // 은행 이체 처리
        // 'else' 필요 없음; 컴파일러가 exhaustiveness를 보장
    }
    ```

<!-- /source-chunk -->

<!-- source-chunk: sha256=83b85432cf78a549a38f93d84e6930c9985f9e0630c6075330de38a2c337c8d9 topic=jvm-java-kotlin-runtime sources=interview_questions.md:3747-3770, interviews.md:3747-3770 -->

> Source: `interview_questions.md:3747-3770`
> Duplicate source aliases: `interview_questions.md:3747-3770, interviews.md:3747-3770`

## Kotlin Coroutine

Kotlin의 Coroutine은 협업적 멀티 스레딩(Cooperative Multithreading)을 기반으로 합니다.

> 협업적 멀티태스킹 (Cooperative Multitasking)
>
> 협업적 멀티태스킹 모델에서 작업은 스스로 적절한 지점에서 중단하고, 다른 작업이 실행되도록 양보합니다.
> 이 과정에서 커널의 스케줄러 개입 없이도 효율적인 비동기 동작을 수행할 수 있습니다.
>
> Kotlin Coroutines은 협업적 멀티태스킹의 대표적인 예입니다.
> 코루틴은 명시적으로 일시 중단(`suspend`)되고, 필요한 시점에 다시 실행되어 비동기 작업을 처리합니다.

협업적 멀티 스레딩은 스레드가 명시적으로 스케줄링 제어를 양보하는 방식입니다.
이를 통해 제어권을 명시적으로 반환하여 다른 Coroutine이 실행되도록 합니다.
즉, Coroutine은 자신이 직접 제어하여 언제 일시 중단하고 언제 재개할지 결정합니다.

이는 코루틴을 명시적으로 일시 중단하는 `suspend` 키워드나 `yield`와 같은 함수를 통해 이루어집니다.

비교적 적은 리소스를 사용하며, *여러 코루틴이 하나의 스레드에서 순차적으로 실행*됩니다.
Kotlin의 Dispatchers를 통해 다양한 스레드 풀에서 코루틴이 실행되도록 제어할 수 있지만, 기본적으로는 협업적 방식으로 실행됩니다.

- 코루틴은 하나의 스레드에서 여러 코루틴을 관리할 수 있지만, 협업적 모델이므로 코루틴은 강제로 멈춰지지 않으며, 명시적으로 멈춰져야 합니다.
- Kotlin 코루틴은 경량 쓰레드로 작동하며, JVM의 `Thread`와 비교했을 때 훨씬 적은 메모리 오버헤드로 수천 개의 코루틴을 동시에 실행할 수 있습니다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=50235565c4d721960fb1b2483d267199799ea865bdd8edc38ee71bd4d259142d topic=jvm-java-kotlin-runtime sources=interview_questions.md:5287-5319, interviews.md:5287-5319 -->

> Source: `interview_questions.md:5287-5319`
> Duplicate source aliases: `interview_questions.md:5287-5319, interviews.md:5287-5319`

## VM, GC, 그리고 런타임

VM(Virtual Machine)과 GC(Garbage Collection)는 별개의 개념입니다.
GC는 메모리 관리와 관련된 기술이고, VM은 프로그램을 실행하기 위한 가상 환경을 제공합니다.

예를 들어, JVM(Java Virtual Machine)은 Java 프로그램을 바이트코드로 실행하는 가상 환경입니다.
JVM은 반드시 GC를 포함하지는 않지만, Java와 같은 언어는 JVM 내부에 GC가 구현되어 자동으로 메모리를 관리합니다.

언어가 VM을 사용하지 않고도 GC를 사용할 수 있습니다.
예를 들어, Go 언어는 네이티브 코드로 컴파일되어 하드웨어에서 바로 실행되지만, Go 런타임은 자체적으로 GC를 포함해 메모리를 자동으로 관리합니다.
이 경우, VM은 없지만 GC는 존재합니다.

Go는 VM을 사용하지 않고, 네이티브 머신 코드로 컴파일됩니다.
하지만 Go 런타임은 VM과 유사한 기능을 제공하는데, 이는 프로그램 실행 중에 메모리 관리, 스케줄링, GC를 포함합니다.
이 런타임은 Go 프로그램이 안전하고 효율적으로 실행되도록 돕지만, 전통적인 VM처럼 바이트코드를 실행하는 것은 아닙니다.

VM이 없어도 언어 자체에서 메모리 관리를 위해 GC를 사용할 수 있습니다.
VM이 없는 경우, GC는 언어의 런타임 환경에서 구현되어 작동합니다.
Go는 네이티브로 컴파일되어 하드웨어에서 직접 실행되지만, Go 런타임에 GC가 포함되어 있어 메모리 관리를 자동으로 처리합니다.

VM과 런타임은 겹치는 부분이 있지만, 역할과 기능에서 차이가 있습니다.

- VM(Virtual Machine):

    프로그램이 하드웨어와 직접 상호작용하는 대신, 가상화된 하드웨어 환경에서 프로그램을 실행할 수 있게 해주는 소프트웨어입니다.
    예를 들어, Java 프로그램은 JVM이라는 VM에서 실행되어 특정 플랫폼에 독립적인 실행 환경을 제공합니다.

- 런타임(Runtime):

    프로그램이 실행되는 동안 필요한 기능을 제공하는 실행 환경입니다.
    런타임은 메모리 관리, 오류 처리, 스레드 관리 등을 포함하며, 이는 네이티브 코드 실행이거나 VM 위에서 실행될 수도 있습니다.
    Go와 같은 언어는 VM 없이도 런타임을 통해 메모리 관리와 스레드 관리 등을 처리합니다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=9bc80b8dd4af126ddfb4d1a276d414438a416b2f930309548700c6bb286cdb95 topic=jvm-java-kotlin-runtime sources=interview_questions.md:5545-5614, interviews.md:5545-5614 -->

> Source: `interview_questions.md:5545-5614`
> Duplicate source aliases: `interview_questions.md:5545-5614, interviews.md:5545-5614`

## JVM, PHP runtime, Go runtime

1. Java의 실행 환경과 메모리 관리

    Java는 "Write Once, Run Anywhere" 철학을 구현하기 위해 JVM(Java Virtual Machine)을 사용합니다.

    JVM은 Java 바이트코드를 실행하는 가상 머신입니다.
    1. Java 소스코드(.java)가 컴파일되어 바이트코드(.class)로 변환됩니다.
    2. JVM이 이 바이트코드를 로드하고 실행합니다.
    3. JVM은 처음에는 인터프리터 방식으로 바이트코드를 실행합니다.
    4. JIT(Just-In-Time) 컴파일러가 자주 실행되는 코드(핫스팟)를 감지하고 네이티브 코드로 컴파일합니다.

    메모리 관리:
    - JVM은 가비지 컬렉션(GC)을 통해 자동으로 메모리를 관리합니다.
    - 다양한 GC 알고리즘(예: Serial, Parallel, CMS, G1)을 제공합니다.

2. PHP의 실행 환경과 메모리 관리

    PHP는 주로 웹 서버 환경에서 실행되며, 스크립트 언어로 설계되었습니다.
    엄밀히 말하면 VM은 아니지만, Zend Engine이 PHP 코드를 실행하는 역할을 합니다.

    실행 과정은 다음과 같습니다:
    1. PHP 소스코드가 Zend Engine에 의해 파싱되고 컴파일됩니다.
    2. 생성된 opcode가 실행됩니다.
    3. OPcache 확장을 사용하면 컴파일된 opcode를 메모리에 캐시하여 성능을 향상시킵니다.

    메모리 관리:
    - PHP는 참조 카운팅 방식의 가비지 컬렉션을 주로 사용합니다.
    - PHP 5.3 이후로는 순환 참조를 처리하기 위한 Cycle Collector도 도입되었습니다.

3. Go의 실행 환경과 메모리 관리

    Go는 컴파일 언어이지만, 컴파일된 네이티브 코드와 함께 런타임 시스템이 포함됩니다.

    실행 과정은 다음과 같습니다.
    1. Go 소스코드가 컴파일되어 실행 파일로 변환됩니다.
    2. 실행 파일에는 Go 런타임이 포함되어 있습니다.

    메모리 관리:
    - Go 런타임은 동시성 지원과 가비지 컬렉션을 담당합니다.
    - Go의 GC는 동시성을 고려하여 설계되었으며, 짧은 stop-the-world 시간을 가집니다.

Runtime과 VM은 다른 건가요? 다르다면 어떻게 다른가요?
- Runtime: 프로그램 실행에 필요한 라이브러리와 지원 기능들의 집합입니다.
- VM: 가상의 컴퓨터 환경을 소프트웨어로 구현한 것입니다.
- 차이점: VM은 일반적으로 더 복잡하고 완전한 실행 환경을 제공하며, 종종 중간 언어(예: 바이트코드)를 실행합니다. Runtime은 보통 더 가벼우며, 특정 언어나 플랫폼에 특화된 지원 기능을 제공합니다.

VM이 있는 Java 경우 VM이 메모리 할당과 GC를 전적으로 관리합니다.
VM이 없지만 GC는 존재하는 Golang의 경우 런타임 시스템의 일부로 GC가 구현되어 있습니다.
VM과 GC 모두 없는 C++ 경우 개발자가 직접 메모리를 관리하거나, 외부 GC 라이브러리를 사용할 수 있습니다.

1. Apache + mod_php와 nginx + php-fpm의 동작 방식은 어떻게 다른가요?
   - Apache + mod_php:
     1. Apache 웹 서버가 요청을 받습니다.
     2. mod_php가 Apache 프로세스 내에서 직접 PHP 코드를 실행합니다.
     3. 각 Apache 프로세스가 PHP 인터프리터를 포함하게 됩니다.

   - nginx + php-fpm:
     1. nginx 웹 서버가 요청을 받습니다.
     2. PHP 관련 요청을 FastCGI 프로토콜을 통해 php-fpm에 전달합니다.
     3. php-fpm이 별도의 프로세스 풀에서 PHP 코드를 실행합니다.
     4. 실행 결과를 nginx에 반환하고, nginx가 클라이언트에게 응답합니다.

   주요 차이점:
   - 리소스 사용: mod_php는 각 Apache 프로세스에 PHP를 로드하므로 메모리 사용량이 높을 수 있습니다. php-fpm은 필요한 만큼만 PHP 프로세스를 유지하므로 더 효율적일 수 있습니다.
   - 확장성: php-fpm은 PHP 프로세스를 독립적으로 관리할 수 있어 더 나은 확장성을 제공합니다.
   - 보안: php-fpm은 웹 서버와 PHP 실행 환경을 분리하여 추가적인 보안 계층을 제공할 수 있습니다.

이러한 차이점들로 인해 최근에는 nginx + php-fpm 조합이 더 많이 선호되는 추세입니다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=3dd949b9fc91a70e74ea140bff2938b64c3227d11f9b00f4d52dea9ae11f9596 topic=jvm-java-kotlin-runtime sources=interview_questions.md:5647-6498, interviews.md:5647-6498 -->

> Source: `interview_questions.md:5647-6498`
> Duplicate source aliases: `interview_questions.md:5647-6498, interviews.md:5647-6498`

## java 프로그램이 프로세스로 실행되는 과정: 시스템 레벨부터 JVM까지

1. 시스템 레벨 프로세스 실행 메커니즘

    1. 셸의 명령어 처리 시스템

        셸은 사용자의 명령어를 처리하기 위해 복잡한 내부 시스템을 운영합니다.
        `bash -c "java -jar some_boot_app.jar"`가 실행될 때, 다음과 같은 과정이 발생합니다:

        1. 프로세스 컨텍스트 초기화
            - 셸은 먼저 자신의 실행 컨텍스트를 설정합니다.
            - 프로세스 환경(environment)은 다음 구조를 가집니다:

            ```c
            struct process_context {
                char environ;          // 환경 변수 배열
                struct rlimit *rlimits;  // 리소스 제한
                sigset_t sigmask;        // 시그널 마스크
                uid_t real_uid;          // 실제 사용자 ID
                gid_t real_gid;          // 실제 그룹 ID
            };
            ```

        2. Java 실행 파일 경로 해석
            시스템은 다음과 같은 계층적 구조로 Java 실행 파일을 찾습니다:

            ```plaintext
            사용자 입력: java
            ↓
            /usr/bin/java (심볼릭 링크)
            ↓
            /etc/alternatives/java (심볼릭 링크)
            ↓
            /usr/lib/jvm/java-11-openjdk-amd64/bin/java (실제 바이너리)
            ```

            이 과정에서 각 심볼릭 링크는 다음과 같은 메타데이터를 포함합니다:

            ```plaintext
            파일 타입: ELF 64-bit LSB shared object
            아키텍처: x86-64
            동적 링크: 필요 (uses shared libs)
            ASLR: 활성화
            스택 보호: 활성화
            ```

            - ASLR 활성화:

                ASLR이 활성화되어 있어 JVM의 베이스 주소가 매번 달라집니다.
                가령 아래 ELF 구조 분석에서 나오는 Entry point address(0x1040)는 파일 내에서의 상대적인 오프셋입니다.
                ASLR이 적용되는 실제 베이스 주소는 프로세스가 메모리에 로드될 때 결정됩니다.
                실제 실행 시 Entry point의 절대 주소는 `베이스 주소 + 0x1040`이 됩니다.

                ```sh
                # 같은 프로세스를 여러 번 실행했을 때의 베이스 주소
                $ cat /proc/<pid>/maps
                첫 번째 실행: 0x555555554000-...
                두 번째 실행: 0x7f1234567000-...
                세 번째 실행: 0x7ff987654000-...

                # 더 자세히 보려면 readelf 명령어로 확인할 수 있습니다:
                ❯ readelf -a /usr/bin/java | grep -i stack -A 2
                GNU_STACK      0x0000000000000000 0x0000000000000000 0x0000000000000000
                                0x0000000000000000 0x0000000000000000  RW     0x10
                GNU_RELRO      0x0000000000002d40 0x0000000000003d40 0x0000000000003d40
                ```

                - GNU_STACK:
                    프로그램의 실행 스택에 대한 속성을 지정하는 특별한 세그먼트입니다

                    1. 주소값이 전부 0:
                        - 실제 스택 주소는 런타임에 동적으로 할당됨
                        - ASLR에 의해 실행할 때마다 다른 주소에 위치하게 됨
                    2. 권한 플래그 "RW"
                        - `R`(Read): 읽기 가능
                        - `W`(Write): 쓰기 가능
                        - `X` 플래그가 없음: 실행 불가능 = NX(No eXecute) bit 활성화
                            - NX bit가 활성화되어 있다는 것은 스택에서 코드를 실행할 수 없음을 의미
                            - 이는 스택 기반 공격(예: 버퍼 오버플로우를 통한 셸코드 실행)을 방지
                            - 정상적인 Java 프로그램은 스택에서 코드를 실행할 필요가 없음

                - GNU_RELRO

                    Relocation Read-Only의 약자로, GOT(Global Offset Table) 덮어쓰기 공격을 방지하기 위한 또 다른 보안 메커니즘입니다.

            - 스택 보호

                스택 보호는 주로 다음 두 가지 방식으로 작동합니다:

                1. Stack Canary:

                    ```plaintext
                    스택 프레임 구조:
                    [지역 변수들]
                    [Canary 값]  <- 무작위 값
                    [이전 EBP]
                    [리턴 주소]
                    ```

                    - 함수 시작 시 Canary 값을 스택에 저장
                    - 함수 종료 시 Canary 값이 변경되었는지 확인
                    - 변경되었다면 스택 오버플로우 공격 시도로 간주

                    ```c
                    void someFunction() {
                        // 컴파일러가 자동으로 삽입하는 코드
                        long canary = __stack_chk_guard;

                        char buffer[64];
                        // ... 함수 로직 ...

                        // 종료 전 검사
                        if (canary != __stack_chk_guard) {
                            __stack_chk_fail();  // 프로세스 종료
                        }
                    }
                    ```

                2. 실행 방지(NX bit):
                    - 스택 영역을 실행 불가능하게 마킹
                    - 버퍼 오버플로우를 통한 코드 실행 방지
                    - 앞서 본 `LOAD` 세그먼트의 권한 플래그(`r--`, `rw-` 등)로 확인 가능

            ELF(Executable and Linkable Format)는 실행 파일, 객체 파일, 공유 라이브러리, 코어 덤프 등을 위한 표준 파일 포맷입니다.
            Unix 계열 시스템에서 가장 널리 사용되는 실행 파일 포맷입니다.
            - 유연성: 다양한 타입의 파일(실행 파일, 라이브러리 등)을 하나의 포맷으로 표현
            - 확장성: 새로운 섹션이나 세그먼트 추가 가능
            - 플랫폼 독립성: 다양한 프로세서와 아키텍처 지원

    2. ELF 바이너리 구조 분석

        Java 실행 파일은 ELF(Executable and Linkable Format) 형식을 따릅니다.
        다음과 같은 ELF 구조의 특징들은 JVM이 시작될 때 기반이 되는 환경을 제공하며, JVM의 성능과 보안에 직접적인 영향을 미칩니다.
        예를 들어, JIT 컴파일러가 생성하는 네이티브 코드는 이러한 ELF 구조가 제공하는 제약 조건 내에서 동작해야 합니다.

        1. ELF 헤더 구조

            ELF 헤더는 파일의 가장 처음에 위치하며, 파일의 구성과 특성을 정의합니다.

            ```plaintext
            ELF Header:
                Magic:   7f 45 4c 46 02 01 01 00
                Class:                ELF64                         → 64비트 실행 파일임을 나타냄
                Data:                 2's complement, little endian → 데이터 인코딩 방식: 2의 보수를 사용하는 리틀 엔디안
                Version:              1 (current)                   → ELF 규격 버전, 현재는 버전 1만 존재
                OS/ABI:               UNIX - System V               → 목표 운영체제와 ABI(Application Binary Interface)
                ABI Version:          0                             → ABI의 버전 번호
                Type:                 DYN (Shared object file)      → 파일 타입: 동적 링크 라이브러리 또는 PIE(Position Independent Executable)
                Machine:              Advanced Micro Devices X86-64 → 목표 아키텍처: AMD64/Intel 64
                Entry point address:  0x1040                        → 프로그램 실행이 시작될 메모리 주소
            ```

            - Magic
                - `7f`: ELF 파일의 시작을 나타내는 특별한 바이트
                - `45 4c 46`: ASCII로 "ELF"
                - `02`: 64비트 포맷(01은 32비트)
                - `01`: 리틀 엔디안
                - `01`: ELF 버전 1
                - `00`: System V ABI

        2. 주요 세그먼트

            프로그램 헤더는 실행을 위해 메모리에 어떻게 파일을 매핑할지 정의합니다.

            ```plaintext
            프로그램 헤더:
            PHDR     0x0000000000000040 0x0000000000000040 r--
            INTERP   0x0000000000000238 0x0000000000000238 r--
                [/lib64/ld-linux-x86-64.so.2]
            LOAD     0x0000000000000000 0x0000000000000000 r--
            LOAD     0x0000000000001000 0x0000000000001000 r-x
            LOAD     0x0000000000002000 0x0000000000002000 r--
            LOAD     0x0000000000002d68 0x0000000000003d68 rw-
            DYNAMIC  0x0000000000002d78 0x0000000000003d78 rw-
            ```

            Java 실행 파일(예: /usr/bin/java)의 ELF 구조는 일반적인 네이티브 실행 파일과 다른 몇 가지 특징적인 구조를 가집니다:

            - `PHDR` 세그먼트
                - 목적: 프로그램 헤더 테이블 자체를 메모리에 매핑
                - 권한: r--(읽기 전용)
                    > - `r`: read (읽기)
                    > - `w`: write (쓰기)
                    > - `x`: execute (실행)
                    > - `-`: 해당 권한 없음
                - 의미: 런타임에 프로그램 헤더 정보 접근 가능

            - `INTERP` 세그먼트
                - 목적: 동적 링커의 경로 지정
                - 권한: r--(읽기 전용)
                - 내용: /lib64/ld-linux-x86-64.so.2
                - 의미: 이 프로그램이 동적 링킹을 필요로 함을 나타냄

                `INTERP` 세그먼트가 지정한 동적 링커는 JVM 시작 시 필요한 모든 공유 라이브러리를 찾아 매핑합니다

            - `LOAD` 세그먼트들
                - `LOAD r--`: 읽기 전용 데이터. ELF 헤더, 프로그램 헤더 등 실행 파일 메타데이터
                - `LOAD r-x`: 실행 가능한 코드. 실제 프로그램 코드(텍스트 세그먼트)

                    다음 코드들이 실제 JVM을 시작하기 전에 필요한 환경을 구성합니다.
                    - JVM 부트스트래퍼 코드
                    - 클래스 로더 초기화 코드
                    - JNI 인터페이스 초기화
                    - 시그널 핸들러 설정

                - `LOAD r--`: 읽기 전용 데이터. 상수, 문자열 리터럴 등
                - `LOAD rw-`: 읽기/쓰기 데이터. 전역 변수, 정적 변수 등

                    다음과 같은 초기 설정값들이 저장되어 있습니다.
                    - JVM 설정을 위한 전역 변수들
                    - GC 옵션 기본값
                    - 힙 크기 기본값
                    - 스레드 스택 크기 기본값

                `LOAD` 세그먼트들의 메모리 매핑이 완료된 후에야 JVM이 자신의 힙 공간을 할당할 수 있습니다
                PLT/GOT를 통한 동적 심볼 해석으로 인해 JNI 호출 시 약간의 오버헤드가 발생합니다

            - `DYNAMIC` 세그먼트
                - 목적: 동적 링킹 정보 저장
                - 내용: 공유 라이브러리 의존성, 심볼 테이블 등
                - 권한: rw-(읽기/쓰기)

                Java 실행 파일은 매우 많은 공유 라이브러리 의존성을 가집니다.
                다음과 같은 공유 라이브러리들이 `DYNAMIC` 세그먼트에 명시되어 있습니다.
                - libjvm.so (JVM 코어)
                - libjava.so (`네이티브 메소드` 구현)

                    `네이티브 메소드`는 Java Native Interface(JNI)를 통해 호출되는 C/C++ 로 작성된 메소드입니다.

                    ```java
                    public class Example {
                        // native 키워드로 선언된 메소드
                        private native void someNativeMethod();

                        static {
                            // 네이티브 라이브러리 로드
                            System.loadLibrary("example");
                        }
                    }
                    ```

                - libverify.so (클래스 파일 검증)
                - libzip.so (JAR 파일 처리)
                - libnio.so (네이티브 I/O)

            - 특별한 섹션들:
                - `.debug_java`: JVM 디버깅을 위한 심볼 정보
                - `.note.jvm`: JVM 버전 및 구현 정보
                - `.rodata.hotspot`: HotSpot JVM 관련 상수 데이터

    3. 메모리 매핑 프로세스

        ELF 파일이 실행될 때 다음과 같은 순서로 메모리에 매핑됩니다:

        1. 초기 매핑

            ```plaintext
            Virtual Address Space
            +----------------------+ 높은 주소
            |      Stack           |
            +----------------------+
            |        ↓             |
            +----------------------+
            |                      |
            +----------------------+
            |                      |
            +----------------------+
            |        ↑             |
            +----------------------+
            |      Heap            |
            +----------------------+
            |   LOAD (rw-)         | → .data, .bss
            +----------------------+
            |   LOAD (r--)         | → .rodata
            +----------------------+
            |   LOAD (r-x)         | → .text
            +----------------------+
            |   LOAD (r--)         | → ELF 헤더 등
            +----------------------+ 낮은 주소
            ```

        2. 동적 링커 로드
            - `INTERP` 세그먼트가 지정한 동적 링커 로드
            - 공유 라이브러리 의존성 해석
            - 필요한 라이브러리 매핑

        3. 재배치(Relocation)
            - 심볼 주소 해석
            - 코드/데이터 참조 수정
            - PLT/GOT 테이블 설정

2. 프로세스 생성 및 메모리 초기화

    1. Fork-Exec 메커니즘:

        운영체제는 새로운 프로세스를 생성하기 위해 Fork-Exec 패턴을 사용합니다.
        이 과정은 다음과 같은 단계로 이루어집니다:

        1. 프로세스 복제 (Fork)
            - 현재 프로세스의 전체 메모리 공간이 복사됩니다
            - Copy-on-Write(CoW) 메커니즘이 사용됩니다
            - 파일 디스크립터가 복제됩니다

        2. 새 프로그램 로딩 (Exec)
            커널은 다음과 같은 작업을 수행합니다:

            ```c
            // 핵심 시스템 콜 구조
            int execve(const char *pathname, char *const argv[], char *const envp[]);
            ```

            실제 실행 시 다음 단계가 수행됩니다:
            1. 현재 프로세스의 메모리 정리
            2. ELF 파일 검증
            3. 프로그램 헤더 파싱
            4. 메모리 세그먼트 매핑
            5. 동적 링커 초기화

    2. 메모리 레이아웃 변환

        프로세스의 메모리 레이아웃은 다음과 같이 변환됩니다:

        ```plaintext
        변환 전 (bash):              변환 후 (JVM):
        +------------------+        +------------------+
        | Kernel Space     |        | Kernel Space     |
        +------------------+        +------------------+
        | Stack            |        | Stack            |
        |                  |        | (Thread Stacks)  |
        +------------------+        +------------------+
        | Shared Libraries |  →     | Shared Libraries |
        |                  |        | (JVM + Native)   |
        +------------------+        +------------------+
        | Heap             |        | Heap             |
        |                  |        | (Java Heap)      |
        +------------------+        +------------------+
        | Data/BSS         |        | Data/BSS         |
        +------------------+        +------------------+
        | Text             |        | Text             |
        +------------------+        +------------------+
        ```

3. JVM 초기화 및 구성

    1. JVM 프로세스 초기화:

        JVM이 시작되면 다음과 같은 초기화 단계를 거칩니다.

        1. JVM 코어 컴포넌트 초기화
            1. 메모리 서브시스템
                - 힙 영역 할당
                - 가비지 컬렉터 초기화
                - 메타스페이스 설정

            2. 실행 엔진
                - 인터프리터 초기화
                - JIT 컴파일러 준비
                - 최적화 시스템 설정

            3. 런타임 데이터 영역
                - 메서드 영역 설정
                - 힙 영역 구성
                - 스레드 로컬 영역 준비

        2. 클래스로더 시스템 초기화:

            JVM은 계층적 클래스로더 시스템을 구성합니다.

            ```plaintext
            Bootstrap ClassLoader (네이티브)
            ↓
            Extension ClassLoader (Java)
            ↓
            System ClassLoader (Java)
            ↓
            Custom ClassLoaders (필요시)
            ```

    2. 메모리 관리 시스템

        JVM의 메모리 관리 시스템은 다음과 같은 구조를 가집니다:

        1. 힙 구조

            ```plaintext
            힙 메모리 구조:
            +----------------------------------------+
            |                                        |
            |             Old Generation             |
            |                                        |
            +----------------------------------------+
            |          |          |                  |
            | Eden     | S0       | S1               |
            | Space    | Space    | Space            |
            |          |          |                  |
            +----------------------------------------+
            Young Generation
            ```

        2. 가비지 컬렉션 알고리즘
            - Minor GC (Young Generation)
            - Major GC (Old Generation)
            - Full GC (전체 힙)

            각 GC는 다음과 같은 기본 단계를 가집니다:

            1. Marking Phase
                - 살아있는 객체 식별
                - 참조 그래프 순회
                - 도달 가능성 분석

            2. Sweeping Phase
                - 죽은 객체 식별
                - 메모리 회수
                - 프리 리스트 업데이트

            3. Compacting Phase (필요시)
                - 살아있는 객체 재배치
                - 메모리 단편화 제거
                - 포인터 업데이트

4. JAR 파일 실행 및 클래스 로딩

    1. JAR 파일 구조 분석:
        Spring Boot JAR 파일은 다음과 같은 특별한 구조를 가집니다.

        ```plaintext
        backend-0.0.1-SNAPSHOT.jar
        .
        ├── org/springframework/boot/loader
        │   ├── ref
        │   ├── net
        │   │   ├── util
        │   │   └── protocol
        │   │       ├── jar
        │   │       └── nested
        │   ├── jarmode
        │   ├── launch
        │   ├── jar
        │   ├── zip
        │   ├── nio
        │   │   └── file
        │   └── log
        ├── META-INF
        │   ├── MANIFEST.MF
        │   └── services
        │       └── java.nio.file.spi.FileSystemProvider
        └── BOOT-INF
            ├── classes
            │   ├── META-INF
            │   │   └── backend.kotlin_module
            │   ├── me
            │   │   └── aimpugn
            │   │       └── backend
            │   │           ├── controller
            │   │           │   └── HomeController.class
            │   │           ├── BackendApplication.class
            │   │           └── BackendApplicationKt.class
            │   └── application.properties
            ├── layers.idx
            ├── classpath.idx
            └── lib
        ```

        주요 컴포넌트:
        1. MANIFEST.MF: 실행 정보 포함
        2. BOOT-INF/classes: 컴파일된 애플리케이션 클래스
        3. BOOT-INF/lib: 의존성 라이브러리
        4. Spring Boot Loader: 특수 클래스로더 및 실행 메커니즘

    2. 클래스 로딩 메커니즘

        Spring Boot의 클래스 로딩은 다음 단계로 진행됩니다:

        1. `LaunchedURLClassLoader` 초기화
            - JAR 파일 내부 구조 분석
            - 클래스패스 구성
            - 리소스 로딩 전략 설정

        2. 클래스 로딩 순서
            1. Bootstrap Classes (JVM 코어)
            2. Extension Classes (JDK 확장)
            3. Application Classes (BOOT-INF/classes)
            4. Dependency Classes (BOOT-INF/lib)

5. JIT 컴파일과 최적화

    1. JIT 컴파일 프로세스

        JIT 컴파일러는 다음과 같은 단계로 작동합니다:

        1. 프로파일링 단계
            1. 메서드 호출 빈도 측정
            2. 루프 실행 횟수 추적
            3. 브랜치 예측 데이터 수집
            4. 타입 정보 수집

        2. 최적화 단계

        3. 인라이닝
            - 메서드 크기 분석
            - 호출 빈도 확인
            - 컨텍스트 특화 복제

        4. 루프 최적화
            - 루프 언롤링
            - 벡터화
            - 범위 체크 제거

        5. 탈출 분석
            - 객체 할당 제거
            - 락 제거
            - 스택 할당

    2. 코드 캐시 관리

        JIT 컴파일된 코드는 다음과 같이 관리됩니다:

        ```plaintext
        코드 캐시 구조:
        +--------------------------------+
        | Non-profiled Code              |
        | (일반 컴파일 코드)                 |
        +--------------------------------+
        | Profiled Code                  |
        | (프로파일링 데이터 포함)             |
        +--------------------------------+
        | Non-method Code                |
        | (스텁, 어댑터 등)                 |
        +--------------------------------+
        ```

6. 스레드 관리 및 동기화
    1. 스레드 모델

        JVM의 스레드 시스템은 다음과 같은 구조를 가집니다:
        1. 스레드 타입
            1. VM 스레드
                - GC 작업 수행
                - JIT 컴파일 수행
                - 내부 최적화 작업

            2. Java 스레드
                - 사용자 코드 실행
                - 애플리케이션 로직 처리
                - 동기화 작업 수행

            3. Compiler 스레드
                - JIT 컴파일 수행
                - 코드 최적화
                - 프로파일링 데이터 수집

        2. 스레드 상태 전이

            ```plaintext
            NEW → RUNNABLE → BLOCKED ↔ WAITING ↔ TIMED_WAITING → TERMINATED
            ```

    2. 동기화 메커니즘

        JVM은 다음과 같은 동기화 시스템을 제공합니다:
        1. 모니터 구현
            1. Thin Lock
                - 단순 동기화
                - 스핀락 사용
                - 빠른 획득/해제

            2. Fat Lock
                - 복잡한 동기화
                - OS 뮤텍스 사용
                - 스레드 대기열 관리

        2. Biased Locking
            - 단일 스레드 최적화
            - 락 획득/해제 오버헤드 제거
            - 자동 리바이어스

7. 메모리 관리 상세 분석

    1. 세대별 가비지 컬렉션 구현

        JVM의 가비지 컬렉션은 "Generational Hypothesis"에 기반하여 설계되었습니다.
        이는 다음과 같은 특성을 가집니다:

        1. Young Generation 관리

            Eden Space 할당 프로세스:
            1. TLAB (Thread Local Allocation Buffer) 할당
                - 스레드별 독립 할당 영역
                - 동기화 오버헤드 감소
                - 일반적으로 Eden의 약 1% 크기

            2. 객체 이동 경로:
                Eden → Survivor 0/1 → Old Generation

            3. 에이징(Aging) 메커니즘:
                - 객체당 나이 카운터 유지
                - Survivor 공간 이동시 증가
                - 임계값(기본 15) 도달시 승격

        2. 메모리 할당 최적화

            ```plaintext
            Fast Path 할당:
            +------------------+
            | TLAB 할당 시도     |
            +------------------+
                    ↓ (실패시)
            +------------------+
            | Eden 직접 할당     |
            +------------------+
                    ↓ (실패시)
            +------------------+
            | Slow Path 할당    |
            +------------------+
            ```

            크기별 할당 전략:
            - 작은 객체 (<128KB): TLAB/Eden
            - 중간 객체 (128KB-2MB): Eden 직접
            - 큰 객체 (>2MB): Old Generation 직접

    2. 메타스페이스 관리

        Java 8 이후 도입된 메타스페이스는 다음과 같은 특성을 가집니다:

        ```plaintext
        메타스페이스 구조:
        +--------------------------------+
        | Klass Metaspace                |
        | - 클래스 메타데이터                |
        | - 메서드 메타데이터                |
        | - 상수 풀                        |
        +--------------------------------+
        | Non-Klass Metaspace            |
        | - 런타임 상수                     |
        | - 메서드 데이터                   |
        | - 기타 메타데이터                  |
        +--------------------------------+
        ```

        메모리 할당 단위:
        - Metachunk: 기본 할당 단위
        - Metablock: Chunk 내부 할당 단위
        - Metacache: Chunk 관리 캐시

8. JIT 컴파일러 상세 분석

    1. 컴파일 단계별 최적화

        JIT 컴파일러는 다음과 같은 단계적 최적화를 수행합니다:
        1. IR(Intermediate Representation) 생성

            ```plaintext
            바이트코드 → HIR → LIR → 머신 코드
            ```

            - HIR (High-level IR) 최적화:
                - 메서드 인라이닝
                - 루프 최적화
                - 탈출 분석
                - 타입 특화

            - LIR (Low-level IR) 최적화:
                - 레지스터 할당
                - 명령어 스케줄링
                - 피프홀 최적화

        2. 컴파일 티어 시스템

            ```plaintext
            Tier 0: 인터프리터
            ↓
            Tier 1: C1 컴파일러 (최적화 없음)
            ↓
            Tier 2: C1 컴파일러 (제한적 최적화)
            ↓
            Tier 3: C1 컴파일러 (전체 최적화)
            ↓
            Tier 4: C2 컴파일러 (서버 컴파일러)
            ```

            각 티어별 특성:
            - 컴파일 시간
            - 최적화 수준
            - 메모리 사용량
            - 코드 품질

    2. OSR (On-Stack Replacement)

        실행 중인 메서드의 최적화 버전 교체 메커니즘:

        - OSR 트리거 조건:
            1. 루프 카운터 임계값 초과
            2. 메서드 호출 횟수 임계값 초과
            3. 백엣지 카운터 임계값 초과

        - 교체 프로세스:
            1. 현재 프레임 상태 캡처
            2. 최적화 코드 생성
            3. 상태 전이 코드 생성
            4. 프레임 상태 복원

9. Spring Boot 애플리케이션 실행 프로세스

    1. 스프링 부트 로더 메커니즘

        Spring Boot JAR의 실행은 다음 단계로 이루어집니다:

        1. JarFile URL 처리

            ```plaintext
            archive:jar:file:/app.jar!/BOOT-INF/classes!/
            ↓
            JarFileUrlConnection
            ↓
            NestedJarFile
            ↓
            실제 클래스/리소스 로딩
            ```

        2. 클래스로더 계층

            ```plaintext
            LaunchedURLClassLoader
            ↓
            ExtClassLoader
            ↓
            Bootstrap ClassLoader
            ```

            검색 순서:
            1. BOOT-INF/classes/
            2. BOOT-INF/lib/*.jar
            3. JVM 시스템 클래스

    2. 스프링 컨텍스트 초기화

        애플리케이션 컨텍스트 초기화 프로세스:

        1. 환경 준비
            1. 시스템 속성 로드
            2. OS 환경 변수 로드
            3. 애플리케이션 속성 파일 로드
            4. 활성 프로파일 결정

        2. 빈 생성 주기

            ```plaintext
            생성 단계:
            Constructor → @PostConstruct → InitializingBean
                                                ↓
            소멸 단계:    DisposableBean ← @PreDestroy
            ```

10. 네이티브 인터페이스 (JNI) 통합

    1. JNI 구조

        JVM과 네이티브 코드의 통합:

        ```plaintext
        JNI 호출 스택:
        Java 메서드
        ↓
        JNI 스텁
        ↓
        네이티브 메서드 스텁
        ↓
        네이티브 코드
        ```

        JNI 참조 관리:
        - Local References
        - Global References
        - Weak Global References

    2. 네이티브 메모리 관리

        네이티브 메모리 영역의 구성:

        ```plaintext
        네이티브 메모리 레이아웃:
        +---------------------------+
        | 코드 캐시                   |
        +---------------------------+
        | 네이티브 힙                  |
        +---------------------------+
        | 스레드 스택                  |
        +---------------------------+
        | 컴파일러 힙                  |
        +---------------------------+
        | 직접 버퍼                   |
        +---------------------------+
        ```

11. 성능 모니터링 및 프로파일링

    1. JVM은 다양한 모니터링 인터페이스를 제공합니다.
        1. JMX (Java Management Extensions): 모니터링 메트릭
            - 힙 메모리 사용량
            - 스레드 상태
            - 클래스 로딩 통계
            - GC 통계
            - JIT 컴파일 통계

        2. Flight Recorder 데이터

            이벤트 타입:
            - JVM 내부 이벤트
            - GC 이벤트
            - 컴파일 이벤트
            - 스레드 이벤트
            - I/O 이벤트

    2. JVM 제공 성능 분석 도구:
        - `jstat`: GC 통계
        - `jmap`: 힙 덤프
        - `jstack`: 스레드 덤프
        - `jcmd`: 진단 명령
        - `jinfo`: 설정 조회

12. 장애 처리 및 디버깅

    1. 예외 처리 메커니즘

        JVM의 예외 처리 구조:

        ```plaintext
        예외 처리 스택:
        Java 예외
        ↓
        JVM 시그널 핸들러
        ↓
        OS 시그널
        ```

        예외 테이블 구조:
        - try 범위
        - catch 타입
        - 핸들러 위치
        - 스택 언와인딩 정보

    2. 크래시 덤프 분석

        시스템 장애 시 생성되는 덤프 컨텐츠:
        1. 스레드 상태
        2. 메모리 맵
        3. 로드된 라이브러리
        4. 시스템 정보
        5. GC 상태

<!-- /source-chunk -->

<!-- source-chunk: sha256=60694b1d89a84a70b64b4693327d5525eba0b2ee177bd81d88a27d0ca242cc08 topic=jvm-java-kotlin-runtime sources=interview_questions.md:9461-9751, interviews.md:9409-9699 -->

> Source: `interview_questions.md:9461-9751`
> Duplicate source aliases: `interview_questions.md:9461-9751, interviews.md:9409-9699`

## java, php에서 mock 생성 원리

Java나 PHP에서 사용되는 Mock 객체는 주로 리플렉션(reflection)과 프록시(proxy) 패턴을 사용하여 동적으로 생성됩니다.
Mocking 프레임워크들은 이러한 기술을 이용해 런타임 시점에 가짜 객체를 생성하고, 특정 메서드 호출 시 지정된 동작을 수행하도록 만들어줍니다.

- 리플렉션 (Reflection)

    리플렉션은 프로그램이 실행 중에 자신의 구조에 대해 알 수 있게 해주고, 런타임에 객체의 메서드나 필드에 접근하고 수정할 수 있는 기능을 제공합니다.
    Java나 PHP의 Mocking 프레임워크는 리플렉션을 통해 객체의 메서드와 필드를 동적으로 조작하고, 호출된 메서드의 정보를 추적할 수 있습니다.

    ```java
    /*
     * 리플렉션을 사용하여 메서드를 동적으로 호출하는 예제
     *
     * 이 메서드는 주어진 객체의 메서드를 이름으로 찾아 호출합니다.
     * 리플렉션을 사용하여 런타임에 메서드의 존재를 확인하고 호출할 수 있습니다.
     *
     * @param obj 메서드를 호출할 객체
     * @param methodName 호출할 메서드의 이름
     * @param args 메서드에 전달할 인자들
     * @return 메서드 호출의 결과
     * @throws Exception 메서드를 찾지 못하거나 호출 중 예외가 발생한 경우
     */
    public static Object invokeMethod(Object obj, String methodName, Object... args) throws Exception {
        // 객체의 클래스 정보를 가져옵니다.
        Class<?> clazz = obj.getClass();

        // 인자들의 클래스 타입을 저장할 배열을 생성합니다.
        Class<?>[] parameterTypes = new Class<?>[args.length];
        for (int i = 0; i < args.length; i++) {
            parameterTypes[i] = args[i].getClass();
        }

        // 지정된 이름과 파라미터 타입을 가진 메서드를 찾습니다.
        Method method = clazz.getMethod(methodName, parameterTypes);

        // 메서드를 호출하고 결과를 반환합니다.
        return method.invoke(obj, args);
    }

    public static void main(String[] args) {
        String str = "Hello, World!";
        try {
            // length() 메서드를 동적으로 호출합니다.
            int length = (int) invokeMethod(str, "length");
            System.out.println("String length: " + length);

            // charAt(int) 메서드를 동적으로 호출합니다.
            char ch = (char) invokeMethod(str, "charAt", 7);
            System.out.println("Character at index 7: " + ch);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
    ```

    이 예제는 리플렉션을 사용하여 메서드를 동적으로 호출하는 방법을 보여줍니다.
    Mock 객체 생성 시 이와 유사한 기술이 사용되어 메서드 호출을 가로채고 사용자 정의 동작을 주입할 수 있습니다.

- 프록시 (Proxy) 패턴

    프록시 패턴은 대리 객체를 생성하여 실제 객체와 상호작용할 때 그 사이에서 중간 역할을 수행하는 방식입니다.
    Mock 객체는 프록시 패턴을 이용하여 실제 객체처럼 보이지만, 실제로는 모킹된 동작을 수행합니다.
    프록시 객체는 실제 객체의 메서드를 호출하기 전에 추가적인 로직을 실행하거나, 아예 다른 동작을 수행하도록 설정될 수 있습니다.

    Java의 `java.lang.reflect.Proxy` 클래스는 인터페이스 기반의 프록시 객체를 동적으로 생성할 수 있도록 해줍니다.

    ```java
    import java.lang.reflect.InvocationHandler;
    import java.lang.reflect.Method;
    import java.lang.reflect.Proxy;

    /*
     * 동적 프록시를 사용하여 인터페이스의 메서드 호출을 가로채는 예제
     *
     * 이 예제는 MyInterface의 모든 메서드 호출을 가로채고 로깅하는 프록시를 생성합니다.
     * 동적 프록시는 런타임에 인터페이스를 구현하는 클래스를 생성하여 메서드 호출을 중개합니다.
     */

    // 프록시할 인터페이스
    interface MyInterface {
        void doSomething();
        String getSomething();
    }

    // 실제 구현 클래스
    class MyImplementation implements MyInterface {
        @Override
        public void doSomething() {
            System.out.println("Doing something");
        }

        @Override
        public String getSomething() {
            return "Something";
        }
    }

    // InvocationHandler 구현
    class LoggingInvocationHandler implements InvocationHandler {
        private final Object target;

        public LoggingInvocationHandler(Object target) {
            this.target = target;
        }

        @Override
        public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
            System.out.println("Before method: " + method.getName());
            Object result = method.invoke(target, args);
            System.out.println("After method: " + method.getName() + ", result: " + result);
            return result;
        }
    }

    public class DynamicProxyExample {
        public static void main(String[] args) {
            MyInterface realObject = new MyImplementation();

            MyInterface proxyObject = (MyInterface) Proxy.newProxyInstance(
                MyInterface.class.getClassLoader(),
                new Class<?>[] { MyInterface.class },
                new LoggingInvocationHandler(realObject)
            );

            // 프록시 객체를 통한 메서드 호출
            proxyObject.doSomething();
            String result = proxyObject.getSomething();
            System.out.println("Main got result: " + result);
        }
    }
    ```

    이 예제에서 `LoggingInvocationHandler`는 모든 메서드 호출을 가로채고 로깅합니다.
    이는 Mock 객체가 메서드 호출을 가로채고 사용자 정의 동작을 수행하는 방식과 유사합니다.

- 바이트코드 조작

    바이트코드 조작은 컴파일된 클래스의 바이트코드를 직접 수정하는 기술입니다.
    이 방법은 더 강력하고 유연하지만, 복잡성도 높습니다.
    Java에서는 ASM, javassist 같은 라이브러리가 이 기능을 제공합니다.

    ```java
    import javassist.*;

    /*
     * Javassist를 사용하여 런타임에 클래스를 수정하는 예제
     *
     * MyClass의 doSomething 메서드를 런타임에 수정하여
     * 메서드 실행 전후에 로깅을 추가합니다.
     */
    public class BytecodeManipulationExample {
        public static void main(String[] args) throws Exception {
            // 클래스 풀 생성
            ClassPool pool = ClassPool.getDefault();

            // 수정할 클래스 가져오기
            CtClass cc = pool.get("MyClass");

            // 수정할 메서드 가져오기
            CtMethod m = cc.getDeclaredMethod("doSomething");

            // 메서드 내용 수정
            m.insertBefore("System.out.println(\"Before doSomething\");");
            m.insertAfter("System.out.println(\"After doSomething\");");

            // 수정된 클래스를 새 이름으로 저장
            cc.setName("MyClass_Modified");
            cc.writeFile();

            // 수정된 클래스 로드 및 인스턴스 생성
            Class<?> modifiedClass = cc.toClass();
            Object obj = modifiedClass.newInstance();

            // 수정된 메서드 호출
            modifiedClass.getMethod("doSomething").invoke(obj);
        }
    }

    // 원본 클래스
    class MyClass {
        public void doSomething() {
            System.out.println("Doing something");
        }
    }
    ```

- Mocking 프레임워크의 내부 동작 원리

    Mocking 프레임워크들은 리플렉션과 프록시를 사용하여 가짜 객체를 동적으로 생성하고, 다음과 같은 방식으로 동작합니다:

    1. 클래스 분석:
        Mocking 프레임워크는 모킹할 클래스나 인터페이스의 메서드와 필드를 리플렉션을 사용해 분석합니다.

    2. 프록시 생성:

        Mocking 프레임워크는 프록시 객체를 생성하여 실제 클래스처럼 동작하도록 만듭니다.
        Java에서는 `Proxy.newProxyInstance()` 메서드를 통해 동적으로 프록시를 생성하고,
        PHP에서는 `__call()` 메서드를 오버라이드하는 방식으로 비슷한 기능을 제공합니다.

    3. 메서드 호출 가로채기:

        프록시 객체의 메서드가 호출되면 `InvocationHandler`나 `MethodInterceptor`와 같은 핸들러가 호출을 가로채고, 미리 정의된 동작(예: 지정된 값 반환, 예외 발생 등)을 수행합니다.
        이 과정에서 호출 횟수나 인자 등을 기록하여 나중에 검증할 수 있습니다.

    4. 결과 반환 또는 예외 발생:

        모킹된 메서드는 사전에 설정된 동작을 수행하고, 그 결과를 반환하거나 특정 예외를 발생시킵니다.
        이를 통해 실제 객체의 복잡한 로직이나 외부 의존성 없이도 테스트를 진행할 수 있습니다.

Java와 PHP에서의 Mocking 프레임워크

- Java - Mockito

    Java에서 널리 사용되는 Mockito는 프록시와 리플렉션을 결합하여 Mock 객체를 생성합니다.
    Mockito는 다음과 같은 방식으로 작동합니다:

    - `Mockito.mock()`: 특정 클래스나 인터페이스의 Mock 객체를 생성합니다.
    - `when().thenReturn()`: 특정 메서드 호출 시 반환될 값을 지정합니다.
    - `verify()`: 메서드가 호출되었는지, 특정 인자와 함께 호출되었는지 등을 검증할 수 있습니다.

    ```java
    // Mockito 사용 예시
    import static org.mockito.Mockito.*;

    public class MockitoExample {
        public static void main(String[] args) {
            // List 인터페이스의 mock 객체 생성
            List<String> mockedList = mock(List.class);

            // mock 객체의 동작 정의
            when(mockedList.get(0)).thenReturn("first");
            when(mockedList.get(1)).thenThrow(new RuntimeException());

            // mock 객체 사용
            System.out.println(mockedList.get(0)); // 출력: first
            try {
                mockedList.get(1); // RuntimeException 발생
            } catch (RuntimeException e) {
                System.out.println("Exception caught as expected");
            }

            // 메서드 호출 검증
            verify(mockedList).get(0);
            verify(mockedList).get(1);
            verify(mockedList, never()).clear();

            // Mock 객체 생성
            UserRepository mockRepository = mock(UserRepository.class);

            // 특정 메서드의 반환값 설정
            when(mockRepository.findUserById(1)).thenReturn(new User(1, "Test User"));

            // Mock 객체 사용
            User user = mockRepository.findUserById(1);
            System.out.println(user.getName()); // "Test User" 출력

            // 메서드 호출 검증
            verify(mockRepository).findUserById(1);
        }
    }
    ```

- PHP - PHPUnit

    PHP에서는 PHPUnit의 Mock 기능을 통해 가짜 객체를 생성할 수 있습니다.
    PHPUnit은 프록시 객체를 만들어 인터페이스나 클래스의 메서드를 모킹할 수 있도록 도와줍니다.
    PHP는 동적으로 메서드를 호출할 수 있는 `__call()` 메서드와 리플렉션을 활용하여 비슷한 원리로 작동합니다.

    ```php
    // PHPUnit 사용 예시
    use PHPUnit\Framework\TestCase;

    class UserServiceTest extends TestCase {
        public function testGetUserName() {
            // Mock 객체 생성
            $mockRepository = $this->createMock(UserRepository::class);

            // 특정 메서드의 반환값 설정
            $mockRepository->method('findUserById')
                        ->willReturn(new User(1, 'Test User'));

            // Mock 객체 사용
            $userService = new UserService($mockRepository);
            $this->assertEquals('Test User', $userService->getUserName(1));
        }
    }
    ```

    이 코드에서는 `createMock()` 메서드를 사용해 `UserRepository` 인터페이스의 Mock 객체를 생성하고, `findUserById` 메서드가 호출될 때 `User` 객체를 반환하도록 설정합니다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=9c428c3213103c7d65c1545975794a9d752dc7f0745326de57513d415813574d topic=jvm-java-kotlin-runtime sources=interview_questions.md:10148-10266, interviews.md:10096-10214 -->

> Source: `interview_questions.md:10148-10266`
> Duplicate source aliases: `interview_questions.md:10148-10266, interviews.md:10096-10214`

## 부모 위임 모델

Java의 클래스 로딩(Class Loading) 모델 중 하나로, 클래스 로더(ClassLoader)들이 계층적인 구조를 가지며, 클래스를 로드할 때 먼저 부모 클래스 로더에게 위임(delegate)하는 방식을 따릅니다.

이 방식은 Java의 기본적인 보안 모델과 클래스 충돌 방지를 위해 설계된 구조입니다.

> 💡 핵심 원칙:
> "클래스를 로드하려면, 먼저 부모 ClassLoader에게 위임하고, 부모가 해당 클래스를 찾지 못할 때만 자신이 직접 로드한다."

이 모델을 활용하면, JVM의 기본 클래스(`java.lang.*`, `java.util.*` 등) 및 공유 클래스가 일관되게 유지되며, 서로 다른 클래스 로더 간의 중복 로딩 및 충돌을 방지할 수 있습니다.

- 부모 위임 모델의 동작 방식

    🔹 클래스 로딩의 기본 흐름
    1️⃣ 애플리케이션에서 특정 클래스를 로드하려고 할 때,
    2️⃣ 해당 요청을 가장 상위 부모 클래스 로더(Bootstrap ClassLoader)에게 위임
    3️⃣ 부모 클래스 로더가 해당 클래스를 찾으면 반환, 찾을 수 없으면 자식(ClassLoader)에게 요청을 넘김
    4️⃣ 최종적으로 해당 클래스를 로드할 수 있는 클래스 로더가 직접 로드하여 반환

    이러한 재귀적 방식으로 클래스를 탐색하며, 부모가 먼저 클래스를 찾을 기회를 가지므로, Java의 핵심 클래스(`java.lang.String` 등)가 항상 표준 클래스 로더에서 먼저 로드되는 구조를 유지할 수 있습니다.

- Tomcat의 ClassLoader 계층 구조

    Tomcat의 `ClassLoader`는 부모 위임 모델을 따르며, 다음과 같은 계층을 가집니다.

    | ClassLoader | 설명 | 클래스 로딩 경로 |
    |---------------|-----------------------------------|-------------------------------|
    | Bootstrap ClassLoader | JDK의 핵심 클래스를 로드 | `$JAVA_HOME/lib/rt.jar` (Java 8 이하) |
    | System ClassLoader | `java.ext.dirs`에 위치한 라이브러리 로드 | `$JAVA_HOME/lib/ext` |
    | Common ClassLoader | Tomcat 내부 및 공유 클래스 로드 | `$CATALINA_HOME/lib` |
    | Catalina ClassLoader | Tomcat의 핵심 실행 코드 로드 | `$CATALINA_HOME/lib/catalina.jar` |
    | Shared ClassLoader | 모든 애플리케이션이 공유하는 라이브러리 로드 | `$CATALINA_HOME/shared/lib` |
    | Webapp ClassLoader | 각 웹 애플리케이션별 클래스 로드 | `WEB-INF/classes` 및 `WEB-INF/lib/*.jar` |

    📌 동작 예시:
    - `java.lang.String`을 로드할 경우:
        → Bootstrap ClassLoader에서 로드 (JVM 핵심 클래스이므로 최상위에서 관리)
    - `javax.servlet.http.HttpServlet`을 로드할 경우:
        → Common ClassLoader에서 로드 (Tomcat의 공유 클래스이므로 공통 로드)
    - `com.example.MyController`를 로드할 경우:
        → Webapp ClassLoader에서 로드 (각 애플리케이션별 고유한 클래스이므로 개별 관리)

- 부모 위임 모델을 사용하는 이유 (Why Parent Delegation Model?)

    1. 보안(Security) 강화
        - JVM의 핵심 클래스(`java.lang.Object`, `java.lang.String`)를 애플리케이션에서 재정의하지 못하도록 보호
        - 악의적인 코드가 JVM의 기본 클래스를 덮어쓰거나 변조하는 것을 방지

        📌 예제: `java.lang.String`을 재정의하는 악성 코드 차단

        ```java
        package java.lang;

        public class String {
            public static void main(String[] args) {
                System.out.println("Hacked!");
            }
        }
        ```

        ➡ 부모 위임 모델 덕분에, `Bootstrap ClassLoader`에서 `java.lang.String`을 먼저 찾고, 위 코드를 무시하여 보안을 유지함.

    2. 클래스 충돌 방지
        - 서로 다른 애플리케이션이 같은 라이브러리를 사용할 때, 중복 로딩을 방지
        - 모든 애플리케이션이 공통적으로 사용하는 클래스(`javax.servlet.*`, `javax.sql.*`)는 공용 `ClassLoader`에서 로드하여 불필요한 메모리 낭비 방지

        📌 예제: 서블릿 API 충돌 방지
        - 만약 `javax.servlet.Servlet`이 각 애플리케이션의 `WEB-INF/lib`에서 로드된다면, 서로 다른 버전이 존재할 가능성이 있음.
        - 부모 위임 모델 덕분에, Tomcat의 공통 ClassLoader에서 한 번만 로드됨.

    3. 애플리케이션 간 클래스 격리 (Class Isolation)
        - 서로 다른 웹 애플리케이션이 같은 패키지를 사용하더라도 클래스 로더가 다르면 서로 독립적으로 동작
        - 애플리케이션 간 의도치 않은 클래스 공유 및 간섭을 방지

        📌 예제: 두 개의 WAR 파일이 같은 패키지 구조를 사용할 때

        ```sh
        webapps/
        ├── app1.war
        │   ├── WEB-INF/classes/com/example/MyService.class
        │   ├── WEB-INF/lib/library-1.0.jar
        ├── app2.war
        │   ├── WEB-INF/classes/com/example/MyService.class
        │   ├── WEB-INF/lib/library-2.0.jar
        ```

        ➡ `app1.war`의 `MyService.class`가 `library-1.0.jar`을 사용하고,
        ➡ `app2.war`의 `MyService.class`가 `library-2.0.jar`을 사용함.
        ✅ 부모 위임 모델 덕분에, 각 애플리케이션이 자신의 라이브러리만 참조하도록 보장됨.

- 부모 위임 모델의 한계 및 해결 방법

    1. 특정 라이브러리를 무조건 부모에서 찾도록 강제하는 문제
        - Spring Boot의 경우, `WEB-INF/lib`에 있는 특정 라이브러리를 우선적으로 로드하고 싶을 때 문제가 발생할 수 있음.

        ✅ 해결 방법: `tomcat.util.scan.StandardJarScanner` 설정 변경

        ```xml
        <Context>
            <JarScanner scanClassPath="false"/>
        </Context>
        ```

        ➡ `scanClassPath="false"`로 설정하면, 부모 클래스 로더에서 라이브러리를 찾지 않고, 애플리케이션의 `WEB-INF/lib`에서 우선 로드함.

    2. Spring Boot와 같은 Fat JAR 환경에서 문제 발생
        - Spring Boot는 `fat JAR`로 실행되므로, Tomcat의 기본적인 `ClassLoader`와 충돌할 가능성이 있음.
        - 이를 해결하기 위해 Spring Boot의 `LaunchedURLClassLoader`가 사용되며, 내부적으로 클래스를 동적으로 로드함.

        ✅ 해결 방법: `Spring Boot ClassLoader`를 활용

        ```java
        ClassLoader cl = Thread.currentThread().getContextClassLoader();
        System.out.println(cl.getClass().getName());
        // org.springframework.boot.loader.LaunchedURLClassLoader
        ```

        ➡ Spring Boot는 자체 ClassLoader를 사용하여 부모 위임 모델을 일부 우회함.

<!-- /source-chunk -->

<!-- source-chunk: sha256=21d3928f49c0b7c1820fb38a5130d9de92f36c9f5d35820a269fe916564f3242 topic=jvm-java-kotlin-runtime sources=interview_questions2.md:250-356, interviews2.md:250-356 -->

> Source: `interview_questions2.md:250-356`
> Duplicate source aliases: `interview_questions2.md:250-356, interviews2.md:250-356`

## 코루틴을 Cooperative Multitasking 이라고 하는 이유?

### Kotlin Coroutine

코루틴은 경량 쓰레드(lightweight thread)로, *협력적(cooperative)*으로 작업을 스케줄링합니다.
- 코루틴은 실제 OS 스레드가 아니며, JVM의 단일 스레드에서 여러 코루틴이 협력하여 실행됩니다.
- 명시적으로 `suspend` 키워드나 특정 지점에서 작업을 양보하며 다른 코루틴이 실행되도록 합니다.
- 컨텍스트 스위칭이 스레드 수준이 아니라 코루틴 수준에서 이루어집니다.

코루틴은 명시적으로 자신의 실행을 중단(`suspend`)하고, 다른 작업이 실행되도록 협력합니다.
이는 OS 스케줄러에 의한 선점형(Preemptive) 스레드 스케줄링과 달리, 프로그램 코드 수준에서 명시적으로 중단을 관리하는 방식입니다.
- 중단 지점(Suspension Point): `suspend` 키워드가 있는 함수나 작업은 실행 중간에 상태를 저장하고, 다음 실행 흐름으로 제어권을 넘깁니다.
- 상태 머신(State Machine):
    Kotlin 컴파일러는 코루틴을 상태 머신으로 변환하여 실행 흐름을 관리합니다.
    각 `suspend` 호출은 상태 전이를 나타냅니다.

    코루틴은 컴파일러가 코드를 변환하여 상태 머신으로 구현됩니다.
    이는 코루틴의 각 `suspend` 지점이 상태(state)로 변환되고, 실행 흐름이 상태 간 전환으로 표현되는 것을 의미합니다.

    상태 머신은 코루틴의 실행 흐름을 관리하며, 각 `suspend` 호출은 상태 전환을 유발합니다.

    ```kotlin
    suspend fun exampleCoroutine() {
        println("Step 1")
        delay(1000) // Suspension Point
        println("Step 2")
    }
    // 위 코드는
    // 아래와 같이 상태 머신으로 변환 됩니다.
    class ExampleCoroutine : Continuation<Unit> {
        var state = 0
        override fun resumeWith(result: Result<Unit>) {
            when (state) {
                0 -> {
                    println("Step 1")
                    state = 1
                    delay(1000, this) // Suspension Point
                }
                1 -> {
                    println("Step 2")
                }
            }
        }
    }
    ```

    `delay`는 코루틴을 일시 중단(suspend)하고, 나중에 재개(resume)될 수 있도록 `Continuation` 객체에 상태를 저장합니다.

    Kotlin의 코루틴은 Continuation-passing style (CPS)을 활용합니다.
    모든 suspend 함수는 컴파일러에 의해 암묵적으로 `Continuation` 객체를 받는 형태로 변환됩니다.
    `Continuation` 객체는 다음 실행 지점(상태)을 기억하고, 나중에 `resumeWith` 메서드를 호출하여 실행을 이어갑니다.

- 비동기 처리: 코루틴은 CPU를 점유하지 않고, 비동기적으로 작업을 대기하거나 다른 작업으로 전환됩니다.

코루틴은 스레드 위에서 실행됩니다.
즉, 코루틴의 실행은 하나 이상의 스레드에서 처리됩니다.

코루틴의 상태는 `Continuation` 객체에 저장됩니다.
`suspend` 함수는 호출된 지점에서 상태를 저장하고, 다음 작업을 스케줄링합니다.
작업 간 컨텍스트 스위칭은 매우 가볍고, 특정 지점에서 명시적으로 이루어집니다.

```kotlin
suspend fun performTask() {
    println("Task start on ${Thread.currentThread().name}")
    delay(1000) // Suspend here and allow other tasks to run
    println("Task end on ${Thread.currentThread().name}")
}
```

코루틴은 작업을 협력적으로 양보해야 다른 코루틴이 실행됩니다.
명시적으로 `delay`, `yield`, `withContext` 등을 사용해 CPU를 다른 작업으로 넘깁니다.
Dispatcher를 통해 코루틴의 실행 스레드를 관리합니다:
- `Dispatchers.Default`: 백그라운드 스레드 풀에서 실행.
- `Dispatchers.Main`: UI 스레드에서 실행 (Android).
- `Dispatchers.IO`: I/O 최적화 스레드 풀에서 실행.

### 일반적인 Java의 멀티스레딩

일반적인 Java의 멀티스레딩은 OS에서 제공하는 물리적 스레드를 사용하며, *선점형(preemptive)*으로 스케줄링됩니다.
- JVM 스레드는 OS 스레드로 매핑되며, 스레드는 OS 커널에 의해 스케줄링됩니다.
- 작업 간의 컨텍스트 스위칭은 OS가 자동으로 관리합니다.
- 각 스레드는 자체 스택과 메모리를 가지며, 독립적으로 실행됩니다.

Java 스레드는 OS에서 직접 관리되며, 각 스레드는 독립적으로 실행됩니다.
작업 간의 컨텍스트 스위칭은 OS 커널에 의해 자동으로 이루어집니다.
작업은 특정 시점에서 중단되고 다른 스레드가 실행될 수 있습니다.

```java
public class Task extends Thread {
    @Override
    public void run() {
        System.out.println("Task start on " + Thread.currentThread().getName());
        try {
            Thread.sleep(1000); // Pause thread
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        System.out.println("Task end on " + Thread.currentThread().getName());
    }
}
```

OS 커널은 작업을 선점적으로 스케줄링합니다.
실행 중인 스레드는 중단되고 다른 스레드가 실행될 수 있습니다.
Thread Scheduler는 CPU 시간을 각 스레드에 할당합니다.
사용자는 스레드 우선순위(`Thread.setPriority`)를 설정할 수 있지만, OS 스케줄러의 결정에 의존합니다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=ffbe6742cabd9b009c81d5584ca5541c668bd6e0aac7616a146862a6bc079ae2 topic=jvm-java-kotlin-runtime sources=interview_questions2.md:357-408, interviews2.md:357-408 -->

> Source: `interview_questions2.md:357-408`
> Duplicate source aliases: `interview_questions2.md:357-408, interviews2.md:357-408`

## 코루틴 제어의 양보

코루틴은 상태 머신(State Machine)을 통해 비동기 실행을 효율적으로 관리하며, 실행 중인 제어를 다른 코루틴에게 양보하는 것이 가능합니다.
- 코루틴은 단일 스레드 내에서 실행 중인 작업을 중단하고, 다른 작업(코루틴)이 실행될 수 있도록 제어권을 넘기는 방식으로 동작합니다.
- 제어의 양보는 비선점형 협력(cooperative multitasking)의 한 형태로, 명시적인 지점에서 실행 흐름이 중단됩니다.

제어 양보의 원리는 다음과 같습니다:
1. 코루틴의 상태 머신 설계:
    - 컴파일러는 코루틴을 상태 머신으로 변환합니다.
    - 상태 머신은 각 실행 지점(중단점)을 상태로 모델링하며, 다음 상태로 이동할 수 있는 로직을 포함합니다.
    - 이 구조는 `suspend` 키워드가 사용된 지점을 상태 변경의 기준으로 삼습니다.

    예제 (단순 코루틴):

    ```kotlin
    suspend fun fetchData() {
        println("Fetching data...")
        delay(1000)  // 중단점 (suspend)
        println("Data fetched!")
    }
    ```

    - `fetchData()`는 내부적으로 상태 머신으로 변환됩니다.
    - `delay(1000)`은 제어권을 현재 코루틴에서 반환하도록 지시하고, 특정 상태로 복귀 가능한 지점을 기록합니다.

2. 코루틴 디스패처와 실행 컨텍스트:
    - 코루틴 디스패처는 코루틴의 실행을 관리합니다. 주요 디스패처:
        - `Dispatchers.Default`: 백그라운드에서 CPU 집약적인 작업 실행.
        - `Dispatchers.IO`: I/O 작업 처리.
        - `Dispatchers.Main`: UI 업데이트 작업.
    - 실행 중단(`suspend`) 시, 디스패처가 제어권을 받아 다른 대기 중인 작업을 실행하거나 리소스를 다른 작업에 재할당합니다.

3. `suspend`와 컨티뉴에이션(Continuation):
    - `suspend` 함수는 중단 가능한 지점을 정의하며, 호출 스택 대신 컨티뉴에이션 객체를 생성해 호출 상태를 저장합니다.
    - 컨티뉴에이션은 "어디에서부터 다시 실행을 시작할지"를 나타냅니다.

    컨티뉴에이션 구조:

    ```kotlin
    suspend fun example() {
        println("Start")
        delay(1000) // `delay` 호출 시
                    // - 현재 상태를 저장하고,
                    // - 디스패처에게 제어권을 넘깁니다.
        println("End")
    }
    ```

4. 실제 동작 예:
    - `delay`나 `yield`와 같은 함수 호출은 스레드를 차단(blocking)하지 않습니다.
    - 대신, 현재 상태를 기록하고 다른 코루틴에게 실행 기회를 제공합니다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=d42f5b06f742e1addd1018ce374d6395a1bc25c2d300634df1feec047bf09026 topic=jvm-java-kotlin-runtime sources=interview_questions2.md:1499-1629, interviews2.md:1499-1629 -->

> Source: `interview_questions2.md:1499-1629`
> Duplicate source aliases: `interview_questions2.md:1499-1629, interviews2.md:1499-1629`

## infix, inline, noinline, crossinline

Kotlin에서 제공하는 `infix`, `inline`, `noinline`, `crossinline` 키워드는 각각 특정한 기능과 의도를 반영하여 설계된 특수한 언어 구성 요소입니다.

### Infix

`infix`는 함수 호출을 보다 직관적이고 읽기 쉽게 하기 위해 제공되는 기능입니다.
"중간(infix)"에 위치한다는 의미에서 이름이 붙었습니다. 이는 기존의 전위(prefix), 후위(postfix) 표현과 구분됩니다.
일반적으로 객체 메서드는 `object.method(argument)` 형식으로 호출되지만, `infix`를 사용하면 중위 표현식(`object method argument`)으로 사용할 수 있습니다.
주로 DSL(Domain-Specific Language)나 간결한 표현식이 필요한 경우에 유용합니다.

사용할 수 있는 조건은 다음과 같습니다:
- 멤버 함수나 확장 함수여야 합니다.
- 함수는 하나의 매개변수만 받아야 합니다.
- `infix` 키워드로 선언되어야 합니다.

`infix`는 JVM에서 특별한 바이트코드를 생성하지 않습니다.
단순히 호출 구문을 간결하게 바꿀 수 있는 문법적 설탕(Syntactic Sugar)입니다.
JVM에서는 여전히 일반 메서드 호출로 처리됩니다.

사용례:
- 집합 연산 DSL:

    ```kotlin
    infix fun String.append(other: String): String = this + other

    val result = "Hello" append "World"
    println(result)  // "HelloWorld"
    ```

- Kotlin 표준 라이브러리의 `to` 함수:

    ```kotlin
    val map = mapOf(1 to "one", 2 to "two")
    ```

### Inline

`inline` 함수는 호출 시점에 함수의 바디를 호출자 코드로 대체(inline)하여 런타임 오버헤드를 줄입니다.
"라인 안에 삽입"한다는 뜻에서 붙여졌습니다.
이는 JVM의 `method call overhead`를 줄이는 것을 목표로 합니다.
특히 람다식을 매개변수로 전달할 때 유용합니다.
고차 함수를 사용하는 코틀린 표준 라이브러리(`forEach`, `filter`, `map`)에서 성능 최적화를 위해 널리 사용됩니다.

컴파일러는 `inline`으로 표시된 함수를 호출하는 모든 지점에서 함수의 실제 코드를 삽입합니다.
람다식이 매개변수로 전달될 경우, 람다 표현식도 인라인됩니다.

장점:
- 함수 호출 비용을 제거하여 성능을 개선합니다.
- 람다 캡처로 인해 생성되는 추가 객체를 줄입니다.

단점:
- 코드 크기가 증가할 수 있습니다(Code Bloat).
- 지나치게 많은 인라인 함수는 성능에 부정적인 영향을 미칠 수 있습니다.

사용례:

- 고차 함수 최적화:

    ```kotlin
    inline fun performAction(action: () -> Unit) {
        println("Starting action")
        action()
        println("Ending action")
    }

    performAction {
        println("Executing action")
    }
    ```

    위 코드는 `performAction` 함수 바디를 호출자로 대체하여 추가 함수 호출을 방지합니다.

### Noinline

`noinline`은 `inline` 함수의 매개변수 중 특정 람다식을 인라인 처리하지 않도록 지정합니다.
"인라인되지 않는다(no-inline)"는 것을 명확히 하기 위해 붙여졌습니다.
`inline` 함수에서 모든 람다가 인라인되는 기본 동작을 제어할 때 사용됩니다.
람다식을 동적으로 조작해야 하는 프레임워크에서 사용됩니다.

JVM은 다음과 같이 동작합니다.
- `noinline` 람다 매개변수는 일반적인 객체로 처리되며, 실제 호출 시점에 생성됩니다.
- 호출자는 해당 람다를 별도의 객체로 전달받아 사용합니다.

사용례:
- 동작 분리:

    ```kotlin
    inline fun example(block1: () -> Unit, noinline block2: () -> Unit) {
        block1()
        block2()
    }

    example({
        println("Inline block")
    }, {
        println("Noinline block")
    })
    ```

    위 예제에서 `block2`는 일반 객체로 처리됩니다.

### Crossinline

`crossinline`은 `inline` 함수에서 람다식이 비지역(non-local) 리턴을 수행하지 못하도록 제한합니다.
"교차 컨텍스트에서의 인라인 제한"이라는 의미에서 이름이 붙었습니다.
보통 `inline` 함수 내부에서 람다를 다른 컨텍스트로 전달할 때 사용됩니다.
코루틴과 같은 비동기 처리에서 비지역 리턴을 방지하여 안전성을 높입니다.

JVM의 동작 방식은 다음과 같습니다:
- 람다의 리턴 동작을 제한하기 위해 추가적인 검사를 삽입합니다.
- `crossinline` 람다식은 단순히 값으로 캡처되며 비지역 리턴이 허용되지 않습니다.

사용례:
- 안전한 리턴 제한:

    ```kotlin
    inline fun safeExecution(crossinline action: () -> Unit) {
        println("Starting safe execution")
        val runnable = Runnable {
            action()  // 비지역 리턴 불가
        }
        runnable.run()
    }

    safeExecution {
        println("Executing safely")
        // return  // 컴파일 에러
    }
    ```

<!-- /source-chunk -->

<!-- source-chunk: sha256=bd3bb9b7fec3aa36236c525e46c19775fa6a88eec9d187c0697824747c9fb0f3 topic=jvm-java-kotlin-runtime sources=interview_questions3.md:92-129 -->

> Source: `interview_questions3.md:92-129`

## 1. Java 관련 질문

### 질문 1. 자바에서 메모리는 어떻게 구성되어 있으며, 각 영역은 어떤 목적을 가지나요?

#### **답변**

1. 자바 애플리케이션이 실행될 때, **JVM(Java Virtual Machine)**은 크게 여러 메모리 영역을 관리합니다. 그중 대표적으로 **힙(Heap) 영역**, **스택(Stack) 영역**, 그리고 **메타스페이스(Metaspace)** 영역이 존재합니다.
2. **힙(Heap) 영역**은 객체와 배열이 동적으로 할당되는 곳입니다. 자바의 모든 객체(예: new 키워드로 생성된 인스턴스)가 대부분 힙에 위치합니다. 가비지 컬렉터(GC)가 이 영역을 주기적으로 검사하여, 더 이상 참조되지 않는 객체를 해제합니다.
3. **스택(Stack) 영역**은 각 쓰레드마다 분리되어 있습니다. 한 쓰레드 안에서 메서드가 호출될 때마다 스택 프레임이 쌓이고, 지역 변수와 메서드 실행 컨텍스트가 이 스택 프레임에 보관됩니다. 메서드가 종료되면 해당 스택 프레임이 사라지면서 그 안의 지역 변수가 해제됩니다.
4. **메타스페이스(Metaspace)**는 자바 8부터 도입된 영역으로, 이전에 PermGen으로 불렸던 부분을 대체합니다. 클래스 메타데이터, 리플렉션 정보, 동적 프록시 정보 등을 저장합니다. 이 영역은 OS 메모리를 동적으로 사용할 수 있으므로, PermGen과 달리 OutOfMemoryError 위험이 어느 정도 줄었습니다.
5. 이처럼 자바 메모리가 여러 영역으로 나뉘어 있는 이유는, **가비지 컬렉션 효율성**과 **프로그램 구조화**에 있습니다. 힙은 객체를 통합적으로 관리하고, 스택은 메서드 호출 구조와 밀접하게 연관되며, 메타스페이스는 클래스 정보를 별도로 관리함으로써 JVM이 메모리를 더욱 체계적으로 운용할 수 있게 해줍니다.

---

### 질문 2. 자바 가비지 컬렉션 알고리즘에는 어떤 종류가 있으며, 어떤 특징을 갖나요?

#### **답변**

1. 자바는 객체 생명주기를 자동으로 관리하기 위해 가비지 컬렉션(Garbage Collection)을 사용합니다. 가비지 컬렉션 알고리즘이란, **더 이상 사용되지 않는(참조되지 않는) 객체를 탐색**하여 힙에서 제거하는 로직을 의미합니다.
2. Java 8까지 자주 사용되던 알고리즘으로는 **Serial GC**, **Parallel GC**, **CMS(Concurrent Mark-Sweep)** 등이 있습니다. Serial GC는 단일 쓰레드로 동작하여 간단하지만, 멀티코어 환경에서는 확장성이 떨어집니다. Parallel GC는 여러 쓰레드를 통해 동시에 객체를 수집하기 때문에 처리량(Throughput)이 높아집니다. CMS는 애플리케이션 스레드와 GC 스레드를 부분적으로 병행하여(Concurrent) 멈추는 시간(Stop-the-world pause)을 줄이려 합니다.
3. Java 9 이상에서는 **G1 GC**가 기본이 되었습니다. G1(Garbage First) GC는 힙을 여러 영역(Region)으로 나누고, 객체 참조도를 추적하여 우선적으로 수거해야 하는 영역부터 처리(garbage first)합니다. 이 방식은 대규모 힙에서 멈춤 시간을 줄이는 데 효과적입니다.
4. 최근에는 **ZGC**나 **Shenandoah** 같은 더 최신의 저지연 GC도 등장했습니다. 이들은 매우 큰 힙에서도 짧은 멈춤 시간(Stop-the-world)을 추구합니다. 예를 들어 ZGC는 큰 힙(수십 GB~TB급)에서도 최대 몇 ms 정도의 멈춤 시간을 목표로 합니다.
5. 이러한 다양한 GC 알고리즘이 존재하는 이유는, **애플리케이션 특성**(예: 대규모 서비스, 짧은 응답 시간 요구, CPU 자원 크기)에 따라 GC가 성능에 큰 영향을 미치기 때문입니다. 실제 서비스에서는 GC 로그를 모니터링하여, 적절한 GC 모드를 선택하고 필요하면 JVM 옵션을 조정해 튜닝합니다.

---

### 질문 3. Java 동시성(멀티쓰레드) 프로그래밍에서 주의해야 할 점은 무엇인가요?

#### **답변**

1. 멀티쓰레드 환경에서는 **공유 자원(shared resource)**에 여러 쓰레드가 동시에 접근할 수 있기 때문에, 동기화(synchronization) 문제가 핵심입니다. 즉, 서로 다른 쓰레드가 같은 데이터를 변경하는 순간, 올바르지 않은 결과가 발생할 수 있습니다(예: Race Condition).
2. 이를 방지하기 위해, 자바는 `synchronized` 키워드나 `ReentrantLock` 클래스를 통해 한 번에 하나의 쓰레드만 임계영역(critical section)에 들어가도록 제한합니다.
3. 또한 **volatile** 키워드는 변수의 가시성(visibility)을 보장하기 위한 것입니다. volatile로 선언된 변수에 쓰기가 일어나면, 모든 쓰레드가 즉시 최신 값을 확인하도록 보장합니다.
4. 자바 메모리 모델(JMM)에 따르면, 쓰레드 간에 값이 즉시 전파되지 않을 수도 있고(캐시 혹은 레지스터 최적화 등), 리오더(reordering)가 발생할 수도 있습니다. synchronized나 volatile을 사용하면 happens-before 관계를 형성하여 이러한 문제를 해결할 수 있습니다.
5. **동시성 프로그래밍**은 CPU 코어 수가 늘어난 현대 환경에서 필수적이지만, 설계가 잘못되면 교착상태(Deadlock)나 리소스 경합(Contention)으로 인해 성능이 급격히 하락할 수 있습니다. 따라서 멀티쓰레드를 사용할 때는 임계구역 최소화, 락 분리, 락 프리 알고리즘(Atomic 클래스) 고려 등 세밀한 접근이 필요합니다.

---

<!-- /source-chunk -->
