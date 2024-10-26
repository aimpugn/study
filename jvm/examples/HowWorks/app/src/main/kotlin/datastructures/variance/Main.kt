/**
 * - variance: 변성
 * - covariant: 공변성
 *      생산자 역할에서의 안전성 때문에 서브타입 관계가 유지됩니다.
 *      공변성에서는 out 키워드를 사용하여 하위 타입을 상위 타입으로 안전하게 반환할 수 있습니다.
 *
 *      - 하위 타입 관계를 유지되므로, 생산시 `Child` 타입의 제네릭 객체가 `Parent` 타입의 제네릭 객체로 안전하게 대체될 수 있습니다.
 *      - `T`를 생산할 때만 사용하는 경우 `out`을 사용하여 공변성을 유지할 수 있습니다.
 *      - 공변성에서 중요한 점은, 타입이 *생산*(반환)되는 상황에서는 *상위* 타입을 사용할 수 있다는 것입니다.
 *
 * - contravariant: 반공변성
 *      반공변성에서는 소비자 역할에서 타입이 안전하게 역전됩니다.
 *      이는 상위 타입을 하위 타입에서 안전하게 소비할 수 있음을 나타냅니다.
 *
 *      - 하위 타입 관계가 반전되므로, 소비시 `Parent` 타입의 제네릭 객체가 `Child` 타입의 제네릭 객체로 대체될 수 있습니다
 *      - T를 소비할 때만 사용하는 경우 in을 사용하여 반공변성을 유지할 수 있습니다.
 *      - 반공변성에서 중요한 점은, 타입이 *소비*되는 상황에서는 *하위* 타입을 사용할 수 있다는 것입니다.
 *
 * - invairant: 불공변성
 *
 * Refernces:
 * - [제네릭의 공변성(Covariance) 및 반공변성(Contravariance)](https://learn.microsoft.com/ko-kr/dotnet/standard/generics/covariance-and-contravariance)
 * - [타입 시스템에서의 변성(Variance)](https://driip.me/d875a384-3fb9-471b-a53b-b3ca52f8238e)
 */
package variance

// `out`: 타입 T를 '생산'하는 메서드만 허용됩니다.
class CovariantTest<out T : Parent>(private val arg: T) {
    fun produce(): T = arg

    // `T`가 소비자 역할인 'in' 위치에 있어 공변성 규칙을 위반합니다.
    // fun print(arg: T) = println(arg) => Type parameter T is declared as 'out' but occurs in 'in' position in type T

    override fun toString(): String = "Covariant"
}

// `in`: 타입 T를 '소비'하는 메서드만 허용됩니다.
// 반공변성은 상위 타입으로부터 값을 안전하게 소비할 수 있음을 나타냅니다
class ContravariantTest<in T : Parent> {
    // 반공변성(in)은 타입 T를 '소비'하는 메서드만 허용합니다.
    fun consume(arg: T) = println(arg)

    // T가 생산자 역할인 'out' 위치에 있어 반공변성 규칙을 위반합니다.
    // fun produce(): T = error("Not allowed") => Type parameter T is declared as 'in' but occurs in 'out' position in type T

    override fun toString(): String = "Contravariant"
}

open class Parent {
    override fun toString() = "Parent instance"
}

class Child : Parent() {
    override fun toString() = "Child instance"
}

fun main() {
    // 1. 공변성(covariant)
    //   서브타입 관계가 제네릭 타입에서도 유지되는 것을 의미합니다.
    //   예를 들어, `String` 타입은 `Any`의 서브타입이고, `Child`는 `Parent`의 서브타입입니다.
    //   `out` 키워드시에는 공변적이므로 서브타입 관계가 유지됩니다.
    //   따라서 `Child` 타입과 `Spuer` 타입이 그대로 처리됩니다.
    //
    println(CovariantTest<Parent>(Parent()).produce()) // Parent instance
    println(CovariantTest<Parent>(Child()).produce()) // Child instance
    //   `CovariantTest<Child>`가 `CovariantTest<Parent>`를 대체
    val covariantChild: CovariantTest<Parent> = CovariantTest<Child>(Child())
    println(covariantChild.produce()) // Child instance
    // 하지만 `fun print(arg: T) = println(arg)`처럼 입력으로 받는 것은 안됩니다.

    // 2. 반공변성(contravariant)
    //   반면 제네릭에 `in` 키워드를 사용하여 서브타입 관계를 역전시키는 것을 의미합니다.
    //   상위 타입으로부터 값을 안전하게 소비할 수 있음을 나타내고,
    //   `in T : Parent` 경우 Parent를 상속하는 어떤 타입을 입력으로 받겠다는 것을 의미합니다.
    //   반공변적인 관계인 경우 서브타입 관계가 역전되므로 `ContravariantTest<Child>`에 `ContravariantTest<Parent>`가
    //   할당될 수 있습니다. 반공변성은 "소비자" 또는 "쓰기 전용" 위치에서 안전합니다.
    ContravariantTest<Parent>().consume(Parent()) // Parent instance
    ContravariantTest<Parent>().consume(Child()) // Child instance
    ContravariantTest<Child>().consume(Child()) // Child instance
    //   이 경우에는 `ContravariantTest<Parent>` 타입이 `ContravariantTest<Child>`으로 할당되었으므로,
    //   더이상 `Parent`를 소비할 수 없습니다.
    val contravariantParent: ContravariantTest<Child> = ContravariantTest<Parent>()
    contravariantParent.consume(Child()) // Child instance
    // contravariantParent.consume(Parent()) => Type mismatch: inferred type is Parent but Child was expected

    // 정리:
    //
    //    [Parent]
    //     ╱   ╲
    //   ╱       ╲
    // [ C h i l d ]
    // 일반적으로 자식 클래스에서 더 많은 기능이나 속성이 있습니다.
    // - out T : Parent
    //     자식 클래스는 항상 부모 클래스를 확장하여 부모 클래스의 모든 속성과 기능을 포함하므로,
    //     생산(리턴)시에 안전하게 부모 클래스로 리턴할 수 있음을 의미합니다.
    //
    // - in T : Parent
    //     반면 in T : Parent 한다는 것은 `Parent` 타입과 그 하위 타입을 소비할 수 있음을 의미합니다.
    //     - `ContravariantTest<Parent>`는 Parent와 그 하위 타입인 Child 모두를 소비할 수 있습니다.
    //     - `ContravariantTest<Parent>`를 `ContravariantTest<Child>` 타입의 변수에 할당하면,
    //        이제 `Parent`는 소비할 수 없고, 오직 `Child`만 소비할 수 있게 됩니다.
}
