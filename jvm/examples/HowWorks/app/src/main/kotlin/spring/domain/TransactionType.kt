package spring.domain

enum class TransactionType {
    /**
     * 입금
     */
    DEPOSIT,

    /**
     * 출금
     */
    WITHDRAWAL,

    /**
     * 계좌 간 송금
     */
    TRANSFER,

    /**
     * ATM 통한 입출금
     */
    AMT,
}