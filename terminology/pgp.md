# PGP

- [PGP](#pgp)
    - [PGP(Pretty Good Privacy)](#pgppretty-good-privacy)
    - [PGP의 역할](#pgp의-역할)
    - [적용되는 암호화 기법](#적용되는-암호화-기법)
        - [비대칭 암호화(Public-Key Cryptography)](#비대칭-암호화public-key-cryptography)
        - [대칭 암호화(Symmetric Encryption)](#대칭-암호화symmetric-encryption)
        - [디지털 서명(Digital Signature)](#디지털-서명digital-signature)
        - [웹 오브 트러스트(Web of Trust)](#웹-오브-트러스트web-of-trust)
    - [사용례](#사용례)
        - [이메일 보안](#이메일-보안)
        - [데이터 암호화](#데이터-암호화)
        - [파일 및 소프트웨어 서명](#파일-및-소프트웨어-서명)
            - [메이븐 아티팩트 다운로드 및 PGP 검증](#메이븐-아티팩트-다운로드-및-pgp-검증)
    - [기타](#기타)

## PGP(Pretty Good Privacy)

PGP(Pretty Good Privacy)는 데이터 암호화 및 디지털 서명을 통한 보안 강화를 목표로 하는 소프트웨어입니다.
필 짐머만이 1991년에 개발했으며, 이메일 보안, 파일 암호화, 소프트웨어 서명, 데이터 무결성 보장 등에 널리 사용되었습니다.

1990년대 초, 인터넷이 확산되면서 데이터 보안과 프라이버시 보호의 필요성이 증가했습니다.
- 이메일과 파일 공유가 보편화되면서 도청, 감시, 변조 위험 증가
- 특히 정부기관, 해커, 악의적 행위자들이 데이터를 탈취하거나 조작하는 사례 발생

이러한 위협을 방지하고 일반 사용자도 강력한 보안 기술을 사용할 수 있도록 개발된 것이 PGP입니다.

필 짐머만은 이러한 문제를 해결하기 위해 비대칭 암호화(Public-Key Cryptography)를 기반으로 하는 PGP를 개발했습니다.
PGP는 무료로 배포되어 전 세계적으로 널리 사용되기 시작했습니다.

> [PGP의 강력한 암호화 기능으로 인해 미국 정부는 이를 국제 무기 거래 규정(ITAR: International Traffic in Arms Regulations) 위반으로 간주하고 법적 조치를 고려했습니다. 짐머만은 이에 대응하여 PGP의 소스 코드를 책으로 출판하여 미국 내에서 법적 문제를 우회하는 방식을 사용했습니다.](https://namu.wiki/w/PGP#s-3.1)

## PGP의 역할

공식적인 출처에서 제공하는 PGP 서명 파일을 사용하여 다운로드하는 동안 변경되지 않았음을 확인합니다.
이를 통해 다음과 같은 보안 기능을 제공합니다.
- 기밀성(Confidentiality): 데이터가 인가되지 않은 사용자에게 노출되지 않도록 보호 (암호화)
- 무결성(Integrity): 데이터가 전송 중 변조되지 않았음을 보장 (디지털 서명)
- 인증(Authentication): 데이터의 발신자가 신뢰할 수 있는 사람인지 확인 (전자 서명)
- 부인 방지(Non-repudiation): 발신자가 자신의 메시지를 부인할 수 없도록 보장

이를 통해 다음과 같은 보안 문제들을 방지할 수 있습니다.
- 공격자가 파일을 가로채 변조하는 MITM(Man-in-the-Middle) 공격
- 다운로드한 파일이 악성 코드에 감염되었는지 확인
- 신뢰할 수 없는 소스로부터 온 수정된 파일을 감지(공급망 공격(Supply Chain Attack))

## 적용되는 암호화 기법

PGP는 효율적으로 보안성을 높이기 위해 여러 암호화 기법을 조합하여 사용합니다.
- 비대칭 암호화
- 대칭 암호화
- 디지털 서명
- 웹 오브 트러스트

### 비대칭 암호화(Public-Key Cryptography)

PGP는 공개 키(public key)와 개인 키(private key) 쌍을 사용합니다.
이를 통해 기존 대칭 암호화(Symmetric Encryption) 방식과 달리 키 분배 문제를 해결하면서 보안성을 강화합니다.

- 공개 키(Public Key): 누구나 접근할 수 있으며, 데이터를 *암호화하는 데 사용*됩니다.
- 개인 키(Private Key): 오직 소유자만이 보유하며, 암호화된 데이터를 *복호화하는 데 사용*됩니다.

A가 B에게 안전하게 메시지를 보내기 위해 B의 공개 키로 메시지를 암호화하고, 수신자인 B는 자신만 갖고 있는 개인 키로 복호화합니다. 예를 들어, Alice가 Bob에게 메시지를 보낼 때:
1. Bob의 공개 키로 메시지를 암호화
2. 암호화된 메시지를 Bob에게 전송
3. Bob이 자신의 개인 키로 메시지를 복호화

이 과정에서 Alice는 Bob의 개인 키를 몰라도 메시지를 안전하게 전달할 수 있습니다.

### 대칭 암호화(Symmetric Encryption)

비대칭 암호화는 보안성이 뛰어나지만, 연산 속도가 느리다는 단점이 있습니다.
이를 해결하기 위해 PGP는 하이브리드 암호화(Hybrid Encryption) 기법을 사용합니다.

1. 메시지를 빠른 대칭 키(세션 키)로 암호화
2. 대칭 키를 수신자의 공개 키로 암호화하여 함께 전송
3. 수신자가 개인 키로 대칭 키를 복호화한 후, 해당 키로 메시지를 복호화

### 디지털 서명(Digital Signature)

PGP는 메시지의 *무결성(integrity)*과 *인증(authentication)*을 보장하기 위해 디지털 서명을 사용합니다.
디지털 서명은 파일 변조 탐지 및 발신자 신원 확인에 사용됩니다.

- 발신자(Alice)가 메시지의 해시 값을 계산하여 발신자(Alice)의 개인 키로 서명을 생성
- 수신자(Bob)는 발신자(Alice)의 공개 키로 서명을 검증하여 메시지가 변조되지 않았음을 확인

가령, Alice가 Bob에게 문서를 보낼 때:
1. 발신자(Alice)가 문서의 해시값(SHA-256)을 계산
2. *발신자(Alice)의 개인 키*로 해시값을 암호화하여 서명을 생성하고 문서에 추가
3. 수신자(Bob)은 *발신자(Alice)의 공개 키*로 서명을 검증하여 문서의 무결성을 확인

### 웹 오브 트러스트(Web of Trust)

기존 인증 기관(CA) 기반의 PKI(Public Key Infrastructure)와 달리, PGP는 사용자가 직접 서로의 키를 신뢰할 수 있도록 설계된 분산형 인증 모델인 '웹 오브 트러스트'([Web of Trust](https://en.wikipedia.org/wiki/Web_of_trust)) 모델을 사용합니다. 사용자는 서로의 공개 키를 신뢰할 수 있도록 서명하고, 이러한 신뢰 체계를 바탕으로 네트워크를 형성합니다.

즉, *신뢰할 수 있는 사람이 서명한 키*라면 *해당 키의 소유자도 신뢰*할 수 있음을 의미합니다.
중앙 기관에 의존하지 않고도 공개 키의 신뢰성을 확보합니다.

## 사용례

### 이메일 보안

PGP는 이메일 내용을 암호화하고 서명을 추가하여 보안성을 높입니다.
Gmail, Outlook 등 다양한 이메일 클라이언트에서 'PGP/MIME' 형식을 지원됩니다.
참고로 요즘에는 'S/MIME'(Secure MIME)가 널리 사용되고 있습니다.

### 데이터 암호화

PGP는 파일을 암호화하여 저장하거나, 클라우드에 업로드할 때 보안성을 높이는 데 사용됩니다.

### 파일 및 소프트웨어 서명

소프트웨어 패키지 배포 시 PGP 서명을 추가하여 다운로드한 파일이 변조되지 않았음을 보장할 수 있습니다.

예를 들어, 메이븐(Maven) 저장소의 아티팩트는 [`.asc` 확장자로 된 PGP 서명 파일](https://www.adobe.com/uk/acrobat/resources/document-files/text-files/asc.html)을 함께 제공하며, 사용자는 이를 검증하여 공식적인 배포자인지 확인할 수 있습니다.

> An ASC file is a variant of the ASCII format, which is an encryption file used by Pretty Good Privacy (PGP) for secure online communication.

#### 메이븐 아티팩트 다운로드 및 PGP 검증

메이븐 아티팩트 파일을 검증하는 과정은 다음과 같습니다.
1. `*.jar`와 `*.asc` 파일 다운로드
2. 서명 파일(`*.asc`)에서 PGP 키 ID 확인
3. 키 서버에서 PGP 공개 키 다운로드
    - 공개 키 서버에서 `gpg --recv-keys` 명령어로 가져오기
    - 만약 PGP 키 서버에 공개 키가 없다면, 공식 출처(GitHub 등)에서 수동으로 가져오기
4. 서명 검증

메이븐 저장소에서 제공하는 아티팩트에는 `.asc` 확장자를 가진 PGP 서명 파일이 함께 제공됩니다.
이를 통해 해당 아티팩트가 공식적인 배포자가 제공한 것임을 확인할 수 있습니다.

- 사용하려는 패키지와 `.asc` 파일을 다운로드 받습니다.

    ```sh
    curl -O https://repo1.maven.org/maven2/org/springframework/spring-web/6.2.3/spring-web-6.2.3.jar
    curl -O https://repo1.maven.org/maven2/org/springframework/spring-web/6.2.3/spring-web-6.2.3.jar.asc
    ```

- 서평 파일(`.asc`)을 사용하여 어떤 PGP 키로 서명되었는지 확인합니다.

    `--verify` 옵션을 사용하여 RSA 키를 얻을 수 있습니다.

    ```sh
    ❯ gpg --verify spring-web-6.2.3.jar.asc
    gpg: assuming signed data in 'spring-web-6.2.3.jar'
    gpg: Signature made 목  2/13 21:54:55 2025 KST
    gpg:                using RSA key 48B086A7D843CFA258E83286928FBF39003C0425
    gpg: Good signature from "Spring Builds (JAR Signing) <builds@springframework.org>" [unknown]
    gpg:                 aka "Spring Builds (JAR Signing) <buildmaster@springframework.org>" [unknown]
    gpg:                 aka "Spring Builds (JAR Signing) <spring-builds@vmware.com>" [unknown]
    gpg: WARNING: This key is not certified with a trusted signature!
    gpg:          There is no indication that the signature belongs to the owner.
    Primary key fingerprint: 48B0 86A7 D843 CFA2 58E8  3286 928F BF39 003C 0425
    ```

    또는 `--list-packets`를 사용합니다. 이 옵션을 사용하면 PGP 데이터 구조의 내부 패킷을 분석하여 출력합니다.

    ```sh
    ❯ ll
    total 4088
    -rw-r--r--@ 1 rody  staff   2.0M  2 26 22:11 spring-web-6.2.3.jar
    -rw-r--r--@ 1 rody  staff   833B  2 26 22:11 spring-web-6.2.3.jar.asc

    ❯ gpg --list-packets spring-web-6.2.3.jar.asc
    # off=0 ctb=89 tag=2 hlen=3 plen=563
    :signature packet: algo 1, keyid 928FBF39003C0425
        version 4, created 1739451295, md5len 0, sigclass 0x00
        digest algo 8, begin of digest 0b 90
        hashed subpkt 33 len 21 (issuer fpr v4 48B086A7D843CFA258E83286928FBF39003C0425)
        hashed subpkt 2 len 4 (sig created 2025-02-13)
        subpkt 16 len 8 (issuer key ID 928FBF39003C0425)
        data: [4096 bits]
    ```

    - 전체 키 지문(Fingerprint): '48B086A7D843CFA258E83286928FBF39003C0425'

        > PGP 키 지문(Fingerprint)?
        > - 공개 키 전체를 SHA-1 또는 SHA-256으로 해싱하여 생성한 문자열입니다.
        >     - SHA-1: 160비트, 고유한 40자리 문자열입니다.
        >     - SHA-256: 256비트, 고유한 64자리 문자열입니다.
        > - 공개 키 데이터를 해싱하여 SHA-1 기반으로 생성됩니다.
        >
        > ```sh
        > # 16진수는 1 Byte(8 Bits)이므로 160비트에 대해 총 20개의 16진수 생성 가능
        > # 16진수는 두 글자로 구성되므로, 총 40자
        > perl -E "say length('48B086A7D843CFA258E83286928FBF39003C0425')"
        > 40
        > ```

    - PGP 키 ID: '928FBF39003C0425'

        > PGP 키 ID(Key ID)?
        > - 키 지문의 마지막 16자리만을 사용한 값으로 키를 간단하게 식별하는 용도로 사용됩니다.

    - 서명된 날짜: '2025-02-13'

- PGP 키 서버에서 PGP 공개 키가 등록된 경우 다운로드받을 수 있습니다.

    RSA 키를 사용하여 공개 키를 다운로드 받을 수 있습니다.

    ```sh
    ❯ gpg --keyserver hkps://keys.openpgp.org --recv-keys 48B086A7D843CFA258E83286928FBF39003C0425
    gpg: key 928FBF39003C0425: "Spring Builds (JAR Signing) <builds@springframework.org>" not changed
    gpg: Total number processed: 1
    gpg:              unchanged: 1
    ```

    또는 `--list-packets` 결과로 얻은 `keyid`를 사용하여 공개 키를 다운로드 받을 수 있습니다.

    ```sh
    ❯ gpg --keyserver hkps://keys.openpgp.org --recv-keys 928FBF39003C0425
    gpg: /Users/rody/.gnupg/trustdb.gpg: trustdb created
    gpg: key 928FBF39003C0425: public key "Spring Builds (JAR Signing) <builds@springframework.org>" imported
    gpg: Total number processed: 1
    gpg:               imported: 1
    ```

    키 서버에서 공개 키를 가져오면, 해당 키를 사용하여 서명을 검증할 수 있습니다.

    만약 이 공개 키 다운로드 과정을 거치지 않으면 "Can't check signature: No public key"라는 결과가 출력됩니다.

    ```sh
    ❯ gpg --verify jackson-databind-2.18.2.jar.asc jackson-databind-2.18.2.jar
    gpg: Signature made 목 11/28 10:12:00 2024 KST
    gpg:                using RSA key 28118C070CB22A0175A2E8D43D12CA2AC19F3181
    gpg: Can't check signature: No public key
    ```

- 다운로드한 파일의 서명을 검증합니다.

    ```sh
    ❯ gpg --verify spring-web-6.2.3.jar.asc spring-web-6.2.3.jar
    gpg: Signature made 목  2/13 21:54:55 2025 KST
    gpg:                using RSA key 48B086A7D843CFA258E83286928FBF39003C0425
    gpg: Good signature from "Spring Builds (JAR Signing) <builds@springframework.org>" [unknown]
    gpg:                 aka "Spring Builds (JAR Signing) <buildmaster@springframework.org>" [unknown]
    gpg:                 aka "Spring Builds (JAR Signing) <spring-builds@vmware.com>" [unknown]
    gpg: WARNING: This key is not certified with a trusted signature!
    gpg:          There is no indication that the signature belongs to the owner.
    Primary key fingerprint: 48B0 86A7 D843 CFA2 58E8  3286 928F BF39 003C 0425
    ```

    'Good signature' 메시지를 통해 파일이 변조되지 않았음을 확인할 수 있습니다.

    > "gpg: WARNING: This key is not certified with a trusted signature!"?
    >
    > 'A trusted signature' is one that was made by a valid key:
    > - Key validity defines whether this key belongs to the person that it claims.
    > - Key trust defines whether this key is allowed to sign other keys (Web-of-Trust). In other words, a trusted key may act as a CA and mark other keys as valid, transitively.
    >
    > References:
    > - [How to suppress "WARNING: This key is not certified with a trusted signature!"](https://superuser.com/a/1435150)
    > - [How to verify an imported GPG key](https://serverfault.com/a/569923)
    > - [Validating other keys on your public keyring](https://www.gnupg.org/gph/en/manual/x334.html)
    > - [GPG why is my trusted key not certified with a trusted signature?](https://security.stackexchange.com/a/147467)

참고로 `.sha1` 파일도 같이 제공될 경우, 이를 통해 검증할 수도 있습니다.

```sh
❯ curl -O https://repo1.maven.org/maven2/org/springframework/spring-web/6.2.3/spring-web-6.2.3.jar.sha1
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    40  100    40    0     0     72      0 --:--:-- --:--:-- --:--:--    72

❯ echo $(cat spring-web-6.2.3.jar.sha1)
662ac5ee41af27d183f97032b2fec2b652d379f5

❯ sha1sum spring-web-6.2.3.jar
662ac5ee41af27d183f97032b2fec2b652d379f5  spring-web-6.2.3.jar
```

## 기타

- [Verifying Apache Software Foundation Releases](https://www.apache.org/info/verification.html)
- [What are PGP/MIME and PGP/Inline?](https://proton.me/support/pgp-mime-pgp-inline)
- [MIME Security with Pretty Good Privacy (PGP)](https://datatracker.ietf.org/doc/html/rfc2015)
- [S/MIME(Security Services for Multipurpose Internet Mail Extension)](http://www.ktword.co.kr/test/view/view.php?m_temp1=3040&id=1579)
