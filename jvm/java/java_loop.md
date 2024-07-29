# java loop

- [java loop](#java-loop)
    - [기본 for 루프](#기본-for-루프)
    - [역순 for 루프](#역순-for-루프)
    - [증분 값 변경](#증분-값-변경)
    - [다중 변수 사용](#다중-변수-사용)
    - [슬라이딩 윈도우](#슬라이딩-윈도우)
    - [연속 요소 비교 (Pairwise Comparison)](#연속-요소-비교-pairwise-comparison)
    - [중첩 루프](#중첩-루프)

## 기본 for 루프

가장 일반적인 for 루프 패턴입니다.

```java
int[] numbers = {1, 2, 3, 4, 5};
for (int i = 0; i < numbers.length; i++) {
    System.out.println(numbers[i]);
}
```

## 역순 for 루프

배열을 역순으로 순회합니다.

```java
int[] numbers = {1, 2, 3, 4, 5};
for (int i = numbers.length - 1; i >= 0; i--) {
    System.out.println(numbers[i]);
}
```

## 증분 값 변경

2씩 증가하는 등 증분 값을 변경할 수 있습니다.

```java
for (int i = 0; i < 10; i += 2) {
    System.out.println(i); // 0, 2, 4, 6, 8 출력
}
```

## 다중 변수 사용

여러 변수를 동시에 사용할 수 있습니다.

```java
int[] numbers = {1, 2, 3, 4, 5};
for (int i = 0, j = numbers.length - 1; i < j; i++, j--) {
    int temp = numbers[i];
    numbers[i] = numbers[j];
    numbers[j] = temp;
}
```

## 슬라이딩 윈도우

고정된 크기의 `윈도우`(아래 예제에서 부분배열)를 배열을 따라 이동시키며 연속된 부분 배열을 처리하는 기법입니다.

```plaintext
배열:       [A, B, C, D, E, F, G]
부분배열 1:  [A, B, C]
부분배열 2:     [B, C, D]
부분배열 3:        [C, D, E]
부분배열 4:           [D, E, F]
부분배열 5:              [E, F, G]
```

```java
int[] numbers = {1, 2, 3, 4, 5, 6, 7};
int windowSize = 3;
for (int i = 0; i <= numbers.length - windowSize; i++) {
    int sum = 0;
    for (int j = i; j < i + windowSize; j++) {
        sum += numbers[j];
    }
    System.out.println("Window sum: " + sum);
}
```

설명:
- 외부 루프는 `n - windowSize + 1`번 실행됩니다 (여기서는 5번).
- 각 반복에서 `windowSize`개의 연속된 요소를 처리합니다.
- 윈도우는 한 번에 한 요소씩 오른쪽으로 이동합니다.
- 이 방법으로 모든 가능한 연속된 `windowSize` 크기의 부분 배열을 처리할 수 있습니다.

## 연속 요소 비교 (Pairwise Comparison)

이 패턴은 배열의 연속된 두 요소를 비교할 때 사용됩니다.
비교 횟수는 요소의 수보다 하나 적지만, 모든 요소를 커버합니다.

```plaintext
배열: [A, B, C, D, E]
비교:  ^--^
         ^--^
            ^--^
               ^--^
```

- 비교 횟수: n-1 (여기서는 4번)
- 커버되는 요소: n (여기서는 5개)

```java
int[] numbers = {1, 3, 2, 4, 5};
for (int i = 0; i < numbers.length - 1; i++) {
    if (numbers[i] > numbers[i + 1]) {
        System.out.println("Decrease at index " + i);
    }
}
```

설명:
- 루프는 `n-1`번 실행되지만, 각 반복에서 두 개의 요소를 비교합니다.
- 첫 번째 요소(`numbers[0]`)는 첫 비교에서, 마지막 요소(`numbers[4]`)는 마지막 비교에서 처리됩니다.
- 이 방식으로 `n-1`번의 비교로 `n`개의 모든 요소를 처리할 수 있습니다.

## 중첩 루프

2차원 배열이나 복잡한 구조를 순회할 때 사용합니다.

```java
int[][] matrix = {{1, 2, 3}, {4, 5, 6}, {7, 8, 9}};
for (int i = 0; i < matrix.length; i++) {
    for (int j = 0; j < matrix[i].length; j++) {
        System.out.print(matrix[i][j] + " ");
    }
    System.out.println();
}
```
