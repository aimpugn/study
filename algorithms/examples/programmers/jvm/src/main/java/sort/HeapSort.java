package sort;

/**
 * 힙 정렬은 먼저 max heap을 만든 뒤, heap의 root를 정렬된 오른쪽 구간으로
 * 반복해서 보냅니다.
 *
 * 시간: O(n log n).
 * 공간: O(1).
 * 안정 정렬: 아니요.
 * 제자리 정렬: 예.
 */
public final class HeapSort {
    private HeapSort() {
    }

    public static void sort(int[] values) {
        buildMaxHeap(values);

        for (int end = values.length - 1; end > 0; end--) {
            // values[0]은 heap 구간의 최댓값입니다.
            // 이 값을 end로 보내면 end..last 구간은 오른쪽부터 정렬된 suffix가 됩니다.
            SortSupport.swap(values, 0, end);
            siftDown(values, 0, end);
        }
    }

    private static void buildMaxHeap(int[] values) {
        for (int parent = values.length / 2 - 1; parent >= 0; parent--) {
            siftDown(values, parent, values.length);
        }
    }

    private static void siftDown(int[] values, int root, int heapSize) {
        while (true) {
            int leftChild = root * 2 + 1;
            int rightChild = leftChild + 1;
            int largest = root;

            if (leftChild < heapSize && values[leftChild] > values[largest]) {
                largest = leftChild;
            }
            if (rightChild < heapSize && values[rightChild] > values[largest]) {
                largest = rightChild;
            }
            if (largest == root) {
                return;
            }

            // root가 두 자식보다 크거나 같아질 때까지 아래로 내려보냅니다.
            // 이 과정을 마치면 root에서 시작한 부분 트리가 다시 max heap 조건을 만족합니다.
            SortSupport.swap(values, root, largest);
            root = largest;
        }
    }
}
