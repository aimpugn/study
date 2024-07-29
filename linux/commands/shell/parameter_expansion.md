# Parameter Expansion

- [Parameter Expansion](#parameter-expansion)
    - [Parameter Expansion](#parameter-expansion-1)
    - [셸 파라미터 확장: `${parameter/pattern/string}`](#셸-파라미터-확장-parameterpatternstring)
    - [Pattern Substitution](#pattern-substitution)
        - [`${1// /}`](#1-)
        - [`${1// }`](#1--1)

## Parameter Expansion

셸 파라미터 확장은 Unix 계열 운영 체제에서 변수의 값을 직접 조작할 수 있는 강력한 기능입니다.
이를 통해 문자열 추출, 패턴 매칭 및 문자열 치환 등의 다양한 작업을 수행할 수 있습니다.

- **Parameter(파라미터)**: 수정할 문자열을 포함하는 셸 변수 또는 위치 매개 변수.
- **Pattern(패턴)**: 매칭 기준을 지정하는 문자열. 패턴은 정규 표현식과 유사하지만 다른 구문을 사용합니다.
- **String(문자열)**: `Pattern`과 일치하는 부분을 대체할 문자열.

파라미터 확장을 사용하면 다음과 같은 장점이 있습니다:
- **코드 가독성**: 파라미터 확장을 사용하면 스크립트가 짧고 읽기 쉬워집니다.
- **성능**: 간단한 문자열 조작에는 외부 명령어(`sed`나 `awk`)를 호출하는 것보다 빠릅니다.
- **다양성**: 이 기능은 패턴 기반 치환을 포함한 복잡한 조작을 지원하여 매우 다양하게 사용할 수 있습니다.

## 셸 파라미터 확장: `${parameter/pattern/string}`

셸 파라미터 확장은 Unix 계열 운영 체제에서 변수의 값을 직접 조작할 수 있는 강력한 기능입니다.
이를 통해 문자열 추출, 패턴 매칭 및 문자열 치환 등의 다양한 작업을 수행할 수 있습니다.

## Pattern Substitution

이 구문은 매개변수 값에서 패턴이 처음 나오는 부분을 문자열로 대체하는 데 사용됩니다.

```sh
${parameter/pattern/string}
${parameter//pattern/string}
${parameter/#pattern/string}
${parameter/%pattern/string}
```

• `/${pattern}/`: 패턴이 처음 나오는 부분을 문자열로 대체
• `//pattern/`: 패턴이 나오는 모든 부분을 문자열로 대체
• `/#pattern/`: 매개변수 값의 시작 부분이 패턴과 매친되면 해당 부분을 대체
• `/%pattern/`: 매개변수 값의 끝 부분이 패턴과 매친되면 해당 부분을 대체

### `${1// /}`

이 표현식은 파라미터 `$1`에 저장된 문자열에서 모든 공백을 제거합니다.

`/ /`(하나 이상의 공백) 패턴을 사용하고 빈 문자열(`/}`)로 대체합니다,
첫 번째 위치 매개변수(`$1`)의 값에서 모든 공백을 효과적으로 제거합니다.

- **`1`**: 첫 번째 위치 매개 변수를 나타내며, 문자열 값을 저장합니다.
- **`//`**: 패턴의 모든 발생을 교체함을 나타냅니다(글로벌 교체).
- **`` (공백)**: 매칭할 패턴(단일 공백 문자).
- **`(빈 문자열)`**: 대체 문자열(아무것도 없음, 공백을 사실상 제거).

```sh
#!/bin/bash
input="A quick brown fox jumps over the lazy dog"
# 모든 공백 제거
output="${input// /}"
echo "Original: $input"
echo "Without spaces: $output"

# Original: A quick brown fox jumps over the lazy dog
# Without spaces: Aquickbrownfoxjumpsoverthelazydog
```

```sh
TEST="    A quick brown fox   "
result="${TEST// /}"
echo "##$result##"
##Aquickbrownfox##
```

### `${1// }`

string을 지정하지 않은 경우입니다.
이 경우 pattern에 해당하는 모든 부분을 빈 문자열로 치환합니다.
즉, pattern에 해당하는 부분을 삭제하는 효과가 있습니다.

따라서 파라미터 `$1`에 저장된 문자열에서 모든 공백을 제거합니다.

```sh
❯ TEST="    A quick brown fox   "
result="${TEST// }"
echo "##$result##"
##Aquickbrownfox##
```
