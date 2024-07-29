# Authorization

- [Authorization](#authorization)
    - [Authorization 헤더](#authorization-헤더)
    - [Bearer](#bearer)

## [Authorization](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization) 헤더

HTTP에서 `Authorization` 헤더는 클라이언트가 서버에게 자신의 인증 정보를 제공할 때 사용되는 메커니즘입니다.
`Bearer` 토큰은 이 `Authorization` 헤더를 사용하는 여러 인증 방식 중 하나로, 주로 OAuth 2.0이나 JWT(JSON Web Token) 인증에서 사용됩니다.

## Bearer

`Bearer`라는 단어는 "소지자", "보유자"를 의미하며,
`Bearer 토큰`은 "토큰을 제시하는 자"가 인증의 주체임을 나타냅니다.

즉, 이 토큰을 소지하고 있는 클라이언트는 해당 토큰에 담긴 권한을 가지고 서비스를 이용할 수 있습니다.

`Bearer` 토큰을 사용할 때의 `Authorization` 헤더 형식은 다음과 같습니다.

```text
Authorization: Bearer <token>
```
