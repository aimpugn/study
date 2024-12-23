package p42583

import java.util.LinkedList

/**
 * [다리를 지나는 트럭](https://programmers.co.kr/learn/courses/30/lessons/42583)
 */
class Solution {
    /**
     * - 트럭 여러 대, 강을 가로지르는 일차선 다리, 정해진 순으로 건너기
     * - 다리에 완전히 오르지 않은 트럭의 무게는 무시
     *
     * 실행 시간
     * ```
     * 테스트 1 〉	통과 (1.12ms, 62.5MB)
     * 테스트 2 〉	통과 (7.03ms, 64.1MB)
     * 테스트 3 〉	통과 (0.07ms, 63MB)
     * 테스트 4 〉	통과 (6.23ms, 65MB)
     * 테스트 5 〉	통과 (45.67ms, 70MB)
     * 테스트 6 〉	통과 (13.04ms, 63.9MB)
     * 테스트 7 〉	통과 (1.46ms, 63.1MB)
     * 테스트 8 〉	통과 (0.71ms, 62.3MB)
     * 테스트 9 〉	통과 (4.09ms, 64MB)
     * 테스트 10 〉	통과 (0.60ms, 62.4MB)
     * 테스트 11 〉	통과 (0.18ms, 61.9MB)
     * 테스트 12 〉	통과 (0.77ms, 63.3MB)
     * 테스트 13 〉	통과 (1.94ms, 62.5MB)
     * 테스트 14 〉	통과 (0.07ms, 62.4MB)
     * ```
     *
     * @param bridge_length 다리에 올라갈 수 있는 최대 트럭 대수
     * - 1 <= bridge_length <= 10,000
     * @param weight 다리가 견딜 수 있는 무게
     * - 1 <= weight <= 10,000
     * @param truck_weights 트럭 별 무게
     * - 1 <= truck_weights.length <= 10,000
     * - 1 <= truck_weights.i <= weight
     * @return 모든 트럭이 다리를 건너려면 최소 몇 초가 걸리는지
     */
    fun solution(bridge_length: Int, weight: Int, truck_weights: IntArray): Int {
        var answer = 0
        // bridge_length: 2, weight: 10, truck_weights: [7, 4, 5, 6]
        // 1. 시작
        //   대기 트럭: [7, 4, 5, 6]
        //   다리를 건너는 트럭: []
        //   다리를 지난 트럭: []
        //
        // 2. 1초
        //               [7]  [4, 5, 6]
        //          #===#===#:10
        //   다리를 건너는 트럭: [7]
        //   다리를 지난 트럭: []
        //
        // 3. 2초
        //           [7]      [4, 5, 6]
        //          #===#===#:10
        //   다리를 건너는 트럭: [7]
        //   다리를 지난 트럭: []
        //
        // 4. 3초
        //       [7]     [4]  [5, 6]
        //          #===#===#:10
        //   다리를 건너는 트럭: [4]
        //   다리를 지난 트럭: [7]
        //
        // 5. 4초
        //       [7] [4 , 5]  [6]
        //          #===#===#:10
        //   다리를 건너는 트럭: [4, 5]
        //   다리를 지난 트럭: [7]
        //
        // 6. 5초
        //     [7,4] [5]      [6]
        //          #===#===#:10
        //   다리를 건너는 트럭: [5]
        //   다리를 지난 트럭: [7, 4]
        //
        // 7. 6초
        //   [7,4,5]     [6]  []
        //          #===#===#:10
        //   다리를 건너는 트럭: [6]
        //   다리를 지난 트럭: [7, 4, 5]
        //
        // 8. 7초
        //  [7,4,5]  [6]      []
        //          #===#===#:10
        //   다리를 건너는 트럭: [6]
        //   다리를 지난 트럭: [7, 4, 5]
        //
        // 9. 8초
        // [7,4,5,6]          []
        //          #===#===#:10
        //   다리를 건너는 트럭: []
        //   다리를 지난 트럭: [7, 4, 5, 6]
        //
        // 길이 1당 1초가 소요됩니다.
        // 이렇게 하면 총 8초가 소요됩니다.
        //
        // bridge_length: 100, weight: 100, truck_weights: [10]
        // 1. 시작
        //   대기 트럭: [10]
        //   다리를 건너는 트럭: []
        //   다리를 지난 트럭: []
        //
        // 2. 1초
        //   대기 트럭: []
        //   다리를 건너는 트럭: [10]
        //   다리를 지난 트럭: []
        //
        // 길이 1당 1초가 소요됩니다.
        // 트럭 한 대가 올라가고, 길이 100을 건넙니다.
        //
        // 즉, 1대가 온전히 다리 길이(bridge_length)를 건더는 데는 bridge_length + 1초가 소요됩니다.
        // 이때, 다리 길이(bridge_length)와 다리 허용 무게(weight)가 허용하는 한도에서 트럭이 올라갈 수 있습니다.
        //
        // bridge_length: 100, weight: 100, truck_weights: [10, 10, 10, 10, 10, 10, 10, 10, 10, 10](길이: 10) 경우
        // 10 대의 트럭이 한번에 다리에 모두 올라갈 수 있고,
        // 마지막 트럭 기준으로 다리를 지나는 데에는 bridge_length + 10초가 소요됩니다.
        // [10, 10, 10, 10, 10, 10, 10, 10, 10, 10]
        // #===#===#===#===#===#===#===#===#===#===#===#===#== ...
        // ^-끝
        // 어쨌든 다리 끝까지 오기 위해 bridge_length 초 소요
        // 그 후 한 대씩 넘어갈 때마다 1초씩 소요해서 +10 초
        //
        // 패턴 찾아보기:
        // [7, 4, 5, 6], weight: 10, bridge_length: 2
        //  init: [] [] [7, 4, 5, 6]
        //  1. [] [7] [4, 5, 6]
        //  2. [] [7,null] [4, 5, 6]
        //  3. [7] [null,4] [5, 6]
        //  4. [7] [4,5] [6]
        //  5. [7,4] [5,null] [6]
        //  6. [7,4,5] [null,6] []
        //  7. [7,4,5] [6,null] []
        //  8. [7,4,5,6] [] []
        //
        // [5, 5, 5], weight: 10, bridge_length: 2
        //  init: [] [] [5,5,5]
        //  1. [] [5] [5,5]
        //  2. [] [5,5] [5]
        //  3. [5] [5,5] []
        //  4. [5,5] [5,null] []
        //  5. [5,5,5] [null,null] []
        //
        // [5, 5, 5, 5, 5], weight: 15, bridge_length: 3
        //  init: [] [] [5,5,5,5,5]
        //  1. [] [5] [5,5,5,5]
        //  2. [] [5,5] [5,5,5]
        //  3. [] [5,5,5] [5,5]
        //  4. [5] [5,5,5] [5]
        //  5. [5,5] [5,5,5] []
        //  6. [5,5,5] [5,5,null] []
        //  7. [5,5,5,5] [5,null,null] []
        //  8. [5,5,5,5,5] [null,null,null] [] => 8초
        //
        // [7, 7, 3, 2, 2, 5, 4], weight: 10, bridge_length: 3
        //  init: [] [] [7,7,3,2,2,5,4]
        //  1. [] [7] [7,3,2,2,5,4]
        //  2. [] [7,null] [7,3,2,2,5,4]
        //  3. [] [7,null,null] [7,3,2,2,5,4]
        //  4. [7] [null,null,7] [3,2,2,5,4]
        //  5. [7] [null,7,3] [2,2,5,4]
        //  6. [7] [7,3,null] [2,2,5,4]
        //  7. [7,7] [3,null,2] [2,5,4]
        //  8. [7,7,3] [null,2,2] [5,4]
        //  9. [7,7,3] [2,2,5] [4]
        // 10. [7,7,3,2] [2,5,null] [4]
        // 11. [7,7,3,2,2] [5,null,4] []
        // 12. [7,7,3,2,2,5] [null,4,null] []
        // 13. [7,7,3,2,2,5] [4,null,null] []
        // 14. [7,7,3,2,2,5,4] [null,null,null] []

        // 고민되는 점은, 결국 큐를 해소하면서 최소가 되는 시간을 구해야 하는데
        // 1. 큐를 어떻게 해소해 나갈 것인지
        // 2. 큐를 해소하면서 어느 타이밍에 answer 값을 증가시킬 것인지
        var totalWeight = 0
        val queue = LinkedList<Int>().apply {
            truck_weights.forEach { truckWeight ->
                totalWeight += truckWeight
                add(truckWeight)
            }
        }

        // 한번에 모두 도로 위에 올라갈 수 있는 경우
        if (truck_weights.size <= bridge_length && totalWeight <= weight) {
            // 다리 길이(bridge_length)만큼은 다리를 전부 지나는 데 필요한 시간이 되고,
            // 트럭이 다리 끝에서 하나씩 건너편으로 넘어가는 데 드는 시간을 합칩니다.
            return bridge_length + truck_weights.size
        }

        val ongoing = LinkedList<Int?>()
        val done = LinkedList<Int>()

        // while(done.size != truck_weights.size)
        //
        //  중간 큐 쌓기
        //
        //  중간 큐 빼고 done 큐 쌓기
        //
        var weightOnBridge = 0
        while (done.size != truck_weights.size) {
            // 중간 큐 쌓기
            // 현재 다리 위에 다음 트럭을 올리 수 있다면 진행중인 큐에 추가합니다.
            if (queue.isNotEmpty() && weightOnBridge + queue.peek() <= weight) {
                var currentTruckWeight = queue.poll()
                weightOnBridge += currentTruckWeight
                ongoing.offer(currentTruckWeight)
            } else {
                // 트럭을 더 올릴 수 없지만 공간이 남았다면 null을 추가합니다.
                // 큐에 더이상 처리할 트럭이 없는 경우, 남은 공간을 채워 둡니다.
                ongoing.offer(null)
            }

            // ongoing 큐에 쌓이는 과정 자체가 다리를 지나는 시간을 의미합니다.
            //
            // 하나씩 밀어내는 경우에는 2번과 3번이 구별되어 있습니다.
            //  2. [] [7,null] [4, 5, 6]
            //  3. [7] [null,4] [5, 6]
            // 하지만 ongoing 큐 [7,null].poll() 결과 7을 꺼내면 [null]이 되고,
            // done 큐에 추가하여 [7]이 됩니다.
            // 즉, 2번과 3번이 켭치는 부분이 있어서, answer++ 외에 마지막 트럭이 빠져나가는 것을 +1로 보정해야 합니다.
            answer++
            // println("$answer. queue: $queue, ongoing: $ongoing")

            if (ongoing.size < bridge_length) {
                continue
            }

            val headOnBridge = ongoing.poll()
            if (headOnBridge != null) {
                weightOnBridge -= headOnBridge
                done.offer(headOnBridge)
            }
            // println("\tongoing: $ongoing\n\tdone: $done")
        }

        return answer + 1
    }

    /**
     * 실행 시간
     * ```
     * 테스트 1 〉	통과 (22.88ms, 65MB)
     * 테스트 2 〉	통과 (124.10ms, 65.2MB)
     * 테스트 3 〉	통과 (14.15ms, 65MB)
     * 테스트 4 〉	통과 (57.94ms, 64.7MB)
     * 테스트 5 〉	통과 (575.85ms, 64.8MB)
     * 테스트 6 〉	통과 (156.27ms, 65MB)
     * 테스트 7 〉	통과 (26.48ms, 64.8MB)
     * 테스트 8 〉	통과 (16.07ms, 64.3MB)
     * 테스트 9 〉	통과 (18.72ms, 65.2MB)
     * 테스트 10 〉	통과 (16.19ms, 66.1MB)
     * 테스트 11 〉	통과 (20.65ms, 64.6MB)
     * 테스트 12 〉	통과 (17.41ms, 64.7MB)
     * 테스트 13 〉	통과 (17.15ms, 65.2MB)
     * 테스트 14 〉	통과 (19.51ms, 65.4MB)
     * ```
     *
     * 배열을 슬라이딩하는 부분으로 인해 시간이 더 오래 걸립니다.
     */
    fun solution2(bridge_length: Int, weight: Int, truck_weights: IntArray): Int {
        // 한번에 모두 도로 위에 올라갈 수 있는 경우
        if (truck_weights.size <= bridge_length && truck_weights.sum() <= weight) {
            // 다리 길이(bridge_length)만큼은 다리를 전부 지나는 데 필요한 시간이 되고,
            // 다리 끝에서 트럭이 하나씩 건너편으로 넘어가는 데 드는 시간을 합칩니다.
            return bridge_length + truck_weights.size
        }

        // 현재 다리의 트럭 상태를 관리합니다.
        // 트럭이 없는 경우 null 대신 0으로 표시합니다.
        // [7, 7, 3, 2, 2, 5, 4], bridge_length: 3, weight: 10
        //  init: [0, 0, 0]
        //  1. [0, 0, 7]
        //  2. [0, 7, 0]
        //  3. [7, 0, 0]
        //  4. [0, 0, 7]
        //  5. [0, 7, 3]
        //  6. [7, 3, 0]
        //  7. [3, 0, 2]
        //  8. [0, 2, 2]
        //  9. [2, 2, 5]
        // 10. [2, 5, 0]
        // 11. [5, 0, 4]
        // 12. [0, 4, 0]
        // 13. [4, 0, 0]
        // 14. [0, 0, 0]
        val bridge = IntArray(bridge_length) { 0 }
        var time = 0 // 경과 시간
        var currTotalWeight = 0 // 현재 다리 위의 트럭 무게 합
        val waitingTrucksQueue = LinkedList(truck_weights.toList()) // 대기 트럭 큐

        val lastIdx = bridge_length - 1
        // '대기 트럭이 있음 OR 다리가 비어있지 않음' 동안 계속 처리합니다
        while (waitingTrucksQueue.isNotEmpty() || currTotalWeight > 0) {
            time++

            // 1. 다리에서 트럭 제거
            // println("bridge: ${bridge.contentToString()}, currTotalWeight: $currTotalWeight")
            currTotalWeight -= bridge[0]
            for (i in 1 until bridge_length) { // 다리의 상태를 한 칸씩 앞으로 이동시킵니다.
                bridge[i - 1] = bridge[i]
            }
            bridge[lastIdx] = 0 // 다리 끝부분을 초기화

            // 2. 다리가 견딜 수 있다면 새 트럭 추가합니다.
            if (waitingTrucksQueue.isNotEmpty() && currTotalWeight + waitingTrucksQueue.peek() <= weight) {
                val truck = waitingTrucksQueue.poll()
                bridge[lastIdx] = truck // 다리 출입 지점에 트럭을 진입시킵니다.
                currTotalWeight += truck // 그리고 현재 다리 위 전체 트럭 무게에 추가합니다.
            }
        }

        return time
    }

    /**
     * 실행 시간
     * ```
     * 테스트 1 〉	통과 (16.61ms, 64.7MB)
     * 테스트 2 〉	통과 (25.41ms, 64.4MB)
     * 테스트 3 〉	통과 (20.22ms, 63.8MB)
     * 테스트 4 〉	통과 (22.11ms, 65.9MB)
     * 테스트 5 〉	통과 (39.34ms, 65.2MB)
     * 테스트 6 〉	통과 (27.98ms, 64.6MB)
     * 테스트 7 〉	통과 (15.01ms, 64.4MB)
     * 테스트 8 〉	통과 (15.12ms, 64.4MB)
     * 테스트 9 〉	통과 (19.86ms, 65.4MB)
     * 테스트 10 〉	통과 (16.28ms, 64MB)
     * 테스트 11 〉	통과 (14.33ms, 65.3MB)
     * 테스트 12 〉	통과 (14.51ms, 64.4MB)
     * 테스트 13 〉	통과 (16.13ms, 65MB)
     * 테스트 14 〉	통과 (15.54ms, 64.6MB)
     * ```
     */
    fun solution3(bridge_length: Int, weight: Int, truck_weights: IntArray): Int {
        // 한번에 모두 도로 위에 올라갈 수 있는 경우
        if (truck_weights.size <= bridge_length && truck_weights.sum() <= weight) {
            // 다리 길이(bridge_length)만큼은 다리를 전부 지나는 데 필요한 시간이 되고,
            // 다리 끝에서 트럭이 하나씩 건너편으로 넘어가는 데 드는 시간을 합칩니다.
            return bridge_length + truck_weights.size
        }

        val queue = LinkedList<Pair<Int, Int>>() // 트럭의 무게와 다리에서 나가는 시간을 저장
        var time = 0                            // 현재 시간
        var currentWeight = 0                   // 다리 위의 총 무게
        val waitingTrucks = LinkedList(truck_weights.toList()) // 대기 중인 트럭

        while (waitingTrucks.isNotEmpty() || queue.isNotEmpty()) {
            time++

            // 1. 다리에서 나갈 트럭을 처리 (현재 시간이 다리에서 나가는 시간과 같다면)
            if (queue.isNotEmpty() && queue.peek().second == time) {
                val (weight, _) = queue.poll()
                currentWeight -= weight
            }

            // 2. 새로운 트럭이 다리에 올라갈 수 있는지 확인
            if (waitingTrucks.isNotEmpty() && currentWeight + waitingTrucks.peek() <= weight) {
                val truck = waitingTrucks.poll()
                queue.offer(Pair(truck, time + bridge_length)) // 트럭 무게와 나가는 시간 추가
                currentWeight += truck
            }
        }

        return time
    }
}

data class TestCase(val bridgeLength: Int, val weight: Int, val truckWeights: IntArray, val expected: Int)

fun main() {
    val s = Solution()

    listOf(
        TestCase(2, 10, intArrayOf(7, 4, 5, 6), 8),
        TestCase(3, 10, intArrayOf(7, 7, 3, 2, 2, 5, 4), 14),
        TestCase(100, 100, intArrayOf(10), 101),
        TestCase(100, 100, intArrayOf(10, 10, 10, 10, 10, 10, 10, 10, 10, 10), 110),
    ).forEach { tc ->
        println("solution ${s.solution(tc.bridgeLength, tc.weight, tc.truckWeights) == tc.expected}")
        println("solution2 ${s.solution2(tc.bridgeLength, tc.weight, tc.truckWeights) == tc.expected}")
        println("solution3 ${s.solution3(tc.bridgeLength, tc.weight, tc.truckWeights) == tc.expected}")
    }
}