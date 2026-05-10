package sort;

/**
 * 버블 정렬은 이웃한 두 값을 비교하면서, 아직 확정되지 않은 값 중 가장 큰 값을
 * 오른쪽 끝으로 밀어냅니다.
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

            // 이 pass가 끝나면 values[end]에는 아직 확정되지 않은 값 중 가장 큰 값이 놓입니다.
            // 그래서 다음 pass는 end 오른쪽을 보지 않아도 됩니다.
            for (int i = 0; i < end; i++) {
                if (values[i] > values[i + 1]) {
                    SortSupport.swap(values, i, i + 1);
                    swapped = true;
                }
            }

            // 한 번도 바꾸지 않았다면 모든 이웃이 이미 오름차순입니다.
            // 남은 pass를 더 돌려도 배열은 달라지지 않습니다.
            if (!swapped) {
                return;
            }
        }
    }
}
