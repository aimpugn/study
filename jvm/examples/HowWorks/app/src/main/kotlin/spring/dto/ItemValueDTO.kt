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

        fun of(itemValue: String, created: Instant, modified: Instant): ItemValueDTO = ItemValueDTO(
            itemValue = itemValue,
            created = created,
            modified = modified,
        )

        fun setFrom(itemKey: ItemKey): Set<ItemValueDTO> {
            return itemKey.itemValues.map { of(it) }.toSet()
        }
    }
}
