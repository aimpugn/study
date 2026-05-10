package sort;

/**
 * 삽입 정렬은 왼쪽을 이미 정렬된 구간으로 보고, 다음 값을 그 구간의 제자리에
 * 끼워 넣습니다.
 *
 * 시간: O(n^2), 거의 정렬된 입력에서는 O(n)에 가깝습니다.
 * 공간: O(1).
 * 안정 정렬: 예. 같은 값은 서로를 지나 밀리지 않습니다.
 * 제자리 정렬: 예.
 */
public final class InsertionSort {
    private InsertionSort() {
    }

    public static void sort(int[] values) {
        for (int next = 1; next < values.length; next++) {
            int valueToInsert = values[next];
            int cursor = next - 1;

            // values[0..next-1]은 이미 정렬된 구간입니다.
            // valueToInsert보다 큰 값만 한 칸씩 오른쪽으로 밀면, 빈자리가 곧 삽입 위치가 됩니다.
            while (cursor >= 0 && values[cursor] > valueToInsert) {
                values[cursor + 1] = values[cursor];
                cursor--;
            }

            values[cursor + 1] = valueToInsert;
        }
    }
}
