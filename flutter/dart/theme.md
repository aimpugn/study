# Theme

- [Theme](#theme)
    - [`Theme.of(context).primaryColorDark`와 `primaryColorDark`의 차이](#themeofcontextprimarycolordark와-primarycolordark의-차이)
        - [`Theme.of(context).primaryColorDark`](#themeofcontextprimarycolordark)
        - [`primaryColorDark`](#primarycolordark)
        - [결론](#결론)

## `Theme.of(context).primaryColorDark`와 `primaryColorDark`의 차이

- `Theme.of(context).primaryColorDark`와 `primaryColorDark`는 Flutter에서 색상을 참조하는 두 가지 다른 방법
- 이들의 차이점을 이해하려면 Flutter의 `테마 시스템`과 `직접 정의한 색상 변`수에 대해 알고 있어야 한다

### `Theme.of(context).primaryColorDark`

- `Theme.of(context).primaryColorDark`는 현재 앱의 `ThemeData`에서 `primaryColorDark`를 참조한다
- 이는 앱 전반에 걸쳐 정의된 테마 설정에서 어두운 주 색상을 가져온다
- `MaterialApp` 위젯에서 `theme` 속성을 통해 정의한 `ThemeData` 객체 내에서 설정된 값
- `context`를 사용하여 현재 위젯 트리의 `ThemeData`에 접근하고, 이를 통해 다양한 위젯에서 일관된 색상과 스타일을 사용할 수 있다

예시:

```dart
// 현재 앱 테마의 `primaryColorDark` 값을 가져옵니다.
Color color = Theme.of(context).primaryColorDark;
```

### `primaryColorDark`

- `primaryColorDark`는 직접 정의한 색상 변수일 수 있다
- 이는 개발자가 앱 내 특정 부분에서 사용하기 위해 별도로 정의한 색상 값
- 이 색상은 `ThemeData`와 무관하게 직접 설정되고 관리된다

예시:

```dart
// `primaryColorDark`라는 이름의 변수를 직접 만들고, 
// 그 값을 `Colors.blueGrey[800]`으로 설정
Color primaryColorDark = Colors.blueGrey[800];
```

### 결론

- `Theme.of(context).primaryColorDark`는 **앱의 전역 테마 설정**에서 정의된 어두운 주 색상을 참조한다
- `primaryColorDark`는 개발자가 직접 정의한 색상 변수일 수 있으며, 이는 특정한 값으로 초기화되어 사용된다
