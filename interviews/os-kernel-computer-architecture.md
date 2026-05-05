# OS, 커널, 컴퓨터 구조

프로세스, 커널, 부팅, CPU, 메모리, 파일 디스크립터처럼 애플리케이션 아래층의 실행 조건을 다룹니다.

> 원문 배치본입니다. source chunk의 문장은 유지하고, 대분류/중분류/소분류 계층에 맞게 Markdown heading depth만 조정했습니다. 원본 span과 SHA-256은 manifest에서 검증할 수 있습니다.

## CPU와 숫자 표현

### 부동소수점

#### 원문: 부동소수점

<!-- curriculum-chunk: sha256=1d7853f2897ac15e68cd3ad7bf28a9e13b4579e97454088d2aa7693d00cef214 major=os-kernel-computer-architecture mid=CPU와 숫자 표현 sub=부동소수점 sources=source/interview_questions.md:233-634, source/interviews.md:233-634 -->

> Source: `source/interview_questions.md:233-634`
> Classification reason: cpu/numeric representation
> Duplicate source aliases: `source/interview_questions.md:233-634, source/interviews.md:233-634`

##### 부동소수점

1. 부동 소수점(Floating Point)의 기본 개념

    부동 소수점은 매우 크거나 매우 작은 실수를 표현하기 위한 컴퓨터의 수 표현 방식입니다.
    "부동(浮動, floating)"이란 소수점이 움직일 수 있다는 의미로, 이는 지수 표기법과 유사합니다.

    예를 들어:
    - 12345.6789는 1.23456789 × 10⁴로 표현 가능
    - 0.0000123456은 1.23456 × 10⁻⁵로 표현 가능

2. IEEE 754 표준

    IEEE 754는 부동 소수점 표현의 국제 표준으로, 다음과 같은 형식을 정의합니다:

    ```plaintext
    단정밀도(32비트)
    ┌──────┬──────────┬─────────────────────┐
    │ 부호  │  지수부    │        가수부        │
    │ (1)  │   (8)    │        (23)         │
    └──────┴──────────┴─────────────────────┘

    배정밀도(64비트)
    ┌──────┬───────────┬────────────────────────────────────────┐
    │ 부호  │  지수부     │               가수부                     │
    │ (1)  │   (11)    │               (52)                     │
    └──────┴───────────┴────────────────────────────────────────┘
    ```

    ```plaintext
    [부호 비트(1)] [지수부(k)] [가수부(n)]
    ```

    - 부호 비트(Sign bit):
        - 0: 양수
        - 1: 음수

    - 지수부(Exponent):
        - 단정밀도: -126 ~ +127 (바이어스: 127)
        - 배정밀도: -1022 ~ +1023 (바이어스: 1023)

        ```plaintext
        실제 지수 = 저장된 값 - 바이어스
        예: 단정밀도에서 129가 저장되어 있다면
        실제 지수 = 129 - 127 = 2
        ```

        bias를 사용하여 음수 지수 표현합니다:

    - 가수부(Mantissa/Significand):
        - 1.xxxxx 형태로 정규화
        - 첫 번째 1은 암묵적으로 저장 (정규화된 수의 경우)

    예를 들어, 32비트 단정밀도(single precision)의 경우:

    ```plaintext
    [부호(1)][지수부(8)][가수부(23)] = 총 32비트
    ```

    ```plaintext
    # 실제 값 계산
    값 = (-1)^부호비트 × 2^(지수부-127) × (1.가수부)

    # 예를 들어 10.625 변환
    10.625(10) = 1010.101(2)
               = 1.010101(2) × 2^3

    부호비트: 0 (양수)
    지수부: 3 + 127 = 130 = 10000010(2)
    가수부: 010101... = 01010100000000000000000

    최종: 0 10000010 01010100000000000000000

    # 예를 들어 12.375 변환
    1. 12.375를 이진수로 변환
        - 12(정수부) = 1100(2진수)
        - 0.375(소수부) = 0.011(2진수)

        => 1100.011(2진수)

    2. 정규화 (소수점을 이동하여 1.xxx 형태로)

        1100.011 = 1.100011 × 2³

    3. 각 필드의 비트 계산
        - 부호비트: 0 (양수)
        - 지수부: 3 + 127(바이어스) = 130
                130 = 10000010(2진수)
        - 가수부: 100011(나머지는 0으로 채움)

    3. 최종 32비트 표현

        0 10000010 10001100000000000000000

    # 예를 들어 3.25 변환
    1. 3.25를 이진수로 변환
        - 3(정수부) = 11(2진수)
        - 0.25(소수부) = 01(2진수)

        => 11.01(2진수)

    2. 정규화

        11.01 = 1.101 × 2¹

    3. 각 필드의 비트 계산
        - 부호비트: 0 (양수)
        - 지수부: 1 + 127 = 128
                128 = 10000000(2진수)

    4. 가수부: 101(나머지는 0으로 채움)

    Step 4: 최종 32비트 표현
    0 10000000 10100000000000000000000
    ```

3. 특수한 값들의 표현

    32비트 부동 소수점에서:
    - 0의 표현

        ```plaintext
        [0][00000000][00000000000000000000000] = +0
        [1][00000000][00000000000000000000000] = -0
        ```

    - 무한대

        ```plaintext
        [0][11111111][00000000000000000000000] = +∞
        [1][11111111][00000000000000000000000] = -∞
        ```

    - NaN (Not a Number)

        ```plaintext
        [x][11111111][non-zero mantissa] = NaN
        ```

4. 부동 소수점 연산의 구현

    - 덧셈/뺄셈의 경우:

        '12.375 + 3.25' 덧셈 과정의 비트 단위 연산은 다음과 같습니다.

        ```plaintext
        12.375 = 0 10000010 (1.)10001100000000000000000(32)
        3.25   = 0 10000000 (1.)10100000000000000000000(32)
                            ^^^^ '1.'이 있다고 가정

        1. 지수 비교
            - 첫 번째 수 (12.375)의 지수: 10000010 (130₁₀ - 127 = 3)
            - 두 번째 수 (3.25)의 지수:   10000000 (128₁₀ - 127 = 1)

            지수 차이: 2

        2. 가수부 덧셈 준비

            작은 지수를 가진 수의 가수 조정합니다.
            IEEE 754에서 유효한 덧셈을 하려면 두 수의 지수가 같아야 합니다.
            덧셈은 같은 자리의 숫자끼리만 가능하기 때문입니다

            따라서, 3.25의 가수부를 2비트 오른쪽 시프트합니다.
            2비트 오른쪽 시프트는 지수가 2 감소하는 것과 같습니다.

                1.101 × 2¹ = 0.01101 × 2³

            - 3.25  원래 가수:  1.10100 000000000000000000(23)
            - 3.25  시프트 후:  0.01101 000000000000000000(23)

            12.375와 조정된 3.25의 가수를 비교하면 아래와 같습니다.
            - 12.375의 가수: 1.100011 00000000000000
            - 3.25의 가수:   0.011010 00000000000000

        3. 이진 덧셈 수행

              1.100011 00000000000000
            + 0.011010 00000000000000
            -------------------------
              1.111111 00000000000000

        4. 결과 정규화 확인

            이미 1.xxx 형태이므로 추가 정규화 불필요

        5. 최종 비트 패턴 구성
            - 부호: 0 (양수)
            - 지수: 10000010 (130₁₀, 즉 3 + 127)
            - 가수: 11111100000000000000000

            최종 32비트 표현:
            0 10000010 11111100000000000000000
        ```

        ```java
        class FloatingPointOperation {
            /
             * 부동 소수점 덧셈의 의사 코드
            * @param a 첫 번째 피연산자
            * @param b 두 번째 피연산자
            * @return a + b의 결과
            */
            float add(float a, float b) {
                // 1. 지수 비교
                int expA = getExponent(a);
                int expB = getExponent(b);

                // 2. 작은 지수를 가진 수의 가수를 오른쪽으로 시프트
                int shift = Math.abs(expA - expB);
                float smaller = (expA < expB) ? a : b;
                float larger = (expA < expB) ? b : a;

                // 3. 가수부 덧셈
                float mantissaSum = alignAndAdd(smaller, larger, shift);

                // 4. 결과 정규화
                return normalize(mantissaSum, Math.max(expA, expB));
            }
        }
        ```

    - 곱셈의 경우:

        original: 40.218750
        0 10000100 01000001110000000000000

        IEEE 754 분석:
        Sign bit: 0
        Exponent: 10000100 (132)
        Mantissa: 1.01000001110000000000000

        '12.375 x 3.25' 곱셈 과정의 비트 단위 연산은 다음과 같습니다.

        ```plaintext
        12.375 = 0 10000010 (1.)10001100000000000000000(32)
        3.25   = 0 10000000 (1.)10100000000000000000000(32)
                            ^^^^ '1.'이 있다고 가정

        1. 부호 비트 계산 (XOR 연산)

            0 XOR 0 = 0 (양수)

        2. 지수 덧셈
            - 첫 번째 지수: 10000010 (130₁₀ - 127 = 3)
            - 두 번째 지수: 10000000 (128₁₀ - 127 = 1)

            130 + 128 - 127 = 131 (10000011)
            - Note: 이 값은 정규화 과정에서 조정됩니다.

        3. 가수부 곱셈

            이진수 곱셈은 십진수 곱셈과 유사하게, 각 비트에 대해 0 또는 1로 곱한 후 자리 이동(시프트)하여 더해주는 방식입니다.

               1.100011
            ×     1.101
            -----------
                1100011   (× 1)
               0000000    (× 0)
              1100011     (× 1)
             1100011      (× 1)
                  └┘
            -----------------
            10100000111

        4. 정규화

            10.100000111... × 2⁴ = 1.0100000111... × 2⁵

            - 새 지수: 5 + 127 = 132 (10000100)

        5. 가수 반올림

            1.0010111111... → 1.001011111100000000000

        6. 최종 비트 패턴
            - 부호: 0
            - 지수: 10000100
            - 가수: 01000001110000000000000

        최종 32비트 표현:
        0 10000100 01000001110000000000000
        ```

        ```java
        class FloatingPointMultiplication {
            /
             * 부동 소수점 곱셈의 의사 코드
            * @param a 첫 번째 피연산자
            * @param b 두 번째 피연산자
            * @return a * b의 결과
            */
            float multiply(float a, float b) {
                // 1. 부호 결정
                int sign = getSign(a) ^ getSign(b);

                // 2. 지수 더하기
                int exp = getExponent(a) + getExponent(b) - bias;

                // 3. 가수부 곱하기
                float mantissa = multiplyMantissas(getMantissa(a), getMantissa(b));

                // 4. 결과 정규화
                return normalize(sign, mantissa, exp);
            }
        }
        ```

5. 32비트와 64비트의 차이점

    - 정밀도:
        - 32비트: 약 7자리의 십진 정밀도
        - 64비트: 약 15-17자리의 십진 정밀도

    - 범위:
        - 32비트: ±1.18 × 10⁻³⁸ ~ ±3.4 × 10³⁸
        - 64비트: ±2.23 × 10⁻³⁰⁸ ~ ±1.80 × 10³⁰⁸

    - 메모리 사용:

        ```java
        float f = 1.0f;    // 32비트 = 4바이트
        double d = 1.0;    // 64비트 = 8바이트
        ```

6. 정밀도 손실과 오차

    ```java
    class PrecisionLoss {
        void demonstratePrecisionLoss() {
            float f = 0.1f;
            double d = 0.1;

            // 0.1을 10번 더하기
            float sumF = 0.0f;
            double sumD = 0.0;

            for (int i = 0; i < 10; i++) {
                sumF += f;
                sumD += d;
            }

            // sumF ≈ 1.0000001
            // sumD ≈ 1.0000000000000007
            // 정확한 값 1.0과는 약간의 차이가 있음
        }
    }
    ```

7. 실제 응용에서의 고려사항

    - 금융 계산:

        ```java
        class FinancialCalculations {
            // 금융 계산에는 부동 소수점 대신 BigDecimal 사용 권장
            void calculateMoney() {
                // 잘못된 방법
                double price = 19.99;
                double tax = price * 0.06;

                // 올바른 방법
                BigDecimal price = new BigDecimal("19.99");
                BigDecimal tax = price.multiply(new BigDecimal("0.06"));
            }
        }
        ```

    - 과학 계산:

        ```java
        class ScientificCalculations {
            // 과학 계산에서는 double 사용이 일반적
            void calculatePhysics() {
                double gravity = 9.81;
                double time = 2.5;
                double distance = 0.5 * gravity * time * time;
            }
        }
        ```

8. 오차 처리:

    ```java
    class ErrorHandling {
        /
         * 부동 소수점 비교를 위한 안전한 방법
         */
        boolean approximatelyEqual(double a, double b) {
            final double EPSILON = 1e-10;
            return Math.abs(a - b) <= EPSILON;
        }

        /
         * 상대 오차를 사용한 비교
         */
        boolean relativelyEqual(double a, double b) {
            final double EPSILON = 1e-10;
            return Math.abs(a - b) <= EPSILON * Math.max(Math.abs(a), Math.abs(b));
        }
    }
    ```

<!-- /curriculum-chunk -->

## 부팅과 init 시스템

### systemd가 다른 시스템 데몬들을 실행하는 과정

#### 원문: systemd가 다른 시스템 데몬들을 실행하는 과정

<!-- curriculum-chunk: sha256=ffac218aec5744a3d057b787644357f3a9529029f62bf2a141c28adc4fb97ff5 major=os-kernel-computer-architecture mid=부팅과 init 시스템 sub=systemd가 다른 시스템 데몬들을 실행하는 과정 sources=source/interview_questions.md:3149-3225, source/interviews.md:3149-3225 -->

> Source: `source/interview_questions.md:3149-3225`
> Classification reason: boot/init
> Duplicate source aliases: `source/interview_questions.md:3149-3225, source/interviews.md:3149-3225`

##### systemd가 다른 시스템 데몬들을 실행하는 과정

systemd가 다른 시스템 데몬들을 실행하는 과정은 기본적으로 fork와 exec의 조합을 사용합니다.
하지만 systemd는 전통적인 Unix 방식에 몇 가지 최적화와 추가 기능을 더했습니다.

1. 기본 원리: fork와 exec

    - fork(): systemd는 자신의 프로세스를 복제(fork)합니다.
    - exec(): 그 후 fork된 자식 프로세스에서 목표 데몬(예: sshd)으로 실행 이미지를 교체(exec)합니다.

2. systemd의 최적화:

    - 소켓 활성화 (Socket Activation):
        - systemd는 서비스를 위한 소켓을 미리 생성하고 관리합니다.
        - 실제 서비스 프로세스는 필요할 때만 시작됩니다.

    - D-Bus 활성화:
        - 특정 D-Bus 인터페이스가 요청될 때 서비스를 시작합니다.

    - 병렬 실행:
        - 의존성을 고려하여 여러 서비스를 동시에 시작합니다.

3. 실제 프로세스 생성 과정:

   ```sh
   systemd
     |
     ├─ fork()
     |    |
     |    └─ exec(sshd)  -> sshd 프로세스
     |
     ├─ fork()
     |    |
     |    └─ exec(httpd) -> httpd 프로세스
     |
     ├─ fork()
          |
          └─ exec(crond) -> crond 프로세스
   ```

4. 추가적인 고려사항:

    - Cgroups: systemd는 각 서비스를 별도의 cgroup에 배치하여 리소스 관리와 격리를 수행합니다.

    - Namespaces: 일부 서비스는 별도의 namespace에서 실행되어 추가적인 격리를 제공할 수 있습니다.

    - 서비스 유형: systemd는 다양한 서비스 유형(simple, forking, oneshot 등)을 지원하며, 각 유형에 따라 프로세스 생성 방식이 약간 다를 수 있습니다.

5. 코드 예시 (systemd의 단순화된 서비스 시작 로직):

    ```c
    pid_t pid = fork();
    if (pid == 0) {  // 자식 프로세스
        // 환경 설정, 리소스 제한 등 적용
        setup_environment();
        setup_resource_limits();

        // 소켓 활성화의 경우, 미리 생성된 소켓 파일 디스크립터 전달
        if (socket_activated) {
            pass_socket_fds();
        }

        // 실제 서비스 실행
        execv(service_path, service_args);
        _exit(EXIT_FAILURE);  // exec 실패 시
    } else if (pid > 0) {  // 부모 프로세스 (systemd)
        // 자식 프로세스 추적, 상태 관리 등
        monitor_child_process(pid);
    } else {
        // fork 실패 처리
        handle_fork_error();
    }
    ```

이렇게 systemd는 기본적인 fork-exec 모델을 사용하면서도, 추가적인 기능과 최적화를 통해 더 효율적이고 유연한 서비스 관리를 제공합니다.
소켓 활성화, D-Bus 활성화, cgroups 사용 등의 기능은 전통적인 init 시스템에서는 볼 수 없었던 systemd의 특징적인 기능들입니다.

<!-- /curriculum-chunk -->

### 서버 부팅 과정

#### 원문: 서버 부팅 과정

<!-- curriculum-chunk: sha256=297fc253351de9246b662e928e0800bf4cc46ecae95cb4f6f906152b8df49326 major=os-kernel-computer-architecture mid=부팅과 init 시스템 sub=서버 부팅 과정 sources=source/interview_questions.md:3025-3148, source/interviews.md:3025-3148 -->

> Source: `source/interview_questions.md:3025-3148`
> Classification reason: boot/init
> Duplicate source aliases: `source/interview_questions.md:3025-3148, source/interviews.md:3025-3148`

##### 서버 부팅 과정

1. 전원 공급 및 하드웨어 초기화

    서버에 전원이 공급되면 다음 과정이 진행됩니다:

    - PSU (Power Supply Unit, 전원 공급 장치):
        서버의 모든 컴포넌트에 적절한 전압과 전류를 공급하는 하드웨어입니다.

    - CPU (Central Processing Unit, 중앙 처리 장치):
        컴퓨터의 '두뇌' 역할을 하는 핵심 부품으로, 연산과 제어를 담당합니다.

    - 리셋 벡터 (Reset Vector):
        CPU가 리셋될 때 실행을 시작하는 메모리 주소입니다. x86 아키텍처에서는 일반적으로 0xFFFFFFF0입니다.

    - 메모리 컨트롤러 초기화: RAM을 사용 가능한 상태로 준비합니다.
    - 마더보드 칩셋 초기화: 다양한 하드웨어 컴포넌트 간의 통신을 설정합니다.

2. BIOS/UEFI 실행

    대부분의 현대 시스템은 BIOS나 UEFI 중 하나만 사용합니다.
    - BIOS (Basic Input/Output System):
        컴퓨터 하드웨어를 초기화하고 제어하는 펌웨어입니다. 레거시 시스템에서 주로 사용됩니다.

    - UEFI (Unified Extensible Firmware Interface):
        BIOS의 현대적인 대체제로, 더 빠른 부팅 시간과 향상된 보안 기능을 제공합니다.

        - UEFI 보안 부팅: 디지털 서명된 부트 로더와 커널만 실행을 허용하여 악성 소프트웨어로부터 부팅 과정을 보호합니다.

    - POST (Power-On Self-Test):
        컴퓨터가 부팅될 때 하드웨어의 정상 작동 여부를 확인하는 진단 과정입니다.

    - 부팅 순서 결정: BIOS/UEFI 설정에 따라 부팅 장치의 우선순위를 결정합니다.
    - 하드웨어 열거 및 초기화: 연결된 모든 장치를 식별하고 초기화합니다.

3. 부트 로더 로딩

    부트 로더는 운영 체제를 로드하는 소프트웨어 프로그램입니다.
    운영 체제가 올바르게 시작될 수 있도록 필요한 초기 환경을 제공합니다.

    디스크는 MBR이나 GPT 중 하나의 파티션 테이블만 사용합니다.
    UEFI 시스템은 주로 GPT를 사용하며, 레거시 BIOS 시스템은 MBR을 사용합니다.
    - MBR (Master Boot Record): 하드 디스크의 첫 번째 섹터로, 부트 로더의 첫 단계와 파티션 테이블을 포함합니다.
    - GPT (GUID Partition Table): MBR의 현대적 대안으로, 더 큰 디스크와 더 많은 파티션을 지원합니다.

    일반적으로 하나의 부트 로더만 사용됩니다 (예: GRUB2).
    - GRUB2 (GRand Unified Bootloader version 2): 가장 널리 사용되는 오픈 소스 부트 로더입니다.

    서버는 로컬 디스크에서 부팅하거나 PXE를 통해 네트워크에서 부팅합니다.
    - PXE (Preboot eXecution Environment): 네트워크를 통한 부팅을 지원하는 프로토콜로, 서버 환경에서 중앙 집중식 OS 배포에 사용됩니다.

4. 부트 로더 실행

    부트 로더는 다양한 커널 버전이나 운영 체제 중 선택할 수 있는 메뉴를 제공하고, 선택된 커널에 필요한 파라미터를 전달합니다
    부트 로더의 작동 원리는 다음과 같습니다.

    ```plaintext
    BIOS/UEFI가 부트 로더 로드
        |
        ˅
    부트 로더 초기화
        |
        ˅
    설정 파일 읽기 (예: /boot/grub/grub.cfg)
        |
        ˅
    사용자 선택 또는 기본 옵션 사용
        |
        ˅
    커널 및 initrd 로딩
        |
        ˅
    커널에 제어권 전달
    ```

    1. 커널 로딩:

        운영 체제의 핵심 부분으로, 하드웨어와 소프트웨어 사이의 중개자 역할을 하는 커널을 메모리에 로드합니다.

    2. 커널 로딩 후:

        - 커널 압축 해제: 대부분의 커널 이미지는 압축되어 있어, 메모리에 로드 후 압축을 해제합니다.
        - 커널 초기화: 메모리 관리, 장치 드라이버, 스케줄러 등 핵심 서브시스템을 초기화합니다.

    3. 초기 RAM 디스크 (initrd/initramfs) 로딩:

        임시 루트 파일 시스템을 로드합니다.
        대부분의 현대 시스템은 initramfs를 사용하며, initrd는 레거시 시스템에서 사용됩니다.

        - initrd (initial ramdisk):
            임시 루트 파일 시스템으로, 실제 루트 파일 시스템을 마운트하기 전에 필요한 드라이버와 커널 모듈을 제공합니다.

        - initramfs (initial RAM filesystem):
            initrd의 현대적 버전으로, 더 유연하고 효율적인 구조를 가집니다.

        - 필수 드라이버 로드: 실제 루트 파일 시스템에 접근하기 위한 드라이버를 로드합니다.

    4. 커널 파라미터 전달: 부팅 옵션을 커널에 전달합니다.
    5. 멀티부팅: 여러 운영 체제 중 선택할 수 있게 합니다.

5. 시스템 초기화 및 서비스 시작

    - 실제 루트 파일 시스템 마운트: initrd/initramfs에서 실제 루트 파일 시스템으로 전환합니다.
    - systemd:

        systemd는 많은 Linux 배포판에서 사용하는 초기화 시스템(init system)이자 시스템 및 서비스 관리자입니다.
        이전의 SysV init 시스템을 대체했습니다.

        시스템 부팅 시 프로세스 ID 1로 실행되며, 다른 모든 프로세스의 부모 프로세스 역할을 합니다.
        systemd의 주요 역할은 시스템 데몬을 포함한 다양한 서비스를 관리하고 시작/중지하는 것입니다.

        systemd는 시스템 상태 및 로그 관리, 네트워크 설정, 전원 관리 등 다양한 시스템 관리 작업을 통합적으로 처리합니다.

        - 병렬 서비스 시작: systemd는 의존성을 고려하여 여러 서비스를 동시에 시작할 수 있습니다.
        - 소켓 활성화: 필요 시 서비스를 시작하는 소켓 기반 활성화를 지원합니다.

    - 시스템 데몬 시작:

        시스템 데몬은 백그라운드에서 실행되는 프로세스로, 특정 서비스나 기능을 제공합니다.
        예를 들어 sshd (SSH 서버), httpd (웹 서버), crond (작업 스케줄러) 등이 있습니다.
        systemd가 이러한 네트워크, 로깅, 보안 등 기본 시스템 서비스(데몬)를 시작합니다.

    - 사용자 공간 초기화: 로그인 프롬프트 또는 그래픽 사용자 인터페이스를 시작합니다.

<!-- /curriculum-chunk -->

## 서버 하드웨어와 운영 환경

### 노트북도 서버가 될 수 있는데 왜 굳이 서버 호스팅을 받나요?

#### 원문: 노트북도 서버가 될 수 있는데 왜 굳이 서버 호스팅을 받나요?

<!-- curriculum-chunk: sha256=5546e984111fe51a018fa17df1962d0af140af3eceeec9f3624a27f8e5d97daa major=os-kernel-computer-architecture mid=서버 하드웨어와 운영 환경 sub=노트북도 서버가 될 수 있는데 왜 굳이 서버 호스팅을 받나요? sources=source/interview_questions.md:7419-7420, source/interviews.md:7367-7368 -->

> Source: `source/interview_questions.md:7419-7420`
> Classification reason: server environment
> Duplicate source aliases: `source/interview_questions.md:7419-7420, source/interviews.md:7367-7368`

##### 노트북도 서버가 될 수 있는데 왜 굳이 서버 호스팅을 받나요?

<!-- /curriculum-chunk -->

### 왜 맥북이나 가정용 PC로 서비스하지 않고 별도의 서버용 컴퓨터를 사용하나요?

#### 원문: 왜 맥북이나 가정용 PC로 서비스하지 않고 별도의 서버용 컴퓨터를 사용하나요?

<!-- curriculum-chunk: sha256=06633e3e034ce7f7754993d679630cee6cc9289d4312818dacd9d410d426df8d major=os-kernel-computer-architecture mid=서버 하드웨어와 운영 환경 sub=왜 맥북이나 가정용 PC로 서비스하지 않고 별도의 서버용 컴퓨터를 사용하나요? sources=source/interview_questions.md:2754-2923, source/interviews.md:2754-2923 -->

> Source: `source/interview_questions.md:2754-2923`
> Classification reason: server environment
> Duplicate source aliases: `source/interview_questions.md:2754-2923, source/interviews.md:2754-2923`

##### 왜 맥북이나 가정용 PC로 서비스하지 않고 별도의 서버용 컴퓨터를 사용하나요?

서버용 컴퓨터와 가정용 PC는 하드웨어 설계 목적에 따라 많은 차이점을 가지고 있습니다.
서버는 안정성, 확장성, 고가용성, 성능을 목표로 설계되었으며, 가정용 PC는 일반적으로 비용과 성능의 균형을 중시합니다.

- 프로세서 (CPU)

    - 가정용 PC 프로세서
        - 예시 모델: Intel Core i9-13900K, AMD Ryzen 9 7950X
        - 코어/스레드 수: 16코어 32스레드 (Ryzen 9), 24코어 32스레드 (Intel Core)
        - 클럭 속도: 최대 5.8GHz (터보 부스트), 기본 클럭 3.0~4.0GHz
        - 캐시: L3 캐시 32MB (Ryzen 9), L3 캐시 36MB (Intel Core)
        - 지원 메모리: DDR5 최대 128GB, 비-ECC 메모리 지원
        - 특징:
            - 높은 클럭 속도로 단일 스레드 성능을 극대화.
            - 비-ECC 메모리만 지원, ECC 메모리 필요 시 제한적.
            - 소비 전력(일반적으로 125W~250W)은 높은 편이며, 주로 게이밍 및 멀티미디어 작업에 최적화됨.

    - 서버용 프로세서
        - 예시 모델: Intel Xeon Gold 6258R, AMD EPYC 7773X
        - 코어/스레드 수: 28코어 56스레드 (Xeon), 64코어 128스레드 (EPYC)
        - 클럭 속도: 2.7GHz (기본), 터보 부스트 시 4.0GHz 이하
        - 캐시: L3 캐시 38.5MB (Xeon), L3 캐시 256MB (EPYC)
        - 지원 메모리: 최대 4TB 이상 (DDR4 ECC 지원), 멀티 채널 메모리 컨트롤러
        - 특징:
            - 다중 코어에 최적화된 설계로 멀티스레드 작업에서 성능 극대화.
            - ECC 메모리 지원으로 메모리 오류를 감지 및 수정 가능.
            - 서버급 프로세서는 NUMA (Non-Uniform Memory Access) 아키텍처를 통해 다수의 메모리 채널에 고속으로 접근 가능.
            - 전력 소비는 200~400W로 높지만, 장기적인 고부하 작업에 최적화.

    가정용 PC는 일반적으로 더 높은 클럭 속도를 제공하여 단일 스레드 성능에 유리하지만, 서버용 CPU는 다중 스레드 처리와 안정성을 중시합니다.
    서버용 CPU는 가정용 PC보다 훨씬 많은 코어를 지원하여 멀티태스킹 성능이 뛰어납니다.
    서버용 CPU는 수십 개의 메모리 슬롯을 지원하며, 수 TB의 ECC 메모리를 사용할 수 있습니다.

- 메모리 (RAM)

    - 가정용 PC 메모리
        - 예시 모델: Corsair Vengeance LPX DDR5 128GB (4 x 32GB)
        - 메모리 용량: 최대 128GB (일반적으로 32GB~64GB 사용)
        - 메모리 유형: DDR5, 비-ECC
        - 특징:
            - ECC 기능이 없어 메모리 오류에 취약함.
            - 주로 게이밍, 멀티미디어 작업 등에서 고속으로 작동하도록 설계됨.
            - 비-ECC 메모리는 주기적으로 메모리 오류가 발생할 수 있으며, 서버 환경에서는 데이터 무결성이 보장되지 않음.

    - 서버용 메모리
        - 예시 모델: Samsung DDR4 ECC Registered RAM 256GB (8 x 32GB)
        - 메모리 용량: 최대 4TB (서버용 마더보드와 CPU에 따라 달라짐)
        - 메모리 유형: ECC (Error-Correcting Code) 메모리
        - 특징:
            - ECC 기능: 메모리 비트 오류를 실시간으로 감지하고 수정하여 데이터 무결성을 보장.
            - 대용량 지원: 멀티 소켓 시스템에서 테라바이트 단위의 메모리 지원 가능.
            - 메모리 채널 수가 많아 멀티채널 메모리 구성이 가능하여 데이터 전송 대역폭이 훨씬 넓음.

    서버 메모리는 ECC 기능을 통해 메모리 오류를 감지하고 수정하여 데이터 무결성이 보장되지만, 가정용 PC 메모리는 ECC를 지원하지 않아 안정성 면에서 차이가 있습니다.
    서버는 수 TB의 메모리를 지원하는 반면, 가정용 PC는 128GB 정도가 최대입니다.

- 저장 장치 (Storage)

    - 가정용 PC 저장 장치
        - 예시 모델: Samsung 980 PRO 2TB NVMe SSD
        - 인터페이스: PCIe 4.0 NVMe
        - 속도: 읽기 속도 7,000 MB/s, 쓰기 속도 5,000 MB/s
        - 신뢰성: 일반적인 소비자용 SSD는 보통 600 TBW (Total Bytes Written)의 수명을 가짐.
        - 특징:
            - 고성능 SSD지만, 내구성과 신뢰성은 서버용 SSD에 비해 떨어짐.
            - 고용량의 데이터 기록/삭제가 반복될 경우 성능 저하가 발생할 수 있음.

    - 서버용 저장 장치
        - 예시 모델: Intel DC P4510 4TB NVMe SSD (엔터프라이즈급)
        - 인터페이스: PCIe 3.0 NVMe
        - 속도: 읽기 속도 3,200 MB/s, 쓰기 속도 3,000 MB/s
        - 신뢰성: 20,000 TBW 이상의 내구성을 가짐.
        - 특징:
            - RAID 구성 가능: 데이터를 안전하게 보호하기 위한 RAID 1, 5, 6, 10 구성이 가능함.
            - 핫스왑 지원: 하드웨어 오류 발생 시 시스템 중단 없이 디스크 교체 가능.
            - 엔터프라이즈급 SSD는 데이터 센터 환경에서 지속적인 쓰기 작업에도 내구성이 뛰어남.

    서버용 저장 장치는 높은 TBW 수명과 강력한 내구성을 제공하여 데이터 무결성을 유지합니다.
    PC용 저장 장치는 성능은 뛰어나지만, 내구성에서 서버급 장치보다 낮습니다.

    서버용 저장 장치는 RAID 구성을 통해 데이터 보호를 강화하지만, PC는 일반적으로 단일 디스크로 구성됩니다.

- 전원 공급 장치 (PSU)

    - 가정용 PC 전원 공급 장치
        - 예시 모델: Corsair RM850x 850W
        - 출력: 850W
        - 특징:
            - 단일 PSU 구성으로 설계되어 고장 시 전력 공급 중단.
            - 게이밍 및 고성능 작업용으로 설계되었으나, 서버처럼 지속적인 고부하 작업에 적합하지 않음.

    - 서버용 전원 공급 장치
        - 예시 모델: HPE 1200W 플래티넘 핫스왑 전원 공급 장치 (Dual PSU)
        - 출력: 1200W
        - 특징:
            - 중복 전원 공급 장치: 서버는 이중화된 PSU 구성을 지원하여 하나의 PSU가 고장 나도 다른 PSU가 전력을 공급하여 시스템을 중단 없이 운영 가능.
            - 핫스왑 기능: PSU 고장 시 시스템 운영 중에도 PSU를 교체 가능.
            - 에너지 효율: 서버용 PSU는 일반적으로 플래티넘 등급 이상의 고효율 제품을 사용하여 전력 낭비를 줄임.

    서버는 PSU가 중복되어 전원 문제 시에도 지속적으로 운영될 수 있습니다.
    가정용 PC는 단일 PSU만 사용하므로 PSU 고장 시 시스템이 중단될 위험이 있습니다.

    서버는 PSU 고장 시 교체를 위한 시스템 중단이 필요 없지만, PC는 반드시 시스템을 종료해야 합니다.

- 네트워크 인터페이스

    - 가정용 PC 네트워크
        - 예시 모델: Intel I219-V 기가비트 이더넷 컨트롤러
        - 속도: 최대 1Gbps
        - 특징:
            - 일반적인 가정용 인터넷 속도를 지원하며, 다중 포트를 지원하지 않음.
            - 주로 단일 네트워크 연결을 지원하며, 고속 네트워크 사용이 제한됨.

    - 서버용 네트워크
        - 예시 모델: Intel X710 10GbE Quad-Port 네트워크 어댑터
        - 속도: 10Gbps, 40Gbps 이상 지원 가능
        - 특징:
            - 다중 포트: 서버는 일반적으로 다중 네트워크 인터페이스를 갖추고 있으며, 10Gbps 이상의 고속 네트워크를 지원함.
            - 리던던시 지원: 네트워크 장애 시 자동으로 백업 네트워크로 전환되는 기능 제공.
            - 대규모 트래픽을 처리할 수 있으며, 데이터 센터 환경에서 안정적인 대역폭을 제공함.

    서버 네트워크 인터페이스는 PC에 비해 훨씬 빠른 네트워크 속도와 대역폭을 제공합니다.
    서버는 다중 네트워크 포트를 사용하여 더 높은 가용성과 안정성을 제공하지만, PC는 단일 기가비트 이더넷만을 지원합니다.

같은 코어 수와 RAM 크기를 가진 PC와 서버 컴퓨터가 있다고 가정해도, 서버 컴퓨터를 선택하는 데에는 여러 중요한 이유가 있습니다:

1. 프로세서 아키텍처의 차이:
    서버용 프로세서(예: Intel Xeon)는 데스크톱 프로세서와 다른 아키텍처를 가집니다.
    이들은 다음과 같은 특징을 갖습니다:
    - 더 큰 캐시 메모리: L3 캐시가 훨씬 크며, 이는 대량의 데이터를 처리할 때 성능을 향상시킵니다.
    - 고급 명령어 세트: AVX-512와 같은 고급 벡터 처리 명령어를 지원하여 과학 계산, 암호화 등에서 우수한 성능을 발휘합니다.
    - 더 높은 메모리 대역폭: 멀티 채널 메모리 컨트롤러를 통해 더 많은 동시 메모리 액세스를 지원합니다.

2. ECC (Error-Correcting Code) 메모리:
    서버용 시스템은 ECC 메모리를 사용합니다.
    ECC 메모리는 다음과 같은 이점을 제공합니다:
    - 비트 플립 오류 감지 및 수정: 우주 방사선 등으로 인한 메모리 오류를 실시간으로 수정합니다.
    - 데이터 무결성 보장: 금융 거래, 과학 계산 등 정확성이 중요한 작업에 필수적입니다.

3. 고급 I/O 기능:
    서버 하드웨어는 더 발전된 I/O 기능을 제공합니다:
    - PCIe 레인 수 증가: 더 많은 고성능 장치(NVMe SSD, GPU 등)를 동시에 사용할 수 있습니다.
    - NUMA (Non-Uniform Memory Access) 아키텍처: 대규모 멀티프로세서 시스템에서 메모리 액세스 최적화를 제공합니다.

4. 전원 관리 및 발열 처리:
    서버 하드웨어는 지속적인 고부하 작업을 위해 설계되었습니다:
    - 고급 전력 관리 기능: 부하에 따른 동적 전압 및 주파수 조정(DVFS)이 더 세밀하게 제어됩니다.
    - 효율적인 열 설계: 고성능 방열 시스템으로 장시간 최대 성능 유지가 가능합니다.

5. 확장성:
    서버 시스템은 미래의 확장을 고려하여 설계되었습니다:
    - 다중 소켓 지원: 단일 시스템에서 여러 물리적 프로세서를 사용할 수 있습니다.
    - 대용량 메모리 지원: 테라바이트 단위의 RAM을 지원하는 메모리 컨트롤러를 갖추고 있습니다.

6. 신뢰성 및 가용성 기능:
    서버 하드웨어는 중단 없는 운영을 위한 기능을 제공합니다:
    - 핫스왑 가능한 구성 요소: 전원 공급 장치, 팬, 드라이브 등을 시스템 운영 중에 교체할 수 있습니다.
    - 예측 오류 분석(PFA): 하드웨어 구성 요소의 잠재적 오류를 미리 감지하고 경고합니다.

7. 관리 기능:
    서버는 원격 및 자동화된 관리를 위한 고급 기능을 제공합니다:
    - 내장 관리 컨트롤러(예: iLO, iDRAC): 하드웨어 수준의 모니터링 및 제어가 가능합니다.
    - 고급 시스템 이벤트 로깅: 하드웨어 및 시스템 수준의 이벤트를 상세히 기록하여 문제 해결을 용이하게 합니다.

8. 보안 기능:
    서버 하드웨어는 엔터프라이즈급 보안 기능을 제공합니다:
    - TPM(Trusted Platform Module): 하드웨어 기반 암호화 및 키 관리를 제공합니다.
    - 보안 부팅: 부팅 프로세스의 무결성을 보장하여 부트킷과 같은 위협을 방지합니다.

<!-- /curriculum-chunk -->

## 스케줄링과 선점

### 선점형(Pre-emption)OS와 비선점형(Nonpre-emption)OS 차이

#### 원문: 선점형(Pre-emption)OS와 비선점형(Nonpre-emption)OS 차이

<!-- curriculum-chunk: sha256=cb78802b01debcbf551954cb6cf3f781871cfd65b5b1fdcf52d569623386fba4 major=os-kernel-computer-architecture mid=스케줄링과 선점 sub=선점형(Pre-emption)OS와 비선점형(Nonpre-emption)OS 차이 sources=source/interview_questions.md:5285-5286, source/interviews.md:5285-5286 -->

> Source: `source/interview_questions.md:5285-5286`
> Classification reason: scheduling
> Duplicate source aliases: `source/interview_questions.md:5285-5286, source/interviews.md:5285-5286`

##### 선점형(Pre-emption)OS와 비선점형(Nonpre-emption)OS 차이

<!-- /curriculum-chunk -->

## 프로세스 생성과 실행

### `&`와 background process

#### 원문: `&`와 background process

<!-- curriculum-chunk: sha256=97ffd8acfc1c17b7bf58b993c4f1169e9d3fb1fd5eca01e197caa0b3c71039e6 major=os-kernel-computer-architecture mid=프로세스 생성과 실행 sub=`&`와 background process sources=source/interview_questions.md:9214-9215, source/interviews.md:9162-9163 -->

> Source: `source/interview_questions.md:9214-9215`
> Classification reason: process execution
> Duplicate source aliases: `source/interview_questions.md:9214-9215, source/interviews.md:9162-9163`

##### `&`와 background process

<!-- /curriculum-chunk -->

### exec

#### 원문: exec

<!-- curriculum-chunk: sha256=8bdc10ff86de0e60c5a1cf20556e3b379bfae7eb9fdddce1dc57dc69628f9fc2 major=os-kernel-computer-architecture mid=프로세스 생성과 실행 sub=exec sources=source/interview_questions.md:5635-5646, source/interviews.md:5635-5646 -->

> Source: `source/interview_questions.md:5635-5646`
> Classification reason: process execution
> Duplicate source aliases: `source/interview_questions.md:5635-5646, source/interviews.md:5635-5646`

##### exec

`exec`와 `fork`는 서로 다른 시스템 호출로, 역할과 동작 방식에 차이가 있습니다.
`fork`는 현재 프로세스를 복사하여 새로운 자식 프로세스를 생성하는 반면, `exec`은 *현재 프로세스의 메모리 공간을 대체하는 시스템 호출*합니다.

- `exec()`은 현재 프로세스에서 호출된 프로그램을 새로운 프로그램으로 대체합니다. 즉, `exec()`가 호출되면 현재 프로세스의 메모리 공간, 코드, 데이터 섹션 등이 모두 새로운 프로그램으로 교체됩니다.
- `exec()`은 새로운 프로세스를 생성하지 않으며, 기존 프로세스가 실행 중이던 메모리 공간을 완전히 덮어씁니다. 프로세스 ID(PID)는 그대로 유지되지만, 새로운 프로그램이 해당 프로세스를 덮어쓰고 실행됩니다.

`exec()`가 호출되면, 기존 프로세스의 메모리 공간은 완전히 새로운 프로그램으로 대체되므로, COW 방식이 적용되지 않습니다.
`exec()` 호출은 메모리 페이지를 복사하는 것이 아니라, 기존의 메모리 맵을 지우고 새로운 메모리 맵을 로드하기 때문에, 기존의 부모 프로세스 메모리를 재사용하지 않습니다.
따라서 `exec()` 호출 후에는 이전 프로세스의 메모리 상태가 남아 있지 않고, 새로운 프로그램이 해당 프로세스의 메모리 공간을 차지하게 됩니다.

<!-- /curriculum-chunk -->

### fork

#### 원문: fork

<!-- curriculum-chunk: sha256=f2137b84b0417cd2e4cea1d1393e9c3125f6cba7b3163127e34e1a4cebcdb75d major=os-kernel-computer-architecture mid=프로세스 생성과 실행 sub=fork sources=source/interview_questions.md:5615-5634, source/interviews.md:5615-5634 -->

> Source: `source/interview_questions.md:5615-5634`
> Classification reason: process execution
> Duplicate source aliases: `source/interview_questions.md:5615-5634, source/interviews.md:5615-5634`

##### fork

- `fork()`는 현재 프로세스의 복사본(자식 프로세스)을 생성합니다. 자식 프로세스는 부모 프로세스와 동일한 메모리 공간을 공유하되, Copy-On-Write(COW) 방식을 사용해 메모리 페이지가 실제로 복사되는 시점은 메모리 변경이 발생할 때입니다.
- `fork()`는 새로운 프로세스를 생성하는 것만을 수행하며, 자식 프로세스는 부모 프로세스와 동일한 코드를 실행합니다. 이때 부모와 자식은 별도의 프로세스로 나뉘어 각자 독립적으로 실행됩니다.

`fork`는 현재 실행 중인 부모 프로세스(Apache 마스터 프로세스)를 COW(Copy-On-Write) 방식으로 복사하여 자식 프로세스를 만듭니다.
`fork` 호출 시, 자식 프로세스는 부모 프로세스의 메모리 공간을 그대로 공유합니다.
즉, 자식 프로세스는 부모 프로세스와 동일한 메모리 맵을 가지고 있지만, 메모리 페이지는 실제로 복사되지 않습니다.
이때의 복사는 논리적인 복사로, 실제 메모리 페이지는 자식 프로세스가 해당 페이지를 수정하려고 할 때까지 물리적으로 복사되지 않습니다.

COW 방식으로 동작하므로, 자식 프로세스가 메모리 내용을 수정할 때까지는 부모와 자식 프로세스가 동일한 메모리 페이지를 공유합니다.
그러나 자식 프로세스가 메모리를 수정하려고 하면, 해당 메모리 페이지가 실제로 복사되어 자식 프로세스에서만 수정됩니다.
자식 프로세스가 처음 생성되었을 때는 Apache 마스터 프로세스의 메모리 상태를 공유합니다.
하지만 PHP를 처리하는 동안 PHP 코드가 실행되면서 메모리 수정이 일어나면, 해당 메모리 페이지는 COW에 의해 독립적으로 복사됩니다.

Apache 마스터 프로세스는 미리 설정된 수만큼의 자식 프로세스를 생성합니다.
이때, 자식 프로세스는 Apache 마스터 프로세스의 상태를 그대로 상속받습니다.
이 프로세스들은 PHP를 처리하기 위한 프로세스로 동작합니다.
mod_php 모듈을 사용할 경우, PHP 인터프리터가 Apache 마스터 프로세스에 이미 로드되어 있으면, 자식 프로세스가 부모 프로세스의 메모리 공간을 상속하게 됩니다.

<!-- /curriculum-chunk -->

### java 프로그램이 프로세스로 실행되는 과정: 시스템 레벨부터 JVM까지

#### 원문: java 프로그램이 프로세스로 실행되는 과정: 시스템 레벨부터 JVM까지

<!-- curriculum-chunk: sha256=3dd949b9fc91a70e74ea140bff2938b64c3227d11f9b00f4d52dea9ae11f9596 major=os-kernel-computer-architecture mid=프로세스 생성과 실행 sub=java 프로그램이 프로세스로 실행되는 과정: 시스템 레벨부터 JVM까지 sources=source/interview_questions.md:5647-6498, source/interviews.md:5647-6498 -->

> Source: `source/interview_questions.md:5647-6498`
> Classification reason: process execution
> Duplicate source aliases: `source/interview_questions.md:5647-6498, source/interviews.md:5647-6498`

##### java 프로그램이 프로세스로 실행되는 과정: 시스템 레벨부터 JVM까지

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

<!-- /curriculum-chunk -->

---

아래 항목들은 현재 원문 배치 내용과 별개로 앞으로 정식 답변 자산으로 준비할 중주제 후보입니다.
이미 정리된 원문은 위에 그대로 두고, 누락 위험을 줄이기 위해 OS, 커널, 하드웨어, 실행 파일, 관측, 운영 질문까지 넓게 펼쳐 둡니다.

## 면접에서 OS 질문을 30초 안에 답하는 기본 구조

정리 대상: 짧은 직답, 핵심 구분, 아래층으로 내려가는 첫 기술 단위, 꼬리 질문을 받았을 때의 전개 순서.

## 컴퓨터는 전원이 들어온 뒤 첫 명령을 어떻게 실행하는가

정리 대상: 전원 공급, reset vector, firmware, BIOS/UEFI, bootloader, kernel entry까지 이어지는 첫 실행 경로.

## BIOS와 UEFI는 무엇이 다르고 왜 UEFI가 등장했는가

정리 대상: legacy BIOS 한계, GPT, EFI System Partition, Secure Boot, firmware와 OS 사이의 계약.

## 부트로더는 커널을 어떻게 찾아 메모리에 올리는가

정리 대상: GRUB, boot entry, kernel image, kernel parameter, initramfs, root filesystem 전환.

## 커널은 부팅 직후 어떤 순서로 자기 자신을 초기화하는가

정리 대상: 압축 해제, 초기 page table, interrupt 설정, scheduler 준비, init 프로세스 실행.

## systemd는 왜 PID 1로 실행되고 무엇을 책임지는가

정리 대상: init system, service dependency, target, unit, socket activation, process reaping.

## 데몬은 일반 프로세스와 무엇이 다르고 어떻게 관리되는가

정리 대상: session, controlling terminal, background service, restart policy, log, signal handling.

## CPU는 명령어를 어떤 주기로 실행하는가

정리 대상: fetch-decode-execute cycle, program counter, register, ALU, pipeline의 기본 흐름.

## ISA와 마이크로아키텍처는 어떻게 다른가

정리 대상: x86-64, ARM64 같은 명령어 집합과 실제 CPU 내부 구현의 분리.

## 레지스터는 왜 필요하고 메모리보다 왜 빠른가

정리 대상: general-purpose register, instruction operand, stack pointer, frame pointer, syscall register convention.

## CPU 캐시는 왜 필요하고 어떤 계층으로 구성되는가

정리 대상: L1/L2/L3, cache line, locality, cache miss, 백엔드 성능 설명에서 캐시가 등장하는 지점.

## 캐시 일관성은 멀티코어에서 어떤 문제를 해결하는가

정리 대상: core별 cache, coherence protocol, invalidation, false sharing, memory barrier와의 연결.

## 메모리 계층 구조는 성능 병목을 어떻게 만든다

정리 대상: register, cache, RAM, SSD, network storage 사이의 지연 시간 차이와 설계 판단.

## 정수는 컴퓨터 안에서 어떻게 표현되고 overflow는 왜 생기는가

정리 대상: two's complement, signed/unsigned, overflow, underflow, 언어별 정수 연산 차이.

## 부동소수점은 왜 정확한 십진수가 아니고 언제 위험한가

정리 대상: IEEE 754, binary fraction, rounding, NaN, infinity, 금융/정산 코드에서의 위험.

## 엔디언은 무엇이고 네트워크 바이트 오더와 어떻게 연결되는가

정리 대상: little-endian, big-endian, serialization, packet parsing, binary protocol.

## alignment와 padding은 메모리 구조에 어떤 영향을 주는가

정리 대상: CPU 접근 단위, struct layout, padding, cache line, native interop에서의 위험.

## 사용자 모드와 커널 모드는 왜 분리되어 있는가

정리 대상: privilege level, protection ring, unsafe instruction, kernel boundary, 보안과 안정성.

## 시스템 콜은 함수 호출과 무엇이 다른가

정리 대상: user mode to kernel mode 전환, syscall ABI, trap instruction, context 저장과 복원.

## interrupt, trap, exception은 어떻게 구분하는가

정리 대상: 외부 장치 interrupt, 의도적 trap, CPU exception, page fault, syscall과의 관계.

## 커널은 하드웨어 장치를 어떻게 추상화하는가

정리 대상: device driver, block device, character device, major/minor number, `/dev`.

## 프로세스는 실행 중인 프로그램 이상의 무엇인가

정리 대상: address space, file descriptor table, signal handler, credentials, resource limit, kernel task 구조.

## 프로세스 주소 공간은 어떤 구역으로 나뉘는가

정리 대상: text, data, bss, heap, stack, mmap region, shared library mapping.

## fork는 무엇을 복사하고 무엇을 공유하는가

정리 대상: process table entry, virtual memory, file descriptor, copy-on-write, parent-child 관계.

## exec는 왜 새 프로세스를 만드는 것이 아니라 현재 프로세스를 갈아끼우는가

정리 대상: execve, ELF loader, argv/envp, address space replacement, PID 유지.

## wait와 exit는 부모-자식 프로세스 관계를 어떻게 닫는가

정리 대상: exit status, zombie process, orphan process, reaping, PID 1의 역할.

## shell에서 `&`를 붙이면 실제로 무엇이 달라지는가

정리 대상: foreground/background job, process group, terminal control, job control signal.

## 프로세스와 스레드는 무엇을 공유하고 무엇을 따로 가지는가

정리 대상: address space, file descriptor, stack, register context, scheduling unit.

## 스레드가 많아지면 왜 context switch 비용이 문제가 되는가

정리 대상: CPU register save/restore, scheduler, cache pollution, run queue, latency.

## 선점형 스케줄링은 왜 필요하고 어떤 위험을 만든다

정리 대상: time slice, preemption, fairness, starvation, race condition, kernel preemption.

## 스케줄러는 어떤 기준으로 다음 실행 대상을 고르는가

정리 대상: runnable state, priority, nice, CFS, run queue, CPU affinity.

## load average는 CPU 사용률과 무엇이 다른가

정리 대상: runnable task, uninterruptible sleep, CPU saturation, I/O wait, 운영 관측 해석.

## 가상 메모리는 왜 필요하고 실제 메모리와 어떻게 연결되는가

정리 대상: virtual address, physical address, address translation, isolation, overcommit.

## MMU와 TLB는 주소 변환에서 어떤 역할을 하는가

정리 대상: page table walk, TLB hit/miss, context switch와 TLB flush.

## page table은 프로세스마다 왜 따로 필요한가

정리 대상: per-process address space, permission bit, user/kernel mapping, page table hierarchy.

## page fault는 항상 장애인가

정리 대상: demand paging, copy-on-write, swapped-out page, invalid access, SIGSEGV.

## copy-on-write는 fork 비용을 어떻게 줄이는가

정리 대상: shared physical page, write fault, page duplication, fork-heavy 서버와의 연결.

## swap은 메모리 부족을 어떻게 완화하고 왜 성능을 무너뜨릴 수 있는가

정리 대상: swap out/in, page reclaim, thrashing, memory pressure, latency spike.

## OOM Killer는 언제 프로세스를 죽이는가

정리 대상: overcommit, memory cgroup, oom_score, kill decision, container 환경의 차이.

## 커널 메모리와 사용자 메모리는 왜 분리해서 봐야 하는가

정리 대상: kernel heap, slab allocator, page cache, pinned memory, user copy.

## 메모리 할당기는 malloc 요청을 실제 메모리와 어떻게 연결하는가

정리 대상: brk, mmap, arena, fragmentation, allocator와 OS 경계.

## 파일 디스크립터는 파일이 아니라 무엇을 가리키는가

정리 대상: descriptor table, open file description, inode, offset, dup, close.

## inode, dentry, file object는 파일 시스템에서 어떻게 연결되는가

정리 대상: path lookup, directory entry, metadata, hard link, open file handle.

## VFS는 왜 여러 파일 시스템을 하나처럼 보이게 하는가

정리 대상: ext4, xfs, tmpfs, procfs, sysfs를 공통 API로 다루는 구조.

## page cache는 파일 I/O 성능을 어떻게 바꾸는가

정리 대상: read cache, write-back, dirty page, cache hit, 메모리 사용량 해석.

## `write()`가 성공하면 데이터는 디스크에 저장된 것인가

정리 대상: buffer, page cache, flush, fsync, fdatasync, durability와 성능 tradeoff.

## 파일 시스템 journaling은 어떤 장애를 막고 무엇을 보장하지 않는가

정리 대상: metadata consistency, crash recovery, ordered/writeback mode, application-level durability.

## block device와 character device는 어떻게 다른가

정리 대상: disk, terminal, pipe, socket, buffering, random access 가능 여부.

## I/O scheduler는 디스크 요청을 왜 재정렬하는가

정리 대상: HDD seek, SSD/NVMe queue, fairness, latency, throughput.

## DMA는 CPU 개입 없이 데이터를 어떻게 옮기는가

정리 대상: device memory access, interrupt completion, network/disk I/O, zero-copy의 기반.

## zero-copy는 실제로 어떤 복사를 줄이는가

정리 대상: user-kernel copy, sendfile, mmap, splice, NIC 전송 경로.

## blocking I/O는 스레드를 정확히 어디에서 멈추게 하는가

정리 대상: syscall wait, sleep state, wait queue, wakeup, scheduler handoff.

## non-blocking I/O는 blocking I/O와 무엇이 다른가

정리 대상: O_NONBLOCK, EAGAIN/EWOULDBLOCK, readiness, retry loop.

## select, poll, epoll은 왜 등장했고 무엇이 달라졌는가

정리 대상: fd set scan, interest list, ready list, level-triggered, edge-triggered.

## kqueue와 io_uring은 epoll과 어떤 관점에서 비교해야 하는가

정리 대상: BSD event queue, Linux submission/completion queue, readiness vs completion.

## 소켓은 파일 디스크립터인데 왜 네트워크 통신이 되는가

정리 대상: socket object, protocol stack, send/receive buffer, TCP state.

## 네트워크 패킷은 NIC에서 애플리케이션 버퍼까지 어떻게 이동하는가

정리 대상: interrupt, softirq, kernel network stack, socket buffer, copy to user.

## TCP 연결 상태는 커널 안에서 어떻게 관리되는가

정리 대상: SYN backlog, accept queue, ESTABLISHED, TIME_WAIT, keepalive.

## pipe와 redirect는 shell 명령 사이에 어떤 커널 객체를 만든다

정리 대상: anonymous pipe, stdin/stdout/stderr, file descriptor inheritance.

## signal은 프로세스에게 어떤 방식으로 비동기 이벤트를 전달하는가

정리 대상: SIGTERM, SIGKILL, SIGCHLD, signal mask, handler, async-signal safety.

## IPC는 프로세스 사이의 격리를 어떻게 우회해서 협력하게 하는가

정리 대상: pipe, Unix domain socket, shared memory, message queue, semaphore.

## mutex, spinlock, futex는 어느 계층의 동기화 도구인가

정리 대상: user-space lock, kernel wait, busy waiting, futex fast path/slow path.

## RCU는 커널에서 읽기 많은 자료구조를 어떻게 보호하는가

정리 대상: read-copy-update, grace period, lock-free read path, kernel data structure.

## atomic operation과 memory barrier는 왜 하드웨어 지식이 필요한가

정리 대상: compare-and-swap, instruction reordering, visibility, CPU memory model.

## NUMA는 서버 메모리 접근 비용을 어떻게 바꾸는가

정리 대상: socket, local/remote memory, NUMA node, CPU pinning, database 성능.

## SMT와 물리 코어는 성능 설명에서 어떻게 구분해야 하는가

정리 대상: hyper-threading, execution unit sharing, throughput, latency-sensitive workload.

## 서버용 하드웨어는 가정용 PC와 어떤 장애 모델이 다른가

정리 대상: ECC, redundant PSU, hot swap, RAID, BMC, IPMI, data center 운영.

## RAID와 백업은 왜 같은 말이 아닌가

정리 대상: disk redundancy, availability, accidental deletion, corruption, restore test.

## 가상화는 OS 위에 또 다른 OS를 어떻게 실행하게 하는가

정리 대상: hypervisor, VM, trap and emulate, hardware virtualization, guest kernel.

## container는 VM과 무엇이 다르고 커널을 어떻게 공유하는가

정리 대상: namespace, cgroup, image filesystem, process isolation, shared kernel.

## namespace는 프로세스가 보는 세계를 어떻게 분리하는가

정리 대상: PID, mount, network, user, IPC namespace.

## cgroup은 리소스 사용량을 어떻게 제한하고 관측하는가

정리 대상: CPU quota, memory limit, blkio, pids, container OOM.

## container에서 PID 1 문제는 왜 생기는가

정리 대상: signal forwarding, zombie reaping, init process behavior, Docker entrypoint.

## 권한 모델은 UID, GID, capability로 어떻게 나뉘는가

정리 대상: root privilege, setuid, file permission, Linux capability, least privilege.

## SELinux와 AppArmor는 일반 파일 권한과 무엇이 다른가

정리 대상: mandatory access control, policy, profile, denial log, 운영 장애 진단.

## ASLR, NX, stack canary는 실행 파일 보안을 어떻게 강화하는가

정리 대상: address randomization, non-executable memory, stack overwrite detection.

## ELF 실행 파일은 어떤 구조를 가지고 커널에 로딩되는가

정리 대상: ELF header, program header, section, segment, interpreter, dynamic linking.

## 동적 링커는 프로그램 시작 전에 무엇을 해석하는가

정리 대상: shared library, PLT/GOT, relocation, symbol resolution, `ld.so`.

## ABI와 calling convention은 함수 호출을 어떤 규칙으로 고정하는가

정리 대상: argument register, stack alignment, return value, syscall ABI, FFI/JNI.

## core dump는 프로세스가 죽을 때 어떤 정보를 남기는가

정리 대상: memory image, register, stack trace, ulimit, gdb 분석.

## strace는 시스템 콜 관점에서 프로그램을 어떻게 보여 주는가

정리 대상: syscall sequence, blocking point, errno, file/network troubleshooting.

## lsof는 열린 파일과 포트를 어떤 기준으로 보여 주는가

정리 대상: PID, FD, inode, socket, deleted file, port owner.

## perf는 CPU 병목을 어떤 샘플링 정보로 설명하는가

정리 대상: sampling profiler, call graph, kernel/user stack, CPU cycles.

## vmstat, iostat, pidstat는 각각 어떤 병목을 드러내는가

정리 대상: run queue, context switch, page fault, disk utilization, per-process CPU/I/O.

## 로그와 메트릭만으로 OS 병목을 오판하기 쉬운 경우는 무엇인가

정리 대상: CPU busy vs I/O wait, memory used vs page cache, load average vs throughput.

## Linux와 macOS, Windows의 OS 개념은 어디까지 공통이고 어디서 달라지는가

정리 대상: POSIX, syscall 차이, Mach/BSD, NT kernel, 개발자가 말할 수 있는 공통 추상화.

## POSIX는 운영체제 지식을 어떤 공통 인터페이스로 묶는가

정리 대상: process, file, signal, pipe, socket API, Unix-like portability.

## OS 지식은 Java, Spring, Netty 설명으로 어떻게 이어지는가

정리 대상: JVM process, native thread, file descriptor, socket, epoll, GC pause와 OS scheduling.

## OS 지식은 데이터베이스 성능 설명으로 어떻게 이어지는가

정리 대상: page cache, fsync, lock wait, context switch, NUMA, disk I/O.

## OS 지식은 시스템 아키텍처 답변에서 어떤 판단 축이 되는가

정리 대상: connection limit, process/thread model, resource isolation, failure domain, capacity planning.
