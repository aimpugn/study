# 오픈소스 내부 작업대

## 개요

작업대는 오픈소스 서버가 요청 하나를 처리하는 동안 실제로 지나간 소스 줄을 브라우저에서 한 단계씩 넘겨 보는 도구입니다. 화면 왼쪽에는 디버거가 멈춘 단계들이 차례로 나오고, 각 단계가 어느 소스 디렉터리에 속하는지가 가로 위치로 그려집니다. 오른쪽에는 고른 단계에서 멈춘 함수의 소스와, 그 순간 디버거가 읽은 값과 호출 스택이 나옵니다. 지금 다루는 주제는 [PostgreSQL 18.6 실습](../database/postgresql/README.md) 하나입니다.

작업대를 쓰는 방법은 둘입니다.

- 재생본: 저장해 둔 실행 기록 하나를 HTML 파일 하나로 만든 것입니다. 서버도 Docker도 필요 없이 브라우저로 파일을 열면 됩니다. 지금 있는 재생본은 [`select-where-id-2.html`](../database/postgresql/replays/select-where-id-2.html)입니다.
- 실시간 실행: 내 컴퓨터에서 작업대 서버를 켜고, 브라우저에 SQL을 입력해 그 자리에서 추적합니다. 실습 이미지와 컨테이너가 필요합니다.

화면의 문구는 실행 기록, 소스 파일, 작업대가 붙인 "분류", 그리고 도구를 쓰는 방법과 실행 상태를 알리는 안내 가운데 하나입니다. 작업대는 PostgreSQL이 무슨 일을 하는지 설명하는 문장을 쓰지 않습니다. 각 단계가 무슨 일을 하는지는 실습 README가 설명하고, 작업대는 그 설명이 기대는 실제 기록과 소스를 보여 줍니다.

이 문서를 읽고 나면 다음을 할 수 있어야 합니다.

- 화면의 각 부분이 기록의 어느 줄에서 왔는지 짚기
- 재생본을 열고, 새 기록으로 재생본을 만들고, 그 재생본의 소스가 원본과 같은지 확인하기
- 작업대 서버를 켜서 SQL을 추적하고, 실패 응답이 무엇을 뜻하는지 읽기
- 추적을 중간에 끊을 때 PostgreSQL 서버가 왜 다시 시작될 수 있고, 작업대가 그것을 어떻게 막는지 설명하기

## 목차

- [화면의 각 부분은 어디서 오는가](#화면의-각-부분은-어디서-오는가)
- [재생본 열기와 만들기](#재생본-열기와-만들기)
- [실시간 실행](#실시간-실행)
- [실행 한 번이 지나는 길](#실행-한-번이-지나는-길)
- [안전 경계](#안전-경계)
- [실패 응답 읽기](#실패-응답-읽기)
- [추적을 끊을 때의 함정 네 가지](#추적을-끊을-때의-함정-네-가지)
- [검증](#검증)
- [한계와 다시 열 조건](#한계와-다시-열-조건)
- [근거](#근거)

## 화면의 각 부분은 어디서 오는가

작업대가 읽는 것은 실습 스크립트 [`run-trace.sh`](../database/postgresql/scripts/run-trace.sh)가 표준 출력에 남기는 기록입니다. 이 스크립트는 psql 연결 하나를 열고, 그 연결을 맡은 PostgreSQL 백엔드 프로세스에 gdb를 붙인 뒤, 감시 목록의 함수에 멈출 때마다 한 줄씩 적습니다. 감시 목록은 "어느 함수에서 멈추고 무슨 값을 읽을지"를 적은 파일입니다(`scripts/select.watch`, `scripts/index.watch`, `scripts/commit.watch`). 기록의 실제 줄은 이렇게 생겼습니다([`evidence/select-where-id-2.trace`](../database/postgresql/evidence/select-where-id-2.trace)에서 가져왔습니다).

```text
STOP 1 exec_simple_query src/backend/tcop/postgres.c:1014 query_string=0xaaaafe32e8e8 "SELECT * FROM t WHERE id = 2;"
STOP 18 PinBufferForBlock src/backend/storage/buffer/bufmgr.c:1120 (char *) rel->rd_rel->relname.data=0xffff93ad50cc "t" blockNum=0
BT 0 PinBufferForBlock src/backend/storage/buffer/bufmgr.c:1120
BT 28 main src/backend/main/main.c:227
END ExecSeqScanWithQual=2 ExecutePlan=1 PinBufferForBlock=1 ...
```

`STOP` 줄 하나가 화면의 단계 하나입니다. 순번, 멈춘 함수, 소스 상대경로와 줄 번호, 감시 목록이 읽으라고 한 식과 그 값이 차례로 있습니다. `BT` 줄은 호출 스택의 한 칸이고, 0번이 멈춘 함수, 숫자가 커질수록 그 함수를 부른 바깥 함수입니다. 호출 스택은 주제 정의에서 고른 함수(SELECT 감시 목록이면 `PinBufferForBlock`)에 처음 멈췄을 때만 남습니다. `END` 줄은 함수별로 멈춘 횟수입니다. 기록 형식의 정본은 [`trace.py`](../database/postgresql/scripts/trace.py)의 머리 주석이고, 작업대는 이 형식을 바꾸지 않고 읽기만 합니다.

화면의 각 부분과 그 출처는 아래와 같습니다.

| 화면 | 출처 | 누가 정했는가 |
| --- | --- | --- |
| 실행 흐름의 행과 함수 이름 | `STOP` 줄의 순번과 함수 | gdb 기록 |
| 행의 가로 위치와 위쪽 범례(명령 처리, 구문 분석, 실행기 등) | `STOP` 줄의 소스 경로를 [`workbench.json`](../database/postgresql/workbench.json)의 `layers` 접두사와 맞춘 결과 | 작업대의 분류. 오른쪽 위, 호출 스택, 행의 툴팁에는 "분류:"를 붙이고, 범례는 바로 위 안내에서 디렉터리를 보고 붙인 분류라고 밝힙니다. |
| 오른쪽 위의 함수, 경로와 줄 | `STOP` 줄 | gdb 기록 |
| 소스 줄 | 실습 이미지 안의 18.6 소스 트리(`/src/postgresql-18.6`) | 소스 파일 |
| 파란 줄과 노란 줄 | 파란 줄은 함수 정의가 시작하는 줄, 노란 줄은 멈춘 줄 | 작업대가 소스에서 찾은 위치와 gdb 기록 |
| 멈춘 순간 gdb가 읽은 값 | `STOP` 줄의 `식=값` | gdb 기록 |
| 호출 스택 | `BT` 줄 | gdb 기록 |
| 함수별 멈춘 횟수 | `END` 줄 | gdb 기록 |
| psql 출력 | 기록 뒤의 `--- psql` 아래 | psql |

오른쪽 소스 창은 파일 전체가 아니라 구간입니다. PostgreSQL 소스는 함수 이름을 줄의 맨 앞에 쓰는 관례를 따르므로, 작업대는 멈춘 줄에서 위로 올라가 `함수이름(`으로 시작하는 줄을 찾고, 그 위의 반환형 줄과 머리 주석부터 멈춘 줄 아래 12줄까지를 보여 줍니다. 이 범위가 70줄을 넘으면 함수 머리(주석부터 여는 중괄호까지)와 멈춘 줄 앞뒤 12줄만 남기고, 그 사이는 "N~M번 줄 생략"으로 표시합니다([`internal/source/window.go`](internal/source/window.go)).

같은 SQL은 두 번 실행됩니다. 세션 스크립트가 같은 SQL로 먼저 카탈로그 캐시를 데운 뒤에 gdb를 붙이기 때문입니다. 스크립트의 머리 주석은 그 목적을 테이블 정의 조회가 기록에 섞이지 않게 하는 것이라고 적습니다(데우지 않은 기록과 나란히 비교한 적은 없습니다). 화면에 나오는 기록은 두 번째 실행의 것이고, 실행 양식과 기록 제목 아래에 이 사실을 적어 둡니다. INSERT를 실행하면 행이 두 개 들어간다는 뜻이기도 합니다.

## 재생본 열기와 만들기

재생본은 저장소를 받은 뒤 파일을 브라우저로 열면 됩니다. GitHub 화면에서 열면 HTML 원문이 보이므로, 내려받은 파일을 직접 엽니다. 재생본은 화면 코드와 실행 자료와 필요한 소스 줄을 모두 파일 안에 담고, 외부에서 아무것도 불러오지 않습니다. 주소 끝에 `#step-18`처럼 단계 번호를 붙이면 그 단계를 연 채로 시작합니다. 오른쪽 화살표나 `j`로 다음 단계, 왼쪽 화살표나 `k`로 앞 단계로 가고, `Home`과 `End`로 처음과 끝으로 갑니다.

새 기록으로 재생본을 만들려면 아래처럼 실행합니다. 소스는 실습 컨테이너가 아니라 이미지에서 직접 읽으므로(`docker run --rm --network none pg-internals:18.6 cat ...`) 컨테이너가 꺼져 있어도 되지만, 이미지는 있어야 합니다.

```bash
cd open_source_internals/workbench && go run ./cmd/workbench export -subject ../database/postgresql/workbench.json -trace ../database/postgresql/evidence/select-where-id-2.trace -watch select -out ../database/postgresql/replays/select-where-id-2.html -guide '../README.md#질문-1-select-한-줄이-지나는-경로'
```

성공하면 `재생본 ...: 단계 27개, 소스 파일 20개, 소스 줄 1576개, 110660바이트`처럼 크기를 알립니다. `-watch`는 그 기록을 만들 때 쓴 감시 목록의 id이고, `STOP` 줄의 값을 식마다 나누는 데 씁니다. `-sql`을 주지 않으면 첫 단계의 `query_string` 값에서 SQL을 꺼내고, 꺼내지 못하면 멈춥니다. 기록의 형식이 맞지 않으면 줄 번호를 밝힌 오류로, `STOP` 줄이 하나도 없으면 그 사실을 밝힌 오류로 멈춥니다.

재생본에 담긴 소스 줄이 원본과 같은지는 `check`로 확인합니다. 재생본을 만들 때는 파일 전체를 `cat`으로 읽어 줄을 나눴으므로, `check`는 다른 방법인 `sed -n '줄번호{=;p;}'`로 그 줄만 다시 읽어 비교합니다. 두 경로가 같은 방식으로 줄 번호를 잘못 세는 실수를 함께 숨기지 않게 하기 위해서입니다.

```bash
cd open_source_internals/workbench && go run ./cmd/workbench check -subject ../database/postgresql/workbench.json -replay ../database/postgresql/replays/select-where-id-2.html
```

PASS는 `소스 대조: 파일 20개, 줄 1576개, 다른 줄 0개`와 종료 코드 0입니다. 한 줄이라도 다르면 `src/backend/tcop/postgres.c:1014`처럼 다른 줄의 목록을 보이고 종료 코드 1로 끝납니다. 재생본의 소스 한 글자를 일부러 바꿔 이 실패가 실제로 나는 것을 확인했습니다.

## 실시간 실행

실시간 실행에는 실습 README의 [준비 절](../database/postgresql/README.md#준비-이미지와-실습-컨테이너)에서 만든 이미지 `pg-internals:18.6`과, 그 이미지로 띄운 컨테이너 `pg-lab`이 필요합니다. 저장소 루트에서 컨테이너를 띄우고, 작업대 서버를 켭니다.

```bash
sh open_source_internals/database/postgresql/scripts/start-lab.sh
```

```bash
cd open_source_internals/workbench && go run ./cmd/workbench serve -subject ../database/postgresql/workbench.json
```

서버가 뜨면 `작업대: http://127.0.0.1:8742/ (끝내려면 Ctrl+C)`가 나옵니다. 이 주소를 브라우저로 열고, 감시 목록을 고르고, SQL을 입력하고, "실행하고 추적"을 누릅니다. 2026-10-12에 이 컴퓨터에서 `SELECT * FROM t WHERE id = 2;`는 약 3초 만에 27단계로 돌아왔고, 같은 SQL을 `run-trace.sh`로 직접 실행한 기록과 함수, 위치, 멈춘 횟수가 모두 같았습니다. INSERT 감시 목록의 예시는 12단계, 인덱스 감시 목록은 24단계와 호출 스택 25칸이 모두 같았습니다.

감시 목록은 셋입니다. SELECT 경로, 인덱스 스캔 경로(실습 3), INSERT와 커밋 경로입니다. 인덱스 스캔 경로는 실습 3과 같은 조건을 만들려고 순차 스캔을 끈 세션(`-c enable_seqscan=off`)에서 돕니다. 행 3개짜리 테이블에서는 인덱스가 있어도 계획기가 순차 스캔을 고르기 때문입니다(2026-10-12에 인덱스를 만든 뒤 설정 없이 `EXPLAIN`하자 `Seq Scan on t`가 나왔습니다). 이 설정은 학습자가 바꾸는 값이 아니라 `workbench.json`의 감시 목록이 정하는 값(`pgOptions`)입니다. 인덱스는 작업대가 만들지 않으므로, 실습 3의 `CREATE INDEX t_id_idx ON t(id);` 명령을 터미널에서 먼저 실행합니다. 인덱스 없이 실행하면 순차 스캔 경로가 기록됩니다.

옵션은 셋입니다. `-subject`는 주제 정의 파일입니다. `-addr`의 기본값은 `127.0.0.1:8742`이고, 호스트는 `127.0.0.1`만 받습니다(포트 0이면 빈 포트를 고릅니다). `-timeout`은 추적 한 번의 시간 한도이고 기본값은 90초입니다. SQL 한 문장은 서버가 정한 `statement_timeout` 30초를 넘으면 취소됩니다. 다만 이 한도는 문장마다 따로 걸리고 세션 준비와 gdb를 붙이는 시간은 세지 않으므로, 두 번 실행하는 긴 SQL은 90초에 닿을 수 있습니다. 시간 초과 처리를 실제로 시험할 때 `-timeout 10s`처럼 줄입니다.

실시간 실행에서 알아 둘 점이 둘 있습니다.

1. SQL이 두 번 실행되므로 INSERT나 UPDATE는 테이블을 두 번 바꿉니다.

    실습 README의 숫자(예: `heap_getnextslot` 4번)는 행 3개짜리 새 테이블을 전제합니다. 행을 더한 뒤에 SELECT를 추적하면 행 수만큼 횟수가 달라집니다. 같은 숫자를 다시 보려면 `docker rm -f pg-lab` 뒤에 `start-lab.sh`로 다시 띄웁니다.

2. 터미널에서 `run-trace.sh`가 도는 동안에는 작업대가 실행을 거부합니다. 반대로 작업대가 도는 동안 터미널에서 시작한 실행은 막지 못합니다.

    세션 스크립트는 컨테이너 안의 고정된 `/tmp` 파일을 쓰므로 두 추적이 동시에 돌면 서로의 표시 파일을 읽게 됩니다. 작업대는 실행 전에 컨테이너 안에 gdb나 세션 psql이 있는지 보고, 있으면 터미널 쪽 실행을 건드리지 않고 409로 거부합니다. 터미널의 스크립트에는 이런 확인이 없습니다. 또 작업대 실행이 실패해 정리할 때는 컨테이너 안의 gdb와 세션 psql을 누가 시작했는지 가리지 않고 끝냅니다.

## 실행 한 번이 지나는 길

브라우저에서 "실행하고 추적"을 누르면 아래 순서로 일이 일어납니다.

1. 화면 코드([`web/app.js`](web/app.js))가 `POST /api/runs`에 `{"subject", "watch", "sql"}`을 보냅니다. 머리글 `X-Workbench: 1`을 붙입니다.
2. 서버([`internal/server`](internal/server/server.go))가 요청의 출처를 검사합니다. 아래 [안전 경계](#안전-경계)의 검사를 통과하지 못하면 여기서 403으로 끝납니다.
3. 실행기([`internal/run`](internal/run/run.go))가 SQL의 꼴을 확인하고, 이 서버의 다른 실행이 진행 중이면 409로 거부합니다. 컨테이너가 켜져 있는지, 컨테이너 안에 다른 추적이 없는지 확인합니다.
4. 실행기가 `sh run-trace.sh <감시 목록> <SQL> <호출 스택을 남길 함수>`를 최소 환경 변수로 실행합니다. 감시 목록과 함수는 학습자 입력이 아니라 `workbench.json`에서 고릅니다. 세션 설정 `PG_OPTIONS`는 서버가 정한 `statement_timeout` 30초에 감시 목록의 `pgOptions`를 더한 값입니다.
5. 스크립트가 끝나면 기록 해석기([`internal/trace`](internal/trace/trace.go))가 `STOP`, `BT`, `END`, `NOBREAK` 줄을 단계, 호출 스택, 횟수로 바꿉니다. 모르는 꼴의 줄이 있으면 오류로 멈춥니다.
6. 묶음기([`internal/replay`](internal/replay/bundle.go))가 단계마다 소스 구간을 계산합니다. 소스 파일은 컨테이너에서 통째로 읽고(`docker exec pg-lab cat <파일>`, 한 번 읽은 파일은 기억), 그 구간의 줄만 실행 자료 하나에 담습니다. 실행 자료(bundle)는 단계, 분류, 횟수, psql 출력, 화면에 필요한 소스 줄을 담은 JSON입니다.
7. 서버가 실행 자료를 돌려주고, 화면이 그것을 그립니다. 재생본은 이 실행 자료를 HTML 안에 넣은 것이므로, 실시간 화면과 재생본은 같은 그리기 코드를 씁니다.

폴더는 이 순서를 따라 나뉩니다.

| 위치 | 맡는 일 |
| --- | --- |
| `cmd/workbench/` | `serve`, `export`, `check` 명령과 옵션 검사 |
| `internal/server/` | 출처 검사, `GET /api/subjects`, `POST /api/runs`, 화면 파일 제공 |
| `internal/run/` | SQL 꼴 검사, 한 번에 한 실행, 컨테이너 안 추적 확인, 스크립트 실행, 실패 뒤 정리 |
| `internal/trace/` | 기록 줄 해석, gdb 문자열 값 해석 |
| `internal/source/` | 소스 경로 검사, Docker에서 파일 읽기, 함수 머리부터의 구간 계산 |
| `internal/subject/` | `workbench.json` 읽기와 디렉터리 분류 |
| `internal/replay/` | 기록과 소스 구간을 실행 자료로 묶기 |
| `web/` | 화면(HTML, CSS, JS, 외부 의존 없음)과 재생본 만들기 |

주제 정의 [`workbench.json`](../database/postgresql/workbench.json)에는 컨테이너와 이미지 이름, 이미지 안 소스 위치, 실행 스크립트, 감시 목록과 예시 SQL과 감시 목록별 세션 설정, 디렉터리별 분류 이름, 소스 저작권 문구가 있습니다. `pgOptions`는 `-c 이름=값`을 공백으로 이은 꼴만 받고, 다른 꼴이면 서버가 뜨지 않습니다. 새 주제를 더할 때는 이 파일과 그 주제의 추적 스크립트를 만들고, 기록 형식을 `trace.py`와 같게 맞춥니다.

## 안전 경계

실습의 psql 사용자 `postgres`는 슈퍼유저이므로, `COPY ... TO PROGRAM`처럼 SQL만으로도 컨테이너 안에서 셸 명령을 실행할 수 있습니다. 그래서 SQL을 걸러 내는 것으로는 보안을 지킬 수 없습니다. 작업대의 보안 경계는 두 가지입니다.

첫째는 컨테이너 격리입니다. `start-lab.sh`는 컨테이너를 `--network none`으로 띄우고, 디버거에 필요한 `SYS_PTRACE` 권한만 더하고, 호스트에서는 `scripts` 폴더만 읽기 전용으로 연결합니다. SQL이 컨테이너 안에서 무엇을 하든 네트워크로 나가거나 호스트의 파일을 쓸 수 없습니다. 컨테이너 안의 PostgreSQL은 `postgres` 사용자로 돌고, 이 사용자가 쓸 수 있는 소스 트리 안의 항목은 0개였습니다(`docker run --rm --network none -u postgres pg-internals:18.6 find /src -writable`의 출력 줄 수). 그래서 SQL로 화면에 보일 소스를 바꿀 수도 없습니다.

둘째는 다른 웹사이트가 학습자의 브라우저를 통해 보내는 요청의 차단입니다. 서버는 `127.0.0.1`에만 열리므로 다른 컴퓨터는 접속할 수 없지만, 학습자가 연 다른 사이트의 스크립트는 같은 컴퓨터의 `127.0.0.1:8742`로 요청을 보낼 수 있습니다. 서버는 아래 검사로 그런 요청을 막습니다. 마지막 열은 실제 서버에 curl이나 브라우저로 보내 거부되는 것을 확인한 요청이고, 단위 시험으로만 확인한 줄은 그렇게 적었습니다.

| 검사 | 막는 것 | 확인한 반례와 응답 |
| --- | --- | --- |
| 모든 요청의 `Host`가 `127.0.0.1:포트`나 `localhost:포트`여야 함 | DNS 재바인딩: 다른 사이트가 자기 도메인 이름을 127.0.0.1로 바꿔 같은 출처처럼 행세하는 공격 | `Host: evil.example:8742` → 403 |
| API 요청에 `X-Workbench: 1` 머리글이 있어야 함 | 다른 사이트의 폼 제출이나 단순 요청. 브라우저는 다른 출처로 이런 머리글을 붙이기 전에 사전 확인(preflight)을 하는데, 서버가 어떤 응답에도 CORS 허용 머리글을 붙이지 않으므로 사전 확인은 언제나 실패합니다. | 머리글 없는 POST, `text/plain` POST → 403. 다른 출처(`127.0.0.1:8799`) 페이지의 `fetch`는 브라우저가 "No 'Access-Control-Allow-Origin' header"로 막음 |
| `Origin`이 있으면 이 서버 주소와 같아야 함 | 사전 확인을 거치지 않는 경로로 온 다른 출처 요청 | `Origin: http://evil.example` → 403 |
| `Sec-Fetch-Site`가 있으면 `same-origin`이어야 함 | 같은 이유 | `cross-site` → 403 |
| 경로를 정리한 뒤 검사 | `/x/../api/runs`처럼 돌아 들어오는 경로 | 단위 시험: 머리글 없는 POST와 GET → 403 |
| 화면은 다른 사이트의 틀(iframe) 안에 들어가지 않음 | 작업대를 투명하게 덮어 실행 단추를 누르게 하는 공격 | 단위 시험: 모든 응답에 `X-Frame-Options: DENY`와 `frame-ancestors 'none'`이 있음(틀에 넣는 시험은 하지 않음) |

SQL 검사는 보안 장치가 아니라 세션 형식을 지키는 장치입니다. 세션 스크립트는 SQL을 psql 스크립트의 한 줄로 쓰고 그다음 줄에 psql 명령을 이어 쓰므로, psql이 그 줄 끝의 `;`를 문장 끝으로 보지 못하면 다음 줄들이 SQL에 이어 붙습니다. 실제로 `;`를 뺀 `SELECT * FROM t WHERE id = 2`를 직접 실행하자 `query_string`이 같은 SQL 두 줄을 이은 한 문장이 되어 구문 오류로 끝났고, 캐시를 데우는 실행은 일어나지 않았습니다. 그래서 아래를 400으로 거부합니다.

- 끝에 `;`가 없는 SQL, 끝이 아닌 자리의 `;`(문장이 둘이면 추적이 첫 문장에서 끝남). 따옴표 안의 `;`도 거부하는데, SQL을 해석하지 않고 판단하기 위한 의도된 제한입니다.
- 닫히지 않은 따옴표, 짝이 맞지 않는 괄호, 주석(`--`, `/*`), 따옴표 밖의 `$`(달러 따옴표). psql은 이것들 안의 `;`를 문장 끝으로 보지 않습니다.
- `COPY ... FROM STDIN`. psql이 다음 줄들을 복사할 자료로 읽습니다.
- 백슬래시(psql 메타 명령을 막으려고 따옴표 안이든 밖이든 모두 거부), 줄바꿈과 제어 문자, 빈 입력, 1000바이트 초과. 감시 목록은 이름으로만 고르므로 `"watch":"../../../etc/passwd"`도 400입니다.

## 실패 응답 읽기

| 응답 | 뜻 | 할 일 |
| --- | --- | --- |
| 400 | SQL이나 요청의 꼴이 맞지 않습니다. 이유가 화면에 나옵니다. | 문구대로 SQL을 고칩니다. |
| 403 | 요청이 작업대 화면에서 온 것으로 보이지 않습니다. | 서버가 알려 준 주소(`127.0.0.1:8742` 또는 `localhost:8742`)로 화면을 다시 엽니다. |
| 409, "다른 실행이 진행 중" | 이 서버의 다른 추적이 진행 중입니다. | 끝난 뒤 다시 실행합니다. |
| 409, "다른 추적(gdb나 세션 psql)이 진행 중" | 컨테이너 안에서 이 서버가 시작하지 않은 추적이 돕니다. 대개 터미널에서 직접 실행한 `run-trace.sh`입니다. | 그 실행이 끝난 뒤 다시 실행합니다. 직접 실행한 것이 없는데 계속 나오면 `docker rm -f pg-lab` 뒤 `start-lab.sh`로 다시 띄웁니다. |
| 415 | 본문이 JSON이 아닙니다. | 화면에서 보낸 요청이면 나오지 않습니다. |
| 503 | 실습 컨테이너가 꺼져 있습니다. 화면에 켜는 명령이 나옵니다. | 그 명령을 저장소 루트에서 실행한 뒤 다시 실행합니다. |
| 504 | 추적이 시간 한도 안에 끝나지 않았습니다. 작업대는 컨테이너 안에 남은 gdb와 psql을 정리했습니다. | 더 짧은 SQL로 다시 실행합니다. 다음 실행은 바로 받습니다. |
| 500 | 스크립트가 실패했거나 기록 형식이 예상과 다릅니다. 스크립트의 표준 오류 끝부분이 함께 나옵니다. | 그 문구를 보고 실습 README의 같은 명령을 터미널에서 직접 실행해 봅니다. |

기록에 멈춘 단계가 하나도 없거나, gdb가 중단점을 만들지 못한 함수(`NOBREAK`)가 있으면 실패가 아니라 화면 위쪽의 알림으로 보여 줍니다. 단계가 없으면 앞선 실행의 소스 창이 이번 기록처럼 남지 않도록 소스 창을 숨깁니다.

## 추적을 끊을 때의 함정 네 가지

추적이 시간 한도를 넘으면 작업대는 스크립트를 끝내고 컨테이너 안에 남은 gdb와 psql을 정리합니다. 이 정리를 처음에는 `pkill`로 단순하게 만들었는데, 실제 시간 초과 시험에서 PostgreSQL 서버가 두 가지 이유로 세 번 모든 연결을 끊고 복구를 거쳐 다시 시작했습니다. 그 두 이유와, 그 전후에 찾은 두 문제는 디버거나 컨테이너를 다루는 다른 도구에서도 나타날 수 있습니다.

### 1. 디버거를 거칠게 끝내면 대상 프로그램이 죽는다

gdb의 소프트웨어 중단점은 대상 프로그램의 기계어를 고쳐 쓰는 방식으로 동작합니다. 중단점 자리의 명령을 트랩을 일으키는 명령으로 바꿔 두면, 프로그램이 그 자리에 오는 순간 커널이 SIGTRAP을 보내고, ptrace로 붙어 있던 gdb가 그 신호를 받아 멈춘 것으로 처리합니다. 앞의 두 문장은 디버거의 일반적인 동작 방식을 옮긴 배경 설명이고, 이 실습에서 백엔드의 메모리를 직접 들여다보지는 않았습니다. 이 컨테이너의 gdb 13.1은 중단점을 언제 써 넣는지를 스스로 이렇게 밝힙니다(`help set breakpoint always-inserted`). 기본값(off)에서는 "실행이 계속될 때만 중단점을 써 넣고, 실행이 멈추면 빼냅니다".

따라서 백엔드가 달리는 동안 gdb를 SIGTERM으로 끝내면, 트랩 명령이 백엔드 메모리에 남은 채로 디버거만 사라질 수 있습니다(위 두 사실에서 끌어낸 추론). 실제로 그렇게 끝낸 시험에서 서버 기록은 `client backend (PID 360) was terminated by signal 5: Trace/breakpoint trap`에 이어 `terminating any other active server processes`와 `automatic recovery in progress`를 남겼습니다. 18.6의 `postmaster.c` 머리 주석은 개별 백엔드가 죽으면 postmaster가 공유 메모리를 다시 초기화해 복구한다고 설명합니다. 그 처리를 맡은 `HandleChildCrash`가 남은 자식 프로세스 모두에게 끝내라는 신호를 보내므로, 백엔드 하나가 죽으면 다른 연결까지 모두 끊깁니다.

작업대는 gdb에 SIGINT를 먼저 보냅니다. gdb는 SIGINT를 받으면 대상 프로그램을 멈추고, 위 기본값에 따라 멈출 때 중단점을 빼냅니다. 멈춤과 함께 `trace.py`의 `continue`가 돌아오고, 스크립트는 `END`를 찍은 뒤 `detach`로 분리합니다. 백엔드에게 SIGINT는 쿼리 취소 신호이므로 그 신호가 백엔드에 전달되는지도 따져야 합니다. gdb의 신호 표(`info signals SIGINT`)는 대상 프로그램이 받은 SIGINT를 `Pass to program No`로 다루지만, gdb가 백엔드를 멈출 때 정확히 어떤 신호를 쓰는지는 확인하지 않았습니다. 대신 결과를 관측했습니다. 백엔드가 `pg_sleep(6)`을 실행하는 중에 SIGINT를 보낸 시험에서, 0.5초 안에 gdb가 사라졌고, 기록은 `STOP 12` 뒤의 `END`로 끝났고, 백엔드는 살아 있었습니다. psql 출력에는 쿼리 취소 오류가 없었고(두 번의 `pg_sleep`이 모두 빈 결과를 돌려줌) 서버 기록에는 아무 줄도 남지 않았습니다. 5초 안에 gdb가 끝나지 않을 때만 마지막 수단으로 SIGKILL을 보냅니다.

### 2. 컨테이너에서는 데이터베이스가 고아 프로세스를 거둔다

리눅스에서 부모를 잃은 프로세스는 부모가 속한 PID 네임스페이스의 init 프로세스, 즉 그 네임스페이스의 1번 프로세스에게 넘어갑니다(`pid_namespaces(7)`). 이 실습 컨테이너의 1번은 postmaster(`postgres -D /var/lib/pgsql/data`)입니다. `docker exec`로 띄운 프로세스는 부모가 네임스페이스 밖의 Docker 쪽에 있어서 부모 번호가 0으로 보입니다. 시험에서 `docker exec -d`로 띄운 `sh -c 'psql ...'`의 부모 번호는 0이었고, psql의 부모는 그 sh였습니다. 그 sh를 죽이자 psql의 부모 번호가 1로 바뀌었습니다.

postmaster는 자기가 만들지 않은 자식이 끝나도 그 종료 상태를 봅니다. 18.6의 `postmaster.c` 2527~2533행은 종료 코드가 0이나 1이 아니면(신호로 죽은 경우도 여기에 듭니다) 백엔드가 죽었을 때와 같은 `HandleChildCrash`를 부릅니다. 종료 코드 0으로 끝나면 `untracked child process ... exited with exit code 0`이라는 기록만 남깁니다. 처음 정리 명령의 `pkill -f '[p]sql -X -At -f /tmp/session.sql'`는 psql뿐 아니라 그 명령줄을 품은 sh까지 끝냈습니다. 그래서 sh가 먼저 죽어 psql이 postmaster의 자식이 됐고, 이어서 SIGTERM으로 죽은 psql 때문에 서버가 다시 시작했습니다(`untracked child process (PID 769) was terminated by signal 15`).

작업대는 psql만 정확히 골라(`pkill -f '^psql -X -At -f /tmp/session\.sql$'`) 감싼 sh가 살아서 거두게 합니다. 또 psql을 끝내기 전에 `/tmp/ready`를 만들어, psql이 `\!`로 띄운 대기 반복이 부모를 잃더라도 신호 없이 스스로 끝나게 합니다. `docker run --init`으로 컨테이너의 1번 자리에 작은 init 프로그램을 두면, 같은 규칙에 따라 고아는 postmaster가 아니라 그 프로그램에게 넘어갑니다. 다만 그렇게 하면 실습 스크립트와 그 스크립트로 남긴 기록의 전제가 바뀌므로, 이번에는 정리 순서로 막았습니다.

### 3. 자식이 출력 파이프를 붙잡으면 시간 한도가 지켜지지 않는다

Go의 `exec.CommandContext`는 시간 한도가 지나면 직접 띄운 프로세스(`sh`)만 끝냅니다. `run-trace.sh`의 `docker exec ... | grep`처럼 sh가 띄운 자식은 살아서 표준 출력을 계속 열어 둡니다. 표준 출력을 파일이 아닌 버퍼로 모으면 `Wait`는 그 복사가 끝날 때까지, 즉 파이프가 닫힐 때까지 기다립니다(`go doc os/exec.Cmd.Wait`). 실제 프로세스로 만든 시험에서 0.2초 한도가 3.0초 뒤에야 돌아왔고, 3.0초는 자식 `sleep 3`이 끝난 시각이었습니다. 작업대는 `cmd.WaitDelay`를 2초로 둡니다. 이 값이 있으면 시간 한도가 지나거나 프로세스가 끝난 것을 본 때부터 2초 뒤에 남은 파이프를 닫고 돌아옵니다. 같은 시험은 2.2초에 돌아옵니다. 정상 실행에서는 `run-trace.sh`가 파이프라인이 모두 끝난 뒤에야 끝나므로, 이 2초가 출력을 자르지 않습니다.

### 4. 실행 전에 정리하면 다른 사람의 실행을 끝낸다

처음 설계는 실행 전에도 정리했습니다. 앞선 실행이 남긴 psql이 다음 실행의 표시 파일에 함께 깨어나 출력을 섞는 것을 막기 위해서였습니다. 그런데 작업대의 잠금은 이 서버의 실행만 보므로, 터미널에서 직접 실행한 `run-trace.sh`를 실행 전 정리가 끝내 버렸습니다(2번에 적은 재시작 가운데 23:10의 것이 이 충돌에서 일어났습니다). 이제 작업대는 실행 전에는 확인만 하고 다른 추적이 있으면 409로 거부하며, 정리는 자기 실행이 실패하거나 시간 한도를 넘거나 취소됐을 때만 합니다.

이 네 가지를 고친 뒤, 10초 한도로 gdb가 붙은 뒤의 시간 초과와 캐시를 데우는 중의 시간 초과를 각각 실제로 일으켰습니다. 두 경우 모두 504로 끝났고, 컨테이너에 gdb와 세션 psql이 남지 않았고, 바로 뒤의 정상 실행이 27단계와 `2|b` 두 줄을 돌려줬고, 서버 기록에는 다시 시작이나 사고 줄이 없었습니다.

## 검증

코드 검증은 작업대 폴더에서 실행합니다.

```bash
cd open_source_internals/workbench && go vet ./... && go test -race ./... && golangci-lint run ./...
```

PASS는 세 명령이 모두 종료 코드 0이고 마지막 줄이 `0 issues.`인 것입니다. 시험은 기록 해석(실제 기록의 단계 27개와 호출 스택 29칸 재현), 소스 구간, 분류, SQL 검사, 시간 초과와 잠금과 정리 횟수, 실제 프로세스의 시간 한도, 출처 검사의 반례 표, 실패 응답의 상태 코드를 다룹니다. 출처 검사는 검사 하나씩을 지운 코드에서 시험이 실패하는 것을 확인했습니다.

문서 게이트는 [`.golangci.yml`](.golangci.yml)의 revive(`exported`, `package-comments`)와 godoclint(`require-doc`, `start-with-name`)입니다. revive의 `exported` 규칙은 공개 식별자만 보고 `package main`에서는 걸리지 않으므로, 비공개 함수와 `main` 패키지까지 문서를 요구하는 godoclint를 함께 켰습니다. 공개 식별자의 문서 주석은 `Parse 함수는`처럼 이름과 공백으로 시작해야 하고(revive), 비공개 식별자의 문서 주석도 이름으로 시작해야 합니다(godoclint의 `start-with-name`은 이름 뒤의 공백까지는 보지 않습니다). 이 폴더는 둘 다 이름 뒤에 공백을 둡니다. 이 게이트가 결함을 실제로 잡는지는 문서 주석을 지운 코드로 먼저 확인했습니다.

화면과 실시간 실행의 판정 기준과 그 결과는 [작업대 WORK 문서](../../docs/works/WORK_20261012_OSS_INTERNALS_WORKBENCH.md)에 있습니다.

## 한계와 다시 열 조건

- 주제는 PostgreSQL 하나이고, 시각화는 단계가 지나는 디렉터리 층의 경로 그림 하나입니다. 버퍼나 페이지 안의 상태처럼 주제마다 다른 시각화는 아직 없습니다. 실습 README의 질문 2와 3이 쓰는 페이지 조회나 strace 기록도 화면에 나오지 않습니다.
- 실시간 실행도 재생본처럼 소스 구간만 보여 줍니다. 함수 전체를 스크롤해 보고 싶다는 학습 필요가 생기면 파일 전체를 보내는 방식을 다시 엽니다.
- `NOBREAK`가 있는 실제 기록과 멈춘 단계가 없는 실제 기록은 만들지 않았습니다. 앞의 것은 설정에 없는 함수를 넣은 감시 목록이 필요하고, 뒤의 것은 지금 감시 목록으로는 생기지 않습니다. 모든 단순 쿼리가 `exec_simple_query`와 `ReadyForQuery`를 지나기 때문입니다. 두 경우의 화면 알림은 서버 응답을 바꿔 넣는 방식으로 확인했습니다.
- psql이 `\!` 대기 반복을 도는 순간에 시간 한도가 닿는 경우는 시각을 맞추기 어려워 실제로 일으키지 않았고, 위의 정리 순서로 막힌다고 추론했습니다.
- 실행 전 확인과 스크립트 시작 사이에 터미널 실행이 끼어드는 아주 짧은 틈은 막지 못합니다.
- SQL 꼴 검사는 SQL을 해석하지 않는 단순한 규칙이라, 따옴표 안의 `;`와 `stdin`, 주석, 달러 따옴표를 쓰는 올바른 SQL도 거부합니다. 이런 SQL은 실습 README의 명령처럼 터미널에서 실행합니다.
- 셸 스크립트와 Docker를 쓰므로 macOS와 Linux를 전제합니다.

## 근거

- 실행 기록: [`evidence/select-where-id-2.trace`](../database/postgresql/evidence/select-where-id-2.trace), 기록 형식은 [`scripts/trace.py`](../database/postgresql/scripts/trace.py)의 머리 주석
- gdb 13.1(실습 컨테이너 안)의 `help set breakpoint always-inserted`, `info signals SIGINT` 출력
- PostgreSQL 18.6 소스 `src/backend/postmaster/postmaster.c`: 머리 주석(개별 백엔드가 죽은 뒤의 복구), `HandleChildCrash`, 2527~2533행(모르는 자식 프로세스의 종료 처리)
- Linux `pid_namespaces(7)`(<https://man7.org/linux/man-pages/man7/pid_namespaces.7.html>): 고아 프로세스가 넘어가는 곳, 네임스페이스 밖 부모의 번호가 0으로 보이는 것
- Go 1.27 `os/exec`의 `Cmd.Wait`와 `Cmd.WaitDelay` 문서(`go doc -all os/exec`, <https://pkg.go.dev/os/exec#Cmd>)
- 판정 기준, 시험 기록, 설계 변경: [작업대 WORK 문서](../../docs/works/WORK_20261012_OSS_INTERNALS_WORKBENCH.md)
