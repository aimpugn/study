package sort;

/**
 * 삽입 정렬은 왼쪽을 이미 정렬된 구간으로 보고, 오른쪽에서 하나를 꺼내 그 구간의
 * 알맞은 위치에 끼워 넣는 정렬입니다. 손에 든 카드 한 장을 이미 정렬된 카드들 사이에
 * 넣는 장면과 거의 같습니다.
 *
 * <pre>
 * 입력: [5, 1, 4, 2]
 *
 * next = 1, valueToInsert = 1
 * 정렬된 왼쪽: [5] / 끼워 넣을 값: 1
 * [5, 1, 4, 2]  5가 1보다 크므로 5를 오른쪽으로 밀기  -> [5, 5, 4, 2]
 * 빈자리 0번에 1 삽입                           -> [1, 5, 4, 2]
 *
 * next = 2, valueToInsert = 4
 * [1, 5, 4, 2]  5가 4보다 크므로 5를 오른쪽으로 밀기  -> [1, 5, 5, 2]
 * 빈자리 1번에 4 삽입                           -> [1, 4, 5, 2]
 * </pre>
 *
 * 이동 중에 같은 값이 잠깐 두 번 보이는 것은 버그가 아닙니다.
 * {@code valueToInsert}를 지역 변수에 따로 보관해 두었기 때문에, 큰 값을 오른쪽으로
 * 복사하면서 빈자리를 만들 수 있습니다.
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
            // cursor는 그 구간의 오른쪽 끝에서 왼쪽으로 이동하며 삽입 위치를 찾습니다.
            //
            // valueToInsert보다 큰 값만 오른쪽으로 밀고, 같은 값은 밀지 않습니다.
            // 그래서 같은 값끼리는 원래 앞에 있던 값이 계속 앞에 남아 안정 정렬이 됩니다.
            while (cursor >= 0 && values[cursor] > valueToInsert) {
                values[cursor + 1] = values[cursor];
                cursor--;
            }

            // while이 멈추면 cursor는 배열 밖(-1)이거나 valueToInsert 이하의 값을 가리킵니다.
            // 따라서 실제 빈자리는 cursor 바로 오른쪽입니다.
            values[cursor + 1] = valueToInsert;
        }
    }
}
