package p42587.attempt260511;

import support.TestCase2;

import java.util.Comparator;
import java.util.LinkedList;
import java.util.List;
import java.util.PriorityQueue;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42587?language=java">프로세스</a>
 * <p>
 * 준비 메타
 * - 공식 레벨: Lv.2
 * - 분류/트랙: Stage1/StackQueue
 * - 핵심 패턴: Queue/Priority
 * - 체감 난이도: 체감 Lv.2 중
 * - 예상 풀이 시간: 실전 숙련자 20~35분 / 학습 모드 50~80분
 * - 첫 접근: 더 높은 우선순위가 남아 있으면 현재 문서를 큐 뒤로 보낸다.
 */
class Solution {
    /**
     * 운영체게의 프로세스 관리 규칙이 다음과 같을 때, 특정 프로세스가 몇 번째로 실행되는지
     * 1. 실행 대기 큐에서 대기중인 프로세스를 하나 꺼냅니다.
     * 2. 큐에 대기중인 프로세스 중 우선순위가 더 높은 프로세스가 있다면 방금 꺼낸 프로세스를 다시 큐에 넣습니다.
     * 3. 만약 그런 프로세스가 없다면 방금 꺼낸 프로세스를 실행합니다.
     * 3.1. 한 번 실행한 프로세스는 다시 큐에 넣지 않고 그대로 종료됩니다.
     * <p>
     * 예: [A(2), B(1), C(3), D(2)]
     * => [C, D, A, B] 순으로 실행
     *
     * @param priorities 프로세스의 중요도가 순서대로 담긴 배열
     * - 1 <= priorities.length <= 100
     * - 1 <= priorities[i] <= 9
     * - priorities[i]는 우선순위를 나타내며, 숫자가 클수록 우선순위가 높습니다.
     * @param location 몇 번째로 실행되는지 알고싶은 프로세스의 위치
     * - 0 <= location <= priorities.length - 1
     *
     * @return 해당 프로세스가 몇 번째로 실행되는지
     */
    public int solution(int[] priorities, int location) {

        // 예: [1, 1, 9, 1, 1, 1], location=0
        // => 5
        // [A(1), B(1), C(9), D(1), E(1), F(1)]
        // [C, D, E, F, A, B]
        //
        // 우선순위가 가장 높은 것을 실행하고, 그 다음에는 그냥 순차적으로 실행합니다.
        // - 우선순위는 우선순위 큐로 해결합니다.
        // - location 위치의 프로세스를 나중에 언제 실행됐는지 확인할 수 있도록 식별할 수 있어야 합니다.

        var pq = new PriorityQueue<Integer>(Comparator.comparingInt(x -> -x));
        var queue = new LinkedList<Integer>();
        for (var i = 0; i < priorities.length; i++) {
            pq.offer(priorities[i]);
            queue.add(i);
        }

        var cnt = 0;
        while (!queue.isEmpty()) {
            var mostPriority = pq.peek();
            var idx = queue.poll();

            if (mostPriority == null) {
                break;
            }

            // 우선 순위 프로세스 실행
            if (priorities[idx] == mostPriority) {
                pq.poll();
                cnt++;

                // 실행된 프로세스가 찾던 거라면 cnt 리턴
                if (idx == location) {
                    return cnt;
                }
            } else {
                // 다시 넣는다는 건 대기 순서만 바뀌는 일이지,
                // 남은 우선순위 집합이 바뀌는 일이 아닙니다.
                queue.add(idx);
            }
        }

        return cnt;
    }

    /*
     * 테스트 1 〉	통과 (1.50ms, 65.5MB)
     * 테스트 2 〉	통과 (2.19ms, 62.4MB)
     * 테스트 3 〉	통과 (1.48ms, 60.7MB)
     * 테스트 4 〉	통과 (1.39ms, 62.5MB)
     * 테스트 5 〉	통과 (1.19ms, 64.2MB)
     * 테스트 6 〉	통과 (2.06ms, 61.8MB)
     * 테스트 7 〉	통과 (1.30ms, 62.2MB)
     * 테스트 8 〉	통과 (2.53ms, 59.8MB)
     * 테스트 9 〉	통과 (1.58ms, 62.9MB)
     * 테스트 10 〉	통과 (2.06ms, 58MB)
     * 테스트 11 〉	통과 (1.86ms, 61.7MB)
     * 테스트 12 〉	통과 (1.76ms, 62.8MB)
     * 테스트 13 〉	통과 (1.84ms, 61.2MB)
     * 테스트 14 〉	통과 (1.34ms, 62.8MB)
     * 테스트 15 〉	통과 (1.20ms, 63.7MB)
     * 테스트 16 〉	통과 (1.30ms, 60.8MB)
     * 테스트 17 〉	통과 (1.86ms, 61.9MB)
     * 테스트 18 〉	통과 (1.74ms, 62.7MB)
     * 테스트 19 〉	통과 (2.60ms, 62MB)
     * 테스트 20 〉	통과 (1.34ms, 63.1MB)
     */
}

public class Main260511 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        // TODO: Programmers 예제를 옮겨 적고 결과를 직접 확인하세요.
        // System.out.println(Arrays.toString(solution.solution(...)));

        var testCases = List.of(
                new TestCase2<>(new int[]{2, 1, 3, 2}, 2, 1),
                new TestCase2<>(new int[]{1, 1, 9, 1, 1, 1}, 0, 5)
        );

        for (var testCase : testCases) {
            System.out.println(solution.solution(testCase.input(), testCase.input2()));
        }
    }
}
