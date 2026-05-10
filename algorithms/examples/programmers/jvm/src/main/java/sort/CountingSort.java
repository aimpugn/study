package sort;

/**
 * 계수 정렬은 각 정수가 몇 번 나왔는지 센 뒤, 작은 값부터 count만큼 다시 씁니다.
 *
 * 시간: O(n + k). 여기서 k는 max - min + 1입니다.
 * 공간: O(k).
 * 안정 정렬: int 값만 다시 쓰는 이 예제에서는 관찰 대상이 아닙니다.
 * 제자리 정렬: 아니요. count 배열이 핵심 작업 공간입니다.
 */
public final class CountingSort {
    private static final int MAX_COUNTING_RANGE = 1_000_000;

    private CountingSort() {
    }

    public static void sort(int[] values) {
        if (values.length < 2) {
            return;
        }

        int min = values[0];
        int max = values[0];
        for (int value : values) {
            if (value < min) {
                min = value;
            }
            if (value > max) {
                max = value;
            }
        }

        long range = (long) max - min + 1;
        if (range > MAX_COUNTING_RANGE) {
            throw new IllegalArgumentException("value range is too wide for this counting sort example");
        }

        int[] counts = new int[(int) range];
        for (int value : values) {
            counts[value - min]++;
        }

        int write = 0;
        for (int offset = 0; offset < counts.length; offset++) {
            int value = min + offset;

            // counts[offset]는 value가 몇 번 나왔는지를 기억합니다.
            // 값을 비교하지 않고 count만큼 다시 쓰기 때문에, 값 범위가 작을 때만 이 방식이 유리합니다.
            for (int remaining = counts[offset]; remaining > 0; remaining--) {
                values[write++] = value;
            }
        }
    }
}
