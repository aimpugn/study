package p42586;

import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Queue;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42586?language=java">기능개발</a>
 * <p>
 * 준비 메타
 * - 공식 레벨: Lv.2
 * - 분류/트랙: Stage1/StackQueue
 * - 핵심 패턴: Queue
 * - 체감 난이도: 체감 Lv.2 하~중
 * - 예상 풀이 시간: 실전 숙련자 15~25분 / 학습 모드 35~60분
 * - 첫 접근: 각 작업의 완료일을 먼저 만들고, 앞 작업의 배포일 기준으로 묶는다.
 * <p>
 * 이 문제의 Stack/Queue 분류는 "반드시 Queue를 써야 한다"는 뜻이라기보다,
 * 먼저 들어온 기능이 뒤 기능의 배포를 막는 선입선출 흐름을 읽어야 한다는 신호에 가깝다.
 * 배열 인덱스로 풀어도 같은 흐름을 표현할 수 있고, Queue로 풀면 그 흐름이 자료구조 이름으로 드러난다.
 */
class Solution {
    /*
     * 이 문제는 테스트 케이스의 모양을 따라가면 쉽게 흔들린다.
     * 먼저 문제 문장에서 변하지 않는 규칙을 뽑고, 그 규칙을 코드 상태로 바꾸는 편이 안전하다.
     * <p>
     * 문제 분석은 이렇게 시작한다.
     * 1. 반환값은 "각 기능의 완료일"이 아니라 "각 배포마다 함께 나가는 기능 수"다.
     * 2. 뒤 기능이 먼저 끝날 수 있지만, 앞 기능이 배포되기 전에는 뒤 기능도 먼저 배포되지 못한다.
     * 3. 따라서 비교 기준은 직전 기능 하나가 아니라 "현재 배포 묶음을 이끄는 앞 기능의 완료일"이다.
     * 4. 이 구조는 먼저 온 기능이 뒤 기능을 막는 선입선출 흐름이라 Queue 패턴으로 읽을 수 있다.
     * <p>
     * 여기서 세울 수 있는 루프 불변식(loop invariant)은 풀이 방식마다 조금 다르다.
     * 루프 불변식은 반복문이 한 바퀴 돌기 전과 돈 뒤에도 계속 참이어야 하는 조건이다.
     * <p>
     * - 인덱스 풀이: `start`보다 앞의 기능들은 정확히 한 번씩 배포 묶음에 들어갔고,
     *   `start`는 아직 처리하지 않은 첫 기능이다.
     * - Queue 풀이: Queue에는 아직 배포 묶음에 넣지 않은 기능들의 완료일만 원래 순서대로 남아 있고,
     *   Queue의 맨 앞 값이 이번 배포 묶음의 기준일이다.
     * <p>
     * 이 불변식이 있으면 `테스트 1은 [2, 1]이니까 이렇게 세면 되겠다`가 아니라,
     * `이번 묶음을 닫는 순간 다음 시작점은 어디인가`를 먼저 묻게 된다.
     * 그 질문이 바로 기존 풀이에서 `i = j - 1`이 필요했던 이유이고,
     * 개선 풀이에서는 `start = next` 또는 `poll()`로 더 직접 표현된다.
     * <p>
     * 조건과 루프는 다음 순서로 정한다.
     * 1. 바깥 반복의 한 단위는 "배포 묶음 하나를 완성하는 일"이다.
     *    그래서 바깥 루프는 아직 처리하지 않은 기능이 남아 있는 동안만 돈다.
     * 2. 안쪽 반복의 한 단위는 "다음 기능이 현재 배포 묶음에 들어갈 수 있는지 확인하는 일"이다.
     *    그래서 안쪽 루프 조건은 `다음 기능이 존재한다`와 `그 기능의 완료일이 기준일보다 늦지 않다`가 된다.
     * 3. 안쪽 루프가 멈추는 순간은 현재 묶음이 닫히는 순간이다.
     *    이때 멈춘 위치가 다음 배포 묶음의 시작점이다.
     * <p>
     * 자료구조는 이 규칙을 어떻게 표현하고 싶은지에 따라 고른다.
     * - 배열 인덱스는 `start`, `next`로 다음 시작 위치를 직접 보여 주기 좋다.
     * - Queue는 "맨 앞 기능을 꺼내고, 그 기능이 허락하는 뒤 기능만 같이 꺼낸다"는 선입선출 흐름을 보여 주기 좋다.
     * - Stack은 마지막에 들어온 기능을 먼저 보는 구조라 이 문제의 배포 순서와 맞지 않는다.
     * <p>
     * 대표 반례를 손으로 따라가면, 왜 "다음 시작점"이 핵심인지 더 잘 보인다.
     * <pre>
     * remained = [5, 10, 1, 1, 1, 20]
     *
     * start=0, 기준일=5
     *   next=1의 10은 5보다 늦게 끝난다.
     *   이번 묶음은 [5] 하나이고, 다음 start는 1이다.
     *
     * start=1, 기준일=10
     *   next=2,3,4의 1은 10보다 늦지 않으므로 같은 묶음이다.
     *   next=5의 20은 10보다 늦게 끝난다.
     *   이번 묶음은 [10, 1, 1, 1] 네 개이고, 다음 start는 5다.
     *
     * start=5, 기준일=20
     *   남은 기능은 [20] 하나다.
     *
     * answer = [1, 4, 1]
     * </pre>
     */

    /**
     * - 기능 개선 작업 중이며 각 기능은 진도가 100%일 때 서비스에 반영 가능합니다.
     * - 각 기능 개발 속도는 모두 다르므로 뒤에 있는 기능이 앞에 있는 기능보다 먼저 개발될 수 있습니다. 이때 뒤에 있는 기능은 앞에 있는 기능이 배포될 때 함께 배포됩니다.
     * - 배포는 하루에 한 번만 가능하고, 하루의 끝에 이루어진다고 가정.
     * - 예를 들어 진도율 95%의 개발 속도가 하루 4%라면 배포는 2일 뒤에 이뤄집니다.
     *
     * @param progresses 먼저 배포되어야 하는 순서대로 작업의 진도가 적힌 정수 배열
     * - progresses.length <= 100
     * - 0 < progresses[i] < 100
     * @param speeds 각 작업의 개발 속도가 적힌 정수 배열
     * - speeds.length <= 100
     * - 0 < speeds[i] <= 100
     *
     * @return 각 배포마다 몇 개의 기능이 배포되는지
     */
    public int[] solution(int[] progresses, int[] speeds) {
        // 특정 날짜가 아니라, 배포마다이므로, 먼저 나가는대로 카운트 합니다.
        //
        // 예: 93(1), 30(30), 55(5)
        // - 93은 7일 뒤
        // - 30은 3일 뒤
        // - 55는 9일 뒤
        // => 30은 93과 같이 배포되므로 2, 55는 단독 배포되므로 1
        // => [2, 1]
        //
        // 예: 95(1), 90(1), 99(1), 99(1), 80(1), 99(1)
        // - (1*5), (1*10, 1*1, 1*1), (1*20, 1*1)

        // 각 작업별 속도로 남은 일수들을 계산
        var remained = new int[progresses.length];
        for (var i = 0; i < remained.length; i++) {
            int days = (100 - progresses[i]) / speeds[i];
            days += ((100 - progresses[i]) % speeds[i]) != 0 ? 1 : 0;
            remained[i] = days;
        }
        // 93(1), 30(30), 55(5) 경우 remained: [7, 3, 9]
        // 95(1), 90(1), 99(1), 99(1), 80(1), 99(1) 경우 remained: [5, 10, 1, 1, 20, 1]
        // 95(1), 90(1), 99(1), 99(1), 93(1), 99(1) 경우 remained: [5, 10, 1, 1, 7, 1]
        var answer = new ArrayList<Integer>();
        var total = 0;
        for (var i = 0; i < remained.length; i++) {
            // 현재 풀이의 핵심 생각은 맞습니다.
            // 기준 완료일보다 늦게 끝나지 않는 뒤 기능은 앞 기능과 같은 날 배포됩니다.
            // 다만 이 방식은 바깥 for문이 i를 자동 증가시키므로, 묶음을 닫을 때 인덱스 보정이 필요합니다.
            if (total == remained.length) {
                break;
            }
            var curr = remained[i];
            var cnt = 1;
            // 다음 작업을 검토
            for (var j = i + 1; j < remained.length; j++) {
                if (curr < remained[j]) {
                    // j는 다음 배포 묶음의 첫 위치입니다.
                    // 바깥 for문이 곧 i++를 실행하므로, 지금은 j - 1로 맞춰 둡니다.
                    i = j - 1;
                    break;
                }
                cnt++;
            }
            total += cnt;
            answer.add(cnt);
        }

        return answer.stream().mapToInt(Integer::intValue).toArray();
    }

    /**
     * 현재 풀이와 같은 생각, 즉 "완료일까지 계산한 뒤 앞 기능 기준으로 묶는다"를 유지하되
     * 바깥 for문의 자동 증가에 기대는 `i = j - 1` 보정을 없앤 버전이다.
     * <p>
     * 여기서 가장 중요한 불변식은 다음 한 문장이다.
     * `start`는 아직 어떤 배포 묶음에도 넣지 않은 첫 번째 기능을 가리킨다.
     * 이 불변식이 보이면 `total` 같은 별도 개수 보정 없이도 다음 시작 위치를 직접 정할 수 있다.
     */
    public int[] solutionByExplicitStartIndex(int[] progresses, int[] speeds) {
        // 먼저 각 기능이 며칠 뒤 완료되는지 같은 단위의 숫자로 바꾼다.
        // 진도율과 속도를 그대로 비교하면 "앞 기능 때문에 기다린다"는 규칙을 적용하기 어렵다.
        int[] remained = calculateRemainedDays(progresses, speeds);

        // answer에는 각 배포 시점에 함께 나가는 기능 개수만 쌓는다.
        // 실제 배포 날짜는 반환하지 않으므로, 날짜 자체보다 묶음 크기가 최종 산출물이다.
        var answer = new ArrayList<Integer>();

        // 바깥 루프의 반복 단위는 "배포 묶음 하나 만들기"다.
        // start는 아직 처리하지 않은 첫 기능이고, 한 묶음을 만들 때마다 다음 묶음의 첫 위치로 이동한다.
        int start = 0;
        while (start < remained.length) {
            // 배포 묶음의 기준일은 그 묶음에서 가장 앞에 있는 기능의 완료일이다.
            // 뒤 기능이 더 빨리 끝나도 이 기준일 전에는 먼저 배포될 수 없다.
            int deployDay = remained[start];

            // next는 현재 묶음에 들어갈 수 있는지 검사할 후보 위치다.
            // start는 묶음의 대표 기능으로 이미 포함됐으므로 바로 다음 칸에서 시작한다.
            int next = start + 1;

            // 안쪽 루프의 조건은 두 부분으로 나뉜다.
            // next < remained.length: 검사할 후보 기능이 실제로 남아 있어야 한다.
            // remained[next] <= deployDay: 후보 기능이 기준일보다 늦게 끝나지 않아야 같은 묶음에 들어간다.
            // 조건이 깨지는 순간 next는 이번 묶음에 들어가지 못하는 첫 위치, 즉 다음 묶음의 시작점이다.
            while (next < remained.length && remained[next] <= deployDay) {
                next++;
            }

            // start는 묶음의 첫 위치, next는 묶음이 끝난 다음 위치다.
            // 따라서 next - start가 이번 배포에 포함된 기능 수가 된다.
            answer.add(next - start);

            // 다음 반복은 아직 묶지 않은 첫 기능에서 시작한다.
            // 이 줄이 기존 풀이의 `i = j - 1`이 표현하려던 의미를 더 직접 드러낸다.
            start = next;
        }

        // Programmers 제출 형식은 int[]라서, 학습 중 다루기 쉬운 List<Integer>를 마지막에만 변환한다.
        return toIntArray(answer);
    }

    /**
     * Queue를 사용한 풀이.
     * <p>
     * Queue는 먼저 들어온 값이 먼저 나가는 자료구조다. 이 문제에서는 앞 기능이 아직 배포되지 않으면
     * 뒤 기능도 먼저 나갈 수 없으므로, 완료일을 앞에서부터 꺼내며 배포 묶음을 만드는 흐름과 잘 맞는다.
     * 다만 Queue가 배열 인덱스 풀이보다 시간복잡도를 낮추지는 않는다. 둘 다 O(n)이고,
     * Queue 버전은 완료일을 담는 별도 자료구조 때문에 메모리를 조금 더 쓴다.
     */
    public int[] solutionByQueue(int[] progresses, int[] speeds) {
        // 각 기능의 완료일을 배포 대기열에 넣는다.
        // Queue의 앞쪽 값은 "아직 배포되지 않은 가장 앞 기능"의 완료일이다.
        Queue<Integer> waitingDays = new ArrayDeque<>();
        for (int i = 0; i < progresses.length; i++) {
            waitingDays.add(daysToFinish(progresses[i], speeds[i]));
        }

        // 각 배포마다 몇 개가 함께 나가는지 저장한다.
        var answer = new ArrayList<Integer>();

        // 바깥 루프의 반복 단위는 인덱스 풀이와 같다.
        // Queue가 비지 않았다는 것은 아직 어떤 배포 묶음에도 넣지 않은 기능이 남아 있다는 뜻이다.
        while (!waitingDays.isEmpty()) {
            // 맨 앞 기능은 이번 배포 묶음의 기준이 된다.
            // poll은 값을 꺼내면서 Queue에서 제거하므로, 같은 기능을 다시 세지 않는다.
            int deployDay = waitingDays.poll();

            // 기준 기능 하나는 이미 이번 배포에 포함됐다.
            int count = 1;

            // 안쪽 루프는 Queue의 맨 앞 기능이 현재 묶음에 들어갈 수 있는 동안만 돈다.
            // !waitingDays.isEmpty()를 먼저 확인해야 peek 결과가 없는 상태를 피할 수 있다.
            // waitingDays.peek() <= deployDay는 "앞 기능이 배포될 때 같이 나갈 수 있는가"라는 문제 규칙을 그대로 옮긴 조건이다.
            while (!waitingDays.isEmpty() && waitingDays.peek() <= deployDay) {
                waitingDays.poll();
                count++;
            }

            // 기준일보다 더 늦게 끝나는 기능을 만나면 이번 묶음은 닫힌다.
            answer.add(count);
        }

        // List<Integer>로 쌓은 배포 개수를 제출 형식인 int[]로 바꾼다.
        return toIntArray(answer);
    }

    private int[] calculateRemainedDays(int[] progresses, int[] speeds) {
        // 각 기능마다 완료일까지 걸리는 일수를 담는다.
        // remained[i]는 i번째 기능이 혼자 배포 가능해지는 날짜다.
        int[] remained = new int[progresses.length];
        for (int i = 0; i < progresses.length; i++) {
            remained[i] = daysToFinish(progresses[i], speeds[i]);
        }
        return remained;
    }

    private int daysToFinish(int progress, int speed) {
        // 남은 진도율을 하루 개발 속도로 나누면 필요한 날짜의 기본값이 나온다.
        int remainingProgress = 100 - progress;

        // 나누어떨어지지 않는 경우에는 하루가 더 필요하다.
        // 예를 들어 5%가 남고 하루 4%씩 진행하면 1일로는 부족하고 2일 뒤에 배포 가능하다.
        return (remainingProgress + speed - 1) / speed;
    }

    private int[] toIntArray(List<Integer> values) {
        // Programmers는 기본형 int[]를 기대한다.
        // List<Integer>는 묶음을 동적으로 추가하기 편하고, 마지막에만 제출 형식으로 바꾼다.
        return values.stream().mapToInt(Integer::intValue).toArray();
    }
}

public class Main260511 {
    public static void main(String[] args) {
        Solution solution = new Solution();

        // System.out.println(Arrays.toString(solution.solution(...)));
        var testCases = List.of(
            new TestCase(List.of(new int[]{93, 30, 55}, new int[]{1, 30, 5}), new int[]{2, 1}),
            new TestCase(List.of(new int[]{95, 90, 99, 99, 80, 99}, new int[]{1, 1, 1, 1, 1, 1}), new int[]{1, 3, 2}),
            new TestCase(List.of(new int[]{95, 90, 99, 99, 93, 99}, new int[]{1, 1, 1, 1, 1, 1}), new int[]{1, 5}),
            new TestCase(List.of(new int[]{95, 90, 99, 99, 99, 80}, new int[]{1, 1, 1, 1, 1, 1}), new int[]{1, 4, 1}),
            new TestCase(List.of(new int[]{99, 98, 97}, new int[]{1, 1, 1}), new int[]{1, 1, 1}),
            new TestCase(List.of(new int[]{99, 99, 99}, new int[]{1, 1, 1}), new int[]{3}),
            new TestCase(List.of(new int[]{99}, new int[]{1}), new int[]{1})
        );

        for (var testCase : testCases) {
            var input = testCase.input();
            assertAnswer("현재 풀이", testCase.answer(), solution.solution(input.getFirst(), input.getLast()));
            assertAnswer("명시적 인덱스 풀이", testCase.answer(), solution.solutionByExplicitStartIndex(input.getFirst(), input.getLast()));
            assertAnswer("Queue 풀이", testCase.answer(), solution.solutionByQueue(input.getFirst(), input.getLast()));
        }

        System.out.println("All p42586 Main260511 checks passed.");
    }

    private static void assertAnswer(String label, int[] expected, int[] actual) {
        if (!Arrays.equals(expected, actual)) {
            throw new AssertionError(
                label + " failed. expected=" + Arrays.toString(expected)
                    + ", actual=" + Arrays.toString(actual)
            );
        }
    }

    private record TestCase(List<int[]> input, int[] answer) {
    }
}
