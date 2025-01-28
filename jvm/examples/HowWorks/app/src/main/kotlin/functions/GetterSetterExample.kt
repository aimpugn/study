package functions

import util.RunExample
import java.time.Instant
import java.time.LocalDateTime
import java.time.ZoneId
import java.util.*

interface Person {
    /**
     * 인터페이스에서 `val`로 선언된 프로퍼티는 반드시 읽을 수 있는 `getter`를 제공해야 한다는 의미일 뿐입니다.
     * 따라서 이를 반드시 `val`로만 구현해야 한다는 제한은 없습니다.
     *
     * 하지만 `var`로 선언된 프로퍼티를 `val`로 오버라이드할 수는 없습니다:
     *
     *    Val-property cannot override var-property
     */
    val name: String?
    val age: Int
    val gender: String
    // 인터페이스 내부에서만 접근 가능한 private 프로퍼티는 불가능합니다.
    // private val internalCode: String -> Abstract property in an interface cannot be private

    // 하지만 private 함수로 기본 구현 제공은 가능합니다.
    private fun info(): String = "Name: $name, Age: $age, Gender: $gender"

    fun printInfo() {
        println(info())
    }
}

class Asian(override val gender: String) : Person {
    /**
     * 프로퍼티는 클래스 속성처럼 보이지만, 실제로는 접근자(accessor)로 이루어져 있습니다.
     * Kotlin에서 프로퍼티는 단순히 값을 저장하는 필드를 가리키는 것이 아니라, 값을 가져오거나 설정하기 위한 함수의 집합으로 동작합니다.
     * 이 때문에 프로퍼티는 꼭 필드를 필요로 하지 않으며, 접근자만으로 구현될 수 있습니다.
     *
     * 프로퍼티는 필드 없이도 동작할 수 있기 때문에, Java와 달리 Kotlin은 인터페이스에 프로퍼티를 정의할 수 있습니다.
     * 프로퍼티가 단순한 필드가 아니라 접근자 메서드로 구성되기 때문입니다.
     */
    override var name: String? = null
        /**
         * 여기서 `field`를 백킹 필드(backing field)라고 합니다.
         * 프로퍼티의 값을 실제로 저장하는 내부 저장소를 의미합니다.
         * 프로퍼티를 정의할 때 기본적으로 생성되는 `getter`와 `setter`는 이 백킹 필드를 통해 값을 읽고 쓰게 됩니다.
         * `field`는 백킹 필드에 접근하기 위한 예약어로, 커스텀 `getter`와 `setter` 내부에서만 사용됩니다.
         *
         */
        get() = field?.uppercase()
        set(value) {
            if (!value.isNullOrBlank()) {
                field = value
            }
        }

    override val age: Int
        get() = (1..100).random()

    /**
     * 백킹 필드 없이도 프로퍼티를 정의할 수 있습니다.
     * 백킹 필드 없이 구현된 `getter`와 `setter`는 프로퍼티 값을 저장할 내부 공간이 없으므로,
     * 해당 프로퍼티는 읽거나 쓸 때마다 다른 로직을 통해 값을 계산하거나 외부 데이터(ex: [name])를 참조하는 경우에 사용됩니다.
     *
     * [length] 프로퍼티는 [name]의 길이를 반환하지만 자체적으로 값을 저장하지는 않습니다.
     * 대신 호출될 때마다 특정 계산을 통해 결과를 반환하게 됩니다.
     */
    val length: Int
        get() = name?.length ?: 0

    private var millis: Long = System.currentTimeMillis()  // millis 변수를 초기화

    /**
     * 파생 프로퍼티(derived property)는 실제 데이터를 저장하지 않고 특정 로직에 의해 다른 값으로 변환되는 프로퍼티를 의미합니다.
     * 예를 들어, [date] 프로퍼티는 [millis]라는 별도의 프로퍼티를 이용하여 값을 계산해 반환하는 파생 프로퍼티입니다.
     * 이렇게 파생 프로퍼티를 만들면 프로퍼티 자체는 데이터를 직접 저장하지 않으면서도 특정 형식이나 타입의 데이터를 제공할 수 있습니다.
     *
     * [date]와 같은 프로퍼티를 만들 때 직접적으로 값을 저장하지 않고 별도의 [millis] 프로퍼티를 이용해 래핑 및 언래핑을 수행하는 방식입니다.
     * 이를 통해 다음과 같은 이점이 있습니다.
     *
     * - 캡슐화로 인해 구현이 숨겨짐:
     *   [date] 프로퍼티를 통해 외부에서 [Date] 타입으로 접근할 수 있지만,
     *   실제 값은 [millis]라는 [Long] 타입의 숫자로 저장됩니다.
     *   이를 통해 내부 구현을 숨길 수 있습니다.
     *
     * - 유연한 구현 변경:
     *   예를 들어 [Date] 대신 다른 타입(예: [LocalDateTime])을 사용해야 하는 상황이 생길 때,
     *   [millis] 속성을 사용해서 [localDateTime] 속성을 제공할 수 있습니다.
     */
    var date: Date
        get() = Date(millis)
        set(value) {
            millis = value.time
        }

    val localDateTime: LocalDateTime
        get() = LocalDateTime.ofInstant(Instant.ofEpochMilli(millis), ZoneId.systemDefault())
}

@RunExample
fun gettersAndSetters() {
    val asian = Asian("male")
    println("Before: ${asian.name}(length: ${asian.length})") // Before: null(length: 0)
    asian.name = "new name"
    println("After: ${asian.name}(length: ${asian.length})") // After: NEW NAME(length: 8)
    println("Date: ${asian.date}") // Date: Thu Nov 14 17:38:08 KST 2024
    println("LocalDateTime: ${asian.localDateTime}") // LocalDateTime: 2024-11-14T17:38:08.461
    asian.printInfo()
}