package p84512

/**
 * [모음사전](https://school.programmers.co.kr/learn/courses/30/lessons/84512?language=kotlin)
 */
class Solution {
    val vowels = listOf("A", "E", "I", "O", "U")

    /**
     * 모든 경우의 수를 세야 하므로, 클래스 속성으로 카운트합니다.
     */
    var counter = 0

    /**
     * 사전에 수록된 단어
     * - 알파벳 모음 A, E, I, O, U만 사용하여 만들 수 있는
     * - 길이 5 이상의 모든 단어
     *
     * 사전의 첫번째 단어는 A, 그 다음은 AA, ... 마지막 단어는 UUUUU
     *
     * @param word 찾으려는 단어
     * @return 주어진 단어가 사전에서 몇 번째 단어인지
     */
    fun solution(word: String): Int {
        // word: AAAAE
        // A => AA => AAA => AAAA => AAAAA
        //                        => AAAAE
        // : 6번째
        //
        // word: AAAE
        // A => AA => AAA => AAAA => AAAAA
        //                        => AAAAE
        //                        => AAAAI
        //                        => AAAAO
        //                        => AAAAU
        //                => AAAE
        // : 10번째

        // 시작 단어는 A, E, I, O, U 입니다.
        // 각 시작 단어부터 시작하여, 중복을 허용하여, 길이 5가 완성될 때까지 계속 붙여 나갑니다.
        // 그리고 각 길이별로 끝 단어가 5개의 문자를 다 소진하면 끝납니다.
        // 일단, 주어진 조건에 따라 단어들을 만들어 보고, 그 중간에 찾는 대상 일치 여부 조건을 추가합니다.
        //
        // 주어진 패턴을 보면 A, AA, AAA 처럼 같은 문자를 먼저 깊게 파고 들어가는 모양이라
        // DFS 방식으로 구현합니다.
        //
        // dfs(문자열빌더, 카운터, word): Int
        //   if 문자열빌더.length == word.length && 문자열 == word
        //      return 카운터
        //
        //   if 문자열빌더.length == 5
        //      if 문자열 == word
        //          return 카운터
        //      else
        //          return -1
        //
        //   [A, E, I, O, U].forEach { vowel ->
        //      문자열빌더.add(vowel)
        //      if dfs(문자열빌더, 카운터++) != -1
        //          return counter
        //      문자열빌더.delete(마지막문자)
        //   }
        //
        //   return -1

        dfs(StringBuilder(), word)

        return counter
    }

    fun dfs(strBuilder: StringBuilder, word: String): Boolean {
        if (strBuilder.length == word.length && strBuilder.toString() == word) {
            return true
        }
        if (strBuilder.length == 5) {
            return strBuilder.toString() == word
        }

        vowels.forEach { vowel ->
            counter++
            strBuilder.append(vowel)
            if (dfs(strBuilder, word)) {
                return true
            }
            // 마지막 문자를 제거합니다.
            // ```
            // val strBuilder = StringBuilder()
            // strBuilder.append("A")
            // strBuilder.append("B")
            // strBuilder.append("C")
            // println(strBuilder) // ABC
            //
            // strBuilder.deleteAt(strBuilder.length - 1)
            // println(strBuilder) // AB
            // ```
            strBuilder.deleteAt(strBuilder.length - 1)
        }

        return false
    }
}

fun main() {

    listOf(
        Pair("AAAAE", 6),
        Pair("AAAE", 10),
        Pair("AAAI", 16),
        Pair("AAAU", 28),
        Pair("I", 1563),
        Pair("EIO", 1189),
    ).forEach { tc ->
        val s = Solution()
        println(s.solution(tc.first) == tc.second)
    }

}