# T1 마지막 입력 E-03 Spring: 이어서 할 세션의 작업 지시

이 파일은 로컬 세션이 클라우드 세션에 넘기는 지시서입니다. 학습 지도 플랫폼 T1(공통 형식 v0로 기존 기록 옮기기)에서 gdb 8벌, 요청의 여정 11벌, 트랜잭션 실험 44벌은 끝났고 이 브랜치에 커밋되어 있습니다. 남은 입력은 Spring 14벌입니다.

## 0. 시작과 브랜치

- 이 브랜치 `learning-map-t1`에서만 일하고 커밋합니다. GitHub의 `main`에는 이 작업이 없습니다. `main`에 올리지 않고, 강제로 올리지 않습니다.
- 단위(결정, 적재기, 검토)가 끝날 때마다 `origin learning-map-t1`에 올립니다.
- 사용자에게는 한국어로 답합니다.

## 1. 먼저 읽을 것(모두 이 브랜치에 있음)

- `docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md`의 「T1 공통 형식 v0와 기존 기록 이식」 → 「설계 메모」 1~12. 특히 2(형식), 5의 3(트랜잭션 실험 규칙, 가장 가까운 선례), 5의 4(Spring 계획), 6(예외 부류 X1~X7), 7(도출과 추정, `rule:` 무리), 9(검사 1~6)와 「실행 기록」.
- 재고 조사 `docs/works/evidence/learning-map-20261017/t1-inventory.md` 2절(Spring 키 트리, 시계, 모호함 2.6).
- 결정 표 `docs/works/evidence/learning-map-20261017/t1-structural-roles.tsv`와 생성기 `t1-inventory/roles.py`. 다시 만드는 명령: `cd docs/works/evidence/learning-map-20261017/t1-inventory && python3 -I roles.py marks_e05.txt e05 marks_c3.txt c3 marks_c4.txt c4 marks_tx.txt tx marks_e03.txt e03 > ../t1-structural-roles.tsv`.
- 코드: `open_source_internals/workbench/internal/atlas/`(형식과 검사), `topics/pgsql/{load,gdbtrace,journey,txlab}`(txlab을 본보기로), `topics/pgsql/facts.json`, `topics/pgsql/cmd/pgsql-atlas/main.go`. 기록을 만든 해석기는 `workbench/design/feasibility/fulltrace/decode/spring*.go`, 화면은 `workbench/design/spring-view.tmpl.html`.
- 입력: `workbench/design/concept-9-spring.html` 129행의 `RUNS` literal(열쇠 예: `5-nested/a`). txlab처럼 `load.Literal`로 읽습니다. `.lab-tmp`의 JSON 사본은 이 checkout에 없습니다.

## 2. 할 일(차례대로)

1. **근거.** e03의 `TBD` 44행과, 이미 정했지만 구현하기에 모호한 약 35행(예: `place/target` "P3 실행 위치", 정규식이 잘못 맞은 `kids[]`·`branches.lines[]`·`pauses[]`·`f.at`·`checks[].path[].at`, X1로 잘못 둔 `waits[].seq`)마다 뜻(해석기 파일:줄), 14벌의 수, 가리키는 기록과 늘 하나로 풀리는지, 정확히 다시 셀 수 있는지, 결정을 적습니다. 보고서는 `docs/works/evidence/learning-map-20261017/t1-e03-evidence.md`이고, 셈 스크립트는 `t1-inventory/e03/`에 커밋해 모든 수를 다시 돌릴 수 있게 합니다(`python3 -I`).
   앞선 조사(스크립트를 잃어 모두 다시 확인해야 함)의 결과:
   - 결정의 구성은 recompute 23, link 6, X2 8, stack 2, 시작/끝 2, place 1, target 1, row 1이었고, 규칙은 모두 n/n으로 맞았습니다.
   - 관찰마다 `afterStep = n`으로 멈춤 자리 하나(68/68)나 `blocked` 하나(6/6)와 짝이 맞고, `tCapture`는 관찰 `t`와 같습니다.
   - `deriveTxs`를 다시 돌리면 트랜잭션 20개가 모두 같고, `xids`·`clog`는 서버 사건의 사본입니다(22/22).
   - `view.*`(branch key/alt 1364/1364, lane, kids, seen 54/54, binding·slots 14/14, waits, waitCycles 6/6)는 정확히 다시 셀 수 있습니다.
   - `overlap[]`은 `conns[].messages[]`의 사본입니다(17/17).
   - `conns[].thread`는 해석기 `threadAt`(24/24), `events[].f.at`은 스택 1번 칸입니다(16/16).
   - JFR 프레임 글에는 서술자가 없습니다. 오버로드 메서드가 넷이고, 예외 프레임 12칸은 어느 오버로드인지 가를 수 없습니다.
   - 줄이 −1인 JFR 프레임이 272개, `conn`이 빈 글인 빌림이 4개이고, `parks[].under`는 4/4가 `at[]` 밖입니다. 시각 값 0은 없습니다.
2. **결정**을 적재기 코드보다 먼저 커밋합니다. `roles.py`에 e03 덮어쓰기를 이유와 함께 더해(미정 0) 표를 다시 만들고, 메모 5의 4를 고치고 실행 기록을 남깁니다. 권하는 결정(검증하고, 근거가 있을 때만 바꾸며 이유를 적음):
   - 관찰과 멈춤 자리·`blocked`: `same` 연결(근거 txlab, 관찰). `afterStep`과 두 `tCapture`는 다시 셈(같은 순간을 사건 둘로 만들지 않음).
   - `xids`·`clog`: 사본 행이 아니라 다시 셈(메모 5의 4의 copy 문구를 고침).
   - 트랜잭션은 해석기 출처의 `span` 행이고, 범위 안 요청마다 `contains` 연결. `first`·`last`는 다시 셈, `next`는 연결.
   - `view.binding`, `view.slots`, return·suspend·resume 효과는 **추정** 출처(입력인 빌림이 추정이고 화면도 ≈로 표시). 메모 7의 목록에 더함.
   - 스택에만 보인 트리 마디 54개(`t1 = null`): `t0`는 시작이 아니라 처음 본 시각이므로 시각을 두지 않고 `t0`·`seen`을 자손에서 다시 셈.
   - `branch.alt`: run attrs의 X4에서 다시 셈. 맞는 길이 정확히 하나가 아니면 적재기가 멈춤.
   - `parks.under`: X5. JFR 줄 −1은 위치 없는 프레임으로 두고 −1로 내보냄(pgsql의 `?`와 같은 처리). 빌림 `conn ""`은 target 없음과 ambiguity note.
   - JFR 장소: 스레드 → 메서드(`fn`) → 서술자(`desc`)에 구간과 트리 마디를 두고, 프레임(`fn` + 줄)은 메서드 아래. 정적 key는 pgsql 선례대로 부모 key를 넣음.
   - 두 시계(맥 `t`, 컨테이너 `tc`)와 JFR은 메모 3대로. 연결 위 메시지는 사건 둘과 `message` 연결.
3. **결정의 분리 검토.** 적재기 전에 독립 하위 에이전트에게 메모와 표를 주고 검토받습니다. 답 전문과 지적마다의 처분을 WORK 문서에 남깁니다.
4. **적재기** `topics/pgsql/spring`(`ReadRuns`, `Import`, `Export`, `InferredGroups`)와 시험을 txlab 수준으로 만듭니다.
   - 새로 다시 읽은 원본과의 되돌림 비교.
   - `roles.Load("e03")`로 `Validate`·`ScanStructure`·`CheckNature`.
   - 원본에서 바로 센 사실과 견주는 주제를 모르는 재추출. 재고 조사의 검사 5 사실(요청과 트리 마디, 연결 위 메시지, `view.waits`를 포함한 기다림, 막힘 관찰, JFR 잠듦, 결과 값)을 facts.json의 `spring` 목록으로 더함.
   - 검사마다 판별력을 보이는 변이 시험.
   - 시계마다 어긋난 구간 0.
   - `cmd/pgsql-atlas`에 넣음(8+11+44+14 = 77벌 기대). `.lab-tmp/`는 여기서 git이 무시하지 않으므로 `-out`으로 임시 폴더에 쓰고 절대 커밋하지 않습니다.
5. **독립 코드 검토.** 하위 에이전트로 검토합니다(구조가 attrs에 숨지 않았는지, 출처 표시, 내보내기가 적재 결과 밖을 읽지 않는지, X1~X7). 답 전문과 처분을 WORK 문서에 남깁니다. Claude가 아닌 모델(Codex)의 검토와 myai 작업 지도의 T1 체크 제출은 뒤에 사용자 컴퓨터에서 하므로 시도하지 말고 남은 일로 적습니다.

## 3. 커밋마다 통과할 것

- `cd open_source_internals/workbench && go test ./... -count=1`(Go 1.27 이상).
- `golangci-lint run ./internal/atlas/... ./topics/...`가 0건. 없으면 golangci-lint v2를 설치하고, 설치하지 못하면 `go vet`을 돌리고 lint를 못 돌렸다고 밝힙니다.
- `node topics/pgsql/check-screens.mjs` 통과.
- 파일을 하나씩 골라 올리고(explicit staging), 메시지 끝에 `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. 커밋과 파일에 회사 이름 같은 식별 정보를 쓰지 않습니다.
- 코드 꼴은 기존 패키지와 같게: 한국어 문서 주석은 "이름 함수는/메서드는/값은"으로 시작, 시험 함수에도 문서 주석, 시험은 Given/When/Then. 배관은 직접 짜지 말고 라이브러리를 씁니다.

## 4. 일하는 원칙(사용자의 실행·개발 기준을 줄인 것)

- 최상위 기준은 자원 효율입니다. 다만 데이터 무결성, 정확성, 요구한 정밀도는 시간을 아끼려고 낮추지 않습니다.
- 사실, 추론, 모르는 것을 섞지 않고, 핵심 주장마다 확인 경로를 둡니다.
- 중요한 결정은 구현 전에 설계하고, 새 사실이 설계를 깨면 설계부터 고칩니다.
- 중요한 결과는 만든 쪽이 스스로 인증하지 않고 분리된 검토를 받습니다. 재검토는 앞선 지적의 수정과 회귀만 봅니다.
- 시험은 결과와 상태를 보고, 틀린 규칙도 통과할 우려가 있으면 변이나 반례로 판별력을 확인합니다.
- 작고 응집된 커밋. 돌리지 않은 검사를 돌린 것처럼 쓰지 않습니다.
- 요청한 결과가 있고, 중요한 성질을 확인했고, 요구를 낮추지 않았고, 남은 잔여물이 없고, 확인하지 못한 범위를 밝혔을 때만 완료라고 합니다.

## 5. 완료 조건과 마지막 보고

- Spring 행이 다시 돌릴 수 있는 근거와 함께 모두 정해짐.
- 적재기가 14/14벌에서 검사 1~5를 통과하고 변이가 모두 잡힘.
- 명령이 77벌을 씀.
- 두 검토의 처분이 기록됨. 메모와 실행 기록이 고쳐짐. 모두 `learning-map-t1`에 올라감.

마지막에 사용자에게 한국어로 요약합니다. 커밋, 시험과 lint 결과, 메모에서 바뀐 것, 남은 일(Codex 검토, T1 체크 제출, 사용자의 T3 시안 확인)을 적습니다.
