package com.example.testing.services;

import com.example.testing.utils.TransactionHelper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Isolation;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.util.HashMap;
import java.util.Map;

@Service
@Transactional(
    transactionManager = "txManager",
    propagation = Propagation.REQUIRED,
    isolation = Isolation.READ_COMMITTED,
    rollbackFor = Exception.class
)
public class OuterService {

    @Autowired
    InnerService innerService;

    @Autowired
    TransactionHelper txHelper;

    public Map<String, Object> unexpectedRollbackException() throws Exception {
        var result = new HashMap<String, Object>();

        try {
            innerService.innerMethod();
        } catch (Exception e) {
            System.out.println("outerMethod: innerMethod failed");
        }

        var body = new HashMap<String, String>();
        body.put("key", "value");
        body.put("key2", "value2");
        result.put("body", body);

        return result;
    }

    public Map<String, Object> resolveUnexpectedRollbackException() throws Exception {
        var result = new HashMap<String, Object>();

        var nestedResult = "";

        try {
            txHelper.<Void>nested(() -> { // NESTED 트랜잭션 시작하고 save point
                innerService.innerMethod(); // 익셉션 발생 시 save point로 롤백
                return null;
            });
        } catch (Exception ex) { // 이 시점에서는 NESTED 트랜잭션만 롤백된 상태
            System.out.println("caught -> " + ex.getClass().getSimpleName());
            nestedResult = "NOT_OK"; // 바깥 트랜잭션은 여전히 커밋 가능
        }

        var body = new HashMap<String, String>();
        body.put("result", nestedResult);
        body.put("key", "value");
        body.put("key2", "value2");
        result.put("body", body);

        return result;
    }
}
