# cURL timeout

## `CURLOPT_CONNECTTIMEOUT`

```php
curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, 1);
```

이 설정은 연결 수립 단계에서의 시간 제한을 1초로 설정합니다.

적용 구간:
- DNS 조회
- TCP 연결 수립 (TCP 3-way handshake)
- SSL/TLS 핸드셰이크 (HTTPS의 경우)

동작 과정:
a) cURL이 호스트 이름을 IP 주소로 변환 (DNS 조회)
b) 대상 서버와 TCP 연결 시도
c) HTTPS인 경우 SSL/TLS 핸드셰이크 수행
d) 이 모든 과정이 1초 이내에 완료되어야 함
e) 1초가 지나면 연결 시도를 중단하고 오류 반환

## `CURLOPT_TIMEOUT`

```php
curl_setopt($ch, CURLOPT_TIMEOUT, 10);
```

이 설정은 전체 cURL 작업의 최대 허용 시간을 10초로 설정합니다.

적용 구간:
- 연결 수립 (위의 `CURLOPT_CONNECTTIMEOUT` 포함)
- 요청 전송
- 서버의 응답 대기
- 응답 데이터 수신

동작 과정:
a) 연결 수립 시작 (`CURLOPT_CONNECTTIMEOUT의` 1초 포함)
b) HTTP 요청 헤더 및 본문 전송
c) 서버의 응답 대기
d) 응답 헤더 및 본문 수신
e) 전체 과정이 10초 이내에 완료되어야 함
f) 10초가 지나면 작업을 중단하고 오류 반환

시간 경과에 따른 동작:

1. 0-1초: 연결 수립 시도 (`CURLOPT_CONNECTTIMEOUT`)
   - 1초 내에 연결되지 않으면 즉시 실패
   - 1초 내에 연결되면 다음 단계로 진행

2. 1-10초: 요청 전송 및 응답 수신 (`CURLOPT_TIMEOUT`)
   - 연결 후 남은 시간(최대 9초) 동안 요청 전송 및 응답 수신
   - 10초 총 시간 제한에 도달하면 작업 중단

주의사항:
- `CURLOPT_TIMEOUT`은 항상 `CURLOPT_CONNECTTIMEOUT`보다 크거나 같아야 합니다.
- 네트워크 상태, 서버 부하 등에 따라 실제 동작 시간은 다를 수 있습니다.
- 대용량 데이터 전송 시 CURLOPT_TIMEOUT을 더 길게 설정해야 할 수 있습니다.
