package sort;

import java.util.Arrays;

public final class SortLearningRunner {
    private SortLearningRunner() {
    }

    public static void main(String[] args) {
        IntSorter[] sorters = {
            new IntSorter("bubble", BubbleSort::sort),
            new IntSorter("selection", SelectionSort::sort),
            new IntSorter("insertion", InsertionSort::sort),
            new IntSorter("merge", MergeSort::sort),
            new IntSorter("quick", QuickSort::sort),
            new IntSorter("heap", HeapSort::sort),
            new IntSorter("counting", CountingSort::sort)
        };

        int[][] cases = {
            {},
            {7},
            {1, 2, 3, 4, 5},
            {5, 4, 3, 2, 1},
            {3, 1, 2, 3, 1, 2},
            {-4, 0, 3, -1, 3, 2},
            {64, 34, 25, 12, 22, 11, 90},
            {9, 9, 9, 1, 0, -1, 8, 7, 6, 5}
        };

        for (IntSorter sorter : sorters) {
            for (int[] input : cases) {
                assertSorted(sorter, input);
            }
        }

        assertCountingSortRangeGuard();
        System.out.println("All sort learning tests passed");
    }

    private static void assertSorted(IntSorter sorter, int[] input) {
        int[] expected = input.clone();
        Arrays.sort(expected);

        int[] actual = input.clone();
        sorter.sort(actual);

        if (!Arrays.equals(expected, actual)) {
            throw new AssertionError(
                sorter.name + " failed: input=" + Arrays.toString(input)
                    + ", expected=" + Arrays.toString(expected)
                    + ", actual=" + Arrays.toString(actual)
            );
        }
    }

    private static void assertCountingSortRangeGuard() {
        try {
            CountingSort.sort(new int[]{-500_001, 500_000});
            throw new AssertionError("counting sort range guard did not fail");
        } catch (IllegalArgumentException expected) {
            // 값 범위가 너무 넓으면 count 배열이 정렬 대상보다 훨씬 커집니다.
            // 이 예제 구현은 그 상황을 조용히 진행하지 않고 명시적으로 거절합니다.
        }
    }

    private record IntSorter(String name, SortFunction sortFunction) {
        void sort(int[] values) {
            sortFunction.sort(values);
        }
    }

    @FunctionalInterface
    private interface SortFunction {
        void sort(int[] values);
    }
}
