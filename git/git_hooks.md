# git hooks

- [git hooks](#git-hooks)
    - [git pre-push](#git-pre-push)

## git pre-push

`git push` 전에 실행되는 스크립트입니다.
push 전에 테스트 등을 실행하여 테스트가 실패할 경우 push 되는 것을 방지할 수 있습니다.

```bash
#!/bin/bash

# 임시 파일 생성
temp_file=$(mktemp)

docker_command="docker exec php5.6 /bin/bash -c"
command_string="source /home/ubuntu/test-resources/scripts/functions && cd /var/www/someservice && run-test All"

# 출력과 동시에 결과를 파일에 저장하기 위해 `| tee`를 사용합니다.
$docker_command "$command_string" | tee "$temp_file"

# 임시 파일에서 결과를 읽어 변수에 저장
phpunit_result=$(<"$temp_file")

# 임시 파일 삭제
rm "$temp_file"

# 이제 $phpunit_result 변수에 phpunit의 전체 출력이 저장되어 있습니다.
# 이를 활용하여 추가 작업을 수행할 수 있습니다.

# 예: 결과에서 특정 문자열 찾기
if echo "$phpunit_result" | grep -q "FAILURES!"; then
    echo "테스트에 실패했습니다."
else
    echo "모든 테스트가 통과했습니다."
fi

# 예: 결과를 파일로 저장
echo "$phpunit_result" > phpunit_results.txt

# 예: 결과의 줄 수 계산
line_count=$(echo "$phpunit_result" | wc -l)
echo "총 $line_count 줄의 출력이 있었습니다."
```
