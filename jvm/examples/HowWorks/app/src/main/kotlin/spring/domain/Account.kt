package spring.domain

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Table
import java.math.BigDecimal
import java.util.*

@Entity
@Table(name = "accounts")
data class Account(
    @Id
    val id: Long,

    @Column(nullable = false)
    val name: String,

    @Column(nullable = false)
    var balance: BigDecimal,

    @Column(nullable = false)
    val currency: Currency,
) : Auditable()
