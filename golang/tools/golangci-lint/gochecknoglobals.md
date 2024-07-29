# gochecknoglobals

- [gochecknoglobals](#gochecknoglobals)
    - [gochecknoglobals?](#gochecknoglobals-1)
    - [Enum 사용 시 전역 변수 문제 해결 방법](#enum-사용-시-전역-변수-문제-해결-방법)
    - [READ ONLY 전역 변수](#read-only-전역-변수)
    - [전역 변수를 함수로 변경할 경우](#전역-변수를-함수로-변경할-경우)
    - [전역 변수를 함수로 바꿀 때의 단점](#전역-변수를-함수로-바꿀-때의-단점)
    - [대안 및 고려사항](#대안-및-고려사항)
    - [결론](#결론)

## gochecknoglobals?

`gochecknoglobals` 린트는 Go에서 전역 변수의 사용을 경고하는 도구입니다.
전역 변수는 여러 고루틴에서 동시에 접근할 경우 예상치 못한 부작용이나 데이터 경쟁(race condition)을 일으킬 수 있기 때문에, 가능한 한 사용을 피하는 것이 좋습니다.
그러나, 설정 값이나 상수와 같이 변경되지 않는 데이터를 전역으로 선언하는 것은 일반적인 패턴입니다.

전역 변수 사용은 주의 깊게 고려해야 하며, 가능한 한 상수나 구조체를 통한 캡슐화를 사용하는 것이 좋습니다. `gochecknoglobals` 린트는 코드의 안정성과 유지보수성을 높이기 위해 전역 변수 사용을 제한하려는 목적으로 사용됩니다.

## Enum 사용 시 전역 변수 문제 해결 방법

1. **상수로 선언하기**:
   Enum과 같은 경우, 값이 변경되지 않도록 상수(`const`)로 선언하는 것이 일반적입니다. Go에서는 `iota`를 사용하여 enum 비슷한 패턴을 구현할 수 있습니다.

   ```go
   type EnumType int

   const (
       EnumAAA EnumType = iota
       EnumBBB
       // ...
   )
   ```

   이 방법은 `EnumAAA`, `EnumBBB` 등이 컴파일 타임에 결정되며, 런타임에 변경될 수 없기 때문에 안전합니다.

2. **구조체와 메서드 사용하기**:
   전역 변수 대신에 구조체를 사용하여 상태를 관리하고, 필요한 enum 값을 메서드를 통해 제공할 수 있습니다. 이 방법은 객체 지향적 접근 방식을 통해 데이터 캡슐화를 강화할 수 있습니다.

   ```go
   type Enums struct {
       AAA EnumType
       BBB EnumType
       // ...
   }

   func NewEnums() *Enums {
       return &Enums{
           AAA: 1, // 예시 값
           BBB: 2, // 예시 값
           // ...
       }
   }
   ```

   이렇게 하면, `Enums` 타입의 인스턴스를 생성하여 사용할 때 각 인스턴스마다 독립적인 상태를 유지할 수 있습니다.

3. **린트 경고 무시하기**:
   경우에 따라 전역 변수 사용이 정당화될 수 있으며, `gochecknoglobals` 경고를 무시해야 할 수도 있습니다. 이를 위해 특정 라인이나 파일에 대해 린트 경고를 무시하는 주석을 추가할 수 있습니다.

   ```go
   //nolint:gochecknoglobals
   var (
       EnumAAA EnumType = 1
       EnumBBB EnumType = 2
       // ...
   )
   ```

   이 주석은 해당 파일이나 코드 블록에서 `gochecknoglobals` 경고를 무시하도록 지시합니다.

## READ ONLY 전역 변수

이 상황에 대한 몇 가지 해결 방법을 제안드리겠습니다:

1. 패키지 초기화 함수 사용:

    Go에서는 `init()` 함수를 사용하여 패키지 레벨 변수를 초기화할 수 있습니다.
    이 방법을 사용하면 전역 변수를 피하면서도 한 번만 초기화할 수 있습니다.

    ```go
    var currencyScalingFactors map[int]*big.Int

    func init() {
        currencyScalingFactors = map[int]*big.Int{
            0: new(big.Int).Exp(big.NewInt(10), big.NewInt(0), nil),
            1: new(big.Int).Exp(big.NewInt(10), big.NewInt(1), nil),
            2: new(big.Int).Exp(big.NewInt(10), big.NewInt(2), nil),
            3: new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil),
            4: new(big.Int).Exp(big.NewInt(10), big.NewInt(4), nil),
        }
    }
    ```

2. 싱글톤 패턴 사용:
    싱글톤 패턴을 사용하여 한 번만 초기화되는 인스턴스를 만들 수 있습니다.

    ```go
    var (
        currencyScalingFactorsOnce sync.Once
        currencyScalingFactors     map[int]*big.Int
    )

    func getCurrencyScalingFactors() map[int]*big.Int {
        currencyScalingFactorsOnce.Do(func() {
            currencyScalingFactors = map[int]*big.Int{
                0: new(big.Int).Exp(big.NewInt(10), big.NewInt(0), nil),
                1: new(big.Int).Exp(big.NewInt(10), big.NewInt(1), nil),
                2: new(big.Int).Exp(big.NewInt(10), big.NewInt(2), nil),
                3: new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil),
                4: new(big.Int).Exp(big.NewInt(10), big.NewInt(4), nil),
            }
        })
        return currencyScalingFactors
    }
    ```

    그리고 `ToFloat64` 함수에서 다음과 같이 사용합니다:

    ```go
    f.Quo(f, new(big.Float).SetInt(getCurrencyScalingFactors()[minorUnit]))
    ```

3. 상수 사용:
    만약 이 값들이 정말로 상수라면, `const`를 사용할 수 있습니다. 단, `big.Int`는 상수로 사용할 수 없으므로 일반 정수를 사용해야 합니다.

    ```go
    const (
        CurrencyScalingFactor0 = 1
        CurrencyScalingFactor1 = 10
        CurrencyScalingFactor2 = 100
        CurrencyScalingFactor3 = 1000
        CurrencyScalingFactor4 = 10000
    )

    func getCurrencyScalingFactor(minorUnit int) int64 {
        switch minorUnit {
        case 0:
            return CurrencyScalingFactor0
        case 1:
            return CurrencyScalingFactor1
        case 2:
            return CurrencyScalingFactor2
        case 3:
            return CurrencyScalingFactor3
        case 4:
            return CurrencyScalingFactor4
        default:
            return 1
        }
    }
    ```

    그리고 `ToFloat64` 함수에서 다음과 같이 사용합니다:

    ```go
    f.Quo(f, new(big.Float).SetInt64(getCurrencyScalingFactor(minorUnit)))
    ```

## 전역 변수를 함수로 변경할 경우

네, 귀하의 분석이 정확합니다. 전역 변수를 함수로 바꾸는 것에는 몇 가지 단점이 있습니다. 이에 대해 상세히 설명드리겠습니다.

## 전역 변수를 함수로 바꿀 때의 단점

1. **메모리 사용 증가**
    - 전역 변수: 프로그램 시작 시 한 번만 초기화되어 메모리에 상주합니다.
    - 함수: 호출될 때마다 맵을 새로 생성하고 메모리에 할당합니다.

2. **성능 저하**
    - 전역 변수: 접근 시 즉시 사용 가능합니다.
    - 함수: 호출될 때마다 맵을 재생성해야 하므로, 특히 큰 맵의 경우 성능 저하가 발생할 수 있습니다.

3. **가비지 컬렉션 부하**
    - 전역 변수: 프로그램 종료 시까지 메모리에 남아있어 가비지 컬렉션 대상이 아닙니다.
    - 함수: 함수 호출마다 새로운 맵 객체가 생성되어 가비지 컬렉션의 대상이 됩니다. 이는 가비지 컬렉터에 추가적인 부하를 줄 수 있습니다.

4. **동시성 고려사항**
    - 전역 변수: 읽기 전용이라면 동시성 문제가 없습니다.
    - 함수: 매번 새로운 인스턴스를 생성하므로 동시성 문제는 없지만, 여러 고루틴에서 동시에 호출될 경우 각각 별도의 맵 인스턴스를 생성하게 됩니다.

5. **코드 복잡성 증가**
    - 전역 변수: 직접 접근이 가능하여 사용이 간단합니다.
    - 함수: 매번 함수를 호출해야 하며, 이는 코드를 약간 더 복잡하게 만들 수 있습니다.

6. **힙 할당**
    - 전역 변수: 정적 메모리에 할당됩니다.
    - 함수: 반환된 맵은 힙에 할당되며, 이는 추가적인 메모리 관리를 필요로 합니다.

7. **초기화 지연**
    - 전역 변수: 프로그램 시작 시 즉시 사용 가능합니다.
    - 함수: 처음 호출될 때까지 초기화가 지연됩니다. 이는 때때로 장점이 될 수도 있지만, 초기 접근 시 지연을 발생시킬 수 있습니다.

## 대안 및 고려사항

1. **sync.Once 사용**

   ```go
   var (
       aSpecificInfoMapOnce sync.Once
       aSpecificInfoMapInstance map[string]*SomeInfo
   )

   func getASpecificInfoMap() map[string]*SomeInfo {
       aSpecificInfoMapOnce.Do(func() {
           aSpecificInfoMapInstance = map[string]*SomeInfo{
               // ... 맵 초기화
           }
       })
       return aSpecificInfoMapInstance
   }
   ```

   이 방법은 전역 변수의 이점(한 번만 초기화)과 함수의 이점(지연 초기화)을 결합합니다.

2. **상수 맵 사용 고려**
   Go 1.21부터는 맵 리터럴을 `const`로 선언할 수 있습니다. 이는 컴파일 타임에 맵을 초기화하여 런타임 오버헤드를 줄일 수 있습니다.

3. **코드 생성 도구 사용**
   맵이 정적이고 변경되지 않는다면, 코드 생성 도구를 사용하여 맵을 상수 또는 함수로 생성할 수 있습니다.

## 결론

전역 변수를 함수로 바꾸는 것은 린트 경고를 해결할 수 있지만, 성능과 메모리 사용 측면에서 단점이 있습니다. 특히 맵이 크고 자주 접근되는 경우 이러한 단점이 더 두드러질 수 있습니다.

최적의 해결책은 사용 패턴, 성능 요구사항, 그리고 코드 유지보수성을 고려하여 결정해야 합니다. `sync.Once`를 사용하는 방법이나 Go 1.21 이상에서 제공하는 상수 맵 기능을 고려해보는 것도 좋은 대안이 될 수 있습니다.
