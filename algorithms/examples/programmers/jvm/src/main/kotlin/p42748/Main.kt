package p42748

/**
 * [K번째수](https://school.programmers.co.kr/learn/courses/30/lessons/42748)
 */
class Solution {
    /**
     * 배열의 i번째 숫자부터 j번째 숫자까지 자르고 정렬했을 때, k번째 있는 수.
     *
     * Example:
     * - array: [1, 5, 2, 6, 3, 7]
     *   - i: 2번째
     *   - j: 5번째
     *   - k: 3
     *
     *   1. array[i:j] = array[2:5] = [5, 2, 6, 3]
     *   2. 정렬하면 [2, 3, 5, 6]
     *   3. k(3)번째 수는 5
     *
     *
     * @param array 모든 요소가 포함된 배열
     * - 1 <= array.size <= 100
     * - 1 <= array.i <= 100
     * @param commands [i, j, k]를 원소로 갖는 2차원 배열
     * - 1 <= commands.size <= 50
     * - commands.i.size == 3
     *
     * @return commands의 모든 원소에 대해 연산을 적용한 결과를 배열에 담아서 리턴
     */
    fun solution(array: IntArray, commands: Array<IntArray>): IntArray {
        var answer = IntArray(commands.size)

        commands.forEachIndexed { idx, command ->
            val i = command[0]
            val j = command[1]
            val k = command[2]

            val subArray = subtract(array, i - 1, j)
            quickSort(subArray, 0, subArray.size - 1)
            answer[idx] = subArray[k - 1]
        }

        return answer
    }

    /**
     * @param from 시작 인덱스. [0, 1, 2, 3]에서 0, 1, 2, 3이 시작 인덱스가 될 수 있습니다.
     * @param to 도달할 인덱스. [0, 1, 2, 3]에서 0, 1, 2, 3이 끝 인덱스가 될 수 있습니다.
     */
    fun subtract(array: IntArray, from: Int, to: Int): IntArray {
        val subArray = IntArray(to - from)

        var idx = 0
        for (i in from..to - 1) {
            subArray[idx++] = array[i]
        }

        return subArray
    }

    /**
     * 퀵 정렬은 분할 정복 방식으로 배열을 정리합니다.
     * 최선의 경우 O(n log n), 최악의 경우 O(n^2)의 시간 복잡도를 갖습니다.
     *
     * 비어있지 않은, 두 개의 연속된 배열로 분할합니다.
     * 첫번째 부분 배열의 요소는 두번째 부분 배열의 요소보다 크지 않아야 합니다.
     * 파티션 후 재귀적으로 부분 배열을 정렬하는데, 가능하면 분할 시점에, 이미 정렬된 요소를 제외한 후 정렬합니다.
     *
     * 재귀적 특성으로 인해 퀵 정렬은 더 큰 배열 내의 범위에 대해 호출 가능하도록 공식화되어야 합니다.
     *
     * Reference:
     * - [Quick Sort](https://en.wikipedia.org/wiki/Quicksort)
     */
    fun quickSort(array: IntArray, low: Int, high: Int) {
        // 1. 요소가 2개 미만인 경우 즉시 리턴합니다.
        //    high 인덱스는 `pivotIdx - 1`로 계속 차감될 수 있으므로 0보다 같거나 큰지 확인합니다.(high==0 경우 요소가 1개)
        //    low 인덱스는 `pivotIdx + 1`로 계속 증가될 수 있으므로 high 보다 같거나 작은지 확인합니다.(low==high 경우 요소가 1개)
        if (low >= high) return

        // 2. `pivot`을 선택합니다.
        //    - 무작위(Random)
        //    - 첫 원소(First)
        //    - 마지막 원소(Last)
        //    - 3개 이상의 후보 중 중앙값(Median-of-three) 등
        //
        // 3. 하위 배열로 분할합니다.
        //    피벗을 기준으로 분할할 지점을 결정하면서 요소의 순서를 변경합니다.
        //    - 요소 < pivot: 분할 지점 앞에 위치시킵니다.
        //    - pivot < 요소: 분할 지점 뒤에 위치시킵니다.
        //    - pivot == 요소: 앞이든 뒤든 위치할 수 있습니다.
        //    피벗의 인스턴스가 하나 이상 존재하므로, 대부분의 분할 루틴은 분할 지점에 도달하는 값이 피벗과 동일하고
        //    최종 지점에 도달했다는 것을 보장합니다.
        val pivotIdx = lomutoPartition(array, low, high)

        // 4. 분할 지점까지의 하위 범위와 그 다음 하위 범위까지 퀵 정렬을 재귀적으로 적용합니다.
        //    가능하면 `pivot`과 값이 같은 요소는 분할 지점에서 제외합니다.

        // hoarePartition 경우 `do {} while()`로 통해 범위 바깥부터 시작하여 인덱스 이동하게 하므로
        // 우측 파티션의 끝을 `pivotIdx - 1`이 아닌 `pivotIdx`로 넘겨야 합니다.
        // quickSort(array, low, pivotIdx)
        // 반면 lomutoPartition 경우, 피벗 기준으로 좌/우가 피벗 이하, 피벗 이상으로 정확히 나뉘고,
        // 피벗은 더이상 정렬의 대상이 아니게 됩니다.
        quickSort(array, low, pivotIdx - 1)
        quickSort(array, pivotIdx + 1, high)
    }

    /**
     * 피벗의 위치를 확정하면서 정렬합니다.
     *
     * ```
     * [(10, 25, 7, 11, 3, 4, 21, 12, 1, 9), 43, (111, 87, 88)], low: 0, pivotIdx: 10(43), high: 13
     * [(7, 3, 4, 1), 9, (10, 21, 12, 11, 25), 43, 111, 87, 88], low: 0, pivotIdx: 4(9), high: 9
     * [(1), (3, 4, 7), 9, 10, 21, 12, 11, 25, 43, 111, 87, 88], low: 0, pivotIdx: 0(1), high: 3
     * [1, (3, 4), (7), 9, 10, 21, 12, 11, 25, 43, 111, 87, 88], low: 1, pivotIdx: 3(7), high: 3
     * [1, (3), (4), 7, 9, 10, 21, 12, 11, 25, 43, 111, 87, 88], low: 1, pivotIdx: 2(4), high: 2
     * [1, 3, 4, 7, 9, (10, 21, 12, 11), (25), 43, 111, 87, 88], low: 5, pivotIdx: 9(25), high: 9
     * [1, 3, 4, 7, 9, (10), 11, (12, 21), 25, 43, 111, 87, 88], low: 5, pivotIdx: 6(11), high: 8
     * [1, 3, 4, 7, 9, 10, 11, (12), (21), 25, 43, 111, 87, 88], low: 7, pivotIdx: 8(21), high: 8
     * [1, 3, 4, 7, 9, 10, 11, 12, 21, 25, 43, (87), 88, (111)], low: 11, pivotIdx: 12(88), high: 13
     * ```
     *
     * 왼쪽은 모두 pivot 이하, 오른쪽은 모두 pivot 초과한 값들이 위치하게 만들고,
     * 피벗은 좌/우 파티션 가운데에 자리를 잡음으로써 위치가 정렬된 상태라 할 수 있습니다.
     * 즉, 파티션 종료 시, 피벗의 왼쪽과 오른쪽에 각각 더 이상 피벗이 갈 수 없는 값들로 나뉘어집니다.
     *
     * Reference:
     * - [Lomuto partition scheme](https://en.wikipedia.org/wiki/Quicksort#Lomuto_partition_scheme)
     */
    fun lomutoPartition(array: IntArray, low: Int, high: Int): Int {
        val pivot = array[high] // 마지막 요소를 피벗으로 선택합니다.

        var i = low

        for (j in low until high) {
            //  현재 원소가 `pivot` 이하이면, i 위치와 swap 해서
            //  'pivot 이하' 구간을 한 칸 확장합니다.
            if (array[j] <= pivot) {
                swap(array, i, j)
                i++
            }
        }

        // 반복문이 종료해도 `pivot`은 여전히 array[high]에 위치합니다.
        // `pivot`을 'i' 위치로 옮김으로써,
        // [`pivot`보다 작은 값 | `pivot`보다 큰 값 | `pivot`] 배열을
        // [`pivot`보다 작은 값 | `pivot` | `pivot`보다 큰 값] 배열로 만듭니다.
        swap(array, i, high)

        return i
    }

    /**
     * left 및 right 두 포인터를 사용해서 교차할 때까지 이동하고, 교차 전이면 swap, 교차하면 종료합니다.
     * 그 결과 아래와 같이 두 구간으로 분할됩니다.
     * ```
     * [low..right 범위는 피벗 이하 | right+1..high 범위는 피벗 이상]
     * ```
     *
     * 예를 들면:
     * ```
     *    [10, 25, 7, 88, 11, 3, 4, 21, 12, 1, 9, 111, 87, 43], swap(0, 9)
     * => [1, 25, 7, 88, 11, 3, 4, 21, 12, 10, 9, 111, 87, 43], swap(1, 6)
     * => [1, 4, 7, 88, 11, 3, 25, 21, 12, 10, 9, 111, 87, 43], swap(2, 5)
     * => [1, 4, 3, 88, 11, 7, 25, 21, 12, 10, 9, 111, 87, 43], right = 2 = 다음 파티셔닝 위한 새로운 피벗
     *    |피벗 이하|<---------------- 피벗 이상 --------------->|
     * ```
     *
     * Reference:
     * - [Hoare partition scheme](https://en.wikipedia.org/wiki/Quicksort#Hoare_partition_scheme)
     */
    fun hoarePartition(array: IntArray, low: Int, high: Int): Int {
        println(array.contentToString())
        val pivot = array[((high - low) / 2) + low]
        var left = low - 1
        var right = high + 1

        while (true) {
            do {
                // var left = low
                // while(array[left] < pivot) {left++} 하지 않는 이유:
                // while(true)로 무한 루프를 돌게 되는데, 이때 만약 시작부터
                // - array[left] >= pivot: 좌측에 피벗보다 크거나 같은 값이 있는 경우
                // - pivot >= array[right]: 우측에 피벗보다 작거나 같은 값이 있는 경우
                // 이런 경우 left 또는 right 인덱스의 변화 없이 무한히 반복문을 돌게 됩니다.
                left++
            } while (array[left] < pivot) // 좌측에 pivot 보다 작은 값이 있는 동안 left 인덱스 증가
            // 좌측에는 피벗보다 작은 값들만 있어야 하므로, 좌측에서 피벗보다 크거나 같은 값을 발견하면 중지.
            do {
                right--
            } while (pivot < array[right]) // 우측에 pivot 보다 큰 값이 있는 동안 right 인덱스 증가
            // 우측에는 피벗보다 큰 값들만 있어야 하므로, 우측에서 피벗보다 작거나 같은 값을 발견하면 중지

            // 우측 인덱스가 좌측 인덱스와 접하게 되면 종료합니다.
            // [요소1, 요소2, 요소3, ..., 요소n-2, 요소n-1, 요소n]
            //                ^   ^
            //                |  left ->
            //            <- right
            // left 인덱스가 right 인덱스보다 크다는 것은,
            // 피벗 기준으로 좌/우 분할이 끝났음을 의미합니다.
            // 그리고 피벗 또는 피벗보다 작은 값을 새로운 파티션 피봇 기준으로 잡을 수 있도록,
            // right 인덱스를 반환합니다.
            if (left >= right) {
                return right
            }
            swap(array, left, right)
        }
    }

    /**
     * 가운데 피벗을 두고, 피벗 좌측에 작은 값, 우측에 큰 값을 두는 방식으로 분할을 구현해봅니다.
     * 다만, 피벗 인덱스에 left 또는 right 인덱스가 도달하면 한 칸씩 옮기는 로직이 있는데,
     * 같은 값이 여럿 존재하는 경우 제대로 정렬이 이뤄지지 않습니다.
     *
     * - Before: [6, 6, 7, 6, 5, 6, 8]
     * - After:  [5, 6, 6, 6, 7, 6, 8]
     *
     * 이는 [swap]을 하면서 피벗 인덱스를 계속 옮기기 때문입니다.
     *
     * 파티션 종료 시점(`pivot`이 확정된 시점)에 다음 조건이 보장되어야 합니다.
     * - 피벗 왼쪽에 있는 모든 원소 <= `pivot` <= 피벗 오른쪽에 있는 모든 원소
     *
     * 피벗이 자기 자리를 차지하여 피벗의 위치는 더 이상 정렬이 필요하지 않게 만들어야 합니다.
     *
     */
    fun partition(array: IntArray, low: Int, high: Int): Int {
        var pivotIdx = ((high - low) / 2) + low
        val pivot = array[pivotIdx]

        var left = low
        var right = high
        while (left < right) {
            println(array.contentToString())
            // left 인덱스가 pivot 인덱스까지 도달했으면, pivot 좌측에 작은 값들을 놓기 위해
            // pivot 인덱스 위치를 우측으로 옮깁니다.
            if (left == pivotIdx) {
                print("\tleft == pivotIdx, low: $low, high: $high, left: $left, right: $right, pivotIdx: $pivotIdx")
                swap(array, pivotIdx, pivotIdx + 1)
                pivotIdx++
                print(" => $pivotIdx\n")
                println("\t => ${array.contentToString()}")
            } else if (right == pivotIdx) {
                print("\tright == pivotIdx, low: $low, high: $high, left: $left, right: $right, pivotIdx: $pivotIdx")
                // right 인덱스가 pivot 인덱스까지 도달했으면, pivot 우측에 큰 값들을 놓기 위해
                // pivot 인덱스 위치를 좌측으로 옮깁니다.
                swap(array, pivotIdx, pivotIdx - 1)
                pivotIdx--
                print(" => $pivotIdx\n")
                println("\t => ${array.contentToString()}")
            }

            if (array[right] < pivot) {  // 3.1. 우측 요소가 `pivot`보다 작다면, 좌측으로 넘깁니다.
                swap(array, left, right)
                // array[left]에 `pivot`보다 작은 값이 들어간 게 확실하므로, left 인덱스를 증가시킵니다.
                left++
            } else if (pivot < array[left]) {  // 3.2. 좌측 요소가 `pivot`보다 크다면, 우측으로 넘깁니다.
                swap(array, left, right)
                // array[right]에 `pivot`보다 큰 값이 들어간 게 확실하므로, right 인덱스를 감소시킵니다.
                right--
            } else { // 3.2. left, right 모두 pivot 기준으로 잘 분할되었다면, 각각 다음 인덱스를 이동시킵니다.
                left++
                right--
            }
        }

        return pivotIdx
    }

    fun swap(array: IntArray, i: Int, j: Int) {
        val tmp = array[i]
        array[i] = array[j]
        array[j] = tmp
    }
}

fun main() {
    val s = Solution()

    val arr1 = intArrayOf(10, 25, 7, 88, 11, 3, 4, 21, 12, 1, 9, 111, 87, 43)
    val answer1 = s.subtract(arr1, 0, arr1.size)
    answer1.sort()
    s.quickSort(arr1, 0, arr1.size - 1)
    println(arr1.contentToString())
    println(answer1.contentEquals(arr1))

    val arr2 = intArrayOf(6, 6, 7, 6, 5, 6, 8)
    val answer2 = s.subtract(arr1, 0, arr1.size)
    answer2.sort()
    s.quickSort(arr2, 0, arr2.size - 1)
    println(arr2.contentToString())

    listOf(
        Triple(
            intArrayOf(1, 5, 2, 6, 3, 7, 4),
            arrayOf(
                intArrayOf(2, 5, 3),
                intArrayOf(4, 4, 1),
                intArrayOf(1, 7, 3),
            ),
            intArrayOf(5, 6, 3),
        ),
    ).forEach { tc ->
        println(s.solution(tc.first, tc.second).contentEquals(tc.third))
    }
}