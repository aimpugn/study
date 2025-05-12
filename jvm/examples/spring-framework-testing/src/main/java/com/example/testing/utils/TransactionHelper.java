package com.example.testing.utils;

import org.springframework.transaction.annotation.Isolation;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

public class TransactionHelper {
    @Transactional(
        propagation = Propagation.NESTED,
        isolation = Isolation.READ_COMMITTED,
        rollbackFor = Exception.class
    )
    public <R> R nested(TxCallback<R> callback) throws Exception {
        return callback.doInTx();
    }
}
