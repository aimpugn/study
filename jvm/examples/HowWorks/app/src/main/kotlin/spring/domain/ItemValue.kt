package spring.domain

import com.fasterxml.jackson.annotation.JsonBackReference
import jakarta.persistence.*
import java.time.Instant

@Entity
@Table(name = "item_values")
data class ItemValue(
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    var id: Long? = null,

    /**
     * [JoinColumn]을 사용하면 외래 키를 필드로 추가할 필요 없습니다.
     * ```
     * @Column(name = "item_key_id", nullable = false)
     * var itemKeysId: Long? = null,
     * ```
     */
    @ManyToOne(
        // - https://docs.jboss.org/hibernate/orm/current/userguide/html_single/Hibernate_User_Guide.html#basic-basic-annotation
        fetch = FetchType.LAZY,
    )
    @JsonBackReference // https://www.baeldung.com/jackson-bidirectional-relationships-and-infinite-recursion#managed-back-reference
    var itemKey: ItemKey?,

    @Column(name = "item_value", nullable = false)
    var itemValue: String,

    @Column(name = "created", updatable = false, nullable = false)
    var created: Instant?,

    @Column(name = "modified", nullable = false)
    var modified: Instant?,
)