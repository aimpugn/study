package functions

import util.RunExample

/**
 * `infix` 함수는 특수한 형태의 함수 호출 구문을 제공하여, 호출 시 점(`.`)과 괄호(`()`)를 생략하고 자연어처럼 표현할 수 있게 합니다.
 * 이를 통해 가독성을 높일 수 있으며, 특히 도메인 특정 언어(DSL)를 만들 때 유용합니다.
 * infix 함수는 다음 조건을 만족해야 합니다:
 * 1. 멤버 함수 또는 확장 함수여야 합니다.
 * 2. 하나의 매개변수만 가질 수 있습니다.
 * 2. 해당 매개변수는 기본값을 가질 수 없고, 가변 인자를 받을 수 없습니다.
 *
 * References:
 * - [Infix notation](https://kotlinlang.org/docs/functions.html#infix-notation)
 */
@RunExample
fun infixFunctions() {
    println("1 shl 2: ${1 shl 2}")
    println("1.shl(2): ${1.shl(2)}")
}

infix fun Int.shl(x: Int): Int {
    return this shl x
}