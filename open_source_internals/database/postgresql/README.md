# PostgreSQL 내부를 실행하며 따라가기

## 개요

이 폴더는 PostgreSQL 18.6을 디버그 정보와 함께 빌드하고, 쿼리를 실제로 실행하면서 서버가 어느 함수와 어느 줄을 지나는지 기록하는 실습입니다. 다루는 질문은 세 가지입니다.

1. `SELECT` 한 줄은 서버 소스의 어떤 경로를 지나는가.
2. PostgreSQL은 왜 행의 버전 정보(`xmin`, `xmax`)를 행 안에 두는가.
3. `synchronous_commit`을 끄면 COMMIT에서 무엇이 달라지는가.

설명의 사실은 세 곳에서 가져옵니다. 첫째는 디버거(gdb), 시스템 콜 추적기(strace), SQL이 이 컨테이너에서 실제로 기록한 내용이고, 둘째는 빌드에 쓴 18.6 소스 트리, 셋째는 PostgreSQL과 Linux의 공식 문서입니다. 실행 경로와 관측값은 첫째와 둘째에서만 가져옵니다. "왜 그렇게 만들었는가"에 답하는 문장에는 소스 안의 주석과 설계 문서, 또는 공식 문서를 함께 적습니다. 출처로 확인하지 못한 설명은 추론이라고 밝힙니다.

다 읽고 실습을 마치면 다음을 자기 말로 설명할 수 있어야 합니다.

- SQL 문자열 하나가 백엔드 프로세스 안에서 거치는 단계와, 각 단계를 맡은 소스 파일
- UPDATE가 행을 고쳐 쓰지 않고 새 버전을 더하는 이유와, 스냅숏이 그중 하나를 고르는 규칙
- `synchronous_commit` 설정 하나가 바꾸는 것: 누가 WAL(데이터 파일보다 먼저 변경 내용을 적는 로그, 질문 3에서 설명)을 디스크에 내리는지, 언제 내리는지, 그래서 무엇을 얻고 무엇을 잃을 수 있는지

## 목차

- [왜 실행해서 확인하는가](#왜-실행해서-확인하는가)
- [준비: 이미지와 실습 컨테이너](#준비-이미지와-실습-컨테이너)
- [추적 기록을 읽는 법](#추적-기록을-읽는-법)
- [질문 1. SELECT 한 줄이 지나는 경로](#질문-1-select-한-줄이-지나는-경로)
- [질문 2. 행 버전과 스냅숏](#질문-2-행-버전과-스냅숏)
- [질문 3. synchronous_commit이 바꾸는 것](#질문-3-synchronous_commit이-바꾸는-것)
- [덮고 다시 확인하기](#덮고-다시-확인하기)
- [이 실습이 확인하지 않은 것](#이-실습이-확인하지-않은-것)
- [근거와 점검](#근거와-점검)

## 왜 실행해서 확인하는가

이 실습을 만들기 전에 Claude가 기억만으로 함수 이름 30개와 그 파일을 예측했고, 빌드한 소스와 대조했습니다([`evidence/predictions-vs-source.tsv`](evidence/predictions-vs-source.tsv)). 30개 모두 예측한 파일에 실제로 있었습니다. 그런데 실제 `SELECT`를 추적하자 그중 두 함수, `ExecSeqScan()`과 `ReadBufferExtended()`는 한 번도 불리지 않았습니다. 18.6은 같은 일을 다른 함수로 하고 있었습니다(질문 1에서 설명합니다).

이 결과가 이 실습의 방법을 정합니다. "이 함수가 소스에 있다"와 "이 쿼리가 이 함수를 지난다"는 서로 다른 주장입니다. 앞의 것은 소스 검색으로 확인되지만, 뒤의 것은 실행해야 확인됩니다. 그래서 이 문서는 경로에 관한 문장을 모두 실행 기록에 묶고, 실습 문제도 예측한 다음 실행으로 맞춰 보는 방식으로 냅니다.

## 준비: 이미지와 실습 컨테이너

명령은 모두 저장소 루트에서 실행합니다. 이미지를 만들 때 PostgreSQL 공식 tarball과 Debian 패키지를 내려받으므로 네트워크가 필요합니다.

```bash
docker build -t pg-internals:18.6 open_source_internals/database/postgresql
```

이미지가 이미 있으면 이 단계는 건너뜁니다. 빌드 캐시가 남아 있지 않으면 같은 명령이 처음부터 다시 빌드하고 다시 내려받습니다. 2026-10-12에 이 문서의 명령을 다시 실행했을 때 실제로 그랬고, 약 131MB를 다시 받았습니다. 다시 빌드하면 옛 이미지가 이름 없이 남으므로 `docker images --filter dangling=true`로 확인해 필요 없으면 지웁니다.

성공하면 `pg-internals:18.6` 이미지가 생기고, 이미지 안의 `/src/postgresql-18.6`에 빌드한 소스가 남습니다. 빌드 플래그 `-O0`은 최적화를 꺼서 변수와 실행 순서가 소스에 가깝게 남게 하고, `-g3`은 줄 번호와 변수에 더해 매크로 정의까지 디버그 정보에 넣습니다. 대신 이 바이너리로 성능을 재면 안 됩니다. tarball의 SHA-256이 공개값과 다르면 빌드가 `sha256sum: WARNING`과 함께 실패합니다. 이 맥에서는 빌드에 약 3분, 디스크 약 2.3GiB가 들었습니다(시험 WORK의 S1 기록). 2026-10-12에는 호스트 디스크가 가득 차서 첫 빌드 도중 Docker VM이 멈춘 적이 있으므로, 빌드 전에 `df -h /System/Volumes/Data`로 남은 공간을 확인하세요.

실습은 항상 새 컨테이너에서 시작합니다. 아래 기록은 행 3개짜리 테이블에서 얻은 것이고, 질문 3의 INSERT를 실행하면 행 수가 바뀌기 때문입니다. 이전 컨테이너가 없으면 첫 명령은 아무것도 하지 않습니다.

```bash
docker rm -f pg-lab
```

```bash
sh open_source_internals/database/postgresql/scripts/start-lab.sh
```

`pg-lab 준비 완료`가 나오면 컨테이너 `pg-lab` 안에서 서버가 돌고, 테이블 `t(id int, name text)`에 행 `(1,'a')`, `(2,'b')`, `(3,'c')`가 들어 있습니다. 같은 이름의 컨테이너가 이미 있거나 서버가 30초 안에 준비되지 않으면 0이 아닌 종료 코드로 끝납니다.

이 문서가 기대는 서버 설정값을 먼저 확인합니다. 값이 다르면 아래 기록과 결과가 달라질 수 있습니다.

```bash
docker exec pg-lab psql -X -c "SELECT name, setting, unit FROM pg_settings WHERE name IN ('block_size', 'synchronous_commit', 'wal_level', 'wal_segment_size', 'wal_sync_method', 'wal_writer_delay') ORDER BY name;"
```

2026-10-12의 결과([`evidence/settings.out`](evidence/settings.out))는 `block_size` 8192, `synchronous_commit` `on`, `wal_level` `replica`, `wal_segment_size` 16777216바이트(16MB), `wal_sync_method` `fdatasync`, `wal_writer_delay` 200ms였습니다. 각 값의 뜻은 그 값을 쓰는 자리에서 설명합니다.

이 컨테이너는 디버거에 필요한 `SYS_PTRACE` 권한 하나만 더 받습니다. 네트워크를 끊고(`--network none`) 포트를 열지 않으며, 서버는 컨테이너 안의 Unix socket으로만 연결을 받습니다. 호스트에서는 `scripts` 폴더만 읽기 전용으로 `/lab`에 연결합니다. 호스트나 다른 컨테이너와 프로세스 목록(`--pid=host`)이나 볼륨을 공유하지 않으므로, 같은 Docker VM에서 도는 다른 컨테이너의 프로세스와 데이터는 이 실습의 디버거 대상이 되지 않습니다.

## 추적 기록을 읽는 법

추적 스크립트는 세 가지입니다. 셋 다 컨테이너 안에서 `psql` 연결 하나를 열고, 그 연결을 맡은 서버 프로세스를 관찰합니다.

- [`run-trace.sh`](scripts/run-trace.sh)는 gdb를 그 프로세스에 붙이고, 감시 목록 파일에 적힌 함수마다 중단점을 겁니다. 중단점에 닿으면 기록을 한 줄 남기고 곧바로 계속 실행합니다.
- [`run-strace.sh`](scripts/run-strace.sh)는 같은 프로세스와 WAL writer 프로세스가 부르는 시스템 콜을 strace로 기록합니다.
- [`mvcc-page.sh`](scripts/mvcc-page.sh)는 행을 INSERT, UPDATE, VACUUM하면서 테이블 페이지 안의 내용을 `pageinspect` 확장으로 보여 줍니다.

공식 문서 51.2절에 따르면 PostgreSQL은 클라이언트 연결 하나마다 서버 프로세스 하나를 둡니다. 연결 요청을 받는 감독 프로세스(postmaster)가 요청이 올 때마다 새 프로세스를 띄웁니다. 이 문서에서는 연결 하나를 맡은 프로세스를 소스의 용어대로 **백엔드**라고 부릅니다. 질문 3의 strace 기록에서도 두 연결의 백엔드가 서로 다른 프로세스 번호(PID 260과 344)를 받았습니다.

질문 1에서 볼 백엔드의 호출 스택 맨 아래에는 postmaster의 함수(`PostmasterMain()`, `ServerLoop()`, `BackendStartup()`)가 남아 있습니다. postmaster는 `postmaster_child_launch()` 안에서 `fork_process()`(`launch_backend.c` 246행)로 백엔드를 만듭니다. `fork(2)` 매뉴얼에 따르면 fork는 호출한 프로세스를 복제해 새 프로세스를 만들고, 그 순간 두 프로세스의 메모리 내용은 같습니다. 스택도 메모리의 일부이므로, 자식인 백엔드의 스택에는 부모가 fork를 부르기까지의 함수 프레임이 그대로 남습니다.

`run-trace.sh`는 같은 SQL을 먼저 한 번 실행한 뒤에 gdb를 붙입니다. 첫 실행에는 테이블 정의를 카탈로그에서 읽어 캐시에 넣는 일처럼 연결마다 처음 한 번만 하는 일이 섞일 수 있습니다. 기록을 SQL 자체가 하는 일로 좁히려고 이렇게 설계했고, 첫 실행에 실제로 무엇이 섞이는지는 재지 않았습니다. 그래서 `psql` 출력에는 결과가 두 번 나옵니다.

기록의 한 줄은 이렇게 생겼습니다.

```text
STOP 1 exec_simple_query src/backend/tcop/postgres.c:1014 query_string=0xaaaafe32e8e8 "SELECT * FROM t WHERE id = 2;"
```

앞에서부터 멈춘 순번, 함수 이름, 멈춘 소스 위치, 감시 목록에 적은 식과 그 값입니다. `postgres.c`에서 `exec_simple_query()`의 이름은 1012행, 여는 중괄호는 1013행에 있고, 1014행이 본문의 첫 문장입니다. 함수 이름에 건 중단점은 이처럼 함수 본문의 첫 문장에서 멈추므로, 기록의 줄 번호는 정의 줄보다 조금 뒤입니다. `0xaaaa...` 같은 메모리 주소와 프로세스 번호는 실행할 때마다 바뀝니다. 함수 이름, 파일, 줄 번호와 순서는 같은 이미지에서 바뀌지 않습니다.

`STOP` 줄은 감시 목록에 적은 함수만 보여 줍니다. 목록에 없는 함수는 불려도 기록에 나오지 않으므로, `STOP` 줄의 순서를 전체 호출 경로로 읽으면 안 됩니다. 그 사이에 무엇이 있었는지는 `BT` 줄이 보여 줍니다. `BT` 줄은 지정한 함수에 처음 멈췄을 때의 호출 스택 전체이고, `BT 0`이 가장 안쪽 함수입니다. 마지막 `END` 줄은 함수마다 멈춘 횟수입니다. 한 번도 멈추지 않은 함수는 `END` 줄에 나오지 않습니다(`trace.py` 130행). 이 문서에서 어떤 함수가 "0회"였다는 말은 감시 목록에 있었지만 `END` 줄에 없었다는 뜻입니다.

## 질문 1. SELECT 한 줄이 지나는 경로

짧은 답부터 적습니다. `SELECT * FROM t WHERE id = 2;`는 연결을 맡은 백엔드 프로세스 하나 안에서 다음 순서로 처리됩니다.

1. 클라이언트가 보낸 SQL 문자열을 받습니다(`tcop`).
2. 문법만 보고 구문 트리를 만듭니다(`parser`).
3. 카탈로그를 보고 테이블과 열의 의미를 확정합니다(`parser`의 분석 단계). 바로 뒤에 뷰 정의 같은 규칙으로 쿼리를 고쳐 쓰는 재작성 단계(`rewrite`)가 있지만, 이번 감시 목록에는 넣지 않았습니다.
4. 실행 계획을 고릅니다(`optimizer`).
5. 계획 트리를 실행합니다(`executor`).
6. 실행기가 테이블 접근 계층을 통해 행을 하나씩 받습니다(`access/heap`).
7. 테이블 접근 계층이 행이 든 페이지를 공유 버퍼(모든 백엔드가 함께 쓰는 메모리 안의 페이지 캐시)에서 찾습니다(`storage/buffer`).
8. 조건을 통과한 행을 클라이언트에 보내고, 다음 쿼리를 받을 준비가 됐다고 알립니다.

아래 명령으로 이 경로를 직접 기록합니다. 세 번째 인자 `PinBufferForBlock`은 그 함수에 처음 멈췄을 때 호출 스택을 남기라는 뜻입니다.

```bash
sh open_source_internals/database/postgresql/scripts/run-trace.sh select.watch "SELECT * FROM t WHERE id = 2;" PinBufferForBlock
```

정상이면 `STOP` 27줄, `BT` 29줄, `END` 한 줄이 나오고 마지막에 `psql` 결과 `2|b`가 두 번 나옵니다. 60초 안에 세션이 준비되지 않으면 `세션이 준비되지 않았습니다`와 함께 종료 코드 1로 끝납니다. `NOBREAK` 줄이 보이면 감시 목록의 함수 이름이 이 빌드에 없다는 뜻입니다. 2026-10-12의 기록은 [`evidence/select-where-id-2.trace`](evidence/select-where-id-2.trace)에 있습니다.

이 기록은 브라우저에서 단계별로 넘겨 볼 수도 있습니다. [`replays/select-where-id-2.html`](replays/select-where-id-2.html)을 내려받아 브라우저로 열면, 왼쪽에는 27단계가 소스 디렉터리 층 사이를 오가는 경로가, 오른쪽에는 고른 단계에서 멈춘 함수의 소스와 값과 호출 스택이 나옵니다. 서버와 Docker 없이 열리지만, GitHub 화면에서는 HTML 원문이 보입니다. SQL을 바꿔 가며 같은 화면으로 보려면 [작업대](../../workbench/README.md)의 실시간 실행을 씁니다. 아래 단계별 설명을 읽으며 재생본의 같은 단계를 함께 열어 두면, 설명이 가리키는 줄을 바로 확인할 수 있습니다.

### 단계별로 따라가기

기록의 `STOP` 줄을 단계별로 묶으면 다음과 같습니다. 줄 번호는 멈춘 위치입니다.

| 단계 | 멈춘 함수와 위치 | 기록에서 보이는 것 |
| --- | --- | --- |
| 받기 | `exec_simple_query` `postgres.c:1014` | SQL 문자열 그대로 |
| 문법 해석 | `pg_parse_query` `postgres.c:610`, `raw_parser` `parser.c:49` | 같은 문자열 |
| 의미 분석 | `parse_analyze_fixedparams` `analyze.c:109` | 구문 트리의 종류 `T_SelectStmt` |
| 계획 | `pg_plan_queries` `postgres.c:974`, `pg_plan_query` `postgres.c:889`, `standard_planner` `planner.c:335` | 명령 종류 `CMD_SELECT` |
| 실행 준비 | `PortalStart` `pquery.c:450`, `standard_ExecutorStart` `execMain.c:168` | |
| 실행 | `PortalRun` `pquery.c:703`, `standard_ExecutorRun` `execMain.c:319`, `ExecutePlan` `execMain.c:1667`, `ExecSeqScanWithQual` `nodeSeqscan.c:132` | 계획 트리 맨 위 노드 `T_SeqScan` |
| 테이블 접근 | `heap_getnextslot` `heapam.c:1390`, `heap_fetch_next_buffer` `heapam.c:652` | 테이블 이름 `t` |
| 버퍼 | `read_stream_next_buffer` `read_stream.c:813`, `StartReadBuffers` `bufmgr.c:1504`, `PinBufferForBlock` `bufmgr.c:1120` | 테이블 `t`, 블록 0 |
| 결과 전송 | `printtup` `printtup.c:306` | 1회 |
| 끝 | `ReadyForQuery` `dest.c:258` | |

받기 단계의 `exec_simple_query()`는 클라이언트가 SQL 문자열 하나를 통째로 보내는 방식(simple query)의 입구입니다. 이 함수가 아래 단계들을 차례로 부르고, 모든 단계가 끝나면 `ReadyForQuery()`로 다음 쿼리를 받을 준비가 됐다고 클라이언트에 알립니다. 실행 준비와 실행 단계의 함수 이름에 나오는 portal은 `portal.h` 머리 주석의 표현대로 실행 중이거나 실행할 수 있는 쿼리의 실행 상태를 나타내는 객체입니다. SQL 커서도 같은 구조를 씁니다.

문법 해석과 의미 분석이 나뉘어 있는 데에는 소스가 직접 밝힌 이유가 있습니다. `postgres.c` 590~602행의 주석에 따르면, 의미 분석과 규칙 재작성(뷰 정의 같은 규칙으로 쿼리를 고쳐 쓰는 단계)은 데이터베이스의 테이블을 읽어야 하므로 이미 실패한(abort) 트랜잭션 안에서는 할 수 없습니다. 그래서 문법 해석을 따로 떼어 두고, 실패한 트랜잭션에서는 문법 해석만으로 들어온 문장이 `COMMIT`이나 `ROLLBACK`인지 가립니다. `raw_parser()`가 만든 트리는 아직 `t`가 어떤 테이블인지 모릅니다. `analyze.c` 머리 주석이 설명하듯 이 트리를 Query 트리로 바꾸는 단계에서 참조하는 테이블에 잠금을 잡고 의미를 확정합니다. `exec_simple_query()`는 1190행에서 `pg_analyze_and_rewrite_fixedparams()`를 부르고, 이 함수가 분석(683행)과 재작성(692행)을 차례로 부릅니다.

계획 단계의 `standard_planner()`는 가능한 실행 방법 가운데 하나를 골라 계획 트리를 만듭니다. 이번 계획의 맨 위 노드는 `T_SeqScan`, 곧 테이블을 처음부터 끝까지 읽는 순차 스캔입니다. 테이블 `t`에는 아직 인덱스가 없습니다.

실행 단계에서 일이 실제로 일어납니다. `src/backend/executor/README`는 계획 트리를 수요가 당기는 파이프라인(demand-pull pipeline)이라고 설명합니다. 노드를 부르면 그 노드가 출력할 다음 행 하나를 돌려주고, 더 없으면 NULL을 돌려줍니다. 기록의 횟수가 이 모델을 그대로 보여 줍니다. 순차 스캔 노드 `ExecSeqScanWithQual()`은 2번 불렸고, 그 아래 `heap_getnextslot()`은 4번 불렸습니다. 첫 호출에서 노드는 행 1을 받아 조건 `id = 2`로 거르고, 행 2를 받아 조건을 통과시켜 돌려줍니다. 두 번째 호출에서는 행 3을 걸러 내고, 네 번째 `heap_getnextslot()`에서 행이 끝났다는 답을 받아 NULL을 돌려줍니다. 결과를 보내는 `printtup()`이 1번뿐인 것도 같은 이유입니다.

테이블 접근 단계와 버퍼 단계 사이의 관계는 `BT` 줄에서 보입니다. 아래는 `PinBufferForBlock()`에 멈췄을 때의 스택 가운데 실행기부터 버퍼까지입니다.

```text
BT 0 PinBufferForBlock src/backend/storage/buffer/bufmgr.c:1120
BT 1 StartReadBuffersImpl src/backend/storage/buffer/bufmgr.c:1328
BT 2 StartReadBuffers src/backend/storage/buffer/bufmgr.c:1504
BT 3 read_stream_start_pending_read src/backend/storage/aio/read_stream.c:356
BT 4 read_stream_look_ahead src/backend/storage/aio/read_stream.c:514
BT 5 read_stream_next_buffer src/backend/storage/aio/read_stream.c:899
BT 6 heap_fetch_next_buffer src/backend/access/heap/heapam.c:679
BT 7 heapgettup_pagemode src/backend/access/heap/heapam.c:1041
BT 8 heap_getnextslot src/backend/access/heap/heapam.c:1395
BT 9 table_scan_getnextslot src/include/access/tableam.h:1036
BT 10 SeqNext src/backend/executor/nodeSeqscan.c:81
BT 11 ExecScanFetch src/include/executor/execScan.h:134
BT 12 ExecScanExtended src/include/executor/execScan.h:195
BT 13 ExecSeqScanWithQual src/backend/executor/nodeSeqscan.c:138
```

아래에서 위로 읽습니다. 실행기의 순차 스캔 노드(BT 13)는 테이블을 직접 읽지 않고 `table_scan_getnextslot()`(BT 9)을 부릅니다. 이 함수는 `tableam.h`가 정의하는 테이블 접근 방법(table access method) 계층이고, 힙 테이블에서는 그 구현인 `heap_getnextslot()`(BT 8)으로 이어집니다. 힙 코드는 행이 든 페이지가 필요해지면 `heap_fetch_next_buffer()`에서 read stream(BT 5)에 다음 페이지를 달라고 하고, read stream이 버퍼 관리자(BT 2)를 거쳐 공유 버퍼에서 테이블 `t`의 블록 0을 찾아 고정(pin)합니다(BT 0). 스택의 나머지 15줄은 `ExecProcNodeFirst()`부터 `main()`까지로, [기록 파일](evidence/select-where-id-2.trace)에서 볼 수 있습니다.

`END` 줄에서 `read_stream_next_buffer()`는 3번, `heap_fetch_next_buffer()`는 2번입니다. 앞의 두 번은 힙 코드가 페이지를 달라고 한 호출이고, 두 번째에는 새 블록을 고정하지 않았습니다(`PinBufferForBlock()` 1회). 남은 한 번은 스캔을 끝낼 때 나옵니다. 감시 목록에 `heap_endscan()`, `read_stream_end()`, `read_stream_reset()`을 더해 같은 SQL을 다시 추적하면([`evidence/scan-end-order.trace`](evidence/scan-end-order.trace)), 마지막 `read_stream_next_buffer()`는 `heap_endscan()`이 `read_stream_end()`를 부르고 이 함수가 `read_stream_reset()`을 부른 다음에 불립니다. `read_stream_reset()`은 미리 읽어 두고 아직 쓰지 않은 버퍼의 고정을 풀려고, 버퍼가 더 나오지 않을 때까지 `read_stream_next_buffer()`를 부릅니다(`read_stream.c` 1087~1089행).

### 예측이 틀린 두 곳

이 경로에서 사전 예측이 틀린 곳이 두 군데입니다. 둘 다 함수가 소스에 있었기 때문에 이름 대조만으로는 드러나지 않았습니다.

첫째, 순차 스캔 노드로 `ExecSeqScan()`을 예측했지만 실제로는 `ExecSeqScanWithQual()`이 불렸습니다. 18.6의 `nodeSeqscan.c` 257~277행은 실행 준비 단계에서 노드의 실행 함수를 정합니다. 동시에 갱신된 행을 다시 검사하는 특수한 경우(EvalPlanQual)에는 `ExecSeqScanEPQ()`를 쓰고(262~263행), 그 밖에는 두 조건을 보고 네 함수 가운데 하나를 고릅니다. 하나는 WHERE 조건식(소스의 용어로 qual)이 있는지이고, 다른 하나는 출력할 행을 새로 만들어야 하는지(projection)입니다. `SELECT *`는 테이블의 행을 그대로 내보내므로 새로 만들 필요가 없습니다. 같은 자리의 주석은 각 함수가 해당 조건 조합에 맞춰 최적화되어 있다고 설명합니다. 조건이 없는 경우를 위한 함수가 조건 검사를 아예 하지 않는 식입니다.

둘째, 페이지를 읽는 함수로 `ReadBufferExtended()`를 예측했지만 한 번도 불리지 않았습니다. 18.6의 순차 스캔은 `heapam.c` 679행에서 read stream(`read_stream_next_buffer()`)으로 페이지를 받습니다. `read_stream.c` 머리 주석에 따르면, 페이지마다 `ReadBuffer()`를 부르면 버퍼에 없는 페이지마다 작은 I/O를 따로 기다려야 하므로 비효율적입니다. read stream은 앞으로 읽을 블록 번호를 미리 받아 이웃한 블록을 묶어서 미리 읽습니다. 다만 이 테이블은 페이지가 하나뿐이고 이미 버퍼에 있었으므로, 이번 기록에서는 묶어 읽기가 일어나지 않았습니다. 같은 주석은 I/O가 필요 없을 때는 한 블록보다 더 앞을 내다볼 이유가 없다고 설명합니다.

기록에서 `WaitReadBuffers()`는 0회였습니다. 디스크 읽기를 기다리지 않았다는 뜻입니다. 캐시를 데우는 첫 실행이 페이지를 이미 공유 버퍼에 올려 두었기 때문으로 보이지만, 이는 확인하지 않은 추론입니다. 페이지가 버퍼에 없을 때의 처음 읽기 경로는 이 실습에서 추적하지 않았습니다.

### 실습 1. 조건이 없으면

`SELECT * FROM t;`를 실행하면 순차 스캔 노드로 어느 함수가 불릴지, 그 함수와 `heap_getnextslot()`, `printtup()`이 각각 몇 번 불릴지 먼저 적은 다음 실행합니다.

```bash
sh open_source_internals/database/postgresql/scripts/run-trace.sh select.watch "SELECT * FROM t;"
```

<details>
<summary>실행 결과와 설명</summary>

`ExecSeqScan`이 4번, `heap_getnextslot`이 4번, `printtup`이 3번 불렸습니다([`evidence/exercise-1-select-all.trace`](evidence/exercise-1-select-all.trace)). 조건도 projection도 없으므로 네 함수 가운데 `ExecSeqScan()`이 골라집니다. 이번에는 노드가 받은 행을 모두 그대로 돌려주므로, 노드 호출 횟수가 출력 행 3개에 끝을 알리는 NULL 하나를 더한 4번이 됩니다. 본문의 `WHERE id = 2`에서 노드가 2번만 불린 것과 비교하면, 조건 검사가 노드 안에서 반복된다는 것이 보입니다.

</details>

### 실습 2. 열 하나만 고르면

`SELECT name FROM t WHERE id = 2;`는 네 함수 가운데 무엇을 고를지 먼저 적은 다음 실행합니다.

```bash
sh open_source_internals/database/postgresql/scripts/run-trace.sh select.watch "SELECT name FROM t WHERE id = 2;"
```

<details>
<summary>실행 결과와 설명</summary>

`ExecSeqScanWithQualProject`가 불렸습니다([`evidence/exercise-2-select-name.trace`](evidence/exercise-2-select-name.trace)). 조건 `id = 2`가 있고, 테이블의 행 `(id, name)`에서 `name`만 뽑은 새 행을 만들어야 하므로 두 조건이 모두 참입니다.

</details>

### 실습 3. 인덱스가 있으면

`id`에 인덱스를 만들고 순차 스캔을 끄면 같은 `SELECT * FROM t WHERE id = 2;`가 어떤 경로를 지날지 예측합니다. 특히 두 가지를 먼저 적어 봅니다. 하나는 인덱스 스캔이 페이지를 몇 개, 어느 relation(테이블이나 인덱스처럼 페이지로 저장되는 객체)에서 읽을지입니다. 다른 하나는 힙 페이지를 질문 1처럼 read stream으로 받을지입니다. 결과에 나오는 메타 페이지 캐시는 아직 설명하지 않은 내용이므로, 그 부분은 예측보다 결과에서 새로 발견하는 문제입니다.

인덱스를 만들고, 순차 스캔을 끈 계획이 Index Scan인지 확인합니다.

```bash
docker exec pg-lab psql -X -q -c "CREATE INDEX t_id_idx ON t(id);"
```

```bash
docker exec -e PGOPTIONS='-c enable_seqscan=off' pg-lab psql -X -c "EXPLAIN SELECT * FROM t WHERE id = 2;"
```

`Index Scan using t_id_idx on t`가 보이면 다음으로 넘어갑니다. `Bitmap Heap Scan`이 보이면 감시 목록과 다른 경로이므로 결과가 맞지 않습니다. 이 실습의 이미지에서는 Index Scan이 나왔습니다. 이어서 같은 설정으로 추적합니다. 바로 위 명령의 `PGOPTIONS`는 `psql`이 연결을 열 때 서버에 넘길 설정을 읽는 환경 변수입니다. 아래의 `PG_OPTIONS`는 실습 스크립트의 환경 변수이고, 스크립트가 이 값을 추적할 연결의 `PGOPTIONS`로 넘깁니다(`session.sh` 12행).

```bash
PG_OPTIONS='-c enable_seqscan=off' sh open_source_internals/database/postgresql/scripts/run-trace.sh index.watch "SELECT * FROM t WHERE id = 2;" heap_hot_search_buffer
```

인덱스 페이지의 구성을 본 다음 인덱스를 지웁니다. 질문 3의 기록은 인덱스가 없는 테이블에서 얻은 것이므로, 같은 조건을 맞추려는 것입니다.

```bash
docker exec pg-lab psql -X -c "CREATE EXTENSION IF NOT EXISTS pageinspect;" -c "SELECT root, level FROM bt_metap('t_id_idx');" -c "SELECT itemoffset, ctid, data FROM bt_page_items('t_id_idx', 1);"
```

```bash
docker exec pg-lab psql -X -q -c "DROP INDEX t_id_idx;"
```

<details>
<summary>실행 결과와 설명</summary>

기록은 [`evidence/exercise-3-index-scan.trace`](evidence/exercise-3-index-scan.trace), 인덱스 페이지 조회는 [`evidence/exercise-3-index-pages.out`](evidence/exercise-3-index-pages.out)에 있습니다. 계획 트리 맨 위 노드는 `T_IndexScan`이고, 경로는 다음과 같습니다. 경로의 `tid`는 테이블 행의 주소 `(블록, 줄 번호)`이고, `bt_page_items()`의 `ctid` 열이 바로 이 값입니다. 줄 번호는 질문 2에서 페이지 구조와 함께 설명합니다.

```text
ExecIndexScan -> IndexNext -> index_getnext_slot -> index_getnext_tid -> btgettuple -> _bt_first -> _bt_search
  -> _bt_getroot (t_id_idx) -> PinBufferForBlock (t_id_idx, 블록 1)
-> index_fetch_heap -> heapam_index_fetch_tuple (tid 오프셋 2) -> ReleaseAndReadBuffer (블록 0)
  -> PinBufferForBlock (t, 블록 0) -> heap_hot_search_buffer (오프셋 2)
-> printtup
-> ExecIndexScan -> ... -> btgettuple -> _bt_next -> (더 없음)
```

읽은 페이지는 두 개입니다. `bt_metap()`에서 루트가 블록 1이고 높이(level)가 0이므로, 블록 1은 루트이면서 키가 든 리프입니다. `bt_page_items()`는 그 페이지의 항목 세 개와 각 항목이 가리키는 행의 위치 `(0,1)`, `(0,2)`, `(0,3)`을 보여 줍니다. `data` 열은 키를 바이트로 보여 줍니다. `01 00 00 00 00 00 00 00`은 4바이트 정수 1을 낮은 바이트부터 적고 8바이트로 채운 것으로 읽히고, 같은 식으로 다음 두 항목은 키 2와 3입니다. B-tree는 블록 1에서 `id = 2`의 위치 `(0,2)`를 찾고, 힙 코드는 테이블 `t`의 블록 0에서 오프셋 2를 읽습니다.

블록 0인 메타 페이지는 이번 기록에서 읽지 않았습니다. `nbtpage.c` 361~363행이 보듯 `_bt_getroot()`는 먼저 relcache(테이블과 인덱스 정의의 캐시)의 `rd_amcache`에 메타 페이지 내용이 있는지 보고, 있으면 그 내용으로 루트를 찾습니다. 캐시를 데운 첫 실행이 메타 페이지 내용을 그곳에 남긴 것으로 보입니다. 첫 실행은 추적하지 않았으므로 이 부분은 추론입니다.

힙 페이지는 read stream이 아니라 `ReleaseAndReadBuffer()`로 받습니다. 인덱스가 위치를 하나씩 알려 주므로 다음에 읽을 블록을 미리 알기 어렵기 때문으로 보이지만, 이 이유는 소스 주석으로 확인하지 않은 추론입니다. 마지막 단계의 `heap_hot_search_buffer()`는 머리 주석(`heapam.c` 1698~1704행)의 설명대로 인덱스가 가리킨 행에서 시작해 같은 페이지 안의 갱신 사슬(HOT chain)을 따라가며 스냅숏에 보이는 첫 버전을 찾습니다. 이 사슬은 질문 2에서 직접 만들어 봅니다.

</details>

## 질문 2. 행 버전과 스냅숏

짧은 답은 이렇습니다. PostgreSQL의 UPDATE는 행을 제자리에서 고치지 않습니다. 옛 버전에는 "이 트랜잭션이 끝냈다"는 표시(`xmax`)를 남기고, 같은 테이블에 새 버전을 더합니다. 행마다 자신을 만든 트랜잭션(`xmin`)과 끝낸 트랜잭션(`xmax`)이 적혀 있으므로, 읽는 쪽은 자기 스냅숏과 이 두 번호만 비교해서 어느 버전을 볼지 정합니다. 스냅숏은 어떤 트랜잭션이 이미 끝났고 어떤 트랜잭션이 아직 진행 중인지를 한 시점에 찍어 둔 기록입니다. 그래서 읽기가 쓰기를 기다리지 않습니다. 대가는 옛 버전이 테이블에 남는다는 것이고, 이를 나중에 VACUUM이 치웁니다.

이 답을 페이지 안에서 직접 확인합니다.

```bash
sh open_source_internals/database/postgresql/scripts/mvcc-page.sh
```

스크립트는 새 테이블 `u(id int, v text)`에 행 하나를 넣고, 두 번 갱신하고, VACUUM합니다. 단계마다 테이블 `u`의 0번 페이지를 보여 줍니다. `NOTICE: table "u" does not exist, skipping`(처음 실행할 때)과 `NOTICE: extension "pageinspect" already exists, skipping`(실습 3을 한 뒤) 줄은 정상입니다. SQL이 실패하면 그 자리에서 0이 아닌 종료 코드로 끝납니다. 2026-10-12의 기록은 [`evidence/mvcc-page.out`](evidence/mvcc-page.out)에 있습니다. 트랜잭션 번호(xid)는 그때까지 실행한 명령에 따라 달라지고, 번호 사이의 관계만 같습니다. 아래 설명은 기록 파일의 번호를 씁니다. 새 컨테이너에서 이 문서의 순서대로 실행하면 같은 번호가 나왔습니다. 기록 파일에서는 NOTICE 줄이 실행한 자리가 아니라 출력 끝에 모여 있습니다. 표준 오류를 따로 모아 붙였기 때문입니다. 터미널에서는 해당 SQL을 실행한 자리에 나옵니다.

트랜잭션이 번호를 받는 시점도 알아 둡니다. `src/backend/access/transam/README` 193~195행에 따르면 트랜잭션은 번호가 필요한 일, 대개 행을 넣거나 고치거나 지우는 일을 처음 할 때에만 번호를 받습니다. 스크립트가 단계마다 부르는 `pg_current_xact_id()`는 공식 문서 표 9.84의 설명대로 번호가 아직 없으면 새로 받게 하므로, 이 스크립트에서는 모든 트랜잭션의 번호가 보입니다. 번호가 없으면 NULL을 돌려주는 `pg_current_xact_id_if_assigned()`도 있으며, 실습 5에서 씁니다.

### 페이지를 읽는 열

`heap_page_items()`가 보여 주는 열은 페이지의 두 층을 함께 보여 줍니다. 테이블 페이지 앞부분에는 줄 포인터(line pointer)의 배열이 있고, 각 줄 포인터가 페이지 뒤쪽에 놓인 행 데이터를 가리킵니다. 인덱스 항목도, 다른 버전의 `t_ctid`도 행 데이터의 위치가 아니라 이 줄 번호를 가리킵니다.

| 열 | 뜻 |
| --- | --- |
| `lp` | 줄 포인터 번호. 행의 주소 `(블록, 줄 번호)`의 뒷부분입니다. 소스는 이 주소를 tid라고 부릅니다. |
| `lp_flags` | 줄 포인터 상태. `itemid.h` 38~41행에 따라 0은 사용 안 함(`LP_UNUSED`), 1은 정상(`LP_NORMAL`), 2는 다른 줄로 넘김(`LP_REDIRECT`), 3은 죽음(`LP_DEAD`)입니다. |
| `lp_off` | 정상 상태에서는 행 데이터의 페이지 안 위치이고, 넘김 상태에서는 이어지는 줄 번호입니다. |
| `t_xmin`, `t_xmax` | 행 머리에 적힌 트랜잭션 번호. `htup_details.h` 122~125행의 `HeapTupleFields`에 선언되어 있으며, 주석은 각각 넣은 트랜잭션과 지우거나 잠근 트랜잭션이라고 설명합니다. |
| `t_ctid` | 이 버전의 주소, 또는 이 버전을 대체한 새 버전의 주소 |
| `flags` | 행 머리의 표시 비트 이름 |

### 네 단계로 따라가기

INSERT(xid 757) 뒤의 페이지에는 버전이 하나 있습니다. `t_xmin`은 757이고, `t_xmax`가 0이며 `HEAP_XMAX_INVALID` 표시가 있으므로 이 버전을 끝낸 트랜잭션은 없습니다. `t_ctid` `(0,1)`은 자기 자신을 가리킵니다. `lp_off` 8160은 행 데이터가 페이지 안 8160바이트 위치에 있다는 뜻입니다. 이 컨테이너의 페이지 크기는 8192바이트(준비 단계에서 확인한 `block_size`)이고, 행 데이터는 페이지 끝에서부터 앞쪽으로 쌓입니다. 뒤의 단계에서 새 버전이 8128, 8096에 놓이는 것으로 확인할 수 있습니다.

```text
 lp | lp_flags | lp_off | t_xmin | t_xmax | t_ctid |                flags
  1 |        1 |   8160 |    757 |      0 | (0,1)  | {HEAP_HASVARWIDTH,HEAP_XMAX_INVALID}
```

UPDATE(xid 758) 뒤에는 버전이 둘입니다. 줄 1의 행 데이터는 같은 위치(8160)에 그대로 있고, `t_xmax`에 758이, `t_ctid`에 새 버전의 주소 `(0,2)`가 적혔습니다. 줄 2가 새 버전이고 `t_xmin`이 758입니다. 소스에서는 `heap_update()`가 새 버전의 xmin(`heapam.c` 3861행)과 옛 버전의 xmax(4260행)에 같은 트랜잭션 번호를 넣습니다.

```text
 lp | lp_flags | lp_off | t_xmin | t_xmax | t_ctid |                               flags
  1 |        1 |   8160 |    757 |    758 | (0,2)  | {HEAP_HASVARWIDTH,HEAP_XMIN_COMMITTED,HEAP_HOT_UPDATED}
  2 |        1 |   8128 |    758 |      0 | (0,2)  | {HEAP_HASVARWIDTH,HEAP_XMAX_INVALID,HEAP_UPDATED,HEAP_ONLY_TUPLE}
```

새로 보이는 표시 비트가 넷입니다. `HEAP_UPDATED`(줄 2)는 `htup_details.h` 210행의 정의대로 이 버전이 UPDATE로 생긴 버전이라는 표시입니다. `HEAP_HOT_UPDATED`(줄 1)와 `HEAP_ONLY_TUPLE`(줄 2)은 이 갱신이 HOT(Heap Only Tuple) 갱신이라는 뜻입니다. `README.HOT`에 따르면 HOT 갱신은 인덱스가 걸린 열을 바꾸지 않는 갱신에서 새 버전을 같은 페이지에 두고, 새 버전에는 인덱스 항목을 만들지 않습니다. 테이블 `u`에는 인덱스가 없으므로 인덱스 열 조건은 늘 만족합니다. 또 하나의 조건은 새 버전이 같은 페이지에 놓이는 것입니다(`README.HOT` 56~60행). `heap_update()`는 새 버전이 페이지의 남은 공간(`pagefree`)보다 크면 다른 페이지를 구하고(`heapam.c` 4049~4055행), 새 버전이 같은 페이지에 들어갈 때만 HOT 갱신을 검토합니다(4116~4125행). 이번에는 페이지가 거의 비어 있어 같은 페이지에 들어갔습니다. 인덱스는 사슬의 첫 줄만 가리키고, 찾는 쪽은 `t_ctid`를 따라 같은 페이지 안에서 다음 버전으로 넘어갑니다. 실습 3의 `heap_hot_search_buffer()`가 이 사슬을 따라가는 함수입니다.

줄 1에 생긴 `HEAP_XMIN_COMMITTED`는 "757이 커밋했다"는 확인 결과를 행에 적어 둔 표시(hint bit)입니다. `htup_details.h` 204행이 이 비트를 `t_xmin committed`라고 정의합니다. `heapam_visibility.c` 머리 주석(6~10행)에 따르면 가시성 판정 함수는 넣거나 지운 트랜잭션이 끝난 것을 확인하면 이 비트를 적고 버퍼를 바뀐 상태로 표시합니다. 행을 읽기만 해도 페이지가 바뀔 수 있다는 뜻입니다.

세 번째 단계가 이 질문의 중심입니다. 스크립트는 REPEATABLE READ 읽기 트랜잭션을 먼저 열고 행을 한 번 읽습니다. 그 읽기가 열려 있는 동안 다른 세션이 xid 759로 `v`를 `c`로 바꾸고 커밋합니다. 기록에서 먼저 시작한 읽기는 갱신 전의 첫 번째 읽기에서 `b`를 읽었고, 갱신이 커밋된 뒤의 두 번째 읽기에서도 `b`를 읽었습니다. 갱신 뒤에 연 새 세션은 `c`를 읽었습니다. 아래 페이지는 갱신이 커밋된 직후, 두 번째 읽기 전의 상태입니다.

```text
 lp | lp_flags | lp_off | t_xmin | t_xmax | t_ctid |   (v)
  1 |        1 |   8160 |    757 |    758 | (0,2)  |   a
  2 |        1 |   8128 |    758 |    759 | (0,3)  |   b
  3 |        1 |   8096 |    759 |      0 | (0,3)  |   c

 reader 1st read | b | 759:759:
 reader 2nd read | b
```

`(v)` 열은 이해를 돕기 위해 붙인 것으로, 각 버전의 값입니다. 기록 파일에는 이 열 대신 표시 비트가 있습니다.

먼저 시작한 읽기가 `b`를 고른 이유는 스냅숏 `759:759:`에 있습니다. 공식 문서 9.27.8절의 표 9.85에 따르면 스냅숏은 `xmin:xmax:진행 중 목록` 형식입니다. `xmin`은 아직 진행 중인 가장 작은 번호이고, 그보다 작은 번호는 모두 끝났습니다. `xmax`는 끝난 번호 가운데 가장 큰 것에 1을 더한 값이고, `xmax` 이상인 번호는 스냅숏을 잡을 때 끝나지 않았으므로 보이지 않습니다. 첫 번째 읽기 때 끝난 번호 가운데 가장 큰 것은 758이므로 `xmax`는 759입니다. 진행 중인 쓰기 트랜잭션이 없었으므로 진행 중 목록은 비었고, `xmin`도 759입니다. `procarray.c` 2252~2253행은 `xmin`을 `xmax`로 시작해, 진행 중인 번호가 있을 때만 낮춥니다. REPEATABLE READ는 공식 문서 13.2.2절의 설명대로 트랜잭션의 첫 문장이 시작할 때 잡은 이 스냅숏을 트랜잭션이 끝날 때까지 씁니다.

두 번째 읽기는 위 페이지를 같은 스냅숏으로 판정합니다.

1. 줄 1(`a`)은 758이 끝냈습니다. 758은 스냅숏의 `xmax`보다 작고 진행 중 목록에도 없으므로 이미 커밋한 삭제로 보고, 이 버전은 보이지 않습니다(`heapam_visibility.c` 1153행).
2. 줄 2(`b`)는 759가 끝냈습니다. 그런데 `XidInMVCCSnapshot()`은 번호가 스냅숏의 `xmax` 이상이면 진행 중으로 판정합니다(`snapmgr.c` 1883~1885행). `HeapTupleSatisfiesMVCC()`는 지운 트랜잭션이 아직 진행 중인 것처럼 다루므로 이 버전은 보입니다(`heapam_visibility.c` 1146~1148행).
3. 줄 3(`c`)은 759가 만들었고, 같은 이유로 759를 진행 중으로 보므로 보이지 않습니다(`heapam_visibility.c` 1078~1081행).

첫 번째 읽기 때는 줄 3이 아직 없었고, 줄 2에는 끝낸 트랜잭션이 없다는 `HEAP_XMAX_INVALID` 표시가 있었습니다. 그래서 줄 2는 1086~1087행에서 바로 보이는 버전으로 판정됐습니다. 두 번의 읽기는 같은 `b`를 읽었지만, 판정에 쓰인 규칙은 서로 다릅니다.

같은 페이지, 같은 행 데이터에서 두 세션이 다른 값을 읽은 것은 행이 바뀌었기 때문이 아니라 두 세션의 스냅숏이 달랐기 때문입니다. 새 세션의 스냅숏은 759가 커밋한 뒤에 잡혔으므로 같은 규칙으로 줄 3을 고릅니다.

네 번째 단계에 들어갈 때는 먼저 시작한 읽기가 이미 끝났으므로, 줄 1이나 줄 2를 볼 수 있는 스냅숏이 남아 있지 않습니다. 이때 VACUUM을 실행한 뒤의 페이지는 이렇습니다.

```text
 lp | lp_flags | lp_off | t_xmin | t_xmax | t_ctid
  1 |        2 |      3 |        |        |
  2 |        0 |      0 |        |        |
  3 |        1 |   8160 |    759 |      0 | (0,3)
```

줄 2는 행 데이터와 함께 사라져 다시 쓸 수 있는 줄(`LP_UNUSED`)이 됐습니다. 줄 1은 행 데이터는 사라졌지만 줄 포인터가 넘김 상태(`LP_REDIRECT`)로 남았고, `lp_off` 3이 이어지는 줄 3을 가리킵니다. 정리 코드인 `pruneheap.c` 1184~1189행은 사슬에 DEAD 버전이 있으면 사슬의 첫 줄 포인터를 DEAD가 아닌 첫 버전으로 넘기고, 그 사이에서 지울 수 있는 줄은 사용 안 함으로 표시합니다. 여기서 DEAD는 어떤 스냅숏에도 더는 보이지 않아 지워도 되는 버전입니다. 다른 트랜잭션이 이미 지운 버전이라도 그 버전을 아직 볼 수 있는 스냅숏이 있으면, 죽었지만 아직 지울 수 없는 상태(`HEAPTUPLE_RECENTLY_DEAD`, `heapam.h` 126행)라서 DEAD가 아닙니다. 이번에는 VACUUM 전에 읽기 트랜잭션이 끝났으므로 줄 1과 줄 2가 모두 DEAD였고, 넘김은 줄 3으로 갔습니다. 이 결정은 사슬의 상태만 보고 하므로, 인덱스가 없는 테이블 `u`에서도 넘김이 생겼습니다. 이런 설계의 이유는 `README.HOT`이 인덱스로 설명합니다. 인덱스는 사슬의 첫 줄을 가리키므로, 첫 버전이 죽어도 그 줄 포인터를 없애면 인덱스로 살아 있는 버전을 찾을 수 없게 됩니다. 그래서 첫 줄 포인터만 남겨 다음 줄로 넘기고, 행 데이터의 공간은 회수합니다.

줄 3의 `lp_off`도 8096에서 8160으로 바뀌었습니다. VACUUM이 죽은 버전의 공간을 회수하면서 남은 행 데이터를 페이지 끝으로 옮긴 것입니다(`README.HOT` 첫 문단은 이 페이지 단위 정리를 defragmentation 또는 pruning이라고 부릅니다). 그런데 이 행의 주소 `(0,3)`은 그대로입니다. 인덱스와 `t_ctid`는 행 데이터의 위치가 아니라 줄 번호를 가리키므로, 페이지 안에서 데이터를 옮겨도 그 주소를 가진 쪽은 아무것도 고칠 필요가 없습니다. 줄 포인터 배열이 페이지 안의 간접 주소 역할을 하는 셈입니다.

### 왜 이렇게 설계했는가

공식 문서 13.1절은 이 방식(MVCC, 다중 버전 동시성 제어)의 주된 이점을 읽기용 잠금과 쓰기용 잠금이 충돌하지 않는 것, 곧 읽기가 쓰기를 막지 않고 쓰기가 읽기를 막지 않는 것으로 설명합니다. 세 번째 단계에서 갱신이 먼저 열린 읽기를 기다리지 않고 커밋하고, 읽기도 갱신을 기다리지 않은 것이 그 모습입니다. 24.1.2절은 대가를 설명합니다. UPDATE나 DELETE가 옛 버전을 바로 지우지 않는 것은 다른 트랜잭션에 아직 보일 수 있기 때문이고, 그래서 더 이상 아무에게도 보이지 않는 버전을 치우는 VACUUM이 필요합니다.

같은 목표를 다르게 이룬 설계와 비교하면 이 선택이 더 분명해집니다. MySQL 8.4 문서 17.3절에 따르면 InnoDB는 행(클러스터드 인덱스 레코드)을 제자리에서 고치고, 옛 버전을 다시 만들 정보는 undo 로그에 따로 둡니다. 행에는 마지막으로 넣거나 고친 트랜잭션 번호(`DB_TRX_ID`)와 undo 로그 레코드를 가리키는 포인터(`DB_ROLL_PTR`)가 있고, 옛 스냅숏으로 읽는 쪽은 그 포인터를 따라가 옛 값을 다시 만듭니다. 갱신용 undo 로그는 그것이 필요할 수 있는 스냅숏이 모두 사라진 뒤에야 버릴 수 있습니다. 삭제 표시된 행과 그 인덱스 레코드를 실제로 지우는 작업을 InnoDB는 purge라고 부릅니다.

| 비교 | PostgreSQL | InnoDB |
| --- | --- | --- |
| 옛 버전이 있는 곳 | 테이블 페이지 안의 다른 줄 | undo 로그 |
| 최신 버전을 읽을 때 | 버전마다 스냅숏으로 판정해 보이는 것을 고름 | 행을 그대로 읽음 |
| 옛 버전을 읽을 때 | 그 버전이 페이지에 그대로 있음 | undo 레코드로 옛 값을 다시 만듦 |
| 정리하는 쪽 | VACUUM, 그리고 페이지 하나 단위의 정리(pruning) | 갱신용 undo 로그 버리기, 그리고 purge |

POSTGRES 시절의 저장 설계를 다룬 1차 자료로 Stonebraker의 1987년 논문 *The Design of the POSTGRES Storage System*이 있습니다. 이 실습에서는 그 논문의 본문을 읽지 못했으므로(스캔 이미지 PDF), 지금의 버전 보관 방식이 그 설계에서 이어졌는지는 확인하지 않았습니다.

VACUUM이 하는 다른 일(visibility map, freeze, 통계), 오래 열린 트랜잭션이 정리를 막는 문제, 운영에서 보는 지표는 [`deep-dive/postgresql/18-postgresql-deep-dive.md`](../../../database/deep-dive/postgresql/18-postgresql-deep-dive.md)와 [`deep-dive/transactions/12-mvcc-snapshot-visibility.md`](../../../database/deep-dive/transactions/12-mvcc-snapshot-visibility.md)가 다룹니다. MVCC의 역사와 다른 DBMS의 방식은 [`database/mvcc.md`](../../../database/mvcc.md)에 정리되어 있습니다.

### 실습 4. 격리 수준을 바꾸면

공식 문서 13.2.1절에 따르면 READ COMMITTED의 SELECT는 그 쿼리가 시작한 순간의 스냅숏을 봅니다. 먼저 시작한 읽기를 READ COMMITTED로 바꾸면 두 번째 읽기에서 무엇을 읽을지, 이 규칙과 본문의 판정 규칙으로 예측합니다. `ISOLATION` 값이 `READ COMMITTED`, `REPEATABLE READ`, `SERIALIZABLE` 가운데 하나가 아니면 스크립트는 종료 코드 2로 끝납니다.

```bash
ISOLATION='READ COMMITTED' sh open_source_internals/database/postgresql/scripts/mvcc-page.sh
```

<details>
<summary>실행 결과와 설명</summary>

첫 번째 읽기는 `b`, 두 번째 읽기는 `c`였습니다([`evidence/exercise-4-read-committed.out`](evidence/exercise-4-read-committed.out)). 페이지 내용은 REPEATABLE READ 때와 같습니다. 달라진 것은 스냅숏을 잡는 시점입니다. READ COMMITTED에서는 한 트랜잭션 안의 두 SELECT가 각자 스냅숏을 잡으므로 다른 데이터를 볼 수 있습니다(공식 문서 13.2.1절). 두 번째 읽기가 새 스냅숏을 잡을 때는 세 번째 버전을 만든 트랜잭션이 이미 커밋했으므로, 같은 판정 규칙이 이번에는 줄 3을 고릅니다. 기록 파일의 트랜잭션 번호(761~763)는 본문과 다르지만 번호 사이의 관계는 같습니다.

</details>

## 질문 3. synchronous_commit이 바꾸는 것

먼저 WAL(Write-Ahead Log)이 무엇인지 적습니다. 공식 문서 28.3절에 따르면 WAL의 중심 개념은 데이터 파일을 고치기 전에 그 변경을 설명하는 WAL 레코드를 먼저 영구 저장소에 내려 두는 것입니다. 그러면 서버가 죽어도 WAL 레코드로 변경을 다시 적용할 수 있으므로, 커밋할 때마다 데이터 페이지를 디스크에 쓰지 않아도 됩니다. 그래서 커밋이 확실히 남았다고 말하려면 그 커밋 레코드까지의 WAL이 디스크에 있어야 합니다.

짧은 답은 이렇습니다. 기본값 `on`에서는 COMMIT을 실행한 백엔드가 커밋 레코드까지 WAL을 직접 디스크에 내리고(`fdatasync`), 그것이 끝난 뒤에 클라이언트에 성공을 알립니다. `off`에서는 백엔드가 WAL을 내리지 않고 바로 성공을 알리며, 별도 프로세스인 WAL writer가 조금 뒤에 내립니다. 그 사이에 서버가 죽으면 성공을 받은 트랜잭션이 사라질 수 있습니다. 공식 문서 28.4절은 이 선택의 이점을 첫 문장에 적습니다. 비동기 커밋은 트랜잭션이 더 빨리 끝나게 해 주고, 그 대가로 서버가 죽으면 가장 최근의 트랜잭션을 잃을 수 있습니다. 아래 기록에서 보듯 빨라지는 몫은 커밋마다 장치 쓰기를 기다리지 않는 데서 나옵니다.

### 소스의 분기

갈림길은 `xact.c`의 `RecordTransactionCommit()` 한 곳입니다. 1498~1502행은 WAL을 썼고 트랜잭션 번호가 있으며 `synchronous_commit`이 `SYNCHRONOUS_COMMIT_OFF`보다 클 때 `XLogFlush()`를 부릅니다. 지울 파일이 있는 경우처럼 반드시 동기로 처리해야 하는 조건에서도 같은 쪽으로 갑니다. 나머지 경우에는 1523행에서 `XLogSetAsyncXactLSN()`을 불러 "이 위치까지 내려 달라"는 표시만 남깁니다. 트랜잭션 번호도 없고 WAL도 만들지 않은 트랜잭션은 이 분기에 오기 전에 1404~1405행에서 빠져나갑니다(실습 5).

그 위의 주석(1473~1496행)은 언제 비동기 커밋을 허용하고 언제 반드시 즉시 내려야 하는지를 설명하고, 1512~1522행은 대가를 적습니다. 비동기 커밋은 postmaster가 비정상 종료하면 아직 쓰지 않은 WAL 버퍼와 함께 커밋한 트랜잭션을 잃을 수 있고, 남은 WAL은 WAL writer가 내립니다. 공식 문서 28.4절은 잃는 것이 데이터 손상이 아니라 최근 트랜잭션이라는 점을 분명히 합니다. 복구된 데이터베이스는 스스로 일관된 상태이고, 아직 내리지 못한 트랜잭션만 빠집니다. 위험 구간은 WAL writer가 `wal_writer_delay`(준비 단계에서 확인한 200ms)마다 WAL을 내리므로 제한되며, 문서는 최대 길이를 그 세 배로 적습니다. 같은 절은 이 설정을 트랜잭션마다 바꿀 수 있다고 설명합니다. 잃어도 되는 기록과 잃으면 안 되는 기록을 한 서버에서 섞어 쓸 수 있다는 뜻입니다.

### gdb로 본 차이

두 설정에서 같은 INSERT를 추적합니다.

```bash
PG_OPTIONS='-c synchronous_commit=on' sh open_source_internals/database/postgresql/scripts/run-trace.sh commit.watch "INSERT INTO t VALUES (4, 'd');" pg_fdatasync
```

```bash
PG_OPTIONS='-c synchronous_commit=off' sh open_source_internals/database/postgresql/scripts/run-trace.sh commit.watch "INSERT INTO t VALUES (4, 'd');"
```

`psql` 결과는 두 번 모두 `INSERT 0 1`입니다. 캐시를 데우는 실행까지 포함해 한 번 추적할 때마다 행이 2개씩 늘어납니다. 기록은 [`evidence/commit-sync-on.trace`](evidence/commit-sync-on.trace)와 [`evidence/commit-sync-off.trace`](evidence/commit-sync-off.trace)에 있습니다. WAL 레코드는 종류마다 그것을 만들고 다시 적용하는 모듈이 정해져 있고, 소스는 이 모듈을 resource manager, 줄여서 rmgr라고 부릅니다(`rmgrlist.h`의 목록은 rmgr마다 재적용 함수와 설명 함수를 적습니다). 아래 표에서 `XLogInsert`에 붙은 rmgr이 그 호출이 만든 레코드의 종류입니다.

| 순번 | `on` | `off` |
| --- | --- | --- |
| 1~7 | `exec_simple_query`, `ExecInsert`, `heap_insert`(테이블 `t`), `XLogInsert`(rmgr `Heap`), `CommitTransaction`, `RecordTransactionCommit`, `XLogInsert`(rmgr `Transaction`) | 같음 |
| 8 | `XLogFlush` | `XLogSetAsyncXactLSN` |
| 9 | `XLogWrite` | `ReadyForQuery` |
| 10 | `issue_xlog_fsync`(fd=10) | |
| 11 | `pg_fdatasync`(fd=10) | |
| 12 | `ReadyForQuery` | |

행을 넣는 일과 WAL 레코드를 만드는 일은 두 설정에서 똑같습니다. 첫 `XLogInsert()`는 행 삽입의 기록(rmgr `Heap`)이고, 두 번째는 커밋 레코드(rmgr `Transaction`)입니다. 달라지는 것은 8번부터입니다. `on`의 `pg_fdatasync()` 호출 스택에는 `RecordTransactionCommit src/backend/access/transam/xact.c:1502`가 있어, 소스의 분기에서 `XLogFlush()`로 들어간 바로 그 줄을 지났음을 확인할 수 있습니다. `issue_xlog_fsync()`가 `pg_fdatasync()`를 부르는 것은 `wal_sync_method`가 `fdatasync`이기 때문입니다(`xlog.c` 8781행). `off`에서는 `XLogSetAsyncXactLSN()` 한 번 뒤에 바로 `ReadyForQuery()`로 끝납니다.

### 시스템 콜로 본 차이

gdb 기록은 백엔드 안에서 일어난 일만 보여 줍니다. `off`에서 WAL을 누가 언제 내리는지 보려면 WAL writer 프로세스도 함께 봐야 합니다.

먼저 두 시스템 콜의 역할을 구분합니다. `pwrite64`는 파일의 지정한 위치에 데이터를 씁니다. 이것만으로는 데이터가 아직 커널의 메모리에 있을 수 있습니다. `fdatasync`는 `fdatasync(2)` 매뉴얼의 설명대로 그 파일의 바뀐 데이터를 저장 장치로 내리고, 장치가 끝났다고 알릴 때까지 돌아오지 않습니다. 그래서 커밋이 디스크에 남았다고 말할 수 있는 시점은 `pwrite64`가 아니라 `fdatasync`가 끝난 뒤입니다.

```bash
PG_OPTIONS='-c synchronous_commit=on' sh open_source_internals/database/postgresql/scripts/run-strace.sh pwrite64,fdatasync "INSERT INTO t VALUES (4, 'd');"
```

```bash
PG_OPTIONS='-c synchronous_commit=off' sh open_source_internals/database/postgresql/scripts/run-strace.sh pwrite64,fdatasync "INSERT INTO t VALUES (4, 'd');"
```

출력의 마지막 부분이 `역할 시각 시스템 콜` 형식의 기록입니다. 2026-10-12의 기록([`evidence/commit-sync-on.strace`](evidence/commit-sync-on.strace), [`evidence/commit-sync-off.strace`](evidence/commit-sync-off.strace))에서 핵심 줄은 이렇습니다(쓴 내용은 줄였습니다).

```text
on:
backend   11:50:58.909812 pwrite64(10</var/lib/pgsql/data/pg_wal/000000010000000000000001>, "...", 8192, 7790592) = 8192
backend   11:50:58.909983 fdatasync(10</var/lib/pgsql/data/pg_wal/000000010000000000000001>) = 0

off:
backend   11:51:01.724293 +++ exited with 0 +++
walwriter    11:51:01.921745 pwrite64(6</var/lib/pgsql/data/pg_wal/000000010000000000000001>, "...", 8192, 7790592) = 8192
walwriter    11:51:01.922128 fdatasync(6</var/lib/pgsql/data/pg_wal/000000010000000000000001>) = 0
```

`on`에서는 백엔드가 WAL 파일에 8192바이트를 쓰고(`pwrite64`) 이어서 `fdatasync`를 부릅니다. 커밋을 알리기 전에 이 장치 쓰기를 기다리는 것이 `on`의 비용입니다. 다만 이 기록으로는 그 비용의 크기를 알 수 없습니다. strace가 줄 앞에 적는 시각은 시스템 콜이 시작된 시각이라, 두 줄의 차이 0.17ms는 `pwrite64`가 시작되고 `fdatasync`가 시작되기까지의 시간, 곧 거의 `pwrite64`에 걸린 시간입니다. 시각이 시작 시각이라는 것은 `docker exec pg-lab strace -tt -T -e trace=clock_nanosleep,nanosleep,exit_group sleep 0.3`으로 확인할 수 있습니다. `clock_nanosleep` 줄의 시각에 그 줄 끝에 적힌 걸린 시간(약 0.3초)을 더하면 다음 줄의 시각과 거의 같습니다. `fdatasync`가 걸린 시간은 재지 않았습니다(호출마다 걸린 시간을 적는 `-T`를 쓰지 않았습니다). 백엔드가 `fdatasync`를 시작하고 5.5ms 뒤(11:50:58.915492)에 끝났으므로 그보다 짧았다는 것만 압니다. 이 시간은 저장 장치에 따라 달라집니다.

`off`에서는 백엔드가 WAL을 한 번도 쓰지 않고 끝납니다. WAL writer가 같은 파일의 같은 위치(오프셋 7790592)에 같은 크기로 쓰고 `fdatasync`를 부른 것은 백엔드 프로세스가 끝나고 0.197초 뒤입니다. 커밋은 프로세스가 끝나기 전에 일어나므로, 커밋에서 WAL writer의 쓰기까지는 0.197초보다 깁니다. 이 실행은 커밋 시각을 남기지 않아 얼마나 더 긴지는 모릅니다. 아래 재실행 기록에서는 커밋 레코드의 시각부터 백엔드 종료까지 1.7ms였습니다.

이 간격은 고정값이 아닙니다. 이 문서의 명령을 다시 실행한 기록([`evidence/commit-sync-off-rerun.strace`](evidence/commit-sync-off-rerun.strace))에서는 커밋 레코드에 적힌 시각 12:18:05.695110([`evidence/waldump-after-replay.txt`](evidence/waldump-after-replay.txt))부터 WAL writer의 쓰기 12:18:05.780552까지 0.085초였습니다. 이 기록에는 WAL writer의 쓰기가 한 번 더 있습니다(12:18:05.991471). 두 번 모두 오프셋 7987200부터 8192바이트, 곧 WAL 파일의 같은 페이지입니다. 두 번째 쓰기가 무엇을 내렸는지는 확인하지 않았습니다. WAL writer가 같은 페이지를 다시 쓰는 일은 [on인데 WAL writer 줄이 보일 때](#on인데-wal-writer-줄이-보일-때)에서 다룹니다. 두 값 모두 공식 문서 28.4절이 적은 최대 위험 구간(`wal_writer_delay`의 세 배, 600ms) 안에 있습니다. `xlog.c` 2630~2637행의 주석에 따르면 `XLogSetAsyncXactLSN()`은 WAL writer가 잠들어 있으면 깨워서, 비동기 커밋이 예상한 시간 안에 디스크에 닿게 합니다. 이번 간격이 WAL writer의 주기와 깨우기 가운데 어느 쪽으로 정해졌는지는 재지 않았습니다. strace와 `pg_waldump`의 시각은 모두 컨테이너의 UTC 기준입니다.

두 설정 모두 같은 오프셋에 8192바이트를 썼다는 점도 눈여겨볼 만합니다. 이번 INSERT의 WAL 레코드는 수십 바이트이지만(아래 `pg_waldump` 출력에서 행 삽입 61바이트, 커밋 34바이트), 파일에는 WAL 페이지 크기(`pg_config.h` 786행의 `XLOG_BLCKSZ`, 8192)만큼 썼습니다.

### on인데 WAL writer 줄이 보일 때

같은 명령을 다시 실행하면 `on`에서도 WAL writer 줄이 보일 수 있습니다. 2026-10-12에 다시 실행했을 때 실제로 그랬습니다([`evidence/commit-sync-on-rerun.strace`](evidence/commit-sync-on-rerun.strace)).

```text
backend   12:04:16.847283 pwrite64(10<.../000000010000000000000001>, "...", 8192, 7987200) = 8192
backend   12:04:16.847889 fdatasync(10<.../000000010000000000000001>) = 0
walwriter    12:04:17.787390 pwrite64(6<.../000000010000000000000001>, "...", 8192, 7987200) = 8192
walwriter    12:04:17.787741 fdatasync(6<.../000000010000000000000001>) = 0
```

이 줄을 보고 커밋이 비동기로 처리됐다고 읽으면 안 됩니다. 백엔드는 WAL writer보다 먼저 자기 커밋까지의 WAL을 직접 쓰고 `fdatasync`까지 마쳤습니다. WAL writer가 무엇을 내렸는지는 WAL 내용을 `pg_waldump`로 직접 보면 알 수 있습니다.

```bash
docker exec -u postgres pg-lab sh -c 'pg_waldump -p /var/lib/pgsql/data/pg_wal 000000010000000000000001 2>/dev/null | tail -12'
```

출력의 한 줄은 WAL 레코드 하나입니다. `rmgr`는 레코드의 종류, `tx`는 트랜잭션 번호, `lsn`은 WAL 안의 위치, `desc`는 내용입니다. 이 문서의 순서대로 실행했다면 질문 3의 INSERT마다 `Heap`의 `INSERT`와 `Transaction`의 `COMMIT`이 한 쌍씩 보입니다. `Standby`의 `RUNNING_XACTS`는 실행 시점에 따라 보이지 않을 수 있습니다. 출력이 비어 있으면 WAL 파일 이름이 다른 것이므로 `docker exec pg-lab ls /var/lib/pgsql/data/pg_wal`로 확인합니다.

다시 실행했을 때의 WAL([`evidence/waldump-after-commit.txt`](evidence/waldump-after-commit.txt))에서는 백엔드가 내린 커밋 레코드(xid 767, 12:04:16.847219) 바로 뒤에 `Standby` rmgr의 `RUNNING_XACTS` 레코드 하나가 있고, 그다음이 다음 INSERT(xid 768)입니다. xid 768의 커밋(12:04:18.705289)은 WAL writer의 쓰기(12:04:17.787390)보다 뒤이므로 그 쓰기에 들어 있을 수 없습니다. 그 사이에 들어온 레코드는 `RUNNING_XACTS` 하나뿐입니다.

위치도 맞습니다. LSN `0/0179F4C8`은 WAL 전체에서의 바이트 위치 0x0179F4C8입니다. WAL 파일 하나의 크기는 16MB(0x01000000바이트, 준비 단계의 `wal_segment_size`)이고, 파일 `000000010000000000000001`은 위치 0x01000000부터를 담습니다. 그래서 파일 안 오프셋은 0x0179F4C8 - 0x01000000 = 0x79F4C8, 십진수로 7992520입니다. 이 오프셋은 WAL writer가 다시 쓴 8KB 페이지(7987200부터 8192바이트) 안에 있습니다.

`RUNNING_XACTS`는 진행 중인 트랜잭션 목록을 적는 레코드입니다. 소스에서 이 레코드를 정기적으로 남기는 곳은 bgwriter입니다. bgwriter는 `bgwriter.c` 머리 주석의 설명대로 백엔드가 공유 버퍼의 바뀐 페이지를 직접 쓰지 않아도 되게 미리 써 두는 백그라운드 프로세스입니다. 같은 파일 254~258행의 주석은 복제 서버가 일관된 상태에 더 빨리 도달하도록 이 기록을 `LOG_SNAPSHOT_INTERVAL_MS`(15초, 70행)마다 남긴다고 설명하고, 264~267행은 그 사이 WAL 기록이 있었을 때만 남긴다고 덧붙입니다. 이 컨테이너의 `wal_level`은 `replica`입니다.

여기까지에서 두 가지는 추론입니다. 체크포인트(checkpoint)는 그때까지 바뀐 데이터 페이지를 모두 디스크에 쓰고 WAL에 체크포인트 레코드를 남기는 작업이고, 장애 뒤 복구는 마지막 체크포인트 레코드를 보고 WAL을 어디서부터 다시 적용할지 정합니다(공식 문서 28.5절). 체크포인트도 같은 `RUNNING_XACTS` 레코드를 남기므로(`xlog.c` 7245행), 이번 레코드를 남긴 프로세스가 bgwriter인지는 확인하지 않았습니다. WAL writer가 그 페이지를 다시 쓴 것이 이 레코드 때문이라는 것도 레코드의 순서와 위치로 판단했고, 쓴 내용을 직접 대조하지는 않았습니다. 첫 실행에서 WAL writer 줄이 없었던 것은 이런 레코드가 기록 구간에 들어오지 않았기 때문으로 보입니다.

### 비동기 커밋이 다른 곳에 미치는 영향

질문 2에서 본 hint bit도 이 설정의 영향을 받습니다. `heapam_visibility.c` 82~112행의 `SetHintBits()` 주석에 따르면, 커밋을 확인했다는 표시는 그 트랜잭션의 커밋 레코드가 페이지보다 먼저 디스크에 내려간다고 보장될 때만 남길 수 있습니다. 비동기 커밋은 커밋 레코드가 아직 내려가지 않았을 수 있으므로, 그동안 읽는 쪽은 표시를 남기지 못하고 나중으로 미룹니다. 설정 하나가 COMMIT의 대기 시간만이 아니라 다른 세션의 읽기 경로에도 닿는 셈입니다. 이 영향은 소스로만 확인했고 실행으로 관측하지는 않았습니다.

### 실습 5. SELECT의 자동 커밋

`synchronous_commit=on`에서 `SELECT * FROM t;`를 자동 커밋으로 실행할 때도 백엔드가 `fdatasync`를 부를지 예측합니다. `commit.watch`의 함수 가운데 어디까지 지날지도 적어 봅니다. 질문 2에서 본, 트랜잭션이 번호를 받는 시점이 실마리입니다.

```bash
PG_OPTIONS='-c synchronous_commit=on' sh open_source_internals/database/postgresql/scripts/run-trace.sh commit.watch "SELECT * FROM t;"
```

<details>
<summary>실행 결과와 설명</summary>

`exec_simple_query`, `CommitTransaction`, `RecordTransactionCommit`, `ReadyForQuery`만 지났고, `XLogInsert`, `XLogFlush`, `pg_fdatasync`는 0회였습니다([`evidence/exercise-5-select-commit.trace`](evidence/exercise-5-select-commit.trace)). 읽기만 한 트랜잭션도 끝날 때 커밋 처리를 거칩니다. 하지만 행을 쓰지 않았으므로 트랜잭션 번호를 받지 않았을 것입니다. 추적한 트랜잭션의 번호는 기록에 없으므로, 같은 종류의 문장에서 확인합니다.

```bash
docker exec pg-lab psql -X -At -c "SELECT pg_current_xact_id_if_assigned() IS NULL AS no_xid FROM t LIMIT 1;"
```

`t`(참)가 나오면 이 SELECT의 트랜잭션은 번호를 받지 않은 것입니다([`evidence/exercise-5-xid.out`](evidence/exercise-5-xid.out)). `RecordTransactionCommit()`의 1350~1354행은 번호가 없으면 커밋 레코드를 쓸 수도 없고 쓸 필요도 없다고 설명하고, WAL을 하나도 만들지 않았으면 1404~1405행에서 그대로 빠져나갑니다. 그래서 `synchronous_commit`이 `on`이어도 기다릴 것이 없습니다.

번호 없이 WAL을 만드는 경우도 있습니다. 읽는 도중 페이지 정리(HOT pruning)가 일어난 경우입니다. 1398~1403행의 주석은 이때 남은 WAL을 커밋 레코드처럼 내리게 한다고 설명하지만, 트랜잭션 번호가 없으므로 1498행의 조건이 거짓이 되어 비동기 쪽으로 갑니다. 1479~1482행의 주석이 그 이유를 적어 두었습니다. 서버가 죽어 이런 트랜잭션을 잃어도 상관이 없고, 페이지 정리는 나중에 다시 하면 됩니다. 이 경우는 이번 실습에서 실행하지 않았습니다.

</details>

## 덮고 다시 확인하기

문서를 덮고 아래 질문에 답합니다. 1번과 3번은 이 문서의 명령을 조금 바꿔 실행해 확인할 수 있고, 2번, 4번, 5번은 문서와 소스의 근거로 답하는 질문입니다.

1. `SELECT name FROM t;`는 순차 스캔 함수 가운데 무엇을 고르고, 노드는 몇 번 불릴까요. 새 컨테이너에서 `run-trace.sh select.watch "SELECT name FROM t;"`로 확인합니다. 질문 3을 실행한 컨테이너에서는 행 수가 달라 횟수도 달라집니다.
2. 질문 1의 기록에서 `heap_getnextslot()`이 4번 불렸습니다. 행이 3개인데 4번인 이유를 실행기의 호출 방식으로 설명하세요.
3. 먼저 열어 둔 REPEATABLE READ 읽기 트랜잭션이 VACUUM할 때까지 끝나지 않았다면, VACUUM 뒤 줄 1과 줄 2는 어떤 상태일까요. 그 읽기의 스냅숏이 어느 버전을 아직 볼 수 있는지로 판단합니다. `mvcc-page.sh`의 VACUUM 단계를 읽기 트랜잭션이 끝나기 전으로 옮기면 확인할 수 있습니다.
4. `synchronous_commit=off`로 커밋한 직후 서버 전체(postmaster와 모든 백엔드)가 비정상 종료하면, 그 트랜잭션은 남을까요. 데이터베이스가 깨질까요. 이 실습은 이 상황을 실행하지 않았으므로 공식 문서 28.4절로 답합니다.
5. strace 기록에서 WAL writer 줄이 보인다는 사실만으로 그 커밋이 비동기였다고 말할 수 없는 이유는 무엇일까요.

## 이 실습이 확인하지 않은 것

- 페이지가 공유 버퍼에 없을 때의 처음 읽기 경로입니다. 추적 스크립트가 캐시를 먼저 데우므로 `WaitReadBuffers()`가 불리지 않았습니다. 18.6은 비동기 I/O 방식의 기본값이 `worker`이고(`aio.h` 42행의 `DEFAULT_IO_METHOD`), `method_worker.c` 머리 주석은 이 방식이 별도의 I/O worker 프로세스가 읽기를 수행하는 것이라고 설명합니다. 그래서 처음 읽기에서는 실제 시스템 콜을 백엔드가 아닌 다른 프로세스가 부를 수 있습니다. 이 부분은 소스로만 확인했고 실행으로 보지 않았습니다.
- `fdatasync` 아래, 곧 커널과 장치에서 일어나는 일입니다. 이 실습의 컨테이너는 macOS 위의 OrbStack Linux VM(arm64)에서 돌고, 커널 추적은 하지 않았습니다.
- 성능입니다. `-O0` 빌드이고 gdb와 strace가 붙어 있으므로, 이 실습의 시각 차이는 순서를 보는 데만 씁니다.
- 비동기 커밋 뒤 서버가 비정상 종료했을 때 실제로 커밋이 사라지는지입니다. 질문 3의 위험은 소스 주석과 공식 문서로만 설명했습니다.
- PostgreSQL이 버전을 덮어쓰지 않는 저장 방식을 택한 역사적 경위입니다. 1987년 논문을 읽지 못했습니다.

## 근거와 점검

| 종류 | 위치 |
| --- | --- |
| 실행 기록 | [`evidence/`](evidence/) 폴더. 파일마다 이 문서에서 처음 인용한 자리에 링크했습니다. |
| 소스 | 이미지 안의 `/src/postgresql-18.6`. 함수의 정의 위치는 `docker exec pg-lab sh /lab/find-def.sh 함수이름`으로 찾습니다. |
| 설계 문서 | `src/backend/executor/README`, `src/backend/parser/README`, `src/backend/access/heap/README.HOT`, `src/backend/access/transam/README` |
| 공식 문서 | PostgreSQL 18 [51.2](https://www.postgresql.org/docs/18/connect-estab.html), [13.1](https://www.postgresql.org/docs/18/mvcc-intro.html), [13.2](https://www.postgresql.org/docs/18/transaction-iso.html), [24.1.2](https://www.postgresql.org/docs/18/routine-vacuuming.html), [28.3](https://www.postgresql.org/docs/18/wal-intro.html), [28.4](https://www.postgresql.org/docs/18/wal-async-commit.html), [28.5](https://www.postgresql.org/docs/18/wal-configuration.html), [9.27.8](https://www.postgresql.org/docs/18/functions-info.html)(표 9.84, 9.85), MySQL 8.4 [17.3](https://dev.mysql.com/doc/refman/8.4/en/innodb-multi-versioning.html), Linux man-pages [fork(2)](https://man7.org/linux/man-pages/man2/fork.2.html), [fdatasync(2)](https://man7.org/linux/man-pages/man2/fdatasync.2.html) |
| 시험 계획과 판정 | [시험 WORK](../../../docs/works/WORK_20261012_POSTGRESQL_SOURCE_LAB_PILOT.md) |

이 문서 본문이 함수 이름(`이름()`), 구조체 이름, 대문자 상수로 인용한 C 이름이 빌드한 소스에 실제로 정의되어 있는지는 아래 명령으로 확인합니다. 정의를 찾지 못한 이름이 있으면 그 이름을 출력하고 종료 코드 1로 끝납니다. 줄 번호 인용은 이 명령이 확인하지 않습니다.

```bash
sh open_source_internals/database/postgresql/scripts/check-readme.sh
```

실습을 마치면 컨테이너를 지웁니다.

```bash
docker rm -f pg-lab
```
