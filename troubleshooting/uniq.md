# uniq

## 출력 결과가 일정하지 않음

### 문제

```bash
rg -i 'isDirect' --files-with-matches | rg -r '$1' '^([^/]+).*' | uniq | sort
```

### 원인

### 해결

`sort`를 먼저 수행합니다.

```bash
rg -i 'isDirect' --files-with-matches | rg -r '$1' '^([^/]+).*' | sort | uniq
```

또는 `sort`의 `-u` 옵션을 사용합니다.

```bash
rg -i 'isDirect' --files-with-matches | rg -r '$1' '^([^/]+).*' | sort -u
```
