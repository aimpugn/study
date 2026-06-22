# import는 무엇을 하나 — 링킹·로딩·프로세스 메모리 (Go vs JVM vs Python)

"안 쓰는 import가 메모리를 먹나? Go는 왜 가볍지? 라이브러리는 프로세스의 어디에 올라가지?" — 이 혼동은 사실 **세 가지 다른 일을 하나로 뭉쳐서** 생긴다. 분리하면 한 번에 풀린다.

1. **import** — *소스에서 짧은 이름을 어떤 정의로 해소할지*. 대개 컴파일 타임의 이름 문제다.
2. **링킹(linking)** — *여러 코드 조각(심볼)을 하나의 실행 이미지로 결합*. 정적이면 빌드 시, 동적이면 로드/실행 시.
3. **로딩(loading)** — *코드·데이터를 실행 중 프로세스의 메모리에 실제로 올리는 것*.

언어마다 이 셋을 하는 *시점*이 다르다. 그래서 "import"라는 한 단어가 언어마다 전혀 다른 일을 가리킨다. 아래에서는 프로세스 메모리 구조부터 짚고, 그 위에서 Go(정적·AOT)·JVM(동적·지연)·Python(런타임 실행)을 같은 틀로 비교한다.

## 목차

- [1. 프로세스 메모리 — 코드와 데이터는 어디에 사나](#1-프로세스-메모리--코드와-데이터는-어디에-사나)
- [2. 분리하라 — import vs 링킹 vs 로딩](#2-분리하라--import-vs-링킹-vs-로딩)
- [3. Go — 정적 링크, AOT, 미사용 import는 컴파일 에러](#3-go--정적-링크-aot-미사용-import는-컴파일-에러)
- [4. JVM — 바이트코드, 동적·지연 클래스 로딩](#4-jvm--바이트코드-동적지연-클래스-로딩)
- [5. Python(대조) — import가 런타임에 실행된다](#5-python대조--import가-런타임에-실행된다)
- [6. 한눈 비교표](#6-한눈-비교표)
- [7. 처음 두 질문에 답하기](#7-처음-두-질문에-답하기)
- [8. 실무 함의](#8-실무-함의)
- [9. 스스로 확인하기](#9-스스로-확인하기)
- [출처](#출처)
- [관련 노트](#관련-노트)

---

## 1. 프로세스 메모리 — 코드와 데이터는 어디에 사나

직관("스택 말고 외부 코드/데이터가 올라가는 영역이 있다")은 맞다. 프로세스 가상 주소공간은 보통 이렇게 나뉜다.

```
 높은 주소  ┌───────────────────────────────┐
            │  stack  (호출 프레임·지역변수)    │   ↓ 아래로 성장
            ├───────────────────────────────┤
            │              ...               │
            │  mmap 영역                      │   공유 라이브러리(.so/.dll), mmap 파일,
            │  (shared libs, mmap, JIT 등)    │   런타임이 잡는 큰 블록
            │              ...               │
            ├───────────────────────────────┤
            │  heap   (동적 할당: malloc/new)  │   ↑ 위로 성장
            ├───────────────────────────────┤
            │  BSS    (0으로 초기화된 전역/정적) │   파일엔 크기만, 메모리에서 0으로 채움
            ├───────────────────────────────┤
            │  data   (값 있는 전역/정적)       │
            ├───────────────────────────────┤
            │  rodata (상수·문자열 리터럴)      │   읽기전용
            ├───────────────────────────────┤
 낮은 주소  │  text   (기계어 코드)            │   읽기전용·실행가능
            └───────────────────────────────┘
```

구체적으로 어느 칸인지:

| 코드/데이터 | 들어가는 곳 |
|---|---|
| 함수의 기계어 | text |
| `static const char* s = "hi";` 의 `"hi"` | rodata |
| `int g = 7;` (전역) | data |
| `int g;` / `static int n;` (0 초기화 전역) | BSS |
| `malloc(...)` / `new ...` 가 준 메모리 | heap |
| 함수 안 지역변수, 리턴주소 | stack |
| `libc.so`, `libssl.so` 같은 동적 라이브러리 | mmap 영역에 매핑 |

이론적 근거: 실행 파일 포맷(ELF/PE/Mach-O)이 이 **세그먼트(섹션)** 를 기술하고, `exec` 시 로더가 각 세그먼트를 페이지 테이블로 주소공간에 매핑한다. "라이브러리가 올라가는 곳"이란 정적이면 바이너리의 text/data, 동적이면 mmap된 공유 라이브러리 영역이다.

> **"네이티브 코드"·"네이티브 라이브러리"란?** 여기서 *네이티브*란 JVM 바이트코드가 아니라 **OS가 CPU에서 직접 실행하는 기계어**(보통 C/C++로 작성)를 말한다. 운영체제의 시스템 콜을 직접 부르며 메모리 할당·스레드 관리·GC·파일/네트워크 I/O 같은 저수준 일을 한다. **JVM 자체가 그런 네이티브 프로그램이다** — `java` 런처가 실제로 띄우는 알맹이는 **libjvm**(리눅스 `libjvm.so`, macOS `libjvm.dylib`, 윈도 `jvm.dll`)이라는 네이티브 동적 라이브러리이고, 바이트코드 인터프리터·JIT·GC·클래스로더가 모두 그 안의 C/C++ 코드다. 그래서 위 그림에서 `libjvm`·`libc` 같은 공유 라이브러리가 mmap 영역에 올라가고, 자바 코드는 이 네이티브 토대 *위에서* 돈다. (OS가 `libjvm`을 어떻게 찾아 올리는지 — 동적 링커·`NEEDED`·재배치 — 는 [graalvm_native_image.md](./graalvm_native_image.md)와 예제 문서 [GRAAL.md](../examples/fully-portable-spring-boot/GRAAL.md)에서 깊게 다룬다.)

---

## 2. 분리하라 — import vs 링킹 vs 로딩

| 단계 | 무엇을 하나 | 결정 시점(언어별로 다름) | 다루는 자료구조 |
|---|---|---|---|
| **import** | 소스의 짧은 이름 → 정의 해소 | 보통 컴파일 타임 | 심볼 테이블, 패키지/네임스페이스 |
| **링킹** | 심볼/코드 조각을 한 이미지로 결합, 주소 확정 | 정적=빌드 시 / 동적=로드·실행 시 | 심볼 테이블, 재배치(relocation) 테이블 |
| **로딩** | 코드·데이터를 프로세스 메모리에 적재 | 정적=exec 시 일괄 / 동적=필요 시 | 페이지 테이블, (JVM)Metaspace, (Py)sys.modules |

핵심은 하나다. **import한다고 메모리에 올라가는 게 아니다.** import는 대개 *이름 문제*에서 끝나고, 실제로 메모리에 올리는 일은 링킹과 로딩이 한다. 언어마다 이 둘을 얼마나 일찍 또는 늦게 묶느냐가 차이의 전부다.

---

## 3. Go — 정적 링크, AOT, 미사용 import는 컴파일 에러

- **import = 컴파일 타임 이름 해소**(패키지 단위). 런타임에 실행되는 게 아니다.
- **미사용 import는 *컴파일 에러*** 다. 즉 "안 쓰는 import"는 존재 자체가 불가능하다.
- **정적 링크**: 컴파일러+링커가 *쓰이는* 패키지들과 Go 런타임을 **하나의 네이티브 바이너리**로 묶는다. 링커는 **dead-code elimination**으로 도달 불가능한 함수/심볼을 떼어낸다.
- **실행**: OS가 그 바이너리의 text/data/BSS를 주소공간에 매핑하고 바로 기계어를 실행한다. 추가 동적 로딩이 없다(순수 Go, CGO 미사용 기준).

```
[빌드]   main.go + fmt + runtime + ...
            │  컴파일 → 오브젝트 → 정적 링크 + DCE
            ▼
         app  (단일 ELF/PE 바이너리)
            ├ .text   : main + fmt + runtime 의 기계어 (도달 가능한 것만)
            ├ .rodata : 문자열·상수
            └ .data/.bss : 전역/정적

[실행 ./app]
         OS exec → .text/.data/.bss 를 주소공간에 매핑 → 즉시 실행
                   (다 들어있으므로 런타임 동적 로딩 없음)
```

구체 예시:

```go
package main
import "fmt"
func main() { fmt.Println("hi") }
```

```
$ go build -o app .
$ ls -la app           # 대략 1.5~2MB대 (버전마다 다름) — 런타임이 정적으로 들어있어 hello도 큼

# 안 쓰는 import를 넣으면 빌드 자체가 실패한다:
import "os"            →  ./main.go:4:2: "os" imported and not used

# 순환 import도 컴파일 단계에서 막힌다:
                       →  import cycle not allowed
```

데이터 구조 관점: 미사용 import는 컴파일러의 의미 분석에서 걸리고, DCE는 호출 그래프(call graph) 도달성으로 죽은 심볼을 제거한다. 그래서 "안 쓰는 라이브러리가 data 영역에 적재"되는 일은 Go에선 일어나지 않는다 — 미사용은 컴파일이 막고, 사용분도 안 닿는 부분은 링커가 솎는다.

> **흔한 오해 교정**: blank import `import _ "pkg"` 는 *허용되는* "미사용처럼 보이는" import다. 이름은 안 쓰지만 그 패키지의 `init()` 부수효과를 위해 *일부러* 링크해 넣는 것(예: DB 드라이버 등록). 즉 이건 의도적 적재이지 실수가 아니다.

---

## 4. JVM — 바이트코드, 동적·지연 클래스 로딩

- `javac`가 `*.java` → `*.class` **바이트코드**로 컴파일한다. **import문은 컴파일 후 사라진다.** `.class`에는 *실제로 쓰인* 타입만 **상수 풀(constant pool)** 에 FQN(내부형 `java/util/HashMap`)으로 남는다.
- 클래스는 **런타임에 *처음 능동적으로 쓰일 때*** ClassLoader가 **클래스패스**(디렉터리/JAR)에서 찾아 로드한다(lazy). import가 로딩을 트리거하지 않는다 — *사용*이 트리거한다.
- 로드된 것이 가는 곳:
  - **Metaspace**(네이티브, off-heap): 클래스 메타데이터(메서드 바이트코드, 상수 풀, 필드 정보 등).
  - **Heap**: 객체 인스턴스.
  - **Code Cache**(네이티브): JIT가 만든 기계어.
  - **Thread Stacks**: 스레드별 호출 프레임.

```
[빌드]  Foo.java ──javac──▶ Foo.class  (import 사라짐, 쓰인 타입만 상수풀에 FQN)

[실행 java App]
  JVM 프로세스 = java 런처 + libjvm (이건 그 자체로 네이티브 text/data를 가짐)
     │  "처음 쓰일 때" 클래스패스에서 클래스를 찾아 로드 (lazy)
     ▼
  ┌──────────── JVM이 관리하는 영역 ────────────┐
  │ Metaspace (네이티브) : 로드된 클래스 메타데이터 │
  │ Heap                : 객체 인스턴스           │
  │ Code Cache(네이티브) : JIT 기계어             │
  │ Thread Stacks       : 스레드별 프레임          │
  └────────────────────────────────────────────┘
  안 쓰는 클래스 → 로드 안 됨 → Metaspace/Heap 어디에도 안 올라옴
```

이론(JVM 명세의 클래스 수명): 한 클래스는 **로딩 → 링킹(검증·준비·해소) → 초기화** 단계를 거치는데, 이 과정은 **첫 능동적 사용**(인스턴스 생성, 정적 필드/메서드 접근, 리플렉션 등)에서 비로소 시작된다. 그래서 "참조조차 안 되는" 클래스는 단계 자체가 시작되지 않는다.

구체 예시:

```
# 1) 바이트코드엔 import가 없고 '쓰인' 타입만 상수풀에 있다
$ javap -v Foo.class
  Constant pool:
     #7 = Class    #8        // java/util/HashMap     ← 코드에서 쓴 타입
     #8 = Utf8     java/util/HashMap
     ...                                              ← import 문 같은 건 없음

# 2) 클래스는 '쓰는 순서대로' lazy 하게 로드된다 (Java 9+; 예전엔 -verbose:class)
$ java -Xlog:class+load=info App
  [info][class,load] App           source: file:.../
  [info][class,load] java.util.HashMap  source: shared objects file
  ...                              ← 안 쓰는 클래스는 이 목록에 안 나온다
```

데이터 구조 관점: `.class`의 *상수 풀*이 "이 클래스가 참조하는 다른 타입/필드/메서드" 목록이다. 미사용 import는 여기에 항목을 만들지 않으므로(코드에서 안 쓰였으니), 해소(resolution)될 대상이 없고 따라서 로드될 일도 없다. ClassLoader들은 **부모 위임 트리**(Bootstrap → Platform → Application)를 이루며, "어떤 클래스로더가 무엇을 로드했나"가 클래스 정체성의 일부다.

결론: **안 쓰는 import → 바이트코드에 참조 없음 → 그 클래스 영영 로드 안 됨 → 런타임 메모리 0.** 클래스패스에 JAR이 아무리 많아도 *쓰는 클래스만* 로드된다(큰 클래스패스 ≠ 큰 메모리).

---

## 5. Python(대조) — import가 런타임에 실행된다

여기가 많은 사람이 Java/Go에 잘못 투영하는 모델이다.

- Python `import x`는 **런타임에 실행**된다: 모듈을 찾고 → 그 모듈의 *최상위 코드를 실행*하고 → 이름을 바인딩하고 → `sys.modules` 딕셔너리에 캐시한다.
- 따라서 **미사용 import도 "실행"된다** — 모듈이 로드·실행되어 메모리(와 시간)를 *실제로* 쓴다.

```python
# mod.py
print("mod 로드됨")     # 모듈 최상위 코드 → import 시점에 실행된다
HEAVY = [0] * 10_000_000

# main.py
import mod              # 이 줄에서 "mod 로드됨" 출력 + HEAVY 가 메모리에 올라간다
                        # mod 의 무엇도 안 써도 그렇다
```

데이터 구조 관점: `sys.modules`(이미 로드된 모듈 캐시 dict)가 핵심이다. import는 이 캐시를 보고 없으면 모듈을 실행해 채운다. 그래서 Python에선 "안 쓰는 import 제거"가 *런타임* 의미가 있다(메모리·기동 시간). Java/Go와 정반대다.

---

## 6. 한눈 비교표

| | Go | JVM (Java) | Python |
|---|---|---|---|
| import의 정체 | 컴파일 타임 이름 해소 | 컴파일 타임 이름 해소 | **런타임 실행** |
| 링킹 | 정적(빌드 시 바이너리에) | 동적(런타임 클래스 로딩) | 런타임(모듈 실행·캐시) |
| 코드가 메모리에 오는 때 | `exec` 시 바이너리 일괄 매핑 | 클래스 **처음 쓰일 때** lazy | `import` 실행 시 모듈 |
| 미사용 import | **컴파일 에러**(존재 불가) | 무해(바이트코드에 없음) | **실행됨**(로드·캐시) |
| 미사용의 런타임 비용 | (존재 불가) | **0** | 메모리 + 시간 |
| 기동 비용 | 매우 낮음(네이티브 즉시 실행) | 높음(VM 초기화+코어 클래스+JIT) | 중간(인터프리터+import 실행) |
| 산출물 | 단일 네이티브 바이너리 | `.class`/JAR + JVM 필요 | `.py` + 인터프리터 필요 |

---

## 7. 처음 두 질문에 답하기

**Q. 안 쓰는 import가 메모리를 먹나?**
- Go: 애초에 컴파일이 막아 *존재 불가*.
- JVM: 무해. 바이트코드에 참조가 없어 클래스가 *로드되지 않음* → 0.
- Python: 먹는다. import가 실행되어 모듈이 적재됨.
- 즉 Java/Go에선 "안 쓰는 import 제거"가 *런타임 효율*과 무관하다(가독성·위생 목적). Python에선 런타임 의미가 있다.

**Q. Go가 경량인 이유가 "순환 참조를 컴파일에서 막아서"인가?**
- 순환 import를 막는 건 사실이지만(`import cycle not allowed`), 그 목적은 컴파일·링크를 단순하게 하고, `init` 순서를 결정적으로 만들고, 의존성 구조를 깨끗이 유지하려는 데 있다. 그래프가 깔끔하면 *간접적으로* 도움은 되지만, 경량의 핵심은 아니다.
- Go가 가벼운 진짜 이유는 따로 있다. AOT로 네이티브 코드까지 미리 컴파일하니 VM·JIT 워밍업이 없고, 정적 링크로 단일 바이너리가 나오니 런타임 동적 로딩도 별도 런타임 설치도 없다. 거기에 런타임 자체가 작고 기동이 빠르다.
- JVM이 "무겁다"고 느껴지는 진짜 지점도 미사용 import가 아니라 **기동 비용**이다 — VM 초기화, 코어 클래스 수백~수천 개 로드, JIT 워밍업. 그 대신 안 쓰는 코드는 애초에 올라오지도 않는다.

---

## 8. 실무 함의

- **optimize-imports(IDE)** 는 *성능*이 아니라 *위생*이다(Java/Go). 안 쓰는 import 제거·정렬·일관성으로 가독성과 diff 품질을 높이는 것이지 런타임 메모리와는 무관하다. (Python에선 미사용 import가 런타임 비용 + 린터 `F401` 대상이라 의미가 다르다.)
- **Go의 강점 활용처**: 단일 바이너리 + 즉시 기동 → 작은 컨테이너 이미지, serverless 콜드스타트 단축, CLI 단일 파일 배포.
- **JVM 기동 비용 완화책**: AppCDS(클래스 데이터 공유 — 클래스 메타데이터를 아카이브로 mmap해 기동 단축), GraalVM **native-image**(AOT로 Java를 네이티브 바이너리화 → Go처럼 즉시 기동), 프레임워크 지연 초기화.
- **Metaspace는 실무 함정**: 클래스/클래스로더가 누수되면(핫 리디플로이, 런타임 동적 클래스 생성 과다) `OutOfMemoryError: Metaspace`가 난다. "클래스 메타데이터는 힙이 아니라 네이티브 Metaspace"라는 사실이 이 진단의 출발점이다([resource_management_and_leaks.md](./resource_management_and_leaks.md)의 "GC는 힙만 책임진다"와 연결, 진단 레시피는 [metaspace_leak_diagnosis.md](./metaspace_leak_diagnosis.md)).

---

## 9. 스스로 확인하기

- "import가 메모리를 먹나?"를 물을 때, 그 언어가 import를 *컴파일 타임 이름 해소*로 쓰는지(Go/Java) *런타임 실행*으로 쓰는지(Python)를 먼저 구분했는가?
- JVM에서 클래스를 메모리에 올리는 트리거는 import인가 *사용*인가? (사용 — 첫 능동적 사용.)
- Go 바이너리가 hello-world인데도 큰(수 MB) 이유를 한 문장으로? (런타임을 정적으로 포함하므로.)
- "Go가 가볍다"를 설명할 때 순환참조 금지를 핵심으로 들면 무엇으로 바로잡아야 하나? (AOT+정적링크+작은 런타임+VM 기동 없음.)
- 클래스패스에 JAR 100개를 추가하면 메모리가 그만큼 느나? (아니다 — 쓰는 클래스만 로드된다.)

## 출처

- ELF/실행 파일 세그먼트, `exec` 매핑: OS·링커 일반 이론(예: ostep, 링커/로더 문서).
- JVM 클래스 수명(로딩·링킹·초기화)과 초기화 트리거: Java Virtual Machine Specification §5.
- Metaspace(Java 8+ PermGen 대체): HotSpot 문서.
- Go: "imported and not used" / "import cycle not allowed"는 컴파일러 진단, 정적 링크·dead-code elimination은 Go 링커 동작.
- Python import 시스템과 `sys.modules`: Python 언어 레퍼런스 "The import system".
- JVM 내부 동작 전반(클래스 로딩·GC·JIT·성능)의 심화 읽을거리: Aleksey Shipilëv, "JVM Anatomy Quarks" — <https://shipilev.net/jvm/anatomy-quarks/> .

## 관련 노트

- [resource_management_and_leaks.md](./resource_management_and_leaks.md) — 힙 vs 네이티브 자원, GC가 못 치우는 것(같은 "JVM 런타임이 무엇을 어떻게 다루나" 결).
- [graalvm_native_image.md](./graalvm_native_image.md) — JVM을 Go의 정적·AOT 모델로 옮기는 법(이 노트의 "동적 로딩"을 빌드 때 굳힘).
- [metaspace_leak_diagnosis.md](./metaspace_leak_diagnosis.md) — 동적 클래스 로딩이 Metaspace에 쌓이는 문제와 진단.
- [java_jdk_tool.md](./java_jdk_tool.md) — `javac`/`java`(11+ 소스 직접 실행 포함)·`javap` 등 JDK 도구.
- [../../computer_architecture/virtual_memory.md](../../computer_architecture/virtual_memory.md) — 프로세스 주소공간과 페이지 매핑.
- [../../computer_architecture/stack_frame.md](../../computer_architecture/stack_frame.md) — stack 영역의 호출 프레임.
