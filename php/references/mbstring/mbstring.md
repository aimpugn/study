# mbstring

## `mb_check_encoding`

- 이 함수는 주어진 문자열이 주어진 인코딩으로 유효한 바이트 시퀀스를 가지고 있는지 확인
- 이 함수는 문자열이 특정 인코딩으로 유효하게 인코딩될 수 있는지만 확인
- 그 인코딩이 실제 문자열의 인코딩인지는 확인하지 않는다

## `mb_detect_encoding`

- 이 함수는 주어진 문자열의 인코딩을 탐지하려고 시도
- 이 함수는 가능한 인코딩 목록을 제공하고, 문자열이 어떤 인코딩으로 되어 있는지를 탐지하려고 시도

```php
public static function detectCharSet($str, $encodings = ['EUC-KR', 'UTF-8', 'UHC'])
{
    $detected = mb_detect_encoding($str, $encodings, true);
    if ($detected !== false) {
        return $detected;
    }
    foreach ($encodings as $encoding) {
        if (mb_check_encoding($str, $encoding)) {
            return $encoding;
        }
    }

    return null;
}
```
