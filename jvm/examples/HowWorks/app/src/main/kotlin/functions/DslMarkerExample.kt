package functions

import util.RunExample

/**
 * Type-safe 빌더를 사용하면 복잡한 계층적 데이터 구조를 구축하는 데 적합한
 * kotlin 기반의 DSL을 만들 수 있습니다.
 * ```
 * html {
 *     head {
 *         title("Page Title")
 *     }
 *     body {
 *         h1("Header")
 *         p("This is a paragraph.")
 *     }
 * }
 * ```
 *
 * 여기서 `html`은 사실 람다 표현식을 인자로 받는 함수 호출입니다.
 * 그리고 인자로 받는 `init`은 [리시버 있는 함수 타입](https://kotlinlang.org/docs/lambdas.html#function-literals-with-receiver)입니다.
 * ```
 * class HTML {
 *     fun body() { ... }
 * }
 *
 * fun html(init: HTML.() -> Unit): HTML {
 *     //         ^^^^^^^^^^^^^^^
 *     //         function type with receiver
 *     val html = HTML() // 리시버 인스턴스를 생성합니다.
 *     html.init() // 생성된 리시버 인스턴스를 람다로 전달합니다.
 *     return html
 * }
 *
 * html { // 리시버 있는 람다가 여기부터 시작합니다.
 *     ┌> `HTML` 타입의 인스턴스(리시버)는 `this` 키워드를 사용하여 접근할 수 있습니다.
 *     this.head { ... }
 *           ├> `HTML`의 멤버 함수
 *     this.body { ... }
 * }
 * ```
 *
 * 여기서 리시버 인스턴스는 확장 함수 또는 람다 리시버가 작업을 수행할 때 암시적으로 참조하는 객체입니다.
 * ```
 * fun String.shout(): String = this.uppercase() + "!!!"
 * // "hello".shout() 호출 시, 리시버 객체는 "hello".
 * ```
 *
 * 그런데 여러 DSL 블록을 중첩해서 사용하는 경우, 여러 리시버 객체가 스코프에 존재하게 됩니다.
 * 이로 인해, 어떤 리시버를 사용해야 할지 혼동이 발생할 수 있습니다.
 *
 * [DslMarker]는 가장 가까운 리시버를 사용하도록 강제하거나, 명시적 리시버 사용을 강제하는 메커니즘으로,
 * Kotlin DSL에서 스코프 간 혼동을 방지하는 데 사용됩니다.
 * [DslMarker]를 사용하면 특정 DSL의 리시버가 다른 DSL 리시버와 혼동되지 않도록 제한할 수 있습니다.
 *
 * ```
 * // `body`는 `html` 리시버에 정의된 함수이지만,
 * // `head` 리시버가 우선 적용되어 잘못된 스코프에서 호출됩니다.
 * html {
 *     head {
 *         title("Page Title")
 *         body { // 스펙상 body는 head 내부에서 호출될 수 없지만,
 *                // `DslMarker` 없는 경우 호출될 수 있는 것처럼 보일 수 있습니다.
 *             h1("Header")
 *         }
 *     }
 * }
 * ```
 *
 * [DslMarker]를 사용할 경우, [Body]는 [HTML] 리시버 객체 내에서 사용해야 하므로,
 * `head`나 [Table] 같은 스코프에서 사용하려고 하면 아래와 같은 에러가 발생합니다.
 * ```
 * ...cannot be called in this context with an implicit receiver. Use an explicit receiver if necessary.
 * ```
 *
 * 즉, [DslMarker]를 사용하면 명시적인 리시버 없이도 리시버 간 혼동 방지할 수 있습니다.
 *
 * 마킹 규칙: 암시적 리시버가 `@Ann`으로 마킹되었다고 간주되는 경우.
 * - 리시버 타입 자체가 마킹된 경우
 * - 리시버 타입의 분류자가 마킹된 경우(ex: `List<T>` 경우 `List`가 `classifier`)
 * - 리시버의 상위 클래스나 상위 인터페이스가 마킹된 경우
 *
 * References:
 * - [Type-safe builders](https://kotlinlang.org/docs/type-safe-builders.html)
 */
@DslMarker
annotation class HtmlDsl

@HtmlDsl
open class HTML {
    /**
     * [HTML]의 멤버 함수입니다.
     *
     * 리시버 있는 함수는 다음과 같이 다양한 방식으로 호출할 수 있습니다.
     *
     * ```
     * Body().init()
     * init.invoke(Body())
     * init(Body())
     * // 셋 모두 모두 아래와 같이 컴파일 됩니다.
     * // Compiled:
     * //   init.invoke(new Body());
     * ```
     *
     * ```
     * Body().apply(init)
     * // Compiled:
     * //   Body var2 = new Body();
     * //   init.invoke(var2);
     * ```
     */
    fun body(init: Body.() -> Unit) {
        Body().init()
    }

    fun head(init: Head.() -> Unit) {
        Head().init()
    }
}

@HtmlDsl
open class Head

@HtmlDsl
open class Body {
    fun table(init: Table.() -> Unit) {
        println("Body.table called")
        Table().init()
    }
}

// HTML DSL 구조 정의
@HtmlDsl
open class Table {
    fun tr(init: Tr.() -> Unit) {
        println("Table.tr called")
        Tr().init()
    }
}

@HtmlDsl
class Tr {
    fun td(init: Td.() -> Unit) {
        println("Tr.td called")
        Td().init()
    }
}

@HtmlDsl
class Td {
    fun text(value: String) {
        println("Td.text: $value")
    }
}

/**
 * DSL 빌더 함수입니다.
 * [html] 함수 안에서는 [Table.tr] 메서드 등을 자유롭게 사용할 수 있습니다.
 */
fun html(init: HTML.() -> Unit): HTML {
    return HTML().apply(init)
}

/**
 * [Table] 클래스에 [HtmlDsl]이 마킹되어 있으므로,
 * [Table]을 상속하는 [TableExtension]도 [HtmlDsl]이 마킹된 것으로 간주됩니다.
 */
class TableExtension : Table()

fun tableExtension(init: TableExtension.() -> Unit): TableExtension {
    return TableExtension().apply(init)
}

/**
 * "type's classifier"는 보통 타입의 "컴파일 타임 클래스"를 의미합니다.
 * 즉, 컴파일 타임에 타입이 어떤 클래스나 인터페이스로 분류되는지를 나타내는 메타데이터로 이해할 수 있습니다.
 * 가령 `classifier`는 [kotlin.reflect.KType.classifier]를 통해 확인할 수 있습니다.
 *
 * ```
 * println("Classifier of Table: ${tb::class.createType().classifier}") // Classifier of Table: class functions.Table
 * println("Classifier of TableExtension: ${tbe::class.createType().classifier}") // Classifier of TableExtension: class functions.TableExtension
 * ```
 *
 * 따라서 `Box<String>`, `Box<Int>` 등 [Box]를 포함하는 모든 인스턴스는 동일하게 [HtmlDsl]가 마킹된 것으로 인식됩니다.
 *
 */
@HtmlDsl
class Box<T> {
    fun doSomething(something: T) {
        println("Do something: $something")
    }
}

fun <T> box(init: Box<T>.() -> Unit): Box<T> {
    return Box<T>().apply(init)
}

@RunExample
fun htmlDsl() {
    // DSL 사용
    html {
        head {}
        body {
            table {
                tr {
                    td {
                        text("Cell 1")
                        // tr 태그는 tr 태그 하위에 중첩될 수 없습니다.
                        // - https://html.spec.whatwg.org/multipage/tables.html#the-tr-element
                        // tr { }  // @DslMarker 사용시 상위 스코프인 Table.tr 호출 불가능
                        // ^ 'fun tr(init: Tr.() -> Unit): Unit' cannot be called in this context with an implicit receiver. Use an explicit receiver if necessary.
                        this@table.tr {} // 단, explicit receiver 통해서 Table.tr 호출 가능
                    }
                }
            }
        }
    }

    tableExtension {
        tr { }
        box<String> {
            doSomething("test")
            // tr {} // 'fun tr(init: Tr.() -> Unit): Unit' can't be called in this context by implicit receiver. Use the explicit one if necessary
        }
    }
}