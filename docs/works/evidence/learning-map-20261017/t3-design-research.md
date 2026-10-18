# 지도형 학습 화면 설계 조사 (design3)

확인일은 모두 2026-10-17입니다. 표기 규칙은 다음과 같습니다.

- 「확인」은 문서나 원문 본문을 직접 읽은 것입니다.
- 「검색 요약만」은 검색 결과 요약으로만 본 것이며 본문을 읽지 못했으므로 미확인으로 취급합니다.
- 「추론」은 이 플랫폼에 적용하려는 우리 쪽 판단이며 출처가 말한 내용이 아닙니다.

---

## 1. 정보 탐색 만트라 (Shneiderman 1996)

- 출처: Ben Shneiderman, "The Eyes Have It: A Task by Data Type Taxonomy for Information Visualizations", Proc. Visual Languages 96, 1996년 9월 (UMD 기술보고서 CS-TR-3665 / ISR-TR-96-66). https://drum.lib.umd.edu/handle/1903/5784 , 본문 PDF 사본 https://webspace.science.uu.nl/~telea001/uploads/VACourse/Shneiderman96.pdf
- 확인한 내용: 본문 PDF의 표지와 2~3절을 직접 읽었습니다. 만트라는 "Overview first, zoom and filter, then details-on-demand"이고, 저자는 이것을 "시작점일 뿐"이라고 선을 긋습니다. 이어서 7개 작업(overview, zoom, filter, details-on-demand, relate, history, extract)을 정의하는데, history는 "undo, replay, progressive refinement를 위한 행동 이력 보관"이고 relate는 항목 사이의 관계 보기입니다. 2차원(지도) 데이터의 사용자 문제로 "항목 사이의 경로"를 명시합니다. IEEE 학회 DOI(10.1109/VL.1996.545307)는 검색 요약만 있어 미확인입니다.
- 빌릴 것: 첫 화면은 개요(시스템 수준)이고, 줌과 필터를 거쳐, 세부(함수·커널)는 요청할 때만 보입니다. history 작업은 탐색 이력(되돌리기, 다시 재생)의 근거가 됩니다. extract는 "구간을 뽑아 저장"하는 범위 설명과 맞닿습니다.
- 빌리지 않을 것 / 주의: 만트라는 선형 순서처럼 읽히지만, 이 플랫폼은 경로 한 건이 세 수준을 오르내리므로 순서가 한 번에 끝나지 않습니다. 논문은 "시작점일 뿐"이라고 했으니 만트라 자체를 법칙으로 인용하면 안 됩니다.

## 2. 시맨틱 줌 (Perlin & Fox / Bederson & Hollan)과 Furnas fisheye

- 출처 (a): Ken Perlin, David Fox, "Pad: An Alternative Approach to the Computer Interface", SIGGRAPH '93. https://www.cs.umd.edu/projects/hcil/jazz/learn/papers/siggraph-93-origpad.pdf . 쪽수(57–64)와 DOI는 검색 요약만이라 미확인입니다.
- 출처 (b): Benjamin Bederson, James Hollan, "Pad++: A Zoomable Graphical Interface System", CHI '95 Companion(Demonstrations), p.23. https://cs.umd.edu/~bederson/images/pubs_pdfs/p23-bederson.pdf . (UIST 1994 논문은 읽지 않았으므로 인용하지 않습니다.)
- 출처 (c): George Furnas, "Generalized Fisheye Views", CHI '86 Proceedings, pp.16–23. https://pages.cpsc.ucalgary.ca/~saul/581/exer.eps/4furnas86.pdf (수업 사본)
- 확인한 내용: Perlin & Fox는 문서가 작게 보이면 제목만, 커지면 요약, 더 커지면 전문을 보여 주는 동작을 "We call this semantic zooming."이라 부르며, 이를 객체의 표시 메서드가 배율에 따라 달라지는 방식으로 구현합니다. Bederson & Hollan은 같은 개념을 "줌아웃하면 단순화된 렌더링, 줌인하면 더 많은 세부"로 설명하고 크기·뷰 복잡도·과업·렌즈 등에 따라 모습이 바뀐다고 확장합니다. 두 문서에는 "기하 줌(geometric zoom)"을 정의한 문장이 없습니다. 기하 줌은 모양은 그대로 크기만 바뀌는 것이라는 대비는 이 맥락의 추론입니다. Furnas는 관심도 DOI(x|y) = API(x) − D(x,y)를 제시합니다. API는 선험적 중요도이고 D는 현재 초점 y까지의 거리입니다. Perlin의 논문이 Furnas의 fisheye를 직접 인용하므로 두 계보가 이어진다는 것은 사실입니다.
- 빌릴 것: 줌 수준마다 같은 대상의 표시가 달라집니다(시스템 박스, 프로세스 내부 구조, 함수·커널 줄). Furnas식 DOI는 "현재 초점 단계에서 가까운 단계는 자세히, 먼 단계는 요약"이라는 규칙으로 쓸 수 있습니다. D를 지도상 거리가 아니라 경로상 단계 거리로 정의하는 것은 추론입니다.
- 빌리지 않을 것 / 주의: 위치와 크기가 의미가 있는 Pad의 무한 평면은 계층 구조에는 과합니다. 원논문에서 검증된 것은 정성적 설명과 프로토타입뿐이라 사용성 효과를 근거로 삼지 못합니다.

## 3. 지도 앱의 경로 UX (Google Maps, Apple Maps)

- 출처: "Get directions & show routes in Google Maps - Computer", Google Maps Help. https://support.google.com/maps/answer/144339
- 확인한 내용: 경로가 여럿이면 가장 좋은 경로는 파란색, 나머지는 회색이고 각 경로에 예상 소요 시간이 지도에 표시됩니다. 인용: "To choose another route in any transportation mode, select it on the map." 데스크톱에서는 경로 아래 Details로 단계 목록을 열고 Expand로 상세를 펼치며, 단계에 포인터를 올리면 Street View 미리보기가 뜹니다. 모바일 안내 페이지는 경로 정보를 위로 스와이프해 목록을 연다고 합니다(검색 요약).
- Apple Maps: "Select other route options in Maps on iPhone"과 "View a route overview or a list of turns" 두 페이지를 직접 읽으려 했으나 목차만 반환되어 본문은 미확인입니다. 검색 요약상 대안 경로를 탭하면 선택되고 소요 시간이 카드 상단에 나옵니다. Google 쪽으로 요구사항("또는")은 충족했습니다.
- 빌릴 것: 대안 경로(같은 요청의 다른 기록)는 기본 경로와 구분되는 회색으로 두고, 각 경로에 총 소요 시간 같은 한 숫자 요약을 붙입니다. 경로를 지도 위에서 직접 눌러 바꿉니다. 단계 목록은 "경로 → 단계 목록 → 한 단계 미리보기" 순서의 접힌 패널로 둡니다.
- 빌리지 않을 것 / 주의: 지도 앱의 대안 경로는 "서로 다른 길"이고 우리 대안은 "같은 요청의 다른 실행 기록"이라 시간 차이의 원인이 다릅니다. 시간 라벨만으로는 왜 다른지 알 수 없으므로 갈라지는 지점을 표시해야 합니다(추론).

## 4. 트레이스 뷰어 (Perfetto, Chrome DevTools, Firefox Profiler)

**Perfetto**
- 출처: https://perfetto.dev/docs/visualization/perfetto-ui , https://perfetto.dev/docs/instrumentation/track-events , https://perfetto.dev/docs/getting-started/converting , https://perfetto.dev/docs/contributing/ui-plugins
- 확인한 내용: 영역 선택은 "Click and drag over the timeline to make an area selection." 방식이며 시작·끝 시각과 트랙 목록으로 구성되고, 마커를 끌어 시간을, 체크박스로 트랙을 바꿉니다. 이벤트 상세는 Q로 여닫는 탭 서랍의 "Current Selection" 탭에 나옵니다(문서에 "하단 패널"이라는 말은 없음). 트랙 핀 아이콘으로 트랙을 맨 위에 고정합니다. 플로우 이벤트는 같은 flow id를 가진 슬라이스를 화살표로 잇고, 이벤트 하나를 선택했을 때 화살표가 표시된다고 track-events 문서가 말합니다. 사용자용 수동 메모 기능은 못 찾았고, 플러그인 API에만 점 노트와 구간 노트(span note: 강조 시간 범위)가 있습니다.
- 빌릴 것: 구간 = 시작+끝+포함 트랙의 묶음. 핀 고정은 "항상 보이는 계층 한 줄". 화살표는 선택했을 때만 그려 시각 소음을 줄입니다.
- 주의: 영역 선택은 임시이며 저장·버전·채택 개념이 없습니다.

**Chrome DevTools**
- 출처: https://developer.chrome.com/docs/devtools/performance/reference , https://developer.chrome.com/docs/devtools/performance/annotations , https://developer.chrome.com/blog/devtools-annotations (2024-11-13), https://developer.chrome.com/blog/new-in-devtools-131 (2024-10-30 갱신)
- 확인한 내용: 개요 영역에서 드래그로 범위를 고르고 줌 버튼으로 "중첩 breadcrumb"을 쌓으며, 위 단계를 다시 누르면 돌아가고 부모 breadcrumb을 우클릭하면 자식이 지워집니다(문서가 breadcrumb이라 명시). Main 트랙은 flame chart이며 x는 시간, y는 호출 스택입니다. 주석은 항목 라벨, 항목 사이 화살표 연결, 시간 범위 라벨 세 종류이고 트레이스 파일 안에 저장됩니다. 인용: "Annotations can be saved in performance traces for sharing." 블로그에 따르면 기존 시간 범위 주석의 범위는 조정할 수 없고 지우고 다시 만들어야 합니다. Chrome 131 릴리스 노트에 Annotations 탭이 실려 있습니다.
- 빌릴 것: 줌 이력 = breadcrumb. 주석이 기록(트레이스) 파일에 함께 실립니다. 두 항목을 잇는 화살표 주석은 "경로 위 두 지점의 관계 표시"로 응용할 수 있습니다.
- 주의: 부모 breadcrumb 삭제가 자식을 지우는 파괴적 동작은 저장된 설명에는 위험합니다. 범위 조정 불가는 불편한 선례입니다.

**Firefox Profiler**
- 출처: https://raw.githubusercontent.com/firefox-devtools/profiler/main/docs-user/guide-ui-tour-timeline.md (공식 저장소 사용자 문서 원본)
- 확인한 내용: 인용: "These range selections can be committed by clicking the zoom button." 미리보기 선택은 드래그하는 동안 아래 패널을 갱신하고, 확정(commit)한 범위가 여러 개 쌓입니다. 이미지 설명문에 "위쪽 툴바의 확정 범위"라는 말이 있지만 breadcrumb이라는 단어는 문서에 없습니다. 확정 범위를 눌러 되돌아가는 동작은 문서로 확인하지 못했으므로 미확인입니다.
- 빌릴 것: "미리보기 선택(임시)"과 "확정 범위(줌 이력)"의 두 단계 구분.

## 5. 범위 주석과 겹침 (Hypothesis, W3C, Google Docs, GitHub)

- 출처: W3C Web Annotation Data Model (Recommendation, 2017-02-23) https://www.w3.org/TR/annotation-model/ ; Hypothesis 블로그 "Fuzzy Anchoring" (2013-04-22) https://web.hypothes.is/blog/fuzzy-anchoring/ ; Hypothesis 클라이언트 소스 https://raw.githubusercontent.com/hypothesis/client/main/src/styles/annotator/highlights.scss ; Google Docs 도움말 https://support.google.com/docs/answer/65129 ; GitHub Docs https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/commenting-on-a-pull-request
- 확인한 내용:
  - W3C: TextQuoteSelector는 exact와 앞뒤 문맥(prefix/suffix)으로, TextPositionSelector는 start/end 위치로 대상을 가리킵니다. 위치 선택자는 대상이 바뀌면 깨지기 쉬워 State(시점·판본 기술)를 함께 쓰라고 하고, 인용: "Multiple Selectors should select the same content". `refinedBy`는 선택자를 사슬로 정밀화하며, TimeState는 sourceDate로 대상 판본을 고정합니다. 문서의 뒤 36,619자는 읽지 못했습니다.
  - Hypothesis(블로그): 범위·위치·인용(앞뒤 32자) 선택자 세 가지를 저장하고 네 가지 전략을 순서대로 시도해 다시 붙입니다. 모두 실패했을 때의 처리는 글에 없습니다.
  - Hypothesis(소스 코드이며 도움말이 아님): 하이라이트가 겹치면 DOM 깊이에 따라 둘째·셋째 색을 쓰고, 주석에 "Highlights more than two levels deep are transparent by default."라고 적혀 있습니다. 즉 색 겹침으로 구분하는 방식은 두 단계가 한계입니다.
  - Google Docs: 선택한 범위에 댓글을 달고, 해결(Resolve)하면 닫히며 다시 열 수 있고 패널에서 열림/해결 필터가 있습니다. 원문이 바뀔 때 앵커가 어떻게 되는지, 댓글 이력은 도움말에 없어 미확인입니다.
  - GitHub: 줄 번호 클릭 후 드래그하거나 Shift를 누르고 마지막 줄 번호를 눌러 줄 범위를 고릅니다(인용: "hold down Shift and click on the last line number"). 기존 댓글에서 범위가 어떻게 표시되는지는 페이지에 없어 미확인입니다.
- 빌릴 것: 구간은 주 선택자(단계 id)와 보조 선택자(시각·오프셋)를 함께 저장하고, 대상 기록의 판본(State)을 붙입니다. 단계 이름 앞뒤를 문맥으로 저장하면 같은 요청의 다른 기록에서도 다시 붙일 수 있습니다(추론). 겹침이 둘을 넘으면 색 중첩 대신 레인을 씁니다(추론). Docs의 해결/재열기는 채택/폐기 토글과 유사합니다.
- 주의: 이 조사에서 "저장·버전·채택된 AI 설명"과 같은 모델의 선례는 찾지 못했으므로 그 부분은 우리가 설계해야 합니다.

## 6. 계층 단면 / 계층 가로지르는 선례

- 결과: 요청 한 건이 계층을 내려갔다 올라오는 동적 궤적을 보여 주는 단단한 선례는 찾지 못했습니다.
- 가장 가까운 확인 사례(모두 정적 구조도): Werner Fischer의 Linux Storage Stack Diagram(Thomas-Krenn, v6.18, CC BY-SA 4.0) https://www.thomas-krenn.com/en/wiki/Linux_Storage_Stack_Diagram 는 VFS에서 블록 계층·드라이버까지의 배치를 보여 줍니다. 그림 속 화살표의 존재와 요청 경로 표시 여부는 이미지를 열어 보지 않아 미확인입니다. Constantine Shulyupin의 인터랙티브 리눅스 커널 맵 https://makelinux.github.io/kernel/map/ 은 사용자 공간에서 하드웨어까지 계층을 SVG로 두고 드래그 이동, 휠 줌, 항목 링크를 지원합니다. 라이선스 표기는 확인하지 못했습니다.
- Brendan Gregg의 flame graph 설명 https://www.brendangregg.com/flamegraphs.html 은 구분을 분명히 합니다. flame graph의 x축은 알파벳순 모집단이며 인용: "is not the passage of time." y축은 스택 깊이이고, 시간을 x에 두는 것은 flame chart입니다. icicle은 flame graph를 위아래로 뒤집은 것입니다. 이 구분은 컨셉 C의 용어를 정할 때 중요합니다.
- 주의: Jaeger UI의 트레이스 폭포(waterfall)와 미니맵은 페이지를 열었으나 해당 내용이 없어 미확인입니다.

## 7. 점진적 공개 (NN/g)

- 출처: Jakob Nielsen, "Progressive Disclosure", NN/g, 2006-12-03. https://www.nngroup.com/articles/progressive-disclosure/
- 확인한 내용: 인용: "Progressive disclosure defers advanced or rarely used features to a secondary screen". 단계식 변형은 선형 순서로 일부 옵션씩 보이는 것이며, 단계가 서로 독립적일 때 잘 되고 서로 얽혀 계속 오가야 하면 실패한다고 합니다.
- 빌릴 것: 세부(함수·커널)는 요청이 있을 때만 펼칩니다. 기본 화면에는 개요와 경로만 둡니다.
- 주의: 이 경로는 단계가 서로 얽혀(커널 안에서 프로세스로 돌아오는 등) 있어서 NN/g가 말한 실패 조건에 가깝습니다. 단계를 오가는 전환 비용을 줄이는 장치가 필요합니다(추론).

---

## 세 컨셉에 주는 시사점

**컨셉 A (글자 그대로의 지도 + 시맨틱 줌 + 경로선)**
- 근거: Pad의 시맨틱 줌, Furnas의 DOI, Google Maps의 파란 기본/회색 대안 + 시간 라벨 + Details 목록.
- 장점: 줌 수준마다 표시가 달라지는 것(시스템 → 프로세스 → 함수)이 가장 직접적입니다. 대안 경로 선택 UX가 검증된 관용구입니다.
- 위험: 공간 거리가 의미 없는 계층 구조에 지리적 은유를 쓰면 위치·크기가 거짓 정보를 줄 수 있습니다(추론). 경로가 계층을 오갈 때 지도 위에서 어디서 출발해 어디서 끝나는지 읽히는지가 검증 대상입니다.

**컨셉 B (수직 계층 단면, 경로가 내려갔다 올라옴)**
- 근거: 동적 선례는 없고 정적 계층도(Thomas-Krenn, makelinux 맵)뿐입니다. 선례 부재는 실패의 증거가 아니라 미검증이라는 뜻입니다.
- 장점: 모든 계층을 동시에 보이는 구조라 Furnas DOI를 "초점 계층은 크게, 나머지는 얇게"로 곧바로 쓸 수 있습니다.
- 위험: NN/g가 말한 단계 간 얽힘 문제가 가장 크게 나타납니다. 내려갔다 올라오는 같은 지점을 한 경로로 읽히게 하는 장치가 필요합니다.

**컨셉 C (시간 트랙 트레이스 + flame/icicle 깊이)**
- 근거: Perfetto의 트랙·핀·영역 선택(시작+끝+트랙 목록)·선택 시 플로우 화살표·탭 서랍, Chrome의 flame chart(y = 호출 깊이)와 주석 저장.
- 장점: 구간 선택과 중첩(깊이)이 이미 해결된 어휘를 가집니다. 겹침은 트랙/레인으로 풉니다.
- 위험: x가 시간이면 마이크로초 단위 커널 구간은 점이 되고 프로세스 대기는 길어져 균형이 깨집니다(추론). 시간 비례 축과 단계 균등 축을 선택할 수 있어야 합니다. 이 컨셉의 "flame/icicle"은 엄밀히 flame chart(x가 시간)이고 Gregg의 flame graph가 아닙니다.

**범위 설명 규칙에 대한 매핑**

| 규칙 | 선례에서 확인한 것 | 우리 규칙 제안 (추론) |
| --- | --- | --- |
| 단계 경계로 정규화 | Perfetto 영역 선택 = 시작+끝+트랙. Chrome은 기존 시간 범위 주석을 조정 못 함. W3C: 선택자 여럿이 같은 내용을 가리켜야 함 | 드래그를 놓는 순간 가장 가까운 단계 경계로 맞추고 미리보기를 보여 줍니다. 저장은 단계 id(주) + 시각(보조) + 문맥(앞뒤 단계명)으로 합니다 |
| 겹침 레인 | Hypothesis는 색 겹침이 두 단계에서 한계. Perfetto는 트랙으로 분리 | 겹치는 구간은 마커 레인을 늘려 배치하고 색 중첩에 의존하지 않습니다 |
| 중첩 | flame chart y축 = 호출 깊이. Hypothesis는 DOM 깊이로 둘째·셋째 색 | 포함 관계(시작·끝)로 계산하고 바깥 → 안쪽 순으로 들여써 보입니다. 깊이 상한을 두고 초과분은 접습니다 |
| 버전 / 채택 | W3C State/TimeState가 판본을 고정. Chrome 주석은 트레이스 파일에 저장. Docs의 해결/재열기 | 설명은 대상 기록 판본에 묶고 판본 번호를 둡니다. 채택/폐기는 토글로 하고 이전 판본은 지우지 않습니다. 이 모델의 직접 선례는 찾지 못했습니다 |
| 줌 이력 = breadcrumb | Chrome: 중첩 breadcrumb(문서 명시). Firefox: 확정 범위 여럿(breadcrumb 용어는 문서에 없음). Shneiderman: history 작업 | 확정한 줌/범위를 쌓아 이력으로 두되 부모를 누르면 자식을 지우지 말고 앞으로 가기를 남깁니다. 저장된 설명은 이력과 별도 데이터입니다 |

**미확인 모음**: Apple Maps 본문, Perlin & Fox의 쪽수·DOI, Shneiderman의 IEEE DOI, Firefox Profiler의 breadcrumb 동작 문구, Google Docs의 앵커 변경 처리와 댓글 이력, GitHub의 기존 범위 댓글 표시, W3C 문서 후반부, Hypothesis의 앵커 실패 처리와 도움말 수준의 겹침 설명, Jaeger 폭포 뷰, Thomas-Krenn 그림의 경로 표시.
