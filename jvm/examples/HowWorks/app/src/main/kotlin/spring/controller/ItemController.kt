package spring.controller

import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RestController
import spring.service.ItemService

@RestController
class ItemController(
    private val itemService: ItemService,
) {
    @GetMapping("items")
    fun items() = itemService.findAllItems()

    @GetMapping("items-fetch-join")
    fun itemsByFetchJoin() = itemService.findAllItemsByFetchJoin()
}