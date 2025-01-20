package spring.domain

import jakarta.persistence.*

@Entity
@Table(name = "user_accounts")
data class UserAccount(
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    val id: Long?,

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    val user: User,

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "account_id", nullable = false)
    val account: Account,

    @Column(nullable = false)
    val roles: Set<Role>,

    @Column(nullable = false)
    val active: Boolean,
) : Auditable()
