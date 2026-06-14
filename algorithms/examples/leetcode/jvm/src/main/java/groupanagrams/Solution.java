package groupanagrams;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

/**
 * https://leetcode.com/problems/group-anagrams/
 * <p>
 * 1바퀴, 유형 공개: 해시. 1차 풀이(쌍 비교)가 TLE(Time Limit Exceeded, 시간 초과 - 128개
 * 테스트 중 127개만 통과)에 막혀 표준형 키로 재작성했습니다.
 * 1차 풀이 원본과 "무엇이 무엇으로, 왜 바뀌었는지"는 파일 하단 보존 블록에 있습니다.
 * 회고, 복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md로.
 */
class Solution {
    /**
     * 문자열 배열을 받아 애너그램끼리 묶은 그룹들을 돌려줍니다. 그룹 순서, 그룹 안 순서는 자유입니다.
     *
     * @param strs 영어 소문자 문자열로 이뤄진 배열
     * - 1 <= strs.length <= 10^4
     * - 0 <= strs[i].length <= 100
     *
     * <p>핵심 한 줄: 단어를 표준형(글자를 정렬한 문자열)으로 바꿔 맵의 키로 쓰면, "같은
     * 그룹인가?"라는 쌍 비교가 "키가 같은가?"라는 조회 한 번으로 바뀝니다.
     */
    public List<List<String>> groupAnagrams(String[] strs) {
        // > 불변식: "i번째 단어를 처리한 직후, 맵에는 (표준형 키 -> 그 키의 단어 전부)가 들어 있다."
        //
        // 아래 루프는 이 문장의 받아쓰기입니다. 예제를 흘려 보내면 맵이 이렇게 자랍니다
        // (바뀐 그룹만 적고 나머지는 ...로 줄입니다):
        //
        //   단어       키           맵의 변화                       일어난 일
        //   "eat"  ->  "aet"  ->  {"aet"=[eat]}                   새 그룹 생성
        //   "tea"  ->  "aet"  ->  {"aet"=[eat, tea]}              기존 그룹 합류 - 비교 없이
        //   "tan"  ->  "ant"  ->  {..., "ant"=[tan]}              새 그룹 생성
        //   "ate"  ->  "aet"  ->  {"aet"=[eat, tea, ate], ...}    기존 그룹 합류
        //   "nat"  ->  "ant"  ->  {..., "ant"=[tan, nat]}         기존 그룹 합류
        //   "bat"  ->  "abt"  ->  {..., "abt"=[bat]}              새 그룹 생성
        //
        //   최종 상태: {"aet"=[eat, tea, ate], "ant"=[tan, nat], "abt"=[bat]}
        //   맵의 값들이 곧 답입니다: [[eat, tea, ate], [tan, nat], [bat]]
        var groups = new HashMap<String, List<String>>();

        for (var word : strs) {
            // > 표준형 = 글자를 정렬한 문자열. 애너그램이면 정렬 결과가 같고, 아니면 다르다.
            //
            // 애너그램은 글자 구성이 같고 순서만 다른 단어라, 정렬하면 모두 한 문자열로 모입니다.
            //   [e,a,t] -> [a,e,t],  [t,e,a] -> [a,e,t]  =>  둘 다 "aet"
            // char[]로 복사해 정렬한 뒤 다시 String을 만드는 건, String이 불변이라 제자리
            // 정렬이 안 되고, char[]는 내용이 같아도 참조로만 비교돼 맵 키 구실을 못 하기 때문입니다.
            //
            // 비용은 단어당 O(L log L)입니다(L = 단어 길이 <= 100). "정렬은 N log N"에는 두
            // 층위가 있는데 둘 다 닫고 가겠습니다.
            //
            // (1) 수학적 하한 - 비교로 순서를 가리는 정렬은 어떤 방법이든 N log N보다 빠를 수
            // 없습니다. 결정 트리로 보면 왜인지 닫힙니다. 정렬이란 N개를 어떤 순서로 늘어놓을지
            // 가리는 일이고, 가능한 순서는 N!가지입니다. 비교 한 번("a가 b보다 작나?")의 답은
            // 예/아니오 둘뿐이라 후보 순서를 한 번에 최대 절반까지만 줄입니다. N!개를 절반씩
            // 줄여 1개로 좁히려면 최소 log2(N!)번 비교해야 하고 - 이게 하한입니다 - 그 log2(N!)이
            // 약 N log N입니다(스털링 근사: N!은 대략 (N/e)^N이라, 로그를 취하면 N log N에서 낮은
            // 차수 항을 뺀 꼴이고, 지배하는 큰 항이 N log N). 그래서 N log N은 "이 구현이 그렇다"가
            // 아니라 "비교 정렬이면 수학적으로 그렇다"입니다(쉽게 배우는 알고리즘 4장, CLRS 8장).
            //
            // (2) 이 구현 - 자바 Arrays.sort(char[])는 primitive 배열이라 dual-pivot 퀵정렬(피벗
            // 2개로 나눠 정복하는 자바 기본 정렬)을 씁니다. 평균 N log N, 최악 N^2입니다(피벗이
            // 계속 한쪽으로 치우쳐 분할이 1:나머지로만 쪼개지면 가르는 단계가 N번까지 늘기 때문).
            // 객체 배열이면 TimSort(병합 정렬 변형)라 최악도 N log N 보장이지만, 우리 건 char[]라
            // 퀵정렬입니다. 여기선 N = L <= 100이라 최악 100^2 = 만 단위로도 무시되어 평균
            // L log L로 봅니다.
            //
            // 곱 구조의 직관(N log N이 왜 곱인가): "매번 절반씩 가르는" 균형 분할정복을 그려
            // 보면 됩니다(병합 정렬, 또는 운 좋게 피벗이 가운데 떨어진 퀵정렬). N개를 절반씩
            // 가르면 가르는 깊이가 log2 N(약 7단계)이고, 각 깊이에서 N개를 한 번씩 훑어 자리를
            // 맞추니 log2 N 곱하기 N입니다. 퀵정렬은 이 절반 분할이 보장되진 않지만(그래서 위처럼
            // 최악 N^2) 평균적으로는 이 균형에 가까워 N log N입니다. 여기서 정렬되는 N은 단어
            // 목록(n개)이 아니라 한 단어의 글자(L개)라, 단어당 L log L = 약 100 * 7 = 700번입니다.
            //
            // 더 줄이려면 26칸 등장 횟수 배열을 키로 쓰는 길도 있습니다(정렬 없이 세기만 하니
            // 단어당 O(L)). 단 int[26] 자체는 위에서 char[]가 키가 못 된 것과 똑같은 이유(배열은
            // 내용이 같아도 참조로만 비교됨)로 키가 못 되니, 26칸을 "1#0#2#..." 같은 문자열로
            // 이어붙여 키로 씁니다.
            var letters = word.toCharArray();
            Arrays.sort(letters);
            var canonical = new String(letters);

            // 키가 처음이면 빈 그룹을 만들고, 이미 있으면 그 그룹에 합류시킵니다. 1차 풀이가
            // "배열 전체를 다시 훑으며 같은 그룹 찾기"로 하던 일이 이 조회 한 번으로 끝납니다.
            //
            // 이 조회, 삽입은 평균 O(1)입니다. HashMap은 키의 해시값으로 저장 위치(버킷)를
            // 계산해 바로 찾아가니, 들어 있는 그룹이 몇 개든 단번에 도착합니다. 그런데 왜
            // "보장"이 아니라 "평균"일까요 - 서로 다른 키가 같은 버킷에 떨어지는 충돌(collision)
            // 때문입니다. 한 버킷에 여러 키가 쌓이면 그 안을 차례로 훑어야 하고, 운 나쁘게 다
            // 같은 버킷에 몰린 최악에는 O(n)까지 늘어집니다. 다만 해시 함수가 키를 버킷에 고르게
            // 흩뿌리면 한 버킷의 줄이 짧게 유지되어, 평균적으로는 O(1)에 가깝습니다. 그래서
            // 보장이 아니라 평균입니다(쉽게 배우는 알고리즘 7장 해시 테이블 p.223~).
            groups.computeIfAbsent(canonical, key -> new ArrayList<>()).add(word);
        }

        return new ArrayList<>(groups.values());
        // 시간 복잡도는 O(n * L log L)입니다 - "단어 수 n, 곱하기 글자 수 L, 곱하기 로그 L"로
        // 읽습니다(변수 정의: n = 단어 개수 <= 10^4, L = 한 단어의 글자 수 <= 100). O(...)는
        // 입력이 커질 때 비용이 자라는 '모양'만 남기고 상수를 버린 표기이고, log L은 정렬이
        // 배열을 반씩 나누며 일해서 생기는 항입니다 - L을 반으로 log2 L번 나누면 1이 되고,
        // 100은 7번이면 충분합니다(100->50->25->12->6->3->1).
        // 유도: 바깥 루프가 단어 수 n번 돌고, 반복마다 가장 비싼 일이 길이 L짜리 글자 배열의
        // 정렬 O(L log L)이며, 맵 조회, 삽입은 평균 O(1)이라 정렬 비용에 흡수됩니다. 곱하면
        // 10^4 * 100 * 7 = 약 7*10^6번의 기본 연산이고, 자바가 초당 ~10^8번쯤 처리하니 수십
        // ms입니다. 그리고 이 식은 O(n * n log n)이 아닙니다 - 루프 횟수는 n이 맞지만, 반복
        // 안에서 정렬되는 것은 단어 목록(n개)이 아니라 한 단어의 글자들(L개)입니다. 정렬
        // 비용의 변수가 무엇인지 물을 때는 "무엇이 정렬되는가"부터 확인해야 합니다.
        //
        // 공간 복잡도는 O(n * L)입니다. 맵의 값 리스트들이 모든 단어를 정확히 한 번씩 담고
        // (합쳐서 단어 n개 * 길이 L), 그룹마다 길이 L짜리 키 문자열이 하나씩 있기 때문입니다.
        // 루프 안의 char[] 복사본도 단어당 L이지만 반복이 끝나면 버려져 동시에 하나만
        // 존재합니다 - 누적되지 않는 O(L) 추가분입니다.
        // 1차 풀이는 시간 O(G * n * L), 최악(G = n)에서 수 초였습니다 - 비교는 아래 보존 블록 참조.
    }

    /* 1차 풀이 보존 (TLE, 127/128). 여기서 볼 것은 코드 대비가 아니라 사고의 갈림길입니다.

       어디서 갈렸나. 1차 풀이는 애너그램을 "두 단어가 같은 부류인가?"라는 관계로 봤고,
       관계로 보면 모든 쌍을 맞대보는 이중 루프가 자연히 따라옵니다. 정답 사고는 애너그램을
       "각 단어 하나가 가진 표준형(정렬 결과)"이라는 속성으로 보고, 속성으로 보면 그 표준형을
       키 삼아 묶는 것이 따라옵니다. 관계로 보면 비교가, 속성으로 보면 키가 나옵니다.

       그런데 정답을 이미 스쳤습니다. isAnagram 안에 "차라리 정렬하는 게 낫나?"라고 적어
       두셨는데, 그 정렬이 바로 표준형입니다. 다만 그 정렬을 "두 단어를 맞대보는 도구"로만
       쓰고 "한 단어를 대표하는 키"로 승격시키지 못했습니다. 도구에서 키로 한 걸음이었습니다.

       그 한 걸음을 놓친 까닭은 문제를 관계로 본 프레임입니다. 그래서 다음에 잡을 신호 하나 -
       "A와 B가 같은 부류인가?"라는 관계 질문이 보이면 멈추고 물어보세요. "각 원소를 하나의
       대표값으로 바꿔서, 그 값으로 묶을 수 있나?" 묶이면 쌍 비교 O(n^2)가 키 조회 O(n)로
       내려갑니다.

       > 같은 부류는 둘씩 맞대보지 말고, 각자의 대표값(키)으로 묶어라.

       이 사고를 코드로 보면 - 같은 "tea"가 자기 그룹을 찾는 동작:
         1차(관계->비교): "tea" <-> "eat"? <-> "tan"? <-> "ate"? <-> ...   O(G * n)
         새 풀이(속성->키): "tea" -> 정렬 -> "aet" -> groups["aet"] 직행    비교 0, 조회 1
       바뀐 자취: 이중 루프 + bitmap -> for 하나 + computeIfAbsent(쌍 비교도 "이미 묶였나"를
       세던 bitmap도 맵이 키로 알아서 소멸), isAnagram(카운트 맵 2개 맞대보기) -> 표준형 1개를
       키로(비교기가 키 생성기로), Integer != 지뢰(캐시 덕에 우연히만 정답)는 코드가 사라지며
       함께 소멸. 복잡도는 O(G * n * L)인데 - G개 그룹 대표가 각각 n개를 훑고(G * n번 비교),
       비교 한 번인 isAnagram이 길이 L 단어를 세느라 O(L)이라 L이 곱해집니다 - 최악(G=n)에서
       수 초였습니다. 새 풀이는 O(n * L log L) 약 10^7, 수십 ms.

       아래는 당시의 분석, 시행착오 주석과 검수 주석을 그대로 보존한 원본입니다.

    public List<List<String>> groupAnagrams(String[] strs) {
        // 1. ["eat","tea","tan","ate","nat","bat"]
        //    => [["bat"],["nat","tan"],["ate","eat","tea"]]
        //    "bat"으로 재배열 될 수 있는 문자열 없음.
        //    "nat", "tan"은 서로 재배열하여 만들 수 있는 애너그램
        //    "ate", "eat", "tea"는 서로 재배열하여 만들 수 있는 애너그램
        // 이름은 grouped나 visited가 더 정확합니다. 실체가 boolean[]라 비트맵(비트 단위 압축)이
        // 아니고, 역할도 "이미 그룹에 묶였는가"이기 때문입니다 - 이름이 역할을 말하면 읽는
        // 사람이 자료구조가 아니라 의도를 읽습니다.
        var bitmap = new boolean[strs.length];
        var answer = new ArrayList<List<String>>();

        // 꺼내서 다른 것과 비교하고, 애너그램이면 리스트에 넣습니다.
        // 그러면 현재 문자열이 있어야 하고, 비교할 대상도 있어야 합니다.
        // 하나의 자료 구조로 가능할지? 큐? pop해서 현재 요소를 꺼내서 다른 요소들과 비교해서 애너그램이면 같은 리스트에 넣어야 합니다.
        // 중첩 배열을 도는 게 더 직관적일 거 같습니다.
        // 어떤 애너그램 리스트에 들어갔으면 그 대상은 체크할 필요 없습니다. 어차피 애너그램이면 같은 리스트에 존재할 것이므로.

        // 위 판단은 정확합니다. 애너그램은 동치 관계라서(같은 그룹이면 서로 전부 애너그램)
        // 한 번 묶인 원소는 다시 볼 필요가 없습니다.
        //
        // 그런데 TLE의 원인이 바로 이 이중 루프입니다. isAnagram이 느려서가 아니라 "쌍 비교"라는
        // 구조 자체가 문제라서, 탈출구도 비교를 빠르게 만드는 것이 아니라 없애는 것입니다.
        //
        // 비용은 isAnagram 호출 O(G * n)번이고(G = 그룹 수), 실패 케이스는 G = n이라 최악입니다.
        // 그룹 대표마다 배열 전체를 다시 훑는 구조인데, 위의 가지치기는 그룹이 클 때만 일하고,
        // 실패 케이스는 서로 묶이는 단어가 거의 없는 입력이라 그룹 크기가 전부 1 - 가지치기가
        // 한 번도 일하지 못합니다. n <= 10^4이면 쌍 비교 약 5*10^7회에 무거운 호출 비용(아래
        // isAnagram 주석 참조)이 곱해져 자바 기준 수 초가 걸립니다. 최적화를 넣을 때는 그것이
        // 평균을 돕는지 최악을 돕는지를 채점자 입장에서 물어야 합니다.
        //
        // 비교를 없애는 길은 표준형입니다. 같은 그룹의 모든 단어는 변하지 않는 표준형(글자를
        // 정렬한 문자열, 또는 26칸 등장 횟수)을 공유하므로, 표준형을 맵의 키로 쓰면 "같은
        // 그룹인가?"라는 쌍 질문이 "키가 같은가?"라는 조회 한 번으로 바뀝니다. 비교로 찾지 말고
        // 키를 계산해 한 번에 도착하라는 해시의 핵심 사고입니다(쉽게 배우는 알고리즘 7장 p.223~,
        // CLRS 11장).
        for (var i = 0; i < strs.length; i++) {
            if (bitmap[i]) {
                continue;
            }
            var str1 = strs[i];
            var anagrams = new ArrayList<String>();
            bitmap[i] = true;
            anagrams.add(str1);

            for (var j = 0; j < strs.length; j++) {
                if (!bitmap[j] && isAnagram(str1, strs[j])) {
                    anagrams.add(strs[j]);
                    // 이미 처리했다고 표시합니다.
                    bitmap[j] = true;
                }
            }

            answer.add(anagrams);
        }

        // System.out.println(answer);

        return answer;
    }

    // 답은 이미 이 메서드 안에 있었습니다. 여기서 만드는 카운트 맵이 곧 그 문자열의
    // 표준형이라서, 두 개 만들어 '비교'하는 대신 하나를 그룹의 '키'로 쓰면 이 메서드 전체가
    // "단어 -> 키" 변환 하나로 줄어듭니다. 본문의 자기 질문("차라리 정렬하는 게 낫나?")도
    // 같은 정답 방향이었습니다 - 정렬 결과 역시 표준형이니까요. 알파벳이 26개로 고정이면
    // 비교하는 대신 센다는 계수 정렬의 사고와 같은 축입니다(쉽게 배우는 알고리즘 4장 특수
    // 정렬 p.120~).
    //
    // 호출 한 번의 비용도 무겁습니다. 길이가 같은 쌍마다 HashMap 2개를 만들고 문자당
    // 박싱(Character/Integer) 연산을 약 2L회 치르는데, 이것이 바깥 루프의 호출 횟수와
    // 곱해지며 TLE의 상수를 키웠습니다. 쌍 비교가 정말 필요한 자리라면 HashMap 대신
    // int[26] 두 개를 채워 Arrays.equals로 비교하는 쪽이 박싱 없이 O(L)입니다.
    private boolean isAnagram(String str1, String str2) {
        if (str1.length() != str2.length()) return false;

        // cca caa => 서로 달라야 함
        // 해당 문자의 등장 횟수가 동일해야 함
        // 차라리 정렬하는 게 낫나? 아니면 그냥 맵을 두 개 만들어서 비교하는 게 낫나?
        var map1 = new HashMap<Character, Integer>();
        for (var ch : str1.toCharArray()) {
            map1.compute(ch, (k, v) -> v == null ? 1 : v + 1);
        }

        var map2 = new HashMap<Character, Integer>();
        for (var ch : str2.toCharArray()) {
            map2.compute(ch, (k, v) -> v == null ? 1 : v + 1);
        }

        for (var entry : map1.entrySet()) {
            if (!map2.containsKey(entry.getKey())) {
                return false;
            }
            // 이 비교는 지금 우연히만 맞습니다. 박싱된 Integer를 !=로 비교하면 값이 아니라 참조
            // (객체의 메모리 주소)를 비교합니다. 서로 다른 곳에서 따로 센 두 카운트가 어떻게
            // 같은 객체가 될 수 있냐면, 오토박싱 때문입니다. int를 Integer로 자동 변환할 때 자바는
            // 속으로 Integer.valueOf를 부르고, 이 메서드는 -128~127 범위의 값에 대해서는 미리
            // 만들어 둔 캐시 객체를 매번 똑같이 돌려줍니다. 그래서 두 곳에서 독립적으로 센 "3"이
            // 둘 다 동일한 캐시 인스턴스가 되어, 참조 비교(!=)가 "다르지 않다"로 우연히 통과합니다.
            // 등장 횟수가 L <= 100이라 늘 이 캐시 범위 안이라 안 들킬 뿐, 제약이 1000이었다면
            // 128부터는 매번 새 객체라 조용히 틀렸을 겁니다. 이미 손에 쥔 entry.getValue()를 써서
            // .equals(map2.get(...))로 비교하면 값 비교가 되고 불필요한 재조회도 사라집니다.
            if (map1.get(entry.getKey()) != map2.get(entry.getKey())) {
                return false;
            }
        }

        return true;
    }

    - 1차 풀이 보존 끝 */

    static void main() {
        var s = new Solution();
        check(s.groupAnagrams(new String[]{"eat", "tea", "tan", "ate", "nat", "bat"}),
            List.of(List.of("bat"), List.of("nat", "tan"), List.of("ate", "eat", "tea")));
        check(s.groupAnagrams(new String[]{""}), List.of(List.of("")));
        check(s.groupAnagrams(new String[]{"a"}), List.of(List.of("a")));
        // 길이가 같아도 애너그램이 아니면 다른 그룹이어야 합니다 - 정렬 키가 다름을 검증.
        check(s.groupAnagrams(new String[]{"abc", "bca", "xyz"}),
            List.of(List.of("abc", "bca"), List.of("xyz")));
        // 같은 단어가 중복돼도 한 그룹에 모두 들어가야 합니다.
        check(s.groupAnagrams(new String[]{"aa", "aa"}), List.of(List.of("aa", "aa")));

        // TLE를 일으켰던 적대 형태의 실측 검증입니다 - 단일 문자를 1~100번 반복해 길이가
        // 100종으로 흩어진 단어 1만 개("a", "bb", "ccc", ...). 같은 길이 안에서는 글자가 달라
        // 대부분 서로 비-애너그램이라 그룹이 잘게 쪼개집니다. 1차 풀이라면 길이가 다른 쌍은
        // 길이 가드로 즉시 걸러지지만, 같은 길이 쌍(약 10^6쌍)마다 무거운 isAnagram 본문이
        // 돌아 수 초가 걸렸을 형태입니다. 표준형 키는 단어마다 한 번 정렬해 키로 꽂으니 순회
        // 한 번, 수십 ms 안에 끝나야 합니다.
        var adversarial = new String[10_000];
        for (var i = 0; i < adversarial.length; i++) {
            adversarial[i] = String.valueOf((char) ('a' + i % 26)).repeat(i % 100 + 1);
        }
        var kinds = new HashSet<String>(); // 기대 그룹 수 = 서로 다른 (글자, 길이) 조합의 수
        for (var word : adversarial) {
            kinds.add(word.charAt(0) + ":" + word.length());
        }
        var started = System.nanoTime();
        var groupCount = s.groupAnagrams(adversarial).size();
        var elapsedMs = (System.nanoTime() - started) / 1_000_000;
        if (groupCount != kinds.size()) {
            throw new AssertionError("expected groups=" + kinds.size() + ", actual=" + groupCount);
        }
        System.out.println("PASS 적대 케이스 " + groupCount + "그룹, " + elapsedMs + "ms");
    }

    /**
     * 그룹 순서와 그룹 안 순서는 채점 대상이 아니므로, 각 그룹을 정렬해 집합으로 비교합니다.
     */
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
