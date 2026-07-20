public final class SynchronizedBytecodeExample {

    private int value;

    public synchronized void instanceMethod() {
        value++;
    }

    public void instanceBlock() {
        synchronized (this) {
            value++;
        }
    }

    public static synchronized void staticMethod() {
        // The monitor is SynchronizedBytecodeExample.class.
    }
}
