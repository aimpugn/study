# forward

- [forward](#forward)
    - [forward?](#forward-1)
    - [fast-forward](#fast-forward)
    - [forward-compatible](#forward-compatible)
    - [forward declaration](#forward-declaration)
    - [forward secrecy](#forward-secrecy)
    - [forwarding (in networking)](#forwarding-in-networking)
    - [Forward Proxy vs. Reverse Proxy](#forward-proxy-vs-reverse-proxy)

## forward?

"Forward"라는 개념은 개발과 컴퓨팅 분야에서 다양한 맥락에서 사용되며, 여러 용어와 결합하여 특정 기술적 작업이나 원칙을 나타낸다.
"fast-forward"와 "forward compatibility", "forward declaration", "forward secrecy", "forwarding" 등이 있다.

## fast-forward

"Fast-forward"라는 용어는 여러 맥락에서 사용될 수 있으나, 소프트웨어 개발, 특히 버전 관리 시스템에서 자주 사용된다.
가장 대표적으로 Git 같은 분산 버전 관리 시스템에서 볼 수 있는 용어다.

Git에서 "fast-forward"는 특정 브랜치가 다른 브랜치의 최신 커밋을 가리킬 때, 추가적인 커밋 없이 후자를 전자로 "빠르게 앞당겨" 업데이트하는 과정을 의미한다. 이는 두 브랜치 사이에 충돌이나 추가적인 변화가 없을 때 가능하며, 병합 과정을 간소화하여 히스토리를 깔끔하게 유지할 수 있게 한다.

예를 들어, `feature` 브랜치를 `master` 브랜치로 병합할 때, `feature` 브랜치의 모든 변경 사항이 `master` 브랜치의 현재 상태 이후에 이루어졌다면, Git은 단순히 `master` 브랜치 포인터를 `feature` 브랜치의 최신 커밋으로 "fast-forward"할 수 있다.

## forward-compatible

> Deprecated: Use RegisteredClaims instead for a *forward-compatible* way to access registered claims

"Forward-compatible way"라는 표현은 특정 기술, 시스템, 플랫폼, 라이브러리, API, 프로토콜, 소프트웨어 구성요소 등이 미래의 버전이나 개발에도 계속 호환될 수 있도록 설계되었다는 것을 의미한다. 즉, 현재 사용하는 기능이나 방법이 나중에 나올 업데이트나 버전에서도 문제 없이 작동하도록 만들어졌다는 것이다.

소프트웨어 개발에서 "forward compatibility"는 주로 라이브러리나 API를 설계할 때 중요하게 고려되는 요소다.
예를 들어, 새로운 기능을 추가하거나 API를 변경할 때, 이전 버전의 클라이언트 응용 프로그램이 계속해서 현재 버전의 API와 상호 작용할 수 있도록 설계하는 것이 중요하다. 이를 통해 클라이언트 측에서 대규모의 업데이트나 변경을 요구하지 않고도, 서버 측에서의 개선이나 기능 추가가 가능해진다.

개발자들은 종종 미래의 변경 사항을 예측하여 현재의 코드나 시스템이 그 변경에도 유연하게 대응할 수 있도록 한다.
이를 통해 나중에 업데이트를 할 때 기존의 코드나 시스템을 대대적으로 수정하지 않아도 되는 이점이 있다.

여기서 언급된 "Use RegisteredClaims instead for a forward-compatible way to access registered claims"라는 문장은, 미래의 변경사항이나 업데이트에도 불구하고 등록된 클레임을 안정적으로 액세스하기 위한 방법으로 `RegisteredClaims`를 사용하라는 권장 사항이다.
이는 소프트웨어 라이브러리나 API를 사용할 때 특정 기능이나 구성요소가 미래의 업데이트에서도 문제없이 작동하도록 하는 방식을 선택하라는 의미를 담고 있다.
이렇게 하면 추후에 시스템이나 응용 프로그램을 업데이트할 때 발생할 수 있는 호환성 문제를 최소화할 수 있다.

## forward declaration

프로그래밍 언어, 특히 C나 C++에서 주로 사용된다.

함수나 변수, 클래스 등을 실제로 구현하기 전에 미리 선언하는 것을 말한다.
이를 통해 컴파일러가 해당 식별자의 존재를 인식할 수 있게 하며, 나중에 코드의 다른 부분에서 해당 식별자를 사용할 수 있도록 한다.
예를 들어, 두 클래스가 서로를 참조할 때, forward declaration을 사용하여 순환 참조 문제를 해결할 수 있다.

## forward secrecy

암호학, 네트워크 보안에서 사용된다.

"Forward secrecy" (또는 "perfect forward secrecy", PFS)는 현재 사용 중인 세션 키가 공개되더라도 과거의 통신 내용이 보호되는 보안 원칙이다. 이는 각 세션마다 고유한 세션 키를 생성하고, 통신이 종료되면 해당 키를 폐기함으로써 달성된다.
이 방식으로, 공격자가 장기적인 비밀 키를 손에 넣었다 하더라도, 과거의 통신 내용을 복호화할 수 없다.

## forwarding (in networking)

네트워크, 서버 관리에서 사용된다.

"Forwarding"은 네트워크 패킷이 목적지로 전송될 때 사용되는 용어다.
예를 들어, 이메일 포워딩이나 포트 포워딩에서 볼 수 있다.
"포트 포워딩"은 네트워크 트래픽을 특정 IP 주소와 포트 번호로 전달하는 과정을 의미하며, 보통 NAT (Network Address Translation) 뒤에 있는 로컬 네트워크의 서버에 외부에서 접근할 수 있게 하는 데 사용된다.

## Forward Proxy vs. Reverse Proxy

네트워크, 웹 개발에서 사용된다.

프록시 서버는 클라이언트와 서버 사이의 중계 역할을 합니다.
- "Forward proxy": 클라이언트가 인터넷상의 다양한 서버에 접속할 때 중계 역할을 하며, 사용자의 실제 IP 주소를 숨기거나 콘텐츠 필터링, 접근 제어 등의 기능을 수행한다
- "Reverse proxy": 인터넷에서 서버로의 요청을 중계하며, 로드 밸런싱, 캐싱, SSL 암호화 등을 제공하여 서버의 보안과 효율성을 높인다.
