package lambda;

import com.google.common.base.Function;
import kotlin.random.Random;
import util.RunExample;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.function.*;

interface JavaFunctionalInterface<T> {
    int doSomething(T msg);
}

public class LambdaExample {
    private static final int testKotlinFunctionalInterfaceOfKotlin$lambda$0(Object o) {
        System.out.println("It's a FunctionalInterface type anonymous instance " + o);
        return Random.Default.nextInt(100, 1000);
    }

    /**
     * static 방식으로 메서드 참조하기 위해 staic 메서드로 선언합니다.
     *
     * <pre>
     * {@code
     * Non-static method cannot be referenced from a static context
     *
     * }
     * </pre>
     *
     * @param a 앞에 위치할 문자열
     * @param b 뒤에 위치할 문자열
     * @return a와 b를 합친 문자열
     */
    private static String concatString(String a, String b) {
        return a + b;
    }

    /**
     * Kotlin에서 람다를 Java 코드로 컴파일 할 경우 메서드 참조를 사용합니다.
     * <p>
     * 메서드 참조는 기존에 정의된 함수나 메서드를 참조하여 람다 표현식을 대체할 수 있는 간결한 문법입니다.
     * 주로 기존에 존재하는 메서드를 참조하여 람다 표현식을 사용할 때 사용됩니다.
     * <p>
     * 네 종류의 메서드 참조가 있습니다:
     * <ul>
     *  <li>static method</li>
     *  <li>특정 객체의 인스턴스 메서드</li>
     *  <li>특정 타입의 임의의 객체의 인스턴스 메서드</li>
     *  <li>생성자</li>
     * </ul>
     *
     * @see <a href="https://www.baeldung.com/java-method-references">Method References in Java</a>
     */
    @RunExample
    public void kotlinFunctionalInterfaceInJava() {
        // 함수형 인터페이스를 구현하는 익명 클래스 인스턴스를 생성할 때, 메서드 참조를 사용합니다.
        Function<Object, Integer> whenKotlinDiffWithJavaCompiled = LambdaExample::testKotlinFunctionalInterfaceOfKotlin$lambda$0;
        System.out.printf(
                "LambdaExample::testKotlinFunctionalInterfaceOfKotlin$lambda$0's result: %d\n",
                whenKotlinDiffWithJavaCompiled.apply("Function<Object, Integer>")
        );

        // 메서드 참조의 시그니처가 함수형 인터페이스의 추상 메서드 시그니처와 일치하기 때문에,
        // Java 컴파일러는 이를 함수형 인터페이스의 익명 구현체로 변환할 수 있습니다.
        JavaFunctionalInterface<String> javaFunctionalInterface = LambdaExample::testKotlinFunctionalInterfaceOfKotlin$lambda$0;
        System.out.printf("javaFunctionalInterface's result: %d\n", javaFunctionalInterface.doSomething("JavaFunctionalInterface"));

        KotlinFunctionalInterface<String> kotlinFunctionalInterface = LambdaExample::testKotlinFunctionalInterfaceOfKotlin$lambda$0;
        System.out.printf("kotlinFunctionalInterface's result: %d\n", javaFunctionalInterface.doSomething("KotlinFunctionalInterface"));
    }

    @RunExample
    public void testIntBinaryOperators() {
        IntBinaryOperator add = (int x, int y) -> {
            return x + y;
        };
        System.out.printf("IntBinaryOperator: %b\n", add.applyAsInt(1, 2) == 3);
    }

    @RunExample
    public void testBiFunctions() {
        BiFunction<Integer, Integer, Integer> add = (a, b) -> a + b; // `Integer::sum` 메서드 참조도 가능
        System.out.println("BiFunction add's result: " + add.apply(5, 10)); // BiFunction add's result: 15

        // Math 클래스의 정적 메서드 참조
        BiFunction<Integer, Integer, Integer> multiply = Math::multiplyExact;
        System.out.println(
                "BiFunction multiply's result: " + multiply.apply(5, 10)
        ); // BiFunction multiply's result: 50

        BiFunction<String, String, String> concat = LambdaExample::concatString;
        System.out.println(
                "BiFunction concat's result: " + concat.apply("Hello", ", World")
        ); // BiFunction concat's result: Hello, World

        System.out.println("Result of combined BiFunctions: " + concat.apply(
                "Add: " + add.apply(5, 10),
                ", Multiply: " + multiply.apply(5, 10)
        )); // Result of combined BiFunctions: Add: 15, Multiply: 50

        Function<Integer, String> toResultMsg = integer -> "Added result is " + integer;
        System.out.println(
                "BiFunction andThen result: " + add.andThen(toResultMsg).apply(5, 10)
        ); // BiFunction andThen result: Added result is 15
    }

    @RunExample
    public void testFunctions() {
        // 문자열 길이를 반환하는 람다
        Function<String, Integer> lengthFunction = str -> str.length();

        System.out.println("lengthFunction's length: " + lengthFunction.apply("lengthFunction")); // lengthFunction's length: 14
    }

    @RunExample
    public void testConsumers() {
        // 문자열을 출력하는 람다
        Consumer<String> printConsumer = str -> System.out.println("printConsumer: " + str);

        printConsumer.accept("Hello, Consumer!"); // printConsumer: Hello, Consumer!
    }

    @RunExample
    public void testPredicates() {
        // 문자열이 비어 있는지 확인하는 람다
        Predicate<String> isEmpty = str -> str.isEmpty();

        System.out.println("Predicate isEmpty: " + isEmpty.test("")); // Predicate isEmpty: true
    }

    @RunExample
    public void testSuppliers() {
        Supplier<List<String>> strListSupplier = ArrayList::new;

        System.out.println("listSupplier make new ArrayList: " + strListSupplier.get()); // listSupplier make new ArrayList: []

        List<String> strList = strListSupplier.get();
        Collections.addAll(strList, "A", "B", "C");
        System.out.println("strList: " + strList); // strList: [A, B, C]
    }


}
