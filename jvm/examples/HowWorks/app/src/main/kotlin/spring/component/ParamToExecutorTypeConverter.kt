package spring.component

import org.springframework.core.convert.converter.Converter
import org.springframework.stereotype.Component
import spring.service.ExecutorType
import spring.service.ExecutorType.FIXED_THREAD_POOL

@Component
class ParamToExecutorTypeConverter : Converter<String, ExecutorType> {
    override fun convert(source: String) =
        ExecutorType.entries
            .find { it.value == source }
            ?: FIXED_THREAD_POOL
}