package sandbox;

import java.util.ArrayDeque;
import java.util.Deque;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.Objects;

/**
 * ArrayDeque는 양쪽 끝에서 값을 넣고 뺄 수 있는 Deque 구현체입니다.
 *
 * <p>알고리즘 풀이에서는 보통 {@link Deque} 타입으로 선언하고,
 * 구현체만 {@link ArrayDeque}로 고정합니다. 이렇게 하면 코드는
 * "양끝 자료구조를 쓴다"는 의도를 드러내고, 구현체는 배열 기반의 빠른
 * 단일 스레드 자료구조를 선택했다는 사실만 남습니다.</p>
 *
 * <p>사용법을 고르는 핵심 질문은 하나입니다. 어느 쪽 끝을 넣고 빼는
 * 입구로 볼 것인가. 큐는 뒤에 넣고 앞에서 빼며, 공식 스택 메서드는 앞쪽을
 * top으로 씁니다. 알고리즘 추적에서는 오른쪽 끝을 top처럼 쓰기 위해
 * addLast/peekLast/pollLast를 쓰기도 합니다.</p>
 */
public final class ArrayDequeTest {
    private ArrayDequeTest() {
    }

    public static void main(String[] args) {
        stackWithOfficialMethods();
        stackWithTailAsTopForAlgorithmTrace();
        fifoQueue();
        doubleEndedQueue();
        emptyAndNullPolicy();

        System.out.println("All ArrayDeque guide checks passed");
    }

    private static void stackWithOfficialMethods() {
        Deque<String> stack = new ArrayDeque<>(3);

        // Deque의 공식 스택 메서드는 앞쪽(first/head)을 top으로 봅니다.
        // push(e) == addFirst(e), pop() == removeFirst()입니다.
        // add(e)는 addLast(e)와 같아서 뒤쪽에 넣습니다. 그래서 add/pop을 섞으면
        // "마지막에 넣은 값이 먼저 나온다"는 스택 의도가 코드에서 흐려집니다.
        // top 확인은 비어 있어도 괜찮으면 peek(), 비어 있으면 버그라면 getFirst()를 씁니다.
        stack.push("1");
        stack.push("2");
        stack.push("3");

        print("stack after push 1,2,3", stack);
        assertContents(stack, List.of("3", "2", "1"), "push는 새 값을 앞쪽 top에 둔다");
        assertEquals("3", stack.peek(), "peek는 top을 확인만 한다");
        assertEquals("3", stack.pop(), "pop은 top을 꺼낸다");
        assertContents(stack, List.of("2", "1"), "pop 이후 top 하나만 제거되어야 한다");

        Deque<String> addBased = new ArrayDeque<>(3);
        addBased.add("1");
        addBased.add("2");
        addBased.add("3");
        assertContents(addBased, List.of("1", "2", "3"), "add는 스택 top이 아니라 뒤쪽 tail에 넣는다");
    }

    private static void stackWithTailAsTopForAlgorithmTrace() {
        Deque<String> stack = new ArrayDeque<>(3);

        // 단조 스택처럼 입력을 왼쪽에서 오른쪽으로 보면서 "오른쪽 끝이 top"이라고
        // 생각하면 addLast/peekLast/pollLast가 눈에 더 잘 들어옵니다.
        // push/pop도 스택으로는 정석이지만, push는 앞쪽에 넣기 때문에 출력이 [3, 2, 1]처럼
        // 뒤집혀 보입니다. 입력 순서와 스택 상태를 같이 추적하려면 last 계열이 더 읽기 쉽습니다.
        stack.addLast("1");
        stack.addLast("2");
        stack.addLast("3");

        print("tail-top stack after addLast 1,2,3", stack);
        assertContents(stack, List.of("1", "2", "3"), "addLast는 입력 순서를 그대로 보여 준다");
        assertEquals("3", stack.peekLast(), "peekLast는 오른쪽 top을 확인만 한다");
        assertEquals("3", stack.pollLast(), "pollLast는 오른쪽 top을 꺼낸다");
        assertContents(stack, List.of("1", "2"), "pollLast 이후 오른쪽 끝 하나만 제거되어야 한다");
    }

    private static void fifoQueue() {
        Deque<String> queue = new ArrayDeque<>(3);

        // Queue 관점에서는 offer(e) == offerLast(e), poll() == pollFirst()입니다.
        // add(e)도 뒤에 넣지만, Collection의 "추가"라는 말에 가깝습니다.
        // 큐 역할을 드러낼 때는 offer/poll/peek 묶음이 더 선명합니다.
        queue.offer("1");
        queue.offer("2");
        queue.offer("3");

        print("queue after offer 1,2,3", queue);
        assertContents(queue, List.of("1", "2", "3"), "큐는 들어온 순서를 유지한다");
        assertEquals("1", queue.peek(), "peek는 가장 먼저 나갈 값을 확인만 한다");
        assertEquals("1", queue.poll(), "poll은 가장 먼저 들어온 값을 꺼낸다");
        assertContents(queue, List.of("2", "3"), "poll 이후 앞쪽 값 하나만 제거되어야 한다");

        Deque<String> addBased = new ArrayDeque<>(3);
        assertEquals(true, addBased.add("1"), "add도 ArrayDeque에서는 뒤쪽 삽입에 성공한다");
        assertEquals(true, addBased.offer("2"), "offer는 큐 언어로 뒤쪽 삽입 성공 여부를 돌려준다");
        assertContents(addBased, List.of("1", "2"), "add와 offer는 여기서는 같은 끝에 넣지만 의도가 다르다");
    }

    private static void doubleEndedQueue() {
        Deque<String> deque = new ArrayDeque<>(3);

        // 양끝을 모두 쓰는 문제라면 first/last를 메서드 이름에 직접 드러내는 편이 안전합니다.
        // offerFirst/offerLast는 실패를 false로 표현하는 계열이고,
        // addFirst/addLast는 실패를 예외로 표현하는 계열입니다.
        // ArrayDeque는 크기 제한이 없어서 보통 둘 다 성공하지만, 양끝 큐 예제에서는
        // first와 last 방향을 드러내는 이름 자체가 가장 중요합니다.
        deque.offerLast("middle");
        deque.offerFirst("front");
        deque.offerLast("back");

        print("deque after offerFirst/offerLast", deque);
        assertContents(deque, List.of("front", "middle", "back"), "first와 last 위치가 분명해야 한다");
        assertEquals("front", deque.pollFirst(), "pollFirst는 앞쪽 값을 꺼낸다");
        assertEquals("back", deque.pollLast(), "pollLast는 뒤쪽 값을 꺼낸다");
        assertContents(deque, List.of("middle"), "양끝을 꺼내면 가운데 값만 남아야 한다");
    }

    private static void emptyAndNullPolicy() {
        Deque<String> deque = new ArrayDeque<>();

        // poll/peek 계열은 비어 있으면 null을 돌려줍니다.
        // ArrayDeque는 null 원소를 금지하므로, 여기서 null은 "비어 있음"이라는 신호로만 해석됩니다.
        assertEquals(null, deque.poll(), "poll은 빈 deque에서 null을 반환한다");
        assertEquals(null, deque.peek(), "peek는 빈 deque에서 null을 반환한다");

        // remove/get/pop 계열은 비어 있는 상태를 버그로 보고 즉시 예외를 던집니다.
        // 알고리즘 코드에서 비어 있으면 안 되는 불변식을 확인하고 싶을 때 유용합니다.
        assertThrows(NoSuchElementException.class, deque::remove, "remove는 빈 deque에서 실패해야 한다");
        assertThrows(NoSuchElementException.class, deque::getFirst, "getFirst는 빈 deque에서 실패해야 한다");
        assertThrows(NoSuchElementException.class, deque::pop, "pop은 빈 stack에서 실패해야 한다");

        assertThrows(NullPointerException.class, () -> deque.offer(null), "ArrayDeque는 null 원소를 허용하지 않는다");
    }

    private static void print(String label, Deque<String> deque) {
        System.out.println(label + ": " + deque);
    }

    private static void assertContents(Deque<String> actual, List<String> expected, String message) {
        var actualValues = List.copyOf(actual);
        if (!actualValues.equals(expected)) {
            throw new AssertionError(
                message + ": expected=" + expected + ", actual=" + actualValues
            );
        }
    }

    private static void assertEquals(Object expected, Object actual, String message) {
        if (!Objects.equals(expected, actual)) {
            throw new AssertionError(
                message + ": expected=" + expected + ", actual=" + actual
            );
        }
    }

    private static void assertThrows(Class<? extends Throwable> expectedType, ThrowingRunnable runnable, String message) {
        try {
            runnable.run();
        } catch (Throwable actual) {
            if (expectedType.isInstance(actual)) {
                return;
            }
            throw new AssertionError(
                message + ": expected exception=" + expectedType.getSimpleName()
                    + ", actual exception=" + actual.getClass().getSimpleName(),
                actual
            );
        }

        throw new AssertionError(message + ": expected exception=" + expectedType.getSimpleName());
    }

    @FunctionalInterface
    private interface ThrowingRunnable {
        void run() throws Throwable;
    }
}
