# set

- [set](#set)
    - [set 명령어](#set-명령어)
    - [`set -e`](#set--e)
    - [`set +e`](#set-e)
    - [`set -u`](#set--u)
    - [`set +u`](#set-u)
    - [`set -x`](#set--x)
    - [`set +x`](#set-x)
    - [`set -o pipefail`](#set--o-pipefail)
    - [`set +o pipefail`](#set-o-pipefail)
    - [예제 스크립트](#예제-스크립트)

## set 명령어

`set` 명령어는 셸 스크립트에서 셸의 동작을 제어하는 데 사용되는 중요한 명령어입니다.
이 명령어는 스크립트의 실행 환경을 설정하고, 오류 처리, 디버깅, 변수 사용 등의 다양한 측면에서 셸의 동작을 세밀하게 조정할 수 있게 해줍니다.
이를 통해 스크립트의 안정성과 가독성을 높일 수 있습니다.

## `set -e`

- **설명**: 명령어가 실패하면 스크립트를 종료합니다.
- **사용 예시**:

  ```bash
  set -e
  command1
  command2  # command1이 실패하면 이 명령어는 실행되지 않습니다.
  ```

```bash
# set -e와 set +e 사용
set +e
SRC=$(echo "$SRC" | grep -s -E "$proto_pattern")
set -e
```

## `set +e`

- **설명**: `set -e` 옵션을 비활성화합니다. 명령어가 실패해도 스크립트가 계속 실행됩니다.
- **사용 예시**:

  ```bash
  set +e
  command1
  command2  # command1이 실패해도 이 명령어는 실행됩니다.
  ```

## `set -u`

- **설명**: 선언되지 않은 변수를 사용하면 스크립트를 종료합니다.
- **사용 예시**:

  ```bash
  set -u
  echo $UNDEFINED_VAR  # UNDEFINED_VAR가 선언되지 않았으므로 스크립트가 종료됩니다.
  ```

## `set +u`

- **설명**: `set -u` 옵션을 비활성화합니다. 선언되지 않은 변수를 사용해도 스크립트가 계속 실행됩니다.
- **사용 예시**:

  ```bash
  set +u
  echo $UNDEFINED_VAR  # UNDEFINED_VAR가 선언되지 않았지만 스크립트가 계속 실행됩니다.
  ```

## `set -x`

- **설명**: 실행되는 각 명령어를 실행하기 전에 출력합니다. 디버깅에 유용합니다.
- **사용 예시**:

  ```bash
  set -x
  command1
  command2  # 각 명령어가 실행되기 전에 명령어가 출력됩니다.
  ```

## `set +x`

- **설명**: `set -x` 옵션을 비활성화합니다. 명령어가 실행되기 전에 출력되지 않습니다.
- **사용 예시**:

  ```bash
  set +x
  command1
  command2  # 명령어가 실행되기 전에 출력되지 않습니다.
  ```

## `set -o pipefail`

- **설명**: 파이프라인의 어느 명령어라도 실패하면 전체 파이프라인이 실패로 간주됩니다. 파이프라인의 중간 명령어가 실패했을 때도 이를 감지하고 처리할 수 있게 해줍니다.
- **사용 예시**:

  ```bash
  set -o pipefail
  command1 | command2  # command1 또는 command2가 실패하면 전체 파이프라인이 실패로 간주됩니다.
  ```

## `set +o pipefail`

- **설명**: `set -o pipefail` 옵션을 비활성화합니다. 파이프라인의 마지막 명령어의 종료 상태만 고려됩니다.
- **사용 예시**:

  ```bash
  set +o pipefail
  command1 | command2  # command2의 종료 상태만 고려됩니다.
  ```

## 예제 스크립트

다음은 `set` 명령어와 다양한 옵션들을 사용하는 예제 스크립트입니다:

```bash
#!/bin/bash

set -e  # 명령어가 실패하면 스크립트를 종료합니다.
set -u  # 선언되지 않은 변수를 사용하면 스크립트를 종료합니다.
set -x  # 실행되는 각 명령어를 실행하기 전에 출력합니다.
set -o pipefail  # 파이프라인의 어느 명령어라도 실패하면 전체 파이프라인이 실패로 간주됩니다.

echo "This is a test script."

# 선언되지 않은 변수 사용 (스크립트가 종료됩니다)
echo $UNDEFINED_VAR

# 명령어 실패 (스크립트가 종료됩니다)
false

# 파이프라인 실패 (스크립트가 종료됩니다)
false | true
```
