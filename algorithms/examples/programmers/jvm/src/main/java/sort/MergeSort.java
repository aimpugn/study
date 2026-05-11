package sort;

/**
 * 병합 정렬은 배열을 계속 반으로 나누어 길이 1짜리 구간까지 내려간 뒤,
 * 이미 정렬된 두 구간을 하나의 정렬된 구간으로 합치며 올라오는 정렬입니다.
 *
 * <pre>
 * 입력: [5, 1, 4, 2]
 *
 * 나누기:
 * [5, 1, 4, 2]
 * [5, 1] [4, 2]
 * [5] [1] [4] [2]
 *
 * 합치기:
 * [5] + [1] -> [1, 5]
 * [4] + [2] -> [2, 4]
 * [1, 5] + [2, 4]
 *   후보 1과 2 중 1 선택 -> [1, _, _, _]
 *   후보 5와 2 중 2 선택 -> [1, 2, _, _]
 *   후보 5와 4 중 4 선택 -> [1, 2, 4, _]
 *   남은 5 복사          -> [1, 2, 4, 5]
 * </pre>
 *
 * 이 구현은 구간을 {@code [left, rightExclusive)}로 표현합니다.
 * 오른쪽 끝을 포함하지 않는 구간이라 길이는 {@code rightExclusive - left}가 되고,
 * 왼쪽 절반은 {@code [left, mid)}, 오른쪽 절반은 {@code [mid, rightExclusive)}로
 * 겹치지 않게 나뉩니다.
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
        // 병합할 때마다 새 배열을 만들면 메모리 할당이 반복됩니다.
        // 전체 크기의 buffer 하나를 공유하면 병합 결과를 잠시 담았다가 원본 구간에 되돌릴 수 있습니다.
        int[] buffer = new int[values.length];
        sort(values, buffer, 0, values.length);
    }

    private static void sort(int[] values, int[] buffer, int left, int rightExclusive) {
        // 길이가 0 또는 1인 구간은 이미 정렬되어 있습니다.
        // 재귀의 바닥 조건이므로 여기서 더 나누지 않습니다.
        if (rightExclusive - left < 2) {
            return;
        }

        int mid = left + (rightExclusive - left) / 2;

        // 먼저 두 절반을 각각 정렬해 둡니다.
        // merge는 "이미 정렬된 두 구간"이라는 전제가 있어야 선형 시간으로 합칠 수 있습니다.
        sort(values, buffer, left, mid);
        sort(values, buffer, mid, rightExclusive);
        merge(values, buffer, left, mid, rightExclusive);
    }

    private static void merge(int[] values, int[] buffer, int left, int mid, int rightExclusive) {
        int leftCursor = left;
        int rightCursor = mid;
        int write = left;

        // values[left..mid)와 values[mid..rightExclusive)는 각각 이미 정렬되어 있습니다.
        // 그래서 각 구간의 맨 앞 값만 비교하면, 두 구간 전체에서 가장 작은 남은 값을 알 수 있습니다.
        //
        // left half:  [1, 5]
        // right half: [2, 4]
        // 후보 비교: 1 vs 2 -> 1을 buffer에 쓰고 leftCursor만 이동
        //
        // 같은 값이면 왼쪽 값을 먼저 쓰기 위해 <=를 사용합니다.
        // 이 선택이 병합 정렬의 안정성을 지켜 줍니다.
        while (leftCursor < mid && rightCursor < rightExclusive) {
            if (values[leftCursor] <= values[rightCursor]) {
                buffer[write++] = values[leftCursor++];
            } else {
                buffer[write++] = values[rightCursor++];
            }
        }

        // 한쪽 구간이 먼저 비면 다른 구간의 남은 값들은 이미 정렬된 상태입니다.
        // 그대로 뒤에 붙이면 buffer[left..rightExclusive)도 정렬된 구간이 됩니다.
        while (leftCursor < mid) {
            buffer[write++] = values[leftCursor++];
        }
        while (rightCursor < rightExclusive) {
            buffer[write++] = values[rightCursor++];
        }

        // buffer에 만든 정렬 결과를 원래 배열의 같은 구간으로 되돌립니다.
        // 다음 상위 merge는 이 구간이 정렬되어 있다는 전제를 소비합니다.
        for (int i = left; i < rightExclusive; i++) {
            values[i] = buffer[i];
        }
    }
}
