package sort;

/**
 * 퀵 정렬은 기준값 pivot을 하나 고른 뒤, pivot보다 작은 값은 왼쪽 방향으로,
 * 큰 값은 오른쪽 방향으로 보내도록 구간을 나누고 남은 두 구간을 다시 정렬합니다.
 *
 * <pre>
 * 입력: [5, 1, 4, 2, 8], pivot = 4
 *
 * i는 왼쪽에서 오른쪽으로, j는 오른쪽에서 왼쪽으로 움직입니다.
 *
 * [5, 1, 4, 2, 8]
 *  i           j
 * 5는 pivot보다 크고, 2는 pivot보다 작으므로 둘을 교환합니다.
 *
 * [2, 1, 4, 5, 8]
 *        i
 *        j
 * 4는 pivot과 같으므로 제자리에 둔 채 i와 j를 지나가게 합니다.
 *
 * 분할 뒤에는 왼쪽 구간에 pivot보다 큰 값이 남지 않고,
 * 오른쪽 구간에 pivot보다 작은 값이 남지 않습니다.
 * 두 구간 내부는 아직 완전히 정렬되지 않았으므로 재귀가 이어집니다.
 * </pre>
 *
 * 이 구현의 pivot은 "기준값"입니다. pivot 값이 들어 있던 실제 원소를 마지막 위치로
 * 고정하는 방식이 아니라, 두 포인터가 값을 교환하며 경계를 만드는 방식입니다.
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
        // 정렬할 원소가 0개 또는 1개인 구간은 이미 정렬되어 있습니다.
        // 재귀 호출이 더 작아지다가 여기서 멈춥니다.
        if (left >= right) {
            return;
        }

        int pivot = values[left + (right - left) / 2];
        int i = left;
        int j = right;

        // i는 왼쪽에서 pivot보다 작아 이미 제자리에 가까운 값들을 건너뜁니다.
        // j는 오른쪽에서 pivot보다 커 이미 제자리에 가까운 값들을 건너뜁니다.
        //
        // i와 j가 멈췄다는 것은 왼쪽에는 너무 큰 값, 오른쪽에는 너무 작은 값이
        // 발견됐다는 뜻입니다. 둘을 교환하면 두 값 모두 더 맞는 쪽으로 이동합니다.
        //
        // 반복이 끝나면:
        //   left..j  : pivot보다 큰 값이 경계 밖으로 밀려난 왼쪽 후보 구간
        //   i..right : pivot보다 작은 값이 경계 밖으로 밀려난 오른쪽 후보 구간
        // 각 구간 내부의 순서는 아직 보장하지 않으므로 아래 재귀 호출이 필요합니다.
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

        // 분할이 끝난 뒤 남은 두 구간을 같은 방식으로 다시 정렬합니다.
        // 이미 엇갈린 경계 때문에 left..j와 i..right는 서로 겹치지 않습니다.
        quickSort(values, left, j);
        quickSort(values, i, right);
    }
}
