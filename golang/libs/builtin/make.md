# make

- [make](#make)
    - [make 함수](#make-함수)
        - [make Slice (슬라이스)](#make-slice-슬라이스)
        - [make Map](#make-map)
        - [make Channel (채널)](#make-channel-채널)
    - [make로 용량 미리 할당시 장점](#make로-용량-미리-할당시-장점)

## make 함수

> The make built-in function allocates and initializes an object of type `slice`, `map`, or `chan` (only).
> Like `new`, the first argument is a type, not a value. Unlike `new`, `make`'s return type is the same as the type of its argument, not a pointer to it.
> The specification of the result depends on the type:
>
> - Slice: The size specifies the length. The capacity of the slice is equal to its length. A second integer argument may be provided to specify a different capacity; it must be no smaller than the length. For example, make([] int, 0, 10) allocates an underlying array of size 10 and returns a slice of length 0 and capacity 10 that is backed by this underlying array.
> - Map: An empty map is allocated with enough space to hold the specified number of elements. The size may be omitted, in which case a small starting size is allocated.
> - Channel: The channel's buffer is initialized with the specified buffer capacity. If zero, or the size is omitted, the channel is unbuffered.

```go
func make(t Type, size ...IntegerType) Type
```

make 내장 함수는 slice, map, 또는 chan (채널) 타입의 객체만을 할당하고 초기화합니다.
`new` 함수와 마찬가지로 첫 번째 인자는 값이 아닌 타입입니다.
그러나 `new`와 달리 `make`의 반환 타입은 그 인자의 타입과 동일하며, 포인터가 아닙니다.
결과의 명세는 타입에 따라 다릅니다:

1. Slice (슬라이스):

    `size`는 길이를 지정합니다.
    슬라이스의 용량은 길이와 같습니다.
    두 번째 정수 인자를 제공하여 다른 용량을 지정할 수 있습니다; 이 용량은 길이보다 작아서는 안 됩니다.

    예를 들어, `make([]int, 0, 10)`은 크기가 10인 기본 배열을 할당하고, 이 기본 배열에 의해 지원되는 길이 0, 용량 10의 슬라이스를 반환합니다.

2. Map (맵):

    지정된 수의 요소를 저장할 수 있는 충분한 공간이 할당된 빈 맵이 생성됩니다.
    `size`를 생략할 수 있으며, 이 경우 작은 초기 크기가 할당됩니다.

3. Channel (채널):

    채널의 버퍼가 지정된 버퍼 용량으로 초기화됩니다.
    0이거나 `size`가 생략되면, 채널은 버퍼가 없는 상태가 됩니다.

### make Slice (슬라이스)

슬라이스는 동적 배열과 유사한 데이터 구조입니다. `make` 함수로 슬라이스를 생성할 때, 두 가지 또는 세 가지 인자를 사용할 수 있습니다.

- `make([]T, length)`: 길이가 length인 슬라이스를 생성합니다. 용량도 length와 같습니다.
- `make([]T, length, capacity)`: 길이가 length이고 용량이 capacity인 슬라이스를 생성합니다.

예제 설명:

```go
slice := make([]int, 0, 10)
```

이 코드는:
1. 크기가 10인 정수 배열을 메모리에 할당합니다 (backing array).
2. 이 배열을 참조하는 슬라이스를 생성합니다.
3. 슬라이스의 길이는 0입니다 (현재 사용 중인 요소 없음).
4. 슬라이스의 용량은 10입니다 (추가로 할당 없이 10개 요소까지 추가 가능).

이렇게 생성된 슬라이스는 초기에는 비어 있지만, `append` 함수를 사용하여 최대 10개의 요소를 추가할 때까지 새로운 메모리 할당 없이 효율적으로 데이터를 추가할 수 있습니다.

### make Map

맵은 키-값 쌍을 저장하는 해시 테이블 기반의 자료구조입니다.

- `make(map[K]V)`: 기본 크기의 맵을 생성합니다.
- `make(map[K]V, size)`: 최소 size 개의 요소를 저장할 수 있는 초기 공간이 할당된 맵을 생성합니다.

예제:

```go
m := make(map[string]int, 100)
```

이 코드는 string 타입의 키와 int 타입의 값을 가지며, 초기에 100개의 요소를 효율적으로 저장할 수 있는 맵을 생성합니다. 맵은 필요에 따라 자동으로 크기가 조정되지만, 예상되는 요소 수를 지정하면 초기 해시 충돌을 줄이고 성능을 향상시킬 수 있습니다.

### make Channel (채널)

채널은 고루틴 간 통신 및 동기화에 사용되는 데이터 구조입니다.

- `make(chan T)`: 버퍼가 없는 채널을 생성합니다.
- `make(chan T, capacity)`: 지정된 용량의 버퍼를 가진 채널을 생성합니다.

예제:

```go
ch := make(chan int, 5)
```

이 코드는 정수를 주고받을 수 있고, 최대 5개의 값을 버퍼에 저장할 수 있는 채널을 생성합니다.
버퍼가 가득 차기 전까지는 송신 작업이 블록되지 않으며, 버퍼가 비어있지 않은 한 수신 작업도 블록되지 않습니다.

make 함수로 용량을 미리 할당했을 때의 주요 장점들을 자세히 설명해드리겠습니다.

## make로 용량 미리 할당시 장점

1. 메모리 재할당 횟수 감소

    슬라이스는 내부적으로 배열을 사용합니다.
    만약 용량이 부족하다면 새로운 더 큰 배열을 할당하고 데이터를 복사합니다.

    따라서 예상 크기만큼 미리 할당하면 재할당 횟수가 줄어 성능이 향상됩니다.
    특히 대량의 데이터를 다룰 때 효과가 큽니다.

    예시:

    ```go
    // 비효율적인 방식
    s := make([]int, 0)
    for i := 0; i < 10000; i++ {
        s = append(s, i)  // 여러 번의 재할당 발생
    }

    // 효율적인 방식
    s := make([]int, 0, 10000)
    for i := 0; i < 10000; i++ {
        s = append(s, i)  // 재할당 없음
    }
    ```

2. 성능 향상

    메모리 할당은 비용이 큰 연산입니다.
    데이터 복사도 CPU 시간을 소모합니다.

    불필요한 메모리 할당과 복사 연산을 줄여 전체적인 실행 시간이 단축됩니다.
    특히 시간이 중요한 애플리케이션에서 유용합니다.

    ```go
    func BenchmarkSliceAllocation(b *testing.B) {
        for i := 0; i < b.N; i++ {
            s := make([]int, 0)
            for j := 0; j < 10000; j++ {
                s = append(s, j)
            }
        }
    }

    func BenchmarkSlicePreallocation(b *testing.B) {
        for i := 0; i < b.N; i++ {
            s := make([]int, 0, 10000)
            for j := 0; j < 10000; j++ {
                s = append(s, j)
            }
        }
    }
    ```

    미리 할당한 버전이 일반적으로 더 빠른 성능을 보입니다.

3. 예측 가능한 메모리 사용

    Go의 슬라이스 확장 알고리즘은 현재 용량의 약 2배로 증가시킵니다.
    이는 예측하기 어려운 메모리 사용 패턴을 만들 수 있습니다.

    미리 할당한다면 메모리 사용량을 정확히 예측할 수 있습니다.
    메모리 제한이 있는 환경에서 특히 유용합니다.

    ```go
    // 예측하기 어려운 메모리 사용
    s := make([]int, 0)
    for i := 0; i < 1000000; i++ {
        s = append(s, i)  // 용량이 여러 번 증가함
    }

    // 예측 가능한 메모리 사용
    s := make([]int, 0, 1000000)
    for i := 0; i < 1000000; i++ {
        s = append(s, i)  // 용량이 변하지 않음
    }
    ```

4. 가비지 컬렉션 부하 감소

    재할당 시 이전 배열은 가비지 컬렉션의 대상이 됩니다.
    잦은 재할당은 가비지 컬렉션 횟수를 증가시킵니다.

    미리 할당하면 가비지 컬렉션 대상이 되는 객체 수가 줄어듭니다.
    이는 전체적인 애플리케이션의 성능과 반응성을 향상시킵니다.

5. 메모리 단편화 감소

    작은 크기의 메모리를 자주 할당하고 해제하면 메모리 단편화가 발생할 수 있습니다.
    큰 블록을 한 번에 할당함으로써 메모리 단편화를 줄일 수 있습니다.
    이는 전체적인 메모리 사용 효율성을 높입니다.

6. 코드의 의도 명확화

    코드는 실행뿐만 아니라 다른 개발자들이 읽고 이해해야 합니다.

    용량을 미리 지정함으로써 예상되는 데이터의 크기를 명확히 표현할 수 있습니다.
    이는 코드의 가독성과 유지보수성을 향상시킵니다.

    ```go
    // 의도가 불분명
    users := make([]User, 0)

    // 의도가 명확 (약 1000명의 사용자를 처리할 것으로 예상)
    users := make([]User, 0, 1000)
    ```
