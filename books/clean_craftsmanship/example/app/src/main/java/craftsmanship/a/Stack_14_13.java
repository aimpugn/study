package craftsmanship.a;

public class Stack_14_13 {
    private final int[] elements = new int[2];
    private int size = 0;

    public boolean isEmpty() {
        return size == 0;
    }

    public void push(int element) {
        // 컴파일 에러가 발생합니다.
        // this.elements = element; // Required type: int[]
        this.elements[size++] = element;
    }

    public int pop() {
        if (size == 0) {
            throw new Underflow();
        }

        // 컴파일 에러가 발생합니다.
        // return elements; // Required type: int
        return elements[--size];
    }

    public int getSize() {
        return size;
    }

    /**
     * 자바에서 클래스 내부에 정의된 클래스(Inner Class)는 기본적으로 바깥 클래스의 인스턴스에 종속됩니다.
     * 즉, `static`이 아닌 내부 클래스는 반드시 바깥 클래스의 인스턴스와 연결되어야 하므로, 독립적으로 생성할 수 없습니다.
     * 따라서 `static`이 아닌 상태에서 {@link Underflow} 클래스를 생성하려면 다음과 같이 {@link Stack_14_13} 클래스의 인스턴스가 필요합니다.
     * <pre>
     * {@code
     * throw (new Stack_10_36()).new Underflow();
     * }
     * </pre>
     * <p>
     * 반면, 정적 내부 클래스(static inner class)는 바깥 클래스의 인스턴스와 독립적으로 존재할 수 있습니다.
     */
    static class Underflow extends RuntimeException {
    }
}
