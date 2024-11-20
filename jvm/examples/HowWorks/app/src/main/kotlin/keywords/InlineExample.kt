package keywords

import util.RunExample

/**
 * ```
 * public static final void performAction(@NotNull Function0 action) {
 *    Intrinsics.checkNotNullParameter(action, "action");
 *    int $i$f$performAction = 0;
 *    String var2 = "Starting action";
 *    System.out.println(var2);
 *    action.invoke();
 *    var2 = "Ending action";
 *    System.out.println(var2);
 * }
 * ```
 */
inline fun performAction(action: () -> Unit) {
    println("Starting action")
    action()
    println("Ending action")
}

inline fun example(block1: () -> Unit, noinline block2: () -> Unit) {
    block1()
    block2()
}

@RunExample
fun inlineExample() {
    /**
     * ```
     * int $i$f$performAction = 0;
     * String var1 = "Starting action";
     * System.out.println(var1);
     * int var2 = 0;
     * String var3 = "Executing action";
     * System.out.println(var3);
     * var1 = "Ending action";
     * System.out.println(var1);
     * ```
     */
    performAction {
        println("Executing action")
    }

    /**
     * ```
     *     Function0 var4 = InlineExampleKt::inlineExample$lambda$2;
     *     int $i$f$example = 0;
     *     var2 = 0;
     *     var3 = "Inline block";
     *     System.out.println(var3);
     *     var4.invoke();
     * }
     *
     * private static final Unit inlineExample$lambda$2() {
     *    String var0 = "Noinline block";
     *    System.out.println(var0);
     *    return Unit.INSTANCE;
     * }
     * ```
     */
    example({
        println("Inline block")
    }, {
        println("Noinline block")
    })
}