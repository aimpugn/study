# `WidgetsBinding`

- [`WidgetsBinding`](#widgetsbinding)
    - [`WidgetsBinding` mixin?](#widgetsbinding-mixin)
    - [역할과 사용 사례](#역할과-사용-사례)
    - [예시](#예시)
    - [중요 사항](#중요-사항)

## `WidgetsBinding` mixin?

The glue between the widgets layer and the Flutter engine.

The `WidgetsBinding` manages a single Element tree rooted at rootElement. Calling runApp (which indirectly calls attachRootWidget) bootstraps that element tree.

- `WidgetsBinding.instance.addPostFrameCallback`은 Flutter에서 매우 유용한 기능으로, 현재 프레임의 렌더링 작업이 완료된 후에 실행되어야 하는 콜백을 등록하는 데 사용됩니다. 이 함수는 특히 UI 렌더링이 완료된 직후에 어떤 작업을 수행해야 할 때 매우 유용합니다.

## 역할과 사용 사례

- `addPostFrameCallback`을 사용하면 **현재 프레임의 렌더링이 완료된 후**에 콜백이 실행된다. 이는 렌더링 프로세스가 완전히 완료된 시점이므로, 이 때 UI의 최종 상태에 접근할 수 있다.
- 따라서 UI 업데이트 후 추가 작업을 할 수 있다. 예를 들어,
    - 위젯이 렌더링된 후에 어떤 조건에 따라 다른 페이지로 이동할 경우
    - 대화 상자를 표시하는 등의 작업을 수행할 경우
- 애니메이션 및 레이아웃 계산시 사용할 수 있다. 화면에 위젯이 표시된 후에 그 위젯의 크기나 위치 등을 계산해야 할 때 유용하다. 예를 들어,
    - 특정 위젯의 위치를 기반으로 팝업을 표시하려면, 위젯의 레이아웃이 완전히 결정된 후에 그 위치 정보를 얻어야 한다

## 예시

```dart
WidgetsBinding.instance.addPostFrameCallback((_) {
  if (someCondition) {
    Navigator.of(context).pushNamed('/someRoute');
  }
});
```

이 코드는 현재 프레임의 렌더링이 완료된 후에 조건에 따라 새로운 루트로 이동하는 로직을 구현합니다. `addPostFrameCallback`을 사용함으로써, 렌더링 프로세스가 완전히 끝나고 나서 페이지 이동이 발생하므로, 렌더링 중에 발생할 수 있는 문제들을 예방할 수 있다.

## 중요 사항

- `addPostFrameCallback`은 현재 프레임이 렌더링된 후에 한 번만 실행되며, 이후에는 자동으로 콜백이 해제된다.
- UI 업데이트 중에 페이지 이동, 대화 상자 표시 등의 작업을 수행하려면, 직접적으로 해당 작업을 수행하는 대신 `addPostFrameCallback`을 사용하여 UI 렌더링이 완료된 후에 작업을 수행하는 것이 좋다
