package fraudulentactivity;

import support.Judge;

import java.util.List;

/**
 * <a href="https://www.hackerrank.com/challenges/fraudulent-activity-notifications/problem">Fraudulent Activity Notifications</a>
 * <p>
 * HackerRank 사이트에서 푼 코드를 이 파일의 activityNotifications 본문에 복붙한 뒤, java로 돌려 검증하고 첨삭받습니다.
 * 시그니처는 HackerRank 원문과 같게 맞췄습니다(List&lt;Integer&gt; expenditure, int d 받아 int 반환).
 */
class Solution {
    public int activityNotifications(List<Integer> expenditure, int d) {
        return 0;
    }

    static void main() {
        var sol = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(sol.activityNotifications(List.of(2, 3, 4, 2, 3, 6, 8, 4, 5), 5), 2);
        Judge.check(sol.activityNotifications(List.of(1, 2, 3, 4, 4), 4), 0);
        // 반례를 여기에 추가하세요:
    }
}
