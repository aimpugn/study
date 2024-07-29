# gosec

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

Citations:
