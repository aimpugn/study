# script path

## 현재 스크립트 경로

```bash
#!/bin/bash

# [zsh] 현재 스크립트
# https://unix.stackexchange.com/questions/76505/unix-portable-way-to-get-scripts-absolute-path-in-zsh
# zsh에서 ${BASH_SOURCE[0]}과 같은 코드는?
# https://stackoverflow.com/questions/9901210/bash-source0-equivalent-in-zsh
# ${(%):-%N}
# ${(%):-%x}
CURRENT_SCRIPT="${0// }"
echo "${BASH_SOURCE[@]}"
echo "1. using current script: $(cd "$(dirname "$CURRENT_SCRIPT")" && pwd -P)"
echo "2. using bash source: $(dirname "${BASH_SOURCE[0]}")"

# A. when local m1 mac(oh my zsh)
#   # (공백)
# 1. using current script: /Users/rody/IdeaProjects/${DIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_B}
# 2. using bash source: .

# B. when ubuntu container
# /home/ubuntu/${SUBDIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_B}/functions
# 1. using current script: /usr/bin
# 2. using bash source: /home/ubuntu/${SUBDIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_B}

# C. when github actions
# /home/runner/work/${CURRENT_SERVICE_NAME}/${CURRENT_SERVICE_NAME}/${DIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_B}/functions /home/runner/work/_temp/909ae778-8037-4624-b4c8-a5db5399c9c9.sh
# 1. using current script: /home/runner/work/_temp
# 2. using bash source: /home/runner/work/${CURRENT_SERVICE_NAME}/${CURRENT_SERVICE_NAME}/${DIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_B}
# /home/runner/work/${CURRENT_SERVICE_NAME}/${CURRENT_SERVICE_NAME}/${DIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_B}/functions /home/runner/work/${CURRENT_SERVICE_NAME}/${CURRENT_SERVICE_NAME}/${DIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_B}/shell_script
# 1. using current script: /home/runner/work/${CURRENT_SERVICE_NAME}/${CURRENT_SERVICE_NAME}/${DIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_B}
# 2. using bash source: /home/runner/work/${CURRENT_SERVICE_NAME}/${CURRENT_SERVICE_NAME}/${DIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_A}/${SUBDIRECTORY_NAME_B}

# Can a shell script tell what directory it's in. No bashisms
# https://stackoverflow.com/questions/8468519/can-a-shell-script-tell-what-directory-its-in-no-bashisms
`dirname -- "$0"`
```
