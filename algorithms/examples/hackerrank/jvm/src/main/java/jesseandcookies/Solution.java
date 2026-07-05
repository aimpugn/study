package jesseandcookies;

import support.Judge;

import java.util.List;
import java.util.PriorityQueue;

/**
 * <a href="https://www.hackerrank.com/challenges/jesse-and-cookies/problem">Jesse and Cookies</a>
 * <p>
 * HackerRank 사이트에서 푼 코드를 이 파일의 cookies 본문에 복붙한 뒤, java로 돌려 검증하고 첨삭받습니다.
 * 시그니처는 HackerRank 원문과 같게 맞췄습니다(int k, List&lt;Integer&gt; A 받아 int 반환).
 */
class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.cookies(7, List.of(1, 2, 3, 9, 10, 12)), 2);
        Judge.check(Result.cookies(10, List.of(1, 1, 1)), -1);
        // 반례를 여기에 추가하세요:
    }
}

class Result {

    /*
     * Complete the 'cookies' function below.
     *
     * The function is expected to return an INTEGER.
     * The function accepts following parameters:
     *  1. INTEGER k
     *  2. INTEGER_ARRAY A
     *
     * Jesse loves cookies and wants the sweetness of some cookies to be greater than value `k`
     * To do this,
     * - two cookies with the least sweetness are repeatedly mixed.
     * - This creates a special combined cookie with:
     *   sweetness  = 1 Least sweet cookie + 2 * 2nd least sweet cookie
     *   This occurs until all the cookies have a sweetness `k`
     *
     * @param A the sweetness of a number of cookies
     * @return the minimum number of operations required. If it is not possible -1
     */

    public static int cookies(int k, List<Integer> A) {
        // Write your code here
        // [2(x), 7, 3(x), 6, 4, 6], k = 9
        // smallest: 2, 3 => 2 + 3*2 = 8
        // => [8, 7, 6(x), 4(x), 6]
        //    smallest: 4, 6 => 4 + 6*2 = 16
        // => [16, 8, 7(x), 6(x)]
        //    smallest: 7, 6 => 6 + 7*2 = 20
        // => [20, 16, 8]
        //
        // Use priority queue ascending
        // validation
        if (A.size() <= 1) return -1;
        if (A.size() == 2) {
            if (A.get(0) + A.get(1) < k) {
                return -1;
            } else {
                return 1;
            }
        }

        // 첨삭: 접근은 맞습니다 — 우선순위 큐(min-heap)로 두 최솟값을 꺼내 섞는 골격이 정확하고 샘플도
        //   통과합니다. 좋은 힙 직감이에요. 다만 2/28 실패(WA)에는 버그가 셋 있고, 셋 다 "큰 입력/경계"에서만
        //   드러나 작은 샘플론 안 보입니다.
        //
        //   1) int 오버플로 (제일 큰 원인). 아래 q.offer(least1 + (least2 * 2))가 int 연산인데, 그 결합값이
        //      int가 담을 수 있는 한계를 넘어섭니다. 무엇이 얼마인지 차례로:
        //      - int 범위: -2,147,483,648 부터 +2,147,483,647 까지 (약 ±21억; 4바이트=32비트=2^32가지 값)
        //      - 값 A[i]: 최대 10^6(1,000,000, 백만)
        //      - k: 최대 10^9(1,000,000,000, 10억)
        //      - 결합값(least1 + 2*least2): 섞을수록 커져 최대 약 3*10^9(3,000,000,000, 30억) ('약'은 '대략')
        //      - 넘침: 30억 > 21억(int 최대)이라 int로는 표현 못 하고, 값이 '한 바퀴 돌아'(2^32로 나눈 나머지)
        //        엉뚱해집니다 — 21억~43억이면 음수로, 더 크면 작은 양수로 뒤집힘.
        //        ex) int에서 15억 + 2*15억 = 45억(4,500,000,000)이어야 하지만 실제로는 205,032,704가 나옴
        //            (45억에서 2^32(4,294,967,296, 약 43억)를 한 번 빼고 남은 값)
        //      이렇게 한 번 엉뚱한 값이 나오면, 그게 힙에서 '가장 작은 값'인 척 맨 앞에 앉아 정렬 순서를
        //      무너뜨리고, 결국 엉뚱한 답(주로 -1)이 납니다. 그래서 더 큰 정수 타입 long으로 받아야 합니다 —
        //      long은 최대 9,223,372,036,854,775,807(약 922경; 8바이트=64비트)까지라 30억쯤은 여유롭게 담습니다.
        //      고침 = 힙을 PriorityQueue<Long>로. 실측: n=200, k=10억에서 이 코드는 -1, long으로 고치면 198.
        //      (한 끗 신호: 제약을 읽어 타입을 정한다 — "k가 10억, 결합으로 더 커짐 -> int(21억) 넘침 -> long".)
        //
        //   2) "이미 만족"을 0으로 안 본다. 처음부터 모든 값이 k 이상이면 답은 0인데, 아래 while은 poll->offer를
        //      먼저 하고 검사는 나중이라 무조건 한 번 섞어 cnt>=1을 돌려줍니다. 실측: cookies(5,[10,20,30])이
        //      코드는 1, 정답은 0.
        //
        //   3) 위 size==2 특수분기가 틀렸습니다. `A.get(0)+A.get(1) < k`(합)로 판정하는데, 한 번 섞으면 값은
        //      합이 아니라 min + 2*max입니다. 그래서 [1,5], k=10에서 이 코드는 -1(1+5=6<10)이지만 실제론
        //      1+2*5=11>=10이라 1회로 됩니다. 둘 다 이미 >=k여도 0이 아니라 1을 줍니다. 이 분기는 버그일 뿐
        //      아니라 불필요합니다 — 아래 일반 루프가 size 2도 올바르게 처리합니다.
        //
        //   고치는 핵심은 구조입니다: "섞고 나서 검사" 말고 "min이 k 미만인 동안만 섞는다"로 뒤집습니다.
        //     while (q.peek() < k) { if (q.size() < 2) return -1;  두 최솟값 섞기;  ops++; }  return ops;
        //   이러면 (2)이미 만족=루프 0회=0, (3)size 2도 자동, 섞을 둘 없으면 -1이 한 줄로 정리됩니다.
        //
        //   [데이터 움직임] min-heap (k=9, A=[2,7,3,6,4,6]) — 예제의 중간 배열 오타는 무시하고 실제로:
        //     [2,3,4,6,6,7]  poll 2,3 -> offer 2+2*3=8   -> [4,6,6,7,8]   ops1
        //     [4,6,6,7,8]    poll 4,6 -> offer 4+2*6=16  -> [6,7,8,16]    ops2
        //     [6,7,8,16]     poll 6,7 -> offer 6+2*7=20  -> [8,16,20]     ops3  (= 당신이 의문 가진 [20,16,8] 그 자리, 정확)
        //     [8,16,20]      min 8<9  -> poll 8,16 -> offer 8+2*16=40 -> [20,40]   ops4
        //     min 20>=9 -> 멈춤. 4회.
        //
        // > 카드: 힙 결합 문제 = (a) 결합값이 int 넘는지 제약으로 보고 long, (b) "이미 만족=0"을 while(min<k)로 자연히, (c) 섞을 둘 없으면 -1. 섞고나서 검사 말고 조건을 루프에 건다.
        var q = new PriorityQueue<>(A);     // O(n): A의 n개로 힙 빌드(heapify는 한 번에 O(n))
        var cnt = 0;
        while (true) {                       // 최대 n-1번 반복(섞을 때마다 원소가 1개씩 줆)
            // if after all, remained value is lower than k?
            if (q.size() == 1 && q.peek() < k) {
                return -1;
            }
            // if one least value greater or equal than k???
            var least1 = q.poll();           // O(log n): 최소 꺼내기
            var least2 = q.poll();           // O(log n): 최소 꺼내기
            q.offer(least1 + (least2 * 2));   // O(log n): 다시 넣기
            cnt++;
            if (k <= q.peek()) {
                break;
            }
        }
        // 첨삭 — 최종 시간/공간 복잡도(위 마커 합성):
        //   시간 = heapify O(n)  +  [반복 최대 n번] × [한 번에 poll·poll·offer = O(log n)]
        //        = O(n) + O(n log n).  순차 합이라 지배항만 남겨 -> O(n log n).
        //   공간 = 힙이 원소 n개를 들고 있음 -> O(n).
        return cnt;
    }

    // 첨삭 — 고친 동작 버전(접근 그대로, 셋 다 해결):
    //   힙을 long으로(오버플로), 특수분기 제거(불필요·버그), 루프 조건을 "min<k인 동안"으로 뒤집기.
    //
    //     public static int cookies(int k, List<Integer> A) {
    //         PriorityQueue<Long> q = new PriorityQueue<>();   // long: 결합값이 int를 넘는다
    //         for (int a : A) {                 // O(n): n번 add (총 O(n))
    //             q.add((long) a);
    //         }
    //         int ops = 0;
    //         while (q.peek() < k) {            // 최대 n-1번 (섞을 때마다 원소 1개 줆)
    //             if (q.size() < 2) return -1;  // 섞을 둘이 없으면 불가능
    //             long least1 = q.poll();        // O(log n): 최소 꺼내기
    //             long least2 = q.poll();        // O(log n): 최소 꺼내기
    //             q.offer(least1 + 2 * least2);  // O(log n): 다시 넣기
    //             ops++;
    //         }
    //         return ops;                       // 루프 0회 = 이미 만족 = 0
    //     }
    //
    //   검증: 샘플 2/-1 통과, cookies(5,[10,20,30])=0, [1,5] k=10 = 1, n=200 k=10억(오버플로 케이스)도 정답.
    //
    // 첨삭 — 왜 이 구조가 더 나은가:
    //   핵심은 "불변식을 루프 조건에 거는 것"입니다. 목표 "모든 값 >= k"는 곧 "최솟값 >= k"이고, 그 부정
    //   "최솟값 < k"가 바로 "더 섞어야 한다"입니다. while (q.peek() < k)가 그 목표를 그대로 코드로 옮긴 것이라,
    //   이미 만족(0회)·도중 종료·불가능(-1)이 전부 한 조건에서 따라 나옵니다. "섞고 나서 검사"는 그 자연스러운
    //   순서를 뒤집어 경계(0회, 마지막 원소)를 놓칩니다.
    //
    //   최종 시간/공간 복잡도(위 마커 합성): 시간 = heapify O(n) + [최대 n-1회] × [poll·poll·offer = O(log n)]
    //   = O(n) + O(n log n) -> 지배항만 남겨 O(n log n). 공간 = 힙 원소 n개 -> O(n).
    //   n=10^6(1,000,000, 백만)이라 O(n log n)이 적정선.
    //
    // > 카드: 목표가 "최소>=k"면 루프 조건은 그 부정 "while(min<k)". 불변식을 조건에 걸면 0회·불가능·종료가 공짜로 닫힌다.
}

//public class Solution {
//    public static void main(String[] args) throws IOException {
//        BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(System.in));
//        BufferedWriter bufferedWriter = new BufferedWriter(new FileWriter(System.getenv("OUTPUT_PATH")));
//
//        String[] firstMultipleInput = bufferedReader.readLine().replaceAll("\\s+$", "").split(" ");
//
//        int n = Integer.parseInt(firstMultipleInput[0]);
//
//        int k = Integer.parseInt(firstMultipleInput[1]);
//
//        List<Integer> A = Stream.of(bufferedReader.readLine().replaceAll("\\s+$", "").split(" "))
//                .map(Integer::parseInt)
//                .collect(toList());
//
//        int result = Result.cookies(k, A);
//
//        bufferedWriter.write(String.valueOf(result));
//        bufferedWriter.newLine();
//
//        bufferedReader.close();
//        bufferedWriter.close();
//    }
//}

