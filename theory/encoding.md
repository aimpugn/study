# encoding

- [encoding](#encoding)
    - [encoding이란?](#encoding이란)
    - [예시를 통한 이해](#예시를-통한-이해)
    - [비트 패턴의 변환과 활용](#비트-패턴의-변환과-활용)
    - [비유를 통한 설명](#비유를-통한-설명)
        - [언어 번역](#언어-번역)
        - [음악 악보와 연주](#음악-악보와-연주)
        - [비밀번호와 암호화](#비밀번호와-암호화)
    - [문자열 인코딩](#문자열-인코딩)
        - [UTF-8](#utf-8)
            - [UTF-8 인코딩 방식](#utf-8-인코딩-방식)
        - [Quoted-Printable](#quoted-printable)
            - [Quoted-Printable 인코딩이 필요한 이유](#quoted-printable-인코딩이-필요한-이유)
            - [Quoted-Printable 인코딩 사용 시점](#quoted-printable-인코딩-사용-시점)
            - [인코딩 방식](#인코딩-방식)
            - [Quoted-Printable 인코딩의 예](#quoted-printable-인코딩의-예)
            - [Quoted-Printable 인코딩 과정](#quoted-printable-인코딩-과정)
            - [Quoted-Printable 인코딩의 예시 코드](#quoted-printable-인코딩의-예시-코드)
        - [base64](#base64)
            - [Base64 인코딩의 길이 증가 원리](#base64-인코딩의-길이-증가-원리)
            - [데이터는 늘어나지만 문자열은 줄어들 수 있음](#데이터는-늘어나지만-문자열은-줄어들-수-있음)

## encoding이란?

인코딩(Encoding)은 데이터를 특정 형식으로 변환하는 과정을 의미합니다.
이는 다양한 유형의 데이터를 컴퓨터 시스템에서 효율적으로 저장, 전송 및 처리할 수 있도록 하는 데 중요한 역할을 합니다.
인코딩은 여러 형태로 존재하며, 텍스트, 이미지, 비디오, 오디오 등의 다양한 데이터를 다룰 때 사용됩니다.

> *인코딩은 다른 비트 패턴을 표현하기 위해 사용하는 비트 패턴을 뜻한다.*
>
> - **다른 비트 패턴**: 원래 데이터를 의미합니다. 즉, 문자, 숫자, 이미지, 오디오 등의 원본 데이터를 나타냅니다.
> - **표현하기 위해 사용하는 비트 패턴**: 원래 데이터를 인코딩한 결과입니다. 즉, 컴퓨터가 이해할 수 있도록 변환된 비트 패턴을 의미합니다.
>
> 다시 말하자면, "인코딩은 데이터를 컴퓨터가 이해하고 처리할 수 있도록 특정 비트 패턴으로 변환하는 과정입니다. 이 비트 패턴은 원래 데이터(예: 문자, 이미지, 소리)를 표현하기 위한 것입니다."

이는 인코딩을 통해 우리가 이해하는 실제 데이터(문자, 이미지, 오디오 등)를 컴퓨터가 이해하고 처리할 수 있는 비트 패턴으로 변환하는 과정을 설명하는 것입니다. *각 데이터는 해당 데이터를 컴퓨터가 처리할 수 있는 형식으로 변환하기 위해 특정 비트 패턴으로 인코딩*됩니다. 이 과정은 데이터의 저장, 전송, 처리에서 중요한 역할을 하며, 다양한 인코딩 방식이 존재합니다.

즉, 인코딩이란 실제 데이터(이미지, 텍스트, 숫자, 오디오 등)를 컴퓨터가 이해할 수 있는 형태인 비트 패턴으로 변환하는 방식을 뜻합니다.

따라서, 인코딩은 실제 데이터를 표현하기 위해 사용하는 일련의 비트 패턴을 정의하는 과정입니다.
이를 통해 컴퓨터는 다양한 유형의 데이터를 저장, 전송 및 처리할 수 있게 됩니다.

1. **비트와 비트 패턴**

   컴퓨터는 모든 데이터를 비트(bit)라는 0과 1로 구성된 이진수(binary)를 사용해 저장하고 처리합니다.
   비트 패턴(bit pattern)은 이러한 비트들이 특정한 순서로 배열된 것을 말합니다.

   예를 들어, `01100001`이라는 비트 패턴은 8비트(1바이트)로 구성되어 있습니다.

2. **비트 패턴의 해석**

    특정 비트 패턴은 서로 다른 데이터 형식을 나타낼 수 있습니다.

    예를 들어, `01100001`이라는 비트 패턴은 ASCII 인코딩 방식에서 문자 `a`를 의미합니다.
    하지만 다른 인코딩 방식에서는 다른 의미를 가질 수 있습니다.

3. **인코딩의 역할**

    인코딩은 특정 데이터를 비트 패턴으로 변환하는 방법을 정의합니다.
    또한 데이터를 특정 비트 패턴으로 변환하는 과정입니다.
    인코딩 자체는 변환의 과정이지만, 결과적으로 데이터는 비트 패턴으로 표현됩니다.
    즉, 인코딩은 특정 데이터를 특정 비트 패턴으로 매핑(mapping)하는 방법을 정의합니다.

    예를 들어, 문자를 숫자로, 숫자를 비트로, 이미지를 픽셀 값으로 변환하는 과정이 인코딩입니다.

    - **문자 데이터**는 ASCII나 UTF-8과 같은 인코딩 방식을 통해 비트 패턴으로 변환됩니다.
    - **이미지 데이터**는 JPEG, PNG와 같은 형식으로 인코딩되어 비트 패턴으로 변환됩니다.
    - **오디오 데이터**는 MP3, WAV와 같은 형식으로 인코딩되어 비트 패턴으로 변환됩니다.

## 예시를 통한 이해

- 텍스트 인코딩

    - **ASCII 인코딩**:

    ASCII는 American Standard Code for Information Interchange의 약자로, 128개의 문자 집합을 정의합니다.

    예를 들어, 문자 `A`는 ASCII 코드에서 65에 해당하며, 비트 패턴으로는 `01000001`입니다.
    문자 `B`는 ASCII에서 `01000010`으로 인코딩됩니다. 따라서, 'B'를 표현하는 비트 패턴은 `01000010`입니다.

    - **UTF-8 인코딩**:

    UTF-8은 유니코드 문자를 인코딩하는 방식 중 하나로, 가변 길이 인코딩을 사용합니다.

    예를 들어, 유니코드 문자 `가`는 UTF-8에서 `EAB080`으로 인코딩됩니다.

- 이미지 인코딩

    이미지의 각 픽셀은 특정 색상을 가집니다. 이 색상 데이터는 컴퓨터가 처리할 수 있도록 비트 패턴으로 변환되어야 합니다.

    예를 들어, RGB 값 (255, 0, 0)은 빨간색을 의미합니다.

    - **JPEG 인코딩**:

    이미지를 압축하여 저장하는 방식입니다.
    각 픽셀의 색상을 특정 비트 패턴으로 표현하고, 이를 압축하여 비트 패턴의 길이를 줄입니다.
    RGB 값을 특정 비트 패턴으로 변환합니다.
    여기서 비트 패턴은 JPEG 형식에 따라 달라질 수 있으며, 압축 과정에서 더 복잡한 비트 패턴이 사용됩니다.

- 오디오 인코딩

    - **오디오 데이터**:

    오디오 신호는 시간에 따른 음압의 변화를 나타냅니다.

    - **MP3 인코딩**

    아날로그 오디오 신호를 디지털 신호로 변환하고, 이를 특정 비트 패턴으로 인코딩합니다.
    이 과정에서 압축 알고리즘이 적용되어 데이터의 크기를 줄이지만, 여전히 원래의 오디오 신호를 표현하는 비트 패턴을 사용합니다.

## 비트 패턴의 변환과 활용

- **네트워크 전송**:

    데이터를 네트워크를 통해 전송할 때, 데이터를 전송하기 적합한 비트 패턴으로 인코딩합니다.

    예를 들어, 텍스트 데이터를 네트워크 패킷으로 전송하기 위해 Base64 인코딩을 사용할 수 있습니다.

- **파일 저장**:

    데이터를 파일에 저장할 때, 파일 포맷에 맞는 비트 패턴으로 인코딩합니다.

    예를 들어, 텍스트 파일은 UTF-8 인코딩으로, 이미지 파일은 PNG 또는 JPEG 인코딩으로 저장됩니다.

## 비유를 통한 설명

### 언어 번역

- **비트 패턴**을 **다른 언어**로 생각해봅시다. 예를 들어, 영어와 한국어를 사용하는 두 사람이 있다고 가정합니다.
- **데이터**는 이 두 사람 사이에서 전달해야 할 정보입니다. 예를 들어, 영어로 된 문장 "Hello"가 있습니다.
- **인코딩**은 이 영어 문장을 한국어로 번역하는 과정입니다. 즉, "Hello"를 한국어로 "안녕하세요"로 번역합니다.
- 여기서 "Hello"라는 영어 단어는 하나의 비트 패턴(영어라는 언어에서의 비트 패턴)이고, "안녕하세요"라는 한국어 단어는 다른 비트 패턴(한국어라는 언어에서의 비트 패턴)입니다.
- 인코딩은 "Hello"를 "안녕하세요"로 변환하여 다른 비트 패턴을 표현합니다.

### 음악 악보와 연주

- **비트 패턴**을 **악보**로 생각해봅시다. 예를 들어, 피아노 악보가 있습니다.
- **데이터**는 실제 연주할 음악입니다. 즉, 악보에 따라 연주할 음악입니다.
- **인코딩**은 이 악보를 피아노 연주로 변환하는 과정입니다. 즉, 악보에 적힌 노트를 실제 피아노 건반을 눌러 소리로 변환합니다.
- 여기서 악보에 적힌 각 노트는 하나의 비트 패턴(음악을 표기하는 방식에서의 비트 패턴)이고, 피아노에서 연주되는 실제 소리는 다른 비트 패턴(소리라는 형태에서의 비트 패턴)입니다.
- 인코딩은 악보의 노트를 실제 소리로 변환하여 다른 비트 패턴을 표현합니다.

### 비밀번호와 암호화

- **비트 패턴**을 **비밀번호**로 생각해봅시다. 예를 들어, 간단한 비밀번호 "1234"가 있습니다.
- **데이터**는 보호해야 할 정보입니다. 즉, 비밀번호 자체입니다.
- **인코딩**은 이 비밀번호를 암호화된 문자열로 변환하는 과정입니다. 즉, "1234"를 암호화하여 "e99a18c428cb38d5f260853678922e03"와 같은 해시 값으로 변환합니다.
- 여기서 "1234"라는 비밀번호는 하나의 비트 패턴이고, 암호화된 해시 값 "e99a18c428cb38d5f260853678922e03"는 다른 비트 패턴입니다.
- 인코딩은 비밀번호를 해시 값으로 변환하여 다른 비트 패턴을 표현합니다.

## 문자열 인코딩

### UTF-8

UTF-8 인코딩은 가변 길이 문자 인코딩 방식으로, *Unicode 문자를 1바이트에서 4바이트 사이의 길이로 인코딩*합니다.
사용되는 바이트 수는 문자의 코드 포인트 값에 따라 달라집니다. 이 설명에서는 특히 ASCII 범위(0x0000 ~ 0x007F)에 속하는 문자가 UTF-8로 인코딩될 때 어떻게 처리되는지를 중점적으로 설명하겠습니다.

1. **Unicode**:
  
    전 세계의 모든 문자를 표현하기 위해 고안된 표준입니다.
    8비트 덩어리(chunk)로 구성됩니다.

    각 문자는 고유한 코드 포인트(code point)로 식별됩니다.
    코드 포인트란 Unicode에서 각 문자를 식별하기 위해 부여된 숫자입니다.
    예를 들어
    - 'A'는 U+0041, '한'은 U+D55C입니다.
    -문자 'A'의 Unicode 코드 포인트는 U+0041입니다.

2. **UTF-8**:

    Unicode 문자를 1바이트에서 4바이트 길이의 바이트 시퀀스로 인코딩하는 방식입니다.
    UTF-8은 ASCII와 호환성을 유지하면서, ASCII 범위 밖의 문자도 표현할 수 있도록 설계되었습니다.

#### UTF-8 인코딩 방식

UTF-8에서 문자의 코드 포인트 값에 따라 인코딩되는 바이트 수가 달라집니다:

- 1바이트: U+0000 ~ U+007F
- 2바이트: U+0080 ~ U+07FF
- 3바이트: U+0800 ~ U+FFFF
- 4바이트: U+10000 ~ U+10FFFF

예를 들어 ASCII 문자의 바이트는 1바이트이고, 그 범위는 U+0000부터 U+007F(0x00 ~ 0x7F)까지입니다.
이는 7비트로 표현할 수 있으며, UTF-8로 인코딩될 때는 1바이트(8비트)를 사용합니다.
이때 가장 상위 비트(MSB)는 0으로 설정됩니다.

- **문자 'A'**:
    - Unicode 코드 포인트: U+0041
    - 16진수로 표현: 0x0041
    - 2진수로 표현: 0100 0001 (7비트, 상위 비트는 0)
    - UTF-8 인코딩: 0100 0001 (1바이트, 상위 비트는 0)

### Quoted-Printable

Quoted-Printable 인코딩은 8비트 데이터를 전송 가능한 7비트 ASCII 형식으로 안전하게 변환하기 위해 사용되는 인코딩 방식입니다.
주로 이메일에서 MIME(Multipurpose Internet Mail Extensions) 구조를 사용하여 8비트 텍스트를 7비트로 안전하게 전송할 때 사용됩니다.

예를 들어, 이메일 본문이 UTF-8 인코딩된 다국어 텍스트를 포함할 때 Quoted-Printable 인코딩을 사용하여 호환성을 유지합니다.

Quoted-Printable 인코딩은 사람이 읽을 수 있는 텍스트 형식을 유지하면서, 8비트 문자나 특정 제어 문자를 안전하게 인코딩할 수 있도록 합니다.

1. **출력 가능 문자**:
   - ASCII 문자 중 33~60(33-`!`, 60-`<`), 62~126(62-`>`, 126-`~`) 범위의 출력 가능한 문자들은 그대로 유지됩니다.
   - ASCII의 공백 문자(32, `SP`)와 탭(9, `TAB`)도 그대로 유지됩니다.

2. **인코딩 방식**:
   - 비출력 가능 문자, 즉 7비트 ASCII 범위를 벗어나는 문자나 제어 문자는 `=` 기호 다음에 16진수로 인코딩됩니다.
   - 예를 들어, 문자 `=` 자체는 `=3D`로 인코딩됩니다.

3. **줄 길이 제한**:
   - 줄 길이가 76 문자를 초과하면 등호(`=`)를 줄 끝에 추가하여 줄 바꿈을 표시합니다. 이는 긴 줄이 전송 중에 잘리는 것을 방지합니다.

#### Quoted-Printable 인코딩이 필요한 이유

1. **7비트 전송 시스템 호환성**:

   전통적인 이메일 시스템은 7비트 전송을 기반으로 설계되었습니다.
   하지만 현대 텍스트 데이터는 8비트를 사용하는 경우가 많아 7비트 전송 시스템과 호환되지 않습니다.
   QP 인코딩은 8비트 데이터를 7비트로 안전하게 변환하여 전송할 수 있도록 합니다.

2. **데이터 손상 방지**:

   이메일 전송 과정에서 제어 문자가 잘못 해석되거나, 특정 문자들이 전송 도중 손상되는 것을 방지합니다.
   QP 인코딩은 제어 문자나 비출력 가능 문자를 안전하게 인코딩하여 전송합니다.

3. **사람이 읽을 수 있는 형태 유지**:

    QP 인코딩은 가능한 한 출력 가능한 ASCII 문자(예: 알파벳, 숫자)를 그대로 유지하므로, 인코딩된 텍스트도 사람이 읽을 수 있는 형태로 남습니다.

#### Quoted-Printable 인코딩 사용 시점

- **이메일 본문 인코딩**:

    이메일 클라이언트가 본문 텍스트를 QP 인코딩하여 전송합니다.
    이는 본문이 특수 문자를 포함하거나, 8비트 문자를 포함하는 경우에 주로 사용됩니다.

- **SMTP 프로토콜**:

    Simple Mail Transfer Protocol(SMTP)은 7비트 전송을 기반으로 하므로,
    8비트 데이터(특히 다국어 문자)를 전송할 때 QP 인코딩이 필요합니다.

- **MIME(Multipurpose Internet Mail Extensions)**:

    MIME은 이메일에서 텍스트와 비텍스트 데이터를 다룰 수 있게 하며,
    QP 인코딩은 MIME의 텍스트 데이터 인코딩 방식 중 하나로 사용됩니다.

#### 인코딩 방식

- QP 인코딩에서 비출력 가능 문자는 `=` 기호 다음에 각 니블을 표현하는 16진수 숫자 두 개를 추가하여 8비트 값을 표현합니다.
- 예를 들어, ASCII 코드 61(`=`)은 QP 인코딩에서 `=3D`로 표현됩니다.

"각 니블을 표현하는 16진수 숫자 두 개"를 조금 더 자세하게 설명하면 다음과 같습니다.

- **니블(Nibble)**:
    - 8비트(1바이트)는 두 개의 4비트 블록으로 나눌 수 있습니다. 각 4비트 블록을 니블(nibble)이라고 부릅니다.
    - 예를 들어, 8비트 값 `01011100`은 두 개의 니블로 나눌 수 있습니다: `0101`과 `1100`.

- **16진수 표현**:
    - 각 니블을 16진수로 표현하면, 4비트 값은 0부터 F까지의 16진수 값으로 변환됩니다.
    - 예를 들어, `0101`은 16진수로 `5`, `1100`은 16진수로 `C`입니다. 따라서, `01011100`은 16진수로 `5C`가 됩니다.

#### Quoted-Printable 인코딩의 예

- 일반 텍스트

    출력 가능한 문자들은 그대로 유지됩니다.

    원본 텍스트:

    ```plaintext
    Hello, World!
    ```

    Quoted-Printable 인코딩:

    ```plaintext
    Hello, World!
    ```

- 특수 문자 포함

    문자 `=`는 `=3D`로 인코딩됩니다.

    - 문자: `=`
    - ASCII 코드: 61
    - 2진수: `00111101`
    - 두 개의 니블: `0011` (16진수 3), `1101` (16진수 D)
    - QP 인코딩: `=3D`

    원본 텍스트:

    ```plaintext
    Hello, World! = Hello, again!
    ```

    Quoted-Printable 인코딩:

    ```plaintext
    Hello, World! =3D Hello, again!
    ```

- 8비트 문자 포함

    'ñ'는 UTF-8로 `C3 B1`이므로 각각 `=C3`와 `=B1`로 인코딩됩니다.

    - 문자: `ñ` (Unicode U+00F1, UTF-8로 인코딩 시 0xC3 0xB1)
    - UTF-8 바이트 1: 0xC3 (2진수: 11000011, QP 인코딩: `=C3`)
    - UTF-8 바이트 2: 0xB1 (2진수: 10110001, QP 인코딩: `=B1`)
    - QP 인코딩 결과: `=C3=B1`

    원본 텍스트(UTF-8 인코딩된 'ñ'):

    ```plaintext
    Piñata
    ```

    Quoted-Printable 인코딩:

    ```plaintext
    Pi=C3=B1ata
    ```

#### Quoted-Printable 인코딩 과정

1. **문자열 분석**:
   - 문자열의 각 문자를 분석하여 출력 가능한 문자인지 확인합니다.
   - 출력 가능한 문자는 그대로 유지하고, 그렇지 않은 문자는 인코딩합니다.

2. **인코딩**:
   - 비출력 가능 문자는 `=` 기호 다음에 16진수로 인코딩합니다. 예를 들어, ASCII 코드 61(`=`)는 `=3D`로 인코딩됩니다.
   - 줄 바꿈을 추가하여 각 줄의 길이가 76자를 넘지 않도록 합니다. 줄 바꿈 시 줄 끝에 `=`를 추가하여 줄 바꿈을 표시합니다.

#### Quoted-Printable 인코딩의 예시 코드

Python에서 Quoted-Printable 인코딩과 디코딩을 수행하는 예시 코드를 보여드리겠습니다.

- 인코딩

    ```python
    import quopri

    # 원본 텍스트
    text = "Hello, World! = Hello, again! Piñata"

    # Quoted-Printable 인코딩
    encoded_text = quopri.encodestring(text.encode('utf-8')).decode('utf-8')

    print(encoded_text)
    ```

- 디코딩

    ```python
    import quopri

    # Quoted-Printable 인코딩된 텍스트
    encoded_text = "Hello, World! =3D Hello, again! Pi=C3=B1ata"

    # Quoted-Printable 디코딩
    decoded_text = quopri.decodestring(encoded_text.encode('utf-8')).decode('utf-8')

    print(decoded_text)
    ```

### base64

Base64 인코딩은 데이터를 안전하게 전송하거나 저장할 수 있도록 바이너리 데이터를 ASCII 문자열로 변환하는 과정입니다.
이 과정에서 6비트씩 데이터를 잘라서 각 6비트 조각을 64개의 ASCII 문자 중 하나로 변환합니다.
따라서, Base64 인코딩을 하면 데이터의 길이가 줄어드는 것이 아니라 오히려 길어집니다.

1. **3바이트(24비트) 단위로 분할**:
   - 원본 데이터를 3바이트씩 분할합니다. 3바이트는 3 * 8 = 24비트입니다.

2. **6비트 단위로 분할**:
   - 24비트를 6비트씩 네 개의 그룹으로 나눕니다.

3. **64개의 ASCII 문자로 변환**:
   - 각 6비트 조각을 Base64 인코딩 표에 따라 64개의 ASCII 문자 중 하나로 변환합니다.

4. **패딩 추가**:
   - 원본 데이터의 길이가 3바이트의 배수가 아닌 경우, 남은 비트를 0으로 채우고, 결과 문자열 끝에 `=` 문자를 추가하여 패딩합니다.

예를 들어 원본 데이터가 "Hello"라고 해보겠습니다.

- ASCII 값: 72, 101, 108, 108, 111
- 2진수: `01001000 01100101 01101100 01101100 01101111`

이 데이터는 6비트씩 나누어집니다:
- `010010 000110 010101 101100 011011 001101 101111`

Base64 인코딩 표에 따라 변환하면 `SGVsbG8=`가 됩니다.

#### Base64 인코딩의 길이 증가 원리

1. **인코딩 비율**:
   - 원본 데이터가 3바이트(24비트)인 경우, Base64 인코딩 결과는 4바이트(32비트)입니다.
   - 즉, 3바이트를 4바이트로 변환하므로, 데이터의 길이가 약 33% 증가합니다.

2. **패딩**:
   - 원본 데이터의 길이가 3의 배수가 아닌 경우, Base64 인코딩은 `=` 문자를 사용하여 패딩을 추가합니다.

이를 수식으로 나타내면 다음과 같습니다.

- 원본 데이터 길이: $\text{L}$ 바이트
- Base64 인코딩된 데이터 길이: $\left( \frac{4}{3} \times \text{L} \right)$ 바이트

예를 들어, 원본 데이터가 9바이트인 경우:
- 인코딩된 데이터 길이: $\left( \frac{4}{3} \times 9 \right) = 12$ 바이트가 됩니다.

UUID를 Base64로 인코딩하면 실제로 길이가 줄어드는 것처럼 보일 수 있습니다. 이를 이해하기 위해 UUID와 Base64 인코딩의 원리를 비교해보겠습니다.

#### 데이터는 늘어나지만 문자열은 줄어들 수 있음

가령 UUID(Universally Unique Identifier)는 128비트(16바이트) 길이의 식별자입니다.
일반적으로 16진수로 표현하며, 하이픈으로 구분된 8-4-4-4-12 형식으로 나타냅니다.

```plaintext
123e4567-e89b-12d3-a456-426614174000
```

이 UUID는 32개의 16진수 문자와 4개의 하이픈, 총 36개의 문자를 포함합니다.

이 UUI를 Base64 인코딩하면 데이터의 길이가 증가하지만, UUID와 같은 경우에는 전체 문자열 길이가 줄어드는 것처럼 보일 수 있습니다.
이는 UUID를 16바이트 바이너리 데이터로 간주하고 Base64로 인코딩할 때 발생합니다.

1. **UUID의 바이너리 표현**:
   - UUID는 128비트(16바이트)입니다.

2. **Base64 인코딩**:
   - 3바이트(24비트) 단위로 4바이트(32비트)로 인코딩됩니다.
   - 16바이트 UUID는 Base64로 인코딩될 때, $\left( \frac{16}{3} \times 4 \right) = 22$바이트가 됩니다.

```python
import uuid
import base64

# UUID 생성
original_uuid = uuid.UUID('123e4567-e89b-12d3-a456-426614174000')

# UUID를 바이너리로 변환
uuid_bytes = original_uuid.bytes

# Base64 인코딩
base64_encoded = base64.urlsafe_b64encode(uuid_bytes).rstrip(b'=')

# 결과 출력
print("Original UUID:", str(original_uuid)) # 123e4567-e89b-12d3-a456-426614174000
print("Base64 Encoded UUID:", base64_encoded.decode('utf-8')) # Ej5FZ-ibEtOkVkJmFBdAAA
print("Length of Original UUID:", len(str(original_uuid))) # 36
print("Length of Base64 Encoded UUID:", len(base64_encoded)) # 22
```

1. **UUID 기본 형식의 길이**:
   - UUID는 16진수로 표현되므로, 각 16진수 문자는 4비트를 나타냅니다.
   - UUID 문자열은 32자(16진수) + 4자(하이픈) = 36자
   - 비트: 36자 (32자 * 4비트/자 = 128비트)

2. **Base64 인코딩의 길이**:
   - 128비트(16바이트) 데이터를 3바이트(24비트)씩 나누어 4바이트(32비트)로 인코딩합니다.
   - 16바이트 데이터를 Base64로 인코딩하면 $\left( \frac{16}{3} \times 4 \right) = 22$자
   - 비트: 22자 (22자 * 6비트/자 = 132비트)

UUID를 Base64로 인코딩하면 문자열 길이는 줄어들지만, 실제 데이터 크기(비트 수)는 약간 증가할 수 있습니다.
이는 UUID가 128비트(16바이트) 데이터이며, Base64 인코딩은 3바이트를 4바이트로 인코딩하기 때문에 발생합니다.
이로 인해 원래 36자의 UUID 문자열이 Base64 인코딩 후 22자로 줄어드는 것입니다.

따라서 UUID를 Base64로 인코딩하면 데이터의 크기는 늘어나지만, 텍스트로 표현되는 문자열의 길이는 줄어들 수 있습니다.
이는 Base64 인코딩이 효율적인 텍스트 표현 방법을 제공하기 때문입니다.
