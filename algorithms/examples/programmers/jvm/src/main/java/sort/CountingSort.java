package sort;

/**
 * 계수 정렬은 값을 서로 비교하지 않고, 각 정수가 몇 번 나왔는지 센 뒤 작은 값부터
 * 나온 횟수만큼 다시 쓰는 정렬입니다.
 *
 * <pre>
 * 입력: [3, 1, 3, 2]
 *
 * min = 1, max = 3 이므로 count 배열은 값 1, 2, 3의 출현 횟수를 담습니다.
 *
 * value:  1  2  3
 * count: [1, 1, 2]
 *
 * 다시 쓰기:
 * value 1을 1번 -> [1, _, _, _]
 * value 2를 1번 -> [1, 2, _, _]
 * value 3을 2번 -> [1, 2, 3, 3]
 * </pre>
 *
 * 비교 정렬이 아니라서 입력 크기 n뿐 아니라 값의 범위 k도 비용에 들어갑니다.
 * 값이 0부터 100까지처럼 좁으면 빠르지만, 값이 -1_000_000_000부터 1_000_000_000까지
 * 넓게 흩어져 있으면 count 배열이 너무 커져 적합하지 않습니다.
 *
 * 시간: O(n + k). 여기서 k는 max - min + 1입니다.
 * 공간: O(k).
 * 안정 정렬: int 값만 다시 쓰는 이 예제에서는 관찰 대상이 아닙니다. 객체 정렬에서 안정성이
 * 필요하다면 누적 count와 별도 output 배열을 써서 같은 값의 원래 순서를 보존해야 합니다.
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

        // count 배열을 만들려면 먼저 값의 범위를 알아야 합니다.
        // min을 빼서 offset으로 바꾸면 음수 값도 count 배열 인덱스로 표현할 수 있습니다.
        //
        // 예: min = -2라면 value -2는 offset 0, value 0은 offset 2가 됩니다.
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

        // counts[value - min]은 value가 입력에서 몇 번 나왔는지를 기억합니다.
        // 이 배열이 "값 -> 출현 횟수" 표 역할을 하므로, 이후에는 값끼리 비교하지 않습니다.
        int[] counts = new int[(int) range];
        for (int value : values) {
            counts[value - min]++;
        }

        int write = 0;
        for (int offset = 0; offset < counts.length; offset++) {
            int value = min + offset;

            // offset을 작은 값부터 큰 값까지 훑기 때문에, 다시 쓰는 순서 자체가 오름차순입니다.
            // counts[offset]번만큼 value를 쓰고 나면 그 값의 모든 출현이 결과 배열에 반영됩니다.
            for (int remaining = counts[offset]; remaining > 0; remaining--) {
                values[write++] = value;
            }
        }
    }
}
