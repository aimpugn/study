# Low-Level Design 연습장

면접 LLD(객체지향·컴포넌트 설계) 문제를 **암기 카드 + 정답지 + 내 구현 + 공용 테스트**로 굳히는 공간.
문서를 늘리지 않는다. 외울 핵심은 각 인터페이스 파일 헤더(암기 카드)에만 둔다.

## 사용 루프

1. 문제의 인터페이스 파일 헤더(암기 카드)를 먼저 읽는다. 이것만은 외운다.
2. `My*.java`에 직접 구현한다. 막히기 전엔 `Reference*.java`를 열지 않는다.
3. 해당 `My*Test`의 `@Disabled`를 지우고 테스트로 같은 케이스를 통과시킨다.
4. 통과하면 정답지와 비교하고, 새로 깨달은 차이를 카드에 한 줄로 남긴다. 막히면 리뷰를 요청한다.

```bash
gradle test                          # 전체 (정답지는 통과, 내 구현은 시작 전이면 skip)
gradle test --tests '*MyLruCacheTest'
```

## 문제 목록 (우선순위 티어)

| 티어 | 문제 | 패키지 | 핵심으로 박을 것 | 상태 |
|---|---|---|---|---|
| 자료구조 | LRU 캐시 | `lld.lrucache` | HashMap+이중연결로 모든 연산 O(1) | ✅ 정답지+테스트 |
| 자료구조 | LFU 캐시 | `lld.lfucache` | 빈도 버킷+minFreq, 동률은 LRU로 | ✅ 정답지+테스트 |
| 자료구조 | Rate limiter | `lld.ratelimiter` | 토큰 버킷 lazy refill, 시계 주입 | ✅ 정답지+테스트 |
| 동시성 | 스레드 풀 | `lld.threadpool` | 큐+워커, 거부 정책(전략), poison pill 종료 | ✅ 정답지+테스트 |
| 동시성 | 생산자-소비자 | `lld.producerconsumer` | 락+두 Condition, while로 조건 재확인 | ✅ 정답지+테스트 |
| 모델링 | 주차장 | `lld.parkinglot` | 추상 크기+best-fit, OCP로 확장 | ✅ 정답지+테스트 |
| 패턴 | 로깅 라이브러리 | `lld.logging` | Formatter/Appender 전략 분리 | ✅ 정답지+테스트 |

> JVM·백엔드면 **자료구조·동시성** 티어가 면접에서 더 자주, 더 깊게 들어온다. 거기부터.

### 직접 변형해 볼 숙제 (정답지 없음)

같은 패턴 축이라 정답지를 따로 두지 않았다. 위 문제로 감을 잡은 뒤 직접 변형하면 가장 많이 남는다.

- **엘리베이터** — 주차장과 같은 모델링 축(상태 기계 + 스케줄링 전략 추가).
- **알림 시스템** — 로깅과 같은 전략/옵저버 축(채널별 Sender 전략 + 구독자 Observer).
- **Sliding window rate limiter** — `lld.ratelimiter`의 토큰 버킷을 슬라이딩 윈도우 카운터로.

## 공통 체크리스트 (어느 문제든)

- 연산별 시간·공간 복잡도를 한 줄로 말할 수 있는가 (왜 이 자료구조인가)
- 확장 요구("이제 X를 추가")에 어디만 고치면 되는지 답할 수 있는가 (개방-폐쇄 원칙)
- 상속 대신 합성을 쓴 자리와 그 이유
- 동시성에서 깨지는 지점과 막는 법
- 실무 대안(표준 라이브러리·검증된 구현)을 한 문장으로

## 바닥부터 다시 만들기 (live coding·과제 대비)

라이브 코딩이나 과제에서 "구현해 보세요"를 받으면 빈 디렉터리에서 시작할 수 있다.
알고리즘만이 아니라 **테스트가 도는 프로젝트를 빠르게 세우는 것**도 연습 대상이라, 이 스캐폴드를 맨바닥에서 재현하는 법을 적어 둔다.

**어떻게 만들어졌나**: LRU 하나를 `인터페이스(카드) + 정답지 + 내 구현 + 계약 테스트`의 여섯 파일로 먼저 완성하고, 같은 여섯 파일 구조를 문제마다 복제했다. 그래서 외울 것은 "문제 하나 = 여섯 파일 + 계약 테스트 패턴" 하나뿐이다.

### 1. 스택 결정 (30초)

Java 21 + JUnit 5 + Gradle(Kotlin DSL). 이유는 표준 빌드·테스트라 IDE·CI 어디서나 `gradle test` 한 줄로 끝나고, 테스트 의존성도 한 묶음이면 충분하기 때문이다.

### 2. 최소 파일 (이 둘만 외우면 프로젝트가 산다)

`settings.gradle.kts`:

```kotlin
rootProject.name = "low-level-design"
```

`build.gradle.kts`:

```kotlin
plugins { java }
repositories { mavenCentral() }
dependencies {
    testImplementation(platform("org.junit:junit-bom:5.11.0"))
    testImplementation("org.junit.jupiter:junit-jupiter")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}
java { toolchain { languageVersion = JavaLanguageVersion.of(21) } }
tasks.test { useJUnitPlatform() }
```

디렉터리는 `src/main/java/<패키지>/`(구현)와 `src/test/java/<패키지>/`(테스트), 그리고 `gradle test`.

**한 줄 스캐폴드 (도구·네트워크 있을 때).** 표준 템플릿 명령이 위 파일 + wrapper + 샘플을 한 번에 만든다. 요즘은 둘 다 JUnit 5를 기본 생성한다(2026-06 확인: Maven quickstart도 junit-jupiter + Java 17).

```bash
# Gradle: 이 프로젝트와 같은 구성(Kotlin DSL + JUnit 5 + Java 21)
gradle init --type java-library --test-framework junit-jupiter --dsl kotlin --java-version 21 --use-defaults

# Maven: quickstart archetype (현재 버전은 JUnit 5 BOM + Java 17 생성)
mvn archetype:generate -DarchetypeArtifactId=maven-archetype-quickstart -DinteractiveMode=false
```

생성물엔 샘플 클래스(Gradle은 version catalog도)가 딸려 오니 지우고 쓴다. 그래도 위의 손으로 쓴 최소 build 파일을 외워 두는 이유는 둘이다. (1) 도구·네트워크가 막힌 면접 환경(CoderPad 등) 대비, (2) 템플릿이 무엇을 만드는지 이해하고 직접 고치기 위해.

### 3. 핵심 재사용 트릭 — 계약 테스트 패턴

정답지와 내 구현을 **같은 테스트로** 검증하는 비결은 추상 테스트 + 팩토리 구멍 하나다.

```java
abstract class XContractTest {
    protected abstract X newX(int arg);          // 구현체가 바꿔 끼우는 구멍
    @Test void 동작() { X x = newX(2); /* ...단언... */ }
}
class ReferenceXTest extends XContractTest {      // 정답지 연결, 항상 초록
    protected X newX(int a) { return new ReferenceX(a); }
}
@Disabled("구현 시작하면 이 줄 삭제")
class MyXTest extends XContractTest {              // 내 구현 연결
    protected X newX(int a) { return new MyX(a); }
}
```

테스트 케이스는 한 번만 쓰고, 정답지와 내 구현이 그대로 같은 통과 대상이 된다.

### 4. 의존성이 없을 때 (화이트보드·제한된 환경)

JUnit·Gradle을 못 쓰는 자리면, 구현 클래스 옆에 `main` 하나로 즉시 검증한다.

```java
static void check(boolean ok, String name) {
    if (!ok) throw new AssertionError("FAIL: " + name);
    System.out.println("ok: " + name);
}
// main 안에서:
var c = new ReferenceLruCache<Integer, String>(2);
c.put(1, "a"); c.put(2, "b"); c.put(3, "c");
check(c.get(1) == null, "LRU 제거");
check("b".equals(c.get(2)), "값 유지");
```

`javac *.java && java <메인클래스>`. 프레임워크 0개로 "돌아간다"를 보인다.
