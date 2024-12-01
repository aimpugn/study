package main.spring

data class MyTestProperty(
    var stringProperty: String = "",
    var boolProperty: Boolean = false
) {
    companion object {
        const val KEY = "my.test"
    }

    override fun toString(): String {
        return "string: $stringProperty, bool: $boolProperty"
    }
}