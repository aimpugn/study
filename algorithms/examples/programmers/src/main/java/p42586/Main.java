package p42586;

import java.util.*;

/**
 * @see <a href=
 * "https://school.programmers.co.kr/learn/courses/30/lessons/42586">
 * 가능개발</a>
 */
public class Main {

    public static void main(String[] args) {
        Main main = new Main();
        System.out.println(Arrays.toString(
                        main.solution(new int[]{93, 30, 55}, new int[]{1, 30, 5}))
                .equals("[2, 1]"));
        System.out.println(
                Arrays.toString(main.solution(new int[]{93}, new int[]{1}))
                        .equals("[1]"));
        System.out.println(Arrays
                .toString(main.solution(new int[]{93, 93}, new int[]{5, 2}))
                .equals("[1, 1]"));
        System.out.println(Arrays
                .toString(main.solution(new int[]{95, 90, 99, 99, 80, 99},
                        new int[]{1, 1, 1, 1, 1, 1}))
                .equals("[1, 3, 2]"));
        System.out.println(Arrays
                .toString(main.solution(new int[]{95, 90, 99, 99, 99, 80},
                        new int[]{1, 1, 1, 1, 1, 1}))
                .equals("[1, 4, 1]"));
        System.out.println(Arrays
                .toString(main.solution(new int[]{95, 95, 90, 99, 80, 99, 80},
                        new int[]{1, 1, 1, 1, 1, 1, 1}))
                .equals("[2, 2, 3]"));
        System.out.println(Arrays
                .toString(main.solution(
                        new int[]{95, 95, 90, 99, 99, 80, 99, 80},
                        new int[]{1, 1, 1, 100, 1, 1, 1, 1}))
                .equals("[2, 3, 3]"));
        System.out.println(Arrays
                .toString(main.solution(
                        new int[]{95, 95, 90, 99, 99, 90, 99, 80},
                        new int[]{1, 1, 1, 1, 1, 1, 1, 1}))
                .equals("[2, 5, 1]"));
        System.out.println(Arrays
                .toString(main.solution(new int[]{99, 90, 93, 95, 95, 90, 90},
                        new int[]{1, 1, 1, 1, 1, 1, 1}))
                .equals("[1, 6]"));

    }

    /**
     * <pre>
     * - 기능 배포 if 진도 100%
     * - 각 기능 개발 속도 다를 수 있음
     * - 배포는 하루 한 번, 하루의 끝에 이뤄진다고 가정
     *   ex: if progresses[N] = 95 && speed[N] = 4, then 2일 뒤 배포
     *
     * ex:
     * - progresses: [93, 30, 55]
     *   speeds:     [1,  30, 5 ]
     *                7d  3d  9d
     *                ^^^^^^  (1)
     *                앞의 7d에
     *                포함(2)
     *   return:     [2, 1]
     *
     * - progresses: [95, 90, 99, 99, 80, 99]
     *   speeds:     [1,  1,  1,  1,  1,  1 ]
     *                5d  10d 1d  1d  20d 1d
     *                (1) ^^^^^^^^^^  ^^^^^^
     *                    10d짜리에     20d짜리에
     *                    포함(3)       포함(2)
     *   return:      [1, 3, 2]
     * </pre>
     *
     * @param progresses 순서: 먼저 배포되어야 하는 순서. 1 <= 요소: 각 작업의 진도 < 100. 1 <=
     *                   progresses.length <= 100.
     * @param speeds     1 <= 요소: 각 작업의 개발 속도 <= 100 ("이하"임에 주의). 1 <= speeds.length
     *                   <= 100.
     */
    public int[] solution(int[] progresses, int[] speeds) {
        // progresses.length와 speeds.length는 같습니다.
        //
        // 계산된 잔여 날짜를 기반으로 비교합니다.
        // progresses: [95, 90, 99, 99, 80, 99]
        // speeds:     [1,  1,  1,  1,  1,  1 ]
        //              5d  10d 1d  1d  20d 1d
        //                 <   >   ==  <    >
        //                  ^^^^^^^^^   ^^^^^
        //              `<` 관계 생길 때부터 새로운 그룹 생김
        //
        // progresses: [95, 90, 99, 99, 99, 80]
        // speeds:     [1,  1,  1,  1,  1,  1 ]
        //              5d  10d 1d  1d  1d  20d
        //                <    >  ==  ==  <
        //               1   ^^^^^^^^^^^^^   1
        //                         4
        //
        // [!NOTE] 놓친 케이스: 단순히 직전/직후만 비교하면 안 됩니다.
        // progresses: [95, 90, 99, 99, 93, 99]
        // speeds:     [1,  1,  1,  1,  1,  1 ]
        //              5d  10d 1d  1d  7d  1d
        //                 <   >   ==  <    >
        //                  ^^^^^^^^^^^^^^^^^^
        //              1          5
        // 현재 배포 그룹을 lead하는 기능, 그 peek 값을 기준으로 비교합니다.
        if (progresses.length == 1) {
            return new int[]{1};
        }

        ArrayList<Integer> deploymentsCounter = new ArrayList<>();
        // 첫번째 작업을 초기화합니다.
        int dayToFinish = calculateDaysToFinish(progresses[0], speeds[0]);
        // 첫 번째 배포 그룹의 카운트를 시작합니다.
        int count = 1;

        for (int i = 1; i < progresses.length; i++) {
            int currentDayToFinish =
                    calculateDaysToFinish(progresses[i], speeds[i]);

            if (dayToFinish < currentDayToFinish) {
                // 새로운 배포 그룹이면 이전 배포 그룹의 count를 stack에 추가하고
                deploymentsCounter.add(count);
                // 새로운 배포 그룹의 count를 초기화합니다.
                count = 1;
                // 그리고 새로운 배포 그룹을 lead하는 기능을 비교 기준으로 설정합니다.
                dayToFinish = currentDayToFinish;
            } else {
                // 현재 배포 그룹의 배포 건수를 계속 카운팅 합니다.
                count++;
            }
        }
        // 마지막 배포 그룹의 경우에는
        // 1. `dayToFinish < currentDayToFinish` 경우: 스택에 count가 추가된 후 초기화된 1이 추가되지 않았습니다.
        // 2. `else` 경우: stack에 추가되지 않습니다.
        //
        // 따라서 어떤 경우든 마지막 배포 그룹의 count가 스택에 아직 추가되지 않았으므로,
        // for loop가 끝나고 이를 추가합니다.
        deploymentsCounter.add(count);
        // System.out.println(stack.toString());

        return deploymentsCounter.stream().mapToInt(Integer::intValue)
                .toArray();
    }

    public int[] solution_1st(int[] progresses, int[] speeds) {
        if (progresses.length == 1) {
            return new int[]{1};
        }

        Stack<Integer> stack = new Stack<>();
        stack.add(1);

        int dayToFinish = calculateDaysToFinish(progresses[0], speeds[0]);

        for (int i = 0; i < progresses.length - 1; i++) {
            int nextDayToFinish =
                    calculateDaysToFinish(progresses[i + 1], speeds[i + 1]);

            if (dayToFinish < nextDayToFinish) {
                // 새로운 배포 그룹
                stack.add(1);
                dayToFinish = nextDayToFinish;
            } else {
                // 현재 배포 그룹
                stack.add(stack.pop() + 1);
            }
        }
        // System.out.println(stack.toString());

        return stack.stream().mapToInt(el -> el).toArray();
    }

    // 잔여 진도: 100 - 95 = 5
    //
    // 작업 완료까지 남은 일수: 5 / 1 = 5
    // 만약 나눠 떨어지지 않는다면?
    // - 5 / 2 = 2(x) => 3
    // - 8 / 3 = 2(x) => 3
    private int calculateDaysToFinish(int progress, int speed) {
        return (int) Math.ceil((100.0 - progress) / speed);
    }
}
