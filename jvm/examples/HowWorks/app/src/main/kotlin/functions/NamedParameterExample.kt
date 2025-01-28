package functions

import util.RunExample

fun call(before: (message: String) -> Unit = {}, after: () -> Unit = {}) {
    before("Before")
    println("Middle")
    after()
}

/**
 * Java에서는 메서드에 기본값을 사용할 수 없지만, Kotlin에서는 기본값을 사용할 수 있습니다.
 * 컴파일 된 바이트코드는 아래와 같습니다:
 * ```
 *   // access flags 0x19
 *   public final static functionWithNamedParameters(Ljava/lang/String;ILjava/lang/String;)V
 *     // annotable parameter count: 3 (invisible)
 *     @Lorg/jetbrains/annotations/NotNull;() // invisible, parameter 0
 *     @Lorg/jetbrains/annotations/NotNull;() // invisible, parameter 2
 *    L0
 *     ALOAD 0
 *     LDC "title"
 *     INVOKESTATIC kotlin/jvm/internal/Intrinsics.checkNotNullParameter (Ljava/lang/Object;Ljava/lang/String;)V
 *     ALOAD 2
 *     LDC "decoration"
 *     INVOKESTATIC kotlin/jvm/internal/Intrinsics.checkNotNullParameter (Ljava/lang/Object;Ljava/lang/String;)V
 *    L1
 *     LINENUMBER 74 L1
 *     ALOAD 2
 *     CHECKCAST java/lang/CharSequence
 *     ILOAD 1
 *     INVOKESTATIC kotlin/text/StringsKt.repeat (Ljava/lang/CharSequence;I)Ljava/lang/String;
 *     ASTORE 3
 *    L2
 *     LINENUMBER 75 L2
 *     NEW java/lang/StringBuilder
 *     DUP
 *     INVOKESPECIAL java/lang/StringBuilder.<init> ()V
 *     ALOAD 3
 *     INVOKEVIRTUAL java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
 *     BIPUSH 32
 *     INVOKEVIRTUAL java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
 *     ALOAD 0
 *     INVOKEVIRTUAL java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
 *     BIPUSH 32
 *     INVOKEVIRTUAL java/lang/StringBuilder.append (C)Ljava/lang/StringBuilder;
 *     ALOAD 3
 *     INVOKEVIRTUAL java/lang/StringBuilder.append (Ljava/lang/String;)Ljava/lang/StringBuilder;
 *     INVOKEVIRTUAL java/lang/StringBuilder.toString ()Ljava/lang/String;
 *     GETSTATIC java/lang/System.out : Ljava/io/PrintStream;
 *     SWAP
 *     INVOKEVIRTUAL java/io/PrintStream.println (Ljava/lang/Object;)V
 *    L3
 *     LINENUMBER 76 L3
 *     RETURN
 *    L4
 *     LOCALVARIABLE surround Ljava/lang/String; L2 L4 3
 *     LOCALVARIABLE title Ljava/lang/String; L0 L4 0
 *     LOCALVARIABLE loopCnt I L0 L4 1
 *     LOCALVARIABLE decoration Ljava/lang/String; L0 L4 2
 *     MAXSTACK = 2
 *     MAXLOCALS = 4
 *
 *   // access flags 0x1009
 *   public static synthetic functionWithNamedParameters$default(Ljava/lang/String;ILjava/lang/String;ILjava/lang/Object;)V
 *    L0
 *     LINENUMBER 69 L0
 *     ILOAD 3
 *     ICONST_1
 *     IAND
 *     IFEQ L1
 *    L2
 *     LINENUMBER 70 L2
 *     LDC "default title"
 *     ASTORE 0
 *    L1
 *     LINENUMBER 69 L1
 *    FRAME SAME
 *     ILOAD 3
 *     ICONST_2
 *     IAND
 *     IFEQ L3
 *    L4
 *     LINENUMBER 71 L4
 *     ICONST_1
 *     ISTORE 1
 *    L3
 *     LINENUMBER 69 L3
 *    FRAME SAME
 *     ILOAD 3
 *     ICONST_4
 *     IAND
 *     IFEQ L5
 *    L6
 *     LINENUMBER 72 L6
 *     LDC "\u2b50"
 *     ASTORE 2
 *    L5
 *     LINENUMBER 69 L5
 *    FRAME SAME
 *     ALOAD 0
 *     ILOAD 1
 *     ALOAD 2
 *     INVOKESTATIC functions/NamedParameterExampleKt.functionWithNamedParameters (Ljava/lang/String;ILjava/lang/String;)V
 *     RETURN
 *     MAXSTACK = 3
 *     MAXLOCALS = 5
 * ```
 * 컴파일 시 두 가지 메서드가 생성됩니다.
 * - 기본값이 없는 원래 형태의 메서드입니다.
 * - `$default`가 붙은 메서드는 기본값을 처리하는 메서드로, 호출 시 기본값을 자동으로 적용합니다.
 *   이 메서드는 호출 시 비트마스크를 사용해 어떤 매개변수에 기본값이 필요한지 결정합니다.
 *   비트마스크의 동작 방식은 [classes.SyntheticConstructorExample]와 비슷합니다.
 *
 *    ```
 *    // 원래 메서드
 *    public static final void functionWithNamedParameters(String title, int loopCnt, String decoration) {
 *        // 실제 메서드의 동작
 *    }
 *
 *    // $default 메서드
 *    public static final void functionWithNamedParameters$default(
 *        String title,
 *        int loopCnt,
 *        String decoration,
 *        int mask,
 *        Object unused // 호출 시 자바 호출 규칙에 따라 자리 채우기(placeholder) 용도로 사용됩니다.
 *    ) {
 *        if ((mask & 0x1) != 0) title = "default title";  // 기본값 처리
 *        if ((mask & 0x2) != 0) loopCnt = 1;
 *        if ((mask & 0x4) != 0) decoration = "⭐";
 *        functionWithNamedParameters(title, loopCnt, decoration);
 *    }
 *    ```
 *
 *
 * @param decoration [emoji.json](https://raw.githubusercontent.com/omnidan/node-emoji/master/lib/emoji.json)
 *
 * References:
 * - [콘솔에 emoji 출력되지 않을 경우](https://youtrack.jetbrains.com/issue/IJPL-106386/Console-does-not-display-some-emoji-characters-correctly#focus=Comments-27-5666689.0-0)
 */
fun functionWithNamedParameters(
    title: String = "default title",
    loopCnt: Int = 1,
    decoration: String = "⭐"
) {
    val surround = decoration.repeat(loopCnt)
    println("$surround $title $surround")
}

@RunExample
fun namedParameter() {
    val before = fun(
        message: String
    ): Unit {
        println("message: $message")
        println("and then...")
    }

    call(before = before)

    functionWithNamedParameters(
        title = "Pass value with named parameter"
    )
}
