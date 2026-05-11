package sort;

/**
 * 선택 정렬은 아직 확정되지 않은 구간에서 가장 작은 값을 직접 골라 왼쪽부터 고정합니다.
 * 이름 그대로 "이번 자리에 들어갈 값을 선택한다"는 생각으로 읽으면 됩니다.
 *
 * <pre>
 * 입력: [5, 1, 4, 2]
 *
 * fixed = 0, 0번 자리에 들어갈 최솟값을 0..3에서 찾습니다.
 * [5, 1, 4, 2]  최솟값 1을 찾고 0번과 교환  -> [1 | 5, 4, 2]
 *
 * fixed = 1, 1번 자리에 들어갈 다음 최솟값을 1..3에서 찾습니다.
 * [1 | 5, 4, 2]  최솟값 2를 찾고 1번과 교환  -> [1, 2 | 4, 5]
 * </pre>
 *
 * 핵심 불변식은 {@code values[0..fixed-1]}이 이미 "가장 작은 값들부터 순서대로"
 * 확정된 prefix라는 점입니다. prefix는 배열의 왼쪽 끝에 붙은 확정 구간입니다.
 *
 * 시간: O(n^2).
 * 공간: O(1).
 * 안정 정렬: 아니요. 이 swap 기반 구현은 같은 값의 원래 순서를 보장하지 않습니다.
 * 제자리 정렬: 예.
 */
public final class SelectionSort {
    private SelectionSort() {
    }

    public static void sort(int[] values) {
        for (int fixed = 0; fixed < values.length - 1; fixed++) {
            int minIndex = fixed;

            // fixed 왼쪽은 더 이상 볼 필요가 없습니다.
            // 이번 반복의 일은 fixed..last 구간을 훑어서 fixed 자리에 들어갈 다음 최솟값의 위치를
            // minIndex에 기억하는 것입니다.
            //
            // [확정 prefix 0..fixed-1 | 후보 구간 fixed..last]
            for (int candidate = fixed + 1; candidate < values.length; candidate++) {
                if (values[candidate] < values[minIndex]) {
                    minIndex = candidate;
                }
            }

            // 최솟값을 찾은 뒤에는 한 번만 교환합니다.
            // 이 단순한 교환 때문에 같은 값을 가진 "객체"를 정렬한다면 안정성이 깨질 수 있습니다.
            // 예를 들어 [2a, 2b, 1]에서 1을 앞으로 보내면 2a와 2b의 상대 순서가 밀릴 수 있습니다.
            SortSupport.swap(values, fixed, minIndex);
        }
    }
}
