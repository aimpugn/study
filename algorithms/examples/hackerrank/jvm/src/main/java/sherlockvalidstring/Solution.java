package sherlockvalidstring;

import support.Judge;

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
 *   두 종류면 한 글자만 지워 맞출 수 있는지 — (빈도 1짜리가 딱 하나) 또는 (딱 하나가 남들보다 정확히 1 큼).
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
    public static String isValid(String s) {
        // Write your code here
        return "";
    }
}
