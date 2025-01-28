package functions

import util.RunExample

/**
 * 제네릭은 컴파일시에 지워집니다.
 * JVM은 제네릭 정보를 런타임에 유지하지 않으며, 컴파일 시점에 제네릭 타입을 제거하고 필요시 `Object` 타입으로 변환합니다.
 * -  `List<T>` -> `List`
 * - `T : Comparable<T>` -> `Comparable`
 *
 * ```
 * @NotNull
 * public static final List quickSort(@NotNull List $this$quickSort) {
 *    Intrinsics.checkNotNullParameter($this$quickSort, "<this>");
 *    if ($this$quickSort.size() < 2) {
 *       return $this$quickSort;
 *    } else {
 *       Comparable pivot = (Comparable)CollectionsKt.first($this$quickSort);
 *       Iterable $this$partition$iv = (Iterable)CollectionsKt.drop((Iterable)$this$quickSort, 1);
 *       int $i$f$partition = 0;
 *       ArrayList first$iv = new ArrayList();
 *       ArrayList second$iv = new ArrayList();
 *
 *       for(Object element$iv : $this$partition$iv) {
 *          Comparable it = (Comparable)element$iv;
 *          int var10 = 0;
 *          if (it.compareTo(pivot) < 0) {
 *             first$iv.add(element$iv);
 *          } else {
 *             second$iv.add(element$iv);
 *          }
 *       }
 *
 *       Pair var2 = new Pair(first$iv, second$iv);
 *       List smaller = (List)var2.component1();
 *       List bigger = (List)var2.component2();
 *       return CollectionsKt.plus((Collection)CollectionsKt.plus((Collection)quickSort(smaller), pivot), (Iterable)quickSort(bigger));
 *    }
 * }
 * ```
 *
 * 참고로 JVM에서 `$` 문자는 컴파일러가 생성한 내부 요소들을 구분하기 위한 특별한 문자로 사용됩니다.
 * 컴파일러는 변수명, 메서드명, 클래스명에 `$`를 붙여 내부적으로 사용하는 임시 변수나 클래스, 익명 클래스, 람다 표현식 등을 구분합니다.
 *
 * Kotlin의 로컬 함수나 Java의 지역 클래스와 같이 스코프 내부에 존재하는 함수 및 클래스에도 `$`가 포함된 이름이 부여되기도 합니다.
 */
fun <T : Comparable<T>> List<T>.quickSort(): List<T> {
    if (this.size < 2) return this

    val pivot = this.first()
    val (smaller, bigger) = this.drop(1).partition { it < pivot }

    return smaller.quickSort() + pivot + bigger.quickSort()
}

class Node(private val name: String) {
    fun makeChild(childName: String) =
        create("$name.$childName")
            .apply {
                println("Parent's name: $name OR ${this@Node.name}")
                println("Child's name ${this?.name}")
            }
            .also {
                println("In also scope, parent's name: $name OR ${this@Node.name}")
                println("In also scope, child's name: ${it?.name}")
            }

    private fun create(name: String): Node? = if (name.isNotBlank()) Node(name) else null
}

@RunExample
fun explicitReceiversWithLabel() {
    println(listOf("B", "X", "A", "D", "C").quickSort())

    val parent = Node("parent")
    parent.makeChild("child")
}
