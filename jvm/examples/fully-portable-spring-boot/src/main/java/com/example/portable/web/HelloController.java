package com.example.portable.web;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;

@RestController
public class HelloController {

    @GetMapping("/hello")
    public Map<String, Object> hello(@RequestParam(defaultValue = "world") String name) {
        Map<String, Object> response = new HashMap<>();
        response.put("message", "Hello, " + name);
        response.put("time", Instant.now().toString());
        return response;
    }
}

