package spring.controller

import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.RestController
import spring.aop.LoggingBefore

@RestController
open class HomeController {
    @GetMapping("/")
    open fun home() = "Hello, World!"


    @LoggingBefore
    @GetMapping("/{name}")
    open fun printName(@PathVariable name: String) = "Hello, $name"
}