# E-03 Spring 기록의 미정·모호 행: 근거와 결정

## 개요

이 문서는 T1 구조 역할 대응표([`t1-structural-roles.tsv`](t1-structural-roles.tsv))의 E-03(Spring 14벌, concept-9의 `RUNS`) 행 가운데, 규칙이 정하지 못한 44행(`TBD`)과 규칙이 정했지만 뜻이 맞지 않거나 구현하기에 모호한 행을 하나씩 정한 근거입니다. 행마다 그 칸을 만든 해석기의 줄, 14벌에서의 수, 그 칸이 가리키는 기록과 그 기록이 늘 하나로 풀리는지, 다른 칸에서 정확히 다시 셀 수 있는지, 그리고 결정을 적습니다.

결정은 [`t1-inventory/roles.py`](t1-inventory/roles.py)의 `_E03` 덮어쓰기에 이유와 함께 있고, 표는 그 생성기로 다시 만들었습니다. 덮어쓴 뒤 미정은 다섯 입력 모두 0행입니다.

앞선 조사(로컬 세션의 하위 에이전트)는 스크립트를 남기지 않아, 여기의 수는 모두 이 문서와 함께 커밋한 스크립트로 처음부터 다시 셌습니다. 앞선 조사와 다른 것은 6절에 있습니다.

## 목차

1. 다시 세는 법
2. 시계와 출처
3. 장소
4. 행별 근거
5. 검사 3·4와 맞물리는 제약
6. 앞선 조사와 다른 것
7. 확인하지 않은 것
8. 결정 검토로 바뀐 것

## 1. 다시 세는 법

```sh
cd <저장소 뿌리>
python3 -I docs/works/evidence/learning-map-20261017/t1-inventory/e03/count.py \
  open_source_internals/workbench/design/concept-9-spring.html
```

스크립트는 concept-9 129행의 `RUNS` literal을 읽고, 해석기(`workbench/design/feasibility/fulltrace/decode/`)가 그 칸을 만든 규칙을 Python으로 다시 써서 원본 값과 견줍니다. 줄마다 `검사 번호  맞음/전체  설명`을 쓰고, 마지막 줄이 `# failed 0`이면 PASS, 하나라도 다르면 그 줄의 맞음이 전체보다 작고 종료 코드가 1입니다(FAIL). `I`로 시작하는 줄은 판정하지 않는 셈입니다. 2026-10-18에 돌린 결과는 판정하는 검사 78개가 모두 맞고 `# failed 0`, 종료 코드 0입니다(결정 검토 뒤 C8b·C9f·C10f를 더함, 8절). 아래 표의 C·R 번호가 그 검사입니다.

해석기의 규칙은 대부분 Go 코드를 줄 단위로 옮겼고, 각 함수의 문서 글에 옮긴 줄 범위를 적었습니다. 규칙을 잘못 옮겼다면 원본과 맞지 않아 그 검사가 실패하므로, 맞음은 "옮긴 규칙이 14벌 모두에서 원본을 만든다"는 뜻입니다. 규칙이 원본의 의미까지 맞다는 뜻은 아닙니다(예: 시간 창으로 붙인 짝은 다시 셀 수 있어도 추정입니다). 둘은 옮긴 것이 아닙니다. C15k(스택에만 보인 마디의 `t0`·`seen` = 기록된 자손의 t0 최솟값·t1 최댓값)와 C16f(throw 효과의 `at` = 그 예외 at[]에서 마디와 메서드가 같은 첫 칸)는 해석기와 다른 규칙이 이 기록에서 같은 값을 내는 것이고, 적재기는 그 규칙을 다시 셈으로 쓰되 맞지 않는 원본이 오면 멈춥니다.

## 2. 시계와 출처

### 2.1 시계 셋

모든 시각은 txlab이 Spring 시나리오를 시작한 순간(`spring.json`의 `startNs`, 맥의 `time.Now`)부터 잰 마이크로초입니다(`decode/spring.go` 407~413). 값이 같은 축에 있어도 찍은 시계가 셋입니다.

| 시계 | 찍은 것 | 근거 |
| --- | --- | --- |
| `mac` | 맥 쪽 중계기의 메시지 `t`, 요청의 `sent`·`done`, 트랜잭션의 `t0`·`t1`, 연결의 `accept`·`close`, 막힘 관찰의 `t`, 멈춤 자리의 `t`·`tGo`, 예외 알림의 `t` | txlab의 `time.Now`(`txlab/spring.go` 217·230·327·353). 멈춤 자리 `t`와 예외 알림 `t`는 앱이 `Instant.now()`로 찍어 알린 값(`spring/src/lab/Lab.java` 37·67·71~75)이고 같은 맥의 벽시계입니다. 정밀도 1 µs(14벌의 이 칸은 모두 소수점 없음, 재고 조사 2.2) |
| `container` | 컨테이너 쪽 중계기의 `tc`, 함수 기록기·uprobe의 서버 사건과 서버 창, 관찰의 `t`·`span`·`reads`, 멈춤 자리·막힘 관찰의 `tCapture` | `txlab/spring.go` 111 주석(`CaptureNs`는 컨테이너 시계), `decode/spring.go` 487 |
| `jfr` | JFR 구간·예외·잠듦, 그것으로 만든 호출 트리, 빌림, 연결 묶임, 슬롯 | `decode/spring_jfr.go` 236~240. `Lab.java` 71행 주석은 JFR과 맥 쪽 기록이 같은 벽시계라고 적지만 확인한 기록은 없습니다 |

관계는 둘입니다. `container − mac`은 연결마다 `clock.wire`의 `[lo, hi]`(밀리초)에 들고, 24연결 모두 같은 순번으로 잘렸으며(`same`) 구간이 비지 않습니다(C8 24/24). 값은 정수 ns를 1e6으로 나눈 것이라 ns로 정확히 돌아갑니다(C8). run의 관계는 그 run 연결들의 구간을 모두 덮는 `[min lo, max hi]`로 둡니다. 차이가 연결마다 일정하다면 참값은 모든 구간의 겹침 안에 있으므로 합집합은 더 넓은(보수적인) 범위이고, 연결마다의 값은 run attrs의 `clock.wire`로 남습니다. `jfr − mac`은 모름(`known: false`)입니다.

### 2.2 출처

| 출처 id | 성격 | 시계 | 무엇 |
| --- | --- | --- | --- |
| `txlab` | observed | mac | `spring.json`: 멈춤 자리, 예외 알림, 접속의 받음·닫음, 막힘 관찰 |
| `observe` | observed | container | `steps.json`의 관찰과, 관찰용 접속의 strace로 해석기가 정한 읽기 구간 |
| `wire-mac` | observed | mac | 맥 쪽 중계기 `wire-mac-<n>.jsonl` |
| `wire-cont` | observed | container | 컨테이너 쪽 중계기 `wire-<n>.jsonl`(해석 결과에는 시각만 남음) |
| `recorder` | observed | container | 함수 기록기와 uprobe: 서버 사건, 스냅숏, 줄 판정, 서버 창 안의 카탈로그 스냅숏 수 |
| `jfr` | observed | jfr | `spring.jfr`의 jdk.MethodTrace, jdk.JavaExceptionThrow, jdk.ThreadPark |
| `spring` | derived | (없음) | 해석기의 규칙: 두 중계기 짝, 요청 묶음과 서버 창의 짝, 창 묶음, 트랜잭션, 관찰 겹침, run attrs의 해석기 값 |
| `spring-mac` | derived | mac | 요청 묶음(splitRequests), 트랜잭션(deriveTxs) |
| `spring-cont` | derived | container | 서버 창(두 ReadyForQuery 입구 사이를 요청 하나로 봄), 서버 기다림(serverWaits), 관찰용 접속이 마지막 답을 보낸 순간(captureSpans) |
| `spring-jfr` | derived | jfr | 호출 트리 마디와 해석기 효과(autocommit, validate, rollbackOnly) |
| `spring-guess` | inferred | jfr | 빌림의 연결(100 ms 창), 연결 묶임·슬롯·묶임 효과(그 빌림 위의 다시 셈), 요청을 보낸 JFR 호출(시계가 다른 두 값의 창) |
| `spring-compare` | derived | (없음) | 14벌을 견준 갈림(`view.branches`, `view.compared`) |
| `src`, `jsrc` | observed | (없음) | PostgreSQL 소스 발췌, 자바 소스 발췌 |
| `topic`, `pgsql-topic` | human | (없음) | 주제가 세운 장소(txlab, 앱), ambiguity note |

해석기 출처를 시계마다 나눈 것은 형식이 사건의 시계를 출처로 정하기 때문입니다(설계 메모 3). 한 출처에 시계가 다른 사건이 섞이면 시각 비교와 Chrome 내보내기가 틀린 축에 놓습니다.

## 3. 장소

| 장소 | kind | 부모 | 출처 | 근거 |
| --- | --- | --- | --- | --- |
| txlab | `client` | 없음 | topic | 막힘 관찰, 관찰이 놓이는 지휘 장소(트랜잭션 실험과 같음) |
| 앱 `lab.Main` | `app` | 없음 | topic | txlab이 띄운 JVM의 main 클래스(`txlab/spring.go` 245). 멈춤 자리와 예외 알림의 장소 |
| 스레드 | `thread` | 앱 | jfr | JFR의 `eventThread`. 14벌에 main, lab connection adder, lab housekeeper, lab connection closer |
| 메서드 | `method` | 스레드 | jfr(트리에만 나온 것은 spring-jfr) | `fn`("패키지.클래스::메서드"). 정적 key, 판은 미기록 |
| 서술자 | `descriptor` | 메서드 | 같음 | `desc`. 구간과 트리 마디의 장소. 정적 key |
| 자바 프레임 | `frame` | 메서드 | 같음 | 프레임 글 "fn:줄". 글에 서술자가 없어 메서드 아래(C25b: 오버로드 메서드 다섯, 그 메서드를 가리키는 예외 프레임 칸 12). 줄 −1(272칸, C25a)은 위치 없는 프레임이고 −1로 내보냄. 경로는 기록에 없어 빈 글 |
| 연결 | `connection` | 없음 | txlab | `C<conn>`(C9a 24/24) |
| 백엔드 프로세스 | `process` | 없음 | txlab | 이름은 pid. txlab이 BackendKeyData로 안 값(`txlab/spring.go` 224~226, C9b 24/24) |
| PostgreSQL 함수·프레임·코드 자리 | `function`·`frame`·`site` | 프로세스 | recorder | 공통 장부(`topics/pgsql/load`). 줄 모름 칸(4칸)과 줄 판정 안의 부른 함수(398)는 같은 백엔드의 스택에서 경로를 찾고, 경로가 둘 이상인 경우는 없음(C26 402/402: 경로 하나 46, 없음 356) |

연결과 백엔드의 짝은 따로 두지 않습니다. 연결 위 메시지가 연결 장소에 놓이고 그 백엔드 쪽 끝(앱이 보낸 F는 `target`, 서버가 보낸 B는 `from`)이 백엔드 프로세스이므로, 메시지마다 짝이 구조에 있습니다. `conns[].pid`와 `sessions`는 이 짝으로 다시 셉니다.

## 4. 행별 근거

"다시 셈"은 내보내기가 적재 결과의 구조와 attrs만으로 그 칸을 만든다는 뜻이고, 적재기는 원본의 값이 그 규칙과 다르면 멈춥니다. 수는 14벌의 합입니다.

### 4.1 멈춤 자리, 막힘 관찰, 관찰

멈춤 자리(68)는 앱이 `Lab.at`에서 알린 순간(`t`)부터 txlab이 go를 보낸 순간(`tGo`)까지 앱이 멈춘 구간입니다. 둘 다 맥 시계이고 68행 모두 `tGo > t`라(C2b) 앱 장소의 `span`(txlab 출처) 하나로 둡니다. 그 사이 txlab이 컨테이너의 관찰기에게 "capture n"을 묻고 받은 시각이 `tCapture`이고, 그 관찰이 `captures[]`의 한 행입니다.

| 경로 | 뜻 | 수 | 가리키는 기록, 하나로 풀림 | 다시 셈 | 결정 |
| --- | --- | --- | --- | --- | --- |
| `checkpoints[].tGo` | go를 보낸 맥 시각(`txlab/spring.go` 327) | 68 | | | `end_ns`(멈춤 자리 구간의 끝) |
| `checkpoints[].tCapture` | 관찰기가 그 관찰을 시작한 컨테이너 시각(`txlab/spring.go` 325) | 68 | 관찰 하나: `t`가 같고 `afterStep = n`, 이름이 같은 관찰이 정확히 하나(C1a 68/68) | 그 관찰의 `t` | `same` 연결(멈춤 자리 → 관찰, 근거 txlab)과 다시 셈 |
| `blocked[].tCapture` | 막힘 관찰의 관찰 시작 | 6 | 이름이 blocked이고 `afterStep = after`인 관찰 하나(C1b 6/6) | 그 관찰의 `t` | `same` 연결과 다시 셈 |
| `blocked[].after` | 그 앞 멈춤 자리의 순번 | 6 | `n`이 같은 멈춤 자리 하나(C1c 6/6) | | `after` 연결(막힘 → 멈춤 자리, 근거 txlab) |
| `blocked[].waiters`, `[]` | 락을 기다린 백엔드(txlab은 pid, 해석기가 이름으로, `spring.go` 627~633) | 6행에 2개 | 0개 또는 1개, 모두 연결 이름(C21 6/6) | | 막힘 사건의 `target` = 그 백엔드 프로세스. 둘 이상이면 적재기가 멈춤 |
| `captures[].afterStep` | 관찰 앞 멈춤 자리의 순번 | 74 | 모든 관찰이 멈춤 자리나 막힘 관찰 하나와 짝(C1d 74/74) | `same`의 상대가 멈춤 자리면 그 `n`, 막힘이면 그 `after` | 다시 셈 |
| `captures[].span`, `[]` | 관찰용 접속이 실제로 일한 구간(`spring.go` 772~803) | 74 | | `span[0] = t`(C5 74/74) | `span[1]`은 `spring-cont` 출처의 `instant`(txlab 장소, 이름 `capture_end`)와 `end` 연결(관찰 → 그 순간, 근거 해석기). `span[0]`은 다시 셈 |
| `captures[].overlap[]` | 관찰 span 안에 컨테이너 쪽이 읽은 메시지(`spring.go` 601~605) | 3관찰에 17 | 규칙으로 다시 센 목록과 차례까지 같고(C7 74/74) 원소마다 메시지 하나와 모든 칸이 같음(C7b 17/17) | 연결한 메시지에서 | `overlap` 연결(관찰 → 컨테이너 쪽 메시지 사건, 근거 해석기). 원소의 칸 7개는 다시 셈 |

관찰은 구간이 아니라 끝이 없는 `sample`(observe 출처)로 둡니다. 읽기 구간(`reads`)의 끝은 "다음 질의를 받은 때"라 관찰 span의 끝보다 늦을 수 있어(370개 가운데 60개, I 줄. 예: 5-requires-new-same-row/a의 둘째 관찰은 span 끝 3,631,723 µs, xacts 읽기 끝 3,637,988 µs), 관찰을 구간으로 두면 같은 장소·시계에서 자식 구간과 어긋납니다. 첫 판은 `span[1]`을 관찰 sample의 끝으로 두었으나, 그 값은 관찰기가 적은 것이 아니라 해석기가 strace에서 "관찰용 접속이 마지막으로 답을 보낸 때"를 규칙으로 고른 것이라(`spring.go` 772~803) 관찰 출처의 칸에 둘 수 없습니다(결정 검토 지적 2). 그래서 끝은 해석기 출처의 순간으로 따로 두고 관찰에서 연결로 잇습니다. 읽기 구간의 끝이 시작과 같은 것은 모두 마지막 관찰입니다(C6b 14/14). 트랜잭션 실험과 같이 그 끝은 비웁니다.

### 4.2 연결, 메시지, 요청, 서버 창

| 경로 | 뜻 | 수 | 가리키는 기록, 하나로 풀림 | 다시 셈 | 결정 |
| --- | --- | --- | --- | --- | --- |
| `conns[].name` | `C<conn>`(`spring.go` 473) | 24 | | | 연결 장소의 이름 |
| `conns[].conn` | txlab이 받은 순번 | 24 | | 이름에서(C9a 24/24) | 다시 셈 |
| `conns[].pid` | 그 연결의 백엔드 | 24 | BackendKeyData 메시지의 `f.pid`와 같음(C9b 24/24) | | 메시지의 `target`(F)·`from`(B) = 백엔드 프로세스 장소 |
| `conns[].thread` | 연결을 연 앱 스레드(`spring.go` 511~514) | 24 | 첫 Q의 맥 시각을 감싼 `QueryExecutorImpl::execute` JFR 구간(구간 차례로 처음, `spring_jfr.go` 414~421)의 스레드(C13 24/24). JFR 구간은 (t0 오름, t1 내림) 차례로 저장돼 있어 사건 id 차례가 그 차례(C13b 14/14) | 그 규칙으로 | 다시 셈. 시계가 다른 두 값(맥과 JFR)을 견준 짝짓기라 구조로 두지 않음 |
| `sessions.{*}` | {pid: 연결 이름}(`spring.go` 474) | 24 | | 연결 장소와 메시지의 백엔드 끝에서(C9e 14/14) | 다시 셈 |
| `conns[].accept`·`close` | txlab이 접속을 받은·닫은 맥 시각 | 24 | | | 연결 장소의 `span`(txlab 출처) |
| `conns[].params`, `startup`·`lastMac`·`lastCont`·`join` | 맥 쪽 중계기가 본 StartupMessage의 매개변수, 그리고 해석기의 요약: 첫 Z까지의 접속 인증 메시지 차례(`startup`), 두 중계기가 마지막으로 적은 메시지 종류(`lastMac`·`lastCont`), 함수 기록의 ReadyForQuery 입구 수와 연결 위 Z 수(`join`)(`spring.go` 150~166·470~520) | 24씩 | | | 연결 수명 구간의 attrs(`params`는 wire-mac, 나머지는 해석기). 첫 판에 자리가 없었음(결정 검토 지적 6) |
| `conns[].messages[].t`·`tc` | 같은 메시지를 맥 쪽·컨테이너 쪽 중계기가 읽은 시각 | 1,402 | 두 기록을 같은 순번으로 맞댔고 방향·번호·종류·길이가 같음(`spring.go` 204~220, C8 24/24). 0인 시각 없음(C9d 24/24) | | 사건 둘(`wire-mac`, `wire-cont`)과 `message` 연결(근거 해석기) |
| `conns[].messages[].dir` | F 앱 → 서버, B 서버 → 앱 | 1,402 | | | F는 `target`, B는 `from`이 백엔드 프로세스. 사건 이름은 `type` |
| `conns[].messages[].name` | Parse·Bind에 적힌 준비 문장 이름 | 38(값은 `S_1`뿐) | | | X2(데이터 대상의 이름) |
| `conns[].messages[].f` | 메시지의 나머지 칸(`spring.go` 222~249 wireText): `auth`, `code`, `pid`, `severity`, `where` | | | | 통째로 두지 않고 맥 쪽 사건 attrs의 `f.auth`·`f.code`·`f.severity`·`f.where` 열쇠로 나눔. 그릇을 통째로 두면 그 아래의 `f.pid`(장소 칸)가 attrs에 숨음(5절 제약 1) |
| `conns[].messages[].f.pid` | BackendKeyData가 알린 백엔드 pid | 24 | 모두 B·K 메시지이고 값이 그 연결의 백엔드 pid(C9f 24/24) | 메시지의 `from`(백엔드 프로세스 장소)의 이름 | 다시 셈. 기록기 칸 `f`의 X2 규칙이 잘못 맞았던 행(결정 검토 지적 3). K가 아닌 메시지에 pid가 있거나 값이 다르면 적재기가 멈춤 |
| `conns[].requests[]` | 앱 쪽 묶음(Sync나 Q로 끝)과 그 답(ReadyForQuery까지, `spring.go` 258~308) | 166 | 메시지에서 규칙으로 다시 센 요청과 seq·sent·done·types·sql·stmt·status·tags·error가 모두 같음(C10c 166/166). 메시지는 많아야 한 요청에 들고(C10b), 어느 요청에도 들지 않는 메시지는 모두 끝내기 X(C10e 24/24) | sent·done은 품은 첫 F와 Z의 시각(C10d 166/166) | `spring-mac` 출처의 구간(연결 장소)과 `contains` 연결(요청 → 메시지, 근거 해석기). 칸은 모두 품은 메시지에서 다시 셈 |
| `conns[].requests[].seq` | 요청 번호(0은 접속 인증) | 166 | | 그 연결의 요청 차례(C10c) | 다시 셈(요청 사건에 attrs가 없음) |
| `conns[].requests[].server` | 백엔드가 그 요청을 처리한 창: 앞 ReadyForQuery 입구부터 이번 입구까지(`spring.go` 516~535) | 142 | seq ≥ 1에만 있고(C11b 166/166) 24연결 모두 함수 기록과 연결의 Z 수가 같음(C11a 24/24). 창은 앞 창의 끝에서 이어 시작(C11e 118/118) | `sql` = 요청의 sql을 "; "로 이은 것(C11c 142/142). `in`·`out`은 142행 모두 null(C11d) | `spring-cont` 출처의 구간(백엔드 프로세스 장소, 경로 무리 `$.conns[].requests[].server.t0`). 요청 → 창은 `server` 연결(근거 해석기, 순서 짝). `catalogSnapshots`는 창이 품는 `recorder` 출처의 `sample`(시각 미기록)의 attrs. `in`·`out`은 null로 다시 세고, null이 아니면 적재기가 멈춤(결정 검토 지적 7) |
| `server.events[].f.at` | hint·ereport 사건의 부른 줄 | 16 | 스택(안쪽부터) 1번 칸의 위치와 같음(C14 16/16) | | X2(기록기 칸 `f` 안의 값, 메모 6). 규칙 순서 탓에 JFR 스택 규칙이 먼저 맞았던 행 |
| `server.checks[].path`·`path[]`·`table`, `lines[]`·`path[]`·`inner[]`의 `fn`·`at` | 줄 판정(트랜잭션 실험과 같은 꼴) | 26 | | | 트랜잭션 실험과 같음: path는 다시 셈(경로 사건이 없으면 null), path[]는 자식 행, table은 X2, fn은 장소, at은 스택 1번 코드 자리 |

요청을 원본 경로 `$.conns[].requests[]`가 아니라 `$.conns[].requests[].sent`의 경로 무리로 적습니다. 표시 파일은 `requests[]`를 관찰로, `seq`·`sent`·`done`을 도출로 표시했고, 요청의 경계는 해석기가 정하므로(재고 조사 2.6의 12) 도출이 맞습니다. 여정 적재기가 질의 구간을 `$.data.endUs`로 적은 것과 같은 선택입니다. 그 대신 요청 사건에는 attrs가 없어야 하고(그 경로 아래의 열쇠는 표시 파일에 없음), 그래서 요청의 칸을 모두 품은 메시지에서 다시 세는 C10이 이 결정의 전제입니다.

서버 창도 도출입니다. 창의 양 끝은 함수 기록기가 본 ReadyForQuery 입구 둘이지만, "앞 입구부터 이번 입구까지를 요청 하나의 처리로 본다"는 것은 해석기의 규칙입니다. 142개 가운데 122개가 그 요청의 첫 F가 컨테이너 쪽 중계기에 닿기 전에 시작합니다(I 줄). 곧 창의 대부분은 백엔드가 다음 질의를 기다린 시간으로 시작하고, 기록기가 본 "그 요청을 처리한 구간"이 아닙니다. 첫 판은 창을 원본 경로(`server`, 관찰)대로 recorder 관찰로 두고 트랜잭션 실험의 요청(`exec_simple_query`의 입구와 반환)과 같은 자리라 했으나, 그 요청은 호출 하나의 입구와 반환이고 이 창은 두 호출 사이라 같은 자리가 아닙니다(결정 검토 지적 1). 그래서 창은 해석기 출처(`spring-cont`)의 구간이고, 창 안에서 기록기가 센 값(카탈로그 스냅숏 수)만 관찰로 남깁니다. 요청과 같이 경로 무리를 원본 경로의 도출 칸(`server.t0`)으로 적어 출처 성격 검사가 도출을 기대하게 합니다.

요청 구간은 같은 연결 장소에서 포개질 수 있습니다. pgjdbc는 BEGIN과 첫 질의를 한 번에 보내(`spring.go` 252~253) 앞 요청의 답이 오기 전에 보낸 요청이 20쌍이고, 20쌍 모두 맥 쪽 중계기가 두 묶음을 한 번에 읽어 보낸 시각이 같습니다(C10f 20/20). 그래서 지금은 BEGIN 요청이 다음 요청 안에 든 것처럼 포개질 뿐 어긋나지 않습니다. 두 묶음을 따로 읽은 기록이 오면 같은 장소의 구간이 어긋나 실행 명령이 그 기록을 쓰지 않고 멈추며(검사 2), 그때 요청을 놓을 자리를 다시 정합니다(결정 검토 지적 5, 메모 9의 2).

### 4.3 트랜잭션

| 경로 | 뜻 | 수 | 가리키는 기록, 하나로 풀림 | 다시 셈 | 결정 |
| --- | --- | --- | --- | --- | --- |
| `transactions[]` | ReadyForQuery 상태가 I를 떠났다 돌아온 구간(`spring.go` 311~338) | 20 | 요청에서 규칙으로 다시 센 트랜잭션과 conn·first·last·t0·t1·end·next가 모두 같음(C12a 14/14, C12b 20/20). t0·t1은 첫 요청의 sent, 끝 요청의 done(C12d 20/20) | | `spring-mac` 구간(연결 장소), 품은 요청마다 `contains`(경로 무리 `transactions[].first`, 근거 해석기) |
| `transactions[].conn` | 연결 번호 | 20 | | | 장소 |
| `first`·`last` | 첫·끝 요청의 seq | 20 | | 품은 요청의 seq 최솟값·최댓값 | 다시 셈 |
| `xids[]`·`clog[]`와 그 칸 | 서버 사건 xid(앞 assign_xid의 level, 요청마다 1부터)와 commit_clog·abort_clog의 사본(`spring.go` 537~553) | 22 + 22 | 다시 센 목록과 같음(C12c 20/20) | 품은 요청의 서버 창 사건에서 | 다시 셈(사본 행을 두지 않음) |
| `next`와 그 칸 | 끝난 뒤 그 연결에 앱이 처음 보낸 메시지의 사본(`spring.go` 326~332) | 20 | 메시지 하나(C12b) | 연결한 메시지에서 | `next` 연결(트랜잭션 → 맥 쪽 메시지, 근거 해석기)과 칸의 다시 셈 |

### 4.4 JFR

| 경로 | 뜻 | 수 | 결정 |
| --- | --- | --- | --- |
| `java.spans[].fn`·`desc`·`thread` | 메서드 호출 하나의 메서드·서술자·스레드 | 1,592 | 사건의 장소 = 서술자 장소, 그 위가 메서드와 스레드 |
| `java.spans[].caller` | 그 호출을 부른 자리, JFR 스택의 0번 칸("fn:줄") | 1,592(빈 것 0, I 줄) | 구간의 스택 1번 칸(부른 쪽 메서드 아래 그 줄의 프레임). 스택은 [서술자, 부른 쪽 프레임]. 첫 판의 X5는 "같은 관계가 트리에 따로 있다"는 전제였으나 트리 마디가 없는 구간 258개(장면 구간 밖의 main 238, main이 아닌 스레드 20, I 줄)에서는 이것이 유일한 호출 사슬 기록이라 틀림(결정 검토 지적 4) |
| `java.spans[].depth` | 같은 스레드에서 이 구간을 감싼 기록된 호출의 수(해석기가 셈) | 1,592 | X5. 트리 마디가 있는 1,334개 가운데 750개가 트리 깊이와 달라(I 줄) 트리에서 다시 셀 수 없음 |
| `java.throws[].thread`, `java.parks[].thread` | 예외·잠듦의 스레드 | 62, 4 | 사건 장소(at[0] 프레임)에서 올라간 스레드 |
| `java.throws[].at`, `java.parks[].at` | 스택 앞 8칸·6칸(안쪽부터) | | 스택(그대로, 0번이 사건의 장소) |
| `java.parks[].under` | 스택 전체에서 첫 추적 대상 칸(`spring_jfr.go` 299~308) | 4 | X5. 4행 모두 at[] 밖이라(C22 4/4) 다시 셀 수 없음 |

### 4.5 호출 트리

트리 마디(1,388)는 `spring-jfr` 출처의 구간이고, 장소는 main 스레드(`spring_tree.go` 22의 `sceneThread`) 아래 서술자, 부모는 `parent`(같은 출처)입니다. 스택은 [마디의 서술자, 숨은 칸(안쪽부터), 부모 메서드의 부른 줄 프레임]이라 `site`와 `hidden`이 그 안에 있습니다. 기록된 마디 1,334개는 저마다 (fn, desc, t0, t1)이 같은 main 스레드의 JFR 구간이 정확히 하나 있습니다(C15l, I 1,334). 해석기가 구간에서 마디를 만들었기 때문입니다(`spring_tree.go` 156~165). 원본에 그 대응이 글로 남지는 않아, 설계 메모 11대로 v0에서는 연결로 두지 않습니다(6절).

| 경로 | 뜻 | 수 | 다시 셈 | 결정 |
| --- | --- | --- | --- | --- |
| `tree[].fn`·`desc` | 마디의 메서드·서술자 | 1,388 | | 장소 |
| `tree[].lane` | 그릴 칸(`spring_tree.go` 265~279) | 1,388 | fn의 앞부분 규칙(C15d 1,388/1,388) | 다시 셈 |
| `tree[].site` | 부모 메서드 안에서 이 호출로 이어진 줄 | 1,388 | 뿌리는 0(C15f 16/16) | 스택 마지막 칸(부모 메서드 아래 그 줄의 프레임) |
| `tree[].hidden`, `[]` | 부모와 이 마디 사이의 추적하지 않은 칸(바깥부터) | 978 | | 스택 가운데 칸(뒤집음) |
| `tree[].t0` | 기록된 마디는 구간의 시작, 스택에만 보인 마디는 처음 본 시각 | 1,388 | 스택에만 보인 54마디는 기록된 자손의 t0 최솟값(C15k 54/54) | 기록된 마디만 `start_ns`, 54마디는 시작을 비우고 다시 셈 |
| `tree[].seen` | 스택에만 보인 마디가 열려 있던 마지막 시각 | 54 | 기록된 자손의 t1 최댓값(C15k 54/54). t1이 null이면 끝난 꼴이 passed이고 seen이 있음(C15e 1,388/1,388) | 다시 셈 |
| `tree[].kids`, `[]` | 자식 | 1,372 | parent가 이 마디인 마디를 t0 차례로(C15b 1,388/1,388) | parent(목록 원소 규칙이 `kids[]`에 잘못 맞았음) |
| `tree[].depth` | 깊이 | 1,388 | 부모 사슬(C15c 1,388/1,388) | 사건의 depth에서 다시 셈 |
| `roots` | 뿌리 | 16 | 부모 없는 마디를 t0 차례로(C15j 14/14) | parent |
| `tree[].branch`(`key`, `alt`) | 이 마디가 간 길(`spring_compare.go` 82~111) | 1,364 | key = fn + desc(1,364/1,364, lab. 마디 24개는 없음, C15g), alt = run attrs의 `view.branches[key].alts`에서 자식 차례·끝난 꼴·마지막 자식 꼴이 맞는 길(C15i 1,364/1,364). `view.branches`·`compared`는 14벌이 같은 값(C23 14/14) | 다시 셈. 맞는 길이 하나가 아니면 적재기가 멈춤 |
| `view.ambiguous` | 맞춤이 둘 이상이었던 수 | 14 | 모두 0(C15m) | run attrs |

### 4.6 효과

효과(224)는 종류마다 자리가 다릅니다. 한 마디 안의 차례는 늘 묶임 효과(bindingReplay가 먼저 붙임, `spring.go` 730~731) → 종류 효과 → 잠듦 → 예외입니다(C16a 1,388/1,388). 그래서 내보내기는 이 차례로 모읍니다.

| 종류 | 수 | 무엇의 사본인가, 하나로 풀림 | 결정 |
| --- | --- | --- | --- |
| wire | 70 | 그 연결의 요청 가운데 sent가 마디 구간 안인 것(`spring_tree.go` 852~863, C16b 1,388/1,388). 요청의 맥 시각을 JFR 구간에 견줌 | 마디 → 요청의 `wire` 연결(근거 `spring-guess`). `conn`·`seq`는 다시 셈 |
| borrow | 24 | 그 마디를 감싸는 main 스레드의 빌림 하나(`spring_tree.go` 864~869, C16c 24/24) | 마디 → 빌림의 `borrow` 연결(근거 해석기). 칸은 빌림에서 다시 셈 |
| park | 4 | main 스레드의 잠듦 하나, `at` = 그 잠듦의 under(C16e 4/4). main 잠듦마다 효과 하나(C16i 14/14) | 마디 → JFR 잠듦의 `park` 연결(근거 해석기). `t0`·`t1`·`at` 다시 셈 |
| throw | 30 | main 스레드의 예외 하나, `at` = 그 예외 at[]에서 마디와 메서드가 같은 첫 칸(C16f 30/30). main 예외마다 효과 하나(C16h 14/14) | 마디 → JFR 예외의 `throw` 연결(근거 해석기). `t0`·`class`·`message`·`at` 다시 셈 |
| autocommit | 40 | `value`는 부른 줄의 소스 글에서 읽은 값(`spring_tree.go` 870~884). `sent`는 40행 모두 0이라 칸이 없음(C16d 40/40) | 마디의 자식 instant(`spring-jfr`), attrs `value` |
| validate, rollbackOnly | 18, 2 | 종류뿐(C16g 20/20) | 마디의 자식 instant(`spring-jfr`) |
| return, suspend, resume | 20, 8, 8 | bindingReplay가 옮긴 연결(C17c 1,388/1,388) | `spring-guess` 출처의 instant(마디의 장소), `target` = 연결, 마디 → 효과의 `contains`(근거 `spring-guess`) |

### 4.7 빌림, 연결 묶임, 슬롯, 기다림, 기다림 고리

| 경로 | 뜻 | 수 | 다시 셈 | 결정 |
| --- | --- | --- | --- | --- |
| `borrows[]` | HikariPool.getConnection(long) 호출마다 받은 연결(`spring_jfr.go` 435~479) | 24 | 시각은 같은 JFR 구간의 시각(C20b 24/24) | `spring-guess` 구간(스레드 장소). conn은 `target`. 빈 글 4행은 모두 시간 초과·idleMs 0·검사 안 함(C20a 4/4)이라 target 없음과 ambiguity note |
| `view.binding[]` | 스레드의 묶임 상태(`spring_tree.go` 403~465) | 70 | 빌림·트리로 다시 센 목록과 같음(C17a 14/14). 첫 상태는 늘 장면 시작의 빈 상태(C17d 14/14) | 첫 상태는 다시 셈(맥 시계의 장면 시작이라 JFR 시계 출처에 둘 수 없음), 나머지 56은 `spring-guess` instant(main 스레드) |
| `binding[].bound` | 묶인 연결 | 70 | | `target` |
| `binding[].by` | 상태를 바꾼 트리 마디(-1은 빌림으로 묶음) | 70 | | `by` 연결(상태 → 마디, 근거 `spring-guess`), -1이면 연결 없음 |
| `binding[].held`, `[]` | 떼어 둔 연결 | 16 | 저장한 상태만으로: by 마디가 doSuspend면 앞 상태의 bound를 쌓고 doResume이면 뺌(C17f 56/56) | 다시 셈 |
| `view.slots[]` | 풀의 연결 하나를 main이 쓴 구간 | 20 | 다시 센 목록과 같음(C17b 14/14). 돌려주지 않은 슬롯(t1 0)은 0개(I) | `spring-guess` 구간(main 스레드), conn은 `target` |
| `view.waits[]` | 서버의 기다림(`spring_tree.go` 478~500) | 2 | 다시 센 목록과 같음(C18a 14/14). 끝은 모두 오류(ereport, 컨테이너 시각, C18b 2/2) | `spring-cont` 구간. 장소는 그 연결의 백엔드 프로세스, 서버 창이 `contains`로 품음 |
| `waits[].conn` | 기다린 연결 | 2 | 장소의 백엔드에서 | 장소의 부모 쪽 짝으로 다시 셈 |
| `waits[].seq` | 기다린 요청의 seq(앞 판의 X1은 틀림: 요청을 가리킴) | 2 | 기다림을 품은 서버 창의 요청에서 | 다시 셈. 끝이 요청의 done(맥 시계)인 기다림은 컨테이너 시계 구간이 될 수 없어 적재기가 멈춤(지금 0행) |
| `view.waitCycles[]`와 그 칸 | 기다림 고리(`spring_tree.go` 532~626) | 6 | 기다림·트리 마디·효과·묶임 상태·슬롯·xid 사건으로 다시 센 목록과 같음(C19 14/14) | 다시 셈. 시각이 종류마다 다른 시계이고(server는 컨테이너, pool은 JFR) 칸이 모두 다른 행의 사본 |
| `view.window` | 장면 구간 [start의 tGo, end의 t](`spring.go` 714~720) | 14 | C3 14/14 | 다시 셈 |
| `view.pauses` | 멈춤 구간(`spring.go` 721~723) | 68 | 멈춤 자리마다 [t, tGo](C2 14/14) | 다시 셈(목록 원소 규칙이 `pauses[]`에 잘못 맞았음) |
| `view.branches.{*}.lines[]` | 갈린 줄 | 252 | | X4(목록 원소 규칙이 잘못 맞았음) |

시각 칸 11,826개에 0은 없습니다(C24). 그래서 `rel(0) = 0`이 "기준 순간"과 "값 없음"을 섞는 함정(`spring.go` 408~411)은 이 기록에서는 생기지 않고, 적재기는 0인 시각을 만나면 ambiguity note를 남깁니다.

### 4.8 표 정의

`tables[]`와 그 열·색인(14벌, 각각 표 하나)은 트랜잭션 실험과 같게 X2(데이터 대상의 정의, run attrs)입니다.

## 5. 검사 3·4와 맞물리는 제약

결정을 고르며 지킨 제약 둘을 적습니다. 구현이 이 제약을 어기면 검사가 실패하므로, 결정이 구현 가능하다는 근거이기도 합니다.

1. attrs에 그릇을 통째로 두면 그 아래의 구조 역할은 X 부류여야 합니다(검사 3의 아래 경로 규칙). 그래서 서버 사건의 `f`를 attrs에 두는 이상 `f.at`은 다시 셈이 아니라 X2입니다. 연결 위 메시지의 `f`는 반대로 그 아래 `f.pid`가 장소(메시지의 `from`)라 X 부류로 둘 수 없어, 그릇을 통째로 두지 않고 칸마다 `f.<칸>` 열쇠로 나누고 `f.pid`는 다시 셉니다(결정 검토 지적 3). 효과·묶임 상태·고리처럼 그 아래에 다시 셀 칸이 있는 목록도 attrs에 두지 않고 행이나 다시 셈으로 둡니다.
2. 출처 성격 검사는 행의 경로 무리에 붙은 표시 파일의 출처를 기대합니다. 그래서 추정 행은 경로 무리가 겹치지 않게 나눕니다: 효과 가운데 해석기 효과는 `$.view.tree[].effects[]`(도출), 묶임 효과는 `$.view.tree[].effects[].conn`(추정 목록), wire 연결은 `$.view.tree[].effects[].seq`(추정 목록)입니다. 추정 목록에 더하는 경로 무리는 `$.borrows[]`, `$.view.binding[]`, `$.view.binding[].by`, `$.view.slots[]`, `$.view.tree[].effects[].conn`, `$.view.tree[].effects[].seq`입니다.

## 6. 앞선 조사와 다른 것

- 오버로드 메서드는 다섯입니다. 넷(HikariPool.getConnection, QueryExecutorImpl.execute, PgConnection.rollback, JdbcTemplate.execute)은 트리에도 서술자 둘로 나오고, 다섯째(DataSourceTransactionManager.&lt;init&gt;)는 장면 구간 밖의 JFR 구간에만 나옵니다. 앞선 조사의 "넷"은 트리 쪽 수와 맞습니다. 그 메서드를 가리키는 예외 프레임 칸이 12인 것은 같습니다(C25b).
- 기록된 트리 마디마다 값이 같은 JFR 구간이 정확히 하나 있습니다(1,334/1,334). 설계 메모 11의 "원본에 마디와 구간의 대응이 없다"는 글로 된 대응이 없다는 뜻으로는 맞지만, 값으로는 하나로 다시 찾을 수 있습니다. 메모 11의 까닭을 그렇게 고치고, 연결로 둘지는 T4가 줌 이동에 필요하면 정합니다.
- `waits[].seq`는 X1(원래 번호)로 표시돼 있었으나 요청을 가리키는 번호라 다시 셈입니다(4.7).
- 기다림 고리는 행이 아니라 다시 셈으로 정했습니다. 시각이 종류마다 다른 시계이고 칸이 모두 다른 행의 사본이기 때문입니다(4.7).
- 시계마다 같은 장소의 구간이 어긋나는지 원본에서 미리 셌습니다(요청과 트랜잭션과 연결 수명, 빌림과 슬롯, 멈춤 자리, JFR 구간과 트리 마디, 서버 창과 기다림). 어긋난 쌍은 0입니다. 적재기의 시험이 같은 것을 형식 위에서 다시 봅니다. 요청 구간의 0은 구조가 보장한 것이 아니라 포개진 20쌍의 보낸 시각이 같아서입니다(C10f, 4.2).

## 7. 확인하지 않은 것

- JFR 시계와 맥 시계의 차이는 모릅니다. 해석기의 짝짓기 셋(빌림의 연결, wire 효과, 연결을 연 스레드)이 이 차이를 0으로 놓고 셈하므로 추정으로 둡니다.
- 원시 기록(`spring.jfr`, `wire-*.jsonl`, `.bin`)은 이 checkout에 없습니다. 판정의 기준은 해석 결과 JSON(concept-9 literal)입니다(재고 조사 5절 함정 11).
- 컨테이너 쪽 중계기가 적은 메시지 칸(SQL, 태그 등)은 해석 결과에 없습니다. 해석기가 맥 쪽 칸만 썼고 두 쪽의 방향·번호·종류·길이가 같음만 확인했습니다(C8). 연결마다의 시계 구간 `clock.wire`는 저장된 메시지의 `tc − t`로 다시 셀 수 있습니다(C8b 24/24: B의 최댓값이 `lo`, F의 최솟값이 `hi`).

## 8. 결정 검토로 바뀐 것

결정(커밋 9f6662c)을 만든 쪽이 아닌 하위 에이전트에게 이 보고서와 표, 셈 스크립트, 해석기, 플랫폼 검사 코드를 주고 검토받았습니다. 요청과 답 전문은 [`t1-e03-decision-review.md`](t1-e03-decision-review.md), 처분은 WORK 문서의 「Spring 결정 검토와 처분」입니다. 이 보고서에서 바뀐 결정은 넷입니다.

1. 서버 창은 recorder 관찰이 아니라 `spring-cont` 도출 구간입니다(지적 1, 4.2). 142개 가운데 122개가 첫 F가 닿기 전에 시작합니다.
2. 관찰의 끝 `span[1]`은 관찰 sample의 끝이 아니라 해석기 순간과 `end` 연결입니다(지적 2, 4.1).
3. 연결 위 메시지의 `f.pid`는 X2가 아니라 `from` 장소에서 다시 셉니다(지적 3, 4.2, C9f).
4. JFR 구간의 `caller`는 X5가 아니라 스택 1번 칸입니다. `depth`는 X5로 남깁니다(지적 4, 4.4).

결정은 그대로이고 근거나 자리를 더한 것: 연결의 칸 다섯의 자리(지적 6), `in`·`out`이 null이 아니면 멈춤(지적 7), 요청 구간이 포개지는 사정과 다시 열 조건(지적 5, C10f), C15k·C16f가 옮긴 규칙이 아니라는 것과 C8b(지적 10). 검토자가 센 수(창 122/142, 읽기 60/370, 트리 마디 없는 구간 258, depth가 다른 750)는 모두 셈 스크립트의 `I` 줄로 다시 셌고 같은 값입니다.
