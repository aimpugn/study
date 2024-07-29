# Authentication

- [Authentication](#authentication)
    - [Authentication?](#authentication-1)
        - [인증(Authentication)과 인가(Authorization)](#인증authentication과-인가authorization)
    - [Basic](#basic)
    - [Bearer](#bearer)

## Authentication?

### 인증(Authentication)과 인가(Authorization)

- 인증(Authentication)
    - 사용자의 신원을 확인하는 과정
    - 예를 들어, 사용자가 제공한 아이디와 비밀번호를 검증하여 해당 사용자가 누구인지 확인하는 것이 일종의 인증
- 인가(Authorization)
    - **이미 인증된** 사용자에 대해 특정 리소스에 접근하거나 특정 작업을 수행할 권한이 있는지를 결정하는 과정
    - 예를 들어, 인증된 사용자가 특정 문서를 읽거나 수정할 권한이 있는지 확인하는 것이 일종의 인가

## Basic

- Basic Authentication은 클라이언트가 사용자 이름과 비밀번호를 조합하여 Base64 인코딩된 문자열을 생성하고, 이를 Authorization 헤더에 포함시켜 서버로 전송하는 방식

```shell
# HTTP Header
Authorization: Basic base64(username:password)
```

- 상대적으로 덜 안전하며, 무엇보다도 통신이 암호화되지 않은 경우에 사용자 이름과 비밀번호가 노출될 수 있다
- 이러한 이유로, Basic Authentication은 HTTPS와 같은 암호화된 연결을 통해 사용되어야 한다
- 서버가 사용자 이름과 비밀번호를 검증하며, 이러한 자격 증명은 사용자 또는 시스템 관리자에 의해 직접 관리

## Bearer

> `Bearer`?
>
> "Bearer"란 단어 자체는 "소지자" 또는 "지닌 사람"이라는 의미를 가지고 있다.
> 원래 "Bearer"는 금융 분야에서 사용되며, "Bearer bond" 또는 "Bearer check"와 같이 해당 문서를 소지하고 있는 사람에게 특정 권리를 부여한다.
> 비슷한 방식으로, HTTP 인증에서 "Bearer Token"은 토큰을 소지하고 있는 사람에게 특정 권리를 부여한다
>
> "Bearer Token"이라는 용어는 토큰이 단순히 소지된 것만으로 해당 서비스 또는 리소스에 대한 액세스 권한을 부여받게 됨을 의미한다
> 다시 말해, 토큰을 "지닌"(bear) 클라이언트는 토큰이 부여하는 권한을 가지게 되며, 서버는 이 토큰을 기반으로 클라이언트의 요청을 인증하고 인가(Authorization)한다
>
> 이 "Bearer" 접두사는 토큰의 타입을 명시하며, 서버가 이 토큰을 어떻게 처리해야 하는지를 알려준다.
> "Bearer Token" 방식은 많은 인증 프레임워크와 표준, 특히 OAuth 2.0에서 널리 사용되며, 이는 클라이언트가 토큰을 단순히 "지니고" 있어야만 서버의 리소스에 액세스할 수 있음을 나타낸다

- Bearer Authentication은 클라이언트가 Authorization 헤더에 "Bearer"라는 단어와 함께 토큰을 포함시켜 서버로 전송하는 방식

```shell
# HTTP Header
Authorization: Bearer your-token
```

- 토큰 기반 인증 방식으로, 보안 토큰이 서버와 클라이언트 사이에 전송된다
- Bearer 토큰도 암호화된 연결을 통해 전송되어야 한다
- Bearer Authentication은 OAuth 2.0과 같은 토큰 기반 인증 프레임워크에서 더 복잡한 인증 및 인가 시나리오에 사용된다
- 토큰이 인증 서버에 의해 발급되며, 토큰의 수명, 취소 및 재발급 등이 관리될 수 있다
