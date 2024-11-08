package classes


/**
 * `synthetic constructor`란 컴파일러가 특별한 목적을 가지고 자동으로 만드는 생성자입니다.
 * Kotlin 경우 Java와의 상호 운용성을 위해 다음과 같은 경우 `synthetic constructor`를 생성합니다:
 * - `data class`
 * - `companion object`
 * - 기본 매개변수(default parameter)를 포함한 생성자들
 *
 * 가령, Java 경우에는 기본값을 처리할 수 없기 때문에, Kotlin 컴파일러가 추가적인 `synthetic constructor`를 생성하여
 * 기본값이 있는 매개변수들을 Java에서도 지원할 수 있게 합니다.
 *
 * 이때 사용되는 [kotlin.jvm.internal.DefaultConstructorMarker]는 컴파일러가 자동으로 생성하는 `synthetic constructor`의 인자로 전달되는데,
 * JVM에서 `synthetic constructor`가 일반 생성자와 충돌하지 않도록 구분자 역할을 하며, 항상 `null`로 설정됩니다.
 * `bit mask`와 함께 사용되며, 기본값 매개변수를 다루는 `synthetic constructor`임을 식별하는 데 필요합니다.
 *
 * `bit mask`는 매개변수의 기본값 사용 여부를 나타내는 플래그로, 각 비트는 해당 매개변수가 기본값을 사용할지 여부를 나타냅니다.
 * 생성자의 파라미터가 두 개인 경우 각 비트마스크는 다음과 같은 의미입니다:
 * - `0b00`: 모든 파라미터가 명시적으로 제공됨
 * - `0b01`: 첫 번째 파라미터가 기본값 사용
 * - `0b10`: 두 번째 파라미터가 기본값 사용
 * - `0b11`: 모든 파라미터가 기본값 사용
 *
 * 기본값을 처리하기 위한 비트 마스크와 함께 사용되며, JVM 바이트코드에서 다른 생성자와 충돌하지 않게 합니다.
 * 컴파일러가 기본값 지원을 위한 `synthetic constructor`임을 식별하는 데 필요합니다.
 *
 * 예를 들어, 아래 바이트코드는 [SyntheticConstructorExample] 코드를 `kotlinc`로 컴파일하여 얻은 class 파일을
 * [Bytecode Viewer 플러그인](https://blog.jetbrains.com/ko/2020/04/10/java-bytecode-decompiler-ko/)으로 열어본 결과입니다.
 *
 * ```
 * // class version 52.0 (52)
 * // access flags 0x31
 * public final class classes/SyntheticConstructorExample {
 *
 *   // compiled from: SyntheticConstructorExample.kt
 *
 *   @Lkotlin/Metadata;(mv={2, 0, 0}, k=1, xi=48, d1={"\u0000\u0018\n\u0002\u0018\u0002\n\u0002\u0010\u0000\n\u0000\n\u0002\u0010\u0008\n\u0000\n\u0002\u0010\u000e\n\u0002\u0008\u0006\u0018\u00002\u00020\u0001B\u001b\u0012\u0008\u0008\u0002\u0010\u0002\u001a\u00020\u0003\u0012\u0008\u0008\u0002\u0010\u0004\u001a\u00020\u0005\u00a2\u0006\u0004\u0008\u0006\u0010\u0007R\u0011\u0010\u0002\u001a\u00020\u0003\u00a2\u0006\u0008\n\u0000\u001a\u0004\u0008\u0008\u0010\u0009R\u0011\u0010\u0004\u001a\u00020\u0005\u00a2\u0006\u0008\n\u0000\u001a\u0004\u0008\n\u0010\u000b"}, d2={"Lclasses/SyntheticConstructorExample;", "", "a", "", "b", "", "<init>", "(ILjava/lang/String;)V", "getA", "()I", "getB", "()Ljava/lang/String;"})
 *
 *   // access flags 0x12
 *   private final I a
 *
 *   // access flags 0x12
 *   private final Ljava/lang/String; b
 *   @Lorg/jetbrains/annotations/NotNull;() // invisible
 *
 *   // access flags 0x1
 *   public <init>(ILjava/lang/String;)V    ; 일반적인 생성자입니다. 이 경우에는 기본값을 지원하지 않으며, 인자를 명시적으로 전달해야 합니다.
 *     // annotable parameter count: 2 (invisible)
 *     @Lorg/jetbrains/annotations/NotNull;() // invisible, parameter 1
 *    L0
 *     ALOAD 2
 *     LDC "b"
 *     INVOKESTATIC kotlin/jvm/internal/Intrinsics.checkNotNullParameter (Ljava/lang/Object;Ljava/lang/String;)V
 *    L1
 *     LINENUMBER 143 L1
 *     ALOAD 0
 *     INVOKESPECIAL java/lang/Object.<init> ()V
 *     ALOAD 0
 *     ILOAD 1
 *     PUTFIELD classes/SyntheticConstructorExample.a : I
 *     ALOAD 0
 *     ALOAD 2
 *     PUTFIELD classes/SyntheticConstructorExample.b : Ljava/lang/String;
 *     RETURN
 *    L2
 *     LOCALVARIABLE this Lclasses/SyntheticConstructorExample; L0 L2 0
 *     LOCALVARIABLE a I L0 L2 1
 *     LOCALVARIABLE b Ljava/lang/String; L0 L2 2
 *     MAXSTACK = 2
 *     MAXLOCALS = 3
 *
 *   // access flags 0x1001
 *   public synthetic <init>(ILjava/lang/String;ILkotlin/jvm/internal/DefaultConstructorMarker;)V
 *                           ^ ^                ^   ^
 *                   (slot 1)a |                |   DefaultConstructorMarker 클래스(slot 4)
 *                             b(slot 2)        bit mask(slot 3)
 *    ; SyntheticConstructorExample(10) 경우 실제로 호출되는 것은
 *    ; SyntheticConstructorExample(10, "", 0b10, null) 입니다.
 *    ; 두번째 파라미터가 기본값을 사용하므로, 비트마스크는 0b10 입니다.
 *    L0
 *     LINENUMBER 143 L0
 *     ILOAD 3          ; 3번째 로컬 변수 슬롯의 값, 즉 bit mask 값을 스택에 로드합니다. 스택: [0b10]
 *     ICONST_1         ; 첫 번째 비트를 나타내는 값 정수 1(0b01)을 스택에 푸시합니다. 스택: [0b10, 0b01]
 *     IAND             ; IAND로 `0b1 & 0b01` 연산하여, 첫 번째 비트가 0인지 확인합니다. 스택: [0b10 & 0b01]
 *     IFEQ L1          ; IAND 결과 0이면 전달된 값을 사용한다는 의미이므로, L1으로 점프합니다. 스택: []
 *     ICONST_0         ; 기본값 정수 0을 스택에 푸시합니다.
 *     ISTORE 1         ; 기본값을 a 파라미터(slot 1)에 저장합니다.
 *    L1
 *    FRAME SAME
 *     ILOAD 3          ; 3번째 로컬 변수 슬롯의 값, 즉 bit mask 값을 스택에 로드합니다. 스택: [bitmask]
 *     ICONST_2         ; 두 번째 비트를 나타내는 값 정수 2(0b10)를 스택에 푸시합니다. 스택: [bitmask, 2]
 *     IAND             ; IAND로 bit mask와 1을 비트 AND 연산하여, 두 번째 비트가 0인지 확인합니다. 스택: [bitmask & 2]
 *     IFEQ L2          ; IAND 결과 0이면 전달된 값을 사용한다는 의미이므로, L2으로 점프합니다. 스택: []
 *     LDC ""           ; 기본값 빈 문자열("")을 스택에 푸시합니다.
 *     ASTORE 2         ; 기본값을 b 파라미터(slot 2)에 저장합니다.
 *    L2
 *    FRAME SAME
 *     ALOAD 0          ; this 객체를 스택에 로드합니다.
 *     ILOAD 1          ; a 매개변수의 값 10을 스택에 로드합니다.
 *     ALOAD 2          ; b 매개변수의 값 빈 문자열("")을 스택에 로드합니다.
 *     INVOKESPECIAL classes/SyntheticConstructorExample.<init> (ILjava/lang/String;)V ; 실제 생성자를 호출합니다.
 *     RETURN           ; 메서드 종료.
 *     MAXSTACK = 3     ; 이 메서드가 사용하는 최대 스택 크기입니다.
 *     MAXLOCALS = 5    ; 파라미터와 지역 변수를 포함한 총 변수 수입니다.
 *                      ; 여기서는 this, a, b, bitmask, marker 총 5개입니다.
 *
 *   // access flags 0x11
 *   public final getA()I
 *    L0
 *     LINENUMBER 143 L0
 *     ALOAD 0
 *     GETFIELD classes/SyntheticConstructorExample.a : I
 *     IRETURN
 *    L1
 *     LOCALVARIABLE this Lclasses/SyntheticConstructorExample; L0 L1 0
 *     MAXSTACK = 1
 *     MAXLOCALS = 1
 *
 *   // access flags 0x11
 *   public final getB()Ljava/lang/String;
 *   @Lorg/jetbrains/annotations/NotNull;() // invisible
 *    L0
 *     LINENUMBER 143 L0
 *     ALOAD 0
 *     GETFIELD classes/SyntheticConstructorExample.b : Ljava/lang/String;
 *     ARETURN
 *    L1
 *     LOCALVARIABLE this Lclasses/SyntheticConstructorExample; L0 L1 0
 *     MAXSTACK = 1
 *     MAXLOCALS = 1
 *
 *   // access flags 0x1
 *   public <init>()V
 *    L0
 *     ALOAD 0
 *     ICONST_0
 *     ACONST_NULL
 *     ICONST_3
 *     ACONST_NULL
 *     INVOKESPECIAL classes/SyntheticConstructorExample.<init> (ILjava/lang/String;ILkotlin/jvm/internal/DefaultConstructorMarker;)V
 *     RETURN
 *    L1
 *     LOCALVARIABLE this Lclasses/SyntheticConstructorExample; L0 L1 0
 *     MAXSTACK = 5
 *     MAXLOCALS = 1
 * }
 * ```
 *
 * References:
 * - [Make kotlin.jvm.internal.DefaultConstructorMarker public](https://github.com/JetBrains/kotlin/pull/4194)
 * - [Two Additional types in default constructor in Kotlin?]
 * - [Quirks of Kotlin - Synthetic Constructors](https://ximedes.com/blog/2023-04-14/kotlin-quirks-synthetic-constructors)
 * - [Chapter 6. The Java Virtual Machine Instruction Set](https://docs.oracle.com/javase/specs/jvms/se21/html/jvms-6.html)
 */
class SyntheticConstructorExample(val a: Int = 0, val b: String = "")

fun main() {
    println(SyntheticConstructorExample(10).a)
    // 컴파일러가 생성하는 실제 호출:
    // SyntheticConstructorExample(10, 0, 0b10, null)

    SyntheticConstructorExample::class.java.constructors.forEach { constructor ->
        println(constructor)
    }
    // Output:
    //  public classes.SyntheticConstructorExample() ->  파라미터 없는 기본 생성자 (모든 파라미터가 선택적이므로 생성됨)
    //  public classes.SyntheticConstructorExample(int,java.lang.String,int,kotlin.jvm.internal.DefaultConstructorMarker) -> synthetic 생성자 (기본값 처리용)
    //  public classes.SyntheticConstructorExample(int,java.lang.String) -> 주 생성자 (모든 파라미터)
}