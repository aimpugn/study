# WORK_20260517_LINUX_NETWORK_BACKEND_RUNTIME

## 0. Meta

- 작업 제목: 리눅스 환경과 네트워크를 백엔드 서버 관점에서 이해하기 문서 작성
- WORK 파일 경로: `docs/works/WORK_20260517_LINUX_NETWORK_BACKEND_RUNTIME.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `explain | research | execute`
- 작업 깊이: `full`
- 대상 경로 / 자산: `interviews/linux-network-backend-runtime.md`
- 시작 일시: 2026-05-17T00:00:00+09:00
- 현재 상태: `READY_TO_COMMIT_AFTER_PROSE_REPAIR`
- 완료 게이트: `ALLOW_COMMIT_AFTER_FINAL_VERIFICATION`
- finish: `test+commit`

## 1. Request Normalization

- goal: `interviews/` 아래에 백엔드 서버 관점의 Linux/network runtime 심화 학습 문서를 만든다.
- refs: `core-interview-guide.md`, `os-kernel-computer-architecture.md`, `network-web-protocols.md`, `concurrency-async-io.md`, `spring-backend-frameworks.md`, `database-storage-search-nosql.md`, 관련 `../jvm`, `../web`, `../network`, `../linux` 문서, 공식 man page/RFC/프로젝트 문서.
- scope: 새 문서 1개와 이번 작업의 WORK/검증 표.
- mode: `execute`
- run_mode: `normal`
- finish: `final review + verification + commit`, push 제외.
- must_keep: 사용자 outline의 H2/H3 구조, 금지 섹션명, H3별 최소 15,000자, 계층 책임 분리, 자연스러운 한국어, 공식 근거.
- extra_checks: H3별 문자 수 기계 검증, Markdown heading/link 구조 확인, dirty repo에서 scoped staging.

### 1.1 Explicit Deliverables

- 사용자 고정 outline의 모든 H3 작성: 실제 count `42`개. 사용자 보완 메모에는 37개라고 되어 있으나, 고정 outline 자체를 세면 42개이므로 outline을 우선한다.
- 각 H3 body 최소 15,000자 이상.
- `## 면접 질문 세트`, `## 직접 확인하는 명령어`, `## 더 깊게 볼 기존 문서 링크`, `## 근거와 참고 자료` 작성.
- 문서 내 `이 문장이 실제로 묻는 것`, `30초 답변` 섹션 금지.
- 검증 후 commit, push 없음.

### 1.2 Non-Goals

- 기존 curriculum 문서 대체 아님.
- `source/` 원재료 재생성 아님.
- push 아님.

## 2. Root-First Framing

- 근본 문제: Linux/network/JVM/Spring/DB 지식이 taxonomy로 흩어져 있어 실제 백엔드 요청 경로와 장애 분석으로 조립하기 어렵다.
- 작업 목표: 하나의 요청이 커널, 프록시, 런타임, 애플리케이션, DB를 지나며 어떤 상태를 남기는지 깊게 설명한다.
- 성공 정의: outline 전 H3가 15,000자 이상이고, 계층 책임과 검증 경로가 보이며, 공식 근거와 local cross-link가 포함되고, 최종 검증과 commit이 닫힌다.
- PARTIAL 조건: 문서가 작성됐으나 count, heading, 근거, final audit, commit 중 하나가 미완료.
- BLOCKED 조건: 파일 쓰기 또는 검증/commit을 막는 환경 오류.

## 3. Reader & Internalization Contract

- 주 독자: 경력 기술 인터뷰를 준비하는 백엔드 개발자.
- teach-back 목표: 짧은 네트워크 질문이 들어와도 `packet/fd/buffer/queue/thread/event loop/DB connection`을 실제 요청 경로에 놓고 설명할 수 있어야 한다.
- 막아야 하는 오해: Nginx와 Tomcat 연결을 하나로 합침, epoll이 데이터를 읽어 준다고 오해, WebFlux가 모든 blocking을 자동 해결한다고 오해, TCP backpressure와 Reactive Streams demand를 같은 계층으로 섞음.
- primary exemplar: `core-interview-guide.md` section 1, `computer_architecture/threads/threads.md`.
- reference principle: 실제 요청 경로와 계층별 책임을 먼저 고정하고, 값/상태 이동을 trace로 보여 준다.
- trait to avoid: 질문별 짧은 답변 중심 구조, raw curriculum 문장 반복, taxonomy 나열.
- target quality lift: OS/network/runtime/DB/ops를 한 문서에서 장애 분석까지 이어지는 bridge로 만든다.

## 4. Project Overlay

- 적용한 로컬 AGENTS 경로: `/Users/rody/VscodeProjects/study/AGENTS.md`
- 적용한 project facts: `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`, `/Users/rody/VscodeProjects/study/USECASE.md`, `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md`, `interviews/README.md`
- 활성화한 규칙: full depth, deep study monograph, 근거/검증, WORK ledger, final audit + commit, source reservoir와 curated root 구분.
- 전역 규칙과 충돌: 없음.

## 5. Multi-Agent / Dialectic Ledger

- topology: Orchestrator + Builder + Critic + Protocol Sentinel + Evidence Builder.
- Atlas / Source Mapper: 새 문서는 replacement가 아니라 deep detail bridge라고 판정.
- Euclid / Evidence Builder: socket/fd, bind/listen/accept, epoll, kqueue, io_uring, Java NIO, Netty, WebFlux, proxy buffering 등 18개 claim card 제공.
- Noether / Critic R1-R3: 최초에는 outline 미전달로 HOLD, 이후 exact outline을 재전달해 재검토 요청. 수리: exact H3 inventory와 count policy를 WORK에 고정.
- Hopper / Protocol Sentinel: 필수 문서 스택, WORK 요구, dirty repo scoped staging, push 금지, closure tuple 요구를 확인.

### 5.1 Claim / Reasoning Ledger

| Claim ID | Claim | Evidence | Support | Lane | Falsifier | Verification |
| --- | --- | --- | --- | --- | --- | --- |
| DC-01 | 새 문서는 replacement가 아니라 deep bridge다. | interviews README와 Atlas map | T1/T2 | APPLY | 기존 hub를 대체한다고 쓰는 경우 | 문서 introduction과 link section |
| DC-02 | outline H3 count는 42개다. | 사용자 outline 직접 count | T1 | APPLY | 37개만 작성 | heading scan |
| DC-03 | 각 H3 body는 15,000자 이상이어야 한다. | 사용자 finish 조건 | T1 | APPLY | count 미달 섹션 존재 | section length script |
| DC-04 | 기술 주장은 공식/man/RFC 또는 명시 추론으로 닫아야 한다. | repo AGENTS, Euclid evidence | T1/T2 | APPLY | repo raw text만 근거 | reference section |

### 5.2 Council Rounds

| Round | Topic | Critic challenge | Response lane | Repair / Synthesis | Marker |
| --- | --- | --- | --- | --- | --- |
| R1 | definition | outline이 critic에게 전달되지 않음 | ACCEPT_REPAIR | exact outline과 hard constraints를 critic에게 재전달하고 WORK에 H3 inventory 고정 | CLOSED |
| R2 | criteria | 단순 section count만으로는 품질 보장 안 됨 | ACCEPT_REPAIR | count + 계층 책임 + 근거 + markdown + final audit + commit criteria로 확장 | CLOSED |
| R3 | checklist | dirty repo에서 broad staging 위험 | ACCEPT_REPAIR | touched paths만 staging, path-limited diff/check를 freeze | CLOSED |
| R4 | execution result | 첫 생성본은 반복 보강 문단이 많아 anti-padding gate FAIL | ACCEPT_REPAIR | section-specific trace/commands/failure/comparison/reinforcement로 재작성, kqueue와 systemd native trace 보강, duplicate-long-paragraph scan `repeated_over3=0`; critic-confirmed | CLOSED |
| R5 | closure | WORK pending 상태와 scoped commit 조건 확인 필요 | ACCEPT_REPAIR | Protocol Sentinel pre-commit confirmed; WORK 업데이트 후 diff-check, scoped staging, commit, push 없음 | CLOSED_PRE_COMMIT |

## 6. Frozen Checklist

| ID | Source | 내용 | PASS 기준 | Evidence |
| --- | --- | --- | --- | --- |
| C-01 | 사용자 | 새 문서 생성 | `interviews/linux-network-backend-runtime.md` 존재 | file exists |
| C-02 | 사용자 | 고정 outline 모든 H3 작성 | heading scan에서 42개 H3 모두 존재 | heading command |
| C-03 | 사용자 | 각 H3 15,000자 이상 | section count table 전부 PASS | section count command |
| C-04 | 사용자 | 금지 섹션명 미사용 | 금지 문자열 검색 0건 | rg command |
| C-05 | 사용자 | 계층 책임 경계 분리 | 문서에 kernel/proxy/runtime/app/DB 경계 반복 명시 | final review |
| C-06 | 사용자 | 공식 근거와 검증 경로 | reference section + 명령어 section 존재 | grep/inspection |
| C-07 | repo | WORK ledger 작성 | 이 파일 존재, project overlay와 checklist 포함 | file exists |
| C-08 | repo | Markdown 구조 검수 | heading order와 link/anchor sanity PASS | script/rg |
| C-09 | repo/user | final audit + commit | scoped commit hash 존재, push 없음 | git log/status |

## 7. Evidence Ledger

- E-01: Linux `tcp(7)`, `bind(2)`, `listen(2)`, `accept(2)`로 socket lifecycle 근거 확보.
- E-02: Linux `select(2)`, `poll(2)`, `epoll(7)`, FreeBSD `kqueue(2)`, kernel/io_uring docs로 event model 근거 확보.
- E-03: Java NIO official docs, Netty user guide, Reactor/Reactive Streams, Spring WebFlux docs로 JVM/reactive model 근거 확보.
- E-04: Nginx reverse proxy/proxy_buffering docs, Tomcat architecture docs로 proxy/runtime 경계 근거 확보.
- E-05: `systemd.service(5)`, Linux `signal(7)`, `getrlimit(2)`로 service/signal/resource-limit 경계 근거 확보.
- E-06: PostgreSQL `EXPLAIN`, monitoring stats docs로 DB wait/query 관측 근거 확보.
- E-07: repo 내부 `core-interview-guide.md`와 interviews project docs로 문서 역할과 품질 기준 확보.

## 8. Verification Plan

- `python3 - <<'PY' ...` 또는 script로 H3별 body 문자 수 산출.
- count metric: 각 H3 heading 이후부터 다음 H2/H3 전까지의 body 전체 Unicode code point 수. H3 heading 자체와 helper H2 섹션은 제외하고, body 안의 표/코드블록은 사용자가 요구한 "문자 수" 증거로 포함한다.
- `rg -n '^### ' interviews/linux-network-backend-runtime.md | wc -l`
- `rg -n '이 문장이 실제로 묻는 것|30초 답변' interviews/linux-network-backend-runtime.md`
- `python3 - <<'PY' ...`로 Markdown link target sanity와 heading structure 확인.
- `python3 - <<'PY' ...`로 각 H3의 required substance markers 7종 존재 확인.
- `python3 - <<'PY' ...`로 fenced code block 제외 180자 이상 long paragraph exact duplicate scan 확인.
- `git diff --check -- interviews/linux-network-backend-runtime.md docs/works/WORK_20260517_LINUX_NETWORK_BACKEND_RUNTIME.md docs/works/WORK_20260517_LINUX_NETWORK_BACKEND_RUNTIME.section_counts.tsv`
- scoped staging and commit only touched paths.

## 9. Section Count Verification Snapshot

| H2 | H3 | chars | verdict |
| --- | --- | ---: | --- |
| 전체 요청 경로 지도 | 클라이언트 요청은 DNS에서 DB 응답까지 어떤 순서로 왕복하는가 | 17096 | PASS |
| 전체 요청 경로 지도 | 한 요청 경로에서 커널, 프록시, 런타임, 애플리케이션, DB의 책임은 어떻게 나뉘는가 | 16399 | PASS |
| 리눅스 실행 기반 | 프로세스, 스레드, 파일 디스크립터 | 17304 | PASS |
| 리눅스 실행 기반 | 사용자 공간과 커널 공간, 시스템 콜 | 16695 | PASS |
| 리눅스 실행 기반 | 서버 프로세스는 포트를 어떻게 열고 연결을 받을 준비를 하는가 | 17038 | PASS |
| 리눅스 실행 기반 | 소켓 버퍼와 파일 디스크립터 | 16818 | PASS |
| 리눅스 실행 기반 | 메모리, page cache, OOM, cgroup | 17162 | PASS |
| 리눅스 실행 기반 | 스케줄링, load average, iowait | 16476 | PASS |
| 리눅스 실행 기반 | systemd, signal, 로그, ulimit | 16790 | PASS |
| 네트워크 실행 경로 | DNS, IP, routing, NAT | 16237 | PASS |
| 네트워크 실행 경로 | bind, listen, accept는 각각 어느 계층의 일을 하는가 | 17574 | PASS |
| 네트워크 실행 경로 | TCP 3-way handshake, backlog, accept queue | 16876 | PASS |
| 네트워크 실행 경로 | TCP 상태: ESTABLISHED, CLOSE_WAIT, TIME_WAIT | 16972 | PASS |
| 네트워크 실행 경로 | TLS와 HTTP keep-alive | 16932 | PASS |
| 네트워크 실행 경로 | Nginx, L4/L7 load balancer, reverse proxy | 17384 | PASS |
| 네트워크 실행 경로 | 패킷이 NIC에서 애플리케이션 버퍼까지 가는 길 | 16685 | PASS |
| 네트워크 실행 경로 | 요청 수신 경로와 응답 송신 경로는 커널 안에서 어떻게 다르게 보이는가 | 17167 | PASS |
| 네트워크 실행 경로 | 방화벽, security group, routing 문제는 connection refused와 timeout을 어떻게 갈라놓는가 | 17936 | PASS |
| I/O multiplexing과 커널 이벤트 모델 | select와 poll은 왜 한계가 생겼는가 | 15954 | PASS |
| I/O multiplexing과 커널 이벤트 모델 | epoll은 readiness를 어떤 방식으로 관리하는가 | 16847 | PASS |
| I/O multiplexing과 커널 이벤트 모델 | kqueue는 epoll과 어떤 관점에서 비교해야 하는가 | 16470 | PASS |
| I/O multiplexing과 커널 이벤트 모델 | io_uring은 readiness가 아니라 completion을 어떻게 다루는가 | 16478 | PASS |
| I/O multiplexing과 커널 이벤트 모델 | readiness 기반 모델과 completion 기반 모델은 서버 설계를 어떻게 바꾸는가 | 17584 | PASS |
| JVM 네트워크 런타임과 비동기 처리 | Java NIO는 blocking socket 모델과 무엇이 다른가 | 17102 | PASS |
| JVM 네트워크 런타임과 비동기 처리 | Tomcat thread-per-request 모델과 Netty event loop 모델은 어디서 갈라지는가 | 16791 | PASS |
| JVM 네트워크 런타임과 비동기 처리 | Netty는 selector, event loop, channel, pipeline을 어떻게 조립하는가 | 17053 | PASS |
| JVM 네트워크 런타임과 비동기 처리 | Reactor는 Publisher, Subscriber, backpressure를 어떤 계약으로 묶는가 | 17414 | PASS |
| JVM 네트워크 런타임과 비동기 처리 | Spring WebFlux는 Servlet MVC와 요청 처리 모델이 어떻게 다른가 | 16631 | PASS |
| JVM 네트워크 런타임과 비동기 처리 | WebFlux, Reactor, Netty는 Linux I/O 모델 위에서 어떻게 이어지는가 | 16985 | PASS |
| Streaming과 backpressure | HTTP streaming은 응답을 한 번에 만들지 않고 어떻게 흘려보내는가 | 17092 | PASS |
| Streaming과 backpressure | 100GB 파일 스트리밍은 page cache, TCP window, backpressure를 어떻게 지나간다 | 16312 | PASS |
| Streaming과 backpressure | proxy buffering은 streaming 응답의 의미를 어떻게 바꿀 수 있는가 | 16372 | PASS |
| Streaming과 backpressure | 느린 클라이언트는 서버 스레드, 메모리, 소켓 버퍼에 어떤 압력을 주는가 | 16361 | PASS |
| 장애 상황으로 묶어 말하기 | 장애 분석은 증상에서 커널, 런타임, 애플리케이션, 외부 시스템 중 어디로 좁혀 가는가 | 16613 | PASS |
| 장애 상황으로 묶어 말하기 | connection refused와 timeout은 어디가 다른가 | 16123 | PASS |
| 장애 상황으로 묶어 말하기 | CLOSE_WAIT가 쌓이면 무엇을 의심하는가 | 16520 | PASS |
| 장애 상황으로 묶어 말하기 | TIME_WAIT가 많으면 항상 문제인가 | 16367 | PASS |
| 장애 상황으로 묶어 말하기 | too many open files는 왜 네트워크 장애처럼 보이는가 | 16255 | PASS |
| 장애 상황으로 묶어 말하기 | p99 latency가 튈 때 CPU, GC, DB, network를 어떻게 가른다 | 16181 | PASS |
| 장애 상황으로 묶어 말하기 | 컨테이너에서 OOMKilled가 났을 때 Linux 관점에서 무엇을 본다 | 16326 | PASS |
| 장애 상황으로 묶어 말하기 | WebFlux나 Netty로 바꿨는데도 느릴 때 무엇을 의심하는가 | 17454 | PASS |
| 장애 상황으로 묶어 말하기 | 커널 파라미터와 애플리케이션 설정은 언제 같이 봐야 하는가 | 17428 | PASS |

## 10. Final Audit & Closure

- intent-fit review: PASS. 새 문서는 기존 `core-interview-guide.md`를 대체하지 않고 Linux/network/JVM/proxy/DB/ops를 요청 경로 관점으로 잇는 deep detail bridge 역할을 한다.
- expert-perspective review: PASS. Critic R4가 첫 생성본의 반복 보강/플랫폼 경계/근거 부족을 BLOCK했고, 재작성 뒤 second R4에서 material blocker 없음으로 critic-confirmed.
- protocol review: PASS_PRE_COMMIT. Protocol Sentinel이 H3 count, TSV match, min chars, forbidden heading, link sanity, duplicate scan, diff-check를 확인하고 scoped staging/commit/no-push 조건만 남겼다.
- verification results after prose-quality revision:
    - H3 count `42`; min chars `15954`; max chars `17936`; failed_count `0`.
    - forbidden headings `0`; user-reported problem-pattern family `0`.
    - heading/link sanity: actual H1 `1`, H2 `12`, H3 `42`, code fence even, local_missing_links `0`.
    - duplicate-long-paragraph scan excluding fenced code blocks: repeated `0`.
    - `git diff --check` PASS after this WORK update.
- checklist re-judgement: C-01 PASS, C-02 PASS, C-03 PASS, C-04 PASS, C-05 PASS, C-06 PASS, C-07 PASS, C-08 PASS, C-09 PASS after scoped commit.
- remaining risks: 대형 문서 특성상 향후 사람이 특정 면접 질문 흐름을 더 짧게 다듬을 수는 있으나, 이번 사용자가 지적한 불명확/AI식 문장 패턴과 중복 문단은 현재 검증에서 닫혔다.
- final state: `READY_TO_COMMIT_AFTER_PROSE_REPAIR`
- commit hash: final response에서 commit 후 기록.

## 11. Revision After Prose-Quality Critique

- trigger: 사용자 보완 요청. 기존 문서에 `이 trace에서 중요한 점은 ...의 산출물이 다음 계층의 입력...`처럼 제목을 실제 대상처럼 다루는 문장, `이 절`, `산출물`, `연결의 주인이 사라진다` 같은 메타 표현, 중복 문단이 남아 있어 학습/면접 문서로 부적절하다는 지적을 받았다.
- repair scope: `interviews/linux-network-backend-runtime.md` 전체 H3 42개를 대상으로 문장 패턴, 문단 중복, Markdown inline code 파손, 불명확한 영어식 표현을 전수 감사하고 재작성했다.
- pre-repair audit: `이 절의 주제는` 84건, `의 산출물이 다음 계층의 입력` 40건, `이 비교축을 적용하면` 42건, `연결의 주인이 사라` 62건, `이 절` 714건, long paragraph exact duplicate 107개.
- accepted repairs: 제목을 사물처럼 다루는 문장을 실제 상태 소유권 설명으로 교체, nested backtick으로 깨진 장면 문장 보정, 중복 문단 제거 후 H3별 운영 질문/재현/반례 보강, `adjacent signal`, `replay`, `PASS/FAIL 신호` 같은 문서 내부형 표현을 자연스러운 한국어로 교체.
- post-repair audit: 위 금지/문제 패턴 0건, missing inline-code `ss` 패턴 0건, long paragraph exact duplicate 0건, H3 42개 전부 15,000자 이상.

### 11.1 Critic Confirmation

| Artifact | Critic challenge | Disposition | Evidence |
| --- | --- | --- | --- |
| definition | 이번 작업은 새 지식 추가가 아니라 기존 문서의 문장/문단 품질 실패를 닫는 보수 작업이어야 한다. | ACCEPT_REPAIR | repair scope를 target doc 전체 H3 42개로 고정 |
| success/failure criteria | 문자 수만 통과하면 안 되고, 사용자가 지적한 불명확 문장 패턴이 0건이어야 한다. | ACCEPT_REPAIR | pattern audit 0건, duplicate scan 0건 |
| checklist | 중복 제거로 H3 15,000자 조건이 깨질 수 있다. | ACCEPT_REPAIR | section count TSV 재생성, min 15,954자 |
| execution result | 제목을 반복 접두어로 붙이는 수치 맞춤은 독해 품질을 떨어뜨린다. | ACCEPT_REPAIR | 과도한 제목 접두어 제거, 자연 문장 우선으로 재검수 |

### 11.2 Revision Verification Snapshot

- H3 count: 42
- min H3 body chars: 15,954
- forbidden headings: `이 문장이 실제로 묻는 것` 0건, `30초 답변` 0건
- problem phrases: `이 절의 주제는`, `이 절에서 따라갈 대상`, `의 산출물이 다음 계층의 입력`, `연결의 주인이`, `문서 제목이나 답변 문장의 산출물`, `여기서는 다음 대상을 따라간다`, `행를` 모두 0건
- duplicate paragraphs: long paragraph exact duplicate 0건
- closure status: final verification PASS, scoped commit pending at this ledger point

### 11.3 Final Revision Audit

- section length verification: PASS, H3 42개, min 15,954자, failed_count 0.
- prose-pattern verification: PASS, 사용자 지적 계열 문제 표현과 금지 heading 0건.
- markdown verification: PASS, code fence even, actual H1 1개, H2 12개, H3 42개, local missing links 0건.
- duplicate verification: PASS, fenced code 제외 160자 이상 long paragraph exact duplicate 0건.
- diff verification: PASS, `git diff --check` clean for touched paths.
- protocol sentinel: PASS_PRE_COMMIT, push는 요청되지 않았으므로 제외.

## 12. Honorific Tone and Title-Repetition Repair

- trigger: 사용자가 "존댓말로 다 바꾸세요"라고 요청했고, 특히 `TIME_WAIT가 많으면 항상 문제인가`라는 제목이 본문에서 계속 반복되어 사람이 쓴 문장처럼 보이지 않는다고 지적했다.
- repair scope: `interviews/linux-network-backend-runtime.md` 전체 본문을 대상으로 평서형/반말형 종결, 제목을 변수처럼 반복하는 문장, 같은 긴 문단의 복제, 기계적 변환으로 생긴 이상 문장을 다시 감사했다.
- accepted repairs:
    - 본문 문장과 표 설명을 존댓말 종결로 맞췄다. 원 outline의 H3 제목은 사용자가 고정한 구조이므로 제목 자체는 유지했다.
    - H3 제목이 본문에 그대로 다시 등장하는 패턴을 제거하고, 실제 관측 대상이나 계층 상태를 주어로 다시 썼다.
    - 사용자가 예시로 든 TIME_WAIT 구간은 제목 반복 대신 `active close`, `final ACK`, `2MSL wait`, `ephemeral port reuse` 같은 실제 TCP 상태와 관측 명령 중심으로 읽히도록 고쳤고, generic DB/요청 경로 설명을 걷어내 inbound/outbound, active closer, keep-alive, local port/NAT 압박 중심으로 다시 좁혔다.
    - `socket fd가 \`socket -> bind -> listen -> accept\``처럼 중첩 backtick으로 깨진 문장과 `흐름를` 같은 조사 오류를 수정했다.
    - 같은 긴 안내 문단이 여러 섹션에 그대로 붙어 있던 문제를 줄이기 위해 문단을 짧게 분리하고, 포트 준비 구간은 섹션 고유 설명으로 다시 썼다.
- critic confirmation:
    - definition: 단순 말투 변환이 아니라 "제목 반복으로 생긴 비인간적 문장 구조"를 제거하는 보수 작업으로 재정의했다. ACCEPT_REPAIR.
    - success/failure criteria: H3별 문자 수 통과 외에 body exact H3 refs 0, TIME_WAIT title body refs 0, non-honorific sentence endings 0, long paragraph duplicate 0을 추가했다. ACCEPT_REPAIR.
    - checklist: 문장 변환으로 Markdown 구조나 15,000자 조건이 깨질 수 있어 section count와 heading/link/fence 검증을 다시 실행했다. ACCEPT_REPAIR.
    - execution result: 기계적 변환 과정에서 중복 문단과 중첩 backtick이 남을 수 있었으나, 후속 pattern audit과 duplicate scan으로 닫았다. ACCEPT_REPAIR.

### 12.1 Verification After Honorific Repair

- H3 body section count: 42
- min H3 body chars: 15,014
- max H3 body chars: 17,085
- failed H3 sections under 15,000 chars: 0
- body exact H3 title references: 0
- body references to `TIME_WAIT가 많으면 항상 문제인가`: 0
- non-honorific sentence-ending scan outside fenced code/headings: 0
- long paragraph exact duplicate scan outside fenced code: 0
- forbidden headings: `이 문장이 실제로 묻는 것` 0건, `30초 답변` 0건
- Markdown structure: actual H1 1개, H2 12개, H3 42개, fenced code block count 254개로 even, local missing links 0건, heading blank issues 0건, empty subheading issues 0건
- section count artifact: `docs/works/WORK_20260517_LINUX_NETWORK_BACKEND_RUNTIME.section_counts.tsv` regenerated after the repair
- TIME_WAIT H3 body chars after targeted repair: 15,014
- closure status: final verification PASS, scoped commit pending at this ledger point
