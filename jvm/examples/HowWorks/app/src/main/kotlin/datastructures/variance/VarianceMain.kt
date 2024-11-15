package datastructures.variance

open class Parent {
    override fun toString() = "Parent instance"
}

class Child : Parent() {
    override fun toString() = "Child instance"
}

class InvariantBox<T>(val value: T)


class CovariantBox<out T>(private var value: T? = null) {
    /*fun set(value: T) {} -> Type parameter T is declared as 'out' but occurs in 'in' position in type T*/

    /**
     * 하지만 가시성을 `private`으로 만들면 setter 선언이 가능합니다.
     * 객체 내부에서는 공변성을 사용하여 객체를 상위 타입으로 업캐스팅할 수 없기 때문입니다
     */
    private fun set(value: T?) {
        this.value = value
    }

    fun get(): T = value ?: error("Value not set")
}


class ContravariantBox<in T> {
    private var value: T? = null

    fun set(value: T) {
        this.value = value
        println("set done")
    }
}

/**
 * ### 변성(variance)
 *
 * 변성(Variance)은 타입 계층 구조에서 제네릭 타입 간의 관계를 정의하는 규칙을 설명하는 개념입니다.
 * 제네릭 타입이 상위 타입-하위 타입 관계를 유지하거나, 특정 방향에서 허용되는지를 정의합니다.
 *
 * - 공변성(Covariance): 하위 타입이 상위 타입으로 안전하게 대체될 수 있음.
 * - 반공변성(Contravariance): 상위 타입이 하위 타입으로 안전하게 대체될 수 있음.
 * - 불공변성(Invariance): 타입 간의 대체가 허용되지 않음.
 *
 * ### 공변성(covariant): `out` 키워드
 *
 * 서브타입 관계가 제네릭 타입에서도 유지되는 것을 의미합니다.
 * 하위 타입을 상위 타입으로 대체될 수 있는 성질입니다.
 * 공변성에서는 `out` 키워드를 사용하여 하위 타입을 상위 타입으로 안전하게 반환할 수 있습니다.
 *
 * - 하위 타입 관계를 유지되므로, 생산시 [Child] 타입의 제네릭 객체가 [Parent] 타입의 제네릭 객체로 안전하게 대체될 수 있습니다.
 * - `T`를 리턴(생산)할 때만 사용하는 경우 `out`을 사용하여 공변성을 유지할 수 있습니다.
 * - 공변성에서 중요한 점은, 타입이 리턴(생산)되는 상황에서는 상위 타입을 사용할 수 있다는 것입니다.
 *
 * 즉, covariant는 리턴할 수 있는 타입의 상한을 정해두고, 그 상한이 되는 타입 및 하위 타입 내에서 리턴할 수 있게 합니다.
 *
 * ### 반공변성(contravariant): `in` 키워드
 *
 * 반공변성은 제네릭에 `in` 키워드를 사용하여 서브타입 관계를 역전시키는 것을 의미합니다.
 * 상위 타입으로부터 값을 안전하게 소비할 수 있음을 나타냅니다.
 * `in T : Parent` 경우 [Parent]를 상속하는 어떤 타입을 입력으로 받겠다는 것을 의미합니다.
 *
 * - 하위 타입 관계가 반전되므로, 소비시 [Parent] 타입의 제네릭 객체가 더 구체적인 [Child] 타입의 객체로 대체될 수 있습니다.
 * - `T`를 소비할 때만 사용하는 경우 in을 사용하여 반공변성을 유지할 수 있습니다.
 * - 반공변성에서 중요한 점은, 타입이 소비되는 상황에서는 하위 타입을 사용할 수 있다는 것입니다.
 *
 * 즉, contravariant는 파라미터로 수용할 수 있는 타입의 상한을 정해두고, 그 상한이 되는 타입 및 하위 타입 내에서 파라미터로 받을 수 있게 합니다.
 *
 * ### 불공변성(invariant)
 *
 * 불공변성은 타입 안정성을 보장하기 위해 중요합니다.
 *
 * ```
 * class Box<T>(val value: T)
 *
 * val intBox: Box<Int> = Box(1)
 * val numBox: Box<Number> = intBox // 컴파일 에러! Box<Int>는 Box<Number>의 서브타입이 아님
 * ```
 *
 * ### Examples
 *
 * - default: [Int] 값을 [Number] 타입의 변수에 설정할 수 있습니다.
 *    ```
 *    // 공변성(covariant)
 *    val integer: Int = 100
 *    val number: Number = integer
 *    ```
 *
 * - invariance:
 *    ```
 *    class Box<T>
 *    // 아무 관계도 없습니다.
 *    class Box<Int> <-X-> class Box<Number>
 *    ```
 *
 * - covariance:
 *    ```
 *    class Box<out T>
 *    // Box<Int> 타입의 값을 Box<Number> 타입의 변수에 할당할 수 있습니다.
 *    class Box<Int> ---> class Box<Number>
 *
 *    // val a: Box<Int> = Box<Number>() Error: Type mismatch
 *    val b: Box<Number> = Box<Int>()
 *    ```
 *
 * - contravariance:
 *    ```
 *    class Box<in T>
 *    // Box<Number> 타입의 값을 Box<Int> 타입의 변수에 할당할 수 있습니다.
 *    class Box<Int> <--- class Box<Number>
 *
 *    val a: Box<Int> = Box<Number>()
 *    // val b: Box<Number> = Box<Int>() Error: Type mismatch
 *    ```
 *
 * Refernces:
 * - [제네릭의 공변성(Covariance) 및 반공변성(Contravariance)](https://learn.microsoft.com/ko-kr/dotnet/standard/generics/covariance-and-contravariance)
 * - [타입 시스템에서의 변성(Variance)](https://driip.me/d875a384-3fb9-471b-a53b-b3ca52f8238e)
 */
fun main() {
    // 공변성: 타입 계층 구조에서 하위 타입의 값이 상위 타입 변수에 할당될 수 있습니다.
    val integer: Int = 100
    val number: Number = integer
    val any: Any = number
    println("integer as any: $any")

    // val a: CovariantBox<Int> = CovariantBox<Number>() Type mismatch.
    val b: CovariantBox<Number> = CovariantBox<Int>(100) // Ok
    println("val b: CovariantBox<Number> = CovariantBox<Int>(): $b") // val b: CovariantBox<Number> = CovariantBox<Int>(): datastructures.variance.CovariantBox@69d9c55

    val c: ContravariantBox<Int> = ContravariantBox<Number>() // Ok
    // val d: ContravariantBox<Number> = ContravariantBox<Int>() Type mismatch.
    println("val c: ContravariantBox<Int> = ContravariantBox<Number>(): $c") // val c: ContravariantBox<Int> = ContravariantBox<Number>(): datastructures.variance.ContravariantBox@7ca48474

    val e: InvariantBox<Int> = InvariantBox(100)
    /*
    val f: InvariantBox<Number> = e // Type mismatch.
    */
    println("e.value: ${e.value}") // e.value: 100

    open class Animal {
        private val name = "animal"
        open fun feed() {
            println("Feeding animal")
        }

        override fun toString(): String {
            return name
        }
    }

    open class Dog : Animal() {
        private val name = "dog"
        override fun feed() {
            println("Feeding dog")
        }

        override fun toString(): String {
            return name
        }
    }

    class Puppy : Dog() {
        private val name = "puppy"
        override fun feed() {
            println("Feeding Puppy")
        }

        override fun toString(): String {
            return name
        }
    }

    class Hound : Dog() {
        private val name = "hound"
        override fun feed() {
            println("Feeding Hound")
        }

        override fun toString(): String {
            return name
        }
    }

    /**
     * 서브타이핑(subtyping) 개념에 기반한 암시적 업캐스팅(implicit upcasting)
     */
    fun takeDog(dog: Dog) {}

    takeDog(Dog())
    takeDog(Puppy()) // implicit upcasting
    takeDog(Hound()) // implicit upcasting

    val puppyBox = CovariantBox<Puppy>(Puppy())
    val dogBox: CovariantBox<Dog> = puppyBox
    print("Covariant: ")
    dogBox.get().feed()

    val dogHouse = CovariantBox<Dog>(Dog())
    val anyBox: CovariantBox<Any> = dogHouse
    println("Covariant: anyBox act only as Any type ${anyBox.get()}")

    /**
     * 주의할 점은 `out T: Dog`로 공변성을 설정했다고 해서,
     * 클래스 내부에서 반드시 모든 반환 타입이 [T]와 관련되어야 한다는 의미는 아닙니다.
     * 즉, 공변성이든 반공변성이든 타입 파라미터 [T]와 직접적으로 관련된 메서드에만 적용됩니다.
     */
    class DogOrPuppyBox<out T : Dog> {
        fun getDog(): Dog = Dog()
        fun getPuppy(): Puppy = Puppy()
        fun getAny(): Any = Any()
        fun getRandomInt(): Int = (1..100).random()
    }

    val dogOrPuppy = DogOrPuppyBox<Dog>()
    println("dogOrPuppy.getDog() returns ${dogOrPuppy.getDog()}") // dogOrPuppy.getDog() returns dog
    println("dogOrPuppy.getPuppy() returns ${dogOrPuppy.getPuppy()}") // dogOrPuppy.getPuppy() returns puppy
    println("dogOrPuppy.getAny() returns ${dogOrPuppy.getAny()}") // dogOrPuppy.getPuppy() returns java.lang.Object@7946e1f4
    println("dogOrPuppy.getInt() returns ${dogOrPuppy.getRandomInt()}") // dogOrPuppy.getInt() returns 66

    /**
     * 반공변성
     */
    val handleDog = fun(processor: (Animal) -> Unit) {
        val dog = Dog()
        // 하위 타입인 Dog를 파라미터로 받는 함수가 필요할 때,
        // 상위 타입 Animal을 파라미터로 받는 함수를 사용할 수 있습니다.
        processor(dog) // Dog 타입이 Animal 타입으로 암시적으로 업캐스팅
    }
    val animalProcessor: (Animal) -> Unit = { animal -> animal.feed() }
    handleDog(animalProcessor) // Feeding dog

    /**
     * 공변성
     */
    fun provideAnimal(provider: () -> Animal) {
        // Puppy 타입이 Animal 타입으로 암시적으로 업캐스팅
        val animal: Animal = provider()
        animal.feed()
    }

    // 상위 타입 Animal을 리턴해야 하는 함수가 필요할 때,
    // 하위 타입인 Puppy를 리턴하는 함수를 사용할 수 있습니다.
    val dogProvider: () -> Puppy = { Puppy() }
    provideAnimal(dogProvider) // Feeding Puppy

    class DogHandler<in T : Dog> {
        fun feedDog(dog: T) {
            dog.feed()
        }
    }

    DogHandler<Dog>().feedDog(Puppy()) // Feeding Puppy
    // DogHandler<Dog>().feedDog(Animal()) // Type mismatch.

    // 즉, covariant는 리턴할 수 있는 최대치를 열어두고, 그 하위 타입 내에서 리턴할 수 있게 하는 것으로 보이고
    val covariantPuppyBox = CovariantBox<Puppy>(Puppy())
    val covariantDogBox: CovariantBox<Dog> = covariantPuppyBox
    val covariantAnimalBox: CovariantBox<Animal> = covariantDogBox
    covariantAnimalBox.get().feed() // Feeding Puppy

    // contravariant는 수용할 수 있는 최대치를 열어두고, 그 하위 타입 내에서 파라미터로 받는 것으로 보입니다.
    val contravariantAnimalBox = ContravariantBox<Animal>()
    val contravariantDogBox: ContravariantBox<Dog> = contravariantAnimalBox
    val contravariantPuppyBox: ContravariantBox<Puppy> = contravariantDogBox
    contravariantPuppyBox.set(Puppy()) // set done

    val appendDogs = fun(list: MutableList<in Dog>) { // Dog 이하 클래스의 인스턴스들 추가 가능합니다.
        // list.add(Animal()) // Type mismatch.
        list.add(Puppy())
        list.add(Hound())
        list.add(Dog())
    }
    val animalList = mutableListOf(Animal())
    appendDogs(animalList)
    println(animalList) // [animal, puppy, hound, dog]

    /**
     * ```
     * (T1, T2) -> T
     *   ╲  ╱      └ out (covariant)
     *    in
     * (contravariant)
     * ```
     *
     * 코틀린 함수에서
     * - 모든 파라미터 타입은 반공변적(contravariant)입니다.
     *   반공변적(contravariant)이란 상위 타입을 처리하는 함수를 하위 타입을 처리하는 함수로 사용할 수 있는 성질을 의미합니다.
     *
     *   ex: `(Animal) -> Unit` 타입의 함수를 `(Dog) -> Unit` 타입이 필요한 곳에 사용 가능
     *
     *   ```
     *   fun handleDog(processor: (Animal) -> Unit) {
     *       val dog = Dog()
     *       // 상위 타입 Animal을 처리하는 함수를 파라미터로 받아서,
     *       // 하위 타입인 Dog를 처리하는 함수로 사용할 수 있습니다.
     *       processor(dog)
     *   }
     *   ```
     *
     * - 모든 리턴 타입은 공변적(covariant)입니다.
     *   공변적(covariance)이란 하위 타입을 상위 타입으로 안전하게 대체할 수 있는 성질을 의미합니다.
     *   즉, 가령 함수 시그니처에서 리턴 타입을 [Any], [Exception]으로 하고,
     *   실제 리턴 시에는 더 구체적인 [String], [IllegalArgumentException] 등으로 리턴할 수 있습니다.
     *
     *   ex: `() -> Dog` 타입의 함수를 `() -> Animal` 타입이 필요한 곳에 사용 가능
     *
     *   ```
     *   fun provideAnimal(provider: () -> Animal) {
     *       val animal = provider()
     *       println(animal)
     *   }
     *
     *   val dogProvider: () -> Dog = { Dog() }
     *   provideAnimal(dogProvider) // () -> Dog는 () -> Animal로 사용 가능
     *   ```
     *
     * 함수 타입 `(A) -> B`가 있고, `(C) -> D`가 있다고 할 때, 이 둘이 호환되려면:
     * - 파라미터 타입 C는 A의 슈퍼타입이어야 합니다 (반공변성).
     * - 리턴 타입 B는 D의 서브타입이어야 합니다 (공변성).
     *
     * 아래 다이어그램에서, 다이어그램에서 화살표는 호환 가능성을 나타냅니다.
     * 이 호환 가능성은 다음 반공변성과 공변성에 따라 결정됩니다.
     * - '파라미터 타입'은 넓어지고(상위 타입으로 확장 가능): 반공변성
     * - '반환 타입'은 구체화됩니다(하위 타입으로 좁아질 수 있음): 공변성
     *
     * ```
     * 파라미터 타입 하위             (Int) -> Any                     리턴 타입 상위
     *        ↑                   ↗         ↖                        ↑
     *        │                 ╱             ╲                      │
     *        │        (Int) -> Number      (Number) -> Any          │
     *        │                ↖                ↗                    │
     *        │                  ╲            ╱                      │
     *        │                (Number) -> Number                    │
     *        │                   ↗          ↖                       │
     *        ↓                 ╱              ╲                     ↓
     * 파라미터 타입 하위     (Any) -> Number      (Number) -> Int     리턴 타입 하위
     *                     │                               │
     *                     │                               └ 파라미터 타입 `Number`는 동일합니다.
     *                     │                                 리턴 타입 `Int`는 `Number`의 서브타입입니다. (공변성)
     *                     │
     *                     └ 파라미터 타입 `Any`는 `Number`의 슈퍼타입입니다. (반공변성)
     *                       리턴 타입 `Number`는 동일합니다.
     *  ```
     *
     * 다이어그램에서 위에서 아래로 내려갈수록 '파라미터 타입'은 더 상위 타입으로 이동합니다.
     * '파라미터 타입'은 더 넓은 타입으로 변할 수 있으며, 이는 반공변성(contravariant)을 나타냅니다.
     * 반공변성이란, '파라미터 타입'이 더 상위 타입으로 이동할 수 있다는 의미입니다.
     * ex: `(Int) -> Number`에서 `(Number) -> Number` 경우 [Int]에서 [Number]로 올라갑니다.
     *
     * 다이어그램에서 위에서 아래로 내려갈수록 '반환 타입'은 더 하위 타입으로 이동합니다.
     * '반환 타입'은 더 구체적인 타입으로 변할 수 있으며, 이는 공변성(covariance)을 나타냅니다.
     * 공변성이란, '반환 타입'이 더 하위 타입으로 변할 수 있다는 의미입니다.
     * ex: `(Int) -> Any`에서 `(Int) -> Number` 경우 [Any]에서 [Number]로 더 구체화됩니다.
     *
     * - 파라미터 타입으로 들어갈 수 있는 타입들: [Any], [Number], [Int]. 즉, [Int]를 포함한 그 슈퍼 타입들.
     * - 리턴 타입으로 들어갈 수 있는 타입들: [Int], [Number], [Any]. 즉, [Any]를 포함한 그 하위 타입들.
     */
    val printProcessedNumber = fun(transition: (Int) -> Any) {
        println(transition(42))
    }

    val intToDouble: (Int) -> Number = { it.toDouble() }
    val numberAsText: (Number) -> Any = { it.toShort() }
    val identity: (Number) -> Number = { it }
    val numberToInt: (Number) -> Int = { it.toInt() }
    val numberHash: (Any) -> Number = { it.hashCode() }
    val stringHash: (Int) -> String = { "converted: $it" }
    val anyToInt: (Any) -> Int = { 100 }
    printProcessedNumber(intToDouble) // 42.0
    printProcessedNumber(numberAsText) // 42
    printProcessedNumber(identity) // 42
    printProcessedNumber(numberToInt) // 42
    printProcessedNumber(numberHash) // 42
    printProcessedNumber(stringHash) // converted: 42
    printProcessedNumber(anyToInt) // 100
}
