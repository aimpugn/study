# gochecknoglobals

- [gochecknoglobals](#gochecknoglobals)
    - [gochecknoglobals?](#gochecknoglobals-1)
    - [Enum 사용 시 전역 변수 문제 해결 방법](#enum-사용-시-전역-변수-문제-해결-방법)

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
