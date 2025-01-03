# environments

## `IFS`

`IFS`는 Internal Field Separator의 약자로, 쉘에서 문자열을 분리할 때 사용되는 구분자를 결정하는 환경 변수입니다.
기본값은 공백, 탭, 줄바꿈을 포함하는 세 문자열입니다.
즉, 기본적으로 쉘은 공백, 탭, 줄바꿈을 기준으로 문자열을 분리합니다.

```bash
❯ declare -p IFS
typeset IFS=$' \t\n\C-@'
```

`IFS=' '`를 선언하는 이유는 문자열을 공백 기준으로 분리하고자 할 때입니다.
이 경우, `IFS` 값을 공백(' ')으로 설정함으로써 문자열을 공백 기준으로 분리할 수 있습니다.
이렇게 하면, 문자열을 공백으로 구분된 여러 부분을 각각의 요소로 취급할 수 있습니다.

`IFS`를 원래대로 돌리지 않으면, 이후의 스크립트에서 문자열 분리가 예상치 못한 방식으로 이루어질 수 있습니다.

예를 들어, `IFS`를 공백으로 설정한 후 다른 부분에서 공백이 아닌 다른 문자(예: 콜론)를 기준으로 문자열을 분리하려고 하면, 예상치 못한 결과를 얻게 받을 수 있습니다.

`IFS`를 원래대로 돌리려면, 원래 값으로 복원해야 합니다.
이를 위해 `IFS`의 원래 값을 다른 변수에 저장한 후, 작업이 끝난 후에 `IFS`를 원래 값으로 복원할 수 있습니다.

예를 들어, 다음과 같이 할 수 있습니다:

```bash
old_ifs="$IFS"
IFS=' '
# 작업 수행
IFS="$old_ifs"
```

또는, `IFS`를 초기 상태로 되돌리려면 `unset` 명령어를 사용할 수도 있습니다. 하지만, `IFS`를 빈 문자열("")로 설정한 경우와 `unset`한 경우의 동작 방식이 다르므로 주의해야 합니다. `unset`한 경우, `IFS`는 기본값(공백, 탭, 줄바꿈)을 가진 것으로 처리됩니다. 반면, 빈 문자열로 설정한 경우에는 어떤 분리 작업도 수행되지 않습니다 [1].
