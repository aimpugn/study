package test.consumer.controller

import org.springframework.web.bind.annotation.RestController
import org.springframework.web.bind.annotation.GetMapping

@RestController
class Home {

    @GetMapping("/")
    fun index(): String = "Hello, World"
}