# curl

## curl?

cURL(Client URL) 이란 커맨드 라인에서 URL을 다루는 프로그램이다. cURL은 URL을 통해 데이터를 전송하거나 받을 수 있다. cURL은 다양한 프로토콜을 지원한다. HTTP, HTTPS, FTP, FTPS, SCP, SFTP, TFTP, LDAP, DAP, DICT, TELNET, FILE, IMAP, POP3, SMTP, RTMP, RTSP, SMB, SMBS, MQTT, GOPHER, LDAP, LDAPS, MQTT, SCP, SFTP 등을 지원한다.

## PHP에서 cURL 호출 과정

### 1. cURL 초기화 (`curl_init()`)

- cURL이 네트워크 통신을 위해 필요한 운영 체제의 자원과 설정에 접근할 수 있게 해준다.
- 내부 작업: `curl_init()` 함수는 cURL 세션을 시작한다. 내부적으로 cURL 라이브러리를 초기화하고, 후속 cURL 작업에 사용되는 새로운 cURL 핸들을 생성한다.
- 리소스 할당: cURL 핸들에 필요한 리소스가 할당되며, 이는 네트워크 연결 및 HTTP 요청 처리에 필요한 구성 요소를 포함한다.
- 옵션 설정 준비: 이 단계에서는 아직 실제 네트워크 요청을 보내지 않고, 요청을 위한 준비만 수행한다.
- `curl_init()` 함수가 호출될 때 할당되는 필수적인 리소스
    1. cURL 핸들 생성: `curl_init()`은 cURL 라이브러리에 대한 핸들을 생성한다. 이 핸들은 이후 모든 cURL 작업의 기반으로 사용되며, cURL 라이브러리와의 상호작용을 관리하는 데 중요한 역할을 한다.
    2. 기본 설정 초기화: cURL 핸들은 기본적인 설정과 상태로 초기화 된다. 이는 cURL 라이브러리의 기본 동작을 정의한다.
    3. 메모리 할당: cURL 핸들에 필요한 메모리가 할당된다. 이 메모리는 후속 cURL 작업(옵션 설정, 요청 실행 등)에서 사용된다.
    4. 네트워크 스택 초기화: cURL은 네트워크 통신을 위한 내부 네트워크 스택을 초기화한다. 이는 TCP/IP 연결 및 데이터 전송에 사용된다.
    5. 기본 네트워크 매개변수 설정: 기본적인 네트워크 매개변수가 설정된다. 이에는 기본 타임아웃 값, 네트워크 프로토콜 선택(예: HTTP, HTTPS) 등이 포함될 수 있다.
    6. 내부 데이터 구조 생성: cURL은 내부적으로 요청과 응답을 관리하기 위한 다양한 데이터 구조를 생성한다. 이 구조들은 요청의 전송, 응답의 수신, 헤더 처리 등에 사용된다.
    7. 에러 핸들링 메커니즘 설정: 요청 중 발생할 수 있는 에러를 감지하고 처리하기 위한 기본 메커니즘이 설정된다.

> **네트워크 스택?**
>
> 네트워크 연결을 관리하고 데이터를 전송 및 수신하는 데 사용되는 소프트웨어의 집합을 의미
>
> 1. TCP/IP 프로토콜 스택: 네트워크 스택은 TCP/IP 프로토콜 스택을 기반으로 한다. TCP/IP 스택은 네트워크 통신을 가능하게 하는 여러 계층(예: 네트워크 계층, 전송 계층)의 프로토콜로 구성된다.
> 2. 소프트웨어 구현: 이 스택은 *운영 체제에 의해* 소프트웨어적으로 구현된다. Linux, Windows, macOS 등 각 운영 체제는 자체적인 네트워크 스택 구현을 가지고 있으며, 이를 통해 네트워크 연결과 데이터 전송을 관리한다.
> 3. 데이터 전송 및 수신 관리: 네트워크 스택은 데이터 패킷을 만들고, 주소(네트워크 통신 과정에서 사용되는 IP 주소와 포트 번호)를 할당하고, 데이터를 전송하고 수신하는 전체 과정을 관리한다.
>
> **cURL에서의 네트워크 스택 초기화**
>
> - cURL과 네트워크 스택: cURL 라이브러리는 네트워크 요청을 수행하기 위해 *운영 체제의 네트워크 스택을 사용*한다. `curl_init()` 함수를 호출할 때, cURL은 내부적으로 이 네트워크 스택과 연결되고, 필요한 네트워크 관련 설정을 초기화한다.
> - TCP/IP 연결 준비: 이 초기화 과정에서 cURL은 TCP 연결을 준비하고, 필요한 경우 SSL/TLS 계층(HTTPS 요청의 경우)도 설정한다. 이는 서버와의 연결을 위한 준비 단계다.

### 2. 옵션 설정 (`curl_setopt()`)

- 핸들 구성: `curl_setopt()` 함수는 cURL 핸들에 다양한 옵션을 설정합니다. 이 옵션에는 URL, HTTP 메소드, 헤더, 타임아웃, 쿠키, SSL 설정 등이 포함될 수 있습니다.
- 네트워크 구성: 프록시 설정, 인증 방법, 사용할 네트워크 인터페이스 등 네트워크 관련 구성이 이루어집니다.
- 요청 페이로드 준비: POST 또는 PUT 요청의 경우, 요청 본문(body)이 이 단계에서 준비됩니다.

### 3. 요청 실행 (`curl_exec()`)

- DNS 조회: 지정된 URL의 호스트 이름을 IP 주소로 변환하기 위해 DNS 조회가 수행된다. 이 과정은 네트워크 상의 DNS 서버를 통해 이루어진다.
- TCP 연결: DNS 조회 후 얻은 IP 주소를 사용하여 서버와의 TCP 연결을 시도한다. 이는 TCP 3-way handshake 과정을 포함한다.
- SSL/TLS 핸드셰이크 (HTTPS 요청의 경우): 암호화된 연결을 위해 SSL 또는 TLS 핸드셰이크가 수행된다. 이 과정에서 서버의 인증서가 검증되며, 암호화된 채널이 설정된다.
- HTTP 요청 전송: HTTP 프로토콜에 따라 구성된 요청 메시지가 서버로 전송된다. 헤더, 요청 방식, 본문 데이터 등이 포함된다.
- 응답 대기 및 처리: 서버로부터의 응답을 기다리고, 도착하면 응답 데이터를 수신한다. 헤더, 상태 코드, 응답 본문 등이 포함된 응답을 처리한다.

### 4. 응답 처리 및 cURL 세션 종료

- 응답 분석: 응답 데이터는 PHP 스크립트에서 분석된다. HTTP 상태 코드, 응답 헤더, 본문 등이 사용자에게 제공되거나 후속 처리를 위해 사용된다.
- 오류 처리: 네트워크 오류, 타임아웃, HTTP 오류 등이 발생했을 경우, 이를 처리하는 코드가 실행된다.
- 리소스 해제: `curl_close()` 함수를 호출하여 cURL 핸들과 관련된 모든 리소스를 해제하고, 네트워크 연결을 종료한다.

## debug curl

```php
public function debugCURL()
{
    print_r('debugCURL' . PHP_EOL);
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, 'https://google.com');
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, 0);
    curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, 5);
    curl_setopt($ch, CURLOPT_TIMEOUT, 5);
    curl_setopt($ch, CURLINFO_HEADER_OUT, 1); // cURL 실행 후 요청에 실제로 보낸 헤더 정보를 추출하기 위해 
                                              // curl_getinfo 함수와 함께 사용된다. 이는 디버깅 정보를 출력하는 것이 아니라, 
                                              // 프로그램 내에서 요청 헤더 정보를 검사할 목적으로 사용된다.
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    curl_setopt($ch, CURLOPT_VERBOSE, true);

    // 디버그 정보를 메모리에 저장하기 위한 스트림 열기
    $verbose = fopen('php://temp', 'w+');
    curl_setopt($ch, CURLOPT_STDERR, $verbose);

    $response = curl_exec($ch);
    // 요청이 실패했는지 확인
    if ($response === false) {
        printf("cURL error (%d): %s\n", curl_errno($ch), htmlspecialchars(curl_error($ch)));
    }

    // 디버그 출력을 위해 스트림의 포인터를 처음으로 되돌림
    rewind($verbose);
    $verboseLog = stream_get_contents($verbose);

    fclose($verbose);
    curl_close($ch);

    print_r($verboseLog);
    print_r('DONE' . PHP_EOL);
}
```
