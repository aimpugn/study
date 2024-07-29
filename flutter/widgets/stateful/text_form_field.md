# TextFormField

- [TextFormField](#textformfield)
    - [TextFormField](#textformfield-1)
    - [예시](#예시)
    - [속성](#속성)
    - [`decoration`](#decoration)
        - [`errorText`](#errortext)
    - [`validator`](#validator)
        - [`textInputAction`](#textinputaction)
            - [주요 특징](#주요-특징)
            - [`TextInputAction.next`](#textinputactionnext)
            - [`TextInputAction.none`](#textinputactionnone)
            - [`TextInputAction.unspecified`](#textinputactionunspecified)
            - [`TextInputAction.done`](#textinputactiondone)
            - [`TextInputAction.go`](#textinputactiongo)
            - [`TextInputAction.search`](#textinputactionsearch)
            - [`TextInputAction.send`](#textinputactionsend)
            - [`TextInputAction.previous`](#textinputactionprevious)
            - [`TextInputAction.continueAction`](#textinputactioncontinueaction)
            - [`TextInputAction.join`](#textinputactionjoin)
            - [`TextInputAction.route`](#textinputactionroute)
            - [`TextInputAction.emergencyCall`](#textinputactionemergencycall)
            - [`TextInputAction.newline`](#textinputactionnewline)

## TextFormField

## 예시

```dart
class MyForm extends StatefulWidget {
  @override
  _MyFormState createState() => _MyFormState();
}

class _MyFormState extends State<MyForm> {
  final _formKey = GlobalKey<FormState>();

  @override
  Widget build(BuildContext context) {
    return Form(
      key: _formKey,
      child: Column(
        children: <Widget>[
          TextFormField(
            decoration: InputDecoration(hintText: 'Enter your nickname'),
            validator: (value) {
              if (value == null || value.isEmpty) {
                return 'Please enter some text';
              }
              return null;
            },
          ),
          // ... 다른 TextFormField 위젯들
          ElevatedButton(
            onPressed: () {
              if (_formKey.currentState!.validate()) {
                // Form data is valid
              }
            },
            child: Text('Submit'),
          ),
        ],
      ),
    );
  }
}
```

## 속성

## `decoration`

### `errorText`

- Flutter에서 `TextFormField`와 같은 입력 필드에서 실시간으로 오류 메시지를 표시하고자 할 때, 일반적으로 `validator` 함수를 사용한다
- 하지만 `validator`는 폼이 제출될 때 (예: 사용자가 로그인 버튼을 클릭할 때) 실행되며, 사용자가 입력을 하는 동안 실시간으로 반응하지는 않는다
- 실시간으로 오류 메시지를 표시하고자 할 때에는 몇 가지 방법이 있다:
    - **상태 관리를 이용한 방법**
        1. `TextEditingController`의 리스너를 사용하여 입력값이 변경될 때마다 유효성을 검사하고,
        2. 이 결과를 상태 변수에 저장한 다음,
        3. 이 상태에 따라 `TextFormField`의 `decoration` 속성에서 오류 메시지를 동적으로 표시
    - **Global Key를 사용한 방법**
        1. `Form` 위젯에 `GlobalKey<FormState>`를 할당하고,
        2. 이 키를 사용하여 폼의 상태에 접근한다
        3. 사용자가 입력을 할 때마다 `formKey.currentState.validate()`를 호출하여 `validator` 함수를 강제로 실행시키면,
        4. 입력 필드에 실시간으로 오류 메시지를 표시할 수 있다

```dart
class MyForm extends StatefulWidget {
  @override
  _MyFormState createState() => _MyFormState();
}

class _MyFormState extends State<MyForm> {
  final _passwordController = TextEditingController();
  final _passwordCheckController = TextEditingController();
  String _passwordError;

  @override
  void initState() {
    super.initState();
    _passwordCheckController.addListener(_validatePassword);
  }

  void _validatePassword() {
    if (_passwordController.text != _passwordCheckController.text) {
      setState(() {
        _passwordError = "비밀번호가 일치하지 않습니다.";
      });
    } else {
      setState(() {
        _passwordError = null;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: _passwordCheckController,
      decoration: InputDecoration(
        errorText: _passwordError,
      ),
    );
  }
}
```

이 코드는 `_passwordController`와 `_passwordCheckController`의 값이 일치하지 않을 경우 `_passwordError` 상태를 업데이트하여 `TextFormField`의 `errorText`에 오류 메시지를 표시합니다.

실시간으로 유효성 검사를 수행하는 것은 사용자 경험을 향상시키는 좋은 방법이지만, 동시에 사용자 인터페이스가 너무 자주 변경되어 사용자에게 혼란을 줄 수도 있습니다. 따라서 이런 기능을 구현할 때는 사용자 인터페이스의 일관성과 사용자 경험을 고려하는 것이 중요합니다.

## `validator`

```dart
/// Signature for validating a form field.
///
/// Returns an error string to display if the input is invalid, or null
/// otherwise.
///
/// Used by [FormField.validator].
typedef FormFieldValidator<T> = String? Function(T? value);


final FormFieldValidator<T>? validator;
```

### `textInputAction`

- `TextInputAction` 열거형(enum)은 Flutter에서 텍스트 입력 필드와 관련된 키보드의 동작을 정의한다
- 이 열거형은 사용자가 키보드의 특정 버튼(예: "완료", "다음", "검색" 등)을 눌렀을 때 어떠한 동작을 할 것인지를 지정한다
- 각각의 `TextInputAction` 값은 논리적인 의미를 가지며, 이에 따라 소프트 키보드는 특정 종류의 동작 버튼을 표시한다

#### 주요 특징

1. **논리적 의미**: 각 액션은 논리적인 의미를 가지고 있으며, 키보드에 특정한 종류의 동작 버튼을 표시하게 한다
2. **플랫폼별 차이**:
    - Android와 iOS 모두 대부분의 `TextInputAction`을 지원하지만, 완벽하게 1:1로 매핑되는 것은 아니다
    - 예를 들어, Android의 `IME_ACTION_NEXT`와 iOS의 `UIReturnKeyNext`는 둘 다 "다음"으로의 이동을 의미하지만, 실제 표시되는 버튼의 모양이나 동작은 플랫폼에 따라 다를 수 있다
3. **개발자의 책임**:
    - 특정 `TextInputAction`을 선택했다고 해서 반드시 특정 동작이 발생하는 것은 아니다.
    - 예를 들어, "Emergency Call" 버튼을 눌렀을 때 다음 텍스트 필드로 포커스가 이동하는 것은 논리적으로 적절하지 않으며, 개발자는 액션 버튼이 눌렸을 때 적절한 동작이 이루어지도록 해야 한다
4. **디버그 모드에서의 오류 처리**:
    - 디버그 모드에서 플랫폼에 부적절한 `TextInputAction`을 선택하면 오류가 발생한다
    - 릴리스 모드에서는 부적절한 값 대신 플랫폼별 기본값("unspecified" 또는 "default")이 사용된다

#### `TextInputAction.next`

현재 입력 소스를 완료하고 "다음" 입력 필드로 이동하고자 할 때 사용됩니다. 사용자가 키보드의 "다음" 버튼을 누를 때 다음 텍스트 입력 필드로 포커스가 이동합니다.

#### `TextInputAction.none`

현재 입력 소스에 관련된 입력 액션이 없음을 의미합니다.

#### `TextInputAction.unspecified`

OS가 가장 적절한 액션을 결정하게 합니다.

#### `TextInputAction.done`

사용자가 입력을 완료하고 어떤 최종화 동작을 수행해야 함을 나타냅니다.

#### `TextInputAction.go`

사용자가 입력한 텍스트(예: 레스토랑 이름)에 해당하는 앱의 부분으로 이동하길 원할 때 사용됩니다.

#### `TextInputAction.search`

검색 쿼리를 실행하고자 할 때 사용됩니다.

#### `TextInputAction.send`

사용자가 작성한 내용(예: 이메일, 문자 메시지)을 보내고자 할 때 사용됩니다.

#### `TextInputAction.previous`

이전 입력 소스로 돌아가고자 할 때 사용됩니다.

#### `TextInputAction.continueAction`

사용자가 텍스트 입력을 계속하고자 할 때 사용됩니다.
iOS 전용이며,'계속하기(Continue)' 버튼이 표시됩니다. Android에서는 사용할 수 없습니다.
여러 단계의 폼을 작성하는 경우, 한 단계에서 다음 단계로 넘어가기를 원할 때 사용될 수 있습니다.

#### `TextInputAction.join`

사용자가 어떤 것에 가입하고자 할 때 사용됩니다.
iOS 전용이먀, '가입하기(Join)' 버튼이 표시됩니다. Android에서는 사용할 수 없습니다.
네트워크에 접속하거나, 그룹에 가입하는 등의 상황에서 사용될 수 있습니다.

#### `TextInputAction.route`

사용자가 경로나 방향을 찾고자 할 때 사용됩니다.
iOS 전용이며, '경로(Route)' 버튼이 표시됩니다. Android에서는 사용할 수 없습니다.
운전 방향을 검색하거나, 지도 앱에서 목적지를 찾을 때 사용될 수 있습니다.

#### `TextInputAction.emergencyCall`

사용자가 긴급 전화를 걸고자 할 때 사용됩니다.
iOS 전용이며, '긴급 전화(Emergency Call)' 버튼이 표시됩니다. Android에서는 사용할 수 없습니다.
긴급 상황에서 빠르게 도움을 요청하고자 할 때 사용될 수 있습니다.

#### `TextInputAction.newline`

사용자가 새로운 줄로 이동하고자 할 때 사용됩니다.
Android와 iOS 모두에서 사용 가능. 키보드에 '줄바꿈(New Line)' 또는 '리턴(Return)' 버튼이 표시됩니다.
멀티라인 텍스트 필드에서 새 줄을 시작하고자 할 때 사용됩니다. 예를 들어, 채팅 앱에서 메시지를 작성할 때 줄바꿈을 위해 사용될 수 있습니다.
