# filter-branch

## `filter-branch`?

- 저장소의 히스토리를 필터링하고 재작성하는데 사용

## 특정 파일을 모든 커밋에서 제거

```bash
git filter-branch \
    --force \
    --index-filter \
    "git rm --cached --ignore-unmatch 파일명" \
    --prune-empty \
    --tag-name-filter cat -- --all
```
