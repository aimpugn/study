# icon checkbox

## 아이콘과 텍스트를 모두 포함하는 탭 가능한 영역을 만들기

```text
✔️ 로그인 상태 유지하기
```

- `IconButton`은 아이콘을 위해 최적화된 위젯이기 때문에, `IconButton` 내에 텍스트를 넣는 것은 일반적으로 권장되지 않는다.
- 대신, 아이콘과 텍스트를 함께 사용하려면 `Row` 위젯 내에 `Icon`과 `Text` 또는 `CustomTextButton`을 별도의 위젯으로 두고, 각각에 탭 기능을 추가하는 방법을 사용할 수 있다

### 방법 1: `Row` 내 `GestureDetector` 사용

- `Row` 위젯 내에 `Icon`과 `Text`를 배치하고, 전체 `Row`를 `GestureDetector`로 감싸 탭 이벤트를 처리한다
- 이 방법은 클릭 가능한 영역을 아이콘과 텍스트 모두로 확장한다.

```dart
Row(
  mainAxisAlignment: MainAxisAlignment.end,
  children: [
    GestureDetector(
      onTap: () {
        setState(() {
          _keepSignIn = !_keepSignIn;
        });
      },
      child: Row(
        children: [
          Icon(
            _keepSignIn ? Icons.check_circle : Icons.check_circle_outline,
            color: _keepSignIn ? primaryBackgroundColor : Colors.grey,
          ),
          Padding(
            padding: EdgeInsets.only(left: 8.0),
            child: Text(l10n.labelKeepSignIn),
          ),
        ],
      ),
    ),
  ],
),
```

### 방법 2: `InkWell` 또는 `TextButton` 사용

- 아이콘과 텍스트를 `InkWell` 또는 `TextButton` 위젯으로 감싸 클릭 이벤트를 처리
- 이 방법은 머티리얼 디자인의 잉크 효과를 제공한다

```dart
Row(
  mainAxisAlignment: MainAxisAlignment.end,
  children: [
    InkWell(
      onTap: () {
        setState(() {
          _keepSignIn = !_keepSignIn;
        });
      },
      child: Row(
        children: [
          Icon(
            _keepSignIn ? Icons.check_circle : Icons.check_circle_outline,
            color: _keepSignIn ? primaryBackgroundColor : Colors.grey,
          ),
          Padding(
            padding: EdgeInsets.only(left: 8.0),
            child: Text(l10n.labelKeepSignIn),
          ),
        ],
      ),
    ),
  ],
),
```

이러한 방법들은 사용자가 아이콘 또는 텍스트 어느 쪽을 탭하더라도 동일한 동작을 수행하도록 해줍니다. `GestureDetector`, `InkWell`, 또는 `TextButton` 위젯은 클릭 이벤트를 처리하는 데 유용하며, `Row` 위젯을 사용하면 아이콘과 텍스트를 수평으로 배치할 수 있습니다.
