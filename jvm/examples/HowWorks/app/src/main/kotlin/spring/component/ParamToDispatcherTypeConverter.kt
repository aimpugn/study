package spring.component

import org.springframework.core.convert.converter.Converter
import org.springframework.stereotype.Component
import spring.service.DispatcherType

@Component
class ParamToDispatcherTypeConverter : Converter<String, DispatcherType> {
    override fun convert(source: String) =
        DispatcherType.entries
            .find { it.value == source }
            ?: DispatcherType.DEFAULT
}