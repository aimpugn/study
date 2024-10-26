/**
 * [Classes and objects](https://kotlinlang.org/docs/classes.html)
 */
package classes

class InitOrderDemo(name: String) {
    val firstProperty = "First property: $name".also(::println)

    init {
        // When want to run some code during object creation
        println("First initializer block that prints $name")
    }

    val secondProperty = "Second property: ${name.length}".also(::println)

    init {
        println("Second initializer block that prints ${name.length}")
    }
}

/**
 * Any를 상속합니다.
 */
class Example // Implicitly inherits from Any

/**
 * 다음 코드와 동일합니다.
 * ```
 * class Person constructor(firstName: String) { /*...*/ }
 * ```
 */
class Person(
    // 생성자는 기본적으로 public 입니다.
    firstName: String, // 외부에서 이 클래스 속성에 접근할 수 없습니다.
    val lastName: String, // val 로 선언되어 수정이 불가능합니다.
    var isEmployed: Boolean = true // 기본값과 함께 선언이 가능합니다.
) {
    // 주 생성자 파라미터는 initializer block에서 사용될 수 있습니다.
    val firstNameUpper = firstName.uppercase()
}

class PrivateConstructor private constructor(val field: String = "private") {
    private constructor() : this("by secondary constructor")

    companion object {
        fun new(field: String = "by new") = PrivateConstructor(field)
        fun create() = PrivateConstructor()
    }

    override fun toString() = "field: $field"
}

// 기본적으로 kotlin의 클래스는 final입니다.
// class SubExample : Example() => This type is final, so it cannot be inherited from

// 상속을 위해서는 클래스를 open으로 선언해야 합니다.
open class Base(p: Int) // Class is open for inheritance
class Derived(p: Int) : Base(p)

fun main() {
    println(Example() is Any) // true

    println(Person("Kim", "SangHyun").lastName) // SangHyun
    println(Person("Kim", "SangHyun").firstNameUpper) // KIM

    InitOrderDemo("hello")
    // Output:
    //  First property: hello
    //  First initializer block that prints hello
    //  Second property: 5
    //  Second initializer block that prints 5

    println(PrivateConstructor.new()) // field: by new
    println(PrivateConstructor.create()) // field: by secondary constructor
}