# read

## read

## read 동작 과정

1. 사용자 프롬프트:

    스크립트나 프로그램이 실행될 때 사용자로부터 정보를 얻어야 할 수 있습니다.
    일반적으로 `read` 명령은 어떤 종류의 입력이 필요한지 알려주는 프롬프트 메시지와 함께 사용됩니다.

    예를 들어

    ```sh
    read -p "Please enter your name: " name
    ```

    - `-p`: 입력을 읽기 전에 프롬프트 메시지를 표시하는 데 사용됩니다.

2. 입력 대기

    프롬프트 메시지를 출력한 후, `read` 명령어는 사용자가 어떤 입력하고 Enter 키를 누를 때까지 대기합니다.

3. 입력 읽기

    사용자가 Enter 키를 입력하면, `read` 명령어는 사용자가 제공한 입력을 캡처합니다.

4. 입력 저장

    `read` 명령어는 사용자의 입력을 지정된 변수에 저장합니다.
    가령 1번 예제에서 `name` 변수에 저장됩니다.

5. 단어 분할:

    기본적으로 읽기 명령은 입력에 대해 단어 분할을 수행합니다.
    즉, 사용자가 공백으로 구별되는 여러 단어를 입력하는 경우, 각 단어는 별개의 필드로 처리되며
    서로 다른 변수에 저장됩니다.

    ```sh
    ❯ read first second
    John Doe

    ❯ echo "first:$first, second:$second"
    first:John, second:Doe
    ```

내부적으로 `read` 명령은 표준 입력 스트림(`STDIN`)에서 입력을 읽는 방식으로 작동합니다.
터미널에서 스크립트나 프로그램을 실행하면 *터미널 자체가 표준 입력 스트림으로 작동*하여 사용자가 직접 입력을 입력할 수 있습니다.

```sh
# !/bin/zsh

# Ask for the user's name and store it in the 'name' variable

echo "Please enter your name:"
read name

# Print a personalized greeting using the stored name
echo "Hello, $name!"
```

`read` 명령어는 입력 리디렉션을 통해 파일이나 파이프 같은 다른 소스로부터 입력을 읽을 수 있습니다.

test
