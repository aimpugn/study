package p42885;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * [구명보트](https://school.programmers.co.kr/learn/courses/30/lessons/42885?language=java)
 */
class Solution {

    /**
     * 무인도 갇힌 사람들을 구명보트로 구출
     * 구명보드 최대 2명, 무게 제한 있음
     * <p>
     * 목표: 구명 보트 최대한 적게 사용하여 모든 사람 구출
     *
     * <p>
     * 예제:
     * - people: [70kg, 50kg, 80kg, 50kg]
     * - 보트 무게 limit: 100kg
     * <p>
     * people[1], people[3] 같이 탈 수 있음
     * people[0] + people[2] = 150kg => 같이 탈 수 없음
     *
     * @param people 무인도에 갇힌 사람
     *               - 1 <= people.length <= 50,000
     *               - 40 <= people[i] <= 240
     * @param limit  구명보트 무게 제한
     *               - 40 <= limit <= 240
     *               - 항상 사람 몸무게 중 최댓값보다 크게 주어지므로, 구출할 수 없는 경우는 없음
     * @return 필요한 구명보트 개수의 최솟값
     */
    public int solution(int[] people, int limit) {
        int answer = 0;

        // 정렬하여 몸무게가 작은 사람부터 보트에 태워봅니다.
        // in place로 정렬
        Arrays.sort(people);

        // 구명 보트 무게 제한을 넘지 않게 보트에 태워 봅니다.
        // 조건에 따라 한 명은 무조건 태울 수 있습니다.
        // 최대 두 명이므로, 두 수의 합이 최대가 되게 합니다.
        int left = 0;
        int right = people.length - 1;

        while (left <= right) {
            // 두 사람을 태울 수 있는 경우 없는 경우를 if ~ else 로 나누는 경우
            // if (people[left] + people[right] <= limit) {
            //     left++;
            //     right--;
            // } else {
            //     // 두 사람을 태울 수 없다면:
            //     // 가장 무거운 사람만 태워서, 점차 가벼운 사람과 같이 탈 수 있는 사람만 남깁니다.
            //     right--;
            // }
            // 이를 줄이면 아래와 같이 줄일 수 있습니다:
            if (people[left] + people[right] <= limit) {
                left++;
            }
            right--; // 무거운 사람은 무조건 타게 됩니다.
            answer++;
        }

        return answer;
    }
}

class TestCase {
    public int[] people;
    public int limit;
    public int answer;

    TestCase(int[] people, int limit, int answer) {
        this.people = people;
        this.limit = limit;
        this.answer = answer;
    }
}

public class Main {
    public static void main(String[] args) {
        List<TestCase> cases = new ArrayList<>();
        cases.add(new TestCase(new int[]{70, 50, 80, 50}, 100, 3));
        cases.add(new TestCase(new int[]{70, 80, 50}, 100, 3));
        cases.add(new TestCase(new int[]{50, 50, 50, 50, 50}, 100, 3));
        cases.add(new TestCase(new int[]{30, 30, 30, 30, 30}, 100, 3));
        cases.add(new TestCase(new int[]{50, 50, 50, 50, 50, 50}, 100, 3));
        cases.add(new TestCase(new int[]{10, 20, 30, 40, 50, 60, 70}, 100, 4));

        Solution s = new Solution();
        for (TestCase tc : cases) {
            System.out.println("answer: " + (s.solution(tc.people, tc.limit) == tc.answer));
        }
    }
}
