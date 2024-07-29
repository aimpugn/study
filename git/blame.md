# blame

## blame?

> the stupid content tracker

## 특정 파일의 전체 내용에 대한 Blame 정보 얻기

```bash
git blame <파일명>
```

## 특정 라인 범위에 대한 Blame 정보 얻기

```bash
# `-L`: 라인 번호를 지정하여 특정 범위에 대한 정보를 얻기
git blame -L <시작 라인>,<끝 라인> <파일명>
```

## 특정 커밋 이후의 변경 사항만 보기

```bash
# `-s`: 특정 커밋 이후에 어떤 변경이 발생했는지
git blame -s <커밋 해시> <파일명>
```

## 더 많은 상세 정보 보기

```bash
# `-e`: 각 라인을 마지막으로 수정한 사람의 이메일 주소도 함께 표시
git blame -e <파일명>
```
