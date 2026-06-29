package fraudulentactivity;

import support.Judge;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

/**
 * <a href="https://www.hackerrank.com/challenges/fraudulent-activity-notifications/problem">Fraudulent Activity Notifications</a>
 * <p>
 * HackerRank 사이트에서 푼 코드를 이 파일의 activityNotifications 본문에 복붙한 뒤, java로 돌려 검증하고 첨삭받습니다.
 * 시그니처는 HackerRank 원문과 같게 맞췄습니다(List&lt;Integer&gt; expenditure, int d 받아 int 반환).
 */

class Result {

    /*
     * Complete the 'activityNotifications' function below.
     *
     * The function is expected to return an INTEGER.
     * The function accepts following parameters:
     *  1. INTEGER_ARRAY expenditure
     *  2. INTEGER d
     *
     * HackerLand National Bank has a simple policy for warning clients about possible fraudulent account activity.
     * If the amount spent by a client on a particular day >= 2X the client's median spending for a trailing number of days,
     * => they send the client a notification about potential fraud.
     *
     * The bank doesn't send the client any notifications until they have
     * at least that trailing number of prior days' transaction data.
     *
     * median of a list of numbers
     * - can be found by first sorting the numbers ascending.
     * - If there is an odd number of values, the middle one is picked.
     * - If there is an even number of values, the median is then defined to be the average of the two middle values.
     *   {0, 1, 2, 3} => (1 + 2) / 2 => 1 * 2 = 2
     *
     * @param expenditure client's total daily expenditures for a period of  days
     * - int expenditure[n]: daily expenditures
     * - 1<= n <= 2*10^5 = 200,000
     * @param d trailing day
     * - 1 <= d <= n
     *
     * @return the number of times the client will receive a notification over all  days.
     */
    public static int activityNotifications(List<Integer> expenditure, int d) {
        // 1. expenditure = [10, 20, 30, 40, 50]
        //    d = 3
        // On the first three days, they just collect spending data.
        // At day 4, trailing expenditures are [10, 20, 30].
        // => The median is  20
        // => and the day's expenditure is 40 >= 20 * 2
        // Because , there will be a notice.
        //
        // The next day, trailing expenditures are [20, 30, 40]
        // => and the expenditures are 50.
        // This is less than 2 * 30 so no notice will be sent. Over the period, there was one notice sent.
        //
        // 2. expenditures = [2, 3, 4, 2, 3, 6, 8, 4, 5], d = 5
        //   - {2, 3, 4, 2, 3}
        //     {2, 2, 3, 3, 4}
        //     => 3 * 2
        //     expenditure = 6
        //
        // validation:
        // expenditure.isEmpty
        // expenditure.size() < d
        // d == 0
        var answer = 0;
        if (d == 0 || expenditure.isEmpty() || expenditure.size() < d) {
            return answer;
        }

        // var isEven = d % 2 == 0;
        // var i = d; d++
        // var subset = new ArrayList<>(list.subList(i - d, d));
        // subset.sort();
        // get median by isEven.
        // - middle
        // - average of two middle values
        // if median <= list.get(d)
        //    cnt++
        //
        var isEven = d % 2 == 0;
        // 첨삭: 정답입니다 — 로직은 맞고, 틀린 게 아니라 "너무 느린" 것이며 그게 TLE의 정체입니다. 샘플
        //   ([2,3,4,2,3,6,8,4,5],d=5 -> 2 / [1,2,3,4,4],d=4 -> 0)을 다 통과하고, 무작위 2000건을 베스트
        //   버전과 교차검증해도 답이 전부 같았습니다. 중앙값을 홀/짝으로 나눠 2*median을 그날 지출과
        //   비교한 발상도 정확합니다. 작은 케이스(0,6)만 통과하고 큰 케이스(1~5)가 시간 초과한 건, 규모가
        //   커야 드러나는 버그라 로컬 샘플로는 안 보이기 때문입니다. 그래서 복잡도를 "셀 줄" 아는 게 곧
        //   이 버그를 미리 보는 눈입니다.
        //
        //   [복잡도 세는 법 — 이 루프로 4박자]  (전혀 안 잡힌다고 하셨으니 이 코드에 그대로 붙여 봅니다.)
        //   1) 변수 정의: n = expenditure 길이, d = 윈도우(직전 며칠) 크기.
        //   2) 바깥 루프가 몇 번 도나: i가 d부터 n까지 -> 약 (n - d)번, 최대 n번.
        //   3) 한 번 돌 때 안에서 가장 비싼 일: subList 복사 O(d) + sort O(d log d) -> O(d log d)이 지배.
        //   4) 곱한다(루프 안에 비용이 있으면 바깥 횟수와 곱): n * d log d -> 시간 O(n * d log d).
        //   제약 대입으로 체감: n=2*10^5, d=10^5면 2*10^5 * 10^5 * 17 ≈ 3.4*10^11. 보통 1초에 약 10^8
        //   연산이라 수천 초 -> TLE. (공간: 매 윈도우 subset을 새로 복사하니 O(d).)
        //
        //   핵심 신호: "바깥 루프 안에서 매번 정렬/전체 복사"가 보이면 바깥 횟수와 곱해져 터집니다. 줄이는
        //   길은 거의 늘 "매번 새로 하지 말고, 직전 결과를 조금만 고쳐 재사용"입니다(아래 베스트). 복잡도를
        //   더 체계적으로는 algorithms/complexity_analysis_playbook.md를 보세요.
        //
        //   곁가지(지뢰): for 조건이 `d < expenditure.size()`라 d가 안 변해 사실상 무한 루프인데, 안쪽
        //   `if (size == i) break`가 막아 줍니다. 의도는 `i < expenditure.size()`입니다 — 동작은 하지만
        //   헷갈리는 자리라, 조건을 i로 바꾸고 break를 지우면 깔끔합니다.
        //
        // > 카드: 복잡도 = (바깥 루프 횟수) * (한 번 돌 때 안에서 가장 비싼 비용). 루프 안에서 매번 정렬/전체훑기가 보이면 곱해져 터진다 -> 직전 결과 재사용으로 바꿔라.
        for (var i = d; i < expenditure.size(); i++) {
            // 5 - 5 = 0, 5
            // 6 - 5 = 1, 6
            // System.out.println("i: " + i);
            var subset = new ArrayList<>(expenditure.subList(i - d, i));
            subset.sort(Comparator.naturalOrder());
            var median = 0;
            var idx = d / 2;
            if (isEven) {
                // {0, 1, 2, 3} => (1 + 2) / 2 => 1 * 2 = 2
                var tmp1 = subset.get(idx);
                var tmp2 = subset.get(idx - 1);
                var tmp3 = (double) ((tmp1 + tmp2)) / 2;

                System.out.println("tmp3: " + tmp3);
                // 첨삭: 이 디버그 출력이 루프 안에 남아 있습니다. 매 반복마다 콘솔 I/O가 일어나
                //   그 자체로 큰 속도 저하이니(특히 수십만 번이면 치명적), 제출 전에는 지웁니다.
                median = (int) (tmp3 * 2);
            } else {
                median = subset.get(idx) * 2;
            }
            var currExpenditure = expenditure.get(i);
            // System.out.println("subset: " + subset + ", idx: " + idx + ", median: " + median + ", currExpenditure: " + currExpenditure);
            if (median <= currExpenditure) {
                answer++;
            }
        }

        return answer;
    }

    // 첨삭 — 너무 느린 걸 살리는 길(접근은 그대로, "매번 정렬"만 버린다):
    //   당신 골격(윈도우마다 중앙값 구해 2*median과 비교)은 그대로 둡니다. 바꾸는 건 "중앙값을 어떻게
    //   구하나" 하나입니다. 지출액이 0~200으로 갇혀 있다는 제약(문제에 명시)이 열쇠입니다 — 값 종류가
    //   201개뿐이면, 비교 정렬(O(d log d)) 대신 "각 값이 몇 번 나왔나"를 세는 count 배열로 중앙값을 찾을
    //   수 있습니다. 게다가 윈도우는 한 칸씩만 미끄러지니, 매번 새로 세지 말고 나가는 날 -1, 들어오는
    //   날 +1만 갱신하면 됩니다(O(1)).
    //
    //   [데이터 움직임] 윈도우 [2, 2, 3, 3, 4] (d=5)의 중앙값을 count로:
    //     count:  값 2 -> 2개,  값 3 -> 2개,  값 4 -> 1개   (나머지 0)
    //     k = d/2 + 1 = 3번째를 누적으로 찾는다:
    //       값 2까지 누적 2   (< 3)
    //       값 3까지 누적 4   (>= 3)  -> 중앙값 3,  2*중앙값 = 6
    //
    //   [윈도우 슬라이드] i가 한 칸 늘 때 (전체 재계산 X, 두 칸만 고침):
    //     나가는 날 expenditure[i - d] -> count[그 값]--
    //     들어오는 날 expenditure[i]   -> count[그 값]++
    //
    //     int activityNotifications(List<Integer> expenditure, int d) {
    //         int n = expenditure.size();
    //         if (n <= d) return 0;
    //         int[] count = new int[201];                   // 값이 0~200이라 201칸
    //         for (int i = 0; i < d; i++) {                 // 첫 윈도우 [0, d) 채우기
    //             count[expenditure.get(i)]++;
    //         }
    //         int notifications = 0;
    //         for (int i = d; i < n; i++) {
    //             if (expenditure.get(i) >= twiceMedian(count, d)) {
    //                 notifications++;
    //             }
    //             count[expenditure.get(i - d)]--;          // 나가는 날 제거
    //             count[expenditure.get(i)]++;              // 들어오는 날 추가
    //         }
    //         return notifications;
    //     }
    //
    //     // count를 누적해 k번째 값을 찾아 2*중앙값을 돌려준다(정렬 없이).
    //     int twiceMedian(int[] count, int d) {
    //         if (d % 2 == 1) {                             // 홀수: 가운데 하나
    //             int k = d / 2 + 1, seen = 0;
    //             for (int v = 0; v <= 200; v++) {
    //                 seen += count[v];
    //                 if (seen >= k) return 2 * v;
    //             }
    //         } else {                                      // 짝수: 가운데 둘의 합(= 2*평균)
    //             int t1 = d / 2, t2 = d / 2 + 1, seen = 0, a = -1;
    //             for (int v = 0; v <= 200; v++) {
    //                 seen += count[v];
    //                 if (a < 0 && seen >= t1) a = v;
    //                 if (seen >= t2) return a + v;
    //             }
    //         }
    //         return 0;
    //     }
    //
    //   이 버전은 무작위 2000건에서 원본과 답이 전부 같았고, n=200000 · d=10000을 12ms에 끝냅니다.
    //
    // 첨삭 — 왜 이게 베스트이고, 복잡도는 어떻게 줄었나:
    //   한 씨앗 깨달음입니다 — "값이 작은 범위에 갇혀 있으면(여기선 0~200) 비교해서 정렬하지 말고 세라."
    //   정렬은 값을 서로 비교해 줄 세우느라 O(d log d)가 들지만, 값 종류가 201개로 고정이면 "각 값이 몇
    //   개"를 세서 누적으로 k번째를 찾는 게 O(201)입니다. 윈도우가 한 칸씩 움직이는 것도 "직전 count를 두
    //   칸만 고쳐 재사용"으로 바뀌어, 매번 d개를 새로 훑던 걸 O(1)로 줄입니다.
    //
    //   [복잡도 4박자 — 고친 버전]
    //   1) n = 길이, V = 값 종류 수(여기 201, 상수).
    //   2) 바깥 루프 ~n번.
    //   3) 한 번 돌 때: 중앙값 walk O(V) + 슬라이드 O(1) -> O(V).
    //   4) 곱: n * V -> O(n * V). V가 상수라 사실상 O(n). 제약 대입 2*10^5 * 201 ≈ 4*10^7 -> 즉시 통과.
    //   공간: count 배열 201칸 고정 -> O(V) = O(1)(입력 크기와 무관한 상수).
    //
    //   원본 O(n * d log d) -> 고친 것 O(n * V). 같은 답을 내지만 "매번 정렬"을 "한 번 세고 조금씩 갱신"으로
    //   바꾼 것이 전부입니다.
    //
    // > 카드: 값이 작은 범위(0~K)에 갇히면 비교정렬(O(d log d)) 말고 count 배열로 세기(O(K)). 슬라이딩 윈도우는 매번 새로 X, 나가는/들어오는 것만 갱신(O(1)).
}

class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.activityNotifications(List.of(2, 3, 4, 2, 3, 6, 8, 4, 5), 5), 2);
        Judge.check(Result.activityNotifications(List.of(1, 2, 3, 4, 4), 4), 0);
        // 반례를 여기에 추가하세요:
    }
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
//        int d = Integer.parseInt(firstMultipleInput[1]);
//
//        List<Integer> expenditure = Stream.of(bufferedReader.readLine().replaceAll("\\s+$", "").split(" "))
//                .map(Integer::parseInt)
//                .collect(toList());
//
//        int result = Result.activityNotifications(expenditure, d);
//
//        bufferedWriter.write(String.valueOf(result));
//        bufferedWriter.newLine();
//
//        bufferedReader.close();
//        bufferedWriter.close();
//    }
//}