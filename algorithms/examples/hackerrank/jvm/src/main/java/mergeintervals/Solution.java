package mergeintervals;

import support.Judge;

/**
 * Merge Intervals — LeetCode 56 (대표 Interval 문제, 라이브코딩 단골).
 * <a href="https://leetcode.com/problems/merge-intervals/">문제</a>
 * <p>
 * 겹치는 구간들을 합쳐, 겹치지 않는 구간 목록을 시작값 기준 오름차순으로 반환합니다.
 * 예: [[1,3],[2,6],[8,10],[15,18]] -> [[1,6],[8,10],[15,18]]  ([1,3]과 [2,6]이 겹쳐 [1,6])
 * <p>
 * 막히면(10분 룰): 손으로 구간을 수직선 위에 그려 "언제 둘이 겹치나"를 먼저 말로 정의해 보세요.
 * 시작값으로 정렬해 두면 겹침 판정이 한 방향으로 단순해진다는 게 이 문제의 씨앗입니다.
 */
class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.merge(new int[][]{{1, 3}, {2, 6}, {8, 10}, {15, 18}}),
                new int[][]{{1, 6}, {8, 10}, {15, 18}});
        Judge.check(Result.merge(new int[][]{{1, 4}, {4, 5}}),
                new int[][]{{1, 5}});
        Judge.check(Result.merge(new int[][]{{1, 4}, {2, 3}}),
                new int[][]{{1, 4}});
        // 반례를 여기에 추가하세요:
    }
}

class Result {
    public static int[][] merge(int[][] intervals) {
        // Write your code here
        return new int[0][];
    }
}
