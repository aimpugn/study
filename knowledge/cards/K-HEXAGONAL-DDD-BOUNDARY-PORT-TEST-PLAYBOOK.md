# K-HEXAGONAL-DDD-BOUNDARY-PORT-TEST-PLAYBOOK

## 1) 이 문서가 답하는 질문

Hexagonal + DDD로 설계/리팩토링/테스트할 때 반복되는 핵심 질문을 한 문서로 고정합니다.

- 중복 코드를 언제 공통화하고, 언제 중복을 유지해야 안전한가?
- 공통 유틸/매핑/파싱 코드는 `core`/`adapter`/`server` 중 어디에 둬야 하는가?
- Repository Port 시그니처에는 어떤 타입이 등장해야 하는가?
- DB row/JSON payload 같은 저장 표현(representation)의 해석 책임은 누구에게 두는가?
- `core` 정책 테스트, `adapter` primitive 테스트, `server` 통합 테스트를 어떻게 분리해야 하는가?
- fake(in-memory)와 JDBC 구현의 의미 drift를 어떻게 막는가?

핵심 목표는 “좋아 보이는 구조”가 아니라, 변경 비용/실패 모드/검증 절차를 재현 가능하게 만드는 것입니다.

---

## 2) 배경: 왜 이 주제가 비싸게 터지는가

이 영역의 비용은 두 축에서 반복적으로 발생합니다.

첫째, **중복 방치 비용(drift)** 입니다. 같은 의미의 로직이 여러 곳에 있으면 시간이 지나며 구현이 갈라지고, 특정 경로/특정 데이터에서만 실패하는 비결정성이 생깁니다.

둘째, **성급한 공통화 비용(wrong abstraction)** 입니다. 의미가 다른 구현을 억지로 하나로 합치면 결합도가 커지고, 한 번의 실수가 여러 모듈에 동시에 전파됩니다.

Hexagonal 경계에서는 여기에 추가 비용이 생깁니다.

- `core`가 인프라 의미(DB/JDBC/JSON 프레임워크)를 알게 되면 의존성 방향이 깨집니다.
- `adapter` 테스트가 `core.internal`을 직접 import 하면 경계가 테스트에서 먼저 붕괴합니다.
- 테스트가 목적별로 분리되지 않으면 정책 회귀와 저장소 회귀가 섞여 원인 축소가 어려워집니다.

---

## 3) 제어점: 판단을 고정하는 6개 축

### 3.1 중복의 의미 동등성

겉보기로 비슷한 코드가 아니라, **같은 의미를 표현하는 로직인지** 먼저 판정합니다.
의미가 다르면 공통화는 wrong abstraction입니다.

### 3.2 의미 소유권(Owner of Meaning)

그 규칙이 바뀔 때 누가 정당한 변경 주체인지로 위치를 정합니다.

- 도메인 규칙: `core`
- 저장/전송/직렬화 표현: `adapter` 또는 `server`(인프라 경계)
- 특정 VAN 프로토콜/필드 문법: VAN `adapter`

### 3.3 Port 시그니처 타입 소유권

Port는 core가 외부에 요구하는 능력 선언입니다.
따라서 Port에 등장하는 타입은 **core 소유 타입**이어야 합니다.
Port가 adapter 타입을 노출하면 core가 adapter를 알아야 하므로 의존성 방향이 붕괴합니다.

### 3.4 표현(Representation) 소유권

DB row/JSON payload가 “도메인 계약”인지 “인프라 스냅샷”인지 분리합니다.

- 인프라 스냅샷이면 adapter가 파싱/직렬화를 소유
- 도메인 계약이면 core가 파싱을 소유 가능
  단, 스키마 버전/하위 호환/필수 필드 규칙까지 core가 함께 소유해야 함

### 3.5 정책(Policy) vs Primitive 분리

- 정책: core 유스케이스가 조합하는 불변식 실행 규칙(멱등 수렴, 충돌, replay 등)
- primitive: adapter 저장소가 제공하는 원자 연산(유니크 INSERT, 조건부 UPDATE, claim 등)

정책 테스트가 primitive 구현에 붙거나, primitive 테스트가 정책 구현에 붙으면 실패 원인 분리가 깨집니다.

### 3.6 테스트 결정성(Determinism)

랜덤/`Thread.sleep`/환경 의존 테스트는 회귀 탐지기보다 잡음원이 되기 쉽습니다.
가능하면 결정적 입력/결정적 지터/명시적 시간 주입으로 재현성을 고정합니다.

---

## 4) 기본 결론(운영 기본값)

### 4.1 코드 배치 기본값

- 중복 발견 즉시 공통화하지 않고, 의미가 안정화될 때까지 제한적 중복 허용
- 공통화 시 “재사용 최대화”보다 “의미 소유 경계”를 우선
- 도메인 타입이 DB 레코드 타입을 직접 아는 구조(`fromRecord`)는 기본적으로 피함

### 4.2 Port/매핑/파싱 기본값

- Port 시그니처에는 core 타입만 노출
- adapter 내부 DB row/entity는 adapter에 캡슐화
- adapter가 row -> core 타입으로 매핑해 Port 구현
- 인프라 스냅샷 payload(JSON dump 등) 파싱은 adapter에서 처리

### 4.3 테스트 배치 기본값

- **core 정책 테스트**: in-memory port double 사용, 빠르고 결정적
- **adapter primitive 테스트**: H2 등 실제 스키마/제약으로 DB 동작 관측
- **server 통합 테스트**: 조립/엣지(HTTP/TCP/스케줄러/트랜잭션 프록시) 검증에 한정
- fake/JDBC 의미 동등성: 공용 contract suite로 자동 검증

---

## 5) 실패 모드와 관측 신호

### 5.1 의미가 다른 로직을 공통화한 경우

- 증상: 길이/프레임/체크섬 오류, 특정 데이터만 파싱 실패
- 관측: “expected length != actual length”, decode error 증가

### 5.2 의미가 같은 로직을 중복 방치한 경우

- 증상: 경로별 결과 불일치(한 경로는 trim, 다른 경로는 no-trim)
- 관측: 동일 입력의 경로별 출력 차이, 특정 케이스만 실패

### 5.3 Port가 adapter 타입을 노출한 경우

- 증상: core 단독 컴파일 불가, adapter 변경이 core 수정으로 전파
- 관측: core 모듈에 adapter 의존 추가 필요, 빌드 규칙 붕괴

### 5.4 adapter 테스트가 core internal에 결합된 경우

- 증상: core 내부 리팩토링이 adapter 테스트 컴파일 실패 유발
- 관측: adapter 테스트 import에 `core.internal` 등장

### 5.5 정책 테스트가 DB에 과결합된 경우

- 증상: 느리고 간헐적인 실패, 원인 분리 불가
- 관측: 정책 회귀와 DB 제약 회귀가 같은 테스트에서 동시 발생

---

## 6) 권장 구조(요약)

```text
core
  - Port (core-owned types only)
  - UseCase/Policy
  - InMemory Port Doubles
  - Contract Test Suite (Port semantics)

adapter-jdbc
  - JDBC Row/Entity/SQL (adapter-owned)
  - Port Implementation
  - H2 Test Fixture + Primitive Tests
  - Contract Test Suite Adapter Binding

server
  - Composition Root
  - HTTP/TCP/Scheduler Edge
  - Integration Tests (wiring/proxy/edge behavior)
```

---

## 7) 최소 검증 실험(빠른 체크)

### 7.1 경계 위반 탐지

```bash
rg -n "import .*core\\.internal" adapters -S --glob "*Test.java"
```

기대: 0건

### 7.2 server 테스트의 adapter concrete 결합 탐지

```bash
rg -n "^import .*jdbc\\." bridge-server/src/test/java -S
```

기대: 정책 테스트 영역에서는 0건

### 7.3 계약 동등성 확인

```bash
./gradlew test
```

기대: Port contract suite가 in-memory/JDBC 양쪽에서 PASS

### 7.4 wrong abstraction 관측 미니랩

```java
String field = "A         "; // fixed-length
String normalized = field.trim();
assert field.length() != normalized.length(); // 길이 보존 불변식 깨짐
```

---

## 8) 트레이드오프 정리

- 중복 유지:
  - 장점: 결합도 낮음, 변경 반경 작음
  - 단점: 의미 drift 위험

- 공통화:
  - 장점: 일관성 향상, 샷건 수술 감소
  - 단점: wrong abstraction 시 폭발 반경 확대

선택 기준은 “코드 줄이기”가 아니라 “동시에 같이 바뀌어야 하는 의미가 실제로 안정적인가”입니다.

---

## 9) 외부 리뷰 요청 시 사용할 질문

1. 정책/primitive/core-server 경계 분리가 실패 원인 분리를 충분히 보장하는가?
2. fake 유지 전략의 가장 큰 drift 지점은 어디인가?
3. contract suite 항목(동시성 claim race, timeout 경계, replay 의미)이 충분한가?
4. 어떤 테스트가 통합으로 승격되어야 하고, 어떤 테스트는 단위에 남겨야 하는가?

---

## 10) 실행 체크리스트

- [ ] Port 시그니처에 adapter 타입이 없다.
- [ ] adapter 테스트가 `core.internal`을 import 하지 않는다.
- [ ] 정책 테스트는 port double로 결정적으로 재현된다.
- [ ] adapter primitive 테스트가 DB 제약을 직접 관측한다.
- [ ] fake/JDBC contract 동등성 테스트가 자동화되어 있다.
- [ ] 랜덤/수면 기반 불안정 테스트를 결정적 방식으로 대체했다.

이 체크리스트를 통과하면, Hexagonal + DDD의 경계 이점(변경 영향 국소화, 실패 원인 분리, 테스트 신뢰성)을 유지하면서 확장할 수 있습니다.
