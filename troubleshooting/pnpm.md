# pnpm

- [pnpm](#pnpm)
    - [Error: `env` and `jsc.target` cannot be used together](#error-env-and-jsctarget-cannot-be-used-together)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)

## Error: `env` and `jsc.target` cannot be used together

### 문제

```log
../.. build: Build completed!
../.. build: [Error: `env` and `jsc.target` cannot be used together] {
../.. build:   code: 'GenericFailure'
../.. build: }
../.. build:  ELIFECYCLE  Command failed with exit code 1.
../.. build: Failed
```

### 원인

이 오류 메시지는 `env`와 `jsc.target` 옵션을 동시에 사용할 수 없다는 것을 나타냅니다.
이는 Babel 또는 다른 JavaScript 컴파일러 설정에서 발생할 수 있는 일반적인 문제입니다.
`env`와 `jsc.target`은 서로 충돌하는 설정이므로 둘 중 하나만 사용해야 합니다.

- `env`와 `jsc.target`은 JavaScript 컴파일러 설정에서 서로 다른 방식으로 타겟 환경을 지정합니다.
- 이 두 옵션을 동시에 사용하면 충돌이 발생하여 컴파일러가 어떤 설정을 우선해야 할지 알 수 없게 됩니다.

근데 pnpm 버전을 9.1.2로 업그레이드하고 pnpm install 다시 실행하니 해결 됐다

### 해결

1. root 경로의 `pnpm-lock.yaml` 를 롤백
2. `pnpm install`
3. `pnpm start:dev`
