package craftsmanship.a;

public class Stack_06_06 {
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
}
