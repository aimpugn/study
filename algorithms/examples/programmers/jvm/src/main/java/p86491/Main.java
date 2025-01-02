package p86491;

public class Main {

    public static void main(String[] args) {
        System.out.println(Main.solution(
                new int[][]{{60, 50}, {30, 70}, {60, 30}, {80, 40}}) == 4000);

    }

    /**
     * <pre>
     * 목표: 모든 명함을 수납할 수 있는 가장 작은 지갑 만들기
     *
     * 지갑 크기 정하기
     * - 다양한 모양과 크기 명함 수납 가능
     * - 작아서 들고 다니기 편해야 함
     *
     * 가로 x 세로
     * 60 x 50
     * 30 x 70
     * 60 x 30
     * 80 x 40
     *
     * 가장 긴 가로 길이와 세로 길이가 각각 80, 70이므로 80 * 70 크기의 지갑을 만들면 모든 명함 수납 가능.
     * 하지만 30 x 70 경우 가로로 눕혀 수납할 수 있음.
     * 2번 케이스에 가로로 수납한다고 하면 80 x 50 크기의 지갑ㅂ으로 모든 명함 수납 가능.
     *
     * </pre>
     *
     * @param sizes 모든 명함의 가로 x 세로 길이 나타내는 배열. [{width, height},..] 형식. 1 <=
     *              sizes.length <= 10,000
     * @return int 지갑의 크기
     * @see <a href=
     * "https://school.programmers.co.kr/learn/courses/30/lessons/86491">최소
     * 직사각형</a>
     */
    public static int solution(int[][] sizes) {
        if (sizes.length == 1) {
            return sizes[0][0] * sizes[0][1];
        }

        // 가장 짧은 가로 길이,
        // 가잘 짧은 세로 길이
        int maxLong = Integer.MIN_VALUE;
        int maxShort = Integer.MIN_VALUE;

        for (int[] size : sizes) {
            int longSide = Math.max(size[0], size[1]);
            int shortSide = Math.min(size[0], size[1]);

            maxLong = Math.max(maxLong, longSide);
            maxShort = Math.max(maxShort, shortSide);
        }

        return maxLong * maxShort;
    }
}
