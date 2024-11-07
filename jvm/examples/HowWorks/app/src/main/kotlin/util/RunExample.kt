package util

import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.launch
import org.springframework.context.annotation.ClassPathScanningCandidateComponentProvider
import org.springframework.core.io.support.PathMatchingResourcePatternResolver
import safety.unsafeCalculateByCoroutineScope
import java.io.File
import java.net.URL
import java.security.SecureClassLoader
import java.util.jar.JarFile
import kotlin.reflect.full.callSuspend
import kotlin.reflect.jvm.kotlinFunction

/**
 * 이 어노테이션이 붙은 메서드를 실행합니다.
 * 해당 메서드는 파라미터가 없어야 합니다.
 *
 * ```
 * @RunExample
 * fun unsafeCalculateByThread() {}
 * ```
 *
 * - 함수에만 사용할 것이므로 [Target]은 [AnnotationTarget.FUNCTION]으로 설정합니다.
 * - 유지되는 기간인 [Retention] 경우, 리플렉션에서 활용할 것이므로 [AnnotationRetention.RUNTIME]으로 설정합니다.
 *
 * 참고로 Kotlin 경우 [org.springframework.web.bind.annotation.RequestMapping]에서처럼
 * 어떤 메서드를 정의할 수 없습니다.
 *
 * ```
 * public @interface RequestMapping {
 *      String name() default ""; -> Kotlin에서는 불가능. "Members are prohibited in annotation classes"
 * ```
 *
 * @param printTitle 실행되는 메서드의 이름을 출력할 것인지 여부. `true`면 메서드 실행 직전 [RunExampleInvoker.printTitle] 메서드가 호출됩니다.
 */
@Target(AnnotationTarget.FUNCTION) // 함수에만 적용할 수 있도록 설정
@Retention(AnnotationRetention.RUNTIME)
annotation class RunExample(
    val printTitle: Boolean = true,
    val title: String = ""
)

class RunExampleInvoker {
    companion object {
        private const val CLASS_SUFFIX = ".class"

        /**
         * Java는 기본적으로 패키지 내의 모든 클래스를 나열하거나 탐색하는 기능을 제공하지 않습니다.
         * 따라서, 클래스패스에 있는 모든 클래스 파일을 찾기 위해 파일 시스템과 클래스 로더를 활용하여 특정 패키지를 검색해야 합니다.
         *
         * Examples:
         * - [ClassPathScanningCandidateComponentProvider.scanCandidateComponents]
         * - [PathMatchingResourcePatternResolver.getResources]
         */
        suspend fun invoke(packageName: String) {
            var currClassName = ""
            scanTargetClasses(packageName).forEach { clazz ->
                /**
                 * Kotlin에서 인스턴스화 가능한 타입들은 다음과 같습니다.
                 * - 일반 클래스 (`class`)
                 * - 데이터 클래스 (`data class`)
                 * - 내부 클래스 (`inner class`)
                 * - `sealed class`의 구체 구현체
                 *
                 * 그리고 단일 인스턴스만 생성되는 타입들은 다음과 같습니다.
                 * - `object` (독립적인 싱글톤 객체)
                 * - `companion object` (특정 클래스와 연관된 싱글톤 객체로, 클래스의 static 멤버**처럼** 호출 가능)
                 *
                 * 하지만, java와 다르게 Kotlin의 경우 별도의 클래스 없이 파일 수준(file-level)에서 함수, 프로퍼티, 타입 별칭 등을 자유롭게 사용할 수 있습니다:
                 * ```
                 * // Example.kt
                 * fun topLevelFunction() { } // static 메서드와 같습니다.
                 * const val CONSTANT = 42
                 * var mutableProperty = "Hello"
                 * ```
                 *
                 * 최상위 수준에서 정의된 함수와 속성은 Kotlin에서 File Facade 클래스라는 방식으로
                 * 패키지 내의 모든 클래스 파일을 감싸며, 기본적으로 `static` 멤버로 변환됩니다.
                 *
                 * 이는 실제로 클래스가 없는 것이 아니라, Kotlin JVM 컴파일러가 "file facades"라고 불리는 wrapper 클래스를 생성해주기 때문입니다.
                 * 이때 `File Facade`는 단순히 최상위 함수와 프로퍼티를 감싸기 위해 Kotlin 컴파일러가 생성하는 클래스일 뿐입니다.
                 *
                 * `File Facade`는 다음 규칙을 따라 생성됩니다:
                 * - 파일 이름을 기반으로 클래스가 생성됩니다. ex: Example.kt -> ExampleKt.class
                 * - `final class`로 생성되어 상속이 방지됩니다.
                 * - @file:[JvmName] 어노테이션을 사용하여 생성될 클래스 이름을 지정할 수도 있습니다.
                 * - 기본 생성자가 존재하지 않으므로 인스턴스화가 불가능합니다. 그래서 이 경우 ([Class.getDeclaredConstructor]) 호출하면
                 *   [Class.getConstructor0]에서 다음과 같은 익셉션이 발생합니다:
                 *    ```
                 *    java.lang.NoSuchMethodException: safety.SafetyExampleKt.<init>()
                 *    ```
                 * - 모든 멤버가 `static final`로 컴파일되어 불변성을 보장합니다.
                 *
                 * [`javap`](https://docs.oracle.com/en/java/javase/21/docs/specs/man/javap.html) 디스어셈블러를 사용해서 생성자 존재하지 않는 경우와 존재하는 경우의
                 * `.class` 파일의 구조를 디스어셈블리 해보면 다음과 같은 결과가 나옵니다:
                 * ```
                 * # 생성자 존재하지 않는 경우 아예 SafetyExampleKt 클래스에 대한 생성자가 없음을 알 수 있습니다.
                 * ❯ javap -v SafetyExampleKt.class | rg "public|private|protected"
                 * public final class safety.SafetyExampleKt
                 *   public static final void unsafeCalculateByThread();
                 *   public static final java.lang.Object calculateByCoroutineScope(kotlin.coroutines.Continuation<? super kotlin.Unit>);
                 *   public static final int randomInt();
                 *   public static final int getFizz();
                 *   public static final int getBuzz();
                 *   public static final int getFoo();
                 *   public static final safety.Example getBar();
                 *   public static final void valIsReadOnlyNotImmutable();
                 *   public static final void safeCalculateBySynchronizedThread();
                 *   public static final java.lang.String getPart1();
                 *   public static final void setPart1(java.lang.String);
                 *   public static final java.lang.String getPart2();
                 *   public static final void setPart2(java.lang.String);
                 *   public static final java.lang.String getPartsConcat();
                 *   public static final void valueOfReadOnlyValCanBeMutable();
                 *   public static final void useImmutableAsImmutable();
                 *   public static final void immutableDoNotNeedDefensiveCopy();
                 *   public static final void dataClassForCopy();
                 *   public static final java.lang.Object main(kotlin.coroutines.Continuation<? super kotlin.Unit>);
                 *   public static void main(java.lang.String[]);
                 *   public static final #548= #13 of #547;  // IntRef=class kotlin/jvm/internal/Ref$IntRef of class kotlin/jvm/internal/Ref
                 *   public static final #182= #187 of #181; // Default=class kotlin/random/Random$Default of class kotlin/random/Random
                 *   public static final #469= #476 of #468; // Companion=class util/PrintTitleScanner$Companion of class util/PrintTitleScanner
                 *
                 * # 또는 k 값이 2이면 FileFacade 입니다.
                 * ❯ javap -v SafetyExampleKt.class | rg '^\s+k='
                 *       k=2
                 *
                 * # 반면 생성자가 존재하는 경우에는 다음과 같이 생성자가 존재함을 알 수 있습니다.
                 * ❯ javap -v Person.class | rg "public|private|protected"
                 * public final class safety.Person
                 *   public safety.Person(java.lang.String, java.util.List<java.lang.String>, java.util.List<java.lang.String>);
                 *   public final java.lang.String getName();
                 *   public final java.util.List<java.lang.String> getMutableHobbies();
                 *   public final java.util.List<java.lang.String> getImmutableHobbies();
                 *   public final java.util.List<java.lang.String> getMutableHobbiesAsImmutable();
                 *
                 * # 또는 k 값이 1이면 일반 클래스입니다.
                 * ❯ javap -v Person.class | rg '^\s+k='
                 *       k=1
                 * ```
                 *
                 * References:
                 * - [Package-level functions](https://kotlinlang.org/docs/java-to-kotlin-interop.html#package-level-functions)
                 * - [Why is possible to write a function outside a class in Kotlin?](https://stackoverflow.com/a/49015268)
                 * - [Multiplatform projects - Technical details](https://kotlinlang.org/docs/coding-conventions.html#multiplatform-projects)
                 * - [FileFacade](https://kotlinlang.org/api/kotlinx-metadata-jvm/kotlin-metadata-jvm/kotlin.metadata.jvm/-kotlin-class-metadata/-file-facade)
                 */
                var instance: Any? = null
                clazz.declaredMethods.forEach methods@{ method ->
                    /**
                     * `RunExample::class`은 Kotlin의 [kotlin.reflect.KClass] 객체를 참조하는 표현입니다.
                     * `RunExample::class.java`는 [kotlin.reflect.KClass]를 [Class] 타입으로 리턴합니다.
                     */
                    if (method.isAnnotationPresent(RunExample::class.java)) {
                        if (instance == null || currClassName != clazz.name) {
                            currClassName = clazz.name

                            instance = if (
                                clazz.kotlin.isCompanion
                                || clazz.kotlin.objectInstance != null // https://stackoverflow.com/a/55391118
                                || clazz.simpleName.endsWith("Kt")
                            ) {
                                /**
                                 * `companion object`, `object`, 그리고 `File Facade`는 인스턴스 없이 정적 메서드처럼 호출이 가능합니다.
                                 * 단, `companion object` 경우 실제로 static 메서드인 것은 아닙니다.
                                 *
                                 * ```
                                 * class OuterClass {
                                 *     companion object {
                                 *         fun staticOrNot() {
                                 *             println(::staticOrNot.name)
                                 *         }
                                 *     }
                                 * }
                                 * ```
                                 *
                                 * 위의 `staticOrNot` 메서드는 `public final void staticOrNot()`로 컴파일됩니다.
                                 *
                                 * ```
                                 * Compiled from "OuterClass.kt"
                                 * public final class tmp.OuterClass$Companion {
                                 *   public final void staticOrNot();
                                 *   public tmp.OuterClass$Companion(kotlin.jvm.internal.DefaultConstructorMarker);
                                 * }
                                 * ```
                                 *
                                 * [JvmStatic] 어노테이션으로 표시된 경우에만 정적 메서드처럼 직접 호출이 가능합니다.
                                 * 그렇지 않은 경우, 가령 Java에서 해당 코드를 사용하려고 한다면, Companion 객체를 통해 접근해야 합니다.
                                 *
                                 * `static` 메서드 특징은 다음과 같습니다:
                                 * - 클래스에 바인딩되어 있지 인스턴스에 바인딩되어 있지 않습니다.
                                 * - 메서드 영역(Method Area)에 저장되어 있어 인스턴스 없이도 접근 가능합니다.
                                 *
                                 * 따라서 [`static` 메서드는 `receiver`가 불필요](https://docs.oracle.com/javase/specs/jls/se21/html/jls-8.html#jls-8.4.3.2)합니다.
                                 * > A class method is always invoked without reference to a particular object.
                                 *
                                 * `static` 메서드 호출 시 첫 번째 매개변수인 `receiver`는 인스턴스가 아닌 클래스 자체에 바인딩되기 때문입니다.
                                 * 인스턴스를 필요로 하지 않는 메서드이기 때문에 `receiver`가 불필요하고, 따라서 [java.lang.reflect.Method.invoke] 시에 첫번째 매개변수에 `null`을 전달합니다.
                                 */
                                null  // 패키지 레벨 함수 또는 컴패니언 객체는 인스턴스 생성 없이 null로 호출
                            } else {
                                clazz.getDeclaredConstructor().newInstance()
                            }
                        }

                        method.annotations.forEach methodAnnotations@{
                            if (it is RunExample && it.printTitle) {
                                printTitle(if (it.title == "") method.name else it.title)
                            }
                        }

                        /*method.invoke(instance)*/
                        // Method.invoke 대신 kotlinFunction 을 사용합니다.
                        method.kotlinFunction?.let {
                            if (it.isSuspend) {
                                it.callSuspend()
                                return@methods
                            }

                            it.call()
                        }
                    }
                }
            }
        }

        private fun printTitle(name: String) {
            val banner = "=".repeat(20)
            println("$banner $name $banner")
        }

        private fun scanTargetClasses(packageName: String): Set<Class<*>> {
            val classes = mutableSetOf<Class<*>>()
            packageName.replace('.', '/')

            val classLoader = Thread.currentThread().contextClassLoader as SecureClassLoader

            for (resource in classLoader.getResources(packageName)) {
                when (resource.protocol) {
                    "file" -> classes += findClassesInDirectory(File(resource.file), packageName)
                    "jar" -> classes += findClassesInJar(resource, packageName)
                }
            }

            return classes
        }

        /**
         * @param directory 컴파일된 클래스 파일이 위치하는 디렉토리 경로입니다.
         *
         *      /path/to/project/app/build/classes/kotlin/main/safety
         *
         * @param packageName
         */
        private fun findClassesInDirectory(directory: File, packageName: String): Set<Class<*>> {
            if (!directory.exists()) {
                return emptySet()
            }

            val classes = mutableSetOf<Class<*>>()

            directory.listFiles()?.forEach { fileOrDirectory ->
                if (fileOrDirectory.isDirectory) {
                    classes += findClassesInDirectory(
                        fileOrDirectory,
                        "$packageName.${fileOrDirectory.name}"
                    )
                    return classes
                }

                if (!fileOrDirectory.name.endsWith(CLASS_SUFFIX)) {
                    println("It's not compiled class file: $fileOrDirectory")
                    return classes
                }

                fileOrDirectory.name.removeSuffix(CLASS_SUFFIX).let { className ->
                    /**
                     * Examples:
                     * - `SafetyExampleKt$calculateByCoroutineScope$2$1$1$1.class` -> `safety.SafetyExampleKt$calculateByCoroutineScope$2$1$1$1`
                     * - `Person.class` -> `safety.Person`
                     * - `SafetyExampleKt.class` -> `safety.SafetyExampleKt`
                     *
                     * [unsafeCalculateByCoroutineScope] 경우 함수지만, 코루틴 빌더 ([coroutineScope]와 [launch])는 익명 내부 클래스 및 람다 표현식으로 변환됩니다.
                     * 그리고 컴파일된 클래스 파일에서는 `$`와 숫자 표기법을 사용하여 이 구조를 나타내는 클래스들이 생성됩니다.
                     *
                     * ```
                     * class OuterClass {
                     *     inner class InnerClass
                     *
                     *     fun method() {
                     *         val anonymous = object : Runnable {
                     *             override fun run() {}
                     *         }
                     *         anonymous.run()
                     *
                     *         val lambda = { x: Int -> x * 2 }
                     *         lambda(2)
                     *     }
                     * }
                     * ```
                     *
                     * ```
                     * OuterClass$InnerClass.class          // 명시적 내부 클래스: `inner class InnerClass`
                     * OuterClass$method$anonymous$1.class  // `val anonymous = object : Runnable`
                     * OuterClass.class                     // `class OuterClass`
                     * ```
                     */
                    val fullyQualifiedClassName = "$packageName.$className"
                    classes.add(Class.forName(fullyQualifiedClassName))
                }
            }
            return classes
        }

        private fun findClassesInJar(resource: URL, packageName: String): Set<Class<*>> {
            val classes = mutableSetOf<Class<*>>()

            /**
             * Example:
             *
             *    file:/path/to/study/jvm/examples/HowWorks/app/build/libs/safety-example.jar!/safety
             */
            val resourcePath = resource.path

            /**
             * Example:
             *
             *    /path/to/study/jvm/examples/HowWorks/app/build/libs/safety-example.jar
             */
            val jarPath = resourcePath.substring(5, resource.path.indexOf("!"))

            JarFile(jarPath).use { jar ->
                val path = packageName.replace('.', '/')
                jar.entries().asSequence().forEach { entry ->
                    if (entry.name.startsWith(path) && entry.name.endsWith(CLASS_SUFFIX) && !entry.isDirectory) {
                        val className = entry.name.replace('/', '.').removeSuffix(CLASS_SUFFIX)
                        classes.add(Class.forName(className))
                    }
                }
            }

            return classes
        }
    }
}
