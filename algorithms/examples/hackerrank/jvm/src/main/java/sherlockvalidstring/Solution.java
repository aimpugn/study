package sherlockvalidstring;

import support.Judge;

import java.util.HashMap;

/**
 * Sherlock and the Valid String — HackerRank (Interview Prep Kit, Dictionaries and Hashmaps).
 * <a href="https://www.hackerrank.com/challenges/sherlock-and-valid-string/problem">문제</a>
 * <p>
 * 문자열 s가 "valid"이면 "YES", 아니면 "NO"를 반환합니다. valid의 정의:
 * - 모든 문자의 등장 횟수가 같거나,
 * - 딱 한 글자(한 위치)를 지우면 남은 모든 문자의 등장 횟수가 같아지는 경우.
 * 예: "aabbc" -> c 하나 지우면 a=2,b=2 -> YES.  "aabbcd" -> 한 번 지워선 못 맞춤 -> NO.
 * <p>
 * 힌트(막히면): 각 문자의 빈도를 센 뒤, 그 "빈도들의 빈도"를 봐라. 빈도가 한 종류면 valid.
 * 두 종류면 한 글자만 지워 맞출 수 있는지 — (빈도 1짜리가 딱 하나) 또는 (딱 하나가 남들보다 정확히 1 큼).
 */
class Solution {
    static void main() {
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(Result.isValid("aabbcd"), "NO");     // a2 b2 c1 d1 - 한 번 지워선 못 맞춤
        Judge.check(Result.isValid("aabbc"), "YES");     // a2 b2 c1 - c 하나 지우면 다 2
        Judge.check(Result.isValid("aaaabbbb"), "YES");  // a4 b4 - 이미 다 같음
        Judge.check(Result.isValid("aaaabbcc"), "NO");   // a4 b2 c2 - 한 번 지워선 못 맞춤
        // 반례를 여기에 추가하세요:
    }
}

class Result {

    /*
     * Complete the 'isValid' function below.
     *
     * The function is expected to return a STRING.
     * The function accepts STRING s as parameter.
     *
     * Sherlock considers a string to be valid if
     * - all characters of the string appear the same number of times.
     * - It is also valid if remove just 1 character at 1 index in the string,
     *   and the remaining characters will occur the same number of times.
     *
     * @param s
     * - 1 <= s.length <= 100,000
     * - s[i] = ASCII a-z
     *
     * @return if s is valid "YES" or "NO"
     */
    public static String isValid(String s) {
        // Write your code here
        // abc -> valid. {1, 1, 1}
        // abcc -> valid. because remove `c`(at 2 index). {1, 1, 2}
        // abccc -> invalid {1, 1, 3}
        // aabbcd -> invalid {2, 2, 1, 1}
        // aabbccddeefghi -> invalid. {2, 2, 2, 2, 2, 1, 1, 1, 1}
        // abcdefghhgfedecba -> {2, 2, 2, 2, 3, 2, 2, 2}
        //
        // all same || all same except one case which can be same if decrease 1
        if (s == null || s.length() == 0) {
            return "NO";
        }
        if (s.isBlank()) {
            return "YES";
        }

        // arrange String into Map<Character, number of characters>
        var numberOfChars = new HashMap<Character, Integer>();
        for (var ch : s.toCharArray()) {        // O(n): 문자열을 한 번 훑어 문자별 빈도 (n = 길이)
            numberOfChars.merge(ch, 1, Integer::sum);
        }
        // System.out.println(numberOfChars);
        // and Map<number of characters, count>
        var countOfNumberOfChars = new HashMap<Integer, Integer>();
        for (var numberOfChar : numberOfChars.entrySet()) {  // O(σ): 서로 다른 문자 수만큼 (σ=distinct char, a-z라 <=26 -> O(1))
            var number = numberOfChar.getValue();
            countOfNumberOfChars.merge(number, 1, Integer::sum);
        }
        // aabbcd: {1=2, 2=2} -> NO
        // aabbccddeefghi: {1=4, 2=5} => NO
        // System.out.println(countOfNumberOfChars);

        var size = countOfNumberOfChars.size();
        // if 1 == Map<number of characters, count>.size => YES
        if (size == 1) {
            return "YES";
        }
        // 첨삭: 정답입니다 — 완전하고 깔끔합니다(brute force로 15,300개 문자열 전수 대조, 불일치 0).
        //   "문자 빈도 -> 그 빈도들의 빈도"로 줄인 게 이 문제의 씨앗이고, 그 뒤 판정도 맞습니다.
        //   막판에 추가해 통과시킨 (A) 조건(val2==1 && count(val2)==1)이 놓치기 딱 좋은 자리라, 왜 그런지 짚어 둡니다.
        //
        //   빈도의 빈도(countOfNumberOfChars)로 줄이면 판정이 size로 세 갈래로 떨어집니다:
        //   - size 1: 모든 문자가 같은 횟수 -> 손 안 대도 valid -> YES
        //   - size 2: 딱 한 번 지워서 균일하게 만들 수 있나? 그 방법이 정확히 둘뿐 (아래 A, B)
        //   - size 3+: 빈도 뭉치가 셋 이상 -> 한 번 지워선 절대 하나로 못 합침 -> NO
        //
        //   size 2에서 "한 번 지우기"로 균일해지는 형태는 딱 두 가지 — 이걸 빠짐없이 나눠 보는 게 요점:
        //   - (A) 작은 빈도가 1이고, 그 1을 가진 글자가 딱 하나  -> 그 글자를 통째로 삭제(빈도 1 -> 0, 사라짐)
        //         남은 건 다 큰 빈도라 균일.   코드: val2 == 1 && count(val2) == 1
        //   - (B) 큰 빈도가 작은 빈도보다 정확히 1 크고, 그 큰 빈도 글자가 딱 하나 -> 그 글자에서 한 개만 삭제
        //         (큰 빈도 -> 큰 빈도 - 1 = 작은 빈도) 라 균일.   코드: count(val1) == 1 && val1 - 1 == val2
        //
        //   놓칠 뻔한 이유(간극): A와 B 둘 다 "한 번 지우기"지만 성격이 다릅니다. B는 "많은 놈 카운트를 1 깎기"
        //   (수를 조정), A는 "딱 한 번 나온 희귀 글자를 통째로 지우기"(글자를 삭제). '카운트 조정'(B)은 자연히
        //   떠오르는데 '희귀 글자 삭제'(A)는 관점이 달라 빠지기 쉽습니다. 신호로 고정: 빈도의 빈도가 size 2면
        //   "한 번 제거로 균일해지는 형태"를 A(하나 통째 삭제) / B(하나만 감소) 둘로 나눠 둘 다 확인하라.
        //
        //   [맵이 자라는 모습] 같은 입력이 두 맵을 거쳐 판정으로 이어지는 실제 상태:
        //     "aabbc"      빈도{a=2,b=2,c=1}      빈도의빈도{2=2, 1=1}  -> size2, (A) c 통째삭제  -> YES
        //     "aabbccddd"  빈도{a=2,b=2,c=2,d=3}  빈도의빈도{2=3, 3=1}  -> size2, (B) d 하나삭제   -> YES
        //     "aabbcd"     빈도{a=2,b=2,c=1,d=1}  빈도의빈도{2=2, 1=2}  -> size2, A아님 B아님      -> NO
        //   (aabbcd가 NO인 까닭: 빈도 1인 글자가 c,d 둘이라 count(1)=2. 하나 지워도 나머지 하나가 1로 남음.)
        //
        //   [최종 복잡도] 빈도 맵 O(n) + 빈도의빈도 O(σ) + 판정 O(1) = O(n) 시간 (n = 문자열 길이 <= 100,000).
        //   공간 O(σ) = O(1) (σ=distinct char, a-z뿐이라 두 맵 합쳐도 최대 26칸). n=10^5 제약이라 O(n)이면 넉넉.
        //
        // > 카드: 문자 빈도가 다 같은지 물으면 -> "빈도의 빈도"로 줄여라. size 1=YES, size 3+=NO,
        // >       size 2=한 번 제거 두 형태 확인 (A: 빈도1짜리가 하나면 통째삭제 / B: 최대가 +1이고 하나면 하나감소).
        //
        // if 2 == Map<number of characters, count>.size:
        // {2=7, 3=1}:
        // {2=2, 1=2}: aabc
        if (size == 2) {
            // 1 < Math.abs(values[0] - values[1]) ? NO : YES
            var keys = countOfNumberOfChars.keySet().toArray(new Integer[0]);
            var val1 = Math.max(keys[0], keys[1]);
            var val2 = Math.min(keys[0], keys[1]);
            // if 2 - 4, then even though decrease 1, it can not be same
            // {3=2, 2=1} ? "aaabbbcc": invalid
            // {2=3, 3=1} ? "aabbccddd": valid
            // map[3] == 1 && (3 - 1 == 2)
            //
            // {2=2, 1=1} ? "aabbc": valid
            if (val2 == 1 && countOfNumberOfChars.get(val2) == 1) {
                return "YES";
            }
            return countOfNumberOfChars.get(val1) == 1 && (val1 - 1 == val2) ? "YES" : "NO";
        }
        // if 2 < Map<number of characters, count>.size => NO
        // {3=2, 2=1, 1=1}
        // aaabbbccd

        return "NO";
    }
}