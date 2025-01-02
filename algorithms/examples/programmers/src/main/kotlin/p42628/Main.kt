package p42628

import java.util.PriorityQueue

/**
 * [이중우선순위큐](https://school.programmers.co.kr/learn/courses/30/lessons/42628?language=kotlin)
 */
class Solution {
    /**
     * 이중 우선순위 큐 연산:
     * - I 숫자: 큐에 주어진 숫자를 삽입.
     * - D 1: 큐에서 최댓값을 삭제.
     * - D -1: 큐에서 최솟값을 삭제.
     *   단, 빈 큐에 데이터를 삭제하라는 연산이 주어질 경우 해당 연산은 무시.
     *
     * 불확실한 점들
     * - 중복된 숫자 insert 가능한지?
     * - 숫자가 하나만 남는 경우 어떻게 할지?
     *
     * @param operations 이중 우선순위 큐가 할 연산 목록
     * - 1 <= operations.length <= 1,000,000
     * - operations.i: "명령어 데이터" 형식
     * @return 모든 연산을 처리한 후
     * - 큐가 비어있으면 [0,0]
     * - 비어있지 않으면 [최댓값, 최솟값]
     */
    fun solution(operations: Array<String>): IntArray {
        // ["I 16", "I -5643", "D -1", "D 1", "D 1", "I 123", "D -1"]
        // 1. [16]
        // 2. [16, -5643]
        // 3. [16]
        // 4. []
        // 5. []
        // 6. [123]
        // 7. []

        // 최소힙
        val minPQ = PriorityQueue<Int>()
        // 최대힙
        val maxPQ = PriorityQueue<Int>(compareByDescending { it })

        // 값이 유효한지 관리합니다. 중복이 있을 수 있으므로 개수를 셉니다.
        // <숫자, 개수>
        val numbers = mutableMapOf<Int, Int>()

        // if 명령어 == I
        //  최대힙.offer(Pair(idx, 숫자))
        //  최소힙.offer(Pair(idx, 숫자))
        // else if 명령어 == D
        //  if 숫자 == 1
        //      맵.removeAt(최대힙.poll().idx)
        //  if 숫자 == -1
        //      맵.removeAt(최소힙.poll().idx)
        operations.forEach { operation ->
            val (command, rawNumber) = operation.split(" ")
            val number = rawNumber.toInt()

            when (command) {
                "I" -> {
                    minPQ.offer(number)
                    maxPQ.offer(number)
                    numbers[number] = numbers.getOrDefault(number, 0) + 1
                }

                "D" -> {
                    // 최솟값을 삭제합니다.
                    var targetPQ = if (number == -1) {
                        minPQ
                    } else {
                        maxPQ
                    }

                    if (targetPQ.isEmpty()) {
                        return@forEach
                    }

                    // 최소힙, 최대힙 두 개의 큐를 사용하는 반면 `poll`은 한 번만 이뤄집니다.
                    // 따라서 각 큐에는 다른 큐에서 이미 제거된 숫자가 남아있을 수 있습니다.
                    // 큐에 있는 숫자가 `numbers`에 있다면 처리해야 하는 쌍이므로 최소값이든 최대값이든 삭제합니다.
                    // 큐에 있는 숫자가 `numbers`에 없다면 이미 처리된 값이므로 제거합니다.
                    while (targetPQ.isNotEmpty()) {
                        val currValue = targetPQ.poll()

                        if (numbers.contains(currValue)) {
                            numbers[currValue] = numbers[currValue]!! - 1
                            if (numbers[currValue] == 0) {
                                numbers.remove(currValue)
                            }
                            break
                        }
                    }
                }
            }
        }

        var minVal = 0
        while (minPQ.isNotEmpty()) {
            val minCandidate = minPQ.poll()
            if (numbers.contains(minCandidate)) {
                minVal = minCandidate
                break
            }
        }
        var maxVal = 0
        while (maxPQ.isNotEmpty()) {
            val maxCandidate = maxPQ.poll()
            if (numbers.contains(maxCandidate)) {
                maxVal = maxCandidate
                break
            }
        }
        // println("minPQ: $minPQ, maxPQ: $maxPQ, numbers: $numbers, minVal: $minVal, maxVal: $maxVal")

        return intArrayOf(maxVal, minVal)
    }
}

fun main() {
    val s = Solution()
    listOf(
        Pair(
            arrayOf("I 16", "I -5643", "D -1", "D 1", "D 1", "I 123", "D -1"),
            intArrayOf(0, 0)
        ),
        Pair(
            arrayOf("I -45", "I 653", "D 1", "I -642", "I 45", "I 97", "D 1", "D -1", "I 333"),
            intArrayOf(333, -45)
        ),
        Pair(
            arrayOf("I 3", "I 3", "I 3"),
            intArrayOf(3, 3)
        ),
    ).forEach { tc ->
        println(s.solution(tc.first).contentEquals(tc.second))
    }
}