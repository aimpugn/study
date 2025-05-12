package com.example.testing.controller;

import com.example.testing.services.OuterService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("/tx")
public class TransactionController {
    private final OuterService outerService;

    public TransactionController(OuterService outerService) {
        this.outerService = outerService;
    }

    @GetMapping("/unexpectedRollbackException")
    public Map<String, Object> unexpectedRollbackException() throws Exception {
        return outerService.unexpectedRollbackException();
    }

    @GetMapping("/resolveUnexpectedRollbackException")
    public Map<String, Object> resolveUnexpectedRollbackException() throws Exception {
        return outerService.resolveUnexpectedRollbackException();
    }
}
