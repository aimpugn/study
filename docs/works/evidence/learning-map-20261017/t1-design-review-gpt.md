다시 설계

1. **주제별 구조를 `attrs`에 숨겨도 검사를 통과할 수 있습니다.**

   **심각도: 높음**

   **근거:** 설계 메모의 형식 표는 `attrs`를 구조가 정해지지 않은 칸으로 두고, 주제 값은 그 안에 넣습니다(설계 메모 §1·§2, [WORK 문서](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:360>). 검사는 잎 값 소비와 되읽기, 일부 관계를 확인합니다(§7, [WORK 문서](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:398>)). 끝 기준은 “주제 칸은 확장 칸에만”입니다([T1 기준](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:281>).)

   **문제:** 적재기가 구조상 필요한 필드—예를 들어 메시지 방향·연결 종류·요청 창 소속—를 `attrs`에 넣어도 잎 값 대응과 되읽기는 통과할 수 있습니다. 관계 검사는 정한 배열 무리만 대상으로 하므로, 모든 재사용 가능한 구조가 정식 필드나 연결로 표현됐는지 판정하지 않습니다. 구조 칸을 Go 타입으로 닫아 둔 것과 구조 사실을 `attrs`에 숨기지 못하게 막는 것은 별개의 조건입니다.

   **고칠 제안:** 어떤 의미가 구조인지 먼저 열거하고, 그 사실을 `attrs`에 넣은 적재는 실패하는 규칙을 추가하세요. `attrs`에 허용되는 주제 사실과 `place`·`parent`·`link`·시간 관계로 표현해야 하는 구조 사실을 구분한 스키마를 두고, 재추출기가 그 경계까지 검사해야 합니다.

2. **출처 표지만으로 추정 연결이 관측 사실로 표시되지 않는다고 보장하기 어렵습니다.**

   **심각도: 높음**

   **근거:** 사건은 출처를, 연결은 `basis`를 가리키고 출처 성격은 `observed`·`derived`·`inferred`·`human` 중 하나로 둡니다(§1, §7. [형식·검사](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:364>) 및 [출처 검사](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:358>)). 재고 조사에는 메시지·시스템 콜 연결을 해석기가 순서나 인자로 맞추는 사례가 있습니다([재고 조사 §3.4.5](</Users/rody/VscodeProjects/study/docs/works/evidence/learning-map-20261017/t1-inventory.md:1452>), [설계 메모 §4](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:388>)).

   **문제:** “출처가 `inferred`라고 기록됐다”는 것과 “그 연결을 화면·재추출에서 관측 사실로 취급하지 않는다”는 것은 다릅니다. 다섯 검사에는 연결이 실제로 어떤 표지로 노출되는지, 연결의 성격이 근거 출처에서 일관되게 이어지는지 확인하는 항목이 없습니다. 입력에 없는 짝을 순서나 인자로 추정해 연결하고도 출처 성격을 잘못 달면 적재 결과가 통과할 여지가 있습니다.

   **고칠 제안:** 연결의 근거 출처뿐 아니라 관계의 주장 수준(직접 관측, 해석기 도출, 추정된 대응)을 명시하고, 그 수준이 재추출 결과와 사용자 표지까지 보존되는지 검사하세요. 원본이 보장하지 않는 연결은 “미기록” 또는 “추정”으로만 통과하도록 반례를 넣어야 합니다.

3. **시간 필드는 시계가 여러 개인 기록의 사건 순서와 간격을 안전하게 표현하지 못합니다.**

   **심각도: 높음**

   **근거:** T0는 run 시작 기준 상대 나노초와 run의 시계 정보를 정합니다(T0 결정 1, [WORK 문서](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:311>)). 그런데 재고 조사는 E-03에서 맥·컨테이너 시계가 섞이고 `clock.wire`는 밀리초라고 기록합니다([재고 조사 §5.1](</Users/rody/VscodeProjects/study/docs/works/evidence/learning-map-20261017/t1-inventory.md:1859>)). 메모는 출처별 시계 이름·정밀도를 두되 “사건마다 시계 칸을 둘 필요가 없다”고 합니다([WORK 문서](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:378>)).

   **문제:** 출처별 시계 이름과 정밀도만으로는 서로 다른 시계의 원점 차이나 보정 관계를 나타내지 못합니다. 서로 다른 시계의 사건을 같은 run의 `start_ns`로 정렬하거나 간격을 비교하면 관측 근거가 없는 순서·지연이 생길 수 있습니다. 나노초 변환의 왕복 검사는 단위 손실만 검사하고, 시계 간 정렬이 참인지는 검사하지 않습니다.

   **고칠 제안:** 사건의 시각을 원시 시계 좌표와 시계 ID로 보존하고, run 공통 시간축에 투영할 때는 보정 근거·오차·정밀도를 별도로 기록하세요. 보정 불가 시에는 비교 불가로 표시하고, 이를 Chrome 내보내기·경로 재생이 어떻게 다루는지도 정해야 합니다.

4. **합쳐진 이동을 `span`으로 바꾸면 사건의 종류와 개별 발생 정보를 잃을 수 있습니다.**

   **심각도: 높음**

   **근거:** 재고 조사에 개별 시각이 없는 `count > 1` 이동이 있으며, 이를 펼칠 수 없다고 명시합니다([재고 조사 §3.4.5](</Users/rody/VscodeProjects/study/docs/works/evidence/learning-map-20261017/t1-inventory.md:1456>), [재고 조사 §5.7](</Users/rody/VscodeProjects/study/docs/works/evidence/learning-map-20261017/t1-inventory.md:1873>)). 메모는 이를 여러 번의 이동이 아니라 `span` 하나와 `attrs`의 `count`로 표현합니다([WORK 문서](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:388>)).

   **문제:** `span`은 시작·끝이 있는 구간 사건으로 정의됐지만, 합친 이동은 개별 발생 시각이나 구간이 없습니다. 하나의 `span`에 횟수만 넣으면 시간상 구간과 집계 묶음을 같은 종류로 취급하게 됩니다. 이후 지도에서 사건 수, 반복 횟수, 경로 순서를 재구성할 때 의미가 달라질 수 있습니다.

   **고칠 제안:** 개별 사건과 집계 사건을 형식에서 구분하세요. 집계 사건에는 개별 발생 수, 알려진 시간 범위 여부, 원본에서 분리 불가능하다는 상태를 두고, 개별 사건처럼 시간 재생·호출 순서에 투입되지 않는 검사를 추가하세요.

5. **다섯 검사는 모든 화면 사실의 보존과 올바른 의미 대응을 판정하지 않습니다.**

   **심각도: 높음**

   **근거:** 검사 ①은 읽지 않은 잎 값, ②는 원본 값의 되읽기, ③은 일부 정한 관계, ④는 참조와 시간 형식, ⑤는 목록에 든 화면 계산을 비교합니다([검사 목록](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:396>)). 기존 화면에는 표에 적힌 계산 외에도 층·노드·관계 이름·단계 라벨·상태 계산 등 화면 자체의 사실이 있습니다([재고 조사 §3.5.4](</Users/rody/VscodeProjects/study/docs/works/evidence/learning-map-20261017/t1-inventory.md:1618>), [재고 조사 §5.9–10](</Users/rody/VscodeProjects/study/docs/works/evidence/learning-map-20261017/t1-inventory.md:1875>)).

   **문제:** 잎 값이 목적지 어딘가에서 왕복된다는 사실은 그 값이 올바른 사건·연결·출처에 붙었다는 뜻이 아닙니다. 또한 검사 ⑤의 정해진 재추출 항목과 화면 상수 평가가 T1 기준의 “기존 화면의 사실” 전체를 덮는다는 보장은 없습니다. 필드 사이 관계가 뒤바뀌거나, 상수·파생 결과가 일부 빠져도 선택한 비교만 통과할 수 있습니다.

   **고칠 제안:** 입력별로 화면이 소비하는 사실을 완전한 목록으로 만들고, 각 사실을 원본의 경로·변환 규칙·목적지·검사에 연결하세요. 최소한 이벤트 순서, 연결 끝점과 방향, 기다림의 시작·끝, 값의 실행 구분, 출처 표지, 화면 계산 결과를 포함한 독립 기대값과 결과를 비교해야 합니다. 값 왕복 검사와 의미 대응 검사는 분리하세요.

6. **`null`·빈 문자열·0·미기록·관측된 0개를 표현하는 규칙이 없습니다.**

   **심각도: 높음**

   **근거:** 재고 조사에는 시각 0, `null` 스냅숏, 빈 연결 문자열, `stack: null`, END에 없는 함수(0회)가 각각 다른 뜻으로 등장합니다([재고 조사 §5.8](</Users/rody/VscodeProjects/study/docs/works/evidence/learning-map-20261017/t1-inventory.md:1874>)). 메모는 “빈 목록·`null`·0회도 값”이라고 제외 목록에서 정합니다([WORK 문서](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:394>)).

   **문제:** 값을 그대로 보존해도 의미 차이는 정해지지 않습니다. 특히 `coverage[]`가 “무엇을 지켜봤는가”만 담으면, 관측 결과가 0건인 경우, 관측하지 않은 경우, 해당 없음, 알 수 없는 경우를 구분하기 어렵습니다. 지도에서 기록하지 않은 층을 빗금으로 보여 주는 요구와도 연결됩니다.

   **고칠 제안:** 선택 필드마다 존재 여부와 의미 상태를 정하세요(예: 관측됨·관측 결과 0개·수집 안 함·해석 불가·해당 없음). `coverage`도 장소만이 아니라 관측 대상 종류와 범위·구간·완전성을 포함해야 “안 봄”과 “봤지만 0건”을 판별할 수 있습니다.

7. **장소와 사건의 출처·버전이 형식에서 일관되게 추적되지 않습니다.**

   **심각도: 중간**

   **근거:** T1 기준은 실행·사건·장소 식별자와 코드 위치·버전을 보존하라고 합니다([T1 기준](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:281>)). 그러나 `place`에는 `source`가 없고, 장소 출처를 어디서 판정하는지 명시되어 있지 않습니다. 재고 조사에는 `procsGuess`, 분류된 층·역할, 추론된 커널 호출 후보처럼 장소와 밀접한 해석이 있습니다([재고 조사 §3.5.4–5](</Users/rody/VscodeProjects/study/docs/works/evidence/learning-map-20261017/t1-inventory.md:1620>), [재고 조사 §3.5.5](</Users/rody/VscodeProjects/study/docs/works/evidence/learning-map-20261017/t1-inventory.md:1653>)).

   **문제:** `place.attrs`에 출처 ID를 넣는다고 해도 이를 강제하는 규칙이나 검사가 없습니다. 따라서 추론된 장소 역할이나 소스 위치를 관측 사실처럼 소비할 수 있고, 소스 버전별 장소 key와 그 key의 근거도 검증하기 어렵습니다.

   **고칠 제안:** 장소 식별 정보와 분류·배치 주장을 각각 어느 출처가 제공했는지 연결하고, `key`가 포함하는 소스 판·부모 식별 규칙을 명시하세요. 장소에 대한 출처 표지도 사건·연결과 같은 형식 검사 대상에 포함하세요.

8. **다음 단계의 가지 공유와 설명 재사용에 필요한 구조적 기준이 아직 모호합니다.**

   **심각도: 중간**

   **근거:** T2는 반복 가지 공유와 세 질의를 측정하고, 설명 재사용은 범위·장소·가지 모양 해시 등을 열쇠로 씁니다([T2 기준](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:282>), [설명 층](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:233>)). T1은 부모 포함과 `contains` 링크를 함께 쓰고, 사건마다 스택도 별도로 둡니다([WORK 문서](</Users/rody/VscodeProjects/study/docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md:370>)).

   **문제:** 가지 모양을 해시할 때 `parent`, `contains` 링크, `stack` 중 무엇이 정본인지와 어떤 연결 종류가 가지 모양에 들어가는지 정해져 있지 않습니다. 같은 호출 경로가 서로 다른 출처로 쪼개진 기록에서 동일한 모양인지 비교할 규칙이 없으면 T2의 공유율과 설명 재사용 열쇠가 적재기별 판단에 흔들릴 수 있습니다.

   **고칠 제안:** 호출 포함·연결·스택의 역할을 분리하고, 가지 해시 입력을 형식 수준에서 정하세요. 출처 판이 바뀌어도 유지되는 구조와 바뀌어야 하는 구조의 예를 만들어 Q1~Q3 및 해설 재사용 검사에 포함하세요.

### 확인하지 못한 것

- 요청하신 대로 프로젝트 코드나 테스트는 실행하지 않았고 네트워크도 사용하지 않았습니다.
- 설계 메모가 언급한 원본 해석기·작업대 코드도 읽지 않았습니다. 재고 조사와 설계 문서만으로 위 지적을 검토할 수 있다고 판단했습니다.
- 전역 `AGENTS.md`가 가리키는 myai 기준의 조회 명령은 실행하지 않았습니다. 해당 조회는 사용자가 금지한 실행에 해당할 수 있어 진행하지 않았으며, 그 기준을 적용했다고 주장하지 않습니다.