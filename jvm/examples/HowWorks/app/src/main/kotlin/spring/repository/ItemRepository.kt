package spring.repository

import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.data.jpa.repository.Query
import org.springframework.data.jpa.repository.support.SimpleJpaRepository
import org.springframework.data.repository.CrudRepository
import spring.domain.ItemKey

/**
 * [JpaRepository]는 Spring Data가 제공하는 표준 메서드들(ex: `findAll`, `save` 등)을 자동으로 구현해주는데,
 * [CrudRepository] 인터페이스를 구현하는 [SimpleJpaRepository] 클래스를 사용합니다.
 *
 * References:
 * - [Defining Repository Interfaces](https://docs.spring.io/spring-data/rest/reference/data-commons/repositories/definition.html)
 */
interface ItemRepository : JpaRepository<ItemKey, Long>, ItemRepositoryCustom {
    @Query(
        value = "SELECT ik FROM ItemKey ik JOIN FETCH ik.itemValues",
        countQuery = "SELECT COUNT(1) FROM ItemKey ik JOIN FETCH ik.itemValues"
    )
    fun findAllFetchJoin(): List<ItemKey>
}