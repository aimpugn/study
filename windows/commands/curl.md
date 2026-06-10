# curl

## troubleshooting

### curl: (35) schannel: next InitializeSecurityContext failed: CRYPT_E_NO_REVOCATION_CHECK (0x80092012)

이 에러는 **서버 인증서가 반드시 폐기됐다는 뜻이 아니라**, Windows의 TLS 엔진인 **Schannel/CryptoAPI가 "이 인증서가 폐기됐는지" 확인하는 CRL/OCSP 조회를 완료하지 못했다**는 뜻입니다. Microsoft의 에러 코드 정의도 `CRYPT_E_NO_REVOCATION_CHECK (0x80092012)`를 "revocation function was unable to check revocation for the certificate"라고 설명합니다.

흐름은 이렇게 봐야 합니다.

```text
curl.exe
  -> TCP 연결
  -> TLS handshake 시작
  -> 서버 인증서 체인 수신
  -> Windows Schannel/CryptoAPI가 인증서 체인 검증
      -> 신뢰 루트/중간 CA 확인
      -> 인증서 폐기 여부 확인
          -> OCSP 응답 확인
          -> 로컬 Cryptnet URL cache 확인
          -> 시스템 CRL store 확인
          -> 인증서의 AIA OCSP URL 조회
          -> 인증서의 CDP CRL URL 조회
  -> 이 중 폐기 여부 확인 단계가 실패
  -> InitializeSecurityContext 실패
  -> curl: (35) schannel ... CRYPT_E_NO_REVOCATION_CHECK
```

Microsoft 문서 기준으로 Windows의 revocation 확인은 stapled OCSP, Cryptnet URL cache, 시스템 CRL store, 인증서의 AIA OCSP URL, 인증서의 CRL Distribution Point URL을 차례로 활용합니다. 즉 대상 서버와는 TLS 연결이 됐더라도, **별도의 CA/OCSP/CRL 서버로 나가는 조회가 막히면** 이 에러가 날 수 있습니다.

가장 흔한 원인은 회사망, 프록시, 보안 장비, TLS inspection, 폐쇄망입니다. 브라우저는 되는데 `curl`만 실패하는 경우도 이상하지 않습니다. 브라우저는 자체 캐시, 자체 네트워크 스택, 프록시 설정, 정책 처리가 다를 수 있고, 반대로 `curl.exe`는 Windows Schannel을 통해 인증서 검증을 하다가 CRL/OCSP 조회에서 막힐 수 있습니다.

먼저 실제로 Windows Schannel 기반 curl인지 확인하세요.

```powershell
# PowerShell에서 curl이 alias인지 실제 curl.exe인지 확인합니다.
# 이 에러 메시지가 "curl: (35) schannel" 형태라면 대개 실제 curl.exe입니다.
Get-Command curl

# curl이 어떤 TLS backend로 빌드됐는지 확인합니다.
# 출력에 Schannel이 보이면 Windows 인증서 저장소와 Schannel 검증 경로를 사용합니다.
curl.exe -V

# 실패 지점을 보기 위해 verbose로 재현합니다.
curl.exe -v https://대상호스트
```

가장 안전한 임시 완화책은 `--ssl-revoke-best-effort`입니다. 이 옵션은 Schannel에서 revocation 확인이 "missing/offline distribution point" 때문에 실패한 경우 그 실패를 무시합니다. curl 공식 문서에 명시된 Schannel 전용 옵션입니다.

```powershell
# 폐기 확인을 완전히 끄는 것이 아니라,
# CRL/OCSP 배포 지점이 없거나 오프라인이라 확인할 수 없는 경우만 완화합니다.
curl.exe --ssl-revoke-best-effort -v https://대상호스트
```

반대로 `--ssl-no-revoke`는 revocation check 자체를 끕니다. curl 공식 문서는 이 옵션을 "Disable certificate revocation checks"라고 설명하고, SSL 보안을 느슨하게 만든다고 경고합니다.

```powershell
# 더 강한 우회입니다.
# 내부망 테스트, 일회성 설치, 원인 분리 용도로만 쓰는 편이 안전합니다.
curl.exe --ssl-no-revoke -v https://대상호스트
```

`-k` 또는 `--insecure`는 더 위험합니다. 이 옵션은 revocation 확인만 건너뛰는 것이 아니라 인증서 검증 자체를 약하게 만들기 때문에, "회사망에서 CRL만 막힌 상황"을 해결하려고 쓰기에는 범위가 너무 큽니다. 이 케이스에서는 우선순위가 `--ssl-revoke-best-effort` → `--ssl-no-revoke` → 정말 불가피한 경우의 `--insecure` 순서입니다.

근본 해결은 curl 옵션이 아니라 **Windows가 인증서 폐기 정보를 가져올 수 있게 만드는 것**입니다. 외부 사이트라면 방화벽/프록시에서 대상 도메인뿐 아니라 해당 인증서 체인의 OCSP/CRL URL도 허용해야 합니다. 내부 사설 CA라면 발급된 서버 인증서에 들어 있는 CDP/AIA URL이 실제로 접근 가능해야 하고, CRL이 만료되지 않아야 하며, 중간 CA 인증서가 Windows의 적절한 저장소에 배포되어야 합니다.

진단은 이렇게 진행하면 됩니다.

```powershell
# 1. 일반 접속은 실패하는지 확인합니다.
curl.exe -v https://대상호스트

# 2. best-effort로 성공하는지 확인합니다.
# 여기서 성공하면 "서버 인증서 자체"보다 "revocation 조회 경로" 문제가 강해집니다.
curl.exe --ssl-revoke-best-effort -v https://대상호스트

# 3. no-revoke에서만 성공하는지 확인합니다.
# best-effort도 실패하고 no-revoke만 성공하면, revocation 확인 실패가 더 강하게 의심됩니다.
curl.exe --ssl-no-revoke -v https://대상호스트
```

인증서를 파일로 저장할 수 있다면 Windows의 `certutil`로 체인과 URL 조회를 검사할 수 있습니다. Microsoft 쪽 PKI 문서에서도 `certutil -urlfetch -verify`를 인증서 검증과 chain build 정보 확인에 사용하는 예를 듭니다.

```powershell
# server.cer는 브라우저에서 내보내거나, 운영자가 제공한 서버 인증서 파일입니다.
# 이 명령은 인증서 체인 구성과 CRL/OCSP URL 조회 실패 여부를 드러내는 데 유용합니다.
certutil -urlfetch -verify .\server.cer > .\cert-verify.txt

# GUI로 AIA/CDP/OCSP/CRL URL 접근 가능성을 확인할 때 사용합니다.
certutil -url .\server.cer
```

캐시가 꼬인 경우도 있습니다. Windows는 CRL/OCSP 응답을 캐시하므로, 이전 실패나 만료된 응답이 계속 영향을 줄 수 있습니다. 이때는 관리자 PowerShell 또는 CMD에서 캐시를 비우고 재시도합니다.

```powershell
# CRL 캐시 삭제
certutil -urlcache crl delete

# OCSP 캐시 삭제
certutil -urlcache ocsp delete

# 다시 재현
curl.exe -v https://대상호스트
```

프록시가 있는 회사망이라면 "curl이 대상 서버로 나가는 프록시"와 "Windows가 CRL/OCSP를 조회하는 경로"를 따로 봐야 합니다. 대상 서버 요청만 프록시를 타도, revocation 조회 URL이 별도 도메인이라 막힐 수 있습니다.

```powershell
# 현재 세션에서 curl 대상 요청에 프록시를 태웁니다.
$env:HTTPS_PROXY = "http://proxy.company.local:8080"
$env:HTTP_PROXY  = "http://proxy.company.local:8080"

# Windows WinHTTP 프록시 상태를 확인합니다.
netsh winhttp show proxy

# 조직에서 IE/Windows 인터넷 옵션 프록시를 표준으로 쓴다면 WinHTTP에 가져옵니다.
# 회사 정책에 따라 관리자 권한이 필요할 수 있습니다.
netsh winhttp import proxy source=ie

# 다시 확인합니다.
curl.exe --ssl-revoke-best-effort -v https://대상호스트
```

판단 기준은 간단합니다. `curl.exe --ssl-revoke-best-effort`로 성공하면 네트워크나 인증서의 revocation distribution point 접근 문제가 핵심입니다. `--ssl-no-revoke`에서만 성공하면 revocation 검증을 수행할 수 없는 상태가 더 강하고, 운영 환경에서는 CRL/OCSP URL 허용, 사설 CA의 CRL 게시, 중간 CA 배포를 고쳐야 합니다. 둘 다 실패하면 revocation 문제가 아니라 루트 CA 불신, 중간 인증서 누락, TLS inspection CA 미설치, SNI/프록시/방화벽 문제까지 넓혀 봐야 합니다.

실무적으로는 이렇게 처리하는 것이 가장 안전합니다.

```powershell
# 일회성 설치/다운로드이고 회사망에서 CRL/OCSP가 막힌 상황이면 우선 이 방식
curl.exe --ssl-revoke-best-effort -fsSL "https://example.com/install.cmd" -o install.cmd

# 내부 API/운영 자동화라면 옵션으로 숨기지 말고,
# 방화벽/프록시/사설 CA의 CRL, OCSP, AIA, CDP 경로를 정상화하는 것이 맞습니다.
```

정리하면, 이 에러의 본질은 **TLS 인증서 체인 검증 중 "폐기 여부 확인"을 Windows가 완료하지 못한 상태**입니다. 급한 우회는 `--ssl-revoke-best-effort`, 원인 분리는 `--ssl-no-revoke`, 근본 해결은 CRL/OCSP 접근성, 사설 CA 배포, 프록시/방화벽 정책, Windows 인증서 저장소를 바로잡는 것입니다.

---

이 용어들은 전부 **"TLS 서버 인증서를 받은 Windows 클라이언트가, 이 인증서를 지금 믿어도 되는지 판단하는 경로"**에 등장합니다. `curl.exe`가 Windows에서 Schannel backend로 빌드되어 있으면, TLS 연결과 인증서 검증의 상당 부분을 Windows Schannel/CryptoAPI 쪽에 맡깁니다. Schannel은 Windows의 TLS/SSL 구현인 Security Support Provider이고, Microsoft 문서도 Schannel이 SSL/TLS 프로토콜을 구현하며 SSPI를 통해 사용된다고 설명합니다.

전체 위치를 먼저 보면 이렇습니다.

```text
curl.exe
  -> Windows Schannel
       TLS handshake, 암호 스위트 협상, 서버 인증서 수신
  -> CryptoAPI / Crypt32
       인증서 체인 구성
       루트 CA 신뢰 확인
       인증서 이름 / 기간 / 용도 확인
       폐기 여부 확인
           -> stapled OCSP
           -> Cryptnet URL cache
           -> CRL store, 예: HKLM\CA
           -> AIA 안의 OCSP URL
           -> CDP 안의 CRL URL
```

**Windows Schannel**은 Windows가 제공하는 TLS/SSL 엔진입니다. "엔진"이라고 부르는 이유는 애플리케이션이 직접 TLS record, handshake, cipher suite, 인증서 교환을 모두 구현하지 않고, Windows의 보안 패키지인 Schannel을 호출해서 보안 채널을 만들 수 있기 때문입니다. Microsoft는 Secure Channel, 즉 Schannel을 "identity authentication과 encryption을 통한 secure private communication을 제공하는 보안 프로토콜 집합"이라고 설명합니다.

그래서 `curl: (35) schannel: next InitializeSecurityContext failed`라는 문구에서 핵심은 `curl` 자체가 "인증서 폐기 목록을 직접 파싱하다가 실패했다"가 아닙니다. `curl`이 Windows Schannel에 TLS 보안 컨텍스트 생성을 맡겼고, Schannel이 그 과정에서 인증서 검증을 진행하다가 실패한 것입니다. 여기서 `InitializeSecurityContext`는 클라이언트 쪽 보안 컨텍스트를 한 단계 진행하는 Windows SSPI 계열 호출이라고 이해하면 됩니다.

**CryptoAPI**는 Windows의 암호화와 인증서 관련 API 집합입니다. Schannel은 암호 연산이나 인증서 저장소 접근 같은 일을 할 때 CryptoAPI를 사용합니다. Microsoft 문서도 Schannel이 public/private key 저장 같은 암호 작업에 CryptoAPI를 사용한다고 설명합니다. 이 맥락에서 `CryptoAPI`는 "RSA/AES 같은 암호 알고리즘만 제공하는 라이브러리"가 아니라, 인증서 저장소, 인증서 체인, 키 컨테이너, 폐기 확인 같은 PKI 작업까지 포함하는 Windows 보안 기반으로 보는 편이 맞습니다.

그중 **Crypt32**는 인증서 검증에서 자주 보이는 Windows 구성 요소입니다. Microsoft 문서는 온라인 폐기 확인이 켜진 상태에서 `Crypt32 CertGetCertificateChain` API가 호출되면 유효한 OCSP 응답 또는 CRL을 찾는다고 설명합니다. 이때 Windows는 먼저 TLS handshake 안에 포함된 stapled OCSP 응답을 보고, 그다음 사용자 Cryptnet URL cache, 시스템 저장소의 CRL, AIA 확장의 OCSP URL, CDP 확장의 CRL URL 순서로 확인을 시도합니다.

여기서 중요한 점은 "인증서가 유효하다"는 판단이 단순히 날짜만 보는 일이 아니라는 것입니다. 서버가 보낸 인증서가 아직 만료되지 않았고, 이름도 맞고, 루트 CA까지 체인이 이어져도, 그 인증서가 발급 후에 폐기됐을 수 있습니다. 예를 들어 개인키가 유출되었거나, 잘못 발급되었거나, 조직이 더 이상 해당 도메인을 소유하지 않는 상황이 생기면 CA는 인증서를 폐기할 수 있습니다. 그래서 Windows는 "이 인증서가 폐기 목록에 올라가 있지 않은가"를 따로 확인합니다.

**CRL**은 Certificate Revocation List, 즉 인증서 폐기 목록입니다. 쉽게 말하면 CA가 "이 일련번호(serial number)를 가진 인증서는 더 이상 믿지 말라"고 서명해서 배포하는 목록입니다. RFC 5280은 인터넷 PKI에서 X.509 v3 인증서와 X.509 v2 CRL의 형식과 의미를 정의하는 표준 문서입니다.

**OCSP**는 Online Certificate Status Protocol입니다. CRL이 "폐기된 인증서 목록 전체를 다운로드해서 내 인증서 일련번호가 있는지 확인하는 방식"이라면, OCSP는 "이 인증서 하나의 현재 상태를 OCSP responder에게 물어보는 방식"입니다. RFC 6960은 OCSP를 CRL 없이 디지털 인증서의 현재 상태를 확인하는 프로토콜이라고 정의합니다.

**Cryptnet URL cache**는 Windows가 CRL이나 OCSP 응답을 URL 단위로 저장해두는 캐시입니다. "브라우저 캐시"가 아니라 Windows 인증서 검증 경로에서 쓰는 캐시라고 보는 것이 정확합니다. Microsoft 문서는 Crypt32가 온라인 폐기 확인을 할 때 사용자의 Cryptnet URL cache에서 시간상 유효한 OCSP 응답 또는 CRL을 먼저 확인하고, AIA OCSP URL이나 CDP CRL URL에서 새로 받은 응답도 성공 시 사용자의 Cryptnet URL cache에 추가한다고 설명합니다.

이 캐시가 필요한 이유는 CRL 다운로드가 비용이 크기 때문입니다. CRL은 CA가 발급한 인증서가 많고 폐기된 인증서가 많을수록 커질 수 있습니다. Microsoft 문서도 첫 유효성 검사에서는 CRL 다운로드가 끝날 때까지 호출 애플리케이션이 차단될 수 있고, 이후 검증은 CRL이 만료될 때까지 캐시된 CRL을 사용한다고 설명합니다. 그래서 회사망에서 `curl`이 갑자기 느려지거나 실패할 때, 실제 장애 지점은 대상 서버가 아니라 CRL/OCSP URL 접근 또는 이 캐시의 만료 상태일 수 있습니다.

**CRL store**는 CRL이 저장되는 Windows 인증서 저장소 쪽 개념입니다. Microsoft 문서는 Schannel 또는 애플리케이션이 인증서 체인을 만들고 수신한 인증서 체인을 검증할 때 Root store와 CA store를 사용한다고 설명합니다.

Cryptnet URL cache와 CRL store는 비슷해 보이지만 역할이 다릅니다. Cryptnet URL cache는 네트워크 URL에서 받아온 OCSP/CRL 응답을 재사용하기 위한 캐시에 가깝고, CRL store는 Windows 인증서 저장소 체계 안에 있는 CRL 저장 위치입니다. 운영 관점에서 이 차이가 중요한 이유는, 캐시 삭제로 해결되는 문제와 GPO/사설 CA/중간 CA 배포를 고쳐야 하는 문제가 다르기 때문입니다.

**AIA OCSP**에서 AIA는 Authority Information Access입니다. RFC 5280은 AIA 확장이 인증서 발급자에 대한 정보와 서비스에 접근하는 방법을 나타내며, 온라인 검증 서비스와 CA 정책 데이터가 포함될 수 있다고 설명합니다. 또 CRL 위치는 AIA가 아니라 `cRLDistributionPoints` 확장에서 제공된다고 분리해서 설명합니다.

즉 AIA 안에 OCSP URL이 있으면 Windows는 이렇게 해석합니다.

```text
서버 인증서 안의 AIA 확장
  -> id-ad-ocsp accessMethod 발견
  -> accessLocation에 있는 OCSP responder URL 확인
  -> "이 인증서 serial number가 현재 good/revoked/unknown 중 무엇인가?" 질의
  -> 응답이 유효하면 Cryptnet URL cache에 저장
```

이 경로가 막히면 인증서 자체가 정상이어도 `CRYPT_E_NO_REVOCATION_CHECK`가 날 수 있습니다. 예를 들어 회사 프록시가 대상 API 도메인은 허용했지만 `ocsp.digicert.com`, `ocsp.sectigo.com`, 사설 CA의 OCSP 서버 같은 URL은 차단한 경우입니다. 이때 사용자는 "나는 `https://example.com`에 접속했을 뿐인데 왜 다른 도메인 접근이 필요하지?"라고 느끼지만, Windows 입장에서는 인증서 폐기 여부 확인을 위해 인증서 안에 적힌 별도 URL로 조회해야 합니다.

**CDP CRL URL**에서 CDP는 CRL Distribution Point입니다. RFC 5280은 CRL Distribution Points 확장이 CRL 정보를 어떻게 얻는지 식별한다고 설명합니다. 다시 말해 CDP는 인증서 안에 들어 있는 "이 인증서의 폐기 여부를 확인하려면 여기에서 CRL을 받아라"라는 위치 정보입니다.

CDP CRL 방식은 이렇게 움직입니다.

```text
서버 인증서 안의 CDP 확장
  -> CRL URL 발견
  -> CA의 CRL 서버에서 .crl 파일 다운로드
  -> CRL 서명 검증
  -> CRL의 ThisUpdate / NextUpdate 시간 확인
  -> 서버 인증서 serial number가 CRL 안에 있는지 확인
  -> 없으면 일단 폐기되지 않은 것으로 판단
  -> 결과 CRL을 Cryptnet URL cache에 저장
```

이 방식은 OCSP보다 단순하지만, 목록을 다운로드해야 하므로 CRL이 커지면 느려질 수 있습니다. Microsoft 문서도 CA가 많은 인증서를 발급하면 CRL에 있는 폐기 항목 수가 커지고 CRL 크기가 증가할 수 있다고 설명합니다. 그래서 Windows는 캐시, prefetch, partitioned CRL 같은 최적화 경로를 둡니다.

이제 앞의 `curl` 에러와 연결하면, `CRYPT_E_NO_REVOCATION_CHECK`는 보통 이 지점에서 터집니다.

```text
서버 인증서 수신은 성공
인증서 체인 구성도 어느 정도 진행
하지만 폐기 여부 확인 단계에서
  stapled OCSP 없음
  Cryptnet URL cache에 유효한 응답 없음
  CRL store에도 유효한 CRL 없음
  AIA OCSP URL 접근 실패
  CDP CRL URL 접근 실패
따라서 Windows가 "폐기됐는지 아닌지 판단 불가" 상태가 됨
```

여기서 "판단 불가"와 "폐기됨"은 다릅니다. `revoked`는 CA가 실제로 폐기했다고 확인된 상태입니다. 반면 `NO_REVOCATION_CHECK`는 확인 경로가 없거나 막혀서 판단할 수 없는 상태입니다. 운영 로그에서 이 둘을 혼동하면 조치가 완전히 달라집니다. 실제 폐기라면 인증서를 재발급해야 하지만, 판단 불가라면 프록시, 방화벽, OCSP/CRL URL 허용, 사설 CA의 AIA/CDP 게시 상태, Windows 저장소 배포를 봐야 합니다.

눈으로 확인하려면 이 명령들이 도움이 됩니다.

```powershell
# curl이 Schannel backend를 쓰는지 확인합니다.
# 출력의 Features/SSL backend 쪽에 Schannel이 보이면 Windows 인증서 검증 경로를 탑니다.
curl.exe -V

# 어떤 TLS 단계에서 실패하는지 봅니다.
# schannel, InitializeSecurityContext, certificate revocation 같은 문구가 관측 포인트입니다.
curl.exe -v https://대상호스트

# 서버 인증서를 파일로 확보한 뒤 Windows 인증서 체인과 revocation URL 조회를 확인합니다.
# 실패하면 출력에 AIA, CDP, OCSP, CRL URL 접근 실패 흔적이 남는 경우가 많습니다.
certutil -urlfetch -verify .\server.cer > .\cert-verify.txt

# GUI로 인증서 안의 AIA/CDP URL과 조회 가능 여부를 확인합니다.
certutil -url .\server.cer

# Cryptnet URL cache 쪽 문제를 분리할 때 사용합니다.
certutil -urlcache crl delete
certutil -urlcache ocsp delete
```

정리하면, **Schannel은 Windows의 TLS 엔진**, **CryptoAPI/Crypt32는 그 TLS 검증 중 인증서/체인/폐기 확인을 수행하는 Windows PKI 기반**, **Cryptnet URL cache는 OCSP/CRL 조회 결과 캐시**, **CRL store는 Windows 인증서 저장소 안의 CRL 보관 위치**, **AIA OCSP는 인증서 안에 적힌 OCSP 조회 위치**, **CDP CRL URL은 인증서 안에 적힌 CRL 다운로드 위치**입니다. Windows curl의 revocation 오류는 이 중 AIA/CDP 조회 또는 캐시/저장소 확인 경로가 막힌 상황에서 가장 자주 드러납니다.
