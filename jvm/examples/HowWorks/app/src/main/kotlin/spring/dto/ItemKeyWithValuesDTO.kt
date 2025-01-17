package spring.dto

import java.time.Instant

open class ItemKeyWithValuesDTO(
    val id: Long,
    val itemKey: String,
    val created: Instant,
    val modified: Instant,
    val itemValueId: Long,
    val itemValue: String,
    val itemValueCreated: Instant,
    val itemValueModified: Instant,
) {
    companion object {
        const val MAPPING_NAME = "ItemKeyWithValuesDTO"
    }
}