package com.example.testing.utils;

@FunctionalInterface
public interface TxCallback<R> {
    R doInTx() throws Exception;   // ★ 예외 전파 허용
}