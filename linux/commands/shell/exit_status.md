# exit status

```bash
is_failure=$(grep -q Failures "$temp_file" && echo $?) # is_failure에 0 또는 1이 저장 안 됨

grep -q Failures "$temp_file"
is_failure=$(echo $?) # is_failure에 0 또는 1이 저장됨
```

1. 첫 번째 방식:

```bash
is_failure=$(grep -q Failures "$temp_file" && echo $?)
```

   이 방식의 동작:
   a) `grep -q Failures "$temp_file"`를 실행합니다.
   b) grep 명령이 성공하면 (즉, "Failures"를 찾으면), `&&` 연산자 때문에 `echo $?`가 실행됩니다.
   c) `echo $?`는 grep 명령의 종료 상태를 출력합니다 (이 경우 0).
   d) grep 명령이 실패하면 (즉, "Failures"를 찾지 못하면), `&&` 연산자 때문에 `echo $?`가 실행되지 않습니다.
   e) 전체 명령의 결과가 `is_failure`에 할당됩니다.

   문제점:
- grep이 실패할 경우 (Failures를 찾지 못함), `echo $?`가 실행되지 않아 `is_failure`에 아무 값도 할당되지 않습니다.
- 이로 인해 `is_failure`는 비어있게 되거나, 이전에 할당된 값을 유지하게 됩니다.

2. 두 번째 방식:

```bash
grep -q Failures "$temp_file"
is_failure=$(echo $?)
```

   이 방식의 동작:
   a) `grep -q Failures "$temp_file"`를 실행합니다.
   b) grep 명령의 결과와 관계없이, 다음 줄의 `echo $?`가 항상 실행됩니다.
   c) `echo $?`는 직전에 실행된 grep 명령의 종료 상태를 출력합니다.
   d) 이 출력 결과가 `is_failure`에 할당됩니다.

   장점:
- grep 명령의 성공 여부와 관계없이 항상 종료 상태가 `is_failure`에 저장됩니다.
- grep이 성공하면 (Failures를 찾음) `is_failure`는 0이 됩니다.
- grep이 실패하면 (Failures를 찾지 못함) `is_failure`는 1이 됩니다.

주요 차이점:
1. 첫 번째 방식은 조건부 실행(`&&`)을 사용하여 grep이 성공할 때만 `echo $?`를 실행합니다. 이로 인해 grep이 실패할 경우 결과가 저장되지 않습니다.
2. 두 번째 방식은 grep의 결과와 관계없이 항상 `echo $?`를 실행하므로, grep의 종료 상태가 항상 `is_failure`에 저장됩니다.

결론적으로, 두 번째 방식이 더 안정적이고 예측 가능한 결과를 제공합니다. 이 방식은 grep 명령의 성공 여부와 관계없이 항상 정확한 종료 상태를 `is_failure` 변수에 저장합니다.
