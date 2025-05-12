package com.example.testing;

import com.example.testing.utils.TransactionHelper;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class TransactionHelperConfig {
    @Bean
    public TransactionHelper txHelper() {
        return new TransactionHelper();
    }
}
