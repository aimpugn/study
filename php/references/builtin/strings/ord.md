# [ord](https://www.php.net/manual/en/function.ord.php)

## `ord`?

PHP `ord` 함수는 문자열의 첫 번째 바이트를 0 ~ 255 사이의 정수로 반환한다.

> ASCII 문자열의 경우, 첫 번째 바이트는 문자의 코드 포인트와 동일하다.
>
> - 0-31 are the ASCII control characters
> - 32-127 are the printable characters.
> - 128-255 is the extended ASCII set

## 멀티 바이트 문자열

```php
public function main()
{
    $str = '🐘';
    for ($pos = 0; $pos < strlen($str); $pos++) {
        $byte = substr($str, $pos);
        echo 'Byte ' . $pos . ' of $str has value ' . ord($byte) . PHP_EOL;
    }
}
// Byte 0 of $str has value 240
// Byte 1 of $str has value 159
// Byte 2 of $str has value 144
// Byte 3 of $str has value 152
```
