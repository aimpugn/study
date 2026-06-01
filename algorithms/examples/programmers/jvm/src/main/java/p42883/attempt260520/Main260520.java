package p42883.attempt260520;

import support.TestCase2;

import java.util.Arrays;
import java.util.List;
import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42883?language=java">큰 수 만들기</a>
 */
class Solution {
    public String solution(String number, int k) {
        int targetLength = number.length() - k;
        int remove = k;
        var kept = new StringBuilder(number.length());

        for (int i = 0; i < number.length(); i++) {
            char digit = number.charAt(i);

            while (
                remove > 0
                    && !kept.isEmpty()
                    && kept.charAt(kept.length() - 1) < digit
            ) {
                kept.setLength(kept.length() - 1);
                remove--;
            }

            kept.append(digit);
        }

        return kept.substring(0, targetLength);
    }

    /**
     * @param number 문자열 숫자
     * - 2 <= number <= 1,000,000
     * @param k 제거할 개수
     * - {@code number}의 자릿수 미만인 자연수
     *
     * @return 어떤 숫자에서 k개의 수를 제거했을 때 얻을 수 있는 가장 큰 숫자
     */
    public String solution_bad(String number, int k) {
        // number=1924, k=2
        // => [19, 12, 14, 92, 94, 24]
        // => 94
        //
        // 순서를 유지해야 합니다.
        // 그리고 루프를 돌면서 큰 수가 앞으로 와야 하고,
        // 남은 자리수도 고려해야 합니다.
        //
        // 항상 가장 큰 수대로 와야 합니다.
        // 남은 자리중 가장 큰 수를 찾습니다.
        // 결국 앞의 자리가 크면 뒤의 자리는 작아도 됩니다.
        // 배열을 소비하여 시작점을 찾아 갑니다.
        //
        // 첨삭: 여기까지의 문제 해석은 좋습니다. "순서를 유지한 채 앞자리를 최대화한다"는 핵심을 잡았습니다.
        // 다만 이 풀이가 찝찝한 이유는, 매 자리마다 남은 구간을 다시 훑어서 이번 자리에 올 최댓값을 찾기 때문입니다.
        // 길이가 최대 1,000,000이므로, 같은 숫자들을 여러 번 다시 보는 구조는 정답이 나오더라도 시간 여유가 얇아집니다.
        //
        // 이 문제에서 한 번 더 넘어가야 하는 생각은 "이번 자리에 무엇을 고를까?"가 아니라
        // "현재 숫자가 들어왔을 때, 직전에 남겨 둔 더 작은 숫자를 지워도 되는가?"입니다.
        // 더 큰 숫자가 뒤에서 왔고 삭제 횟수가 남아 있다면, 바로 앞의 작은 숫자를 지우는 것이 앞자리를 키우는 가장 직접적인 행동입니다.

        // 첨삭: split -> stream -> Integer 변환은 숫자 하나를 보기 위해 너무 많은 객체를 만듭니다.
        // 이 문제에서는 각 자리가 문자 하나이므로 number.charAt(i)만으로 비교해도 충분합니다.
        // 특히 입력 길이가 1,000,000까지 열려 있을 때는 이런 변환 비용도 테스트 10 같은 큰 케이스에서 체감됩니다.
        var arr = Arrays.stream(number.split("")).map(Integer::parseInt).toList();

        var len = number.length() - k;
        var answer = new StringBuilder();
        var nextIdx = 0;
        var currMax = 0;

        // 첨삭: 이 while은 answer의 한 자리를 확정할 때마다 남은 후보 구간을 다시 검색합니다.
        // 그래서 겉으로는 그리디지만, 구현 모양은 "자리 수만큼 반복 검색"에 가깝습니다.
        // 더 좋은 불변식은 answer를 스택처럼 보고, answer의 마지막 숫자가 현재 숫자보다 작으면 pop한다는 것입니다.
        // 그러면 각 숫자는 한 번 들어가고 많아야 한 번 빠지므로 전체 흐름이 O(n)으로 닫힙니다.
        while (answer.length() < len) {
            currMax = arr.get(nextIdx);

            var currMaxUpdated = false;
            for (var i = nextIdx; i < number.length(); i++) {
                // 빠져나오는 조건
                // - i와 len에서 answer.length 뺀 값을 합쳐서 number 길이보다 크면 안 됨
                // 1: 0 + 4 - 0 <= 7
                //    1 + 4 - 0 <= 7
                //    2 + 4 - 0 <= 7
                //    3 + 4 - 0 <= 7 -> break
                // 2: 3 + 4 - 1 <= 7
                //    4 + 4 - 1 <= 7
                // 첨삭: 이 조건은 "지금 i를 고른 뒤에도 남은 자리를 채울 수 있는가"를 확인하려는 장치입니다.
                // 판단 자체는 맞지만 식이 거꾸로 읽혀서 디버깅할 때 머릿속에서 다시 해석해야 합니다.
                // 같은 뜻을 더 직접 쓰면 i <= number.length() - (len - answer.length()) 입니다.
                // 다만 스택 관점으로 풀면 이 범위 계산 자체가 사라지고, "작은 이전 숫자를 지울 수 있을 때 지운다"만 남습니다.
                if (number.length() < (i + len - answer.length())) {
                    break;
                }

                if (currMax < arr.get(i)) {
                    // 첨삭: 여기서 nextIdx를 i + 1로 옮기는 것은 "앞자리에 더 큰 수를 놓기 위해 이전 구간을 버린다"는 뜻입니다.
                    // 이 선택은 그리디하게 맞지만, 더 큰 수를 찾을 때마다 앞쪽을 통째로 재검색한 뒤 버리는 모양이 됩니다.
                    // 정석 풀이에서는 현재 digit을 보자마자 answer 끝의 작은 숫자들을 지워서 같은 결정을 즉시 수행합니다.
                    currMaxUpdated = true;
                    currMax = arr.get(i);
                    nextIdx = i + 1;
                }
            }
            if (!currMaxUpdated) {
                // 첨삭: 현재 구간의 첫 숫자가 그대로 선택될 때만 한 칸 전진합니다.
                // 이 분기 때문에 nextIdx의 의미가 "다음 검색 시작점"과 "선택된 숫자 다음 위치" 사이를 오가서 읽기 어렵습니다.
                // 스택 방식은 입력을 왼쪽에서 오른쪽으로 한 번만 소비하므로 이런 별도 시작점 상태가 필요하지 않습니다.
                nextIdx++;
            }

            answer.append(currMax);
        }

        // 첨삭: 이 풀이의 핵심 문제를 한 문장으로 줄이면,
        // "앞자리를 키운다"는 좋은 기준을 잡았지만, 이전 선택을 되돌리는 대신 다음 최댓값을 반복 검색했다는 점입니다.
        //
        // 권장 사고 흐름:
        // - answer는 지금까지 남기기로 한 숫자들의 스택입니다.
        // - 새 digit이 들어왔을 때 answer의 마지막 숫자가 digit보다 작고 삭제 횟수가 남아 있으면, 그 마지막 숫자는 지웁니다.
        // - 더 이상 지울 수 없거나 마지막 숫자가 digit 이상이면 digit을 붙입니다.
        // - 끝까지 봤는데 삭제 횟수가 남아 있으면, 앞자리는 이미 최선이므로 뒤에서 남은 만큼 자릅니다.
        //
        // 4177252841, k=4:
        // [4] -> [4,1] -> 7이 오면 1,4 제거 -> [7,7,2] -> 5가 오면 2 제거
        // -> [7,7,5,2] -> 8이 오면 2 제거 -> [7,7,5,8,4,1]
        //
        // 기억 문장: 더 큰 숫자가 뒤에서 오면, 삭제 횟수가 허용하는 동안 직전에 남긴 작은 숫자부터 지운다.
        return answer.toString();
    }
}

public class Main260520 {
    public static void main(String[] args) {
        var solution = new Solution();

        var testCases = List.of(
            new TestCase2<>("1924", 2, "94"),
            new TestCase2<>("1231234", 3, "3234"),
            new TestCase2<>("4177252841", 4, "775841")
        );

        for (var testCase : testCases) {
            var result = solution.solution(testCase.input(), testCase.input2());
            assertEquals(testCase, result);
        }
    }

    private static void assertEquals(TestCase2<?, ?, ?> testCase, Object actual) {
        if (!valuesEqual(testCase.answer(), actual)) {
            throw new AssertionError(
                "input=" + valueToString(testCase.input())
                    + ", input2=" + valueToString(testCase.input2())
                    + ", expected=" + valueToString(testCase.answer())
                    + ", actual=" + valueToString(actual)
            );
        }
    }

    private static boolean valuesEqual(Object expected, Object actual) {
        if (expected instanceof int[] expectedArray && actual instanceof int[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof long[] expectedArray && actual instanceof long[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof Object[] expectedArray && actual instanceof Object[] actualArray) {
            return Arrays.deepEquals(expectedArray, actualArray);
        }
        return Objects.equals(expected, actual);
    }

    private static String valueToString(Object value) {
        if (value instanceof int[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof long[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof Object[] values) {
            return Arrays.deepToString(values);
        }
        return String.valueOf(value);
    }
}
