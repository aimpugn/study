package sort;

/**
 * 병합 정렬은 배열을 반으로 나누고, 이미 정렬된 두 구간을 다시 하나로 합칩니다.
 *
 * 시간: O(n log n).
 * 공간: O(n). 병합용 임시 배열을 하나 공유합니다.
 * 안정 정렬: 예. 같은 값이면 왼쪽 구간의 값을 먼저 씁니다.
 * 제자리 정렬: 아니요.
 */
public final class MergeSort {
    private MergeSort() {
    }

    public static void sort(int[] values) {
        int[] buffer = new int[values.length];
        sort(values, buffer, 0, values.length);
    }

    private static void sort(int[] values, int[] buffer, int left, int rightExclusive) {
        if (rightExclusive - left < 2) {
            return;
        }

        int mid = left + (rightExclusive - left) / 2;
        sort(values, buffer, left, mid);
        sort(values, buffer, mid, rightExclusive);
        merge(values, buffer, left, mid, rightExclusive);
    }

    private static void merge(int[] values, int[] buffer, int left, int mid, int rightExclusive) {
        int leftCursor = left;
        int rightCursor = mid;
        int write = left;

        // values[left..mid)와 values[mid..rightExclusive)는 각각 이미 정렬되어 있습니다.
        // 두 구간의 맨 앞 후보 중 작은 값을 buffer에 쓰면, buffer도 왼쪽부터 정렬됩니다.
        while (leftCursor < mid && rightCursor < rightExclusive) {
            if (values[leftCursor] <= values[rightCursor]) {
                buffer[write++] = values[leftCursor++];
            } else {
                buffer[write++] = values[rightCursor++];
            }
        }

        while (leftCursor < mid) {
            buffer[write++] = values[leftCursor++];
        }
        while (rightCursor < rightExclusive) {
            buffer[write++] = values[rightCursor++];
        }

        for (int i = left; i < rightExclusive; i++) {
            values[i] = buffer[i];
        }
    }
}
