# Wrapping Class

- [Wrapping Class](#wrapping-class)
    - [Wrapping class 생성?](#wrapping-class-생성)
    - [일반 클래스 vs StatelessWidget/StatefulWidget](#일반-클래스-vs-statelesswidgetstatefulwidget)
    - [결정을 내릴 때 고려해야 할 요소](#결정을-내릴-때-고려해야-할-요소)
    - [예시](#예시)
        - [`TextFormField`를 wrapping 하는 클래스 만들기](#textformfield를-wrapping-하는-클래스-만들기)
            - [`StatelessWidget`으로 리팩토링](#statelesswidget으로-리팩토링)
            - [사용 예시](#사용-예시)
        - [단순 데이터 클래스인 `InputDecoration` 클래스 경우?](#단순-데이터-클래스인-inputdecoration-클래스-경우)
            - [비밀번호 필드의 표시/숨김 기능을 포함하는 `InputDecoration` 생성 함수](#비밀번호-필드의-표시숨김-기능을-포함하는-inputdecoration-생성-함수)
            - [사용 예시](#사용-예시-1)

## Wrapping class 생성?

- Flutter에서 위젯을 래핑하는 사용자 정의 클래스를 만들 때 `StatelessWidget`이나 `StatefulWidget`을 상속받는 것이 일반적인 방법.
- 하지만 특정 경우에는 단순한 Dart 클래스를 사용하여 위젯을 생성하는 "헬퍼" 또는 "빌더" 클래스를 만들 수도 있다
    - 이런 접근 방식은 특히 위젯 구성이 **상태를 관리할 필요가 없고 단순히 프로퍼티를 전달하여 위젯을 생성하는 경우에 적합**합니다.
- 결국, 사용자 정의 위젯 클래스를 구현할 때는 다음 사항들을 고려하여 가장 적합한 방식을 선택해야 한다
    - 위젯의 사용 목적
    - 상태 관리의 필요성
    - 재사용성,
    - 그리고 구현의 복잡성 등
- 단순한 위젯 래퍼의 경우 일반 클래스로 충분할 수 있으며, 보다 복잡한 동적 위젯의 경우 `StatelessWidget` 또는 `StatefulWidget`을 상속받는 것이 적합하다

## 일반 클래스 vs StatelessWidget/StatefulWidget

1. **일반 클래스 사용:**
    - 상태 관리가 필요 없고, 단순히 프로퍼티를 기반으로 위젯을 생성하는 경우.
    - 코드를 단순화하고 재사용하기 쉬운 작은 빌더 또는 헬퍼 클래스를 만들 때 적합하다
    - 사용례:
        - 주어진 프로퍼티를 기반으로 `TextFormField` 위젯을 생성하는 경우
2. **StatelessWidget/StatefulWidget 상속:**
    - 위젯이 자체적인 상태를 가지거나, 생명주기 메서드(`initState`, `dispose` 등)에 접근해야 하는 경우.
    - Flutter **위젯 트리에 직접 포함되는 복잡한 위젯**을 구현할 때 적합하다
    - 사용례: 동적인 상호작용을 포함하는 위젯
        - **사용자 입력을 처리**
        - 애니메이션과 등

## 결정을 내릴 때 고려해야 할 요소

1. **상태 관리의 필요성:**
    - 위젯이 내부 상태를 관리해야 하거나, 사용자 상호작용에 응답해야 하는지 여부를 고려
2. **재사용성:**
    - 클래스를 재사용하기 쉽게 만들고자 하는 목적이 있는지, 그리고 재사용성을 높이기 위해 어떤 방식이 적합한지 고려
3. **복잡성:**
    - 구현하려는 위젯의 복잡성에 따라 상속받을 클래스를 결정한다
    - 단순한 위젯 구성은 일반 클래스로 충분할 수 있으며, 복잡한 위젯은 `StatelessWidget` 또는 `StatefulWidget`을 사용하는 것이 좋다
4. **Flutter 위젯 트리와의 통합:**
    - 위젯이 Flutter 위젯 트리의 일부로서 작동해야 하는지, 아니면 단순히 위젯 생성을 위한 도구로만 사용되는지 고려

## 예시

### `TextFormField`를 wrapping 하는 클래스 만들기

```dart
class GrayFilledTextFormField {
  final String labelName; // 필드 선언
  final Key textFormFieldKey;
  final TextEditingController controller;
  final TextStyle? labelTextStyle;
  final List<TextInputFormatter>? inputFormatters;
  final Widget? suffix;

  GrayFilledTextFormField({
    required this.labelName,
    required this.textFormFieldKey,
    required this.controller,
    this.labelTextStyle,
    this.inputFormatters,
    this.suffix,
  });

  Widget of() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          labelName, // 라벨 텍스트
          style: labelTextStyle,
        ),
        gap12,
        Row(
          children: [
            Expanded(
              child: TextFormField(
                key: SignUpScreen.name,
                controller: controller,
                autovalidateMode: AutovalidateMode.onUserInteraction,
                autocorrect: false,
                textInputAction: TextInputAction.next,
                keyboardType: TextInputType.emailAddress,
                keyboardAppearance: Brightness.light,
                inputFormatters: inputFormatters,
                decoration: InputDecoration(
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(5.0),
                    borderSide: BorderSide.none,
                  ),
                  fillColor: inputGrayColor,
                  filled: true,
                  suffix: suffix,
                ),
              ),
            )
          ],
        )
      ],
    );
  }
}
```

- 제시된 `GrayFilledTextFormField` 클래스는 사용자 입력을 받는 `TextFormField`를 재사용하기 위한 목적으로 설계된 클래스
- 이 클래스는 상태를 직접 관리하지 않고, 외부에서 전달받은 데이터 (`labelName`, `controller`, `textFormFieldKey` 등)를 기반으로 위젯을 구 성
- 이러한 특성을 감안할 때, 이 클래스는 다음과 같이 분류할 수 있다
    - **상태가 없는 (Stateless) 위젯:**
        - 클래스 자체가 내부 상태를 관리하거나 변경하지 않기 때문에, 이는 상태가 없는 위젯의 특성을 갖는다
        - 주어진 매개변수를 바탕으로 위젯을 생성하는 역할을 하며, **내부 상태의 변화를 추적하거나 업데이트하지 않는다**
    - **일반 클래스로의 구현:**
        - `GrayFilledTextFormField`는 Flutter의 `StatelessWidget`을 상속받지 않고 일반 Dart 클래스로 구현되어 있다
        - 이는 클래스가 단순히 `TextFormField` 위젯을 생성하는 데 사용되며, 별도의 UI 또는 상태 관련 기능이 포함되지 않았음을 의미
    - **적합성:**
        - 이 경우, `StatelessWidget`이나 `StatefulWidget`의 상속 없이도 충분히 기능적이다
        - 클래스의 목적이 단순히 특정 위젯을 구성하고 반환하는 것이라면, 일반 Dart 클래스로 구현하는 것이 더 간결하고 목적에 부합하다

#### `StatelessWidget`으로 리팩토링

- `StatelessWidget`을 상속하면 **Flutter 위젯 트리에 쉽게 통합**할 수 있으며, 더 일관된 Flutter 위젯 작성 방식을 따를 수 있다
- **생성자에 `Key? key` 추가:** 모든 `StatelessWidget`은 선택적으로 `Key`를 받을 수 있습니다. 이는 위젯을 고유하게 식별하는 데 도움이 된다
- **`build` 메서드 구현:**
    - 모든 `StatelessWidget`은 `build` 메서드를 구현해야 한다
    - 이 메서드는 위젯이 화면에 어떻게 그려질지를 정의한다

```dart
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class GrayFilledTextFormField extends StatelessWidget {
  final String labelName;
  final Key textFormFieldKey;
  final TextEditingController controller;
  final TextStyle? labelTextStyle;
  final List<TextInputFormatter>? inputFormatters;
  final Widget? suffix;

  GrayFilledTextFormField({
    required this.labelName,
    required this.textFormFieldKey,
    required this.controller,
    this.labelTextStyle,
    this.inputFormatters,
    this.suffix,
    Key? key,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          labelName,
          style: labelTextStyle,
        ),
        SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: TextFormField(
                key: textFormFieldKey, // 수정: 각 인스턴스에 고유한 key 적용
                controller: controller,
                autovalidateMode: AutovalidateMode.onUserInteraction,
                autocorrect: false,
                textInputAction: TextInputAction.next,
                keyboardType: TextInputType.emailAddress,
                keyboardAppearance: Brightness.light,
                inputFormatters: inputFormatters,
                decoration: InputDecoration(
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(5.0),
                    borderSide: BorderSide.none,
                  ),
                  fillColor: Colors.grey[200], // 예시: 회색 배경 색상
                  filled: true,
                  suffix: suffix,
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }
}
```

#### 사용 예시

- Flutter에서 `StatelessWidget` 또는 `StatefulWidget`을 사용하여 커스텀 위젯을 만들면, **`build` 메서드는 Flutter 프레임워크에 의해 자동으로 호출**된다
- 이 **`build` 메서드는 위젯이 화면에 그려질 때마다 호출**되며, 위젯의 레이아웃과 UI를 정의한다
- `GrayFilledTextFormField` 위젯을 사용하여 기존의 `TextFormField` 부분을 대체하려면, 다음과 같이 `GrayFilledTextFormField`의 인스턴스를 생성하고 필요한 매개변수를 전달해야 한다

```dart
List<Widget> _fieldsAfterPhoneVerified(AppLocalizations l10n) {
  return [
    GrayFilledTextFormField(
      labelName: l10n.labelNickname,
      textFormFieldKey: SignUpScreen.name, // 이 부분은 고유한 Key를 제공해야 합니다.
      controller: _nameController,
      labelTextStyle: smallTextStyle,
      inputFormatters: <TextInputFormatter>[defaultNoSpaceTextInputValidator],
      suffix: InkWell(
        onTap: () {
          // 중복 확인 로직 구현
        },
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8.0),
          child: Text(
            l10n.labelCheckDuplicate,
            style: smallTextStyle,
          ),
        ),
      ),
    ),
    gap8,
    // Email 필드는 GrayFilledTextFormField를 사용하여 대체
    GrayFilledTextFormField(
      labelName: l10n.labelEmail,
      textFormFieldKey: SignUpScreen.emailKey,
      controller: _emailController,
      // 다른 속성들 추가
    ),
    gap8,
    // Password 필드도 GrayFilledTextFormField를 사용하여 대체
    GrayFilledTextFormField(
      labelName: l10n.labelPassword,
      textFormFieldKey: SignUpScreen.passwordKey,
      controller: _passwordController,
      // 다른 속성들 추가
    ),
    // 나머지 위젯들...
  ];
}
```

### 단순 데이터 클래스인 `InputDecoration` 클래스 경우?

- Flutter에서 `InputDecoration`을 재사용하는 것은 매우 일반적인 상황이라고 한다
- 특히 비밀번호 입력 필드와 같이 공통적인 스타일과 기능을 여러 곳에서 사용할 경우, `InputDecoration`을 별도의 함수나 클래스로 래핑하여 재사용성을 높이는 것이 좋다

```dart
InputDecoration(
    labelText: l10n.labelPassword,
    enabled: !state.isLoading,
    helperText: state.value is SignInError
        ? l10n.loginFailedMessage
        : null,
    helperStyle: const TextStyle(color: primaryBackgroundColor),
    suffixIcon: IconButton(
        icon: Icon(
            // Based on passwordVisible state choose the icon
            _passwordVisible
                ? Icons.visibility
                : Icons.visibility_off,
            color: Theme.of(context).primaryColorDark,
        ),
        onPressed: () {
            setState(() => _passwordVisible = !_passwordVisible);
        },
    ),
),
```

#### 비밀번호 필드의 표시/숨김 기능을 포함하는 `InputDecoration` 생성 함수

- `InputDecoration`은 `StatelessWidget`이나 `StatefulWidget`과는 독립적인, 단순한 데이터 클래스
- 따라서, `InputDecoration`을 생성하는 함수나 클래스를 만들어 관리할 수 있다.

```dart
InputDecoration buildPasswordInputDecoration({
  required String labelText,
  required String helperText,
  required bool isPasswordVisible,
  required VoidCallback onTogglePasswordVisibility,
}) {
  return InputDecoration(
    labelText: labelText,
    helperText: helperText,
    suffixIcon: IconButton(
      icon: Icon(
        isPasswordVisible ? Icons.visibility : Icons.visibility_off,
      ),
      onPressed: onTogglePasswordVisibility,
    ),
    // 기타 필요한 스타일링...
  );
}
```

- 이 함수는 비밀번호 필드의 `InputDecoration`을 생성하며, 표시/숨김 상태(`isPasswordVisible`)와 상태 변경을 위한 콜백(`onTogglePasswordVisibility`)을 매개변수로 받는다

#### 사용 예시

```dart
TextFormField(
  obscureText: !_passwordVisible,
  decoration: buildPasswordInputDecoration(
    labelText: 'Password',
    helperText: 'Enter your password',
    isPasswordVisible: _passwordVisible,
    onTogglePasswordVisibility: () {
      setState(() {
        _passwordVisible = !_passwordVisible;
      });
    },
  ),
)
```

이 방법은 코드의 재사용성을 높이고, 일관된 UI 스타일을 유지하는 데 도움이 됩니다. 필요에 따라 `buildPasswordInputDecoration` 함수를 더 많은 매개변수로 확장하여 다양한 사용 사례에 맞게 조정할 수도 있습니다.
