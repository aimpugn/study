package p178871

/**
 * [달리기 경주](https://school.programmers.co.kr/learn/courses/30/lessons/178871)
 */
class Solution {
    /**
     * 자기 바로 앞 선수 추월 시, 추월한 선수의 이름 부릅니다.
     * 1     2    3
     * mumu, soe, poe
     *
     * => soe 부름
     *    1    2     3
     * => soe, mumu, poe
     *
     * @param players 1등부터 현재 등수 순서대로 선수들 이름이 담긴 배열
     * - 5 <= players.length <= 50,000
     * - players.i: i 번째 선수의 이름
     * - 선수 이름은 알파벳 소문자로만 이루어져 있음
     * - 중복된 이름 없음
     * - 3 <= players.i <= 10
     * @param callings 해설진이 부른 이름을 담은 문자열 배열
     * - 2 <= callings <= 1,000,000
     * - players 원소들로만 이루어져 있음
     * - 경주 진행중 1등인 선수 이름은 불리지 않음
     * @return 경주 종료 시 선수들의 이름을 1등부터 등수 순서대로 담은 배열
     */
    fun solution(players: Array<String>, callings: Array<String>): Array<String> {
        var answer: Array<String> = arrayOf<String>()
        // 플레이어: [mumu, soe, poe, kai, mine]
        // 호명: [kai, kai, mine, mine]
        // 1. kai
        // => [mumu, soe, kai, poe, mine]
        // 2. kai
        // => [mumu, kai, soe, poe, mine]
        // 3. mine
        // => [mumu, kai, soe, mine, poe]
        // 4. mine
        // => [mumu, kai, mine, soe, poe]

        // 호명하면 player[i-1], player[i] = player[i], player[i-1] 로 서로 바꾸도록 합니다.
        // 호명한 선수의 현재 인덱스를 알아야 하므로, 선수별로 현재 인덱스를 관리합니다.
        val playerIndex = mutableMapOf<String, Int>().apply {
            players.forEachIndexed { idx, playerName ->
                put(playerName, idx)
            }
        }

        // 호출한 순서대로 처리합니다.
        callings.forEach { calledPlayerName ->
            playerIndex[calledPlayerName]?.let { calledPlayerIdx ->
                val idxToSwap = calledPlayerIdx - 1
                val playerNameToSwap = players[idxToSwap]
                // 위치를 서로 바꿔줍니다.
                players[idxToSwap] = players[calledPlayerIdx]
                players[calledPlayerIdx] = playerNameToSwap

                // 변경된 인덱스를 업데이트합니다.
                playerIndex.put(calledPlayerName, idxToSwap)
                playerIndex.put(playerNameToSwap, calledPlayerIdx)
            }
        }

        return players
    }
}


fun main() {

    val s = Solution()

    listOf(
        Triple(
            arrayOf("mumu", "soe", "poe", "kai", "mine"),
            arrayOf("kai", "kai", "mine", "mine"),
            arrayOf("mumu", "kai", "mine", "soe", "poe")
        )
    ).forEach { tc ->
        println(s.solution(tc.first, tc.second).contentEquals(tc.third))
    }

}