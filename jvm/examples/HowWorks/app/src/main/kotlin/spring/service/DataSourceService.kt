package spring.service

import io.agroal.api.AgroalDataSource
import jakarta.persistence.EntityManager
import jakarta.persistence.EntityManagerFactory
import jakarta.persistence.PersistenceContext
import jakarta.persistence.spi.PersistenceUnitTransactionType
import org.hibernate.engine.jdbc.connections.spi.ConnectionProvider
import org.hibernate.engine.spi.SessionFactoryImplementor
import org.springframework.aop.support.AopUtils
import org.springframework.orm.jpa.EntityManagerProxy
import org.springframework.orm.jpa.LocalContainerEntityManagerFactoryBean
import org.springframework.orm.jpa.SharedEntityManagerCreator
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.lang.reflect.Proxy
import javax.sql.DataSource

/**
 *
 * ```
 * EntityManagerFactory
 *    ├── ConnectionProvider
 *    │      └── DataSource (javax.sql.DataSource: HikariCP, Agroal 등)
 *    │            └── Connection (JDBC 연결)
 *    │
 *    └── EntityManager
 *          ├── Persistence Context
 *          │      ├── 1차 캐시 (엔티티 객체 저장)
 *          │      ├── 엔티티 상태 관리 (Transient, Managed, Detached, Removed)
 *          │      ├── 변경 감지(Dirty Checking)
 *          │      └── 지연 로딩(Lazy Loading)
 *          └── Query (JPQL, Criteria, Native SQL 실행)
 *                └── SQL (Hibernate가 생성한 SQL 명령어)
 * ```
 *
 * - [EntityManagerFactory]:
 *   데이터베이스와 관련된 설정([ConnectionProvider], [DataSource] 등)을 초기화하고 관리합니다.
 *
 * - [ConnectionProvider]:
 *   JDBC 연결을 생성 및 관리하기 위해 사용하는 추상화 계층으로, 어떻게 커넥션을 생성/관리할 것인지를 결정합니다.
 *   [DataSource]와 상호작용하여 JDBC 연결([java.sql.Connection])을 제공합니다.
 *    - [`DriverManagerConnectionProviderImpl`](https://docs.jboss.org/hibernate/orm/current/javadocs/org/hibernate/engine/jdbc/connections/internal/DriverManagerConnectionProviderImpl.html)
 *    - [`HikariConnectionProvider`](https://docs.jboss.org/hibernate/orm/current/javadocs/org/hibernate/hikaricp/internal/HikariCPConnectionProvider.html)
 *    - [`AgroalConnectionProvider`](https://docs.jboss.org/hibernate/orm/current/javadocs/org/hibernate/agroal/internal/AgroalConnectionProvider.html)
 *    - 그 외 DBCP, C3P0 등
 *
 * - [DataSource]:
 *   표준 JDBC 인터페이스로, [java.sql.Connection] 객체를 얻어오는 방법을 정의합니다.
 *   DB 연결을 얻기 위해 커넥션 풀(예: Hikari, Agroal, DBCP)을 사용하거나,
 *   [java.sql.DriverManager.getConnection]을 직접 호출하는 것도 가능합니다.
 *
 * - [java.sql.Connection]:
 *   데이터베이스와의 실제 연결을 나타냅니다.
 *
 * - [EntityManager]:
 *   JPA 표준의 인터페이스로, 엔티티 객체를 관리하고 데이터베이스와 상호작용합니다.
 *
 * - [PersistenceContext]:
 *   엔티티 객체를 관리하는 메모리 내 저장소로, 엔티티의 상태(`Transient`, `Managed`, `Detached`, `Removed`)를 관리합니다.
 *   1차 캐시, 변경 감지(Dirty Checking), 지연 로딩(Lazy Loading) 등의 기능을 제공합니다.
 *
 * - [jakarta.persistence.Query]:
 *   JPA 표준의 JPQL, Criteria API, Native SQL을 실행합니다.
 */
@Service
class DataSourceService(
    private val entityManagerFactory: EntityManagerFactory,
    private val entityManagerByConstructor: EntityManager,
) {
    /**
     * 엔티티(객체)를 관리하는 [EntityManager]는 다양한 작업들을 수행합니다.
     * - 어떤 엔티티 클래스를 어떤 테이블과 매핑해야 하는지
     * - lazy 로딩을 어떻게 처리할지
     * - 트랜잭션 경계를 어떻게 처리할지
     * - DDL 생성 전략(ddl-auto)은 무엇인지
     * - 변경 감지(Dirty Checking)를 언제, 어떻게 적용할지
     * - 트랜잭션 범위([PersistenceUnitTransactionType.RESOURCE_LOCAL] vs [PersistenceUnitTransactionType.JTA])는 어떻게 할 것인지 등
     * 하이버네이트 경우 [org.hibernate.internal.SessionImpl]로 구현되어 있습니다.
     *
     * 엔티티 매니저를 생성할 때마다 매번 이를 직접 다루기에는 부담이 크기 때문에 설정을 한번 로드하고 필요한 시점에 새로운 [EntityManager]를 생성하는
     * [jakarta.persistence.EntityManagerFactory]가 존재하며, 스프링 실행 시 싱글톤 빈으로 생성되어 사용 및 관리됩니다.
     * 하이버네이트 경우 [org.hibernate.internal.SessionFactoryImpl]로 구현되어 있습니다.
     *
     * [jakarta.persistence.EntityManagerFactory]는 내부적으로 Connection Pool(예: HikariCP)을 통해 데이터베이스 연결을 관리하는데,
     * Connection Pool 설정에 따라 성능 최적화가 이루어질 수 있습니다.
     *
     * 가령 스프링 컨테이너 경우 [LocalContainerEntityManagerFactoryBean.createNativeEntityManagerFactory]를 통해서 팩토리가 생성됩니다.
     * 엔티티 매니저 팩토리는 설정에 따라 적절한 벤더의 팩토리를 생성합니다.
     * 이때 하이버네이트는 커넥션 프로바이더를 생성합니다.
     * - [org.springframework.orm.jpa.vendor.SpringHibernateJpaPersistenceProvider.createContainerEntityManagerFactory]
     * - [org.hibernate.engine.jdbc.env.internal.JdbcEnvironmentInitiator.buildJdbcConnectionAccess]
     *
     * 이 예제의 경우 [org.hibernate.agroal.internal.AgroalConnectionProvider]가 생성되고,
     * [org.hibernate.agroal.internal.AgroalConnectionProvider.configure]에서 [AgroalDataSource] 커넥션 풀을 생성합니다.
     *
     * [EntityManager]는 다음과 같은 작업들을 수행합니다.
     * - 데이터베이스와 상호작용하는 인터페이스로, 엔티티와 DB를 연결
     *    - `persist`: 새로운 엔티티를 영속성 컨텍스트에 추가하고 `INSERT` 작업을 예약합니다.
     *    - `merge`: 준영속(detached) 상태의 엔티티를 다시 영속성 컨텍스트로 가져옵니다.
     *    - `remove`: 엔티티를 삭제 대상으로 표시합니다.
     *    - `find`: 기본 키로 엔티티를 조회하며, 영속성 컨텍스트에 캐시된 데이터가 우선 반환됩니다
     *    - `flush`: 변경 감지(Dirty Checking), 엔티티 변경 사항을 SQL로 변환, JDBC 배치로 예약 등을 수행합니다.
     * - 엔티티 상태 추적하고 변화 감지
     * - 트랜잭션 종료 시점에만 쿼리를 몰아서 수행 등
     *
     * 이를 위해서 엔티티 매니저는 엔티티의 '현재 상태'를 알아야 하는데, 내부적으로
     * [org.hibernate.engine.spi.PersistenceContext]을 저장소로 사용하여 엔티티의 상태가 바뀌는 과정을 일관되게 관리합니다.
     * JPA 표준에서는 두 가지 유형의 영속성 컨텍스트를 정의합니다:
     * - Transaction-scoped Persistence Context (기본):
     *    - 트랜잭션 경계 내에서만 영속성 컨텍스트가 유지됩니다.
     *    - 트랜잭션 종료 시 모든 엔티티는 "detached" 상태로 전환됩니다
     * - Extended Persistence Context:
     *    - 영속성 컨텍스트가 여러 트랜잭션에 걸쳐 유지됩니다.
     *    - 주로 Stateful Session 빈과 함께 사용되며, 트랜잭션 없이도 엔티티를 관리할 수 있습니다
     *
     * 하이버네이트 경우 [org.hibernate.engine.internal.StatefulPersistenceContext]로 구현되어 있습니다.
     * 이 영속성 컨텍스트는 [org.hibernate.internal.SessionImpl]의 수명 동안 유지됩니다.
     *
     * JPA 표준 상태는 하이버네이트 경우 [org.hibernate.event.internal.EntityState]에 정의되어 있습니다.
     * - `Transient` (비영속): 아직 영속성 컨텍스트에 포함되지 않은 상태.
     * - `Managed` (영속): 영속성 컨텍스트에서 관리되는 상태.
     * - `Detached` (준영속): 영속성 컨텍스트에서 분리된 상태.
     * - `Removed` (삭제): 삭제 대상으로 표시된 상태.
     * 다만 1차 캐싱 및 변경 감지 등 다양한 작업을 위해 [org.hibernate.engine.spi.Status]에
     * `READ_ONLY`, `LOADING` 등의 추가적인 상태 정의가 되어 있습니다.
     *
     * 영속성 컨텍스트는 관리하는 엔티티 상태를 활용하여 최적화를 수행합니다.
     * - 1차 캐시로서 동일 엔티티를 반복 조회하더라도 DB에 매번 쿼리를 날리지 않고, 이미 메모리에 있는 엔티티를 재활용합니다.
     *   가령 조회 시 [org.hibernate.event.internal.DefaultLoadEventListener.doOnLoad]가 호출되는데,
     *   내부적으로 [org.hibernate.engine.spi.PersistenceContext.getEntityHolder]로 엔티티 홀더를 조회하여
     *   "managed"인 엔티티가 있는지를 체크합니다.
     *
     * - 변경 감지(Dirty Checking)를 통해 엔티티의 변경 사항을 자동으로 감지하고,
     *   트랜잭션이 끝나는 DB 커밋 시점에 한꺼번에 반영(쓰기 지연, Write-Behind)합니다.
     *
     * 구현된 구조를 보면, 이벤트 기반으로 동작하되, 실제 이벤트를 처리하는 리스너에서는 이벤트와 함께 전달된 영속성 컨텍스트를
     * 사용하도록 되어 있습니다.
     *
     * [EntityManager]는 스레드 세이프하지 않기에 멀티 스레딩의 경우 데이터 불일치 및 동기화 문제가 발생할 수 있습니다.
     * 가령 여러 스레드가 동시에 한 엔티티 매니저를 사용할 경우, 내부적인 저장소인 영속성 컨텍스트의 캐시나 상태 추적 로직이 오염될 위험이 있습니다.
     * 따라서 보통 "하나의 트랜잭션에 하나의 엔티티 매니저"를 사용하는 것이 안전합니다.
     *
     * 그렇다고 "스레드:[EntityManager] = 1:1" 관계 역시 안전하지는 않습니다.
     * 스레드 풀을 사용하는 경우 스레드가 재사용되므로, 하나의 스레드가 재사용되며 여러 트랜잭션을 처리할 수 있기 때문입니다.
     * 만약 동일한 엔티티 매니저 인스턴스를 재사용한다면, 이전 요청 또는 이전 트랜잭션의 흔적이 남아있거나,
     * 이미 닫혀 있거나, 새 트랜잭션에는 부적합한 상태일 수 있습니다.
     *
     * 따라서 "트랜잭션:[EntityManager] = 1:1" 관계, 즉 "트랜잭션 범위로 엔티티 매니저를 새로 할당"하는 것이 가장 좋습니다.
     * 트랜잭션이 시작될 때 영속성 컨텍스트가 생성되어 엔티티 매니저와 연결되고, 트랜잭션이 끝난 후 영속성 컨텍스트를 정리하면
     * 문제가 발생될 여지가 없습니다.
     *
     * 스프링은 이를 자동으로 처리하기 위해서 엔티티 매니저를 주입할 때 실제로는 프록시로 감싼 엔티티 매니저 인스턴스를 주입합니다.
     * - [org.springframework.orm.jpa.support.PersistenceAnnotationBeanPostProcessor.PersistenceElement.getResourceToInject]
     * - [org.springframework.orm.jpa.support.PersistenceAnnotationBeanPostProcessor.PersistenceElement.resolveEntityManager]
     * - [org.springframework.orm.jpa.SharedEntityManagerCreator.createSharedEntityManager]
     *
     * 이는 '생성자로 주입([entityManagerByConstructor])'하든
     * '[PersistenceContext] 어노테이션을 통해 주입([entityManagerByAnnotation])'하든 마찬가지입니다.
     * - 생성자 주입: [org.springframework.beans.factory.support.ConstructorResolver.autowireConstructor]
     * - 어노테이션 주입: [org.springframework.orm.jpa.support.PersistenceAnnotationBeanPostProcessor]
     *
     * > 다만 스프링 문서를 보면 엔티티 매니저를 생성자에 주입하는 것보다는
     * > [PersistenceContext] 어노테이션을 통해 주입하는 예제가 대다수인 것으로 보아,
     * > 어노테이션을 통해 주입하는 것이 일반적인 방식으로 보입니다.
     *
     * 이 프록시 엔티티 매니저는 트랜잭션과 관련된 엔티티 매니저의 기능이 필요할 때 현재 스레드([ThreadLocal])에서 실행 중인 트랜잭션을 확인하고,
     * 그 트랜잭션에 연결된 실제 [EntityManager]를 찾아서 요청을 위임합니다.
     * 이 덕분에 트랜잭션 범위로 [EntityManager]가 관리되고 각 트랜잭션마다 올바른 [EntityManager]를 안전하게 사용할 수 있습니다.
     * 또한 주입 시점에 엔티티 매니저가 실제로 생성되는 것은 아니고, 필요할 때 실제 객체와 영속성 컨텍스트를 초기화하기 때문에
     * 효율적인 리소스 관리가 가능합니다.(lazy loading)
     * - [org.springframework.orm.jpa.SharedEntityManagerCreator.SharedEntityManagerInvocationHandler.invoke] 호출될 때
     *   [org.springframework.orm.jpa.EntityManagerFactoryUtils.doGetTransactionalEntityManager]를 통해
     *   현재 사용해야 하는 엔티티 매니저를 판단합니다.
     *
     * 그리고 트랜잭션이 끝나면 해당 엔티티 매니저와 영속성 컨텍스트는 자동으로 정리됩니다.
     * - [org.springframework.orm.jpa.support.PersistenceAnnotationBeanPostProcessor.postProcessBeforeDestruction] 참고
     *
     * 마찬가지로 [Transactional] 어노테이션은 트랜잭션 경계를 정의하며,
     * 트랜잭션이 시작될 때 [EntityManager]와 그 내부에 영속성 컨텍스트가 생성되고,
     * 트랜잭션이 종료되면 자동으로 정리됩니다.
     *
     * References:
     * - https://docs.spring.io/spring-data/jpa/docs/current-SNAPSHOT/reference/html/#jpa.misc.jpa-context
     * - https://docs.spring.io/spring-framework/reference/data-access/dao.html#dao-annotations
     * - [7.7. Container-managed Persistence Contexts](https://jakarta.ee/specifications/persistence/3.2/jakarta-persistence-spec-3.2#a11791)
     * - [10.5.1. PersistenceContext Annotation](https://jakarta.ee/specifications/persistence/3.2/jakarta-persistence-spec-3.2#persistencecontext-annotation)
     */
    @PersistenceContext
    private lateinit var entityManagerByAnnotation: EntityManager

    /**
     * - [org.springframework.orm.jpa.SharedEntityManagerCreator.SharedEntityManagerInvocationHandler.invoke]
     */
    fun entityManagerInfo(): Map<String, Any> {

        val method = EntityManagerProxy::class.java.getDeclaredMethod("getTargetEntityManager")

        val emByAnnotationHandler = Proxy.getInvocationHandler(entityManagerByAnnotation)
        val actualEntityManagerOfEmByAnnotation = emByAnnotationHandler.invoke(entityManagerByAnnotation, method, null)

        val emByConstructorHandler = Proxy.getInvocationHandler(entityManagerByConstructor)
        val actualEntityManagerOfEmByConstructorHandler =
            emByConstructorHandler.invoke(entityManagerByConstructor, method, null)

        return mapOf(
            "actualEntityManagerOfEmByAnnotation" to actualEntityManagerOfEmByAnnotation.hashCode(),
            "actualEntityManagerOfEmByConstructorHandler" to actualEntityManagerOfEmByConstructorHandler.hashCode(),
            "actualEntityManagerOfEmByAnnotation.equals(actualEntityManagerOfEmByConstructorHandler)" to (
                    actualEntityManagerOfEmByAnnotation == actualEntityManagerOfEmByConstructorHandler // 같은 스레드이므로 true여야 합니다.
                    ),
            "entityManagerByAnnotation" to entityManagerByAnnotation.toString(),
            "entityManagerByAnnotation.hashCode" to entityManagerByAnnotation.hashCode(),
            "entityManagerByAnnotation.isAopProxy" to AopUtils.isAopProxy(entityManagerByAnnotation),
            "entityManagerByAnnotation.isCglibProxy" to AopUtils.isCglibProxy(entityManagerByAnnotation),
            "entityManagerByAnnotation.isEntityManagerProxy" to (entityManagerByAnnotation is EntityManagerProxy),
            "entityManagerByAnnotation.isJdkDynamicProxy" to AopUtils.isJdkDynamicProxy(entityManagerByAnnotation),
            "entityManagerByAnnotation.isProxyClass" to Proxy.isProxyClass(entityManagerByAnnotation.javaClass),
            "entityManagerByAnnotation.isSharedEntityManagerCreator" to (entityManagerByAnnotation is SharedEntityManagerCreator),
            "entityManagerByConstructor" to entityManagerByConstructor.toString(),
            "entityManagerByConstructor.hashCode" to entityManagerByConstructor.hashCode(),
            "entityManagerByConstructor.isAopProxy" to AopUtils.isAopProxy(entityManagerByConstructor),
            "entityManagerByConstructor.isCglibProxy" to AopUtils.isCglibProxy(entityManagerByConstructor),
            "entityManagerByConstructor.isEntityManagerProxy" to (entityManagerByConstructor is EntityManagerProxy),
            "entityManagerByConstructor.isJdkDynamicProxy" to AopUtils.isJdkDynamicProxy(entityManagerByConstructor),
            "entityManagerByConstructor.isProxyClass" to Proxy.isProxyClass(entityManagerByConstructor.javaClass),
            "entityManagerByConstructor.isSharedEntityManagerCreator" to (entityManagerByConstructor is SharedEntityManagerCreator),
        )
    }

    fun dataSourceInfo(): Map<String, Any> {
        // Hibernate SessionFactory 가져오기
        val sessionFactory = entityManagerFactory.unwrap(SessionFactoryImplementor::class.java)

        // JdbcServices에서 ConnectionProvider 가져오기
        val connectionProvider: ConnectionProvider? =
            sessionFactory.serviceRegistry.getService(ConnectionProvider::class.java)

        val dataSource: DataSource? = connectionProvider?.unwrap(DataSource::class.java)

        if (dataSource == null) {
            return emptyMap()
        }

        val agroalDataSource = dataSource as AgroalDataSource
        val connectionPoolConfig = agroalDataSource.configuration.connectionPoolConfiguration()
        val connectionFactoryConfig = connectionPoolConfig.connectionFactoryConfiguration()

        return mapOf(
            "dataSource.config.connectionPool.acquisitionTimeout" to connectionPoolConfig.acquisitionTimeout(),
            "dataSource.config.connectionPool.connectionFactoryConfig.connectionProviderClass" to connectionFactoryConfig.connectionProviderClass(),
            "dataSource.config.connectionPool.connectionFactoryConfig.credentials" to connectionFactoryConfig.credentials(),
            "dataSource.config.connectionPool.connectionFactoryConfig.jdbcProperties" to connectionFactoryConfig.jdbcProperties(),
            "dataSource.config.connectionPool.connectionFactoryConfig.jdbcProperties" to connectionFactoryConfig.jdbcProperties(),
            "dataSource.config.connectionPool.connectionFactoryConfig.jdbcUrl" to connectionFactoryConfig.jdbcUrl(),
            "dataSource.config.connectionPool.connectionFactoryConfig.principal" to connectionFactoryConfig.principal(),
            "dataSource.config.connectionPool.idleValidationTimeout" to connectionPoolConfig.idleValidationTimeout(),
            "dataSource.config.connectionPool.initialSize" to connectionPoolConfig.initialSize(),
            "dataSource.config.connectionPool.leakTimeout" to connectionPoolConfig.leakTimeout(),
            "dataSource.config.connectionPool.maxLifetime" to connectionPoolConfig.maxLifetime(),
            "dataSource.config.connectionPool.maxSize" to connectionPoolConfig.maxSize(),
            "dataSource.config.connectionPool.minSize" to connectionPoolConfig.minSize(),
            "dataSource.config.connectionPool.reapTimeout" to connectionPoolConfig.reapTimeout(),
            "dataSource.config.connectionPool.validateOnBorrow" to connectionPoolConfig.validateOnBorrow(),
            "dataSource.config.connectionPool.validationTimeout" to connectionPoolConfig.validationTimeout(),
            "dataSource.connection.clientInfo.size" to dataSource.connection.clientInfo.size,
        )
    }

    fun getConnectionProviderName(): String {
        // Hibernate SessionFactory 가져오기
        val sessionFactory = entityManagerFactory.unwrap(SessionFactoryImplementor::class.java)

        // JdbcServices에서 ConnectionProvider 가져오기
        val connectionProvider: ConnectionProvider? =
            sessionFactory.serviceRegistry.getService(ConnectionProvider::class.java)

        return "Loaded ConnectionProvider: ${connectionProvider?.javaClass?.name}"
    }
}