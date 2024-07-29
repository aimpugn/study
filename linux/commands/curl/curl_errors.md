# cURL Errors

- [cURL Errors](#curl-errors)
    - [curl: (35) Recv failure: Connection reset by peer](#curl-35-recv-failure-connection-reset-by-peer)
    - [cURL error 56, errno 104](#curl-error-56-errno-104)

## curl: (35) Recv failure: Connection reset by peer

```shell
curl -v --tls-max 1.0 https://uas.teledit.com:443
*   Trying 150.242.135.225:443...
* Connected to uas.teledit.com (150.242.135.225) port 443 (\#0) # 서버에 성공적으로 연결되었음을 의미
* ALPN: offers h2,http/1.1 # `ALPN`(Application-Layer Protocol Negotiation)은 
                           # 클라이언트가 지원하는 프로토콜(h2는 HTTP/2를, http/1.1은 HTTP/1.1을 의미)을 서버에 알리는 과정
* (304) (OUT), TLS handshake, Client hello (1):
*  CAfile: /etc/ssl/cert.pem # 클라이언트가 SSL/TLS 인증을 위해 사용하는 CA(인증 기관) 인증서 파일의 경로
*  CApath: none
* Recv failure: Connection reset by peer # 서버 또는 중간 네트워크 노드에 의해 연결이 갑자기 닫혔음을 나타낸다
                                         # 이는 여러 원인에 의해 발생할 수 있다.
* LibreSSL/3.3.6: error:02FFF036:system library:func(4095):Connection reset by peer # SSL 라이브러리(LibreSSL)에서 발생한 
                                                                                    # 에러 메시지. 연결이 리셋되었음을 의미
* Closing connection 0 # curl이 연결을 닫고 있음을 나타낸다
curl: (35) Recv failure: Connection reset by peer
```

## cURL error 56, errno 104

만약 서버가 연결을 강제로 종료한다면, 클라이언트는 TCP RST 패킷을 받아들이고, 이는 네트워크 시스템 호출의 실패로 이어져 `ECONNRESET`(errno 104) 오류가 발생합니다. cURL은 이 오류를 감지하고, `cURL error 56: SSL read: errno 104`와 같은 메시지로 보고합니다.

cURL로 HTTPS 요청을 보낼 때의 통신 과정과 errno를 받는 경우를 단계별로 설명하겠습니다.
여기서는 `https://google.com`에 대한 요청을 예로 들어, 만약 서버가 연결을 강제로 끊는 경우까지 설명합니다.

1. cURL 명령 실행

    ```bash
    curl https://google.com
    ```

2. DNS 조회

    - **클라이언트(cURL)**:
        - 입력된 URL에서 도메인 이름(`google.com`)을 추출합니다.
        - DNS 서버에 도메인 이름의 IP 주소를 요청합니다.
    - **DNS 서버**:
        - 도메인 이름에 해당하는 IP 주소를 응답합니다.
        - 예: `142.250.196.78` (Google 서버의 IP 주소)

3. TCP 연결 설정

    - **클라이언트**:
        - Google 서버의 IP 주소와 TCP 포트 443(HTTPS)을 사용하여 TCP 연결을 설정합니다.
        - 3-way 핸드셰이크 과정을 수행합니다:
            1. **SYN**: 클라이언트가 서버에 SYN 패킷을 보냅니다.
            2. **SYN-ACK**: 서버가 SYN-ACK 패킷으로 응답합니다.
            3. **ACK**: 클라이언트가 ACK 패킷으로 응답합니다.
    - **서버**:
        - 클라이언트와의 TCP 연결을 수락합니다.

4. SSL/TLS 핸드셰이크

    - **클라이언트**:
        - 서버와 SSL/TLS 핸드셰이크를 시작합니다.
        - `ClientHello` 메시지를 서버에 보냅니다.
    - **서버**:
        - `ServerHello` 메시지로 응답합니다.
        - 서버 인증서를 클라이언트에 보냅니다.
        - 키 교환 및 암호화 알고리즘을 협상합니다.
    - **클라이언트**:
        - 서버 인증서를 검증합니다.
        - 비밀 키 자료를 생성하고, 이를 서버와 공유합니다.
    - **서버**:
        - 비밀 키 자료를 받아들입니다.
        - 양쪽에서 대칭 암호화 키를 생성합니다.
    - **클라이언트와 서버**:
        - 핸드셰이크 완료 메시지를 교환합니다.

5. HTTPS 요청 전송

    - **클라이언트**:
        - 암호화된 HTTP 요청을 생성하고, 서버에 보냅니다.
        - 예: `GET / HTTP/1.1 Host: google.com`

6. 서버의 응답 처리

    - **서버**:
        - HTTP 요청을 처리합니다.
        - 응답을 생성하고 암호화된 형태로 클라이언트에 보냅니다.
        - 예: `HTTP/1.1 200 OK`

    서버가 클라이언트의 요청에 대해 200 OK 응답을 보내는 과정에서 연결을 강제로 종료하는 경우는 일반적으로 두 가지 상황으로 나눌 수 있습니다:

    1. **정상적인 200 OK 응답 후 강제 종료**:
        - 서버가 클라이언트의 요청을 정상적으로 처리하고, 200 OK 응답을 보낸 후에 연결을 강제로 종료할 수 있습니다.
        - 이 경우, 클라이언트는 200 OK 응답을 정상적으로 받고, 그 이후에 연결이 끊어진 것을 인지합니다.
        - 서버가 응답을 보낸 후 TCP RST 패킷을 보내 연결을 강제로 종료하면 클라이언트는 200 OK 응답을 받은 후 연결이 끊어진 것을 인지.
        - 클라이언트는 정상적인 HTTP 응답을 받은 후에 연결이 끊어질 경우, 응답 데이터를 정상적으로 수신하고 연결이 종료되었음을 인지합니다. 이 경우, 클라이언트는 데이터 수신 후 연결이 끊어진 것으로 처리합니다.

    2. **200 OK 응답을 보내기 전에 강제 종료**:
        - 서버가 클라이언트의 요청을 처리하는 도중, 어떤 이유로 인해 200 OK 응답을 보내기 전에 연결을 강제로 종료할 수 있습니다.
        - 이 경우, 클라이언트는 200 OK 응답을 받지 못하고 연결이 끊어진 것을 인지합니다.
        - 서버가 요청을 처리하는 도중 TCP RST 패킷을 보내 연결을 강제로 종료하면 클라이언트는 200 OK 응답을 받지 못하고 연결이 끊어진 것을 인지.
        - 클라이언트가 서버로부터 응답을 받기 전에 연결이 강제로 종료되면, 클라이언트는 데이터를 수신하는 과정에서 실패하게 됩니다.
        - cURL은 `recv()` 시스템 호출에서 실패하고, `ECONNRESET`(errno 104) 오류를 반환합니다.

7. 연결 강제 종료 (가정)

    - **서버**:
        - 특정 이유로 클라이언트와의 연결을 강제로 종료합니다.
        - TCP RST 패킷을 클라이언트에 보냅니다.

8. 클라이언트에서의 오류 처리

    - **클라이언트**:
        - 서버로부터 TCP RST 패킷을 받습니다.
        - 이 시점에서 클라이언트는 연결이 강제로 종료되었음을 인식합니다.
        - 네트워크 시스템 호출(`recv()` 등)이 실패하며, `ECONNRESET`(errno 104) 오류를 반환합니다.

9. cURL의 오류 보고

    - **cURL**:
        - 내부적으로 네트워크 시스템 호출이 실패할 때, 해당 `errno`를 읽습니다.
        - 오류 메시지로 `cURL error 56: SSL read: errno 104`를 보고합니다.
        - 이 메시지는 SSL 읽기 과정에서 네트워크 데이터 수신에 실패했음을 나타냅니다.
