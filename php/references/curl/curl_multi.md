# curl_multi

- [curl\_multi](#curl_multi)
    - [비동기 실행 과정](#비동기-실행-과정)
    - [`curl_multi_init`](#curl_multi_init)
    - [`curl_multi_add_handle`](#curl_multi_add_handle)
    - [`curl_multi_exec`](#curl_multi_exec)
        - [`$still_running` 매개변수](#still_running-매개변수)
    - [상수들](#상수들)
        - [`CURLM_CALL_MULTI_PERFORM`](#curlm_call_multi_perform)
        - [`CURLM_OK`](#curlm_ok)
        - [기타 cURL 멀티 관련 상수](#기타-curl-멀티-관련-상수)

## 비동기 실행 과정

```php
$mh = curl_multi_init();
// curl_init 및 curl_setopt으로 여러 cURL 핸들을 설정하고 $mh에 추가

$still_running = null;
do {
    $mrc = curl_multi_exec($mh, $still_running);
    //                          ^^^^^^^^^^^^^^
    //                        `curl_multi_exec`는 처음 호출될 때 
    //                        `$still_running`을 멀티 핸들에 있는 활성 cURL 핸들의 수로 설정
} while ($mrc == CURLM_CALL_MULTI_PERFORM);
//^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
// 현재 수행할 작업이 더 있음을 나타내는 `CURLM_CALL_MULTI_PERFORM`를 
// 반환하는 동안 계속 실행

while ($still_running && $mrc == CURLM_OK) {
    // ... 중략 ...
}
```

1. cURL 요청 실행
    - `curl_multi_exec` 함수를 호출하여 멀티 핸들에 추가된 모든 cURL 요청을 실행한다
    - 이 함수는 비동기적으로 요청을 처리하므로, PHP 스크립트는 각 요청이 완료될 때까지 기다리지 않고 즉시 다음 코드로 넘어간다
2. `$still_running` 카운터
    - `$still_running` 변수는 아직 처리 중인 요청의 수를 나타낸다
    - 이 변수는 `curl_multi_exec` 함수가 호출될 때마다 업데이트된다
3. 요청 상태 확인
    - `curl_multi_exec`는 모든 요청이 완료될 때까지 반복적으로 호출된다
    - 각 호출은 아직 완료되지 않은 요청을 처리하고 `$still_running` 변수를 업데이트한다
4. 완료된 요청 처리
    - 요청이 완료되면 `$still_running` 카운터가 감소한다
    - 모든 요청이 완료되면 `$still_running`은 0이 된다

## `curl_multi_init`

- `curl_multi_init` 함수를 호출하여 cURL 멀티 핸들을 생성
- 이 멀티 핸들은 여러 cURL 핸들을 동시에 관리한다

## `curl_multi_add_handle`

- `curl_multi_add_handle` 함수를 사용하여 멀티 핸들에 개별 cURL 요청 핸들들을 추가
- 각 cURL 핸들은 개별적인 HTTP 요청을 나타낸다

## `curl_multi_exec`

- cURL 멀티 핸들에 등록된 모든 cURL 핸들의 처리를 시작하거나 계속 처리한다
- `curl_multi_exec` 함수가 호출될 때마다, `$still_running` 매개변수는 처리 중인 cURL 핸들의 수로 업데이트된다

### `$still_running` 매개변수

- 함수에 의해 참조로 전달되는 정수형 변수. 함수 호출 시마다 변할 수 있다
- 초기에 `$still_running`은 멀티 핸들에 추가된 cURL 핸들의 총 수로 설정된다
- 함수 호출이 반환될 때, 이 변수는 처리 중인 cURL 핸들의 수를 나타낸다
- 멀티 핸들에 있는 각 cURL 요청이 완료될 때마다 `$still_running`의 값은 감소하고, 모든 요청이 완료되면 0이 된다
- 이 변수는 멀티 핸들에 있는 각 cURL 요청의 상태를 추적하는 데 사용된다

## 상수들

### `CURLM_CALL_MULTI_PERFORM`

- `curl_multi_exec` 함수가 현재 수행할 작업이 더 있음을 나타낸다
- 즉, cURL 멀티 핸들이 여전히 데이터를 전송하거나 수신 중이며, 추가적인 처리가 필요함을 의미
- `CURLM_CALL_MULTI_PERFORM`가 반환되면, 작업이 완료될 때까지 `curl_multi_exec`를 계속 호출해야 한다. 멀티 핸들이 관련된 모든 작업을 완료할 때까지 반복적으로 처리되어야 함을 나타낸다

### `CURLM_OK`

- `curl_multi_exec` 함수가 성공적으로 수행되었음을 나타낸다
    - 현재 수행 중인 작업이 없거나,
    - 모든 작업이 정상적으로 진행 중임을 의미
- 이 상수가 반환될 때, 현재 멀티 핸들에서 처리 중인 요청이 없거나, **아직 처리해야 할 요청이 남아있을 수** 있다. 따라서, `$still_running` 변수를 확인하여 아직 완료되지 않은 요청이 있는지 확인해야 한다.
- `CURLM_OK`는 `while` 루프와 함께 사용되어, 모든 요청이 완료되거나 오류가 발생할 때까지 `curl_multi_exec`를 주기적으로 호출한다
    - `curl_multi_exec`는 멀티 핸들의 상태를 지속적으로 체크하고 업데이트한다. `CURLM_OK`가 반환되더라도, 모든 요청이 완료될 때까지 이 함수를 반복적으로 호출해야 한다.
    - 즉, 모든 요청이 완료될 때까지 (`$still_running`이 0이 될 때까지) `curl_multi_exec`를 호출해야 한다

> 왜 `CURLM_OK`임에도 아직 처리해야 할 요청이 남아있을 수 있을까?
>
> 멀티 핸들의 비동기적 특성 때문. `curl_multi_exec` 함수를 사용할 때, 여러 cURL 핸들이 동시에 비동기적으로 처리되고 있는 중이다. 이때 `CURLM_OK`는 `curl_multi_exec` 함수 자체가 성공적으로 호출되었음을 나타내지만, 이것만으로는 모든 개별 cURL 요청이 완료되었는지 여부를 나타내지는 않는다.

### 기타 cURL 멀티 관련 상수

- `CURLM_BAD_HANDLE`: 잘못된 멀티 핸들이 제공되었을 때 반환된다
- `CURLM_BAD_EASY_HANDLE`: 멀티 핸들에 잘못된 쉬운 핸들(easy handle)이 추가됐을 때 반환된다
- `CURLM_OUT_OF_MEMORY`: 메모리 할당에 실패했을 때 반환된다
- `CURLM_INTERNAL_ERROR`: 내부적인 오류가 발생했을 때 반환된다
