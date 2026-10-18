# PostgreSQL 소스 실행 학습 시험(파일럿)

- 시작: 2026-10-12 20:19 KST · 실행자: Claude Code(`claude-opus-5-5`) · 깊이: full · finish: 단계별 검증 + 범위 한정 commit, push 없음
- 상태: IN_PROGRESS(S5 완료, S6 사용자 실습 대기) · 작업 상태의 정본은 [question-led-learning PLAN](../../ai/authoring/question-led-learning/PLAN.md)의 QLL-07 행
- 산출물 위치: [`open_source_internals/database/postgresql/`](../../open_source_internals/database/postgresql/README.md). 2026-10-12 사용자 지시로 `database/postgresql/internals-lab/`에서 옮겼다. 아래 실행 기록의 옛 경로는 당시 사실이므로 고치지 않았다.

## 요청과 목적

사용자의 목적 원문(2026-10-12): “리눅스, 카프카, spring, postgresql, nodejs 등등 다양한 오픈 소스들을 이렇게 interactive로 보면서 그리고 왜 그렇게 동작하는지를 배우고 싶은 건데요. … 제가 사용은 하지만 그 내부가 어떤지, 왜 그렇게 구현 됐는지, 어떤 선택의 결과물인지 등을 직접 보고 그 코드가 어떻게 동작하는지 앎으로써 많이 배울 수 있다고 생각하는데요. 물론 실제 활용해보는 게 가장 좋고요. … 보통 할 수 있다고 하고 막상 해보면 AI도 버거워하는 경우가 많아서, 지금 할 수 있다고 말하는 게 정말 되는 건지 걱정이긴 합니다.”

그 앞의 화면 원문은 “브라우저로 치면 좌측에는 시각화가 있고 우측에는 소스코드가 나옵니다. … 쿼리를 실행합니다. 그러면 쿼리 요청을 받는 소스 코드 라인으로 이동하고, 시각화는 어떤 일이 이뤄지는지 보여줍니다”이다.

이번 요청 원문: “PostgreSQL 시험을 진행해봐 주세요. 근데 마구 진행하지 말고... study 말고 별도 저장소를 만드는 게 좋을까요? 계획이나 전략이나 뭔가 기록하면서 진행할 필요는 없나요? 바로 진행하는 게 좋나요?”

이 시험이 판단할 것은 하나다. **실제 실행으로 AI의 설명을 검사하는 방식이 큰 오픈소스의 내부·이유·선택을 배우는 데 실제로 작동하는가.** 왼쪽 시각화·오른쪽 소스 화면은 이번 범위가 아니다. 이 시험에서 반복해서 필요한 그림이 확인된 뒤에 만든다(AI 판단, 2026-10-12 대화에서 사용자에게 순서로 제안).

## 진행 방식의 결정

| 질문 | 결정 | 이유 | 다시 열 조건 |
| --- | --- | --- | --- |
| 별도 저장소인가 | study의 `database/postgresql/internals-lab/` | 이번 산출물은 학습 문서와 작은 실험이고, 같은 주제의 문서가 `database/postgresql/`에 있다. 소스 트리와 빌드 결과는 Docker 이미지 안에만 있고 git에 넣지 않는다. 프로젝트 README도 독립 저장소를 “외부 사용자·독립 버전·배포가 실제로 필요해진 단계”로 두었다. | 실행기·소스 뷰어·시나리오 형식 같은 도구 코드가 두 개 이상의 오픈소스에서 쓰이거나, 따로 공개·버전 관리할 필요가 생기면 **도구만** 별도 저장소로 옮기고 학습 문서는 study에 둔다. |
| 기록하며 진행하는가 | 이 WORK에 예측과 판정 기준을 **실행 전에** 적는다 | 시험의 목적이 “정말 되는가”의 판정이다. 결과를 본 뒤 기준을 정하면 무엇이든 통과처럼 해석하기 쉽다. | — |
| 바로 진행하는가 | 아래 단계별 관문을 통과할 때만 다음 단계로 간다 | 실패를 고치기 쉬운 단계에서 드러내고, 사용자가 방향을 바꿀 지점을 남긴다. | 관문의 실패 신호가 나오면 멈추고 보고한다. |

## 시험 질문

- Q1(어떻게): `SELECT` 한 줄은 서버 소스의 어떤 경로를 거치는가.
- Q2(왜): PostgreSQL은 왜 행의 버전 정보(`xmin`/`xmax`)를 행 안에 저장하는가.
- Q3(선택): `synchronous_commit`을 끄면 COMMIT에서 무엇이 달라지는가.

## 사전 등록: AI의 예측(실행·소스 확인 전)

아래는 Claude가 PostgreSQL 18.6 소스를 보기 전에 기억으로 적은 예측이다. S1b에서 실제 소스 트리와 기계적으로 대조해 존재·이름 변경·부재를 기록한다. 이 대조 결과는 “AI가 그럴듯하게 지어낸다”는 걱정에 대한 첫 측정값이다.

| ID | 예측한 함수·구조 | 예측한 파일 |
| --- | --- | --- |
| P1 | `PostgresMain` | `src/backend/tcop/postgres.c` |
| P2 | `exec_simple_query` | `src/backend/tcop/postgres.c` |
| P3 | `pg_parse_query` | `src/backend/tcop/postgres.c` |
| P4 | `raw_parser` | `src/backend/parser/parser.c` |
| P5 | `pg_analyze_and_rewrite_fixedparams` | `src/backend/tcop/postgres.c` |
| P6 | `parse_analyze_fixedparams` | `src/backend/parser/analyze.c` |
| P7 | `pg_plan_queries` | `src/backend/tcop/postgres.c` |
| P8 | `pg_plan_query` | `src/backend/tcop/postgres.c` |
| P9 | `standard_planner` | `src/backend/optimizer/plan/planner.c` |
| P10 | `PortalStart` | `src/backend/tcop/pquery.c` |
| P11 | `PortalRun` | `src/backend/tcop/pquery.c` |
| P12 | `standard_ExecutorStart` | `src/backend/executor/execMain.c` |
| P13 | `standard_ExecutorRun` | `src/backend/executor/execMain.c` |
| P14 | `ExecutePlan` | `src/backend/executor/execMain.c` |
| P15 | `ExecSeqScan` | `src/backend/executor/nodeSeqscan.c` |
| P16 | `heap_getnextslot` | `src/backend/access/heap/heapam.c` |
| P17 | `ReadBufferExtended` | `src/backend/storage/buffer/bufmgr.c` |
| P18 | `printtup` | `src/backend/access/common/printtup.c` |
| P19 | `ReadyForQuery` | `src/backend/tcop/dest.c` |
| P20 | `CommitTransaction` | `src/backend/access/transam/xact.c` |
| P21 | `RecordTransactionCommit` | `src/backend/access/transam/xact.c` |
| P22 | `XLogInsert` | `src/backend/access/transam/xloginsert.c` |
| P23 | `XLogFlush` | `src/backend/access/transam/xlog.c` |
| P24 | `issue_xlog_fsync` | `src/backend/access/transam/xlog.c` |
| P25 | `XLogSetAsyncXactLSN` | `src/backend/access/transam/xlog.c` |
| P26 | `heap_insert` | `src/backend/access/heap/heapam.c` |
| P27 | `ExecInsert` | `src/backend/executor/nodeModifyTable.c` |
| P28 | `HeapTupleHeaderData`의 `t_xmin`·`t_xmax` | `src/include/access/htup_details.h` |
| P29 | `heap_update` | `src/backend/access/heap/heapam.c` |
| P30 | `HeapTupleSatisfiesMVCC` | `src/backend/access/heap/heapam_visibility.c` |

예측한 설계 문서: `src/backend/access/heap/README.HOT`, `src/backend/access/heap/README.tuplock`, `src/backend/access/nbtree/README`, `src/backend/access/transam/README`, `src/backend/storage/buffer/README`.

Q3의 예측 결과: `synchronous_commit=on`이면 COMMIT을 실행한 백엔드가 `XLogFlush`를 호출하고 WAL 파일에 `fdatasync`(Linux 기본 `wal_sync_method`)를 직접 실행한다. `off`이면 그 백엔드는 COMMIT 레코드의 flush를 기다리지 않고 `XLogSetAsyncXactLSN`으로 WAL writer에 맡기므로, COMMIT 시점의 백엔드에서는 `fdatasync`가 관측되지 않는다.

## 판정 기준과 실패 신호

| 기준 | 통과 조건 | 근거의 종류 |
| --- | --- | --- |
| M1 메커니즘 | `-O0 -g3` 빌드 완료. 특권 없는 컨테이너(`SYS_PTRACE`만 추가)에서 gdb가 백엔드에 붙고, `SELECT` 실행 중 중단점이 실제 파일·줄을 출력 | 디버거 출력 |
| M2 예측 정확도 | P1~P30과 README 다섯 개의 존재·이름 변경·부재를 기록. 통과 기준치는 두지 않는 측정값 | 소스 트리 대조 |
| M3 소스·실행 일치 | 학습 문서에 인용하는 모든 함수·줄이 디버거 로그나 빌드한 소스 트리에서 확인됨 | 디버거 출력, 소스 트리 |
| W 이유 | 모든 “왜”에 1차 자료(소스 안의 주석·README, 공식 문서, 원 설계 논문, 개발 메일링 리스트)가 붙거나 “추론” 표시 | 1차 자료 |
| T 선택 | `synchronous_commit` on/off에서 COMMIT 시점의 차이를 디버거나 시스템 콜 추적으로 직접 관측 | 디버거 출력, strace |
| L 학습 | 사용자가 직접 명령을 실행하고, 다음 중단점과 변형(인덱스가 있을 때)의 경로를 먼저 예측한 뒤 확인 | 사용자의 실제 응답 |
| C 비용 | 단계별 소요 시간, 다운로드 크기, 막힌 곳 | 실행 기록 |

실패 신호: 한 번 재시도해도 끝나지 않는 빌드, 승인 범위(300MB)를 넘는 다운로드를 재승인 없이 진행, 빌드한 소스에 없는 이름의 인용, 출처도 “추론” 표시도 없는 이유, 특권 컨테이너가 있어야만 하는 디버깅. 마지막 신호가 나오면 안전 경계를 바꾸지 않고 멈춘다.

L은 사용자가 참여해야 닫힌다. 그 전까지 이 시험의 정직한 상태는 PARTIAL이다(메커니즘 확인, 학습 효과 미확인).

## 환경과 안전 경계

- 소스: PostgreSQL 18.6 공식 tarball `postgresql-18.6.tar.bz2`(22,632,314바이트). 공식 배포처가 HTTPS로 공개한 SHA-256 `555610c24d53e4316da5b7d3fc25c279d96856d5e0e23ee308c328c5fa881d9f`를 빌드 중에 대조한다. PGP 서명 검증은 하지 않는다.
- 빌드 도구·gdb·strace: Debian bookworm 공식 저장소. 설치 전에 패키지 묶음의 실제 다운로드 크기를 계산해 이 문서에 적고, 소스와 합쳐 300MB를 넘으면 멈추고 다시 묻는다.
- 빌드 플래그: `--enable-debug --without-icu`, `CFLAGS="-O0 -g3 -fno-omit-frame-pointer"`. 최적화를 끈 빌드이므로 성능 측정에 쓰지 않는다.
- 실행 컨테이너: `--cap-add=SYS_PTRACE`만 추가하고 `--privileged`·`--pid=host`는 쓰지 않는다. 포트를 공개하지 않고 `--network none`으로 실행하며 서버는 Unix socket만 연다. 호스트 경로는 실습 스크립트만 읽기 전용으로 연결한다. 2026-10-12의 앞선 커널 추적 탐색과 달리 같은 VM의 다른 컨테이너 데이터에 닿지 않는다.

## 단계와 관문

1. S0 계획: 이 문서, PLAN의 QLL-07 행, 실습 폴더의 Dockerfile과 안내를 커밋한다.
2. S1 환경: 호스트 여유 공간 확인(예상 사용량의 두 배 이상, 2026-10-12 빌드 실패 뒤 추가) → 패키지 크기 계산 → 이미지 빌드 → 디버그 정보와 빌드 플래그 확인. 관문: 빌드 완료, `pg_config`에 플래그가 보이고 gdb가 `exec_simple_query`의 소스 줄을 찾는다.
3. S1b 예측 대조: P1~P30과 README를 소스 트리에서 찾는다. 관문: 대조표 기록.
4. S2 Q1 추적: 소스에서 확인한 함수로 중단점을 정하고 `SELECT`의 실제 경로를 기록한다. 관문: M1. **여기서 사용자에게 중간 보고한다.**
5. S3 Q3 추적: `synchronous_commit` on/off를 비교한다.
6. S4 Q2 근거: `pageinspect`로 INSERT·UPDATE 뒤 `xmin`/`xmax`를 관측하고 1차 자료로 이유를 정리한다.
7. S5 학습 문서: 사용자가 직접 실행하고 예측할 과정을 포함해 실습 README를 쓴다.
8. S6 사용자 실습과 S7 판정: L을 관측하고 M1~C를 판정한다.

## 근거 표시 원칙

결과를 적을 때 근거를 세 종류로 구분한다. **[소스]**는 빌드한 18.6 트리에서 확인한 것, **[실행]**은 디버거·strace·SQL 출력으로 관측한 것, **[기억]**은 아직 확인하지 않은 AI의 지식이다. [기억]은 학습 문서에 사실로 쓰지 않는다.

## 실행 기록

(단계마다 시각, 실제로 실행한 명령, 결과와 막힌 곳을 덧붙인다.)

- S0: 2026-10-12 20:19 시작. 공식 배포처 목록에서 18.x의 최신 판이 18.6임을 확인했다. 19.x 목록은 없었다. 20:21 계획을 `0e20f16ba`로 커밋했다.
- S1 패키지 크기: 20:21:39 Debian 패키지 106개, 100,424,600바이트(95.8MB). 소스 tarball과 패키지 목록을 합쳐 약 130MB로 승인 범위 안이었다.
- S1 빌드 실패와 VM 정지: 20:21:48 이미지 빌드를 시작했고, 20:21:53 apt 다운로드 도중 OrbStack VM이 강제로 멈췄다. `~/.orbstack/log/vmgr.log`에는 가상 디스크 쓰기가 macOS 쪽에서 `No space left on device`(StorageFull)로 실패한 기록과 `VM stopped`가 있다. 같은 VM에서 실행 중이던 postgres·vaultwarden 컨테이너도 함께 멈췄다.
- 원인: 호스트 APFS 데이터 볼륨 460Gi 중 425Gi를 쓰고 있었고, 여유 공간은 약 4GB(사용률 100%)였다. 빌드가 승인 범위를 넘은 것이 아니라 호스트에 남은 공간이 거의 없었다.
- 계획의 빈틈: 사전 점검은 다운로드 크기(사용자 승인 범위)만 셌다. 호스트의 여유 공간과 빌드가 실제로 차지할 공간은 세지 않았다. `-O0 -g3` 디버그 빌드는 다운로드보다 훨씬 큰 공간을 쓴다(추정 3~5GB, 실제로 재지 않음).
- 조치: 20:25 `orbctl start`로 OrbStack을 다시 켰다. postgres:18.6과 vaultwarden이 healthy로 돌아온 것을 `docker ps`로 확인했다. Docker build cache는 0B로, 실패한 빌드가 남긴 것은 없다. 아무것도 삭제하지 않았다.
- `scripts/find-def.sh`: 작은 가짜 소스 트리로 동작을 확인했다. 정의 줄만 찾고 extern 선언과 호출은 세지 않으며, 트리가 없으면 종료 코드 2로 끝난다. 실제 18.6 트리에서는 아직 실행하지 않았다.
- 상태: BLOCKED(호스트 디스크 부족). 재개 조건은 호스트 여유 공간 확보다(권고 20GB 이상). 판정 기준(M1~C)은 바꾸지 않는다. S1 관문의 첫 단계로 “호스트 여유 공간이 예상 사용량의 두 배 이상인지 확인”을 추가한다.
- 재개: 20:34 사용자가 공간을 정리한 뒤 다시 확인했다. 호스트 여유 공간은 44Gi(사용률 90%), VM 안 디스크는 41G였다. 확인 명령은 `df -h /System/Volumes/Data`, `orbctl status`, `docker ps`, `docker system df`, `docker run --rm debian:bookworm-slim df -h /`이다.
- S1 빌드: 20:34:39~20:37:29(2분 50초). 5초마다 호스트 여유 공간을 재고 10GiB 아래면 중단하는 감시 스크립트로 실행했다. tarball 체크섬 `OK`. 여유 공간은 46,524,404KiB에서 최저 44,114,080KiB까지 내려가 **약 2.3GiB를 썼다**(사전 추정 3~5GB보다 적음). 이미지 크기 1.49GB.
- S1 관문 PASS [실행]: `postgres --version` 18.6, `pg_config --configure`에 `--enable-debug --without-icu 'CFLAGS=-O0 -g3 -fno-omit-frame-pointer'`, 바이너리에 `.debug_info`·`.debug_line`·`.debug_macro` 섹션, gdb `info line exec_simple_query`가 `postgres.c` 1013행을 반환, pageinspect 설치.
- S1b 예측 대조 [소스]: P1~P30이 모두 예측한 파일에 정의되어 있었다(`evidence/predictions-vs-source.tsv`). README 다섯 개도 모두 있었다. 단 P28의 `t_xmin`·`t_xmax`는 `HeapTupleHeaderData`(153행)에 직접 선언된 것이 아니라, 그 안에 `t_heap`으로 끼워진 `HeapTupleFields`(122~125행)에 선언되어 있다. 이 대조는 이름과 파일의 존재만 확인한다. 실행 경로에 있는지는 S2에서 확인한다.
- S2 감시 목록 설계 [소스]: 함수 인자 이름은 18.6 소스의 서명에서 읽었다. 그 과정에서 18.6의 순차 스캔이 WHERE 조건과 열 재구성 유무에 따라 `ExecSeqScan`·`ExecSeqScanWithQual`·`ExecSeqScanWithProject`·`ExecSeqScanWithQualProject` 중 하나를 고른다는 것을 확인했다(`nodeSeqscan.c` 262~277행). `SELECT * … WHERE id = 2`에는 `ExecSeqScanWithQual`을 예상했다.
- S2 실행 [실행]: 20:39 `start-lab.sh`로 컨테이너를 띄웠다(`SYS_PTRACE`만 추가, `--network none`, 커널 `ptrace_scope=0`). 첫 추적에서 `ReadBufferExtended`가 한 번도 호출되지 않아, 소스에서 `heap_fetch_next_buffer`(heapam.c 647행)가 `read_stream_next_buffer`(679행)로 페이지를 받는 것을 확인하고 감시 목록에 read stream 경로와 P7·P8을 더했다. 두 번째 실행 전 `run-trace.sh`가 root 소유의 `/tmp/ready`를 지우지 못해 멈추는 결함을 발견해, 정리 단계를 root로 바꿨다. 최종 기록은 `evidence/select-where-id-2.trace`다.
- S2 결과 [실행]: 실행 순서는 `exec_simple_query` → `pg_parse_query` → `raw_parser` → `parse_analyze_fixedparams`(`T_SelectStmt`) → `pg_plan_queries` → `pg_plan_query` → `standard_planner`(`CMD_SELECT`) → `PortalStart` → `standard_ExecutorStart` → `PortalRun` → `standard_ExecutorRun`(계획 `T_SeqScan`) → `ExecutePlan` → `ExecSeqScanWithQual` → `heap_getnextslot`(4회: 행 3개와 끝) → `heap_fetch_next_buffer` → `read_stream_next_buffer` → `StartReadBuffers` → `PinBufferForBlock`(테이블 `t`, 블록 0) → `printtup`(1회, id=2 행) → `ReadyForQuery`다. `main`부터 `PinBufferForBlock`까지 호출 스택 전체가 파일·줄과 함께 남았다. `WaitReadBuffers` 0회: 캐시를 데운 뒤라 디스크 읽기를 기다리지 않았다(추론: 페이지가 이미 공유 버퍼에 있었다).
- S2 예측 판정: 실행 경로에 대한 예측 중 **P15 `ExecSeqScan`(0회, 실제는 `ExecSeqScanWithQual`)과 P17 `ReadBufferExtended`(0회, 실제는 read stream → `PinBufferForBlock`)가 틀렸다.** 두 함수 모두 소스에는 있으므로 S1b의 이름 대조만으로는 이 오류가 잡히지 않았다. P20~P27은 Q3·Q2에서 확인할 대상이라 이번 추적에 넣지 않았다.
- 환경 사건: 20:49:13 OrbStack이 VM 관리자를 새 빌드로 교체하며(`vmgr.log`의 `replaced running vmgr … newBuild`) 컨테이너를 재시작했다. 재시작 정책이 없는 `pg-lab`은 정상 종료(smart shutdown) 상태로 남아 `docker start`로 다시 켰다. 실습 명령이 원인이 아니며, 테이블 `t`의 행 3개는 보존됐다.
- 스크립트 정리: 두 추적기가 같은 세션 준비 코드를 쓰도록 `session.sh`로 분리했다. 분리한 `run-trace.sh`로 S2를 다시 실행해, 메모리 주소를 빼면 `evidence/select-where-id-2.trace`와 한 줄도 다르지 않음을 `diff`로 확인했다.
- S3 소스 [소스]: `xact.c` 1498~1502행은 `wrote_xlog && markXidCommitted && synchronous_commit > SYNCHRONOUS_COMMIT_OFF`이거나 강제 동기 조건(`forceSyncCommit`, 지울 non-temp 파일 `nrels > 0`)일 때 `XLogFlush(XactLastRecEnd)`를 호출하고, 아니면 1523행에서 `XLogSetAsyncXactLSN(XactLastRecEnd)`를 호출한다. 그 위 주석(1473~1496행)과 1512~1520행이 이유를 직접 설명한다. 비동기 커밋은 postmaster가 비정상 종료하면 WAL 버퍼가 쓰이지 않아 커밋된 트랜잭션을 잃을 수 있고, WAL writer가 그 LSN까지 flush한다. CLOG 갱신은 WAL flush 뒤로 미룬다. `issue_xlog_fsync`(xlog.c 8746행)는 `wal_sync_method=fdatasync`일 때 8781행에서 `pg_fdatasync`를 부른다. 현재 설정은 `wal_sync_method=fdatasync`, `wal_writer_delay=200ms`다.
- S3 첫 시도의 실패: `SET synchronous_commit = on; INSERT …`를 한 줄로 보내면 한 Query 메시지가 될 거라 가정했다. 그러나 psql은 파일 입력을 문장마다 따로 보내므로 추적이 `SET`에서 끝났다. 설정을 연결 시작 옵션(`PG_OPTIONS` → `PGOPTIONS`)으로 넘기도록 `session.sh`를 고쳤다.
- S3 gdb 결과 [실행]: 두 설정 모두 `ExecInsert` → `heap_insert`(테이블 `t`) → `XLogInsert`(rmgr "Heap") → `CommitTransaction` → `RecordTransactionCommit` → `XLogInsert`(rmgr "Transaction", 커밋 레코드)까지 같다. **on**에서는 이어서 `XLogFlush` → `XLogWrite` → `issue_xlog_fsync`(fd=10) → `pg_fdatasync`(fd=10)를 백엔드가 직접 실행했다. 호출 스택의 `RecordTransactionCommit` 위치가 `xact.c:1502`로, 소스의 분기와 일치한다. **off**에서는 `XLogSetAsyncXactLSN` 한 번 뒤 바로 `ReadyForQuery`였고, flush·write·fdatasync가 백엔드에서 0회였다(`evidence/commit-sync-on.trace`, `commit-sync-off.trace`).
- S3 strace 결과 [실행]: **on**에서는 백엔드가 `pg_wal/000000010000000000000001`에 8192바이트를 `pwrite64`로 쓰고, 0.17ms 뒤 `fdatasync`를 호출했다. **off**에서는 백엔드의 WAL 쓰기가 0회였고, 약 0.2초 뒤 WAL writer(PID 17)가 같은 파일에 `pwrite64`와 `fdatasync`를 실행했다(`evidence/commit-sync-*.strace`). 간격이 `wal_writer_delay`와 맞는다는 것은 시각이 일치한다는 데서 나온 추론이다.
- S3 판정: Q3의 사전 예측과 일치한다. **T PASS**(설정 하나의 차이를 함수와 시스템 콜 두 수준에서 관측). 앞서 같은 맥의 커널 추적에서 본 블록 계층의 동기 쓰기와 장치 캐시 비우기가 이 `fdatasync`의 아래쪽이다. 다만 이번 실습 컨테이너에서는 커널 추적을 하지 않았다.
- S4 소스 [소스]: 줄 포인터 상태는 `LP_UNUSED`=0·`LP_NORMAL`=1·`LP_REDIRECT`=2·`LP_DEAD`=3이다(`itemid.h` 38~41행). `heap_update`는 새 버전의 xmin에 갱신 트랜잭션 번호를 넣고(`heapam.c` 3861행), 옛 버전의 xmax에도 같은 번호를 넣는다(4260행). `htup_details.h` 124~125행의 주석은 `t_xmin`을 inserting xact ID로, `t_xmax`를 deleting or locking xact ID로 설명한다.
- S4 실행 [실행]: `scripts/mvcc-page.sh`의 기록은 `evidence/mvcc-page.out`에 있다. INSERT(xid 774)는 버전 1개를 만든다(xmin=774, xmax=0). UPDATE(xid 775)는 행을 제자리에서 고치지 않는다. 옛 버전 lp1에 xmax=775와 `t_ctid`=(0,2)를 남기고, 같은 페이지에 새 버전 lp2(xmin=775)를 추가했다. 플래그는 `HEAP_HOT_UPDATED`/`HEAP_ONLY_TUPLE`이다. REPEATABLE READ 읽기 트랜잭션(스냅숏 `776:776:`)을 연 채 다른 세션이 xid 776으로 갱신하자, 새 세션은 `c`를 읽었다. 먼저 시작한 읽기는 갱신 커밋 뒤에도 `b`를 읽었고, 페이지에는 세 버전이 모두 남아 있었다. VACUUM 뒤 lp1은 REDIRECT(`lp_off`=3)가 됐고, lp2는 UNUSED, lp3만 살아 있는 버전으로 남았다.
- S4 가시성 판정 [소스]: 읽기의 스냅숏은 xmin=776, xmax=776이다. lp2(`b`)를 지운 xid 776은 `XidInMVCCSnapshot`에서 “xid >= xmax면 진행 중”으로 처리된다(`snapmgr.c` 1883~1885행). 그래서 `HeapTupleSatisfiesMVCC`가 “treat as still in progress”로 lp2를 보이게 판정한다(`heapam_visibility.c` 1146~1148행). lp3(`c`)는 xmin 776이 진행 중이므로 보이지 않는다(1078~1081행). lp1(`a`)은 xid 775가 이미 커밋한 삭제이므로 보이지 않는다(1153행).
- S4 이유 [공식 문서]: PostgreSQL 18 문서 13.1절은 MVCC의 주된 이점을 읽기용 잠금과 쓰기용 잠금이 충돌하지 않는 것, 곧 읽기와 쓰기가 서로를 막지 않는 것으로 설명한다. 24.1.2절은 UPDATE·DELETE가 옛 버전을 바로 지우지 않는 이유를 다른 트랜잭션에 아직 보일 수 있기 때문이라고 설명하고, 그래서 VACUUM이 필요하다고 말한다(https://www.postgresql.org/docs/18/mvcc-intro.html, https://www.postgresql.org/docs/18/routine-vacuuming.html, 2026-10-12 확인). 설계 문서 `README.HOT`는 이 방식의 비용을 다룬다. 갱신마다 새 버전이 생기면 인덱스 항목과 정리 비용이 늘어나고, HOT는 인덱스 열을 바꾸지 않는 갱신에서 그 비용을 줄인다.
- S4 대안 [공식 문서]: MySQL 8.4 문서 17.3절(InnoDB Multi-Versioning)은 반대 선택을 설명한다. 클러스터드 인덱스 행을 제자리에서 고치고, 옛 버전은 undo 로그(rollback segment)에 둔다. 행에는 `DB_TRX_ID`와 `DB_ROLL_PTR`가 있고, purge가 더는 필요 없는 버전을 지운다(https://dev.mysql.com/doc/refman/8.4/en/innodb-multi-versioning.html, WebFetch 요약으로 확인).
- S4 미확인: Stonebraker의 1987년 논문(The Design of the POSTGRES Storage System, VLDB)은 스캔 이미지 PDF라 본문을 읽지 못했다. PostgreSQL의 버전 보관 방식이 그 no-overwrite 설계에서 이어졌다는 설명은 **[추론]**으로 표시한다. 이 읽기 시도에서 WebFetch가 PDF 사본(1.5MB)을 세션 임시 폴더에 저장했다. 저장소 밖이고 더 쓰지 않았다. undo 기반 PostgreSQL 저장소 제안(zheap)은 확인하지 않았다.
- S4 판정: W는 학습 문서의 모든 “왜”에 위 출처나 [추론] 표시를 붙여야 통과한다. S5에서 판정한다.
- S5 실습 문제의 사전 예측(21:03 첫 실행 전): 학습 문서의 문제 다섯 개는 Claude가 새 컨테이너에서 먼저 실행하고, 그 출력만 정답으로 쓴다. 아래는 실행 전에 적은 예측이다.
    - E1 `SELECT * FROM t;`: `ExecSeqScan`(조건 없음, 열 재구성 없음). `heap_getnextslot` 4회, `printtup` 3회.
    - E2 `SELECT name FROM t WHERE id = 2;`: `ExecSeqScanWithQualProject`.
    - E3 인덱스 `t_id_idx(id)`와 `enable_seqscan=off`에서 `SELECT * FROM t WHERE id = 2;`: 계획은 Index Scan이고 `ExecIndexScan` → `IndexNext` → `index_getnext_slot` → `index_getnext_tid` → `btgettuple` → `_bt_first` → `_bt_search` → `index_fetch_heap` → `heapam_index_fetch_tuple` → `heap_hot_search_buffer` → `printtup` 1회를 지난다. 인덱스 페이지와 힙 페이지를 각각 읽는다. 메타 페이지는 캐시를 데울 때 relcache(`rd_amcache`)에 남으므로 추적 중에는 다시 읽지 않는다.
    - E4 `READ COMMITTED` 읽기: 두 번째 읽기는 새 값 `c`를 읽는다(문장마다 새 스냅숏).
    - E5 `synchronous_commit=on`에서 `SELECT * FROM t;`의 자동 커밋: `CommitTransaction`·`RecordTransactionCommit`은 지나지만 `XLogInsert`·`XLogFlush`·`pg_fdatasync`는 0회다(트랜잭션 번호를 받지 않아 커밋 레코드가 없다).
- S5 정답 확보 [실행]: 21:03~21:07 새 컨테이너에서 문서 순서대로 실행했다. 다섯 문제 모두 예측과 같았다(5/5). E1 `ExecSeqScan` 4회·`heap_getnextslot` 4회·`printtup` 3회. E2 `ExecSeqScanWithQualProject`. E3은 Index Scan이고, 인덱스 블록 1(`bt_metap` 루트 1·level 0)과 힙 블록 0만 읽었으며 메타 페이지는 다시 읽지 않았다. 힙 페이지는 `ReleaseAndReadBuffer`로 받았다. E4 두 번째 읽기 `c`. E5 `XLogInsert`·`XLogFlush`·`pg_fdatasync` 0회. E3 감시 목록의 함수 17개는 실행 전에 `find-def.sh`로 모두 확인했다. 같은 실행에서 Q1·Q2·Q3의 기존 기록도 주소·xid·LSN을 빼면 같았다.
- S5 새 관측 [실행]: `on`의 strace를 다시 실행하자 WAL writer 줄이 함께 나왔다. `pg_waldump`로 보니 백엔드가 내린 커밋 레코드(xid 767) 바로 뒤에 `Standby`의 `RUNNING_XACTS` 레코드가 있었다. 그 LSN 0/0179F4C8은 WAL writer가 다시 쓴 페이지(오프셋 7987200부터 8KB) 안에 있다. 이 레코드는 bgwriter가 15초마다 남긴다(`bgwriter.c` 70행, 254~258행 주석). WAL writer가 내린 것이 이 레코드라는 것은 순서와 위치로 본 추론이다(`evidence/commit-sync-on-rerun.strace`, `evidence/waldump-after-commit.txt`). 학습 문서에서는 오해 바로잡기로 쓴다. `off`에서 백엔드가 끝난 뒤 WAL writer가 쓰기까지 걸린 시간은 0.197초와 0.084초로, 고정값이 아니었다.
- S5 스크립트 변경: `index.watch`를 추가했다. `mvcc-page.sh`에는 격리 수준을 바꾸는 `ISOLATION`과 `lp_off` 열을 더했다. S4 기록의 `lp_off`=3은 당시 증거 파일에 없던 값이었다. 이번에 열을 더해 증거로 남겼고(VACUUM 뒤 줄 1 `lp_off`=3), `evidence/mvcc-page.out`을 새로 만들었다(xid 757~759). `find-def.sh`는 `pageinspect`의 SQL 함수를 찾도록 `contrib`까지 찾는다. 학습 문서의 C 이름 인용을 검사하는 `check-readme.sh`도 추가했다. 이 검사기는 처음 실행에서 `contrib`에 있는 이름 3개를 잡아 `find-def.sh`의 범위 문제를 드러냈다.
- S5 학습 문서: README를 질문 1·2·3 순서로 다시 썼다. 각 질문은 짧은 답, 실제 기록, 단계별 설명, 오해 바로잡기, 예측 후 실행 문제로 이어진다. 기준 문서는 `computer_architecture/ostep/xv6-riscv/kernel/entry.S`와 `git/git_rebase.md`다. entry.S에서는 첫 상태와 전제를 먼저 고정하는 방식을 가져왔다(새 컨테이너, 행 3개, 캐시 데우기, `-O0`). 줄마다 설명하는 밀도는 따르지 않았다. git_rebase.md에서는 구체적인 전후 상태를 먼저 보이는 방식을 가져왔고, 화면 캡처 중심 구성은 따르지 않았다. 배경 개념(VACUUM 운영, freeze, bloat)은 `deep-dive/postgresql/18-postgresql-deep-dive.md`와 `mvcc.md`로 연결했다.
- S5 근거 확인 [소스]: 새로 인용한 줄을 18.6 트리에서 직접 확인했다.
    - 실행 경로: `postgres.c` 590~602·1012~1014, `nodeSeqscan.c` 257~277, `read_stream.c` 머리 주석, `executor/README` 1~11, `analyze.c` 머리 주석, `nbtpage.c` 361~363, `heapam.c` 1698~1704
    - 행 버전과 hint bit: `heapam_visibility.c` 6~10·80~112, `htup_details.h` 204·210
    - 커밋: `xact.c` 1350~1354·1398~1405·1473~1530, `bgwriter.c` 70·254~258
    - 설정과 상수: `pg_config.h` 786(`XLOG_BLCKSZ` 8192), `aio.h` 42, `method_worker.c` 머리 주석, `wal_segment_size` 16MB, `block_size` 8192
    - 공식 문서 [공식 문서]: PostgreSQL 18의 51.2(연결마다 프로세스), 13.2.1·13.2.2(격리 수준별 스냅숏 시점), 28.4(비동기 커밋의 위험 구간, 최대 `wal_writer_delay`의 3배), 9.27.8 표 9.85(스냅숏 형식)를 WebFetch로 확인했다.
    - 학습 문서에 추론으로 표시한 것은 여덟 가지다. `WaitReadBuffers` 0회의 원인, 메타 페이지 캐시를 남긴 주체, 인덱스 스캔이 read stream을 쓰지 않는 이유, WAL writer 간격과 `wal_writer_delay`의 관계, WAL writer가 내린 레코드, 1987년 논문, hint bit에 대한 비동기 커밋의 영향(소스만 확인), I/O worker의 처음 읽기(소스만 확인)다.
- S5 재생 [실행]: 21:16~21:18 README의 bash 블록 21개를 문서 순서대로 새 컨테이너에서 실행했다(세션 임시 폴더의 `replay.py`). 모두 종료 코드 0이었다. 출력은 기존 기록과 같았고(주소·xid·LSN·시각 제외), 재생한 명령은 README의 블록과 문자 그대로 같다. 실습 정답 파일 `evidence/exercise-*`는 이 재생의 출력이다. `check-readme.sh`는 C 이름 66개를 모두 정의로 찾았다. 상대 링크 29개와 목차 앵커도 모두 맞았다.
- S5 계획에 없던 다운로드: 재생의 첫 블록 `docker build`가 캐시를 쓰지 않고 처음부터 다시 빌드했다. Debian 패키지 약 109MB와 tarball 22MB를 다시 받았고(체크섬 `OK`, 약 100초), 끝까지 마친 두 번의 빌드로 받은 양은 합쳐 약 260MB다. 20:21에 VM이 멈추며 중단된 첫 시도가 받은 양은 기록이 없어 모르므로, 누적량이 승인 범위 300MB 안이라고 단정할 수 없다. 이 재다운로드는 승인을 다시 묻지 않고 일어났다. 빌드 캐시가 왜 비어 있었는지는 확인하지 않았다(20:49 OrbStack VM 관리자 교체와의 관계는 미확인). 옛 이미지 `4c1fde28c970`(20:37 빌드, 1.49GB)이 이름 없이 남았고, 지우지 않았다. 호스트 여유 공간은 42Gi에서 40Gi가 됐다. 같은 재생을 다시 할 때는 이미지가 있으면 빌드 블록을 건너뛰어야 한다. README에 이 동작을 적었다.
- S5 영향 범위: `deep-dive/postgresql/18-postgresql-deep-dive.md`의 두 "출처와 검증 경계" 절에 이 실습으로 오는 링크를 한 문장씩 달았다. `validate_deep_dive.py` 출력은 변경 전과 같은 PASS(56단위)다. 사용자 원재료 노트인 `database/mvcc.md`는 고치지 않았다. deep-dive 계획이 이 파일을 보존 후 재구성할 원재료로 지정했고, deep-dive 문서에서 링크가 이어지면 충분하다고 판단했다(AI 판단).
- S5 미실행: `trace.py`의 ruff `--select D` 게이트는 이 맥에 ruff가 없고 uv 캐시에도 없어 실행하지 않았다. 내려받으려면 승인이 필요하다. 셸 스크립트에는 §9.2 표에 확인한 게이트가 없다.
- S5 독립 검토 1차 [검토]: 21:22~21:40 별도 검토 에이전트가 README를 고치지 않고 읽으며 소스 줄 번호 약 35개, evidence의 숫자, 공식 문서 절 번호, C 이름 67개, 상대 링크를 대조했다. 잘못된 줄 번호나 숫자는 없었다. 대신 다섯 갈래의 문제를 지적했다.
    1. M3 위반 두 곳. `bt_page_items()`가 키를 보여 준다고 썼지만 명령이 키 열을 고르지 않았고, 0.084초는 evidence 파일 없이 이 WORK의 서술에만 있었다.
    2. 사실 오류. 먼저 시작한 읽기가 갱신 커밋 뒤에 두 번 모두 `b`를 읽었다고 썼지만, 첫 읽기는 UPDATE 전에 일어난다.
    3. 측정 기준점. 0.197초는 백엔드 프로세스가 끝난 시각부터 잰 간격인데, 커밋부터의 간격처럼 200ms와 비교했다.
    4. W. `off`의 이점(28.4절)이 빠지고 대가만 있었다. REDIRECT의 이유와 `RUNNING_XACTS`를 남긴 프로세스를 출처나 추론 표시 없이 사실처럼 썼다.
    5. 학습 순서. WAL·rmgr·portal·공유 버퍼 같은 용어가 정의 없이 나왔다. `pwrite64`와 `fdatasync`의 차이, LSN을 파일 오프셋으로 바꾸는 과정, 호출 스택에 `PostmasterMain()`이 있는 이유가 빠졌다. 실습 4·5는 정답에 필요한 전제(격리 수준별 스냅숏 시점, 트랜잭션 번호의 늦은 할당)를 정답보다 먼저 가르치지 않았다. `mvcc-page.sh`는 `ISOLATION` 오타를 조용히 통과시켰다.
- S5 검토 반영(21:40~21:49): 지적을 모두 README에 반영했다. 판단이 들어간 처리만 적는다.
    - 키 열은 명령에 `data` 열을 더하고 출력을 `evidence/exercise-3-index-pages.out`으로 남겼다.
    - 간격은 재생 기록을 `evidence/commit-sync-off-rerun.strace`와 `evidence/waldump-after-replay.txt`로 남기고, 커밋 레코드의 시각부터 WAL writer의 쓰기까지 0.085초로 다시 쟀다. 151행의 0.084초는 같은 실행을 백엔드 종료부터 잰 값이다. 0.197초에는 백엔드 종료부터 잰 값이라고 밝혔고, 비교 대상은 28.4절의 최대 위험 구간(`wal_writer_delay`의 세 배)으로 바꿨다. 시각은 컨테이너의 UTC로 통일했다.
    - REDIRECT는 인덱스를 이유로 든 문장을 지우고, `pruneheap.c`가 사슬의 상태만 보고 정한다는 소스 사실로 바꿨다.
    - `RUNNING_XACTS`를 남긴 프로세스는 추론으로 낮췄다. 체크포인트도 같은 레코드를 남긴다는 사실(`xlog.c` 7245행)과, 그 사이 WAL 기록이 있었을 때만 남긴다는 조건(`bgwriter.c` 264~267행)을 더했다.
    - 설정값은 학습자가 직접 확인하는 `pg_settings` 명령을 더하고 출력을 `evidence/settings.out`으로 남겼다.
    - 실습 4·5의 전제는 본문에서 먼저 설명하도록 옮겼다. 실습 5에는 트랜잭션 번호가 없음을 `pg_current_xact_id_if_assigned()`로 직접 확인하는 명령을 더했다(`evidence/exercise-5-xid.out`).
    - 복습 질문은 이 문서의 명령으로 확인할 질문(1·3)과 문서·소스로 답할 질문(2·4·5)을 나누어 밝혔다. "이 실습이 확인하지 않은 것"에는 비동기 커밋 뒤 비정상 종료를 더했다.
    - `mvcc-page.sh`에는 읽기 세션의 `ON_ERROR_STOP`, 실행 전 `ISOLATION` 검사(틀리면 종료 코드 2), 두 기다림의 30초 한도를 더했다.
- S5 내 실수: 기다림 한도를 넣으면서 조건을 `sh -c`로 실행했는데, 그러면 `pgrep -f`가 그 sh 자신의 명령줄과 맞는다. 그대로 두었다면 두 번째 기다림은 읽기 세션이 끝난 뒤에도 30초 한도에 걸렸을 것이다. 고친 직후 스크립트를 다시 읽다가 발견해 `[p]sql` 패턴으로 바꿨다. 고치기 전 판은 실행하지 않았으므로 "한도에 걸렸을 것"은 소스로 본 추론이다.
- S5 재생 3회차 [실행]: 21:48~21:49 고친 README의 bash 블록 23개 가운데 빌드(1번)와 마지막 정리(23번)를 뺀 21개를 새 컨테이너에서 문서 순서대로 실행했다. 모두 종료 코드 0이었고, 30초 한도에 걸린 기다림은 없었다. 재생한 명령은 지금 README의 블록과 문자 그대로 같다. 실습 정답과 기록 9개(`settings.out`, `exercise-1`~`5`, `mvcc-page.out`)는 주소·xid·LSN·시각을 빼면 이 재생의 출력과 같았다. `exercise-3-index-pages.out`은 블록 9와 11의 출력을 이어 붙인 파일이라 그 차이만 있다. 반례 시험으로 `ISOLATION='REPEATABLE REED'`를 주자 안내문을 내고 종료 코드 2로 끝났다. 30초 한도에 걸리는 실패 경로는 실행해 보지 않았다.
- S5 독립 검토 2차 [검토]: 21:53~21:56 같은 검토 에이전트가 고친 곳만 다시 봤다. 1차 지적은 대부분 FIXED였다. 사소한 지적은 rmgr·tid를 정의보다 먼저 쓴 곳, HOT의 빈 공간 조건이 `README.HOT`에 직접 적혀 있지 않다는 점, 0.197초의 "조금", 재실행 기록의 두 번째 WAL writer 쓰기, 정의 없는 "체크포인트", 그리고 1차부터 남은 `read_stream_next_buffer` 3회 대 `heap_fetch_next_buffer` 2회였다. 새로 생긴 문제 가운데 고칠 것은 둘이었다.
    1. "0.17ms가 `on`의 비용"이라는 문장은 틀렸다. strace의 시각은 호출이 시작된 시각이라, 0.17ms는 `pwrite64`가 시작되고 `fdatasync`가 시작되기까지의 시간이다.
    2. `pruneheap.c`의 "first non-DEAD tuple"을 "첫 번째 살아 있는 버전"으로 옮겨, 읽기 트랜잭션이 열려 있을 때의 넘김 목적지를 잘못 예측하게 만들었다.
- S5 2차 반영과 추가 확인 [실행·소스·공식 문서]: 21:54~22:01.
    - strace 시각: 컨테이너에서 `strace -tt -T -e trace=clock_nanosleep,nanosleep,exit_group sleep 0.3`을 실행했다. `clock_nanosleep` 줄의 시각 12:56:48.982228에 걸린 시간 0.300491초를 더하면 다음 줄의 시각 12:56:49.284609와 거의 같다. 시각이 시작 시각임을 확인했다. README는 0.17ms를 거의 `pwrite64`의 시간으로 고쳤다. `fdatasync`의 시간은 재지 않았고, 백엔드 종료까지 5.5ms 이하라는 상한만 적었다. 이 때문에 T의 "on의 비용"은 순서와 존재로만 관측됐고 크기는 관측되지 않았다.
    - 넘김 목적지: 복습 질문 3의 변형, 곧 VACUUM을 읽기 트랜잭션이 끝나기 전으로 옮긴 `mvcc-page.sh` 임시 사본을 실행했다(21:54). VACUUM 뒤 줄 1은 REDIRECT(`lp_off`=2), 줄 2(`b`, xmin 765·xmax 766)는 NORMAL로 남아 위치만 8128에서 8160으로 옮겨졌다. 읽기의 두 번째 결과는 VACUUM 뒤에도 `b`였다. 검토의 정답 예측과 같다. README는 "DEAD가 아닌 첫 버전"으로 고치고 `HEAPTUPLE_RECENTLY_DEAD`(`heapam.h` 126행)로 경계를 설명했다. 이 출력은 복습 질문의 정답이므로 evidence에 넣지 않고 세션 임시 폴더에만 두었다.
    - 3회 대 2회: 감시 목록에 `heap_endscan`, `read_stream_end`, `read_stream_reset`을 더한 사본을 컨테이너의 /tmp에만 넣고 같은 SQL을 추적했다(21:57). 세 번째 `read_stream_next_buffer`는 `heap_endscan` → `read_stream_end` → `read_stream_reset` 다음에 불렸다. `read_stream.c` 1087~1089행은 쓰지 않은 버퍼의 고정을 풀려고 이 함수를 반복해 부른다. 기록은 `evidence/scan-end-order.trace`에 남기고 README에 한 문단을 더했다.
    - HOT 조건은 `heap_update()`의 `pagefree` 비교(`heapam.c` 4049~4055행)와 같은 버퍼일 때만 HOT을 검토하는 분기(4116~4125행)를 근거로 고쳤다. 체크포인트는 공식 문서 28.5절의 정의로 설명했다.
    - README의 "evidence 파일마다 처음 인용한 자리에 링크했다"는 문장이 원래 strace 기록 두 개에는 맞지 않았다. 링크를 더해 모든 evidence 파일이 README에서 링크되게 했다.
    - 결과: `check-readme.sh` 이름 78개 모두 정의 있음, 상대 링크 40개와 앵커 10개 모두 맞음, 키보드 밖 문장부호 0개.
- S5 독립 검토 3차 [검토]: 22:02~22:03 같은 검토 에이전트가 2차 반영의 11곳만 다시 보고 PASS로 판정했다. 새 M3 오류는 없었다. 용어는 쓰기 전에 정의되었고, 인용한 소스 줄도 맞았다. 검토자는 README의 strace 확인 명령을 직접 실행해 결과를 재현했다. 선택 사항 두 가지도 반영했다. 시각 계산에 "거의"를 붙였고, `heapam.c` 인용은 `RelationGetBufferForTuple` 호출이 끝나는 4055행까지 늘렸다(검토자는 4054행을 제안했지만 소스에서 호출은 4055행에서 끝난다).
- S5 판정:
    - M3 PASS. C 이름 78개는 `check-readme.sh`가 기계적으로 확인했다. 줄 번호 인용과 evidence의 숫자는 검토자가 세 차례 손으로 대조했다. 줄 번호를 기계적으로 확인하는 검사는 없다.
    - W PASS. 2·3차 검토에서 출처도 추론 표시도 없는 "왜"는 남지 않았다. 이 판정은 검토자의 대조에 기대며, 자동 검사는 없다.
    - T PASS(S3)를 유지하되 범위를 좁힌다. `on`이 커밋 전에 `fdatasync`를 부르고 `off`는 부르지 않는다는 순서와 존재는 관측했다. `fdatasync` 한 번이 걸린 시간, 곧 `on`의 비용의 크기는 재지 않았다.
    - L은 열려 있다. S6에서 사용자가 실습 1~5를 직접 예측하고 실행해야 판정할 수 있으므로, 시험 전체는 PARTIAL이다.
- S5 비용 [C]: S5는 21:03~22:03, 약 1시간이 걸렸다. 그중 독립 검토 세 차례가 약 18분, 3분, 1분이었다. 시험 전체는 20:19 시작부터 약 1시간 45분이고, 디스크 부족으로 멈춘 13분(20:21~20:34)을 포함한다. 토큰 사용량은 재지 않았다. 다운로드는 위 "계획에 없던 다운로드" 항목에 적었고, 이후 새 다운로드는 없다. 고칠 것을 찾은 것은 대부분 실행이 아니라 독립 검토였다. 다만 검토의 예측(넘김 목적지, 세 번째 호출의 출처, strace 시각의 뜻)은 이번에 실행으로 확인했다.
- M1 PASS. M2: 이름·파일 30/30, 실행 경로 예측 중 2개 오류, P28 부분 일치. 시간: S0 약 2분, S1(재개 뒤) 약 4분, S1b·S2 약 3분, 스크립트 수정과 재실행 포함.
