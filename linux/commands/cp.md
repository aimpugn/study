# cp

## 날짜 붙여서 복사본 만들기

```sh
cp tmp{,.$(date +'%Y%m%d_%H%M')}
```

Bash shell의 중괄호 확장(brace expansion) 기능을 활용합니다.
`,` 콤마로 구별되며:
- 첫번째 공백은 원래 `tmp`라는 파일 이름을 의미합니다.
- 두번째는 `.YYYYMMDD_HHMM` 형식의 날짜가 붙습니다.

rename 할 때도 사용할 수 있습니다.

```sh
mv tmp{,.$(date +'%Y%m%d_%H%M%S')}
```

- [Copy a file and append a timestamp](https://unix.stackexchange.com/questions/202570/copy-a-file-and-append-a-timestamp)
