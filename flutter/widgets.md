# widgets

- [widgets](#widgets)
    - [`Stateless` Widgets와 `Stateful` Widgets](#stateless-widgets와-stateful-widgets)
        - [`Stateless` Widgets](#stateless-widgets)
        - [`Stateful` Widgets](#stateful-widgets)
        - [`StatelessWidget` 내에서 `StatefulWidget`을 사용하는 게 괜찮을까?](#statelesswidget-내에서-statefulwidget을-사용하는-게-괜찮을까)
            - [왜 가능한가?](#왜-가능한가)
    - [life cycle](#life-cycle)
        - [1. 위젯의 생성 (Instantiation)](#1-위젯의-생성-instantiation)
        - [2. 위젯의 빌드 과정 (Building)](#2-위젯의-빌드-과정-building)
        - [3. 상태 변화에 따른 재빌드 (Rebuilding)](#3-상태-변화에-따른-재빌드-rebuilding)
        - [4. 위젯의 파괴 (Destruction)](#4-위젯의-파괴-destruction)
    - [디버깅과 분석](#디버깅과-분석)
    - [ETC](#etc)

## `Stateless` Widgets와 `Stateful` Widgets

### `Stateless` Widgets

- 상태가 일단 생성되면 변경될 수 없는 위젯
- 변수, 버튼, 아이콘 등 앱에서 변경되지 않는 상태를 나타낸다
- 모든 Flutter 위젯은 build 메서드를 가지고 있다. 이 메서드는 해당 위젯이 화면에 어떻게 그려질지를 결정한다. `Stateless` 위젯에서는 이 `build` 메서드가 한 번만 호출되어 화면에 그려지고, 이후에는 상태가 변경되지 않으므로 다시 호출되지 않는다
- 데이터를 가져오거나 다른 방식으로 상태를 변경할 수 없다. 예를 들어, `Stateless` Widget 내부에서 API 호출을 통해 데이터를 가져오고 그 데이터를 표시하도록 변경할 수 없다. `Stateless` Widget이 앱에서 데이터를 가져오기 위해 그 상태를 변경할 수 없다
- UI가 객체 자체 내부의 정보에 의존할 때 이를 사용

```dart
import 'package:flutter/material.dart';

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container();
  }
}
```

### `Stateful` Widgets

### `StatelessWidget` 내에서 `StatefulWidget`을 사용하는 게 괜찮을까?

```dart
class ProfileSettingScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Column(
      // 각각의 위젯이 필요에 따라 자체적인 상태를 관리
      children: [
        ProfileImagePicker(), // StatefulWidget
        CharmPointTags(),    // StatefulWidget
        MyDropdown(),        // StatefulWidget
        // ... 기타 UI 구성요소
      ],
    );
  }
}
```

- `StatefulWidget`을 `StatelessWidget` 내에서 사용하는 것은 완전히 괜찮고, Flutter 앱 개발에서 일반적인 패턴이다
- Flutter의 위젯 트리에서 위젯들은 서로 다른 유형이 될 수 있으며, `StatelessWidget` 안에 `StatefulWidget`을 포함하는 것은 흔한 일이다

#### 왜 가능한가?

1. 위젯 구성의 유연성:
    - Flutter의 위젯 트리는 다양한 유형의 위젯을 혼합하여 구성할 수 있도록 설계되어 있다.
    - `StatefulWidget`과 `StatelessWidget`은 각각의 사용 사례에 따라 선택되며, 서로 다른 유형의 위젯이 혼합되어 사용될 수 있다
2. 상태 관리의 분리
    - `StatefulWidget`은 내부 상태를 가지고 있으며, 상태 변경에 따라 UI를 업데이트할 수 있다
    - 반면, `StatelessWidget`은 불변 상태를 가지며, 주로 상태가 없거나 상위 위젯으로부터 전달받은 데이터를 표시하는 데 사용된다
    - `StatelessWidget` 내에서 `StatefulWidget`을 사용하면, 상태 관리의 필요성에 따라 위젯의 책임을 적절히 분리할 수 있다
3. 재사용성 및 구조화:
    - `StatefulWidget`을 `StatelessWidget` 내에서 사용함으로써, 복잡한 UI 구성 요소를 재사용 가능한 작은 단위로 캡슐화할 수 있다
    - 이는 코드의 재사용성과 유지보수성을 향상시킨다

## life cycle

- Flutter의 모든 것은 위젯이므로 라이프사이클에 대해 알기 전에 Flutter의 위젯에 대해서 잘 알아야 한다
- Flutter의 위젯 라이프사이클은 크게 `생성`, `업데이트`, `파괴`의 세 단계로 나눌 수 있으며, 각 단계에서 위젯의 상태와 행동이 어떻게 변화하는지 이해하는 것이 중요하다

### 1. 위젯의 생성 (Instantiation)

- **위젯 생성**: 위젯은 `new` 또는 `const` 키워드를 사용하여 생성됩니다. 이때 위젯의 생성자가 호출되며, 위젯의 초기 구성이 정의됩니다.
- **BuildContext 생성**: Flutter 프레임워크는 위젯을 위젯 트리에 추가할 때 `BuildContext`를 생성합니다. 이 컨텍스트는 위젯의 위치를 위젯 트리에서 식별하는 데 사용됩니다.

### 2. 위젯의 빌드 과정 (Building)

- **`build` 메서드 호출**: Flutter 프레임워크는 위젯이 화면에 표시되어야 할 때 `build` 메서드를 호출합니다. `build` 메서드는 위젯의 UI를 구성하는 데 사용됩니다.
- **위젯의 렌더링**: `build` 메서드가 반환하는 위젯(주로 `Widget`의 서브클래스)은 Flutter의 렌더링 엔진에 의해 화면에 그려집니다.

### 3. 상태 변화에 따른 재빌드 (Rebuilding)

- **상태 변경**: 상태가 변경될 때 (예: 사용자 상호작용, 데이터 변경 등), 관련된 위젯은 새로운 상태를 반영하기 위해 재빌드됩니다.
- **`setState` 호출**: `StatefulWidget`에서는 `setState`를 호출하여 위젯을 재빌드할 수 있습니다. 이는 위젯의 `build` 메서드를 다시 호출하도록 합니다.

### 4. 위젯의 파괴 (Destruction)

- **위젯의 제거**: 위젯이 위젯 트리에서 제거될 때, 위젯과 관련된 리소스는 정리됩니다.
- **라이프사이클 종료**: `StatefulWidget`의 경우, `dispose` 메서드가 호출되어 위젯의 상태와 관련된 모든 리소스를 정리합니다.

## 디버깅과 분석

Flutter 애플리케이션에서 위젯의 `build` 메서드가 여러 번 호출되는 경우, 위젯의 상태 변화, 부모 위젯의 영향, 또는 Flutter 프레임워크의 최적화 동작 등을 고려하여 분석해야 합니다. `build` 메서드 내에서 로그를 추가하거나 디버거를 사용하여, 위젯의 라이프사이클 단계와 상태 변화를 추적할 수 있습니다.

Flutter의 위젯 라이프사이클을 이해하는 것은 애플리케이션의 성능 최적화와 문제 해결에 중요하며, 효과적인 앱 개발을 위한 기본적인 지식입니다. 위젯의 라이프사이클을 고려하여 위젯의 구조와 동작을 설계하는 것이 Flutter 개발의 핵심입니다.

## ETC

- [Explore Widget Lifecycle In Flutter](https://medium.flutterdevs.com/explore-widget-lifecycle-in-flutter-e36031c697d0)
