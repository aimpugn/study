package spring.repository

import jakarta.persistence.EntityManager
import jakarta.persistence.PersistenceContext
import jakarta.persistence.SqlResultSetMapping
import org.springframework.data.jpa.repository.config.EnableJpaRepositories
import org.springframework.stereotype.Repository
import spring.domain.ItemKey
import spring.dto.ItemKeyWithValuesDTO

/**
 * 프래그먼트 인터페이스입니다.
 *
 * [`@Query`](https://docs.spring.io/spring-data/jpa/reference/jpa/query-methods.html#jpa.query-methods.at-query)로 부족한 경우,
 * 임의의 리파지토리를 구현하여 사용할 수 있습니다.
 */
interface ItemRepositoryCustom {
    fun findAllByEntityManager(): List<ItemKey>
    fun findAllItemsByNativeQuery(): List<ItemKeyWithValuesDTO>
}

/**
 * 프래그먼트 인터페이스의 구현체입니다.
 *
 * `Impl`라는 접미사를 붙여야 합니다.
 * 단, `Enable<StoreModule>Repositories` 어노테이션의 `repositoryImplementationPostfix` 속성을 사용하여
 * 임의의 접미사를 사용할 수 있습니다.
 * JPA 경우 [EnableJpaRepositories] 어노테이션을 사용합니다.
 *
 * ```
 * @EnableJpaRepositories(repositoryImplementationPostfix = "")
 * ```
 *
 * 전통적으로 스프링 데이터는 [명명 패턴](https://docs.spring.io/spring-data/commons/docs/1.9.0.RELEASE/reference/html/#repositories.single-repository-behaviour)을 따랐습니다.
 *
 * 리파지토리 인터페이스와 동일한 패키지에 위치하며, 리파지토리 인터페이스 이름 뒤에 implementation 접미사가 붙는 경우
 * 커스텀 구현체로 간주하여 처리했습니다. 하지만, 이렇게 이름을 따르는 클래스는 원치 않는 동작을 초래할 수 있습니다.
 * 과거에는 하나의 리포지토리 인터페이스에 대해 단 하나의 커스텀 구현체만 연결할 수 있었기 때문에,
 * 만약 여러 개의 커스텀 구현체를 사용하려면 별도의 리포지토리 인터페이스를 만들어야 했습니다.
 *
 * 이러한 단일 커스텀 구현체 네이밍 방식을 더 이상 권장하지 않으며,
 * 프래그먼트 기반 프로그래밍 모델(fragment-based programming model)로 마이그레이션할 것을 권장합니다.
 *
 * 스프링 데이터 리파지토리는 리파지토리 합성(repository composition)을 형성하는 프래그먼트를 사용하여 구현됩니다.
 * 프래그먼트는 다음과 같습니다:
 * - base repository
 * - functional aspects(ex: [QueryDSL](https://docs.spring.io/spring-data/jpa/reference/repositories/core-extensions.html#core.extensions.querydsl))
 * - 임의의 인터페이스 및 그 구현체
 *
 *
 * References:
 * - [Custom Repository Implementations](https://docs.spring.io/spring-data/jpa/reference/repositories/custom-implementations.html)
 * - [spring-data-examples/jpa repo](https://github.com/spring-projects/spring-data-examples/tree/main/jpa)
 */
@Repository
class ItemRepositoryImpl : ItemRepositoryCustom {

    @PersistenceContext
    private lateinit var entityManager: EntityManager

    override fun findAllByEntityManager(): List<ItemKey> {
        val query = entityManager.createQuery(
            "SELECT ik FROM ItemKey ik JOIN FETCH ik.itemValues", ItemKey::class.java
        )
        return query.resultList
    }

    /**
     * [ItemKey]에 설정된 [SqlResultSetMapping]을 통해
     * 각 로우는 [ItemKeyWithValuesDTO] 인스턴스가 되고, 그 전체 리스트가 반환됩니다.
     */
    override fun findAllItemsByNativeQuery(): List<ItemKeyWithValuesDTO> {
        val query = entityManager.createNativeQuery(
            """
                SELECT 
                    ik.id AS id,
                    ik.item_key AS itemKey,
                    ik.created AS created,
                    ik.modified AS modified,
                    iv.id AS itemValueId,
                    iv.item_value AS itemValue,
                    iv.created AS itemValueCreated,
                    iv.modified AS itemValueModified
                FROM item_keys ik 
                JOIN item_values iv ON ik.id = iv.item_key_id;
            """.trimIndent(),
            ItemKeyWithValuesDTO.MAPPING_NAME
        )

        // SqlResultSetMapping 대신 직접 매핑하는 것도 가능합니다:
        // ```
        // query.unwrap(NativeQuery::class.java)
        //     .setTupleTransformer { tuple, aliases ->
        //         // 결과를 직접 매핑
        //         ItemKeyWithValuesDTO(
        //             id = tuple[0] as Long,
        //             itemKey = tuple[1] as String,
        //             created = tuple[2] as Instant,
        //             modified = tuple[3] as Instant,
        //             itemValue = tuple[4] as String,
        //             ...
        //         )
        //     }
        // ```

        // query.resultList 는 List<Any?>를 리턴하므로, 그 요소가 `ItemKeyWithValuesDTO`인지를 체크합니다.
        return query.resultList.filterIsInstance<ItemKeyWithValuesDTO>()
    }
}