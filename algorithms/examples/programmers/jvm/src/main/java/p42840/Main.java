package p42840;

import java.util.*;

public class Main {
    public static void main(String[] args) {
        Main main = new Main();
        System.out.println(main.solution(new int[]{1, 2, 3, 4, 5}));
        System.out.println(main.solution(new int[]{1, 3, 2, 4, 2}));
    }


    /**
     * <pre>
     * 수포자 3인 수학 문제 전부 찍기
     * - 1번: 1, 2, 3, 4, 5, 1, 2, 3, 4, 5...
     * - 2번: 2, 1, 2, 3, 2, 4, 2, 5, 2, 1, 2, 3, 2, 4, 2, 5, ...
     * - 3번: 3, 3, 1, 1, 2, 2, 4, 4, 5, 5, 3, 3, 1, 1, 2, 2, 4, 4, 5, 5, ...
     * </pre>
     *
     * @param answers 1번~마지막 문제까지의 정답이 순서대로 담긴 배열. 1 <= answers.length <= 10,000
     *                정답은 1, 2, 3, 4, 5중 하나
     * @return 가장 많은 문제를 맞힌 사람. 여럿인 경우 오름차순 정렬.
     */
    public int[] solution(int[] answers) {
        int[] firstPattern = new int[]{1, 2, 3, 4, 5};
        int[] secondPattern = new int[]{2, 1, 2, 3, 2, 4, 2, 5};
        int[] thirdPattern = new int[]{3, 3, 1, 1, 2, 2, 4, 4, 5, 5};

        int firstCount = 0;
        int secondCount = 0;
        int thirdCount = 0;

        // 추상화 & 모델링
        // - 일단 패턴이 있으니 패턴을 활용?
        // - 아니면 세 사람이니까 그냥 for문을 돌면서, 일일이 비교한다?
        //
        // for answer in answers:
        //   // 정답의 순서와 각 수포자의 찍기 패턴 순서를 활용
        //   answer와 1번 수포자 비교
        //   answer와 2번 수포자 비교
        //   answer와 3번 수포자 비교
        //
        for (int i = 0; i < answers.length; i++) {
            int answer = answers[i];
            int first = firstPattern[i % firstPattern.length];
            int second = secondPattern[i % secondPattern.length];
            int third = thirdPattern[i % thirdPattern.length];

            if (answer == first) {
                firstCount++;
            }
            if (answer == second) {
                secondCount++;
            }
            if (answer == third) {
                thirdCount++;
            }
        }

        List<Integer> scoreList =
                Arrays.asList(firstCount, secondCount, thirdCount);
        // scoreList.sort((a, b) -> b.compareTo(a));
        scoreList.sort(Comparator.reverseOrder());
        int maxScore = scoreList.get(0);

        ArrayList<Integer> topUsers = new ArrayList<>();

        if (maxScore == firstCount) {
            topUsers.add(1);
        }
        if (maxScore == secondCount) {
            topUsers.add(2);
        }
        if (maxScore == thirdCount) {
            topUsers.add(3);
        }


        return topUsers.stream().mapToInt(i -> i).toArray();
    }
}
