package cheatsheet;

import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Comparator;
import java.util.Deque;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.PriorityQueue;
import java.util.Set;

/**
 * 코딩테스트 문법 치트시트 (러너블 — `java src/main/java/cheatsheet/Cheatsheet.java`로 돌려 출력 확인).
 * <p>
 * 시험장에서 자주 막히는 자바 문법만 모아, 각 줄이 실제로 컴파일·실행되게 둡니다. 외우는 게 아니라
 * "이 자리에서 뭐였더라" 할 때 펴 보고, 한 번씩 돌려 손에 익히는 용도입니다. 섹션별 메서드로 나눠
 * 두었으니 필요한 곳만 읽으세요.
 */
public class Cheatsheet {

    public static void main(String[] args) {
        priorityQueue();
        stackAndQueue();
        arrays2DandFill();
        listAndStream();
        mapAndSet();
        stringStuff();
        comparatorAndSort();
        overflowAndType();
        patternSkeletons();
    }

    // ─────────────────────────────────────────────────────────────
    // 1) PriorityQueue (힙) — 정렬 방향이 헷갈리는 그 부분
    //    기본은 min-heap: 가장 "작은" 게 먼저 poll. max는 Comparator를 뒤집는다.
    static void priorityQueue() {
        System.out.println("== PriorityQueue ==");

        // min-heap (기본): poll 하면 작은 값부터
        PriorityQueue<Integer> min = new PriorityQueue<>();
        min.offer(5); min.offer(1); min.offer(3);
        System.out.println("min peek=" + min.peek());      // 1 (가장 작은 게 꼭대기)
        System.out.println("min poll=" + min.poll());       // 1

        // max-heap: 비교자를 뒤집는다. 둘 다 같은 뜻.
        PriorityQueue<Integer> max1 = new PriorityQueue<>(Comparator.reverseOrder());
        PriorityQueue<Integer> max2 = new PriorityQueue<>((a, b) -> b - a); // 작은 음수면 a가 앞 -> b-a는 큰 게 앞
        max1.offer(5); max1.offer(1); max1.offer(3);
        System.out.println("max peek=" + max1.peek());      // 5

        // 객체를 특정 필드로 정렬 (예: int[] 를 [0] 기준 min)
        PriorityQueue<int[]> byFirst = new PriorityQueue<>(Comparator.comparingInt(a -> a[0]));
        byFirst.offer(new int[]{3, 9}); byFirst.offer(new int[]{1, 8});
        System.out.println("byField peek=" + Arrays.toString(byFirst.peek())); // [1, 8]

        // 외우는 트리거: "가장 작은/큰 것을 반복해서 꺼낸다" -> 힙. min이 기본, max는 reverseOrder().
    }

    // ─────────────────────────────────────────────────────────────
    // 2) ArrayDeque — 스택도 큐도 이걸로. (java.util.Stack은 느리고 구식이라 안 씀)
    static void stackAndQueue() {
        System.out.println("== ArrayDeque (Stack / Queue) ==");

        // 스택(LIFO): push(머리에 넣기) / pop(머리에서 빼기) / peek(머리 보기)
        Deque<Integer> stack = new ArrayDeque<>();
        stack.push(1); stack.push(2); stack.push(3);
        System.out.println("stack pop=" + stack.pop());     // 3 (마지막에 넣은 게 먼저)
        System.out.println("stack peek=" + stack.peek());   // 2

        // 큐(FIFO): offer(꼬리에 넣기) / poll(머리에서 빼기) / peek(머리 보기)
        Deque<Integer> queue = new ArrayDeque<>();
        queue.offer(1); queue.offer(2); queue.offer(3);
        System.out.println("queue poll=" + queue.poll());   // 1 (먼저 넣은 게 먼저)
        System.out.println("queue peek=" + queue.peek());   // 2

        // 비었는지: isEmpty(). 크기: size(). 빈 데서 pop/poll 하면 예외(stack)/null(queue)니 항상 검사.
        System.out.println("empty? " + new ArrayDeque<>().isEmpty()); // true
    }

    // ─────────────────────────────────────────────────────────────
    // 3) 배열 — 2D 생성, Arrays.fill, 정렬, 복사, 합계
    static void arrays2DandFill() {
        System.out.println("== Arrays / 2D / fill ==");

        int[] a = new int[5];
        Arrays.fill(a, -1);                                 // 전부 -1 (메모 초기화에 자주)
        System.out.println("fill=" + Arrays.toString(a));   // [-1, -1, -1, -1, -1]

        int[][] grid = new int[3][4];                        // 3행 4열, 0으로 채워짐
        System.out.println("rows=" + grid.length + " cols=" + grid[0].length); // 3, 4

        int[] b = {5, 2, 8, 1};
        Arrays.sort(b);                                      // 오름차순 정렬 (제자리)
        System.out.println("sort=" + Arrays.toString(b));   // [1, 2, 5, 8]

        int[] c = Arrays.copyOf(b, 6);                       // 길이 6으로 복사 (뒤는 0)
        System.out.println("copyOf=" + Arrays.toString(c));  // [1, 2, 5, 8, 0, 0]

        int sum = Arrays.stream(b).sum();                    // 합계 (max(), min()도 있음)
        System.out.println("sum=" + sum);                    // 16
    }

    // ─────────────────────────────────────────────────────────────
    // 4) List + Stream — 생성, 정렬, stream 합/변환
    static void listAndStream() {
        System.out.println("== List / Stream ==");

        List<Integer> list = new ArrayList<>(List.of(5, 2, 8, 1)); // List.of는 불변이라 감싸서 가변으로
        list.add(9);
        list.sort((x, y) -> x - y);                          // 또는 Collections.sort(list)
        System.out.println("sorted=" + list);                // [1, 2, 5, 8, 9]
        System.out.println("first=" + list.getFirst());      // 1 (JDK 21+: get(0) 대신)

        int sum = list.stream().mapToInt(Integer::intValue).sum(); // List<Integer> 합계
        System.out.println("stream sum=" + sum);             // 25

        List<Integer> evens = list.stream().filter(x -> x % 2 == 0).toList(); // 거르기
        System.out.println("evens=" + evens);                // [2, 8]

        // 주의: 핫 루프(많이 도는 곳)에선 스트림보다 for문이 빠르고 디버깅 쉽다. 깔끔한 변환에만.
    }

    // ─────────────────────────────────────────────────────────────
    // 5) Map / Set — 빈도 세기, 그룹핑, 멤버십
    static void mapAndSet() {
        System.out.println("== Map / Set ==");

        Map<String, Integer> freq = new HashMap<>();
        for (String w : new String[]{"a", "b", "a"}) {
            freq.put(w, freq.getOrDefault(w, 0) + 1);        // 빈도 세기 정석
            // 또는: freq.merge(w, 1, Integer::sum);
        }
        System.out.println("freq=" + freq);                  // {a=2, b=1}

        // 그룹핑: 키 없으면 빈 리스트 만들어 add
        Map<Integer, List<String>> byLen = new HashMap<>();
        for (String w : new String[]{"hi", "go", "yes"}) {
            byLen.computeIfAbsent(w.length(), k -> new ArrayList<>()).add(w);
        }
        System.out.println("group=" + byLen);                // {2=[hi, go], 3=[yes]}

        Set<Integer> seen = new HashSet<>();
        System.out.println("add new? " + seen.add(7));       // true (처음)
        System.out.println("add again? " + seen.add(7));     // false (이미 있음) -> 중복 검출에 유용
    }

    // ─────────────────────────────────────────────────────────────
    // 6) 문자열 — char 다루기, StringBuilder, 변환
    static void stringStuff() {
        System.out.println("== String ==");

        String s = "abc";
        System.out.println("charAt=" + s.charAt(0));         // 'a'
        System.out.println("sub=" + s.substring(1, 3));      // "bc" (1 이상 3 미만)

        char[] chars = s.toCharArray();
        Arrays.sort(chars);                                  // 문자 정렬 (애너그램 키)
        System.out.println("fromChars=" + new String(chars)); // "abc"

        // char <-> 숫자: 알파벳 인덱스
        System.out.println("c-'a'=" + ('c' - 'a'));          // 2
        System.out.println("'a'+2=" + (char) ('a' + 2));     // 'c'

        StringBuilder sb = new StringBuilder();
        sb.append("ab").append(1);
        sb.reverse();                                        // 뒤집기 (제자리)
        System.out.println("sb=" + sb);                      // "1ba"

        System.out.println("split=" + Arrays.toString("a,b,c".split(","))); // [a, b, c]
        System.out.println("join=" + String.join("-", "x", "y"));            // "x-y"
    }

    // ─────────────────────────────────────────────────────────────
    // 7) Comparator / 정렬 — 2D 정렬, 다중 키, 오버플로 함정
    static void comparatorAndSort() {
        System.out.println("== Comparator / sort ==");

        // 2D 배열을 [0] 기준 오름차순 (Interval 문제 등) — 오버플로 피하려 빼기 대신 compare
        int[][] iv = {{3, 4}, {1, 9}, {1, 2}};
        Arrays.sort(iv, (x, y) -> Integer.compare(x[0], y[0]));     // a[0]-b[0]은 큰 음수에서 오버플로 위험
        System.out.println("by[0]=" + Arrays.deepToString(iv));    // [[1, 9], [1, 2], [3, 4]]

        // 다중 키: [0] 같으면 [1]로 (comparingInt + thenComparingInt)
        Arrays.sort(iv, Comparator.<int[]>comparingInt(x -> x[0]).thenComparingInt(x -> x[1]));
        System.out.println("by[0]then[1]=" + Arrays.deepToString(iv)); // [[1, 2], [1, 9], [3, 4]]

        // 내림차순: reversed()
        List<Integer> nums = new ArrayList<>(List.of(1, 3, 2));
        nums.sort(Comparator.reverseOrder());
        System.out.println("desc=" + nums);                  // [3, 2, 1]
    }

    // ─────────────────────────────────────────────────────────────
    // 8) 오버플로 / 타입 — int 범위와 long (Jesse and Cookies에서 데인 그것)
    static void overflowAndType() {
        System.out.println("== overflow / type ==");

        System.out.println("int max=" + Integer.MAX_VALUE);  // 2,147,483,647 (약 21억)
        System.out.println("long max=" + Long.MAX_VALUE);    // 9,223,372,036,854,775,807 (약 922경)

        // 합/곱이 21억을 넘을 수 있으면 long으로. 리터럴에 L 붙이기.
        long big = 2_000_000_000L + 2_000_000_000L;          // L 없으면 int로 더해 오버플로
        System.out.println("big=" + big);                    // 4,000,000,000

        // 트리거: "제약을 읽고 타입을 정한다" — 값/합/곱의 최대가 21억 넘으면 long.
    }

    // ─────────────────────────────────────────────────────────────
    // 9) 패턴 골격 — 막히면 여기서 베껴 시작 (BFS, DFS, DP, 이분탐색, 투포인터, 슬라이딩윈도우)
    static void patternSkeletons() {
        System.out.println("== pattern skeletons ==");

        // BFS (격자 최단거리/단계 전파): 큐 + 방문표 + 4방향
        int[][] grid = {{0, 0, 0}, {0, 1, 0}, {0, 0, 0}};
        int rows = grid.length, cols = grid[0].length;
        boolean[][] visited = new boolean[rows][cols];
        Deque<int[]> q = new ArrayDeque<>();
        q.offer(new int[]{0, 0});
        visited[0][0] = true;
        int[][] dirs = {{-1, 0}, {1, 0}, {0, -1}, {0, 1}};   // 상하좌우
        int reached = 0;
        while (!q.isEmpty()) {
            int[] cur = q.poll();
            reached++;
            for (int[] d : dirs) {
                int nr = cur[0] + d[0], nc = cur[1] + d[1];
                if (nr < 0 || nr >= rows || nc < 0 || nc >= cols) continue; // 경계 밖
                if (visited[nr][nc] || grid[nr][nc] == 1) continue;          // 방문/벽
                visited[nr][nc] = true;
                q.offer(new int[]{nr, nc});
            }
        }
        System.out.println("BFS reached=" + reached);        // 8 (벽 1칸 빼고 다)

        // 이분탐색 (정렬된 배열에서 target 위치): 닫힌 구간 [lo, hi]
        int[] sorted = {1, 3, 5, 7, 9};
        int target = 7, lo = 0, hi = sorted.length - 1, found = -1;
        while (lo <= hi) {
            int mid = lo + (hi - lo) / 2;                     // (lo+hi)/2는 오버플로 위험
            if (sorted[mid] == target) { found = mid; break; }
            else if (sorted[mid] < target) lo = mid + 1;
            else hi = mid - 1;
        }
        System.out.println("binarySearch idx=" + found);     // 3

        // 투포인터 (정렬 배열에서 두 수의 합 = target)
        int[] arr = {1, 2, 4, 7, 11};
        int want = 15, l = 0, r = arr.length - 1;
        String pair = "none";
        while (l < r) {
            int s = arr[l] + arr[r];
            if (s == want) { pair = arr[l] + "+" + arr[r]; break; }
            else if (s < want) l++;
            else r--;
        }
        System.out.println("twoPointer=" + pair);            // 4+11

        // 슬라이딩 윈도우 (길이 k 구간의 최대 합)
        int[] vals = {2, 1, 5, 1, 3, 2};
        int k = 3, windowSum = 0, best = Integer.MIN_VALUE;
        for (int i = 0; i < vals.length; i++) {
            windowSum += vals[i];
            if (i >= k) windowSum -= vals[i - k];            // 창을 벗어난 왼쪽 빼기
            if (i >= k - 1) best = Math.max(best, windowSum);
        }
        System.out.println("windowMaxSum=" + best);          // 9 (5+1+3)

        // DP 1D (Davis 같은 점화식): dp[i] = 이전 상태들의 조합
        int n = 7;
        int[] dp = new int[n + 1];
        dp[0] = 1;
        for (int i = 1; i <= n; i++) {
            dp[i] = dp[i - 1];
            if (i >= 2) dp[i] += dp[i - 2];
            if (i >= 3) dp[i] += dp[i - 3];
        }
        System.out.println("dp ways(7)=" + dp[n]);           // 44
    }
}
