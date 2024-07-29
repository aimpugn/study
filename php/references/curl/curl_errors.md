# curl errors

- [curl errors](#curl-errors)
    - [libcurl error codes](#libcurl-error-codes)
        - [CURLE\_OPERATION\_TIMEDOUT (28)](#curle_operation_timedout-28)
        - [CURLE\_SSL\_CONNECT\_ERROR (35)](#curle_ssl_connect_error-35)
        - [SSL/TLS 핸드셰이크 단계에서 발생할 수 있는 오류](#ssltls-핸드셰이크-단계에서-발생할-수-있는-오류)
    - [error messages](#error-messages)
        - [Operation timed out after 0 milliseconds with 0 out of 0 bytes received](#operation-timed-out-after-0-milliseconds-with-0-out-of-0-bytes-received)
            - [가능한 원인](#가능한-원인)
        - [Resolving timed out after N milliseconds](#resolving-timed-out-after-n-milliseconds)
            - [가능한 원인](#가능한-원인-1)
        - [SSL connection timeout](#ssl-connection-timeout)
            - [가능한 원인](#가능한-원인-2)

## [libcurl error codes](https://curl.se/libcurl/c/libcurl-errors.html)

### CURLE_OPERATION_TIMEDOUT (28)

> Operation timeout. The specified time-out period was reached according to the conditions.

cURL에서 요청의 실행 중에 설정된 시간 제한에 도달했을 때 발생한다.

1. 연결 시도: cURL이 대상 서버에 연결을 시도할 때, 이 단계에서의 타임아웃은 DNS 조회, TCP 핸드셰이크 등 서버에 연결을 시도하는 데 걸리는 시간이 너무 길 때 발생한다. `CURLOPT_CONNECTTIMEOUT` 옵션을 통해 이 시간을 설정할 수 있다.

2. 데이터 전송: 서버에 연결된 후 실제 데이터 전송 단계에서의 타임아웃. 이는 데이터를 보내고 받는 과정에서 지정된 시간 동안 어떠한 진전도 없을 때 발생한다. `CURLOPT_TIMEOUT` 옵션으로 전체 작업에 대한 시간 제한을 설정할 수 있다.

3. 응답 대기: 서버로부터의 응답을 기다리는 동안의 타임아웃도 포함될 수 있다. 요청을 보낸 후 서버로부터의 응답이 설정된 시간 내에 도착하지 않으면 타임아웃이 발생한다.

### CURLE_SSL_CONNECT_ERROR (35)

> A problem occurred somewhere in the SSL/TLS handshake. You really want the error buffer and read the message there as it pinpoints the problem slightly more. Could be certificates (file formats, paths, permissions), passwords, and others.

`CURLE_SSL_CONNECT_ERROR (35)` 오류는 cURL 통신 과정 중 SSL/TLS 핸드셰이크 단계에서 문제가 발생했을 때 나타납니다. 이 오류는 SSL/TLS 연결 설정 과정에서 다양한 이유로 발생할 수 있는데, 주요 원인은 다음과 같습니다:

### SSL/TLS 핸드셰이크 단계에서 발생할 수 있는 오류

- 인증서 문제:
    - 가장 흔한 원인 중 하나
    - 클라이언트나 서버 측에서 사용하는 인증서가 유효하지 않거나,
    - 신뢰할 수 없는 인증 기관에 의해 발급된 경우,
    - 또는 인증서가 만료된 경우
- 인증서 형식 및 경로 문제
    - 인증서의 파일 형식이 잘못되었거나,
    - 파일 경로에 접근할 수 없는 경우 (예: 잘못된 경로, 파일 권한 문제 등)
- SSL/TLS 프로토콜 불일치
    - 클라이언트와 서버 간에 사용하는 SSL/TLS 버전이 서로 호환되지 않는 경우
    - 예를 들어, 서버가 TLS 1.2만을 지원하는데 클라이언트가 이전 버전의 SSL 프로토콜을 사용하는 경우 등.
- 암호화 알고리즘 불일치
    - 클라이언트와 서버가 지원하는 암호화 알고리즘이 서로 맞지 않을 때 발생
    - SSL/TLS 핸드셰이크 과정에서는 서로 지원하는 알고리즘을 협상하는데, 이 과정에서 일치하는 알고리즘을 찾지 못하면 오류가 발생할 수 있다
- SSL/TLS 라이브러리 설정 오류
    - 클라이언트 측에서 사용하는 SSL/TLS 라이브러리의 설정 문제로 인해 핸드셰이크가 실패할 수도 있다
    - 이는 구성 오류나 라이브러리의 버그로 인한 것일 수 있다.

## error messages

### Operation timed out after 0 milliseconds with 0 out of 0 bytes received

cURL 요청이 지정된 시간 내에 완료되지 않았음을 나타냅니다. 구체적으로는 cURL이 서버로부터 응답을 전혀 받지 못했을 때 발생합니다.

#### 가능한 원인

1. 타임아웃 설정 오류: cURL의 타임아웃 값이 0으로 설정되었거나, 너무 낮게 설정되어 있을 수 있습니다.
2. 서버 과부하: 요청을 처리하는 서버가 과부하 상태에 있어 응답을 제대로 하지 못하는 경우입니다.
3. 네트워크 연결 문제: 클라이언트와 서버 간의 네트워크 연결에 문제가 있을 수 있습니다.

### Resolving timed out after N milliseconds

이 에러는 DNS 조회 과정에서 지정된 시간 내에 서버의 IP 주소를 확인하지 못했음을 의미합니다.

#### 가능한 원인

1. DNS 서버 응답 지연: DNS 서버가 주소 조회에 응답하는 데 시간이 오래 걸릴 수 있습니다.
2. 네트워크 문제: 네트워크 연결이 불안정하거나 지연되는 문제가 있을 수 있습니다.
3. DNS 조회 타임아웃 설정: DNS 조회에 대한 타임아웃 설정이 너무 짧게 설정되어 있을 수 있습니다.

### SSL connection timeout

이 에러는 SSL/TLS 핸드셰이크 과정 중 타임아웃이 발생했음을 나타내며, cURL이 안전한 SSL 연결을 설정하지 못했음을 의미한다.
SSL/TLS 핸드셰이크는 주로 서버에서 처리되는 작업이며, 이 과정에서 목적지 서버의 자원과 설정이 중요한 역할을 한다.

#### 가능한 원인

1. SSL/TLS 설정 문제: 서버의 SSL/TLS 설정에 문제가 있을 수 있다.
2. 서버 과부하: SSL 핸드셰이크를 처리하는 목적지 서버의 과부하 상태가 원인일 수 있다.
3. 네트워크 지연: 네트워크 지연으로 인해 SSL 핸드셰이크가 완료되지 못했을 수 있다.
4. SSL 연결 타임아웃 설정: SSL 연결에 대한 타임아웃 설정이 너무 짧을 수 있다.
