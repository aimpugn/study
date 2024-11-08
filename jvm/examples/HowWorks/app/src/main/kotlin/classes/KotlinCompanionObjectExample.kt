package classes

/**
 * `companion object`는 정적 필드처럼 보이지만 실제로 객체 인스턴스로 간주되며 더 유연한 기능을 제공합니다.
 * 일반적으로 Java의 `static`과 비교되지만, 실제로는 객체이므로 다형성 및 인터페이스 구현 등 더 유연하게 사용할 수 있습니다.
 *
 * Java의 `static` 키워드는 클래스 레벨의 정적 멤버를 정의하기 위한 메커니즘입니다.
 * ```
 * public class Example {
 *     public static int staticValue = 42;
 *
 *     public static void printStatic() {
 *         System.out.println("Static Value: " + staticValue);
 *     }
 * }
 * ```
 * Java에서 `static`은 다음과 같은 특징들을 갖습니다:
 * - Java에서 `static` 멤버는 클래스와 강하게 연결됩니다. 특정 클래스와 직접 연관된 정적 메서드와 정적 변수를 정의하기 위한 용도입니다.
 *   정적 멤버는 클래스 로드 시점에 메모리에 로드되며, 인스턴스와 무관하게 사용됩니다.
 * - Java에서 `static` 메서드는 다형성을 지원하지 않으며, 상속받을 수 없습니다.
 *   따라서 인터페이스 기반 설계와 같이 목(Mock)을 적용할 수 있는 유연한 구조를 만들 수 없습니다.
 * - Java에서 `static` 메서드는 정적 컨텍스트에서만 동작하기 때문에 유연성이 부족합니다.
 *
 * 반면 Kotlin의 `companion object`는 **클래스 내부에 정의되는 정적 객체**입니다.
 * - 단순히 정적 데이터와 메서드를 관리하는 도구가 아니라, 클래스와 관련된 공유 객체(인스턴스)를 정의합니다.
 *    ```
 *    class Example {
 *        companion object {
 *            val value = 42
 *            fun greet() = "Hello"
 *        }
 *    }
 *
 *    fun main() {
 *        // Java에서는 `static`은 클래스에 바인딩 되며 객체로 간주되지 않습니다.
 *        // 하지만 Kotlin에서는 `companion object`를 명시적으로 객체처럼 다룰 수 있습니다.
 *        println(Example.Companion.greet()) // 객체처럼 접근 가능
 *    }
 *    ```
 *
 * - 확장 함수를 정의할 수 있습니다.
 *    ```
 *    class Example {
 *        companion object {
 *            fun originalFunction() = "Original"
 *        }
 *    }
 *
 *    fun Example.Companion.newFunction() = "Extended"
 *
 *    fun main() {
 *        println(Example.originalFunction()) // Original
 *        println(Example.newFunction())      // Extended
 *    }
 *    ```
 *
 * - 다형성 및 인터페이스 구현이 가능합니다.
 *    ```
 *    interface Factory<T> {
 *        fun create(): T
 *    }
 *
 *    // Example 클래스의 생성자는 `private`으로 숨기고, `companion object` 통해 생성 로직을 제어할 수 있습니다.(캡슐화)
 *    class Example private constructor() {
 *
 *        // Java의 `static` 메서드는 클래스에 바인딩되므로 다형성을 지원하지 않으며, 오버라이딩할 수 없습니다.
 *        // 반면 `companion object`는 객체로서 인터페이스를 구현할 수 있습니다.
 *        companion object : Factory<Example> {
 *            private val log: Logger by lazy { Logger() }
 *
 *            override fun create(): Example = Example()
 *        }
 *    }
 *
 *    fun main() {
 *        val example = Example.create()
 *        println(example)
 *    }
 *    ```
 *
 * - Java의 `static` 메서드는 테스트가 어렵고, 목(Mock)을 사용할 수 없습니다.
 *   반면, `companion object`는 객체로 간주되기 때문에 목(Mock)을 통해 테스트할 수 있습니다.
 *    ```
 *    class Example {
 *        companion object {
 *            fun fetchData(): String = "Real Data"
 *        }
 *    }
 *
 *    fun main() {
 *        // Mocking the Companion object (using a mocking library)
 *        println("Mocked Data") // Example of overriding for testing
 *    }
 *    ```
 *
 * 이런 `companion object`는 Java와의 상호 운용성을 위해 바이트코드 수준에서 별도의 정적 내부 클래스로 컴파일됩니다.
 * 아래와 같이, 컴파일 된 클래스 파일을 `javap`로 디스어셈블 해보면 `KotlinCompanionObjectExample$Companion`라는 클래스를 볼 수 있습니다.
 * 그리고 `KotlinCompanionObjectExample` 클래스의 `static {}` 정적 초기화 블록이 `KotlinCompanionObjectExample$Companion` 객체를 초기화합니다.
 * Java 코드로 구현한다면 [KotlinCompanionObjectExampleByJava]처럼 구현할 수 있습니다.
 *
 * ```
 * ❯ rg --files | rg class | xargs javap
 * Compiled from "KotlinCompanionObjectExample.kt"
 * public final class classes.KotlinCompanionObjectExample$Companion {
 *   public final void finalVoidMethod();
 *   public classes.KotlinCompanionObjectExample$Companion(kotlin.jvm.internal.DefaultConstructorMarker);
 * }
 * Compiled from "KotlinCompanionObjectExample.kt"
 * public final class classes.KotlinCompanionObjectExample {
 *   public static final classes.KotlinCompanionObjectExample$Companion Companion;
 *   public classes.KotlinCompanionObjectExample();
 *   static {};
 * }
 * Compiled from "KotlinCompanionObjectExample.kt"
 * public final class classes.KotlinCompanionObjectExampleKt {
 *   public static final void main();
 *   public static void main(java.lang.String[]);
 * }
 * ```
 *
 * 참고로 [kotlin.jvm.internal.DefaultConstructorMarker]는 컴파일러가 자동으로 생성하는 `synthetic constructor`의 인자로 사용됩니다.
 * 이를 통해 `synthetic constructor`가 일반 생성자와 충돌하지 않도록 구분자 역할을 합니다.
 * 자세한 내용은 [SyntheticConstructorExample]를 참고합니다.
 */
class KotlinCompanionObjectExample {
    companion object {
        fun finalVoidMethod() {
            println("finalVoidMethod is called")
        }
    }
}

fun main() {
    KotlinCompanionObjectExample.Companion.finalVoidMethod()
}
