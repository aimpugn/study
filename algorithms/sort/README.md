# 정렬 코드는 무엇을 하나씩 확정해 가는가

정렬 알고리즘을 다시 공부할 때 먼저 붙잡을 것은 이름 목록이 아니라, 배열 안에서 어느 값과 어느 구간이 믿을 수 있는 상태로 바뀌는지입니다. 같은 오름차순 결과를 만들더라도 버블 정렬은 오른쪽 끝부터 큰 값을 확정하고, 선택 정렬은 왼쪽부터 작은 값을 확정하며, 삽입 정렬은 왼쪽의 정렬된 구간을 한 칸씩 넓힙니다.

이 문서는 [sort 패키지](../examples/programmers/jvm/src/main/java/sort)의 Java 구현을 기준으로 정렬 코드의 상태 변화를 설명합니다. 코드를 읽을 때 `for` 문이 몇 번 도는지만 보면 금방 흐려집니다. 대신 `end`, `fixed`, `next`, `leftCursor`, `rightCursor`, `i`, `j`, `heapSize`, `write` 같은 변수가 배열의 어느 경계를 의미하는지 잡으면, 나중에 비슷한 문제를 만났을 때 다시 손으로 복원할 수 있습니다.

## 목차

- [이 문서에서 계속 따라갈 작은 배열](#이-문서에서-계속-따라갈-작은-배열)
- [정렬 코드를 읽는 공통 질문](#정렬-코드를-읽는-공통-질문)
- [이웃을 바꾸며 오른쪽을 확정하는 Bubble Sort](#이웃을-바꾸며-오른쪽을-확정하는-bubble-sort)
- [최솟값을 골라 왼쪽을 확정하는 Selection Sort](#최솟값을-골라-왼쪽을-확정하는-selection-sort)
- [왼쪽 정렬 구간에 값을 끼워 넣는 Insertion Sort](#왼쪽-정렬-구간에-값을-끼워-넣는-insertion-sort)
- [정렬된 두 구간을 다시 합치는 Merge Sort](#정렬된-두-구간을-다시-합치는-merge-sort)
- [pivot 값으로 경계를 가르는 Quick Sort](#pivot-값으로-경계를-가르는-quick-sort)
- [heap의 root를 오른쪽으로 보내는 Heap Sort](#heap의-root를-오른쪽으로-보내는-heap-sort)
- [값의 개수를 세어 다시 쓰는 Counting Sort](#값의-개수를-세어-다시-쓰는-counting-sort)
- [코딩 테스트에서 정렬을 볼 때 던질 질문](#코딩-테스트에서-정렬을-볼-때-던질-질문)
- [이번 코드에 없는 정렬은 왜 뺐나](#이번-코드에-없는-정렬은-왜-뺐나)
- [직접 실행하는 방법](#직접-실행하는-방법)
- [손으로 다시 그려 볼 연습](#손으로-다시-그려-볼-연습)

## 이 문서에서 계속 따라갈 작은 배열

대부분의 설명은 아래 배열을 기준으로 삼습니다.

```text
초기 상태: [5, 1, 4, 2, 8]
목표 상태: [1, 2, 4, 5, 8]
```

정렬을 "결과 배열을 만든다"라고만 기억하면 각 알고리즘의 차이가 사라집니다. 정렬 코드를 읽을 때 더 중요한 질문은 아래처럼 바뀝니다.

```text
어느 구간은 이미 확정됐는가?
아직 다시 봐야 하는 구간은 어디인가?
이번 반복에서 값 하나가 어디로 이동했는가?
다음 반복은 방금 바뀐 상태를 어떻게 소비하는가?
```

예를 들어 같은 `[5, 1, 4, 2, 8]`도 첫 반복 뒤에 알고리즘마다 다른 약속을 남깁니다.

```text
Bubble:    [1, 4, 2, 5 | 8]  오른쪽 8은 더 보지 않아도 된다.
Selection: [1 | 5, 4, 2, 8]  왼쪽 1은 더 보지 않아도 된다.
Insertion: [1, 5 | 4, 2, 8]  왼쪽 [1, 5]는 이미 정렬되어 있다.
```

이 약속을 불변식이라고 부릅니다. 불변식은 반복문이 돌 때마다 계속 참이어야 하는 조건입니다. 정렬 코드를 배울 때 불변식을 잡으면, 왜 특정 인덱스부터 다시 보지 않아도 되는지 설명할 수 있습니다.

## 정렬 코드를 읽는 공통 질문

이 패키지의 구현은 일부러 Java 표준 라이브러리 정렬을 호출하지 않습니다. 목적은 빠른 제출 코드가 아니라, 코딩 테스트에서 정렬이 문제 해결 도구로 등장할 때 내부 감각을 복원하는 것입니다.

| 알고리즘 | 구현 | 먼저 볼 경계 | 핵심 상태 변화 |
|---|---|---|---|
| Bubble Sort | [BubbleSort.java](../examples/programmers/jvm/src/main/java/sort/BubbleSort.java) | `end` | 큰 값이 이웃 swap으로 오른쪽 끝까지 밀린다. |
| Selection Sort | [SelectionSort.java](../examples/programmers/jvm/src/main/java/sort/SelectionSort.java) | `fixed` | 남은 구간의 최솟값이 왼쪽 확정 구간으로 온다. |
| Insertion Sort | [InsertionSort.java](../examples/programmers/jvm/src/main/java/sort/InsertionSort.java) | `next`, `cursor` | 큰 값들이 오른쪽으로 밀리고 빈자리에 새 값이 들어간다. |
| Merge Sort | [MergeSort.java](../examples/programmers/jvm/src/main/java/sort/MergeSort.java) | `left`, `mid`, `rightExclusive` | 정렬된 두 구간의 앞 후보를 비교해 buffer에 쓴다. |
| Quick Sort | [QuickSort.java](../examples/programmers/jvm/src/main/java/sort/QuickSort.java) | `i`, `j`, `pivot` | pivot 값보다 작은 쪽과 큰 쪽의 경계를 만든다. |
| Heap Sort | [HeapSort.java](../examples/programmers/jvm/src/main/java/sort/HeapSort.java) | `heapSize`, `end` | heap root의 최댓값을 오른쪽 확정 구간으로 보낸다. |
| Counting Sort | [CountingSort.java](../examples/programmers/jvm/src/main/java/sort/CountingSort.java) | `min`, `counts`, `write` | 값을 비교하지 않고 개수를 세어 작은 값부터 다시 쓴다. |

정렬 구현은 크게 두 부류로 나뉩니다. 비교 정렬은 두 값을 비교해서 순서를 정합니다. 버블, 선택, 삽입, 병합, 퀵, 힙 정렬이 여기에 들어갑니다. 계수 정렬은 값을 직접 비교하지 않고, 값의 범위가 작다는 전제를 이용해 각 값이 몇 번 나왔는지 셉니다.

## 이웃을 바꾸며 오른쪽을 확정하는 Bubble Sort

[BubbleSort.sort](../examples/programmers/jvm/src/main/java/sort/BubbleSort.java)는 이웃한 두 값 `values[i]`, `values[i + 1]`을 비교합니다. 왼쪽 값이 더 크면 두 값을 바꿉니다. 한 pass가 끝나면 아직 확정되지 않은 값 중 가장 큰 값이 `end` 위치로 이동합니다.

첫 pass를 손으로 따라가면 값이 어떻게 "오른쪽으로 밀리는지"가 보입니다.

```text
start: [5, 1, 4, 2, 8]

i=0: 5 > 1 이므로 swap
       [1, 5, 4, 2, 8]

i=1: 5 > 4 이므로 swap
       [1, 4, 5, 2, 8]

i=2: 5 > 2 이므로 swap
       [1, 4, 2, 5, 8]

i=3: 5 <= 8 이므로 유지
       [1, 4, 2, 5 | 8]
```

오른쪽의 `8`은 첫 pass가 끝난 뒤 확정됩니다. 다음 pass는 `end = 3`까지만 보면 됩니다.

```text
pass 1 이후: [1, 4, 2, 5 | 8]
pass 2 이후: [1, 2, 4 | 5, 8]
pass 3 이후: [1, 2 | 4, 5, 8]
```

이 공통 예제에서는 가장 큰 값 `8`이 처음부터 끝에 있었습니다. 큰 값이 실제로 끝까지 움직이는 모습만 따로 보면 아래와 같습니다.

```text
start: [5, 1, 4, 2]

5 > 1 -> [1, 5, 4, 2]
5 > 4 -> [1, 4, 5, 2]
5 > 2 -> [1, 4, 2 | 5]
```

이 구현에는 `swapped`가 있습니다. 어떤 pass에서 swap이 한 번도 일어나지 않았다면 모든 이웃 쌍이 이미 오름차순입니다. 그래서 남은 pass는 배열을 바꾸지 못하고, 바로 return해도 안전합니다.

```java
if (!swapped) {
    return;
}
```

버블 정렬은 원리 학습에는 좋지만, 보통 코딩 테스트에서 직접 쓸 정렬은 아닙니다. 그래도 "한 pass가 끝날 때마다 오른쪽 suffix가 확정된다"는 감각은 반복문 불변식을 익히기에 좋습니다.

## 최솟값을 골라 왼쪽을 확정하는 Selection Sort

[SelectionSort.sort](../examples/programmers/jvm/src/main/java/sort/SelectionSort.java)는 `fixed` 위치에 들어갈 값을 고릅니다. `fixed`부터 끝까지 훑으며 가장 작은 값의 위치 `minIndex`를 찾고, 그 값을 `fixed` 위치와 바꿉니다.

첫 두 pass를 보면 버블 정렬과 방향이 반대입니다.

```text
start: [5, 1, 4, 2, 8]

fixed=0:
    후보 구간: [5, 1, 4, 2, 8]
    최솟값: 1(index 1)
    swap values[0], values[1]
    결과: [1 | 5, 4, 2, 8]

fixed=1:
    확정 구간: [1]
    후보 구간: [5, 4, 2, 8]
    최솟값: 2(index 3)
    swap values[1], values[3]
    결과: [1, 2 | 4, 5, 8]
```

선택 정렬의 불변식은 `values[0..fixed-1]`이 이미 가장 작은 값부터 차례대로 확정된 구간이라는 점입니다. 그래서 `fixed` 왼쪽은 다시 볼 필요가 없습니다.

주의할 점은 안정성입니다. 안정 정렬은 같은 key를 가진 원소의 원래 순서를 보존하는 정렬입니다. 이 구현은 멀리 떨어진 값끼리 swap하므로 안정 정렬이 아닙니다. `int`만 보면 같은 값의 순서를 관찰하기 어렵지만, 실제 원소가 `(score, name)` 같은 record라면 같은 `score`끼리 원래 순서가 바뀔 수 있습니다.

## 왼쪽 정렬 구간에 값을 끼워 넣는 Insertion Sort

[InsertionSort.sort](../examples/programmers/jvm/src/main/java/sort/InsertionSort.java)는 왼쪽 구간을 이미 정렬된 작은 배열로 봅니다. `next` 위치의 값을 `valueToInsert`에 잠깐 빼 두고, 그 값보다 큰 원소들을 한 칸씩 오른쪽으로 밉니다.

`next = 1`일 때의 움직임은 아래와 같습니다.

```text
start: [5, 1, 4, 2, 8]
             ^
             valueToInsert = 1

cursor=0: values[0] = 5 > 1
    5를 오른쪽으로 한 칸 복사한다.
    [5, 5, 4, 2, 8]
     ^
     아직 여기에 바로 1을 쓰지 않는다.
     cursor를 더 왼쪽으로 이동한 뒤, 마지막 cursor + 1 위치에 쓴다.

cursor=-1:
    더 비교할 왼쪽 값이 없다.
    cursor + 1 = 0 위치에 1을 쓴다.
    [1, 5, 4, 2, 8]
```

`next = 2`에서는 `4`를 `[1, 5]` 안에 끼워 넣습니다.

```text
before: [1, 5, 4, 2, 8]
              ^
              valueToInsert = 4

5 > 4 이므로 5를 오른쪽으로 이동
        [1, 5, 5, 2, 8]

1 <= 4 이므로 멈추고 cursor + 1 위치에 4를 쓴다.
after:  [1, 4, 5, 2, 8]
```

삽입 정렬의 불변식은 `values[0..next-1]`이 이미 정렬되어 있다는 점입니다. 거의 정렬된 배열에서는 `while`이 금방 멈춥니다. 그래서 최악 시간은 `O(n^2)`이지만, 이미 정렬된 입력에서는 각 `next`마다 비교만 한 번 하고 지나가 `O(n)`에 가깝게 움직입니다.

## 정렬된 두 구간을 다시 합치는 Merge Sort

[MergeSort.sort](../examples/programmers/jvm/src/main/java/sort/MergeSort.java)는 배열을 반으로 나누고, 길이가 1인 구간까지 내려간 뒤, 이미 정렬된 두 구간을 다시 합칩니다. 이 구현은 구간을 `left` 이상 `rightExclusive` 미만으로 표현합니다. `rightExclusive`는 포함되지 않는 오른쪽 끝입니다.

병합의 핵심은 "두 구간이 각각 이미 정렬되어 있다"는 전제입니다. 그러면 각 구간의 맨 앞 후보만 비교해도 다음에 쓸 값을 고를 수 있습니다.

```text
왼쪽 구간: values[0..2) = [1, 5]
오른쪽 구간: values[2..4) = [2, 4]
buffer: [_, _, _, _]

leftCursor=0, rightCursor=2, write=0
    1 <= 2 이므로 왼쪽 1을 쓴다.
    buffer: [1, _, _, _]

leftCursor=1, rightCursor=2, write=1
    5 > 2 이므로 오른쪽 2를 쓴다.
    buffer: [1, 2, _, _]

leftCursor=1, rightCursor=3, write=2
    5 > 4 이므로 오른쪽 4를 쓴다.
    buffer: [1, 2, 4, _]

오른쪽 구간이 비었으므로 왼쪽에 남은 5를 쓴다.
    buffer: [1, 2, 4, 5]

마지막으로 buffer[0..4)를 values[0..4)에 복사한다.
```

`if (values[leftCursor] <= values[rightCursor])`에서 `<=`를 쓰는 것도 중요합니다. 두 값이 같을 때 왼쪽 구간의 값을 먼저 쓰면, 원래 앞에 있던 같은 key가 결과에서도 앞에 남습니다. 그래서 이 병합 정렬 구현은 안정 정렬입니다.

병합 정렬은 매 단계마다 임시 배열 `buffer`를 사용하므로 제자리 정렬은 아닙니다. 대신 `O(n log n)` 시간과 안정성을 예측하기 쉽습니다.

## pivot 값으로 경계를 가르는 Quick Sort

[QuickSort.sort](../examples/programmers/jvm/src/main/java/sort/QuickSort.java)는 가운데 위치의 값을 `pivot` 값으로 잡고, 왼쪽 포인터 `i`와 오른쪽 포인터 `j`를 안쪽으로 움직입니다. 이 구현에서 pivot은 "최종 위치가 바로 확정되는 원소"라기보다, 작은 값 쪽과 큰 값 쪽을 나누는 기준값입니다.

`[5, 1, 4, 2, 8]`에서는 가운데 값 `4`가 pivot입니다.

```text
start: [5, 1, 4, 2, 8]
              ^
              pivot = 4

i=0, j=4
    values[i] = 5 는 pivot보다 작지 않으므로 i는 멈춘다.
    values[j] = 8 은 pivot보다 크므로 j를 왼쪽으로 옮긴다.
    j=3, values[j] = 2 는 pivot보다 크지 않으므로 j는 멈춘다.

    i <= j 이므로 values[0]과 values[3]을 swap한다.
    [2, 1, 4, 5, 8]
     i        j

swap 뒤에는 i++, j--를 한다.
    i=1, j=2

values[1] = 1 < 4 이므로 i를 오른쪽으로 옮긴다.
    i=2, j=2

values[2] = 4 는 pivot보다 작지도 크지도 않다.
    i <= j 이므로 자기 자신과 swap하고 i++, j--로 경계를 넘긴다.
    i=3, j=1
```

여기서 반복이 끝나면 `left..j`와 `i..right`만 다시 정렬하면 됩니다.

```text
partition 뒤: [2, 1] [4] [5, 8]
재귀 대상:    left..j     i..right
              0..1        3..4
```

퀵 정렬을 "pivot이 자기 최종 위치에 박힌다"로만 외우면 이 구현을 잘못 읽을 수 있습니다. Lomuto partition처럼 pivot 원소 하나를 끝에 두고 최종 위치로 보내는 구현도 있지만, 이 패키지의 구현은 양쪽 포인터가 pivot 값 기준으로 엇갈릴 때까지 swap하며 경계를 만듭니다.

평균적으로는 `O(n log n)`이지만, pivot이 계속 나쁘게 갈리면 최악 `O(n^2)`까지 나빠질 수 있습니다. 코딩 테스트에서는 직접 퀵 정렬을 구현하는 일보다, "pivot 기준으로 조건을 만족하는 왼쪽/오른쪽 구간을 만든다"는 partition 감각이 더 자주 쓰입니다.

## heap의 root를 오른쪽으로 보내는 Heap Sort

[HeapSort.sort](../examples/programmers/jvm/src/main/java/sort/HeapSort.java)는 먼저 배열을 max heap으로 만듭니다. max heap은 부모 값이 자식 값보다 크거나 같은 완전 이진 트리 모양의 배열입니다. 배열에서 `root`의 왼쪽 자식은 `root * 2 + 1`, 오른쪽 자식은 `root * 2 + 2` 위치에 있습니다.

`[5, 1, 4, 2, 8]`을 max heap으로 만들면 아래처럼 됩니다.

```text
초기 배열:       [5, 1, 4, 2, 8]
buildMaxHeap 후: [8, 5, 4, 2, 1]

배열 인덱스를 트리처럼 보면:

        8
      /   \
     5     4
    / \
   2   1
```

root인 `values[0]`은 heap 구간에서 가장 큰 값입니다. 이 값을 배열 끝으로 보내면 오른쪽 끝이 확정됩니다.

```text
heap: [8, 5, 4, 2, 1]

end=4:
    root 8과 values[4]를 swap한다.
    [1, 5, 4, 2 | 8]

    heap 구간은 [1, 5, 4, 2]이고, 8은 확정 구간이다.
    root 1이 heap 조건을 깨므로 siftDown으로 아래로 내려보낸다.

    1과 더 큰 자식 5를 swap
    [5, 1, 4, 2 | 8]

    1과 더 큰 자식 2를 swap
    [5, 2, 4, 1 | 8]
```

다음 반복도 같은 구조입니다.

```text
end=3:
    [5, 2, 4, 1 | 8]
    root 5를 end로 보낸다.
    [1, 2, 4 | 5, 8]
    siftDown 후
    [4, 2, 1 | 5, 8]

end=2:
    [1, 2 | 4, 5, 8]
    siftDown 후
    [2, 1 | 4, 5, 8]

end=1:
    [1 | 2, 4, 5, 8]
```

힙 정렬의 불변식은 오른쪽 `end + 1..last` 구간이 이미 정렬된 suffix라는 점입니다. 버블 정렬도 오른쪽을 확정하지만, 버블 정렬은 이웃 swap으로 큰 값을 천천히 보내고, 힙 정렬은 heap root에서 최댓값을 바로 꺼냅니다.

## 값의 개수를 세어 다시 쓰는 Counting Sort

[CountingSort.sort](../examples/programmers/jvm/src/main/java/sort/CountingSort.java)는 비교 정렬이 아닙니다. 먼저 최솟값과 최댓값을 찾고, `value - min`을 count 배열의 offset으로 씁니다.

작은 예제로 보면 이 방식이 왜 가능한지 바로 보입니다.

```text
values: [3, 1, 3, 2]
min=1, max=3

offset = value - min

value 1 -> offset 0
value 2 -> offset 1
value 3 -> offset 2

counts를 채운 뒤:
offset:  0  1  2
value:   1  2  3
count:   1  1  2
```

그 다음에는 작은 offset부터 count만큼 값을 다시 씁니다.

```text
write=0, value=1, count=1 -> values[0] = 1
    [1, _, _, _]

write=1, value=2, count=1 -> values[1] = 2
    [1, 2, _, _]

write=2, value=3, count=2 -> values[2] = 3, values[3] = 3
    [1, 2, 3, 3]
```

이 정렬은 값 범위가 작을 때만 좋습니다. 입력 개수는 10개인데 최솟값과 최댓값 차이가 수십억이면 count 배열이 정렬 대상보다 훨씬 커집니다. 그래서 이 예제 구현은 `MAX_COUNTING_RANGE = 1_000_000`을 넘는 범위를 명시적으로 거절합니다.

```java
long range = (long) max - min + 1;
if (range > MAX_COUNTING_RANGE) {
    throw new IllegalArgumentException("value range is too wide for this counting sort example");
}
```

이 구현은 `int` 값만 다시 쓰므로 안정성을 관찰할 대상이 없습니다. 만약 원소가 `(score, name)` 같은 record이고 `score`로 계수 정렬을 안정적으로 하고 싶다면, count를 누적 위치로 바꾼 뒤 원본을 뒤에서 앞으로 읽어 output 배열에 배치하는 방식이 필요합니다. 그 버전은 "몇 번 나왔는가"에서 한 단계 더 나아가 "각 값이 결과 배열의 어느 구간을 차지하는가"를 계산합니다.

## 코딩 테스트에서 정렬을 볼 때 던질 질문

정렬 문제를 만나면 알고리즘 이름보다 아래 질문을 먼저 던지는 편이 실전적입니다.

1. 정렬하면 비교해야 할 후보가 이웃으로 붙는가

    전화번호 목록처럼 접두어 관계를 찾는 문제는 문자열을 정렬한 뒤 이웃한 값만 보면 됩니다. 전체 쌍을 비교하지 않아도 되는 이유는 같은 prefix를 공유하는 값들이 정렬 후 가까이 모이기 때문입니다.

2. 어떤 기준으로 정렬해야 greedy가 안전해지는가

    단속카메라 유형은 진출 지점 기준으로 정렬해야 카메라 위치를 확정할 수 있습니다. 섬 연결하기 유형은 비용 기준으로 간선을 정렬한 뒤, 사이클을 만들지 않는 간선을 고릅니다. 여기서 핵심은 정렬 자체보다 "정렬한 순서대로 선택해도 나중 선택이 앞선 선택을 깨지 않는가"입니다.

3. 전체 정렬이 필요한가, 상위 K개만 필요한가

    상위 몇 개만 계속 필요하면 전체 배열을 정렬하지 않고 heap이나 priority queue가 더 자연스러울 수 있습니다. 전체 순서가 필요한 문제인지, 최솟값/최댓값을 반복해서 꺼내면 충분한 문제인지 먼저 구분합니다.

4. 값 범위가 작아서 count 배열로 충분한가

    점수, 알파벳, 작은 자연수처럼 값의 범위가 좁으면 비교 정렬보다 개수 세기가 더 직접적일 수 있습니다. 반대로 값 범위가 입력 개수보다 훨씬 크면 count 배열은 좋은 선택이 아닙니다.

5. 같은 key의 원래 순서가 보존되어야 하는가

    같은 key를 가진 원소의 원래 순서가 의미 있다면 안정 정렬이 필요합니다. 이 패키지에서는 bubble sort, insertion sort, merge sort가 안정 정렬입니다. selection sort, quick sort, heap sort는 이 구현 기준으로 안정 정렬이 아닙니다.

## 이번 코드에 없는 정렬은 왜 뺐나

이 패키지는 정렬 알고리즘 전체 사전이 아니라, 코딩 테스트 감각을 회복하기 위한 첫 학습 묶음입니다. 그래서 먼저 비교 정렬의 대표 흐름과 값 범위가 좁을 때 쓰는 계수 정렬만 코드로 넣었습니다.

- Radix sort는 이미 [radix.md](radix.md)에 개념 메모가 있습니다. 자릿수별로 안정 정렬을 반복해야 하므로, counting sort의 count와 write 흐름을 먼저 익힌 뒤 구현하는 편이 자연스럽습니다.
- Bucket sort는 입력이 구간에 고르게 퍼진다는 전제가 중요합니다. 이 전제가 깨지면 bucket 안에서 다시 정렬해야 해서 기대한 이점이 줄어듭니다.
- Java 내장 정렬은 실제 코딩 테스트에서 가장 자주 씁니다. 다만 이번 문서의 목적은 `Arrays.sort` 사용법이 아니라, 정렬 코드 내부의 경계와 상태 변화를 손으로 복원하는 것입니다.

## 직접 실행하는 방법

JVM 예제 프로젝트에서 컴파일하고 runner를 실행합니다.

```bash
cd /Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm
JAVA_HOME=/Users/rody/Library/Java/JavaVirtualMachines/jbr-21.0.10/Contents/Home ./gradlew -q compileJava
java -cp build/classes/java/main sort.SortLearningRunner
```

성공하면 아래 문장이 출력됩니다.

```text
All sort learning tests passed
```

실패하면 어떤 정렬이 어떤 입력에서 `Arrays.sort` 결과와 달랐는지 `AssertionError`로 나옵니다. runner는 빈 배열, 원소 하나, 이미 정렬된 배열, 역순 배열, 중복 값, 음수 포함 배열, 더 긴 섞인 배열을 같은 방식으로 검증합니다. 또한 counting sort가 너무 넓은 값 범위를 조용히 처리하지 않고 예외로 거절하는지도 확인합니다.

## 손으로 다시 그려 볼 연습

아래 입력을 종이에 적고, 알고리즘마다 첫 pass 또는 첫 partition만 직접 그려 보세요.

```text
[4, 1, 3, 1]
```

1. Bubble sort

    첫 pass가 끝났을 때 오른쪽 끝에는 무엇이 확정되는지 표시합니다.

2. Selection sort

    `fixed=0`일 때 어느 값이 선택되고, swap 뒤 왼쪽 확정 구간이 어디까지인지 표시합니다.

3. Insertion sort

    `next=1`, `next=2`에서 어떤 값이 오른쪽으로 밀리고, `valueToInsert`가 어디에 다시 쓰이는지 표시합니다.

4. Quick sort

    이 패키지의 구현처럼 가운데 값을 pivot으로 잡고, `i`와 `j`가 어디에서 멈추는지 표시합니다. pivot 원소 하나가 바로 최종 위치로 확정된다고 가정하지 않는 것이 중요합니다.

5. Counting sort

    `min`, `max`, `counts` 배열을 먼저 만들고, `write`가 왼쪽부터 어떻게 움직이는지 표시합니다.

직접 그린 결과를 코드와 비교할 때는 "최종 배열이 맞는가"만 보지 말고, 중간 상태의 약속이 코드 주석과 맞는지 확인합니다. 이 확인을 할 수 있으면 정렬을 외운 것이 아니라, 정렬 코드가 배열을 바꾸는 방식을 복원할 수 있게 된 것입니다.
