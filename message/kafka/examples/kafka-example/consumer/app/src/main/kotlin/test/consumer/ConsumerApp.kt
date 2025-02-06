package test.consumer

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication

@SpringBootApplication
open class ConsumerApp

fun main(args: Array<String>) {
    runApplication<ConsumerApp>(*args)
}