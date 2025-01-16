package spring.dto

import spring.domain.ItemKey

data class ItemKeyDTO(
    val itemKey: String,
    val itemValues: Set<ItemValueDTO>,
) {
    companion object {
        fun from(itemKey: ItemKey): ItemKeyDTO {
            return ItemKeyDTO(
                itemKey = itemKey.itemKey,
                itemValues = ItemValueDTO.setFrom(itemKey)
            )
        }
    }
}

