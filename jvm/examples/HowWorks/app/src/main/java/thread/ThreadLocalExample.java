package thread;

import util.RunExample;

import java.util.concurrent.atomic.AtomicInteger;

public class ThreadLocalExample {
    private static final ThreadLocal<Integer> threadLocal = ThreadLocal.withInitial(() -> 0);

    @RunExample
    public void testThreadLocal() {
        System.out.printf("First ThreadId.get(): %d\n", ThreadId.get()); // First ThreadId.get(): 0
        System.out.printf("Second ThreadId.get(): %d\n", ThreadId.get()); // Second ThreadId.get(): 0

        new Thread(() -> System.out.printf("Third ThreadId.get(): %d\n", ThreadId.get())).start();
        // initialValue is called
        // Third ThreadId.get(): 1
        new Thread(() -> System.out.printf("Fourth ThreadId.get(): %d\n", ThreadId.get())).start();
        // initialValue is called
        // Fourth ThreadId.get(): 2

        Runnable task = () -> {
            threadLocal.set((int) (Math.random() * 100));
            System.out.println(Thread.currentThread().getName() + " : " + threadLocal.get());
        };

        new Thread(task).start(); // Thread-0 : 12
        new Thread(task).start(); // Thread-1 : 84

        ThreadLocal<String> threadLocal = ThreadLocal.withInitial(() -> "Hello");
        Thread aThread = new Thread(() -> {
            threadLocal.set("A Thread Value");
            System.out.println(Thread.currentThread().getName() + " : " + threadLocal.get());
            // Thread-4 : A Thread Value
        });

        Thread bThread = new Thread(() -> {
            threadLocal.set("B Thread Value");
            System.out.println(Thread.currentThread().getName() + " : " + threadLocal.get());
            // Thread-5 : B Thread Value
        });

        aThread.start();
        bThread.start();
    }
}

class ThreadId {
    // Atomic integer containing the next thread ID to be assigned
    private static final AtomicInteger nextId = new AtomicInteger(0);

    // Thread local variable containing each thread's ID
    private static final ThreadLocal<Integer> threadId =
            new ThreadLocal<Integer>() {
                @Override
                protected Integer initialValue() {
                    System.out.println("initialValue is called");
                    return nextId.getAndIncrement();
                }
            };

    /**
     * 여러 번 {@link #get()}을 호출해도 스레드가 동일하다면 값이 변하지 않습니다.
     * <p>
     * {@link #threadId} 속성은 {@link ThreadLocal} 타입으로, 각 스레드마다 독립적인 저장 공간을 가집니다.
     * 따라서 한 스레드에서 {@link #get()}을 여러 번 호출하더라도, ({@link #threadId})는 한번 초기화되고 반환된 값을 유지합니다.
     * {@link ThreadLocal#initialValue}는 스레드가 처음 {@link #get()}을 호출할 때만 실행됩니다.
     * 이후 {@link #get()}을 호출하면 이미 설정된 값이 반환되므로 증가하지 않습니다.
     * 이 값은 스레드가 살아 있는 동안 바뀌지 않습니다.
     * <p>
     *
     * @return int
     */
    public static int get() {
        return threadId.get();
    }
}