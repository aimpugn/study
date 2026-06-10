package longestsubstring;

/**
 * https://leetcode.com/problems/longest-substring-without-repeating-characters/
 * <p>
 * 1바퀴 · 유형 공개: 슬라이딩 윈도우. 코드 전에 풀이 프로토콜(메서드 안 ①~④)을 거친다.
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md로.
 */
class Solution {
    /**
     * 같은 문자가 두 번 나오지 않는 가장 긴 부분 문자열(연속 구간)의 길이를 돌려준다.
     *
     * @param s 0 <= s.length <= 5 * 10^4, 영문·숫자·기호·공백
     */
    public int lengthOfLongestSubstring(String s) {
        // ① 손 실행: "pwwkew"에서 가장 긴 무중복 구간을 손으로 찾아 본다.
        //    찾는 동안 내 손이 구간을 언제 늘리고, 언제 무엇을 기준으로 줄이는지 관찰.
        // ② 불변식 문장: "오른쪽 끝을 i까지 늘린 직후, 윈도우는 항상 ____다."
        // ③ 반례 공격: 빈 문자열, 전부 같은 문자, 전부 다른 문자, … → main의 슬롯에 추가.
        // ④ 코드 = ②의 받아쓰기. 10분 안에 ②가 안 잡히면 브루트포스로 후퇴.
        return 0; // TODO: ②의 문장이 확정된 뒤에 작성
    }

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상 — 풀면 초록이 된다.
        check(s.lengthOfLongestSubstring("abcabcbb"), 3);
        check(s.lengthOfLongestSubstring("bbbbb"), 1);
        check(s.lengthOfLongestSubstring("pwwkew"), 3);
        // ③에서 직접 만든 반례를 여기에 추가:
    }

    private static void check(int actual, int expected) {
        if (actual != expected) {
            throw new AssertionError("expected=" + expected + ", actual=" + actual);
        }
        System.out.println("PASS " + actual);
    }
}
