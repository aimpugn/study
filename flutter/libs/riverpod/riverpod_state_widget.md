# riverpod state widget

- [riverpod state widget](#riverpod-state-widget)
    - [`ConsumerWidget`과 `ConsumerStatefulWidget`](#consumerwidget과-consumerstatefulwidget)
        - [`ConsumerWidget`](#consumerwidget)
        - [`ConsumerStatefulWidget`](#consumerstatefulwidget)
            - [사용 목적?](#사용-목적)
                - [상태 관리와 UI 업데이트](#상태-관리와-ui-업데이트)
                - [Riverpod 상태 감시의 효율적인 통합](#riverpod-상태-감시의-효율적인-통합)
                - [상태 변화에 따른 사이드 이펙트 관리](#상태-변화에-따른-사이드-이펙트-관리)
            - [`StatelessWidget`와의 비교](#statelesswidget와의-비교)
        - [둘 중 무엇을 선택해야 할까?](#둘-중-무엇을-선택해야-할까)
    - [`ConsumerWidget`과 `StatelessWidget`](#consumerwidget과-statelesswidget)
        - [ConsumerWidget 사용](#consumerwidget-사용)
        - [StatelessWidget 사용](#statelesswidget-사용)
        - [결정 요인](#결정-요인)
        - [결론](#결론)

## `ConsumerWidget`과 `ConsumerStatefulWidget`

- `ConsumerWidget`과 `ConsumerStatefulWidget`을 선택하는 것은 각 위젯의 사용 목적과 상태 관리의 필요성에 따라 달라진다
- 대부분의 경우 `ConsumerWidget`으로 시작하여, 필요에 따라 `ConsumerStatefulWidget`으로 전환하는 것이 일반적인 접근 방식이라고 한다
- 상태 관리의 필요성과 UI의 동적 요소를 고려하여 적절한 위젯을 선택하는 것이 중요하다

### `ConsumerWidget`

- `ConsumerWidget`은 `StatelessWidget`의 변형
- Riverpod의 **provider를 소비(consume)하기 위해** 특별히 설계된 위젯
- `ConsumerWidget`은 다음과 같은 경우에 적합하다고 한다
    1. 정적 UI:
        - 위젯의 UI가 상태 변화에 의존하지 않거나, 내부 상태를 관리할 필요가 없는 경우.
    2. 상태 읽기만 필요:
        - 상태를 변경하지 않고 단순히 읽기만 하는 경우.
        - 예를 들어, 설정 정보나 사용자 데이터를 표시할 때.
    3. 간단한 상태 변화:
        - `provider`의 상태 변화에 의해 UI가 업데이트되는 간단한 상황. 내부적으로 `State`를 관리할 필요가 없을 때.

### `ConsumerStatefulWidget`

- `ConsumerStatefulWidget`은 Flutter의 `StatefulWidget`과 Riverpod의 기능을 결합한 것
    - Riverpod **provider와 상호작용하는 동시에 자체적인 상태를 관리**할 수 있고
    - Riverpod의 상태 관리 기능을 효율적으로 활용할 수 있게 해준다

#### 사용 목적?

Riverpod의 `ref`를 위젯의 상태(`State`)와 결합하여 관리하고, 상태 변화에 따라 UI를 동적으로 업데이트

1. 동적 UI:
    - 위젯의 UI가 사용자 상호작용이나 내부 로직에 의해 동적으로 변화해야 하는 경우.
2. 상태 변경 필요
    - 사용자의 입력이나 액션에 의해 내부 상태를 변경하고, 이에 따라 UI를 업데이트해야 하는 경우.
    - 예를 들어, 폼 입력 처리, 버튼 클릭 이벤트 등.
3. 복잡한 상태 관리:
    - 여러 상태들이 서로 상호작용하고, 복잡한 비즈니스 로직을 포함하는 경우.

##### 상태 관리와 UI 업데이트

- `ConsumerStatefulWidget`은 상태 관리와 UI 업데이트를 Riverpod와 함께 사용하기 위해 특별히 설계된 위젯
- `ConsumerStatefulWidget`은 내부적으로 위젯의 `State`를 관리하면서 Riverpod의 `ref`를 사용해 상태를 감시한다. 이를 통해 상태가 변경될 때마다 `State`가 업데이트되고, `build` 메서드를 통해 UI가 재구축된다
- 상태 변화에 따라 UI를 업데이트해야 하는 경우, `StatefulWidget`의 특성을 활용하여 위젯의 상태를 유지하고 변경에 반응하게 할 수 있다.

##### Riverpod 상태 감시의 효율적인 통합

- `ConsumerStatefulWidget`은 `StatefulWidget`의 기능과 Riverpod의 `ConsumerWidget` 기능을 결합한다. 이를 통해:
    - `StatefulWidget`의 모든 기능(예: 상태 유지, 애니메이션 관리 등) 사용 가능
    - Riverpod의 상태 감시 기능 사용 가능
    - Riverpod의 의존성 주입 기능 사용 가능
- `ConsumerStatefulWidget`의 `ConsumerState`를 사용하면 `build` 메서드에서 직접 `ref.watch`나 `ref.read`를 호출할 필요 없이, 상태를 감시하고 필요에 따라 위젯을 업데이트할 수 있다

##### 상태 변화에 따른 사이드 이펙트 관리

- `ConsumerStatefulWidget`을 사용하면, 상태 변화에 따른 사이드 이펙트(예: 데이터 가져오기, 네비게이션 이벤트, 애니메이션 등)를 `State` 내부에서 관리할 수 있다.
- `initState`, `dispose`, `didUpdateWidget` 등의 생명주기 메서드를 사용하여 복잡한 상태 관리 로직을 구현할 수 있다

#### `StatelessWidget`와의 비교

- `StatelessWidget`은
    - 상태 변화에 따른 UI 업데이트가 필요 없거나,
    - 간단한 상태 관리만 필요한 경우에 적합하다.
- 반면, `ConsumerStatefulWidget`은
    - 상태 변화에 따라 UI를 동적으로 업데이트하거나,
    - 복잡한 상태 관리와 사이드 이펙트를 처리해야 하는 경우에 더 적합하다
- 이러한 이유로, 상태 변화에 따라 UI를 업데이트하거나 복잡한 상태 관리가 필요한 경우 `ConsumerStatefulWidget`을 사용하는 것이 좋다

### 둘 중 무엇을 선택해야 할까?

- 상태 관리의 복잡성
    - 위젯 내부에서 상태 관리의 복잡성이 높고, 여러 상태 변화를 처리해야 한다면 `ConsumerStatefulWidget` 사용을 고려
- UI의 동적 변화
    - 사용자 상호작용에 의한 UI의 동적 변화가 필요하다면 `ConsumerStatefulWidget`이 더 적합할 수 있다
- 성능 고려사항
    - `ConsumerWidget`은 상태 변경 시 전체 위젯을 다시 빌드한다
    - 따라서 성능이 중요한 경우, 불필요한 리빌드를 줄이기 위해 `ConsumerStatefulWidget`을 사용하여 특정 부분만 업데이트할 수 있다

## `ConsumerWidget`과 `StatelessWidget`

- `ConsumerWidget`과 `StatelessWidget` 사이의 선택은 주로 사용하는 상황과 상태 관리의 필요성에 따라 달라진다

### ConsumerWidget 사용

- 상황: 위젯이 Riverpod의 `Provider`를 직접 소비하고, 상태 변경에 반응해야 할 때 사용한다
- 장점: `ConsumerWidget`은 `WidgetRef` 객체를 제공하여, `Provider`로부터 상태를 읽거나, 리스너를 등록하거나, 상태를 변경하는 것을 간편하게 만든다
- 예시: 특정 `Provider`의 상태에 따라 UI가 변경되어야 하거나, 상태 변경을 위한 사용자 인터랙션이 있는 경우.

### StatelessWidget 사용

- 상황: 위젯이 상태 변경에 의존하지 않거나, **부모 위젯으로부터 필요한 모든 데이터를 매개변수로 전달받을 수 있을 때** 사용
- 장점: `StatelessWidget`은 상태 관리 로직이 없거나, 상태가 부모 위젯에서 관리될 때 더 단순하고 이해하기 쉽다
- 예시: 정적인 UI 요소, 데이터 표시 등 상태 변경이나 상호작용이 필요 없는 경우.

### 결정 요인

- 상태 의존성: 위젯이 상태 변경에 따라 업데이트되어야 하는지 여부.
- 상태 접근 빈도: 위젯이 얼마나 자주 `Provider`의 상태에 접근하는지 여부.
- 코드의 간결성 및 재사용성: 각 위젯 유형이 전체적인 코드 구조에 미치는 영향.

### 결론

Riverpod를 사용하는 경우에도 모든 위젯을 `ConsumerWidget`으로 만들 필요는 없다. 상태 관리가 필요한 위젯에서만 `ConsumerWidget`을 사용하고, 그 외에는 `StatelessWidget`을 사용하는 것이 일반적으로 권장된다. 이렇게 하면 코드의 복잡성을 관리할 수 있고, 각 위젯의 목적과 역할을 더 명확히 할 수 있다.
