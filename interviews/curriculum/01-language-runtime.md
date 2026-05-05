# 01. 언어와 런타임

언어 문법이 아니라 코드가 실행 단위가 되어 메모리, GC, 클래스 로딩, 런타임 스케줄러를 만나는 지점을 다룹니다.

> 원문 배치본입니다. source chunk의 문장은 유지하고, 대분류/중분류/소분류 계층에 맞게 Markdown heading depth만 조정했습니다. 원본 span과 SHA-256은 manifest에서 검증할 수 있습니다.

## Class loading과 실행 준비

### 부모 위임 모델

#### 원문: 부모 위임 모델

<!-- curriculum-chunk: sha256=9c428c3213103c7d65c1545975794a9d752dc7f0745326de57513d415813574d major=language-runtime mid=Class loading과 실행 준비 sub=부모 위임 모델 sources=interview_questions.md:10148-10266, interviews.md:10096-10214 -->

> Source: `interview_questions.md:10148-10266`
> Classification reason: class loading
> Duplicate source aliases: `interview_questions.md:10148-10266, interviews.md:10096-10214`

##### 부모 위임 모델

Java의 클래스 로딩(Class Loading) 모델 중 하나로, 클래스 로더(ClassLoader)들이 계층적인 구조를 가지며, 클래스를 로드할 때 먼저 부모 클래스 로더에게 위임(delegate)하는 방식을 따릅니다.

이 방식은 Java의 기본적인 보안 모델과 클래스 충돌 방지를 위해 설계된 구조입니다.

> 💡 핵심 원칙:
> "클래스를 로드하려면, 먼저 부모 ClassLoader에게 위임하고, 부모가 해당 클래스를 찾지 못할 때만 자신이 직접 로드한다."

이 모델을 활용하면, JVM의 기본 클래스(`java.lang.*`, `java.util.*` 등) 및 공유 클래스가 일관되게 유지되며, 서로 다른 클래스 로더 간의 중복 로딩 및 충돌을 방지할 수 있습니다.

- 부모 위임 모델의 동작 방식

    🔹 클래스 로딩의 기본 흐름
    1️⃣ 애플리케이션에서 특정 클래스를 로드하려고 할 때,
    2️⃣ 해당 요청을 가장 상위 부모 클래스 로더(Bootstrap ClassLoader)에게 위임
    3️⃣ 부모 클래스 로더가 해당 클래스를 찾으면 반환, 찾을 수 없으면 자식(ClassLoader)에게 요청을 넘김
    4️⃣ 최종적으로 해당 클래스를 로드할 수 있는 클래스 로더가 직접 로드하여 반환

    이러한 재귀적 방식으로 클래스를 탐색하며, 부모가 먼저 클래스를 찾을 기회를 가지므로, Java의 핵심 클래스(`java.lang.String` 등)가 항상 표준 클래스 로더에서 먼저 로드되는 구조를 유지할 수 있습니다.

- Tomcat의 ClassLoader 계층 구조

    Tomcat의 `ClassLoader`는 부모 위임 모델을 따르며, 다음과 같은 계층을 가집니다.

    | ClassLoader | 설명 | 클래스 로딩 경로 |
    |---------------|-----------------------------------|-------------------------------|
    | Bootstrap ClassLoader | JDK의 핵심 클래스를 로드 | `$JAVA_HOME/lib/rt.jar` (Java 8 이하) |
    | System ClassLoader | `java.ext.dirs`에 위치한 라이브러리 로드 | `$JAVA_HOME/lib/ext` |
    | Common ClassLoader | Tomcat 내부 및 공유 클래스 로드 | `$CATALINA_HOME/lib` |
    | Catalina ClassLoader | Tomcat의 핵심 실행 코드 로드 | `$CATALINA_HOME/lib/catalina.jar` |
    | Shared ClassLoader | 모든 애플리케이션이 공유하는 라이브러리 로드 | `$CATALINA_HOME/shared/lib` |
    | Webapp ClassLoader | 각 웹 애플리케이션별 클래스 로드 | `WEB-INF/classes` 및 `WEB-INF/lib/*.jar` |

    📌 동작 예시:
    - `java.lang.String`을 로드할 경우:
        → Bootstrap ClassLoader에서 로드 (JVM 핵심 클래스이므로 최상위에서 관리)
    - `javax.servlet.http.HttpServlet`을 로드할 경우:
        → Common ClassLoader에서 로드 (Tomcat의 공유 클래스이므로 공통 로드)
    - `com.example.MyController`를 로드할 경우:
        → Webapp ClassLoader에서 로드 (각 애플리케이션별 고유한 클래스이므로 개별 관리)

- 부모 위임 모델을 사용하는 이유 (Why Parent Delegation Model?)

    1. 보안(Security) 강화
        - JVM의 핵심 클래스(`java.lang.Object`, `java.lang.String`)를 애플리케이션에서 재정의하지 못하도록 보호
        - 악의적인 코드가 JVM의 기본 클래스를 덮어쓰거나 변조하는 것을 방지

        📌 예제: `java.lang.String`을 재정의하는 악성 코드 차단

        ```java
        package java.lang;

        public class String {
            public static void main(String[] args) {
                System.out.println("Hacked!");
            }
        }
        ```

        ➡ 부모 위임 모델 덕분에, `Bootstrap ClassLoader`에서 `java.lang.String`을 먼저 찾고, 위 코드를 무시하여 보안을 유지함.

    2. 클래스 충돌 방지
        - 서로 다른 애플리케이션이 같은 라이브러리를 사용할 때, 중복 로딩을 방지
        - 모든 애플리케이션이 공통적으로 사용하는 클래스(`javax.servlet.*`, `javax.sql.*`)는 공용 `ClassLoader`에서 로드하여 불필요한 메모리 낭비 방지

        📌 예제: 서블릿 API 충돌 방지
        - 만약 `javax.servlet.Servlet`이 각 애플리케이션의 `WEB-INF/lib`에서 로드된다면, 서로 다른 버전이 존재할 가능성이 있음.
        - 부모 위임 모델 덕분에, Tomcat의 공통 ClassLoader에서 한 번만 로드됨.

    3. 애플리케이션 간 클래스 격리 (Class Isolation)
        - 서로 다른 웹 애플리케이션이 같은 패키지를 사용하더라도 클래스 로더가 다르면 서로 독립적으로 동작
        - 애플리케이션 간 의도치 않은 클래스 공유 및 간섭을 방지

        📌 예제: 두 개의 WAR 파일이 같은 패키지 구조를 사용할 때

        ```sh
        webapps/
        ├── app1.war
        │   ├── WEB-INF/classes/com/example/MyService.class
        │   ├── WEB-INF/lib/library-1.0.jar
        ├── app2.war
        │   ├── WEB-INF/classes/com/example/MyService.class
        │   ├── WEB-INF/lib/library-2.0.jar
        ```

        ➡ `app1.war`의 `MyService.class`가 `library-1.0.jar`을 사용하고,
        ➡ `app2.war`의 `MyService.class`가 `library-2.0.jar`을 사용함.
        ✅ 부모 위임 모델 덕분에, 각 애플리케이션이 자신의 라이브러리만 참조하도록 보장됨.

- 부모 위임 모델의 한계 및 해결 방법

    1. 특정 라이브러리를 무조건 부모에서 찾도록 강제하는 문제
        - Spring Boot의 경우, `WEB-INF/lib`에 있는 특정 라이브러리를 우선적으로 로드하고 싶을 때 문제가 발생할 수 있음.

        ✅ 해결 방법: `tomcat.util.scan.StandardJarScanner` 설정 변경

        ```xml
        <Context>
            <JarScanner scanClassPath="false"/>
        </Context>
        ```

        ➡ `scanClassPath="false"`로 설정하면, 부모 클래스 로더에서 라이브러리를 찾지 않고, 애플리케이션의 `WEB-INF/lib`에서 우선 로드함.

    2. Spring Boot와 같은 Fat JAR 환경에서 문제 발생
        - Spring Boot는 `fat JAR`로 실행되므로, Tomcat의 기본적인 `ClassLoader`와 충돌할 가능성이 있음.
        - 이를 해결하기 위해 Spring Boot의 `LaunchedURLClassLoader`가 사용되며, 내부적으로 클래스를 동적으로 로드함.

        ✅ 해결 방법: `Spring Boot ClassLoader`를 활용

        ```java
        ClassLoader cl = Thread.currentThread().getContextClassLoader();
        System.out.println(cl.getClass().getName());
        // org.springframework.boot.loader.LaunchedURLClassLoader
        ```

        ➡ Spring Boot는 자체 ClassLoader를 사용하여 부모 위임 모델을 일부 우회함.

<!-- /curriculum-chunk -->

## JVM 실행 모델

### 1. Java 관련 질문

#### 원문: 1. Java 관련 질문

<!-- curriculum-chunk: sha256=bd3bb9b7fec3aa36236c525e46c19775fa6a88eec9d187c0697824747c9fb0f3 major=language-runtime mid=JVM 실행 모델 sub=1. Java 관련 질문 sources=interview_questions3.md:92-129 -->

> Source: `interview_questions3.md:92-129`
> Classification reason: java question cluster

##### 1. Java 관련 질문

###### 질문 1. 자바에서 메모리는 어떻게 구성되어 있으며, 각 영역은 어떤 목적을 가지나요?

###### **답변**

1. 자바 애플리케이션이 실행될 때, **JVM(Java Virtual Machine)**은 크게 여러 메모리 영역을 관리합니다. 그중 대표적으로 **힙(Heap) 영역**, **스택(Stack) 영역**, 그리고 **메타스페이스(Metaspace)** 영역이 존재합니다.
2. **힙(Heap) 영역**은 객체와 배열이 동적으로 할당되는 곳입니다. 자바의 모든 객체(예: new 키워드로 생성된 인스턴스)가 대부분 힙에 위치합니다. 가비지 컬렉터(GC)가 이 영역을 주기적으로 검사하여, 더 이상 참조되지 않는 객체를 해제합니다.
3. **스택(Stack) 영역**은 각 쓰레드마다 분리되어 있습니다. 한 쓰레드 안에서 메서드가 호출될 때마다 스택 프레임이 쌓이고, 지역 변수와 메서드 실행 컨텍스트가 이 스택 프레임에 보관됩니다. 메서드가 종료되면 해당 스택 프레임이 사라지면서 그 안의 지역 변수가 해제됩니다.
4. **메타스페이스(Metaspace)**는 자바 8부터 도입된 영역으로, 이전에 PermGen으로 불렸던 부분을 대체합니다. 클래스 메타데이터, 리플렉션 정보, 동적 프록시 정보 등을 저장합니다. 이 영역은 OS 메모리를 동적으로 사용할 수 있으므로, PermGen과 달리 OutOfMemoryError 위험이 어느 정도 줄었습니다.
5. 이처럼 자바 메모리가 여러 영역으로 나뉘어 있는 이유는, **가비지 컬렉션 효율성**과 **프로그램 구조화**에 있습니다. 힙은 객체를 통합적으로 관리하고, 스택은 메서드 호출 구조와 밀접하게 연관되며, 메타스페이스는 클래스 정보를 별도로 관리함으로써 JVM이 메모리를 더욱 체계적으로 운용할 수 있게 해줍니다.

---

###### 질문 2. 자바 가비지 컬렉션 알고리즘에는 어떤 종류가 있으며, 어떤 특징을 갖나요?

###### **답변**

1. 자바는 객체 생명주기를 자동으로 관리하기 위해 가비지 컬렉션(Garbage Collection)을 사용합니다. 가비지 컬렉션 알고리즘이란, **더 이상 사용되지 않는(참조되지 않는) 객체를 탐색**하여 힙에서 제거하는 로직을 의미합니다.
2. Java 8까지 자주 사용되던 알고리즘으로는 **Serial GC**, **Parallel GC**, **CMS(Concurrent Mark-Sweep)** 등이 있습니다. Serial GC는 단일 쓰레드로 동작하여 간단하지만, 멀티코어 환경에서는 확장성이 떨어집니다. Parallel GC는 여러 쓰레드를 통해 동시에 객체를 수집하기 때문에 처리량(Throughput)이 높아집니다. CMS는 애플리케이션 스레드와 GC 스레드를 부분적으로 병행하여(Concurrent) 멈추는 시간(Stop-the-world pause)을 줄이려 합니다.
3. Java 9 이상에서는 **G1 GC**가 기본이 되었습니다. G1(Garbage First) GC는 힙을 여러 영역(Region)으로 나누고, 객체 참조도를 추적하여 우선적으로 수거해야 하는 영역부터 처리(garbage first)합니다. 이 방식은 대규모 힙에서 멈춤 시간을 줄이는 데 효과적입니다.
4. 최근에는 **ZGC**나 **Shenandoah** 같은 더 최신의 저지연 GC도 등장했습니다. 이들은 매우 큰 힙에서도 짧은 멈춤 시간(Stop-the-world)을 추구합니다. 예를 들어 ZGC는 큰 힙(수십 GB~TB급)에서도 최대 몇 ms 정도의 멈춤 시간을 목표로 합니다.
5. 이러한 다양한 GC 알고리즘이 존재하는 이유는, **애플리케이션 특성**(예: 대규모 서비스, 짧은 응답 시간 요구, CPU 자원 크기)에 따라 GC가 성능에 큰 영향을 미치기 때문입니다. 실제 서비스에서는 GC 로그를 모니터링하여, 적절한 GC 모드를 선택하고 필요하면 JVM 옵션을 조정해 튜닝합니다.

---

###### 질문 3. Java 동시성(멀티쓰레드) 프로그래밍에서 주의해야 할 점은 무엇인가요?

###### **답변**

1. 멀티쓰레드 환경에서는 **공유 자원(shared resource)**에 여러 쓰레드가 동시에 접근할 수 있기 때문에, 동기화(synchronization) 문제가 핵심입니다. 즉, 서로 다른 쓰레드가 같은 데이터를 변경하는 순간, 올바르지 않은 결과가 발생할 수 있습니다(예: Race Condition).
2. 이를 방지하기 위해, 자바는 `synchronized` 키워드나 `ReentrantLock` 클래스를 통해 한 번에 하나의 쓰레드만 임계영역(critical section)에 들어가도록 제한합니다.
3. 또한 **volatile** 키워드는 변수의 가시성(visibility)을 보장하기 위한 것입니다. volatile로 선언된 변수에 쓰기가 일어나면, 모든 쓰레드가 즉시 최신 값을 확인하도록 보장합니다.
4. 자바 메모리 모델(JMM)에 따르면, 쓰레드 간에 값이 즉시 전파되지 않을 수도 있고(캐시 혹은 레지스터 최적화 등), 리오더(reordering)가 발생할 수도 있습니다. synchronized나 volatile을 사용하면 happens-before 관계를 형성하여 이러한 문제를 해결할 수 있습니다.
5. **동시성 프로그래밍**은 CPU 코어 수가 늘어난 현대 환경에서 필수적이지만, 설계가 잘못되면 교착상태(Deadlock)나 리소스 경합(Contention)으로 인해 성능이 급격히 하락할 수 있습니다. 따라서 멀티쓰레드를 사용할 때는 임계구역 최소화, 락 분리, 락 프리 알고리즘(Atomic 클래스) 고려 등 세밀한 접근이 필요합니다.

---

<!-- /curriculum-chunk -->

### JVM, PHP runtime, Go runtime

#### 원문: JVM, PHP runtime, Go runtime

<!-- curriculum-chunk: sha256=9bc80b8dd4af126ddfb4d1a276d414438a416b2f930309548700c6bb286cdb95 major=language-runtime mid=JVM 실행 모델 sub=JVM, PHP runtime, Go runtime sources=interview_questions.md:5545-5614, interviews.md:5545-5614 -->

> Source: `interview_questions.md:5545-5614`
> Classification reason: jvm execution
> Duplicate source aliases: `interview_questions.md:5545-5614, interviews.md:5545-5614`

##### JVM, PHP runtime, Go runtime

1. Java의 실행 환경과 메모리 관리

    Java는 "Write Once, Run Anywhere" 철학을 구현하기 위해 JVM(Java Virtual Machine)을 사용합니다.

    JVM은 Java 바이트코드를 실행하는 가상 머신입니다.
    1. Java 소스코드(.java)가 컴파일되어 바이트코드(.class)로 변환됩니다.
    2. JVM이 이 바이트코드를 로드하고 실행합니다.
    3. JVM은 처음에는 인터프리터 방식으로 바이트코드를 실행합니다.
    4. JIT(Just-In-Time) 컴파일러가 자주 실행되는 코드(핫스팟)를 감지하고 네이티브 코드로 컴파일합니다.

    메모리 관리:
    - JVM은 가비지 컬렉션(GC)을 통해 자동으로 메모리를 관리합니다.
    - 다양한 GC 알고리즘(예: Serial, Parallel, CMS, G1)을 제공합니다.

2. PHP의 실행 환경과 메모리 관리

    PHP는 주로 웹 서버 환경에서 실행되며, 스크립트 언어로 설계되었습니다.
    엄밀히 말하면 VM은 아니지만, Zend Engine이 PHP 코드를 실행하는 역할을 합니다.

    실행 과정은 다음과 같습니다:
    1. PHP 소스코드가 Zend Engine에 의해 파싱되고 컴파일됩니다.
    2. 생성된 opcode가 실행됩니다.
    3. OPcache 확장을 사용하면 컴파일된 opcode를 메모리에 캐시하여 성능을 향상시킵니다.

    메모리 관리:
    - PHP는 참조 카운팅 방식의 가비지 컬렉션을 주로 사용합니다.
    - PHP 5.3 이후로는 순환 참조를 처리하기 위한 Cycle Collector도 도입되었습니다.

3. Go의 실행 환경과 메모리 관리

    Go는 컴파일 언어이지만, 컴파일된 네이티브 코드와 함께 런타임 시스템이 포함됩니다.

    실행 과정은 다음과 같습니다.
    1. Go 소스코드가 컴파일되어 실행 파일로 변환됩니다.
    2. 실행 파일에는 Go 런타임이 포함되어 있습니다.

    메모리 관리:
    - Go 런타임은 동시성 지원과 가비지 컬렉션을 담당합니다.
    - Go의 GC는 동시성을 고려하여 설계되었으며, 짧은 stop-the-world 시간을 가집니다.

Runtime과 VM은 다른 건가요? 다르다면 어떻게 다른가요?
- Runtime: 프로그램 실행에 필요한 라이브러리와 지원 기능들의 집합입니다.
- VM: 가상의 컴퓨터 환경을 소프트웨어로 구현한 것입니다.
- 차이점: VM은 일반적으로 더 복잡하고 완전한 실행 환경을 제공하며, 종종 중간 언어(예: 바이트코드)를 실행합니다. Runtime은 보통 더 가벼우며, 특정 언어나 플랫폼에 특화된 지원 기능을 제공합니다.

VM이 있는 Java 경우 VM이 메모리 할당과 GC를 전적으로 관리합니다.
VM이 없지만 GC는 존재하는 Golang의 경우 런타임 시스템의 일부로 GC가 구현되어 있습니다.
VM과 GC 모두 없는 C++ 경우 개발자가 직접 메모리를 관리하거나, 외부 GC 라이브러리를 사용할 수 있습니다.

1. Apache + mod_php와 nginx + php-fpm의 동작 방식은 어떻게 다른가요?
   - Apache + mod_php:
     1. Apache 웹 서버가 요청을 받습니다.
     2. mod_php가 Apache 프로세스 내에서 직접 PHP 코드를 실행합니다.
     3. 각 Apache 프로세스가 PHP 인터프리터를 포함하게 됩니다.

   - nginx + php-fpm:
     1. nginx 웹 서버가 요청을 받습니다.
     2. PHP 관련 요청을 FastCGI 프로토콜을 통해 php-fpm에 전달합니다.
     3. php-fpm이 별도의 프로세스 풀에서 PHP 코드를 실행합니다.
     4. 실행 결과를 nginx에 반환하고, nginx가 클라이언트에게 응답합니다.

   주요 차이점:
   - 리소스 사용: mod_php는 각 Apache 프로세스에 PHP를 로드하므로 메모리 사용량이 높을 수 있습니다. php-fpm은 필요한 만큼만 PHP 프로세스를 유지하므로 더 효율적일 수 있습니다.
   - 확장성: php-fpm은 PHP 프로세스를 독립적으로 관리할 수 있어 더 나은 확장성을 제공합니다.
   - 보안: php-fpm은 웹 서버와 PHP 실행 환경을 분리하여 추가적인 보안 계층을 제공할 수 있습니다.

이러한 차이점들로 인해 최근에는 nginx + php-fpm 조합이 더 많이 선호되는 추세입니다.

<!-- /curriculum-chunk -->

## Kotlin 언어 기능

### infix, inline, noinline, crossinline

#### 원문: infix, inline, noinline, crossinline

<!-- curriculum-chunk: sha256=d42f5b06f742e1addd1018ce374d6395a1bc25c2d300634df1feec047bf09026 major=language-runtime mid=Kotlin 언어 기능 sub=infix, inline, noinline, crossinline sources=interview_questions2.md:1499-1629, interviews2.md:1499-1629 -->

> Source: `interview_questions2.md:1499-1629`
> Classification reason: kotlin language
> Duplicate source aliases: `interview_questions2.md:1499-1629, interviews2.md:1499-1629`

##### infix, inline, noinline, crossinline

Kotlin에서 제공하는 `infix`, `inline`, `noinline`, `crossinline` 키워드는 각각 특정한 기능과 의도를 반영하여 설계된 특수한 언어 구성 요소입니다.

###### Infix

`infix`는 함수 호출을 보다 직관적이고 읽기 쉽게 하기 위해 제공되는 기능입니다.
"중간(infix)"에 위치한다는 의미에서 이름이 붙었습니다. 이는 기존의 전위(prefix), 후위(postfix) 표현과 구분됩니다.
일반적으로 객체 메서드는 `object.method(argument)` 형식으로 호출되지만, `infix`를 사용하면 중위 표현식(`object method argument`)으로 사용할 수 있습니다.
주로 DSL(Domain-Specific Language)나 간결한 표현식이 필요한 경우에 유용합니다.

사용할 수 있는 조건은 다음과 같습니다:
- 멤버 함수나 확장 함수여야 합니다.
- 함수는 하나의 매개변수만 받아야 합니다.
- `infix` 키워드로 선언되어야 합니다.

`infix`는 JVM에서 특별한 바이트코드를 생성하지 않습니다.
단순히 호출 구문을 간결하게 바꿀 수 있는 문법적 설탕(Syntactic Sugar)입니다.
JVM에서는 여전히 일반 메서드 호출로 처리됩니다.

사용례:
- 집합 연산 DSL:

    ```kotlin
    infix fun String.append(other: String): String = this + other

    val result = "Hello" append "World"
    println(result)  // "HelloWorld"
    ```

- Kotlin 표준 라이브러리의 `to` 함수:

    ```kotlin
    val map = mapOf(1 to "one", 2 to "two")
    ```

###### Inline

`inline` 함수는 호출 시점에 함수의 바디를 호출자 코드로 대체(inline)하여 런타임 오버헤드를 줄입니다.
"라인 안에 삽입"한다는 뜻에서 붙여졌습니다.
이는 JVM의 `method call overhead`를 줄이는 것을 목표로 합니다.
특히 람다식을 매개변수로 전달할 때 유용합니다.
고차 함수를 사용하는 코틀린 표준 라이브러리(`forEach`, `filter`, `map`)에서 성능 최적화를 위해 널리 사용됩니다.

컴파일러는 `inline`으로 표시된 함수를 호출하는 모든 지점에서 함수의 실제 코드를 삽입합니다.
람다식이 매개변수로 전달될 경우, 람다 표현식도 인라인됩니다.

장점:
- 함수 호출 비용을 제거하여 성능을 개선합니다.
- 람다 캡처로 인해 생성되는 추가 객체를 줄입니다.

단점:
- 코드 크기가 증가할 수 있습니다(Code Bloat).
- 지나치게 많은 인라인 함수는 성능에 부정적인 영향을 미칠 수 있습니다.

사용례:

- 고차 함수 최적화:

    ```kotlin
    inline fun performAction(action: () -> Unit) {
        println("Starting action")
        action()
        println("Ending action")
    }

    performAction {
        println("Executing action")
    }
    ```

    위 코드는 `performAction` 함수 바디를 호출자로 대체하여 추가 함수 호출을 방지합니다.

###### Noinline

`noinline`은 `inline` 함수의 매개변수 중 특정 람다식을 인라인 처리하지 않도록 지정합니다.
"인라인되지 않는다(no-inline)"는 것을 명확히 하기 위해 붙여졌습니다.
`inline` 함수에서 모든 람다가 인라인되는 기본 동작을 제어할 때 사용됩니다.
람다식을 동적으로 조작해야 하는 프레임워크에서 사용됩니다.

JVM은 다음과 같이 동작합니다.
- `noinline` 람다 매개변수는 일반적인 객체로 처리되며, 실제 호출 시점에 생성됩니다.
- 호출자는 해당 람다를 별도의 객체로 전달받아 사용합니다.

사용례:
- 동작 분리:

    ```kotlin
    inline fun example(block1: () -> Unit, noinline block2: () -> Unit) {
        block1()
        block2()
    }

    example({
        println("Inline block")
    }, {
        println("Noinline block")
    })
    ```

    위 예제에서 `block2`는 일반 객체로 처리됩니다.

###### Crossinline

`crossinline`은 `inline` 함수에서 람다식이 비지역(non-local) 리턴을 수행하지 못하도록 제한합니다.
"교차 컨텍스트에서의 인라인 제한"이라는 의미에서 이름이 붙었습니다.
보통 `inline` 함수 내부에서 람다를 다른 컨텍스트로 전달할 때 사용됩니다.
코루틴과 같은 비동기 처리에서 비지역 리턴을 방지하여 안전성을 높입니다.

JVM의 동작 방식은 다음과 같습니다:
- 람다의 리턴 동작을 제한하기 위해 추가적인 검사를 삽입합니다.
- `crossinline` 람다식은 단순히 값으로 캡처되며 비지역 리턴이 허용되지 않습니다.

사용례:
- 안전한 리턴 제한:

    ```kotlin
    inline fun safeExecution(crossinline action: () -> Unit) {
        println("Starting safe execution")
        val runnable = Runnable {
            action()  // 비지역 리턴 불가
        }
        runnable.run()
    }

    safeExecution {
        println("Executing safely")
        // return  // 컴파일 에러
    }
    ```

<!-- /curriculum-chunk -->

### kotlin의 `when` 표현식에서의 exhaustiveness 체크

#### 원문: kotlin의 `when` 표현식에서의 exhaustiveness 체크

<!-- curriculum-chunk: sha256=b6ad5f1248b2d1f9a43415eabf943d73258e8d96b61aad6169a0265e44c0142b major=language-runtime mid=Kotlin 언어 기능 sub=kotlin의 `when` 표현식에서의 exhaustiveness 체크 sources=interview_questions.md:3562-3673, interviews.md:3562-3673 -->

> Source: `interview_questions.md:3562-3673`
> Classification reason: kotlin language
> Duplicate source aliases: `interview_questions.md:3562-3673, interviews.md:3562-3673`

##### kotlin의 `when` 표현식에서의 exhaustiveness 체크

Kotlin에서 `when` 표현식의 exhaustiveness 체크는 컴파일러가 모든 가능한 경우를 처리하도록 보장합니다.
처리되지 않은 케이스로 인한 런타임 오류를 방지합니다.
이 기능은 주로 sealed 클래스나 enum과 함께 사용되지만, 다른 상황에서도 exhaustiveness를 달성할 수 있습니다.

- Sealed 클래스:

    sealed 클래스는 상속 가능한 클래스를 제한하여, 동일한 파일 내에서만 서브클래스를 정의할 수 있습니다.
    `when` 표현식에서 sealed 클래스를 사용하면 컴파일러는 모든 가능한 서브클래스를 알고 있으므로 exhaustiveness를 체크할 수 있습니다.

    ```kotlin
    sealed class PaymentResult
    class Success : PaymentResult()
    class Failure : PaymentResult()

    fun handleResult(result: PaymentResult) = when (result) {
        is Success -> // 성공 처리
        is Failure -> // 실패 처리
        // 'else' 필요 없음; 컴파일러가 모든 서브클래스를 알고 있음
    }
    ```

- Enums:

    enum은 고정된 상수 집합을 가지므로, `when` 표현식에서 enum을 사용할 때 컴파일러가 모든 케이스를 알고 있어 exhaustiveness를 체크할 수 있습니다.

    ```kotlin
    enum class PaymentMethod { CreditCard, PayPal, BankTransfer }

    fun processPayment(method: PaymentMethod) = when (method) {
        PaymentMethod.CreditCard -> // 신용카드 처리
        PaymentMethod.PayPal -> // 페이팔 처리
        PaymentMethod.BankTransfer -> // 은행 이체 처리
        // 'else' 필요 없음
    }
    ```

Sealed 클래스나 enum을 사용하지 않더라도 컴파일러가 exhaustiveness 체크를 할 수 있는 몇 가지 경우가 있습니다.

- Boolean 조건을 사용하는 경우

    `when`을 주제(subject) 없이 표현식으로 사용하고 모든 가능한 boolean 조건을 다루면,
    컴파일러가 이를 exhaustiveness로 간주할 수 있습니다.

    ```kotlin
    fun categorizeNumber(number: Int) = when {
        number > 0 -> "양수"
        number == 0 -> "영"
        number < 0 -> "음수"
        // 모든 케이스를 다룸
    }
    ```

- 알려진 타입으로 스마트 캐스팅하는 경우

    컴파일러가 스마트 캐스팅을 통해 모든 가능한 타입이 처리되었음을 알 수 있다면, exhaustiveness 체크가 가능합니다.

    ```kotlin
    interface PaymentGateway
    class PayPalGateway : PaymentGateway
    class CreditCardGateway : PaymentGateway
    class BankTransferGateway : PaymentGateway

    fun processGateway(gateway: PaymentGateway) = when (gateway) {
        is PayPalGateway -> // 페이팔 처리
        is CreditCardGateway -> // 신용카드 처리
        is BankTransferGateway -> // 은행 이체 처리
        else -> throw IllegalArgumentException("알 수 없는 게이트웨이")
    }
    ```

    이 경우, `PaymentGateway`가 열린 인터페이스이기 때문에 컴파일러는 exhaustiveness를 보장하지 않으므로 `else`가 필요합니다.

- Companion Object나 Object 선언을 사용하는 경우

    각 케이스를 나타내는 object 선언을 사용하고, 이를 `when` 표현식에서 사용할 때 모든 object를 다루면 exhaustiveness를 달성할 수 있습니다.

    ```kotlin
    object PayPalGateway : PaymentGateway
    object CreditCardGateway : PaymentGateway
    object BankTransferGateway : PaymentGateway

    fun processGateway(gateway: PaymentGateway) = when (gateway) {
        PayPalGateway -> // 페이팔 처리
        CreditCardGateway -> // 신용카드 처리
        BankTransferGateway -> // 은행 이체 처리
        // 'PaymentGateway'가 sealed인 경우 컴파일러가 exhaustiveness를 체크
    }
    ```

    하지만 `PaymentGateway`가 sealed 클래스나 인터페이스가 아니라면 컴파일러는 모든 구현체를 알지 못하므로 exhaustiveness를 보장하지 않습니다.

- Sealed 인터페이스 (Kotlin 1.5 이상)

    Kotlin 1.5부터 sealed 인터페이스를 지원합니다.
    인터페이스를 sealed로 정의하면, 컴파일러가 모든 구현체를 알고 exhaustiveness를 체크할 수 있습니다.

    ```kotlin
    sealed interface PaymentGateway
    object PayPalGateway : PaymentGateway
    object CreditCardGateway : PaymentGateway
    object BankTransferGateway : PaymentGateway

    fun processGateway(gateway: PaymentGateway) = when (gateway) {
        PayPalGateway -> // 페이팔 처리
        CreditCardGateway -> // 신용카드 처리
        BankTransferGateway -> // 은행 이체 처리
        // 'else' 필요 없음; 컴파일러가 exhaustiveness를 보장
    }
    ```

<!-- /curriculum-chunk -->

## 언어 런타임 비교

### Apache + mod_php와 Nginx + php-fpm의 차이

#### 원문: Apache + mod_php와 Nginx + php-fpm의 차이

<!-- curriculum-chunk: sha256=557e3f66a06efdd93b06acb1ad66f2703591d46278fbcb12a44072660501fe1c major=language-runtime mid=언어 런타임 비교 sub=Apache + mod_php와 Nginx + php-fpm의 차이 sources=interview_questions.md:5420-5544, interviews.md:5420-5544 -->

> Source: `interview_questions.md:5420-5544`
> Classification reason: runtime comparison
> Duplicate source aliases: `interview_questions.md:5420-5544, interviews.md:5420-5544`

##### Apache + mod_php와 Nginx + php-fpm의 차이

- Apache + mod_php:

    Apache는 요청을 처리할 때 사용하는 MPM(Multi-Processing Module) 방식을 통해 요청을 여러 프로세스나 쓰레드로 나눠 처리합니다.
    대표적인 MPM 방식에는 prefork, worker, event가 있습니다.

    각 Apache 프로세스는 PHP 코드를 실행할 수 있도록 PHP 인터프리터와 관련된 모든 데이터를 메모리에 로드해야 합니다.
    이는 PHP의 모듈, 확장 기능, 설정 정보 등이 포함되며, 각 프로세스가 독립적으로 이러한 자원을 할당받습니다.
    이 때문에 각 프로세스에 PHP 인터프리터가 중복으로 로드됩니다. 이는 PHP의 메모리 사용량이 각 프로세스마다 추가됩니다.

    - Pre-fork MPM:

        Apache는 미리 여러 개의 프로세스를 생성해 두고, 각 프로세스가 하나의 요청을 처리합니다.
        이때 각 요청은 독립적인 프로세스에서 처리되며, 동시에 여러 프로세스가 요청을 병렬적으로 처리합니다.

        각 프로세스는 독립적으로 실행되며, 쓰레드를 사용하지 않기 때문에 메모리 보호가 강력합니다.
        그러나 프로세스 간의 메모리 격리로 인해 메모리 사용량이 상대적으로 많습니다.

        mod_php를 사용하는 경우, PHP는 각 요청마다 새롭게 프로세스에서 실행됩니다.
        즉, 각 프로세스는 독립적이므로, 하나의 프로세스가 모든 요청을 처리하는 것은 아니며 여러 프로세스가 요청을 나눠서 처리합니다.

    - Worker MPM:

        Worker MPM은 프로세스와 쓰레드를 혼합하여 요청을 처리합니다.
        각 프로세스는 여러 개의 쓰레드를 실행하며, 각 쓰레드가 하나의 요청을 처리합니다.

        프로세스 수는 제한되지만, 각 프로세스 내에서 여러 쓰레드가 동작하기 때문에 메모리 사용량이 더 효율적입니다.
        다만, 쓰레드를 사용하는 만큼 동시성 문제가 발생할 수 있어 이를 적절히 관리해야 합니다.

        Worker MPM에서도 PHP는 mod_php를 통해 실행되지만, 여러 요청이 동일한 프로세스 내의 쓰레드에서 처리될 수 있습니다.

    - Event MPM:

        Event MPM은 Worker MPM과 유사하지만, 비동기 방식으로 동작하여 더 많은 동시 요청을 처리할 수 있습니다.
        연결이 활성 상태일 때만 쓰레드를 사용하고, 비활성 상태에서는 쓰레드를 해제하여 리소스를 절약합니다.

        비동기 처리가 가능해 더 많은 동시 연결을 효율적으로 처리합니다.

        Event MPM은 비동기 I/O에 최적화되어 있지만, PHP가 쓰레드 안전(thread-safe)하지 않다면 기본적으로 Worker MPM과 유사하게 작동합니다.

    요청을 처리하는 과정은 다음과 같습니다:
    1. 초기화 단계:
        - Apache 웹 서버가 시작될 때, mod_php 모듈이 로드됩니다.
        - 각 Apache 워커 프로세스(또는 스레드)에 PHP 인터프리터가 내장됩니다.
        - PHP 설정 파일(php.ini)이 읽히고 적용됩니다.

    2. 요청 수신:
        - 클라이언트가 HTTP 요청을 보냅니다.
        - Apache의 리스닝 소켓이 요청을 받아들입니다.

    3. 요청 처리:
        - Apache의 요청 처리 파이프라인이 시작됩니다.
        - URL을 분석하여 요청된 리소스가 PHP 파일인지 확인합니다.

    4. PHP 실행:
        - mod_php가 PHP 파일 실행을 담당합니다.
        - PHP 인터프리터가 파일을 읽고 파싱합니다.
        - PHP 코드가 Apache 프로세스의 컨텍스트 내에서 직접 실행됩니다.

    5. 데이터베이스 연결 (필요한 경우):
        - PHP 코드가 데이터베이스 연결을 요청하면, mod_php는 해당 워커 프로세스 내에서 연결을 생성합니다.
        - 이 연결은 다른 요청에서 재사용될 수 있습니다(영구 연결 설정에 따라).

    6. 출력 생성:
        - PHP 코드의 실행 결과(HTML, JSON 등)가 생성됩니다.
        - 출력은 Apache의 출력 버퍼로 직접 전달됩니다.

    7. 응답 전송:
        - Apache가 생성된 출력을 HTTP 응답으로 클라이언트에게 전송합니다.

    8. 정리:
        - 요청 처리가 완료되면, 사용된 리소스(메모리 등)가 해제됩니다.
        - 그러나 PHP 인터프리터는 메모리에 계속 로드된 상태로 유지됩니다.

- Nginx + php-fpm:

    Nginx는 가벼운 웹 서버로, PHP 처리를 php-fpm(PHP FastCGI Process Manager)에 위임합니다.
    즉, Nginx는 정적 파일이나 리버스 프록시 역할을 하면서, PHP 스크립트가 필요할 때는 PHP-FPM에 해당 작업을 요청합니다.

    PHP-FPM은 PHP를 전용 프로세스로 실행하며, 여러 개의 프로세스가 동시에 실행될 수 있어, Nginx와 PHP 처리 간에 명확한 분리가 이루어집니다.
    php-fpm이 별도의 프로세스 풀에서 PHP 코드를 실행합니다.

    1. 초기화 단계:
        - nginx 웹 서버가 시작됩니다.
        - 별도로 php-fpm 프로세스 매니저가 시작됩니다.
        - php-fpm은 미리 정의된 수의 워커 프로세스를 생성합니다.
        - 각 php-fpm 워커는 독립적인 PHP 실행 환경을 가집니다.

    2. 요청 수신:
        - 클라이언트가 HTTP 요청을 보냅니다.
        - nginx의 리스닝 소켓이 요청을 받아들입니다.

    3. nginx의 요청 처리:
        - nginx가 요청을 분석하여 PHP 파일에 대한 요청인지 확인합니다.
        - PHP 파일 요청이면, nginx는 이를 php-fpm에 전달하기 위해 준비합니다.

    4. FastCGI 프로토콜:
        - nginx는 요청을 FastCGI 프로토콜 형식으로 변환합니다.
        - 이 과정에서 HTTP 헤더, 서버 변수, GET/POST 데이터 등이 포함됩니다.

    5. php-fpm과의 통신:
        - nginx는 Unix 소켓 또는 TCP 소켓을 통해 php-fpm에 연결합니다.
        - FastCGI 형식의 요청을 php-fpm에 전송합니다.

    6. PHP 실행:
        - php-fpm 프로세스 매니저가 요청을 받아 유휴 워커에 할당합니다.
        - 선택된 php-fpm 워커가 PHP 스크립트를 실행합니다.
        - PHP 인터프리터가 코드를 파싱하고 실행합니다.

    7. 데이터베이스 연결 (필요한 경우):
        - PHP 코드가 데이터베이스 연결을 요청하면, php-fpm 워커가 연결을 생성합니다.
        - 연결 풀링은 php-fpm 설정에 따라 관리됩니다.

    8. 출력 생성:
        - PHP 코드의 실행 결과가 생성됩니다.
        - 결과는 FastCGI 프로토콜을 통해 nginx로 전송됩니다.

    9. nginx의 응답 처리:
        - nginx가 php-fpm으로부터 받은 결과를 처리합니다.
        - 필요한 경우 추가 헤더를 붙이거나 압축을 수행합니다.

    10. 응답 전송:
        - nginx가 최종 처리된 응답을 클라이언트에게 전송합니다.

<!-- /curriculum-chunk -->

### Golang의 Concurrent Mark and Sweep(CMS)

#### 원문: Golang의 Concurrent Mark and Sweep(CMS)

<!-- curriculum-chunk: sha256=46cffff50deb329343dffa9aa58948ebf7a14dc32d62840f7b10c7cb95456a0e major=language-runtime mid=언어 런타임 비교 sub=Golang의 Concurrent Mark and Sweep(CMS) sources=interview_questions.md:5320-5419, interviews.md:5320-5419 -->

> Source: `interview_questions.md:5320-5419`
> Classification reason: runtime comparison
> Duplicate source aliases: `interview_questions.md:5320-5419, interviews.md:5320-5419`

##### Golang의 Concurrent Mark and Sweep(CMS)

Go 언어에서 사용하는 Concurrent Mark and Sweep (CMS) 가비지 컬렉션(GC) 알고리즘은 성능을 극대화하면서 STW(Stop The World) 시간을 최소화하는 데 중점을 둔 트라이컬러 마킹 알고리즘에 기반을 두고 있습니다.
이는 Go 런타임이 Go의 고루틴(Goroutine)과 같은 동시성 모델을 지원하면서도 원활한 메모리 관리를 가능하게 하는 중요한 메커니즘입니다.
Go의 가비지 컬렉터는 성능, 일관성, 그리고 메모리 사용 간의 균형을 맞추기 위해 설계되었으며, 비동기식으로 작동하는 특성을 가지고 있습니다.
여기에서는 Go의 GC가 어떻게 동작하는지에 대한 구체적이고 전문적인 설명을 다루겠습니다.

1. Go의 GC가 해결하려는 문제

    Go 프로그램은 동시성(concurrency)을 강하게 지원하며, 여러 고루틴이 동시에 실행됩니다.
    따라서 메모리 할당과 해제도 빈번하게 발생하고, 이에 대한 메모리 관리를 효율적으로 해야 합니다.
    Concurrent Mark and Sweep 알고리즘은 이러한 메모리 해제를 효율적으로 수행하면서, 애플리케이션이 가능한 한 중단되지 않도록 설계되었습니다.

2. CMS 알고리즘의 기본 흐름

    Go의 가비지 컬렉터는 크게 3단계로 나눌 수 있습니다:

    1. Mark 단계: 객체가 사용 중인지 확인하여 생존 중인 객체를 식별합니다.
    2. Sweep 단계: 사용되지 않는 객체(즉, 가비지로 간주되는 객체)를 수집하여 메모리를 해제합니다.
    3. STW(Stop The World) 단계: Go 런타임에서 고루틴을 일시 중지하고 GC가 처리할 수 있도록 하는 최소화된 중단 시간이 포함됩니다.

    Go의 CMS는 트라이컬러 마킹 알고리즘을 사용하여 이 단계를 최적화합니다.

3. 트라이컬러 마킹 알고리즘

    Go의 GC는 트라이컬러 추적(Three-color marking)을 사용하여 메모리 수집을 관리합니다.
    이 알고리즘은 객체의 상태를 색상으로 나타내며, 각 객체는 세 가지 색 중 하나로 표시됩니다:
    - 흰색: 아직 방문되지 않은 객체입니다. 가비지 컬렉션이 완료된 후에도 흰색으로 남아 있으면 이 객체는 가비지로 간주됩니다.
    - 회색: 이 객체는 마킹 중이며, 아직 모든 자식 객체가 확인되지 않은 상태입니다.
    - 검은색: 이 객체와 그 자식 객체들이 모두 마킹된 상태이며, 이 객체는 사용 중인 메모리로 간주됩니다.

    트라이컬러 마킹 단계:
    1. 초기 상태에서 모든 객체는 흰색입니다.
    2. 루트 객체(스택 변수, 전역 변수 등)부터 시작해 사용 중인 객체를 찾아 회색으로 표시합니다.
    3. 회색 객체를 확인하며 해당 객체가 참조하는 모든 자식 객체를 회색으로 바꾸고, 자신은 검은색으로 표시합니다.
    4. 이 과정이 반복되면서 점점 더 많은 객체가 검은색으로 바뀌고, 마킹이 끝날 때 모든 사용 중인 객체는 검은색이 됩니다. 아직 흰색으로 남아 있는 객체는 더 이상 사용되지 않으며 가비지로 간주됩니다.

    이 마킹 작업은 동시성을 고려해 설계되었습니다. GC가 실행되는 동안에도 애플리케이션의 코드가 실행될 수 있도록 고루틴과 GC가 병렬로 작업을 진행합니다.

4. Concurrent Mark 단계

    가비지 컬렉션의 마킹 작업은 동시적으로 수행되며, 애플리케이션이 실행되는 동안에도 고루틴과 병렬로 진행됩니다.
    이 과정에서 런타임은 `Write Barrier`라는 메커니즘을 사용하여 애플리케이션이 변경하는 포인터 값을 추적합니다.
    이는 Go 런타임이 객체 참조 관계의 변경 사항을 정확히 기록하고, 동시성이 유지되는 동안에도 정확한 마킹을 가능하게 합니다.

    > `Write Barrier`:
    > - 프로그램이 메모리 위치에 새로운 값을 쓰려고 할 때마다 GC는 `Write Barrier`를 사용하여 해당 참조 관계가 업데이트되는 순간을 인식합니다.
    > - 이를 통해 프로그램이 실행 중일 때도 변경된 참조 관계를 올바르게 추적할 수 있어, 정확한 마킹을 보장합니다.

    예를 들어, 만약 고루틴이 회색으로 마킹된 객체에 새로운 자식 객체를 추가하면, 그 자식 객체는 즉시 회색으로 표시되어 마킹 과정에서 누락되지 않게 됩니다.

5. Sweep 단계

    마킹이 완료되면, 이제 사용되지 않는 메모리를 회수하는 Sweep 단계가 시작됩니다.
    이 단계에서는 흰색으로 남은 객체가 메모리에서 해제됩니다.

    - Go의 스윕 단계는 동시성을 지원합니다. 즉, 가비지 컬렉션이 진행되는 동안에도 새로운 메모리 할당이 가능합니다.
    - 스윕(sweep) 단계에서 흰색 객체가 발견되면 해당 객체는 가비지로 간주되어 메모리에서 해제됩니다.
        이때 흰색 객체는 더 이상 사용되지 않는 객체를 나타내며, 스윕 단계의 주요 목표는 이 흰색 객체를 프리 리스트(free list)에 추가하여 메모리를 회수하는 것입니다.

    스윕 단계에서 Go의 런타임은 메모리를 재사용하는 방식으로 동작하여, 새로 할당할 메모리를 찾을 때 불필요한 시스템 호출을 줄이고 성능을 향상시킵니다.

6. STW(Stop The World)와 최소화 전략

    Go의 CMS 알고리즘은 STW(Stop The World) 시간을 최소화하는 데 중점을 둡니다.
    이 단계에서 애플리케이션의 모든 고루틴이 일시적으로 중단되고 GC 작업이 완료될 때까지 기다립니다.

    - Stop The World는 주로 다음 두 시점에서 발생합니다:
        1. Mark 시작 시점: 루트 객체를 식별하기 위해 잠시 중단이 필요합니다.
        2. Sweep 완료 시점: 최종적으로 마킹과 스윕을 정리하고 마무리하는 짧은 시간 동안 중단됩니다.

    Go 런타임은 STW 시간을 최소화하기 위해 대부분의 가비지 컬렉션 작업을 동시성 있게 처리하여, 긴 중단 없이 애플리케이션이 자연스럽게 실행될 수 있도록 합니다.
    최신 Go 버전에서는 STW 시간이 밀리초 수준으로 매우 짧게 유지됩니다.

7. Heap 성장과 GC 트리거

    Go 런타임은 힙 메모리 사용량을 모니터링하여, 힙 크기가 일정 임계치에 도달하면 GC를 트리거합니다.
    Go의 힙 크기 관리 방식은 동적으로 조절되며, 메모리 사용 패턴에 따라 힙을 확장하거나 축소할 수 있습니다.

    - GC 시작 조건

        기본적으로 Go는 현재 힙 크기의 1.5배에 도달하면 가비지 컬렉션을 시작합니다.
        이 임계치는 GOGC 환경 변수를 통해 조정할 수 있습니다.
        - 예: `GOGC=100`은 힙 크기가 2배로 증가할 때 GC를 트리거하는 설정입니다.

        힙 사용량이 갑작스럽게 증가하거나, 프로그램의 메모리 할당 속도가 빠를 때는 GC가 더 자주 실행되며, 메모리 관리가 안정적으로 유지됩니다.

8. 성능 최적화와 튜닝

    Go는 GOGC라는 환경 변수를 통해 GC의 동작을 조정할 수 있습니다.

    - GOGC=100은 힙 크기가 100% 커지면 GC를 트리거하는 기본 설정입니다.
    - 더 자주 GC가 발생하도록 설정하면 메모리 사용량이 줄어들지만, GC 비용이 높아질 수 있습니다. 반대로 GC 발생 빈도를 줄이면 힙 크기가 커지지만, GC 오버헤드가 감소할 수 있습니다.

    Go의 GC는 실시간 시스템처럼 동작하지는 않지만, 매우 짧은 STW 시간을 유지하면서도 성능과 메모리 사용량의 균형을 맞추도록 설계되었습니다.

Go의 Concurrent Mark and Sweep(CMS) 가비지 컬렉션은 트라이컬러 마킹 알고리즘을 기반으로 설계되어, 동시성을 효율적으로 관리하고, Stop The World 시간을 최소화하면서 GC 성능을 향상시킵니다.
이를 통해 Go 프로그램은 다수의 고루틴이 동시에 실행되는 상황에서도 메모리 관리를 안정적으로 처리할 수 있습니다.
Go의 GC는 `Write Barrier`와 같은 메커니즘을 통해 동시성 문제를 해결하며, GOGC를 사용한 메모리 사용 패턴에 따른 조정 기능을 제공하여 고성능 애플리케이션을 위한 최적의 메모리 관리를 지원합니다.

<!-- /curriculum-chunk -->

### VM, GC, 그리고 런타임

#### 원문: VM, GC, 그리고 런타임

<!-- curriculum-chunk: sha256=50235565c4d721960fb1b2483d267199799ea865bdd8edc38ee71bd4d259142d major=language-runtime mid=언어 런타임 비교 sub=VM, GC, 그리고 런타임 sources=interview_questions.md:5287-5319, interviews.md:5287-5319 -->

> Source: `interview_questions.md:5287-5319`
> Classification reason: runtime comparison
> Duplicate source aliases: `interview_questions.md:5287-5319, interviews.md:5287-5319`

##### VM, GC, 그리고 런타임

VM(Virtual Machine)과 GC(Garbage Collection)는 별개의 개념입니다.
GC는 메모리 관리와 관련된 기술이고, VM은 프로그램을 실행하기 위한 가상 환경을 제공합니다.

예를 들어, JVM(Java Virtual Machine)은 Java 프로그램을 바이트코드로 실행하는 가상 환경입니다.
JVM은 반드시 GC를 포함하지는 않지만, Java와 같은 언어는 JVM 내부에 GC가 구현되어 자동으로 메모리를 관리합니다.

언어가 VM을 사용하지 않고도 GC를 사용할 수 있습니다.
예를 들어, Go 언어는 네이티브 코드로 컴파일되어 하드웨어에서 바로 실행되지만, Go 런타임은 자체적으로 GC를 포함해 메모리를 자동으로 관리합니다.
이 경우, VM은 없지만 GC는 존재합니다.

Go는 VM을 사용하지 않고, 네이티브 머신 코드로 컴파일됩니다.
하지만 Go 런타임은 VM과 유사한 기능을 제공하는데, 이는 프로그램 실행 중에 메모리 관리, 스케줄링, GC를 포함합니다.
이 런타임은 Go 프로그램이 안전하고 효율적으로 실행되도록 돕지만, 전통적인 VM처럼 바이트코드를 실행하는 것은 아닙니다.

VM이 없어도 언어 자체에서 메모리 관리를 위해 GC를 사용할 수 있습니다.
VM이 없는 경우, GC는 언어의 런타임 환경에서 구현되어 작동합니다.
Go는 네이티브로 컴파일되어 하드웨어에서 직접 실행되지만, Go 런타임에 GC가 포함되어 있어 메모리 관리를 자동으로 처리합니다.

VM과 런타임은 겹치는 부분이 있지만, 역할과 기능에서 차이가 있습니다.

- VM(Virtual Machine):

    프로그램이 하드웨어와 직접 상호작용하는 대신, 가상화된 하드웨어 환경에서 프로그램을 실행할 수 있게 해주는 소프트웨어입니다.
    예를 들어, Java 프로그램은 JVM이라는 VM에서 실행되어 특정 플랫폼에 독립적인 실행 환경을 제공합니다.

- 런타임(Runtime):

    프로그램이 실행되는 동안 필요한 기능을 제공하는 실행 환경입니다.
    런타임은 메모리 관리, 오류 처리, 스레드 관리 등을 포함하며, 이는 네이티브 코드 실행이거나 VM 위에서 실행될 수도 있습니다.
    Go와 같은 언어는 VM 없이도 런타임을 통해 메모리 관리와 스레드 관리 등을 처리합니다.

<!-- /curriculum-chunk -->

## 함수형 프로그래밍 모델

### 함수형

#### 원문: 함수형

<!-- curriculum-chunk: sha256=7c5b45abb7d6106d2c28451a32820c77390e1b67a1411b1028aee41bbbd90e87 major=language-runtime mid=함수형 프로그래밍 모델 sub=함수형 sources=interview_questions2.md:1743-2120, interviews2.md:1743-2120 -->

> Source: `interview_questions2.md:1743-2120`
> Classification reason: functional programming
> Duplicate source aliases: `interview_questions2.md:1743-2120, interviews2.md:1743-2120`

##### 함수형

함수형 프로그래밍(Functional Programming, FP)은 수학적 함수의 개념에서 영감을 받은 프로그래밍 패러다임으로, 코드의 명확성, 재사용성, 안전성을 강조합니다.

순수 함수(Pure Function), 불변성(Immutability), 고차 함수(Higher-Order Functions), 일급 객체(First-Class Citizens) 등의 원칙을 중심으로 하는 프로그래밍 패러다임입니다.

1. 순수 함수(Pure Function):
    - 동일한 입력에 대해 항상 동일한 결과를 반환하며, 외부 상태를 변경하지 않음.
    - 부수효과(Side Effects)를 피함으로써 프로그램의 예측 가능성을 증가시킴.

2. 불변성(Immutability):
    - 데이터와 상태는 변경할 수 없으며, 변경하려면 복사본을 생성.
    - 이는 데이터 경쟁(Data Race)과 같은 멀티스레드 이슈를 방지.

3. 고차 함수(Higher-Order Function):
    - 함수를 값처럼 취급하며, 이를 다른 함수의 인자로 전달하거나 반환값으로 사용할 수 있음.

4. 일급 객체(First-Class Citizen):
    - 함수가 변수처럼 할당되고, 다른 함수의 인자나 반환값으로 사용될 수 있음.

5. 참조 투명성(Referential Transparency):
    - 표현식이 항상 동일한 값을 반환하여, 호출을 그 결과값으로 대체해도 프로그램의 의미가 유지됨.

6. 부수효과의 최소화(Side Effects Minimization):
    - 함수의 결과에 영향을 미치지 않는 외부 상태 변경을 최소화.

###### 함수형 프로그래밍의 이론적 개념

- Functor
    - 컨테이너 안의 값을 매핑(map)하는 추상화.
    - 값을 직접 변경하지 않고, 컨테이너의 구조를 유지하면서 값을 변환.

    이론적 정의:
    - `map` 연산을 지원하며, 다음 법칙을 만족:
        - 항등 법칙: `F.map(id) == F`
        - 합성 법칙: `F.map(f).map(g) == F.map(f ∘ g)`

- Monad
    - 연속적인 계산을 모델링하는 추상화.
    - Functor를 확장하여, 컨테이너 안의 값을 연결(bind)하고 중첩을 평평하게(flatten) 만듦.

    이론적 정의:
    - `flatMap`(또는 `bind`) 연산을 지원하며, 다음 법칙을 만족:
        - 좌항등 법칙: `M.pure(a).flatMap(f) == f(a)`
        - 우항등 법칙: `M.flatMap(M.pure) == M`
        - 결합 법칙: `M.flatMap(f).flatMap(g) == M.flatMap { x -> f(x).flatMap(g) }`

- 부수효과의 제어
    - Effect Handling: 부수효과를 함수형으로 모델링하여 안전하게 관리.
        - 예: I/O 작업, 네트워크 요청.

###### 3.6 Either를 통한 에러 처리

###### 이론 적용

`Either`는 성공과 실패를 명시적으로 처리하는 함수형 대안입니다.

###### 실무 예제

```kotlin
import arrow.core.Either
import arrow.core.left
import arrow.core.right

fun divide(a: Int, b: Int): Either<String, Int> =
    if (b == 0) "Division by zero".left()
    else (a / b).right()

val result = divide(10, 2) // Right(5)
val error = divide(10, 0) // Left("Division by zero")
```

###### 3.7 I/O 작업 제어

###### 이론 적용

부수효과를 함수형으로 모델링하여 안전하게 처리.

###### Arrow-kt에서의 I/O 모델링

```kotlin
import arrow.fx.IO

val program = IO {
    println("Performing side-effect")
    "Result"
}

val result = program.unsafeRunSync() // "Performing side-effect" 출력
```

###### 순수 함수 (Pure Function)

동일한 입력에 대해 항상 동일한 출력을 반환하며, 외부 상태를 변경하지 않는 함수.

```kotlin
// 순수 함수는 외부 상태에 의존하지 않으며, 동일한 입력에 대해 항상 동일한 결과를 반환합니다.
fun calculateDiscount(price: Double, discountRate: Double): Double =
    price * discountRate

// 외부 상태에 의존하지 않음
val result = calculateDiscount(100.0, 0.1) // 항상 10.0 반환

fun add(a: Int, b: Int): Int = a + b
```

###### 불변성 (Immutability)

함수형 프로그래밍에서는 상태를 변경하지 않고 새로운 값을 생성합니다.

```kotlin
// 데이터를 변경하지 않고, 복사본을 생성하여 상태 변경을 구현.
val originalList = listOf(1, 2, 3)
val updatedList = originalList + 4 // [1, 2, 3, 4]

// 원본은 변경되지 않음
println(originalList) // [1, 2, 3]

val list = listOf(1, 2, 3)
val newList = list.map { it * 2 } // 원래 list는 변경되지 않음
```

###### 고차 함수 (Higher-Order Functions)

다른 함수를 인자로 받거나 반환하는 함수.

```kotlin
// 고차 함수는 함수를 인자로 받거나 반환값으로 사용하여 동작을 추상화합니다.
fun applyOperation(numbers: List<Int>, operation: (Int) -> Int): List<Int> {
    return numbers.map(operation)
}

// 두 배로 만드는 함수 전달
val doubled = applyOperation(listOf(1, 2, 3)) { it * 2 } // [2, 4, 6]

fun applyTwice(f: (Int) -> Int, x: Int): Int = f(f(x))

val result = applyTwice({ it * 2 }, 5) // 20
```

###### 일급 객체 (First-Class Citizens)

함수가 변수에 할당되거나, 다른 함수의 인자로 전달될 수 있음.

```kotlin
val double: (Int) -> Int = { it * 2 }
println(double(4)) // 8
```

###### Functor

컨테이너 안의 값을 매핑할 수 있는 추상화.

Kotlin의 `map` 함수는 Functor의 예입니다.

```kotlin
val list = listOf(1, 2, 3)
val doubled = list.map { it * 2 } // [2, 4, 6]
```

Arrow는 `Option`, `Either`와 같은 컨테이너 타입에 Functor 연산을 제공합니다.

```kotlin
import arrow.core.Option
import arrow.core.some

val option: Option<Int> = 3.some()
val result = option.map { it * 2 } // Option(6)
```

###### Monad

컨테이너 안의 값을 매핑할 뿐 아니라 컨테이너의 중첩을 평평하게(flatten) 하는 추상화.
`bind` 또는 `flatMap`을 통해 중첩된 컨테이너를 단일 컨테이너로 변환합니다.

```kotlin
// 컨테이너 안의 중첩된 값을 평평하게(flatten) 하여 연결.
val nested = listOf(listOf(1, 2), listOf(3, 4))
val flattened = nested.flatMap { it } // [1, 2, 3, 4]
```

Arrow-kt에서의 Monad:

```kotlin
import arrow.core.Option
import arrow.core.some

val opt1: Option<Int> = 10.some()
val result = opt1.flatMap { x -> Option.just(x * 2) } // Option(20)
```

```kotlin
import arrow.core.Option
import arrow.core.none
import arrow.core.some

val opt1: Option<Int> = 10.some()
val opt2: Option<Int> = none()

val result = opt1.flatMap { x -> opt2.map { y -> x + y } } // None
```

###### Task (Deferred in Kotlin)

비동기 작업을 표현하며, 결과를 지연 평가(Lazy Evaluation)로 처리.

Kotlin의 예:

```kotlin
import kotlinx.coroutines.*

suspend fun fetchUser(): String {
    delay(1000) // Simulating long-running task
    return "User"
}

fun main() = runBlocking {
    val task = async { fetchUser() }
    println(task.await()) // "User"
}
```

Arrow의 `IO` 모나드 또는 `Task`는 비동기 작업을 함수형 스타일로 처리하는 데 사용됩니다.

```kotlin
import arrow.fx.coroutines.*

suspend fun main() {
    val task = IO.invoke { "Result" }
    println(task.unsafeRunSync()) // "Result"
}
```

###### Either

실패 또는 성공을 표현하는 양자 타입.
`Either.Left`는 오류를, `Either.Right`는 성공 값을 나타냅니다.

```kotlin
import arrow.core.Either
import arrow.core.left
import arrow.core.right

fun divide(a: Int, b: Int): Either<String, Int> =
    if (b == 0) "Division by zero".left()
    else (a / b).right()

val result = divide(10, 2)
println(result) // Right(5)
```

`bind()`는 Arrow의 `Monad` 컴퓨테이션을 처리하는 DSL(do notation)에서 사용되는 메서드로, Monad 컨텍스트에서 값을 꺼내고 연속적인 계산을 처리합니다.
Monad 컨텍스트란, 예를 들어, `Either`, `Option`, `IO`와 같은 Monad 타입에서 값을 직접 꺼낼 수 없는 상황에서, `bind()`를 통해 안전하게 값을 꺼내 계산을 이어갑니다.

`Either`는 성공(`Right`)과 실패(`Left`)를 표현하는 타입입니다.

```kotlin
import arrow.core.Either
import arrow.core.continuations.either
import arrow.core.left
import arrow.core.right

fun fetchUserData(userId: String): Either<String, String> =
    if (userId == "123") "User data".right() else "User not found".left()

fun fetchUserPosts(userId: String): Either<String, List<String>> =
    if (userId == "123") listOf("Post1", "Post2").right() else "No posts".left()

suspend fun fetchUserWithPosts(userId: String): Either<String, Pair<String, List<String>>> = either {
    val userData = fetchUserData(userId).bind() // 값을 안전하게 꺼냄
                                                // `bind()`는 Monad(`Either`) 내부에서 값을 꺼내 다음 계산에 전달.
                                                // 만약 `Left`가 발생하면 즉시 실행을 중단하고 에러를 반환.
    val userPosts = fetchUserPosts(userId).bind()
    userData to userPosts
}

fun main() {
    val result = fetchUserWithPosts("123")
    println(result) // Right((User data, [Post1, Post2]))
}
```

Arrow를 사용하여 HTTP 통신 결과를 처리하는 경우도 유사합니다.
예를 들어, HTTP 요청이 `Either` 타입으로 성공과 실패를 반환할 때:

```kotlin
import arrow.core.Either
import arrow.core.continuations.either

suspend fun fetchHttpResponse(url: String): Either<String, String> =
    if (url.startsWith("http")) "Response from $url".right()
    else "Invalid URL".left()

suspend fun fetchData(): Either<String, String> = either {
    val response = fetchHttpResponse("http://example.com").bind() // HTTP 응답을 처리
                                                                  // `bind()`는 성공(`Right`) 값을 꺼내 계산을 이어가며, 실패(`Left`)가 발생하면 즉시 반환.
    "Processed: $response"
}

suspend fun main() {
    val result = fetchData()
    println(result) // Right(Processed: Response from http://example.com)
}
```

`bind()`가 HTTP 통신에서 사용된다면, 이는 Arrow와 같은 함수형 라이브러리를 활용하여 안전한 에러 처리와 Monad 값을 관리하는 방식으로 구현되었을 가능성이 큽니다. 이 방식은 다음과 같은 이점을 제공합니다:

1. 명시적 에러 처리:
    - `Either`와 같은 Monad 타입을 통해, 성공과 실패를 명확히 처리.

2. 안전한 값 추출:
    - `bind()`는 Monad 내부 값을 안전하게 꺼내고, 실패가 발생하면 자동으로 에러를 반환.

3. 연속적인 계산 모델링:
    - 여러 HTTP 요청이 연속적으로 연결될 때, `bind()`를 사용해 간결하고 명확하게 코드 작성 가능.

###### Arrow-kt 주요 데이터 타입 및 함수

Arrow-kt는 함수형 프로그래밍의 개념을 구현하는 여러 데이터 타입과 도구를 제공합니다.

- `Option`
    - 정의: 값이 있을 수도, 없을 수도 있는 경우를 표현.
    - `Option.Some`은 값을 나타내고, `Option.None`은 값이 없음을 나타냅니다.

    ```kotlin
    import arrow.core.Option
    import arrow.core.some
    import arrow.core.none

    val someValue: Option<Int> = 5.some()
    val noValue: Option<Int> = none()

    val result = someValue.getOrElse { 0 } // 5
    val fallback = noValue.getOrElse { 0 } // 0
    ```

- `Validated`

    실패 또는 성공을 표현하며, 실패를 누적할 수 있음.

    ```kotlin
    import arrow.core.Validated
    import arrow.core.invalid
    import arrow.core.valid

    val success = 42.valid()
    val failure = "Error".invalid()

    println(success) // Valid(42)
    println(failure) // Invalid(Error)
    ```

- `IO`

    효과(effect)를 표현하며, 부수 효과(side-effect)를 안전하게 처리.

    ```kotlin
    import arrow.fx.IO

    val io = IO { println("Running IO") }
    io.unsafeRunSync() // 실제로 실행
    ```

- Applicative

    독립적인 계산을 결합하여 병렬 실행 지원.

    ```kotlin
    import arrow.core.Tuple2
    import arrow.core.extensions.list.apply.map

    val listA = listOf(1, 2, 3)
    val listB = listOf("A", "B", "C")

    val combined = listA.map(listB) { a, b -> "$a$b" }
    println(combined) // [1A, 1B, 1C, 2A, 2B, 2C, 3A, 3B, 3C]
    ```

<!-- /curriculum-chunk -->
