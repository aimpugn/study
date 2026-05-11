package sort;

/**
 * 버블 정렬은 서로 붙어 있는 두 칸만 비교하면서 큰 값을 오른쪽으로 밀어내는 정렬입니다.
 * 한 번의 pass가 끝날 때마다 아직 확정되지 않은 값 중 가장 큰 값이 오른쪽 끝에 고정됩니다.
 *
 * <pre>
 * 입력: [5, 1, 4, 2]
 *
 * 첫 번째 pass, end = 3
 * [5, 1, 4, 2]  5와 1을 비교하고 교환  -> [1, 5, 4, 2]
 * [1, 5, 4, 2]  5와 4를 비교하고 교환  -> [1, 4, 5, 2]
 * [1, 4, 5, 2]  5와 2를 비교하고 교환  -> [1, 4, 2 | 5]
 *
 * 오른쪽의 5는 이번 pass에서 가장 큰 값이므로 다음 pass에서 다시 볼 필요가 없습니다.
 * </pre>
 *
 * 핵심 불변식은 {@code values[end + 1..last]}가 이미 정렬된 suffix라는 점입니다.
 * suffix는 배열의 오른쪽 끝에 붙은 확정 구간을 뜻합니다.
 *
 * 시간: O(n^2), swap이 없으면 멈추는 이 구현의 최선은 O(n).
 * 공간: O(1).
 * 안정 정렬: 예. 같은 값은 swap하지 않습니다.
 * 제자리 정렬: 예.
 */
public final class BubbleSort {
    private BubbleSort() {
    }

    public static void sort(int[] values) {
        for (int end = values.length - 1; end > 0; end--) {
            boolean swapped = false;

            // i와 i + 1만 보면서 큰 값을 오른쪽으로 한 칸씩 보냅니다.
            // end를 넘어선 구간은 이미 확정된 suffix이므로 건드리지 않습니다.
            //
            // pass가 시작될 때:
            //   [아직 정렬해야 하는 구간 0..end | 확정된 suffix end+1..last]
            // pass가 끝날 때:
            //   [남은 미정렬 구간 0..end-1 | 이번에 확정된 최댓값 end | 기존 suffix]
            for (int i = 0; i < end; i++) {
                if (values[i] > values[i + 1]) {
                    SortSupport.swap(values, i, i + 1);
                    swapped = true;
                }
            }

            // 한 번도 바꾸지 않았다는 것은 모든 이웃 쌍이 이미 왼쪽 <= 오른쪽이라는 뜻입니다.
            // 이웃한 곳마다 역전이 없으면 전체 배열에도 더 이상 역전이 남아 있지 않으므로,
            // 남은 pass를 생략해도 결과가 달라지지 않습니다.
            if (!swapped) {
                return;
            }
        }
    }
}
