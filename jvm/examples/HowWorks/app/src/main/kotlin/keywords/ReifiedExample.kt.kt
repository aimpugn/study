package keywords

import util.RunExample

/**
 * [T] 타입의 모든 멤버(프로퍼티와 메서드 등)를 리턴합니다.
 *
 * 일반적으로 JVM에서 제네릭 타입 매개변수는 런타임에 타입 정보가 지워지는 소거(Erasure) 방식을 사용합니다.
 * 그래서 보통 제네릭 [T]의 타입 정보를 런타임에 알 수 없고, 따라서 리플렉션을 사용할 때도 제네릭 타입 정보를 바로 사용할 수 없습니다.
 *
 * 하지만 `inline` 함수와 `reified` 키워드를 통해 제네릭 타입을 런타임에 사용할 수 있게 해줍니다.
 * `reified` 키워드는 `inline` 함수에서만 사용할 수 있기 때문에 컴파일 타임에 타입이 확정되어 런타임 소거가 필요하지 않게 됩니다.
 * `reified` 타입은 컴파일 타임에 구체적으로 정의된 타입으로 변환되기 때문에 런타임 리플렉션에서도 타입 정보에 접근이 가능합니다.
 *
 * 컴파일러가 함수 호출 시 제네릭 타입을 실제 타입으로 교체하여 인라인으로 삽입합니다.
 * 이렇게 하면 `T::class`와 같이 리플렉션 API(ex: [kotlin.reflect.KClass.members])에서 타입 정보를 사용할 수 있게 됩니다.
 *
 * References:
 * - [Inline functions - Reified type parameters](https://kotlinlang.org/docs/inline-functions.html#reified-type-parameters)
 */
inline fun <reified T> membersOf() = T::class.members // T 타입의 모든 멤버(프로퍼티와 메서드 등)를 리턴합니다.

/**
 * 일반 팩토리 함수 경우 타입 파라미터 [T]는 런타임에 소거되므로 직접적으로 사용할 수 없습니다.
 * 따라서 새로운 인스턴스를 생성할 클래스를 직접 매개변수로 전달해야 합니다.
 */
fun <T> normalCreateInstance(clazz: Class<T>): T {
    return clazz.getDeclaredConstructor().newInstance()
}

/**
 * 반면 `inline fun <reified T>` 경우 타입 파라미터 [T]가 런타임에도 유지됩니다.
 * `T::class` 방식으로 [kotlin.reflect.KClass]를 얻어서 주어진 제네릭 [T]의 자바 클래스 정보를 얻을 수 있습니다.
 */
inline fun <reified T> reifiedCreateInstance(): T {
    return T::class.java.getDeclaredConstructor().newInstance()
}

@RunExample
fun reifiedExample() {
    println(membersOf<StringBuilder>().joinToString("\n"))
    check(normalCreateInstance(String::class.java) == reifiedCreateInstance<String>()) {
        "It must be equal"
    }
}
