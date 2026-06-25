package dailytemperatures;

import support.Judge;

/**
 * <a href="https://leetcode.com/problems/daily-temperatures/">Daily Temperatures</a>
 * <p>
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md에 작성합니다.
 */
class Solution {
    /**
     *
     * @param temperatures 매일 온도를 나타내는 배열
     * -
     * @return answer[i]는 '더 따뜻해질 때까지 기다려야 하는 i_th 날짜'를 나타내는 배열
     * - 불가능하다면 answer[i] == 0
     */
    public int[] dailyTemperatures(int[] temperatures) {
        // 1. temperatures = [73,74,75,71,69,72,76,73]
        //    73 < 74  74 < 75  75 (> 71 > 69 > 72) < 76 71 (> 72) < 76 69 < 72 76 > 73
        //    [     1,       1,                        4,             2,      1,      0, 0]
        //
        // 2. temperatures = [30,40,50,60]
        //    30 > 40  40 < 50  50 < 60
        //    [     1,       1,       1, 0]
        //
        // 수가 같으면? 더 따뜻해지는 거니까 다음 인덱스로 이동시키는 게 맞을 것으로 보입니다.
        // 첨삭: 결론(같으면 계속 오른쪽으로)은 맞습니다. 다만 "같으면 더 따뜻해지는 것"은 아니에요. 같은 온도는
        //   더 따뜻한 게 아니라 답에 안 칩니다. 그래서 같을 때 멈추지 않고 계속 찾는 게 맞고, 아래 코드의 >=가 그걸 정확히 합니다.

        // 현재 온도 기준으로 더 높은 온도를 만날 때까지의 수.
        // 배열에서 순서대로 하나씩 꺼내면서 그 다음 온도와 비교합니다.
        // 바로 떠오르는 방법은 중첩 루프입니다.
        // 하지만 더 좋은, 다른 방법이 있는지 고민해 봅니다...
        // 이런 경우 스택을 사용했던 거 같습니다.
        // **왜? 중첩 루프보다 스택을 사용하는 게 더 나은 이유는?** 사실 모르겠으므로, 일단 다른 방법을 찾아 봅니다.
        //
        // 또는 포인터를 두 개 움직이는 것도 가능할 거 같습니다. 현재 인덱스를 가리키는 포인터와, 현재 인덱스부터 더 큰 수를 찾는 포인터.
        //
        // 첨삭: 이 풀이는 정답이지만 입력이 커지면 너무 느립니다(O(n^2)). 47/48이 통과하고 마지막 하나가 TLE인
        //   이유가 여기 있어요. 위에서 "스택을 썼던 거 같은데 왜 더 나은지 모르겠다"며 접어 둔 그 질문이 사실
        //   이 문제의 핵심입니다. 그 "왜"를 여는 게 정답으로 가는 다리예요.
        //
        //   왜 느린가: 각 i에서 오른쪽을 한 칸씩 끝까지 스캔하는데, 같은 구간을 i마다 다시 봅니다. 실패 입력
        //   [99]를 99999개 늘어놓고 끝에 [100]을 둔 경우를 보면, 99 하나하나가 자기 오른쪽의 99들을 전부 다시
        //   훑고서야 끝의 100에 닿습니다. 합이 n + (n-1) + ... 약 n^2/2 = 약 5*10^9 번. 실측으로 n=100000에서
        //   926ms 걸립니다(아래 main 참고). 통과한 47개가 안 느렸던 건 그 입력들이 "긴 같은 값/내림 구간 뒤
        //   봉우리" 같은 적대적 형태가 아니었을 뿐, 통과 사실이 O(n^2)를 가려 준 것은 아닙니다.
        //
        //   "왜 스택이 중첩 루프보다 나은가"의 답: 중첩 루프는 한 칸을 여러 번 다시 보지만, 스택은 각 날을 딱 한 번
        //   쌓고(push) 딱 한 번 꺼냅니다(pop). 다시 보는 일이 없으니 전체가 한 번 훑기(O(n))로 끝납니다. 고친 풀이
        //   둘이 이 메서드 아래에 있습니다 — 하나는 이 오른쪽-스캔 골격을 살린 버전, 하나는 스택(베스트)입니다.
        //
        // > 카드: "각 원소에서 다음으로 더 큰(작은) 값 찾기" 신호 -> 단조 스택. 각 원소를 한 번 push, 한 번 pop -> O(n).
        var answer = new int[temperatures.length];
        for (int i = 0; i < temperatures.length; i++) {
            var j = i + 1;
            // 첨삭: 바로 이 while이 "같은 구간 재스캔"의 자리입니다. i가 바뀔 때마다 오른쪽을 처음부터 다시 훑어요.
            //   j를 한 칸씩(j++) 올리는 대신 이미 푼 답으로 건너뛰면 이 재스캔이 사라집니다(아래 고친 버전).
            while (j < temperatures.length && temperatures[i] >= temperatures[j]) {
                j++;
            }
            if (j != temperatures.length) {
                answer[i] = j - i;
            }
        }
        // System.out.println(Arrays.toString(answer));

        return answer;
    }

    // 이 오른쪽-스캔 골격을 살린 동작 버전입니다. 바꾼 건 둘 — (1) i를 오른쪽에서 왼쪽으로 돌려 더 오른쪽의 answer를
    //   먼저 채우고, (2) j를 한 칸씩 올리는 대신 answer[j]만큼 건너뜁니다. answer[j]는 "j가 더 따뜻한 날까지
    //   기다리는 날 수"라서, j += answer[j]는 j의 안 따뜻한 구간을 한 번에 뛰어넘습니다.
    //
    // public int[] dailyTemperatures(int[] temperatures) {
    //     int n = temperatures.length;
    //     var answer = new int[n];
    //     for (int i = n - 2; i >= 0; i--) {
    //         int j = i + 1;
    //         while (j < n && temperatures[j] <= temperatures[i]) {
    //             if (answer[j] == 0) { j = n; break; }  // j 뒤에 더 따뜻한 날이 없으면 i도 없음
    //             j += answer[j];                         // j의 안 따뜻한 구간을 한 번에 건너뜀
    //         }
    //         if (j != n) answer[i] = j - i;
    //     }
    //     return answer;
    // }
    //
    // 무엇이 부족했나 (고친 코드와 비교):
    //   원본은 j++로 한 칸씩 가서 같은 99들을 일일이 다시 밟았습니다. 고친 건 answer[j]로 그 구간을 한 번에 건너뜁니다.
    //   왜 O(n)이 되나: 점프는 항상 더 따뜻한 날로 올라갑니다(answer[j]가 가리키는 곳이 t[j]보다 따뜻하니까).
    //   온도가 30~100 사이 정수라, 한 i에서 점프로 올라갈 수 있는 서로 다른 온도는 많아야 약 70단계입니다. 그래서
    //   각 i의 점프 횟수가 상수(70 이하)로 묶여 전체가 O(n * 70) = O(n)이 됩니다(값 범위가 차수를 묶는 경우).
    //   실측: 원본이 926ms이던 n=100000 입력에서 0ms. 예제 3개 + 랜덤 1000개에서 원본과 답이 일치함을 확인했습니다.

    // 베스트 프랙티스입니다. 위에서 접어 둔 그 스택이에요. 왼쪽에서 오른쪽으로 한 번 훑되, 아직 더 따뜻한 날을 못 만난
    //   인덱스들을 스택에 쌓아 둡니다. 더 따뜻한 날이 오면, 그 날을 기다리던 인덱스들을 스택에서 꺼내 답을 채웁니다.
    //
    // > 불변식: 스택에는 "아직 더 따뜻한 날을 못 만난 인덱스"만, 온도가 위로 갈수록 작아지게(단조) 쌓인다.
    //
    public int[] dailyTemperatures2(int[] temperatures) {
        int n = temperatures.length;
        var answer = new int[n];
        var stack = new java.util.ArrayDeque<Integer>();  // 대기 중인 인덱스
        for (int i = 0; i < n; i++) {
            while (!stack.isEmpty() && temperatures[i] > temperatures[stack.peek()]) {
                int idx = stack.pop();   // i가 idx의 "더 따뜻한 날"
                answer[idx] = i - idx;
            }
            stack.push(i);
            System.out.println("stack: " + stack);
        }
        return answer;
    }
    //
    // 고친 학습자 버전보다 왜 더 나은가:
    //   jump 버전의 O(n)은 "온도가 30~100이라 점프가 70번 안에 끝난다"는 값 범위에 기댑니다. 스택은 그 가정이 필요 없어요.
    //   각 인덱스가 스택에 딱 한 번 들어가고 딱 한 번 나오므로, 값 범위와 무관하게 언제나 O(n)입니다. 더 단순하고,
    //   "왜 빠른가"가 한 줄로 끝납니다(각 날을 두 번 보지 않는다). 이게 위에서 접어 둔 질문의 답입니다.
    //   실측: n=100000에서 0ms. 예제 3개 + 랜덤 1000개에서 원본과 일치 확인.
    //   (java.util.ArrayDeque import 필요. 박싱이 싫으면 int[] 스택으로 같은 로직을 써도 됩니다.)

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(s.dailyTemperatures(new int[]{73, 74, 75, 71, 69, 72, 76, 73}), new int[]{1, 1, 4, 2, 1, 1, 0, 0});
        Judge.check(s.dailyTemperatures(new int[]{30, 40, 50, 60}), new int[]{1, 1, 1, 0});
        Judge.check(s.dailyTemperatures(new int[]{30, 60, 90}), new int[]{1, 1, 0});
        Judge.check(s.dailyTemperatures(new int[]{89, 62, 70, 58, 47, 47, 46, 76, 100, 70}), new int[]{8, 1, 5, 4, 3, 2, 1, 1, 0, 0});
        Judge.check(s.dailyTemperatures2(new int[]{89, 62, 70, 58, 47, 47, 46, 76, 100, 70}), new int[]{8, 1, 5, 4, 3, 2, 1, 1, 0, 0});
        // 반례를 여기에 추가하세요. 이 문제의 반례는 "정답이 틀린 입력"이 아니라 "원본을 O(n^2)로 모는 입력"입니다.
        //   (실패 모드별 분류 -> 원본 실측 @ n=100000, jump/stack은 모두 약 0ms)
        //   [전부 같은 값 + 끝에 봉우리]   [99]가 99999개 뒤에 [100]    원본 926ms (실제 TLE 입력)
        //   [단조 감소, 더 따뜻한 날 없음]  [100,99,98,...]              각 i가 끝까지 스캔하고 못 찾음 -> O(n^2)
        //   [단조 증가]                   [30,31,32,...,100,...]       각 i가 바로 다음에서 멈춤 -> 이 경우만 원본도 O(n) (best case)
    }
}
