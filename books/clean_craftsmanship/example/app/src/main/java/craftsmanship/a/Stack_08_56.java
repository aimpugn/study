package craftsmanship.a;

public class Stack_08_56 {
    private final boolean empty = true;
    private int size = 0;

    public boolean isEmpty() {
        return size == 0;
    }

    public void push(int element) {
        size++;
    }

    public int pop() {
        --size;
        return -1;
    }

    public int getSize() {
        return size;
    }
}
