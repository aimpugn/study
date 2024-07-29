# error

## try ~ catch

PHP에서 `try-catch` 블록 내에서 한 번 예외가 캐치되면, 그 이후의 `catch` 블록은 실행되지 않는다.
즉, 한 `try` 블록에 대해 오직 하나의 `catch` 블록만 실행될 수 있다.

```php
try {
    $req = new \GuzzleHttp\Psr7\Request('GET', 'localhost');
    throw new \GuzzleHttp\Exception\RequestException('test', $req, null, null, []);
} catch (\GuzzleHttp\Exception\RequestException $e) {
    print_r("1" . PHP_EOL);
    throw new Exception('test2');
} catch (Exception $e) {
    print_r("2" . PHP_EOL);
    throw new Exception('test3');
} catch (GuzzleHttp\Exception\GuzzleException $e) {
    print_r("3" . PHP_EOL);
} finally {
    print_r('finally!');
}
```

따라서, 주어진 코드에서 `\GuzzleHttp\Exception\RequestException`이 발생하고 캐치되면, 첫 번째 `catch` 블록이 실행되고, 그 안에서 다시 `Exception`을 던지더라도, 그것을 캐치할 추가적인 `catch` 블록으로 이동하지 않는다. 대신, 그 `Exception`은 현재 `try-catch` 구조를 벗어나 호출 스택에서 상위의 `try-catch` 구조(있다면)로 넘어간다.

주어진 코드를 실행하면 다음과 같은 흐름을 따른다:

1. `RequestException`이 발생하고 첫 번째 `catch` 블록에서 캐치된다
2. 첫 번째 `catch` 블록에서 "1"과 개행 문자(PHP_EOL)를 출력한다.
3. 첫 번째 `catch` 블록 내에서 `Exception`을 새로 던진다(`throw new Exception('test2');`).
4. 이 새로 던진 `Exception`은 현재 `try-catch` 블록 내에서는 캐치되지 않고, 현재 블록을 벗어나 상위 레벨의 에러 핸들러나 다른 `try-catch` 구조로 이동한다. 만약 상위 레벨에 적절한 `catch` 블록이 없다면, 이는 스크립트의 종료를 유발할 수 있다.
5. `finally` 블록은 예외의 발생 여부와 상관없이 항상 실행되며, 여기서 'finally!'와 개행 문자를 출력한다.
