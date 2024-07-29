# go vet

- [go vet](#go-vet)
    - [go vet](#go-vet-1)
    - [description](#description)
    - [Examples](#examples)

## go vet

Go 프로그래밍 언어 툴체인에서 제공하는 `go vet`은 Go 코드에 대한 정적 분석을 수행하여 잠재적인 오류, 버그, 의심스러운 구조, 그리고 관용적인 Go 사용법에서 벗어난 부분을 찾기 위해 사용됩니다.

Go 컴파일러가 성공 시 최소한의 출력만을 생성하도록 설계되었기 때문에 컴파일러의 출력만으로는 즉시 명확하지 않을 수 있는 문제들을 식별하는 데 도움을 줍니다.

`go vet` 도구는 Go 코드베이스의 품질과 정확성을 유지하는 데 도움을 주기 위해 Go 프로그래밍 언어 툴셋의 일부로 도입되었습니다.
다음과 같이 런타임 오류로 이어질 수 있는 일반적인 실수를 자동으로 감지합니다:
- 복사-붙여넣기 오류
- 포인터의 잘못된 사용
- 잘못된 형식 문자열
- 키가 없는 복합 리터럴 등

`go vet`을 사용하는 주요 장점은 다음과 같습니다:
- 개발 단계에서 더 쉽게 수정할 수 있는 오류와 잠재적인 버그를 조기에 감지합니다.
- Go 관용구에 대한 코드 품질 향상과 준수를 통해 다양한 개발자가 읽고 유지보수할 수 있는 코드를 보장합니다.
- 코드 편집 중에 `go vet`을 자동으로 실행할 수 있는 다양한 편집기 및 IDE와의 통합을 통해 개발자에게 실시간 피드백을 제공합니다.
- CI/CD 파이프라인의 일부로 사용되어 코드가 코드베이스에 병합되기 전에 오류에 대해 검증될 수 있습니다.

`go vet`은 Go 컴파일러가 제공하는 철저한 테스팅이나 정적 타입 검사의 필요성을 대체하지 않는다는 점에 주의합니다.

## description

```bash
# usage
go vet [-C dir] [-n] [-x] [-vettool prog] [build flags] [vet flags] [packages]
```

Vet runs the Go vet command on the packages named by the import paths.

For more about vet and its flags, see 'go doc cmd/vet'.
For more about specifying packages, see 'go help packages'.
For a list of checkers and their flags, see 'go tool vet help'.
For details of a specific checker such as 'printf', see 'go tool vet help printf'.

The `-C` flag changes to dir before running the 'go vet' command.
The `-n` flag prints commands that would be executed.
The `-x` flag prints commands as they are executed.

The `-vettool=prog` flag selects a different analysis tool with alternative or additional checks.

For example, the 'shadow' analyzer can be built and run using these commands:

```bash
go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
go vet -vettool=$(which shadow)
```

The build flags supported by go vet are those that control package resolution and execution, such as `-n`, `-x`, `-v`, `-tags`, and `-toolexec`.
For more about these flags, see 'go help build'.

See also: go fmt, go fix.

## Examples

```bash
go vet -vettool=(which containedctx) ./...
```
