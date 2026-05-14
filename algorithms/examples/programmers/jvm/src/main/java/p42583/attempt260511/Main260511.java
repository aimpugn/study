package p42583.attempt260511;

import java.util.ArrayDeque;
import java.util.Arrays;
import java.util.List;
import java.util.Queue;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42583?language=java">다리를 지나는 트럭</a>
 * <p>
 * 준비 메타
 * - 공식 레벨: Lv.2
 * - 분류/트랙: Stage1/StackQueue
 * - 핵심 패턴: Queue/Simulation
 * - 체감 난이도: 체감 Lv.2 중
 * - 예상 풀이 시간: 실전 숙련자 25~40분 / 학습 모드 60~90분
 * - 첫 접근: 시간, 다리 위 트럭, 현재 무게를 함께 상태로 관리한다.
 */
class Solution {
    /**
     * 강을 가로지르는 일차선 다리를 정해진 순으로 건너려 합니다.
     * <p>
     * 목표:
     * 모든 트럭이 다리를 건너려면 최소 몇 초가 걸리는지 알아내기
     *
     * @param bridge_length 최대 다리에 올라갈 수 있는 트럭 대수
     * - 1 <= bridge_length <= 10,000
     * @param weight 다리가 견딜 수 있는 무게. 다리에 완전히 오르지 않은 트럭 무게는 무시합니다.
     * - 1 <= weight <= 10,000
     * @param truck_weights 트럭별 무게 배열
     * - 1 <= truck_weights.length <= 10,000
     * - 1 <= truck_weights[i] <= weight
     *
     * @return 모든 트럭이 다리를 지나는 데 소요되는 최소 초
     */
    public int solution(int bridge_length, int weight, int[] truck_weights) {
        Queue<Integer> waitingTrucks = new ArrayDeque<>();
        for (int truckWeight : truck_weights) {
            waitingTrucks.offer(truckWeight);
        }

        Queue<TruckOnBridge> onBridge = new ArrayDeque<>();
        int time = 0;
        int currentWeight = 0;

        while (!waitingTrucks.isEmpty() || !onBridge.isEmpty()) {
            time++;

            // 루프 한 번은 1초입니다. 먼저 다리 끝에 도착한 맨 앞 트럭을 빼고,
            // 그 뒤에 새 트럭이 올라갈 수 있는지 판단해야 같은 초에 빠짐과 진입을 함께 처리할 수 있습니다.
            if (!onBridge.isEmpty() && onBridge.peek().exitTime() == time) {
                currentWeight -= onBridge.poll().weight();
            }

            // 다리 위 트럭은 들어간 순서대로만 나갑니다.
            // 그래서 모든 트럭의 이동 거리를 매번 올리지 않고, 들어올 때 나갈 시간을 확정해 두면
            // 매초 큐의 맨 앞 트럭만 확인해도 같은 시간 시뮬레이션을 표현할 수 있습니다.
            if (!waitingTrucks.isEmpty()
                && onBridge.size() < bridge_length
                && currentWeight + waitingTrucks.peek() <= weight) {
                int truckWeight = waitingTrucks.poll();
                currentWeight += truckWeight;
                onBridge.offer(new TruckOnBridge(truckWeight, time + bridge_length));
            }
        }

        return time;
    }

    private record TruckOnBridge(int weight, int exitTime) {
    }
}

public class Main260511 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        var testCases = List.of(
            new TestCase(2, 10, new int[]{7, 4, 5, 6}, 8),
            new TestCase(100, 100, new int[]{10}, 101),
            new TestCase(100, 100, new int[]{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}, 110),
            new TestCase(1, 1, new int[]{1}, 2),
            new TestCase(1, 100, new int[]{10, 20, 30}, 4),
            new TestCase(2, 10, new int[]{5, 5, 5}, 5),
            new TestCase(2, 10, new int[]{10, 10}, 5),
            new TestCase(3, 10, new int[]{10, 10}, 7),
            new TestCase(3, 100, new int[]{10, 20, 30}, 6),
            new TestCase(3, 10, new int[]{7, 7, 3, 2, 2, 5, 4}, 14)
        );

        for (var testCase : testCases) {
            var result = solution.solution(testCase.input(), testCase.input2(), testCase.input3());
            assertEquals(testCase, result);
        }
    }

    private static void assertEquals(TestCase testCase, int actual) {
        if (actual != testCase.answer()) {
            throw new AssertionError(
                "bridgeLength=" + testCase.input()
                    + ", weight=" + testCase.input2()
                    + ", truckWeights=" + Arrays.toString(testCase.input3())
                    + ", expected=" + testCase.answer()
                    + ", actual=" + actual
            );
        }
    }

    private record TestCase(int input, int input2, int[] input3, int answer) {
    }
}
