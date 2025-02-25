# characters

## 문자들은 8비트 덩어리 시퀀스로 표현된되고, 각 덩어리는 2개의 16진 문자로 표현 가능

1. **비트와 바이트**:
   - **비트(bit)**: 컴퓨터에서 가장 작은 데이터 단위로, 0 또는 1의 값을 가집니다.
   - **바이트(byte)**: 8비트로 구성된 데이터 단위입니다. 따라서 1바이트는 0에서 255까지의 값을 가질 수 있습니다.

2. **문자의 표현**:
   - 컴퓨터는 문자를 저장하고 표현하기 위해 각 문자를 특정한 바이트 시퀀스로 변환합니다.
   - 예를 들어, ASCII 인코딩에서 문자 'A'는 1바이트(8비트)로 표현되며, 이 1바이트의 이진수 값은 `01000001`입니다.

### 8비트 덩어리와 16진 문자

1. **8비트 덩어리**:
   - 각 문자는 8비트(1바이트)로 표현됩니다.
   - 예를 들어, 문자 'A'의 8비트 표현은 `01000001`입니다.

2. **16진 문자**:
   - 8비트(1바이트)를 16진수로 표현할 수 있습니다.
   - 8비트를 4비트씩 두 부분으로 나누어 각각을 16진수로 변환합니다.
   - 예를 들어, `01000001`을 두 부분으로 나누면 `0100`과 `0001`이 됩니다.
   - `0100`은 16진수로 `4`, `0001`은 16진수로 `1`입니다.
   - 따라서, `01000001`은 16진수로 `41`입니다.

### 예시

#### 문자 'A'

- **8비트 표현**: `01000001`
- **16진수 표현**:
    - `0100` -> `4`
    - `0001` -> `1`
    - 합치면 `41`

#### 문자 'B'

- **8비트 표현**: `01000010`
- **16진수 표현**:
    - `0100` -> `4`
    - `0010` -> `2`
    - 합치면 `42`

### 시퀀스로 표현하기

여러 문자를 나열할 때, 각 문자는 8비트 시퀀스로 표현됩니다. 이를 16진수로 변환하면 2개의 16진 문자가 됩니다.

예를 들어, 문자열 "AB"를 16진수로 표현하면:

- 문자 'A': 8비트 -> `01000001` -> 16진수 `41`
- 문자 'B': 8비트 -> `01000010` -> 16진수 `42`

따라서 문자열 "AB"는 16진수로 `41 42`가 됩니다.

### 정리

- **8비트 덩어리**: 문자는 8비트(1바이트)로 표현됩니다.
- **2개의 16진 문자**: 각 8비트는 2개의 16진 문자로 표현됩니다. 예를 들어, `01000001`은 `41`로, `01000010`은 `42`로 표현됩니다.

이러한 방식은 컴퓨터가 문자 데이터를 저장하고 전송할 때, 특히 16진수로 데이터를 표현할 때 매우 유용합니다.
16진수는 2진수보다 사람이 읽고 이해하기 쉬운 형식으로, 디버깅이나 데이터 분석 시 자주 사용됩니다.
