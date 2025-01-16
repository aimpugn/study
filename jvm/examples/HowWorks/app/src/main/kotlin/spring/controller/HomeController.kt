package spring.controller

import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestParam
import org.springframework.web.bind.annotation.RestController
import spring.aop.LoggingBefore

@RestController
class HomeController {
    @GetMapping("/home")
    fun home() = "Hello, World!"


    @LoggingBefore
    @GetMapping("/printName")
    fun printName(@RequestParam("name") name: String) = "Hello, $name"
}