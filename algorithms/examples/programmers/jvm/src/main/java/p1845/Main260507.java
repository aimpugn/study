package p1845;

import java.util.Arrays;
import java.util.HashSet;
import java.util.List;

/**
 * @see <a href=
 * "https://school.programmers.co.kr/learn/courses/30/lessons/1845">
 * 포켓몬</a>
 */
public class Main260507 {

    /**
     * 총 N 마리의 포켓몬 중에서 N/2 마리를 가져도 좋습니다.
     * 포켓몬은 종류에 따라 번호를 붙여서 구분합니다.
     * <p>
     * 처음에는 N/2마리를 고르는 모든 조합을 세고 싶어질 수 있습니다.
     * 하지만 이 문제의 답은 선택 방법의 개수가 아니라 선택 결과에 들어간
     * 서로 다른 포켓몬 종류의 최댓값입니다.
     * <p>
     * 예를 들어 [3, 1, 2, 3]에서는 2마리만 고를 수 있고, 전체 종류는
     * 3가지입니다. 아무리 종류가 많아도 2마리만 고르므로 답은 2입니다.
     * [3, 3, 3, 2, 2, 4]에서는 3마리를 고를 수 있고, 전체 종류도
     * 3가지이므로 답은 3입니다.
     *
     * @param nums 포켓몬의 종류 번호가 담긴 1차원 배열
     * - 1 <= nums.length <= 10,000. 자연수이며 항상 짝수
     * - 1 <= nums[i] <= 200,000. 자연수.
     *
     * @return N/2 마리 포켓몬을 선택하는 방법 중, 가장 많은 종류의 포켓몬을 선택하는 최댓값
     * - 가장 많은 종류의 포켓몬을 선택하는 방법이 여러가지인 경우에도, 최댓값 하나만 리턴
     */
    public static int solution(int[] nums) {
        // 포켓몬별 개수는 필요하지 않고, 서로 다른 종류가 몇 개인지만 필요합니다.
        var uniqueTypes = new HashSet<Integer>();
        for (var num : nums) {
            uniqueTypes.add(num);
        }

        return Math.min(nums.length / 2, uniqueTypes.size());
    }

    record TestCase(int[] input, int answer) {
    }

    public static void main(String[] args) {
        var testCases = List.of(
            new TestCase(new int[]{3, 1, 2, 3}, 2),
            new TestCase(new int[]{3, 3, 3, 2, 2, 4}, 3),
            new TestCase(new int[]{3, 3, 3, 2, 2, 2}, 2)
        );

        for (var testCase : testCases) {
            var param = testCase.input();
            var answer = testCase.answer();
            var actual = solution(param);

            if (actual != answer) {
                throw new AssertionError(
                    "input=" + Arrays.toString(param)
                        + ", expected=" + answer
                        + ", actual=" + actual);
            }
        }
        System.out.println("Success");
    }
}
