# bash

## `-c`

```bash
# 명령어 및 해당 명령어의 옵션을 문자열로 받아서 실행
# - https://kldp.org/node/165498
# - https://stackoverflow.com/questions/20858381/what-does-bash-c-do
/bin/bash -c "echo \$1" test1 test2
```

## `if`

Bash 스크립트에서 `if` 문은 주어진 명령어의 *종료 상태(exit status)를 평가*하여 조건 분기를 수행한다.
종료 상태는 명령어가 성공적으로 실행되었는지 실패했는지를 나타내는 값으로, 일반적으로 `0`은 성공을, `0`이 아닌 다른 값은 오류나 실패를 의미한다.

```bash
if (mockery); then # `mockery` 명령어를 실행하고 그 종료 상태를 검사
    echo "test";
else
    echo "not passed"
fi
```

- Exit 0 (성공): `mockery` 명령어가 성공적으로 실행되어 종료 상태가 `0`이면, `if` 문 바로 다음에 오는 `then` 절이 실행
- Exit 0 아닌 경우 (실패): `mockery` 명령어 실행 후 종료 상태가 `0`이 아니면, `else` 절이 실행된다.
