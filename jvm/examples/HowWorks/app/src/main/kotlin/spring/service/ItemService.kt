package spring.service

import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import spring.dto.ItemKeyDTO
import spring.repository.ItemRepository

@Service
class ItemService(
    private val itemRepository: ItemRepository,
) {
    @Transactional
    fun findAllItems(): List<ItemKeyDTO> {
        return itemRepository.findAll().map { ItemKeyDTO.from(it) }
    }

    @Transactional
    fun findAllItemsByFetchJoin(): List<ItemKeyDTO> {
        return itemRepository.findAllFetchJoin().map { ItemKeyDTO.from(it) }
    }
}