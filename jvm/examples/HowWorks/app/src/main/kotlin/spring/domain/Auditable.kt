package spring.domain

import jakarta.persistence.Column
import jakarta.persistence.EntityListeners
import jakarta.persistence.MappedSuperclass
import org.springframework.data.annotation.CreatedDate
import org.springframework.data.annotation.LastModifiedDate
import org.springframework.data.jpa.domain.support.AuditingEntityListener
import java.time.Instant


// `MappedSuperclass`: JPA 엔티티의 매핑 정보로 사용하도록 지정합니다.
// - 여러 엔티티 클래스에 공통되는 상태와 매핑 정보를 정의하는 데 사용합니다.
// - "Mapped"는 JPA가 처리하는 데이터베이스 테이블 또는 컬럼 매핑 정보를 의미합니다.
// - 엔티티와 달리 쿼리될 수 없고 `EntityManager` 또는 `Query` 작업의 인자로 사용될 수 없습니다.
// - https://jakarta.ee/specifications/persistence/3.2/jakarta-persistence-spec-3.2#mapped-superclasses
@MappedSuperclass
@EntityListeners(value = [AuditingEntityListener::class])
abstract class Auditable {
    @Column(nullable = false, updatable = false)
    @CreatedDate
    val created: Instant? = null

    @Column(nullable = false)
    @LastModifiedDate
    var modified: Instant? = null
}