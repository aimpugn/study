# Programmers 80 Repetition Queue

이 문서는 1번부터 차근차근 도장깨기 하듯 풀 수 있는 반복 문제장입니다.
고정 로드맵은 아니지만, 무엇을 풀지 모르겠을 때는 앞 번호부터 진행하면 자연스럽게 손이 풀리고 유형 감각이 넓어지도록 배치합니다.

기본 축은 프로그래머스 고득점 Kit입니다.
프로그래머스는 이 Kit을 코딩테스트 결과를 분석해 자주 나오는 유형과 많이 틀리는 유형을 간추린 세트라고 설명합니다.
여기에 카카오, PCCP, 최근 Lv.2~Lv.3 문제를 섞어 실전 감각을 보강합니다.

이 큐는 공식 난이도 `Lv.1~Lv.3` 중심입니다.
현재 목표는 어려운 `Lv.4`를 억지로 뚫는 것이 아니라, `Lv.2`를 안정적으로 풀고 `Lv.3`에 들어가도 무너지지 않는 저점을 만드는 것입니다.

## Stage

1. `1~20`: 재활 구간입니다. 해시, 스택/큐, 정렬, 단순 구현처럼 바로 손이 움직여야 하는 문제를 먼저 풉니다.
2. `21~40`: 핵심 Lv.2 구간입니다. 완전탐색, DFS, 그리디, 문자열 파싱, 간단한 DP가 섞입니다.
3. `41~60`: Lv.2 상단에서 Lv.3 입구로 넘어가는 구간입니다. 전처리, 이분탐색, 격자, 사각형 업데이트 감각을 만듭니다.
4. `61~80`: Lv.3 실전 구간입니다. 그래프, 최단거리, 상태 탐색, DP, 다익스트라처럼 생각의 모양을 먼저 잡아야 하는 문제를 풉니다.

번호가 항상 정확한 난이도 순서는 아닙니다.
그래도 앞 구간은 손풀이와 기본 자료구조, 뒤 구간은 문제 변환과 최적화 감각이 더 많이 필요하도록 배치했습니다.

## 사용법

상태는 가볍게만 표시합니다.

```text
TODO   아직 시작하지 않음
R1     처음 풀었거나 풀이를 이해함
R2     2~3일 뒤 다시 구현함
R3     1주 뒤 풀이 흐름을 설명함
R4     2~3주 뒤 시간 제한으로 다시 풀었음
KEEP   나중에 또 볼 가치가 큼
SKIP   지금은 우선순위 낮음
```

문제별로 최소한 하나의 기억 문장을 남깁니다.
예를 들면 `연속 구간 합이 반복되면 누적합`, `다음에 처리할 가장 작은 값이 필요하면 힙`, `최댓값을 직접 찾기 어려우면 답을 이분탐색`처럼 다음 문제에서 다시 떠올릴 수 있는 문장이어야 합니다.

## 오늘 고르는 법

1. 앞에서부터 아직 TODO인 문제 하나를 고릅니다.
2. 이미 R1 이상인 쉬운 문제 하나를 15~25분 안에 다시 풉니다.
3. 막혔던 문제 하나는 코드 없이 5분 설명만 해 봅니다.

막히면 오래 버티기보다 멈춘 위치에 주석을 남깁니다.
그 주석이 다음 첨삭 지점입니다.

## Queue

| No | Track | ID | Problem | Lv | Main Pattern | First Trigger | Status | Next |
| ---: | --- | ---: | --- | --- | --- | --- | --- | --- |
| 1 | Stage1/Hash | 42576 | 완주하지 못한 선수 | 1 | HashMap | 이름별 개수를 세고 하나만 남는 값을 찾는다. | TODO |  |
| 2 | Stage1/Hash | 1845 | 폰켓몬 | 1 | Set | 고를 수 있는 수와 종류 수 중 작은 값을 고른다. | TODO |  |
| 3 | Stage1/StackQueue | 12906 | 같은 숫자는 싫어 | 1 | Stack | 직전 값과 같으면 버린다. | TODO |  |
| 4 | Stage1/Sort | 42748 | K번째수 | 1 | Sort | 자르고 정렬하고 k번째를 고른다. | TODO |  |
| 5 | Stage1/BruteForce | 86491 | 최소직사각형 | 1 | Normalize | 각 명함을 긴 쪽과 짧은 쪽으로 정규화한다. | TODO |  |
| 6 | Stage1/BruteForce | 42840 | 모의고사 | 1 | Pattern | 반복 패턴의 인덱스를 나머지로 맞춘다. | TODO |  |
| 7 | Stage1/Greedy | 42862 | 체육복 | 1 | Greedy | 잃어버렸고 여벌도 있는 학생을 먼저 제거한다. | TODO |  |
| 8 | Stage1/StackQueue | 12909 | 올바른 괄호 | 2 | Stack/Counter | 열린 수가 음수가 되면 실패, 끝에 0이어야 한다. | TODO |  |
| 9 | Stage1/StackQueue | 42586 | 기능개발 | 2 | Queue | 각 작업의 완료일을 만들고 앞에서부터 묶는다. | TODO |  |
| 10 | Stage1/StackQueue | 42587 | 프로세스 | 2 | Queue/Priority | 더 높은 우선순위가 남아 있으면 뒤로 보낸다. | TODO |  |
| 11 | Stage1/StackQueue | 42583 | 다리를 지나는 트럭 | 2 | Queue/Simulation | 시간 흐름과 다리 위 무게를 같이 관리한다. | TODO |  |
| 12 | Stage1/StackQueue | 42584 | 주식가격 | 2 | Monotonic Stack | 가격이 떨어지는 첫 시점을 스택으로 찾는다. | TODO |  |
| 13 | Stage1/Hash | 42577 | 전화번호 목록 | 2 | Sort/Hash | 접두어 관계는 정렬 후 이웃 비교로 줄일 수 있다. | TODO |  |
| 14 | Stage1/Hash | 42578 | 의상 | 2 | Counting/Product | 종류별 선택 수를 곱하고 아무것도 안 입는 경우를 뺀다. | TODO |  |
| 15 | Stage1/Sort | 42747 | H-Index | 2 | Sort | 인용 수와 논문 수의 경계를 찾는다. | TODO |  |
| 16 | Stage1/BruteForce | 42842 | 카펫 | 2 | Divisor Search | 넓이의 약수 후보 중 테두리 조건을 맞춘다. | TODO |  |
| 17 | Stage1/Stack | 12973 | 짝지어 제거하기 | 2 | Stack | 같은 문자가 연속되면 제거한다. | TODO |  |
| 18 | Stage1/Kakao | 17680 | 캐시 | 2 | LRU | 최근 사용 순서를 갱신한다. | TODO |  |
| 19 | Stage1/Kakao | 64065 | 튜플 | 2 | Parsing/Set | 길이가 짧은 집합부터 새 원소를 찾는다. | TODO |  |
| 20 | Stage1/Simulation | 12985 | 예상 대진표 | 2 | Simulation | 번호를 다음 라운드 번호로 계속 접는다. | TODO |  |
| 21 | Stage2/Heap | 42626 | 더 맵게 | 2 | PriorityQueue | 항상 가장 작은 두 값을 꺼내 섞는다. | TODO |  |
| 22 | Stage2/BruteForce | 42839 | 소수 찾기 | 2 | Permutation/Prime | 숫자 조각으로 가능한 수를 만들고 소수 판정한다. | TODO |  |
| 23 | Stage2/BruteForce | 87946 | 피로도 | 2 | DFS/Permutation | 던전 순서를 모두 시도하되 현재 피로도로 가지치기한다. | TODO |  |
| 24 | Stage2/BruteForce | 86971 | 전력망을 둘로 나누기 | 2 | Tree/BFS | 간선 하나를 끊고 두 컴포넌트 크기를 비교한다. | TODO |  |
| 25 | Stage2/BruteForce | 84512 | 모음사전 | 2 | DFS/Order | 사전 순 DFS 방문 순서가 곧 번호다. | TODO |  |
| 26 | Stage2/Greedy | 42860 | 조이스틱 | 2 | Greedy/String | 알파벳 변경 비용과 좌우 이동 최소를 분리한다. | TODO |  |
| 27 | Stage2/Greedy | 42883 | 큰 수 만들기 | 2 | Monotonic Stack | 앞 숫자가 작고 제거 기회가 있으면 뺀다. | TODO |  |
| 28 | Stage2/Greedy | 42885 | 구명보트 | 2 | Two Pointers | 가장 무거운 사람과 가장 가벼운 사람을 같이 태울 수 있는지 본다. | TODO |  |
| 29 | Stage2/DFS-BFS | 43165 | 타겟 넘버 | 2 | DFS | 각 숫자 앞에 `+` 또는 `-`를 붙이는 결정 트리다. | TODO |  |
| 30 | Stage2/DFS-BFS | 1844 | 게임 맵 최단거리 | 2 | BFS/Grid | 가중치 없는 최단거리는 BFS다. | TODO |  |
| 31 | Stage2/Math | 12924 | 숫자의 표현 | 2 | Two Pointers/Math | 연속된 자연수 합을 구간으로 본다. | TODO |  |
| 32 | Stage2/DP | 12914 | 멀리 뛰기 | 2 | DP | 마지막 점프가 1칸인지 2칸인지로 나눈다. | TODO |  |
| 33 | Stage2/DP | 12913 | 땅따먹기 | 2 | DP | 같은 열을 연속으로 밟지 않는 최대 합을 누적한다. | TODO |  |
| 34 | Stage2/Kakao | 17684 | 압축 | 2 | Map/Simulation | 사전을 늘리며 가장 긴 현재 문자열을 찾는다. | TODO |  |
| 35 | Stage2/Kakao | 17686 | 파일명 정렬 | 2 | Sort/Parsing | HEAD와 NUMBER를 분리해 안정 정렬한다. | TODO |  |
| 36 | Stage2/Kakao | 42888 | 오픈채팅방 | 2 | Map/Log | 마지막 닉네임과 출입 로그를 분리한다. | TODO |  |
| 37 | Stage2/Kakao | 42890 | 후보키 | 2 | Combination/Set | 유일성과 최소성을 둘 다 만족해야 한다. | TODO |  |
| 38 | Stage2/Kakao | 60057 | 문자열 압축 | 2 | String/BruteForce | 자르는 길이를 모두 시도한다. | TODO |  |
| 39 | Stage2/Kakao | 60058 | 괄호 변환 | 2 | Recursion/String | 균형 잡힌 접두사를 기준으로 재귀 변환한다. | TODO |  |
| 40 | Stage2/Kakao | 67257 | 수식 최대화 | 2 | Permutation/Parsing | 연산자 우선순위를 모두 시도한다. | TODO |  |
| 41 | Stage3/Kakao | 72411 | 메뉴 리뉴얼 | 2 | Combination/Counting | 코스 길이별 조합 빈도를 센다. | TODO |  |
| 42 | Stage3/Kakao | 72412 | 순위 검색 | 2 | Precompute/BinarySearch | 조건 조합별 점수 목록을 만들고 이분탐색한다. | TODO |  |
| 43 | Stage3/Kakao | 81302 | 거리두기 확인하기 | 2 | Grid/BFS | 사람 주변 거리 2 안의 파티션 조건을 본다. | TODO |  |
| 44 | Stage3/Kakao | 92341 | 주차 요금 계산 | 2 | Map/Time | 입차 시각을 저장하고 출차 때 누적한다. | TODO |  |
| 45 | Stage3/Kakao | 92342 | 양궁대회 | 2 | DFS/Combination | 점수를 얻을 과녁 조합을 탐색한다. | TODO |  |
| 46 | Stage3/Kakao | 118667 | 두 큐 합 같게 만들기 | 2 | Two Pointers/Queue | 한쪽 합이 크면 그쪽에서 빼서 반대로 보낸다. | TODO |  |
| 47 | Stage3/Kakao | 150368 | 이모티콘 할인행사 | 2 | DFS/BruteForce | 할인율 조합을 모두 시도한다. | TODO |  |
| 48 | Stage3/Kakao | 150369 | 택배 배달과 수거하기 | 2 | Greedy/Stack | 먼 집부터 배달과 수거 잔량을 처리한다. | TODO |  |
| 49 | Stage3/Stack | 154539 | 뒤에 있는 큰 수 찾기 | 2 | Monotonic Stack | 아직 더 큰 수를 못 찾은 인덱스를 스택에 둔다. | TODO |  |
| 50 | Stage3/Grid | 159993 | 미로 탈출 | 2 | BFS/Grid | 시작-레버, 레버-출구를 따로 최단거리로 구한다. | TODO |  |
| 51 | Stage3/Grid | 250136 | 석유 시추 | 2 | BFS/Union | 석유 덩어리 크기를 구하고 열별 합산한다. | TODO |  |
| 52 | Stage3/Recent | 468379 | 선인장 숨기기 | 2 | 2D Window/Prefix | 구역 안 최솟값을 빠르게 구해 가장 늦게 젖는 위치를 찾는다. | TODO |  |
| 53 | Stage3/Hash | 42579 | 베스트앨범 | 3 | HashMap/Sort | 장르 합계와 장르 안 순위를 따로 정렬한다. | TODO |  |
| 54 | Stage3/Heap | 42627 | 디스크 컨트롤러 | 3 | PriorityQueue | 도착한 작업 중 처리 시간이 짧은 것을 고른다. | TODO |  |
| 55 | Stage3/Heap | 42628 | 이중우선순위큐 | 3 | TreeMap/PriorityQueue | 최솟값과 최댓값 삭제 정책을 동시에 유지한다. | TODO |  |
| 56 | Stage3/Greedy | 42861 | 섬 연결하기 | 3 | MST | 모든 섬을 잇는 최소 비용은 최소 신장 트리다. | TODO |  |
| 57 | Stage3/Greedy | 42884 | 단속카메라 | 3 | Interval Greedy | 진출 지점 기준으로 정렬하고 카메라 위치를 고정한다. | TODO |  |
| 58 | Stage3/DP | 42895 | N으로 표현 | 3 | DP/Set | `N`을 i번 써서 만들 수 있는 수 집합을 누적한다. | TODO |  |
| 59 | Stage3/DP | 43105 | 정수 삼각형 | 3 | DP | 위에서 내려온 최대 합만 남긴다. | TODO |  |
| 60 | Stage3/DP | 42898 | 등굣길 | 3 | DP/Grid | 왼쪽과 위쪽에서 오는 경우의 수를 더한다. | TODO |  |
| 61 | Stage4/DFS-BFS | 43162 | 네트워크 | 3 | DFS/BFS | 연결 컴포넌트 개수를 센다. | TODO |  |
| 62 | Stage4/DFS-BFS | 43163 | 단어 변환 | 3 | BFS | 한 글자 차이인 단어를 간선으로 보는 최단거리다. | TODO |  |
| 63 | Stage4/DFS-BFS | 87694 | 아이템 줍기 | 3 | BFS/Geometry | 테두리를 격자로 확대한 뒤 최단거리로 본다. | TODO |  |
| 64 | Stage4/DFS-BFS | 43164 | 여행경로 | 3 | DFS/Backtracking | 티켓을 모두 쓰는 경로 중 사전순 첫 경로를 찾는다. | TODO |  |
| 65 | Stage4/DFS-BFS | 84021 | 퍼즐 조각 채우기 | 3 | BFS/Shape | 빈칸 모양과 조각 모양을 정규화해 비교한다. | TODO |  |
| 66 | Stage4/BinarySearch | 43238 | 입국심사 | 3 | Parametric Search | 시간 `t` 안에 몇 명을 처리할 수 있는지 판정한다. | TODO |  |
| 67 | Stage4/Graph | 49189 | 가장 먼 노드 | 3 | BFS/Graph | 1번 노드에서의 최단거리 최댓값 개수를 센다. | TODO |  |
| 68 | Stage4/Graph | 49191 | 순위 | 3 | Floyd/Graph | 모든 상대와 승패 관계가 정해진 선수를 센다. | TODO |  |
| 69 | Stage4/Heap | 12927 | 야근 지수 | 3 | PriorityQueue | 가장 큰 일을 줄이는 선택을 반복한다. | TODO |  |
| 70 | Stage4/Kakao | 60059 | 자물쇠와 열쇠 | 3 | Matrix/BruteForce | 키 회전과 이동을 확장 보드에서 검사한다. | TODO |  |
| 71 | Stage4/Kakao | 60061 | 기둥과 보 설치 | 3 | Simulation/Validation | 매 명령 후 전체 구조가 유효한지 확인한다. | TODO |  |
| 72 | Stage4/Kakao | 60062 | 외벽 점검 | 3 | Permutation/Circular | 원형을 펼치고 친구 순서를 시도한다. | TODO |  |
| 73 | Stage4/Kakao | 67258 | 보석 쇼핑 | 3 | Two Pointers | 모든 종류를 포함하는 가장 짧은 구간을 유지한다. | TODO |  |
| 74 | Stage4/Kakao | 67259 | 경주로 건설 | 3 | BFS/DP | 위치뿐 아니라 진입 방향까지 비용 상태다. | TODO |  |
| 75 | Stage4/Kakao | 72413 | 합승 택시 요금 | 3 | Floyd/Dijkstra | 합승 지점을 하나 정하고 세 구간 비용을 더한다. | TODO |  |
| 76 | Stage4/Kakao | 92343 | 양과 늑대 | 3 | DFS/State | 갈 수 있는 후보 노드 집합을 상태로 들고 간다. | TODO |  |
| 77 | Stage4/Kakao | 92344 | 파괴되지 않은 건물 | 3 | 2D Prefix Sum | 사각형 업데이트를 차분 배열로 모은다. | TODO |  |
| 78 | Stage4/Kakao | 118668 | 코딩 테스트 공부 | 3 | DP | 알고력과 코딩력을 좌표로 보는 최단 비용이다. | TODO |  |
| 79 | Stage4/Kakao | 118669 | 등산코스 정하기 | 3 | Dijkstra | 출입구에서 산봉우리까지의 최대 간선 비용을 최소화한다. | TODO |  |
| 80 | Stage4/Greedy | 12987 | 숫자 게임 | 3 | Greedy/Sort | 이길 수 있는 가장 작은 수로 상대 카드를 이긴다. | TODO |  |

## Review Slots

아래는 실제로 다시 풀면서 채웁니다.
문제 큐 전체를 한 번에 완성하려고 하지 말고, 막힌 문제만 짧게 남깁니다.

| Date | Problem | Attempt | What got stuck | Takeaway |
| --- | --- | --- | --- | --- |
|  |  |  |  |  |

## Sources

- [Programmers Coding Test High Score Kit](https://school.programmers.co.kr/learn/challenges?tab=algorithm_practice_kit)
- [Programmers Hash Kit](https://school.programmers.co.kr/learn/courses/30/parts/12077)
- [Programmers Stack/Queue Kit](https://school.programmers.co.kr/learn/courses/30/parts/12081)
- [Programmers Heap Kit](https://school.programmers.co.kr/learn/courses/30/parts/12117)
- [Programmers Sort Kit](https://school.programmers.co.kr/learn/courses/30/parts/12198)
- [Programmers Brute Force Kit](https://school.programmers.co.kr/learn/courses/30/parts/12230)
- [Programmers Greedy Kit](https://school.programmers.co.kr/learn/courses/30/parts/12244)
- [Programmers Dynamic Programming Kit](https://school.programmers.co.kr/learn/courses/30/parts/12263)
- [Programmers DFS/BFS Kit](https://school.programmers.co.kr/learn/courses/30/parts/12421)
- [Programmers Binary Search Kit](https://school.programmers.co.kr/learn/courses/30/parts/12486)
- [Programmers Graph Kit](https://school.programmers.co.kr/learn/courses/30/parts/14393)
