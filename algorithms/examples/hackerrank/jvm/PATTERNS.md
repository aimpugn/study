# 패턴 결정 가이드 — 접근이 안 잡히거나 도중에 흔들릴 때

> 문법은 `cheatsheet`, 멘탈·막힘 대응은 `PLAYBOOK`. 여기는 "무엇으로, 어떻게".
> 1번 표로 도구를 고르고, 2번에서 **그 패턴의 움직임 + 골격을 한 카드로** 본다(둘은 붙어 있다).

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

---

## 2. 패턴 카드 (움직임 + 골격을 같이 — 이 둘은 한 몸)

### 2-A. BFS — 격자 최단 거리/횟수  (겪음: Castle)
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

### 2-B. flood fill — 이어진 덩어리 크기/개수  (겪음: Connected Cells, Number of Islands)
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

### 2-C. 미끄러짐(룩) — 격자 최소 이동  (겪음: Castle, 2-B와 구별용)
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

### 2-D. 힙 — 최솟값을 반복해 꺼내 처리  (겪음: Jesse and Cookies)
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

### 2-E. DP/메모 — 겹치는 부분문제  (겪음: Davis' Staircase)
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

### 2-F. 이분탐색 — 정렬된 것에서 반씩 버림
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

### 2-G. 슬라이딩 윈도우 — 연속 구간, 두 포인터
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

### 2-H. 스택 — 최근 것부터 / 짝 맞추기
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

### 2-I. 백트래킹 — 모든 조합/순열 생성
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

---

## 3. 격자 3형제 한눈 구별 (2-A / 2-B / 2-C가 헷갈릴 때)

| 유형 | 이웃 | 무엇을 | while? |
|---|---|---|---|
| flood(2-B) | 인접 칸, 찾은 칸 재출발 | 이어진 칸 수 | 없음 |
| 최단 BFS(2-A) | 인접 칸, dist 파동 | 최소 거리/횟수 | 없음 |
| 미끄러짐(2-C) | 직선 위 모든 칸 | 최소 이동 횟수 | **있음** |

**BFS vs DFS**: 최단이면 반드시 BFS. 그냥 다 방문/개수면 아무거나(DFS 재귀가 짧음).

## 4. 복잡도 4박자 (약점 — 매번 이 순서로 소리 내어)
1. 변수 정의(n, m이 뭔지) 2. 바깥 루프 횟수 x 반복당 비용 3. 곱(중첩)/합(순차), 지배항 4. 제약 대입 어림수(자바 ~10^8/초). 격자면 대개 O(n*m).

## 5. 타입/오버플로 (Jesse에서 데임)
값·합·곱의 최대가 int(약 21억, 2,147,483,647) 넘으면 long. mod 10^10 규모면 long.

## 6. 나의 반복 함정 (이것만 조심하면 절반은 산다)
- **스친 정답 승격**: 번뜩인 생각("차라리 정렬?", "caching?")에서 멈춰 "이게 답 아닌가?" 따져라.
- **직전 모델 전이 금지**: 방금 푼 방식을 옮기지 말고 이 문제의 이동/연결 규칙을 새로 읽어라(2-C를 2-B에 옮긴 실수).
- **최단 = dist 채우기**, 경로 수집 아님.
- **복잡도 먼저**: 브루트포스라도 짜기 전에 복잡도부터 세서 통과 여부 판단.

---
**실전 순서**: 문제 -> 도구 고르기(1) -> 그 카드에서 움직임+골격 같이 보기(2) -> 복잡도(4) -> 골격 베끼기 -> 입으로 말하며 구현. 흔들리면 그 카드로 돌아온다.
