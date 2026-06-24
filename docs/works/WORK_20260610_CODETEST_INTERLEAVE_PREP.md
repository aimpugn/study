# WORK 20260610 — 코테 대비 LC 교차 바퀴 (쿠팡·네이버제트 가정)

상태: ACTIVE · 시작 2026-06-10
연계: `WORK_20260511_PROGRAMMERS_JAVA_REHAB_SETUP.md`(프로그래머스 레인, 계속) · `books/easy_to_learn_algorithms/INDEX.md`(책은 막힌 직후 조회) · 회고 = algorithm-retrospective 스킬

## 운영 규칙

- **풀이 프로토콜**: ⓪ 유형 판별 + 판단 근거 말하기(2바퀴부터) → ① 작은 예제 손 실행(코드 금지) → ② 불변식 한 문장("i 처리 직후 \_\_는 항상 \_\_다" — 세우는 법은 `algorithms/problem_solving_growth.md`의 "구체 시뮬은 발견, 기호 시뮬은 증명": 공존하는 둘을 집어 규칙이 금지하는 관계를 묻고 그 부정을 불변식으로) → ③ 문장을 반례로 공격(빈/1개/전부 겹침/역순/동값) → ④ 코드는 문장의 받아쓰기. **10분 룰**: 문장이 안 잡히면 브루트포스(완전성 자명)로 정답 먼저.
- **바퀴**: k바퀴 = 각 유형의 k번째 문제. 1바퀴는 유형 공개·순서대로, **2바퀴부터 셔플 + 유형 비공개**(AI가 문제만 제시).
- **부스터**: 노힌트 실패 + 불변식 인출 실패한 유형은 다음 세션에 그 유형 1문제 삽입.
- **측정**(회고에 기록): 노힌트 정답 여부 · 첫 불변식 문장까지 시간.
- **난이도 양방향 튜닝**: 노힌트 + ~15분 + 문장이 한 번에 나옴 → 카드만 만들고 그 유형의 다음(더 어려운) 칸으로 점프. 40분 초과 + 문장 실패 → 브루트포스로 마감, 다음 세션에 같은 유형 부스터 1문제. 이 두 규칙이 풀 난이도를 자가 보정한다(기준점: LC 56 = 힌트 동반 완주가 현재 레벨의 훈련 존).
- **세션 운영**: 채팅 1개 = 세션 1개(1~3문제). **회고는 그 채팅을 닫기 전에 완료**(세션의 실수 디테일은 채팅 밖으로 안 나감). 새 채팅 오프너는 "알고리즘 세션 시작" 한 줄이면 충분(메모리가 계획·프로토콜·진도 출처를 들고 있음).
- **라이브 모의**: 3주차부터 주 1회, 45~60분, 영어 지문, think-aloud, AI=면접관.
- 교차 연습은 블록 연습보다 체감 진도가 느리고 자주 막히는 게 정상(작동 신호).

## 단기 집중 코테 보강 (활성, 4~6주 · 주 6~12h, 2026-06-24~)

코테/라이브 데드라인 가정으로 아래 바퀴를 시간 예산 안에 재정렬한 오버레이다. 주 9h × 5주 ≈ 45h라 NeetCode 150 완주가 아니라 **패턴 커버리지**가 목표 — 개수가 아니라 모든 패턴 1회+ 깊게 + 빠진 패턴 우선 + 인식 reps. 아래 바퀴(1·2·3바퀴)는 이 오버레이가 끌어 쓰는 문제 풀이다.

빠진 패턴(이 플랜 바퀴에 없던 것, 코테/라이브 빈출) 우선 — LeetCode, Java:

| 주 | 패턴 | 문제 |
|---|---|---|
| 1 | 연결 리스트 + 트리 기본 | 206 역순, 21 병합, 141 사이클, 143 재배열 / 226 뒤집기, 104 최대깊이, 102 레벨순회 |
| 2 | 트리/BST 심화 + 힙 | 235 LCA(BST), 98 검증, 543 지름 / 215 K번째, 973 K최근접, 295 스트림 중앙값(두 힙) |
| 3 | 트라이 + 행렬 + 비트 | 208 트라이, 211 추가/검색 / 48 회전, 54 나선, 73 0채우기 / 136, 338, 268 |
| 4 | 인터리브 강화(유형 비공개) | 2바퀴 풀에서 셔플: 560, 11, 33, 994, 207(위상), 322, 39 중 6~7 |
| 5 | 모의 + 약점 보강 | 타임드 2문제 mixed ×2~3(45~60분) + 막힌 패턴 reps + 심화(23, 105, 212) 선택 |

(6주차면 5주 내용 R복습 + 모의 버퍼.)

2단 reps(깊이+양 동시):

- 새 패턴 첫 문제 -> 풀이 프로토콜 + 불변식 한 문장 + 회고(PROCESS.md + PROCESS.html).
- 같은 패턴 추가 문제 -> 풀이 + 불변식 한 문장 + 30초 카드만(회고 생략, 빠르게 인식 reps).
- R복습 -> 빈 화면 인출. (회고를 *할 때는* 항상 시각화이되, 추가 reps는 회고를 안 하므로 매 문제 시각화는 아니다.)

막히면 `algorithms/problem_solving_growth.md`의 "구체 시뮬은 발견, 기호 시뮬은 증명"으로 불변식을 세운다. 진도·노힌트·첫 불변식까지 시간은 아래 세션 로그에 남긴다.

## 1바퀴 — 유형 공개, 순서대로 (얇은 유형만 2문제 블록)

기준점 캘리브레이션(2026-06-10): LC 56을 힌트 동반으로 완주했고 막힌 곳이 전부 개념 단계(정렬 키, 비교 상대)였음 → 구현력이 아니라 패턴 선택·불변식 수립이 병목. 따라서 trivial easy는 전부 제외하고 easy-medium~medium에서 시작.

- [ ] 해시: LC 49 Group Anagrams (medium, 애너그램 정규화 — p86491 정규화 전이와 연결)
- [ ] 윈도우: LC 3 Longest Substring Without Repeating Characters (medium)
- [ ] 구간: **mergeintervals R2 (06-12, 키워드 인출) → 직후 LC 57 Insert Interval (medium)** (인출→근전이 콤보, 순서 고정)
- [ ] 스택: LC 739 Daily Temperatures (medium, 단조 스택)
- [ ] 이분탐색②: LC 875 Koko Eating Bananas (medium, 답 이분탐색 = p43238 입국심사와 같은 패턴) + LC 153 Find Minimum in Rotated Sorted Array (medium, 경계 불변식)
- [ ] BFS/DFS: LC 200 Number of Islands (medium)
- [ ] 그리디: LC 55 Jump Game (medium)
- [ ] union-find②: LC 547 Number of Provinces + LC 684 Redundant Connection (medium×2)
- [ ] DP②: LC 198 House Robber + LC 62 Unique Paths (medium×2, 62=책 9장 행렬 경로)
- [ ] 백트래킹②: LC 78 Subsets + LC 46 Permutations (medium×2)

## 2바퀴 — 셔플·유형 비공개 (AI가 임의 순서로 제시)

- [ ] 구간: LC 75 Sort Colors · LC 435 Non-overlapping Intervals
- [ ] 해시: LC 560 Subarray Sum Equals K
- [ ] 투포인터·윈도우: LC 11 Container With Most Water · LC 209 Minimum Size Subarray Sum
- [ ] 스택: LC 155 Min Stack (설계형) · LC 394 Decode String
- [ ] 이분탐색: LC 33 Search in Rotated Sorted Array (153의 확장)
- [ ] BFS: LC 994 Rotting Oranges · LC 207 Course Schedule (위상정렬)
- [ ] 그리디: LC 134 Gas Station
- [ ] DP: LC 322 Coin Change
- [ ] 백트래킹: LC 39 Combination Sum

## 3바퀴+ — 중간 난도 마감

- [ ] LC 215 Kth Largest · LC 424 Longest Repeating Character Replacement · LC 1011 Capacity to Ship(875 패턴 강화) · LC 695 Max Area of Island · LC 1143 LCS(=책 9장) · LC 300 LIS · LC 79 Word Search
- [ ] LC 146 LRU Cache (LLD 트랙 연계, 쿠팡 라이브 빈출 보고)

## 부스터 예비 (쉬움 — 유형 붕괴 시에만 투입, 평소 사용 금지)

LC 1 · 20 · 125 · 704 · 278 · 35 · 455 · 70

## 프로그래머스 레인 — 주 1회

- [ ] 잔여 Lv1(42840 모의고사 + 42862 체육복)은 한 세션에 묶어 빠르게 소진
- [ ] 이후 카카오 기출 Lv2 합류 (네이버식 구현·긴 지문) — 같은 양방향 튜닝 규칙 적용

## 라이브 모의 로그

- [ ] #1 (3주차~):

## 세션 로그

| 날짜 | 문제 | 노힌트 | 첫 불변식까지 | 메모 |
|---|---|---|---|---|
| | | | | |
