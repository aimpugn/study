package p42579

/**
 * [베스트앨범](https://school.programmers.co.kr/learn/courses/30/lessons/42579?language=kotlin)
 */
class Solution {
    /**
     * 장르 별로 가장 많이 재생된 노래 두 개씩
     * 노래는 고유 번호로 구분
     * 노래 수록 기준
     * - 속한 노래가 많이 재생된 장르를 먼저 수록
     * - 장르 내에서 많이 재생된 노래를 먼저 수록
     * - 장르 내에서 재생 횟수가 같은 노래 경우, 고유 번호가 낮은 노래 먼저 수록
     *
     * 장르에 속한 곡이 하나라면, 하나의 곡만 선택
     *
     * 예제:
     *  0       1   2       3       4
     * [classic pop classic classic pop  ]
     * [500     600 150     800     2500 ]
     *
     * - 속한 노래가 많이 재생된 장르? pop(3,100)
     * - 장르 내에서 많이 재생된 노래? 4
     * - 장르 내에서 재생 횟수가 같은 노래 경우, 고유 번호가 낮은 노래 먼저 수록
     * pop  pop  classic   classic
     * 4 -> 1 -> 3      -> 0      -> "장르 별로 가장 많이 재생된 노래 두 개씩"이므로 2는 제외
     * [4, 1, 3, 0]
     *
     * @param genres 노래의 장르 나타내는 문자열 배열
     * - 인덱스 i는 고유 번호
     * - 1 <= genres.length <= 10,000
     * - 1 <= set(genres) < 100
     * - 모든 장르는 재생된 횟수가 다름
     *
     * @param plays 노래별 재생 횟수 나타내는 정수 배열
     * - 인덱스 i는 고유 번호
     * - 1 <= plays.length <= 10,000
     *
     * @return 베스트 앨범에 들어갈 노래의 고유 번호를 순서대로
     */
    fun solution(genres: Array<String>, plays: IntArray): IntArray {
        // 장르별로 카운트합니다.
        // 각 장르 별로 곡과 그 재생 횟수를 별도의 자료 구조로 만듭니다.
        // 장르별로 모은 곡을 재생 횟수별로 역정렬 합니다.

        // 1. 장르별 재생 횟수 합산
        val genrePlayCount = genres.indices
            .groupingBy { genres[it] }
            .fold(0) { acc, idx -> acc + plays[idx] }

        // 2. 장르별 재생 횟수 역순으로 정렬된 노래 목록
        val songsPerGenre = genres.indices
            .groupBy { idx -> genres[idx] }
            .mapValues { entry -> // 장르별 노래 고유 번호. ex: classic=[0, 2, 3]
                // Pair(노래 고유 번호, 노래 재생 횟수) 목록으로 만듭니다.
                val pairs = entry.value
                    .map { idx -> idx to plays[idx] }
                /*pairs.sortedWith { it, other ->
                    val itPlayCount = it.second
                    val otherPlayCount = other.second

                    if (itPlayCount < otherPlayCount) {
                        // 재생 횟수가 it보다 other가 더 크면, other가 앞에 와야 하므로 1을 리턴합니다.
                        return@sortedWith 1
                    } else if(itPlayCount == otherPlayCount) {
                        // 장르 내에서 재생 횟수가 같은 노래 경우, 고유 번호가 낮은 노래 먼저 수록합니다.
                        val itSongIdx = it.first
                        val otherSongIdx = other.first
                        // it이 작은 경우 앞으로 와야 하므로 -1을 리턴 합니다.
                        return@sortedWith if(itSongIdx < otherSongIdx) -1 else 1
                    } else {
                        // 재생 횟수가 it보다 other가 더 작다면, it이 앞에 와야 하므로 -1을 리턴합니다.
                        return@sortedWith -1
                    }
                }*/

                /*
                pairs.sortedWith(compareByDescending<Pair<Int, Int>> {
                    val playCount = it.second
                    playCount
                }.thenBy {
                    val songIdx = it.first
                    songIdx
                })
                */
                pairs.sortedWith { a, b ->
                    compareByDescending<Pair<Int, Int>> {
                        val playCount = it.second
                        playCount
                    }.thenBy {
                        val songIdx = it.first
                        songIdx
                    }.compare(a, b)
                }
            }

        // 3. 가장 많이 플레이된 장르부터 최대 두 개씩 꺼내고, 그 중 노래 고유 번호만 모아서 정수 배열로 만듭니다.
        return genrePlayCount.entries
            .sortedByDescending { it.value }
            .flatMap {
                val genreAndTotal = it
                println("genreAndTotal: $genreAndTotal") // pop=3100
                val genre = genreAndTotal.key
                // 장르별로
                val top2SongsPerGenre = songsPerGenre[genre]?.take(2).orEmpty()
                println("top2SongsPerGenre: $top2SongsPerGenre") // top2SongsPerGenre: [(4, 2500), (1, 600)]
                top2SongsPerGenre
            }
            .map {
                val songIdx = it.first
                songIdx
            }.toIntArray()
    }
}

fun main() {
    listOf(
        Pair(
            arrayOf("classic", "pop", "classic", "classic", "pop"),
            intArrayOf(500, 600, 150, 800, 2500)
        )
    ).forEach { pair ->
        println(Solution().solution(pair.first, pair.second).contentToString())
    }
}