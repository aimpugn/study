package functions

import util.RunExample

operator fun Int.times(operation: () -> Unit): () -> Unit = {
    repeat(this) { operation() }
}

@RunExample
fun overloadingOperator() {
    val triplePrint = 3 * {
        println("Hello, World!")
    }
    triplePrint()
}

