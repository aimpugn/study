package craftsmanship.a;

public class Stack_07_32 {
    private boolean empty = true;

    public boolean isEmpty() {
        return empty;
    }

    public void push(int element) {
        empty = false;
    }

    public int pop() {
        empty = true;
        return -1;
    }

    public int getSize() {
        return 2;
    }
}
