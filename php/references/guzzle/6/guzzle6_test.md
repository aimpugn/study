# Guzzle6 Test

- [Guzzle6 Test](#guzzle6-test)
    - [핸들러 모킹](#핸들러-모킹)

## 핸들러 모킹

- HTTP clients 테스트할 때, 성공 응답, 에러, 특정 순서로 정렬된 응답 등 어떤 시나리오 시뮬레이션 필요할 수 있다
- 유닛 테스트는 예측 가능하고, 부트스트랩 하기 쉬우며, 빨라야 하므로, 실제 원격 API를 호출하는 것은 test smell이다
- Guzzle은 큐에서 반환 값을 꺼내어 HTTP 요청을 응답 또는 예외로 처리할 수 있는 모의(mock) 핸들러를 제공

```php
use GuzzleHttp\Client;
use GuzzleHttp\Handler\MockHandler;
use GuzzleHttp\HandlerStack;
use GuzzleHttp\Psr7\Response;
use GuzzleHttp\Psr7\Request;
use GuzzleHttp\Exception\RequestException;

// Create a mock and queue two responses.
$mock = new MockHandler([
    new Response(200, ['X-Foo' => 'Bar'], 'Hello, World'),
    new Response(202, ['Content-Length' => 0]),
    new RequestException('Error Communicating with Server', new Request('GET', 'test'))
]);

$handlerStack = HandlerStack::create($mock);
$client = new Client(['handler' => $handlerStack]);

// The first request is intercepted with the first response.
$response = $client->request('GET', '/');
echo $response->getStatusCode();
//> 200
echo $response->getBody();
//> Hello, World
// The second request is intercepted with the second response.
echo $client->request('GET', '/')->getStatusCode();
//> 202

// Reset the queue and queue up a new response
$mock->reset();
$mock->append(new Response(201));

// As the mock was reset, the new response is the 201 CREATED,
// instead of the previously queued RequestException
echo $client->request('GET', '/')->getStatusCode();
//> 201
```
