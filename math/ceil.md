# ceil

- [ceil](#ceil)
    - [올림을 수동으로](#올림을-수동으로)
        - [`(length + chunkSize) / chunkSize` 는 안된는 이유](#length--chunksize--chunksize-는-안된는-이유)
        - [예제 1: length = 5, chunkSize = 2](#예제-1-length--5-chunksize--2)
            - [예제 2: length = 7, chunkSize = 3](#예제-2-length--7-chunksize--3)
        - [예제 3: length = 8, chunkSize = 4](#예제-3-length--8-chunksize--4)
        - [예제 4: length = 100, chunkSize = 64](#예제-4-length--100-chunksize--64)
        - [예제 5: length = 64, chunkSize = 64](#예제-5-length--64-chunksize--64)

## 올림을 수동으로

```go
// (length + chunkSize - 1) / chunkSize
math.Ceil(length / chunkSize)
```

먼저, 정수 나눗셈의 기본적인 성질을 이해해야 합니다.
정수 나눗셈은 항상 나눗셈의 결과를 내림합니다.
예를 들어, `5 / 2`는 2.5지만, 정수 나눗셈에서는 결과가 2가 됩니다.

`math.Ceil` 함수는 소수점을 올림하여 가장 작은 정수를 반환합니다.
예를 들어, `length / chunkSize`가 1.5이면 `math.Ceil`은 2를 반환합니다.

하지만 Go에서는 정수 나눗셈을 사용하면 소수점 이하가 잘리므로, 이를 올림 효과로 바꾸기 위해 `(length + chunkSize - 1) / chunkSize`를 사용합니다.

이제 일반적인 경우를 생각해보겠습니다.
`length`를 `chunkSize`로 나눌 때, 올림 효과를 얻기 위해 `chunkSize - 1`을 더합니다.

*`length`가 `chunkSize`의 정수 배수가 아닌 경우, `length`를 `chunkSize`로 나누면 항상 소수점 이하가 생깁니다*.
이때 `chunkSize - 1`을 더하면 나머지 부분이 `chunkSize`를 넘어가게 되어 정수 나눗셈에서 올림 효과를 가져옵니다.

정수 나눗셈에서 소수점을 올림하는 방법을 수식으로 나타내면 다음과 같습니다:

1. `(length + chunkSize - 1) / chunkSize`는 `length`에 `chunkSize - 1`을 더한 값을 `chunkSize`로 나눕니다.
2. 이렇게 하면 `length`가 `chunkSize`의 정수 배수가 아닐 때 올림 효과를 줍니다.

- (length + chunkSize - 1) / chunkSize

    1. `length`를 `chunkSize`로 나눌 때 나머지를 고려합니다.
    2. `length = n * chunkSize + r` (여기서 `n`은 몫, `r`은 나머지, 0 <= r < chunkSize)
    3. `(length + chunkSize - 1) / chunkSize`는 `(n * chunkSize + r + chunkSize - 1) / chunkSize`로 표현됩니다.
    4. 이 수식을 정리하면 `n + (r + chunkSize - 1) / chunkSize`가 됩니다.
    5. `(r + chunkSize - 1) / chunkSize`는 항상 `1`이 되므로 결과는 `n + 1`이 됩니다.

### `(length + chunkSize) / chunkSize` 는 안된는 이유

1. **정수 나눗셈 기본 원리**:
   - `length`를 `chunkSize`로 나누면, 나머지가 생길 수 있습니다.
   - 예를 들어, `length = 7`, `chunkSize = 3`이면 `7 / 3`은 2.333...이고, 정수 나눗셈으로는 2가 됩니다.

2. **올림을 구현**:
   - 우리가 원하는 것은 `7 / 3`의 올림 값인 3입니다.
   - `(length + chunkSize - 1)`을 사용하면, `7 + 3 - 1 = 9`가 됩니다.
   - `9 / 3`는 3입니다. 즉, 올림이 된 결과를 얻을 수 있습니다.

이제, `(length + chunkSize) / chunkSize`를 생각해봅시다.
이 경우, 추가된 `chunkSize`로 인해 예상치 못한 결과를 얻을 수 있습니다. 예제를 통해 살펴보겠습니다:

- **예제**: `length = 7`, `chunkSize = 3`인 경우

    `(length + chunkSize) / chunkSize`는 `(7 + 3) / 3 = 10 / 3`입니다.
    이는 3.333...이고, 정수 나눗셈에서는 3이 됩니다.
    이 경우는 올바른 결과를 얻었지만, 다음과 같은 경우를 생각해봅시다.

- **예제 2**: `length = 6`, `chunkSize = 3`인 경우

    `(length + chunkSize) / chunkSize`는 `(6 + 3) / 3 = 9 / 3`입니다.
    이는 3이 되고, 이는 올림 효과가 필요 없는 경우에도 잘못된 결과를 줍니다.
    하지만, 우리가 원하는 것은 `2`입니다. 즉, 올림이 필요 없는 경우에도 결과가 달라질 수 있습니다.

### 예제 1: length = 5, chunkSize = 2

1. `length / chunkSize`를 계산해봅시다:
   - 5 / 2 = 2.5
   - `math.Ceil(2.5)` = 3

2. `(length + chunkSize - 1) / chunkSize`를 계산해봅시다:
   - length + chunkSize - 1 = 5 + 2 - 1 = 6
   - 6 / 2 = 3 (정수 나눗셈)

#### 예제 2: length = 7, chunkSize = 3

1. `length / chunkSize`를 계산해봅시다:
   - 7 / 3 = 2.3333...
   - `math.Ceil(2.3333...)` = 3

2. `(length + chunkSize - 1) / chunkSize`를 계산해봅시다:
   - length + chunkSize - 1 = 7 + 3 - 1 = 9
   - 9 / 3 = 3 (정수 나눗셈)

### 예제 3: length = 8, chunkSize = 4

1. `length / chunkSize`를 계산해봅시다:
   - 8 / 4 = 2
   - `math.Ceil(2)` = 2

2. `(length + chunkSize - 1) / chunkSize`를 계산해봅시다:
   - length + chunkSize - 1 = 8 + 4 - 1 = 11
   - 11 / 4 = 2 (정수 나눗셈, 나머지 3)

### 예제 4: length = 100, chunkSize = 64

1. `length / chunkSize`를 구해봅시다:
   - `100 / 64 = 1.5625` (실제 나눗셈 값)
   - 정수 나눗셈에서는 `1`이 됩니다.
   - `math.Ceil(1.5625) = 2`입니다.

2. `(length + chunkSize - 1) / chunkSize`를 계산해봅시다:
   - `100 + 64 - 1 = 163`
   - `163 / 64 = 2` (정수 나눗셈)

### 예제 5: length = 64, chunkSize = 64

1. `length / chunkSize`를 구해봅시다:
   - `64 / 64 = 1`
   - 정수 나눗셈에서도 `1`입니다.
   - `math.Ceil(1) = 1`입니다.

2. `(length + chunkSize - 1) / chunkSize`를 계산해봅시다:
   - `64 + 64 - 1 = 127`
   - `127 / 64 = 1` (정수 나눗셈)
