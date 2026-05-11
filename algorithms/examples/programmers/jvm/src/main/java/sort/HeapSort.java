package sort;

/**
 * 힙 정렬은 배열을 먼저 max heap으로 바꾼 뒤, heap의 root에 있는 최댓값을
 * 배열 오른쪽 끝으로 하나씩 보내며 정렬합니다.
 *
 * max heap은 부모가 자식보다 크거나 같은 완전 이진 트리입니다.
 * 이 구현에서는 별도 트리 객체를 만들지 않고 배열 인덱스로 트리를 표현합니다.
 *
 * <pre>
 * index:  0  1  2  3  4
 * value: [8, 5, 4, 2, 1]
 *
 * 0번의 자식: 1번, 2번
 * 1번의 자식: 3번, 4번
 *
 * heap root인 0번에는 항상 heap 구간의 최댓값이 있습니다.
 * [8, 5, 4, 2, 1]  root 8을 end로 교환  -> [1, 5, 4, 2 | 8]
 * [1, 5, 4, 2]     1을 아래로 내려 heap 복구 -> [5, 2, 4, 1 | 8]
 * </pre>
 *
 * 오른쪽의 {@code |} 뒤쪽은 정렬이 끝난 suffix이고, 왼쪽은 아직 heap으로 유지할 구간입니다.
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
            // values[0]은 현재 heap 구간의 최댓값입니다.
            // 이 값을 end로 보내면 end..last 구간은 오른쪽부터 정렬된 suffix가 됩니다.
            //
            // 교환 직후 0번에는 작은 값이 올라올 수 있으므로,
            // heapSize를 end로 줄인 뒤 siftDown으로 heap 조건을 다시 맞춥니다.
            SortSupport.swap(values, 0, end);
            siftDown(values, 0, end);
        }
    }

    private static void buildMaxHeap(int[] values) {
        // 마지막 부모부터 거꾸로 내려오며 siftDown을 하면 전체 배열이 max heap이 됩니다.
        // values.length / 2 이후의 인덱스는 자식이 없는 leaf라서 이미 heap 조건을 만족합니다.
        for (int parent = values.length / 2 - 1; parent >= 0; parent--) {
            siftDown(values, parent, values.length);
        }
    }

    private static void siftDown(int[] values, int root, int heapSize) {
        while (true) {
            int leftChild = root * 2 + 1;
            int rightChild = leftChild + 1;
            int largest = root;

            // root와 두 자식 중 가장 큰 위치를 찾습니다.
            // 가장 큰 값이 root가 아니라면 부모-자식 heap 조건이 깨져 있는 상태입니다.
            if (leftChild < heapSize && values[leftChild] > values[largest]) {
                largest = leftChild;
            }
            if (rightChild < heapSize && values[rightChild] > values[largest]) {
                largest = rightChild;
            }
            if (largest == root) {
                return;
            }

            // root가 두 자식보다 크거나 같아질 때까지 더 큰 자식과 교환하며 아래로 내려갑니다.
            // 한 번 내려간 뒤에는 그 자식 위치의 하위 트리에서 heap 조건이 깨질 수 있으므로
            // root를 largest로 옮겨 같은 검사를 반복합니다.
            SortSupport.swap(values, root, largest);
            root = largest;
        }
    }
}
