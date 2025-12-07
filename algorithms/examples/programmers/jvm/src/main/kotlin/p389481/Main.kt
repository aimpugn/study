package p389481

class Solution {
    companion object {
        const val ALPHABET = "abcdefghijklmnopqrstuvwxyz"
    }

    /**
     * - 각 주문은 알파벳 소문자 11글자 이하로 구성
     * - 실제로 마법적 효과를 지니지 않는 의미 없는 주문들,
     *   즉, 알파벳 소문자 11글자 이하로 쓸 수 있는 모든 문자열이 고대의 규칙에 따라 아래와 같이 정렬됩니다.
     *     1. 글자 수가 적은 주문부터 먼저 기록된다.
     *     2. 글자 수가 같다면, 사전 순서대로 기록된다.
     *
     * 예를 들어,
     * "a" → "b" → "c" → "d" → "e" → "f" →...→ "z"
     * →"aa" → "ab"→...→"az" → "ba" →...→ "by" → "bz" → "ca"→...→"zz"
     * →"aaa" → "aab"→...→"aaz" → "aba" →...→ "azz" → "baa"→...→"zzz"
     * →"aaaa"→...→"aazz" → "abaa" →...→ "czzz" → "daaa" →...→ "zzzz"
     * →"aaaaa"→...
     *
     * 저주받은 주문들이 숨겨져 있었고, 이를 악용하려는 자들을 막기 위해 몇몇 주문을 주문서에서 삭제했습니다.
     * 삭제가 완료된 주문서에서 n번째 주문을 찾아내야 합니다.
     *
     * 예를 들어, 주문서에서 "d", "e", "bb", "aa", "ae" 5개의 주문을 지웠을 때,
     * 주문서에서 30번째 주문을 찾으려고 합니다.
     *
     * - 1~3번째 주문은 "a", "b", "c" 입니다.
     * - "d"와 "e"는 삭제됐으므로 4~24번째 주문은 "f" ~ "z"입니다.
     * - "aa"는 삭제됐으므로 25~27번째 주문은 "ab", "ac", "ad"입니다.
     * - "ae"는 삭제됐으므로 28~30번째 주문은 "af", "ag", "ah"입니다.
     *
     * 삭제된 주문 중 “bb”와 같이 n번째 주문보다 뒤에 위치해 있어서
     * n번째 주문을 찾는 데 영향을 주지 않는 주문도 존재할 수 있습니다.
     *
     * @param n
     * - 1 <= n <= 10^15
     * @param bans 삭제된 주문들을 담은 1차원 문자열 배열
     * - 1 <= bans.length <= 300,000
     * - 알파벳 소문자로만 이루어진 길이가 1 이상 11 이하인 문자열입니다.
     * - bans의 원소는 중복되지 않습니다.
     */
    fun solution(n: Long, bans: Array<String>): String {
        // 11자리 소문자 알파벳이 가능하므로 26^11 경우의 수가 가능합니다.
        // 26^8만으로도 208,827,064,576 이고,
        // 26^1 + 26^2 + ... + 26^11 ≈ 3.8 * 10^15 정도가 됩니다.
        // 이미 단순히 미리 정리해두고 탐색하는 것으로 문제를 해결할 수 있는 정도를 벗어납니다.
        //
        // 사전순으로 정렬되어 있다면 "원래" 순서를 예측할 수 있고,
        // 그 원래 순서의 위치에 있는 알파벳에서 삭제된 수만큼 뒤로 이동시키면 되지 않을까 싶습니다.
        //
        // 근데 이 경우 최악의 경우 300,000개가 영향을 줄 수 있는데...
        // 찾은 원래 문자열의 30만번째 두의 주문을 찾아야 하는 경우가 생길 수 있습니다.
        // 근데 원래 알파벳을 예측할 수 있다면, 찾은 문자열을 시작점으로 해서 다음 30만번째 알파벳도 예측할 수 있습니다.
        //
        // 그러면
        // 1. 원래 위치를 찾을 수 있는지 확인하고
        //    var originalSpell = findSpell(n)
        // 2. bans에서 실제로 영향을 주는 개수를 확인하고
        //    var affectedCnt =  = countAffectedCnt(originalSpell, bans)
        // 3. 'n + 영향을 주는 개수'의 위치에 있는 알파벳을 찾습니다.
        //    var targetSpell = findSpell(n + affectedCnt)
        //
        // 하지만 이 생각에 대한 중간 리뷰를 받아 보니, 다음과 같은 반례가 존재합니다:
        // 예를 들어,
        // a, b, c, d, e, f, g,... 에서 {c, d}를 삭제되었고 n=3이면,
        // a, b, e, f, g,...에서 n=3은 "e"가 되어야 합니다.
        // -> 근데 만약 위 로직을 따르게 되면 원래 n=3번째는 "c"이고,
        // -> 영향을 주는 bans는 "c" 한 개라고 착각하게 되고,
        // -> n + 1 = 4가 되어 "d"를 찾게 됩니다.
        //
        // 구체적으로 알파벳을 쭉 나열해보면 다음과 같은 패턴을 갖습니다:
        //  1
        // [a, b, c, d, ..., z]: 26개
        //  27  28  29  30  31  32  33  34       52
        // [aa, ab, ac, ad, ae, af, ag, ah, ..., az]: 26개
        //  53  54  55
        // [ba, bb, bc, ..., bz]: 26개
        //                  702
        // [.................zz]: 26개
        // ...
        //  703
        // [aaa, aab, aac, ..., aaz]: 26개
        // ...
        // [zza, zzb, zzc, ..., zzz]: 26개
        //
        // 26진법 수를 다루는 것과 비슷하게 느껴집니다.
        //
        // 그러면 아래와 같이 1부터 N까지 알파벳에 대한 인덱스를 맵핑할 수 있습니다:
        // a = 26^0 + 1 = 1
        // b = 26^0 + 2 = 2
        // ...
        // aa = (26^1 * 1) + 1 = 27
        // ab = (26^1 * 1) + 2 = 28
        // ...
        // ba = (26^1 * 2) + 1 = 53
        //

        val max = 11
        val accCntPerSpellLength = getAccCntPerSpellLength(max)

        val bannedSpellIndices = LongArray(bans.size)
        for (i in bans.indices) {
            bannedSpellIndices[i] = toIdx(accCntPerSpellLength, bans[i])
        }

        // 영향을 주는 ban spell인지 체크하기 위해 정렬합니다.
        bannedSpellIndices.sort()

        // n = 3
        // bans = [c, d]
        //
        // a b c d e f g h ...
        //
        // c는 삭제되었고 영향을 주므로 target(=3)은 한 칸 뒤로 밀려서 4가 됩니다.
        // d는 삭제되었고 영향을 주므로 target(=4)은 한 칸 뒤로 밀려서 5가 됩니다.
        // 더이상 인덱스에 영향을 줄 삭제된 주문 없으므로 5를 리턴합니다.
        var target = n
        for (bannedSpellIdx in bannedSpellIndices) {
            if (target < bannedSpellIdx) {
                // 삭제된 주문의 인덱스(bannedSpellIdx)가 target 보다 크다는 것은
                // 찾으려는 주문의 인덱스에 더이상 영향을 주지 않는다는 것을 의미합니다.
                // 찾으려는 주문 이후에 위치하므로 종료합니다.
                break
            }
            // 삭제된 주문의 인덱스(bannedSpellIdx)가 target 보다 작다는 것은
            // 찾으려는 주문의 인덱스에 영향을 준다는 것을 의미합니다.
            // 삭제가 없을 때는 n번째이지만,
            // 삭제가 있으므로 n번째에 n+1번째 주문이 오게 됩니다.
            // 따라서 target을 +1 증가시킵니다.
            target++
        }

        return findSpell(accCntPerSpellLength, 11, target)
    }

    fun getAccCntPerSpellLength(max: Int): LongArray {
        // 3.8 * 10^15를 고려하여 Long을 사용합니다.
        val pow26Arr = LongArray(max + 1)
        pow26Arr[0] = 1
        for (i in 1..max) {
            pow26Arr[i] = pow26Arr[i - 1] * 26L
        }

        // 주문 길이별로 누적 주문 건수를 기록합니다.
        // 1자리까지 주문 = 26
        // 2자리까지 주문 = 26 + 26^2
        // 3자리까지 주문 = 26 + 26^2 + 26^3
        val accCntPerSpellLength = LongArray(max + 1)
        accCntPerSpellLength[0] = 0L
        for (len in 1..max) {
            accCntPerSpellLength[len] = accCntPerSpellLength[len - 1] + pow26Arr[len]
        }

        return accCntPerSpellLength
    }

    fun findSpell(accCntPerSpellLength: LongArray, max: Int, n: Long): String {
        // 찾으려는 n이 속한 문자열 길이, 즉:
        // accCntPerSpellLength[len - 1] < idx <= accCntPerSpellLength[len]
        // 조건에 해당하는 idx를 찾습니다.
        var len = 1
        while (len < max && accCntPerSpellLength[len] < n) {
            len++
        }
        // 주문 길이가 len인 집합에서 n이 몇 번째에 위치하는지 찾습니다.
        // 가령 n=10인 경우, 한 자리 알파벳(len=1)이므로
        // 10 - 0 = 10으로 "j"가 됩니다.
        //
        // 이때 26진수인 점을 고려하여 '-1'을 합니다.
        // 가령,
        // - 2진수의 경우 0, 1 두 가지의 값으로 표현됩니다.
        // - 10진수의 경우 0, 1, 2, ..., 9 열 가지의 값으로 표현됩니다.
        //
        // 그렇다면 마찬가지로 26진수의 경우에는 0, 1, 2, ..., 25 스물 다섯 가지의 값으로 표현합니다.
        var pos0Based = n - accCntPerSpellLength[len - 1] - 1L

        val chars = CharArray(len)
        for (i in len - 1 downTo 0) {
            chars[i] = ALPHABET[(pos0Based % 26L).toInt()]
            pos0Based /= 26L
        }

        return String(chars)
    }

    fun toIdx(accCntPerSpellLength: LongArray, spell: String): Long {
        // 길이 < len인 모든 문자열 개수
        val accCntBeforeLen = accCntPerSpellLength[spell.length - 1]

        // 주문 길이가 len인 집합에서 몇 번째에 위치하는지 구합니다.
        // 알파벳을 char로 계산하기 편하게 0부터 계산하고, 나중에 +1을 합니다.
        var idxInSpellGroup = 0L
        // xyz 이면 x, y, z 순서로 순회합니다.
        // ba이면,
        // (0 * 26L) + ('b' - 'a') = 1
        // (1 * 26L) + ('a' - 'a') = 26
        //
        // aaa이면
        // (0 * 26L) + ('a' - 'a') = 0
        for (ch in spell) {
            idxInSpellGroup = idxInSpellGroup * 26L + (ch - 'a')
        }

        // idxInSpellGroup가 0부터 시작했으므로 1을 더합니다.
        return accCntBeforeLen + idxInSpellGroup + 1
    }
}

fun main() {
    val s = Solution()

    val max = 11;
    val accCntPerSpellLength = s.getAccCntPerSpellLength(max)

    assert(s.findSpell(accCntPerSpellLength, max, 5) == "e")
    assert(s.findSpell(accCntPerSpellLength, max, 27) == "aa")
    assert(s.findSpell(accCntPerSpellLength, max, 53) == "ba")
    assert(s.findSpell(accCntPerSpellLength, max, 703) == "aaa")

    assert(s.toIdx(accCntPerSpellLength, "a") == 1L)
    assert(s.toIdx(accCntPerSpellLength, "aa") == 27L)
    assert(s.toIdx(accCntPerSpellLength, "ba") == 53L)
    assert(s.toIdx(accCntPerSpellLength, "aaa") == 703L)

    listOf(
        Triple(
            30L,
            arrayOf("d", "e", "bb", "aa", "ae"),
            "ah"
        ),
        Triple(
            3L,
            arrayOf("c", "d"),
            "e"
        )

    ).forEachIndexed { idx, tc ->
        val answer = s.solution(tc.first, tc.second)
        println("[tc:$idx] answer: $answer")
        assert(tc.third == answer)
    }
}