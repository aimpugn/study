# uint64

- [uint64](#uint64)
    - [uint64 to bytes](#uint64-to-bytes)

## uint64 to bytes

- `uint64` 값을 `[]byte` 배열로 변환하는 원리는 메모리에 저장된 `uint64` 값을 연속된 8개의 바이트로 해석하는 것
- 이는 바이트 오더(Byte Order) 또는 엔디언(Endian)이라는 개념과 관련이 있다
- 컴퓨터 메모리는 바이트 단위로 데이터를 저장하고, 각 바이트는 고유한 주소를 갖는다
- `uint64` 타입의 값은 메모리 상에서 **연속된 8개의 바이트로 표현**된다. 각 바이트당 8비트, 총 64비트를 구성
- 여기서 중요한 것은 바이트를 배열로 해석할 때 어떤 바이트를 먼저 해석할지 결정하는 엔디언:
    - **빅 엔디언(Big Endian)**: 가장 큰 바이트, 즉 가장 상위 바이트가 메모리의 가장 낮은 주소에 저장됩니다.
    - **리틀 엔디언(Little Endian)**: 가장 작은 바이트, 즉 가장 하위 바이트가 메모리의 가장 낮은 주소에 저장됩니다.
- Go 언어의 `encoding/binary` 패키지는 이러한 엔디언을 다룰 수 있는 함수들을 제공
    - 예를 들어, `binary.BigEndian.PutUint64` 함수는 `uint64` 값을 빅 엔디언 방식으로 바이트 슬라이스에 저장
        1. `uint64` 값의 가장 상위 바이트(7번째 바이트)를 바이트 슬라이스의 첫 번째 요소에 저장
        2. 다음 바이트(6번째 바이트)를 바이트 슬라이스의 두 번째 요소에 저장
        3. 이 과정을 계속하여 마지막 바이트(0번째 바이트)를 바이트 슬라이스의 마지막 요소에 저장
- 이렇게 하면 `uint64` 값이 메모리 상에 8개의 바이트로 나타나며, 각 바이트는 원래 `uint64` 값의 일부를 나타내는데, 이를 배열로 변환하면 바이트 단위로 해당 값을 읽고 쓸 수 있다