<!-- markdownlint-disable MD007 MD012 -->

# WORK_20260409_HTTP_CLIENT_TIMEOUT

## 0. Meta

- 작업 제목: HTTP 클라이언트 timeout 모델의 수렴 이유와 예외 사례 정리
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260409_HTTP_CLIENT_TIMEOUT.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `research | explain | execute`
- 작업 깊이: `full`
- 관련 요청: `study/web` 아래에 사람이 이해하기 쉬운 개념 문서 작성
- 원문 사용자 요청: HTTP 클라이언트 timeout이 왜 요청별보다는 단일 timeout이나 클라이언트 기본값 쪽으로 수렴하는지, 그 배경과 역사와 낮은 계층 맥락, 그리고 다르게 제공하는 라이브러리들까지 조사해서 정리
- 대상 경로 / 자산:
  - `/Users/rody/VscodeProjects/study/web/http/http_client_timeout.md`
- 실행자: Codex
- 시작 일시: 2026-04-09
- 종료 일시: 2026-04-09 23:18:15 KST
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `verify + commit`

## 1. Request Normalization

- goal: HTTP timeout API가 왜 지금처럼 생겼는지 표준, 구현, 역사, 낮은 계층, 라이브러리 사례를 묶어 재구성 가능한 문서로 남긴다.
- refs: 공식 문서, RFC, Linux man page, 주요 라이브러리 공식 문서
- scope: timeout 모델 설명 문서 1개 작성과 근거 ledger 정리
- mode: `research + explain + execute`
- run_mode: `normal`
- finish: `verify + commit`
- must_keep:
  - 표준이 직접 강제했는지 여부를 분리해서 설명
  - 네트워크/소켓/커널/OS 관점 포함
  - 역사와 발전 과정 포함
  - 다르게 제공하는 라이브러리 정리
  - 사람이 이해하기 쉬운 설명, ASCII, pseudo code, 비유 포함
- extra_checks:
  - timeout 종류를 먼저 분리해서 설명
  - "요청별 timeout"이 정확히 무엇을 뜻하는지도 분해

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구:
  - `study/web` 아래에 문서 생성
  - 표준/실무/기술적 배경과 역사 정리
  - 낮은 계층까지 연결
  - 예외적인 라이브러리 사례 정리
  - 사람이 이해하기 쉬운 설명
- 사용자가 명시한 금지 사항:
  - 문서 안에 메타 문구로 설명 방식 자체를 선언하지 않는다
- path / naming / format / finish 관련 요구:
  - Markdown 문서
  - `/Users/rody/VscodeProjects/study/web` 하위
  - 저장소 변경 작업이므로 검수 후 커밋
- 내가 추가한 누락 방지 항목:
  - timeout 종류 표를 먼저 둔다
  - 요청 전체 마감시간과 소켓 timeout을 혼동하지 않게 분리한다
  - HTTP/2, HTTP/3가 왜 중요한지 연결한다

### 1.2 Non-Goals

- 이번 작업의 비범위:
  - 특정 언어에서 timeout 설정 코드를 완전한 튜토리얼로 만드는 일
  - 실제 벤치마크나 패킷 캡처 실험 추가
  - 기존 문서 대규모 개편
- 지금 하지 않는 이유:
  - 사용자 요청은 원인과 배경을 이해하는 문서 작성이 중심이다

## 2. Root-First Framing

- 근본 문제: 개발자는 "request"를 기준으로 사고하지만, 낮은 계층과 연결 관리 계층은 "socket / connection / stream / pool"을 기준으로 동작한다. 이 틈 때문에 timeout API가 직관과 다르게 보인다.
- 왜 이 문제가 지금 중요한가: timeout 설정 실수는 장애, 오해, 과도한 재시도, 잘못된 SLA 판단으로 이어지기 쉽다.
- 작업 목표: timeout이 어느 계층의 책임인지 분리해서 설명하고, 라이브러리 API 차이가 우연이 아니라 배경이 있는 선택임을 보인다.
- 기대 이점:
  - timeout 이름만 보고 성급히 설정하지 않게 됨
  - 클라이언트 기본 설정과 요청별 deadline을 구분하게 됨
  - 다른 언어/프레임워크로도 전이 가능한 모델을 갖게 됨
- 이점이 닫혔다고 판단할 확인 기준:
  - 문서만 읽고도 connect timeout, socket timeout, 전체 deadline의 차이를 설명할 수 있음
  - 왜 HTTP/2 이후 per-request read timeout이 애매해지는지 설명할 수 있음
- 하드 제약 / 호환성 경계:
  - 공식/1차 자료 우선
  - 문서는 기존 study 저장소 톤보다 더 높은 명료성을 목표로 함
- 성공 정의:
  - 문서가 표준 vs 구현 수렴을 구분하고, 낮은 계층과 라이브러리 사례를 연결한다
- PARTIAL 조건:
  - 문서는 썼지만 근거 링크나 예외 사례가 약함
- BLOCKED 조건:
  - 공식 자료 접근 실패나 저장소 쓰기 실패

## 3. Reader & Internalization Contract

- 주 독자: 웹/백엔드 개발을 하는 한국어 독자
- 독자가 이미 알고 있다고 가정하는 것: HTTP 요청/응답의 기본 개념, TCP가 연결형이라는 정도
- 이 작업 결과를 통해 나중에 스스로 설명할 수 있어야 하는 것:
  - 왜 connect timeout은 요청별보다 연결/클라이언트 쪽에 잘 붙는가
  - 왜 request deadline은 요청별로 주기 쉬운가
  - 왜 일부 라이브러리는 per-request override를 제공하지만 기본 모형은 여전히 계층형인가
- 사용자가 내재화해야 할 사고 패턴:
  - API 이름보다 계층과 상태 소유권을 먼저 본다
  - 표준이 규정한 것과 구현이 자연스럽게 수렴한 것을 구분한다
- 특히 막아야 하는 오해:
  - timeout은 하나뿐이라는 오해
  - 요청별 timeout을 지원하지 않으면 라이브러리가 뒤떨어졌다는 오해
  - HTTP RFC가 timeout API 모양까지 정했다고 보는 오해
- 기억 anchor 후보:
  - 도로/차량/배송 마감 비유
  - one request per connection -> pooled connection -> multiplexed streams 흐름
- 반드시 거쳐야 하는 추상화 계층:
  - HTTP RFC
  - 클라이언트 라이브러리
  - connection pool / stream
  - socket / syscall
  - kernel / OS
- 핵심 대조쌍 / 혼동쌍:
  - connect timeout vs request deadline
  - socket read timeout vs response timeout
  - per-request override vs shared client default
  - HTTP/1.0 vs HTTP/1.1 keep-alive vs HTTP/2 multiplexing
- 질문형으로 먼저 답해야 하는 핵심 질문:
  - 왜 요청별 timeout보다 클라이언트 기본값이 더 흔한가
  - 왜 일부 라이브러리는 예외처럼 보이는 모델을 제공하는가
- 이번 작업의 품질 기준 exemplar:
  - 공식 문서의 계층 설명 + 사람이 이해하기 쉬운 서술
- 현재 저장소 평균 품질을 기준으로 삼으면 안 되는 이유:
  - 현재 문서 중에는 개념 구분과 근거 밀도가 약한 것도 있어 기준선으로 삼기 어렵다

## 4. Depth Decision

- 선택한 깊이: `full`
- 왜 이 깊이가 맞는가:
  - 표준, OS, 라이브러리 API, 역사까지 다 연결해야 오해를 줄일 수 있다
- 전체 루프를 켜야 하는 트리거:
  - 주제가 저장소 전체 학습 품질에 큰 영향을 준다
- 축약 가능한 섹션과 그 근거:
  - 기존 문서 동기화는 허브 문서가 없으므로 최소화 가능

## 5. Project Overlay

- 적용한 로컬 AGENTS 경로: `/Users/rody/VscodeProjects/study/AGENTS.md`
- 활성화한 프로젝트 규칙:
  - current 평균 품질을 기준선으로 삼지 않기
  - 공식 자료와 재현 가능한 근거 우선
  - 문서형 작업도 WORK ledger 유지
  - 사람 친화적 설명과 계층 연결 강조
- 특히 중요한 규칙:
  - 근거 없는 단정 금지
  - 설명은 질문형/계층형/비교형을 적극 활용
  - 저장소 변경 작업은 검수 후 커밋
- 전역 규칙과의 충돌 여부 / 해소:
  - 충돌 없음

## 6. Topic Analysis

- 현재 이해한 사용자 의도:
  - 기술적 사실을 넘어서 "왜 그런 설계가 자연스럽게 나왔는지"를 이해하고 싶어 한다
- 현재 보이는 문제 구조:
  - 사용자 직관은 request 중심
  - 실제 구현 제약은 socket/connection 중심
  - 라이브러리는 이 둘을 타협하며 API를 설계
- 핵심 경계:
  - 프로토콜 표준 경계
  - 라이브러리 API 경계
  - OS socket 경계
  - connection pool / multiplexing 경계
- 숨은 가정 / 불확실성:
  - "요청별 timeout"이 connect/read/write를 각각 뜻하는지, 전체 deadline을 뜻하는지 문맥에 따라 다르다
- 성공을 오판하기 쉬운 지점:
  - request-level timeout 사례만 모아놓고 공통 원리를 설명하지 못하는 경우

## 7. Analysis Critique + Repair

### 7.1 Challenge 1

- 어떤 점이 틀릴 수 있는가: 표준이 직접 timeout API 모양을 강제한 것처럼 보일 수 있다
- 보강안: RFC는 프로토콜 의미를, 라이브러리는 실행 모델을 담당한다는 선을 명확히 긋는다
- 왜 이 보강안이 더 강한가: 표준 vs 구현 수렴을 혼동하지 않게 한다

### 7.2 Challenge 2

- 어떤 점이 좁거나 누락될 수 있는가: HTTP/1.1 이야기만 하면 HTTP/2와 HTTP/3 시대 설명이 약해진다
- 보강안: multiplexing과 stream 개념을 반드시 포함한다
- 왜 이 보강안이 더 강한가: 오늘날의 timeout API를 설명하는 데 결정적이다

### 7.3 Challenge 3

- 어떤 점이 설명/학습 목표를 놓칠 수 있는가: 자료를 나열만 하면 사용자가 원리를 못 가져간다
- 보강안: timeout 종류 표, ASCII 그림, 의사코드, 비유를 넣는다
- 왜 이 보강안이 더 강한가: 기억 anchor와 전이 가능성을 만든다

### 7.4 Retained Framing

- 최종 채택한 문제 정의: timeout은 하나가 아니라 계층별 책임의 묶음이고, API 차이는 그 계층을 어디까지 사용자에게 노출하느냐의 차이다
- 폐기한 문제 정의와 이유:
  - "왜 어떤 라이브러리는 불편한가"만 다루는 프레이밍은 역사와 구조를 설명하지 못해 폐기

## 8. Scope Expansion & Impact Sync

- 시작 키워드: `timeout`, `HTTP client`, `per request timeout`
- 확장 키워드(동의어 / 약어 / 원어 / 구명칭 / 관련 에러 / 표준 번호):
  - `connect timeout`
  - `socket timeout`
  - `read timeout`
  - `call timeout`
  - `deadline`
  - `AbortSignal`
  - `SO_RCVTIMEO`
  - `SO_SNDTIMEO`
  - `O_NONBLOCK`
  - `MSG_DONTWAIT`
  - `RFC 7230`
  - `RFC 9113`
  - `RFC 9000`
- 조사한 경로:
  - `web/http`
  - 로컬 AGENTS / WORK template
  - 공식 웹 자료
- 함께 점검한 자산:
  - `/Users/rody/VscodeProjects/study/web/http_connection_flow.md`
  - `/Users/rody/VscodeProjects/study/web/http/http_post.md`
- 함께 움직여야 하는 표면:
  - 이번에는 새 문서 1개면 충분
- 한쪽만 바꾸면 깨질 부분:
  - 별도 허브 문서가 없어 직접적인 링크 드리프트는 적음
- 제외 표면과 근거:
  - 기존 문서 대규모 개편은 사용자 요청 범위를 넘음

## 9. Evidence Gathering

### 9.1 Evidence Ledger

- E-01
  - 주장: HTTP RFC는 timeout API 모양을 직접 표준화하지 않는다
  - 근거 유형: `standard`
  - 자료:
    - RFC 7230
    - RFC 9113
    - RFC 9000
  - 이 자료로 닫힌 것:
    - persistent connection과 stream/multiplexing은 규정하지만, client library API shape는 규정하지 않음
  - 아직 비어 있는 것:
    - 각 언어 생태계가 이를 어떻게 노출하는지는 별도 라이브러리 문서 필요
- E-02
  - 주장: 낮은 계층의 기본 단위는 request가 아니라 socket / file description이다
  - 근거 유형: `official doc`
  - 자료:
    - `socket(7)`
    - `connect(2)`
    - `recv(2)`
  - 이 자료로 닫힌 것:
    - socket options는 socket에 붙고, per-call 예외는 별도 플래그로 존재함
  - 아직 비어 있는 것:
    - HTTP library가 이를 어떻게 추상화하는지는 별도 자료 필요
- E-03
  - 주장: 현대 라이브러리는 timeout을 계층별로 나누거나, 요청 전체 deadline/cancel만 요청별로 노출하는 경향이 있다
  - 근거 유형: `official doc`
  - 자료:
    - JDK HttpClient / HttpRequest
    - Apache HttpComponents config guide
    - Go `net/http`
    - Requests quickstart
    - aiohttp client quickstart
    - libcurl option docs
    - DOM `AbortSignal.timeout`
    - OkHttp Builder Javadoc
  - 이 자료로 닫힌 것:
    - 라이브러리별 차이와 공통 경향
  - 아직 비어 있는 것:
    - 없음

### 9.2 Source Conflicts / Gaps

- 충돌하는 근거:
  - 일부 라이브러리는 per-request override를 제공하지만, 그것이 곧 낮은 계층까지 per-request native라는 뜻은 아님
- 아직 부족한 근거:
  - 없음
- 추론으로만 남는 항목:
  - "실무적으로 왜 이렇게 설계가 선호되었는가" 중 일부는 공식 문서와 구조를 바탕으로 한 추론

## 10. Evidence Critique + Repair

- 소스 품질 리스크:
  - 언어별 공식 문서는 표현 방식이 달라 직접 비교 시 오독 위험이 있음
- 오래되었을 가능성이 있는 가정:
  - OkHttp는 3.x Javadoc을 참고하므로 최신 API와 세부 표현이 조금 다를 수 있음
- 빠진 대안 또는 빠진 근거:
  - Fetch/AbortSignal을 추가해 브라우저/JS 진영의 cancellation 중심 모델 보강
- 근거 세트를 어떻게 보강했는가:
  - RFC, Linux man page, 라이브러리 공식 문서를 함께 사용
- 보강 후에도 남는 한계:
  - 모든 라이브러리를 망라하지는 않음

## 11. Design

- 선택한 접근:
  - 질문형 제목 -> 짧은 직답 -> timeout 종류 분리 -> 낮은 계층 설명 -> 역사 -> 라이브러리 비교 -> 실무 가이드 순서
- 왜 이 접근이 근본 문제와 이점에 가장 잘 맞는가:
  - 독자가 가장 먼저 헷갈리는 지점부터 풀 수 있음
- 고려한 대안:
  - 라이브러리 비교표부터 시작
  - RFC 설명부터 시작
- 대안을 채택하지 않은 이유:
  - 비교표부터 시작하면 원인이 아니라 현상만 보임
  - RFC부터 시작하면 실무 감각이 늦게 옴
- 문서 / 예제 / 자산 구조:
  - 문서 1개
  - ASCII 그림과 pseudo code 포함
- 설명 뼈대: `질문형 + 계층형 + 비교형`
- 계층별 설명 순서:
  - request -> client/pool -> socket -> kernel -> protocol evolution -> library APIs
- 넣을 구체 예시 / 관측 anchor:
  - HTTP/1.0 vs keep-alive vs HTTP/2
  - one socket, many streams 그림
  - socket option vs request deadline pseudo code
- 이 문서를 끌어올릴 목표 수준:
  - 나중에 다른 언어 timeout API를 만나도 구조를 예측할 수 있는 수준
- 실패 모드:
  - timeout 종류 분리가 흐려짐
  - request-level override 사례를 너무 단순화함
- 검증 경로:
  - 직접 읽고 connect timeout / read timeout / request deadline 차이가 재구성되는지 점검
  - 링크와 문장 정확성 재검수

## 12. Design Critique + Repair

### 12.1 Architect View

- 반론: 문서가 너무 설명형으로만 흐르면 표준과 구현 근거가 약해질 수 있음
- 보강 또는 유지 결정: 근거 링크 섹션과 문장별 출처를 명시
- 이유: 신뢰성과 재검증성을 높임

### 12.2 Domain / API Consumer View

- 반론: 사용자는 "실무에서 어떻게 설정할까"를 궁금해할 수 있음
- 보강 또는 유지 결정: 마지막에 실무 판단 기준 섹션 추가
- 이유: 개념 문서를 실무 감각으로 닫기 위함

### 12.3 Newcomer / Learner View

- 반론: socket, stream, file description 개념이 낯설 수 있음
- 보강 또는 유지 결정: 비유, ASCII, 짧은 용어 정의 추가
- 이유: 학습 비용을 낮춤

### 12.4 Final Design Decision

- 최종 채택: 질문형 + 계층형 + 비교형 혼합 문서
- 트레이드오프: 분량은 다소 늘지만 오해 가능성을 줄일 수 있음

## 13. Overall Plan

- 작업 순서:
  - 근거 확인
  - WORK ledger 작성
  - 본문 문서 작성
  - 품질 검수
  - git 검증과 커밋
- 선행 의존성:
  - 공식 자료 링크 확인
- validation order:
  - 자체 문장 검수
  - `git diff --check`
  - 가능하면 markdownlint
- rollback / retry / staging 필요 여부와 이유:
  - docs 작업이므로 별도 staging 단계 불필요

## 14. Plan Critique + Repair

- 계획이 실패할 수 있는 지점:
  - 너무 많은 자료를 넣어 핵심이 흐려질 수 있음
- 순서상 위험:
  - 비교표를 먼저 작성하면 이야기 흐름이 깨질 수 있음
- 빠진 prerequisite:
  - 없음
- 보강안:
  - 문서 앞부분에 직답과 timeout 종류 표를 배치
- 왜 보강된 계획이 더 나은가:
  - 독자가 길을 잃지 않음

## 15. Detailed Task Plan

- 수정 / 생성 / 검토할 파일:
  - 생성: `/Users/rody/VscodeProjects/study/web/http/http_client_timeout.md`
  - 생성: `/Users/rody/VscodeProjects/study/WORK_20260409_HTTP_CLIENT_TIMEOUT.md`
- 각 파일에서 바꿀 논리 또는 구조:
  - WORK: 근거와 판정 ledger
  - 본문: 질문형 개념 문서
- 관련 문서 동기화 계획:
  - 허브 문서 없음. 별도 링크 수정은 생략
- 예제 추가 / 보강 계획:
  - pseudo code와 ASCII diagram 추가
- 근거 섹션 반영 계획:
  - 문서 마지막에 1차 자료 중심 링크 모음

### 15.1 Success / Failure / Regression Cases

- 성공 케이스 최소 3개:
  - S1: 문서만 읽고 connect timeout과 request timeout을 구분할 수 있다
  - S2: 왜 HTTP/2가 timeout API를 더 복잡하게 만드는지 설명할 수 있다
  - S3: 라이브러리별 timeout 모델 차이를 "계층 차이"로 이해할 수 있다
- 실패 케이스 최소 3개:
  - F1: timeout을 하나의 숫자로만 이해하게 만든다
  - F2: RFC가 API 모양까지 정한 것처럼 오해하게 만든다
  - F3: per-request timeout을 지원하는 라이브러리가 왜 가능한지 설명하지 못한다
- 회귀 위험:
  - 기존 문서와 직접 충돌은 적음
- 회귀 방지 확인 경로:
  - 링크와 파일 경로 점검

### 15.2 Code / Doc Quality Review Points

- 단순성: 개념 축을 먼저 세우고 예시는 뒤에 둔다
- 응집도: timeout 모델 하나의 질문으로 문서를 묶는다
- 확장 여지: 이후 Java/Go 개별 문서에서 링크할 수 있다
- 과한 일반화 여부: "모든 라이브러리"라고 쓰지 않고 경향으로 서술
- 설명 누락 위험:
  - request deadline과 socket timeout 차이를 놓치지 않기

## 16. Detailed Plan Critique + Repair

- 누락된 케이스:
  - HTTP/3/QUIC 언급 필요
- fuzzy success criteria:
  - 문서를 읽고 "왜"를 설명할 수 있어야 함
- scope overreach / under-specification:
  - 너무 많은 라이브러리 나열은 피하고 대표 사례 위주로 정리
- 보강안:
  - 라이브러리를 유형별로 묶는다
- 최종 상세 계획:
  - timeout 종류 표 -> 낮은 계층 -> 역사 -> 라이브러리 유형 비교 -> 실무 가이드

## 17. Frozen Checklist

- [x] 문서 위치를 `study/web` 하위로 고정
- [x] 표준 vs 구현 수렴을 분리해서 설명
- [x] 소켓/커널/OS 맥락 포함
- [x] 역사와 발전 과정 포함
- [x] 다르게 제공하는 라이브러리 사례 포함
- [x] 이해를 돕는 pseudo code, ASCII, 비유 포함
- [x] 문서 작성 완료
- [x] 품질 검수 완료
- [x] 검증 명령 완료
- [x] commit 완료

## 18. Execute Log

- 공식 자료 링크와 관련 study 문서 구조를 확인했다.
- timeout 종류를 먼저 분리한 뒤 계층별 원인과 예외 사례를 설명하는 구조로 확정했다.
- 낮은 계층 설명에 kernel, TCP stack, NIC driver 관점을 추가했다.
- 역사 흐름을 별도 요약 절로 보강했다.

## 19. Verification Plan

- `git -C /Users/rody/VscodeProjects/study diff --check`
- `markdownlint`가 있으면 대상 파일 검사
- 수동 검수:
  - 용어 정의가 앞에 나오는지
  - 원인과 결과가 같은 흐름 안에 닫히는지
  - 추론과 사실이 섞이지 않는지

## 20. Final Audit

- 결과:
  - 표준이 API를 직접 강제하지 않는다는 점과 구현 수렴을 분리해 설명했다.
  - socket / kernel / connection / request 계층을 분리했다.
  - HTTP/1.0 -> keep-alive -> pool -> HTTP/2 -> HTTP/3의 흐름을 연결했다.
  - 대표 라이브러리를 유형별로 나눠 비교했다.
- 검증 결과:
  - `git -C /Users/rody/VscodeProjects/study diff --check -- WORK_20260409_HTTP_CLIENT_TIMEOUT.md web/http/http_client_timeout.md` PASS
  - `npx --yes markdownlint-cli /Users/rody/VscodeProjects/study/WORK_20260409_HTTP_CLIENT_TIMEOUT.md /Users/rody/VscodeProjects/study/web/http/http_client_timeout.md` PASS
- 잔여 리스크:
  - OkHttp 링크는 3.x Javadoc 기준이라 최신 버전의 표현은 조금 다를 수 있다. 다만 이 문서에서 필요한 timeout 분류와 설계 경향 설명에는 충분하다.
