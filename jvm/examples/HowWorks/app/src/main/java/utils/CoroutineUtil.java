package utils;

import kotlin.Unit;
import kotlin.coroutines.Continuation;
import kotlin.coroutines.CoroutineContext;
import kotlin.coroutines.EmptyCoroutineContext;
import org.jetbrains.annotations.NotNull;

public class CoroutineUtil {
    public static Continuation<Unit> newContinuation() {
        return new Continuation<>() {
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
    }
}
