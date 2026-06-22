# Metaspace 누수 진단 — 클래스가 안 내려가는 이유와 추적법

[imports_linking_and_loading.md](./imports_linking_and_loading.md)에서 로드된 클래스 메타데이터는 힙이 아니라 네이티브 **Metaspace**에 올라간다고 했다. 이 영역이 차면 `OutOfMemoryError: Metaspace`가 난다. 그럼 클래스는 언제 *내려가는가*(unload)? 그 클래스를 정의한 ClassLoader가 통째로 GC될 때, 오직 그때뿐이다. 그래서 Metaspace 누수는 거의 언제나 **ClassLoader 누수**다. 이 노트는 그 메커니즘과 실전 진단 레시피를 정리한다.

## 목차

- [1. Metaspace는 무엇을 담고 언제 비워지나](#1-metaspace는-무엇을-담고-언제-비워지나)
- [2. 왜 ClassLoader 단위인가 (이론)](#2-왜-classloader-단위인가-이론)
- [3. 누수의 흔한 원인](#3-누수의-흔한-원인)
- [4. 증상](#4-증상)
- [5. 진단 레시피](#5-진단-레시피)
- [6. 흔한 수정](#6-흔한-수정)
- [7. 스스로 확인하기](#7-스스로-확인하기)
- [출처](#출처)
- [관련 노트](#관련-노트)

---

## 1. Metaspace는 무엇을 담고 언제 비워지나

Metaspace에는 로드된 클래스의 메타데이터가 들어간다 — klass 구조, 메서드 바이트코드, 상수 풀, 애너테이션 따위다. 힙이 아니라 네이티브(off-heap) 영역이고, Java 8에서 PermGen을 대체했다. 기본값은 상한이 없어 네이티브 메모리 한도까지 자라지만, `-XX:MaxMetaspaceSize`로 상한을 걸 수 있다.

이 메모리가 회수되는 단위는 클래스 하나가 아니라 ClassLoader 하나다. 어떤 ClassLoader가 GC되면 그 로더가 정의한 클래스가 *전부* 한꺼번에 언로드되고, 그만큼 Metaspace가 풀린다. 거꾸로 말하면 로더가 살아 있는 한 그 클래스는 한 개도 못 내려간다.

그런데 Bootstrap·Application 클래스로더는 사실상 프로세스가 끝날 때까지 산다. 그들이 로드한 JDK·앱 코어 클래스가 안 내려가는 건 그래서 정상이다. 실제로 내려갈 수 있는 건 *수거 가능한* 로더 — 웹앱 배포마다, 플러그인마다, 또는 런타임에 동적으로 만들어지는 로더 — 가 로드한 클래스뿐이다.

## 2. 왜 ClassLoader 단위인가 (이론)

참조 그래프가 한 덩어리로 묶여 있기 때문이다.

```
ClassLoader L
   ├─ defines → Class A ─┐
   ├─ defines → Class B  │  (각 Class 객체는 자신의 정의 ClassLoader L을 역참조)
   └─ defines → Class C ─┘
        ▲                    인스턴스 a:A, b:B ... 는 각자 Class를 참조
        └──────────── 서로가 서로를 붙들고 있음 ────────────
```

각 `Class` 객체는 자신을 정의한 ClassLoader를 거꾸로 참조하고, ClassLoader는 자신이 로드한 Class들을 참조한다. 인스턴스는 또 자기 Class를 붙든다. 셋이 서로를 물고 있는 한 덩어리인 셈이다. 그러니 로더·클래스·인스턴스 중 어느 하나라도 살아 있는 참조(GC root)에 걸리면, 그 로더가 정의한 클래스 메타데이터 *전체*가 도달 가능 상태로 남아 언로드되지 않는다.

[resource_management_and_leaks.md](./resource_management_and_leaks.md)에서 본 원리 그대로다. GC는 *도달 불가능*한 것만 회수하는데, ClassLoader 그래프는 한 군데만 붙들려도 전체가 도달 가능으로 묶인다.

## 3. 누수의 흔한 원인

**(A) ClassLoader 누수 — 가장 흔하다.** 새 로더와 그 클래스들이 무언가에 붙들려 GC되지 못하는 경우다.

전형은 앱서버 hot redeploy다. 재배포할 때마다 새 웹앱 ClassLoader가 생기는데, 옛 로더가 (아래에 적은 이유들로) 누수되면 옛 클래스가 안 내려간다. 그러면 재배포 한 번에 Metaspace가 한 층씩 쌓이다가 결국 OOM에 닿는다. 옛 로더를 붙드는 GC root는 대체로 정해져 있다.

- 살아 있는 스레드나 ThreadLocal. 공유 스레드풀의 스레드가 옛 앱의 인스턴스(또는 클래스)를 ThreadLocal에 담아 두고 비우지 않으면, 그 스레드가 사는 한 옛 로더도 따라 산다.
- 공유 라이브러리의 static 캐시가 옛 앱의 `Class`나 인스턴스를 키 또는 값으로 들고 있는 경우(예: `static Map<Class<?>, ...>`).
- JDBC 드라이버가 `DriverManager`에 등록된 채 언디플로이 때 해제되지 않는 경우. 드라이버 클래스 자체가 옛 로더 소속이라 같이 붙들린다.
- 앱이 띄운 스레드를 언디플로이 때 멈추지 않는 경우.

**(B) 끝없는 동적 클래스 생성.** 런타임에 클래스나 프록시를 계속 찍어내고 풀어주지 않는 경우다. CGLIB·ByteBuddy·동적 프록시를 고유 키마다 무한히 만들거나, 요청마다 새 ClassLoader를 띄우거나, 스크립팅·JSP 재컴파일이 쌓이는 식이다. 생성된 클래스(또는 그 로더)가 해제되지 않으면 Metaspace는 단조 증가한다.

## 4. 증상

- `java.lang.OutOfMemoryError: Metaspace`.
- 모니터링상 Metaspace 사용량이 *시간에 따라 계단식 상승*(재배포 횟수나 요청량과 상관).
- 로드된 클래스 수(loaded − unloaded)가 계속 증가.

## 5. 진단 레시피

**(0) 빨리 실패하게 상한을 건다.** 안 그러면 네이티브 메모리를 다 먹을 때까지 끌린다.

```
-XX:MaxMetaspaceSize=256m      # 누수면 예측 가능한 지점에서 일찍 OOM → 재현·덤프 쉬움
```

**(1) 정말 Metaspace가 자라는지 계측.**

```
jcmd <pid> VM.metaspace                 # Metaspace 상세 분해(사용/용량/로더별)
jstat -gc <pid> 1000                    # MC(메타capacity)/MU(메타used) 컬럼이 계속 오르는지
jstat -gcmetacapacity <pid> 1000
# 또는 JMX: MemoryPoolMXBean "Metaspace"
-XX:NativeMemoryTracking=summary  + jcmd <pid> VM.native_memory   # Metaspace가 자라는 네이티브 영역임을 확인
```

**(2) 클래스가 로드만 되고 언로드는 안 되는지 본다.**

```
-Xlog:class+load=info -Xlog:class+unload=info   # Java 9+ (예전엔 -verbose:class)
-Xlog:gc+metaspace=info
```

재배포/부하 후 *load는 쌓이는데 unload가 거의 없으면* 누수 신호.

**(3) 로더가 수거되는지 / 무엇이 붙드는지 — 힙 덤프로 확정.**

```
jmap -clstats <pid>                      # 클래스로더별 통계(살아있는/죽은 로더, 클래스 수)
jcmd <pid> GC.heap_dump /tmp/heap.hprof   # 힙 덤프
```

덤프를 **Eclipse MAT**(또는 VisualVM)로 열어 세 가지를 본다.

- 문제의 웹앱·플러그인 **ClassLoader 인스턴스가 몇 개**인지. 정상이면 1개여야 하는데 N개로 나오면 옛 로더가 안 죽고 남았다는 뜻이다.
- 그 누수된 ClassLoader에 **"Path to GC Roots"**(weak/soft 참조 제외)를 돌려 무엇이 붙들고 있는지 경로를 따라간다. 경로 끝에는 보통 Thread, ThreadLocal, 또는 어떤 static 필드가 앉아 있는데, 그게 범인이다.
- 같은 이름의 클래스가 여러 로더에 중복 로드돼 있지 않은지. 중복은 옛 로더가 살아남았다는 또 다른 신호다.

`GC.class_stats`(클래스별 메타데이터 크기)는 `-XX:+UnlockDiagnosticVMOptions`가 필요한 데다 버전에 따라 아예 빠져 있기도 하니, 위 흐름의 보조로만 쓴다.

**진단 흐름 요약:**

```
상한 걸기 → Metaspace 증가 계측(jcmd VM.metaspace/jstat)
   → load↑ unload≈0 확인(class+load/unload 로그)
      → 힙 덤프(jmap -clstats / GC.heap_dump)
         → MAT에서 "ClassLoader 인스턴스 N개" + "Path to GC Roots"로 붙드는 참조 추적
            → 그 참조(Thread/ThreadLocal/static)를 끊는다
```

## 6. 흔한 수정

- **붙드는 참조를 끊는다.** ThreadLocal은 다 쓰면 `remove()`로 비우고(풀에 반납할 때 정리), 언디플로이 때 앱이 띄운 스레드를 멈추고, JDBC 드라이버는 `DriverManager.deregisterDriver(...)`로 등록을 해제한다. 공유 라이브러리의 `static Map<Class,...>` 류 캐시에 앱 클래스가 갇히지 않게 하는 것도 같은 맥락이다.
- **동적 클래스 생성을 제한한다.** 만든 프록시·클래스는 캐시해 재사용하고(고유 키마다 새로 찍지 말 것), 요청마다 ClassLoader를 새로 만들지 않는다.
- **누수가 아니라 정말 클래스가 많은 경우**라면 `-XX:MaxMetaspaceSize`를 올린다.
- **앱서버 차원의 위생**도 있다. 많은 WAS가 leak-prevention 기능을 제공하고, 라이브러리 쪽은 언디플로이 때 등록 해제·스레드 정리를 지켜 줘야 한다.

## 7. 스스로 확인하기

- "클래스 하나만 안 쓰면 그 클래스가 Metaspace에서 내려가나?" → 아니다. *정의 ClassLoader 전체*가 도달 불가능해야 그 로더의 클래스가 *한꺼번에* 내려간다.
- 재배포마다 Metaspace가 계단식으로 오른다 — 가장 먼저 의심할 것은? (옛 웹앱 ClassLoader 누수.)
- 힙 덤프에서 누수 범인을 찾는 한 동작은? ("Path to GC Roots"로 죽었어야 할 ClassLoader를 붙드는 참조 추적.)
- native-image엔 왜 이 문제가 (런타임에) 없나? ([graalvm_native_image.md](./graalvm_native_image.md) — 클래스가 런타임에 로드/언로드되지 않고 빌드 때 굳혀지므로.)

## 출처

- HotSpot Metaspace(Java 8+ PermGen 대체), 클래스 언로드 = ClassLoader 수거: JVM/HotSpot 문서, JVM 명세 §5(로딩·언로딩).
- 진단 도구: `jcmd`(VM.metaspace, VM.native_memory, GC.heap_dump), `jstat`, `jmap -clstats`, `-Xlog:class+load/unload`, Eclipse MAT "Path to GC Roots".

## 관련 노트

- [imports_linking_and_loading.md](./imports_linking_and_loading.md) — 클래스가 *어떻게* 로드돼 Metaspace에 오나(이 노트의 전제).
- [resource_management_and_leaks.md](./resource_management_and_leaks.md) — GC 도달성 원리(왜 한 군데만 붙들려도 전체가 안 내려가나).
- [graalvm_native_image.md](./graalvm_native_image.md) — 클래스를 빌드 때 굳혀 이 문제 자체를 없애는 반대 접근.
- [java_jdk_tool.md](./java_jdk_tool.md) — 진단에 쓰는 `jcmd`/`jstat`/`jmap`이 속한 JDK 도구 모음.
- [../options.md](../options.md) — `-XX:MaxMetaspaceSize`·`-Xlog:class+load` 등 메모리·로깅 옵션.
