# Guzzle

- [Guzzle](#guzzle)
    - [Guzzle](#guzzle-1)
    - [Response 사용](#response-사용)
        - [`getBody`](#getbody)
            - [스트림(Stream)과 스트림의 콘텐츠(Stream's Contents)](#스트림stream과-스트림의-콘텐츠streams-contents)
                - [스트림(Stream)](#스트림stream)
                - [스트림의 콘텐츠(Stream's Contents)](#스트림의-콘텐츠streams-contents)
            - [`getBody()` 사용 예제](#getbody-사용-예제)
            - [`getBody()`의 장점](#getbody의-장점)
        - [`getBody()->getContents()`](#getbody-getcontents)
            - [`getBody()->getContents()` 사용 예제](#getbody-getcontents-사용-예제)

## Guzzle

## Response 사용

```php
$code = $response->getStatusCode(); // 200
$reason = $response->getReasonPhrase(); // OK

// Check if a header exists.
if ($response->hasHeader('Content-Length')) {
    echo "It exists";
}

// Get a header from the response.
echo $response->getHeader('Content-Length')[0];

// Get all of the response headers.
foreach ($response->getHeaders() as $name => $values) {
    echo $name . ': ' . implode(', ', $values) . "\r\n";
}
```

### `getBody`

- HTTP 응답의 본문(body)을 나타내는 `스트림 객체`를 반환하고, 이 스트림 객체는 응답 데이터를 순차적으로 읽는 데 사용
- 즉, `getBody`가 반환하는 것은 실제 데이터 자체가 아니라, 그 **데이터에 접근할 수 있는 스트림 인터페이스**
- 응답의 바디는 `getBody` 메서드를 사용해서 가져올 수 있다
- 바디는 여러 경우로 사용될 수 있다
    - 문자열
    - 문자열로 캐스팅
    - stream 같은 오브젝트

#### 스트림(Stream)과 스트림의 콘텐츠(Stream's Contents)

##### 스트림(Stream)

- 스트림?
    - 스트림은 데이터의 연속적인 흐름을 나타낸다
    - 스트림을 사용하는 주된 목적은 데이터의 효율적인 처리에 있고, 특히 대용량 데이터를 처리할 때 유용하다
    - 스트림을 사용하면 전체 데이터를 한꺼번에 메모리에 로드하지 않고, 필요한 부분만 순차적으로 읽을 수 있다
- 스트림의 역할
    - 스트림은 데이터를 순차적으로 처리할 수 있는 인터페이스를 제공
    - 예를 들어, 파일을 읽거나 네트워크 응답을 받는 경우, 전체 내용을 한 번에 메모리에 로드하는 대신 스트림을 통해 조금씩 읽어들일 수 있다

##### 스트림의 콘텐츠(Stream's Contents)

- 스트림의 콘텐츠 정의
    - 스트림의 콘텐츠란 **스트림을 통해 전달되는 실제 데이터**를 의미
    - 이 데이터는 텍스트, 이미지, 비디오 등 다양한 형태일 수 있다
- 콘텐츠의 처리 방식
    - `getBody()->getContents()`를 사용하면 스트림의 현재 위치부터 끝까지의 모든 콘텐츠를 한 번에 읽어들인다
    - 이 방법은 스트림의 전체 데이터를 필요로 할 때 유용하지만, 데이터 크기가 클 경우 메모리 사용량이 증가할 수 있다

#### `getBody()` 사용 예제

```php
$body = $response->getBody();
// Implicitly cast the body to a string and echo it
echo $body;
// Explicitly cast the body to a string
$stringBody = (string) $body;
// Read 10 bytes from the body
$tenBytes = $body->read(10);
// Read the remaining contents of the body as a string
$remainingBytes = $body->getContents();
```

```php
$client = new GuzzleHttp\Client();
$response = $client->request('GET', 'http://example.com');

// `getBody()`는 응답의 본문을 스트림으로 반환
$body = $response->getBody();

// 스트림에서 조금씩 데이터를 읽기. 메모리 사용을 최소화하면서 데이터를 처리할 수 있게 해준다
while (!$body->eof()) {
    // read() 메서드를 사용하여 스트림에서 1024바이트 단위로 데이터를 읽는다
    echo $body->read(1024);
}
```

#### `getBody()`의 장점

- 스트림 객체이므로, 개발자는 데이터를 보다 효율적으로 관리하고 제어할 수 있다
- 대용량 데이터를 처리할 때 필수적일 수 있습니다.
- 전체 데이터를 한 번에 메모리에 로드하는 것이 아니라, 필요한 부분만 읽거나, 데이터를 순차적으로 처리할 수 있다

### `getBody()->getContents()`

- 스트림의 현재 위치에서부터 남아 있는 모든 데이터를 읽고 문자열로 반환한다
- 만약 스트림이 이미 어느 정도 읽혀진 상태라면, `getContents()`는 **이미 읽힌 부분을 제외한 나머지 부분만 반환**한다.
- 전체 내용이 필요하고 데이터 크기가 크지 않다면 `getBody()->getContents()`를 사용하는 것이 좋지만, 스트림의 크기가 클 경우 메모리 문제가 발생할 수 있다.
- 처음부터 끝까지 모든 내용을 읽으려면
    1. 스트림의 포인터를 처음으로 리셋하거나
    2. `__toString()` 메서드를 사용해야 한다

#### `getBody()->getContents()` 사용 예제

- `getBody()->getContents()` 메서드는 스트림의 현재 위치부터 끝까지 모든 데이터를 문자열로 읽어 반환한다
- 이는 스트림 전체의 데이터를 한 번에 가져오고 싶을 때 사용. 스트림의 전체 내용을 바로 사용하고자 할 때 편리하다

```php
$client = new GuzzleHttp\Client();
$response = $client->request('GET', 'http://example.com');

$body = $response->getBody();

// 스트림의 현재 위치부터 끝까지 모든 데이터를 읽기
$contents = $body->getContents();
echo $contents;
```
