# go get

## 패키지 추가하기

```bash
go get github.com/jfcg/sorty/v2

# 버전 지정
go get github.com/jfcg/sorty/v2@v2.1.0
```

## [패키지 삭제](https://stackoverflow.com/a/67620609)

```bash
go get package@none

# example
go get github.com/jfcg/sorty@none
```

```bash
go clean -i importpath...

go clean -i -n github.com/motemen/gore...
```
