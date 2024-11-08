package classes;

/**
 * {@link classes.KotlinCompanionObjectExampleKt}와 동일한 기능을 하는 클래스입니다.
 * <p>
 * `javap` 디스어셈블리로 진입점 `main` 메서드를 제외하고 비교하면 거의 비슷함을 알 수 있습니다.
 * <pre>
 * {@code
 *
 * ❯ rg --files | rg class | xargs javap
 * Compiled from "KotlinCompanionObjectExampleByJava.java"
 * public final class classes.KotlinCompanionObjectExampleByJava$Companion {
 *   public classes.KotlinCompanionObjectExampleByJava$Companion();
 *   public void finalVoidMethod();
 * }
 * Compiled from "KotlinCompanionObjectExampleByJava.java"
 * public class classes.KotlinCompanionObjectExampleByJava {
 *   public static final classes.KotlinCompanionObjectExampleByJava$Companion Companion;
 *   public classes.KotlinCompanionObjectExampleByJava();
 *   public static void main(java.lang.String[]);
 *   static {};
 * }
 *
 * }
 *
 * {@code
 *
 * ❯ rg --files | rg class | xargs javap
 * Compiled from "KotlinCompanionObjectExample.kt"
 * public final class classes.KotlinCompanionObjectExample$Companion {
 *   public final void finalVoidMethod();
 *   public classes.KotlinCompanionObjectExample$Companion(kotlin.jvm.internal.DefaultConstructorMarker);
 * }
 * Compiled from "KotlinCompanionObjectExample.kt"
 * public final class classes.KotlinCompanionObjectExample {
 *   public static final classes.KotlinCompanionObjectExample$Companion Companion;
 *   public classes.KotlinCompanionObjectExample();
 *   static {};
 * }
 *
 * }
 * </pre>
 */
public class KotlinCompanionObjectExampleByJava {
    public static final Companion Companion = new Companion();

    /**
     * A class can be declared final if its definition is complete and no subclasses are desired or required.
     */
    public static final class Companion {
        private Companion() {
            // Singleton 형태로 제한
        }

        public void finalVoidMethod() {
            System.out.println("finalVoidMethod is called");
        }
    }

    public static void main(String[] args) {
        KotlinCompanionObjectExampleByJava.Companion.finalVoidMethod();
    }
}

