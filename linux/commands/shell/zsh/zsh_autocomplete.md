
# zsh-autocomplete

- [zsh-autocomplete](#zsh-autocomplete)
    - [zsh-autocomplete](#zsh-autocomplete-1)
    - [Configuration](#configuration)
        - [설정 예제](#설정-예제)

## [zsh-autocomplete](https://github.com/marlonrichert/zsh-autocomplete)

zsh-autocomplete는 데스크톱 앱에서 볼 수 있는 것과 유사하게 명령줄에 실시간 입력 자동 완성 기능을 추가하는 플러그인입니다.

현재 입력된 명령어와 위치를 기반하여 적절한 자동완성 옵션을 결정합니다.

## [Configuration](https://github.com/marlonrichert/zsh-autocomplete?tab=readme-ov-file#configuration)

1. 스타일 설정:

    ```zsh
    # 자동완성 메뉴를 선택 가능한 형태로 만들기
    zstyle ':completion:*' menu select
    # 파일 유형에 따라 색상을 적용
    zstyle ':completion:*' list-colors ${(s.:.)LS_COLORS}
    ```

2. 자동완성 옵션 조정:

    ```zsh
    # 최소 2글자 입력 후 자동완성 시작
    zstyle ':autocomplete:*' min-input 2
    # 모호하지 않은 경우 자동 삽입
    zstyle ':autocomplete:*' insert-unambiguous yes
    ```

3. 표시 형식 변경:

    ```zsh
    # 최대 5줄까지 표시
    zstyle ':autocomplete:*' list-lines 5
    # 최근 디렉토리 제안 방식
    zstyle ':autocomplete:*' recent-dirs-insert both
    ```

4. 키 바인딩 설정:

    ```zsh
    # Shift-Tab으로 역방향 순회
    bindkey '^[[Z' reverse-menu-complete
    ```

5. 특정 명령어에 대한 자동완성 비활성화:

    ```zsh
    # git 명령어에 대해 자동완성 비활성화
    zstyle ':autocomplete:*' ignored-input "git *"
    ```

### 설정 예제

```sh
# Enable zsh-autocomplete plugin
plugins+=(zsh-autocomplete)

# Source oh-my-zsh
source $ZSH/oh-my-zsh.sh

# 자동 완성 목록의 테마를 설정합니다.
#- `match`: 자동 완성 목록에서 일치하는 텍스트를 강조 표시합니다.
#- `simple`: 기본 테마로, 특별한 강조 없이 목록을 표시합니다.
zstyle ':autocomplete:*' theme 'match'

# 자동 완성 항목이 하나로 명확히 결정되었을 때 자동으로 삽입할지 여부를 설정합니다.
# - `yes`: 명확히 결정된 항목을 자동으로 삽입합니다.
# - `no`: 명확히 결정된 항목을 자동으로 삽입하지 않습니다.
zstyle ':autocomplete:*' insert-unambiguous yes

# 자동 완성 목록의 색상을 설정합니다.
# 색상과 속성을 조합하여 설정할 수 있습니다.
# - `current`: 현재 선택된 항목의 색상과 속성을 설정합니다.
# - `fg=COLOR`: 글꼴 색상을 설정합니다.
# - `bg=COLOR`: 배경 색상을 설정합니다.
# - `bold`: 굵게 표시합니다.
# - `underline`: 밑줄을 표시합니다.
zstyle ':autocomplete:*' list-colors '=(#b) #009900=00:current:#b:fg=blue,bold'

# 자동 완성 목록에 표시되는 최대 항목 수를 설정합니다.
zstyle ':autocomplete:*' max-lines 20

# 자동 완성 위젯의 스타일을 설정합니다.
# - `menu-select`: 메뉴 선택 스타일을 사용합니다. 키보드 화살표 키로 항목을 선택할 수 있습니다.
# - `default`: 기본 스타일을 사용합니다.
zstyle ':autocomplete:*' widget-style menu-select
```
