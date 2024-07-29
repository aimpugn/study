# JDKs

## 너무나도 다양한 JDK

- Java 개발 키트(JDK)는 Java 프로그램을 개발하고 실행하는 데 필요한 소프트웨어의 모음
- 기본적으로 JDK는 아래와 같은 것들을 포함
    - Java 컴파일러(`javac`)
    - Java 런타임 환경(JRE)
    - 클래스 라이브러리 및 기타 도구
- 다양한 조직에서 자체 JDK를 제공하는 이유?
    - **라이선스 및 지원**: 오라클 JDK는 예전에는 무료였지만, 오라클이 상업적 사용에 대해 유료 라이선스 정책을 도입하면서 많은 조직이 무료 대안을 찾기 시작
    - **특정 요구사항 충족**: 일부 JDK는 특정 환경이나 요구사항(예: 가벼움, 빠른 시작, 장기 지원, 특정 시스템 최적화)에 맞추어 설계
    - **공개 기여 및 커뮤니티 주도**: 오픈소스 커뮤니티 주도의 JDK는 더 넓은 공개 기여와 혁신을 가능하게 합한다

## JDK 종류

다양한 JDK의 특징과 차이점은 다음과 같습니다:

### **Amazon Corretto**

- 아마존에서 지원하는 무료, 멀티플랫폼, 상용 지원 JDK
- 오라클 JDK와 호환되며 장기 지원(LTS)을 제공

### **Azul Zulu Community**

- Azul 시스템즈에서 제공하는 무료 오픈소스 JDK
- 상용 버전인 Zulu Enterprise도 제공하며, 이는 추가 성능 향상 및 지원 옵션을 포함

### **Eclipse Temurin (이전 AdoptOpenJDK)**

- 이클립스 재단에서 관리하는 오픈소스 JDK
- 커뮤니티 주도로 개발되며 다양한 플랫폼을 지원

### **BellSoft Liberica JDK**

- 모든 주요 운영 체제 및 클라우드 환경을 지원하는 JDK
- Liberica는 BellSoft에 의해 지원되며, OpenJDK의 공식 릴리즈를 기반으로 한다

### **JetBrains Runtime**

- JetBrains Runtime (JBR)은 IntelliJ IDEA를 비롯한 JetBrains IDE에 최적화된 OpenJDK 기반의 JDK
- JetBrains 툴과의 통합을 위해 특정 수정과 개선이 이루어진 버전
- 여러 버전과 변종이 있는데, 이는 다양한 사용 사례와 개발 요구에 맞추어 최적화되어 있다
    - **SDK 포함 런타임 (JBRSDK)**: 개발자에게 추가적인 기능을 제공하지만, 평균 사용자에게는 필요하지 않은 기능을 포함하여 더 큰 설치 크기를 가질 수 있다
    - **JCEF 포함 런타임**: 더 풍부한 UI를 제공하지만, 추가적인 시스템 자원을 소모할 수 있다
    - **fastdebug 옵션**: 디버깅에 유리하지만, 일반 사용에 있어서는 성능이 저하될 수 있다
    - **vanilla 런타임**: 가장 기본적인 버전으로, 특별한 요구 사항이 없는 사용자에게 적합하며, 다른 변종에 비해 가장 가볍고 간단하다
- J[BR의 배포된 버전들은 깃허브](https://github.com/JetBrains/JetBrainsRuntime/releases)에서 볼 수 있다

#### **JetBrains Runtime (JCEF)**

- JetBrains가 개발한 자체 JDK의 변형 중 하나
- `JCEF`는 **J**ava **C**hromium **E**mbedded **F**ramework의 약자
- JetBrains IDE에서 HTML, CSS, JavaScript를 사용하여 UI 부분을 구현할 수 있게 해주는 라이브러리
- 기본적으로, 이는 Chromium 기반의 웹 브라우저 엔진을 Java 애플리케이션 내에 내장시켜주는 것
- JCEF를 사용하면 개발자는 웹 기술을 활용하여 리치한 사용자 인터페이스를 만들 수 있으며, 이를 통해 웹 컨텐츠와 상호 작용하는 데스크탑 애플리케이션을 만들 수 있다
- JetBrains Runtime에 JCEF를 포함시킨 이유는, IntelliJ와 같은 IDE에서 더 나은 웹 콘텐츠 렌더링과 현대적인 웹 기술을 사용할 수 있도록 지원하기 위함이다
- 이로써 JetBrains의 IDE 사용자는 웹 기반의 플러그인을 사용하거나, 웹 페이지를 IDE 내에서 직접 볼 수 있는 등의 확장된 기능을 사용할 수 있게 된다

#### **JetBrains Runtime JBRSDK with JCEF**

- **JBRSDK**:
    - "JetBrains Runtime Software Development Kit"를 의미
    - 표준 JBR에 추가적인 개발 도구와 라이브러리가 포함되어 있다
- **with JCEF**
    - "Java Chromium Embedded Framework"가 포함되어 있음을 의미
    - 이를 통해 브라우저 기반의 UI 구성 요소를 IDE 내에서 사용할 수 있다

#### **JetBrains Runtime JBR with JCEF (fastdebug)**

- **JBR with JCEF**
    - 이것은 표준 JBR에 JCEF가 포함된 버전
- **fastdebug**: 이
    - 것은 디버그 버전의 JBR
    - 추가적인 디버그 정보와 검증이 포함되어 있어, 개발자가 런타임을 더 쉽게 디버깅할 수 있게 해준다
    - 그러나 일반 사용자에게는 일반 런타임보다 성능이 떨어질 수 있다

#### **JetBrains Runtime JBRSDK**

- 이것은 표준 JBR에 추가적인 개발 도구와 라이브러리가 포함된 버전
- 개발자가 IntelliJ IDEA 플러그인이나 관련 도구를 개발할 때 유용하다. 단, JCEF는 포함되어 있지 않다

#### **JetBrains Runtime JBR (vanilla)**

- **vanilla**
    - "기본", "표준" 또는 "변경되지 않은" 버전을 의미
    - 즉, JBR의 기본 버전이며, JCEF나 추가 개발 도구가 포함되어 있지 않다

### **Oracle JDK**

- 오라클이 제공하는 상용 JDK로, 과거에는 표준으로 많이 사용
- 현재는 상업적 사용을 위해 유료 라이선스가 필요

### **SAP SapMachine**

- SAP의 요구 사항을 충족하기 위해 SAP에서 만든 JDK
- SAP 애플리케이션과의 호환성에 중점을 두고 있다
