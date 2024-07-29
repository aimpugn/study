package main.java.perfect_square;

public class Main {
    /**
     * <pre>
     * 완전 제곱수 판별 알고리즘
     *
     * 이 알고리즘은 주어진 정수가 완전 제곱수인지 판별합니다.
     * 1부터 순차적으로 제곱하는 방식 대신,
     * 이진 탐색 기법을 사용하여 O(log n) 시간 복잡도로 효율적으로 판별합니다.
     *
     * 알고리즘의 주요 단계:
     * 1. 탐색 범위 초기화 (1부터 입력값까지)
     * 2. 이진 탐색을 통한 가능한 제곱근 탐색
     * 3. 중간값의 제곱과 입력값 비교
     * 4. 탐색 범위 조정 및 반복
     *
     * - 시간 복잡도: O(log n): 여기서 n은 입력값입니다.
     * - 공간 복잡도: O(1): 입력 크기와 관계없이 일정한 추가 공간만 사용합니다.
     * </pre>
     *
     * @param num 완전 제곱수인지 판별할 양의 정수
     * @return 입력값이 완전 제곱수이면 true, 아니면 false
     */
    public static boolean isPerfectSquare(int num) {
        // 입력값이 1 미만인 경우 예외 처리
        if (num < 1) {
            throw new IllegalArgumentException("입력값은 1 이상의 양의 정수여야 합니다.");
        }

        // 1은 완전 제곱수입니다 (1^2 = 1)
        if (num == 1) {
            return true;
        }

        // 이진 탐색을 위한 범위 초기화
        long left = 1;
        long right = num;

        // 이진 탐색 시작
        while (left <= right) {
            // 중간값 계산. (left + right) / 2 대신 사용하여 정수 오버플로우 방지
            long mid = left + (right - left) / 2;

            // 중간값의 제곱 계산
            long square = mid * mid;

            // 중간값의 제곱과 입력값 비교
            if (square == num) {
                // 완전 제곱수를 찾은 경우
                return true;
            } else if (square < num) {
                // 중간값의 제곱이 입력값보다 작은 경우, 왼쪽 범위 조정
                left = mid + 1;
            } else {
                // 중간값의 제곱이 입력값보다 큰 경우, 오른쪽 범위 조정
                right = mid - 1;
            }
        }

        // 완전 제곱수를 찾지 못한 경우
        return false;
    }

    /**
     * <pre>
     * 시간 복잡도: O(√n): n의 제곱근까지 순차적으로 검사합니다.
     *
     * - 입력값이 1,000,000일 경우: 최대 1,000번의 반복 필요
     * - 입력값이 1,000,000,000일 경우: 최대 31,623번의 반복 필요
     * </pre>
     *
     * @param num
     * @return
     */
    public boolean isPerfectSquareLinear(int num) {
        for (int i = 1; i * i <= num; i++) {
            if (i * i == num)
                return true;
        }
        return false;
    }

    // 메인 메서드: 예제를 실행하고 결과를 출력합니다.
    public static void main(String[] args) {
        int[] testCases = {1, 4, 8, 9, 16, 25, 64, 100, 121, 10000, 10001};

        for (int num : testCases) {
            boolean result = isPerfectSquare(num);
            System.out.println(
                    num + "은(는) 완전 제곱수" + (result ? "입니다." : "가 아닙니다."));
        }
    }

}
