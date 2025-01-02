package p87946

import kotlin.math.*

class Solution {
    /**
     * 일정 피로도 사용해서 던전 탐험
     * - 최소 필요도: 각 던전마다 탐험 시작 위해 필요
     * - 소모 피로도: 던전 탐험 마쳤을 때 소모
     *
     * 하루 한 번씩 탐험할 수 있는 던전 여러개 존재(중복 탐험 불가)
     * 최대한 많이 탐험하려고
     *
     * Example:
     * - 최소 필요 소모도 80, 소모 피로도 20
     *
     *   현재 남은 피로도가 80 이상이어야 하고,
     *   던전 탐험 후 피로도 20 차감
     *
     * - k=80, [[80,20],[50,40],[30,10]]
     *
     *   1. 0 -> 1 -> 2 인덱스 순서로 탐험
     *      0: 탐험 가능, 탐험 후 k=60
     *      1: 탐험 가능, 탐험 후 k=20
     *      2: 탐험 불가
     *   2. 0 -> 2 -> 1 인덱스 순서로 탐험
     *      0: 탐험 가능, 탐험 후 k=60
     *      2: 탐험 가능, 탐험 후 k=50
     *      1: 탐험 가능, 탐험 후 k=10
     *
     * See:
     * - [완전탐색 - 피도로](https://school.programmers.co.kr/learn/courses/30/lessons/87946)함
     *
     * @param k 유저의 현재 피로도:
     *      - 1 <= k <= 5,000 자연수
     * @param dungeons - [(최소 필요 피로도, 소모 피로도), ...] 2차원 배열
     *      - 1 <= 세로(행) 길이 = 던전의 개수 = dungeons.length <= 8.
     *        최대 8! (40,320) 개의 가능한 경로가 존재.
     *      - 최소 필요 피로도 >= 소모 피로도.
     *        따라서 던전 탐험이 가능하다면 탐험 후 마이너스가 되지는 않음.
     *      - 1 <= 최소 필요 피로도, 소모 피로도 <= 1,000 자연수
     *      - 서로 다른 던전이어도 (최소 필요 피로도, 소모 피로도) 쌍이 같을 수 있
     */
    fun solution(k: Int, dungeons: Array<IntArray>): Int {
        // 값이 같아도 각 던전을 다르게 취급
        // 접근 순서에 따라 탐험할 수 있을 수도 없을 수도 있음 -> 백트래킹을 사용하면 순서에 따라 다른 결과를 체크할 수 있을 거 같음
        return exploreDFS(k, dungeons, 0, BooleanArray(dungeons.size) { false })
    }

    /**
     * 현재까지 탐험한 던전의 수를 기준으로 최대 탐험 가능한 던전 수를 계산합니다.
     *
     * @param currFatigue 현재 피로도. 재귀적으로 호출되면서 차감되는 피로도를 나타냅니다.
     * @param dungeons 탐험 가능한 던전
     * @param visited 방문한 던전의 인덱스 목록
     */
    private fun exploreDFS(
        currFatigue: Int,
        dungeons: Array<IntArray>,
        count: Int,
        visited: BooleanArray
    ): Int {
        // 현재 탐험한 던전의 수와 재귀적으로 탐험한 경로에서 도출된 값을 비교하여 최대값을 유지
        var maxCount = count

        // DFS + 백트래킹은 탐험 경로를 동적으로 생성하고, 고정된 시작점 없이 각 던전의 남은 피로도를 고려하여 탐험 가능한 모든 경로를 탐색하게 됩니다.
        // 각 DFS 호출은 자신의 서브트리에서 탐험 가능한 최대 던전 수를 계산하고 반환합니다.
        dungeons.forEachIndexed { idx, dungeon ->
            // 방문한 적 없음 && 탐험 가능
            if (!visited[idx] && currFatigue >= dungeon[0]) {
                // 탐험 시작 (방문 처리)
                visited[idx] = true
                // 다음 던전 탐험 (현재 던전을 탐험했으므로, 피로도 감소)
                // 중간 단계에서 최종 결과를 아직 모르지만, 계속 max를 비교합니다.
                val newCount = exploreDFS(currFatigue - dungeon[1], dungeons, count + 1, visited)
                maxCount = max(newCount, maxCount)
                visited[idx] = false
            }
        }

        return maxCount
    }
}


class Solution2 {
    /**
     * @param k 초기 피로도
     * @param dungeons 각 던전의 [최소 필요 피로도, 소모 피로도]를 담은 2차원 배열
     * @return 탐험 가능한 최대 던전 수
     */
    fun solution(k: Int, dungeons: Array<IntArray>): Int {
        // 각 던전의 방문 여부를 추적하는 불리언 배열
        // false로 초기화하여 모든 던전을 처음에는 방문하지 않은 상태로 설정
        val visited = BooleanArray(dungeons.size) { false }

        // DFS 함수를 호출하여 최대 탐험 가능 던전 수를 계산하고 반환
        return dfs(k, dungeons, visited)
    }

    /**
     * DFS를 사용하여 모든 가능한 던전 탐험 순서를 탐색합니다.
     *
     * @param fatigue 현재 남은 피로도
     * @param dungeons 던전 정보를 담은 2차원 배열
     * @param visited 각 던전의 방문 여부를 나타내는 불리언 배열
     * @return 현재 상태에서 탐험 가능한 최대 던전 수
     */
    private fun dfs(fatigue: Int, dungeons: Array<IntArray>, visited: BooleanArray): Int {
        // 현재 상태에서 탐험 가능한 최대 던전 수를 저장
        var maxExplored = 0

        // 모든 던전을 순회하며 탐험 가능 여부 확인
        for (i in dungeons.indices) {
            // 아직 방문하지 않음 && 현재 피로도로 탐험 가능한 던전
            if (!visited[i] && fatigue >= dungeons[i][0]) {
                // 현재 던전을 방문했다고 표시
                visited[i] = true

                // 현재 던전을 탐험한 후의 상태로 재귀 호출
                // 1을 더해 현재 던전 탐험을 카운트하고, 소모된 피로도를 차감
                val explored = 1 + dfs(fatigue - dungeons[i][1], dungeons, visited)

                // 최대 탐험 던전 수 갱신
                maxExplored = maxOf(maxExplored, explored)

                // 백트래킹: 현재 던전을 방문하지 않은 상태로 복원
                visited[i] = false
            }
        }

        // 현재 상태에서 탐험 가능한 최대 던전 수 반환
        return maxExplored
    }
}

fun main() {
    println(
        Solution().solution(
            80,
            arrayOf(
                intArrayOf(80, 20),
                intArrayOf(50, 40),
                intArrayOf(30, 10),
            )
        )
    )
}