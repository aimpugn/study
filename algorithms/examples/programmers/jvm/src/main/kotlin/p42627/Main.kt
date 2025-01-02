package p42627

import java.util.PriorityQueue

/**
 * [디스크 컨트롤러](https://school.programmers.co.kr/learn/courses/30/lessons/42627)
 */
class Solution {
    /**
     * - 하드디스크는 한 번에 하나의 작업만 수행
     * - 우선순위 디스크 컨트롤러
     *    1. 작업의 번호, 작업의 요청 시각, 작업의 소요 시간을 저장해두는 대기 큐
     *    2. 하드디스크 작업 X && 대기 큐 있음 => 우선순위가 높은 작업을 꺼내서 하드디스크에 작업 시킴.
     *       우선순위:
     *       1. 작업 소요 시간이 짧은 것
     *       2. 작업 요청 시각이 빠른 것
     *       3. 작업 번호가 작은 것
     *    3. 한 번 작업 시작하면 작업 마칠 때까지 그 작업만 수행.
     *    4. '하드디스크가 어떤 작업을 마치는 시점'과 '다른 작업 요청이 들어오는 시점'이 겹치는 경우
     *       1. 하드디스크가 작업을 마치자마자 디스크 컨트롤러는 요청이 들어온 작업을 대기 큐에 저장하고
     *       2. 우선순위가 높은 작업을 대기 큐에서 꺼내서 하드디스크에 작업 시킴
     *
     *       마치는 시점에 다른 작업이 들어오지 않더라도, 그 작업을 마치자마자 또 다른 작업을 시작할 수 있음.
     *       이 과정에서 걸리는 시간은 없다고 가정.
     *
     * @param jobs 각 작업의 (작업 요청 시점, 작업 소요 시간)을 담은 2차원 배열
     * - 1 <= jobs.length <= 500
     * - jobs.i: i번 작업에 대한 정보.
     * - 0 <= 작업 요청 시점 <= 1,000
     * - 0 <= 작업 소요 시간 <= 1,000
     * @return 모든 요청 작업의 반환 시간의 평균의 정수 부분
     */
    fun solution(jobs: Array<IntArray>): Int {
        // [[0, 3], [1, 9], [3, 5]]
        //
        // 1. 0ms: 큐에 저장
        //    - 대기큐: [[0번, 0ms 요청, 3ms 소요]]
        //    - 대기
        // 2. 0ms: 큐에서 0번 작업 꺼내서 시작
        //    - 대기큐: []
        //    - 작업시작: 0번
        // 3. 1ms: 1번 작업을 큐에 저장
        //    - 대기큐: [[1번, 1ms 요청, 9ms 소요]]
        //    - 작업중: 0번
        // 4. 2~3ms: 0번 작업 완료
        //    - 대기큐: [[1번, 1ms 요청, 9ms 소요]]
        //    - 작업완료: 0번
        // 5. 3ms: 2번 작업을 큐에 저장
        //    - 대기큐: [[1번, 1ms 요청, 9ms 소요], [2번, 3ms 요청, 5ms 소요]]
        //    - 대기
        // 6. 3ms: 큐에서 2번 작업 꺼내서 시작(작업 소요 시간이 짧음)
        //    - 대기큐: [[1번, 1ms 요청, 9ms 소요]]
        //    - 작업시작: 2번
        // 7. 3ms~8ms: 2번 작업 완료
        //    - 대기큐: [[1번, 1ms 요청, 9ms 소요]]
        //    - 작업시작: 2번
        // 8. 8ms: 큐에서 1번 작업을 꺼내서 시작
        //    - 대기큐: []
        //    - 작업시작: 1번
        // 9. 8ms~17ms: 1번 작업 완료
        //    - 대기큐: []
        //    - 대기
        // 모든 요청 작업을 마쳤을 때, 각 작업에 대한 반환 시간은 '작업 요청 ~ 종료 걸린 시간'
        //
        // - 0번 작업: 0ms ~ 3ms => 3 - 0 = 3
        // - 1번 작업: 1ms ~ 17ms => 17 - 1 = 16
        // - 2번 작업: 3ms ~ 8ms => 8 - 3 = 5
        // => (3 + 16 + 5) / 3 = 8

        // 요청 타이밍에 작업이 큐에 들어가야 합니다.
        // 우선 요청 시간순서대로 큐에 쌓습니다.
        // [[0,3], [1, 9], [3, 5]]
        val timelineQueue = PriorityQueue<Task>(
            compareBy<Task> { it.requestedTime }
                // NOTE: requestedTime 순으로만 정렬하면 실패합니다.
                .thenBy { it.elapsedTime } // 요청 시간이 같아도 소요되는 시간이 적으면 그 순서로
                .thenBy { it.idx } // 요청 시간, 소요 시간이 같아도 인덱스가 더 빠른 것으로
        ).apply {
            jobs.forEachIndexed { idx, (requestedTime, elapsedTime) ->
                add(Task(idx, requestedTime, elapsedTime))
            }
        }

        // 작업 우선순위:
        // 1. 작업 소요 시간 짧은 것
        // 2. 작업 요청 시간이 빠른 것
        // 3. 작업 번호가 작은 것
        val waitingQueue = PriorityQueue<Task>(
            compareBy<Task> { it.elapsedTime }
                .thenBy { it.requestedTime }
                .thenBy { it.idx }
        )

        // // 요청이 2ms 시점에 들어온다면? 그때부터 시작해야 하므로,
        // // 현재 시각을 처음 작업의 요청 시간으로 설정합니다.
        // 최초작업 = 타임라인큐.poll()
        // 대기큐.add(최초작업)
        // 현재시각 = 최초작업.requestedTime
        //
        // while(대기큐)
        //
        //  현재작업 = 타임라인큐.poll
        //  // 0 + 3 = 3
        //  // 3 + 5 = 8
        //  // 8 + 9 = 17
        //  작업완료시각 = 현재시각 + 현재작업.elapsedTime
        //
        //  // 작업완료시각 전에 요청들어온 게 있다면
        //  while(타임라인큐.peek().requestedTime < 작업완료시각)
        //      대기큐.add(타임라인큐.poll())
        //  // 그리고?
        //  if 대기큐 비어있음
        //      대기큐.add(타임라인큐.poll())
        //
        //  //     0  1   2
        //  // 1. [3, 0,  0]
        //  // 2. [3, 0,  5]
        //  // 3. [3, 16, 5]
        //  작업완료시각배열[현재작업.idx] = 작업완료시각 - 현재작업.requestedTime
        //  // 3 => 8 => 17 => 끝
        //  현재시각 = 작업완료시각 // 현재 작업의 작업 완료 시점에 도달합니다.

        val firstTask = timelineQueue.poll()
        waitingQueue.add(firstTask)
        var currTime = firstTask.requestedTime
        val taskCompleteAtArray = IntArray(jobs.size) { 0 }

        while (waitingQueue.isNotEmpty()) {
            val currTask = waitingQueue.poll()

            // 현재 시각보다 현재 작업 시각의 요청 시각이 더 뒤라면, 해당 시각으로 변경합니다.
            // 0   1   2   3   4   5   6   7   8
            // |---|---|---|---|---|---|---|---|
            // =====0번====>        =====1번====>
            // 이런 경우 중간에 작업이 없으므로 현재 시각을 5로 바꿉니다.
            if (currTime < currTask.requestedTime) {
                currTime = currTask.requestedTime
            }
            val currTaskCompleteAt = currTime + currTask.elapsedTime

            // 작업 완료 전에 요청이 들어온 게 있다면 대기큐에 추가합니다.
            while (
                timelineQueue.isNotEmpty()
                && timelineQueue.peek().requestedTime <= currTaskCompleteAt
            ) {
                waitingQueue.add(timelineQueue.poll())
            }
            // 대기큐가 비어있고 남은 작업이 있다면 추가합니다.
            if (waitingQueue.isEmpty() && timelineQueue.isNotEmpty()) {
                waitingQueue.add(timelineQueue.poll())
            }

            taskCompleteAtArray[currTask.idx] = currTaskCompleteAt - currTask.requestedTime
            currTime = currTaskCompleteAt
            // println("\ttaskCompleteAtArray: ${taskCompleteAtArray.contentToString()}")
        }

        return taskCompleteAtArray.average().toInt()
    }

    data class Task(val idx: Int, val requestedTime: Int, val elapsedTime: Int)
}

data class TestCase(val jobs: Array<IntArray>, val answer: Int)

fun main() {
    val s = Solution()

    listOf(
        TestCase(
            arrayOf(
                intArrayOf(0, 3),
            ),
            3
        ),
        TestCase(
            arrayOf(
                intArrayOf(0, 3),
                intArrayOf(1, 9),
                intArrayOf(3, 5),
            ),
            8
        ),
        TestCase(
            arrayOf(
                intArrayOf(0, 3), // 1번: 3-0
                intArrayOf(3, 3), // 0번: 6-3
                intArrayOf(6, 3), // 2번: 9-6
            ), // [3, 3, 3]
            3
        ),
        // 시작 시간이 0 아닌 경우
        TestCase(
            arrayOf(
                intArrayOf(2, 3), // 1번: 13-2
                intArrayOf(1, 9), // 0번: 10-1
                intArrayOf(3, 5), // 2번: 18-3
            ), // [11, 9, 15] 평균
            11
        ),
        // 중간에 빈 시간이 있는 경우
        TestCase(
            arrayOf(
                intArrayOf(0, 3), // 0번: 3-0
                intArrayOf(5, 3), // 1번: 8-5
                intArrayOf(6, 3), // 2번: 11-6
            ), // [3, 3, 5] 평균
            3
        ),
        // 작업이 마치는 때 바로 요청이 들어오는 경우
        TestCase(
            arrayOf(
                intArrayOf(0, 3), // 0번: 3-0
                intArrayOf(2, 3), // 1번: 6-2
                intArrayOf(3, 3), // 2번: 9-3
            ), // [3, 4, 6] 평균
            4
        ),
        // 모든 작업이 동시에 들어온 경우
        TestCase(
            arrayOf(
                intArrayOf(0, 5), // 2번: 10-0
                intArrayOf(0, 2), // 0번: 2-0
                intArrayOf(0, 3), // 1번: 5-0
            ), // [10, 2, 5] 평균
            5
        ),
    ).forEach { tc ->
        println(s.solution(tc.jobs) == tc.answer)
    }
}