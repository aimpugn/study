package groupanagrams;

import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

/**
 * https://leetcode.com/problems/group-anagrams/
 * <p>
 * 1바퀴 · 유형 공개: 해시. 코드 전에 풀이 프로토콜(메서드 안 ①~④)을 거친다.
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md로.
 */
class Solution {
    /**
     * 문자열 배열을 받아 애너그램끼리 묶은 그룹들을 돌려준다. 그룹 순서, 그룹 안 순서는 자유.
     *
     * @param strs 1 <= strs.length <= 10^4, 0 <= strs[i].length <= 100, 영어 소문자
     */
    public List<List<String>> groupAnagrams(String[] strs) {
        // ① 손 실행: ["eat","tea","tan","ate","nat","bat"]를 코드 생각 없이 손으로 묶어 본다.
        //    묶는 동안 내 손이 '무엇을 보고' 같은 그룹이라고 판단하는지 관찰.
        // ② 불변식 문장: "i번째 단어를 처리한 직후, ____는 항상 ____다."
        // ③ 반례 공격: 빈 문자열, 한 글자, 전부 같은 단어, 애너그램이 하나도 없는 입력 → main의 슬롯에 추가.
        // ④ 코드 = ②의 받아쓰기. 10분 안에 ②가 안 잡히면 브루트포스로 후퇴.
        return List.of(); // TODO: ②의 문장이 확정된 뒤에 작성
    }

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상 — 풀면 초록이 된다.
        check(s.groupAnagrams(new String[]{"eat", "tea", "tan", "ate", "nat", "bat"}),
                List.of(List.of("bat"), List.of("nat", "tan"), List.of("ate", "eat", "tea")));
        check(s.groupAnagrams(new String[]{""}), List.of(List.of("")));
        check(s.groupAnagrams(new String[]{"a"}), List.of(List.of("a")));
        // ③에서 직접 만든 반례를 여기에 추가:
    }

    /** 그룹 순서와 그룹 안 순서는 채점 대상이 아니므로, 각 그룹을 정렬해 집합으로 비교한다. */
    private static void check(List<List<String>> actual, List<List<String>> expected) {
        if (!normalize(actual).equals(normalize(expected))) {
            throw new AssertionError("expected=" + normalize(expected) + ", actual=" + normalize(actual));
        }
        System.out.println("PASS " + normalize(actual));
    }

    private static Set<List<String>> normalize(List<List<String>> groups) {
        return groups.stream().map(g -> g.stream().sorted().toList()).collect(Collectors.toSet());
    }
}
