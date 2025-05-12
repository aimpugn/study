package com.example.testing.services;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Isolation;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.util.HashSet;
import java.util.Set;
import java.util.random.RandomGenerator;
import java.util.random.RandomGeneratorFactory;

@Service
@Transactional(
    transactionManager = "txManager",
    propagation = Propagation.REQUIRED,
    isolation = Isolation.READ_COMMITTED,
    rollbackFor = Exception.class
)
public class InnerService {
    RandomGenerator rg = RandomGeneratorFactory.getDefault().create();

    public Set<String> innerMethod() throws Exception {
        var result = new HashSet<String>();

        System.out.println("innerMethod: ERROR");
        throw new Exception("innerMethod Exception");

//        System.out.println("innerMethod: SUCCESS");

//        result.add("SUCCESS");

    }
}
