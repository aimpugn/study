package p42747;

import java.util.Arrays;
import java.util.List;

/**
 * [H-Index](https://school.programmers.co.kr/learn/courses/30/lessons/42747?language=java)
 */
class Solution {
    /**
     * H-Index는 과학자의 생산성과 영향력을 나타내는 지표입니다.
     * <p>
     * 발표한 논문 `n`편 중,
     * - `h` 번 이상 인용된 논문이 `h`편 이상이고,
     * - 나머지 논문이 `h`번 이하 인용되었다면,
     * - `h`의 최댓값이 H-Index 입니다.
     *
     * @param citations 어떤 과학자가 발표한 논문의 인용 횟수를 담은 배열
     *                  - 1 <= citations.length <= 1,000
     *                  - 0 <= citations[i] <= 10,000
     * @return 어느 과학자의 H-Index를 나타내는 값
     */
    public int solution(int[] citations) {
        // citations: [3, 0, 6, 1, 5], H-Index: 3
        // - 발표한 논문의 수: 5편
        // - 3편의 논문은 3회 이상 인용
        // - 나머지 2편의 논문은 3회 이하 인용
        // - H-Index: 3
        //
        // h번 인용된 논문이 h편 이상이어야 하므로,
        // 일단 내림차순 정렬이 되어야 할 것으로 보입니다.
        // sortByDescending(citations)
        quickSort(citations, 0, citations.length - 1);

        // 그리고 h번 이상 인용되었는지를 찾아야 합니다.
        // - `h` 번 이상 인용된 논문이 `h`편 이상이고,
        // - 나머지 논문이 `h`번 이하 인용되었다면,
        //
        // 이때 `h`는 "인용 횟수"와 "논문 편수"를 동시에 고려해야 합니다.
        // - "`h`번 이상 인용"되어야 하므로 0 <= `h` <= 10,000 입니다.
        // - 하지만, "`h`편 이상"이어야 하므로 1 <= `h` <= 1,000 입니다.
        // 즉, `h` 값의 범위는 `citations` 배열의 요소 개수, 즉 `citations.length` 내입니다.
        // => `0 ~ citations.length` 범위 내에서 최대인 `h` 값을 찾기 위해 반복문을 순회합니다.
        //
        // 더 구체적으로 본다면, 앞의 예제의 경우 다음과 같이 내림차순으로 정렬됩니다.
        // [6, 5, 3, 1, 0]
        // - `i=0`: `citations[0]=6`회 인용  -> "최소 1편은 1회 이상"
        // - `i=1`: `citations[1]=5`회 인용  -> "최소 2편은 2회 이상"
        // - `i=2`: `citations[2]=3`회 인용  -> "최소 3편은 3회 이상"
        // <내림차순 정렬이므로 반복문을 더 진행해봤자 `i`보다 작은 "인용 횟수"들만 나옵니다.>
        // - `i=3`: `citations[3]=1`회 인용  -> "최소 4편은 4회 이상"(불가능)
        // - `i=4`: `citations[4]=0`회 인용  -> "최소 5편은 5회 이상"(불가능)
        //
        // 이를 일반화하면, "`i=N`: ?회 인용  -> 최소 N+1편은 N+1회 이상" 여부를 판단하면 됩니다.
        // 순회하는 데 사용되는 반복문 변수 `i`(인덱스)가 H-Index(`h`) 후보가 됩니다.
        // - (0+1)편은 (0+1)회 이상
        // - (1+1)편은 (1+1)회 이상
        // - ... 생략 ...
        // - (i+1)편은 (i+1)회 이상.
        //
        // 위의 예제를 보면 `i=2`일 때 멈춰야 합니다.
        // 내림차순 정렬이므로 반복문을 더 진행해봤자, 즉, `i`를 증가시켜 봤자, `i`보다 작은 "인용 횟수"들만 나오기 때문입니다.
        // `i`는 인덱스이고 "(i+1)편 이상"이어야 하므로, `i+1`을 리턴합니다.
        //
        // Before:
        // ```
        // int answer = -1;
        // for (int i = -1; i < citations.length - 1; i++) {
        //     if (citations[i + 1] <= i + 1) {
        //         break;
        //     }
        //     answer++;
        // }
        //
        // return answer + 1;
        // ```

        for (int i = 0; i < citations.length; i++) {
            if (citations[i] <= i) {
                // answer++을 별도로 하지 않고, i 값 그대로 반환합니다.
                return i;
            }
        }

        return citations.length;
    }

    /**
     * 내림차순 퀵정렬 메서드입니다.
     *
     * @param arr  정렬할 배열
     * @param low  낮은 인덱스
     * @param high 높은 인덱스
     */
    protected void quickSort(int[] arr, int low, int high) {
        // 범위를 벗어나면 종료합니다.
        if (low >= high || low < 0) return;

        int partition = partition(arr, low, high);
        quickSort(arr, low, partition - 1);
        quickSort(arr, partition + 1, high);
    }

    /**
     * lomuto partition 방식을 구현합니다.
     *
     * @param arr  논문 인용 배열
     * @param low  시작 인덱스
     * @param high 끝 인덱스
     * @return 파티션을 나누는 인덱스를 리턴합니다. {@link Solution#quickSort}에서 사용됩니다.
     * @see <a href="https://en.wikipedia.org/wiki/Quicksort#Lomuto_partition_scheme">Lomuto partition scheme</a>
     */
    private int partition(int[] arr, int low, int high) {
        int pivot = arr[high];
        // 이 변수는 임시 피벗 인덱스로, 파티션이 끝난 후 리턴할 값입니다.
        int temporaryPivotIdx = low;

        for (int i = low; i < high; i++) {
            // if (arr[i] <= pivot) { // 정방향 정렬
            if (arr[i] >= pivot) { // 역방향 정렬
                swap(arr, i, temporaryPivotIdx);
                temporaryPivotIdx++;
            }
        }
        swap(arr, temporaryPivotIdx, high);

        return temporaryPivotIdx;
    }

    private void swap(int[] arr, int a, int b) {
        int tmp = arr[a];
        arr[a] = arr[b];
        arr[b] = tmp;
    }
}

class TestCase {
    public int[] arr;
    public int expected;

    public TestCase(int[] arr, int expected) {
        this.arr = arr;
        this.expected = expected;
    }
}

public class Main {
    public static void main(String[] args) {
        Solution s = new Solution();
        List<int[]> cases = List.of(
                new int[]{1, 2, 3, 4, 5},
                new int[]{5, 3, 4, 1, 2},
                new int[]{1, 2, 1, 2, 1}
        );
        cases.forEach((tc) -> {
            s.quickSort(tc, 0, tc.length - 1);
            System.out.println(Arrays.toString(tc));
        });

        List<TestCase> testCases = List.of(
                new TestCase(new int[]{3, 0, 6, 1, 5}, 3),
                new TestCase(new int[]{1, 2, 3, 4, 5}, 3),
                new TestCase(new int[]{8}, 1),
                new TestCase(new int[]{3, 0, 6, 1, 5, 2}, 3),
                new TestCase(new int[]{3, 0, 6, 1, 5, 7}, 3),
                new TestCase(new int[]{4, 0, 6, 1, 5}, 3),
                new TestCase(new int[]{0, 0, 0, 0, 0}, 0),
                new TestCase(new int[]{100, 101, 102, 103, 104}, 5)
        );

        testCases.forEach((tc) -> {
            int answer = s.solution(Arrays.copyOf(tc.arr, tc.arr.length));
            System.out.printf("%s => %d(%b)\n", Arrays.toString(tc.arr), answer, answer == tc.expected);
        });

    }
}
