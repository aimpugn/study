package codility.binary_gap

import kotlin.math.max

/**
 * [binary gap](https://app.codility.com/programmers/trainings/9/binary_gap/)
 *
 * 양수 N에서 binary gap은 1로 둘러쌓인 연속된 0의 최대
 *
 * - 9 == 1001
 *   => 길이 2인 binary gap
 * - 529 == 10 0001 0001
 *   => 길이 4 또는 길이 3인 binary gap
 * - 20 => 1 0100
 *   => 길이 1인 binary gap
 * - 15 => 1111
 *   => 없음
 * - 32 => 10 0000
 *   => 없음
 * - 1041 => 100 0001 0001
 *   => 길이 5인 binary gap
 * - 32 => 10 0000
 *   => 없음
 *
 * @param N 양의 정수
 * - 1 <= 2,147,483,647
 * @return 가장 긴 binary gap 리턴합니다.
 * - binary gap 없으면 0 리턴합니다.
 */
fun solution(N: Int): Int {
    // Implement your solution here

    // binary 나타내는 bit array 로 만듭니다.
    val bits = toBinaryArray(N)
    var maxBinaryGap = 0
    var currIdx = -1

    bits.forEachIndexed { idx, bit ->
        if (bit) {
            if (currIdx == -1) {
                currIdx = idx
            } else {
                maxBinaryGap = max(maxBinaryGap, idx - currIdx - 1)
                currIdx = idx
            }
        }
    }

    return maxBinaryGap
}

/**
 * 정수를 boolean 배열로 변환하여 리턴합니다.
 * 1 경우 true, 0 경우 false 입니다.
 *
 * 0번 인덱스부터 첫번째자리를 나타냅니다.
 * 즉, [least significant bit, ... ,most significant bit] 형태로 리턴합니다.
 *
 * 예를 들어,
 * - 9 경우 [true, false, false, true] 입니다.
 * - 32 경우 10 0000 => [false, false, false, false, false, true] 입니다.
 *
 */
fun toBinaryArray(number: Int): BooleanArray {
    var target = number

    var length = 0
    while (target > 0) {
        length++
        target = target shr 1
    }

    val bits = BooleanArray(length)
    target = number

    // - 9 shl 1 경우 1비트 왼쪽으로 쉬프트하므로, 1001 => 10010 = 18
    // - 9 shr 1 경우 1비트 오른쪽으로 쉬프트하므로, 1001 => 100 = 4
    // - 9 ushr 1 경우 부호 없이 1비트 오른쪽으로 쉬프트하므로, 1001 => 100 = 4
    var idx = 0
    val mask = 1
    while (target > 0) {
        bits[idx++] = (target and mask) == 1
        target = target shr 1
    }

    return bits
}

fun main() {
    listOf(
        Pair(1, 0),
        Pair(9, 2),
        Pair(529, 4),
        Pair(20, 1),
        Pair(15, 0),
        Pair(1041, 5),
    ).forEach { tc ->
        println(solution(tc.first) == tc.second)
    }
}
