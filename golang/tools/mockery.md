# mockery

- [mockery](#mockery)
    - [installation](#installation)
    - [configuration](#configuration)
    - [mock 인스턴스 생성](#mock-인스턴스-생성)
        - [`new(mocks.SomeInterface)` vs `mocks.NewMockSomeInterface()`](#newmockssomeinterface-vs-mocksnewmocksomeinterface)
            - [`mocks.NewMockSomeInterface(t)`](#mocksnewmocksomeinterfacet)
            - [`new(mocks.MockSomeInterface)`](#newmocksmocksomeinterface)

## installation

## configuration

## mock 인스턴스 생성

### `new(mocks.SomeInterface)` vs `mocks.NewMockSomeInterface()`

#### `mocks.NewMockSomeInterface(t)`

1. **초기화와 설정**

    이 방식은 `mocks` 패키지에서 `NewMockSomeInterface` 함수를 호출하여 `SomeInterface`의 mock 인스턴스를 생성한다.
    이 함수는 보통 매개변수로 테스트 컨텍스트나 테스팅 객체(`t`)를 받아, 생성된 mock 객체가 테스트의 생명 주기와 연동되게 한다.
    예를 들어, `testify/mock` 라이브러리에서는 mock 객체의 예상이 모두 충족되었는지 검증하는 등의 추가적인 기능을 제공할 수 있다.

2. **기능성**

    `NewMockSomeInterface` 함수는 mock 객체를 초기화하고 필요한 설정을 적용할 수 있는 기회를 제공한다.
    이는 단순히 객체를 생성하는 것 이상의 작업을 할 수 있게 하며, mock 객체가 테스트 동안 필요로 하는 여러 설정이나 초기 상태를 구성할 수 있다.

3. **테스트 통합**

    생성된 mock 객체는 보통 테스트 프레임워크와 밀접하게 통합되어 있으며, 테스트 실패 시 유용한 정보나 에러 메시지를 제공할 수 있다.

#### `new(mocks.MockSomeInterface)`

1. **기본 초기화**

    `new` 키워드는 Go에서 제공하는 기본 내장 함수로, 지정된 타입의 새로운 인스턴스를 생성하고, 그 타입의 포인터를 반환한다.
    이 경우 `mocks.MockSomeInterface` 타입의 새 인스턴스를 생성하고, 그것의 포인터(`*mocks.MockSomeInterface`)를 반환한다.
    `new` 함수는 객체의 필드를 기본값으로 초기화합니다만, 추가적인 설정이나 초기화 작업은 수행하지 않는다.

2. **제한된 기능성**

    `new`를 사용하여 생성된 mock 객체는 `mocks.NewMockSomeInterface(t)`를 통해 생성된 객체와 비교했을 때, 보다 제한된 기능성을 갖는다.
    즉, 추가적인 설정이나 초기화 로직 없이 객체가 생성된다.
    이 방식은 테스트 특화 로직이나 검증 로직 없이 객체의 기본 상태만 필요한 경우에 유용할 수 있다.

3. **테스트 통합 부족**

    `new`를 통해 생성된 객체는 테스트 컨텍스트나 테스팅 객체와 직접적으로 연동되지 않는다.
    따라서 테스트 실패 시 추가적인 정보를 제공하거나, 특정 테스트 기대치를 자동으로 검증하는 등의 기능을 기대하기 어렵다.
