# GraalVM native-image — JVM을 Go처럼 만드는 법

[imports_linking_and_loading.md](./imports_linking_and_loading.md)에서 JVM은 **open-world**(런타임에 클래스를 동적·지연 로딩)이고 Go는 **closed-world**(빌드 때 다 정해 놓고 정적 링크)라고 정리했다. GraalVM의 `native-image`는 JVM 언어를 바로 그 closed-world로 바꾼다. 빌드 타임에 세계를 닫고 네이티브 실행 파일로 AOT 컴파일하는 것이다. 그 결과물은 Go 바이너리처럼 즉시 기동하고 가볍다. 물론 closed-world가 강요하는 제약이 그 대가로 따라붙는다. 이 노트는 *어떻게* 그렇게 되는지와 *무엇을 대가로 치르는지*를 개념 중심으로 정리한다. (libjvm·glibc·musl 정적 링킹, 동적 링커, 배포 전략까지 OS·툴체인 관점에서 예제로 파고드는 심화는 자매 문서 [GRAAL.md](../examples/fully-portable-spring-boot/GRAAL.md)에 있다.)

## 목차

- [1. 무엇인가](#1-무엇인가)
- [2. 어떻게 Go처럼 되나 — closed-world + AOT + 빌드타임 초기화](#2-어떻게-go처럼-되나--closed-world--aot--빌드타임-초기화)
- [3. 기동이 빠른 이유를 분해](#3-기동이-빠른-이유를-분해)
- [4. 대가 — closed-world가 강요하는 것](#4-대가--closed-world가-강요하는-것)
- [5. JVM vs native-image vs Go](#5-jvm-vs-native-image-vs-go)
- [6. 실무 — 언제 쓰나](#6-실무--언제-쓰나)
- [7. 스스로 확인하기](#7-스스로-확인하기)
- [출처](#출처)
- [관련 노트](#관련-노트)

---

## 1. 무엇인가

`native-image`는 Java 바이트코드와 거기서 도달 가능한 모든 의존성을 **독립 실행 네이티브 바이너리**로 AOT(Ahead-Of-Time) 컴파일하는 도구다. 실행할 때 **JVM이 필요 없다.** HotSpot의 바이트코드 인터프리터·JIT·클래스 로더가 들어가는 자리에, 컴파일된 기계어와 **Substrate VM(SVM)** 이라는 *작은* 런타임(GC·스레드 스케줄링 등)만 바이너리에 박혀 있다.

보통의 `java -jar app.jar`를 떠올려 보면 차이가 분명하다. 그쪽은 HotSpot JVM을 띄우고, 클래스를 런타임에 로드하고, JIT로 데워지며 돈다. native-image는 그 일을 전부 *빌드 때* 끝내 둔다.

## 2. 어떻게 Go처럼 되나 — closed-world + AOT + 빌드타임 초기화

핵심은 세 가지다.

**(1) Closed-world 도달성 분석.** 빌드 때 엔트리포인트에서 출발해 **points-to(도달성) 분석**으로 실제로 닿는 코드를 전부 찾아낸다. 닿는 것만 AOT 컴파일하고 나머지는 버린다 — Go 링커의 dead-code elimination과 같은 발상이다. 이 분석이 성립하려면 "온 세상"이 빌드 시점에 다 알려져 있어야 하고, 그게 바로 closed-world라는 말의 뜻이다.

**(2) 빌드타임 초기화와 image heap.** native-image는 클래스 초기화(static 초기화)를 *빌드 때* 미리 돌려, 그렇게 초기화된 객체 상태를 **image heap**이라는 형태로 바이너리의 데이터 영역에 구워 넣을 수 있다. 그러면 런타임에 기동할 때 그 초기화를 다시 할 필요가 없다. 다만 런타임에야 정해지는 값 — 난수 시드, 열린 파일이나 소켓 같은 것 — 은 빌드 때 굳히면 안 되니 런타임 초기화로 미룬다.

**(3) Substrate VM.** 풀 HotSpot 대신 AOT 코드와 함께 컴파일되는 *작은* 런타임이다. GC와 스레드 정도의 최소 기능만 담는다.

```
[빌드: native-image app.jar]
  app.jar + 의존성
      │  ① points-to 분석 (도달 가능한 코드만 선별)
      │  ② 도달 코드 AOT 컴파일 → 기계어
      │  ③ build-time 클래스 초기화 → image heap(초기화된 객체)
      ▼
  ┌──────────────────────────────────────┐
  │  네이티브 실행 파일  app                 │
  │   ├ text   : AOT 기계어                 │
  │   ├ image heap : 미리 초기화된 객체 상태  │
  │   └ Substrate VM : 작은 런타임(GC·스레드) │
  └──────────────────────────────────────┘

[실행: ./app]
  exec → 세그먼트 매핑 → 즉시 실행
  (VM 부트스트랩 X, 런타임 클래스 로딩 X, JIT 워밍업 X)
```

이 그림을 [imports_linking_and_loading.md](./imports_linking_and_loading.md)의 Go 그림과 겹쳐 보면 거의 같다 — native-image가 JVM 언어를 Go의 "정적·AOT·exec하면 끝" 모델로 옮긴 것이다.

## 3. 기동이 빠른 이유를 분해

보통 JVM 기동을 무겁게 만드는 네 가지가 native-image에는 *없다*.

1. **VM 부트스트랩이 없다.** libjvm 초기화나 힙 구조 셋업 같은 준비 과정이 필요 없다.
2. **런타임 클래스 로딩이 없다.** 클래스가 이미 AOT 컴파일돼 image heap에 구워져 있으니, 런타임에 클래스패스를 스캔하고 `defineClass`로 정의하고 검증하는 단계가 통째로 빠진다.
3. **JIT 워밍업이 없다.** 인터프리트 → 프로파일 → 재컴파일로 이어지는 사이클이 없다. 처음부터 기계어로 돈다.
4. **초기화가 미리 반영돼 있다.** static 초기화가 빌드 때 끝나 image heap에 들어가 있다.

그래서 결과가 어떻게 달라지나. 워크로드와 버전마다 다르지만 대략, 기동이 수백 ms에서 초 단위였던 것(JVM에서 도는 Spring이 그렇다)이 **수~수십 ms**로, 상주 메모리도 수백 MB대에서 **수십 MB대**로 떨어진다. Go 네이티브 바이너리에서 보던 것과 같은 결이다.

## 4. 대가 — closed-world가 강요하는 것

공짜가 아니다. "온 세상을 빌드 때 안다"는 가정 자체가 동적 기능과 정면으로 부딪힌다.

가장 먼저 걸리는 게 **리플렉션·동적 프록시·JNI·리소스·직렬화**다. 이것들은 정적 분석에 잡히지 않는다. 그래서 어떤 클래스를 리플렉션으로 부를지를 **reachability metadata**(예: `reflect-config.json`)로 직접 *선언*하거나, **tracing agent**로 실제 실행을 추적해 그 설정을 뽑아내야 한다. 빠뜨리면 빌드는 통과해도 런타임에 `ClassNotFoundException` 류로 터진다.

```
# 실행을 추적해 reflection/resource 설정을 자동 생성
$ java -agentlib:native-image-agent=config-output-dir=META-INF/native-image -jar app.jar
# 그 설정을 안고 네이티브 빌드
$ native-image -jar app.jar
```

나머지 대가는 이렇다.

- **런타임 동적 클래스 로딩이 (대체로) 안 된다.** 빌드 때 몰랐던 클래스를 나중에 끌어다 로드하는 플러그인 패턴이 깨진다.
- **빌드가 비싸다.** 정적 분석 때문에 네이티브 빌드는 분 단위로 느리고 메모리도 많이 먹는다.
- **피크 처리량이 낮을 수 있다.** AOT 코드는 런타임 프로파일을 보고 최적화하는 JIT만큼 피크 throughput을 내지 못할 수 있다. 이 격차는 Oracle GraalVM의 **PGO**(프로파일 기반 최적화)로 어느 정도 메운다. 그래서 짧게 살고 기동에 민감한 쪽 — CLI, serverless, 스케일이 잦은 마이크로서비스 — 은 네이티브가 유리하고, 오래 돌며 throughput을 최대로 짜내야 하는 서버는 JIT를 쓰는 JVM이 여전히 나을 수 있다.
- **GC 선택지가 좁을 수 있다.** HotSpot보다 고를 수 있는 GC가 제한적이다(에디션에 따라 Serial/G1 등). GC별 특성과 선택 기준은 [../gc.md](../gc.md)에 있다.

## 5. JVM vs native-image vs Go

| | HotSpot JVM | GraalVM native-image | Go |
|---|---|---|---|
| 모델 | open-world, 동적 | **closed-world, 정적/AOT** | closed-world, 정적/AOT |
| 코드 형태 | 바이트코드 → 런타임 JIT | 빌드 때 AOT 기계어 | 빌드 때 AOT 기계어 |
| 클래스 로딩 | 런타임 지연 로딩 | 빌드 때 굳힘(런타임 로딩 없음) | (해당 없음) 정적 링크 |
| 기동 | 무거움(VM+로딩+JIT) | **가벼움(ms급)** | 가벼움(ms급) |
| 메모리 | 큼 | 작음 | 작음 |
| 리플렉션/동적 | 자유로움 | **설정 필요(closed-world)** | 제한적(언어 차원) |
| 빌드 | 빠름 | 느림·메모리多 | 빠름 |
| 피크 throughput | 높음(JIT 프로파일) | 보통(PGO로 보완) | 높음 |

요점: native-image는 JVM의 *유연함(open-world)* 을 포기하는 대신 Go의 *가벼움·즉시성* 을 얻는 거래다.

## 6. 실무 — 언제 쓰나

- **프레임워크 지원이 판을 바꿨다.** Spring Boot 3(Spring AOT)·Quarkus·Micronaut가 reachability metadata 생성과 빌드타임 처리를 자동화하면서 native-image가 실용 영역으로 들어왔다. 특히 Quarkus와 Micronaut는 처음부터 native를 염두에 두고 만들어졌다.
- **잘 맞는 곳:** 콜드 스타트를 줄여야 하는 serverless, 단일 바이너리로 뿌리는 CLI, 짧게 살고 자주 뜨는 컨테이너, 메모리가 빡빡한 환경.
- **덜 맞는 곳:** 리플렉션·동적 로딩이 헤비한 레거시, 네이티브 빌드 비용을 감당 못 하는 빌드 파이프라인, throughput 최대화가 목표인 장수 서버(여기선 JIT가 더 낫다).
- **네이티브까지 안 가도 되는 완화책**도 있다. **AppCDS**는 클래스 데이터를 공유한다 — 메타데이터 아카이브를 mmap해서 로딩·검증을 줄이는 식이다. 프레임워크의 지연 초기화도 한 방법이다. 네이티브는 이 방향을 "끝까지 간" 버전인 셈이다.

## 7. 스스로 확인하기

- native-image가 JVM 기동의 무거운 네 가지(VM 부트스트랩·런타임 클래스 로딩·JIT 워밍업·런타임 초기화) 중 무엇을 각각 어떻게 없애는지 말할 수 있나?
- "closed-world 가정"이 왜 리플렉션 설정을 요구하나? (정적 분석에 안 보이므로.)
- 단명 serverless와 장수 throughput 서버 중 네이티브가 유리한 쪽과 그 이유는?
- native-image엔 왜 Metaspace 개념이 (런타임에) 없나? (클래스가 런타임에 로드되지 않고 빌드 때 굳혀지므로 — [metaspace_leak_diagnosis.md](./metaspace_leak_diagnosis.md)와 대비.)

## 출처

- GraalVM Native Image 문서: closed-world assumption, reachability metadata, build-time initialization, Substrate VM, tracing agent.
- Oracle GraalVM PGO(Profile-Guided Optimizations) 문서.
- Spring Boot 3 / Spring AOT, Quarkus, Micronaut native 가이드.

## 관련 노트

- [imports_linking_and_loading.md](./imports_linking_and_loading.md) — JVM(open-world 동적) vs Go(정적 AOT). 이 노트는 "JVM을 그 정적 모델로 옮기는 법".
- [metaspace_leak_diagnosis.md](./metaspace_leak_diagnosis.md) — 동적 클래스 로딩이 만드는 Metaspace 문제. 네이티브는 이 문제 자체가 없어지는 대신 유연성을 잃는다.
- [resource_management_and_leaks.md](./resource_management_and_leaks.md) — 힙 vs 네이티브 자원.
- [GRAAL.md](../examples/fully-portable-spring-boot/GRAAL.md) — 같은 주제의 **예제 기반 심화**: libjvm·glibc·musl 정적 링킹, 동적 링커(`ld` vs `ld-linux`), 플러그인 설계, `java -jar` 대비 배포 전략 등 OS·툴체인 관점.
- [../gc.md](../gc.md) — native-image가 쓸 수 있는 GC(에디션별 제한)를 포함한 GC 선택.
