package sort;

final class SortSupport {
    private SortSupport() {
    }

    static void swap(int[] values, int left, int right) {
        int temp = values[left];
        values[left] = values[right];
        values[right] = temp;
    }
}
