# 기존 기록 재고 조사: E-03 · E-04 · E-05

## 목차

- 0. 이 문서가 답하는 질문과 읽는 법
- 1. 한눈에 보기
- 2. E-03 Spring 해석 결과 14개(= concept-9의 RUNS)
- 3. E-04 화면에 박힌 해석 결과(concept-3, 4, 5~8, 9)
- 4. E-05 gdb 기록 8개
- 5. 공통 형식 설계로 넘길 함정(입력을 가로지르는 것)
- 6. 검증 경로
- 7. 확인하지 않은 것

## 0. 이 문서가 답하는 질문과 읽는 법

이 문서는 공통 사건 형식(place, run, event, link, note 표와 `attrs` 키-값 확장 열)으로 옮길 기존 기록 세 묶음이 정확히 어디에 무엇을 얼마나 담고 있는지, 그리고 지금 화면이 그 가운데 어떤 사실을 보여 주는지를 적습니다. 변환 뒤 화면별 사실을 다시 뽑아 "차이 0, 빠진 사실 0"을 판정하려면, 비교 기준이 되는 키와 수와 화면 함수가 먼저 고정되어 있어야 하기 때문입니다.

이 문서는 읽기만 해서 만들었습니다. 저장소의 파일은 하나도 바꾸지 않았고, 데이터를 읽은 스크립트는 모두 `python3 -I`로 돌렸고, 그 스크립트는 이 문서 옆의 `t1-inventory/` 폴더에 있습니다(처음에는 작업용 임시 폴더에서 돌렸습니다).

문장마다 근거의 종류를 아래 네 가지로 나눕니다.

- **[셈]**: `t1/`의 스크립트로 센 값입니다. 괄호 안에 스크립트 이름을 적습니다. 6절의 명령을 다시 돌리면 같은 수가 나와야 합니다.
- **[코드]**: `파일:행`에서 읽은 의미입니다.
- **[추론]**: 읽은 것에서 이끌어 낸 판단입니다. 데이터나 코드가 직접 말하지 않습니다.
- **[확인 안 함]**: 이번에 열어 보지 않았거나 실행하지 않은 것입니다. 7절에 모았습니다.

키 트리의 표시는 두 열입니다. 역할 열은 공통 형식의 어느 표·열로 갈 후보인지, 출처 열은 그 값이 기록기에서 왔는지 해석기가 만들었는지를 말합니다.

| 역할 | 뜻 | 공통 형식의 후보 자리 |
| --- | --- | --- |
| 장소 | 어디·누구에서 일어났는가: 프로세스, 연결, 세션, 스레드, 함수, 소스 위치, 표, 행(ctid), 트랜잭션 번호 | place 표 |
| 사건 | 기록 하나(목록의 원소) 또는 그 종류(`kind`, `type`, `dir`) | event 표의 행과 kind 열 |
| 연결 | 다른 사건·장소를 가리키는 값: 부모, 자식, 순번 참조, 짝 | link 표 |
| 값 | 관찰한 값 | `attrs` |
| 출처 | 기록기, 시계 맞춤, 합치기·버리기 통계 | run 표 또는 note |
| 시각 | 시각 또는 길이. 단위와 기준은 각 입력의 "시각 체계" 절에 적습니다 | event의 시각 열 |
| 발췌 | 소스 코드 줄 | place에 붙는 부속 자료 |
| 이름 | 사람이 읽으라고 붙인 글(화면 문구, 분류 이름) | note 또는 표시용 사전 |
| 모음 | 값이 없는 그릇(객체·배열 자체) | 표가 아님 |

| 출처 | 뜻 |
| --- | --- |
| 관찰 | 기록기(함수 기록기, uprobe, strace, JFR, 연결 중계기, 관찰 접속, gdb)가 남긴 값을 해석기가 옮겼거나 형식만 바꾼 것 |
| 도출 | 해석기가 셈하거나, 짝을 맞추거나, 규칙으로 분류한 값. 다시 해석하면 달라질 수 있음 |
| 이름 | 사람이 정한 이름·문구. 기록과 무관하게 정해짐 |

키 트리는 `keytree.py`가 문서 전체를 훑어 같은 경로를 합친 결과이고, 표시는 `marks.py`가 명시 규칙으로 붙였습니다. 줄마다 마지막 열이 적용한 규칙 번호이므로, 표시에 이견이 있으면 그 규칙 하나를 고치면 됩니다. 표기법은 다음과 같습니다.

- `n=`: 그 경로가 나타난 횟수(입력 문서 전체 합)
- `obj:`, `list:`, `str:`, `int:`, `float:`, `bool:`, `null:`: 값의 타입별 횟수
- `len[min,max,sum]`: 배열 길이 또는 객체 키 수
- `{*}`: 키가 데이터마다 다른 지도(동적 키). 실제 키 목록은 각 절의 "동적 키" 표에 따로 적습니다
- `kids[]{2+}`: concept-4 커널 트리의 2단 이상 재귀를 한 줄로 합친 것(수는 합)

한계도 미리 밝힙니다. 역할 표시는 판단이고, 같은 키가 두 역할을 할 수 있습니다(예: `xid`는 트랜잭션이라는 장소이면서 잠금 사건의 값). 표시는 "공통 형식으로 옮길 때 첫 후보"로 읽어야 합니다.

myai의 required 기준과 개발 기본 기준은 이번 작업에서 읽지 않았습니다(hook이 `deferred_reference`로 넘김). 그 기준을 적용했다고 보고하지 않습니다.

## 1. 한눈에 보기

| 입력 | 위치 | 문서 | 기록기가 남긴 사건 행 | 해석기가 만든 행(따로 셈) | 기존 화면 |
| --- | --- | --- | --- | --- | --- |
| E-03 | `open_source_internals/database/postgresql/scripts/.lab-tmp/spring/decoded-r5/r5-*.json` | 14 | 3,702 | 호출 트리 마디 1,388, 갈림 1,008, 빌림 24 외 | `workbench/design/spring-view.tmpl.html` → `concept-9-spring.html` |
| E-04 concept-9 | `workbench/design/concept-9-spring.html` 129행 `RUNS` | 14 | E-03과 같음 | E-03과 같음 | 같은 파일 |
| E-04 concept-5~8 | `workbench/design/concept-{5-tx,6-locks,7-durability,8-serializable}.html` 88행 `RUNS` | 44 | 5,082 | 판정 검사 70, 힌트 설명 172, 구경꾼 110 외 | `tx-view.tmpl.html` |
| E-04 concept-3 | `workbench/design/concept-3.html` 103행 `DATA` | 2 | 이동 107행(합친 횟수로는 419) | 단계 14 | 같은 파일 |
| E-04 concept-4 | `workbench/design/concept-4.html` 151행 `QUERIES` | 11 | 이동 1,087행(합친 횟수로는 3,042), 커널 마디 558 | 단계 76 | 같은 파일 |
| E-05 | `open_source_internals/database/postgresql/evidence/*.trace` | 8 | STOP 164, BT 71 | 없음(재생본이 분류·구간을 붙임) | 재생본 1개(`replays/select-where-id-2.html`) + `database/postgresql/README.md`의 서술 |

"기록기가 남긴 사건 행"의 정의는 각 입력의 규모 절에 따로 적습니다. 정의가 입력마다 다르므로 이 열을 서로 더하면 안 됩니다. concept-3과 concept-4의 `update`·`select`는 같은 기록을 두 번 해석한 것입니다(3.6절). 이 둘을 따로 세면 같은 사건을 두 번 셉니다.

## 2. E-03 Spring 해석 결과 14개(= concept-9의 RUNS)

### 2.1 위치와 수

파일은 `open_source_internals/database/postgresql/scripts/.lab-tmp/spring/decoded-r5/`의 14개입니다. 크기는 337,277~411,883 B이고 합은 5,189,984 B입니다 [셈: `stats4.py`].

| 장면 | 실행 | 파일 |
| --- | --- | --- |
| 5-nested | a, b | `r5-nested-{a,b}.json` |
| 5-reborrow | a, b | `r5-reborrow-{a,b}.json` |
| 5-required | a, b | `r5-required-{a,b}.json` |
| 5-requires-new | a, b | `r5-requires-new-{a,b}.json` |
| 5-requires-new-pool1 | a, b, c, d | `r5-requires-new-pool1-{a,b,c,d}.json` |
| 5-requires-new-same-row | a, b | `r5-requires-new-same-row-{a,b}.json` |

14개 모두 concept-9의 `RUNS["5-<장면>/<실행>"]`과 내용이 같습니다 [셈: `cmp9.py`, 14개 IDENTICAL]. 바이트 합(5,189,984)과 concept-9 literal 길이(5,190,265)가 다른 것은 직렬화 차이이고 값 차이가 아닙니다. 따라서 E-03과 concept-9는 한 번만 변환하고 한 번만 셉니다.

출력의 맨 위 키 집합은 `decode/spring.go` 654~676행의 `map[string]any`가 정합니다 [코드]: `scene, pool, javaVersion, javaExit, title, tables, sessions, checkpoints, blocked, notes, conns, transactions, borrows, captures, clock, joins, siteRepair, java, view`, 그리고 소스 폴더가 주어졌을 때만 `src`. map이라 JSON에서 키는 이름순으로 나옵니다. `src`의 발췌는 서버 사건(`server.events[].stack[].at`)이 가리키는 줄만 모은 것이고, 스냅숏·줄 판정의 스택이 가리키는 줄은 따로 모으지 않습니다(같은 행 범위) [코드]. `view`의 `branches`·`compared`는 `decode spring-compare`가 나중에 덧붙입니다(`decode/spring_compare.go` 185~239행) [코드].

실행별 수입니다 [셈: `stats.py`].

| 실행 | 메시지 | 요청 | 서버 사건 | 스냅숏 | 줄 판정 | JFR 구간 | 예외 | 트리 마디 | 효과 | 관찰 | 빌림 | 짝 맞춤 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 5-nested/a | 126 | 14 | 30 | 6 | 2 | 108 | 4 | 90 | 14 | 6 | 1 | 42 |
| 5-nested/b | 126 | 14 | 30 | 6 | 2 | 108 | 5 | 90 | 14 | 6 | 1 | 42 |
| 5-reborrow/a | 109 | 13 | 32 | 6 | 3 | 130 | 2 | 111 | 14 | 3 | 2 | 41 |
| 5-reborrow/b | 109 | 13 | 32 | 6 | 3 | 130 | 2 | 111 | 14 | 3 | 2 | 41 |
| 5-required/a | 100 | 11 | 25 | 6 | 2 | 96 | 5 | 82 | 13 | 6 | 1 | 41 |
| 5-required/b | 100 | 11 | 25 | 6 | 2 | 96 | 5 | 82 | 13 | 6 | 1 | 41 |
| 5-requires-new-pool1/a | 64 | 8 | 18 | 3 | 1 | 93 | 6 | 84 | 15 | 5 | 2 | 27 |
| 5-requires-new-pool1/b | 64 | 8 | 18 | 3 | 1 | 93 | 6 | 84 | 15 | 5 | 2 | 27 |
| 5-requires-new-pool1/c | 64 | 8 | 18 | 3 | 1 | 93 | 6 | 84 | 15 | 5 | 2 | 27 |
| 5-requires-new-pool1/d | 64 | 8 | 18 | 3 | 1 | 93 | 6 | 84 | 15 | 5 | 2 | 27 |
| 5-requires-new-same-row/a | 124 | 15 | 36 | 6 | 2 | 133 | 5 | 119 | 23 | 6 | 2 | 52 |
| 5-requires-new-same-row/b | 124 | 15 | 36 | 6 | 2 | 133 | 5 | 119 | 23 | 6 | 2 | 52 |
| 5-requires-new/a | 114 | 14 | 33 | 6 | 2 | 143 | 3 | 124 | 18 | 6 | 2 | 51 |
| 5-requires-new/b | 114 | 14 | 33 | 6 | 2 | 143 | 2 | 124 | 18 | 6 | 2 | 51 |
| 합 | 1,402 | 166 | 384 | 72 | 26 | 1,592 | 62 | 1,388 | 224 | 74 | 24 | 562 |

그 밖의 합 [셈: `stats.py`]: 연결 24, 트랜잭션 20(`xids` 22, `clog` 22), 멈춤 자리(`checkpoints`) 68, 막힘 관찰(`blocked`) 6, 앱 쪽 오류 메모(`notes`) 12, JFR 잠듦(`parks`) 4, 중복(`dups`) 20, 스택에만 나온 메서드(`stackOnly`) 46, 연결 묶임 재생(`binding`) 70, 기다림 고리(`waitCycles`) 6, 서버 쪽 창이 붙은 요청 142, 카탈로그 스냅숏 수의 합 72.

### 2.2 시각 체계

모든 시각은 `spring.json`의 `startNs`(맥 시계)부터 잰 마이크로초입니다(`decode/spring.go` 21~22행, 407~412행) [코드]. 다만 단위·정밀도·시계가 키마다 다릅니다.

| 키 | 단위 | 시계 | JSON에 나온 꼴 [셈: keytree] | 근거 |
| --- | --- | --- | --- | --- |
| `conns[].messages[].t` | µs | 맥 쪽 중계기 | 1,402개 모두 소수점 없음 | [코드] spring.go 18행, 407~412행 |
| `conns[].messages[].tc` | µs | 컨테이너 쪽 중계기의 ns를 맥 `startNs`로 뺀 값 | float 1,402 | [코드] 같은 곳. 두 시계 차는 `clock.wire`로만 압니다 |
| `conns[].requests[].sent`, `done` | µs | 맥 쪽(요청 묶음의 첫 메시지, ReadyForQuery를 읽은 시각) | 숫자 | [코드] spring.go 108~110행 주석 |
| `...server.t0`, `t1`, `events[].t`, `t1`, `snapshots[].t`, `checks[].t` | µs | 컨테이너(함수 기록기·uprobe) | float | [추론] 컨테이너 쪽 기록기이므로 컨테이너 시계 |
| `java.spans[].t0`, `t1`, `throws[].t`, `parks[]` | µs | JFR | 대부분 소수, 일부 소수점 없음(spans t0 66, t1 59, throws 1) | [추론] 앱이 맥에서 돌므로 맥 시계. 확인 안 함 |
| `checkpoints[].t`, `tGo` | µs | 맥(`javaNs`, `goNs`) | int | [코드] txlab/spring.go `springResult` |
| `checkpoints[].tCapture`, `blocked[].tCapture`, `captures[].t`, `captures[].reads.*` | µs | 컨테이너(`captureNs`) | float | [코드] 같은 곳 |
| `clock.wire[].lo`, `hi` | **밀리초** | (컨테이너 − 맥) 차이의 구간 | float | [코드] spring.go 97~99행 주석 |
| `clock.uprobe.maxDeviationUs`, `medianOffsetUs` | µs | uprobe와 함수 기록기의 차 | float | 이름 |
| `borrows[].idleMs`, `effects[].idleMs`, `poolConfig.*Ms` | 밀리초 | 길이 | float | 이름 |
| `view.tree[].self` | µs | 길이(자기 시간) | float/int/null 70 | [코드] spring_tree.go `annotateTree` |

함정이 셋 있습니다. 첫째, `rel(ns)`는 `ns == 0`이면 0을 돌려줍니다(spring.go 408~411행) [코드]. 그래서 시각 0은 "기준 순간"과 "값 없음"을 가르지 못합니다. 둘째, 같은 순간을 가리키는 값이라도 맥 쪽 값은 소수점이 없고 컨테이너 쪽 값은 소수 µs입니다(예: 5-reborrow/a의 마지막 X 메시지 `t=4456776`, `tc=4479425.793`) [셈]. 맥 쪽 중계기의 ns가 1 µs 단위로만 찍히기 때문으로 보입니다 [추론]. 공통 형식의 시각 열 하나에 넣으려면 시계 열(맥, 컨테이너)과 정밀도를 함께 남겨야 합니다. 셋째, 키 트리의 `int`/`float` 구분은 값의 타입이 아닙니다. Go의 `encoding/json`은 소수 부분이 0인 float64를 소수점 없이 쓰므로(예: `view.tree[].t0`의 int 56), 같은 float64 칸이 JSON에서 int와 float로 갈라져 보입니다 [추론: 같은 칸에 두 꼴이 섞인 것으로 판단]. 변환기는 JSON의 정수/소수 꼴로 타입이나 시계를 판단하면 안 됩니다.

### 2.3 키 트리와 표시

489개 경로입니다 [셈: `keytree.py` → `marks.py e03`].

```text
$  n=14  obj:14 len[min=20,max=20,sum=280]  | 모음 | 관찰 | G9
$.blocked  n=14  list:14 len[min=0,max=1,sum=6]  | 모음 | 관찰 | G9
$.borrows  n=14  list:14 len[min=1,max=2,sum=24]  | 모음 | 도출 | E24+G9
$.borrows[]  n=24  obj:24 len[min=7,max=7,sum=168]  | 사건 | 도출 | E24+G8
$.borrows[].t0  n=24  int:2,float:22  | 시각 | 도출 | E24+G3
$.borrows[].t1  n=24  float:23,int:1  | 시각 | 도출 | E24+G3
$.borrows[].thread  n=24  str:24  | 장소 | 도출 | E24+G4
$.borrows[].validated  n=24  bool:24  | 값 | 도출 | E24+G9
$.borrows[].conn  n=24  str:24  | 연결 | 도출 | E23
$.borrows[].idleMs  n=24  float:20,int:4  | 시각 | 도출 | E24+G3
$.borrows[].result  n=24  str:24  | 값 | 도출 | E24+G9
$.captures  n=14  list:14 len[min=3,max=6,sum=74]  | 모음 | 관찰 | G9
$.captures[]  n=74  obj:74 len[min=13,max=13,sum=962]  | 사건 | 관찰 | G8
$.captures[].afterStep  n=74  int:74  | 연결 | 관찰 | E28
$.captures[].wal  n=74  obj:74 len[min=3,max=3,sum=222]  | 모음 | 관찰 | G9
$.captures[].wal.insert  n=74  str:74  | 값 | 관찰 | G9
$.captures[].wal.write  n=74  str:74  | 값 | 관찰 | G9
$.captures[].wal.flush  n=74  str:74  | 값 | 관찰 | G9
$.captures[].t  n=74  float:74  | 시각 | 관찰 | G3
$.captures[].pages  n=74  obj:74 len[min=1,max=1,sum=74]  | 모음 | 관찰 | G9
$.captures[].pages.acct  n=74  list:74 len[min=2,max=4,sum=236]  | 모음 | 관찰 | G9
$.captures[].pages.acct[]  n=236  obj:236 len[min=8,max=8,sum=1888]  | 값 | 관찰 | G5d
$.captures[].pages.acct[].block  n=236  int:236  | 장소 | 관찰 | G4
$.captures[].pages.acct[].lp  n=236  int:236  | 장소 | 관찰 | G4
$.captures[].pages.acct[].lpFlags  n=236  int:236  | 값 | 관찰 | G9
$.captures[].pages.acct[].ctid  n=236  str:236  | 장소 | 관찰 | G4
$.captures[].pages.acct[].xmin  n=236  str:236  | 값 | 관찰 | G9
$.captures[].pages.acct[].xmax  n=236  str:236  | 값 | 관찰 | G9
$.captures[].pages.acct[].flags  n=236  list:236 len[min=2,max=5,sum=740]  | 모음 | 관찰 | G9
$.captures[].pages.acct[].flags[]  n=740  str:740  | 값 | 관찰 | G8
$.captures[].pages.acct[].values  n=236  obj:236 len[min=3,max=3,sum=708]  | 모음 | 관찰 | G9
$.captures[].pages.acct[].values.balance  n=236  str:236  | 값 | 관찰 | G9
$.captures[].pages.acct[].values.id  n=236  str:236  | 장소 | 관찰 | G4
$.captures[].pages.acct[].values.owner  n=236  str:236  | 값 | 관찰 | G9
$.captures[].activity  n=74  list:60,null:14 len[min=1,max=2,sum=104]  | 모음 | 관찰 | G9
$.captures[].activity[]  n=104  obj:104 len[min=6,max=6,sum=624]  | 값 | 관찰 | G5d
$.captures[].activity[].pid  n=104  int:104  | 장소 | 관찰 | G4
$.captures[].activity[].xid  n=104  str:104  | 장소 | 관찰 | G4
$.captures[].activity[].xmin  n=104  str:104  | 값 | 관찰 | G9
$.captures[].activity[].state  n=104  str:104  | 값 | 관찰 | G9
$.captures[].activity[].waitType  n=104  str:104  | 값 | 관찰 | G9
$.captures[].activity[].waitEvent  n=104  str:104  | 값 | 관찰 | G9
$.captures[].locks  n=74  null:44,list:30 len[min=4,max=10,sum=144]  | 모음 | 관찰 | G9
$.captures[].running  n=74  str:74  | 값 | 관찰 | G9
$.captures[].xacts  n=74  obj:74 len[min=1,max=3,sum=156] (dynamic keys)  | 모음 | 관찰 | G9
$.captures[].xacts.{*}  n=156  str:156  | 값 | 관찰 | G9
$.captures[].reads  n=74  obj:74 len[min=5,max=5,sum=370]  | 시각 | 관찰 | E26
$.captures[].reads.activity  n=74  list:74 len[min=2,max=2,sum=148]  | 시각 | 관찰 | E26
$.captures[].reads.activity[]  n=148  int:148  | 시각 | 관찰 | E26
$.captures[].reads.locks  n=74  list:74 len[min=2,max=2,sum=148]  | 시각 | 관찰 | E26
$.captures[].reads.locks[]  n=148  int:148  | 시각 | 관찰 | E26
$.captures[].reads.pages  n=74  list:74 len[min=2,max=2,sum=148]  | 시각 | 관찰 | E26
$.captures[].reads.pages[]  n=148  int:148  | 시각 | 관찰 | E26
$.captures[].reads.running  n=74  list:74 len[min=2,max=2,sum=148]  | 시각 | 관찰 | E26
$.captures[].reads.running[]  n=148  int:148  | 시각 | 관찰 | E26
$.captures[].reads.xacts  n=74  list:74 len[min=2,max=2,sum=148]  | 시각 | 관찰 | E26
$.captures[].reads.xacts[]  n=148  int:94,float:54  | 시각 | 관찰 | E26
$.captures[].label  n=74  str:74  | 이름 | 관찰 | E27
$.captures[].span  n=74  list:74 len[min=2,max=2,sum=148]  | 시각 | 도출 | E25+G3
$.captures[].span[]  n=148  float:74,int:74  | 시각 | 도출 | E25+G3
$.captures[].overlap  n=74  list:74 len[min=0,max=8,sum=17]  | 모음 | 도출 | E25+G9
$.captures[].visible  n=74  obj:74 len[min=1,max=1,sum=74]  | 모음 | 도출 | E25+G9
$.captures[].visible.acct  n=74  list:74 len[min=2,max=2,sum=148]  | 모음 | 도출 | E25+G9
$.captures[].visible.acct[]  n=148  obj:148 len[min=4,max=4,sum=592]  | 값 | 도출 | E25+G5d
$.captures[].visible.acct[].balance  n=148  str:148  | 값 | 도출 | E25+G9
$.captures[].visible.acct[].ctid  n=148  str:148  | 장소 | 도출 | E25+G4
$.captures[].visible.acct[].id  n=148  str:148  | 장소 | 도출 | E25+G4
$.captures[].visible.acct[].owner  n=148  str:148  | 값 | 도출 | E25+G9
$.captures[].locks[]  n=144  obj:144 len[min=9,max=9,sum=1296]  | 값 | 관찰 | G5d
$.captures[].locks[].pid  n=144  int:144  | 장소 | 관찰 | G4
$.captures[].locks[].type  n=144  str:144  | 값 | 관찰 | G5b
$.captures[].locks[].relation  n=144  str:144  | 장소 | 관찰 | G4
$.captures[].locks[].page  n=144  str:144  | 장소 | 관찰 | G4
$.captures[].locks[].tuple  n=144  str:144  | 장소 | 관찰 | G4
$.captures[].locks[].xid  n=144  str:144  | 장소 | 관찰 | G4
$.captures[].locks[].vxid  n=144  str:144  | 장소 | 관찰 | G4
$.captures[].locks[].mode  n=144  str:144  | 값 | 관찰 | G9
$.captures[].locks[].granted  n=144  bool:144  | 값 | 관찰 | G9
$.checkpoints  n=14  list:14 len[min=3,max=6,sum=68]  | 모음 | 관찰 | G9
$.checkpoints[]  n=68  obj:68 len[min=5,max=5,sum=340]  | 사건 | 관찰 | G8
$.checkpoints[].label  n=68  str:68  | 이름 | 관찰 | E39
$.checkpoints[].n  n=68  int:68  | 값 | 관찰 | G9
$.checkpoints[].t  n=68  int:68  | 시각 | 관찰 | G3
$.checkpoints[].tCapture  n=68  float:68  | 시각 | 관찰 | G3
$.checkpoints[].tGo  n=68  int:68  | 시각 | 관찰 | G3
$.clock  n=14  obj:14 len[min=2,max=2,sum=28]  | 출처 | 도출 | E29
$.clock.uprobe  n=14  obj:14 len[min=4,max=4,sum=56]  | 출처 | 도출 | E29
$.clock.uprobe.events  n=14  int:14  | 출처 | 도출 | E29
$.clock.uprobe.joined  n=14  int:14  | 출처 | 도출 | E29
$.clock.uprobe.maxDeviationUs  n=14  float:14  | 출처 | 도출 | E29
$.clock.uprobe.medianOffsetUs  n=14  float:14  | 출처 | 도출 | E29
$.clock.wire  n=14  list:14 len[min=1,max=2,sum=24]  | 출처 | 도출 | E29
$.clock.wire[]  n=24  obj:24 len[min=5,max=5,sum=120]  | 출처 | 도출 | E29
$.clock.wire[].conn  n=24  int:24  | 출처 | 도출 | E29
$.clock.wire[].n  n=24  int:24  | 출처 | 도출 | E29
$.clock.wire[].same  n=24  bool:24  | 출처 | 도출 | E29
$.clock.wire[].lo  n=24  float:24  | 출처 | 도출 | E29
$.clock.wire[].hi  n=24  float:24  | 출처 | 도출 | E29
$.conns  n=14  list:14 len[min=1,max=2,sum=24]  | 모음 | 관찰 | G9
$.conns[]  n=24  obj:24 len[min=13,max=13,sum=312]  | 장소 | 관찰 | G5c
$.conns[].conn  n=24  int:24  | 장소 | 관찰 | G4
$.conns[].pid  n=24  int:24  | 장소 | 관찰 | G4
$.conns[].name  n=24  str:24  | 장소 | 관찰 | G4
$.conns[].thread  n=24  str:24  | 장소 | 관찰 | G4
$.conns[].startup  n=24  list:24 len[min=6,max=6,sum=144]  | 모음 | 관찰 | G9
$.conns[].startup[]  n=144  str:144  | 값 | 관찰 | G8
$.conns[].params  n=24  obj:24 len[min=6,max=6,sum=144]  | 모음 | 관찰 | G9
$.conns[].params.DateStyle  n=24  str:24  | 값 | 관찰 | G9
$.conns[].params.TimeZone  n=24  str:24  | 값 | 관찰 | G9
$.conns[].params.client_encoding  n=24  str:24  | 값 | 관찰 | G9
$.conns[].params.database  n=24  str:24  | 값 | 관찰 | G9
$.conns[].params.protocol  n=24  str:24  | 값 | 관찰 | G9
$.conns[].params.user  n=24  str:24  | 값 | 관찰 | G9
$.conns[].accept  n=24  int:24  | 시각 | 관찰 | G3
$.conns[].close  n=24  int:24  | 시각 | 관찰 | G3
$.conns[].lastMac  n=24  str:24  | 출처 | 관찰 | E37
$.conns[].lastCont  n=24  str:24  | 출처 | 관찰 | E37
$.conns[].messages  n=24  list:24 len[min=26,max=100,sum=1402]  | 모음 | 관찰 | G9
$.conns[].messages[]  n=1402  obj:1402 len[min=5,max=7,sum=7548]  | 사건 | 관찰 | G8
$.conns[].messages[].t  n=1402  int:1402  | 시각 | 관찰 | G3
$.conns[].messages[].tc  n=1402  float:1402  | 시각 | 관찰 | G3
$.conns[].messages[].dir  n=1402  str:1402  | 사건 | 관찰 | G6
$.conns[].messages[].type  n=1402  str:1402  | 사건 | 관찰 | G6
$.conns[].messages[].len  n=1402  int:1402  | 값 | 관찰 | G9
$.conns[].messages[].text  n=450  str:450  | 값 | 관찰 | G9
$.conns[].messages[].f  n=50  obj:50 len[min=1,max=3,sum=54]  | 모음 | 관찰 | G9
$.conns[].messages[].f.auth  n=24  str:24  | 값 | 관찰 | G9
$.conns[].messages[].f.pid  n=24  str:24  | 장소 | 관찰 | G4
$.conns[].messages[].name  n=38  str:38  | 장소 | 관찰 | G4
$.conns[].requests  n=24  list:24 len[min=2,max=12,sum=166]  | 모음 | 관찰 | G9
$.conns[].requests[]  n=166  obj:166 len[min=6,max=9,sum=1268]  | 사건 | 관찰 | G8
$.conns[].requests[].seq  n=166  int:166  | 연결 | 도출 | E35+G5
$.conns[].requests[].sql  n=166  list:164,null:2 len[min=1,max=1,sum=164]  | 모음 | 관찰 | G9
$.conns[].requests[].sql[]  n=164  str:164  | 값 | 관찰 | G8
$.conns[].requests[].types  n=166  list:166 len[min=0,max=5,sum=380]  | 모음 | 관찰 | G9
$.conns[].requests[].sent  n=166  int:166  | 시각 | 도출 | E35+G3
$.conns[].requests[].done  n=166  int:166  | 시각 | 도출 | E35+G3
$.conns[].requests[].status  n=166  str:166  | 값 | 관찰 | G9
$.conns[].requests[].types[]  n=380  str:380  | 값 | 관찰 | G8
$.conns[].requests[].tags  n=108  list:108 len[min=1,max=1,sum=108]  | 모음 | 관찰 | G9
$.conns[].requests[].tags[]  n=108  str:108  | 값 | 관찰 | G8
$.conns[].requests[].server  n=142  obj:142 len[min=9,max=9,sum=1278]  | 모음 | 관찰 | G9
$.conns[].requests[].server.sql  n=142  str:142  | 값 | 관찰 | G9
$.conns[].requests[].server.t0  n=142  float:142  | 시각 | 도출 | E34
$.conns[].requests[].server.t1  n=142  float:142  | 시각 | 도출 | E34
$.conns[].requests[].server.snapshots  n=142  list:142 len[min=0,max=3,sum=72]  | 모음 | 관찰 | G9
$.conns[].requests[].server.catalogSnapshots  n=142  int:142  | 값 | 관찰 | G9
$.conns[].requests[].server.checks  n=142  list:142 len[min=0,max=2,sum=26]  | 모음 | 관찰 | G9
$.conns[].requests[].server.events  n=142  list:142 len[min=0,max=16,sum=384]  | 모음 | 관찰 | G9
$.conns[].requests[].server.events[]  n=384  obj:384 len[min=2,max=5,sum=1584]  | 사건 | 관찰 | G8
$.conns[].requests[].server.events[].t  n=384  float:384  | 시각 | 관찰 | G3
$.conns[].requests[].server.events[].kind  n=384  str:384  | 사건 | 관찰 | G6
$.conns[].requests[].server.events[].f  n=380  obj:380 len[min=1,max=5,sum=978]  | 모음 | 관찰 | G9
$.conns[].requests[].server.events[].f.sxact  n=80  str:80  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].stack  n=334  list:334 len[min=6,max=30,sum=4660]  | 장소 | 관찰 | G2
$.conns[].requests[].server.events[].stack[]  n=4660  obj:4660 len[min=2,max=2,sum=9320]  | 장소 | 관찰 | G2
$.conns[].requests[].server.events[].stack[].fn  n=4660  str:4660  | 장소 | 관찰 | G2
$.conns[].requests[].server.events[].stack[].at  n=4660  str:4660  | 장소 | 관찰 | G2
$.conns[].requests[].server.in  n=142  null:142  | 모음 | 관찰 | G9
$.conns[].requests[].server.out  n=142  null:142  | 모음 | 관찰 | G9
$.conns[].requests[].server.snapshots[]  n=72  obj:72 len[min=7,max=7,sum=504]  | 사건 | 관찰 | G8
$.conns[].requests[].server.snapshots[].t  n=72  float:72  | 시각 | 관찰 | G3
$.conns[].requests[].server.snapshots[].kind  n=72  str:72  | 사건 | 관찰 | G6
$.conns[].requests[].server.snapshots[].xmin  n=72  int:72  | 값 | 관찰 | G9
$.conns[].requests[].server.snapshots[].xmax  n=72  int:72  | 값 | 관찰 | G9
$.conns[].requests[].server.snapshots[].xip  n=72  null:72  | 모음 | 관찰 | G9
$.conns[].requests[].server.snapshots[].curcid  n=72  int:72  | 값 | 관찰 | G9
$.conns[].requests[].server.snapshots[].stack  n=72  list:72 len[min=3,max=5,sum=264]  | 장소 | 관찰 | G2
$.conns[].requests[].server.snapshots[].stack[]  n=264  obj:264 len[min=2,max=2,sum=528]  | 장소 | 관찰 | G2
$.conns[].requests[].server.snapshots[].stack[].fn  n=264  str:264  | 장소 | 관찰 | G2
$.conns[].requests[].server.snapshots[].stack[].at  n=264  str:264  | 장소 | 관찰 | G2
$.conns[].requests[].server.checks[]  n=26  obj:26 len[min=13,max=13,sum=338]  | 사건 | 관찰 | G8
$.conns[].requests[].server.checks[].t  n=26  float:26  | 시각 | 관찰 | G3
$.conns[].requests[].server.checks[].table  n=26  str:26  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].ctid  n=26  str:26  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].xmin  n=26  int:26  | 값 | 관찰 | G9
$.conns[].requests[].server.checks[].xmax  n=26  int:26  | 값 | 관찰 | G9
$.conns[].requests[].server.checks[].mask  n=26  int:26  | 값 | 관찰 | G9
$.conns[].requests[].server.checks[].snapXmin  n=26  int:26  | 값 | 관찰 | G9
$.conns[].requests[].server.checks[].snapXmax  n=26  int:26  | 값 | 관찰 | G9
$.conns[].requests[].server.checks[].visible  n=26  bool:26  | 값 | 도출 | E33
$.conns[].requests[].server.checks[].tup  n=26  str:26  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].path  n=26  list:26 len[min=2,max=4,sum=78]  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].path[]  n=78  obj:78 len[min=3,max=5,sum=342]  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].path[].fn  n=78  str:78  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].path[].at  n=78  str:78  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].path[].arg  n=78  str:78  | 값 | 관찰 | G9
$.conns[].requests[].server.checks[].path[].ret  n=54  str:54  | 값 | 관찰 | G9
$.conns[].requests[].server.checks[].path[].inner  n=54  list:54 len[min=1,max=2,sum=56]  | 모음 | 관찰 | G9
$.conns[].requests[].server.checks[].path[].inner[]  n=56  obj:56 len[min=2,max=2,sum=112]  | 사건 | 관찰 | G8
$.conns[].requests[].server.checks[].path[].inner[].fn  n=56  str:56  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].path[].inner[].at  n=56  str:56  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].lines  n=26  list:26 len[min=9,max=13,sum=264]  | 모음 | 관찰 | G9
$.conns[].requests[].server.checks[].lines[]  n=264  obj:264 len[min=2,max=2,sum=528]  | 사건 | 관찰 | G8
$.conns[].requests[].server.checks[].lines[].fn  n=264  str:264  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].lines[].at  n=264  str:264  | 장소 | 관찰 | G4
$.conns[].requests[].server.checks[].stack  n=26  list:26 len[min=24,max=24,sum=624]  | 장소 | 관찰 | G2
$.conns[].requests[].server.checks[].stack[]  n=624  obj:624 len[min=2,max=2,sum=1248]  | 장소 | 관찰 | G2
$.conns[].requests[].server.checks[].stack[].fn  n=624  str:624  | 장소 | 관찰 | G2
$.conns[].requests[].server.checks[].stack[].at  n=624  str:624  | 장소 | 관찰 | G2
$.conns[].requests[].server.events[].f.mode  n=72  str:72  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.rel  n=120  str:120  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].f.result  n=98  str:98  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.type  n=72  str:72  | 값 | 관찰 | G5b
$.conns[].requests[].server.events[].t1  n=102  float:102  | 시각 | 관찰 | G3
$.conns[].requests[].server.events[].f.at  n=16  str:16  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].f.bit  n=14  str:14  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.tup  n=14  str:14  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].f.xid  n=82  str:82  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].f.buf  n=28  str:28  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].f.lsn  n=48  str:48  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.created  n=48  str:48  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.ctid  n=50  str:50  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].f.gran  n=48  str:48  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.txid  n=48  str:48  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].f.level  n=22  str:22  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.off  n=24  str:24  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].f.xmax  n=24  str:24  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.xmin  n=24  str:24  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.blk  n=22  str:22  | 장소 | 관찰 | G4
$.conns[].requests[].server.events[].f.nsub  n=22  str:22  | 값 | 관찰 | G9
$.conns[].requests[].stmt  n=20  str:20  | 값 | 관찰 | G9
$.conns[].join  n=24  list:24 len[min=2,max=2,sum=48]  | 출처 | 도출 | E36
$.conns[].join[]  n=48  int:48  | 출처 | 도출 | E36
$.java  n=14  obj:14 len[min=6,max=6,sum=84]  | 모음 | 관찰 | G9
$.java.dropped  n=14  obj:14 len[min=4,max=4,sum=56] (dynamic keys)  | 출처 | 도출 | E30
$.java.dropped.{*}  n=56  int:56  | 출처 | 도출 | E30
$.java.dups  n=14  list:14 len[min=0,max=3,sum=20]  | 출처 | 도출 | E30
$.java.dups[]  n=20  obj:20 len[min=5,max=5,sum=100]  | 출처 | 도출 | E30
$.java.dups[].fn  n=20  str:20  | 출처 | 도출 | E30
$.java.dups[].desc  n=20  str:20  | 출처 | 도출 | E30
$.java.dups[].thread  n=20  str:20  | 출처 | 도출 | E30
$.java.dups[].t0  n=20  float:20  | 출처 | 도출 | E30
$.java.dups[].durs  n=20  list:20 len[min=2,max=2,sum=40]  | 출처 | 도출 | E30
$.java.dups[].durs[]  n=40  float:38,int:2  | 출처 | 도출 | E30
$.java.parks  n=14  list:14 len[min=0,max=1,sum=4]  | 모음 | 관찰 | G9
$.java.spans  n=14  list:14 len[min=93,max=143,sum=1592]  | 모음 | 관찰 | G9
$.java.spans[]  n=1592  obj:1592 len[min=7,max=8,sum=11164]  | 사건 | 관찰 | G8
$.java.spans[].fn  n=1592  str:1592  | 장소 | 관찰 | G4
$.java.spans[].desc  n=1592  str:1592  | 장소 | 관찰 | G4
$.java.spans[].thread  n=1592  str:1592  | 장소 | 관찰 | G4
$.java.spans[].t0  n=1592  float:1526,int:66  | 시각 | 관찰 | G3
$.java.spans[].t1  n=1592  float:1533,int:59  | 시각 | 관찰 | G3
$.java.spans[].caller  n=1592  str:1592  | 연결 | 도출 | E32
$.java.spans[].depth  n=1592  int:1592  | 연결 | 도출 | E32
$.java.spans[].twice  n=20  bool:20  | 출처 | 도출 | E31
$.java.stackOnly  n=14  list:14 len[min=0,max=7,sum=46]  | 출처 | 도출 | E30
$.java.stackOnly[]  n=46  obj:46 len[min=5,max=5,sum=230]  | 출처 | 도출 | E30
$.java.stackOnly[].fn  n=46  str:46  | 출처 | 도출 | E30
$.java.stackOnly[].desc  n=46  str:46  | 출처 | 도출 | E30
$.java.stackOnly[].events  n=46  int:46  | 출처 | 도출 | E30
$.java.stackOnly[].first  n=46  float:45,int:1  | 출처 | 도출 | E30
$.java.stackOnly[].last  n=46  float:45,int:1  | 출처 | 도출 | E30
$.java.throws  n=14  list:14 len[min=2,max=6,sum=62]  | 모음 | 관찰 | G9
$.java.throws[]  n=62  obj:62 len[min=5,max=5,sum=310]  | 사건 | 관찰 | G8
$.java.throws[].t  n=62  float:61,int:1  | 시각 | 관찰 | G3
$.java.throws[].thread  n=62  str:62  | 장소 | 관찰 | G4
$.java.throws[].class  n=62  str:62  | 값 | 관찰 | G9
$.java.throws[].message  n=62  str:62  | 값 | 관찰 | G9
$.java.throws[].at  n=62  list:62 len[min=7,max=8,sum=478]  | 장소 | 관찰 | G4
$.java.throws[].at[]  n=478  str:478  | 장소 | 관찰 | G4
$.javaExit  n=14  int:14  | 값 | 관찰 | E42
$.javaVersion  n=14  str:14  | 값 | 관찰 | E42
$.joins  n=14  list:14 len[min=27,max=52,sum=562]  | 출처 | 도출 | E29
$.joins[]  n=562  obj:562 len[min=4,max=5,sum=2452]  | 출처 | 도출 | E29
$.joins[].pid  n=562  int:562  | 출처 | 도출 | E29
$.joins[].fn  n=562  str:562  | 출처 | 도출 | E29
$.joins[].return  n=204  bool:204  | 출처 | 도출 | E29
$.joins[].uprobe  n=562  int:562  | 출처 | 도출 | E29
$.joins[].recorder  n=562  int:562  | 출처 | 도출 | E29
$.notes  n=14  list:14 len[min=0,max=2,sum=12]  | 모음 | 관찰 | G9
$.notes[]  n=12  obj:12 len[min=4,max=4,sum=48]  | 사건 | 관찰 | G8
$.notes[].chain  n=12  str:12  | 이름 | 관찰 | E39
$.notes[].message  n=12  str:12  | 이름 | 관찰 | E39
$.notes[].t  n=12  int:12  | 시각 | 관찰 | G3
$.notes[].where  n=12  str:12  | 이름 | 관찰 | E39
$.pool  n=14  int:14  | 값 | 관찰 | E42
$.scene  n=14  str:14  | 값 | 관찰 | E42
$.sessions  n=14  obj:14 len[min=1,max=2,sum=24] (dynamic keys)  | 모음 | 관찰 | G9
$.sessions.{*}  n=24  str:24  | 장소 | 관찰 | E38
$.siteRepair  n=14  obj:14 len[min=5,max=5,sum=70]  | 출처 | 도출 | E29
$.siteRepair.repaired  n=14  int:14  | 출처 | 도출 | E29
$.siteRepair.unchanged  n=14  int:14  | 출처 | 도출 | E29
$.siteRepair.before  n=14  obj:14 len[min=1,max=1,sum=14] (dynamic keys)  | 출처 | 도출 | E29
$.siteRepair.before.{*}  n=14  int:14  | 출처 | 도출 | E29
$.siteRepair.leftAfter  n=14  int:14  | 출처 | 도출 | E29
$.siteRepair.leftFns  n=14  obj:14 len[min=1,max=3,sum=18]  | 출처 | 도출 | E29
$.siteRepair.leftFns.XLogBackgroundFlush  n=14  int:14  | 출처 | 도출 | E29
$.src  n=14  obj:14 len[min=100,max=113,sum=1434] (dynamic keys)  | 발췌 | 관찰 | G1
$.src.{*}  n=1434  obj:1434 len[min=2,max=2,sum=2868]  | 발췌 | 관찰 | G1
$.src.{*}.from  n=1434  int:1434  | 발췌 | 관찰 | G1
$.src.{*}.lines  n=1434  list:1434 len[min=21,max=21,sum=30114]  | 발췌 | 관찰 | G1
$.src.{*}.lines[]  n=30114  str:30114  | 발췌 | 관찰 | G1
$.tables  n=14  list:14 len[min=1,max=1,sum=14]  | 모음 | 관찰 | G9
$.tables[]  n=14  obj:14 len[min=4,max=4,sum=56]  | 장소 | 관찰 | G5c
$.tables[].name  n=14  str:14  | 장소 | 관찰 | G4
$.tables[].oid  n=14  int:14  | 값 | 관찰 | G9
$.tables[].columns  n=14  list:14 len[min=3,max=3,sum=42]  | 모음 | 관찰 | G9
$.tables[].columns[]  n=42  obj:42 len[min=2,max=2,sum=84]  | 장소 | 관찰 | G5c
$.tables[].columns[].name  n=42  str:42  | 장소 | 관찰 | G4
$.tables[].columns[].type  n=42  str:42  | 값 | 관찰 | G5b
$.tables[].indexes  n=14  list:14 len[min=1,max=1,sum=14]  | 모음 | 관찰 | G9
$.tables[].indexes[]  n=14  obj:14 len[min=2,max=2,sum=28]  | 장소 | 관찰 | G5c
$.tables[].indexes[].name  n=14  str:14  | 장소 | 관찰 | G4
$.tables[].indexes[].oid  n=14  int:14  | 값 | 관찰 | G9
$.title  n=14  str:14  | 값 | 관찰 | E42
$.transactions  n=14  list:14 len[min=1,max=2,sum=20]  | 모음 | 도출 | E22+G9
$.transactions[]  n=20  obj:20 len[min=9,max=9,sum=180]  | 사건 | 도출 | E22+G8
$.transactions[].conn  n=20  int:20  | 장소 | 도출 | E22+G4
$.transactions[].first  n=20  int:20  | 연결 | 도출 | E19
$.transactions[].last  n=20  int:20  | 연결 | 도출 | E19
$.transactions[].t0  n=20  int:20  | 시각 | 도출 | E22+G3
$.transactions[].t1  n=20  int:20  | 시각 | 도출 | E22+G3
$.transactions[].end  n=20  str:20  | 값 | 도출 | E21
$.transactions[].xids  n=20  list:20 len[min=1,max=2,sum=22]  | 모음 | 도출 | E22+G9
$.transactions[].xids[]  n=22  obj:22 len[min=3,max=3,sum=66]  | 사건 | 도출 | E22+G8
$.transactions[].xids[].xid  n=22  str:22  | 장소 | 도출 | E22+G4
$.transactions[].xids[].level  n=22  str:22  | 값 | 도출 | E22+G9
$.transactions[].xids[].t  n=22  float:22  | 시각 | 도출 | E22+G3
$.transactions[].clog  n=20  list:20 len[min=1,max=2,sum=22]  | 모음 | 도출 | E22+G9
$.transactions[].clog[]  n=22  obj:22 len[min=4,max=4,sum=88]  | 사건 | 도출 | E22+G8
$.transactions[].clog[].kind  n=22  str:22  | 사건 | 도출 | E22+G6
$.transactions[].clog[].nsub  n=22  str:22  | 값 | 도출 | E22+G9
$.transactions[].clog[].t  n=22  float:22  | 시각 | 도출 | E22+G3
$.transactions[].clog[].xid  n=22  str:22  | 장소 | 도출 | E22+G4
$.transactions[].next  n=20  obj:20 len[min=5,max=6,sum=102]  | 연결 | 도출 | E20
$.transactions[].next.t  n=20  int:20  | 연결 | 도출 | E20
$.transactions[].next.tc  n=20  float:20  | 연결 | 도출 | E20
$.transactions[].next.dir  n=20  str:20  | 연결 | 도출 | E20
$.transactions[].next.type  n=20  str:20  | 연결 | 도출 | E20
$.transactions[].next.len  n=20  int:20  | 연결 | 도출 | E20
$.view  n=14  obj:14 len[min=16,max=16,sum=224]  | 모음 | 도출 | E18+G9
$.view.window  n=14  list:14 len[min=2,max=2,sum=28]  | 시각 | 도출 | E17
$.view.window[]  n=28  int:28  | 시각 | 도출 | E17
$.view.tree  n=14  list:14 len[min=82,max=124,sum=1388]  | 모음 | 도출 | E18+G9
$.view.tree[]  n=1388  obj:1388 len[min=17,max=19,sum=25014]  | 사건 | 도출 | E7
$.view.tree[].id  n=1388  int:1388  | 연결 | 도출 | E9c
$.view.tree[].parent  n=1388  int:1388  | 연결 | 도출 | E2
$.view.tree[].fn  n=1388  str:1388  | 장소 | 도출 | E18+G4
$.view.tree[].desc  n=1388  str:1388  | 장소 | 도출 | E18+G4
$.view.tree[].label  n=1388  str:1388  | 이름 | 도출 | E4
$.view.tree[].lane  n=1388  str:1388  | 장소 | 도출 | E18+G4
$.view.tree[].depth  n=1388  int:1388  | 값 | 도출 | E18+G9
$.view.tree[].t0  n=1388  float:1332,int:56  | 시각 | 도출 | E18+G3
$.view.tree[].t1  n=1388  float:1284,null:54,int:50  | 시각 | 도출 | E18+G3
$.view.tree[].site  n=1388  int:1388  | 장소 | 도출 | E18+G4
$.view.tree[].hidden  n=1388  list:1388 len[min=0,max=10,sum=978]  | 장소 | 도출 | E6
$.view.tree[].hidden[]  n=978  str:978  | 장소 | 도출 | E6
$.view.tree[].hiddenKinds  n=1388  list:1388 len[min=0,max=3,sum=326]  | 모음 | 도출 | E5
$.view.tree[].hiddenKinds[]  n=326  str:326  | 값 | 도출 | E5
$.view.tree[].end  n=1388  str:1388  | 값 | 도출 | E5
$.view.tree[].self  n=1388  float:1271,int:47,null:70  | 시각 | 도출 | E18+G3
$.view.tree[].kids  n=1388  list:1388 len[min=0,max=7,sum=1372]  | 연결 | 도출 | E2
$.view.tree[].kids[]  n=1372  int:1372  | 연결 | 도출 | E2
$.view.tree[].effects  n=1388  list:1388 len[min=0,max=2,sum=224]  | 모음 | 도출 | E18+G9
$.view.tree[].why  n=1388  list:1388 len[min=0,max=3,sum=701]  | 이름 | 도출 | E4
$.view.tree[].branch  n=1364  obj:1364 len[min=2,max=2,sum=2728]  | 연결 | 도출 | E2
$.view.tree[].branch.key  n=1364  str:1364  | 연결 | 도출 | E2
$.view.tree[].branch.alt  n=1364  int:1364  | 연결 | 도출 | E2
$.view.tree[].why[]  n=701  str:701  | 이름 | 도출 | E4
$.view.tree[].effects[]  n=224  obj:224 len[min=1,max=5,sum=642]  | 사건 | 도출 | E8
$.view.tree[].effects[].kind  n=224  str:224  | 사건 | 도출 | E18+G6
$.view.tree[].effects[].conn  n=126  str:126  | 장소 | 도출 | E18+G4
$.view.tree[].effects[].result  n=24  str:24  | 값 | 도출 | E18+G9
$.view.tree[].effects[].idleMs  n=20  float:20  | 시각 | 도출 | E18+G3
$.view.tree[].effects[].validated  n=18  bool:18  | 값 | 도출 | E18+G9
$.view.tree[].effects[].seq  n=70  list:70 len[min=1,max=2,sum=90]  | 연결 | 도출 | E3
$.view.tree[].effects[].seq[]  n=90  int:90  | 연결 | 도출 | E3
$.view.tree[].effects[].value  n=40  str:40  | 값 | 도출 | E18+G9
$.view.tree[].seen  n=54  float:54  | 시각 | 도출 | E18+G3
$.view.tree[].effects[].t0  n=34  float:33,int:1  | 시각 | 도출 | E18+G3
$.view.tree[].effects[].at  n=34  str:34  | 장소 | 도출 | E18+G4
$.view.tree[].effects[].class  n=30  str:30  | 값 | 도출 | E18+G9
$.view.tree[].effects[].message  n=18  str:18  | 값 | 도출 | E18+G9
$.view.roots  n=14  list:14 len[min=1,max=2,sum=16]  | 연결 | 도출 | E11
$.view.roots[]  n=16  int:16  | 연결 | 도출 | E11
$.view.ambiguous  n=14  int:14  | 값 | 도출 | E18+G9
$.view.binding  n=14  list:14 len[min=3,max=7,sum=70]  | 모음 | 도출 | E18+G9
$.view.binding[]  n=70  obj:70 len[min=4,max=4,sum=280]  | 사건 | 도출 | E18+G8
$.view.binding[].t  n=70  int:17,float:53  | 시각 | 도출 | E18+G3
$.view.binding[].bound  n=70  str:70  | 연결 | 도출 | E16
$.view.binding[].held  n=70  list:70 len[min=0,max=1,sum=16]  | 연결 | 도출 | E16
$.view.binding[].by  n=70  int:70  | 연결 | 도출 | E16
$.view.slots  n=14  list:14 len[min=1,max=2,sum=20]  | 모음 | 도출 | E18+G9
$.view.slots[]  n=20  obj:20 len[min=3,max=3,sum=60]  | 사건 | 도출 | E18+G8
$.view.slots[].conn  n=20  str:20  | 장소 | 도출 | E18+G4
$.view.slots[].t0  n=20  float:19,int:1  | 시각 | 도출 | E18+G3
$.view.slots[].t1  n=20  float:19,int:1  | 시각 | 도출 | E18+G3
$.view.waits  n=14  list:14 len[min=0,max=1,sum=2]  | 모음 | 도출 | E18+G9
$.view.pauses  n=14  list:14 len[min=3,max=6,sum=68]  | 시각 | 도출 | E17
$.view.pauses[]  n=68  list:68 len[min=2,max=2,sum=136]  | 시각 | 도출 | E17
$.view.pauses[][]  n=136  int:136  | 시각 | 도출 | E17
$.view.java  n=14  obj:14 len[min=1,max=2,sum=22] (dynamic keys)  | 발췌 | 관찰 | E1
$.view.java.{*}  n=22  obj:22 len[min=2,max=2,sum=44]  | 발췌 | 관찰 | E1
$.view.java.{*}.path  n=22  str:22  | 발췌 | 관찰 | E1
$.view.java.{*}.lines  n=22  list:22 len[min=51,max=79,sum=1394]  | 발췌 | 관찰 | E1
$.view.java.{*}.lines[]  n=1394  str:1394  | 발췌 | 관찰 | E1
$.view.jsrc  n=14  obj:14 len[min=53,max=73,sum=882] (dynamic keys)  | 발췌 | 관찰 | E1
$.view.jsrc.{*}  n=882  obj:882 len[min=2,max=2,sum=1764]  | 발췌 | 관찰 | E1
$.view.jsrc.{*}.from  n=882  int:882  | 발췌 | 관찰 | E1
$.view.jsrc.{*}.lines  n=882  list:882 len[min=10,max=15,sum=13220]  | 발췌 | 관찰 | E1
$.view.jsrc.{*}.lines[]  n=13220  str:13220  | 발췌 | 관찰 | E1
$.view.branches  n=14  obj:14 len[min=72,max=72,sum=1008] (dynamic keys)  | 모음 | 도출 | E18+G9
$.view.branches.{*}  n=1008  obj:1008 len[min=7,max=7,sum=7056]  | 모음 | 도출 | E18+G9
$.view.branches.{*}.fn  n=1008  str:1008  | 장소 | 도출 | E18+G4
$.view.branches.{*}.desc  n=1008  str:1008  | 장소 | 도출 | E18+G4
$.view.branches.{*}.label  n=1008  str:1008  | 이름 | 도출 | E10
$.view.branches.{*}.alts  n=1008  list:1008 len[min=1,max=4,sum=1428]  | 모음 | 도출 | E18+G9
$.view.branches.{*}.alts[]  n=1428  obj:1428 len[min=8,max=8,sum=11424]  | 사건 | 도출 | E18+G8
$.view.branches.{*}.alts[].seq  n=1428  list:1428 len[min=0,max=7,sum=1890]  | 연결 | 도출 | E9
$.view.branches.{*}.alts[].seq[]  n=1890  str:1890  | 연결 | 도출 | E9
$.view.branches.{*}.alts[].labels  n=1428  list:1428 len[min=0,max=7,sum=1890]  | 이름 | 도출 | E10
$.view.branches.{*}.alts[].labels[]  n=1890  str:1890  | 이름 | 도출 | E10
$.view.branches.{*}.alts[].sites  n=1428  list:1428 len[min=0,max=7,sum=1890]  | 장소 | 도출 | E9b
$.view.branches.{*}.alts[].sites[]  n=1890  int:1890  | 장소 | 도출 | E9b
$.view.branches.{*}.alts[].end  n=1428  str:1428  | 값 | 도출 | E18+G9
$.view.branches.{*}.alts[].lastExc  n=1428  bool:1428  | 값 | 도출 | E18+G9
$.view.branches.{*}.alts[].cut  n=1428  bool:1428  | 값 | 도출 | E18+G9
$.view.branches.{*}.alts[].n  n=1428  int:1428  | 값 | 도출 | E18+G9
$.view.branches.{*}.alts[].records  n=1428  list:1428 len[min=2,max=14,sum=11368]  | 연결 | 도출 | E9
$.view.branches.{*}.alts[].records[]  n=11368  str:11368  | 연결 | 도출 | E9
$.view.branches.{*}.flagged  n=1008  bool:1008  | 값 | 도출 | E18+G9
$.view.branches.{*}.common  n=1008  int:1008  | 값 | 도출 | E18+G9
$.view.branches.{*}.lines  n=1008  list:1008 len[min=0,max=3,sum=252]  | 장소 | 도출 | E9b
$.view.branches.{*}.lines[]  n=252  int:252  | 장소 | 도출 | E9b
$.view.compared  n=14  list:14 len[min=14,max=14,sum=196]  | 연결 | 도출 | E11
$.view.compared[]  n=196  str:196  | 연결 | 도출 | E11
$.view.javaFrom  n=14  str:14  | 출처 | 도출 | E13
$.view.siteMismatch  n=14  list:14 len[min=0,max=0,sum=0]  | 출처 | 도출 | E13
$.view.waitCycles  n=14  list:14 len[min=0,max=1,sum=6]  | 모음 | 도출 | E18+G9
$.view.poolConfig  n=14  obj:14 len[min=5,max=5,sum=70]  | 모음 | 도출 | E18+G9
$.view.poolConfig.connectionTimeoutMs  n=14  int:14  | 값 | 도출 | E15
$.view.poolConfig.connectionTimeoutAt  n=14  str:14  | 발췌 | 도출 | E14
$.view.poolConfig.aliveBypassMs  n=14  int:14  | 값 | 도출 | E15
$.view.poolConfig.aliveBypassAt  n=14  str:14  | 발췌 | 도출 | E14
$.view.poolConfig.aliveCheckAt  n=14  str:14  | 발췌 | 도출 | E14
$.captures[].overlap[]  n=17  obj:17 len[min=5,max=7,sum=95]  | 사건 | 도출 | E25+G8
$.captures[].overlap[].t  n=17  int:17  | 시각 | 도출 | E25+G3
$.captures[].overlap[].tc  n=17  float:17  | 시각 | 도출 | E25+G3
$.captures[].overlap[].dir  n=17  str:17  | 사건 | 도출 | E25+G6
$.captures[].overlap[].type  n=17  str:17  | 사건 | 도출 | E25+G6
$.captures[].overlap[].len  n=17  int:17  | 값 | 도출 | E25+G9
$.transactions[].next.text  n=2  str:2  | 연결 | 도출 | E20
$.view.binding[].held[]  n=16  str:16  | 연결 | 도출 | E16
$.blocked[]  n=6  obj:6 len[min=4,max=4,sum=24]  | 사건 | 관찰 | G8
$.blocked[].after  n=6  int:6  | 연결 | 관찰 | E40
$.blocked[].t  n=6  int:6  | 시각 | 관찰 | G3
$.blocked[].tCapture  n=6  float:6  | 시각 | 관찰 | G3
$.blocked[].waiters  n=6  list:6 len[min=0,max=1,sum=2]  | 장소 | 관찰 | E41
$.captures[].overlap[].name  n=4  str:4  | 장소 | 도출 | E25+G4
$.captures[].overlap[].text  n=6  str:6  | 값 | 도출 | E25+G9
$.java.parks[]  n=4  obj:4 len[min=5,max=5,sum=20]  | 사건 | 관찰 | G8
$.java.parks[].t0  n=4  float:4  | 시각 | 관찰 | G3
$.java.parks[].t1  n=4  float:4  | 시각 | 관찰 | G3
$.java.parks[].thread  n=4  str:4  | 장소 | 관찰 | G4
$.java.parks[].under  n=4  str:4  | 장소 | 관찰 | G4
$.java.parks[].at  n=4  list:4 len[min=6,max=6,sum=24]  | 장소 | 관찰 | G4
$.java.parks[].at[]  n=24  str:24  | 장소 | 관찰 | G4
$.view.tree[].effects[].t1  n=4  float:4  | 시각 | 도출 | E18+G3
$.view.waitCycles[]  n=6  obj:6 len[min=10,max=11,sum=64]  | 사건 | 도출 | E18+G8
$.view.waitCycles[].kind  n=6  str:6  | 사건 | 도출 | E18+G6
$.view.waitCycles[].pool  n=4  int:4  | 값 | 도출 | E18+G9
$.view.waitCycles[].t0  n=6  float:6  | 시각 | 도출 | E18+G3
$.view.waitCycles[].t1  n=6  float:6  | 시각 | 도출 | E18+G3
$.view.waitCycles[].waiter  n=6  str:6  | 연결 | 도출 | E12
$.view.waitCycles[].xid  n=6  str:6  | 장소 | 도출 | E18+G4
$.view.waitCycles[].holder  n=6  str:6  | 연결 | 도출 | E12
$.view.waitCycles[].thread  n=6  str:6  | 장소 | 도출 | E18+G4
$.view.waitCycles[].call  n=6  int:6  | 연결 | 도출 | E12
$.view.waitCycles[].heldBy  n=6  str:6  | 연결 | 도출 | E12
$.view.waitCycles[].cycle  n=6  bool:6  | 연결 | 도출 | E12
$.blocked[].waiters[]  n=2  str:2  | 장소 | 관찰 | E41
$.conns[].messages[].f.code  n=2  str:2  | 값 | 관찰 | G9
$.conns[].messages[].f.severity  n=2  str:2  | 값 | 관찰 | G9
$.conns[].messages[].f.where  n=2  str:2  | 값 | 관찰 | G9
$.conns[].requests[].error  n=2  obj:2 len[min=3,max=3,sum=6]  | 모음 | 관찰 | G9
$.conns[].requests[].error.code  n=2  str:2  | 값 | 관찰 | G9
$.conns[].requests[].error.message  n=2  str:2  | 값 | 관찰 | G9
$.conns[].requests[].error.where  n=2  str:2  | 값 | 관찰 | G9
$.conns[].requests[].server.events[].f.oper  n=2  str:2  | 값 | 관찰 | G9
$.siteRepair.leftFns.LockAcquireExtended  n=2  int:2  | 출처 | 도출 | E29
$.siteRepair.leftFns.ProcSleep  n=2  int:2  | 출처 | 도출 | E29
$.view.waits[]  n=2  obj:2 len[min=6,max=6,sum=12]  | 사건 | 도출 | E18+G8
$.view.waits[].conn  n=2  str:2  | 장소 | 도출 | E18+G4
$.view.waits[].seq  n=2  int:2  | 연결 | 도출 | E18+G5
$.view.waits[].t0  n=2  float:2  | 시각 | 도출 | E18+G3
$.view.waits[].t1  n=2  float:2  | 시각 | 도출 | E18+G3
$.view.waits[].xid  n=2  str:2  | 장소 | 도출 | E18+G4
$.view.waits[].end  n=2  str:2  | 값 | 도출 | E18+G9
# role x provenance line counts: {('값', '관찰'): 77, ('장소', '관찰'): 71, ('출처', '도출'): 52, ('모음', '관찰'): 41, ('연결', '도출'): 40, ('시각', '도출'): 34, ('시각', '관찰'): 31, ('값', '도출'): 30, ('장소', '도출'): 26, ('모음', '도출'): 19, ('사건', '관찰'): 18, ('사건', '도출'): 17, ('발췌', '관찰'): 15, ('이름', '도출'): 6, ('이름', '관찰'): 5, ('발췌', '도출'): 3, ('연결', '관찰'): 2, ('출처', '관찰'): 2}
```

### 2.4 동적 키와 사건 종류별 속성

서버 사건의 `f`는 종류마다 키가 다릅니다. 값은 모두 문자열입니다 [셈: `dyn.py`, keytree].

| `kind` | 수 | `f` 키(나온 횟수) |
| --- | --- | --- |
| ssi_precommit | 80 | sxact:80 |
| lock | 72 | ctid:2, mode:72, rel:50, result:72, type:72, xid:22 |
| pred_lock | 48 | created:48, ctid:48, gran:48, rel:48, txid:48 |
| hint_fpi | 28 | buf:28, lsn:28 |
| update_check | 24 | off:24, result:24, xmax:24, xmin:24 |
| assign_xid | 22 | level:22 |
| xid | 22 | xid:22 |
| conflict_in | 22 | blk:22, rel:22 |
| hint | 14 | at:14, bit:14, tup:14, xid:14 |
| abort_clog | 12 | nsub:12, xid:12 |
| wal_flush | 10 | lsn:10 |
| commit_clog | 10 | nsub:10, xid:10 |
| async_lsn | 10 | lsn:10 |
| alarm | 4 | (없음) |
| xact_wait | 2 | oper:2, xid:2 |
| deadlock_check | 2 | result:2 |
| ereport | 2 | at:2 |

E-03에는 strace에서 오는 `io`, `sleep_syscall`, `kill` 종류가 0건입니다 [셈]. tx 실행(3.3절)에는 있습니다.

그 밖의 동적 키 [셈: `dyn.py`].

| 경로 | 키 | 값 |
| --- | --- | --- |
| `sessions.{*}` | 백엔드 pid(문자열 키) 24개 | `C1` 또는 `C2`. pool1 장면 4개는 C1 하나뿐 |
| `captures[].xacts.{*}` | xid | committed 96, in progress 36, aborted 24 |
| `java.dropped.{*}` | 버린 까닭 | (장면 앞) 476, java.lang.ClassNotFoundException 354, java.lang.NoSuchMethodError 1,016, jdk.ThreadPark 43 (14개 합) |
| `siteRepair.before.{*}` | 고치기 전 주소 | `0xfffffffff000` 하나, 13,512 (14개 합) |
| `src.{*}` | `src/backend/...:줄` 1,350, `src/include/...:줄` 84 | `{from, lines[]}` 발췌 1,434 |
| `view.jsrc.{*}` | `자바클래스:줄` 882 | `{from, lines[]}` |
| `view.java.{*}` | 앱 클래스 이름(`lab.Inner`, `lab.Outer` 등) 22 | `{path, lines[]}` |
| `view.branches.{*}` | `메서드이름+서술자` 1,008(실행마다 72) | 갈림 정보 |

메시지 `type`은 방향과 함께 봐야 뜻이 정해집니다(같은 `D`가 앞쪽 F에서는 Describe, 뒤쪽 B에서는 DataRow입니다). 분포는 다음과 같습니다 [셈: `stats.py`]: B·S 384, B·Z 166, B·C 108, F·Q 76, B·2 66, F·B 66, F·E 66, F·S 66, B·1 64, F·P 64, F·D 42, B·I 32, B·n 28, B·K 24, B·R 24, B·sslreply 24, F·X 24, F·ssl 24, F·startup 24, B·D 14, B·T 14, B·E 2. 요청 상태는 I 114, T 50, E 2입니다.

### 2.5 기존 화면이 보여 주는 사실

화면은 `workbench/design/spring-view.tmpl.html`(897행)이고, `__RUNS__`를 채운 결과가 `concept-9-spring.html`입니다. concept-9의 129행 literal을 템플릿 자리표시자로 되돌리면 템플릿과 같으므로, 아래 행 번호는 두 파일에 함께 맞습니다 [셈: `stats4.py`]. 템플릿 머리 주석(121~128행)은 "화면은 고르고 그리기만" 하고, 호출 트리·깊이·끝난 꼴·자기 시간·효과·렌즈 이유·갈림·연결 묶임·풀 사용은 해석기(`view`)가 셈했다고 적습니다 [코드].

화면이 데이터에서 꺼내 보이는 사실과 그 함수입니다 [코드].

| 화면 사실 | 함수(행) | 읽는 키 |
| --- | --- | --- |
| 효과 한 줄 글(빌림·돌려줌·떼어 둠·다시 묶음·자동 커밋·롤백 전용·생존 검사·잠듦·예외) | `TOPIC.effect` 184, `effectDetail` 201 | `view.tree[].effects[]`, `view.poolConfig` |
| 연결 위 요청·응답 줄, 서버 사건 줄, 앱 오류 메모 줄, 막힘 관찰 띠 | `TOPIC.rows` 227, `serverText` 473 | `conns[].requests[]`(같은 `sent`끼리 묶음), `server.events[]`, `notes[]`, `captures[]` |
| 시간 구간(연결 묶임 막대) | `TOPIC.intervals` 253 | `view.slots[]`, `view.binding[]` |
| 효과 기호 | `TOPIC.effGlyph` 258 | `effects[].kind` |
| 시각 T의 상태판: 묶인 연결·잡힌 연결·빌려준 연결, 열린 연결, 보낸 뒤 답을 기다리는 요청, 기다림 화살표, xid 알약과 clog 상태, `lock_timeout`, 교착 검사, 고리(↻), 롤백 전용(⚑), 생존 창 막대 | `TOPIC.board` 275~380 | `binding`, `slots`, `conns`, `requests`, `waits`, `transactions`, `server.events`, `waitCycles`, `poolConfig` |
| 간트 줄 | `TOPIC.ganttRows` 383 | `slots`, `waits` |
| 결과 칩(잔액 시작 → 끝) | `TOPIC.outcomeChips` 397 | `captures[]`의 `label`이 start, end인 것의 `pages`/`visible` |
| 시작 줄 | `TOPIC.startRow` 409 | `checkpoints` |
| 기록 흠(막힘 관찰에 겹친 메시지 수) | `TOPIC.runFlaw` 416 | `captures[].overlap` |
| 이야기 순간들 | `TOPIC.moments` 418 | 여러 키 |
| 용어 풀이 | `TOPIC.glossary` 437 | 없음(화면 자체 글) |
| 요청 글, SQL 글, 응답 글 | `reqText` 465, `sqlText` 469, `backText` 470 | `requests[].types`, `sql`, `tags`, `error`, `status` |
| 시각 T 직전 관찰 | `lastCapture` 471 | `captures[].t` |
| 메시지·서버 사건·관찰의 상세 칸 | `wireDetail` 484, `serverDetail` 495, `capDetail` 504 | `messages[]`, `server.*`, `captures[]` |
| 호출 트리 줄(렌즈로 고른 마디) | `visible` 555, `rows` 573, `lanes` 655 | `view.tree[]`, `why`, `hidden`, `hiddenKinds`, `branch` |
| 머리(보이는 마디 수 / 전체) | `header` 590 | `view.tree.length` |
| 장면 설명 | `about` 603 | `TOPIC.scenes`(화면 자체 글) |
| 시간 막대(스크러버) | `scrubber` 611 | `view.window`, `pauses` |
| 범례 | `legend` 729 | 화면 자체 |
| 미니맵 | `minimap` 752 | `view.tree` |
| 소스 창(앱·라이브러리 줄, 갈린 줄), 소스 출처 경고 | `codeWins` 760, `srcBlock` 786 | `view.java`, `view.jsrc`, `view.javaFrom`, `view.siteMismatch` |
| 마디 상세(끝난 꼴, 자기 시간, 효과, 갈림의 갈린 자리와 줄) | `nodeDetail` 788 | `tree[]`, `branches{}` |

화면이 스스로 셈하는 것입니다 [코드]. 변환 뒤 다시 뽑을 때는 이 계산을 화면 쪽에 그대로 남기거나, 같은 결과를 내는 규칙을 별도로 적어야 합니다.

1. 시각을 `view.window[0]` 기준의 ms로 바꿉니다(`rel` 137행, `plus` 138행, `ms` 134행).
2. 렌즈와 펼침으로 보일 마디를 고릅니다(`visible` 555).
3. 칸 안에서 깊이에 따라 들여 씁니다(`lanes` 659행 `indent`).
4. 같은 `sent`를 가진 요청을 한 줄로 묶습니다(`TOPIC.rows` 227).
5. 시각 T에 열린 마디를 찾습니다(`nodeAt` 743).
6. 시각 T의 상태판 전체를 다시 셈합니다(`TOPIC.board` 275~380). 여기서 `lock_timeout` 값은 SQL 글을 정규식으로 읽어 얻습니다.
7. 60 ms보다 긴 빈 구간을 접습니다(`layout` 639).
8. 기록 흠 수를 셉니다(`runFlaw` 416).

화면이 자체로 들고 있는 사실입니다. 이것은 어떤 데이터 literal에도 없으므로, 공통 형식에 넣지 않으면 다시 뽑을 때 "빠진 사실"이 됩니다 [셈: `consts.py`, 코드].

| 상수 | 행 | 내용 |
| --- | --- | --- |
| `TOPIC.scenes` | 156~168 | 장면 6개마다 이름, 질문, 앱 코드가 하는 일을 적은 글(예: "안쪽은 SET LOCAL lock_timeout = '3s'로 …") |
| `TOPIC.lanes` | 169~177 | 칸 8개: app, tx, jdbc, hikari, pg, wire(link), C1, C2(peer). 데이터의 `lane`은 app, tx, jdbc, hikari, pg 다섯만 나오고 `other`는 0건 [셈] |
| `TOPIC.codeNote`, `stackSource`, `effectLegend`, `rowLegend`, `boardLegend`, `glossary` | 178~462 | 범례와 용어 풀이 글. `stackSource`는 "JFR 기록"이라는 출처 문구 |
| `LENSES` | 522 | 렌즈 4개(효과, 갈림, 시간, 고정)의 이름과 설명 |
| `WHYGLYPH`, `WHYCOLOR` | 523~524 | 렌즈 기호와 색 |
| `END` | 525~526 | 끝난 꼴 5가지(return, throw, twice, passed, rethrow)의 뜻 |

화면이 "추론" 표시를 붙이는 곳입니다(`inf` 140행) [코드].

- 211행: 빌림의 쉰 시간(`idleMs`)은 해석기가 그 연결의 직전 메시지부터 잰 값이라고 밝힙니다.
- 215행: 롤백 전용 표시의 뜻은 Spring 소스로 읽은 뜻이라고 밝힙니다.
- 222행: `return`, `suspend`, `resume` 효과는 "호출 순서로 셈"한 값이라고 밝힙니다.
- 831행: 갈림에서 무엇이 길을 정했는지는 기록이 반환값을 남기지 않아 소스와 다음 호출로 읽는다고 밝힙니다.
- 220행(태그 없음, 괄호 문구): 자동 커밋 값은 "부른 줄의 소스 글에서 읽음"이라고 적습니다.

### 2.6 모호함·불일치·사실처럼 저장된 추론

사실처럼 저장된 추론과 도출값입니다. 공통 형식에서 관찰과 같은 열에 두면 "차이 0"이 해석기 버전에 묶입니다.

1. `borrows[]` 24건은 JFR 구간에서 휴리스틱(100 ms 창)으로 연결을 붙인 것입니다(`decode/spring_jfr.go`) [코드]. 그 가운데 4건은 `conn: ""`입니다. 모두 pool1 장면의 `result: "timeout"`이고 `idleMs: 0`, `validated: false`입니다 [셈]. 연결 장소가 없는 사건이므로 link 표에서 빈 끝을 허용해야 합니다.
2. `view.binding[]`(70)은 관찰한 값이 아니라 호출 순서로 다시 셈한 값입니다(`spring_tree.go`의 `bindingReplay`, 주석 "관찰한 값이 아니라 셈한 값") [코드]. 효과 `return` 20, `suspend` 8, `resume` 8도 같습니다 [셈].
3. 끝난 꼴 `passed` 54, `rethrow` 4는 해석기의 판정입니다. `passed` 54건은 모두 `t1 = null`이고 `seen`(마지막으로 본 시각)만 있습니다 [셈]. 끝 시각이 없는 사건을 공통 형식이 허용해야 합니다.
4. `view.waitCycles[]` 6건(server 2, pool 4)의 `heldBy`, `cycle`은 추론입니다(`serverWaits`, `poolWaitCycles`) [코드, 셈].
5. `effects[].value`(autocommit 40건)는 부른 줄의 소스 글에서 읽은 값입니다(2.5절 220행) [코드].
6. `view.poolConfig`는 소스를 정규식으로 읽은 값입니다(`poolConfigOf`) [코드].
7. `captures[].visible`은 `xacts`와 힌트 비트로 셈한 "보일 행"입니다(`visibleRows`) [코드].
8. `server.checks[].visible`도 해석기가 판정 경로로 정한 값입니다 [코드].
9. `java.spans[].twice` 20, `java.dups` 20, `java.stackOnly` 46(사건 1,546), `java.dropped`는 중복 제거와 소음 버리기의 결과입니다(`spring_jfr.go`) [코드, 셈].
10. `view.tree`(1,388)는 `java.spans`(1,592)에서 만든 것입니다(`buildCallTree`) [코드]. 둘을 독립 사건으로 더하면 안 됩니다. 앱 코드(`lab.*`) 마디는 spans에 44, tree에 24입니다 [셈]. 1,592와 1,388의 차 204가 무엇으로 이루어졌는지는 확인하지 않았습니다.
11. `transactions[]`(20)는 ReadyForQuery 상태가 I를 떠났다가 돌아오는 것으로 셈한 것입니다(`deriveTxs`) [코드]. `transactions[].xids`(22)와 `clog`(22)는 서버 사건 `xid` 22건, `commit_clog` 10 + `abort_clog` 12건의 사본입니다 [셈]. 중복으로 세면 안 됩니다.
12. `conns[].requests[]`의 경계(Sync 또는 단순 질의로 끝나는 묶음, Seq 0은 인증)는 해석기가 정합니다(`splitRequests`) [코드]. 서버 쪽 창(`server`)은 ReadyForQuery 입구 사이로 정하고, `join[0] == join[1]`일 때만 믿습니다 [코드]. 이번 데이터는 24개 연결 모두 같습니다 [셈].

불일치와 모호함입니다.

1. 같은 연결을 두 타입으로 부릅니다. `conns[].conn`, `transactions[].conn`, `clock.wire[].conn`은 정수(1, 2)이고, `borrows[].conn`, `effects[].conn`, `slots[].conn`, `waits[].conn`은 문자열(`"C1"`)입니다. 둘을 잇는 것은 `conns[].name`입니다 [셈: keytree].
2. 트랜잭션 번호도 두 타입입니다. `snapshots[].xmin/xmax`, `checks[].xmin/xmax`는 정수이고, `events[].f.xid`, `transactions[].xids[].xid`, `captures[].activity[].xid`, `captures[].locks[].xid`, `waitCycles[].xid`는 문자열입니다 [셈].
3. pid도 두 타입입니다. `conns[].pid`는 정수, `messages[].f.pid`(BackendKeyData)는 문자열, `sessions`의 키는 문자열입니다 [셈].
4. `snapshots[].xip`는 72건 모두 `null`입니다 [셈: `stats4.py`]. 해석기는 `xcnt`만큼 최대 4개를 넣습니다(`decode/scenario.go` 825~827행) [코드]. 0개면 Go의 nil 슬라이스가 `null`로 나가므로, `null`은 "진행 중 트랜잭션 0개"와 "기록 안 됨"을 가르지 못합니다 [추론]. tx 화면은 이 목록을 줄 판정의 xmin·xmax로 스냅숏을 거꾸로 찾아 채웁니다(`tx-view.tmpl.html` 779~790행 `snapOf`) [코드].
5. `requests[].server.in`, `out`은 E-03에서 항상 `null`입니다 [셈: keytree]. 연결 위 메시지는 `conns[].messages`에만 있습니다.
6. `siteRepair.leftAfter`(고치지 못하고 남은 부른 자리 수)는 실행마다 38~86입니다 [셈]. 실행별 값: 38, 46, 48×3, 50×2, 52, 56, 62, 64, 66, 80, 86(합 794). `leftFns`는 14개 실행 모두 `XLogBackgroundFlush`를 담고, `LockAcquireExtended`와 `ProcSleep`은 각각 2개 실행에만 나옵니다. 이 함수들의 `site`는 믿을 수 없습니다 [코드: spring.go `repairSites`].
7. `view.ambiguous`는 14개 모두 0입니다 [셈]. 값이 있는 칸이지만 이번 데이터에서는 정보가 없습니다.
8. `captures[].overlap`(관찰 도중 지나간 메시지)은 3건의 관찰에 17개입니다 [셈]. 5-reborrow/a의 `closed` 관찰에 1개(X), 5-requires-new-pool1/a·b의 `blocked` 관찰에 8개씩(ROLLBACK의 P·B 등)입니다. 이 관찰들은 그 메시지의 앞인지 뒤인지 섞였을 수 있습니다 [추론].
9. `blocked[].waiters`는 txlab이 pid로 기록하고(`txlab/spring.go` 138~143행, `observe.go`의 `lockWaiters`) 해석기가 이름(`C2`)으로 바꾼 것입니다 [코드].
10. 시각 0이 "값 없음"과 섞입니다(2.2절).
11. `clock.wire`의 `lo`/`hi`는 밀리초이고 다른 시각은 마이크로초입니다(2.2절). 주석은 구간이 1 ms를 넘으면 두 시계를 한 축에 놓지 말라고 합니다(spring.go 21~22행) [코드].
12. `clock.uprobe`의 `events` 합은 118,424이고 `joined` 합은 16,530입니다 [셈]. 해석 결과에 남은 서버 사건(384)은 원시 uprobe 사건의 일부입니다. 원시 `.bin`은 지워졌으므로 "빠진 사실 0"은 해석 결과 JSON을 기준으로만 판정할 수 있습니다.

### 2.7 규모

E-03의 "기록기가 남긴 사건 행"은 다음 합으로 정의합니다 [셈: `stats.py`].

| 표 | 수 |
| --- | --- |
| 연결 위 메시지 `conns[].messages[]` | 1,402 |
| 서버 사건 `server.events[]` | 384 |
| 스냅숏 `server.snapshots[]` | 72 |
| 줄 판정 `server.checks[]` | 26 |
| JFR 메서드 구간 `java.spans[]` | 1,592 |
| JFR 예외 `java.throws[]` | 62 |
| JFR 잠듦 `java.parks[]` | 4 |
| 멈춤 자리 `checkpoints[]` | 68 |
| 막힘 관찰 `blocked[]` | 6 |
| 앱 오류 메모 `notes[]` | 12 |
| 관찰 `captures[]` | 74 |
| 합 | 3,702 |

이 합에 넣지 않은 것과 그 까닭입니다.

- `transactions[].xids`, `clog`: 서버 사건의 사본입니다.
- `borrows`, `view.*`: 도출입니다.
- 연결 24는 장소이고, `accept`·`close` 두 시각을 사건으로 펴면 48행이 늘어납니다.
- 관찰 안의 행(`pages.acct[]` 236, `activity[]` 104, `locks[]` 144, `visible.acct[]` 148)은 관찰 한 번의 값입니다. 행마다 사건으로 펴면 632행이 늘어납니다.
- 판정 경로(`checks[].path[]` 78, `lines[]` 264)와 스택 프레임(사건 334개에 4,660프레임)은 사건의 부속입니다.

해석기가 만든 행은 호출 트리 마디 1,388, 효과 224, 갈림 1,008(길 1,428), 연결 묶임 70, 슬롯 20, 기다림 2, 기다림 고리 6, 빌림 24입니다. 소스 발췌는 PostgreSQL 1,434개(30,114줄), 자바 라이브러리 882개(13,220줄), 앱 파일 22개(1,394줄)입니다 [셈: keytree].

## 3. E-04 화면에 박힌 해석 결과

### 3.1 위치

원시 `.bin`은 지워졌고, 해석 결과는 화면 HTML 안의 JavaScript literal로만 남아 있습니다. 바이트 위치는 literal이 시작하는 곳의 UTF-8 오프셋입니다 [셈: `extract.py`].

| 파일 | 변수 | 행 | 바이트 위치 | literal 길이(B) | 파일 크기(B) | 문서 |
| --- | --- | --- | --- | --- | --- | --- |
| `concept-3.html` | `DATA` | 103 | 6,412 | 495,693 | 522,422 | 2(update, select) |
| `concept-4.html` | `QUERIES` | 151 | 11,738 | 5,258,851 | 5,328,161 | 11 |
| `concept-5-tx.html` | `RUNS` | 88 | 7,236 | 1,516,108 | 1,635,753 | 10 |
| `concept-6-locks.html` | `RUNS` | 88 | 7,236 | 2,518,127 | 2,637,772 | 16 |
| `concept-7-durability.html` | `RUNS` | 88 | 7,236 | 2,062,536 | 2,182,181 | 10 |
| `concept-8-serializable.html` | `RUNS` | 88 | 7,236 | 1,882,340 | 2,001,985 | 8 |
| `concept-9-spring.html` | `RUNS` | 129 | 10,359 | 5,190,265 | 5,281,445 | 14 |

모든 파일은 `workbench/design/`에 있습니다. concept-5~8은 `RUNS` literal을 `__RUNS__`로 되돌리면 `tx-view.tmpl.html`과 같고, concept-9는 `spring-view.tmpl.html`과 같습니다 [셈: `stats4.py`]. 그래서 템플릿의 행 번호를 그대로 씁니다. concept-3의 `DATA`는 맨 위 키(`update`, `select`)에 따옴표가 없어 JSON이 아닙니다. 따옴표를 넣어야 읽힙니다(`extract.py` 26~29행) [셈].

`workbench/design/concept-2.html`(53,563 B)도 같은 폴더에 있지만, 받은 범위(concept-3~9)에 없어 열지 않았습니다 [확인 안 함].

### 3.2 concept-9

2절과 같습니다. 따로 변환하지 않습니다.

### 3.3 concept-5~8(트랜잭션 실험, tx-view)

#### 3.3.1 위치와 수

| 파일 | 실행 키 |
| --- | --- |
| concept-5-tx | 1-rc-read, 1-rc-write, 1-rr-read, 1-rr-write, 1-ru-read (각 a, b) |
| concept-6-locks | 2-deadlock, 2-fk-delete, 2-fk-nokey, 2-for-update, 2-idem-commit, 2-idem-onconflict, 2-idem-rollback, 2-lost-update (각 a, b) |
| concept-7-durability | 3-crash-sync-off-wait, 3-crash-sync-off-wait2, 3-crash-sync-off, 3-crash-sync-on, 3-rollback (각 a, b) |
| concept-8-serializable | 4-skew-rr, 4-skew-ssi-row, 4-skew-ssi, 4-ssi-own (각 a, b) |

화면별 수입니다 [셈: `stats.py`].

| 항목 | concept-5 | concept-6 | concept-7 | concept-8 | 합 |
| --- | --- | --- | --- | --- | --- |
| 실행 | 10 | 16 | 10 | 8 | 44 |
| 단계 `steps[]`(= 요청, 단계마다 정확히 1개) | 68 | 104 | 48 | 72 | 292 |
| 요청 사건 `requests[].events[]` | 196 | 390 | 162 | 346 | 1,094 |
| 스냅숏 | 68 | 128 | 76 | 32 | 304 |
| 줄 판정 | 76 | 76 | 42 | 72 | 266 |
| 받은 메시지 `in`(요청마다 1) | 68 | 104 | 48 | 72 | 292 |
| 보낸 메시지 `out[]` | 162 | 204 | 107 | 178 | 651 |
| 서버 사건 `server.events[]` | 94 | 209 | 903 | 395 | 1,601 |
| 서버 프로세스 `server.procs[]` | 120 | 192 | 200 | 96 | 608 |
| 서버 함수 구간 `server.spans[]` | 0 | 0 | 224 | 0 | 224 |
| 관찰 `captures[]` | 78 | 120 | 66 | 80 | 344 |
| 힌트 변화 설명 `hintChanges[]` | 44 | 52 | 28 | 48 | 172 |
| 구경꾼 `bystanders[]` | 10 | 16 | 68 | 16 | 110 |
| kill `kills[]` | 0 | 0 | 8 | 0 | 8 |
| 짝 맞춤 `joins[]` | 316 | 592 | 498 | 416 | 1,822 |

합은 `keytree2.py`의 트리 합(`e04_tx_tree.txt`)과 모두 맞습니다(요청 사건 1,094, 서버 사건 1,601, 관찰 344, 프로세스 608) [셈]. 그 밖에 `pauses` 6, `tupleAddrs` 96, `sessions` 88, `tables` 48, 소스 발췌 `src` 6,018개(126,298줄)입니다.

#### 3.3.2 시각 체계

시각은 모두 첫 단계를 보낸 순간(`steps[0].sentNs`)부터의 마이크로초입니다(`decode/scenario.go` 321행 주석, 439~440행) [코드]. spring과 달리 `ns == 0`을 0으로 바꾸는 예외가 없습니다. 서버 사건 가운데 `f.clock == "ftrace"`인 것은 함수 기록기와 맞추지 못해 ftrace 시계를 중앙값 차이만큼 옮긴 값입니다(`decode/scenario_server.go`) [코드]. 이런 사건은 concept-7에 237건(903 가운데), concept-8에 304건(395 가운데)이고 concept-5, 6에는 0건입니다 [셈: `stats2.py`]. 종류로는 `bg_flush` 425/431, `wal_flush` 116/118입니다 [셈: `dyn.py`]. 같은 시각 열에 두되 출처를 행마다 남겨야 합니다.

#### 3.3.3 키 트리와 표시

44개 실행을 합친 333개 경로입니다 [셈: `keytree2.py` → `marks.py tx`].

```text
$  n=44  obj:44 len[min=16,max=16,sum=704]  | 모음 | 관찰 | G9
$.bystanders  n=44  list:44 len[min=1,max=9,sum=110]  | 모음 | 도출 | T2+G9
$.bystanders[]  n=110  obj:110 len[min=7,max=7,sum=770]  | 사건 | 도출 | T2+G8
$.bystanders[].firstSql  n=110  str:110  | 값 | 도출 | T2+G9
$.bystanders[].hints  n=110  int:110  | 값 | 도출 | T2+G9
$.bystanders[].outsideHints  n=110  int:110  | 값 | 도출 | T2+G9
$.bystanders[].pid  n=110  int:110  | 장소 | 도출 | T2+G4
$.bystanders[].queries  n=110  int:110  | 값 | 도출 | T2+G9
$.bystanders[].watchedChecks  n=110  int:110  | 값 | 도출 | T2+G9
$.bystanders[].watchedHints  n=110  int:110  | 값 | 도출 | T2+G9
$.captures  n=44  list:44 len[min=6,max=10,sum=344]  | 모음 | 관찰 | G9
$.captures[]  n=344  obj:344 len[min=8,max=10,sum=2906]  | 사건 | 관찰 | G8
$.captures[].activity  n=344  list:336,null:8 len[min=1,max=2,sum=626]  | 모음 | 관찰 | G9
$.captures[].activity[]  n=626  obj:626 len[min=6,max=6,sum=3756]  | 값 | 관찰 | G5d
$.captures[].activity[].pid  n=626  int:626  | 장소 | 관찰 | G4
$.captures[].activity[].state  n=626  str:626  | 값 | 관찰 | G9
$.captures[].activity[].waitEvent  n=626  str:626  | 값 | 관찰 | G9
$.captures[].activity[].waitType  n=626  str:626  | 값 | 관찰 | G9
$.captures[].activity[].xid  n=626  str:626  | 장소 | 관찰 | G4
$.captures[].activity[].xmin  n=626  str:626  | 값 | 관찰 | G9
$.captures[].afterStep  n=344  int:344  | 연결 | 관찰 | T8
$.captures[].locks  n=344  null:158,list:186 len[min=1,max=16,sum=992]  | 모음 | 관찰 | G9
$.captures[].pages  n=344  obj:344 len[min=0,max=2,sum=348]  | 모음 | 관찰 | G9
$.captures[].pages.acct  n=204  list:204 len[min=1,max=5,sum=498]  | 모음 | 관찰 | G9
$.captures[].pages.acct[]  n=498  obj:498 len[min=8,max=8,sum=3984]  | 값 | 관찰 | G5d
$.captures[].pages.acct[].block  n=498  int:498  | 장소 | 관찰 | G4
$.captures[].pages.acct[].ctid  n=498  str:498  | 장소 | 관찰 | G4
$.captures[].pages.acct[].flags  n=498  list:498 len[min=2,max=7,sum=1618]  | 모음 | 관찰 | G9
$.captures[].pages.acct[].flags[]  n=1618  str:1618  | 값 | 관찰 | G8
$.captures[].pages.acct[].lp  n=498  int:498  | 장소 | 관찰 | G4
$.captures[].pages.acct[].lpFlags  n=498  int:498  | 값 | 관찰 | G9
$.captures[].pages.acct[].values  n=498  obj:498 len[min=3,max=3,sum=1494]  | 모음 | 관찰 | G9
$.captures[].pages.acct[].values.balance  n=498  str:498  | 값 | 관찰 | G9
$.captures[].pages.acct[].values.id  n=498  str:498  | 장소 | 관찰 | G4
$.captures[].pages.acct[].values.owner  n=498  str:498  | 값 | 관찰 | G9
$.captures[].pages.acct[].xmax  n=498  str:498  | 값 | 관찰 | G9
$.captures[].pages.acct[].xmin  n=498  str:498  | 값 | 관찰 | G9
$.captures[].reads  n=344  obj:344 len[min=4,max=5,sum=1708]  | 시각 | 관찰 | T7
$.captures[].reads.activity  n=344  list:344 len[min=2,max=2,sum=688]  | 시각 | 관찰 | T7
$.captures[].reads.activity[]  n=688  float:688  | 시각 | 관찰 | T7
$.captures[].reads.locks  n=344  list:344 len[min=2,max=2,sum=688]  | 시각 | 관찰 | T7
$.captures[].reads.locks[]  n=688  float:688  | 시각 | 관찰 | T7
$.captures[].reads.pages  n=344  list:344 len[min=2,max=2,sum=688]  | 시각 | 관찰 | T7
$.captures[].reads.pages[]  n=688  float:688  | 시각 | 관찰 | T7
$.captures[].reads.running  n=344  list:344 len[min=2,max=2,sum=688]  | 시각 | 관찰 | T7
$.captures[].reads.running[]  n=688  float:688  | 시각 | 관찰 | T7
$.captures[].reads.xacts  n=332  list:332 len[min=2,max=2,sum=664]  | 시각 | 관찰 | T7
$.captures[].reads.xacts[]  n=664  float:664  | 시각 | 관찰 | T7
$.captures[].running  n=344  str:344  | 값 | 관찰 | G9
$.captures[].t  n=344  float:344  | 시각 | 관찰 | G3
$.captures[].xacts  n=344  obj:344 len[min=0,max=3,sum=640] (dynamic keys)  | 모음 | 관찰 | G9
$.captures[].xacts.{*}  n=640  str:640  | 값 | 관찰 | G9
$.captures[].locks[]  n=992  obj:992 len[min=9,max=9,sum=8928]  | 값 | 관찰 | G5d
$.captures[].locks[].granted  n=992  bool:992  | 값 | 관찰 | G9
$.captures[].locks[].mode  n=992  str:992  | 값 | 관찰 | G9
$.captures[].locks[].page  n=992  str:992  | 장소 | 관찰 | G4
$.captures[].locks[].pid  n=992  int:992  | 장소 | 관찰 | G4
$.captures[].locks[].relation  n=992  str:992  | 장소 | 관찰 | G4
$.captures[].locks[].tuple  n=992  str:992  | 장소 | 관찰 | G4
$.captures[].locks[].type  n=992  str:992  | 값 | 관찰 | G5b
$.captures[].locks[].vxid  n=992  str:992  | 장소 | 관찰 | G4
$.captures[].locks[].xid  n=992  str:992  | 장소 | 관찰 | G4
$.clock  n=44  obj:44 len[min=4,max=4,sum=176]  | 출처 | 도출 | T5
$.clock.events  n=44  int:44  | 출처 | 도출 | T5
$.clock.joined  n=44  int:44  | 출처 | 도출 | T5
$.clock.maxDeviationUs  n=44  float:44  | 출처 | 도출 | T5
$.clock.medianOffsetUs  n=44  float:44  | 출처 | 도출 | T5
$.hintChanges  n=44  list:44 len[min=1,max=6,sum=172]  | 모음 | 도출 | T2+G9
$.hintChanges[]  n=172  obj:172 len[min=8,max=8,sum=1376]  | 사건 | 도출 | T2+G8
$.hintChanges[].at  n=172  str:172  | 장소 | 도출 | T2+G4
$.hintChanges[].bit  n=172  str:172  | 값 | 도출 | T2+G9
$.hintChanges[].by  n=172  str:172  | 연결 | 도출 | T2+G5
$.hintChanges[].ctid  n=172  str:172  | 장소 | 도출 | T2+G4
$.hintChanges[].step  n=172  int:172  | 연결 | 도출 | T2+G5
$.hintChanges[].t  n=172  float:172  | 시각 | 도출 | T2+G3
$.hintChanges[].table  n=172  str:172  | 장소 | 도출 | T2+G4
$.hintChanges[].via  n=172  str:172  | 출처 | 도출 | T2+G7
$.joins  n=44  list:44 len[min=27,max=55,sum=1822]  | 출처 | 도출 | T5
$.joins[]  n=1822  obj:1822 len[min=4,max=5,sum=8068]  | 출처 | 도출 | T5
$.joins[].fn  n=1822  str:1822  | 출처 | 도출 | T5
$.joins[].pid  n=1822  int:1822  | 출처 | 도출 | T5
$.joins[].recorder  n=1822  int:1822  | 출처 | 도출 | T5
$.joins[].uprobe  n=1822  int:1822  | 출처 | 도출 | T5
$.joins[].return  n=780  bool:780  | 출처 | 도출 | T5
$.kills  n=44  list:44 len[min=0,max=1,sum=8]  | 모음 | 관찰 | G9
$.pauses  n=44  list:44 len[min=0,max=1,sum=6]  | 모음 | 관찰 | G9
$.scenario  n=44  str:44  | 값 | 관찰 | T14
$.server  n=44  obj:44 len[min=3,max=3,sum=132]  | 모음 | 관찰 | G9
$.server.events  n=44  list:44 len[min=8,max=155,sum=1601]  | 모음 | 관찰 | G9
$.server.events[]  n=1601  obj:1601 len[min=4,max=6,sum=7812]  | 사건 | 관찰 | G8
$.server.events[].f  n=1601  obj:1601 len[min=1,max=7,sum=5330]  | 모음 | 관찰 | G9
$.server.events[].f.call  n=466  str:466  | 값 | 관찰 | G9
$.server.events[].f.durUs  n=466  str:466  | 시각 | 관찰 | G3
$.server.events[].f.file  n=466  str:466  | 장소 | 관찰 | G4
$.server.events[].f.len  n=374  str:374  | 값 | 관찰 | G9
$.server.events[].f.lsn0  n=218  str:218  | 값 | 관찰 | G9
$.server.events[].f.lsn1  n=218  str:218  | 값 | 관찰 | G9
$.server.events[].f.off  n=218  str:218  | 장소 | 관찰 | G4
$.server.events[].kind  n=1601  str:1601  | 사건 | 관찰 | G6
$.server.events[].pid  n=1601  int:1601  | 장소 | 관찰 | G4
$.server.events[].stack  n=942  list:942 len[min=2,max=14,sum=5101]  | 장소 | 관찰 | G2
$.server.events[].stack[]  n=5101  obj:5101 len[min=2,max=2,sum=10202]  | 장소 | 관찰 | G2
$.server.events[].stack[].at  n=5101  str:5101  | 장소 | 관찰 | G2
$.server.events[].stack[].fn  n=5101  str:5101  | 장소 | 관찰 | G2
$.server.events[].t  n=1601  float:1600,int:1  | 시각 | 관찰 | G3
$.server.events[].t1  n=466  float:466  | 시각 | 관찰 | G3
$.server.procs  n=44  list:44 len[min=12,max=22,sum=608]  | 모음 | 관찰 | G9
$.server.procs[]  n=608  obj:608 len[min=5,max=6,sum=3128]  | 장소 | 관찰 | G5c
$.server.procs[].exit  n=608  str:608  | 값 | 관찰 | G9
$.server.procs[].pid  n=608  int:608  | 장소 | 관찰 | G4
$.server.procs[].role  n=608  str:608  | 이름 | 도출 | T6
$.server.procs[].t0  n=608  float:608  | 시각 | 관찰 | G3
$.server.procs[].t1  n=608  float:608  | 시각 | 관찰 | G3
$.server.procs[].session  n=88  str:88  | 장소 | 관찰 | G4
$.server.spans  n=44  list:44 len[min=0,max=28,sum=224]  | 모음 | 관찰 | G9
$.sessions  n=44  list:44 len[min=2,max=2,sum=88]  | 모음 | 관찰 | G9
$.sessions[]  n=88  obj:88 len[min=2,max=2,sum=176]  | 장소 | 관찰 | G5c
$.sessions[].name  n=88  str:88  | 장소 | 관찰 | G4
$.sessions[].pid  n=88  int:88  | 장소 | 관찰 | G4
$.siteRepair  n=44  obj:44 len[min=4,max=5,sum=200]  | 출처 | 도출 | T5
$.siteRepair.before  n=44  obj:44 len[min=1,max=1,sum=44] (dynamic keys)  | 출처 | 도출 | T5
$.siteRepair.before.{*}  n=44  int:44  | 출처 | 도출 | T5
$.siteRepair.leftAfter  n=44  int:44  | 출처 | 도출 | T5
$.siteRepair.repaired  n=44  int:44  | 출처 | 도출 | T5
$.siteRepair.unchanged  n=44  int:44  | 출처 | 도출 | T5
$.src  n=44  obj:44 len[min=91,max=185,sum=6018] (dynamic keys)  | 발췌 | 관찰 | G1
$.src.{*}  n=6018  obj:6018 len[min=2,max=2,sum=12036]  | 발췌 | 관찰 | G1
$.src.{*}.from  n=6018  int:6018  | 발췌 | 관찰 | G1
$.src.{*}.lines  n=6018  list:6018 len[min=14,max=21,sum=126298]  | 발췌 | 관찰 | G1
$.src.{*}.lines[]  n=126298  str:126298  | 발췌 | 관찰 | G1
$.steps  n=44  list:44 len[min=4,max=9,sum=292]  | 모음 | 관찰 | G9
$.steps[]  n=292  obj:292 len[min=7,max=9,sum=2132]  | 사건 | 관찰 | G8
$.steps[].n  n=292  int:292  | 연결 | 관찰 | T13
$.steps[].output  n=292  list:292 len[min=1,max=7,sum=706]  | 모음 | 관찰 | T12
$.steps[].output[]  n=706  str:706  | 값 | 관찰 | T12
$.steps[].requests  n=292  list:292 len[min=1,max=1,sum=292]  | 모음 | 관찰 | G9
$.steps[].requests[]  n=292  obj:292 len[min=9,max=10,sum=2640]  | 사건 | 관찰 | G8
$.steps[].requests[].catalogSnapshots  n=292  int:292  | 값 | 관찰 | G9
$.steps[].requests[].checks  n=292  list:292 len[min=0,max=4,sum=266]  | 모음 | 관찰 | G9
$.steps[].requests[].events  n=292  list:292 len[min=0,max=20,sum=1094]  | 모음 | 관찰 | G9
$.steps[].requests[].in  n=292  list:292 len[min=1,max=1,sum=292]  | 모음 | 관찰 | T10
$.steps[].requests[].in[]  n=292  obj:292 len[min=3,max=3,sum=876]  | 사건 | 관찰 | T10
$.steps[].requests[].in[].len  n=292  int:292  | 값 | 관찰 | G9
$.steps[].requests[].in[].text  n=292  str:292  | 값 | 관찰 | G9
$.steps[].requests[].in[].type  n=292  str:292  | 사건 | 관찰 | G6
$.steps[].requests[].out  n=292  list:227,null:65 len[min=2,max=6,sum=651]  | 모음 | 관찰 | T10
$.steps[].requests[].out[]  n=651  obj:651 len[min=3,max=3,sum=1953]  | 사건 | 관찰 | T10
$.steps[].requests[].out[].len  n=651  int:651  | 값 | 관찰 | G9
$.steps[].requests[].out[].text  n=651  str:651  | 값 | 관찰 | G9
$.steps[].requests[].out[].type  n=651  str:651  | 사건 | 관찰 | G6
$.steps[].requests[].snapshots  n=292  list:292 len[min=0,max=4,sum=304]  | 모음 | 관찰 | G9
$.steps[].requests[].sql  n=292  str:292  | 값 | 관찰 | G9
$.steps[].requests[].t0  n=292  float:292  | 시각 | 도출 | T9
$.steps[].requests[].t1  n=292  float:292  | 시각 | 도출 | T9
$.steps[].session  n=292  str:292  | 장소 | 관찰 | G4
$.steps[].sql  n=292  str:292  | 값 | 관찰 | G9
$.steps[].t0  n=292  int:45,float:247  | 시각 | 관찰 | G3
$.steps[].t1  n=292  float:292  | 시각 | 관찰 | G3
$.steps[].check  n=70  obj:70 len[min=7,max=7,sum=490]  | 모음 | 도출 | T1
$.steps[].check.checks  n=70  int:70  | 값 | 도출 | T1
$.steps[].check.countMatch  n=70  bool:70  | 값 | 도출 | T1
$.steps[].check.lpNormal  n=70  int:70  | 값 | 도출 | T1
$.steps[].check.result  n=70  list:70 len[min=1,max=3,sum=120]  | 모음 | 도출 | T1
$.steps[].check.result[]  n=120  str:120  | 값 | 도출 | T1
$.steps[].check.setMatch  n=70  bool:70  | 값 | 도출 | T1
$.steps[].check.table  n=70  str:70  | 값 | 도출 | T1
$.steps[].check.visible  n=70  list:70 len[min=1,max=3,sum=120]  | 모음 | 도출 | T1
$.steps[].check.visible[]  n=120  str:120  | 값 | 도출 | T1
$.steps[].requests[].checks[]  n=266  obj:266 len[min=12,max=13,sum=3382]  | 사건 | 관찰 | G8
$.steps[].requests[].checks[].ctid  n=266  str:266  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].lines  n=266  list:266 len[min=2,max=13,sum=2238]  | 모음 | 관찰 | G9
$.steps[].requests[].checks[].lines[]  n=2238  obj:2238 len[min=2,max=2,sum=4476]  | 사건 | 관찰 | G8
$.steps[].requests[].checks[].lines[].at  n=2238  str:2238  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].lines[].fn  n=2238  str:2238  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].mask  n=266  int:266  | 값 | 관찰 | G9
$.steps[].requests[].checks[].path  n=266  list:262,null:4 len[min=1,max=4,sum=638]  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].path[]  n=638  obj:638 len[min=3,max=5,sum=2878]  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].path[].arg  n=638  str:638  | 값 | 관찰 | G9
$.steps[].requests[].checks[].path[].at  n=638  str:638  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].path[].fn  n=638  str:638  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].path[].inner  n=482  list:482 len[min=1,max=2,sum=504]  | 모음 | 관찰 | G9
$.steps[].requests[].checks[].path[].inner[]  n=504  obj:504 len[min=2,max=2,sum=1008]  | 사건 | 관찰 | G8
$.steps[].requests[].checks[].path[].inner[].at  n=504  str:504  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].path[].inner[].fn  n=504  str:504  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].path[].ret  n=482  str:482  | 값 | 관찰 | G9
$.steps[].requests[].checks[].snapXmax  n=266  int:266  | 값 | 관찰 | G9
$.steps[].requests[].checks[].snapXmin  n=266  int:266  | 값 | 관찰 | G9
$.steps[].requests[].checks[].stack  n=266  list:266 len[min=19,max=37,sum=5760]  | 장소 | 관찰 | G2
$.steps[].requests[].checks[].stack[]  n=5760  obj:5760 len[min=2,max=2,sum=11520]  | 장소 | 관찰 | G2
$.steps[].requests[].checks[].stack[].at  n=5760  str:5760  | 장소 | 관찰 | G2
$.steps[].requests[].checks[].stack[].fn  n=5760  str:5760  | 장소 | 관찰 | G2
$.steps[].requests[].checks[].t  n=266  float:266  | 시각 | 관찰 | G3
$.steps[].requests[].checks[].table  n=266  str:266  | 장소 | 관찰 | G4
$.steps[].requests[].checks[].visible  n=266  bool:266  | 값 | 도출 | T11
$.steps[].requests[].checks[].xmax  n=266  int:266  | 값 | 관찰 | G9
$.steps[].requests[].checks[].xmin  n=266  int:266  | 값 | 관찰 | G9
$.steps[].requests[].events[]  n=1094  obj:1094 len[min=2,max=5,sum=4618]  | 사건 | 관찰 | G8
$.steps[].requests[].events[].f  n=1066  obj:1066 len[min=1,max=7,sum=3390] (dynamic keys)  | 모음 | 관찰 | G9
$.steps[].requests[].events[].f.{*}  n=3390  str:3390  | 값 | 관찰 | T15
$.steps[].requests[].events[].kind  n=1094  str:1094  | 사건 | 관찰 | G6
$.steps[].requests[].events[].stack  n=828  list:828 len[min=4,max=33,sum=12920]  | 장소 | 관찰 | G2
$.steps[].requests[].events[].stack[]  n=12920  obj:12920 len[min=2,max=2,sum=25840]  | 장소 | 관찰 | G2
$.steps[].requests[].events[].stack[].at  n=12920  str:12920  | 장소 | 관찰 | G2
$.steps[].requests[].events[].stack[].fn  n=12920  str:12920  | 장소 | 관찰 | G2
$.steps[].requests[].events[].t  n=1094  float:1094  | 시각 | 관찰 | G3
$.steps[].requests[].events[].t1  n=536  float:536  | 시각 | 관찰 | G3
$.steps[].requests[].snapshots[]  n=304  obj:304 len[min=7,max=7,sum=2128]  | 사건 | 관찰 | G8
$.steps[].requests[].snapshots[].curcid  n=304  int:304  | 값 | 관찰 | G9
$.steps[].requests[].snapshots[].kind  n=304  str:304  | 사건 | 관찰 | G6
$.steps[].requests[].snapshots[].stack  n=304  list:304 len[min=3,max=18,sum=1262]  | 장소 | 관찰 | G2
$.steps[].requests[].snapshots[].stack[]  n=1262  obj:1262 len[min=2,max=2,sum=2524]  | 장소 | 관찰 | G2
$.steps[].requests[].snapshots[].stack[].at  n=1262  str:1262  | 장소 | 관찰 | G2
$.steps[].requests[].snapshots[].stack[].fn  n=1262  str:1262  | 장소 | 관찰 | G2
$.steps[].requests[].snapshots[].t  n=304  float:304  | 시각 | 관찰 | G3
$.steps[].requests[].snapshots[].xip  n=304  null:304  | 모음 | 관찰 | G9
$.steps[].requests[].snapshots[].xmax  n=304  int:304  | 값 | 관찰 | G9
$.steps[].requests[].snapshots[].xmin  n=304  int:304  | 값 | 관찰 | G9
$.tables  n=44  list:44 len[min=1,max=2,sum=48]  | 모음 | 관찰 | G9
$.tables[]  n=48  obj:48 len[min=3,max=4,sum=152]  | 장소 | 관찰 | G5c
$.tables[].columns  n=48  list:48 len[min=2,max=3,sum=122]  | 모음 | 관찰 | G9
$.tables[].columns[]  n=122  obj:122 len[min=2,max=2,sum=244]  | 장소 | 관찰 | G5c
$.tables[].columns[].name  n=122  str:122  | 장소 | 관찰 | G4
$.tables[].columns[].type  n=122  str:122  | 값 | 관찰 | G5b
$.tables[].name  n=48  str:48  | 장소 | 관찰 | G4
$.tables[].oid  n=48  int:48  | 값 | 관찰 | G9
$.title  n=44  str:44  | 값 | 관찰 | T14
$.tupleAddrs  n=44  list:44 len[min=0,max=4,sum=96]  | 장소 | 도출 | T4
$.steps[].blocked  n=18  float:18  | 값 | 관찰 | T12
$.server.events[].f.level  n=136  str:136  | 값 | 관찰 | G9
$.server.events[].f.msg  n=136  str:136  | 값 | 관찰 | G9
$.siteRepair.leftFns  n=24  obj:24 len[min=1,max=2,sum=30]  | 출처 | 도출 | T5
$.siteRepair.leftFns.exec_simple_query  n=12  int:12  | 출처 | 도출 | T5
$.steps[].requests[].aborted  n=12  bool:12  | 값 | 관찰 | G9
$.siteRepair.leftFns.LockAcquireExtended  n=2  int:2  | 출처 | 도출 | T5
$.steps[].requests[].checks[].tup  n=190  str:190  | 장소 | 관찰 | G4
$.tupleAddrs[]  n=96  obj:96 len[min=3,max=4,sum=308]  | 장소 | 도출 | T4
$.tupleAddrs[].addrs  n=96  list:96 len[min=1,max=1,sum=96]  | 장소 | 도출 | T4
$.tupleAddrs[].addrs[]  n=96  str:96  | 장소 | 도출 | T4
$.tupleAddrs[].ctid  n=96  str:96  | 장소 | 도출 | T4
$.tupleAddrs[].table  n=96  str:96  | 장소 | 도출 | T4
$.captures[].pages.parent  n=24  list:24 len[min=1,max=2,sum=30]  | 연결 | 관찰 | G5
$.captures[].pages.parent[]  n=30  obj:30 len[min=8,max=8,sum=240]  | 연결 | 관찰 | G5
$.captures[].pages.parent[].block  n=30  int:30  | 장소 | 관찰 | G4
$.captures[].pages.parent[].ctid  n=30  str:30  | 장소 | 관찰 | G4
$.captures[].pages.parent[].flags  n=30  list:30 len[min=2,max=6,sum=116]  | 모음 | 관찰 | G9
$.captures[].pages.parent[].flags[]  n=116  str:116  | 값 | 관찰 | G8
$.captures[].pages.parent[].lp  n=30  int:30  | 장소 | 관찰 | G4
$.captures[].pages.parent[].lpFlags  n=30  int:30  | 값 | 관찰 | G9
$.captures[].pages.parent[].values  n=30  obj:30 len[min=2,max=2,sum=60]  | 모음 | 관찰 | G9
$.captures[].pages.parent[].values.id  n=30  str:30  | 장소 | 관찰 | G4
$.captures[].pages.parent[].values.name  n=30  str:30  | 장소 | 관찰 | G4
$.captures[].pages.parent[].xmax  n=30  str:30  | 값 | 관찰 | G9
$.captures[].pages.parent[].xmin  n=30  str:30  | 값 | 관찰 | G9
$.captures[].pages.child  n=16  list:16 len[min=1,max=1,sum=16]  | 모음 | 관찰 | G9
$.captures[].pages.child[]  n=16  obj:16 len[min=8,max=8,sum=128]  | 값 | 관찰 | G5d
$.captures[].pages.child[].block  n=16  int:16  | 장소 | 관찰 | G4
$.captures[].pages.child[].ctid  n=16  str:16  | 장소 | 관찰 | G4
$.captures[].pages.child[].flags  n=16  list:16 len[min=1,max=3,sum=24]  | 모음 | 관찰 | G9
$.captures[].pages.child[].flags[]  n=24  str:24  | 값 | 관찰 | G8
$.captures[].pages.child[].lp  n=16  int:16  | 장소 | 관찰 | G4
$.captures[].pages.child[].lpFlags  n=16  int:16  | 값 | 관찰 | G9
$.captures[].pages.child[].values  n=16  obj:16 len[min=2,max=2,sum=32]  | 모음 | 관찰 | G9
$.captures[].pages.child[].values.id  n=16  str:16  | 장소 | 관찰 | G4
$.captures[].pages.child[].values.parent_id  n=16  str:16  | 값 | 관찰 | G9
$.captures[].pages.child[].xmax  n=16  str:16  | 값 | 관찰 | G9
$.captures[].pages.child[].xmin  n=16  str:16  | 값 | 관찰 | G9
$.pauses[]  n=6  obj:6 len[min=3,max=3,sum=18]  | 사건 | 관찰 | G8
$.pauses[].beforeStep  n=6  int:6  | 연결 | 관찰 | G5
$.pauses[].t  n=6  float:6  | 시각 | 관찰 | G3
$.pauses[].t1  n=6  float:6  | 시각 | 관찰 | G3
$.captures[].pages.payment  n=24  list:24 len[min=1,max=2,sum=36]  | 모음 | 관찰 | G9
$.captures[].pages.payment[]  n=36  obj:36 len[min=8,max=8,sum=288]  | 값 | 관찰 | G5d
$.captures[].pages.payment[].block  n=36  int:36  | 장소 | 관찰 | G4
$.captures[].pages.payment[].ctid  n=36  str:36  | 장소 | 관찰 | G4
$.captures[].pages.payment[].flags  n=36  list:36 len[min=2,max=3,sum=88]  | 모음 | 관찰 | G9
$.captures[].pages.payment[].flags[]  n=88  str:88  | 값 | 관찰 | G8
$.captures[].pages.payment[].lp  n=36  int:36  | 장소 | 관찰 | G4
$.captures[].pages.payment[].lpFlags  n=36  int:36  | 값 | 관찰 | G9
$.captures[].pages.payment[].values  n=36  obj:36 len[min=2,max=2,sum=72]  | 모음 | 관찰 | G9
$.captures[].pages.payment[].values.amount  n=36  str:36  | 값 | 관찰 | G9
$.captures[].pages.payment[].values.idem_key  n=36  str:36  | 값 | 관찰 | G9
$.captures[].pages.payment[].xmax  n=36  str:36  | 값 | 관찰 | G9
$.captures[].pages.payment[].xmin  n=36  str:36  | 값 | 관찰 | G9
$.captures[].wal  n=146  obj:146 len[min=3,max=3,sum=438]  | 모음 | 관찰 | G9
$.captures[].wal.flush  n=146  str:146  | 값 | 관찰 | G9
$.captures[].wal.insert  n=146  str:146  | 값 | 관찰 | G9
$.captures[].wal.write  n=146  str:146  | 값 | 관찰 | G9
$.captures[].afterKill  n=8  bool:8  | 연결 | 관찰 | T8
$.kills[]  n=8  obj:8 len[min=5,max=8,sum=58]  | 사건 | 관찰 | G8
$.kills[].beforeStep  n=8  int:8  | 연결 | 관찰 | G5
$.kills[].pid  n=8  int:8  | 장소 | 관찰 | G4
$.kills[].session  n=8  str:8  | 장소 | 관찰 | G4
$.kills[].t  n=8  float:8  | 시각 | 관찰 | G3
$.kills[].tReady  n=8  float:8  | 시각 | 관찰 | G3
$.kills[].walSynced  n=6  str:6  | 값 | 관찰 | G9
$.kills[].walWritten  n=6  str:6  | 값 | 관찰 | G9
$.kills[].walWrittenT  n=6  float:6  | 시각 | 관찰 | G3
$.server.events[].f.clock  n=541  str:541  | 출처 | 도출 | T16
$.server.events[].f.ret  n=629  str:629  | 값 | 관찰 | G9
$.server.events[].f.lsn  n=278  str:278  | 값 | 관찰 | G9
$.server.events[].f.how  n=88  str:88  | 값 | 관찰 | G9
$.server.events[].f.sig  n=198  str:198  | 값 | 관찰 | G9
$.server.events[].f.to  n=198  str:198  | 장소 | 관찰 | G4
$.server.events[].f.group  n=72  str:72  | 값 | 관찰 | G9
$.server.events[].f.end  n=156  str:156  | 값 | 관찰 | G9
$.server.events[].f.kind  n=156  str:156  | 값 | 관찰 | G5b
$.server.events[].f.rmgr  n=156  str:156  | 값 | 관찰 | G9
$.server.events[].f.xid  n=160  str:160  | 장소 | 관찰 | G4
$.server.spans[]  n=224  obj:224 len[min=5,max=5,sum=1120]  | 사건 | 관찰 | G8
$.server.spans[].at  n=224  str:224  | 장소 | 관찰 | G4
$.server.spans[].fn  n=224  str:224  | 장소 | 관찰 | G4
$.server.spans[].pid  n=224  int:224  | 장소 | 관찰 | G4
$.server.spans[].t0  n=224  float:224  | 시각 | 관찰 | G3
$.server.spans[].t1  n=224  float:224  | 시각 | 관찰 | G3
$.siteRepair.leftFns.XLogBackgroundFlush  n=16  int:16  | 출처 | 도출 | T5
$.tupleAddrs[].epoch  n=20  int:20  | 값 | 도출 | T4a
$.captures[].pages.doctors  n=80  list:80 len[min=2,max=4,sum=232]  | 모음 | 관찰 | G9
$.captures[].pages.doctors[]  n=232  obj:232 len[min=8,max=8,sum=1856]  | 값 | 관찰 | G5d
$.captures[].pages.doctors[].block  n=232  int:232  | 장소 | 관찰 | G4
$.captures[].pages.doctors[].ctid  n=232  str:232  | 장소 | 관찰 | G4
$.captures[].pages.doctors[].flags  n=232  list:232 len[min=2,max=5,sum=760]  | 모음 | 관찰 | G9
$.captures[].pages.doctors[].flags[]  n=760  str:760  | 값 | 관찰 | G8
$.captures[].pages.doctors[].lp  n=232  int:232  | 장소 | 관찰 | G4
$.captures[].pages.doctors[].lpFlags  n=232  int:232  | 값 | 관찰 | G9
$.captures[].pages.doctors[].values  n=232  obj:232 len[min=2,max=2,sum=464]  | 모음 | 관찰 | G9
$.captures[].pages.doctors[].values.name  n=232  str:232  | 장소 | 관찰 | G4
$.captures[].pages.doctors[].values.on_call  n=232  str:232  | 값 | 관찰 | G9
$.captures[].pages.doctors[].xmax  n=232  str:232  | 값 | 관찰 | G9
$.captures[].pages.doctors[].xmin  n=232  str:232  | 값 | 관찰 | G9
$.tables[].indexes  n=8  list:8 len[min=1,max=1,sum=8]  | 모음 | 관찰 | G9
$.tables[].indexes[]  n=8  obj:8 len[min=2,max=2,sum=16]  | 장소 | 관찰 | G5c
$.tables[].indexes[].name  n=8  str:8  | 장소 | 관찰 | G4
$.tables[].indexes[].oid  n=8  int:8  | 값 | 관찰 | G9
# role x provenance line counts: {('값', '관찰'): 87, ('장소', '관찰'): 78, ('모음', '관찰'): 44, ('시각', '관찰'): 30, ('출처', '도출'): 24, ('사건', '관찰'): 19, ('값', '도출'): 16, ('장소', '도출'): 10, ('연결', '관찰'): 7, ('모음', '도출'): 5, ('발췌', '관찰'): 5, ('시각', '도출'): 3, ('사건', '도출'): 2, ('연결', '도출'): 2, ('이름', '도출'): 1}
```

#### 3.3.4 동적 키와 사건 종류별 속성

요청 사건의 `f`는 키 트리에서 `{*}`로 접혔습니다. 종류별 실제 키입니다. 값은 3,390개 모두 문자열입니다 [셈: `dyn.py`, keytree].

| `kind` | 수 | `f` 키(나온 횟수) |
| --- | --- | --- |
| lock | 314 | ctid:12, mode:314, rel:214, result:314, type:314, xid:100 |
| io | 124 | call:124, durUs:124, file:124, len:62, lsn0:62, lsn1:62, off:62 |
| xid | 84 | xid:84 |
| pred_lock | 80 | blk:24, created:80, ctid:40, gran:80, rel:80, txid:40 |
| update_check | 72 | off:72, result:72, xmax:72, xmin:72 |
| commit_clog | 68 | async:6, lsn:6, xid:68 |
| hint_fpi | 38 | buf:38, lsn:38 |
| conflict_target | 36 | blk:12, ctid:12, flagged:36, gran:36, rel:36 |
| sleep_syscall | 24 | durUs:24 |
| ssi_precommit | 24 | flags:12, sxact:12 |
| kill | 22 | args:22, ret:22 |
| wal_flush | 22 | lsn:22 |
| async_lsn | 20 | lsn:20 |
| xact_wait | 18 | oper:18, xid:18 |
| proc_sleep | 18 | ret:18 |
| wake | 16 | status:16, who:16 |
| abort_clog | 16 | xid:16 |
| hint | 16 | at:16, bit:16, ctid:16, table:16, tup:16, xid:16 |
| conflict_in | 16 | blk:16, rel:16 |
| ereport | 12 | at:12 |
| tuple_lock | 12 | mode:12, off:12, result:12, table:12, wait:12 |
| ssi_begin | 12 | (없음) |
| rw_conflict | 8 | rflags:8, rpid:8, rxid:8, wflags:8, wpid:8, wxid:8 |
| alarm | 6 | (없음) |
| deadlock_check | 6 | result:4 |
| ssi_clear | 6 | (없음) |
| epq | 2 | (없음) |
| multixact | 2 | id:2, st1:2, st2:2, xid1:2, xid2:2 |

화면별 종류 분포는 다음과 같습니다 [셈: `stats.py`].

- concept-5: lock 68, io 36, update_check 20, xid 20, commit_clog 18, kill 6, sleep_syscall 6, proc_sleep 4, wake 4, xact_wait 4, abort_clog 2, alarm 2, deadlock_check 2, epq 2, ereport 2
- concept-6: lock 134, io 44, update_check 34, xid 30, commit_clog 22, sleep_syscall 18, hint 16, kill 16, proc_sleep 14, xact_wait 14, tuple_lock 12, wake 12, abort_clog 8, ereport 6, alarm 4, deadlock_check 4, multixact 2
- concept-7: lock 56, hint_fpi 22, io 20, xid 18, async_lsn 16, commit_clog 16, wal_flush 10, abort_clog 2, update_check 2
- concept-8: pred_lock 80, lock 56, conflict_target 36, io 24, ssi_precommit 24, conflict_in 16, hint_fpi 16, update_check 16, xid 16, commit_clog 12, ssi_begin 12, wal_flush 12, rw_conflict 8, ssi_clear 6, abort_clog 4, async_lsn 4, ereport 4

서버 사건(`server.events[].f`)입니다 [셈: `dyn.py`].

| `kind` | 수 | `f` 키 |
| --- | --- | --- |
| wal_io | 466 | call:466, durUs:466, file:466, len:218, lsn0:218, lsn1:218, off:218 |
| bg_flush | 431 | clock:425, ret:431 |
| signal | 198 | group:72, ret:198, sig:198, to:198 |
| redo | 156 | end:156, kind:156, len:156, lsn:156, rmgr:156, xid:156 |
| log | 136 | level:136, msg:136 |
| wal_flush | 118 | clock:116, lsn:118 |
| exit | 88 | how:88 |
| async_lsn | 4 | lsn:4 |
| commit_clog | 4 | xid:4 |

화면별: concept-5 wal_io 90, log 4 / concept-6 wal_io 187, log 22 / concept-7 signal 198, redo 156, bg_flush 127, wal_flush 118, wal_io 114, log 94, exit 88, async_lsn 4, commit_clog 4 / concept-8 bg_flush 304, wal_io 75, log 16 [셈].

서버 프로세스 역할(`server.procs[].role`)은 io worker 156, backend 88, autovacuum launcher 52, bgwriter 52, checkpointer 52, logical replication launcher 52, observer 52, walwriter 52, postmaster 44, startup 8입니다 [셈]. 역할은 기록이 아니라 그 프로세스의 main 함수로 붙인 분류입니다(`scenario_server.go`) [코드].

서버 함수 구간(`server.spans[].fn`, concept-7만): HandleChildCrash 80, StartChildProcess 64, 그리고 CheckPointBuffers, CreateCheckPoint, CreateSharedMemoryAndSemaphores, HandleFatalError, PerformWalRecovery, ProcessSyncRequests, RequestCheckpoint, StartupXLOG, SyncDataDirectory, TerminateChildren 각 8입니다 [셈].

그 밖의 동적 키 [셈: `dyn.py`].

- `captures[].xacts.{*}`: 키 xid, 값 committed 492, in progress 116, aborted 32
- `siteRepair.before.{*}`: `0xfffffffff000` 하나, 69,992
- `captures[].pages`의 표: acct 204, doctors 80, payment 24, parent 24, child 16(그 표를 담은 관찰 수)

#### 3.3.5 기존 화면이 보여 주는 사실

화면은 `workbench/design/tx-view.tmpl.html`(1,189행)입니다. 머리 주석(84~87행)은 "화면은 계산을 하지 않고 고르고 그리기만" 한다고 하고, 예외를 넷 적습니다 [코드]. 버전 막대의 위치(xmin·xmax), 기다림 길이(ProcSleep 입구·출구 시각의 차), WAL 막대의 위치(LSN을 바이트 수로 읽음), 서버가 죽고 다시 뜨는 장면의 시간 축, 이 장면에서 사라진 SIRead 락입니다.

| 화면 사실 | 함수(행) | 읽는 키 |
| --- | --- | --- |
| 장면 준비: 프레임 나누기(깨어남·끊김), 단계마다 쓸 관찰 고르기, clog 사건으로 `xacts` 고치기, kill로 시대(epoch) 나누기 | `build` 125~186 | `steps`, `captures`, `requests[].events`, `kills` |
| 배경 쓰기(bgwriter·walwriter) | `bgWrites` 193 | `server.events` |
| 트랜잭션 사건 찾기, SSI 시작 여부 | `txEvent` 219, `ssiBegun` 234 | `events[].kind` |
| 사건 → 화면 글(종류마다) | `hopsOf` 242~341 | `requests[].events[]`, `f` |
| 머리 | `header` 347 | `title`, `scenario` |
| psql 터미널(프롬프트를 출력에서 만듦) | `term` 357 | `steps[].output`, `sql` |
| SSI 상태와 상자 | `ssiState` 395, `ssiBox` 417 | `pred_lock`, `conflict_*`, `rw_conflict`, `ssi_*` |
| 행 버전의 xmin·xmax 글 | `xminText` 447, `xmaxText` 450 | `captures[].pages.*[]` |
| 다중 트랜잭션 잠금 | `mxUpTo` 458, `mxById` 462 | `multixact` |
| 미니맵 | `minimap` 465~553 | 여러 키 |
| 세션 칸(기다림 띠, 멈춤) | `lanes` 557~642 | `steps`, `proc_sleep`, `wake`, `pauses` |
| 서버 충돌 장면 | `crashMarks` 656, `crashLayout` 659, `crashSvg` 686, `crashDetail` 1067 | `server.procs`, `server.events`, `server.spans`, `kills`, `captures[].wal` |
| 남의 기다림 | `otherWait` 772 | `xact_wait`, `proc_sleep` |
| 판정에 쓴 스냅숏 찾기 | `snapOf` 781 | `snapshots[]`, `checks[].snapXmin/snapXmax` |
| 자기 xid 찾기 | `ownXid` 793 | `xid` 사건, `sql`의 BEGIN/COMMIT/ROLLBACK |
| SSI 절 | `ssiSection` 807 | SSI 사건들 |
| 상세(배치 이상 검사, 판정 검사 배지, 줄 판정 경로 칩) | `detail` 879~1050 | `steps[].check`, `checks[].path`, `lines` |

화면이 스스로 셈하는 것 [코드]:

- 머리 주석의 넷
- `build`의 프레임 나누기, 관찰 고르기(새는 관찰을 피해 앞으로 물러남), clog로 `xacts` 고치기
- `snapOf`의 xip 되찾기(2.6절 불일치 4)
- `ownXid`는 SQL 글을 정규식(`/^BEGIN/i`, `/^(COMMIT|ROLLBACK)/i`)으로 읽습니다
- `ssiSection` 866~869행은 "그 사이 다른 세션의 커밋이 켰다"를 추론해 글로 씁니다(주석 "추론해 적습니다")

화면이 자체로 들고 있는 사실 [셈: `consts.py`, 코드]:

| 상수 | 행 | 내용 |
| --- | --- | --- |
| `LABEL` | 89~93 | 장면 22개의 이름(예: "RC · 두 번 읽기") |
| `TITLE` | 94 | 실험 4개의 제목 |
| `MODE`, `MXST`, `HINT`, `OPER` | 103~107 | 잠금 모드, 다중 트랜잭션 상태, 힌트 비트, 기다림 연산의 사람 말 |
| `ROLE`, `SPANKO` | 648~652 | 프로세스 역할 10개, 함수 구간 9개의 사람 말 |
| `MASK` | 907~908 | 힌트 비트의 숫자 값 6개(예: `HEAP_XMIN_COMMITTED: 0x0100`) |
| `PART` | 962 | 관찰 부분 3개의 이름 |
| `LOGKO` | 1054~1066 | 서버 로그 정규식 13개와, 그 로그가 어느 함수에서 나왔는지 적은 설명(예: "process_pm_child_exit → CleanupBackend") |

`LOGKO`와 `MASK`는 소스에 대한 주장이라 단순한 표시 문구가 아닙니다. 공통 형식에서 note나 사전으로 옮기지 않으면 다시 뽑을 때 빠집니다.

#### 3.3.6 모호함·불일치·사실처럼 저장된 추론

1. `hintChanges[].via`는 concept-5에서 `value` 44건(값으로 맞춘 설명), concept-6~8에서 `addr` 128건(주소로 맞춘 설명)입니다 [셈: `stats2.py`]. `value`는 같은 xmin·xmax의 다른 행(카탈로그 행 등)에 쓴 것도 설명으로 잡을 수 있다고 해석기가 스스로 적습니다(`scenario_check.go` 200~212행) [코드]. `by`(쓴 세션)는 추론입니다.
2. `steps[].check`(70)는 해석기의 자기 점검입니다. 저장된 데이터 안에서 이 점검이 실패한 자리가 넷 있습니다. 모두 concept-6입니다 [셈: `stats4.py`].
    - 2-deadlock/a, b 단계 9: `countMatch: false`(`checks: 4`, `lpNormal: 5`, `result`·`visible` 모두 `(0,5)`, `(0,4)`, `setMatch: true`)
    - 2-for-update/a, b 단계 4: `setMatch: false`(`result` `(0,2)`, `visible` `(0,1)`, `countMatch: true`)

    판정 결과와 해석기가 셈한 보일 행이 어긋난 것입니다. 화면 `detail`이 이것을 배지로 보입니다 [코드].
3. `deadlock_check` 6건 가운데 2건은 `f.result`가 없습니다 [셈: `dyn.py`].
4. `server.procs[].role`은 main 함수로 붙인 분류입니다(3.3.4절).
5. `bystanders[]`(110)는 관찰 대상이 아닌 백엔드의 활동을 해석기가 요약한 것입니다 [코드].
6. `siteRepair.leftAfter`는 실행마다 다릅니다 [셈: `stats4.py`].
   - concept-5: 0 또는 1
   - concept-6: 0~2
   - concept-7: 0, 2, 4, 50, 70
   - concept-8: 76~82

   `leftFns`에는 `exec_simple_query`(12개 실행), `LockAcquireExtended`(2), `XLogBackgroundFlush`(16)가 나옵니다 [셈: keytree]. 이 함수들의 `site`는 믿을 수 없습니다.
7. 타입 불일치는 E-03과 같습니다. `checks[].xmin/xmax`, `snapshots[].xmin/xmax`는 정수이고, `captures[].pages.*[].xmin/xmax`, `activity[].xid`, `locks[].xid`, `f.xid`는 문자열입니다 [셈: keytree].
8. `snapshots[].xip`는 304건 모두 `null`입니다 [셈: `stats4.py`]. 2.6절 불일치 4와 같은 문제입니다.
9. 요청 `t0`, `t1`은 서버 쪽 `exec_simple_query` 창으로 정해집니다(`scenario.go` 488~502행) [코드]. 단계 `t0`, `t1`(txlab의 `sentNs`, `doneNs`)과 다른 시각입니다. 요청 12건은 `aborted`이고, 단계 18건은 `blocked`입니다 [셈: keytree].
10. concept-7의 관찰 66건 모두 `wal`을 담고, 8건은 `afterKill`입니다. 같은 단계 번호 뒤에 관찰이 둘(단계 뒤, 재시작 뒤)이 되므로(`scenario_check.go` 213~214행 주석), `afterStep`만으로는 관찰을 하나로 정하지 못합니다 [코드, 셈].

#### 3.3.7 규모

tx의 "기록기가 남긴 사건 행"은 다음 합으로 정의합니다 [셈].

| 표 | 수 |
| --- | --- |
| 단계(txlab이 보내고 받은 시각) | 292 |
| 받은 메시지 `in` | 292 |
| 보낸 메시지 `out[]` | 651 |
| 요청 사건 | 1,094 |
| 스냅숏 | 304 |
| 줄 판정 | 266 |
| 서버 사건 | 1,601 |
| 서버 함수 구간 | 224 |
| 관찰 | 344 |
| kill | 8 |
| 멈춤 `pauses[]` | 6 |
| 합 | 5,082 |

이 합에 넣지 않은 것과 그 까닭입니다.

- 서버 프로세스 608, 세션 88, 표 48은 장소입니다.
- `hintChanges` 172, `bystanders` 110, `steps[].check` 70, `tupleAddrs` 96은 도출입니다.
- 관찰 안의 행은 값입니다: `pages.*[]` 812, `activity[]` 626, `locks[]` 992.

원시 규모를 보이는 값으로, `clock.events` 합은 concept-5 78,348, concept-6 121,036, concept-7 172,325, concept-8 76,662이고 `joined` 합은 각각 19,830, 26,592, 12,308, 14,296입니다 [셈: `stats4.py`].

### 3.4 concept-3(질의 하나가 지나는 이동)

#### 3.4.1 위치와 수

`DATA = {update: {...}, select: {...}}`, 103행, 바이트 6,412부터 495,693 B입니다(3.1절). 질의 2개가 있습니다. `update`(pid 5531, `UPDATE t SET name='z' WHERE id=2`, `endUs` 82,826)와 `select`(pid 5534, `endUs` 34,255)입니다 [셈].

| 항목 | update | select | 합 |
| --- | --- | --- | --- |
| 이동 `hops[]` 행 | 97 | 10 | 107 |
| 합친 횟수(`count` 합) | 287 | 132 | 419 |
| 합친 이동(`count > 1`, 그래서 `t0 ≠ t1`) | 44 | 4 | 48 |
| 프로토콜 메시지 `msgs[]` | 5 | 13 | 18 |
| 스택 프레임 | 2,954 | 152 | 3,106 |
| 단계 `stages[]` | 7 | 7 | 14 |
| 소스 발췌 `src{}` | 237 | 100 | 337 (7,074줄) |

이동 종류(`count`로 가중) [셈: `stats.py`]: 버퍼 찾기 359, 페이지 읽기 카탈로그 40, 응답 버퍼에 쓰기 8, WAL 기록 3, 요청 받음 2, 응답 보냄 2, WAL 쓰기 1, 디스크까지 동기화 1, 읽기 맡기기 1, 페이지 읽기(I/O 워커) 테이블·인덱스 1, 행 바꾸기 1. 단계별 행: 의미 분석 48, 계획 27, 실행 21, 응답 6, 커밋 3, 받기 2.

#### 3.4.2 시각 체계

`t0`, `t1`은 그 질의를 받은 순간(`recv.Ns`)부터의 마이크로초입니다(`decode/journey.go` 579행, 885행) [코드]. 이동은 대부분 한 순간이라 `t0 == t1`이고, 연이은 같은 종류를 합친 48행만 `t0 ≠ t1`입니다(`count` 1보다 큼) [셈]. 화면은 시각 옆에 "기록 중이라 실제보다 느림"이라고 적습니다(253행) [코드].

#### 3.4.3 키 트리와 표시

41개 경로입니다 [셈: `keytree2.py` → `marks.py c3`].

```text
$  n=2  obj:2 len[min=6,max=6,sum=12]  | 모음 | 관찰 | G9
$.endUs  n=2  int:2  | 시각 | 도출 | C20
$.hops  n=2  list:2 len[min=10,max=97,sum=107]  | 모음 | 관찰 | G9
$.hops[]  n=107  obj:107 len[min=12,max=15,sum=1466]  | 사건 | 관찰 | G8
$.hops[].n  n=107  int:107  | 연결 | 도출 | C6
$.hops[].pid  n=107  int:107  | 장소 | 관찰 | G4
$.hops[].stage  n=107  str:107  | 값 | 도출 | C3
$.hops[].proc  n=107  str:107  | 장소 | 도출 | C8
$.hops[].from  n=107  str:107  | 장소 | 도출 | C7
$.hops[].to  n=107  str:107  | 장소 | 도출 | C7
$.hops[].kind  n=107  str:107  | 사건 | 이름 | C2
$.hops[].count  n=107  int:107  | 값 | 도출 | C1+G9
$.hops[].bytes  n=46  int:46  | 값 | 관찰 | G9
$.hops[].msgs  n=10  list:10 len[min=1,max=6,sum=18]  | 모음 | 관찰 | G9
$.hops[].msgs[]  n=18  obj:18 len[min=3,max=3,sum=54]  | 사건 | 관찰 | G8
$.hops[].msgs[].type  n=18  str:18  | 사건 | 관찰 | G6
$.hops[].msgs[].len  n=18  int:18  | 값 | 관찰 | G9
$.hops[].msgs[].text  n=18  str:18  | 값 | 관찰 | G9
$.hops[].t0  n=107  int:47,float:60  | 시각 | 관찰 | G3
$.hops[].t1  n=107  int:47,float:60  | 시각 | 도출 | C1+G3
$.hops[].stack  n=107  list:107 len[min=3,max=45,sum=3106]  | 장소 | 관찰 | G2
$.hops[].stack[]  n=3106  obj:3106 len[min=2,max=2,sum=6212]  | 장소 | 관찰 | G2
$.hops[].stack[].fn  n=3106  str:3106  | 장소 | 관찰 | G2
$.hops[].stack[].at  n=3106  str:3106  | 장소 | 관찰 | G2
$.hops[].site  n=107  str:107  | 장소 | 관찰 | G4
$.hops[].catalog  n=83  bool:83  | 값 | 도출 | C3
$.hops[].files  n=43  list:43 len[min=1,max=1,sum=43]  | 장소 | 관찰 | G4
$.hops[].files[]  n=43  str:43  | 장소 | 관찰 | G4
$.pid  n=2  int:2  | 장소 | 관찰 | G4
$.query  n=2  str:2  | 값 | 관찰 | C19
$.src  n=2  obj:2 len[min=100,max=237,sum=337] (dynamic keys)  | 발췌 | 관찰 | G1
$.src.{*}  n=337  obj:337 len[min=2,max=2,sum=674]  | 발췌 | 관찰 | G1
$.src.{*}.from  n=337  int:337  | 발췌 | 관찰 | G1
$.src.{*}.lines  n=337  list:337 len[min=18,max=21,sum=7074]  | 발췌 | 관찰 | G1
$.src.{*}.lines[]  n=7074  str:7074  | 발췌 | 관찰 | G1
$.stages  n=2  list:2 len[min=7,max=7,sum=14]  | 모음 | 관찰 | G9
$.stages[]  n=14  obj:14 len[min=2,max=4,sum=52]  | 사건 | 도출 | C5
$.stages[].name  n=14  str:14  | 이름 | 도출 | C4
$.stages[].t0  n=14  int:2,float:12  | 시각 | 관찰 | G3
$.stages[].fn  n=12  str:12  | 장소 | 관찰 | G4
$.stages[].at  n=12  str:12  | 장소 | 관찰 | G4
# role x provenance line counts: {('장소', '관찰'): 11, ('발췌', '관찰'): 5, ('모음', '관찰'): 4, ('값', '관찰'): 4, ('사건', '관찰'): 3, ('값', '도출'): 3, ('장소', '도출'): 3, ('시각', '도출'): 2, ('시각', '관찰'): 2, ('연결', '도출'): 1, ('사건', '이름'): 1, ('사건', '도출'): 1, ('이름', '도출'): 1}
```

#### 3.4.4 기존 화면이 보여 주는 사실

화면 코드는 `concept-3.html`의 104행부터입니다(데이터 103행 제외) [코드].

| 화면 사실 | 함수(행) |
| --- | --- |
| 줄 만들기(단계 줄, 카탈로그 읽기 접기) | `buildRows` 127 |
| 이동 이름 | `hopLabel` 146, `caller` 157 |
| 칸 머리 | `laneHeads` 159 |
| 그림 전체, 시각 글 | `draw` 186, 253행 |
| psql 결과 표(T·D·C 메시지로 표를 다시 그림) | `terminal` 259 |
| 메시지 덩어리 | `msgBlocks` 281 |
| 상세(파일 → 관계 이름, 송신 버퍼 사용량 셈) | `detail` 290, `relName` 121 |
| 질의 고르기, 이동 | `setQ` 354, `go` 350 |

화면이 자체로 들고 있는 사실 [셈: `consts.py`]:

- `REL`(104행): 파일 경로 → 관계 이름 22개. 이동의 `files`에 나오는 데이터 파일 23개(WAL 파일 1개 포함)를 빠짐없이 덮고, 남는 항목도 없습니다 [셈: `stats4.py`]. 기록에 없는 사람 손 자료입니다.
- `LANES`(105~114행): 칸 8개(client, comm_buffer, backend, shared_buffer, wal_buffer, io_worker, data_files, wal_files)와 각 칸의 설명(예: "8 KB 송·수신")
- `MSGCOLOR`(117행): 메시지 색 6개
- `SYSKINDS`(118행): 시스템 콜에서 온 이동 종류 7개

#### 3.4.5 모호함·불일치·사실처럼 저장된 추론

1. `hops[].kind`는 코드가 아니라 사람 말입니다. 게다가 분류 결과를 글에 섞습니다. `"페이지 읽기 " + relKind(path)`처럼 파일 번호 16,384 미만이면 "카탈로그", 그 이상이면 "테이블·인덱스"를 붙입니다(`journey.go` 835행, 852행 `relKind`) [코드]. 공통 형식에서 `kind`를 키로 쓰려면 글을 나눠야 합니다.
2. `catalog`(83행에 있음)는 `kind`에 "카탈로그"가 있거나, 스택이 카탈로그 캐시·`systable_*`·`catalog/indexing.c`를 지나면 참입니다(`journey.go` 656~668행) [코드].
3. `count`, `t1`은 합친 결과라서, 사건 수를 셀 때 행 수(107)와 횟수(419) 가운데 무엇을 쓸지 정해야 합니다.
4. `stages[].name`은 함수로 붙인 단계 이름입니다(`stageNames`) [코드].
5. `msgs[]`는 버퍼 쓰기와 소켓 송신을 맞춰 나눈 결과입니다(`journey.go` 338행 `parseSend`, 681행에서 부름) [코드].
6. `proc`가 `other`이면 그 pid의 역할은 기록이 모르며, 화면이 기록할 때 본 ps 역할표로 붙입니다(`journey.go` 812~814행) [코드].

### 3.5 concept-4(질의 11개, 커널 포함)

#### 3.5.1 위치와 수

`QUERIES` 151행, 바이트 11,738부터 5,258,851 B입니다. 원소는 `{key, label, typed, x, data, procs, procsGuess}`입니다 [셈].

| key | typed | 이동 행 | 합친 횟수 | 합친 이동 | 메시지 | 스택 프레임 | 커널 마디 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| update | `UPDATE t SET name='z' WHERE id=2;` | 158 | 373 | 93 | 5 | 4,484 | 0 |
| select | `SELECT * FROM t;` | 48 | 180 | 36 | 13 | 1,184 | 0 |
| create_index | `CREATE INDEX t_name_idx ON t(name);` | 329 | 1,041 | 216 | 5 | 8,030 | 0 |
| show | `SHOW work_mem;` | 7 | 7 | 0 | 9 | 53 | 0 |
| show_error | `SHOW status;` | 7 | 7 | 0 | 5 | 51 | 0 |
| insert | `INSERT INTO t VALUES (4,'d');` | 48 | 117 | 20 | 5 | 941 | 0 |
| big | `SELECT length(repeat('x', 20000)), repeat('x', 20000) AS big \gx` | 36 | 86 | 15 | 10 | 875 | 0 |
| bind | `SELECT $1::int + 1 AS answer \bind 41 \g` | 29 | 81 | 14 | 17 | 602 | 0 |
| kernel_insert | `INSERT INTO t VALUES (4,'d');` | 42 | 89 | 17 | 5 | 901 | 176 |
| kernel_index | `CREATE INDEX t_name_idx ON t(name);` | 341 | 972 | 210 | 5 | 8,471 | 193 |
| kernel_insert2 | `INSERT INTO t VALUES (4,'d');` | 42 | 89 | 17 | 5 | 901 | 189 |
| 합 | | 1,087 | 3,042 | 638 | 84 | 26,493 | 558 |

그 밖의 합 [셈: `stats.py`, `stats2.py`, keytree]:

- 단계 76
- 소스 발췌 `src` 2,039개(42,797줄)
- 커널 소스 `ksrc` 3개 질의에 708개(36,813줄)
- `sysUs`가 있는 이동 71(kernel_insert 10, kernel_index 51, kernel_insert2 10)
- 커널 뿌리 8개가 이동 7개에 달림
- `hidden`(2 µs 미만으로 쳐낸 자식 수)이 0이 아닌 커널 마디 416
- `guess`가 참인 마디 31
- `calls` 후보가 있는 마디 316
- 메시지 가운데 `cut`/`cont` 3개(big)
- 스택이 `null`인 이동 5
- `x: true`는 big 하나(`\gx` 확장 표시)
- `procsGuess: true`는 kernel_* 3개, `kernelPid`는 915392(kernel_insert, kernel_index), 934054(kernel_insert2)

이동 종류(`count`로 가중) [셈: `stats.py`]: 버퍼 찾기 2,028, 락 잡기 629, 페이지 읽기 카탈로그 129, WAL 기록 49, 스냅숏 만들기 38, 응답 버퍼에 쓰기 34, 인덱스 항목 넣기 19, 응답 보냄 14, 캐시 무효화 알림 12, 커밋 상태 읽기 12, 요청 받음 11, 트랜잭션 끝 알림 11, WAL 쓰기 10, 디스크까지 동기화 7, 행 넣기 7, 커밋 상태 기록 6, 트랜잭션 번호 받기 6, 페이지 읽기 테이블·인덱스 5, 페이지 쓰기 4, 행 여러 개 넣기 4, 새 파일 2, 정렬 시작 2, 보고(ereport) 1, 읽기 맡기기 1, 행 바꾸기 1. 단계별 행: 실행 746, 의미 분석 190, 계획 62, 커밋 38, 응답 28, 받기 11, 구문 분석 7, 오류 처리 5.

#### 3.5.2 시각 체계

concept-3과 같습니다(`recv.Ns` 기준 µs). 커널 마디의 `us`는 그 함수에서 보낸 시간이고 잠든 시간을 포함합니다(`decode/kernel.go`) [코드]. 이동의 `sysUs`는 시스템 콜 시간입니다. 화면은 커널 시간이 "커널 추적 중에 잰 시간"이라 짧은 함수가 많을수록 길게 나온다고 적습니다(`KTIME_NOTE` 545행) [코드].

#### 3.5.3 키 트리와 표시

107개 경로입니다. 커널 트리의 2단 이상 재귀는 `kids[]{2+}` 한 줄로 합쳤습니다 [셈: `keytree2.py` → `marks.py c4`].

```text
$  n=11  obj:11 len[min=7,max=7,sum=77]  | 모음 | 관찰 | G9
$.key  n=11  str:11  | 값 | 이름 | C17
$.label  n=11  str:11  | 값 | 이름 | C17
$.typed  n=11  str:11  | 값 | 이름 | C17
$.x  n=11  bool:11  | 값 | 이름 | C17
$.data  n=11  obj:11 len[min=6,max=8,sum=72]  | 모음 | 관찰 | G9
$.data.endUs  n=11  int:11  | 시각 | 도출 | C20
$.data.hops  n=11  list:11 len[min=7,max=341,sum=1087]  | 모음 | 관찰 | G9
$.data.hops[]  n=1087  obj:1087 len[min=12,max=16,sum=14388]  | 사건 | 관찰 | G8
$.data.hops[].n  n=1087  int:1087  | 연결 | 도출 | C6
$.data.hops[].pid  n=1087  int:1087  | 장소 | 관찰 | G4
$.data.hops[].stage  n=1087  str:1087  | 값 | 도출 | C3
$.data.hops[].proc  n=1087  str:1087  | 장소 | 도출 | C8
$.data.hops[].from  n=1087  str:1087  | 장소 | 도출 | C7
$.data.hops[].to  n=1087  str:1087  | 장소 | 도출 | C7
$.data.hops[].kind  n=1087  str:1087  | 사건 | 이름 | C2
$.data.hops[].count  n=1087  int:1087  | 값 | 도출 | C1+G9
$.data.hops[].bytes  n=171  int:171  | 값 | 관찰 | G9
$.data.hops[].msgs  n=57  list:57 len[min=1,max=6,sum=84]  | 모음 | 관찰 | G9
$.data.hops[].msgs[]  n=84  obj:84 len[min=2,max=4,sum=250]  | 사건 | 관찰 | G8
$.data.hops[].msgs[].type  n=84  str:84  | 사건 | 관찰 | G6
$.data.hops[].msgs[].len  n=84  int:84  | 값 | 관찰 | G9
$.data.hops[].msgs[].text  n=79  str:79  | 값 | 관찰 | G9
$.data.hops[].t0  n=1087  int:180,float:907  | 시각 | 관찰 | G3
$.data.hops[].t1  n=1087  int:180,float:907  | 시각 | 도출 | C1+G3
$.data.hops[].stack  n=1087  list:1082,null:5 len[min=3,max=47,sum=26493]  | 장소 | 관찰 | G2
$.data.hops[].stack[]  n=26493  obj:26493 len[min=2,max=2,sum=52986]  | 장소 | 관찰 | G2
$.data.hops[].stack[].fn  n=26493  str:26493  | 장소 | 관찰 | G2
$.data.hops[].stack[].at  n=26493  str:26493  | 장소 | 관찰 | G2
$.data.hops[].site  n=1087  str:1087  | 장소 | 관찰 | G4
$.data.hops[].catalog  n=883  bool:883  | 값 | 도출 | C3
$.data.hops[].files  n=155  list:155 len[min=1,max=1,sum=155]  | 장소 | 관찰 | G4
$.data.hops[].files[]  n=155  str:155  | 장소 | 관찰 | G4
$.data.pid  n=11  int:11  | 장소 | 관찰 | G4
$.data.query  n=11  str:11  | 값 | 관찰 | C19
$.data.src  n=11  obj:11 len[min=38,max=359,sum=2039] (dynamic keys)  | 발췌 | 관찰 | G1
$.data.src.{*}  n=2039  obj:2039 len[min=2,max=2,sum=4078]  | 발췌 | 관찰 | G1
$.data.src.{*}.from  n=2039  int:2039  | 발췌 | 관찰 | G1
$.data.src.{*}.lines  n=2039  list:2039 len[min=13,max=21,sum=42797]  | 발췌 | 관찰 | G1
$.data.src.{*}.lines[]  n=42797  str:42797  | 발췌 | 관찰 | G1
$.data.stages  n=11  list:11 len[min=6,max=7,sum=76]  | 모음 | 관찰 | G9
$.data.stages[]  n=76  obj:76 len[min=2,max=5,sum=344]  | 사건 | 도출 | C5
$.data.stages[].name  n=76  str:76  | 이름 | 도출 | C4
$.data.stages[].t0  n=76  int:11,float:65  | 시각 | 관찰 | G3
$.data.stages[].fn  n=64  str:64  | 장소 | 관찰 | G4
$.data.stages[].caller  n=64  str:64  | 장소 | 관찰 | G4
$.data.stages[].at  n=64  str:64  | 장소 | 관찰 | G4
$.procs  n=11  obj:11 len[min=7,max=7,sum=77]  | 장소 | 관찰 | C16
$.procs.io_worker  n=11  list:11 len[min=3,max=3,sum=33]  | 장소 | 관찰 | C16
$.procs.io_worker[]  n=33  int:33  | 장소 | 관찰 | C16
$.procs.walwriter  n=11  list:11 len[min=1,max=1,sum=11]  | 장소 | 관찰 | C16
$.procs.walwriter[]  n=11  int:11  | 장소 | 관찰 | C16
$.procs.checkpointer  n=11  list:11 len[min=1,max=1,sum=11]  | 장소 | 관찰 | C16
$.procs.checkpointer[]  n=11  int:11  | 장소 | 관찰 | C16
$.procs.bgwriter  n=11  list:11 len[min=1,max=1,sum=11]  | 장소 | 관찰 | C16
$.procs.bgwriter[]  n=11  int:11  | 장소 | 관찰 | C16
$.procs.autovacuum  n=11  list:11 len[min=1,max=1,sum=11]  | 장소 | 관찰 | C16
$.procs.autovacuum[]  n=11  int:11  | 장소 | 관찰 | C16
$.procs.logrep  n=11  list:11 len[min=1,max=1,sum=11]  | 장소 | 관찰 | C16
$.procs.logrep[]  n=11  int:11  | 장소 | 관찰 | C16
$.procs.postmaster  n=11  list:11 len[min=1,max=1,sum=11]  | 장소 | 관찰 | C16
$.procs.postmaster[]  n=11  int:11  | 장소 | 관찰 | C16
$.procsGuess  n=11  bool:11  | 출처 | 도출 | C15
$.data.hops[].msgs[].cut  n=2  bool:2  | 연결 | 도출 | C9
$.data.hops[].msgs[].cont  n=1  bool:1  | 연결 | 도출 | C9
$.data.hops[].sysUs  n=71  int:71  | 시각 | 관찰 | G3
$.data.hops[].kernel  n=7  list:7 len[min=1,max=2,sum=8]  | 모음 | 관찰 | C14
$.data.hops[].kernel[]  n=8  obj:8 len[min=5,max=5,sum=40]  | 사건 | 관찰 | C14
$.data.hops[].kernel[].fn  n=8  str:8  | 장소 | 관찰 | G4
$.data.hops[].kernel[].us  n=8  float:8  | 시각 | 관찰 | G3
$.data.hops[].kernel[].layer  n=8  str:8  | 이름 | 도출 | C12
$.data.hops[].kernel[].at  n=8  str:8  | 장소 | 관찰 | G4
$.data.hops[].kernel[].kids  n=8  list:8 len[min=2,max=3,sum=21]  | 연결 | 관찰 | C13
$.data.hops[].kernel[].kids[]  n=21  obj:21 len[min=4,max=6,sum=110]  | 사건 | 관찰 | C13b
$.data.hops[].kernel[].kids[].fn  n=21  str:21  | 장소 | 관찰 | G4
$.data.hops[].kernel[].kids[].us  n=21  float:21  | 시각 | 관찰 | G3
$.data.hops[].kernel[].kids[].layer  n=21  str:21  | 이름 | 도출 | C12
$.data.hops[].kernel[].kids[].at  n=21  str:21  | 장소 | 관찰 | G4
$.data.hops[].kernel[].kids[].calls  n=5  list:5 len[min=1,max=1,sum=5]  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[].calls[]  n=5  list:5 len[min=4,max=4,sum=20]  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[].calls[][]  n=20  obj:20 len[min=3,max=3,sum=60]  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[].calls[][].fn  n=20  str:20  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[].calls[][].at  n=20  str:20  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[].calls[][].text  n=20  str:20  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[].hidden  n=13  int:13  | 출처 | 도출 | C11
$.data.hops[].kernel[].kids[].kids  n=8  list:8 len[min=1,max=1,sum=8]  | 연결 | 관찰 | C13
$.data.hops[].kernel[].kids[]{2+}  n=529  obj:529 (merged 16 levels)  | 모음 | 관찰 | C13b
$.data.hops[].kernel[].kids[]{2+}.fn  n=529  str:529 (merged 16 levels)  | 장소 | 관찰 | G4
$.data.hops[].kernel[].kids[]{2+}.us  n=529  float:509,int:20 (merged 16 levels)  | 시각 | 관찰 | G3
$.data.hops[].kernel[].kids[]{2+}.layer  n=529  str:529 (merged 16 levels)  | 이름 | 도출 | C12
$.data.hops[].kernel[].kids[]{2+}.at  n=498  str:498 (merged 16 levels)  | 장소 | 관찰 | G4
$.data.hops[].kernel[].kids[]{2+}.calls  n=311  list:311 (merged 16 levels)  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[]{2+}.calls[]  n=366  list:366 (merged 16 levels)  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[]{2+}.calls[][]  n=600  obj:600 (merged 16 levels)  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[]{2+}.calls[][].fn  n=600  str:600 (merged 16 levels)  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[]{2+}.calls[][].at  n=600  str:600 (merged 16 levels)  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[]{2+}.calls[][].text  n=600  str:600 (merged 16 levels)  | 발췌 | 도출 | C10
$.data.hops[].kernel[].kids[]{2+}.hidden  n=403  int:403 (merged 16 levels)  | 출처 | 도출 | C11
$.data.hops[].kernel[].kids[]{2+}.kids  n=368  list:368 (merged 15 levels)  | 연결 | 관찰 | C13
$.data.hops[].kernel[].kids[]{2+}.guess  n=31  bool:31 (merged 8 levels)  | 출처 | 도출 | C11
$.data.hops[].kernel[].kids[]{2+}.callsOf  n=30  int:30 (merged 2 levels)  | 발췌 | 도출 | C10
$.data.kernelPid  n=3  int:3  | 장소 | 관찰 | C18
$.data.ksrc  n=3  obj:3 len[min=227,max=249,sum=708] (dynamic keys)  | 발췌 | 관찰 | G1
$.data.ksrc.{*}  n=708  obj:708 len[min=2,max=2,sum=1416]  | 발췌 | 관찰 | G1
$.data.ksrc.{*}.from  n=708  int:708  | 발췌 | 관찰 | G1
$.data.ksrc.{*}.lines  n=708  list:708 len[min=4,max=160,sum=36813]  | 발췌 | 관찰 | G1
$.data.ksrc.{*}.lines[]  n=36813  str:36813  | 발췌 | 관찰 | G1
# role x provenance line counts: {('장소', '관찰'): 34, ('발췌', '도출'): 13, ('발췌', '관찰'): 10, ('모음', '관찰'): 7, ('시각', '관찰'): 6, ('사건', '관찰'): 5, ('값', '이름'): 4, ('값', '관찰'): 4, ('이름', '도출'): 4, ('출처', '도출'): 4, ('연결', '도출'): 3, ('값', '도출'): 3, ('장소', '도출'): 3, ('연결', '관찰'): 3, ('시각', '도출'): 2, ('사건', '이름'): 1, ('사건', '도출'): 1}
```

#### 3.5.4 기존 화면이 보여 주는 사실

| 화면 사실 | 함수(행) |
| --- | --- |
| 칸 합치기 | `lane` 168(`LANE_OF`) |
| 관계 이름(fork 접미사 포함) | `relName` 181(`REL`, `FORK`) |
| 프로세스 역할(추정 표시) | `setProcs` 226, `roleName` 231. `procsGuess`면 " (추정)"을 붙입니다 |
| 줄 만들기 | `buildRows` 239 |
| 노드 그림의 경로 | `routeOf` 259, `edgePath` 330 |
| 이동 이름 | `hopLabel` 284, `baseLabel` 287 |
| 응답 보냄의 몇 번째/몇 번 | 290행. `sendNo`, `sendOf`는 화면이 불러올 때 붙이는 값입니다(788행) |
| 미니맵(지난 길·지금·앞으로) | `minimap` 338 |
| 그림 전체와 카탈로그 요약(찾기·읽기·잠금·쓰기 합) | `draw` 402, 428행 |
| 시각 글 | 465행 |
| psql 결과(Z까지 기다림, 확장 표시, 40자 자르기) | `terminal` 472, 482행 |
| 메시지 덩어리 | `msgBlocks` 507 |
| 커널: 자기 시간, 뜨거운 마디(뿌리의 10% 이상, `KWAIT`), 사실 줄(virtio_queue_rq·filemap_dirty_folio 수), 고드름 그림, 마디 상세(`guess`·`callsOf` 추론 표시), 커널 보기 | `kSelf` 550, `kCount` 553, `kHot` 559, `kFacts` 576, `kIcicle` 586, `kCode` 608, `kNode` 617, `kernelView` 647, `bindKernel` 670 |
| 상세(송신 까닭을 스택에서, 버퍼 셈) | `detail` 685 |
| 이동, 질의 고르기 | `go` 776, `setQ` 782 |

화면이 자체로 들고 있는 사실 [셈: `consts.py`, `stats4.py`]:

- `REL`(153행): 관계 이름 38개. 데이터 파일 39개(WAL 1개 포함)를 빠짐없이 덮음
- `LANES`(156~166행): 칸 9개
- `LANE_OF`(167행): 칸 합치기 6개
- `MSG`(172~175행): 방향 2개(c, s)별 메시지 이름과 색
- `SYSKINDS`(176행): 8개
- `FORK`(180행): fork 접미사 3개
- `STAGES`(185행), `SHORT`(186행): 단계 8개와 그 짧은 이름
- `ZONES`(187~193행): 영역 5개
- `NODES`(194~215행): 노드 20개의 좌표와 역할 글
- `KLAYER`(526행), `KLAYER_ROLE`(529~539행): 커널 층 11개와 그 설명 9개
- `KWAIT`(541~542행): 잠듦 함수 4개와 뜻
- `KSRC_NOTE`(546행): 커널 소스는 원본 Linux v7.2.9이고 실행 커널은 7.2.9-orbstack이라는 출처 문구

#### 3.5.5 모호함·불일치·사실처럼 저장된 추론

1. `procsGuess: true`인 3개 질의의 `procs`는 ps가 아니라 fork 순서와 하는 일로 추정한 역할표입니다(223행 주석) [코드].
2. 커널 마디의 `layer`는 소스 디렉터리로 붙인 이름입니다. `guess`는 그 층을 부모에게서 물려받았다는 표시입니다(`kernel.go`) [코드].
3. 커널 마디의 `calls`는 이름으로 찾은 후보 줄입니다. `callsOf`는 좁히기 전 후보입니다. 둘 다 추론입니다 [코드].
4. `hidden`은 쳐낸 자식의 수일 뿐, 그 자식들은 데이터에 없습니다 [코드].
5. `hops[].kind`가 사람 말이고 분류를 섞는 문제는 concept-3과 같습니다(3.4.5절 1).
6. 스택이 `null`인 이동이 5개 있습니다 [셈: keytree].
7. `msgs[].text`는 84개 가운데 79개에만 있습니다 [셈: keytree].
8. `bytes`는 1,087행 가운데 171행에만 있습니다 [셈: keytree].

### 3.6 concept-3과 concept-4의 겹침

concept-3의 `update`, `select`와 concept-4의 `update`, `select`는 같은 기록입니다. 서로 다른 해석일 뿐입니다 [셈: `stats4.py`].

- pid가 같습니다(5531, 5534).
- `endUs`가 같습니다(82,826, 34,255).
- 질의 글이 같습니다.
- concept-3의 이동을 `(kind, t0)`로 맞추면, update 97행 가운데 96행, select 10행 모두가 concept-4에 있습니다.

다른 점은 둘입니다. 첫째, concept-4는 해석기가 더 많은 종류를 뽑았습니다. update에는 락 잡기, 스냅숏 만들기, 커밋 상태 기록·읽기, 트랜잭션 끝 알림, 트랜잭션 번호 받기, 페이지 읽기 테이블·인덱스가 더 있고, select에는 그 가운데 넷이 더 있습니다. 둘째, 같은 순간(`t0` 77,833)의 이동 하나가 concept-3에서는 "페이지 읽기(I/O 워커) 테이블·인덱스", concept-4에서는 "페이지 읽기 테이블·인덱스"로 이름이 다릅니다.

그래서 공통 형식에서는 이 두 쌍을 같은 run으로 보고, 어느 해석을 정본으로 삼을지 정해야 합니다. 질의 글로 run을 맞추면 반대 방향의 실수도 생깁니다. concept-4의 `insert`(pid 5903, `endUs` 31,580), `kernel_insert`(6421, 376,824), `kernel_insert2`(6576, 148,545)는 질의 글이 같지만 pid와 `endUs`가 다릅니다. `create_index`(5903, 217,700)와 `kernel_index`(6421, 1,213,932)도 그렇습니다 [셈: `stats4.py`]. pid 하나로도 run을 정할 수 없습니다. `insert`와 `create_index`는 백엔드 5903을, `kernel_insert`와 `kernel_index`는 백엔드 6421을 함께 씁니다. 한 접속에서 질의 두 개를 차례로 기록한 것으로 보입니다 [추론]. 그래서 concept-3/4가 같은 기록이라는 판단은 pid만이 아니라 pid, `endUs`, 질의 글, 이동의 `(kind, t0)` 일치를 함께 근거로 삼습니다.

## 4. E-05 gdb 기록 8개

### 4.1 위치와 수

`open_source_internals/database/postgresql/evidence/`의 `.trace` 8개입니다 [셈: `stats.py`, `stats3.py`].

| 파일 | STOP | 값이 있는 STOP | BT | END의 횟수 합 | `#` 설명 줄 | psql 출력 줄 | 쓴 감시 목록 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| commit-sync-off.trace | 9 | 5 | 0 | 9 | 0 | 2 | commit [추론: 함수로 판단] |
| commit-sync-on.trace | 12 | 7 | 17 | 12 | 0 | 2 | commit [추론] |
| exercise-1-select-all.trace | 31 | 10 | 0 | 31 | 0 | 6 | select [추론] |
| exercise-2-select-name.trace | 27 | 10 | 0 | 27 | 0 | 2 | select [추론] |
| exercise-3-index-scan.trace | 24 | 8 | 25 | 24 | 0 | 2 | index [추론] |
| exercise-5-select-commit.trace | 4 | 1 | 0 | 4 | 0 | 22 | commit [추론] |
| scan-end-order.trace | 30 | 10 | 0 | 30 | 4 | 2 | select.watch + 세 줄(파일 안 `#` 줄이 밝힘) |
| select-where-id-2.trace | 27 | 10 | 29 | 27 | 0 | 2 | select(재생본 `meta.watchId`) |
| 합 | 164 | 61 | 71 | 164 | 4 | 40 | |

8개 모두 END 줄이 하나이고 NOBREAK 줄이 없습니다. 함수마다 STOP 수와 END 횟수가 같습니다(차이 0) [셈: `stats3.py`]. BT 71줄은 정확히 세 묶음입니다. 각 묶음은 `workbench.json`의 `stackAt` 함수인 STOP 바로 뒤에 옵니다. commit-sync-on은 11번 `pg_fdatasync` 뒤 17줄, exercise-3은 16번 `heap_hot_search_buffer` 뒤 25줄, select-where-id-2는 18번 `PinBufferForBlock` 뒤 29줄입니다. 깊이는 0부터 빠짐없이 늘어납니다 [셈: `stats4.py`]. END에는 감시 목록에 없는 `ReadyForQuery`가 나옵니다. trace.py가 그 함수에 멈추면 기록을 끝내기 때문입니다(`scripts/trace.py` 9행, 28행의 `TRACE_END` 기본값) [코드].

### 4.2 줄 형식과 읽는 코드

형식의 정본은 trace.py의 머리 주석이고, `workbench/internal/trace/trace.go` 1~10행이 그 형식을 옮겨 적습니다 [코드].

```text
STOP <순번> <함수> <소스 상대경로>:<줄> [식=값 ...]
BT   <깊이> <함수> <소스 상대경로>:<줄>
END  <함수>=<횟수> ...
NOBREAK <함수>
--- psql            (run-trace.sh가 붙임. 아래는 psql 출력 줄)
# ...               (저장한 파일에만 있을 수 있는 설명 줄)
```

실제 첫 줄의 예입니다(select-where-id-2.trace 1행): `STOP 1 exec_simple_query src/backend/tcop/postgres.c:1014 query_string=0xaaaafe32e8e8 "SELECT * FROM t WHERE id = 2;"`.

`trace.Parse`는 감시 목록(`trace.Exprs`)을 받아야 값 부분을 식마다 나눕니다. 나누지 못하면 `Values`를 비우고 `RawValues`에 원문을 남깁니다(trace.go 40~51행) [코드]. 읽는 쪽은 셋입니다.

- `workbench/cmd/workbench/export.go`(`export`): 감시 목록을 `workbench.json`의 `-watch` id로 고르고, `trace.Parse` → `replay.Build` → `web.Standalone`으로 재생본을 만듭니다. `-sql`이 없으면 첫 단계의 `query_string`에서 SQL을 꺼냅니다(25~100행) [코드].
- `workbench/internal/replay/bundle.go`(`Build` 79행~): 위치마다 `subject.Classify`로 층(`layer`, `lane`)을 붙이고, `source.Compute`로 소스 구간(`window`)을 정하고, 그 줄을 `files`에 모읍니다 [코드].
- `workbench/internal/run/run.go`: 실시간 실행 쪽입니다. 이번에 줄 단위로 읽지 않았습니다 [확인 안 함].

### 4.3 키 트리와 표시

`trace.Record`(파싱 결과, trace.go 21~66행)는 기록 파일의 의미를 그대로 담습니다 [코드].

| 경로 | 타입 | 역할 | 출처 | 비고 |
| --- | --- | --- | --- | --- |
| `steps[]` | 객체 | 사건 | 관찰 | STOP 줄 하나 |
| `steps[].seq` | int | 연결(순서) | 관찰 | 시각이 없어 순서가 유일한 시간 정보 |
| `steps[].function` | str | 장소 | 관찰 | |
| `steps[].loc.path`, `loc.line` | str, int | 장소 | 관찰 | 위치를 모르면 `""`, 0 |
| `steps[].values[].expr`, `text` | str | 값 | 관찰(식 이름은 감시 목록) | 감시 목록이 있을 때만 |
| `steps[].rawValues` | str | 값 | 관찰 | 나누지 못했을 때 |
| `steps[].stack[].depth`, `function`, `loc` | int, str, obj | 장소 | 관찰 | BT 줄. 바로 앞 STOP에 붙음 |
| `counts[].function`, `hits` | str, int | 장소, 값 | 관찰 | END 줄 순서. 없는 함수는 0회 |
| `noBreak[]` | str | 출처 | 관찰 | 이번 8개에는 없음 |
| `output[]` | str | 값 | 관찰 | 데우기 실행과 추적 실행의 psql 출력 |

재생본에 박힌 `Bundle`은 `replays/select-where-id-2.html` 261행, 바이트 13,639부터 78,929 B입니다. 맨 위 키는 `subjectId, subjectTitle, sourceRoot, sourceNotice, lanes, meta, steps, counts, output, files`입니다 [셈: `extract5.py`]. 58개 경로입니다 [셈: `keytree.py` → `marks.py e05`].

```text
$  n=1  obj:1 len[min=10,max=10,sum=10]  | 모음 | 관찰 | G9
$.subjectId  n=1  str:1  | 값 | 이름 | X5
$.subjectTitle  n=1  str:1  | 값 | 이름 | X5
$.sourceRoot  n=1  str:1  | 값 | 이름 | X5
$.sourceNotice  n=1  str:1  | 값 | 이름 | X5
$.lanes  n=1  list:1 len[min=14,max=14,sum=14]  | 이름 | 도출 | X4
$.lanes[]  n=14  obj:14 len[min=2,max=2,sum=28]  | 이름 | 도출 | X4
$.lanes[].index  n=14  int:14  | 이름 | 도출 | X4
$.lanes[].label  n=14  str:14  | 이름 | 도출 | X4
$.meta  n=1  obj:1 len[min=5,max=5,sum=5]  | 출처 | 이름 | X7
$.meta.sql  n=1  str:1  | 값 | 도출 | X6
$.meta.watchId  n=1  str:1  | 출처 | 이름 | X7
$.meta.watchLabel  n=1  str:1  | 출처 | 이름 | X7
$.meta.origin  n=1  str:1  | 출처 | 이름 | X7
$.meta.guide  n=1  str:1  | 출처 | 이름 | X7
$.steps  n=1  list:1 len[min=27,max=27,sum=27]  | 모음 | 관찰 | G9
$.steps[]  n=27  obj:27 len[min=2,max=4,sum=65]  | 사건 | 관찰 | G8
$.steps[].seq  n=27  int:27  | 연결 | 관찰 | X8
$.steps[].place  n=27  obj:27 len[min=5,max=5,sum=135]  | 장소 | 관찰 | X12
$.steps[].place.function  n=27  str:27  | 장소 | 관찰 | G4
$.steps[].place.loc  n=27  obj:27 len[min=2,max=2,sum=54]  | 장소 | 관찰 | X10
$.steps[].place.loc.path  n=27  str:27  | 장소 | 관찰 | X10
$.steps[].place.loc.line  n=27  int:27  | 장소 | 관찰 | X10
$.steps[].place.layer  n=27  str:27  | 이름 | 도출 | X1
$.steps[].place.lane  n=27  int:27  | 이름 | 도출 | X1
$.steps[].place.window  n=27  obj:27 len[min=2,max=2,sum=54]  | 발췌 | 도출 | X2
$.steps[].place.window.defLine  n=27  int:27  | 발췌 | 도출 | X2
$.steps[].place.window.ranges  n=27  list:27 len[min=1,max=1,sum=27]  | 발췌 | 도출 | X2
$.steps[].place.window.ranges[]  n=27  obj:27 len[min=2,max=2,sum=54]  | 발췌 | 도출 | X2
$.steps[].place.window.ranges[].from  n=27  int:27  | 발췌 | 도출 | X2
$.steps[].place.window.ranges[].to  n=27  int:27  | 발췌 | 도출 | X2
$.steps[].values  n=10  list:10 len[min=1,max=2,sum=11]  | 모음 | 관찰 | G9
$.steps[].values[]  n=11  obj:11 len[min=2,max=2,sum=22]  | 사건 | 관찰 | G8
$.steps[].values[].expr  n=11  str:11  | 값 | 이름 | X11
$.steps[].values[].text  n=11  str:11  | 값 | 관찰 | G9
$.steps[].stack  n=1  list:1 len[min=29,max=29,sum=29]  | 장소 | 관찰 | G2
$.steps[].stack[]  n=29  obj:29 len[min=5,max=5,sum=145]  | 장소 | 관찰 | G2
$.steps[].stack[].function  n=29  str:29  | 장소 | 관찰 | G4
$.steps[].stack[].loc  n=29  obj:29 len[min=2,max=2,sum=58]  | 장소 | 관찰 | X10
$.steps[].stack[].loc.path  n=29  str:29  | 장소 | 관찰 | X10
$.steps[].stack[].loc.line  n=29  int:29  | 장소 | 관찰 | X10
$.steps[].stack[].layer  n=29  str:29  | 이름 | 도출 | X1
$.steps[].stack[].lane  n=29  int:29  | 이름 | 도출 | X1
$.steps[].stack[].window  n=29  obj:29 len[min=2,max=2,sum=58]  | 발췌 | 도출 | X2
$.steps[].stack[].window.defLine  n=29  int:29  | 발췌 | 도출 | X2
$.steps[].stack[].window.ranges  n=29  list:29 len[min=1,max=2,sum=44]  | 발췌 | 도출 | X2
$.steps[].stack[].window.ranges[]  n=44  obj:44 len[min=2,max=2,sum=88]  | 발췌 | 도출 | X2
$.steps[].stack[].window.ranges[].from  n=44  int:44  | 발췌 | 도출 | X2
$.steps[].stack[].window.ranges[].to  n=44  int:44  | 발췌 | 도출 | X2
$.counts  n=1  list:1 len[min=20,max=20,sum=20]  | 모음 | 관찰 | X9
$.counts[]  n=20  obj:20 len[min=2,max=2,sum=40]  | 값 | 관찰 | X9
$.counts[].function  n=20  str:20  | 장소 | 관찰 | X9a
$.counts[].hits  n=20  int:20  | 값 | 관찰 | X9
$.output  n=1  list:1 len[min=2,max=2,sum=2]  | 모음 | 관찰 | X13
$.output[]  n=2  str:2  | 값 | 관찰 | X13
$.files  n=1  obj:1 len[min=20,max=20,sum=20] (dynamic keys)  | 발췌 | 관찰 | X3
$.files.{*}  n=20  obj:20 len[min=23,max=187,sum=1576] (dynamic keys)  | 발췌 | 관찰 | X3
$.files.{*}.{*}  n=1576  str:1576  | 발췌 | 관찰 | X3
# role x provenance line counts: {('장소', '관찰'): 12, ('발췌', '도출'): 12, ('이름', '도출'): 8, ('모음', '관찰'): 5, ('값', '이름'): 5, ('출처', '이름'): 5, ('값', '관찰'): 4, ('발췌', '관찰'): 3, ('사건', '관찰'): 2, ('값', '도출'): 1, ('연결', '관찰'): 1}
```

`Bundle`의 호출 스택 칸은 `Frame.Depth`를 버리고 `Place` 배열의 순서로만 깊이를 나타냅니다(bundle.go 124~130행) [코드]. 이번 BT는 깊이가 0부터 빠짐없으므로 정보가 줄지 않습니다 [셈].

### 4.4 기존 화면이 보여 주는 사실

재생본이 있는 기록은 8개 가운데 `select-where-id-2.trace` 하나뿐입니다 [셈: `stats4.py`, `replays/`에 파일 1개]. 재생본의 화면 스크립트(262행부터 15,049자)는 `workbench/web/app.js`(427행)와 앞뒤 공백 말고는 같습니다 [셈: `stats4.py`]. 그래서 아래 행 번호는 app.js의 것입니다. HTML 골격은 `workbench/web/index.html`(94행)입니다.

| 화면 사실 | 함수(행) | 읽는 키 |
| --- | --- | --- |
| 제목, 출처 문구 | `show` 165~174 | `subjectTitle`, `sourceNotice` |
| 출처·감시 목록·기록 시각·설명 문서 링크, "이 SQL은 두 번 실행됩니다…" | `renderOrigin` 211~217 | `meta.origin`, `watchLabel`, `recordedAt`, `guide`. 마지막 문장은 고정 글 |
| SQL | `show` 178 | `meta.sql` |
| 쓰인 층만 고른 세로줄과 범례 | `usedColumns` 223, `renderLegend` 229 | `lanes`, `steps[].place.lane` |
| 단계 목록(순번, 함수, 툴팁의 층·경로:줄)과 층 사이를 잇는 꺾은선 | `renderSteps` 236~265 | `steps[].seq`, `place.function`, `place.layer`, `place.loc` |
| END 횟수 표 | `renderCounts` 275 | `counts[]` |
| psql 출력 | `renderOutput` 281 | `output[]` |
| 중단점을 못 만든 함수, 멈춘 단계가 없다는 알림 | `show` 186~191 | `noBreak`, `steps.length` |
| 현재 단계의 함수·위치·층(툴팁: 작업대가 디렉터리로 붙인 이름) | `renderCurrent` 307~332 | `place` 또는 `stack[frame]` |
| 소스 창(구간, 생략 줄 수, 정의 줄, 멈춘 줄, "(이 자료에 없는 줄)") | `renderCode` 335~361 | `place.window`, `files` |
| 멈춘 순간의 값 | `renderValues` 364 | `values[]`, `rawValues` |
| 호출 스택 칸(번호 = 배열 순서) | `renderStack` 374 | `stack[]` |

화면이 스스로 셈하는 것은 셋입니다 [코드]. 쓰인 층만 고르기(`usedColumns`), 꺾은선 좌표(`renderSteps` 250행), 소스 구간 사이의 생략 줄 수(`renderCode` 347행)입니다.

화면이 자체로 들고 있는 사실 [코드]:

- app.js 216행: "이 SQL은 두 번 실행됩니다. 첫 번째는 캐시를 데우는 실행이고, 아래 기록은 두 번째 실행의 것입니다."
- app.js 322행: 층 툴팁 "작업대가 소스 디렉터리를 보고 붙인 이름입니다. 소스가 정한 이름이 아닙니다."
- index.html의 안내 글: 점 하나가 멈춘 단계 하나, 층은 작업대의 분류, "목록에 없으면 0회", "캐시를 데우는 실행과 추적한 실행의 결과가 차례로 나옵니다", "0번 칸이 멈춘 함수"

재생본이 없는 7개는 `open_source_internals/database/postgresql/README.md`의 서술이 소비합니다. 그 서술이 지금 그 기록의 "화면"입니다 [코드].

| README 행 | 쓰는 기록 | 꺼내는 사실 |
| --- | --- | --- |
| 114 | select-where-id-2 | 정상이면 STOP 27줄, BT 29줄, END 한 줄, psql `2\|b` 두 번. NOBREAK의 뜻 |
| 162 | select-where-id-2 | BT 13·9·8·5·2·0번 칸으로 읽는 호출 경로(순차 스캔 → table AM → 힙 → read stream → 버퍼 관리자). 나머지 15줄 |
| 164 | select-where-id-2, scan-end-order | END의 `read_stream_next_buffer` 3회, `heap_fetch_next_buffer` 2회, `PinBufferForBlock` 1회. 마지막 호출이 `heap_endscan` → `read_stream_end` → `read_stream_reset` 뒤에 온다는 순서 |
| 187 | exercise-1-select-all | `ExecSeqScan` 4회, `heap_getnextslot` 4회, `printtup` 3회. 출력 3행 + NULL 1 |
| 202 | exercise-2-select-name | `ExecSeqScanWithQualProject`가 골라짐 |
| 239 | exercise-3-index-scan | 계획 맨 위 노드 `T_IndexScan`과 인덱스 경로. tid와 `bt_page_items()`의 ctid 대응(`exercise-3-index-pages.out` 함께) |
| 401 | commit-sync-on, commit-sync-off | psql `INSERT 0 1` 두 번. `XLogInsert`에 붙은 rmgr 이름. synchronous_commit에 따른 `XLogFlush`와 `XLogSetAsyncXactLSN` 갈림 |
| 491 | exercise-5-select-commit | `exec_simple_query`, `CommitTransaction`, `RecordTransactionCommit`, `ReadyForQuery`만 지났고 `XLogInsert`, `XLogFlush`, `pg_fdatasync`는 0회 |

README의 이 사실들은 셈(횟수), 순서(STOP·BT 순번), 값(`T_IndexScan`, rmgr 이름)과 "0회"(END에 없음)로 이루어집니다. 공통 형식에서 "0회"는 행이 없다는 사실이므로, 감시 목록(무엇을 지켜봤는가)을 run에 남기지 않으면 다시 뽑을 수 없습니다.

### 4.5 모호함·불일치·사실처럼 저장된 추론

1. 기록 파일은 어떤 감시 목록으로 찍었는지 적지 않습니다. 재생본은 사람이 `-watch`로 줍니다(export.go 33행) [코드]. 표의 "쓴 감시 목록"은 select-where-id-2(재생본 `meta`)와 scan-end-order(`#` 줄) 말고는 나온 함수로 판단한 추론입니다.
2. scan-end-order의 감시 목록(`select.watch` + `heap_endscan`, `read_stream_end`, `read_stream_reset`)은 파일로 남아 있지 않고 `#` 설명 줄 1~4행의 글로만 있습니다 [코드].
3. 시각이 없습니다. 순서(`seq`)만 있습니다.
4. 값 글에 실행마다 다른 포인터가 섞입니다(예: `0xaaaafe32e8e8 "SELECT …"`, `relname.data=0xffff93ad50cc "t"`) [셈]. 기록끼리 견주면 이 부분만 다릅니다.
5. 층(`layer`, `lane`)은 `workbench.json`의 `layers` 접두사로 붙인 분류입니다(`subject.Classify`). 화면도 이것을 툴팁으로 밝힙니다(app.js 322행) [코드].
6. 재생본의 `meta.sql`은 `-sql`로 받은 값 또는 첫 `query_string`에서 꺼낸 값입니다(export.go 65~72행) [코드]. 이번 재생본에는 `recordedAt`이 없습니다 [셈].
7. psql 출력에는 데우기 실행과 추적 실행의 결과가 한데 있고, 둘을 가르는 표시가 없습니다(trace.go 63행 주석) [코드]. exercise-5는 22줄이라 두 실행의 경계를 줄 수로만 알 수 있습니다 [추론].

### 4.6 규모

E-05의 사건 행은 STOP 164입니다. BT 71은 STOP 3개에 붙은 부속입니다. END 8줄은 run마다의 횟수 요약입니다(함수 항목 합 164, STOP과 같음). psql 출력 40줄과 `#` 4줄은 run의 note입니다 [셈]. 재생본 하나는 단계 27, 스택 칸 29, 값 11, END 항목 20, 출력 2, 소스 파일 20개(1,576줄), 층 14입니다 [셈: `extract5.py`].

## 5. 공통 형식 설계로 넘길 함정(입력을 가로지르는 것)

1. **시각의 기준·단위·시계가 입력마다 다릅니다.**
   - E-03: 맥 `startNs` 기준 µs. 맥·컨테이너 두 시계가 섞이고, 맥 쪽은 정수, `clock.wire`는 ms입니다.
   - tx: `steps[0].sentNs` 기준 µs. ftrace로 옮긴 사건 541건이 있습니다.
   - concept-3/4: `recv.Ns` 기준 µs.
   - E-05: 시각 없이 순서만 있습니다.

   event 표의 시각 열에는 기준(run별 원점), 시계(맥, 컨테이너, ftrace 보정, 없음), 정밀도를 함께 남겨야 합니다.
2. **같은 대상의 식별자 타입이 다릅니다.** 연결은 정수 또는 `"C1"`이고, xid는 정수 또는 문자열이고, pid는 정수 또는 문자열입니다. 변환할 때 한 타입으로 정규화하고, 원래 타입은 버려도 되는지 정해야 합니다(화면 함수가 `===`로 견주는 곳이 있습니다. 예: `snapOf`의 `x.xmin === c.snapXmin`은 둘 다 정수라 맞음).
3. **사건 종류 이름이 둘입니다.**
   - E-03, tx: `kind`가 코드 이름입니다(`lock`, `xid` …).
   - concept-3/4: `kind`가 사람 말이고, 분류(카탈로그/테이블·인덱스, I/O 워커)를 글에 섞습니다.
4. **같은 기록이 두 번 있습니다.** E-03 = concept-9(14개 모두), concept-3 update/select = concept-4 update/select(해석만 다름). 반대로 같은 SQL 글인데 다른 기록도 있습니다(insert ×3, create_index ×2).
5. **관찰과 도출이 같은 객체 안에 섞여 있습니다.** 도출은 `view.*`, `transactions`, `borrows`, `visible`, `hintChanges`, `bystanders`, `steps[].check`, `procs.role`, 커널 `guess`·`calls`이고, 끝난 꼴 `passed`·`rethrow`, 효과 `return`·`suspend`·`resume`도 도출입니다. 공통 형식에서 이것들을 관찰 사건과 같은 표에 두면, 해석기를 고칠 때마다 "차이 0"이 깨집니다. 도출은 note 또는 별도 run(해석기 버전)으로 두는 편이 비교를 안정시킵니다 [추론].
6. **사본이 있습니다.** `transactions[].xids`·`clog`는 서버 사건의 사본이고, `view.tree`는 `java.spans`에서 만든 것입니다. 둘 다 세면 같은 사실을 두 번 셉니다.
7. **합친 행이 있습니다.** concept-3/4의 `count > 1` 이동 686행(48 + 638)은 여러 사건을 한 행에 담습니다. 펼 수 없으므로(개별 시각이 없음), event 행 하나에 `count` 속성을 두는 것이 정보를 잃지 않는 유일한 길입니다 [추론].
8. **빈 값의 뜻이 여럿입니다.** 시각 0(spring), `t1 = null`(passed), `xip = null`, `conn = ""`(timeout 빌림), `stack = null`(concept-4 5행), END에 없는 함수(0회)가 있습니다. 각각 "없음", "모름", "0개" 가운데 무엇인지 형식이 구분해야 합니다.
9. **화면이 자체로 들고 있는 사실이 있습니다.** REL(22, 38), LANES, NODES(20), TOPIC.scenes(6), LABEL(22), LOGKO(13), MASK(6), KLAYER_ROLE(9), KWAIT(4), app.js 216·322행 고정 글이 그렇습니다. 데이터에 없으므로 place 표(관계 이름, 칸)나 note로 옮기지 않으면, 다시 뽑을 때 빠진 사실로 잡힙니다.
10. **화면이 스스로 셈하는 사실이 있습니다.**
    - spring-view: 상태판, `lock_timeout`, 요청 묶기, 시각 바꾸기
    - tx-view: `build`, `snapOf`, `ownXid`, DOOMED 추론 글
    - concept-4: `sendNo`/`sendOf`, 카탈로그 요약 합
    - app.js: 쓰인 층

    다시 뽑는 비교는 "데이터에서 바로 읽는 사실"과 "화면 계산을 거친 사실"을 나눠야 실패 원인을 가를 수 있습니다 [추론].
11. **원시 기록은 이미 없습니다.** `.bin`이 지워졌으므로 "빠진 사실 0"의 기준은 해석 결과 JSON(E-03, E-04)과 `.trace` 원문(E-05)입니다. uprobe 원시 사건 수(E-03 118,424, tx 448,371)는 기준이 될 수 없습니다.

## 6. 검증 경로

아래 명령을 이 문서 옆의 `t1-inventory/` 폴더에서 다시 돌리면(`data/`는 첫 명령 `extract.py`가 만듭니다) 이 문서의 모든 [셈] 숫자가 출력에 나와야 합니다. 숫자 하나라도 출력에 없거나 다르면 이 문서가 틀린 것입니다(FAIL). 경로는 `DESIGN=/Users/rody/VscodeProjects/study/open_source_internals/workbench/design`, `EV=/Users/rody/VscodeProjects/study/open_source_internals/database/postgresql/evidence`로 둡니다.

| 명령 | 확인하는 것 | PASS |
| --- | --- | --- |
| `python3 -I extract.py $DESIGN data` | 3.1절의 행·바이트 위치·길이·문서 수 | 표와 같은 수 |
| `python3 -I cmp9.py`(E-03 폴더와 `data/concept-9-spring.json`) | E-03 = concept-9 | 14개 모두 IDENTICAL |
| `python3 -I keytree.py <E-03 파일 14개>`, `python3 -I keytree2.py data/concept-*.json` | 키 트리와 `n=` | 2.3, 3.3.3, 3.4.3, 3.5.3절 트리와 같음 |
| `python3 -I marks.py <트리> e03\|tx\|c3\|c4\|e05` | 표시 | 각 절 코드 블록과 같음 |
| `python3 -I stats.py data $EV` | 2.1, 2.4, 2.7, 3.3.1, 3.3.4, 3.4.1, 3.5.1, 4.1절의 수 | 표와 같은 수 |
| `python3 -I stats2.py data` | 힌트 `via`, ftrace 시계, 판정 검사, 관찰 `wal`·`afterKill`, concept-4 `procsGuess`·`kernelPid`·`sysUs` | 3.3.2, 3.3.6, 3.5.1절과 같음 |
| `python3 -I stats3.py $EV` | STOP = END 합, BT 묶음 | 8개 모두 차이 0, BT 묶음 1·1·1 |
| `python3 -I dyn.py data` | 동적 키와 종류별 `f` 키 | 2.4, 3.3.4절 표와 같음 |
| `python3 -I consts.py $DESIGN/concept-3.html $DESIGN/concept-4.html $DESIGN/spring-view.tmpl.html $DESIGN/tx-view.tmpl.html` | 화면 자체 상수의 행과 항목 수 | 표와 같음. 단, 이 도구는 항목 수를 틀리게 낼 때가 있습니다(concept-3 `LANES` 9 → 실제 8, concept-4 `LANES` 10 → 9, `ZONES` 6 → 5, `NODES` 21 → 20, `MSG` 3 → 2, `KLAYER_ROLE` 10 → 9). 그리고 `TOPIC`의 끝을 잘못 잡습니다(898행으로 나옴, 실제 151~462행). 이 일곱은 grep으로 다시 셌고, 표에는 다시 센 값을 적었습니다 |
| `python3 -I stats4.py data /Users/rody/VscodeProjects/study` | 화면 = 템플릿, 재생본 스크립트 = app.js, E-03 파일 수·크기, `xip` 꼴, 빈 `conn` 빌림, `overlap`, `leftAfter`·`leftFns`, 세션 수, `ambiguous`, uprobe 합, `stackOnly` 사건, `lab.*` 마디, xid 사본, `passed`의 `t1`, 실패한 `steps[].check`, tx `clock.events` 합, concept-3/4 겹침, 같은 SQL의 다른 기록, REL 덮음, BT 깊이 연속성 | 2.1, 2.2, 2.6, 2.7, 3.1, 3.3.6, 3.3.7, 3.4.4, 3.5.4, 3.6, 4.1, 4.3, 4.4절과 같음 |
| `python3 -I extract5.py <replays/select-where-id-2.html> data/replay-select-where-id-2.json` | 재생본 위치와 수 | 4.3, 4.6절과 같음 |

## 7. 확인하지 않은 것

- `workbench/internal/run/run.go`의 실시간 실행 경로는 줄 단위로 읽지 않았습니다.
- `web.Standalone`이 app.js 말고 무엇을 넣는지(app.css 등)는 재생본 전체를 대조하지 않았습니다. 화면 스크립트가 app.js와 같다는 것만 확인했습니다.
- `decode`의 테스트(`*_test.go`)는 실행하지 않았습니다.
- 기록 이전 단계는 해석기 주석이 말하는 범위를 넘어 확인하지 않았습니다. txlab record의 실제 동작, trace.py의 동작(머리 주석과 `TRACE_END` 말고), JFR 시계가 맥 시계인지가 여기에 해당합니다.
- `view.tree` 1,388과 `java.spans` 1,592의 차 204가 무엇으로 이루어졌는지 확인하지 않았습니다.
- 커널 마디 558의 층별 분포, 단계·체크포인트 라벨 분포는 세지 않았습니다.
- `concept-2.html`과 E-05 범위 밖의 evidence 파일(`.strace`, `.out`, `.tsv`, waldump `.txt`)은 받은 범위가 아니라 열지 않았습니다.
- 역할 표시(장소·사건·연결 …)는 `marks.py`의 규칙에 따른 판단입니다. 규칙이 한 키를 한 역할로만 표시하므로, 두 역할을 하는 키(예: `xid`)는 공통 형식 설계 때 다시 정해야 합니다.
- myai의 required 기준과 개발 기본 기준은 읽지 않았고, 적용했다고 주장하지 않습니다.
