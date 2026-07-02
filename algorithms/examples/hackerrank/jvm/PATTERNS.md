# 라이브 코테 실전 문서 (파일명은 PATTERNS지만, 이게 최종 준비 문서)

> 문법만 `cheatsheet/Cheatsheet.java`, 나머지는 이 문서 하나. (PLAYBOOK의 멘탈 카드는 7·8장으로 흡수됨)
> **당일 아침**: 0장(대본) -> 6장(함정 표) -> 7장(응급 처치)만 다시 훑는다. 1~5장은 어제까지 몸에 넣은 것 — 실전 중엔 못 펴 본다.
> 문제가 오면: 1번 표로 도구를 고르고, 2번 카드에서 **움직임 + 골격 + 불변식(왜 맞나)** 을 한 카드로 본다.
> `불변식:` = 그 자료구조가 항상 뜻하는 한 문장. 이게 유지되면 답이 맞다 — 카드별 사고 흐름의 핵심.

---

## 0. 실전 대본 — 1시간 2문제, 문제당 ~25분 (버퍼 10분)

> 한 줄 진실: **"얼기"는 몰라서가 아니라 혼자 침묵해서 생긴다. 입을 열면 안 생긴다.**
> 면접관은 심판이 아니라 페어. 평가 = 정답 < 과정·소통. 이 대본을 소리 내는 것 자체가 점수다.

**[0~3분] 명료화 — 코드 창 만지지 않는다. 묻는다 (질문 = 점수 + 생각할 시간):**
- "먼저 입력/출력과 제약을 확인하겠습니다." <- 시작 멘트, 이 한 문장으로 안정 진입
- "n이 최대 얼마인가요? 값 범위는요?" (복잡도 판단 + 오버플로 판단의 재료)
- "빈 입력, 원소 1개, 중복, 음수 같은 엣지도 처리 대상인가요?"
- **샘플 1개를 문제의 규칙대로 손으로 굴려 기대 출력이 나오는지 확인** — 안 나오면 규칙을 잘못 읽은 것 (Castle 오독이 여기서 걸렸어야 했다)

**[3~8분] 접근 + 복잡도 — 아직 코딩 금지. 여기서 합의 받고 시작:**
1. "가장 단순한 완전탐색은 X이고 O(?)입니다. n=? 대입하면 약 ?회라 통과/초과입니다." (4장 4박자)
2. 초과면 병목을 말로: "어디가 비싼가 — 중복 계산? 매번 재스캔? 매번 정렬?"
3. 1번 표로 도구 지목: "최단 거리니까 BFS로 가겠습니다." + 그 카드의 움직임 그림을 머리에
4. 불변식 한 문장: "i 처리 직후 ___는 항상 ___입니다."
5. 합의: "이 접근으로 구현해도 될까요?" <- 방향 틀렸으면 면접관이 여기서 잡아 준다 (공짜 보험)

**[8~20분] 구현 — 코드 = 불변식 받아쓰기. 내레이션 유지:**
- "지금 ...를 만들고 있고, 이 변수는 ...를 뜻합니다."
- 사소한 분기 고민(continue를 넣을까? 0도 마크할까?)이 오면: **정확성에 무관하면 아무거나 골라 선언하고 진행** — "이건 어느 쪽이든 결과가 같으니 A로 가겠습니다." (Connected Cells에서 여기 걸려 생각이 틀어졌었다)
- 손이 멈추면 입이라도: "지금 X 때문에 막혔는데 Y와 Z를 고민 중입니다." (말하다 풀리는 경우가 많고, 침묵보다 항상 낫다)

**[20~25분] 테스트 + 엣지 + 마무리:**
- 샘플 1개를 코드 순서대로 손 트레이스 (변수 값을 실제 숫자로 말하며)
- 엣지 3종을 소리 내며 확인: 빈/최소(n=1), 경계(전부 같음, 최대값), 문제 특이 케이스
- 마무리 멘트: "시간 O(?), 공간 O(?)입니다. 더 줄이려면 ...도 가능합니다."

**막히면 즉시 7장 사다리. 30초 침묵 = 알람.**

---

## 1. 신호 -> 도구 (빠른 인덱스)

| 문제에 이 말이 보이면 | 도구 | 카드 |
|---|---|---|
| 최단 거리/최소 횟수 + 가중치 없음 | BFS | 2-A |
| 연결 덩어리/영역/전부 방문/개수 | flood fill (DFS or BFS) | 2-B |
| 한 방향으로 막힐 때까지 미끄러짐 | BFS + 방향마다 while | 2-C |
| 가장 작은/큰 것을 반복해 꺼냄 | 힙 | 2-D |
| 같은 부분문제를 다시 계산(겹침) | DP/메모 | 2-E |
| 정렬된 것에서 찾기/경계 | 이분탐색 | 2-F |
| 연속 구간/창 | 슬라이딩 윈도우 | 2-G |
| 최근 것부터/짝 맞추기 | 스택 | 2-H |
| 모든 조합·순열 생성 | 백트래킹 | 2-I |
| 문자 빈도가 다 같은지/거의 같은지 | 빈도의 빈도 (해시 2단) | 6장 Sherlock 행 |

---

## 2. 패턴 카드 (움직임 + 골격 + 불변식을 같이 — 한 몸)

### 2-A. BFS — 격자 최단 거리/횟수  (겪음: Castle)
문제: [HR Castle on the Grid](https://www.hackerrank.com/challenges/castle-on-the-grid/problem) · [PRG 게임 맵 최단거리](https://school.programmers.co.kr/learn/courses/30/lessons/1844) · [PRG 단어 변환](https://school.programmers.co.kr/learn/courses/30/lessons/43163) · [PRG 가장 먼 노드](https://school.programmers.co.kr/learn/courses/30/lessons/49189) · [LC 102 Binary Tree Level Order](https://leetcode.com/problems/binary-tree-level-order-traversal/)
움직임 — dist를 파동처럼 채운다. 목표에 처음 닿은 값이 최단:
```
0 1 2     시작 0에서 한 겹씩 +1로 퍼짐. 큐(FIFO)라 가까운 것부터.
1 2 3     경로를 모으지 않는다 — dist만 채우면 답.
2 3 4
```
골격:
```
int[][] dist = new int[n][m]; for(row) fill(row,-1);
queue.add(start); dist[sr][sc]=0;
while(!queue.isEmpty()){
  cur = queue.poll();
  if(cur==goal) return dist[cur.r][cur.c];       // 처음 꺼낸 게 최단
  for(int[] d : DIRS){                            // DIRS=상하좌우
    nr=cur.r+d[0]; nc=cur.c+d[1];
    if(범위 && !막힘 && dist[nr][nc]==-1){         // 넣기 전에 검사+표시
      dist[nr][nc]=dist[cur.r][cur.c]+1; queue.add({nr,nc});
    }
  }
}
```
함정: dist가 방문표 겸함(-1이면 미방문). 이동이 미끄러짐이면 2-C로.
**불변식**: 큐(FIFO) = 거리 오름차순 처리 → 각 칸에 *처음 닿는 순간* 값이 곧 최단(다시 안 바뀜). `==-1` 가드가 "처음 = 최단"을 지키는 자물쇠.

#### ▸ 왜 이 값이 맞나 — 도출 순서 (표-채우기 족보 공통: DP=2-E도 이 순서, 상태만 dp 인덱스로 바뀜)
"dist에 쌓는 값이 최적? 최소? 내가 원한 값?"의 불안은 여기서 닫힌다. 위 골격은 외운 게 아니라 이 순서를 격자·최단에 대입해 **창발한** 것.

**헷갈림의 뿌리 = 서로 다른 세 행위를 한 순간에 하려는 것.** 갈라놓으면 풀린다:

| 행위 | 무엇인가 | 지식 필요? |
|---|---|---|
| **값** (무엇을 담나) | *소원* — 방법 몰라도 적음. 문제가 대개 그냥 줌 | ❌ |
| **모양** (무엇으로 색인: 1D/2D/3D/맵) | *정답 있는 질문* — "한 상황을 콕 집는 좌표 수" | ⭕ 셈 |
| **맞음** (정말 그리 됐나) | *검증* — 불변식으로 맨 뒤에 획득 | ⭕ 확인 |

> 몰라서 못 쓰는 게 아니다. "원한다(소원)"를 먼저 적어야 방법을 찾을 방향이 생긴다. **확신은 맨 뒤(10)에 온다.** 주석은 맹세가 아니라 편집 가능한 가설.

**순서** (→ 는 그 단계가 뽑아내는 데이터/코드):
1. **출력 한 문장** — 결국 뭘 내놓나. → *"시작→목표 최소 걸음 수".*
2. **값의 소원** — 상황마다의 답으로 쪼갬. → `// dist[상황]=시작→그 상황 최소 걸음`.
3. **좌표 세기(=모양)** — "두 순간을 공정히 비교하려면 뭘 알아야?" 그 개수=차원. → 위치=(행,열)=2 → **2차원**.
4. **그릇 타입** — 좌표가 조밀한 정수면 배열, 아니면(문자열·희소) 맵. → `int[][] dist = new int[n][m]`.
5. **점화식(연산 강제)** — 한 상황 값 = 이전/이웃 값으로? `+1`/`min`/`max`는 2번 의미에서 역산. → 한 칸=1걸음 → `dist[이웃]=dist[현재]+1`.
6. **닫힘 검사 ⟳3** — 같은 슬롯에 답 다른 상황이 몰리나? 몰리면 좌표가 하나 빠진 것 → 3으로 가 차원↑ (벽 K개 부수기면 `dist[r][c][k]`). 실패가 빠진 좌표를 알려주니 미리 다 몰라도 됨.
7. **순서 장치(=자료구조)** — 점화식 입력이 읽는 시점에 *이미 최종*이려면 어떤 순서? 그 순서 만드는 게 구조. → 거리 오름차순 → **큐**. (가중치→힙 2-D, DAG→위상순서 2-E)
8. **초기화** — 방법 없이 아는 값 + "미확정" 표식. → `dist[start]=0; 나머지 -1; queue.add(start)`.
9. **채우기 루프(불변식을 코드로)** — 꺼내→점화식→*처음일 때만* 확정→등록. → `if(dist[이웃]==-1){ dist=...+1; add; }`.
10. **검증 ⟳** — 불변식 3박자(초기화 참/한 번 돌려도 참/끝나면 답) + 2×2 손 트레이스. 어긋나면 값 틀림→5·9, 슬롯 충돌→3·6, 순서 문제→7로.

**데이터 흐름 한 컷**: 빈 dist → `start=0` 씨앗 → 점화식이 이웃으로 `+1`씩 전파 → 큐가 "가까운 것 먼저" 보장 → 각 칸 *처음 닿을 때* 확정 → `dist[goal]`이 답.
**창발(사고→코드)**: 소원=`dist` 선언 / 좌표수=`[n][m]` / 초기화=`fill(-1)+start=0` / 점화식=`+1` 줄 / 순서보장=`Queue.poll·add` / 처음만=`==-1` 가드 / 종료=`return dist[goal]`. **골격의 모든 줄은 위 어느 단계의 산물** — 통째 외우지 말고 순서로 재생성.

### 2-B. flood fill — 이어진 덩어리 크기/개수  (겪음: Connected Cells, Number of Islands)
문제: [HR Connected Cells in a Grid](https://www.hackerrank.com/challenges/connected-cell-in-a-grid/problem) · [LC 200 Number of Islands](https://leetcode.com/problems/number-of-islands/) · [PRG 네트워크](https://school.programmers.co.kr/learn/courses/30/lessons/43162)
움직임 — 찾은 칸이 다시 출발점(연쇄). 꺾어가며 이어진 것을 다 방문:
```
1 1 0     (0,0)->이웃 (0,1),(1,1)->그 이웃 (1,2)... "이웃의 이웃의 이웃"
0 1 1     직선으로 긋는 게 아니라 연쇄로 퍼진다.
```
골격 (DFS 재귀 — 짧음, 기본으로 이걸):
```
int flood(r,c){
  if(범위밖 || visited[r][c] || 안채워짐) return 0;  // <- 가드의 집(재귀는 여기 한 곳)
  visited[r][c]=true;
  int size=1;
  for(int[] d : DIRS) size += flood(r+d[0], c+d[1]); // 이웃은 그냥 부름(미리 필터 불필요)
  return size;
}
// driver: for r,c: max = Math.max(max, flood(r,c));
```
골격 (BFS 큐 — 격자 아주 커서 깊은 재귀가 걱정될 때):
```
queue.add(seed); visited[seed]=true; int size=0;
while(!queue.isEmpty()){ cur=poll(); size++;
  for(d:DIRS){ nr,nc; if(범위 && !visited[nr][nc] && 채워짐){ visited[nr][nc]=true; queue.add({nr,nc}); } } }
```
함정(핵심): **가드 위치가 DFS/BFS 다르다** — DFS는 base case에서 거름(부르기 전 필터 불필요·무해), BFS는 넣기 전에 거름(필수, 안 하면 중복·무한). 한 곳에만. 0/벽은 visited 마크 불필요(조건이 이미 막음). 미끄러짐(2-C)과 헷갈리지 말 것 — flood엔 while 슬라이드 없음.
**불변식**: `visited[c]==true` = "이미 센 칸". 각 칸 정확히 1번만 세짐 → size 정확, 무한루프 없음. (최적화가 아니라 도달·개수 문제 — 2-A의 점화식/최소성 도출은 여기 안 씀.)
**바깥 중첩 for의 역할**: 씨앗 심기다 — 서로 안 이어진 덩어리가 여러 개라서 각 덩어리에 한 번씩 불을 붙이는 것. 연결을 따라가는 건 flood가 한다(for는 꺾어 들어가지 못한다 — (1,2)는 (0,1)이나 (1,1)에서 꺾어야 닿는다). 시작점이 하나로 주어지면(Castle) 바깥 for 불필요.

### 2-C. 미끄러짐(룩) — 격자 최소 이동  (겪음: Castle, 2-B와 구별용)
문제: [HR Castle on the Grid](https://www.hackerrank.com/challenges/castle-on-the-grid/problem)
움직임 — 한 방향으로 막힐 때까지 쭉, 그 직선 위 아무 칸에나 설 수 있음:
```
. X .     (0,0)에서 우는 X로 막힘, 아래로 쭉 -> (1,0),(2,0). 한 번에 여러 칸.
. X .     그래서 이웃 = 인접 한 칸이 아니라 '직선 위 모든 칸'.
. . .
```
골격 (2-A와 같은 BFS인데 방향마다 while 하나 더):
```
for(int[] d : DIRS){
  nr=cur.r+d[0]; nc=cur.c+d[1];
  while(범위 && !막힘){                     // <- 미끄러짐 = while
    if(dist[nr][nc]==-1){ dist[nr][nc]=dist[cur.r][cur.c]+1; queue.add({nr,nc}); }
    nr+=d[0]; nc+=d[1];                       // 같은 방향으로 계속
  }
}
```
함정: 2-B(flood)엔 이 while이 없다. 직전 문제(Castle)의 while을 flood에 옮기다 꺾인 연결을 놓친 게 실제 실수.
**불변식**: 2-A와 동일(큐=거리 오름차순, 처음 닿은 값=최단). 달라진 건 "이웃"의 정의뿐 — 인접 한 칸이 아니라 직선 위 모든 칸. 도출 순서도 2-A와 같고 5번(이웃 규칙)만 바뀜.

### 2-D. 힙 — 최솟값을 반복해 꺼내 처리  (겪음: Jesse and Cookies)
문제: [HR Jesse and Cookies](https://www.hackerrank.com/challenges/jesse-and-cookies/problem) · [PRG 더 맵게](https://school.programmers.co.kr/learn/courses/30/lessons/42626)(사실상 같은 문제) · [PRG 디스크 컨트롤러](https://school.programmers.co.kr/learn/courses/30/lessons/42627) · [PRG 이중우선순위큐](https://school.programmers.co.kr/learn/courses/30/lessons/42628)
움직임 — 정렬 전체를 안 하고 매번 극값만:
```
넣기 5,1,3  ->  꺼내면 1,3,5 (항상 최솟값이 꼭대기). 넣고 꺼낼 때마다 O(log n).
```
골격 (목표의 부정을 루프 조건으로):
```
PriorityQueue<Long> pq = new PriorityQueue<>();   // long! 결합값이 int(21억) 넘음
for(int x : a) pq.add((long)x);
int ops=0;
while(pq.peek() < k){            // 목표 "최소>=k"의 부정을 루프 조건에 (불변식)
  if(pq.size()<2) return -1;     // 섞을 둘 없으면 불가
  long x=pq.poll(), y=pq.poll();
  pq.offer(x + 2*y); ops++;
}
return ops;                       // 루프 0회 = 이미 만족 = 0
```
함정: max-heap은 `new PriorityQueue<>(Comparator.reverseOrder())`. 합/곱이 21억 넘으면 long.
**불변식**: 힙 꼭대기 = 현재 최소. 탐욕이 맞는 근거는 "가장 작은 둘을 합치는 게 항상 최선"(교환논법). 루프 불변: `peek()<k`인 동안만 도니 끝나면 전부 ≥k. (2-A식 상태-표가 아님 — 도출은 좌표세기가 아니라 탐욕 정당화.)

### 2-E. DP/메모 — 겹치는 부분문제  (겪음: Davis' Staircase)
문제: [HR Davis' Staircase](https://www.hackerrank.com/challenges/ctci-recursive-staircase/problem) · [PRG 멀리 뛰기](https://school.programmers.co.kr/learn/courses/30/lessons/12914)(같은 점화식) · [PRG 정수 삼각형](https://school.programmers.co.kr/learn/courses/30/lessons/43105) · [PRG 등굣길](https://school.programmers.co.kr/learn/courses/30/lessons/42898) · [PRG 땅따먹기](https://school.programmers.co.kr/learn/courses/30/lessons/12913) · [PRG N으로 표현](https://school.programmers.co.kr/learn/courses/30/lessons/42895)
움직임 — 나열하면 지수, 상태별 1회 저장하면 선형:
```
f(n)=f(n-1)+f(n-2)+f(n-3). 여러 경로가 f(3)을 또 계산 -> 지수 폭발.
한 번 계산해 dp[3]에 저장, 다음엔 꺼내 쓴다.
```
골격 (bottom-up 표 — 제일 깔끔):
```
int[] dp = new int[n+1];
dp[0]=1;
for(int i=1;i<=n;i++){
  dp[i]=dp[i-1];
  if(i>=2) dp[i]+=dp[i-2];
  if(i>=3) dp[i]+=dp[i-3];
}
return dp[n];
```
골격 (top-down 메모 — 재귀가 자연스러울 때):
```
Integer[] memo = new Integer[n+1];
int f(int i){ if(i==0)return 1; if(i<0)return 0;
  if(memo[i]!=null) return memo[i];
  return memo[i] = f(i-1)+f(i-2)+f(i-3); }
```
함정 트리거: "같은 인자로 여러 번 불림 + 답이 경로가 아니라 인자에만 의존" -> 캐시하라.
**불변식**: `dp[i]` = i까지의 최적/개수. 순회 순서(작은 i 먼저)가 dp[i]가 참조하는 것들(i-1,i-2,i-3)이 *이미 확정*임을 보장 — 2-A의 "큐"가 하던 순서 보장을 여기선 for 순서가 한다. **도출은 2-A "왜 맞나"와 같은 족보**: 상태=dp 인덱스, 순서 장치=순회 순서(또는 위상정렬).

### 2-F. 이분탐색 — 정렬된 것에서 반씩 버림
문제: [PRG 입국심사](https://school.programmers.co.kr/learn/courses/30/lessons/43238)(답을 이분탐색 — "시간 T에 다 처리 가능?"을 판정으로)
움직임:
```
[1,3,5,7,9]서 7: 가운데 5<7 -> 오른쪽만 -> 가운데 7 = 찾음. 매번 절반 O(log n).
```
골격:
```
int lo=0, hi=n-1;
while(lo<=hi){
  int mid=lo+(hi-lo)/2;              // (lo+hi)/2는 오버플로 위험
  if(a[mid]==target) return mid;
  else if(a[mid]<target) lo=mid+1;
  else hi=mid-1;
}
return -1;
```
**불변식**: "답이 있다면 반드시 `[lo,hi]` 안에 있다". 매 반복 절반을 버려도 이게 유지되므로 안전. 종료(lo>hi)=구간 비었으니 없음.

### 2-G. 슬라이딩 윈도우 — 연속 구간, 두 포인터
문제: [LC 3 Longest Substring Without Repeating](https://leetcode.com/problems/longest-substring-without-repeating-characters/) · [HR Fraudulent Activity Notifications](https://www.hackerrank.com/challenges/fraudulent-activity-notifications/problem)(창 + 계수 배열)
움직임:
```
[2,1,5,1,3] 길이3 최대합: right로 더하고, 창이 3 넘으면 left 것을 뺀다. O(n).
```
골격:
```
int sum=0, best=Integer.MIN_VALUE;
for(int r=0;r<n;r++){
  sum+=a[r];
  if(r>=k) sum-=a[r-k];             // 창 벗어난 왼쪽 빼기
  if(r>=k-1) best=Math.max(best,sum);
}
```
**불변식**: `sum` = 지금 창(최근 k개)의 합. r 진행 때 더하고 벗어난 왼쪽을 빼는 짝이 이 한 문장을 유지 → 매 위치에서 O(1)로 정답.

### 2-H. 스택 — 최근 것부터 / 짝 맞추기
문제: [HR Balanced Brackets](https://www.hackerrank.com/challenges/balanced-brackets/problem) · [PRG 올바른 괄호](https://school.programmers.co.kr/learn/courses/30/lessons/12909) · [LC 739 Daily Temperatures](https://leetcode.com/problems/daily-temperatures/) · [PRG 뒤에 있는 큰 수 찾기](https://school.programmers.co.kr/learn/courses/30/lessons/154539) · [PRG 주식가격](https://school.programmers.co.kr/learn/courses/30/lessons/42584)
움직임:
```
"([)": '(' push, '[' push, ')' pop해 짝 확인 -> 짝은 '['이라 불일치. 끝에 비면 정상.
```
골격:
```
Deque<Character> st = new ArrayDeque<>();
for(char c : s){
  if(여는 괄호) st.push(c);
  else { if(st.isEmpty() || !맞짝(st.pop(), c)) return false; }
}
return st.isEmpty();
```
**불변식**: 스택 = "아직 안 닫힌 여는 괄호들"(가장 최근이 맨 위). 닫는 걸 만나면 반드시 맨 위와 짝이어야 함 → 끝에 비면 전부 짝 맞음.

### 2-I. 백트래킹 — 모든 조합/순열 생성
문제: [PRG 타겟 넘버](https://school.programmers.co.kr/learn/courses/30/lessons/43165) · [PRG 소수 찾기](https://school.programmers.co.kr/learn/courses/30/lessons/42839)(순열 생성) · [PRG 모의고사](https://school.programmers.co.kr/learn/courses/30/lessons/42840)(완전탐색)
움직임:
```
[1,2,3] 조합: 1 고름 -> {1,2},{1,3} -> 1 빼고(되돌리기) 2 고름 -> {2,3}...
```
골격:
```
void rec(int start, List<Integer> path){
  record(path);                        // 또는 path.size()==k일 때만
  for(int i=start;i<n;i++){
    path.add(a[i]);
    rec(i+1, path);                    // 조합: i+1(지난 건 안 봄). 순열이면 for 0..n + used[]
    path.remove(path.size()-1);        // 되돌리기 = 백트래킹의 심장
  }
}
```
함정: 순서 무관(조합)이면 start로 앞으로만, 순서 중요(순열)면 used[]로 전체에서 쓴 것만 제외.
**불변식**: `path` = 지금까지 고른 부분해. `add → 재귀 → remove` 짝이 "형제 가지로 넘어가기 전 상태 원복"을 보장 → 모든 조합을 정확히 1번씩, 오염 없이.

---

## 3. 격자 3형제 한눈 구별 (2-A / 2-B / 2-C가 헷갈릴 때)

| 유형 | 이웃 | 무엇을 | while? |
|---|---|---|---|
| flood(2-B) | 인접 칸, 찾은 칸 재출발 | 이어진 칸 수 | 없음 |
| 최단 BFS(2-A) | 인접 칸, dist 파동 | 최소 거리/횟수 | 없음 |
| 미끄러짐(2-C) | 직선 위 모든 칸 | 최소 이동 횟수 | **있음** |

**BFS vs DFS**: 최단이면 반드시 BFS. 그냥 다 방문/개수면 아무거나(DFS 재귀가 짧음).
**씨앗 심는 세 가지**: 시작점 하나 주어짐(Castle) = 바깥 for 불필요 / 덩어리 여러 개(Connected Cells) = 중첩 for로 씨앗 / 시작점 여러 개 동시(썩는 오렌지류) = 전부 큐에 넣고 시작.

## 4. 복잡도 4박자 (약점 — 매번 이 순서로 소리 내어)
1. 변수 정의(n, m이 뭔지) 2. 바깥 루프 횟수 x 반복당 비용 3. 곱(중첩)/합(순차), 지배항 4. 제약 대입 어림수(자바 ~10^8/초, ~은 "약"). 격자면 대개 O(n*m).

## 5. 타입/오버플로 (Jesse에서 데임)
값·합·곱의 최대가 int(약 21억, 2,147,483,647) 넘으면 long. mod 10^10 규모면 long.

## 6. 나의 반복 함정 — 실제로 겪은 것만 (이 표가 이 문서의 심장)

무는 순서대로: 읽기 -> 접근 -> 구현.

| 함정 | 실제 사건 | 실전 처방 |
|---|---|---|
| **문제 오독** — 이동/연결 규칙 넘겨짚기 | Castle: 미끄러짐을 "한 칸 이동"으로 오해한 채 구현 | 코딩 전 샘플 1개를 문제 규칙대로 손으로 굴려 기대 출력 재현 (대본 0~3분 칸에 박아둠) |
| **직전 모델 전이** | Castle의 while 슬라이드를 Connected Cells(flood)에 옮겨 꺾인 연결을 놓침 | 새 문제 = 이동/연결 규칙을 백지에서 새로 읽기. 격자면 3장 표로 3형제 중 무엇인지 먼저 판정 |
| **루프 역할 혼동** | Connected Cells: "중첩 for로 다 방문하면 되지 않나"로 샜음 | 중첩 for = 씨앗 심기, 연결 추적 = flood. for는 꺾어 들어가지 못한다 |
| **스친 정답 승격 실패** | Davis: 주석에 "caching?"까지 써놓고 지나감. "차라리 정렬?"도 같은 패턴 | 번뜩이면 거기서 멈춰 소리 내기: "방금 생각이 답 아닌가?" 비용을 따진 뒤에 버린다 |
| **최적화 구조가 안 떠오름** | Davis: 나열(재귀)까지 하고 캐시로 못 넘어감 | 신호 고정: "같은 인자로 또 불림 + 답이 인자에만 의존" -> 상태 저장(2-E). 나열이 먼저, 상태는 그 다음 |
| **추상적으로만 생각하다 시간 소모** | Castle: dist[][] 구조가 안 떠오름 — 데이터 흐름이 머릿속에 없었음 | 막히면 즉시 2x2~3x3 예제로 자료구조 상태를 손으로 2단계 그리기. 값으로 생각하고, 구조가 필요하면 2-A 도출 순서로 |
| **케이스 열거 불완전** | Sherlock: B(최대 빈도 -1)는 떠올렸는데 A(빈도 1 글자 통째 삭제)를 놓칠 뻔 | 케이스 나눴으면 1회 강제 되묻기: "이 조작이 가능한 *다른* 형태는 없나?" 성격이 다른 케이스(수 조정 vs 통째 삭제)가 숨어 있다 |
| **사소한 분기 고민에 생각 틀어짐** | Connected Cells: "0도 visited 마크? driver에 continue?"에서 흔들림 | 정확성에 영향 없으면 아무거나 골라 선언하고 진행: "어느 쪽이든 되니 A로 갑니다" |
| **오버플로** | Jesse: 결합값 x+2y가 int 초과해 WA | 합/곱이 보이면 최댓값 어림 먼저(5장). 21억 근처면 그냥 long |
| **복잡도 나중에 세기** | 여러 번 — 짜고 나서 통과 여부 걱정 | 코딩 전 4박자(4장)를 소리 내는 것으로 고정 — 대본 3~8분 칸이 강제한다 |

상시 리마인더:
- **최단 = dist 채우기**, 경로 수집 아님.
- **값 안 맞으면 그 카드 불변식부터**, 표-채우기면 2-A "왜 맞나"로. "뭘 증가시키지"가 막히면 소원·좌표·점화식이 안 선 것 — 감으로 +1 넣지 말 것.

## 7. 막혔을 때 사다리 + 얼음 응급 처치

**막힘 사다리 (위에서부터 차례로 — 절대 침묵 금지):**
1. **막힌 지점을 말로** — "지금 X 때문에 막혔는데 Y랑 Z 고민 중이에요." (말하다 풀리는 경우 많음)
2. **작은 예제를 손으로** 굴린다 (n=2, 3).
3. **브루트포스라도 짜기 시작** — 느려도 도는 코드 >> 우아한 빈 화면.
4. **힌트 요청** — "이 방향 맞을까요?" / "힌트 하나 주실 수 있나요?" (감점 아님, 소통 점수)

**몸이 굳을 때 (긴장 완화 — 몸 먼저, 생각은 그 다음):**
- 호흡 1세트: 4초 들이쉬고 6초 내쉬기 x 3회. 날숨이 들숨보다 길면 심박이 내려간다. 시작 직전에도 1세트.
- 어깨 내리고 물 한 모금 — 이 동작 자체를 "리셋 버튼"으로 정해 둔다.
- 화면에서 눈 떼고 문제 텍스트로 돌아간다 — "문제를 다시 한번 읽어볼게요"는 자연스러운 멘트이자 합법적 리셋.

**리프레임 3개:**
- **면접관 = 페어.** 끝까지 풀길 바라는 같은 편. (실제 후기: 못 떠올리자 같이 헷갈려주고 다른 길로 유도해 풀게 함)
- **느린 정답 >> 빈 화면.** 브루트포스는 언제나 내놓을 수 있다 — 최악은 비어 있는 1시간이지, 느린 정답이 아니다.
- **30초 이상 침묵 금지.** 생각도 소리 내어. 침묵이 유일한 진짜 실패다.

## 8. 팩트 박스 + 당일 체크리스트

**전형 팩트 (후기 종합):**
- 1시간, 문제 2개, 1:1 화면공유(HackerRank CodePair). 난이도 LeetCode Medium 중심.
- 관찰된 유형: **BFS · Stack · Interval · String** (+ 구현/정렬/해시).
- 평가: 개발 능력만이 아니라 **문제 해결 과정 + 커뮤니케이션 + 프로그래밍 습관** 전체.

**이미 손에 있는 유형 (자신감 근거 — 전부 직접 푼/만진 문제, 관찰된 유형 전부 커버):**

| 유형 | 카드 | 푼/만진 문제 |
|---|---|---|
| 스택 | 2-H | [HR Balanced Brackets](https://www.hackerrank.com/challenges/balanced-brackets/problem) · [LC 739 Daily Temperatures](https://leetcode.com/problems/daily-temperatures/) |
| flood | 2-B | [HR Connected Cells](https://www.hackerrank.com/challenges/connected-cell-in-a-grid/problem) · [LC 200 Number of Islands](https://leetcode.com/problems/number-of-islands/) |
| BFS | 2-A/2-C | [HR Castle on the Grid](https://www.hackerrank.com/challenges/castle-on-the-grid/problem) |
| 슬라이딩 윈도우 | 2-G | [LC 3 Longest Substring](https://leetcode.com/problems/longest-substring-without-repeating-characters/) · [HR Fraudulent Activity](https://www.hackerrank.com/challenges/fraudulent-activity-notifications/problem) |
| 정렬 | — | [HR Quicksort 2](https://www.hackerrank.com/challenges/quicksort2/problem) · [PRG K번째수](https://school.programmers.co.kr/learn/courses/30/lessons/42748) · [PRG H-Index](https://school.programmers.co.kr/learn/courses/30/lessons/42747) |
| 힙 | 2-D | [HR Jesse and Cookies](https://www.hackerrank.com/challenges/jesse-and-cookies/problem) |
| DP | 2-E | [HR Davis' Staircase](https://www.hackerrank.com/challenges/ctci-recursive-staircase/problem) |
| Interval | — | [LC 56 Merge Intervals](https://leetcode.com/problems/merge-intervals/) · [LC 57 Insert Interval](https://leetcode.com/problems/insert-interval/) · [PRG 단속카메라](https://school.programmers.co.kr/learn/courses/30/lessons/42884) |
| 해시 | — | [HR Sherlock Valid String](https://www.hackerrank.com/challenges/sherlock-and-valid-string/problem)(빈도의 빈도) · [LC 49 Group Anagrams](https://leetcode.com/problems/group-anagrams/) · [PRG 의상](https://school.programmers.co.kr/learn/courses/30/lessons/42578) · [PRG 전화번호 목록](https://school.programmers.co.kr/learn/courses/30/lessons/42577) · [PRG 완주하지 못한 선수](https://school.programmers.co.kr/learn/courses/30/lessons/42576) · [PRG 베스트앨범](https://school.programmers.co.kr/learn/courses/30/lessons/42579) |
| 큐/시뮬레이션 | — | [PRG 기능개발](https://school.programmers.co.kr/learn/courses/30/lessons/42586) · [PRG 프로세스](https://school.programmers.co.kr/learn/courses/30/lessons/42587) · [PRG 다리를 지나는 트럭](https://school.programmers.co.kr/learn/courses/30/lessons/42583) |
| 연결 리스트 | — | [LC 206 Reverse Linked List](https://leetcode.com/problems/reverse-linked-list/) · [LC 21 Merge Two Sorted Lists](https://leetcode.com/problems/merge-two-sorted-lists/) |
| 그리디 | — | [PRG 체육복](https://school.programmers.co.kr/learn/courses/30/lessons/42862) · [PRG 큰 수 만들기](https://school.programmers.co.kr/learn/courses/30/lessons/42883)(스택 결합) · [PRG 구명보트](https://school.programmers.co.kr/learn/courses/30/lessons/42885)(투 포인터) |
| 트리 재귀 | — | [LC 104 Maximum Depth](https://leetcode.com/problems/maximum-depth-of-binary-tree/) · [LC 543 Diameter of Binary Tree](https://leetcode.com/problems/diameter-of-binary-tree/) |

**전날 밤:**
- [ ] HackerRank 에디터(CodePair 유사 환경) 5분 만져보기 — 실행 버튼, 단축키. 도구 낯설면 그것만으로 버벅인다.
- [ ] 이 문서 통독 1회 + 치트시트(`cheatsheet/Cheatsheet.java`) 훑기 1회.
- [ ] **새 문제 금지. 일찍 자기.** (지식 한 조각 < 수면이 지키는 워킹메모리)

**당일 아침:**
- [ ] 20분 전 자리 — 카메라/마이크/링크 점검, 물 준비.
- [ ] 0장 대본 + 6장 함정 표만 다시 훑기 (새 내용 금지).
- [ ] 호흡 1세트(4-6 x 3회), 시작 멘트 입에 올리기: **"먼저 입력/출력과 제약을 확인하겠습니다."**

---
**실전 한 줄 요약**: 문제 -> 규칙을 손으로 재현(오독 차단) -> 1번 표로 도구 -> 그 카드의 움직임+골격+불변식 -> 복잡도 4박자 소리 내기 -> 합의 받고 골격 받아쓰기 -> 샘플+엣지 트레이스. 흔들리면 그 카드로, 쌓는 값이 흔들리면 2-A "왜 맞나"로, 몸이 굳으면 7장으로.
