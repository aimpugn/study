package longestsubstring;

import java.util.HashMap;
import java.util.HashSet;

/**
 * https://leetcode.com/problems/longest-substring-without-repeating-characters/
 * <p>
 * 1바퀴, 유형 공개: 슬라이딩 윈도우. 코드 전에 풀이 프로토콜(메서드 안 1~4번)을 거칩니다.
 * 회고, 복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md로.
 * <p>
 * 이 파일에는 두 풀이가 같이 있습니다. 제출용은 슬라이딩 윈도우(lengthOfLongestSubstring),
 * 비교용 보존은 1차 풀이(lengthOfLongestSubstringByScan)입니다. main이 둘을 같은 입력에 돌려
 * 정답이 일치하는지 확인하고, 입력을 키우며 시간을 재서 "윈도우가 더 낫다"를 실측으로 보여줍니다.
 */
class Solution {
    /**
     * 같은 문자가 두 번 나오지 않는 가장 긴 부분 문자열(연속 구간)의 길이를 돌려줍니다.
     *
     * @param s 문자열
     * - 0 <= s.length <= 5 * 10^4
     * - 영문, 숫자, 기호, 공백
     */
    public int lengthOfLongestSubstring(String s) {
        // 슬라이딩 윈도우 O(n)입니다. 핵심은 하나 - 같은 문자를 다시 보지 않으려고 윈도우를 버리고
        // 새로 만드는 게 아니라, 왼쪽 끝만 중복 다음으로 점프시키며 양 끝을 밀어 나갑니다. 정답인 이
        // 발상은 사실 1차 풀이의 39-40줄 주석에 이미 적혀 있습니다 - "중복이 발생하면 어디로 돌아가야
        // 할지? 중복이 발생한 인덱스부터 시작합니다". 흔한 함정은 이 옳은 생각을 "왼쪽 끝만 당기기"가
        // 아니라 "시작점을 1씩 늘려 매번 처음부터 다시 스캔하기"로 옮기는 것이고, 그러면 윈도우를
        // 살려 미는 대신 매번 통째로 버리게 됩니다. (이건 LC49와 똑같은 간극입니다. 거기서도 "차라리
        // 정렬?"이라는 정답이 주석에 적혔지만 정렬이 "두 단어를 맞대보는 비교 도구"로만 쓰였을 뿐
        // "한 단어의 그룹 키"로 승격되지 못했습니다. 스친 정답을 한 걸음 더 미는 것 - 두 문제가 같이
        // 가리키는 지점입니다.)
        //
        // > 가장 긴/짧은 "연속 구간"을 찾으라면 -> 두 포인터로 윈도우의 양 끝을 민다(right는 되감지
        // >   않고 계속 전진, left만 조건이 깨질 때 당긴다). "처음부터 다시"는 윈도우를 버리라는 게
        // >   아니라 왼쪽 끝을 점프하라는 신호다.
        //
        // 불변식 한 문장: "오른쪽 끝 right를 한 칸 전진시킨 직후, 윈도우 [left..right]는 항상 중복 없는
        // 구간이다." 아래 코드는 이 문장의 받아쓰기입니다. 자료구조는 "문자 -> 그 문자를 마지막으로
        // 본 인덱스"(lastSeen). right를 0부터 끝까지 딱 한 번만 전진시키며, s[right]가 맵에 있고 그
        // 위치가 left 이상이면(= 지금 윈도우 안에서 본 중복이면) left를 "그 위치 + 1"로 한 번에
        // 점프시킵니다. 한 칸씩 당기는 게 아니라 바로 건너뜁니다.
        //
        // "abcabcbb"가 통과하는 모습 - 두 끝 left/right와 맵, 답:
        //
        //   s =  a  b  c  a  b  c  b  b
        //   i =  0  1  2  3  4  5  6  7
        //
        //   right=0 'a' 처음     left=0  map{a:0}              len=1  ans=1
        //   right=1 'b' 처음     left=0  map{a:0,b:1}          len=2  ans=2
        //   right=2 'c' 처음     left=0  map{...,c:2}          len=3  ans=3
        //   right=3 'a' 중복@0   0>=0 -> left=1  map{a:3,...}  len=3  ans=3
        //   right=4 'b' 중복@1   1>=1 -> left=2  map{b:4,...}  len=3  ans=3
        //   right=5 'c' 중복@2   2>=2 -> left=3  map{c:5,...}  len=3  ans=3
        //   right=6 'b' 중복@4   4>=3 -> left=5  map{b:6,...}  len=2  ans=3
        //   right=7 'b' 중복@6   6>=5 -> left=7  map{b:7,...}  len=1  ans=3   => 3
        //
        // left가 0->1->2->3->5->7로 전진만 하고 한 번도 되돌아가지 않는 게 핵심입니다. right도 0->7로
        // 전진만 합니다. 두 끝이 각자 최대 n번 움직이고 끝나니 합쳐서 2n, 상수 버리면 O(n).
        //
        // left를 "중복 위치 + 1"로 점프해도 왜 답을 안 놓칠까요? s[right]와 같은 문자가 prev에 있었다면,
        // left를 prev 이하로 두는 순간 윈도우 안에 그 문자가 두 번 들어와 무중복이 깨집니다. 그러니
        // s[right]를 품는 무중복 윈도우의 왼쪽 끝은 아무리 왼쪽이어도 prev+1이고, 그보다 왼쪽은 볼 필요가
        // 없어 통째로 건너뜁니다. 단 prev가 이미 left보다 왼쪽(현재 윈도우 밖)이면 점프는 도리어 left를
        // 뒤로 끌어 윈도우를 늘려버리니, "prev >= left일 때만 점프"라는 조건이 필요합니다.
        var lastSeen = new HashMap<Character, Integer>();
        var answer = 0;
        var left = 0;
        for (var right = 0; right < s.length(); right++) {
            var c = s.charAt(right);
            var prev = lastSeen.get(c);
            if (prev != null && prev >= left) {
                left = prev + 1;
            }
            lastSeen.put(c, right);
            answer = Math.max(answer, right - left + 1);
        }
        return answer;
    }

    /**
     * 1차 풀이 보존(비교용) - 시작점 i마다 set으로 중복까지 다시 스캔.
     * <p>
     * 이 풀이를 "O(n^2)"라고 부르기 쉽지만(아래 45줄 주석도 그렇게 셉니다) 그건 틀립니다. 그리고
     * 그게 틀렸다는 증거는 이미 있습니다 - LeetCode가 5만 입력을 74ms로 통과시켰다는 사실입니다.
     * 진짜 O(n^2)면 5만^2 = 2.5*10^9번이라 수 초가 걸려 TLE가 났어야 합니다. 실제로는 문자 종류가
     * charset(영문/숫자/기호, σ로 적으면 약 95가지)으로 제한돼, 안쪽 for(j)는 무중복 구간이 길어야
     * σ에서 중복을 만나 멈춥니다. 그래서 비용은 바깥 n * 안쪽 최대 σ = O(n * σ)이고, σ가 입력과
     * 무관한 상수라 점근적으로 O(n)입니다 - main 벤치에서 입력을 2배로 키우면 이 풀이 시간도 약 2배로
     * (제곱이면 4배) 늘어 선형임이 확인됩니다.
     * <p>
     * 그러면 윈도우는 무엇이 더 나은가. 점근 차수는 같은 O(n)이지만 상수가 다릅니다. 이 풀이는 같은
     * 문자를 최대 σ번 다시 훑습니다(윈도우를 버리고 시작점마다 새로 스캔하니까). 윈도우는 각 문자를
     * 딱 한 번만 봅니다. worst 입력(무중복 런이 긴 경우)에서 그 차이가 main 표에 수십 배로 찍힙니다.
     * 그리고 charset이 제한되지 않는 문제(정수 배열, 유니코드)라면 σ가 n까지 커질 수 있어 이 풀이는
     * 진짜 O(n^2)로 무너지지만, 윈도우는 charset과 무관하게 O(n)을 지킵니다 - 그래서 윈도우가 더
     * 안전한 정답입니다.
     */
    int lengthOfLongestSubstringByScan(String s) {
        // 1) 손 실행: "pwwkew"에서 가장 긴 무중복 구간을 손으로 찾아 봅니다.
        //    찾는 동안 내 손이 구간을 언제 늘리고, 언제 무엇을 기준으로 줄이는지 관찰합니다.
        // 2) 불변식 문장: "오른쪽 끝을 i까지 늘린 직후, 윈도우는 항상 ____다."
        // 3) 반례 공격: 빈 문자열, 전부 같은 문자, 전부 다른 문자, ... -> main의 슬롯에 추가.
        // 4) 코드 = 2번 문장의 받아쓰기. 10분 안에 2번이 안 잡히면 브루트포스로 후퇴합니다.

        // 1. s = "abcabcbb"
        //    abc, bca, cab 세 글자
        //    bc, cb,
        //    b, b
        //    => 3
        // 2. s = "bbbbb"
        //    b, b, b, b, b
        //    => 1
        // 3. s = "pwwkew"
        //    pw, w, wke, kew
        //    => 3

        // 선형적으로 탐색합니다.
        // 그리고 중복이 발생하는지 set으로 체크합니다.
        // 중복이 발생하면 어디로 돌아가야 할지? 중복이 발생한 인덱스부터 시작합니다.
        // - abc|a -> bca|b -> cab|c -> abc|b -> bc|b -> cb|b -> b|b -> b
        // - pw|w -> w -> wke|w -> w
        // 그러면 인덱스 순서대로 진행하고, 중복이 발생하면 중복 체크하는 set을 초기화합니다.
        // 일단 2중 for문만 생각나는데...
        // 2중 for문을 돈다면? (5*10^4)^2로 최악의 경우 50,000^2 = 2,500,000,000
        var chars = s.toCharArray();
        if (chars.length == 0) return 0;
        if (chars.length == 1) return 1;

        var answer = 0;
        var isDup = new HashSet<Character>();
        for (var i = 0; i < chars.length; i++) {
            isDup.add(chars[i]);

            for (var j = i + 1; j < chars.length; j++) {
                var next = chars[j];
                if (isDup.contains(next)) {
                    break;
                }
                isDup.add(next);
            }

            answer = Math.max(isDup.size(), answer);
            isDup.clear();
        }

        return answer;
    }

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다 - 풀면 초록이 됩니다.
        check(s.lengthOfLongestSubstring("abcabcbb"), 3);
        check(s.lengthOfLongestSubstring("bbbbb"), 1);
        check(s.lengthOfLongestSubstring("pwwkew"), 3);
        check(s.lengthOfLongestSubstring(" "), 1);
        check(s.lengthOfLongestSubstring("au"), 2);
        check(s.lengthOfLongestSubstring(""), 0);
        check(s.lengthOfLongestSubstring("abcdaxz"), 6);
        check(s.lengthOfLongestSubstringByScan("abcdaxz"), 6);
        // 3번에서 직접 만든 반례를 여기에 추가:

        // 두 풀이가 같은 답을 내는지 - 윈도우가 1차 풀이를 망가뜨리지 않았다는 보증
        for (var t : new String[]{"abcabcbb", "bbbbb", "pwwkew", " ", "au", "", "a", "dvdf", "tmmzuxt", "abba"}) {
            sameAnswer(s, t);
        }

        // "더 낫다"를 말이 아니라 측정으로: 입력을 2배씩 키우며 두 풀이의 시간을 잰다
        bench(s);
    }

    private static void check(int actual, int expected) {
        if (actual != expected) {
            throw new AssertionError("expected=" + expected + ", actual=" + actual);
        }
        System.out.println("PASS " + actual);
    }

    private static void sameAnswer(Solution s, String input) {
        var w = s.lengthOfLongestSubstring(input);
        var b = s.lengthOfLongestSubstringByScan(input);
        if (w != b) {
            throw new AssertionError("불일치 input=\"" + input + "\" window=" + w + " scan=" + b);
        }
    }

    // worst case 입력: 무중복 런이 charset 끝까지 가는(= 1차 풀이 안쪽 루프가 가장 길게 도는) 패턴.
    // 32~126(출력 가능 ASCII 95가지)을 주기로 반복하면 어느 시작점에서든 95칸을 가야 중복을 만난다.
    private static String worst(int n) {
        var sb = new StringBuilder(n);
        for (var i = 0; i < n; i++) sb.append((char) (32 + (i % 95)));
        return sb.toString();
    }

    private static void bench(Solution s) {
        for (var i = 0; i < 20; i++) {  // JIT 워밍업
            s.lengthOfLongestSubstringByScan(worst(8000));
            s.lengthOfLongestSubstring(worst(8000));
        }
        System.out.println("--- worst case 벤치 (무중복 런 = charset 95, 입력 2배씩 / 3회 최소 시간) ---");
        var prevScan = 0.0;
        for (var n : new int[]{25000, 50000, 100000, 200000}) {
            var in = worst(n);
            var scanMs = bestOf3(() -> s.lengthOfLongestSubstringByScan(in));
            var winMs = bestOf3(() -> s.lengthOfLongestSubstring(in));
            var scale = prevScan == 0 ? "-" : String.format("입력2배당 %.1fx", scanMs / prevScan);
            System.out.printf("n=%6d  1차(scan)=%7.2fms (%s)  윈도우=%5.2fms  scan/윈도우=%.0f배%n",
                    n, scanMs, scale, winMs, scanMs / winMs);
            prevScan = scanMs;
        }
        System.out.println("판정: 입력 2배에 1차도 약 2배(제곱이면 4배) = 둘 다 선형 O(n). 윈도우는 같은 차수에서 상수가 작아 수십 배 빠름.");
    }

    private static double bestOf3(Runnable r) {
        var best = Long.MAX_VALUE;
        for (var i = 0; i < 3; i++) {
            var t0 = System.nanoTime();
            r.run();
            best = Math.min(best, System.nanoTime() - t0);
        }
        return best / 1e6;
    }
}
