# iota

## [iota](https://en.wikipedia.org/wiki/Iota)?

Iota는 매우 작은 양을 나타내는 용어로 사용되기도 하지만, 프로그래밍 언어 Go에서는 특별한 의미를 가진다.
Go 언어에서 `iota`는 상수 생성기이며, `const` 블록에서 사용될 때 초기화되어 이후에 나오는 각 상수에 대해 자동으로 증가하는 정수 값을 제공한다.

`Iota`라는 단어 자체는 그리스 알파벳의 아홉 번째 글자인 'ι' (iota)에서 유래했으며, 이 글자가 작은 크기로 인해 영어로는 '아주 작은 양'이라는 의미로 사용되었다.

Go 언어에서의 사용은 이 단어가 가진 '작은 것에서 시작한다'는 의미에서 차용된 것으로 보인다.

## Go 언어에서의 Iota 사용례

Go에서 `iota`는 상수 그룹을 선언할 때 간편하게 숫자를 할당할 수 있게 해준다. 이 기능은 주로 열거형(enumeration) 같은 상수의 집합을 정의할 때 유용하게 사용된다.

**기본 예제**:

```go
package main

import "fmt"

const (
    First = iota  // 0
    Second        // 1
    Third         // 2
)

func main() {
    fmt.Println(First, Second, Third)  // 출력: 0 1 2
}
```

**비트 마스크 예제**:
`iota`를 사용하여 각 상수가 이전 상수의 두 배 값이 되도록 할 수 있다. 이는 비트 필드나 비트 마스킹 작업에 유용하다.

```go
package main

import "fmt"

const (
    Flag1 = 1 << iota  // 1 << 0 == 1
    Flag2              // 1 << 1 == 2
    Flag3              // 1 << 2 == 4
    Flag4              // 1 << 3 == 8
)

func main() {
    fmt.Println(Flag1, Flag2, Flag3, Flag4)  // 출력: 1 2 4 8
}
```

**응용 예제**:
`iota`를 사용하여 더 복잡한 수식도 계산할 수 있다. 다음 예제에서는 각 상수 값에 대해 `iota`의 제곱 값을 할당한다.

```go
package main

import "fmt"

const (
    Zero = iota * iota  // 0*0 == 0
    One                 // 1*1 == 1
    Four                // 2*2 == 4
    Nine                // 3*3 == 9
)

func main() {
    fmt.Println(Zero, One, Four, Nine)  // 출력: 0 1 4 9
}
```

이러한 예들을 통해 `iota`의 활용 가능성을 확인할 수 있으며, Go 프로그래밍에서 상수를 효율적으로 관리하는 데 매우 유용한 도구임을 알 수 있다.
