# yield

## yield란?

- PHP 5.6에서 `yield`는 `제너레이터`를 생성하는 데 사용되는 키워드
- `제너레이터`는 함수가 반복 가능한 값을 순차적으로 생성할 수 있게 해주는 특별한 형태의 함수
- `yield`를 사용하면 함수의 실행을 일시 중지하고, 값을 반환한 후, 다음 호출에서 중지된 지점부터 실행을 재개한다

## `yield`와 비동기 처리?

- `yield`는 비동기 프로그래밍과 다르다
- 비동기 프로그래밍은 여러 작업이 동시에 진행되는 것을 의미하지만, `yield`를 사용하는 제너레이터는 **동기적으로 실행**되며, 각 `yield` 표현이 나타날 때마다 실행을 일시 중지한다

```php
<?php

function request($url) {
  // HTTP 요청을 수행하고, 결과를 반환합니다.
  // 동기적으로 실행되며, URL의 데이터가 완전히 로드될 때까지 함수는 여기에서 일시 중지됩니다.
  $response = file_get_contents($url);
  return $response;
}

// 'request' 함수는 단일 HTTP 요청을 수행하며, 각 요청은 순차적으로 실행됩니다.
// 이는 각 'yield' 호출이 완료될 때까지 다음 'yield' 호출이 시작되지 않음을 의미합니다.

$iterator = function() {
  // 'yield'를 사용하여 첫 번째 HTTP 요청을 시작합니다.
  // 이 호출이 완료될 때까지 다음 'yield'는 실행되지 않습니다.
  yield request("https://www.google.com");

  // 첫 번째 요청이 완료된 후, 두 번째 HTTP 요청을 시작합니다.
  yield request("https://www.naver.com");

  // 두 번째 요청이 완료된 후, 세 번째 HTTP 요청을 시작합니다.
  yield request("https://www.daum.net");
};

// 'foreach' 루프를 사용하여 제너레이터의 각 'yield' 결과를 순차적으로 처리합니다.
// 각 'yield'는 해당 HTTP 요청을 수행하고 결과를 반환합니다.
foreach ($iterator as $response) {
  // 여기에서 각 HTTP 요청의 응답을 출력합니다.
  echo $response;
}

// 'foreach' 루프가 종료되면, 모든 HTTP 요청이 순차적으로 처리되었음을 의미합니다.

?>
```
