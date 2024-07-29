# FocusScope

## [FocusScope](https://api.flutter.dev/flutter/widgets/FocusScope-class.html)

- `FocusScope` 위젯은 Flutter에서 키보드 포커스를 관리하는 데 사용된다
- [`Focus`](https://api.flutter.dev/flutter/widgets/Focus-class.html) 클래스와 비슷하지만, 그 하위 항목에 대한 scope 역할을 하여 초점 이동(focus traversal)을 scoped controls로 제한한다
    - 특정 부분의 위젯 트리 내에서 포커스를 관리하고, 포커스를 받을 수 있는 위젯들 간의 포커스 이동을 제어할 수 있다.
    - 간단히 말해서, `FocusScope`는 포커스 컨트롤의 범위를 제한하고 조정하는 역할을 합니다.

## `FocusScope` 주요 기능

### 포커스 관리

- `FocusScope`는 자식 위젯들 중 어느 것이 현재 포커스를 가지고 있는지 추적
- 이를 통해 키보드 입력과 같은 이벤트를 현재 포커스를 가진 위젯에 전달한다

### 포커스 이동

- `FocusScope`를 사용하면 프로그래밍 방식으로 포커스를 특정 위젯으로 이동시킬 수 있다
- 예를 들어, 폼 필드 간에 사용자가 '다음' 버튼을 눌렀을 때 자동으로 포커스를 이동시키는 것이 가능하다

### 포커스 범위 설정

- 애플리케이션의 특정 부분 내에서만 포커스를 관리하고 싶을 때 `FocusScope`를 사용할 수 있다
- 이를 통해 한 부분에서 포커스 관리를 독립적으로 수행할 수 있으며, 다른 부분의 포커스 상태에 영향을 받지 않는다

### 키보드 단축키 관리

- `FocusScope`를 사용하여 특정 키보드 단축키가 활성화될 때 어떤 동작을 할지 결정할 수 있다
- 예를 들어, 특정 키 조합을 누를 때 특정 위젯으로 포커스를 이동시키는 등의 작업을 수행할 수 있다

## `FocusScope` 사용 예시

```dart
FocusScope(
  node: focusScopeNode,
  child: Column(
    children: <Widget>[
      TextField(),
      TextField(),
      ElevatedButton(
        onPressed: () {
          focusScopeNode.nextFocus(); // 다음 텍스트 필드로 포커스 이동
        },
        child: Text('다음'),
      ),
    ],
  ),
)
```

이 예시에서는 `FocusScope`를 사용하여 두 개의 텍스트 필드와 버튼을 감쌌습니다. 버튼을 누를 때마다 `nextFocus()` 메서드를 호출하여 다음 텍스트 필드로 포커스가 자동으로 이동합니다. 이러한 방식으로 `FocusScope`는 사용자 인터페이스의 포커스 관리를 보다 효율적으로 만들어 줍니다.
