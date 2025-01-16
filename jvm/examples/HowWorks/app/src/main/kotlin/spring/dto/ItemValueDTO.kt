package spring.dto

import spring.domain.ItemKey
import spring.domain.ItemValue
import java.time.Instant

data class ItemValueDTO(
    val itemValue: String,
    val created: Instant,
    val modified: Instant,
) {
    companion object {
        fun of(itemValue: ItemValue): ItemValueDTO = ItemValueDTO(
            itemValue = itemValue.itemValue,
            created = itemValue.created ?: Instant.now(),
            modified = itemValue.modified ?: Instant.now(),
        )

        fun setFrom(itemKey: ItemKey): Set<ItemValueDTO> {
            return itemKey.itemValues.map { of(it) }.toSet()
        }
    }
}
