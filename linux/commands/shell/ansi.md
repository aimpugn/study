# ANSI

- [ANSI](#ansi)
    - [ANSI?](#ansi-1)
    - [ANSI Escape Sequences](#ansi-escape-sequences)
    - [ANSI Color Codes](#ansi-color-codes)
        - [요약](#요약)
        - [컬러 변경 예제](#컬러-변경-예제)
    - [터미널과 ANSI 상호작용](#터미널과-ansi-상호작용)

## ANSI?

ANSI(미국 국립 표준 협회, American National Standards Institute)는 다양한 산업 분야의 표준을 개발하고 인증하는 기관입니다. `ANSI`가 제정한 여러 표준 중에는 컴퓨터 터미널과 관련된 ANSI X3.64 (또는 ECMA-48) 표준이 포함되어 있으며, 이는 텍스트 형식 지정과 제어에 사용되는 이스케이프 시퀀스를 정의합니다.

- **ANSI**: 미국 국립 표준 협회, 텍스트 터미널 제어를 위한 표준을 포함.
- **ANSI 이스케이프 시퀀스**: 터미널에서 텍스트 색상, 스타일, 커서 위치 등을 제어하는 코드.
- **터미널과의 상호작용**: 터미널 에뮬레이터가 ANSI 이스케이프 시퀀스를 해석하여 지정된 형식으로 텍스트를 표시.

## ANSI Escape Sequences

ANSI 이스케이프 시퀀스는 텍스트 터미널에서 텍스트 색상, 스타일, 커서 위치 등을 제어하기 위해 사용됩니다. 이스케이프 시퀀스는 다음과 같은 형식을 따릅니다:

- `ESC` (Escape, ASCII 코드 27) 문자로 시작합니다. 이는 보통 `\033` 또는 `\e`로 표시됩니다.
- `[`: 이스케이프 시퀀스의 시작을 나타내는 문자입니다.
- 시퀀스 코드: 제어 기능을 지정하는 숫자와 선택적 매개변수.

예를 들어, 빨간색 텍스트를 나타내는 ANSI 이스케이프 시퀀스는 `\033[0;31m`입니다.

## ANSI Color Codes

`bash`에서 텍스트의 색상을 변경하려면 ANSI escape sequences를 사용할 수 있습니다.
`\033`는 ESC 문자를 나타내며, 그 뒤에 오는 시퀀스가 색상과 스타일을 정의합니다.
`\033[0;31m`는 텍스트를 빨간색으로 설정하는 시퀀스입니다.

ANSI 컬러 코드는 8가지 기본 색상을 제공합니다. 이를 확장하여 16색, 256색, 심지어 true color(24비트)까지 사용할 수 있습니다. 기본 8색과 그 확장된 형태는 다음과 같습니다:

- 기본 색상 (8색)
    - 블랙: `\033[0;30m`
    - 레드: `\033[0;31m`
    - 그린: `\033[0;32m`
    - 옐로우: `\033[0;33m`
    - 블루: `\033[0;34m`
    - 퍼플: `\033[0;35m`
    - 시안: `\033[0;36m`
    - 화이트: `\033[0;37m`

- 밝은 색상 (8색)
    - 밝은 블랙 (회색): `\033[1;30m`
    - 밝은 레드: `\033[1;31m`
    - 밝은 그린: `\033[1;32m`
    - 밝은 옐로우: `\033[1;33m`
    - 밝은 블루: `\033[1;34m`
    - 밝은 퍼플: `\033[1;35m`
    - 밝은 시안: `\033[1;36m`
    - 밝은 화이트: `\033[1;37m`

- ANSI Text Attributes: ANSI 이스케이프 시퀀스는 텍스트 색상 외에도 다양한 텍스트 속성을 제어할 수 있습니다
    - **리셋/기본**: `\033[0m`
    - **굵게**: `\033[1m`
    - **밑줄**: `\033[4m`
    - **반전**: `\033[7m`

아래는 ANSI 이스케이프 시퀀스를 사용하여 다양한 텍스트 형식을 출력하는 예제 스크립트입니다:

```bash
#!/bin/bash

# ANSI 색상 코드
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[0;37m'
RESET='\033[0m'

# ANSI 텍스트 속성
BOLD='\033[1m'
UNDERLINE='\033[4m'
REVERSE='\033[7m'

echo -e "${RED}This is red text${RESET}"
echo -e "${GREEN}This is green text${RESET}"
echo -e "${YELLOW}This is yellow text${RESET}"
echo -e "${BLUE}This is blue text${RESET}"
echo -e "${PURPLE}This is purple text${RESET}"
echo -e "${CYAN}This is cyan text${RESET}"
echo -e "${WHITE}This is white text${RESET}"

echo -e "${BOLD}This is bold text${RESET}"
echo -e "${UNDERLINE}This is underlined text${RESET}"
echo -e "${REVERSE}This is reversed text${RESET}"
```

- `-e` 옵션을 사용하여 escape sequences를 해석하고, `RED` 변수로 정의한 빨간색 코드로 텍스트를 출력합니다. 텍스트 출력 후 `NC` 변수를 사용하여 색상을 기본으로 복원합니다.

### 요약

### 컬러 변경 예제

다음은 `echo` 명령어를 사용하여 텍스트 색상을 변경하는 예제입니다:

```bash
#!/bin/bash

# 텍스트 색상 코드
RED='\033[0;31m'
NC='\033[0m' # No Color (기본 색상 복원)

# 컬러 텍스트 출력
echo -e "${RED}This is red text${NC}"
```

## 터미널과 ANSI 상호작용

터미널 에뮬레이터는 ANSI 이스케이프 시퀀스를 해석하여 텍스트 형식을 지정합니다. 터미널이 이스케이프 시퀀스를 인식하면, 이를 해석하고 지정된 형식으로 텍스트를 표시합니다. 예를 들어, 터미널이 `\033[0;31m`을 받으면, 그 뒤에 오는 텍스트를 빨간색으로 표시합니다.
