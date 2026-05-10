package sort;

/**
 * 퀵 정렬은 pivot을 기준으로 작은 값 쪽과 큰 값 쪽의 경계를 만들고,
 * 남은 두 구간을 재귀적으로 정렬합니다.
 *
 * 시간: 평균 O(n log n), pivot이 계속 나쁘게 갈리면 최악 O(n^2).
 * 공간: 평균 O(log n) 재귀 깊이, 최악 O(n).
 * 안정 정렬: 아니요.
 * 제자리 정렬: 예.
 */
public final class QuickSort {
    private QuickSort() {
    }

    public static void sort(int[] values) {
        quickSort(values, 0, values.length - 1);
    }

    private static void quickSort(int[] values, int left, int right) {
        if (left >= right) {
            return;
        }

        int pivot = values[left + (right - left) / 2];
        int i = left;
        int j = right;

        // i는 pivot보다 작은 값들이 모인 왼쪽 경계를 지나가고,
        // j는 pivot보다 큰 값들이 모인 오른쪽 경계를 지나갑니다.
        // 둘이 엇갈리면 left..j와 i..right 두 구간만 다시 정렬하면 됩니다.
        while (i <= j) {
            while (values[i] < pivot) {
                i++;
            }
            while (values[j] > pivot) {
                j--;
            }
            if (i <= j) {
                SortSupport.swap(values, i, j);
                i++;
                j--;
            }
        }

        quickSort(values, left, j);
        quickSort(values, i, right);
    }
}
