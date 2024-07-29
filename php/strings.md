# strings

- [strings](#strings)
    - [`mb_substr`과 `8bit` 인코딩](#mb_substr과-8bit-인코딩)

## `mb_substr`과 `8bit` 인코딩

```php
$decodedData = base64_decode($encryptedData);

// $data에서 iv 분리
$ivLen = 12;
$iv = substr($decodedData, 0, $ivLen);
$data = substr($decodedData, $ivLen);

// $data에서 태그, ADD 분리
$tagLenInBits = 128 / 8;
$addLenInBits = 128 / 8;

$ciphertext = mb_substr($data, 0, -$tagLenInBits - $addLenInBits, '8bit');
$tagWithADD = mb_substr($data, -$tagLenInBits - $addLenInBits, null, '8bit');
$tag = mb_substr($tagWithADD, 0, $tagLenInBits, '8bit');
$add = mb_substr($tagWithADD, $tagLenInBits, null, '8bit');
```

- `mb_substr` 함수의 `encoding` 파라미터를 `8bit`로 설정하는 이유는 **바이트 단위로 문자열을 처리**하기 위함
    - PHP에서 `substr` 함수는 기본적으로 바이트 단위로 문자열을 처리
    - 하지만 `mb_substr` 함수는 멀티바이트 문자열을 처리하기 위해 설계되었음
    - 멀티바이트 문자열은 여러 바이트로 구성된 문자를 포함할 수 있으며, 이러한 문자들은 `mb_substr` 함수를 사용하지 않으면 잘못 처리될 수 있다
- 위 코드의 경우
    - 암호화된 데이터를 처리하고 있으며, 이 데이터는 임의의 바이트 값을 가질 수 있다
    - 따라서 문자열을 바이트 단위로 처리해야 하며, `mb_substr` 함수의 `encoding` 파라미터를 `8bit`로 설정하여 이를 달성할 수 있다
    - `8bit` 인코딩은 각 문자를 1바이트로 처리하며, 널 종료 문자(`\0`)를 무시한다
