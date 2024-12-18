/**
 * - [final Variables](https://docs.oracle.com/javase/specs/jls/se21/html/jls-4.html#jls-4.12.4)
 * - [sealed, non-sealed, and final Classes](https://docs.oracle.com/javase/specs/jls/se21/html/jls-8.html#jls-8.1.1.2)
 * - [final Fields](https://docs.oracle.com/javase/specs/jls/se21/html/jls-8.html#jls-8.3.1.2)
 */
package keywords

import util.RunExample

const val TOP_LEVEL_PROPERTY = 100

// open object Singleton => Modifier 'open' is not applicable to 'standalone object'
object Singleton {
    const val OBJECT_CONSTANT_VAL = "Singleton constant"
    // const var OBJECT_CONSTANT_VAR = "Cannot const var" => Modifier 'const' is not applicable to 'vars'
}
// 단일 인스턴스를 보장하는 싱글턴에 대해, 기존 클래스 기반으로 새로운 클래스 생성하는 상속은 불가능합니다.
// class DerivedSingleton : Singleton()

open class FinalProperties {
    var FINAL_VAR_CLASS_FIELD = "it's a final var class field"
    val FINAL_VAL_CLASS_FIELD = "it's a final val class field"

    companion object {
        /**
         * const val은 다음과 같은 곳들에서만 사용될 수 있습니다.
         * - top-level properties
         * - objects(including companion objects)
         *
         * 컴파일 시에 값을 알고 있어야 하므로, primitive 또는 문자열 타입에 대해서만 사용할 수 있습니다.
         */
        const val FINAL_COMPILE_TIME_CONSTANT = "resolved at compile-time"

        val FINAL_RUNTIME_CONSTANT = "resolved at runtime"

        var FINAL_MUTABLE_LIST = mutableListOf("mutable", "string", "list")

        // companion object는 싱글톤이므로, 상속과 같이 사용할 수 없습니다.
        // open val OPEN_CONSTANT = "This is a constant" => 'open' has no effect in an object
    }

    val cannotBeOverridden = "This property cannot be overridden"
    open val canBeOverridden = "This property can be overridden"

    fun defaultFunctionIsFinal() = println("메소드는 기본적으로 final 입니다.")

    fun cannotBeOverridden() {
        // final val localValue = "로컬 변수에 final 선언 불가능하고 중복입니다." => Modifier 'final' is not applicable to 'local variable'
        println("Original cannotBeOverridden")
    }

    /**
     * Kotlin에서 `final` 키워드는 Java의 final과 다릅니다.
     *
     * Java `final`:
     * - 변수 경우: 한번 할당하면 끝났음을 의미합니다. 따라서 한번 할당하면 수정이 불가능하지만 하위 클래스에서 오버라이드 할 수 있습니다.
     * - 메서드 경우: 하위 클래스에서 오버라이드가 불가능합니다.
     * - 클래스 경우: 다른 클래스가 상속할 수 없습니다.
     *
     * Kotlin `final`: [상속과 과련해서만 사용](https://kotlinlang.org/docs/inheritance.html#overriding-methods)됩니다. 클래스 멤버들의 오버라이드를 막습니다.
     * - 변수 경우: 수정 여부는 `final` 키워드가 아닌 `val`/`var`에 따라 결정됩니다. `var` 경우 재할당이 가능합니다.
     * - 메서드 경우: 하위 클래스에서 오버라이드가 불가능합니다.
     * - 클래스 경우: 다른 클래스가 상속할 수 없습니다.
     */
    fun modifyFinalVarClassField() {
        FINAL_VAR_CLASS_FIELD = "it's a MODIFIED final var class field"
    }

    open fun canBeOverridden() = println("Original canBeOverridden")
}

open class DerivedFinalProperties : FinalProperties() {
    // 'FINAL_VAR_CLASS_FIELD' in 'FinalProperties' is final and cannot be overridden.
    /* override var FINAL_VAR_CLASS_FIELD = "it CAN NOT be overridden" */

    companion object {
        // companion object는 싱글톤이므로, 상속과 같이 사용할 수 없습니다.
        // override val CONSTANT_CAN_BE_OVERRIDDEN = "" => 'CONSTANT_CAN_BE_OVERRIDDEN' overrides nothing
    }

    // override val cannotBeOverridden  = super.canBeOverridden => 'cannotBeOverridden' in 'FinalProperties' is final and cannot be overridden
    override val canBeOverridden = "overridden result: ${super.canBeOverridden}"

    // override fun cannotBeOverridden() = println("Can NOT override") => 'cannotBeOverridden' in 'FinalProperties' is final and cannot be overridden
    // override fun defaultFunctionIsFinal() = println("Can NOT override") => 'defaultFunctionIsFinal' in 'FinalProperties' is final and cannot be overridden

    override fun canBeOverridden() = println("Overridden canBeOverridden")
}

class DerivedDerivedFinalProperties : DerivedFinalProperties() {
    override val canBeOverridden = "overridden again result: ${super.canBeOverridden}"
}

abstract class AbstractClass {
    val cannotBeOverridden = "This property cannot be overridden"

    abstract fun mustBeImplemented()
}

class ImplementAbstractClass : AbstractClass() {
    override fun mustBeImplemented() = println("ImplementAbstractClass: ${super.cannotBeOverridden}")
}

@RunExample
fun finalKeywordExample() {
    println(TOP_LEVEL_PROPERTY) // 100

    val fp = FinalProperties()
    val dfp = DerivedFinalProperties()
    val ddfp = DerivedDerivedFinalProperties()

    // const val 경우 아래와 같이 문자열로 코드에 하드코딩됩니다.(Bytecode Inlining)
    //  resolved at compile-time  java/lang/System  out Ljava/io/PrintStream;  	   java/io/PrintStream  println (Ljava/lang/Object;)V
    // 이는 Java에서 원시 타입 또는 문자열 static final 속성과 비슷합니다.
    println(FinalProperties.FINAL_COMPILE_TIME_CONSTANT) // resolved at compile-time
    println(FinalProperties.FINAL_RUNTIME_CONSTANT) // resolved at runtime
    println(FinalProperties.FINAL_MUTABLE_LIST) // [mutable, string, list]
    FinalProperties.FINAL_MUTABLE_LIST.add("NewElement")
    println(FinalProperties.FINAL_MUTABLE_LIST) // [mutable, string, list, NewElement]
    FinalProperties.FINAL_MUTABLE_LIST = mutableListOf(
        "Can", "reassign", "new", "list", "when", "final var"
    )
    println(FinalProperties.FINAL_MUTABLE_LIST) // [Can, reassign, new, list, when, final var]

    println(fp.canBeOverridden) // This property can be overridden
    println(fp.cannotBeOverridden) // This property cannot be overridden
    println("Before: ${fp.FINAL_VAR_CLASS_FIELD}") // Before: it's a final var class field
    fp.modifyFinalVarClassField()
    println("After: ${fp.FINAL_VAR_CLASS_FIELD}") // After: it's a MODIFIED final var class field

    println(dfp.canBeOverridden) // overridden result: This property can be overridden
    println(dfp.cannotBeOverridden) // This property cannot be overridden
    dfp.canBeOverridden() // Overridden canBeOverridden

    println(ddfp.canBeOverridden) // overridden again result: overridden result: This property can be overridden

    println(Singleton.OBJECT_CONSTANT_VAL) // Singleton constant
    println(ImplementAbstractClass().cannotBeOverridden) // This property cannot be overridden
    ImplementAbstractClass().mustBeImplemented() // ImplementAbstractClass: This property cannot be overridden
}
