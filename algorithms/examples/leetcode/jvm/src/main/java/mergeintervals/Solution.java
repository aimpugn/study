package mergeintervals;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Comparator;

/**
 * https://leetcode.com/problems/merge-intervals/
 * <p>
 * 풀이 회고·복습 카드·암기 포인트는 같은 폴더의 PROCESS.md 참고.
 */
class Solution {
    /**
     * 겹치는 모든 구간을 병합해, 서로 겹치지 않는 구간 배열을 돌려준다.
     *
     * @param intervals [start, end] 배열 (1 <= length <= 10^4, 0 <= start <= end <= 10^4)
     * @return 병합 후 서로 겹치지 않는 구간들
     *
     * <p>핵심 한 줄: start로 정렬(겹칠 것들을 이웃에 모으는 '정규화') → '지금까지 만든 마지막 구간'을
     * 들고 가며, 겹치면 그 끝을 max로 늘린다. 비교 상대는 원본 이웃(intervals[i-1])이 아니라 결과의 마지막이다.
     */
    public int[][] merge(int[][] intervals) {
        // [문제 분석 — 기록 보존]
        // 1. [[1,3],[2,6],[8,10],[15,18]] → 1~6 사이에 2,3 → merge → [[1,6],[8,10],[15,18]]
        // 2. [[1,4],[4,5]]               → 끝이 닿아도(4=4) merge → [[1,5]]
        // 3. [[4,7],[1,4]]               → 정렬하면 [[1,4],[4,7]] → [[1,7]]
        // 연달아 셋이 오버랩되면 셋을 머지해야 한다. → start로 정렬한 뒤 마지막 구간의 끝을 늘려 간다.

        // 정렬 기준은 start.
        //   이전: (a, b) -> Integer.compare(a[0], b[0])   // 두 인자를 직접 비교하는 '절차'
        //   개선: comparingInt(iv -> iv[0])               // 'iv[0]으로 정렬'이라는 '의도'가 바로 보인다
        Arrays.sort(intervals, Comparator.comparingInt(interval -> interval[0]));

        var merged = new ArrayList<int[]>();
        merged.add(intervals[0]); // 첫 구간을 깔아, 이것을 '마지막 구간'의 출발점으로 삼는다.

        for (var i = 1; i < intervals.length; i++) {
            var current = intervals[i];
            var last = merged.getLast(); // 비교 상대 = 누적한 '마지막 결과 구간'(원본 이웃 아님). Java 21 메서드.

            if (current[0] <= last[1]) {
                // 겹침(끝이 닿기만 해도 <= 로 병합). current를 새 칸으로 넣지 않고 마지막 구간의 끝만 늘린다.
                // last는 merged 안의 배열을 '그대로 가리키는 참조'라, last[1]을 바꾸면 결과도 함께 갱신된다.
                // current가 마지막 구간에 통째로 들어갈 수 있어([1,8] ⊃ [2,4]) 끝이 줄지 않게 max.
                last[1] = Math.max(last[1], current[1]);
            } else {
                merged.add(current); // 안 겹침 → 마지막 구간을 확정하고, current로 새 구간을 시작.
            }
        }

        // toArray는 인자 배열의 각 칸을 리스트 원소(int[] 참조)로 '덮어쓴다'.
        //   이전: new int[size][2] → 안쪽 [2] 배열들이 채워지기도 전에 버려진다(낭비).
        //   개선: new int[0][]     → JVM이 정확한 크기로 한 번에 만든다(표준 관용구).
        return merged.toArray(new int[0][]);
        // 복잡도: 시간 O(n log n) — 정렬이 지배. 공간 O(n).
    }

    static void main() {
        var s = new Solution();
        // 출력 확인 대신 기대값과 비교 — 다양한 겹침 패턴을 한 번에 회귀 검증한다.
        check(s.merge(new int[][]{{1, 3}, {2, 6}, {8, 10}, {15, 18}}), new int[][]{{1, 6}, {8, 10}, {15, 18}});
        check(s.merge(new int[][]{{1, 4}, {4, 5}}),                     new int[][]{{1, 5}});  // 접점 병합
        check(s.merge(new int[][]{{4, 7}, {1, 4}}),                     new int[][]{{1, 7}});  // 정렬이 필요한 경우
        check(s.merge(new int[][]{{4, 7}, {1, 8}, {1, 4}}),             new int[][]{{1, 8}});  // 완전 포함
        check(s.merge(new int[][]{{1, 7}, {1, 8}, {2, 4}}),             new int[][]{{1, 8}});  // 같은 start
    }

    private static void check(int[][] actual, int[][] expected) {
        if (!Arrays.deepEquals(actual, expected)) {
            throw new AssertionError("expected=" + Arrays.deepToString(expected)
                + ", actual=" + Arrays.deepToString(actual));
        }
        System.out.println("PASS " + Arrays.deepToString(actual));
    }
}
