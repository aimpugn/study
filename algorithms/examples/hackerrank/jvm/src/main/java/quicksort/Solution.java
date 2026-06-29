package quicksort;

import support.Judge;

/**
 * <a href="https://www.hackerrank.com/challenges/quicksort2/problem">Quicksort 2 - Sorting</a>
 * <p>
 * HackerRank 사이트에서 푼 코드를 이 파일의 quickSort 본문에 복붙한 뒤, java로 돌려 검증하고 첨삭받습니다.
 * 정렬된 배열을 반환하도록 시그니처를 맞췄습니다(Arrays.sort 호출 말고 퀵정렬을 직접 구현하는 연습).
 */
class Solution {
    public int[] quickSort(int[] arr) {
        return new int[0];
    }

    static void main() {
        var sol = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(sol.quickSort(new int[]{4, 5, 3, 7, 2}), new int[]{2, 3, 4, 5, 7});
        Judge.check(sol.quickSort(new int[]{5, 4, 3, 2, 1}), new int[]{1, 2, 3, 4, 5});
        Judge.check(sol.quickSort(new int[]{1}), new int[]{1});
        // 반례를 여기에 추가하세요:
    }
}
