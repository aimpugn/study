package p389479

import java.util.PriorityQueue

/**
 * [서버 증설 횟수](https://school.programmers.co.kr/learn/courses/30/lessons/389479?language=kotlin)
 */
class Solution {
    /**e
     * 온라인 게임 운영.
     * 같은 시간대에 게임 이용하는 사람이 m명 늘어날 때마다 서버 1대 추가 필요
     * - 어느 시간대 이용자가 m명 미만이면? 서버 증설 불필요
     * - n * m 명 <= 어느 시간대 이용자 수 < (n + 1) * m
     *   => 최소 n대의 증설된 서버가 운영중이어야 함
     * - 한 번 증설한 서버는 k 시간 동안 운영하고 반납. 예를 들어,
     *   k = 5 경우, 10시에 증설한 서버는 10 ~ 15시에만 운영
     *
     * - 변수 정리
     *   - m: 서버 증설 기준
     *   - n: 증설된 서버 수
     *   - k: 서버 운영 시간
     *
     * Q. 하루 동안 모든 게임 이용자가 게임을 하기 위해 서버를 최소 몇 번 증설해야 하는지
     *
     * 같은 시간대에 서버를 x대 증설했다면 해당 시간대의 증설 횟수는 x회
     *
     * ex) m = 3, k = 5일 때 시간대별 증설된 서버의 수와 증설 횟수 예시
     * 시각       게임 이용자의 수    증설된 서버의 수    증설 횟수
     * 0 ~ 1        0               0               0
     * 1 ~ 2        2               0               0
     * 2 ~ 3        3               1               1
     * 3 ~ 4        3               1               0
     * 4 ~ 5        1               1               0
     * 5 ~ 6        2               1               0
     * 6 ~ 7        0               1               0
     * 7 ~ 8        0               0               0
     * 8 ~ 9        0               0               0
     * 9 ~ 10       0               0               0
     * 10 ~ 11      4               1               1
     * 11 ~ 12      2               1               0
     * 12 ~ 13      0               1               1
     * 13 ~ 14      6               2               1
     * 14 ~ 15      0               2               0
     * 15 ~ 16      4               1               0
     * 16 ~ 17      2               1               0
     * 17 ~ 18      13              4               3
     * 18 ~ 19      3               3               0
     * 19 ~ 20      5               3               0
     * 20 ~ 21      10              3               0
     * 21 ~ 22      0               3               0
     * 22 ~ 23      1               0               0
     * 23 ~ 24      5               1               1
     *
     * 모든 이용자 감당 위해 최소 7번 서버 증설이 필요.
     *
     * @param players 00시 ~ 23시까지 시간대별 게임 이용자 수를 나타내는 1차원 정수 배열
     *        - `players.length` == 24
     *        - 0 <= `players[i]` <= 1,000
     *        - `players[i]`는 `i`시 ~ `i+1`시 사이의 이용자 수
     * @param m 서버 한 대로 감당할 수 있는 최대 이용자 수
     *        - 1 <= m <= 1,000
     * @param k 서버 한 대가 운영 가능한 시간
     *        - 1 <= k <=24
     *
     * @return 모든 게임 이용자를 감당하기 위한 최소 서벌 증설 횟수
     */
    fun solution(players: IntArray, m: Int, k: Int): Int {
        var answer: Int = 0

        // 기본적으로 한 대의 서버가 존재하여 m 미만의 유저가 플레이중이라면 증설이 필요 없습니다.
        // 따라서 `1 + (현재 이용자수 / m)`가 현재 있어야 하는 서버의 대수가 됩니다.
        // 가령 m=3 경우 현재 이용자수가 3이면 `1 + (3 / 3) = 2`로 총 두 대의 서버가 존재하고, 1번 증설됩니다.
        //
        // 증설된 서버에는 timeout이 존재해서 k 시간이 지나면 제거되어야 합니다.
        // 그리고 시각별로 증설된 서버별로 몇 시간 운영되었는지 알아야 합니다.
        //
        // 일단 드는 생각은, 현재 몇 대의 서버가 존재하고, 해당 서버로 운영이 가능한지 판단하고,
        // 각 서버의 운영 시간을 관리하는 자료구조가 필요할 것으로 보입니다.
        val servers = PriorityQueue<Server>(compareBy { it.timeout });

        players.forEachIndexed { idx, player ->
            println("[$idx] player: $player, servers: $servers")
            modifyTimeoutAndClear(servers)

            // 플레이어가 m 으로 나눈 값이 0보다 큰지 그리고 현재 서버가 감당 가능한지
            var serversNeeded = player / m;
            if (serversNeeded == 0) {
                return@forEachIndexed
            }

            if (serversNeeded < servers.size) {
                return@forEachIndexed
            }

            var serversAdded = serversNeeded - servers.size
            for (i in 1..serversAdded) {
                servers.add(Server(k))
            }
            answer += serversAdded
        }

        return answer
    }

    data class Server(var timeout: Int) {}

    fun modifyTimeoutAndClear(servers: PriorityQueue<Server>) {
        if (servers.isEmpty()) return

        servers.map { it.timeout-- }

        val iter = servers.iterator()
        while (iter.hasNext()) {
            val item = iter.next()
            if (item.timeout <= 0) {
                iter.remove()
            }
        }
    }
}

fun main() {
    val s = Solution()
    println("1.")
    println(
        "1. " +
                s.solution(
                    intArrayOf(0, 2, 3, 3, 1, 2, 0, 0, 0, 0, 4, 2, 0, 6, 0, 4, 2, 13, 3, 5, 10, 0, 1, 5),
                    3,
                    5
                )
    )
    println(
        "2. " +
                s.solution(
                    intArrayOf(0, 0, 0, 10, 0, 12, 0, 15, 0, 1, 0, 1, 0, 0, 0, 5, 0, 0, 11, 0, 8, 0, 0, 0),
                    5,
                    1
                )
    )
    println(
        "3. " +
                s.solution(
                    intArrayOf(0, 0, 0, 0, 0, 2, 0, 0, 0, 1, 0, 5, 0, 2, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1),
                    1,
                    1
                )
    )
}