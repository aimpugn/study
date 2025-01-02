package codility.demo

/**
 * "Find the smallest positive integer that does not occur in a given sequence."
 *
 * given an array A of N integers, returns the smallest positive integer (greater than 0) that does not occur in A.
 *
 * @param A N 개의 정수를 담은 배열
 * - N is an integer within the range [1..100,000];
 *    - 1 <= A.size <= 100,000
 * - each element of array A is an integer within the range [−1,000,000..1,000,000].
 *    - -1,000,000 <= A.i <= 1,000,000
 * @return A에 없고, 0보다 큰, 가장 작은 양의 정수
 */
fun solution(A: IntArray): Int {
    // Implement your solution here

    // [1, 3, 6, 4, 1, 2] => 5
    // [1, 2, 3] => 4
    // [-1, -3] => 1

    // set 으로 만들고 1부터 체크? 선형적으로 시간 소요
    val dedup = A.toSet()
    // 배열의 크기는 10만, 요소 값은 -1,000,000 ~ 1,000,000
    // 0보다 커야 하므로 음수는 확인하지 않아도 됩니다.
    for(i in 1 .. 1_000_000) {
        if(!dedup.contains(i)) {
            return i
        }
    }

    return 1_000_000
}

fun main() {
    listOf(
        Pair(intArrayOf(1, 3, 6, 4, 1, 2), 5),
        Pair(intArrayOf(1, 2, 3), 4),
        Pair(intArrayOf(-1, -3), 1),
    ).forEach { tc ->
        println(solution(tc.first) == tc.second)
    }
}