package p43163

import java.util.LinkedList

/**
 * [단어 변환](https://school.programmers.co.kr/learn/courses/30/lessons/43163?language=kotlin)
 */
class Solution {
    /**
     * 두 개의 단어 begin, target
     *
     * 규칙 이용하여 begin에서 target으로 변환하는 가장 짧은 변환 과정 찾기
     * 1. 한 번에 한 개의 알파벳만 바꿀 수 있음
     * 2. words에 있는 단어로만 변환할 수 있음
     *
     * begin != target
     *
     * @param begin 변환을 시작할 단어
     * - 알파벳 소문자
     * - 3 <= begin.length <= 10
     * @param target 변환하려는 단어
     * - 알파벳 소문자
     * - 3 <= begin.length <= 10
     * @param words 단어 집합
     * - 3 <= words.length <= 50
     * - 중복 단어 없음
     * @return 최소 몇 단계 과정을 거쳐 begin => target 변환할 수 있는지
     * - 변환 불가능 시 0 리턴
     */
    fun solution(begin: String, target: String, words: Array<String>): Int {
        // hit, cog, [hot, dot, dog, lot, log, cog]
        // 1. hit
        //     ^
        //    hot
        // 2. hot
        //    ^
        //    dot
        // 3. dot
        //      ^
        //    dog
        // 4. dog
        //    ^
        //    cog
        // return 4
        //
        // hit => cog
        // 한 번에 한 개의 알파벳만 바꿀 수 있음.
        //
        // 1. dfs로 한 번에 하나의 알파벳만 바뀌는 목록을 만듭니다.
        // - 순서가 의미가 있으므로 순서가 다르면 다른 것으로 봅니다.
        //   hit -> [hot, dot]
        //   hit -x-> [dot, hot]
        // - 가능한 모든 경우의 수를 모읍니다.
        // - 마지막이 target이 아닌 경우의 수는 제외합니다.
        //   hit -> [hot]
        // - 남은 목록을 순회하며 최소로 변환되는 값을 찾습니다.

        // 단어 목록에 목표 문자열이 없으면 변환 불가
        if (!words.contains(target)) {
            return 0
        }
        if (begin == target) {
            return 1
        }

        val possibleCases = mutableListOf<MutableList<String>>()

        words.forEachIndexed { idx, word ->
            // 각 언어별로 먼저 시작할 수 있습니다.
            // begin과 알파벳 하나 차이인지 확인합니다.
            dfs(begin, mutableListOf(begin), words, BooleanArray(words.size) { false }, possibleCases)
        }

        val candidate = possibleCases
            .filter { it.last() == target }
            .minByOrNull { it.size }

        if (candidate == null) {
            return 0
        }

        return candidate.size - 1 // dfs 과정에서 포함된 begin 을 뺍니다.
    }

    /**
     * ```
     * println(s.isOneCharacterDiff("hit", "hot")) // true
     * println(s.isOneCharacterDiff("hit", "hox")) // false
     * ```
     */
    fun isOneCharacterDiff(a: String, b: String): Boolean {
        var diffCount = 0

        for (i in a.indices) {
            if (a[i] != b[i]) diffCount++
            if (diffCount > 1) return false
        }

        return diffCount == 1
    }

    fun dfs(
        currWord: String,
        currWords: MutableList<String>,
        words: Array<String>,
        visited: BooleanArray,
        candidates: MutableList<MutableList<String>>
    ) {
        for (i in 0 until words.size) {
            // 만약 한 단어 차이가 아니면, 굳이 경우의 수를 계속 이어갈 필요 없으므로 경우의 수에 담지 않습니다
            val nextWord = words[i]
            if (!visited[i] && isOneCharacterDiff(currWord, nextWord)) {
                val currWord = words[i]

                visited[i] = true
                // 방문하지 않았다면 추가했다고 표시하고
                // 현재 단어 목록 기반으로 새로운 목록을 만듭니다.
                val nextWords = currWords.toMutableList()
                // 가능한 새로운 경우의 수로 추가하고
                nextWords.add(currWord)
                candidates.add(nextWords)
                // 다음 단계를 진행합니다.
                dfs(nextWord, nextWords, words, visited, candidates)
                // 백트래킹: 방문하지 않은 것으로 표시함으로써, 다음 반복문에서 순서가 다르게 currWords에 추가되도록 합니다.
                visited[i] = false
            }
        }
    }

    /**
     * BFS 풀이
     */
    fun solution2(begin: String, target: String, words: Array<String>): Int {
        // 변환 가능한 단어가 없는 경우
        if (!words.contains(target)) return 0
        // 바로 변환 가능한 경우
        if (begin == target) return 1

        // BFS를 위한 큐
        val queue = LinkedList<Pair<String, Int>>() // (현재 단어, 변환 단계)
        val visited = BooleanArray(words.size) // 방문한 단어 기록

        queue.add(begin to 0)

        while (queue.isNotEmpty()) {
            // [(hit, 0)]
            // [(hot, 1)]
            // [(dot, 2), (lot, 2)]
            // [(lot, 2), (dog, 3)]
            // [(dog, 3), (log, 3)]
            // [(log, 3), (cog, 4)]
            //
            // 큐에 쌓였던 내역:
            // [(hit, 0), (hot, 1), (dot, 2), (lot, 2), (dog, 3), (log, 3), (cog, 4)]
            //
            // 변환 과정으로 본다면:
            // hit -> hot -> dot -> dog -> cog => 4 리턴
            //            -> lot -> log -> cog
            val (currentWord, step) = queue.poll()

            // 목표 단어에 도달하면 해당 값을 반환합니다.
            if (currentWord == target) return step

            // 변환 가능한 단어 탐색
            words.forEachIndexed { idx, word ->
                if (!visited[idx] && isOneCharacterDiff(currentWord, word)) {
                    visited[idx] = true
                    queue.add(word to step + 1)
                }
            }
        }

        return 0 // 변환 불가능한 경우
    }
}

class TestCase(val begin: String, val target: String, val words: Array<String>)

fun main() {
    val s = Solution()

    listOf(
        TestCase("hit", "cog", arrayOf("hot", "dot", "dog", "lot", "log", "cog")),
        TestCase("hit", "cog", arrayOf("hot", "dot", "dog", "lot", "log")),
    ).forEach { tc ->
        println(s.solution2(tc.begin, tc.target, tc.words))
    }
}