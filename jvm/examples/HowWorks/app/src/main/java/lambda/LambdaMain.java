package lambda;

import kotlin.Unit;
import kotlin.coroutines.Continuation;
import kotlin.coroutines.CoroutineContext;
import kotlin.coroutines.EmptyCoroutineContext;
import org.jetbrains.annotations.NotNull;
import util.RunExampleInvoker;

public class LambdaMain {
    public static void main(String[] args) {
        // Continuation 객체 생성
        Continuation<Unit> continuation = new Continuation<>() {
            @Override
            public @NotNull CoroutineContext getContext() {
                return EmptyCoroutineContext.INSTANCE; // 빈 코루틴 컨텍스트
            }

            @Override
            public void resumeWith(@NotNull Object result) {
                if (result instanceof Throwable) {
                    ((Throwable) result).printStackTrace();
                } else {
                    System.out.println("Finished: " + result);
                }
            }
        };

        RunExampleInvoker.Companion.invoke("lambda", continuation);
    }
}
