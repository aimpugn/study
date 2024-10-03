# HTTP POST

- [HTTP POST](#http-post)
    - [HTTP POST 요청](#http-post-요청)
    - [작동 방식](#작동-방식)
    - [특징](#특징)
    - [서버에서의 처리](#서버에서의-처리)
        - [PHP](#php)
        - [Node.js](#nodejs)

## HTTP POST 요청

POST 요청은 클라이언트가 서버에 데이터를 전송하여 *서버의 리소스를 생성하거나 업데이트*하기 위해 사용되는 HTTP 메서드입니다.

## 작동 방식

1. **클라이언트에서 서버로의 요청**

    클라이언트(예: 웹 브라우저, 모바일 앱)는 HTTP POST 요청을 생성하고, 요청 본문(body)에 데이터를 포함시켜 서버로 전송합니다.
    이 데이터는 [JSON, XML, application/x-www-form-urlencoded, multipart/form-data](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types) 등 다양한 형식으로 전송될 수 있습니다.

2. **서버 처리**

    서버는 받은 요청을 처리하고, 적절한 응답 코드와 함께 응답을 클라이언트에게 전송합니다.

    예를 들어, 요청이 성공적으로 처리되면 일반적으로 `201 Created` 상태 코드를 반환합니다.

    요청 처리 중 오류가 발생한 경우, `4XX` (클라이언트 오류) 또는 `5XX` (서버 오류) 범위의 상태 코드 등을 반환합니다.

## 특징

- **안전하지 않음(Non-safe)**

    POST 요청은 서버의 상태를 변경할 수 있으므로 안전한(safe) 메소드로 간주되지 않습니다.

- **멱등하지 않음(Non-idempotent)**

    같은 POST 요청을 여러 번 보내면, 서버에 여러 번의 영향을 줄 수 있습니다.
    예를 들어, 동일한 POST 요청을 여러 번 보내면 같은 데이터가 여러 번 생성될 수 있습니다.

- **캐시할 수 없음(Non-cacheable)**

    POST 요청의 결과는 일반적으로 캐시되지 않는다.

## 서버에서의 처리

### PHP

```php
<?php
// PHP 예시

class UserController {
    private const FORM_PARAM_USERNAME = 'username';
    private const FORM_PARAM_PASSWORD = 'password';
    private const FORM_PARAM_EMAIL = 'email';

    public function register() {
        $username = $_POST[self::FORM_PARAM_USERNAME] ?? '';
        $password = $_POST[self::FORM_PARAM_PASSWORD] ?? '';
        $email = $_POST[self::FORM_PARAM_EMAIL] ?? '';

        if (empty($username) || empty($password) || empty($email)) {
            return $this->sendError('필수 입력 항목이 누락되었습니다.');
        }

        // 사용자 등록 로직...
    }
}
```

### Node.js

```javascript
// JavaScript 예시 (Node.js with Express)

const express = require('express');
const app = express();

const FORM_PARAM_USERNAME = 'username';
const FORM_PARAM_PASSWORD = 'password';
const FORM_PARAM_EMAIL = 'email';

app.use(express.urlencoded({ extended: true }));

app.post('/register', (req, res) => {
    const username = req.body[FORM_PARAM_USERNAME];
    const password = req.body[FORM_PARAM_PASSWORD];
    const email = req.body[FORM_PARAM_EMAIL];

    if (!username || !password || !email) {
        return res.status(400).json({ error: '필수 입력 항목이 누락되었습니다.' });
    }

    // 사용자 등록 로직...
});
```
