package spring.controller

import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController
import spring.service.DataSourceService

@RestController
@RequestMapping("/datasource")
class DataSourceController(
    val dataSourceService: DataSourceService,
) {
    @GetMapping("info")
    fun dataSourceInfo(): Map<String, Any> {
        return dataSourceService.dataSourceInfo()
    }

    @GetMapping("entity-manager")
    fun entityManagerInfo(): Map<String, Any> {
        return dataSourceService.entityManagerInfo()
    }

    @GetMapping("connection-provider")
    fun connectionProviderName(): Map<String, String> {
        return mapOf(
            "name" to dataSourceService.getConnectionProviderName()
        )
    }
}