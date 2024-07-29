# shell checker

- [shell checker](#shell-checker)
    - [SC2145](#sc2145)
        - [의도하지 않은 결과](#의도하지-않은-결과)
        - [해결 방법](#해결-방법)

## SC2145

> Argument mixes string and array. Use * or separate argument

`Argument mixes string and array. Use * or separate argument` 경고는 문자열과 배열을 혼합하여 사용하는 경우 발생합니다.
이는 `$@`와 같은 *배열 확장이 문자열과 결합될 때 발생*할 수 있습니다.
이 경고는 배열 확장이 문자열의 일부로 처리되어 의도하지 않은 결과를 초래할 수 있기 때문에 발생합니다.

### 의도하지 않은 결과

`$@`는 각 인자를 개별적으로 확장합니다. 즉, 각 인자가 따로따로 처리됩니다.
공백이나 특수 문자가 포함된 인자를 올바르게 처리하기 위해서는 따옴표로 감싸서 사용해야 합니다.

예를 들어, 다음 스크립트를 보겠습니다:

```bash
#!/bin/bash
echo "Arguments: $@"
```

이 스크립트를 `./script.sh "arg1 with spaces" arg2`와 같이 실행하면, `echo` 명령어는 다음과 같이 인자들을 출력합니다:

```plaintext
Arguments: arg1 with spaces arg2
```

여기서 `arg1 with spaces`가 세 개의 개별 인자로 `arg1`, `with`, `spaces`로 분리되어 출력됩니다. 이는 원래 의도와 다르게 출력되는 것입니다.

따라서, `$@`를 사용할 때 공백이나 특수 문자가 포함된 인자를 올바르게 처리하기 위해서는 다음과 같이 따옴표로 감싸야 합니다:

```bash
#!/bin/bash
echo "Arguments: $@"
```

이 스크립트를 `./script.sh "arg1 with spaces" arg2`와 같이 실행하면, `echo` 명령어는 다음과 같이 인자들을 올바르게 출력합니다:

```plaintext
Arguments: arg1 with spaces arg2
```

`$*`는 모든 인자를 하나의 문자열로 확장합니다. 즉, 모든 인자가 공백으로 구분된 하나의 문자열로 합쳐집니다.
공백이나 특수 문자가 포함된 인자도 하나의 인자로 처리됩니다.

예를 들어, 다음 스크립트를 보겠습니다:

```bash
#!/bin/bash
echo "Arguments: $*"
```

이 스크립트를 `./script.sh "arg1 with spaces" arg2`와 같이 실행하면, `echo` 명령어는 다음과 같이 인자들을 출력합니다:

```plaintext
Arguments: arg1 with spaces arg2
```

이 경우, `arg1 with spaces`가 하나의 인자로 처리되어 올바르게 출력됩니다.

- `$@`를 따옴표 없이 사용한 경우:

    ```bash
    #!/bin/bash
    echo "Arguments: $@"
    ```

    실행: `./script.sh "arg1 with spaces" arg2`

    출력:

    ```plaintext
    Arguments: arg1 with spaces arg2
    ```

    `arg1 with spaces`가 `arg1`, `with`, `spaces`로 분리됨.

- `$*`를 사용한 경우:

    ```bash
    #!/bin/bash
    echo "Arguments: $*"
    ```

    실행: `./script.sh "arg1 with spaces" arg2`

    출력:

    ```plaintext
    Arguments: arg1 with spaces arg2
    ```

    `arg1 with spaces`가 하나의 인자로 처리됨.

- `$@`를 따옴표로 감싼 경우:

    ```bash
    #!/bin/bash
    echo "Arguments: $@"
    ```

    실행: `./script.sh "arg1 with spaces" arg2`

    출력:

    ```plaintext
    Arguments: arg1 with spaces arg2
    ```

    `arg1 with spaces`가 하나의 인자로 처리됨.

### 해결 방법

이 문제를 해결하기 위해서는 배열 확장을 별도의 인자로 처리하거나, 배열을 하나의 문자열로 결합하여 사용해야 합니다.
다음은 이를 해결하는 몇 가지 방법입니다:

1. **배열 확장을 별도의 인자로 처리**:

   ```bash
   echo "Arguments:" "$@"
   ```

2. **배열을 하나의 문자열로 결합하여 사용**:

   ```bash
   echo "Arguments: $*"
   ```

3. **배열을 따옴표로 감싸서 사용**:

   ```bash
   echo "Arguments: ${@}"
   ```

따라서, 원래 질문에서 `echo "./run protoc $*"` 대신 `echo "./run protoc $@"`를 사용하면 경고가 발생할 수 있습니다. 이를 해결하기 위해서는 다음과 같이 수정할 수 있습니다:

```bash
#!/bin/bash
echo "./run protoc $*"
exit 0
```

또는, 배열 확장을 별도의 인자로 처리하여 경고를 피할 수 있습니다:

```bash
#!/bin/bash
echo "./run protoc" "$@"
exit 0
```

이렇게 하면 `./run protoc test value` 명령어를 실행했을 때, 올바르게 `./run protoc test value`가 출력되며, 경고도 발생하지 않습니다.

배열과 문자열을 혼합하여 사용할 때는 별도로 인자를 처리하는 것이 좋습니다. 예를 들어:

```bash
#!/bin/bash

args=("arg1" "arg2 with spaces" "arg3")
command_to_run=("echo" "Running command with arguments:")
"${command_to_run[@]}"
for arg in "${args[@]}"; do
    echo "$arg"
done
```

이렇게 하면 각 인자가 개별적으로 처리되어 예상한 대로 동작하게 됩니다.
