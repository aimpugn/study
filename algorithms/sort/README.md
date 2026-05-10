# 정렬 알고리즘은 무엇을 하나씩 확정해 가는가

정렬 알고리즘을 공부할 때 처음 붙잡을 질문은 "어떤 알고리즘 이름이 몇 개인가"가 아니라, 배열 안에서 어느 구간이 이미 믿을 수 있는 상태가 되는가입니다. 같은 오름차순 결과를 만들더라도 어떤 알고리즘은 오른쪽 끝부터 큰 값을 확정하고, 어떤 알고리즘은 왼쪽부터 작은 값을 확정하며, 어떤 알고리즘은 둘로 나눈 구간을 다시 합칩니다.

이 문서는 코딩 테스트에서 자주 만나는 정렬 감각을 Java 코드로 다시 따라가기 위한 학습 허브입니다. 구현은 [sort 패키지](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort)에 있고, 실행 검증은 [SortLearningRunner.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/SortLearningRunner.java)가 맡습니다.

## 목차

- [먼저 붙잡을 작은 배열](#먼저-붙잡을-작은-배열)
- [이번 패키지에 넣은 알고리즘](#이번-패키지에-넣은-알고리즘)
- [비교로 정렬하는 기본 알고리즘](#비교로-정렬하는-기본-알고리즘)
- [나누고 합치거나 경계를 가르는 알고리즘](#나누고-합치거나-경계를-가르는-알고리즘)
- [힙과 개수 세기](#힙과-개수-세기)
- [코딩 테스트에서 실제로 떠올릴 질문](#코딩-테스트에서-실제로-떠올릴-질문)
- [직접 실행하는 방법](#직접-실행하는-방법)
- [다음 학습으로 미룬 정렬](#다음-학습으로-미룬-정렬)

## 먼저 붙잡을 작은 배열

설명은 아래 배열 하나를 계속 기준으로 삼습니다.

```text
[5, 1, 4, 2, 8]
```

정렬이 끝난 모습은 `[1, 2, 4, 5, 8]`입니다. 중요한 부분은 결과보다 중간 상태입니다. 예를 들어 버블 정렬은 첫 pass가 끝나면 `8`이 오른쪽 끝에 확정됩니다. 선택 정렬은 첫 pass가 끝나면 `1`이 왼쪽 끝에 확정됩니다. 삽입 정렬은 왼쪽 구간을 이미 정렬된 작은 배열로 보고, 다음 값을 그 안의 제자리로 밀어 넣습니다.

이 차이를 알아야 코딩 테스트에서 정렬을 단순 호출로만 보지 않습니다. 정렬은 종종 "후보를 이웃으로 붙이는 도구"이고, 때로는 "앞에서부터 안전하게 고르기 위한 순서 만들기"입니다.

## 이번 패키지에 넣은 알고리즘

| 알고리즘 | 구현 | 먼저 볼 질문 | 코테에서의 위치 |
|---|---|---|---|
| Bubble Sort | [BubbleSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/BubbleSort.java) | 오른쪽 끝에 무엇이 확정되는가 | 원리 학습용 |
| Selection Sort | [SelectionSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/SelectionSort.java) | 왼쪽부터 어떤 값을 확정하는가 | 원리 학습용 |
| Insertion Sort | [InsertionSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/InsertionSort.java) | 왼쪽 정렬 구간에 다음 값을 어디에 끼우는가 | 작은 입력, 거의 정렬된 입력 감각 |
| Merge Sort | [MergeSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/MergeSort.java) | 이미 정렬된 두 구간을 어떻게 합치는가 | 안정 정렬, 분할 정복 감각 |
| Quick Sort | [QuickSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/QuickSort.java) | pivot 기준으로 어떤 경계를 만드는가 | partition 사고 학습 |
| Heap Sort | [HeapSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/HeapSort.java) | heap의 root를 어디에 확정하는가 | 우선순위 큐 감각으로 연결 |
| Counting Sort | [CountingSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/CountingSort.java) | 값을 비교하지 않고 개수만 세도 되는가 | 값 범위가 작을 때 |

이번 범위는 일부러 여기서 끊었습니다. Radix sort, bucket sort, Java 내장 정렬은 중요하지만, 처음부터 전부 넣으면 각 알고리즘의 불변식이 흐려집니다. 이번 패키지는 먼저 "구간이 어떻게 확정되는가"를 손에 붙이는 데 초점을 둡니다.

## 비교로 정렬하는 기본 알고리즘

1. Bubble Sort

    버블 정렬은 이웃한 두 값을 계속 비교합니다. `[5, 1, 4, 2, 8]`에서 첫 비교는 `5`와 `1`입니다. 큰 값 `5`가 오른쪽으로 움직이고, 그 다음에는 `5`와 `4`, 다시 `5`와 `2`를 비교합니다. 이 pass가 끝나면 가장 큰 값이 오른쪽 끝에 도착합니다.

    이 알고리즘의 불변식은 `end` 오른쪽이 이미 정렬된 suffix라는 점입니다. [BubbleSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/BubbleSort.java)는 한 pass에서 swap이 없으면 바로 멈춥니다. 이때는 모든 이웃이 이미 오름차순이므로 남은 반복이 배열을 바꾸지 못합니다.

2. Selection Sort

    선택 정렬은 현재 위치에 들어갈 최솟값을 뒤쪽에서 찾습니다. `[5, 1, 4, 2, 8]`의 첫 pass에서는 전체 배열에서 `1`을 찾아 0번 자리에 둡니다. 그 다음 pass에서는 1번 자리부터 끝까지 보며 `2`를 찾습니다.

    이 알고리즘의 불변식은 `0..fixed-1` 구간이 이미 최솟값부터 차례대로 확정되었다는 점입니다. swap 기반 구현은 단순하지만 안정 정렬은 아닙니다. 같은 값이 있을 때 원래 앞에 있던 값이 뒤로 밀릴 수 있기 때문입니다.

3. Insertion Sort

    삽입 정렬은 왼쪽을 이미 정렬된 작은 배열로 봅니다. `[5, 1, 4, 2, 8]`에서 `1`을 볼 때 왼쪽 `[5]` 안에 `1`이 들어갈 자리를 찾고, `5`를 오른쪽으로 민 뒤 `1`을 넣습니다. 그 다음 `4`는 `[1, 5]` 안에서 `5` 앞에 들어갑니다.

    이 알고리즘의 불변식은 `0..next-1`이 이미 정렬되어 있다는 점입니다. 그래서 거의 정렬된 배열에서는 많이 움직이지 않습니다. 코딩 테스트에서 직접 삽입 정렬을 구현할 일은 적지만, "왼쪽 정렬 구간에 하나씩 끼워 넣는다"는 감각은 작은 상태를 확장하는 문제를 볼 때 도움이 됩니다.

## 나누고 합치거나 경계를 가르는 알고리즘

1. Merge Sort

    병합 정렬은 먼저 배열을 반으로 쪼갭니다. `[5, 1, 4, 2, 8]`은 `[5, 1]`, `[4, 2, 8]`처럼 더 작은 구간으로 내려갑니다. 길이가 1인 구간은 이미 정렬되어 있으므로, 그때부터 두 정렬된 구간을 합칩니다.

    핵심은 merge 단계입니다. 왼쪽 구간과 오른쪽 구간이 각각 정렬되어 있다면, 두 구간의 맨 앞 값만 비교해도 다음에 쓸 값을 고를 수 있습니다. [MergeSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/MergeSort.java)는 같은 값이 나오면 왼쪽 값을 먼저 쓰므로 안정성을 유지합니다.

2. Quick Sort

    퀵 정렬은 pivot을 하나 고르고, pivot보다 작은 값은 왼쪽으로, 큰 값은 오른쪽으로 보내는 경계를 만듭니다. 이번 구현은 가운데 값을 pivot으로 삼고 양끝 포인터를 안쪽으로 움직입니다.

    퀵 정렬을 볼 때 조심할 점은 "항상 빠르다"가 아니라는 점입니다. 평균적으로는 `O(n log n)`이지만, pivot이 계속 나쁘게 잡히면 `O(n^2)`까지 나빠질 수 있습니다. 그래서 코딩 테스트에서 직접 퀵 정렬을 제출하려는 목적보다, partition이라는 생각을 이해하는 쪽이 더 중요합니다.

## 힙과 개수 세기

1. Heap Sort

    힙 정렬은 배열을 먼저 max heap으로 만듭니다. max heap은 부모가 자식보다 크거나 같은 구조입니다. 이 구조가 있으면 root인 `values[0]`이 현재 heap 구간의 최댓값입니다.

    [HeapSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/HeapSort.java)는 root를 배열 끝으로 보내고, heap 크기를 하나 줄인 뒤 `siftDown`으로 heap 조건을 복구합니다. 오른쪽 끝에는 큰 값부터 차례대로 확정됩니다. 이 감각은 `PriorityQueue`를 쓰는 문제로 바로 이어집니다.

2. Counting Sort

    계수 정렬은 값을 서로 비교하지 않습니다. 값이 몇 번 나왔는지 세고, 작은 값부터 count만큼 다시 씁니다. 예를 들어 `[3, 1, 3, 2]`라면 `1`은 1번, `2`는 1번, `3`은 2번 나왔으므로 `[1, 2, 3, 3]`을 만들 수 있습니다.

    이 방식은 값 범위가 작을 때만 좋습니다. 입력이 10개인데 최솟값과 최댓값 차이가 수십억이면 count 배열이 너무 커집니다. [CountingSort.java](/Users/rody/VscodeProjects/study/algorithms/examples/programmers/jvm/src/main/java/sort/CountingSort.java)는 그런 입력을 조용히 처리하지 않고 예외로 거절합니다.

## 코딩 테스트에서 실제로 떠올릴 질문

정렬 문제를 만나면 알고리즘 이름보다 아래 질문을 먼저 던지는 편이 실전적입니다.

1. 정렬하면 후보가 이웃으로 붙는가

    전화번호 목록처럼 접두어 관계를 찾는 문제는 정렬 후 이웃만 보면 됩니다. 멀리 떨어진 원소끼리 비교하지 않아도 됩니다.

2. 어떤 기준으로 정렬해야 greedy가 안전해지는가

    단속카메라는 진출 지점 기준으로 정렬해야 카메라 위치를 확정할 수 있습니다. 섬 연결하기는 비용 기준으로 간선을 정렬한 뒤, 사이클을 만들지 않는 간선을 고릅니다.

3. 전체 정렬이 필요한가, 상위 K개만 필요한가

    상위 몇 개만 계속 필요하면 전체 배열을 정렬하지 않고 heap이 더 자연스러울 수 있습니다.

4. 값 범위가 작아서 count 배열로 충분한가

    점수, 알파벳, 작은 자연수처럼 값의 범위가 좁으면 비교 정렬보다 개수 세기가 더 직접적일 때가 있습니다.

5. 안정 정렬이 필요한가

    같은 key를 가진 원소의 원래 순서가 보존되어야 한다면 안정 정렬 여부를 봐야 합니다. 이 패키지에서는 bubble sort, insertion sort, merge sort가 안정 정렬입니다.

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

실패하면 어떤 정렬이 어떤 입력에서 `Arrays.sort` 결과와 달랐는지 `AssertionError`로 나옵니다. 이 runner는 빈 배열, 원소 하나, 이미 정렬된 배열, 역순 배열, 중복 값, 음수 포함 배열을 모두 같은 방식으로 검증합니다.

## 다음 학습으로 미룬 정렬

- Radix sort는 이미 [radix.md](/Users/rody/VscodeProjects/study/algorithms/sort/radix.md)에 개념 메모가 있습니다. 이 정렬은 자릿수별로 안정 정렬을 반복해야 하므로, counting sort를 먼저 이해한 뒤 보는 편이 좋습니다.
- Bucket sort는 입력이 구간에 고르게 퍼진다는 전제가 중요합니다. 전제가 깨지면 bucket 안에서 다시 정렬 비용이 커집니다.
- Java 내장 정렬은 실제 코딩 테스트에서 가장 자주 쓰지만, 이번 패키지의 목적은 라이브러리 사용법이 아니라 정렬 내부의 상태 변화를 이해하는 것입니다.

마지막으로 직접 확인해 볼 질문을 남깁니다. `[4, 1, 3, 1]`을 손으로 따라가며 bubble sort와 selection sort의 첫 pass를 비교해 보세요. 두 알고리즘 모두 배열을 조금 더 정렬된 상태로 만들지만, 하나는 오른쪽 끝을 확정하고 다른 하나는 왼쪽 끝을 확정합니다. 그 차이를 말로 설명할 수 있으면 첫 단계는 통과입니다.
