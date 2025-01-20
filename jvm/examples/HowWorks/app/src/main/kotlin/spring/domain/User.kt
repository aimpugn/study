package spring.domain

import jakarta.persistence.*

@Entity
@Table(name = "users")
data class User(
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    val id: Long,

    @Column(nullable = false)
    var name: String,

    @OneToMany(mappedBy = "user", cascade = [CascadeType.ALL], orphanRemoval = true)
    var accounts: Set<UserAccount>,
) : Auditable()