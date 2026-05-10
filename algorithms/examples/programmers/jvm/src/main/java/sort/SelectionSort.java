package sort;

/**
 * 선택 정렬은 아직 확정되지 않은 구간에서 가장 작은 값을 골라,
 * 왼쪽 자리부터 하나씩 확정합니다.
 *
 * 시간: O(n^2).
 * 공간: O(1).
 * 안정 정렬: 아니요. 이 swap 기반 구현은 같은 값의 원래 순서를 보장하지 않습니다.
 * 제자리 정렬: 예.
 */
public final class SelectionSort {
    private SelectionSort() {
    }

    public static void sort(int[] values) {
        for (int fixed = 0; fixed < values.length - 1; fixed++) {
            int minIndex = fixed;

            // values[0..fixed-1]은 이미 최솟값부터 차례대로 확정된 구간입니다.
            // 이번 반복은 fixed 자리에 들어갈 다음 최솟값을 뒤쪽에서 찾습니다.
            for (int candidate = fixed + 1; candidate < values.length; candidate++) {
                if (values[candidate] < values[minIndex]) {
                    minIndex = candidate;
                }
            }

            SortSupport.swap(values, fixed, minIndex);
        }
    }
}
