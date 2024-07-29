package visibility

open class ProtectedClass {
    protected val protectedProperty: String = "protected"
}

class PublicClass : ProtectedClass() {
    fun publicFunction() {
        val protectedInstance = ProtectedClass()
        println("Protected Property: ${this.protectedProperty}")
        println("Protected Property: ${protectedInstance}")
        // Note: Accessing protected property from another instance is not allowed in Kotlin
        // println("Protected Property from instance: ${protectedInstance.protectedProperty}")
    }
}
