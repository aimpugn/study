package functions

import util.RunExampleInvoker

suspend fun main() {
    RunExampleInvoker.invoke(::main.javaClass.packageName)
}