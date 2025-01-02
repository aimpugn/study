package p42883

import java.util.LinkedList
import kotlin.math.*

// [큰 수 만들기](https://school.programmers.co.kr/learn/courses/30/lessons/42883)
class Solution {
    /**
     * k개의 수를 제거했을 때 얻을 수 있는 가장 큰 수 구하기
     *
     * Example:
     * 1. 1924, 2개 제거하면?
     *      [19, 12, 14, 92, 94, 24]
     *      가장 큰 수는 94
     * 2. "1231234", 3
     *      [1231, 1232, 1233, 1234, ... 3234]
     *
     * @param number 숫자
     *  - 2 <= number.length <= 1,000,000
     * @param k 제거할 수의 개수
     *  - 1 <= k < number.length 자연수
     * @return number에서 k개의 수를 제거했을 때 만들 수 있는 수 중 가장 큰 수
     */
    fun solution(number: String, k: Int): String {
        // 어떻게 풀 것인가?
        // - 조합의 문제일까? 재귀적으로?
        //      숫자의 길이는 1,000,000까지 가능합니다.
        //      만약 재귀적으로 호출한다면 n개의 수에서 n번 반복을 하므로 O(n^2) 가 되어, 굉장히 느려집니다.
        // - 예제를 보면 순서 유지
        //      순서가 유지되므로, 가장 앞에서부터 앞/뒤로 비교해서 큰 수만 남기면 될 거 같습니다.
        // - 처음에 문자열 그 자체를 replace 했지만, 매번 새로운 문자열이 생성돼서 타임아웃이 발생합니다.
        //      StringBuilder를 사용하여 우선 버퍼 내에서 제거하도록 합니다.
        val builder = StringBuilder()
        var removeCount = k

        for (digit in number) {
            // 마지막 문자와 현재 문자를 비교하여 제거 여부 결정합니다.
            // - 4177252841, 4
            //      [4]: removeCount=4
            //      [4, 1]: 4 > 1이므로 while 진입하지 않고 append. removeCount=4
            //      [7]: 1 < 7이므로 1 제거, 4 < 7이므로 4 제거. removeCount=2
            //      [7, 7]:
            //      [7, 7, 2]: 7 > 2이므로 while 진입하지 않고 append. removeCount=2
            //      [7, 7, 5]: 2 < 5이므로 2 제거. removeCount=1
            //      [7, 7, 5, 2]: 5 > 2이므로 while 진입하지 않고 append. removeCount=1
            //      [7, 7, 5, 8]: 2 < 8이므로 2 제거. removeCount=0
            //      이후 removeCount=0이므로 계속 append
            //      [7, 7, 5, 8, 4, 1]
            while (removeCount > 0 && builder.isNotEmpty() && builder.last() < digit) {
                builder.deleteCharAt(builder.length - 1)
                removeCount--
            }
            builder.append(digit)
        }

        // 아직 제거해야 할 숫자가 남아있다면 뒤에서부터 제거합니다.
        // 앞서 로직에서 큰 수는 작은 수를 지워버립니다.
        // 따라서 removeCount가 남았다는 것은, 같은 수 또는 작은 수라는 의미입니다.
        // - 44443, 1
        // - 55555, 1
        if (removeCount > 0) {
            return builder.substring(0, builder.length - removeCount)
        }

        return builder.toString()
    }

    fun solution2(number: String, k: Int): String {
        val stack = LinkedList<Char>()
        var remainK = k

        for (digit in number) {
            while (remainK > 0 && stack.isNotEmpty() && stack.last() < digit) {
                stack.removeLast()
                remainK--
            }
            stack.add(digit)
        }

        // 아직 제거해야 할 숫자가 남아 있으면 뒤에서부터 제거
        while (remainK > 0) {
            stack.removeAt(stack.size - 1)
            remainK--
        }

        return stack.joinToString("")
    }
}

data class TestCase(val arg1: String, val arg2: Int)

fun main() {
    val testCases = listOf(
        TestCase(
            "1924",
            2
        ),
        TestCase(
            "1231234",
            3
        ),
        TestCase(
            "4177252841",
            4
        ),
        TestCase(
            "444456",
            1
        ),
        TestCase(
            "55555",
            1
        ),
        TestCase(
            "44443",
            1
        ),
    )

    for (testCase in testCases) {
        println(Solution().solution(testCase.arg1, testCase.arg2))
        println(Solution().solution2(testCase.arg1, testCase.arg2))
    }

}