package spring

import org.springframework.boot.context.properties.ConfigurationProperties

/**
 * [`spring-boot-configuration-processor`](https://mvnrepository.com/artifact/org.springframework.boot/spring-boot-configuration-processor)를 사용하면
 * 생성자 바인딩(`@ConstructorBinding`)을 사용하고 모든 속성을 `var` 아닌 `val`로 불변으로 만들 수 있습니다.
 */
@ConfigurationProperties(prefix = "my.custom")
data class MyCustomProperty(
    var clientId: String = "",
    var isAdmin: Boolean = false, // Boolean 타입 경우 `is` prefix가 붙습니다.
    var applicationType: ApplicationType = ApplicationType.NONE
) {
    override fun toString(): String {
        return "clientId: $clientId, isAdmin: $isAdmin, applicationType: $applicationType"
    }
}

enum class ApplicationType {
    NONE,
    WEB
}