package com.example.portable;

import com.example.portable.binary.BinaryServerProperties;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.EnableConfigurationProperties;

@SpringBootApplication
@EnableConfigurationProperties(BinaryServerProperties.class)
public class FullyPortableApplication {

    public static void main(String[] args) {
        SpringApplication.run(FullyPortableApplication.class, args);
    }
}
