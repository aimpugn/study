# Grafana JVM 대시보드(Micrometer) 총정리 — 계기판으로 배우는 JVM의 내부

Spring Boot 서비스에 흔히 붙이는 Grafana 공식 대시보드 **"JVM (Micrometer)" (ID 4701)** 의 모든 패널을, 그 숫자가 JVM 내부 어디에서 왜 어떻게 만들어지는지까지 한 권으로 정리한다.
대시보드 읽는 법(운영)과 JVM 동작 원리(이론)를 한 줄기로 엮어서, 패널 하나를 이해하면 JVM의 한 구역이 같이 이해되도록 구성했다.

기준 환경:

- 메트릭 파이프라인: Spring Boot(Micrometer) → Prometheus → Grafana 대시보드 4701
- JVM: HotSpot, JDK 9 이상(캡처에 G1 + 분할 CodeHeap이 보이므로 확실), GC는 G1
- 캡처된 실측값: 4 vCPU, 힙 512 MiB 고정으로 도는 소형 서비스의 2026-06 시점 스냅숏 — 본문 예시와 11장 진단에 그대로 사용

관련 문서:

- [gc.md](./gc.md) — GC 종류별 특성과 선택 기준
- [options.md](./options.md) — JVM 옵션 사전
- [thread_leak.md](./thread_leak.md) — 스레드 누수 진단(live threads 패널이 우상향할 때)
- [java/metaspace_leak_diagnosis.md](./java/metaspace_leak_diagnosis.md) — Metaspace 패널이 계단식으로 오를 때(클래스로더 누수)
- [java/resource_management_and_leaks.md](./java/resource_management_and_leaks.md) — 힙 vs 외부 자원, GC의 책임 범위

목차:

- [0. 이 문서를 읽는 법, 그리고 결론부터](#0-이-문서를-읽는-법-그리고-결론부터)
- [1. 숫자가 화면에 도달하기까지 — 측정 파이프라인](#1-숫자가-화면에-도달하기까지--측정-파이프라인)
- [2. JVM이라는 도시 — 전체 지도](#2-jvm이라는-도시--전체-지도)
- [3. I/O Overview — 도시의 정문](#3-io-overview--도시의-정문)
- [4. JVM Memory — 살림살이 한눈에](#4-jvm-memory--살림살이-한눈에)
- [5. JVM Misc — 활력 징후](#5-jvm-misc--활력-징후)
- [6. JVM Memory Pools (Heap) — 객체의 일생과 G1](#6-jvm-memory-pools-heap--객체의-일생과-g1)
- [7. JVM Memory Pools (Non-Heap) — 설계도 보관소와 번역청](#7-jvm-memory-pools-non-heap--설계도-보관소와-번역청)
- [8. Garbage Collection — 청소국의 작업 일지](#8-garbage-collection--청소국의-작업-일지)
- [9. Classloading — 설계도는 어떻게 반입되는가](#9-classloading--설계도는-어떻게-반입되는가)
- [10. Buffer Pools — 성벽 밖 보세창고](#10-buffer-pools--성벽-밖-보세창고)
- [11. 실측 진단 — 이 대시보드의 숫자 읽기](#11-실측-진단--이-대시보드의-숫자-읽기)
- [12. No data 처방전](#12-no-data-처방전)
- [13. 무엇을 경보로 걸까 — 운영 체크리스트](#13-무엇을-경보로-걸까--운영-체크리스트)
- [부록 A. 패널-메트릭-출처 매핑표](#부록-a-패널-메트릭-출처-매핑표)
- [부록 B. 한 줄 용어 사전](#부록-b-한-줄-용어-사전)
- [부록 C. 참고 자료](#부록-c-참고-자료)

---

## 0. 이 문서를 읽는 법, 그리고 결론부터

상황별 진입점:

| 지금 상황 | 여기로 |
| --- | --- |
| 기초가 없다 — 처음부터 차근차근 | 1장 → 2장 (비유 하나로 끝까지) |
| 특정 패널이 안 읽힌다 | 3~10장에서 해당 행(row) |
| No data 패널을 고치고 싶다 | 12장 (처방전) |
| 지금 우리 서버 괜찮은 건가 | 11장 (실측 진단) |
| 알람을 걸고 싶다 | 13장 |
| 이 패널이 무슨 메트릭·쿼리인지 | 부록 A |

캡처된 대시보드에 대한 결론을 먼저 세 문장으로:

1. **이 JVM은 한가하고 건강하다.** CPU 0 %, load 0.0/4코어, GC 압력 0 %, 힙 점유 16 %(83.7/512 MiB), 락 경합 0, FD 사용률 0.14 %. 즉시 조치할 항목이 없다.
2. **No data 패널들은 고장이 아니라, 원인이 제각각인 "정상"이다.** 5xx가 한 번도 없어서(Errors), 톰캣 메트릭이 꺼져 있거나 톰캣이 아니어서(Utilisation), 별도 라이브러리가 필요해서(Process Memory)다. 12장에 각각의 처방이 있다.
3. **경보로 걸 가치가 있는 것은 순간값이 아니라 추세다.** Old Gen 바닥값의 우상향, error 로그율 급증, FD 사용률, 스레드 수 증가. 13장의 체크리스트로 정리했다.

---

## 1. 숫자가 화면에 도달하기까지 — 측정 파이프라인

### 1.1 네 단계 여행

Grafana 화면의 숫자 하나는 네 단계를 거쳐 도착한 것이다.

```text
┌────────────────────────────────────────────────────┐
│   Spring Boot 애플리케이션 (하나의 JVM 프로세스)   │
│                                                    │
│   ① JVM 내부 계측기 (MXBean)                       │
│      MemoryMXBean, ThreadMXBean,                   │
│      GarbageCollectorMXBean, ...                   │
│              │                                     │
│   ② Micrometer (계측 SDK)                          │
│      MXBean 값을 읽어 표준 "메트릭"으로 변환       │
│              │                                     │
│      GET /actuator/prometheus  (노출 창구)         │
└─────────────┬──────────────────────────────────────┘
              │  ③ Prometheus가 15초마다 긁어감 (scrape)
              ▼
      ┌──────────────┐    PromQL 질의      ┌─────────────┐
      │  Prometheus  │ ◄────────────────── │  ④ Grafana  │
      │  (시계열 DB) │ ──────────────────► │  (대시보드) │
      └──────────────┘    데이터 반환      └─────────────┘
```

자동차에 비유하면 역할이 정확히 갈린다.

- **JVM = 엔진.** 회전수·온도 같은 내부 상태를 갖고 있다.
- **Micrometer = 센서.** 엔진에 붙어 상태를 숫자로 바꾼다.
- **Prometheus = 블랙박스.** 그 숫자를 일정 주기로 기록해 시간순으로 보관한다.
- **Grafana = 계기판.** 기록을 조회해서 그래프로 그릴 뿐, **아무것도 직접 측정하지 않는다.**

이 구분이 중요한 이유: 패널이 비어 있으면("No data") 이 사슬 중 어디가 끊겼는지 거꾸로 따라가면 된다. 12장의 진단이 정확히 이 순서를 탄다.

부수 효과 하나. Prometheus는 보통 15초(설정에 따라 10~60초) 간격으로 샘플을 뜨므로, **그 사이에 일어난 스파이크는 화면에 없다.** "그래프는 잠잠했는데 순간 응답이 튀었다"가 가능한 이유다.

### 1.2 화면 읽는 법 — Last *, Max, 그리고 No data

패널 아래 범례 테이블의 칼럼은 Grafana의 집계 표시다.

- **Last \*** — 시계열의 마지막 **비-null** 샘플. 별표(\*)가 "null 제외"라는 뜻이다. 사실상 "지금 값".
- **Max** — **현재 화면에 잡힌 시간 범위 안에서의** 최댓값. 예: 힙 used가 `Last 83.7 MiB / Max 369 MiB`라면 "지금은 84지만, 보고 있는 기간 중 한때 369까지 찼다"는 뜻이다. 시간 범위를 바꾸면 Max도 바뀐다.
- **No data** — 0이 아니다. **그런 시계열이 저장소에 아예 없다**는 뜻이다.

마지막 항목은 Prometheus의 중요한 성질과 닿아 있다. **카운터는 첫 사건이 일어나기 전까지 존재하지 않는다.** "HTTP 500 응답 수" 카운터는 첫 500이 발생하는 순간 태어난다. 그래서 Errors 패널의 No data는 대개 "한 번도 에러가 없었음"이라는 좋은 소식이다(3.2).

### 1.3 MiB와 MB — 단위가 두 종류다

- **MiB** = 2^20 바이트 = 1,048,576 B (이진 단위. Mi, Gi, Ki)
- **MB** = 10^6 바이트 = 1,000,000 B (십진 단위)

같은 대시보드 안에서도 패널에 따라 섞여 나온다(메모리 패널은 MiB, 버퍼 풀은 MB로 찍히는 식). 512 MiB는 536.9 MB다. 두 수치를 검산할 때 단위부터 맞추지 않으면 5 %쯤 어긋나 보이는데, 고장이 아니라 단위 차이다.

---

## 2. JVM이라는 도시 — 전체 지도

### 2.1 JVM 1분 요약

자바 컴파일러(`javac`)는 소스를 기계어로 바꾸지 않는다. **바이트코드**라는, 실제로는 존재하지 않는 "가상의 CPU"용 명령어로 바꾼다. JVM(Java Virtual Machine)은 그 가상 CPU를 소프트웨어로 구현한 것이고, 덕분에 같은 `.class`/`.jar`가 리눅스·윈도우·맥 어디서든 그대로 돈다.

JVM의 본업은 세 가지다.

1. **설계도 반입** — 클래스 파일을 찾아서 메모리에 올린다 (클래스 로딩)
2. **실행** — 바이트코드를 해석하고, 자주 쓰는 것은 진짜 기계어로 번역해 둔다 (인터프리터 + JIT 컴파일)
3. **메모리 자동 관리** — 개발자가 `free()`를 부르지 않아도, 안 쓰는 객체를 찾아 치운다 (GC)

이 대시보드의 행(row)들은 정확히 이 세 본업에 대응한다.

| JVM의 본업 | 대시보드 행 |
| --- | --- |
| ① 설계도 반입 | Classloading, Memory Pools(Non-Heap)의 Metaspace |
| ② 실행 | JVM Misc(CPU·Threads), Memory Pools(Non-Heap)의 CodeHeap |
| ③ 메모리 관리 | JVM Memory, Memory Pools(Heap), Garbage Collection, Buffer Pools |
| (입구 계측) | I/O Overview |

### 2.2 도시 전체 지도

이 문서 끝까지 끌고 갈 비유: **JVM 프로세스는 하나의 작은 도시다.** OS가 국가, 메모리가 토지, 객체가 주민, 스레드가 일꾼이다. Grafana는 이 도시의 관제센터 모니터다.

OS가 보는 것은 "JVM 도시" 하나의 경계선뿐이지만, 그 안은 이렇게 구획돼 있다.

```text
┌─ JVM 프로세스 (OS가 보는 한 덩어리) ──────────────────────────────────────────────┐
│                                                                                   │
│  ◆ Heap — 주거 지구 (자바 객체 = 주민)             ◆ Non-Heap — 행정 지구         │
│  ┌────────────────────────────────────────┐        ┌───────────────────────────┐  │
│  │ Young   ┌──────────┐ ┌────┐ ┌────┐     │        │ Metaspace = 설계도 보관소 │  │
│  │         │   Eden   │ │ S0 │ │ S1 │     │        │  (클래스 메타데이터)      │  │
│  │         └──────────┘ └────┘ └────┘     │        ├───────────────────────────┤  │
│  │                                        │        │ Compressed Class Space    │  │
│  │ Old     ┌──────────────────────────┐   │        │  (압축 클래스 포인터용)   │  │
│  │         │          Old Gen         │   │        ├───────────────────────────┤  │
│  │         └──────────────────────────┘   │        │ CodeCache (CodeHeap x3)   │  │
│  │                                        │        │  = JIT 번역청             │  │
│  └────────────────────────────────────────┘        └───────────────────────────┘  │
│                                                                                   │
│  ◆ 그 밖의 네이티브 영역 — jvm_memory 장부 밖 (4.4에서 중요해짐)                  │
│  ┌───────────────────────────────────────────────────────────────────────────────┐│
│  │ 스레드 스택 (스레드당 약 1 MiB) · Direct/Mapped 버퍼 (성벽 밖 창고 — 10장)    ││
│  │ GC 부속 자료구조 · JNI/네이티브 malloc · JVM 자체 코드·데이터 ...             ││
│  └───────────────────────────────────────────────────────────────────────────────┘│
│                                                                                   │
└───────────────────────────────────────────────────────────────────────────────────┘
```

대시보드 행과 지도의 대응:

| 대시보드 행 | 지도 위치 | 도시 비유 |
| --- | --- | --- |
| JVM Memory | Heap + Non-Heap 합산 | 시 전체 살림 요약 |
| Memory Pools (Heap) | Eden / Survivor / Old | 주거 지구 동(洞)별 현황 |
| Memory Pools (Non-Heap) | Metaspace / CCS / CodeHeap | 행정 지구 청사별 현황 |
| Garbage Collection | Heap 전체에 작용 | 청소국 작업 일지 |
| Buffer Pools | 성벽 밖 네이티브 | 보세창고 재고 |
| Classloading | Metaspace로의 반입 | 설계도 반입 대장 |
| JVM Misc | 도시 전반 | 활력 징후(인력·전력·출입문) |
| I/O Overview | 도시 정문 | 방문객 통계 |

### 2.3 used · committed · max — 모든 메모리 패널의 공통 문법

메모리 계열 패널은 전부 같은 세 값을 그린다. 주차장으로 비유하면:

- **max** — 구청에서 허가받은 **부지 한도**. 여기까지 확장할 수 있다는 약속일 뿐, 아직 땅도 안 팠다. (가상 주소 공간 예약)
- **committed** — 실제로 **포장 공사를 마친 주차면**. OS가 물리 메모리(또는 스왑)를 내주기로 보증한 부분. OS 관점에서 "이 프로세스가 가져간 메모리"에 가깝다.
- **used** — 지금 **주차된 차**. JVM이 실제 데이터로 채운 양.

```text
0 ──────────────────────────────────────────────────────►  주소 공간
   │■■■■■■■■■■░░░░░░░░░░░░░░░░░░│ · · · · · · · · · · · │
   └── used ──┘                 │                       │
   └──────── committed ─────────┘                       │
   └──────────────────────── max ───────────────────────┘
    ■ 데이터가 든 부분   ░ 받아놨지만 빈 부분   · 예약만 해둔 부분
   항상  used ≤ committed ≤ max
```

세 값 모두 JVM 내부의 `MemoryPoolMXBean`이 보고하며, Micrometer가 `jvm_memory_used_bytes`, `jvm_memory_committed_bytes`, `jvm_memory_max_bytes`로 내보낸다.

캡처에서 힙의 committed와 max가 똑같이 512 MiB로 붙어 있는데, 이는 `-Xms`와 `-Xmx`를 같은 값으로 고정했을 때의 전형적 모양이다. 운영 서버에서 힙을 고정하는 것은 널리 쓰이는 관행으로, 힙이 늘었다 줄었다 하며 생기는 비용과 변동을 없애려는 목적이다.

---

## 3. I/O Overview — 도시의 정문

이 행만 JVM 내부가 아니라 **들어오는 HTTP 트래픽**을 본다. Micrometer가 Spring MVC(또는 WebFlux) 처리 경로에 계측기를 끼워 넣어, 요청이 핸들러에 들어와 응답이 나갈 때까지를 잰 것이 `http_server_requests` 메트릭이다.

식당으로 치면: Rate = 단위 시간당 입장객 수, Duration = 주문부터 서빙까지 걸린 시간, Errors = 주방 실수로 못 내보낸 주문 수, Utilisation = 홀 직원 가동률.

### 3.1 Rate — 초당 요청 수

원본 메트릭은 "기동 이후 누적 요청 수"라는 **단조 증가 카운터**다. 그래프에 보이는 것은 그 카운터의 **기울기**(PromQL `rate()`), 즉 초당 요청 수(RPS)다.

```text
누적 카운터          rate() = 기울기
       ╱│
      ╱ │ 가파름 = 바쁨        ──► 5 req/s
 ────╯  │
─╯       완만함 = 한가         ──► 0.1 req/s
```

카운터를 누적으로 두고 기울기를 나중에 계산하는 이유는, 수집이 한두 번 실패해도(15초 구멍) 누적값은 살아 있어서 평균 기울기를 복원할 수 있기 때문이다.

### 3.2 Errors — No data의 정체

같은 카운터에서 **서버 오류(5xx) 응답만** 골라 센다(대시보드 리비전에 따라 `status=~"5.."` 또는 `outcome="SERVER_ERROR"` 태그 기준).

캡처에서 No data인 이유는 1.2에서 본 그 성질이다. **5xx가 한 번도 발생하지 않아 해당 시계열이 아예 태어나지 않았다.** 무소식이 희소식인 패널이며, 고칠 것이 없다. 굳이 0으로 보이게 하고 싶다면 패널 옵션의 "No value → 0" 표시 설정이나 쿼리에 `or vector(0)`을 더하는 방법이 있다(12.1).

### 3.3 Duration — 평균 응답 시간

카운터는 사실 쌍으로 산다: 누적 횟수(`_count`)와 누적 소요 시간(`_sum`). 평균 = Δsum ÷ Δcount. 즉 이 패널은 **구간 평균** 응답 시간이다.

평균의 함정을 알고 봐야 한다. 요청 100개 중 99개가 10 ms, 1개가 5초면 평균은 60 ms로 "멀쩡해 보인다". 느린 소수를 잡으려면 백분위(p95/p99)가 필요한데, 그러려면 히스토그램 수집을 켜야 한다:

```yaml
management:
  metrics:
    distribution:
      percentiles-histogram:
        http.server.requests: true   # 저장 용량이 늘어나는 대신 p95/p99 질의 가능
```

### 3.4 Utilisation — 일꾼 가동률 (No data)

표준 4701 기준으로 이 패널은 **톰캣 워커 스레드 가동률**(busy ÷ max, `tomcat_threads_*`)을 그린다. 홀 직원 200명 중 몇 명이 지금 서빙 중인가다. 1.0에 붙으면 새 손님이 줄을 서기 시작한다(응답 지연의 직접 원인).

No data의 대표 원인은 두 갈래다 — ① 톰캣인데 JMX 기반 메트릭이 꺼져 있는 경우(Spring Boot 2.2부터 기본 비활성), ② 애초에 톰캣이 아닌 경우(WebFlux/Netty). 처방은 12.2.

---

## 4. JVM Memory — 살림살이 한눈에

이 행은 2.3의 세 값(used/committed/max)을 Heap, Non-Heap, 둘의 합(Total), 그리고 OS 실측(Process Memory)으로 나눠 보여준다.

### 4.1 JVM Heap — 톱니를 읽는 패널

캡처: used `Last 83.7 MiB / Max 369 MiB`, committed 512 MiB, max 512 MiB.

힙 used 그래프의 정상 모양은 **톱니(sawtooth)** 다.

```text
used
512Mi ┤ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  committed = max (고정 힙)
      │        ╱│         ╱│        ╱│
369Mi ┤      ╱  │       ╱  │      ╱  │   ← 톱니 꼭대기 (범례의 Max가 이걸 잡는다)
      │    ╱    │     ╱    │    ╱    │
      │  ╱      │   ╱      │  ╱      │
 84Mi ┤╱        │╱         │╱        ╰──── 지금 (Last *)
      └─────────────────────────────────────────►  시간
        객체 할당으로 차오름 ──► young GC가 한 번에 비움 ──► 반복
```

읽는 법의 핵심 두 가지:

- **꼭대기가 max에 닿는 것 자체는 사고가 아니다.** GC가 돌기 직전까지 차는 것은 설계된 동작이다.
- **사고는 "GC 후 바닥값"이 회차마다 높아지는 것이다.** 청소를 해도 안 비워지는 객체가 쌓인다는 뜻 — 메모리 누수의 그래프 형태다. 이 "바닥값의 추세"가 힙에서 가장 가치 있는 단 하나의 신호다(13장).

캡처의 Max 369 MiB는 "보고 있던 기간 중 한때 369까지 찼다"는 뜻이고(1.2), 지금 84로 내려와 있으므로 GC가 정상적으로 비웠다는 증거다.

### 4.2 JVM Non-Heap

힙 밖 행정 지구(Metaspace + Compressed Class Space + CodeHeap 3개)의 합산이다. 구성 요소는 7장에서 하나씩 본다.

힙과 달리 non-heap의 committed는 used를 바짝 따라가며 야금야금 늘어나는 모양이 정상이다. 필요해질 때마다 조금씩 commit하는 영역이기 때문이다. 또 기동 후 한동안(클래스 로딩·JIT 번역이 활발한 워밍업 구간) 늘다가 수평이 되는 것이 전형이다.

### 4.3 JVM Total — 1.73 GiB라는 숫자의 정체

Total = Heap + Non-Heap. 캡처: used 162 MiB, committed 594 MiB, **max 1.73 GiB**.

힙이 512 MiB뿐인데 max가 1.73 GiB라서 놀라기 쉽다. 산수를 해보면 정체가 드러난다.

```text
힙 max                          512 MiB   (-Xmx)
CodeCache 예약 한도              240 MiB   (JDK 9+ 기본값)
Compressed Class Space 예약     1024 MiB   (기본값, 7.2)
                              ─────────
                               1776 MiB  =  1.734 GiB   ✓ 패널의 1.73 GiB와 일치
(Metaspace는 한도 무제한 = max가 -1로 보고되어 합산에서 빠진다)
```

즉 **1.73 GiB는 "쓸 예정인 메모리"가 아니라 "예약 한도의 합산"** 이다. 실제 점유에 가까운 것은 committed(594 MiB) 쪽이고, max는 2.3의 "허가받은 부지 한도"들을 더한 서류상 숫자다.

### 4.4 JVM Process Memory — JVM이 아는 것과 OS가 보는 것 (No data)

위의 모든 jvm_memory 계열은 **JVM의 자기 신고**다. 회계 장부에 적힌 항목(힙·논힙)만 더한 값이라, 장부 밖 지출이 빠져 있다.

```text
OS가 실측한 거주 메모리(RSS) — 도시가 실제로 차지한 땅
 = 힙 + 논힙                       ◄── 여기까지만 jvm_memory_* 가 안다 (JVM의 자기 신고)
 + 스레드 스택 (73개 x ~1 MiB)     ┐
 + Direct/Mapped Buffer (10장)     │   장부 밖 지출.
 + GC 부속 자료구조                │   이 차이가 계속 자라면
 + JNI·네이티브 라이브러리 malloc  │   "네이티브 누수"를 의심
 + JVM 자체 코드·데이터            ┘
```

이 패널은 그 실측값(RSS·VSS)을 그리라고 만들어졌고, 기본 Micrometer에는 없는 메트릭이라 **별도 라이브러리(micrometer-jvm-extras)를 넣어야 채워진다**(12.3). 그래서 No data다.

채울 가치가 있는 이유: 컨테이너(k8s)의 메모리 limit과 OOMKill은 힙이 아니라 **RSS 기준**으로 집행된다. "힙은 여유인데 파드가 자꾸 죽어요"라는 단골 미스터리의 답이 대부분 이 장부 밖 영역에 있다.

---

## 5. JVM Misc — 활력 징후

사람으로 치면 맥박·혈압·체온 패널이다. 캡처 값 기준으로 하나씩 읽는다.

### 5.1 CPU Usage — system 0.3 %, process 0.0 %

- **system** — 머신(VM) 전체의 CPU 사용률. 이 JVM 말고 다른 프로세스도 포함.
- **process** — 이 JVM 프로세스 몫만.
- **process-15m** — process의 15분 평활(이동평균). 순간 스파이크를 무시한 추세선.

값은 0~1로 정규화되어 **전체 코어 기준**이다. 4코어 머신에서 process 25 % = 코어 하나를 꽉 채워 쓰는 중이라는 뜻이다. 캡처의 process 0.0 %는 사실상 유휴. system Max 8.6 %는 기간 중 다른 무언가(다른 프로세스 또는 순간 부하)가 잠깐 지나간 흔적이다.

### 5.2 Load — 계산대에 선 줄의 길이

`system_load_average_1m`(최근 1분 부하 평균)과 `system_cpu_count`(기준선)를 같이 그린다.

마트 계산대 비유: CPU 코어 4개 = 계산대 4개. load = **계산 중인 사람 + 줄 서 있는 사람**의 평균 수. load 4.0이면 계산대가 정확히 만석, 8.0이면 전원 계산 중 + 4명 대기다. 리눅스의 load는 CPU 대기뿐 아니라 디스크 I/O를 기다리며 잠든(D state) 작업도 포함하므로, load만 높고 CPU가 낮으면 디스크 쪽을 봐야 한다 — 그 갈래는 [storage_io_vmware.md](../linux/commands/metrics/storage_io_vmware.md)의 영역이다.

캡처: load 0.0(Max 0.5) / 4코어. 줄이 없다.

### 5.3 Threads — live 73, daemon 25, peak 77

- **live** — 지금 살아 있는 자바 스레드 수 (daemon 포함)
- **daemon** — 그중 데몬 스레드
- **peak** — 기동 이후 역대 최고치

데몬의 의미는 회사 퇴근에 비유하면 정확하다. **정규 직원(user 스레드)이 모두 퇴근하면 회사(JVM)는 문을 닫는데, 청소 용역(daemon)이 남아 있는지는 묻지 않는다.** GC 보조, 모니터링, 타이머 같은 보조 역할이 데몬으로 돈다.

73개의 전형적 구성(Spring Boot 기준): main 계열 + 톰캣 워커 10개 내외와 acceptor/poller + 커넥션 풀(HikariCP) + 스케줄러 + GC·JIT 관련 자바 스레드들. 참고로 GC 워커처럼 JVM이 직접 띄우는 순수 네이티브 스레드는 이 카운트(자바 스레드)에 잡히지 않는다.

스레드 하나는 스택용으로 약 1 MiB(기본값)의 가상 메모리를 예약하므로, 73개면 약 70 MiB가 4.4의 "장부 밖" 항목으로 들어간다. **live가 수평이 아니라 꾸준히 우상향하면 스레드 누수**이며, 그 진단은 [thread_leak.md](./thread_leak.md)에 정리돼 있다.

### 5.4 Thread States — RUNNABLE 42의 미스터리

자바 스레드는 정확히 6가지 상태(`Thread.State`) 중 하나에 있다.

```text
              ┌─────┐
              │ NEW │  생성만 되고 start() 전
              └──┬──┘
                 ▼ start()
           ┌──────────┐   synchronized 락 대기   ┌─────────┐
     ┌────►│ RUNNABLE │ ───────────────────────► │ BLOCKED │
     │     └────┬─────┘ ◄─────────────────────── └─────────┘
     │          │                락 획득
     │          │ wait() / join() / park()  (+시한부면 TIMED_)
     │          ▼
     │   ┌─────────────────────────┐
     └───┤ WAITING / TIMED_WAITING │   notify / unpark / 시간 만료
         └───────────┬─────────────┘
                     │ run() 종료
                     ▼
              ┌────────────┐
              │ TERMINATED │
              └────────────┘
```

캡처를 보면 CPU는 0 %인데 RUNNABLE이 42개다. 모순처럼 보이지만 아니다. **자바의 RUNNABLE은 "CPU에서 도는 중"이 아니라 "JVM 차원에서는 막힌 게 없음"이라는 뜻**이고, 여기에는 **네이티브 코드로 내려가 OS에서 I/O를 기다리는 스레드가 포함**된다. 톰캣 poller, NIO selector처럼 `epoll`에서 잠들어 있는 스레드들은 OS 입장에선 자고 있지만 자바 분류로는 RUNNABLE이다. 42개의 대부분이 그들이다.

상태별 건강 신호:

- **BLOCKED 0** — `synchronized` 모니터를 기다리는 스레드가 없다 = 락 경합 없음. BLOCKED가 지속적으로 두 자릿수면 병목 락이 있다는 뜻이다.
- **나머지 31개(73−42)** 는 WAITING/TIMED_WAITING — 스레드 풀에서 일감을 기다리는 유휴 일꾼들. 정상.

### 5.5 GC Pressure — 0.0 %

전체 CPU 시간 중 GC가 가져간 비율. "도시 노동력의 몇 %가 청소에 투입되는가"다. 경험칙으로 5 %를 넘으면 주의, 10 %를 넘으면 힙 크기·할당률 튜닝이 필요한 신호로 본다. 캡처의 0.0 %는 청소국이 거의 일이 없다는 뜻이다.

### 5.6 Log Events — 레벨별 로그 발생률

Micrometer가 Logback 파이프라인에 카운터 훅을 걸어, 레벨별(trace/debug/info/warn/error) 초당 로그 건수를 센다(`logback_events_total`).

겉보기엔 수수하지만 **가장 값싼 조기 경보**다. 애플리케이션 오류는 대부분 메트릭에 잡히기 전에 error 로그부터 찍는다. "error 로그율이 평소 대비 급증"은 5xx, 큐 적체, 외부 연동 장애를 가장 먼저 비추는 거울이다(13장).

### 5.7 File Descriptors — open 93 / max 66 K

유닉스의 설계 철학 "모든 것은 파일"에 따라, 진짜 파일뿐 아니라 **소켓, 파이프, epoll 핸들까지 전부 fd**를 하나씩 쓴다. 도시 비유로는 외부와 통하는 **출입문의 개수**다. HTTP 커넥션 하나 = 문 하나, 열린 로그 파일 = 문 하나.

- open 93 — 지금 열려 있는 문
- max 66 K — OS가 이 프로세스에 허가한 한도. 화면의 66 K는 반올림 표시고, 전형적인 설정값은 65,536(`ulimit -n`)이다

한도에 닿으면 새 커넥션 수락 자체가 실패하며 그 유명한 `Too many open files` 예외가 터진다. 전형적 원인은 close하지 않은 스트림/커넥션의 누적(fd 누수)이고, 그래서 절대값보다 **사용률(open÷max)과 추세**를 본다. 캡처는 0.14 % — 아주 여유.

---

## 6. JVM Memory Pools (Heap) — 객체의 일생과 G1

### 6.1 왜 세대를 나누나 — 약한 세대 가설

힙을 한 덩어리로 안 쓰고 Eden/Survivor/Old로 쪼개는 이유는 통계적 관찰 하나에서 나온다.

> **약한 세대 가설(weak generational hypothesis): 대부분의 객체는 어려서 죽는다.**

HTTP 요청 하나를 처리하면 문자열, DTO, 이터레이터, 임시 컬렉션이 우수수 태어나고, 응답이 나가는 순간 전부 쓰레기가 된다. 살아남는 것은 캐시 항목이나 커넥션 같은 극소수다.

이 사실을 설계로 승화한 것이 세대별 GC다. **신생아만 한 구역(Eden)에 모아 두면, 청소할 때 "살아남은 소수만 옮기고 구역 전체를 통째로 리셋"할 수 있다.** 청소 비용이 죽은 객체 수가 아니라 산 객체 수에 비례하게 되는데, 산 객체가 극소수이므로 거의 공짜가 된다. 호텔로 치면 방을 하나하나 청소하는 대신, 투숙객 몇 명만 옆 동으로 옮기고 층 전체를 물청소하는 방식이다.

### 6.2 객체의 일생 — Eden → Survivor → Old

```text
 new Foo()                       (이사 = minor GC 한 번에서 살아남음)
    │
    ▼
┌─────────┐  생존자만 복사   ┌────┐ ⇄ ┌────┐   이사 도장 ≥ 임계값   ┌─────────┐
│  Eden   │ ───────────────► │ S0 │   │ S1 │ ─────────────────────► │ Old Gen │
│ (신생아)│                  └────┴───┴────┘    승격(promotion)     │ (정착촌)│
└─────────┘                  Survivor 두 동을                       └─────────┘
    │                        오가며 나이 도장 +1
    ▼
  대부분(90 %+)은 첫 GC에서 사망 → 비용 0에 가깝게 회수
```

- 모든 객체는 **Eden**에서 태어난다.
- Eden이 차면 **minor(young) GC**: 살아 있는 객체만 Survivor로 복사하고 Eden을 통째로 리셋.
- Survivor에서 GC를 한 번 버틸 때마다 객체 헤더의 **나이(age)가 +1**. 헤더의 나이 칸이 4비트라 상한이 15이며, 임계값(기본 최대 15, 실제로는 GC가 동적으로 조절)을 넘기면 **Old로 승격**된다.
- Old는 "오래 살 것이 검증된" 객체들의 정착촌으로, 청소 주기가 훨씬 길다.

### 6.3 왜 Survivor는 두 개인가 — 복사 GC

S0/S1은 항상 **한쪽만 사용하고 한쪽은 비워 둔다.** GC 때마다 "Eden + 사용 중인 Survivor"의 생존자를 **빈 쪽으로 전부 복사**하고, 역할을 맞바꾼다.

```text
GC 전:  Eden ■■■■■    S0 ■■░░ (사용 중)    S1 ░░░░ (빈 집)
                 │            │
                 └─ 생존자만 ─┴──────► S1로 복사 (나이 +1)
GC 후:  Eden ░░░░░    S0 ░░░░ (빈 집)     S1 ■■░░ (사용 중)
```

이 "통째로 옮겨 적기"의 대가로 얻는 것이 **단편화 제로**다. 살아남은 객체를 빈 공간에 차곡차곡 이어 쓰므로, 청소가 끝나면 항상 연속된 빈 땅이 생긴다. Survivor 한 동을 늘 놀리는 비효율은 청소 시간을 사기 위한 비용이다.

### 6.4 G1 — 바둑판 도시계획

캡처의 풀 이름이 G1 Eden/Survivor/Old인 것은 GC가 **G1**(JDK 9+의 기본 GC)이라는 뜻이다. G1의 혁신은 세대를 **물리적으로 연속된 큰 구역이 아니라, 같은 크기의 리전(region) 바둑판**으로 구현한 것이다.

```text
G1 힙 = 같은 크기 리전의 바둑판 (리전 크기는 힙에 비례해 1~32 MiB,
        캡처처럼 512 MiB 힙이면 1 MiB × 약 512칸)

┌────┬────┬────┬────┬────┬────┬────┐    E = Eden        S = Survivor
│ E  │ O  │ O  │ E  │ S  │ O  │ 빈 │    O = Old         빈 = Free
├────┼────┼────┼────┼────┼────┼────┤    H = Humongous (리전 절반 이상의
│ O  │ E  │ 빈 │ H  │ H→ │ O  │ E  │        대형 객체 — 연속 리전을 통째 점유.
└────┴────┴────┴────┴────┴────┴────┘        대형 배열 남발 시 골칫거리)

칸의 역할(E/S/O)은 고정이 아니라 그때그때 다시 지정된다.
```

이 구조 덕에 G1은 "힙 전체"가 아니라 **수확이 좋은 칸만 골라** 청소할 수 있다. 쓰레기 비율이 높은 리전부터(Garbage **First** — 이름의 유래) 회수해서, 목표 멈춤 시간 안에 끝나는 만큼만 일한다.

### 6.5 세 패널에서 봐야 할 것

| 패널 | 정상 모양 | 이상 신호 |
| --- | --- | --- |
| G1 Eden Space | 가파른 톱니 (차오름→리셋 반복) | 톱니 주기가 점점 짧아짐 = 할당률 급증 (8.5) |
| G1 Survivor Space | GC 때마다 출렁임 | 항상 가득 = 승격 압박, Old로 객체가 밀려 들어가는 중 |
| G1 Old Gen | 수평 또는 완만, 가끔 계단식 하강(mixed GC) | **GC 후 바닥값의 우상향 = 누수의 그래프 서명** |

참고: G1에서는 리전 재지정 때문에 Eden/Survivor 풀의 max가 비어 있거나 -1로 보일 수 있다. 경계가 유동적이라 "이 풀의 한도"라는 개념 자체가 없는 것이며, 버그가 아니다.

---

## 7. JVM Memory Pools (Non-Heap) — 설계도 보관소와 번역청

힙이 "객체(주민)"의 땅이라면, non-heap은 "클래스(설계도)와 기계어(번역물)"의 땅이다.

### 7.1 Metaspace — 클래스의 설계도 보관소

객체가 건물이라면 클래스는 설계도다. `new User()`를 백만 번 해도 건물(인스턴스)이 백만 채 들어설 뿐, 설계도(클래스 메타데이터)는 한 부면 된다. 그 설계도들 — 메서드 바이트코드, 상수 풀, 필드 배치도, vtable — 이 사는 곳이 Metaspace다. 힙의 모든 객체는 헤더에 자기 설계도를 가리키는 포인터를 들고 있다.

역사 한 토막: JDK 7까지는 이 설계도들이 "PermGen"이라는 고정 크기의 칸에 살았고, 크기 산정을 잘못하면 `OutOfMemoryError: PermGen space`가 터지는 단골 장애가 있었다. JDK 8에서 PermGen을 폐지하고 **네이티브 메모리에서 필요한 만큼 자라는 Metaspace**로 옮겼다. 기본 한도가 무제한(max = -1)인 이유이고, 4.3에서 Total max 합산에 안 들어가던 이유다.

운영에서 보는 법: 기동 때 가파르게 차고 그 뒤 수평이 정상. **꾸준히 우상향하면 클래스가 계속 만들어지고 있다는 뜻**인데, 클래스 메타데이터는 그 클래스를 로드한 클래스로더가 통째로 죽어야 회수되므로, 동적 프록시·리플렉션 남용이나 클래스로더 누수를 의심한다. 9.2의 Classes loaded 패널과 반드시 같이 본다.

### 7.2 Compressed Class Space — 주소를 반으로 접는 기술

64비트 시스템에서 포인터는 8바이트다. 객체마다 들어 있는 "설계도 포인터"와 객체 참조들이 전부 8바이트면 힙이 포인터로 비대해진다. HotSpot의 처방은 **압축 포인터**: 전체 주소 대신 **기준점으로부터의 32비트 오프셋**만 기록한다.

아파트 단지 비유가 정확하다. "서울시 ○○구 ○○로 123, ○○아파트 103동 502호"를 매번 다 적는 대신, **단지 안에서는 "103-502"만 적는** 것이다. 단, 이 축약이 성립하려면 조건이 하나 필요하다 — **모두가 같은 단지 안에 살아야 한다.**

- 객체 참조(compressed oops): 8바이트 정렬 덕에 32비트 주소 2^32개 × 8바이트 = **힙 32 GiB까지** 이 축약이 가능하다.
- 클래스 포인터(compressed class pointers): 설계도들이 **하나의 연속된 영역 안에** 모여 살아야 한다. 그 "단지"로 기본 1 GiB의 가상 주소를 예약해 둔 것이 바로 Compressed Class Space다.

그래서 이 패널의 max 1 GiB는 겁먹을 숫자가 아니다. 2.3의 구분으로 말하면 **예약(max)일 뿐이고, committed는 보통 수~수십 MiB**에 그친다. 다만 4.3에서 봤듯 Total max 합산을 1.73 GiB로 부풀리는 주범이다.

### 7.3 CodeHeap 3형제 — JIT 번역청

JVM이 바이트코드를 실행하는 방식은 통역에 비유하면 정확하다.

- **인터프리터** = 동시통역사. 한 문장씩 즉석에서 통역한다. 시작은 빠르지만 같은 문장도 매번 다시 통역한다.
- **JIT 컴파일러** = 번역가. 자주 읽히는 문서는 아예 **번역본 책자(기계어)** 로 만들어 책장에 꽂아 둔다. 만들 땐 시간이 들지만 그 뒤로는 원어민 속도다.

그 책장이 CodeCache이고, JDK 9부터 용도별 세 칸(CodeHeap)으로 분할됐다(총예산 기본 240 MiB). 번역은 2단계로 승급한다.

```text
바이트코드
   │
   ▼
인터프리터 ──핫스팟(자주 호출됨) 감지──► C1 컴파일 ──► CodeHeap 'profiled nmethods' 에 보관
   ▲                                        │             (초벌 번역 + 프로파일 계측 포함)
   │                                        │ 실행 통계가 충분히 쌓이면 (대략 수만 회)
   │                                        ▼
   │                                      C2 컴파일 ──► CodeHeap 'non-profiled nmethods' 에 보관
   │                                        │             (완성 번역, 가장 빠름)
   └────────────────────────────────────────┘
     탈최적화(deopt): 컴파일 때의 가정이 깨지면 번역본을 버리고 인터프리터로 복귀

CodeHeap 'non-nmethods' = 위 둘이 아닌 공용 부품 (인터프리터 스텁, 어댑터 등)
```

- **nmethod** = native method, "컴파일된 자바 메서드"의 HotSpot 내부 용어다. 패널 이름의 정체.
- **profiled nmethods** — C1의 결과물. 실행하면서 통계(어느 분기가 잦은가, 어떤 타입이 오는가)를 모으는 계측이 박혀 있는, 메모 달린 초벌 번역본.
- **non-profiled nmethods** — 그 통계를 근거로 C2가 만든 완성 번역본. 계측이 빠져 있어 가장 빠르다.

운영 함의 두 가지. ① 배포 직후 한동안 느린 **워밍업**은 번역청이 아직 일을 마치지 못한 기간이다. ② CodeCache가 가득 차면 JVM은 컴파일을 멈추고(`CodeCache is full. Compiler has been disabled` 경고) 통역 모드로 살게 되어 성능이 수 배 떨어진다 — 이 패널들의 used가 240 MiB 예산에 다가가는지 봐야 하는 이유다.

---

## 8. Garbage Collection — 청소국의 작업 일지

### 8.1 GC가 쓰레기를 정하는 법 — 도달 가능성

GC는 "안 쓸 객체"를 점쟁이처럼 맞히는 게 아니다. 기준은 기계적이다: **뿌리(GC root)에서 참조를 따라가서 닿을 수 있으면 생존, 못 닿으면 쓰레기.**

```text
GC 루트 = 지금 당장 코드가 손에 쥔 것들
          (각 스레드 스택의 지역변수·파라미터, static 필드, JNI 핸들 ...)

루트 ──► A ──► B            A, B : 루트에서 닿음 → 생존
  │
  └────► C                  C    : 생존
          D ⇄ E             D, E : 서로 꼭 붙들고 있지만 루트에서 닿는
                                   길이 없음 → 둘 다 쓰레기
```

마지막 줄이 핵심이다. **서로 참조하는 순환 고리도, 바깥(루트)과 연이 끊기면 섬째로 회수된다.** 참조 횟수를 세는 방식(예: 파이썬의 1차 메커니즘)이 순환 참조에 약한 것과 대비되는, 추적(tracing) GC의 강점이다.

### 8.2 Stop-The-World — 단체 사진의 시간

객체 그래프를 따라가는 동안 주민들이 계속 이사하고 참조를 바꾸면 지도를 그릴 수 없다. 그래서 GC의 일부 단계는 **모든 애플리케이션 스레드를 세운다.** 단체 사진과 같다 — 셔터가 열린 동안 움직이면 사진이 흐려지므로, 전원 정지.

```text
앱 스레드들:  ━━━━━━━━▌▌━━━━━━━━━━━━▌▌━━━━━━━━━━━▌▌▌▌━━━━━
                      ▲▲            ▲▲           ▲▲▲▲
                   young GC      young GC     mixed GC (조금 김)
                   (수~수십 ms)

▌= Stop-The-World. 이 동안 모든 요청 처리가 멈춘다 → 응답 시간 꼬리(p99)의 주범
```

세우는 시점도 아무 데서나가 아니다. 각 스레드는 JIT가 코드 곳곳에 심어 둔 **안전지점(safepoint)** 까지 달려가서 정차하고, 전원이 정차해야 GC가 시작된다. GC 튜닝이란 결국 **이 멈춤의 길이·빈도와 처리량 사이의 흥정**이다.

### 8.3 G1의 평상시와 비상시

```text
            ┌────────────────── 평상시 ──────────────────┐
            │   young GC 반복 (Eden 차면 비움, 짧은 STW) │
            └──────────────────┬─────────────────────────┘
                               │ Old 영역 점유가 임계치 도달 (기본 45 %, 이후 적응형)
                               ▼
            동시 마킹 (concurrent marking)
            — 앱을 세우지 않고 나란히 달리며 Old 리전들의 생존자 지도 작성
                               ▼
            mixed GC 몇 차례
            — young + "쓰레기 비율 높은 Old 리전"만 골라 회수 (Garbage-First)
                               ▼
            다시 평상시로
            ───────────────────────────────────────────────
            이 모든 게 실패하면 (할당 속도를 청소가 못 따라감):
            Full GC = 힙 전체 STW 대청소. G1에선 "비상사태"이며
            발생 자체가 튜닝 신호다.
```

GC 알고리즘별 비교와 선택 기준은 [gc.md](./gc.md) 참조.

### 8.4 Collections / Pause Durations 패널

- **Collections** — 단위 시간당 GC 횟수. 메트릭의 cause/action 태그로 어떤 GC였는지(`G1 Evacuation Pause`=young, `G1 Humongous Allocation`=대형 객체가 유발 등) 구분된다. young GC가 주기적으로 찍히는 건 호흡 같은 정상 동작이고, 주기가 점점 짧아지는 게 신호다.
- **Pause Durations** — 멈춤의 평균/최대 길이. 경험칙으로 young GC 수십 ms 이하면 양호, max가 갑자기 수백 ms~초 단위로 튀면 mixed/Full GC 또는 힙 부족을 의심한다. **이 값이 곧 응답 시간 p99에 그대로 더해진다**는 점이 이 패널의 존재 이유다.

### 8.5 Allocated/Promoted — 출생률과 이주율

- **Allocated** (`jvm_gc_memory_allocated_bytes_total`의 기울기) — **할당률**: 초당 몇 바이트가 Eden에서 태어나는가. 도시의 출생률. (정확히는 young 영역 할당분만 집계하므로, 리전을 통째로 차지하며 Old에 곧장 들어가는 Humongous 대형 객체는 빠진다.)
- **Promoted** (`jvm_gc_memory_promoted_bytes_total`의 기울기) — **승격률**: 초당 몇 바이트가 Old로 이주하는가.

이 두 값이 GC 리듬을 통째로 결정한다.

```text
young GC 주기 ≈ Eden 크기 ÷ 할당률
   예: Eden 300 MiB, 할당률 30 MB/s  →  약 10초마다 young GC

Old가 차는 속도 ≈ 승격률
   승격률이 높으면 → mixed GC 잦아짐 → 최악엔 Full GC
```

할당률은 코드 품질의 간접 지표이기도 하다(루프 안의 불필요한 객체 생성, 과도한 문자열 연결 등). 승격률이 높은 전형적 원인은 둘이다 — 정말 오래 사는 데이터가 늘었거나(캐시 대량 적재), **임시 객체가 GC를 두세 번 버틸 만큼 요청 처리가 길어져 "조기 승격"** 되고 있거나.

---

## 9. Classloading — 설계도는 어떻게 반입되는가

### 9.1 사서들의 위계 — 로더 계층과 위임

클래스 파일을 찾아 Metaspace에 올리는 주체가 클래스로더다. 도서관 사서의 위계로 구성돼 있고, 요청 처리 방식이 독특하다: **자식은 먼저 부모에게 묻는다(부모 위임 모델).**

```text
        Bootstrap 로더          java.base 핵심 (String, List ...)
              ▲ "그 책 있어요?"
        Platform 로더           JDK 부속 모듈
              ▲ "그 책 있어요?"
        Application 로더        클래스패스: 내 코드 + 라이브러리 jar
              ▲
        (커스텀 로더들)          컨테이너, DevTools, 플러그인 시스템 ...
```

위임의 이유는 효율이 아니라 **신뢰**다. 누가 악의적으로 `java.lang.String`이라는 이름의 클래스를 클래스패스에 심어도, 그 요청이 항상 Bootstrap까지 먼저 올라가므로 진짜 String이 이긴다. 핵심 클래스 위조가 구조적으로 차단된다.

또 하나의 성질은 **게으름**이다. 클래스는 미리 다 로드되지 않고 **처음 사용되는 순간** 로드된다. 기동 직후 그래프가 가파르게 오르다 수평이 되는 이유다.

### 9.2 두 패널 읽기

- **Classes loaded** — 현재 로드돼 있는 클래스 수. Spring Boot 앱은 프레임워크+라이브러리 덕에 보통 1만~3만 개 선에서 수평을 이룬다.
- **Class delta** — 구간당 로드(−언로드) 변화량. 평상시 0 부근이 정상.

이상 신호는 Metaspace(7.1)와 정확히 짝을 이룬다. **정상 기동이 끝난 뒤에도 loaded가 꾸준히 우상향**하면, 런타임에 클래스를 계속 찍어내는 무언가 — CGLIB 프록시 남발, 리플렉션·스크립트 엔진의 동적 클래스 생성, 클래스로더 누수 — 가 있다는 뜻이다. 클래스는 로더째로만 회수되므로 이 우상향은 결국 Metaspace 고갈로 끝난다.

---

## 10. Buffer Pools — 성벽 밖 보세창고

### 10.1 왜 힙 밖 버퍼가 필요한가

네트워크 카드나 디스크에서 데이터를 받는 동안, 받는 쪽 메모리의 **주소가 고정**돼 있어야 한다. 그런데 6.3에서 봤듯 GC는 살아 있는 객체를 수시로 복사하며 **이사**시킨다. 힙 안의 버퍼는 전입신고가 무의미한, 주소가 흔들리는 집이다.

그래서 NIO는 **GC가 건드리지 않는 성벽 밖(네이티브 메모리)에 창고**를 둔다 — 그것이 다이렉트 버퍼(`ByteBuffer.allocateDirect`)다. 주소가 고정되니 OS·하드웨어와 직거래(DMA)가 가능하고, 복사 횟수가 준다.

```text
[힙 버퍼로 읽기]
  디스크/NIC ─DMA→ 커널 버퍼 ─복사→ 임시 다이렉트 버퍼 ─복사→ 힙 ByteBuffer
                                  (이사 다니는 힙에 곧장 못 주므로 한 번 경유)

[다이렉트 버퍼로 읽기]
  디스크/NIC ─DMA→ 커널 버퍼 ─복사→ 다이렉트 버퍼 (끝. 주소 고정 단독주택)
```

### 10.2 direct 패널 읽기 — used 33.6 MB, capacity 33.6 MB, buffers 20

- **count(20)** — 살아 있는 다이렉트 버퍼 개수
- **capacity** — 그 버퍼들 크기의 합
- **used ≈ capacity가 정상이다.** 다이렉트 버퍼는 생성 순간 전액이 네이티브에서 할당되므로 둘이 거의 항상 같다. 힙처럼 "차오르는" 그림이 아니다.

주 고객은 톰캣 NIO 커넥터와 Netty다. 캡처의 20개/33.6 MB는 NIO 기반 서버의 전형적인 평시 모습이다.

조심할 함정 하나 — **회수 타이밍**. 다이렉트 버퍼의 네이티브 메모리는, 힙에 있는 조그만 손잡이 객체(ByteBuffer 핸들)가 GC로 수거될 때 따라서 해제된다. **가방 본체는 성벽 밖 창고에 있고 힙에는 손잡이만 있는** 구조라, 힙이 한가해서 GC가 안 돌면 창고가 안 비워진다. 힙은 텅텅한데 네이티브 메모리 부족(`OutOfMemoryError: Direct buffer memory` — 한도는 `-XX:MaxDirectMemorySize`, 기본값은 대략 -Xmx와 같음)이 나는 기묘한 장애의 원인이다. Netty가 자체 풀링으로 할당·해제 자체를 회피하는 이유이기도 하다.

### 10.3 mapped, 그리고 non-volatile — 0 B

- **mapped** — `FileChannel.map()`으로 **파일을 메모리 주소에 직접 매핑**(mmap)한 영역. 파일을 read() 호출 없이 배열 읽듯 접근하며 OS 페이지 캐시와 직통이다. 창고에 물건을 옮겨오는 대신 **창고 문을 열어 두고 선반을 직접 보는** 방식. Kafka, Lucene/Elasticsearch가 애용한다. 0 B = 이 앱은 안 쓴다는 뜻, 그뿐이다.
- **mapped - 'non-volatile memory'** — 영속 메모리(PMEM) 위에 매핑된 버퍼용 풀(JDK 14+, JEP 352). 해당 하드웨어를 쓰는 극소수 환경이 아니면 평생 0 B다.

---

## 11. 실측 진단 — 이 대시보드의 숫자 읽기

캡처(2026-06 시점 스냅숏) 값을 패널 순서대로 판정한다.

| 패널 | 값 | 해석 | 판정 |
| --- | --- | --- | --- |
| I/O Rate / Duration | (값 있음) | 트래픽 소량 유입, 정상 계측 | 정상 |
| I/O Errors | No data | 5xx가 0건이라 시계열 미생성 (3.2) | 정상 (좋은 무소식) |
| I/O Utilisation | No data | 톰캣 메트릭 미노출 또는 비톰캣 (12.2) | 설정 사안 |
| Heap used | 83.7 MiB (기간 max 369) | 점유 16 %, GC 후 정상 회수 확인 | 양호 |
| Heap committed=max | 512 = 512 MiB | -Xms=-Xmx 고정 힙 관행으로 추정 | 정상 |
| Total max | 1.73 GiB | 512(힙)+240(코드)+1024(CCS) 예약 합산 (4.3) | 숫자만 클 뿐 정상 |
| Process Memory | No data | micrometer-jvm-extras 미설치 (12.3) | 권장 보강 |
| CPU process | 0.0 % (system 0.3 %) | 사실상 유휴 | 한가 |
| Load 1m | 0.0 (max 0.5) / 4 cpus | 계산대 4개에 줄 없음 | 한가 |
| Threads | live 73, daemon 25, peak 77 | peak≈live, 증가 추세 없음 | 안정 |
| Thread States | RUNNABLE 42, BLOCKED 0 | RUNNABLE 다수는 epoll 대기(5.4), 락 경합 없음 | 건강 |
| GC Pressure | 0.0 % | 청소 부담 없음 | 양호 |
| File Descriptors | 93 / 66 K (≈65,536) | 사용률 0.14 % | 여유 |
| Buffer direct | 33.6 MB / 20개 | NIO 서버 전형치, used=capacity 정상 (10.2) | 정상 |
| Buffer mapped | 0 B | mmap 미사용 | 정상 |

검산 메모 하나. 원본 기록에서 Non-Heap 행이 Heap 행과 같은 "committed 512 / max 512"로 적혀 있는데, Total과의 산수가 맞지 않는다. Total이 합산 패널이므로 실제 화면은 다음이었을 가능성이 높다.

```text
(아래는 캡처 실측이 아니라 Total − Heap 역산으로 얻은 추정값이다)
Non-Heap used      = Total used      − Heap used      = 162 − 83.7 ≈  78 MiB
Non-Heap committed = Total committed − Heap committed = 594 − 512  =  82 MiB
Non-Heap max       = Total max       − Heap max       ≈ 1.73 GiB − 512 MiB ≈ 1.23 GiB
                                                        (= CodeCache 240 + CCS 1024)
```

78 MiB의 내용물은 전형적으로 Metaspace 수십 MiB + CodeHeap 십수 MiB + CCS 수 MiB다.

종합하면: **트래픽이 거의 없는 소형 서비스가 고정 힙 512 MiB로 안정 운행 중이며, 메모리·스레드·GC·FD 전 영역에 이상 신호가 없다.** 이 상태에서 할 일은 튜닝이 아니라 ① No data 패널 보강(12장)과 ② 추세 기반 경보 설정(13장)이다.

---

## 12. No data 처방전

1.1의 사슬(JVM → Micrometer → Prometheus → Grafana)을 기억하면, No data 진단은 "이 메트릭이 사슬 어느 단계부터 없는가"를 찾는 일이다. 확인 명령은 공통이다:

```bash
curl -s http://<app>:<port>/actuator/prometheus | grep -E '^(tomcat_|process_memory_|http_server_requests)'
# 여기 없으면 ① ② 단계(앱) 문제, 여기 있는데 Grafana에 없으면 ③ ④ 단계(수집·질의) 문제
```

### 12.1 I/O Errors — 고칠 것 없음

5xx 카운터가 아직 태어나지 않았을 뿐이다(1.2, 3.2). 빈 패널이 싫다면 Grafana 패널 옵션 "No value → 0" 또는 쿼리에 `or vector(0)` 보정. 기능적으로는 그대로 둬도 알람(13장)에는 지장 없다.

### 12.2 I/O Utilisation — 두 갈래 진단

- **톰캣(Spring MVC)인 경우**: `tomcat_threads_*` 메트릭은 JMX MBean에서 읽는데, Spring Boot 2.2부터 톰캣 MBean 레지스트리가 기본 꺼져 있다. 켜면 노출된다.

  ```yaml
  server:
    tomcat:
      mbeanregistry:
        enabled: true
  ```

- **WebFlux/Netty인 경우**: 톰캣 스레드 풀 자체가 없으므로 이 패널은 원래 비는 게 맞다. 이벤트 루프 모델은 "워커 가동률" 개념이 달라서, 필요하면 reactor-netty 메트릭으로 별도 패널을 만든다.

### 12.3 JVM Process Memory — 라이브러리 추가

이 패널이 그리는 `process_memory_rss_bytes`/`process_memory_vss_bytes`는 기본 Micrometer에 없고, 4701 대시보드가 공식 안내하는 **micrometer-jvm-extras**(작성 시점 기준 0.2.2)가 `/proc` 기반으로 제공한다. 리눅스 계열 전용이다.

```groovy
implementation 'io.github.mweirauch:micrometer-jvm-extras:0.2.2'
```

```java
@Bean
public ProcessMemoryMetrics processMemoryMetrics() {
    return new ProcessMemoryMetrics();
}

@Bean
public ProcessThreadMetrics processThreadMetrics() {
    return new ProcessThreadMetrics();
}
```

컨테이너 운영(메모리 limit/OOMKill 대응)이라면 4.4에서 본 이유로 보강 가치가 가장 큰 패널이다.

### 12.4 mapped 0 B는 No data가 아니다

시계열이 존재하고 값이 0인 것이다(mmap 미사용). 1.2의 구분 그대로 — "0"과 "No data"는 다른 진단이다.

---

## 13. 무엇을 경보로 걸까 — 운영 체크리스트

순간값 경보는 오탐이 많다. 이 대시보드에서 가치 있는 신호는 대부분 **추세와 비율**이다.

| 우선순위 | 신호 | 경보 기준 출발점 (경험칙) | 무엇의 징후인가 | 근거 |
| --- | --- | --- | --- | --- |
| ★★★ | Old Gen used의 **GC 후 바닥값** 우상향 | 수 시간~수일 추세 | 메모리 누수 | 4.1, 6.5 |
| ★★★ | Full GC 발생 | 1건이라도 | 힙 부족·누수 말기 | 8.3 |
| ★★★ | 5xx Errors rate | 0 초과 지속 | 장애 | 3.2 |
| ★★☆ | GC pause max | p99 > 0.5~1 s | 응답 꼬리 지연 직결 | 8.4 |
| ★★☆ | GC Pressure | > 5~10 % | 힙 부족, 할당 과다 | 5.5 |
| ★★☆ | error 로그율 | 평시 대비 급증 | 가장 빠른 조기 경보 | 5.6 |
| ★★☆ | live threads | 단조 증가 추세 | 스레드 누수 → [thread_leak.md](./thread_leak.md) | 5.3 |
| ★★☆ | FD open ÷ max | > 80 % | fd 누수, Too many open files 임박 | 5.7 |
| ★☆☆ | Metaspace used + Classes loaded | 동반 우상향 | 클래스로더·프록시 누수 | 7.1, 9.2 |
| ★☆☆ | direct buffer used | 단조 증가 | 네이티브 버퍼 누수 | 10.2 |
| ★☆☆ | CodeHeap used 합 | 240 MiB 근접 | JIT 비활성화 → 성능 급락 | 7.3 |
| ★☆☆ | Tomcat Utilisation | > 0.8 지속 | 워커 고갈, 대기열 형성 | 3.4 |

---

## 부록 A. 패널-메트릭-출처 매핑표

"이 그래프가 정확히 뭘 질의하나"가 필요할 때.

| 패널 | Prometheus 메트릭 (대표) | JVM 안의 출처 |
| --- | --- | --- |
| I/O Rate · Errors · Duration | `http_server_requests_seconds_count/_sum/_max` | Spring MVC/WebFlux 요청 계측 |
| I/O Utilisation | `tomcat_threads_busy_threads` ÷ `tomcat_threads_config_max_threads` | Tomcat JMX MBean |
| JVM Heap / Non-Heap / Total | `jvm_memory_used/committed/max_bytes{area="heap"/"nonheap"}` | MemoryMXBean·MemoryPoolMXBean |
| JVM Process Memory | `process_memory_rss_bytes`, `process_memory_vss_bytes` | micrometer-jvm-extras (/proc/self/status) |
| CPU Usage | `system_cpu_usage`, `process_cpu_usage` | OperatingSystemMXBean |
| Load | `system_load_average_1m`, `system_cpu_count` | OperatingSystemMXBean |
| Threads | `jvm_threads_live/daemon/peak_threads` | ThreadMXBean |
| Thread States | `jvm_threads_states_threads{state=...}` | ThreadMXBean (Thread.State 집계) |
| GC Pressure | `jvm_gc_pause_seconds_sum` 비율 (리비전에 따라 동시 단계 포함) | GC 알림(GarbageCollectorMXBean) |
| Log Events | `logback_events_total{level=...}` | Logback 카운터 훅 |
| File Descriptors | `process_files_open_files`, `process_files_max_files` | UnixOperatingSystemMXBean |
| Memory Pools (Heap/Non-Heap) | `jvm_memory_*_bytes{id="G1 Eden Space" 등}` | 풀별 MemoryPoolMXBean |
| GC Collections / Pauses | `jvm_gc_pause_seconds_count/_sum/_max{action,cause}` | GC 알림 이벤트 |
| GC Allocated/Promoted | `jvm_gc_memory_allocated_bytes_total`, `jvm_gc_memory_promoted_bytes_total` | GC 전후 풀 크기 변화 적산 |
| Classes loaded / delta | `jvm_classes_loaded_classes`, `jvm_classes_unloaded_classes_total` | ClassLoadingMXBean |
| Buffer Pools | `jvm_buffer_memory_used_bytes`, `jvm_buffer_total_capacity_bytes`, `jvm_buffer_count_buffers{id="direct"/"mapped"}` | BufferPoolMXBean |

## 부록 B. 한 줄 용어 사전

- **바이트코드** — javac가 만드는 가상 CPU(JVM)용 명령어. 플랫폼 독립성의 원천.
- **MXBean** — JVM이 자기 상태를 표준 인터페이스로 노출하는 내장 계측 창구.
- **Micrometer** — 그 창구를 읽어 Prometheus 등 다양한 백엔드 형식으로 바꿔 주는 계측 SDK (메트릭계의 SLF4J).
- **scrape** — Prometheus가 `/actuator/prometheus`를 주기적으로 긁어 가는 수집 행위.
- **카운터/게이지** — 단조 증가 누적값(요청 수)/오르내리는 현재값(힙 used). 카운터는 기울기(rate)로 읽는다.
- **used / committed / max** — 주차된 차 / 포장된 주차면 / 허가받은 부지 한도 (2.3).
- **RSS** — OS가 실측한 프로세스 거주 메모리. 컨테이너 OOMKill의 기준.
- **GC root** — 도달 가능성 판정의 출발점(스택 지역변수, static, JNI 핸들).
- **STW(Stop-The-World)** — GC가 객체 그래프를 안전하게 보려고 앱 스레드 전체를 세우는 구간.
- **safepoint** — 스레드가 정차할 수 있도록 JIT가 코드에 심어 둔 안전지점.
- **약한 세대 가설** — "대부분의 객체는 어려서 죽는다". 세대별 GC의 존재 근거.
- **Eden / Survivor / Old** — 신생아 구역 / 나이 도장 찍는 검증 기숙사 2동 / 정착촌.
- **승격(promotion)** — Survivor에서 나이가 찬 객체가 Old로 이주하는 것.
- **조기 승격** — 요청 처리가 길어져 임시 객체가 GC를 버티고 Old로 밀려가는 병리.
- **리전(region)** — G1이 힙을 나눈 같은 크기의 바둑판 칸(1~32 MiB). 역할(E/S/O)이 유동적.
- **Humongous** — 리전 절반 이상 크기의 객체. 연속 리전을 통째로 점유하는 골칫거리.
- **minor(young) / mixed / Full GC** — young만 / young+Old 일부 / 힙 전체(비상) 청소.
- **할당률 / 승격률** — 초당 Eden 출생량 / 초당 Old 이주량. GC 리듬의 결정 변수.
- **Metaspace** — 클래스 메타데이터(설계도) 보관소. JDK 8에서 PermGen을 대체한 네이티브 영역.
- **Compressed Class Space** — 클래스 포인터를 32비트로 접기 위해 설계도를 모아 두는 연속 영역(기본 1 GiB 예약).
- **compressed oops** — 힙 32 GiB 이하에서 객체 참조를 32비트 오프셋으로 압축하는 기법.
- **JIT / C1 / C2** — 실행 중 번역기 / 빠른 초벌 번역(+프로파일 수집) / 프로파일 기반 완성 번역.
- **nmethod** — 컴파일된 자바 메서드의 HotSpot 내부 명칭. CodeHeap 패널 이름의 정체.
- **탈최적화(deoptimization)** — 번역의 전제가 깨져 기계어를 버리고 인터프리터로 복귀하는 것.
- **워밍업** — 배포 직후 JIT 번역이 끝나기 전까지 느린 구간.
- **데몬 스레드** — JVM 종료를 막지 못하는 보조 스레드(유저 스레드가 다 끝나면 같이 종료).
- **fd(파일 디스크립터)** — 파일·소켓·파이프 등 "열려 있는 것"의 핸들. 출입문 수.
- **다이렉트 버퍼** — GC가 이사시키지 않는 힙 밖 고정 주소 버퍼. NIO·DMA의 거점.
- **mmap(mapped buffer)** — 파일을 메모리 주소처럼 접근하게 매핑한 것.

## 부록 C. 참고 자료

- Grafana 대시보드 "JVM (Micrometer)": https://grafana.com/grafana/dashboards/4701
- Micrometer 문서: https://micrometer.io/docs
- micrometer-jvm-extras: https://github.com/mweirauch/micrometer-jvm-extras
- HotSpot GC 튜닝 가이드 (JDK 17): https://docs.oracle.com/en/java/javase/17/gctuning/
- JEP 197 (Segmented Code Cache): https://openjdk.org/jeps/197
- JEP 122 (PermGen 제거): https://openjdk.org/jeps/122
- JEP 352 (Non-Volatile Mapped Byte Buffers): https://openjdk.org/jeps/352
