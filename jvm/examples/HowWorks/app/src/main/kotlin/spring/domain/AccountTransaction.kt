package spring.domain

import jakarta.persistence.*
import java.math.BigDecimal
import java.time.Instant
import java.util.*

@Entity
@Table(name = "account_transactions")
data class AccountTransaction(
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    val id: Long,

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "from_account_id", nullable = false, updatable = false)
    val fromAccount: Account,

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "to_account_id", nullable = false, updatable = false)
    val toAccount: Account,

    @Column(nullable = false, updatable = false)
    val amount: BigDecimal,

    @Column(nullable = false, updatable = false)
    val currency: Currency,

    @Column(nullable = false, updatable = false)
    val description: String,

    @Column(nullable = false, updatable = false)
    val transactionType: TransactionType,

    @Column(nullable = false, updatable = false)
    val created: Instant,
)
// `Auditable`은 created, modified 모두 관리하는데, created만 필요하므로 사용하지 않습니다.
