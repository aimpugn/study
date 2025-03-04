package craftsmanship.a;

public class Stack_03_46 {
    private boolean empty = true;

    public boolean isEmpty() {
        return empty;
    }

    public void push(int element) {
        empty = false;
    }
}
