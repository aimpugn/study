package spring.domain

import com.fasterxml.jackson.annotation.JsonManagedReference
import jakarta.persistence.*
import spring.dto.ItemKeyWithValuesDTO
import java.time.Instant

/**
 * 감사 기능을 통해 엔티티를 생성하거나 변경한 사람, 생성 시점, 변경 시점 등을 투명하게 추적할 수 있습니다.
 * [org.springframework.data.annotation.CreatedBy], [org.springframework.data.annotation.CreatedDate],
 * [org.springframework.data.annotation.LastModifiedDate] 등의 어노테이션을 사용할 수 있습니다.
 *
 * [org.springframework.data.jpa.repository.config.EnableJpaAuditing] 활성화가 필요합니다.
 *
 * ```
 * @EntityListeners(AuditingEntityListener::class)
 * @Entity
 * @Table(name = "pairs")
 * data class Pair(
 *      // ... 생략 ...
 *
 *      @CreatedDate
 *      @Column(name = "created", updatable = false, nullable = false)
 *      var created: Instant? = null,
 *
 *      @LastModifiedDate
 *      @Column(name = "modified", nullable = false)
 *      var modified: Instant? = null
 * )
 * ```
 *
 * Reference:
 * - https://docs.spring.io/spring-data/jpa/reference/auditing.html
 */
@Entity
@Table(name = "item_keys")
@SqlResultSetMapping( // JPA 컨텍스트에 매핑 정보가 로드되도록, JPA 엔티티 클래스에 정의해야 합니다.
    name = ItemKeyWithValuesDTO.MAPPING_NAME,
    // 데이터베이스의 결과를 JPA 엔티티 인스턴스로 매핑하는 데 사용합니다.
    // JPA 엔티티 관계 유지, 영속 상태 관리, 데이터 변경 등이 필요할 때 사용합니다.
    entities = [],
    columns = [],
    // 네이티브 쿼리 결과를 DTO 인스턴스로 매핑하는 데 사용합니다.
    // JPA 엔티티로서의 기능이 필요없고 단순 데이터가 필요할 때 사용합니다.
    classes = [
        ConstructorResult(
            targetClass = ItemKeyWithValuesDTO::class,
            columns = [
                ColumnResult(name = "id", type = Long::class),
                ColumnResult(name = "itemKey", type = String::class),
                ColumnResult(name = "created", type = Instant::class),
                ColumnResult(name = "modified", type = Instant::class),
                ColumnResult(name = "itemValueId", type = Long::class),
                ColumnResult(name = "itemValue", type = String::class),
                ColumnResult(name = "itemValueCreated", type = Instant::class),
                ColumnResult(name = "itemValueModified", type = Instant::class),
            ]
        )
    ]
)
class ItemKey(
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    var id: Long? = null,

    @Column(name = "item_key", nullable = false)
    var itemKey: String,

    /**
     * Note:
     * [Set]을 사용하지 않습니다.
     * [FetchType.LAZY] 경우 내부적으로 [org.hibernate.collection.spi.PersistentSet]를 사용하는데,
     * 다음과 같은 에러가 발생합니다.
     *
     * ```
     * java.util.ConcurrentModificationException: null
     * ```
     *
     * 그리고 디버깅해보면 [org.hibernate.sql.results.spi.ListResultsConsumer.consume]에서 `results`에
     * 아래와 같은 익셉션이 포함되어 있습니다.
     *
     * ```
     * Method threw 'java.util.ConcurrentModificationException' exception. Cannot evaluate spring.domain.ItemKey.toString()
     * ```
     *
     * [ItemKey]는 [`data class`](https://kotlinlang.org/docs/data-classes.html)라 `toString`이 무조건 존재하고,
     * 무엇보다 동시(Concurrent) 수정(Modification)으로 인한 익셉션인데 `toString`과 큰 상관이 없어 보입니다.
     *
     * 하지만 `List<ItemValue>`로 타입을 변환하니 정상 동작을 합니다.
     * Lazy fetch 하기 때문에, 실제 메서드가 호출될 때 조회가 이뤄지고,
     * 조회`PersistentSet`을 초기화하고 데이터를 조회하여 채워 넣을 때 이슈가 있는 것으로 추정됩니다.
     *
     * - [List]일 때 쿼리
     *    ```
     *    Hibernate: select ik1_0.id,ik1_0.created,ik1_0.item_key,ik1_0.modified from item_keys ik1_0
     *    Hibernate: select iv1_0.item_key_id,iv1_0.id,iv1_0.created,iv1_0.item_value,iv1_0.modified from item_values iv1_0 where iv1_0.item_key_id=?
     *    ```
     * - [Set]일 때 쿼리
     *    ```
     *    Hibernate: select ik1_0.id,ik1_0.created,ik1_0.item_key,ik1_0.modified from item_keys ik1_0
     *    Hibernate: select iv1_0.item_key_id,iv1_0.id,iv1_0.created,iv1_0.item_value,iv1_0.modified from item_values iv1_0 where iv1_0.item_key_id=?
     *    Hibernate: select iv1_0.item_key_id,iv1_0.id,iv1_0.created,iv1_0.item_value,iv1_0.modified from item_values iv1_0 where iv1_0.item_key_id=?
     *    ```
     * 테스트 해보니 매핑의 문제는 아니고, [ItemKey]와 [ItemValue] 모두 `data class`인 경우 발생합니다.
     * 둘 중 하나라도 `data class`가 아니면 정상적으로 동작합니다.
     *
     * `data class`는 자동으로 [equals], [hashCode], [toString]을 생성합니다.
     * 그런데 [ItemKey]를 가져올 때 어디선가 [ItemKey.toString]을 호출되고,
     * 내부적으로 선언된 [ItemKey.itemValues]에 접근하게 되어 초기화를 시도하는 것으로 보입니다.
     * 그래서 [ItemKey.itemValues]를 초기화하는 와중에 `Set`을 만들기 위해 [hashCode]를 호출하게 되고,
     * 그러면 완전히 로드되지 않은 시점에서 다시 로딩할 가능성이 있습니다.
     *
     * [equals], [hashCode], [toString]을 오버라이드해서 [FetchType.LAZY] 대상 필드를 접근해보면, 다음과 같은 순서로 출력되다가
     * [java.util.ArrayList.Itr.checkForComodification]에서 [java.util.ArrayList.modCount]와
     * [java.util.ArrayList.Itr.expectedModCount]가 맞지 않아 CME 익셉션이 발생합니다.
     * ```
     * At toString: ItemKey(id=1, itemKey=some_key, itemValues=[], created=2025-01-16T06:59:53.200189Z, modified=2025-01-16T06:59:53.200189Z)
     * At hashCode: ItemKey(id=1, itemKey=some_key, itemValues=[], created=2025-01-16T06:59:53.200189Z, modified=2025-01-16T06:59:53.200189Z)
     * At hashCode: ItemValue(id=1, itemKey=spring.domain.ItemKey@69f08e1d, itemValue=some_value1, created=2025-01-16T06:59:53.201841Z, modified=2025-01-16T06:59:53.201841Z)
     *
     * java.util.ConcurrentModificationException: null
     * ... 생략 ...
     * ```
     * 즉, [toString]에서 이미 로딩을 시도하는데, 다시 [hashCode]에서 로딩을 시도하여 중첩되고,
     * [java.util.ArrayList.modCount]는 4인데, [java.util.ArrayList.Itr.expectedModCount]는 2여서 CME가 발생합니다.
     * 단순한 엔티티는 `data class`를 사용하고, 양방향 매핑에 [FetchType.LAZY]를 사용하는 경우에는 의도치 않은 동작을 방지하기 위해
     * 일반 클래스로 선언하고 [toString], [hashCode], [equals]를 직접 구현합니다.
     *
     * 그리고 'N+1' 문제가 있으므로 Fetch Join, [EntityGraph], [NamedEntityGraphs]를 사용합니다.
     * - Fetch Join: [spring.repository.ItemRepositoryCustom.findAllFetchJoin] 참고
     * - [SpringData N+1 solution with [NamedEntityGraph]](https://medium.com/jpa-java-persistence-api-guide/springdata-n-1-solution-with-namedentitygraph-8b101292261a)
     *
     *
     * References:
     * - [Bidirectional `@OneToMany`](https://docs.jboss.org/hibernate/orm/current/userguide/html_single/Hibernate_User_Guide.html#associations-one-to-many-bidirectional)
     * - [Mapping associations to tables](https://docs.jboss.org/hibernate/orm/current/introduction/html_single/Hibernate_Introduction.html#join-table-mappings)
     * - [12.6. Dynamic fetching via Jakarta Persistence entity graph](https://docs.jboss.org/hibernate/orm/current/userguide/html_single/Hibernate_User_Guide.html#fetching-strategies-dynamic-fetching-entity-graph)
     */
    @OneToMany(
        /**
         * [ItemValue.itemKey] 필드에 의해 맵핑됨(mapped by)을 지정합니다.
         * 이는 양방향(Bidirectional) 매핑에서 부모가 소유자(owner)가 아니고,
         * 자식([ItemValue])이 소유자(owner)임을 의미합니다.
         * 자식이 외래 키를 가리키기 때문에 부모([ItemKey])에는 [JoinColumn]을 사용하지 않습니다.
         *
         * - https://docs.jboss.org/hibernate/orm/current/introduction/html_single/Hibernate_Introduction.html#associations
         */
        mappedBy = "itemKey",
        /**
         * 영속성 컨텍스트와 상호작용하는 [EntityManager]의 작업들을 자식에게 전파하는 기능
         * - https://docs.jboss.org/hibernate/orm/current/introduction/html_single/Hibernate_Introduction.html#cascade
         */
        cascade = [CascadeType.ALL],
        orphanRemoval = true,
    )
    @JsonManagedReference // https://www.baeldung.com/jackson-bidirectional-relationships-and-infinite-recursion#managed-back-reference
    // @ElementCollection // 엔티티가 아닌 기본 또는 `Embeddable` 타입의 값을 갖는 컬렉션을 매핑할 때 사용합니다.
    var itemValues: Set<ItemValue>,

    @Column(name = "created", updatable = false, nullable = false)
    var created: Instant?,

    @Column(name = "modified", nullable = false)
    var modified: Instant?,
) {
    /**
     * 엔티티 객체가 데이터베이스와 동기화되기 전에 시간 정보를 자동으로 설정합니다.
     * [PrePersist]를 사용하여 엔티티가 처음 저장되기 전에 호출되어 [created] 필드를 설정합니다.
     *
     * [PrePersist]는
     * - 엔티티 관리자의 영속화 작업이 실제로 실행되거나 전파(cascaded)되기 전에 실행됩니다.
     * - 이 호출은 영속화 작업과 동기화(synchronous)됩니다.
     * - 동일한 트랜잭션 컨텍스트 내에서 실행되므로, 트랜잭션이 롤백되면 수행된 변경도 함께 롤백됩니다.
     *
     * 즉, [PrePersist] 콜백은 [EntityManager.persist] 작업의 일부로 간주되며,
     * 이 작업이 완료되기 전까지는 [EntityManager.persist]가 끝나지 않습니다.
     *
     * 참고로 [EntityManager.persist]는 엔티티를 영속성 컨텍스트에 추가하고, 데이터베이스에 저장될 준비를 합니다.
     * 즉시 SQL을 실행하지 않고, 트랜잭션이 커밋되거나 [EntityManager.flush]가 호출될 때 `INSERT` 쿼리가 실행됩니다.
     * 엔티티가 "managed" 상태로 전환됩니다.
     *
     * References:
     * - [15.4. Jakarta Persistence Callbacks](https://docs.jboss.org/hibernate/orm/6.6/userguide/html_single/Hibernate_User_Guide.html#events-jpa-callbacks)
     */
    @PrePersist
    fun prePersist() {
        val now = Instant.now()
        created = now
        modified = now
        println("At @PrePersist, created: $created, modified: $modified")
    }

    /**
     * 엔티티 객체가 데이터베이스와 동기화되기 전에 시간 정보를 자동으로 설정합니다.
     * [PreUpdate]를 사용하여 엔티티가 처음 저장되기 전에 호출되어 [modified] 필드를 설정합니다.
     *
     * [PreUpdate]는 DB `UPDATE` 작업 전에 실행됩니다.
     * 더 구체적으로 [EntityManager.flush]가 호출될 때 발생하며,
     * [EntityManager.flush]는 트랜잭션 커밋 시점에 자동으로 이루어지거나 명시적으로 호출될 수 있습니다.
     *
     * References:
     * - [15.4. Jakarta Persistence Callbacks](https://docs.jboss.org/hibernate/orm/6.6/userguide/html_single/Hibernate_User_Guide.html#events-jpa-callbacks)
     */
    @PreUpdate
    fun preUpdate() {
        modified = Instant.now()
        println("At @PreUpdate, modified: $modified")
    }

    override fun toString(): String {
        // println("At toString: ItemKey(id=$id, itemKey=$itemKey, itemValues=$itemValues, created=$created, modified=$modified)")
        return "ItemKey(id=$id, itemKey=$itemKey, created=$created, modified=$modified)"
    }

    override fun equals(other: Any?): Boolean {
        if (this == other) return true
        if (other !is ItemKey) return false

        // println("At equals: ItemKey(id=$id, itemKey=$itemKey, itemValues=$itemValues, created=$created, modified=$modified)")
        return id?.let { id == other.id }
            ?: super.equals(other) // 아직 영속화 전인 경우
    }

    override fun hashCode(): Int {
        // println("At hashCode: ItemKey(id=$id, itemKey=$itemKey, itemValues=$itemValues, created=$created, modified=$modified)")
        return id?.hashCode() ?: 0 // 식별자(PK)를 기반으로만 hashCode를 계산
    }
}