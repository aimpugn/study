package lambda

import util.RunExample
import java.lang.reflect.InvocationHandler
import java.lang.reflect.Proxy
import kotlin.random.Random

interface NormalInterface {
    fun doSomething()
}

class NormalInterfaceImpl : NormalInterface {
    override fun doSomething() {
        println("Normal interface implementation")
    }
}

/**
 * Kotlin의 일반 인터페이스는 기본적으로 SAM(Single Abstract Method) 변환을 지원하지 않습니다.
 *
 * 'SAM 변환'이란 단일 추상 메서드(Single Abstract Method)를 가진 함수형 인터페이스를 구현하기 위해
 * 람다 표현식을 사용하는 방법입니다.
 *
 * SAM 변환 조건:
 * - 인터페이스는 단 하나의 추상 메서드만 가집니다.
 * - 인터페이스에 `default`, `static` 메서드가 포함되어 있어도 괜찮습니다. (ex: [InvocationHandler.invokeDefault])
 * - 람다 표현식은 추상 메서드의 시그니처(매개변수의 개수, 타입, 순서, 그리고 반환 타입 등)에 맞아야 합니다.
 *
 * ```
 * // 이 후행 람다 표현식은 Runnable 인터페이스의 익명 구현체로 변환됩니다.
 * Thread {
 *     // `invokedynamic` 명령어를 통해 `Runnable` 구현체가 생성되고,
 *     //  `Runnable.run()` 메서드를 호출합니다.
 *     println("It's anonymous instance of Runnable type")
 * }
 * ```
 *
 * Kotlin은 Java의 함수형 인터페이스처럼 SAM 변환을 통해 람다 표현식을 함수형 인터페이스의 익명 구현체로 사용할 수 있도록
 * [`fun interface` 문법](https://kotlinlang.org/docs/fun-interfaces.html)을 지원합니다.
 *
 * 다음은 Kotlin에서 '일반 인터페이스 및 그 구현체'와 '함수형 인터페이스 및 람다 익명 구현체' 코드 및 그 컴파일 결과입니다.
 *
 * ```
 * import kotlin.random.Random
 *
 * interface NormalInterface {
 *     fun doSomething()
 * }
 *
 * class NormalInterfaceImpl : NormalInterface {
 *     override fun doSomething() {
 *         println("Normal interface implementation")
 *     }
 * }
 *
 * fun interface KotlinFunctionalInterface<T> {
 *     fun doSomething(msg: T): Int
 * }
 *
 * fun main(args: Array<String>) {
 *     // 런타임에 함수형 인터페이스 `KotlinFunctionalInterface`의 익명 구현체가 동적으로 생성됩니다.
 *     val functionalInterface = KotlinFunctionalInterface<String> { arg1 ->
 *         println("It's a KotlinFunctionalInterface type anonymous instance $arg1")
 *
 *         Random.nextInt(100, 1000)
 *     }
 *     println("functionalInterface.doSomething's result: ${functionalInterface.doSomething("test")}")
 *
 *
 *     NormalInterfaceImpl().doSomething()
 * }
 * // Output:
 * //   It's a KotlinFunctionalInterface type anonymous instance test
 * //   functionalInterface.doSomething's result: 482
 * //   Normal interface implementation
 * ```
 *
 * ```
 * ❯ javap -c TmpKt.class
 * Compiled from "tmp.kt"
 * public final class TmpKt {
 *   public static final void main(java.lang.String[]);
 *     Code:
 *        0: aload_0
 *        1: ldc           #9                  // String args
 *        3: invokestatic  #15                 // Method kotlin/jvm/internal/Intrinsics.checkNotNullParameter:(Ljava/lang/Object;Ljava/lang/String;)V
 *           ┌> JVM이 부트스트랩 메서드를 호출하여 호출 사이트를 초기화하도록 지시.
 *        6: invokedynamic #32,  0             // InvokeDynamic #0:doSomething:()LKotlinFunctionalInterface;
 *                                                               │ │           └> 호출 사이트의 메서드 타입 시그니처로, 인자가 없고 반환 타입이 `KotlinFunctionalInterface`
 *                                                               │ └> `KotlinFunctionalInterface`의 추상 메서드로, 호출될 메서드 이름.
 *                                                               └>  부트스트랩 메서드 테이블에서 첫 번째 항목(인덱스 0)을 참조.
 *       11: astore_1
 *       12: new           #34                 // class java/lang/StringBuilder
 *       15: dup
 *       16: invokespecial #38                 // Method java/lang/StringBuilder."<init>":()V
 *       19: ldc           #40                 // String functionalInterface.doSomething\'s result:
 *       21: invokevirtual #44                 // Method java/lang/StringBuilder.append:(Ljava/lang/String;)Ljava/lang/StringBuilder;
 *       24: aload_1
 *       25: invokeinterface #48,  1           // InterfaceMethod KotlinFunctionalInterface.doSomething:()I
 *       30: invokevirtual #51                 // Method java/lang/StringBuilder.append:(I)Ljava/lang/StringBuilder;
 *       33: invokevirtual #55                 // Method java/lang/StringBuilder.toString:()Ljava/lang/String;
 *       36: getstatic     #61                 // Field java/lang/System.out:Ljava/io/PrintStream;
 *       39: swap
 *       40: invokevirtual #67                 // Method java/io/PrintStream.println:(Ljava/lang/Object;)V
 *       43: new           #69                 // class NormalInterfaceImpl
 *       46: dup
 *       47: invokespecial #70                 // Method NormalInterfaceImpl."<init>":()V
 *       50: invokevirtual #72                 // Method NormalInterfaceImpl.doSomething:()V
 *       53: return
 * }
 * ```
 *
 * 람다 표현식을 처리하려면 결국 컴파일 타임 또는 런타임에서 처리되어야 합니다.
 * 컴파일 타임에 처리하려면 익명 클래스가 필요하고, 런타임에 처리하려면 보통 리플렉션을 사용합니다.
 * 하지만, 각각 다음과 같은 단점들이 있습니다:
 *
 * **익명 클래스의 단점**:
 * - 람다 표현식이 많을수록 바이트코드 크기와 메모리 사용량이 증가합니다.
 *    - 추가적인 메타데이터와 메모리 공간을 요구로 인한 오버헤드가 있는데 항상 메모리에 로드되어 메모리 사용량이 증가합니다.
 *    - 모든 람다 표현식을 컴파일 타임에 구현체로 생성하면 실행되지 않는 람다도 메모리에 로드됩니다.
 * - 호출 경로가 정적으로 정의되므로 최적화가 제한적입니다.
 *    - 컴파일 시점에 명시적으로 생성된 고정된 바이트코드로 존재하며, 호출 경로가 명확히 고정되므로, 메서드 인라이닝이 불가능합니다.
 *
 *      ```
 *      Runnable r = new Runnable() {
 *          @Override
 *          public void run() {
 *              System.out.println("This is an anonymous class");
 *          }
 *      };
 *      // 생성되는 익명 클래스 예제:
 *      // 호출은 항상 정적으로 정의된 Main$1.run() 메서드를 통해 이루어집니다.
 *      // JIT 컴파일러는 동적 디스패치를 제거하지 못하므로 최적화가 제한적입니다.
 *      //
 *      // public final class Main$1 implements Runnable {
 *      //     public void run() {
 *      //         System.out.println("This is an anonymous class");
 *      //     }
 *      // }
 *      ```
 *
 *    - 익명 클래스는 인터페이스를 구현하는데, 메서드 호출할 경우 항상 [동적 디스패치](https://en.wikipedia.org/wiki/Dynamic_dispatch)를 수행해야 합니다.
 *
 * **리플렉션의 단점**:
 * - 특정 클래스의 메타데이터([Class], [java.lang.reflect.Method], [java.lang.reflect.Field] 등)를 검색해야 합니다.
 * - 검색된 메서드 객체를 기반으로 [java.lang.reflect.Method.invoke]를 호출 시 JVM 내부적으로 여러 단계의 검증 및 동적 바인딩 과정을 거쳐야 합니다.
 * - 리플렉션을 사용한 메서드 호출은 JIT 컴파일러에 의해 최적화되지 않습니다.
 *
 * Java 7에서 도입된 `invokedynamic`는 런타임에 효율적인 동적 메서드 호출을 지원합니다.
 * `invokevirtual`, `invokestatic`, `invokeinterface` 등과는 달리, 호출 대상이 컴파일 시점에 고정되지 않습니다.
 * 따라서 익명 클래스처럼 고정된 리소스를 차지하지 않고 최적화의 여지가 있습니다.
 * 또한 런타임에 호출 경로를 최적화하여 효율적인 구현체를 생성하므로 리플렉션처럼 리소스를 낭비하지 않을 수 있습니다.
 *
 * `invokedynamic`는 부트스트랩 메서드(bootstrap method)를 통해 호출 대상을 런타임에 동적으로 결정합니다.
 * 1. JVM이 `invokedynamic` opcode를 처음 볼 경우, 호출 프로세스를 초기화하기 위해 부트스트랩 메서드라는 특별한 메서드를 호출합니다.
 *   그리고 부트스트랩 메서드가 완료되면, 호출 대상 메서드를 나타내는 핸들 [java.lang.invoke.MethodHandle]을 포함하는
 *   런타임 객체 [java.lang.invoke.CallSite] 인스턴스를 반환합니다.
 *
 * 2. 부트스트랩 메서드가 반환하는 [java.lang.invoke.CallSite]는 불변형과 가변형이 있으며,
 *   JVM이 실제로 실행해야 하는 로직에 대한 포인터([java.lang.invoke.MethodHandle])를 포함합니다.
 *    - 불변형([java.lang.invoke.ConstantCallSite])인 경우:
 *      이후 모든 `invokedynamic` 호출은 부트스트랩 과정을 건너뛰고 고정된 메서드 핸들을 사용합니다.
 *      JVM은 이를 JIT 컴파일러를 통해 호출 경로를 최적화하거나 인라인(inlining) 처리할 수 있습니다.
 *    - 가변형([java.lang.invoke.MutableCallSite])인 경우:
 *      호출 대상 메서드 핸들이 런타임에 변경될 수 있으며, 변경 시 JVM은 새로운 메서드 핸들에 맞게 호출 경로를 업데이트합니다.
 *      주로 동적 언어(JRuby, Jython 등)의 메서드 호출을 유연하게 처리하기 위해 사용된다고 합니다.
 *
 * 호출할 때마다 메서드 검색 및 실행 환경을 설정해야 하는 리플렉션과 달리, `invokedynamic`은 JVM 레벨에서 최적화되므로 성능이 좋습니다:
 * - [java.lang.invoke.CallSite]를 통해 동적 호출을 캐싱.
 * - JIT 컴파일러의 [java.lang.invoke.MethodHandle] 최적화
 *    - 호출 경로를 단순화하여 중간 경유 없이 메서드의 메모리 주소를 직접 참조(direct reference).
 *    - 메서드 호출을 인라인(inlining)하여 호출 오버헤드를 최소화.
 *
 * `javap -v` 옵션을 사용하여 부트스트랩 메서드 테이블을 볼 수 있습니다.
 *
 * ```
 * ❯ javap -c -p -v TmpKt.class
 * Classfile /Users/rody/VscodeProjects/study/jvm/kotlin/examples/tmp/TmpKt.class
 *   Last modified 2025. 1. 22.; size 2077 bytes
 *   SHA-256 checksum 5a9cbd6fa3fd68a11f7900301df0dd6e075e85c203668f5458c5f90b99cf7d30
 *   Compiled from "tmp.kt"
 *
 *   ... 생략 ...
 *
 *   BootstrapMethods:
 *   0: #30 REF_invokeStatic java/lang/invoke/LambdaMetafactory.metafactory:(
 *      Ljava/lang/invoke/MethodHandles$Lookup;                         // `invokedynamic`호출이 발생한 클래스의 컨텍스트(호출 클래스와 접근 권한 정보).
 *      Ljava/lang/String;                                              // 런타임 호출 사이트에서 사용된 메서드 이름. 이 경우 `doSomething`.
 *      Ljava/lang/invoke/MethodType;                                   // 런타임 호출 사이트의 메서드 시그니처(인수와 반환 타입). 이 경우 `()LKotlinFunctionalInterface`.
 *      Ljava/lang/invoke/MethodType;                                   // 함수형 인터페이스의 추상 메서드 시그니처 (지워진 형태 포함).
 *      Ljava/lang/invoke/MethodHandle;                                 // 실제 람다 구현체를 참조하는 MethodHandle.
 *      Ljava/lang/invoke/MethodType;                                   // 구체적인 함수형 인터페이스 구현체의 메서드 시그니처.
 *     )Ljava/lang/invoke/CallSite;
 *     Method arguments:
 *       #17 (Ljava/lang/Object;)I                                      // Object 타입을 인자로 받고 정수(`I`)를 리턴하는 지워진 메서드 시그니처.
 *       #22 REF_invokeStatic TmpKt.main$lambda$0:(Ljava/lang/String;)I // 실제 람다 로직을 가리키는 `MethodHandle`.
 *       #23 (Ljava/lang/String;)I                                      // String 타입을 인자로 받고 정수를 리턴하는 지워지지 않은 메서드 시그니처.
 * ```
 *
 * - 모든 람다에 대한 부트스트랩 메서드는 [java.lang.invoke.LambdaMetafactory.metafactory]입니다.
 * - [java.lang.invoke.MethodHandle]은 실제 호출할 메서드(`main$lambda$0`)를 가리키며,
 *   [java.lang.invoke.MethodType]은 이 호출의 형태를 정의합니다.
 * - 제네릭 타입(`T`)은 컴파일 시 JVM의 타입 소거(type erasure)에 의해 `Object`로 변환됩니다.
 *
 * References:
 * - [An Introduction to Invoke Dynamic in the JVM](https://www.baeldung.com/java-invoke-dynamic)
 * - [Type Erasure in Java Explained](https://www.baeldung.com/java-type-erasure)
 * - [Understanding Java method invocation with invokedynamic](https://blogs.oracle.com/javamagazine/post/understanding-java-method-invocation-with-invokedynamic)
 */
fun interface KotlinFunctionalInterface<T> {
    fun doSomething(msg: T): Int
}

/**
 * Java 경우 `execute(() -> { ... })`를 사용하면 컴파일러는 `invokedynamic` opcode를 생성하여
 * 함수형 인터페이스 [Runnable]의 익명 구현체를 동적으로 생성합니다.
 *
 * 반면, Kotlin 경우 `() -> {}` 대신 `{}` 문법을 제공합니다.
 * 이 `{}` 람다 표현식은 기본적으로 함수 타입([Function] Type)이고, `(T) -> R`로 간주됩니다.
 * 인자의 개수에 따라 [Function0], [Function1], ... `FunctionN` 타입으로 컴파일 됩니다. `N`은 인자의 개수를 의미합니다.
 *
 * 단, Kotlin 컴파일러는 Java 함수형 인터페이스([Runnable], [java.util.concurrent.Callable], [java.util.Comparator] 등 보통 [FunctionalInterface]으로 표시되어 있음)를
 * 인자로 받는 경우, 자동으로 람다 표현식을 Java 함수형 인터페이스의 익명 구현체로 변환합니다.(SAM 변환)
 *
 * ### 일반적인 람다 표현식: [Function2] 타입 객체로 컴파일되는 경우
 *
 * ```
 * fun main(args: Array<String>) {
 *     // Java 경우: (int x, int y) -> { return x + y; }
 *     val sum: (Int, Int) -> Int = { x, y -> x + y }
 *     println(sum(1, 2))
 * }
 * ```
 *
 * 위 코드에서 `sum`은 다음과 같이 [Function2] 타입의 `Function2<Integer, Integer, Integer>` 객체로 컴파일됩니다.
 *
 * ```
 * ❯ javap -c TmpKt.class
 * Compiled from "tmp.kt"
 * public final class TmpKt {
 *   public static final void main(java.lang.String[]);
 *     Code:
 *        0: aload_0
 *        1: ldc           #9                  // String args
 *        3: invokestatic  #15                 // Method kotlin/jvm/internal/Intrinsics.checkNotNullParameter:(Ljava/lang/Object;Ljava/lang/String;)V
 *        6: invokedynamic #35,  0             // InvokeDynamic #0:invoke:()Lkotlin/jvm/functions/Function2;
 *       11: astore_1
 *       12: aload_1
 *       13: iconst_1
 *       14: invokestatic  #41                 // Method java/lang/Integer.valueOf:(I)Ljava/lang/Integer;
 *       17: iconst_2
 *       18: invokestatic  #41                 // Method java/lang/Integer.valueOf:(I)Ljava/lang/Integer;
 *       21: invokeinterface #45,  3           // InterfaceMethod kotlin/jvm/functions/Function2.invoke:(Ljava/lang/Object;Ljava/lang/Object;)Ljava/lang/Object;
 *       26: checkcast     #47                 // class java/lang/Number
 *       29: invokevirtual #51                 // Method java/lang/Number.intValue:()I
 *       32: istore_2
 *       33: getstatic     #57                 // Field java/lang/System.out:Ljava/io/PrintStream;
 *       36: iload_2
 *       37: invokevirtual #63                 // Method java/io/PrintStream.println:(I)V
 *       40: return
 * }
 * ```
 *
 * ### SAM 변환: Java 함수형 인터페이스 구현체로 컴파일 되는 경우
 *
 * 반면, [Runnable]을 인자로 받는 코드의 경우:
 *
 * ```
 * fun main(args: Array<String>) {
 *     val thread = Thread {
 *         println("Thread 시작됨")
 *         Thread.sleep(1000) // 1초 대기
 *         println("Thread 종료됨")
 *     }
 *
 *     thread.start()
 *     thread.join()
 * }
 * ```
 *
 * 여기서 원래라면 [Function0] 타입의 `Function0<Unit>` 객체로 컴파일되어야 합니다.
 * 하지만 Java 함수형 인터페이스를 파라미터로 요구하고 있으므로 Kotlin 컴파일러는 SAM 변환을 수행합니다.
 * 즉, Kotlin 람다 표현식을 [Runnable] 익명 구현체로 자동 변환되고, 아래와 같이 컴파일 됩니다.
 *
 * ```
 * ❯ javap -c TmpKt
 * Compiled from "tmp.kt"
 * public final class TmpKt {
 *   public static final void main(java.lang.String[]);
 *     Code:
 *        0: aload_0
 *        1: ldc           #9                  // String args
 *        3: invokestatic  #15                 // Method kotlin/jvm/internal/Intrinsics.checkNotNullParameter:(Ljava/lang/Object;Ljava/lang/String;)V
 *        6: new           #17                 // class java/lang/Thread
 *        9: dup
 *       10: invokedynamic #34,  0             // InvokeDynamic #0:run:()Ljava/lang/Runnable;
 *       15: invokespecial #38                 // Method java/lang/Thread."<init>":(Ljava/lang/Runnable;)V
 *       18: astore_1
 *       19: aload_1
 *       20: invokevirtual #41                 // Method java/lang/Thread.start:()V
 *       23: aload_1
 *       24: invokevirtual #44                 // Method java/lang/Thread.join:()V
 *       27: return
 * }
 * ```
 *
 * `invokedynamic` opcode는 런타임에 람다 표현식에 대응되는 [Runnable]의 익명 구현체를 동적으로 생성합니다.
 *
 * 만약 Kotlin 인터페이스를 Java 함수형 인터페이스처럼 사용하려면 `fun interface` 사용하여
 * 함수형 인터페이스임을 명시적으로 선언해야 합니다.
 * 그러면 Kotlin 람다 표현식은 SAM 변환을 통해 해당 인터페이스의 구현체로 변환됩니다.
 *
 * References:
 * - [Functional Interfaces in Java](https://www.baeldung.com/java-8-functional-interfaces)
 * - [Functional (SAM) interfaces](https://kotlinlang.org/docs/fun-interfaces.html)
 */
@RunExample
fun testKotlinFunctionalInterfaceOfKotlin() {
    // val normalInterface = NormalInterface {
    //                       ^ Error: Interface 'interface NormalInterface : Any' does not have constructors.
    //     println("Work done!")
    //
    //     Random.nextInt(100, 1000)
    // }

    /**
     * '인터페이스가 SAM 변환 대상 && 람다 표현식의 시그니처와 추상 메서드의 매개변수의 개수, 타입, 순서, 그리고 반환 타입 등이 정확히 일치'하는 경우,
     * 해당 인터페이스의 변수에 할당될 수 있습니다.
     *
     * [functionalInterface]를 디컴파일한 결과를 보면 아래와 같습니다.
     * 이 경우 [메서드 참조](https://docs.oracle.com/javase/tutorial/java/javaOO/methodreferences.html)를 사용하고 있습니다.
     *
     * ```
     * public final class LambdaExampleKt {
     *    @RunExample
     *    public static final void testKotlinFunctionalInterfaceOfKotlin() {
     *       KotlinFunctionalInterface functionalInterface = LambdaExampleKt::testKotlinFunctionalInterfaceOfKotlin$lambda$0;
     *       System.out.println("functionalInterface.doSomething's result: " + functionalInterface.doSomething("test"));
     *    }
     *
     *    private static final int testKotlinFunctionalInterfaceOfKotlin$lambda$0(String arg1) {
     *       Intrinsics.checkNotNullParameter(arg1, "arg1");
     *       System.out.println("It's a KotlinFunctionalInterface type anonymous instance " + arg1);
     *       return Random.Default.nextInt(100, 1000);
     *    }
     * }
     * ```
     *
     * Java 코드로 실행되는 [LambdaExample.kotlinFunctionalInterfaceInJava]를 참고합니다.
     */
    val functionalInterface = KotlinFunctionalInterface<String> { arg1 ->
        println("It's a KotlinFunctionalInterface type anonymous instance $arg1")

        Random.nextInt(100, 1000)
    }
    println("functionalInterface.doSomething's result: ${functionalInterface.doSomething("test")}")
}


@RunExample
fun testFunctionTypes() {
    /**
     * [InvocationHandler.invoke] 단일 추상 메서드만 존재하므로,
     * Kotlin 컴파일러는 이 람다 표현식을 `invokedynamic` opcode로 변환하여
     * 런타임에 [InvocationHandler] 인터페이스의 익명 구현체를 만들 수 있습니다.
     *
     * ```
     * ... 생략 ...
     * 22: invokedynamic #40,  0             // InvokeDynamic #0:invoke:()Ljava/lang/reflect/InvocationHandler;
     * 27: invokestatic  #46                 // Method java/lang/reflect/Proxy.newProxyInstance:(Ljava/lang/ClassLoader;[Ljava/lang/Class;Ljava/lang/reflect/InvocationHandler;)Ljava/lang/Object;
     * 30: dup
     * 31: ldc           #48                 // String null cannot be cast to non-null type <root>.PrintAround
     * 33: invokestatic  #51                 // Method kotlin/jvm/internal/Intrinsics.checkNotNull:(Ljava/lang/Object;Ljava/lang/String;)V
     * 36: checkcast     #17                 // class PrintAround
     * ... 생략 ...
     * ```
     *
     * 구현 클래스로 만든다면 아래와 같습니다.
     * ```
     * class PrintAroundInvocationHandler(private val target: Any) : InvocationHandler {
     *     override fun invoke(proxy: Any?, method: Method?, args: Array<out Any?>?): Any? {
     *         println("=== ${method?.name} start ===")
     *         val result = method?.invoke(target, *(args ?: emptyArray()))
     *         println("=== ${method?.name} end ===")
     *
     *         return result
     *     }
     * }
     * ```
     */
    val anonymousInvocationHandler = InvocationHandler { _, method, args ->
        println("=== ${method?.name} start ===")
        // 리플렉션을 통해 메서드를 호출하므로 일반적인 메서드 호출에 비해 오버헤드가 발생합니다.
        val result = method?.invoke(PrintAroundImpl(), *(args ?: emptyArray()))
        println("=== ${method?.name} end ===")

        return@InvocationHandler result
    }

    val proxy = Proxy.newProxyInstance(
        PrintAround::class.java.classLoader,
        arrayOf(PrintAround::class.java),
        anonymousInvocationHandler
    ) as PrintAround

    proxy.higherOrderFunction1 { arg1, arg2 ->
        "Added result: ${arg1 + arg2}"
    }

    proxy.higherOrderFunction2 { arg1, arg2 ->
        "Added result: ${arg1 + arg2}"
    }

    val addNumber: (Int, Int) -> String = { arg1, arg2 -> "Added result by function reference: ${arg1 + arg2}" }

    proxy.higherOrderFunction1(addNumber)
    proxy.higherOrderFunction2(addNumber)

    proxy.higherOrderFunction3()
    proxy.higherOrderFunction3 { arg1, arg2 ->
        "Added result by override function: ${arg1 + arg2}"
    }

    proxy.higherOrderFunction4 { arg1, arg2 ->
        "Added result when AddOperation: ${arg1 + arg2}"
    }

    proxy.higherOrderFunction5 { arg1 ->
        "Added result by extension function: ${this + arg1}"
    }
}

interface PrintAround {
    fun higherOrderFunction1(fn: Function2<Int, Int, String>)
    fun higherOrderFunction2(fn: (Int, Int) -> String)
    fun higherOrderFunction3(fn: (Int, Int) -> String = { arg1, arg2 -> "Added result by default function: ${arg1 + arg2}" })
    fun higherOrderFunction4(fn: AddOperation)
    fun higherOrderFunction5(fn: Int.(Int) -> String)
}

typealias AddOperation = (Int, Int) -> String

class PrintAroundImpl : PrintAround {
    override fun higherOrderFunction1(fn: Function2<Int, Int, String>) {
        println(fn(Random.nextInt(1, 10), Random.nextInt(1, 10)))
    }

    override fun higherOrderFunction2(fn: (Int, Int) -> String) {
        println(fn(Random.nextInt(1, 10), Random.nextInt(1, 10)))
    }

    override fun higherOrderFunction3(fn: (Int, Int) -> String) {
        println(fn(Random.nextInt(1, 10), Random.nextInt(1, 10)))
    }

    override fun higherOrderFunction4(fn: AddOperation) {
        println(fn(Random.nextInt(1, 10), Random.nextInt(1, 10)))
    }

    /**
     * [higherOrderFunction5]의 두 함수 사용 방식은 모두 정상적으로 동작합니다.
     * Java 코드로 컴파일된 결과를 보면 같은 코드로 컴파일되는 것을 확인할 수 있습니다.
     * Kotlin에서 확장 함수 타입이 실제로는 일반 함수 타입으로 처리되기 때문입니다.
     * ```
     * public void method5(@NotNull Function2 fn) {
     *    Intrinsics.checkNotNullParameter(fn, "fn");
     *    System.out.println(fn.invoke(Random.Default.nextInt(1, 10), Random.Default.nextInt(1, 10)));
     *    System.out.println(fn.invoke(Random.Default.nextInt(1, 10), Random.Default.nextInt(1, 10)));
     * }
     * ```
     * - 확장 함수 타입(`T.(U) -> R`)은 `(T, U) -> R`과 동일하게 동작합니다.
     * - `Int.(Int) -> String`는 `Function2<Integer, Integer, String>`로 컴파일.
     *
     * 즉, 확장 함수는 실제로 리시버를 첫 번째 인자로 전달하는 방식으로 동작하며,
     * 이를 통해 JVM에서 별도의 처리 없이 함수 호출이 가능합니다.
     */
    override fun higherOrderFunction5(fn: Int.(Int) -> String) {
        // 이 경우 Kotlin은 첫 번째 인자를 암묵적으로 리시버로 전달합니다.
        println(fn(Random.nextInt(1, 10), Random.nextInt(1, 10)))
        // 리시버를 명시적으로 전달합니다.
        println(Random.nextInt(1, 10).fn(Random.nextInt(1, 10)))
    }
}