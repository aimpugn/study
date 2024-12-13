package p42895;

import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedList;
import java.util.Map;

/**
 * @see <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42895">
 * N으로 표현
 * </a>
 */
public class Main {
    public static void main(String[] args) {

    }

    /**
     * - 5와 사칙연산만으로 12 표현 가능
     * <p>
     * 1. 5 + 5 + (5/5) + (5/5) = 10 + 1 + 1 = 12
     * ^   ^    ^ ^     ^ ^ 6회
     * 2. 55/5 + 5/5 = 11 + 1 = 12
     * ^^ ^   ^ ^ 5회
     * 3. (55 + 5) / 5 = 12
     * ^^   ^    ^ 4회
     * <p>
     * - 2와 사칙연산만으로 11 표현 가능
     * <p>
     * 1. 2 + 2 + 2 + 2 + 2 + (2/2) = 11
     * 2. 2 + 2 + 2 + 2 + (2/2) + (2/2) + (2/2) = 11
     * => 근데, 이미 1번에서 7개 사용해서 해결했기에,
     * 7번 넘어가는 경우는 고려치 않아도 됩니다.
     * 3. 22/2 = 11
     * <p>
     * rule:
     * - 사칙연산만 가능
     * - 나누기에서 나머지 무시
     *
     * @param N      표현에 사용될 숫자. 1 <= N <= 9.
     * @param number 사칙연산으로 도출해낼 수. 1 <= number <= 32,000.
     * @return N 사용횟수 최솟값. if returnValue > 8 then return -1
     */
    public int solution(int N, int number) {
        // 1. 사칙연산을 해야 합니다.
        // 2. 사칙연산 외에 N이 붙어서 자리수가 증가하는 경우도 고려해야 합니다.
        //   - if N = 5, then 5, 55, 555...
        //   - '(5/5)'를 12번 더하면 12가 나올 수 있으므로, 이것도 가능한 경우입니다.
        //     (5/5) + (5/5) + (5/5) + ... + (5/5) = 12
        //
        // 구상...
        // - 도움이 될 거 같은 자료구조 또는 변수
        //   - 연산 결과 캐싱하는 HashMap
        //   - 가장 최근 number 연산 성공한 조합의 N 개수를 저장해둘 변수 => 초과시 백트래킹
        //
        // - 어떻게 자료구조를 탐색하고 제어할 것인지?
        //
        //   시작 지점을 5로 잡고 55, 555, 5555... 로 체크?
        //   - 5555555 / 5, 5555555 / 55 이런 경우?
        //   - 5/5, 55/55, 555/555 처럼 그 결과가 동일한 경우?
        //     => "NNN"/"NNN" 문자열로 만들어서 길이가 같으면 1로 처리
        //
        //   5/5, 55/55. 555/555 모두 1이고,
        //   5+5, 5+(5/5)+(5/5)+(5/5)+(5/5)+(5/5) 모두 10입니다.
        //   하지만 5+5로 10을 만들 수 있다면, 5+(5/5)+(5/5)+(5/5)+(5/5)+(5/5)를 할 필요는 없습니다.
        //   중요한 건 10을 만드는 데 '5'를 두 개 사용했다는 것이고,
        //   1을 만드는 데 '5'를 두 개 사용했단 것입니다.
        //
        //   - 그러면 반대로 '5' 두 개로 만들 수 있는 수:
        //     10, 1, 25(5*5), 0(5-5)입니다.
        //
        //   - '5' 세 개로 만들 수 있는 수는?
        //      11(55/5), 60(55+5, 5+55), 275(55*5, 5*55), 50(55-5), 0(5/55),
        //      15(5 + 5 + 5), 6(5/5 + 5)
        //     '5' 두 개로 만들 수 있는 수가 재사용됨을 알 수 있습니다.
        //
        // 그렇다면 N으로 연산 가능한 수들을 미리 만들어 놓고 재사용한다면, DP가 가능할 거 같습니다.
        // 'N'을 M개로 만들 수 있는 수를 재사용되므로,
        // `DP[i] = 'N'을 M개로 만들 수 있는 수들`이라고 DP 테이블 가정합니다.
        // - DP[2]는 DP[1] 경우를 포함
        // - DP[3]는 DP[2] 경우를 포함

        if (N == number) {
            return 1;
        }

        HashMap<Integer, HashSet<Integer>> dp = new HashMap<>();
        String strN = String.valueOf(N);

        HashSet<Integer> first = new HashSet<>();
        first.add(N);
        dp.put(1, first);

        // 최솟값이 8보다 크면 -1을 리턴하므로, DP[8]을 넘지 않게 합니다.
        for (int i = 2; i <= 8; i++) {
            // N, NN, NNN 등, i는 처음 주어지는 N의 개수를 의미합니다.
            //
            // N에 대해 주어진 개수로 만들 수 있는 수의 DP를 만듭니다.
            HashSet<Integer> currNumberSet = new HashSet<>();

            // DP[1] = {5} 라고 할 때,
            // 가령 DP[4]에서 5 + (5 + 5 + 5) = 20 이 가능합니다.
            // 이는 다시 DP[1] + DP[3] = 20이 가능합니다.
            // DP[3]은 DP[2]와 DP[1]을 통해 만들 수 있습니다.
            // DP[4]는 (DP[3], DP[1]), (DP[2], DP[2]) 을 통해 만들 수 있습니다.
            //        (DP[1], DP[1], DP[1], DP[1]) 경우에는 5 + 5 + 5 + 5인데,
            //        (DP[1] +^^^^^^^^^^^^^^^^^^DP[3])과 같습니다.
            //        즉, 이미 DP[1] + DP[3]에 포함되어 있습니다.
            //
            // 그렇다면 i=6일 경우 6이 될 수 있는 경우의 수는 다음과 같습니다.
            // - if i=6, then {(1, 5), (2, 4), (3, 3)}
            //   ~~이때 DP[5]는 DP[4]를 포함하고 있고, DP[4]는 DP[3]을 포함합니다~~
            //   ~~가령 DP[4]에는 20(5 + 5 + 5 + 5)이 포함되어 있을 텐데~~
            //   ~~이는 DP[3]의 15(5 + 5 + 5)에 5를 더한 것과 같습니다~~
            //   위의 방식은 틀렸습니다.
            //   DP[2]에 있는 숫자 55
            //   DP[4]에 있는 숫자 130
            //   DP[6]에는 185가 들어가려면 55 + 130이 필요합니다.
            //   그러기 위해서는 (DP[1], DP[1]), (DP[2], DP[4]), (DP[3], DP[3])
            //   케이별로 순회하면서 사칙연산을 수행해야 합니다.
            //

            // 이전에 만들어진 값들에 다시 사칙연산
            for (int j = 1; j < i; j++) {
                // if i = 4, then (1, 3), (2, 2), (3, 1)
                for (int left : dp.get(j)) {
                    for (int right : dp.get(i - j)) {
                        // 2=[0, 1, 25, 10], 3=[0,...
                        // 5를 3번 사용해서 0이 나오는 경우 (5-5)*5
                        // ~~하지만 2번 사용해서 0을 만들 수 있는데, 세번 사용할 필요는 없습니다.~~
                        // 만들어진 숫자 집합으로 가능한 모든 조합 연산을 수행합니다.
                        currNumberSet.add(left + right);
                        currNumberSet.add(left - right);
                        currNumberSet.add(left * right);
                        if (right != 0) {
                            currNumberSet.add(left / right);
                        }
                    }
                }
            }
            // NN, NNN, NNNN을 추가해줍니다.
            currNumberSet.add(Integer.parseInt(strN.repeat(i)));

            // 만약 지금까지 구성한 값들중에 결과가 있다면 early return 합니다.
            if (currNumberSet.contains(number)) {
                return i;
            }
            // System.out.println(i + ": " + currNumberSet.toString());
            dp.put(i, currNumberSet);
        }


        // 앞서 early return 하므로 필요 없습니다.
        // for (Map.Entry<Integer, HashSet<Integer>> entry : dp.entrySet()) {
        //     for (int maybeTargetValue : entry.getValue()) {
        //         if (maybeTargetValue == number) {
        //             return entry.getKey();
        //         }
        //     }
        // }

        return -1;
    }
}
