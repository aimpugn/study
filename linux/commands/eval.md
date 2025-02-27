# eval

- [eval](#eval)
    - [`eval`이란?](#eval이란)
    - [왜 존재하나?](#왜-존재하나)
    - [어떤 경우에 사용되나?](#어떤-경우에-사용되나)
    - [어떤 원리로 어떻게 동작하는가?](#어떤-원리로-어떻게-동작하는가)
    - [사용 예시](#사용-예시)
    - [주의사항](#주의사항)

## `eval`이란?

`eval`은 쉘 스크립트나 커맨드 라인에서 매우 강력한 명령어로, *문자열을 쉘 명령어로 평가하고 실행*하는 데 사용된다.
이 명령어는 주어진 문자열을 마치 사용자가 직접 입력한 것처럼 처리하며, 이를 통해 동적으로 명령어를 생성하고 실행할 수 있다.

## 왜 존재하나?

eval 명령어는 동적으로 생성된 셸 명령을 실행할 수 있게 해줍니다.
예를 들어, 사용자 입력이나 스크립트 내 다른 연산에 의해 형성된 명령을 실행하는 경우에 사용됩니다.
이를 통해 스크립트는 더 유연하고 동적으로 작동할 수 있습니다.

## 어떤 경우에 사용되나?

1. 변수에 저장된 명령어를 실행할 때
2. 복잡한 문자열 조작 후 실행 가능한 명령어 생성 시
3. 조건에 따라 다양한 명령어를 실행할 때
4. 환경 변수와 같이 동적으로 변하는 값에 의존하는 스크립트 작성 시 동적으로 명령어를 생성할 수 있다. 변수의 값에 따라 다른 명령어를 실행하고자 할 때 특히 유용하다.
5. 변수 안의 변수 사용. 변수 이름을 저장하고 있는 변수를 평가하여 실제 값을 얻고 싶을 때 사용할 수 있습니다.

    ```bash
    varname="PATH"
    eval echo \$$varname # $PATH 환경 변수의 값을 출력
    ```

6. 함수 정의 전달: 쉘 함수의 정의를 문자열로 저장하고, 이를 다른 쉘 세션에서 `eval`을 사용하여 재정의하고 실행할 수 있다. 이는 복잡한 환경에서 함수를 전달하고 재사용하는 데 유용하다.

## 어떤 원리로 어떻게 동작하는가?

eval 명령은 **주어진 문자열을 셸의 명령어로 해석하여 실행**합니다.
문자열 내에 셸 변수, 파이프, 리다이렉션 등 셸이 해석할 수 있는 모든 요소를 포함할 수 있으며,
eval은 이를 셸이 직접적으로 실행할 수 있는 명령어로 평가하여 실행합니다.

## 사용 예시

```bash
# 동적 변수 이름 생성 및 사용
varname="value"
eval echo \$$varname  value 변수의 값을 출력
```

```bash
# 복잡한 명령어 생성 및 실행
command_part1="ls"
command_part2="-l"
eval $command_part1 $command_part2  'ls -l' 명령 실행
```

## 주의사항

eval을 사용할 때는 코드 인젝션 등의 보안 이슈가 발생할 수 있으므로 사용에 주의가 필요합니다.
