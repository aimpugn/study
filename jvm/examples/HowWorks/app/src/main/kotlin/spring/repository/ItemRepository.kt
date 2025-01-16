package spring.repository

import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.data.jpa.repository.Query
import org.springframework.stereotype.Repository
import spring.domain.ItemKey

@Repository
interface ItemRepository : JpaRepository<ItemKey, Long> {
    @Query("SELECT DISTINCT ik FROM ItemKey ik JOIN FETCH ik.itemValues")
    fun findAllFetchJoin(): List<ItemKey>
}