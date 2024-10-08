package p42578

import java.lang.StringBuilder
import java.util.LinkedList
import kotlin.math.max

/**
 * [의상](https://school.programmers.co.kr/learn/courses/30/lessons/42578)
 */
class Solution {
    fun solution(clothes: Array<Array<String>>): Int {
        // 의상 종류별로 의상 이름들을 그룹핑 합니다.
        val categorized = clothes.groupBy({ it[1] }, { it[0] })
        // {headgear=[yellow_hat, green_turban], eyewear=[blue_sunglasses]}


        return dfsOnlyCount(
            categorized,
            categorized.keys.toList(), // DFS 과정에서 카테고리 인덱스(categoryIdx 매개변수)를 증가시키며 탐색하도록 합니다. 깊이가 깊어질수록 앞선 인덱스의 카테고리는 고려 대상에서 제외됩니다.
            0
        ) - 1 // 아무것도 입지 않는 경우 제외
    }

    /**
     * 원래는 모든 경우의 수를 List로 만들었는데, 그러면 시간 초과가 발생합니다.
     * 이를 해결하기 위해 '현재 경우의 수 * 다음 경우의 수'를 계산하는 방식으로 개선합니다.
     *
     * @param categorized 의상 종류별로 의상 이름들을 정리한 맵입니다.
     * @param categories 의상 종류 목록입니다.
     * @param categoryIdx 처리한 카테고리를 추적하기 위한 인덱스입니다.
     * @return 유효한 조합의 수
     */
    private fun dfsOnlyCount(
        categorized: Map<String, List<String>>,
        categories: List<String>,
        categoryIdx: Int
    ): Int {
        // 모든 카테고리를 처리했을 때 1가지 경우로 계산하도록 리턴합니다.(기저 조건)
        if (categoryIdx == categories.size) {
            return 1
        }
        val currentCategory = categories[categoryIdx]
        val wearables = categorized[currentCategory]!!

        // 현재 카테고리에서 선택 가능한 개수는 (모두 선택 안함 1 + 하나씩 선택)
        val currentCategoryNumberOfCases = 1 + wearables.size
        val nextCategoryNumberOfCases = dfsOnlyCount(categorized, categories, categoryIdx + 1)

        // 현재 카테고리에서 선택 가능한 경우의 수 * 다음 카테고리에서 선택 가능한 경우의 수
        return currentCategoryNumberOfCases * nextCategoryNumberOfCases
    }


    /**
     * 결과: 실패. 모든 조합을 생성할 경우 시간 초과가 발생
     * - 합계: 96.4 / 100.0
     * - 정확성: 96.4
     * ```
     * 테스트 1 〉 실패 (시간 초과)
     * 테스트 2 〉 통과 (6.71ms, 61.1MB)
     * 테스트 3 〉 통과 (11.77ms, 65.3MB)
     * 테스트 4 〉 통과 (689.07ms, 366MB)
     * 테스트 5 〉 통과 (8.68ms, 61.2MB)
     * 테스트 6 〉 통과 (6.32ms, 60.9MB)
     * 테스트 7 〉 통과 (896.43ms, 365MB)
     * 테스트 8 〉 통과 (26.08ms, 68.5MB)
     * 테스트 9 〉 통과 (6.51ms, 60MB)
     * 테스트 10 〉 통과 (8.91ms, 60.6MB)
     * 테스트 11 〉 통과 (8.01ms, 60MB)
     * 테스트 12 〉 통과 (135.39ms, 92.1MB)
     * 테스트 13 〉 통과 (144.40ms, 82MB)
     * 테스트 14 〉 통과 (6.25ms, 61.2MB)
     * 테스트 15 〉 통과 (5.95ms, 60.7MB)
     * 테스트 16 〉 통과 (5.89ms, 61.7MB)
     * 테스트 17 〉 통과 (6.17ms, 60.1MB)
     * 테스트 18 〉 통과 (12.70ms, 67.6MB)
     * 테스트 19 〉 통과 (13.89ms, 67.6MB)
     * 테스트 20 〉 통과 (8.24ms, 60.4MB)
     * 테스트 21 〉 통과 (7.11ms, 60.6MB)
     * 테스트 22 〉 통과 (6.40ms, 61MB)
     * 테스트 23 〉 통과 (5.97ms, 61.3MB)
     * 테스트 24 〉 통과 (8.82ms, 61.2MB)
     * 테스트 25 〉 통과 (14.89ms, 65.2MB)
     * 테스트 26 〉 통과 (867.86ms, 365MB)
     * 테스트 27 〉 통과 (5.96ms, 60.7MB)
     * 테스트 28 〉 통과 (20.05ms, 68.3MB)
     * ```
     *
     * - 매일 다른 옷을 조합
     * - 각 종류별로 최대 1가지 의상만 착용
     * - 의상 일부가 겹쳐도, 다른 의상이 겹치지 않거나 OR 추가 의상 착용 => 서로 다른 방법으로 계산
     * - 하루에 최소 한 개 의상 입음
     *
     * Example
     * - 1d: 동그란 안경, 긴 코트, 파란색 티셔츠
     * - 2d: 동그란 안경, 긴 코트, 파란색 티셔츠, 청바지 OR 검정 선글라스, 긴 코트, 파란색 티셔츠
     *
     * - [["yellow_hat", "headgear"], ["blue_sunglasses", "eyewear"], ["green_turban", "headgear"]]
     *      headgear: yellow_hat, green_turban
     *      eyewear: blue_sunglasses
     *      가능한 조합:
     *      // 하나씩 선택
     *      - yellow_hat
     *      - green_turban
     *      - blue_sunglasses
     *      // 조합하여 선택
     *      - yellow_hat + blue_sunglasses
     *      - green_turban + blue_sunglasses
     *      서로 다른 옷의 조합의 수: 5
     *
     * - [["crow_mask", "face"], ["blue_sunglasses", "face"], ["smoky_makeup", "face"]]
     *      face: crow_mask, blue_sunglasses, smoky_makeup
     *      서로 다른 옷의 조합의 수: 3
     *
     * @param clothes [(의상 이름, 의상 종류), ...] 2차원 배열
     *  - 1 <= clothes.length <= 30
     *  - 같은 이름 의상 없음
     *  - 모든 원소는 문자열
     *  - 1 <= 원소.length <= 20 자연수. 알파벳 소문자 || '_' 로 구성
     *
     * @return 서로 다른 옷의 조합의 수
     */
    fun solutionByActualCombination(clothes: Array<Array<String>>): Int {
        // 1. 우선 옷들을 의상 종류에 따라 분류합니다.
        val categorized = clothes.groupBy(
            keySelector = { it[1] },
            valueTransform = { it[0] }
        )
        // {headgear=[yellow_hat, green_turban], eyewear=[blue_sunglasses]}

        // 실제 조합을 저장할 리스트

        // 2. 의상 종류별 && 중복되지 않게 하나씩 선택하도록 합니다.
        //      - 당장 떠오르는 것은 dfs & 백트래킹으로 조합을 만들면서 경우의 수를 카운트하는 방법
        //        이미 선택한 그룹은 제외하고 나머지 그룹에서 선택하도록 해야 함
        // 순서는 상관 없고 & 가능한 모든 경우의 수는 고려해야 하고
        return dfsActualCombination(categorized, categorized.keys.toList(), 0, mutableListOf())
    }

    /**
     * @param categoryIndex 현재 처리 중인 카테고리의 인덱스입니다. 진행 상태를 추적합니다.
     */
    private fun dfsActualCombination(
        categorized: Map<String, List<String>>,
        categories: List<String>,
        categoryIndex: Int,
        currentCombination: MutableList<String>
    ): Int {
        if (categoryIndex == categories.size) {
            if (currentCombination.isNotEmpty()) {
                return 1
            }

            return 0
        }

        var count = 0
        val currentCategory = categories[categoryIndex]
        val wearables = categorized[currentCategory]!!

        // 현재 카테고리에서 아무것도 선택하지 않는 경우:
        // categoryIndex + 1을 전달하여 다음 카테고리로 이동합니다.
        count += dfsActualCombination(categorized, categories, categoryIndex + 1, currentCombination)

        // 현재 카테고리에서 입을 것을 선택한 경우:
        for (wearable in wearables) {
            currentCombination.add(wearable)
            // categoryIndex + 1을 전달하여 다음 카테고리로 이동합니다.
            count += dfsActualCombination(categorized, categories, categoryIndex + 1, currentCombination)
            currentCombination.removeLast()
        }

        return count
    }


    /**
     * 메모이제이션을 활용해도 시간 초과가 발생합니다. [solution]에서 [dfsOnlyCount]만 사용하는 게 문제 해결 방법으로 보입니다.
     */
    fun solutionByCombinationWithMemo(clothes: Array<Array<String>>): Int {
        val categorized = clothes.groupBy({ it[1] }, { it[0] })
        val categories = categorized.keys.toList()
        val memo = mutableMapOf<Pair<Int, Set<String>>, Int>()

        return countByDFSWithMemo(categorized, categories, 0, setOf(), memo) - 1 // 아무것도 입지 않는 경우 제외
    }

    /**
     * DFS 메서드는 부분 문제(특정 카테고리까지의 조합 수)의 해결입니다. 부분 문제의 해결이 전체 문제(모든 카테고리의 조합 수)의 해결에 독립적으로 기여합니다.
     *
     * @param selectedWearables 같은 이름 의상 없다고 했으므로, 선택한 의류를 Set으로 관리합니다.
     * @param memo '현재 카테고리 인덱스와 선택된 아이템들의 집합'을 키로 삼습니다.
     */
    private fun countByDFSWithMemo(
        categorized: Map<String, List<String>>,
        categories: List<String>,
        categoryIndex: Int,
        selectedWearables: Set<String>,
        memo: MutableMap<Pair<Int, Set<String>>, Int>
    ): Int {
        // 카테고리 인덱스가 사이즈와 같다는 것은, 상위 메서드 호출에서 이미 카테고리 끝에 도달했음을 의미합니다.
        // 따라서 여기서는 카테고리 끝까지 도달하여, 모든 조합을 고려했다는 의미로 1을 반환합니다.(기저 조건)
        if (categoryIndex == categories.size) {
            return 1
        }

        // 메모 체
        val memoKey = Pair(categoryIndex, selectedWearables)
        if (memo.containsKey(memoKey)) {
            return memo[memoKey]!!
        }

        var count = 0
        val currentCategory = categories[categoryIndex]
        val wearables = categorized[currentCategory]!!

        // 경우의 수 1: 현재 카테고리에서 아무것도 선택하지 않는 경우
        count += countByDFSWithMemo(categorized, categories, categoryIndex + 1, selectedWearables, memo)

        // 경우의 수 2: 현재 카테고리에서 의상을 선택하는 경우
        for (item in wearables) {
            if (item !in selectedWearables) {
                val newSelectedWearables = selectedWearables + item
                count += countByDFSWithMemo(
                    categorized,
                    categories,
                    categoryIndex + 1,
                    newSelectedWearables,
                    memo
                )
            }
        }

        // 이번 재귀 호출의 결과를 기록합니다.
        memo[memoKey] = count
        return count
    }
}

data class TestCase(val arg1: Array<Array<String>>);

fun main() {
    val testCases = listOf(
        TestCase(
            arrayOf(
                arrayOf("yellow_hat", "headgear"),
                arrayOf("blue_sunglasses", "eyewear"),
                arrayOf("green_turban", "headgear"),
            )
        )
    )

    for (testCase in testCases) {
        println(Solution().solution(testCase.arg1))
    }
}


