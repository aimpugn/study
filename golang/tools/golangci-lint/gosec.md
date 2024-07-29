# gosec

- [gosec](#gosec)
    - [gosec](#gosec-1)
    - [`// #nosec ***`](#-nosec-)
    - [`G101`](#g101)
        - [false positive](#false-positive)
        - [결론](#결론)
    - [G501](#g501)
    - [G401](#g401)

## gosec

`gosec`는 정적 분석을 통해 코드 내의 패턴을 검사합니다. 특히, 문자열 내에 특정 키워드(예: password, secret, token 등)가 포함되어 있을 경우, 이를 잠재적인 하드코딩된 크리덴셜로 간주할 수 있습니다.

## `// #nosec ***`

`// #nosec` 주석은 `gosec` 도구에서만 사용되는 특별한 주석으로, `gosec`의 정적 분석에서 특정 코드 라인이나 블록을 무시하도록 지시합니다.

`// #nosec` 주석을 사용할 때는 다음과 같이 코드에 직접 주석을 추가합니다. 필요한 경우, 특정 경고 번호도 함께 명시할 수 있습니다.

```go
// #nosec G101
sensitiveFunctionCall()
```

위 예제에서 `// #nosec G101` 주석은 `gosec`에게 이 줄에서 발생할 수 있는 `G101` 경고를 무시하라고 지시합니다.

만약 모든 종류의 경고를 무시하고 싶다면, 경고 번호 없이 `// #nosec`만 사용할 수 있습니다.

다른 린트 도구들, 예를 들어 `golint`나 `staticcheck` 등은 `//nolint` 주석을 사용하여 린트 경고를 무시합니다. 이 주석은 다음과 같이 사용됩니다:

```go
//nolint:staticcheck
complexFunctionCall()
```

이 주석은 `staticcheck` 경고를 무시하도록 지시합니다. `//nolint` 주석은 린트 도구에 따라 다르게 적용되며, 각 도구의 문서를 참조하여 올바른 사용법을 확인해야 합니다.

## `G101`

`G101: Potential hardcoded credentials` 경고는 `gosec` (Go Security Checker) 도구에서 발생하는 것으로, 코드 내에 하드코딩된 크리덴셜(비밀번호, API 키, 토큰 등)이 포함되어 있을 가능성을 지적합니다. 이 경고는 보안상의 위험을 줄이기 위해 중요한 정보를 코드에 직접 하드코딩하지 않도록 권장합니다.

### false positive

```bash
G101: Potential hardcoded credentials (gosec)
    userByAccessTokenSQL = `
SELECT
    aaa.id,
    aaa.column1
FROM aaa
WHERE aaa.column2 = ?
    AND aaa.column3 > ?
    AND aaa.active = TRUE`
```

SQL 쿼리문 내에서는 실제 크리덴셜이 포함되어 있지 않음에도 불구하고, `gosec`가 잘못 감지하는 경우가 발생할 수 있습니다.

제시된 코드에서는 SQL 쿼리문이 하드코딩된 크리덴셜을 포함하고 있지 않습니다.
이 경우, `gosec`가 SQL 쿼리 문자열을 잘못 해석하여 경고를 발생시킬 수 있습니다.
이는 `gosec`의 휴리스틱 분석 방식 때문에 발생하는 오진(false positive)일 수 있습니다.

1. **경고 무시**: 이 경고가 잘못된 것이 확실한 경우, 해당 라인에 주석을 추가하여 `gosec` 경고를 무시할 수 있습니다. 예를 들어:

   ```go
   // #nosec G101
   userByAccessTokenSQL = `
   SELECT
       aaa.id,
       aaa.column1
   FROM aaa
   WHERE aaa.column2 = ?
       AND aaa.column3 > ?
       AND aaa.active = TRUE`
   ```

   `#nosec` 주석은 `gosec`에게 이 줄을 검사하지 말라고 지시합니다.

2. **gosec 설정 조정**:

    `gosec`의 설정을 조정하여 특정 경고 레벨이나 타입을 무시하도록 설정할 수 있습니다.

    예를 들어, `G101` 경고만 무시하도록 설정 파일을 조정할 수 있습니다.

3. **보안 감사 수행**

    잘못된 경고일지라도, 정기적인 보안 감사를 통해 실제 하드코딩된 크리덴셜이 코드에 포함되어 있지 않은지 확인하는 것이 좋습니다.

### 결론

`G101` 경고는 때때로 잘못된 경고를 발생시킬 수 있으며, 특히 SQL 쿼리와 같은 경우에는 크리덴셜과 관련 없는 코드에서도 발생할 수 있습니다. 이러한 경우, 적절한 주석을 추가하거나 도구의 설정을 조정하여 경고를 관리할 수 있습니다.

## G501

G501: Blocklisted import crypto/md5: weak cryptographic primitive (gosec)

이 경고는 `crypto/md5` 패키지를 임포트하는 것에 대한 것입니다.

- 의미: MD5는 약한 암호화 알고리즘으로 간주되어 보안에 취약합니다.
- 문제점:
    - MD5는 충돌 저항성이 낮아 다른 입력으로 같은 해시를 생성할 수 있습니다.
    - 현대의 컴퓨팅 파워로 MD5 해시를 깨는 것이 가능해졌습니다.
- 권장사항: 보안이 중요한 용도로는 MD5 대신 SHA-256과 같은 더 강력한 해시 함수를 사용해야 합니다.

## G401

G401: Use of weak cryptographic primitive (gosec)

이 경고는 실제로 `md5.Sum()` 함수를 사용하는 것에 대한 것입니다.

- 의미: 코드에서 실제로 MD5 해시 함수를 사용하고 있습니다.
- 문제점:
    - MD5로 생성된 해시는 안전하지 않으며, 데이터 무결성이나 인증에 사용하기에 부적절합니다.
    - 악의적인 공격자가 이를 이용해 보안을 우회할 수 있습니다.
- 권장사항:
    - 데이터 무결성 검사에는 SHA-256 또는 SHA-3와 같은 더 강력한 해시 함수를 사용하세요.
    - 암호 해싱에는 bcrypt, scrypt, Argon2와 같은 전용 알고리즘을 사용하세요.

해결 방안:
1. 만약 단순히 체크섬이나 비암호화 용도로 MD5를 사용 중이라면, 목적에 따라 다른 알고리즘으로 대체할 수 있습니다.
2. 보안이 중요한 경우, `crypto/sha256`을 사용하여 다음과 같이 코드를 수정할 수 있습니다:

   ```go
   import "crypto/sha256"

   // ...

   hashed := sha256.Sum256(data)
   ```

3. 암호 해싱의 경우, `golang.org/x/crypto/bcrypt`와 같은 전용 라이브러리를 사용하는 것이 좋습니다.

이러한 변경을 통해 코드의 보안성을 크게 향상시킬 수 있습니다.
단, MD5를 사용하는 특별한 이유가 있다면 (예: 레거시 시스템과의 호환성), 그 사용 목적과 보안 위험을 명확히 문서화하는 것이 중요합니다.
