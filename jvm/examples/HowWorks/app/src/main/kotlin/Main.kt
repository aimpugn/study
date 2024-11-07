package main

/**
 * ```
 * ❯ javap -v MainKt.class
 * Classfile /Users/rody/VscodeProjects/study/jvm/examples/HowWorks/tmp/MainKt/main/MainKt.class
 *   Last modified 2024. 11. 8.; size 487 bytes
 *   SHA-256 checksum 88269fe81b19fe3b9447df598a7d29ec77bb95101cce6501575b85e2131cfeda
 *   Compiled from "Main.kt"
 * public final class main.MainKt
 *   minor version: 0
 *   major version: 65
 *   flags: (0x0031) ACC_PUBLIC, ACC_FINAL, ACC_SUPER
 *   this_class: #2                          // main/MainKt
 *   super_class: #4                         // java/lang/Object
 *   interfaces: 0, fields: 0, methods: 2, attributes: 2
 * Constant pool:
 *    #1 = Utf8               main/MainKt
 *    #2 = Class              #1             // main/MainKt
 *    #3 = Utf8               java/lang/Object
 *    #4 = Class              #3             // java/lang/Object
 *    #5 = Utf8               main
 *    #6 = Utf8               ()V
 *    #7 = Utf8               ([Ljava/lang/String;)V
 *    #8 = NameAndType        #5:#6          // main:()V
 *    #9 = Methodref          #2.#8          // main/MainKt.main:()V
 *   #10 = Utf8               args
 *   #11 = Utf8               [Ljava/lang/String;
 *   #12 = Utf8               Lkotlin/Metadata;
 *   #13 = Utf8               mv
 *   #14 = Integer            2
 *   #15 = Integer            0
 *   #16 = Utf8               k
 *   #17 = Utf8               xi
 *   #18 = Integer            48
 *   #19 = Utf8               d1
 *   #20 = Utf8               \u0000\b\n\u0000\n\u0002\u0010\u0002\n\u0000\u001a\u0006\u0010\u0000\u001a\u00020\u0001¨\u0006\u0002
 *   #21 = Utf8               d2
 *   #22 = Utf8
 *   #23 = Utf8               app
 *   #24 = Utf8               Main.kt
 *   #25 = Utf8               Code
 *   #26 = Utf8               LineNumberTable
 *   #27 = Utf8               LocalVariableTable
 *   #28 = Utf8               SourceFile
 *   #29 = Utf8               RuntimeVisibleAnnotations
 * {
 *   public static final void main();
 *     descriptor: ()V
 *     flags: (0x0019) ACC_PUBLIC, ACC_STATIC, ACC_FINAL
 *     Code:
 *       stack=0, locals=0, args_size=0
 *          0: return
 *       LineNumberTable:
 *         line 5: 0
 *
 *   public static void main(java.lang.String[]);
 *     descriptor: ([Ljava/lang/String;)V
 *     flags: (0x1009) ACC_PUBLIC, ACC_STATIC, ACC_SYNTHETIC
 *     Code:
 *       stack=0, locals=1, args_size=1
 *          0: invokestatic  #9                  // Method main:()V
 *          3: return
 *       LocalVariableTable:
 *         Start  Length  Slot  Name   Signature
 *             0       4     0  args   [Ljava/lang/String;
 * }
 * SourceFile: "Main.kt"
 * RuntimeVisibleAnnotations:
 *   0: #12(#13=[I#14,I#15,I#15],#16=I#14,#17=I#18,#19=[s#20],#21=[s#5,s#22,s#23])
 *     kotlin.Metadata(
 *       mv=[2,0,0]
 *       k=2
 *       xi=48
 *       d1=["\u0000\b\n\u0000\n\u0002\u0010\u0002\n\u0000\u001a\u0006\u0010\u0000\u001a\u00020\u0001¨\u0006\u0002"]
 *       d2=["main","","app"]
 *     )
 * ```
 * - `Constant Pool`: 클래스 파일에서 사용되는 모든 상수들의 집합입니다.
 *      문자열, 클래스 참조, 메소드 참조 등이 포함됩니다.
 * - `mv`: Kotlin 메타데이터 버전입니다.
 * - `k=2`: 파일 타입입니다. `1`이면 일반 클래스, `2`면 `FileFacade`입니다.
 *
 * References:
 * - [Multiplatform projects - Technical details](https://kotlinlang.org/docs/coding-conventions.html#multiplatform-projects)
 * - [FileFacade](https://kotlinlang.org/api/kotlinx-metadata-jvm/kotlin-metadata-jvm/kotlin.metadata.jvm/-kotlin-class-metadata/-file-facade)
 */
fun main() {}
