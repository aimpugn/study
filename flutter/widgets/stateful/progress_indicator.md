# progress indicator

## progress indicator와 `AppBar`

- `AppBar`에 `LinearProgressIndicator`를 사용하는 것은 Flutter에서 일반적으로 권장되는 방법은 아니지만, `AppBar`에 `Stepper`를 직접 배치하는 것과는 다르다
    - `AppBar`에 `LinearProgressIndicator`를 배치하는 경우
        - 사용자에게 현재 앱의 작업 진행 상태를 간결하고 명확하게 보여준다.
        - 예를 들어, 긴 로딩 시간 또는 다단계 프로세스의 진행 상황을 나타내는 데 적합하다.
    - 반면에, `AppBar`에 `Stepper`를 배치하는 경우
        - UI 디자인 측면에서 일반적이지 않으며, 사용자 경험을 해칠 수 있다.
        - `Stepper`는 일반적으로 화면의 본문 부분에 배치되며, 사용자가 각 단계를 쉽게 인식하고 조작할 수 있도록 해야 한다

## `LinearProgressIndicator`
