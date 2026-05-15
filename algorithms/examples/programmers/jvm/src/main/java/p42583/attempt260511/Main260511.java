package p42583.attempt260511;

import java.util.ArrayDeque;
import java.util.Arrays;
import java.util.LinkedList;
import java.util.List;
import java.util.concurrent.ConcurrentLinkedQueue;

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
        // 이 메서드에서 끝까지 지키는 질문은 세 가지입니다.
        // 아직 기다리는 트럭이 누구인가, 다리 위에는 어떤 트럭이 언제까지 남는가,
        // 지금 다리 위 총무게에 다음 트럭을 더해도 안전한가.
        // 이 세 질문이 닫히면 모든 트럭이 다리를 지난 최소 시간이 자연스럽게 나옵니다.
        //
        // 대표 입력 [7, 4, 5, 6], bridge_length = 2, weight = 10을 보면 흐름은 이렇게 됩니다.
        // time 1: 7 진입, 7의 exitTime은 3
        // time 2: 7은 아직 나가지 않고, 4를 더하면 11이라 진입 못 함
        // time 3: 7 퇴장, 4 진입, 4의 exitTime은 5
        // time 4: 5 진입, 5의 exitTime은 6
        // time 5: 4 퇴장, 6은 5 + 6 > 10이라 아직 진입 못 함
        // time 6: 5 퇴장, 6 진입, 6의 exitTime은 8
        // time 8: 6 퇴장, 대기 큐와 다리 큐가 모두 비어 답은 8
        //
        // 이 trace에서 보이는 핵심은 모든 트럭을 매번 훑는 것이 아니라,
        // 현재 시간에 빠질 수 있는 맨 앞 트럭과 지금 들어올 수 있는 다음 트럭만 보면 된다는 점입니다.
        // 이 풀이의 출발점은 "트럭을 실제로 한 칸씩 움직인다"가 아니라
        // "시간이 흐를 때 어떤 사건이 확정되는가"를 먼저 보는 것입니다.
        // 다리를 건너는 트럭은 순서를 바꿀 수 없으므로, 대기 트럭도 FIFO 큐로 두면
        // 다음에 다리에 올라갈 수 있는 후보가 항상 큐의 맨 앞에 있다는 점이 코드에 드러납니다.
        var waitingTrucks = new ArrayDeque<Integer>();
        for (int truckWeight : truck_weights) {
            waitingTrucks.offer(truckWeight);
        }

        // 다리 위 상태도 FIFO 큐입니다.
        // 다리에서는 먼저 들어간 트럭이 먼저 나가므로, 끝에 도달했는지 확인해야 하는 트럭은
        // 매 순간 전체 트럭이 아니라 큐의 맨 앞 트럭 하나뿐입니다.
        var onBridge = new ArrayDeque<TruckOnBridge>();

        // time은 지금까지 흐른 초입니다.
        // currentWeight는 현재 다리 위에 실제로 올라가 있는 트럭들의 무게 합입니다.
        // weight 파라미터는 다리의 최대 허용 무게라는 원래 의미를 유지하고,
        // 남은 무게를 표현하기 위해 weight를 직접 빼고 더하지 않습니다.
        int time = 0;
        int currentWeight = 0;

        // 아직 올라가지 못한 트럭이 남아 있거나, 이미 다리 위에 올라간 트럭이 남아 있으면
        // 시뮬레이션은 끝나지 않았습니다. 모든 트럭이 대기 큐에서도 빠지고 다리 큐에서도 빠져야 답이 됩니다.
        while (!waitingTrucks.isEmpty() || !onBridge.isEmpty()) {
            // 루프 한 번은 "1초가 흐른 뒤의 사건"을 처리합니다.
            // 이 순서를 먼저 고정해야 off-by-one, 즉 1초 차이로 답이 어긋나는 실수를 줄일 수 있습니다.
            time++;

            // 루프 한 번은 1초입니다. 먼저 다리 끝에 도착한 맨 앞 트럭을 빼고,
            // 그 뒤에 새 트럭이 올라갈 수 있는지 판단해야 같은 초에 빠짐과 진입을 함께 처리할 수 있습니다.
            // 원래 풀이의 moved++ 방식으로 말하면, 맨 앞 트럭이 bridge_length만큼 이동을 마친 순간입니다.
            // 개선된 풀이는 그 값을 매번 증가시키는 대신, 진입할 때 계산해 둔 exitTime과 현재 시간을 비교합니다.
            if (!onBridge.isEmpty() && onBridge.peek().exitTime() == time) {
                // poll()이 "다리 위에서 제거한다"는 동작입니다.
                // 제거와 동시에 currentWeight에서도 빼야, 바로 아래 진입 판단에서 빈 무게가 반영됩니다.
                currentWeight -= onBridge.poll().weight();
            }

            // 다리 위 트럭은 들어간 순서대로만 나갑니다.
            // 그래서 모든 트럭의 이동 거리를 매번 올리지 않고, 들어올 때 나갈 시간을 확정해 두면
            // 매초 큐의 맨 앞 트럭만 확인해도 같은 시간 시뮬레이션을 표현할 수 있습니다.
            // 새 트럭은 세 조건을 모두 만족할 때만 올라갑니다.
            // 대기 트럭이 있어야 하고, 다리 위 대수 제한을 넘지 않아야 하며,
            // 새 트럭을 더한 무게가 다리의 최대 허용 무게를 넘지 않아야 합니다.
            if (
                !waitingTrucks.isEmpty()
                    && onBridge.size() < bridge_length
                    && currentWeight + waitingTrucks.peek() <= weight
            ) {
                // 여기서 poll()되는 트럭은 "이번 초에 다리에 진입하는 트럭"입니다.
                // 진입하는 순간 currentWeight에 더해 두어야, 다음 트럭을 같은 초에 또 올리는 실수를 막습니다.
                int truckWeight = waitingTrucks.poll();
                currentWeight += truckWeight;

                // 트럭이 time에 다리에 올라가면 bridge_length초 뒤에 빠집니다.
                // 그래서 moved를 매초 1씩 올리지 않아도 exitTime = time + bridge_length 하나로
                // "언제 다리 끝에 도달하는가"를 같은 의미로 표현할 수 있습니다.
                onBridge.offer(new TruckOnBridge(truckWeight, time + bridge_length));
            }
        }

        // while이 끝났다는 것은 대기 큐도 비었고 다리 위 큐도 비었다는 뜻입니다.
        // 따라서 지금의 time이 "모든 트럭이 다리를 완전히 지난 시각"입니다.
        return time;
    }

    // 다리 위 트럭이 현재 필요한 정보는 두 가지뿐입니다.
    // weight는 다리의 현재 총무게를 유지하기 위해 필요하고,
    // exitTime은 이 트럭이 언제 다리에서 빠지는지 확인하기 위해 필요합니다.
    // moved처럼 매초 바뀌는 값을 들고 있지 않기 때문에 상태가 더 작고, 큐의 맨 앞만 보면 됩니다.
    private record TruckOnBridge(int weight, int exitTime) {
    }

    // 아래 old 버전은 "각 트럭의 이동 거리(moved)를 직접 갱신한다"는 물리적인 사고를 보존한 비교용입니다.
    // 개선된 solution은 이 사고를 버린 것이 아니라, moved가 결국 "언제 나가나"를 알아내기 위한 값이라는 점을 이용합니다.
    // 그래서 매초 모든 트럭의 moved를 올리는 대신, 진입 순간 exitTime을 한 번 계산하고 FIFO 맨 앞만 확인합니다.
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
    public int solution_old(int bridge_length, int weight, int[] truck_weights) {
        // bridge_length 대만큼 올라갈 수 있습니다.
        // 그리고 트럭은 1초에 길이 1만큼 이동합니다.
        //
        // 트럭 큐가 있고, 이게 다리 위에 있고, 루프 하나당 한 칸씩 이동시켜 봅니다.
        // 그리고 그 과정 속에서 다리 위에 올라갈 수 있는 트럭 대수, 무게 등을 고려합니다.
        // 그러면 한 대씩 빼서 이동시키고, 루프마다 +1을 시키고, 가능한 길이와 무게를 확인합니다.
        var q = new LinkedList<TruckInfo>();
        for (var truck_weight : truck_weights) {
            q.add(new TruckInfo(truck_weight, 0));
        }

        // 다리 위에 있는 트럭들을 담을 리스트입니다.
        // 첫번째 트럭을 다리 위에 올리고 1초를 셉니다.
        var sec = 0;
        var firstTruck = q.poll();
        var onBridge = new ConcurrentLinkedQueue<TruckInfo>();
        onBridge.add(firstTruck);
        firstTruck.moved++;
        weight -= firstTruck.truck_weight;
        sec++;

        while (!onBridge.isEmpty()) {
            // 루프 한 번은 1초라고 가정합니다.
            // 그러면 이 루프 안에서 1초 처리해야 하는 일들을 정의합니다.
            //
            // 꺼낸 트럭은 이 루프 안에 있고,
            // 끝에 도달하면 다리를 건넌 트럭이 되어야 하고,
            // 무게에서 제외되어야 합니다.

            // 기존 트럭이 있다면 하나씩 이동시킵니다.
            for (var truck : onBridge) {
                truck.moved++;
                // 이동 거리가 다리 길이를 넘으면 다리를 건넌 것이므로
                // bridge length +1 하고 weight에도 추가합니다.
                if (bridge_length < truck.moved) {
                    weight += truck.truck_weight;

                    // 해당 트럭을 제거합니다.
                    onBridge.remove(truck);
                }
            }

            // 다리 위에 더 올라갈 수 있다면?
            if (onBridge.size() < bridge_length && (!q.isEmpty() && 0 <= weight - q.peek().truck_weight)) {
                var currTruck = q.poll();
                // 현재 트럭이 올라갈 수 있습니다.
                weight -= currTruck.truck_weight;
                // 그리고 현재 트럭이 다리 위에 올라갔으므로 다리 위에 있음을 나타냅니다.
                currTruck.moved++;
                onBridge.add(currTruck);
            }

            sec++;
        }

        return sec;
    }

    class TruckInfo {
        int truck_weight;
        int moved;

        TruckInfo(int truck_weight, int moved) {
            this.truck_weight = truck_weight;
            this.moved = moved;
        }

        @Override
        public String toString() {
            return "TruckInfo{" +
                "truck_weight=" + truck_weight +
                ", moved=" + moved +
                '}';
        }
    }
}

public class Main260511 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        // 앞의 세 케이스는 Programmers 공식 예제입니다.
        // 그 뒤 케이스들은 발상 검증용입니다. 다리 길이가 1인 경우, 무게 때문에 막히는 경우,
        // 여러 트럭이 연속으로 올라갈 수 있는 경우를 함께 두어 exitTime 방식의 시간 감각을 확인합니다.
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
        // 여기까지 예외 없이 도달하면, 준비한 모든 시간 흐름 케이스가 통과한 것입니다.
        // 일부러 출력하지 않는 이유는 성공 로그보다 실패 시점의 expected/actual 정보가 더 중요하기 때문입니다.
    }

    private static void assertEquals(TestCase testCase, int actual) {
        if (actual != testCase.answer()) {
            // 출력만 하고 넘어가면 실패를 눈으로 놓칠 수 있습니다.
            // 학습용 검증도 실패하면 즉시 멈추게 만들어야 어느 입력에서 어떤 시간 계산이 깨졌는지 바로 볼 수 있습니다.
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
