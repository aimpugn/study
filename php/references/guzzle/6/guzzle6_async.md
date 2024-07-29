# guzzle async

- [guzzle async](#guzzle-async)
    - [`Promise`와 비동기 프로그래밍](#promise와-비동기-프로그래밍)
    - [비동기 요청 처리 과정](#비동기-요청-처리-과정)
        - [`Promise`란?](#promise란)
    - [`Settled`, `Resolved`, `Wait`, `Unwrap`](#settled-resolved-wait-unwrap)
        - [`Settled`](#settled)
        - [`Resolved`](#resolved)
        - [`Settled`와 `Resolved` 구분](#settled와-resolved-구분)
        - [`Wait`](#wait)
            - [`wait(false)` 사용 시의 동작](#waitfalse-사용-시의-동작)
        - [`Unwrap`](#unwrap)
            - [Guzzle의 `unwrap` 함수](#guzzle의-unwrap-함수)
            - [Guzzle의 `wait`와 `unwrap`](#guzzle의-wait와-unwrap)
    - [Concurrent Requests](#concurrent-requests)
        - [`settle` 함수](#settle-함수)
            - [`Promise\settle($promises)->wait($unwrap)` 함수와 `$unwrap` 매개변수](#promisesettlepromises-waitunwrap-함수와-unwrap-매개변수)
            - [`Promise\settle($promises)->wait($unwrap)` 후 처리](#promisesettlepromises-waitunwrap-후-처리)
        - [`settle`과 `unwrap`의 차이점](#settle과-unwrap의-차이점)
        - [예시](#예시)
    - [`Resolved`](#resolved-1)

## `Promise`와 비동기 프로그래밍

- `Promise`
    - 비동기 연산의 최종 완료(또는 실패) 및 그 결과값을 나타내는 객체
    - `Promise`는 해당 연산이 완료될 때까지 대기하거나, 완료되면 그 결과를 가져올 수 있는 메커니즘을 제공한다
- 비동기 프로그래밍
    - 프로그램의 주 실행 흐름을 멈추지 않고, 특정 작업이 백그라운드에서 실행되도록 하는 프로그래밍 방식

## 비동기 요청 처리 과정

1. 비동기 요청 (Async Request): 비동기 요청을 시작한다. 백그라운드에서 실행되고, 메인 실행 흐름을 차단하지 않는다.
2. Promise 생성: 비동기 요청에 대한 결과를 나중에 받기 위해 Pending 상태인 `Promise` 객체가 생성된다
3. Fulfilled 또는 Rejected (Pending -> Resolved)
    - Fulfilled: 비동기 요청이 성공적으로 완료되었고, Promise는 'fulfilled' 상태가 된다
    - Rejected: 비동기 요청이 실패했고, Promise는 'rejected' 상태가 된다
4. Unwrapping the Promise: Promise의 최종 상태에 따라 동기적으로 결과를 처리
    - Fulfilled 상태: Promise의 'fulfilled' 상태에서는 그 값을 반환
    - Rejected 상태: Promise가 'rejected' 상태인 경우 예외가 발생

### `Promise`란?

- `Promise`의 개념
    - 비동기 연산의 최종 완료(또는 실패)와 그 결과를 나타내는 객체
    - 즉, 어떤 연산이 나중에 완료될 것이라는 '약속'을 표현한다
- `Promise`는 세 가지 상태 중 하나를 갖는다
    - `Pending`: 아직 결과가 결정되지 않은 초기 상태.
    - `Fulfilled`: 연산이 성공적으로 완료됨.
    - `Rejected`: 연산이 실패함.
- 비동기 연산 처리
    - Promise는 비동기 연산이 완료되었을 때 그 결과를 처리하는 방법을 제공한다
    - 예를 들어, HTTP 요청의 결과를 받은 후 처리하거나, 오류가 발생했을 때 이를 처리하는 등의 작업을 한다

## `Settled`, `Resolved`, `Wait`, `Unwrap`

### `Settled`

- 프라미스가 완전히 처리되었음을 의미한다. "settled" 상태는 "fulfilled"와 "rejected"를 모두 포함하는 개념으로, 프라미스가 최종적으로 완료된 상태를 나타낸다
    - 성공('fulfilled')
    - 또는 실패('rejected').
- 즉, 프라미스의 최종 상태가 결정되었다는 것을 의미.
- 프라미스가 더 이상 상태 변화가 없을 때 사용되며, 즉, 프라미스가 'fulfilled' 되었든 'rejected' 되었든 상관없이 처리가 완료된 상태를 나타냅니다

### `Resolved`

- 프라미스가 "fulfilled" 또는 "rejected" 상태 중 하나로 결정된 상태를 갖고 있음을 나타내는 데 사용된다. 즉, 프라미스의 결과가 결정되었다는 것을 나타낸다.
- 종종 'fulfilled' 상태의 프라미스를 가리키는 데 사용되지만, 'settled'와 마찬가지로 프라미스가 최종 상태에 도달했음을 의미할 수도 있다
- 'Resolved'는 프라미스가 어떤 최종 상태(성공 또는 실패)에 도달했다는 것을 나타내기도 한다

### `Settled`와 `Resolved` 구분

- settled 프라미스가 더 넓은 개념. resolved 프라미스가 settled 프라미스에 포함된다
    - 모든 "resolved" 프라미스는 "settled"이다
    - 모든 "settled" 프라미스가 "resolved"는 아니다
- "resolved" 상태: 프라미스의 결과가 결정되었음을 의미하지만, 그렇다고 **최종적인 완료**를 의미하지는 않는다.
    - 예를 들어, 프라미스가 다른 프라미스로 "resolved" 될 수 있으며, 이 경우 최종적인 "settled" 상태에 이르기까지 추가적인 처리가 필요할 수 있다
- "settled" 상태: 프라미스가 최종적으로 완료된 상태를 의미.

### `Wait`

- 프라미스들이 'settled' 상태가 될 때까지 기다린다
- `wait()` 함수를 호출할 때 기본적으로 프라미스는 'unwrap'된다. 이는 프라미스가 완료되면, 그 결과(값 또는 예외)가 현재 실행 흐름에 합쳐진다는 것을 의미

#### `wait(false)` 사용 시의 동작

- 예외 발생 방지: `wait(false)`를 호출하면 프라미스가 'rejected' 상태여도 예외를 발생시키지 않는다. 이는 프라미스가 처리될 때까지 기다리되, 그 결과(예외)를 'unwrap'하지 않는 것을 의미
- 값 반환 없음: `wait(false)`를 사용할 때는 프라미스의 최종 값이 반환되지 않는다. 즉, 'fulfilled' 상태의 프라미스 결과값도 반환되지 않는다.

```php
$promise = new Promise();
$promise->reject('foo');
// This will not throw an exception. 
// It simply ensures the promise has been resolved.
$promise->wait(false);
```

### `Unwrap`

- 프라미스가 'fulfilled'(이행됨) 되었을 때 그 값을 반환하거나, 'rejected'(거부됨) 되었을 때 예외를 발생시키는 것을 의미
    - 즉, 프라미스의 최종 상태에 따라 동기적으로 값을 반환하거나 예외를 처리
    - 이는 'fulfilled' 상태의 프라미스 결과만을 중요시하는 시나리오에 적합하다

#### Guzzle의 `unwrap` 함수

```php
$promises = [
    'image' => $client->getAsync('/image'),
    'data'  => $client->getAsync('/data')
];

// `unwrap` 함수 'image'와 'data'라는 두 개의 비동기 요청에 대한 Promise를 처리
// 모든 요청이 성공적으로 완료되면, 그 결과를 배열로 반환
// 만약 어떤 요청이 실패하면, 예외가 발생
$results = GuzzleHttp\Promise\unwrap($promises);

echo $results['image']->getBody();
echo $results['data']->getBody();
```

- `unwrap` 함수의 역할
    - `unwrap` 함수는 여러 `Promise` 객체들을 동시에 처리할 수 있도록 해준다
    - 이 함수는 주어진 모든 `Promise`가 완료될 때까지 기다린 후, 그 결과를 배열로 반환한다
- 결과 처리
    - `unwrap` 함수는 각 `Promise`가 `fulfilled` 상태일 때의 결과만을 다룬다
    - 즉, 각 `Promise`가 성공적으로 완료되었을 때의 결과값을 배열로 모아 반환한다
- 예외 처리
    - 만약 어떤 `Promise`가 `rejected` 실패 상태라면, `unwrap` 함수는 예외를 발생시킨다
    - 이는 하나 이상의 `Promise`가 실패했을 때 처리가 중단되어야 한다는 것을 의미한다

#### Guzzle의 `wait`와 `unwrap`

- `wait` 메서드
    - `PromiseInterface`에 정의된 `wait` 메서드는 프라미스가 완료될 때까지 현재 스크립트의 실행을 차단(블로킹)한다
    - 즉, 비동기 작업이 완료될 때까지 기다린다(wait)
- `unwrap` 매개변수
    - `wait` 메서드에 전달되는 `unwrap` 매개변수는, 프라미스가 완료된 후에 그 결과값을 '풀어내고(unwrapping)' 싶은지를 결정한다
    - `true`로 설정하면, 프라미스가 성공적으로 해결된 경우 그 값을 반환하고, 거부된 경우 예외를 발생시킨다
    - `false`인 경우, 프라미스 자체를 반환한다
- `unwrap` 함수
    - Guzzle에서 `unwrap` 함수는 여러 프라미스를 동시에 기다릴 수 있게 해준다
    - 이 함수는 각 프라미스에 대해 `wait`을 호출하고, 모든 프라미스가 해결될 때까지 기다린 후, 각 프라미스의 결과를 배열로 반환한다.

## [Concurrent Requests](https://docs.guzzlephp.org/en/6.5/quickstart.html#concurrent-requests)

```php
use GuzzleHttp\Client;
use GuzzleHttp\Promise;

$client = new Client(['base_uri' => 'http://httpbin.org/']);

// Initiate each request but do not block
$promises = [
    'image' => $client->getAsync('/image'),
    'png'   => $client->getAsync('/image/png'),
    'jpeg'  => $client->getAsync('/image/jpeg'),
    'webp'  => $client->getAsync('/image/webp')
];

// Wait for the requests to complete; throws a ConnectException
// if any of the requests fail
$responses = Promise\unwrap($promises);

// Wait for the requests to complete, even if some of them fail
// $responses = Promise\settle($promises)->wait(true | false);
$responses = Promise\settle($promises)->wait();

// You can access each response using the key of the promise
echo $responses['image']->getHeader('Content-Length')[0];
echo $responses['png']->getHeader('Content-Length')[0];
```

### `settle` 함수

- 이 함수는 여러 프라미스를 동시에 처리할 때 사용된다
- 프라미스가 성공(`fulfilled`) 상태이든 실패(`rejected`) 상태이든 상관하지 않고 주어진 모든 프라미스가 완료될 때까지 기다린다.

```php
Promise\settle($httpRequestPromises);
```

- 주어진 프라미스 배열 `$httpRequestPromises`에 대해 각 요청이 'fulfilled' (이행됨) 또는 'rejected' (거부됨) 상태로 완료될 때까지 기다린다
- 그러나 이 함수 자체는 결과를 반환하지 않는다. 대신, 각 프라미스의 **최종 상태에 대한 정보**를 포함하는 프라미스를 반환한다

#### `Promise\settle($promises)->wait($unwrap)` 함수와 `$unwrap` 매개변수

- `$unwrap`이 `true`
    - 모든 프라미스가 완료될 때까지 기다린다
    - 각 프라미스의 결과를 '풀어내고(unwrapping)' 반환한다
    - 이 경우, 성공적으로 해결된 프라미스의 결과값이 반환되고, 실패한 프라미스는 예외를 발생시킨다
- `$unwrap`이 `false`
    - 모든 프라미스가 완료될 때까지 기다린다
    - 하지만, 결과를 '풀어내지(unwrapping)' 않는다.
    - 프라미스가 어떤 최종 상태에 도달했는지만 확인하고 **결과값을 반환하거나 예외를 발생시키지 않는다**. 단지 모든 요청이 처리될 때까지의 완료 상태만을 확인하는 것
        - 즉, 각 요청이 처리될 때까지 기다리지만, 그 결과는 반환되지 않는다
        - 각 프라미스의 'fulfilled' 또는 'rejected' 상태를 확인하지만, 이에 따른 구체적인 값이나 예외 정보는 반환하지 않는다
- 대부분의 경우에는 `unwrap`을 `true`로 설정하여 프라미스의 결과를 직접 처리한다. Guzzle의 HTTP 요청이 성공적으로 완료되었을 때 해당 결과를 바로 사용하고 싶을 때 유용하다

#### `Promise\settle($promises)->wait($unwrap)` 후 처리

- `Promise\settle($promises)->wait(false)`를 통해 `$promises`의 프라미스들은 'settled' 상태(즉, 'fulfilled' 또는 'rejected' 상태)에 도달했다
- 이때 `$promises`의 각 `$promise->wait()` 시 추가적인 대기 시간이 발생하지 않는다
    - 이미 'settled' 상태에 있는 프라미스에 대해 `wait()`를 호출하면, 해당 프라미스의 최종 결과를 즉시 반환하고, 추가적인 대기 시간은 발생하지 않는다
    - 'fulfilled' 상태의 프라미스는 해당 값이 반환되고, 'rejected' 상태의 프라미스는 예외가 발생한다
- `wait(false)`는 각 프라미스가 완료될 때까지 기다리지만, 프라미스의 결과를 'unwrap'하지 않으므로, 각 개별 프라미스에서 `wait()`를 호출하는 것은 안전하며, 이는 해당 프라미스의 최종 결과값을 동기적으로 반환한다

```php
// 모든 프라미스가 완료될 때까지 기다린다
Promise\settle($promises)->wait(false);

// 각 프라미스의 결과를 추출
foreach ($promises as $key => $promise) {
    try {
        $result = $promise->wait();  // 여기서 추가 대기 없이 결과를 즉시 얻는다
        // 결과 처리
    } catch (Exception $e) {
        // 예외 처리
    }
}
```

이 방법을 통해 각 프라미스의 결과값을 효과적으로 처리할 수 있으며, 이미 'settled' 상태인 프라미스에 대해 `wait()`를 호출하는 것은 추가적인 대기 시간을 발생시키지 않습니다.

### `settle`과 `unwrap`의 차이점

- `unwrap`: 프라미스의 최종 결과값을 가져오는 과정을 의미
    - 모든 프라미스가 성공해야만 결과를 반환
    - 하나라도 실패하면 즉시 예외를 발생
    - 즉, 성공한 프라미스의 결과만을 반환
    - 만약 어떤 프라미스 A가 다른 프라미스 B로 'resolved'된 경우, 프라미스 A를 'unwrap'하면 프라미스 B의 결과값을 기다린다. 즉, 최종 결과가 나올 때까지 연쇄적으로 대기한다. 가령 `promiseA->resolve(promiseB)`와 같이 프라미스 A가 프라미스 B로 'resolved'된 경우, `promiseA->wait()`를 호출하면 프라미스 B의 결과값을 대기한다
- `settle`: 프라미스가 'fulfilled' 또는 'rejected' 상태로 'settled' 되기만을 대기
    - 각 프라미스의 성공 또는 실패에 관계없이 모든 프라미스가 완료되기를 기다린다
    - 그저 프라미스가 처리될 때까지 기다리는 것. 따라서,
        - 결과값을 'unwrap'하지 않는다
        - 결과값 자체를 직접 반환하지 않는다
    - 각 프라미스의 'fulfilled' 또는 'rejected' **상태에 대한 정보는 반환**되지만, 이것은 프라미스의 "결과값"을 반환한다는 것과는 다르다

### 예시

```php
$promises = [
    'image' => $client->getAsync('/image'),
    'data'  => $client->getAsync('/data')
];

$results = Promise\settle($promises)->wait(false);

foreach ($results as $key => $result) {
    if ($result['state'] === 'fulfilled') {
        echo $key . ' 결과: ' . $result['value']->getBody();
    } else {
        echo $key . ' 실패: ' . $result['reason']->getMessage();
    }
}

```

## `Resolved`

- `Resolved`는 `Promise` 상태가 둘 중 하나로 결정된 것을 의미
    - 'fulfilled'(성공적으로 완료)
    - 또는 'rejected'(실패)
- 하지만,`settle($promises)->wait(false)`에서 'resolved'는 단순히 모든 프라미스가 어떤 최종 상태(성공 또는 실패)에 도달했다는 것을 의미
    - 즉, `settle($promises)->wait(false)`는 모든 프라미스가 'settled' 상태, 즉 완료된 상태임을 나타낸다
    - 이 때 각 프라미스는 'fulfilled' 또는 'rejected' 상태일 수 있다
    - 따라서, 이 경우에는 각 프라미스의 최종 상태(성공 또는 실패 여부)를 나타내는 더 포괄적인 용어를 사용하는 것이 적절합니다.
