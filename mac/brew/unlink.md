# unlink

- [unlink](#unlink)
    - [cmd](#cmd)
    - [역할](#역할)

## cmd

```shell
unlink [--dry-run] installed_formula [...]
    Remove symlinks for formula from Homebrew´s prefix. This can be useful for temporarily disabling a formula: brew unlink formula && commands && brew link formula

    -n, --dry-run
            List files which would be unlinked without actually unlinking or deleting any files.
```

## 역할

- 특정 버전의 패키지와 관련된 심볼릭 링크를 제거하여, 그 버전의 실행 파일, 라이브러리, 헤더 파일 등에 대한 접근을 끊는다
- 가령, 실수로 잘못된 버전의 PHP를 실행하는 것을 방지할 수 있다. 이는 특히 여러 버전의 PHP가 설치되어 있고, 그 중 일부 버전이 더 이상 사용되지 않거나 호환성 문제가 있을 때 유용함
