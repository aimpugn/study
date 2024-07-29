# Coroutine

- [Coroutine](#coroutine)
    - [Coroutine](#coroutine-1)
        - [kotlin의 coroutine?](#kotlin의-coroutine)
        - [kotlin의 coroutine과의 차이점?](#kotlin의-coroutine과의-차이점)
    - [예시](#예시)

## Coroutine

- Guzzle의 Coroutine은 **PHP의 제너레이터**를 활용하여 비동기 작업의 흐름을 관리한다.
- 값 또는 프로미스들을 산출(yield)하는 제너레이터를 사용하여 resolve 되는 프로미스를 생성(C#의 비동기 키워드와 다소 유사).
- 코루틴 함수가 호출되면 제너레이터의 인스턴스가 시작되고 최종 산출(yield)된 값으로 fulfilled 된 프로미스를 반환
- 산출된 프로미스가 settle 상태가 되면 제어권이 제너레이터로 다시 반환된다. 이렇게 하면 중간 처리를 최소화하면서 많은 순차적 비동기 호출을 수행할 때 장황한 코드를 줄일 수 있다.
- 제너레이터를 사용하여 비동기 작업을 순차적으로 작성할 수 있게 해주며, `yield` 키워드를 통해 프라미스(promise)의 resolve를 기다리며, 이를 통해 비동기 프로그래밍을 보다 직관적이고 선형적으로 작성할 수 있다.

### kotlin의 coroutine?

- Kotlin의 코루틴은 비동기 프로그래밍과 동시성 프로그래밍을 간결하고 효율적으로 작성할 수 있도록 하는 고급 기능
- Kotlin 코루틴은 언어 차원에서 지원되며, 비동기 작업을 쉽게 관리하고, 리소스를 효율적으로 사용할 수 있게 해준다

### kotlin의 coroutine과의 차이점?

- Kotlin의 코루틴은 언어 차원에서 지원되는 반면, Guzzle의 Coroutine은 PHP의 제너레이터를 활용한 라이브러리 수준의 구현
- Kotlin의 코루틴은 비동기 작업을 쉽게 구성하고 관리할 수 있는 다양한 기능과 도구를 제공한다. 반면, Guzzle의 Coroutine은 주로 비동기 HTTP 요청의 흐름을 관리하는 데 초점을 맞추고 있다
- Kotlin과 PHP는 기본적으로 다른 언어 특성을 가지고 있기 때문에, 두 코루틴의 내부 동작과 효율성에 차이가 있다

## 예시

```php
use GuzzleHttp\Promise;

function createPromise($value) {
    return new Promise\FulfilledPromise($value);
}

$promise = Promise\coroutine(function () {
    $value = (yield createPromise('a'));
    try {
        $value = (yield createPromise($value . 'b'));
    } catch (\Exception $e) {
        // The promise was rejected.
    }
    yield $value . 'c';
});

// Outputs "abc"
$promise->then(function ($v) { echo $v; })
```
