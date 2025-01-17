package spring.service

import org.springframework.aop.framework.AopProxyUtils
import org.springframework.aop.support.AopUtils
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import spring.dto.ItemKeyDTO
import spring.dto.ItemValueDTO
import spring.repository.ItemRepository
import spring.repository.ItemRepositoryCustom
import spring.repository.ItemRepositoryImpl
import java.lang.reflect.InvocationHandler
import java.lang.reflect.Proxy

@Service
class ItemService(
    private val itemRepository: ItemRepository,
) {
    @Transactional(readOnly = true)
    fun findAllItems(): List<ItemKeyDTO> {
        return itemRepository.findAll().map { ItemKeyDTO.from(it) }
    }

    @Transactional(readOnly = true)
    fun findAllItemsByFetchJoin(): List<ItemKeyDTO> {
        return itemRepository.findAllFetchJoin().map { ItemKeyDTO.from(it) }
    }

    @Transactional(readOnly = true)
    fun findAllItemsUsingEntityManager(): List<ItemKeyDTO> {
        println("itemRepository isProxyClass: ${Proxy.isProxyClass(itemRepository::class.java)}")
        println("itemRepository isAopProxy: ${AopUtils.isAopProxy(itemRepository)}")
        println("itemRepository isCglibProxy: ${AopUtils.isCglibProxy(itemRepository)}")
        println("itemRepository isJdkDynamicProxy: ${AopUtils.isJdkDynamicProxy(itemRepository)}")
        println("itemRepository is ItemRepositoryImpl: ${itemRepository is ItemRepositoryImpl}")
        // itemRepository is proxy: true
        // itemRepository isAopProxy: true
        // itemRepository isCglibProxy: false (상속이 아닌 인터페이스 구현 방식이므로)
        // itemRepository isJdkDynamicProxy: true (인터페이스 구현 방식이므로)
        // itemRepository is ItemRepositoryImpl: false (프록시이므로)

        val actualItemRepository = AopProxyUtils.getSingletonTarget(itemRepository)
        println("itemRepository's actualItemRepository: $actualItemRepository")
        println(
            "itemRepository's proxiedUserInterfaces: ${
                AopProxyUtils.proxiedUserInterfaces(itemRepository).contentDeepToString()
            }"
        )
        println("itemRepository is ItemRepository: ${actualItemRepository is ItemRepository}")
        println("itemRepository is ItemRepositoryCustom: ${actualItemRepository is ItemRepositoryCustom}")
        println("itemRepository is ItemRepositoryImpl: ${actualItemRepository is ItemRepositoryImpl}")
        // itemRepository's actualItemRepository: org.springframework.data.jpa.repository.support.SimpleJpaRepository@53089ea3
        // itemRepository's proxiedUserInterfaces: [interface spring.repository.ItemRepository, interface org.springframework.data.repository.Repository]
        // itemRepository is ItemRepository: false (실제 타겟 인스턴스는 SimpleJpaRepository 이므로)
        // itemRepository is ItemRepositoryCustom: false (실제 타겟 인스턴스는 SimpleJpaRepository 이므로)
        // itemRepository is ItemRepositoryImpl: false (실제 타겟 인스턴스는 SimpleJpaRepository 이므로)

        // is 연산자는 객체의 실제 타입과 비교합니다.
        // JDK 동적 프록시가 생성된 경우, 원본 객체의 구체적인 타입(ItemRepositoryImpl 등)을 알지 못하므로 false를 반환합니다.
        // 프록시가 어떤 인터페이스를 구현했는지 확인하려면 `isAssignableFrom`을 사용합니다.
        println("itemRepository implements ItemRepository: ${ItemRepository::class.java.isAssignableFrom(itemRepository::class.java)}")
        println(
            "itemRepository implements ItemRepositoryCustom: ${
                ItemRepositoryCustom::class.java.isAssignableFrom(
                    itemRepository::class.java
                )
            }"
        )
        println(
            "itemRepository implements ItemRepositoryImpl: ${
                ItemRepositoryImpl::class.java.isAssignableFrom(
                    itemRepository::class.java
                )
            }"
        )
        println(
            "Actual itemRepository implements ItemRepository: ${
                ItemRepository::class.java.isAssignableFrom(
                    actualItemRepository::class.java
                )
            }"
        )
        // itemRepository implements ItemRepository: true (프록시는 ItemRepository 인터페이스 구현)
        // itemRepository implements ItemRepositoryCustom: true (프록시는 ItemRepositoryCustom 인터페이스 구현)
        // itemRepository implements ItemRepositoryImpl: false (인터페이스가 아니므로 프록시는 이를 구현하지 않음)
        // Actual itemRepository implements ItemRepository: false (실제 타겟 인스턴스의 타입은 SimpleJpaRepository 이므로)

        val itemRepositoryInvocationHandler = Proxy.getInvocationHandler(itemRepository)
        println("itemRepositoryInvocationHandler is InvocationHandler: ${itemRepositoryInvocationHandler is InvocationHandler}")
        println(
            "proxiedUserInterfaces: ${
                AopProxyUtils.proxiedUserInterfaces(itemRepositoryInvocationHandler).contentDeepToString()
            }"
        )
        // itemRepositoryInvocationHandler is InvocationHandler: true
        // proxiedUserInterfaces: [interface org.springframework.aop.framework.AopProxy, interface java.lang.reflect.InvocationHandler, interface java.io.Serializable]

        return itemRepository.findAllByEntityManager().map { ItemKeyDTO.from(it) }
    }

    @Transactional(readOnly = true)
    fun findAllItemsByNativeQuery(): List<ItemKeyDTO> {
        return itemRepository.findAllItemsByNativeQuery()
            .groupBy { dto -> dto.id }
            .mapValues { (_, groupedDTOs) ->
                val firstDTO = groupedDTOs.first()
                ItemKeyDTO.of(firstDTO.itemKey).apply {
                    groupedDTOs.forEach { dto ->
                        (itemValues as MutableSet).add(
                            ItemValueDTO.of(
                                itemValue = dto.itemValue,
                                created = dto.itemValueCreated,
                                modified = dto.itemValueModified,
                            )
                        )
                    }
                }
            }.values.toList()
    }
}