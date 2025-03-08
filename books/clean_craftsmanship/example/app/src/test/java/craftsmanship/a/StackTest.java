package craftsmanship.a;


import static org.testng.Assert.*;

import org.testng.annotations.Test;

public class StackTest {
    /**
     * 다음과 같이 아무 것도 하지 않는 테스트를 만드는 것으로 시작합니다.
     * <p>
     * 빈 테스트라도 실행해 봄으로써 실행 환경이 잘 작동하는지 확인할 수 있습니다.
     * <p>
     * 이 상태에서 "작성하고 싶은 코드를 알고 있다고 가정"하고 무엇을 테스트할지 고민합니다.
     * 여기서 테스트 하고 싶은 것은 `public class Stack`입니다.
     *
     * @throws Exception
     */
    @Test
    public void empty_00_00() throws Exception {
    }

    /**
     * "작성하고 싶은 코드를 작성하게 만드는 테스트를 쓰라."
     * <p>
     * `:app:test --tests "craftsmanship.a.StackTest"` 테스트 시 실패합니다.
     *
     * @throws Exception
     */
    @Test
    public void canCrateStack_00_44_compile_error() throws Exception {
        // NotImplementedStack stack = new NotImplementedStack();

        // `NotImplementedStack` 클래스는 아직 구현되지 않았으므로, 다음과 같은 컴파일 타임 에러가 발생합니다.
        //
        // /path/to/app/src/test/java/craftsmanship/a/StackTest.java:22: error: cannot find symbol
        //      NotImplementedStack stack = new NotImplementedStack();
        //      ^
        // symbol:   class NotImplementedStack
        // location: class StackTest
    }

    /**
     * {@link StackTest#canCrateStack_00_44_compile_error}가 동작하도록 `public class Stack`을 구현합니다.
     * <p>
     * 이제 `:app:test --tests "craftsmanship.a.StackTest"` 테스트가 통과합니다.
     *
     * @throws Exception
     */
    @Test
    public void canCrateStack_01_04_success() throws Exception {
        Stack_00_54 stack = new Stack_00_54();
    }

    /**
     * {@link Stack_00_54}은 좋은 이름이 아니라고 생각하여 {@link Stack_01_09_Refactored}로 이름을 바꿨다고 가정합니다.
     * <p>
     * 이때에도 테스트는 여전히 통과합니다.
     * 이를 통해 다음과 같은 <a href="https://www.codecademy.com/article/tdd-red-green-refactor">빨강 - 파랑 - 리팩토링</a> 주기를 확인할 수 있습니다.
     * <p>
     * 빨강({@link StackTest#canCrateStack_00_44_compile_error})
     * -> 파랑({@link StackTest#canCrateStack_01_04_success})
     * -> 리팩토링({@link StackTest#canCrateStack_01_09_refactoring_success})
     *
     * @throws Exception
     */
    @Test
    public void canCrateStack_01_09_refactoring_success() throws Exception {
        Stack_01_09_Refactored stack = new Stack_01_09_Refactored();
    }

    /**
     * 단정문(assertion) 테스트를 추가합니다.
     * <p>
     * 01:49에는 이 파일 가장 상단에 `import static org.testng.Assert.*;`를 추가합니다.
     *
     * @throws Exception
     */
    @Test
    public void isEmpty_01_24_compile_error() throws Exception {
        Stack_01_09_Refactored stack = new Stack_01_09_Refactored();
        // assertTrue(stack.isEmpty());

        // 하지만 아직 `isEmpty()` 메서드가 구현되지 않았으므로, 다음과 같은 컴파일 타임 에러가 발생합니다.
        //
        // /path/to/app/src/test/java/craftsmanship/a/StackTest.java:78: error: cannot find symbol
        // assertTrue(stack.isEmpty());
        //                 ^
        // symbol:   method isEmpty()
        // location: variable stack of type Stack_01_09_Refactored
    }

    /**
     * 단정문(assertion) 테스트를 추가합니다.
     * <p>
     * {@link Stack_01_49#isEmpty()}가 구현되어서 컴파일 타임 에러는 발생하지 않지만,
     * 해당 메서드는 항상 `false`를 리턴하므로 테스트는 실패합니다.
     * <p>
     * 이는 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * 이 법칙을 요구하는 이유는 "테스트가 실패해야 할 때 정말 실패하는지" 확인하기 위함입니다.
     *
     * @throws Exception
     */
    @Test
    public void isEmpty_01_49_fail() throws Exception {
        Stack_01_49 stack = new Stack_01_49();
        // 일부러 실패하는 테스트를 작성하여, 정말 실패하는지 확인합니다.
        // assertTrue(stack.isEmpty()); // 다음 코드 진행 위해 주석 처리

        // StackTest > test_01_49_isEmpty FAILED
        //  java.lang.AssertionError at StackTest.java:97
    }

    /**
     * 단정문(assertion) 테스트를 추가합니다.
     * <p>
     * {@link Stack_01_58#isEmpty()}는 항상 `true`를 리턴하므로 테스트는 성공합니다.
     *
     * @throws Exception
     */
    @Test
    public void isEmpty_01_58_success() throws Exception {
        Stack_01_58 stack = new Stack_01_58();
        assertTrue(stack.isEmpty());
        // StackTest > test_01_49_isEmpty FAILED
        //  java.lang.AssertionError at StackTest.java:97
    }

    /**
     * 스택 기능을 구현하기 위해 `push(int)` 메서드를 추가해야 한다는 것을 압니다.
     * 하지만 프로덕션 코드를 구현하기 전에, '규칙 1: 작성하고 싶은 코드를 작성하게 만드는 테스트를 쓰라.'에 따라
     * "`push` 함수를 작성할 수밖에 없게 하는 테스트"를 작성합니다.
     * <p>
     * 이 테스트는 컴파일 에러가 발생합니다.
     *
     * @throws Exception
     */
    @Test
    public void canPush_02_24_compile_error() throws Exception {
        Stack_01_58 stack = new Stack_01_58();
        // stack.push(0); // 다음 코드 진행 위해 주석 처리

        // 아직 `push(int)` 메서드가 구현되지 않았으므로, 다음과 같은 컴파일 타임 에러가 발생합니다.
        //
        // /path/to/app/src/test/java/craftsmanship/a/StackTest.java:137: error: cannot find symbol
        // stack.push(0); // 다음 코드 진행 위해 주석 처리
        //      ^
        // symbol:   method push(int)
        // location: variable stack of type Stack_01_58
    }

    /**
     * 컴파일 성공하도록 프로덕션 코드를 작성합니다.
     * `push` 메서드가 추가되었으므로 이제 컴파일에 성공합니다.
     *
     * @throws Exception
     */
    @Test
    public void canPush_02_31_success() throws Exception {
        Stack_02_31 stack = new Stack_02_31();
        stack.push(0);
    }

    /**
     * `push` 메서드가 추가되어서 {@link StackTest#canPush_02_31_success} 케이스는 성공합니다.
     * 하지만 단정문이 하나도 없기 때문에, 단정문으로 무엇을 확인해야 할지 고민합니다.
     * <p>
     * `push` 후에는 스택이 비어있지 않아야 하므로, `isEmpty()` 여부를 단정문으로 추가합니다.
     * 하지만 `isEmpty()`는 항상 `true`를 리턴하므로 테스트는 실패해야 합니다.
     *
     * @throws Exception
     */
    @Test
    public void canPush_02_54_isEmpty_fail() throws Exception {
        Stack_02_31 stack = new Stack_02_31();
        stack.push(0);
        // 이 시점에서는 `isEmpty()`가 항상 `true`를 리턴하므로, 테스트는 실패해야 합니다.
        // assertFalse(stack.isEmpty()); // 다음 코드 진행 위해 주석 처리
    }

    /**
     * {@link Stack_03_46} 구현 클래스는 {@link Stack_03_46#empty} 플래그를 추가하여
     * {@link StackTest#canPush_02_54_isEmpty_fail}에서 발생하는 테스트 실패를 해결합니다.
     * <p>
     * 이 테스트 통과 후에는 '규칙 2: 실패시켜라. 통과시켜라. 그리고 정리하라.'에 따라 필요하면 코드를 정리할 시간을 갖습니다.
     * <p>
     * 참고로 책에서는 스택 생성 중복 코드가 마음에 안 들어서 필드로 추출하여 정리하는데,
     * 여기서는 순서대로 스택 클래스 구현을 따라가기 위해서 테스트마다 생성하도록 합니다.
     *
     * @throws Exception
     */
    @Test
    public void canPush_03_46_isEmpty_success() throws Exception {
        Stack_03_46 stack = new Stack_03_46();
        stack.push(0);
        assertFalse(stack.isEmpty());
    }

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * 이 법칙을 요구하는 이유는 "테스트가 실패해야 할 때 정말 실패하는지" 확인하기 위함입니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushAndOnePop_05_17_compile_error() throws Exception {
        Stack_03_46 stack = new Stack_03_46();
        stack.push(0);
        // stack.pop(); // 다음 코드 진행 위해 주석 처리

        // 아직 `pop()` 메서드가 구현되지 않았으므로, 다음과 같은 컴파일 타임 에러가 발생합니다.
        //
        // /path/to/app/src/test/java/craftsmanship/a/StackTest.java:206: error: cannot find symbol
        // stack.pop();
        //      ^
        // symbol:   method pop()
        // location: variable stack of type Stack_03_46
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushAndOnePop_05_31_success() throws Exception {
        Stack_05_31 stack = new Stack_05_31();
        stack.push(0);
        stack.pop();
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."을 따릅니다.
     * <p>
     * 스택이 비었는지 확인하는 단정문을 추가합니다.
     * 하지만 `pop()` 메서드는 `empty` 플래그를 업데이트하지 않으므로 테스트는 실패해야 합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushAndOnePop_05_51_isEmpty_fail() throws Exception {
        Stack_05_31 stack = new Stack_05_31();
        stack.push(0);
        stack.pop();
        // 이 시점에서는 `pop` 호출 시에 empty 플래그를 바꾸지 않기 때문에 테스트는 실패해야 합니다.
        // assertTrue(stack.isEmpty()); // 다음 코드 진행 위해 주석 처리

        // StackTest > afterPushAndPop_05_51_isEmpty_fail FAILED
        //     java.lang.AssertionError at StackTest.java:242
    }

    /**
     * {@link Stack_06_06#pop()} 호출 시에 무조건 {@link Stack_06_06#empty} 플래그를 `true`로 업데이트 함으로써
     * {@link StackTest#afterPushAndOnePop_05_51_isEmpty_fail}에서 발생하는 테스트 실패를 해결합니다.
     * <p>
     * 정리할 것이 없으므로 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."로 돌아갑니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushAndOnePop_06_06_isEmpty_success() throws Exception {
        Stack_06_06 stack = new Stack_06_06();
        stack.push(0);
        stack.pop();
        assertTrue(stack.isEmpty());
    }

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * 이제 스택 사이즈를 확인합니다.
     * 마찬가지로 컴파일 에러가 발생하는 테스트 코드를 작성합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterTwoPushes_06_48_sizeIsTwo_compile_error() throws Exception {
        Stack_06_06 stack = new Stack_06_06();
        stack.push(0);
        stack.push(0);
        // assertEquals(stack.getSize(), 2); // 다음 코드 진행 위해 주석 처리

        // 아직 `getSize()` 메서드가 구현되지 않았으므로, 다음과 같은 컴파일 타임 에러가 발생합니다.
        //
        // /path/to/app/src/test/java/craftsmanship/a/StackTest.java:277: error: cannot find symbol
        // assertEquals(stack.getSize(), 2);
        //                      ^
        // symbol:   method getSize()
        // location: variable stack of type Stack_06_06
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link Stack_07_23#getSize()}를 추가하여 {@link StackTest#afterTwoPushes_06_48_sizeIsTwo_compile_error}에서 발생하는 컴파일 에러 해결을 시도합니다.
     * <p>
     * 하지만, `getSize()`는 항상 `0`을 리턴하므로 테스트는 실패합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterTwoPushes_07_23_sizeIsTwo_fail() throws Exception {
        Stack_07_23 stack = new Stack_07_23();
        stack.push(0);
        stack.push(0);
        // assertEquals(stack.getSize(), 2); // 다음 코드 진행 위해 주석 처리

        // Expected :2
        // Actual   :0
        //
        // StackTest > afterTwoPushes_07_23_sizeIsTwo_fail FAILED
        //     java.lang.AssertionError at StackTest.java:302
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link Stack_07_23#getSize()}가 항상 2를 리턴하도록 하여
     * {@link StackTest#afterTwoPushes_07_23_sizeIsTwo_fail}에서 발생하는 테스트 실패 해결을 시도합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterTwoPushes_07_32_sizeIsTwo_success() throws Exception {
        Stack_07_32 stack = new Stack_07_32();
        stack.push(0);
        stack.push(0);
        assertEquals(stack.getSize(), 2);
    }

    /**
     * '규칙 1: 작성하고 싶은 코드를 작성하게 만드는 테스트를 쓰라.'에 따라 "더 나은 해결 방법을 찾을 수밖에 없도록"
     * 이전 테스트 {@link StackTest#afterPushAndOnePop_06_06_isEmpty_success}에 테스트를 추가합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterOnePushAndOnePop_08_06_isEmpty_and_sizeIsOne_fail() throws Exception {
        Stack_07_32 stack = new Stack_07_32();
        stack.push(0);
        stack.pop();
        assertTrue(stack.isEmpty());
        // 이 시점에서는 `getSize`는 항상 2를 리턴하므로 테스트는 실패해야 합니다.
        // assertEquals(stack.getSize(), 1); // 다음 코드 진행 위해 주석 처리

        // Expected :1
        // Actual   :2
        //
        // StackTest > afterOnePushAndOnePop_08_06_isEmpty_and_sizeIsOne_fail FAILED
        //     java.lang.AssertionError at StackTest.java:344
    }

    /**
     * {@link Stack_08_56} 구현 클래스는 {@link Stack_08_56#size} 필드를 추가하여
     * {@link StackTest#afterOnePushAndOnePop_08_06_isEmpty_and_sizeIsOne_fail}에서 발생하는 테스트 실패를 해결합니다.
     * <p>
     * 하지만 책에서는 실수로 `1`과 비교하는 단정문을 작성해서 실패하는 과정을 거치기 때문에, 그 과정을 그대로 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void afterOnePushAndOnePop_08_56_isEmpty_and_sizeIsOne_fail() throws Exception {
        Stack_08_56 stack = new Stack_08_56();
        stack.push(0);
        stack.pop();
        assertTrue(stack.isEmpty());
        // 원래는 0과 비교해야 하는데, 실수로 1과 비교하여 테스트가 실패합니다.
        // assertEquals(stack.getSize(), 1); // 다음 코드 진행 위해 주석 처리

        // Expected :1
        // Actual   :0
        //
        // StackTest > afterOnePushAndOnePop_08_56_isEmpty_and_sizeIsOne_fail FAILED
        //     java.lang.AssertionError at StackTest.java:368
    }

    /**
     * 사이즈 비교하는 단정문의 값을 `0`으로 수정하여
     * {@link StackTest#afterOnePushAndOnePop_08_56_isEmpty_and_sizeIsOne_fail}에서 발생하는 테스트 실패를 해결합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterOnePushAndOnePop_09_28_isEmpty_and_sizeIsOne() throws Exception {
        Stack_08_56 stack = new Stack_08_56();
        stack.push(0);
        stack.pop();
        assertTrue(stack.isEmpty());
        assertEquals(stack.getSize(), 0);
    }

    /**
     * 테스트가 더 완전해지도록 다른 테스트에도 크기 검사를 추가합니다.
     *
     * @throws Exception
     */
    @Test
    public void canPush_09_51_isEmpty_success() throws Exception {
        Stack_08_56 stack = new Stack_08_56();
        stack.push(0);
        assertFalse(stack.isEmpty());
        assertEquals(stack.getSize(), 1);
    }

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * "빈 스택에서 `pop`을 호출하면 어떻게 될까?" 언더플로 예외가 발생해야 합니다.
     * <p>
     * 하지만 {@link Stack_08_56} 클래스에는 `Underflow` 예외 클래스가 없으므로 컴파일 에러가 발생합니다.
     *
     * @throws Exception
     */
    // @Test(expectedExceptions = Stack_08_56.Underflow.class) // 다음 코드 진행 위해 주석 처리
    public void popEmptyStack_10_27_throwsUnderflow_compile_error() throws Exception {
        // /path/to/app/src/test/java/craftsmanship/a/StackTest.java:415: error: cannot find symbol
        // @Test(expected = Stack_08_56.Underflow.class)
        //                             ^
        // symbol:   class Underflow
        // location: class Stack_08_56
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link Stack_10_36.Underflow} 예외 클래스를 추가하여
     * {@link StackTest#popEmptyStack_10_27_throwsUnderflow_compile_error}에서 발생하는 컴파일 에러를 해결합니다.
     * <p>
     * 하지만 `pop()` 메서드에서 언더플로 예외를 발생시키지 않으므로 테스트는 실패합니다.
     *
     * @throws Exception
     */
    // @Test(expected = Stack_10_36.Underflow.class) // 다음 코드 진행 위해 주석 처리
    public void popEmptyStack_10_50_throwsUnderflow_fail() throws Exception {
        Stack_10_36 stack = new Stack_10_36();
        stack.pop();

        // StackTest > popEmptyStack_10_36_throwsUnderflow_success FAILED
        //     java.lang.AssertionError at ExpectException.java:34
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link Stack_11_18#pop()} 메서드에서 사이즈가 0인 경우 언더플로 예외를 발생시켜
     * {@link StackTest#popEmptyStack_10_50_throwsUnderflow_fail}에서 발생하는 테스트 실패를 해결합니다.
     *
     * @throws Exception
     */
    @Test
    public void popEmptyStack_11_18_throwsUnderflow_success() throws Exception {
        Stack_11_18 stack = new Stack_11_18();
        assertThrows(Stack_11_18.Underflow.class, stack::pop);
    }

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * 이제 스택에 `push`된 값을 기억해야 합니다.
     * <p>
     * 하지만 현재 {@link Stack_11_18#pop()}은 항상 `-1`을 리턴하므로 테스트는 실패합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushXThenWillPopX_11_49_fail() throws Exception {
        Stack_11_18 stack = new Stack_11_18();
        stack.push(99);
        // 이 시점에서는 `pop`이 항상 -1을 리턴하므로 테스트는 실패해야 합니다.
        // assertEquals(stack.pop(), 99); // 다음 코드 진행 위해 주석 처리
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 이제 스택에 `push`된 값을 기억해야 합니다.
     * <p>
     * {@link Stack_11_57#pop()} 메서드에서 항상 `99`를 리턴하도록 하여
     * {@link StackTest#afterPushXThenWillPopX_11_49_fail}에서 발생하는 테스트 실패를 해결합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushXThenWillPopX_11_57_success() throws Exception {
        Stack_11_57 stack = new Stack_11_57();
        stack.push(99);
        assertEquals(stack.pop(), 99);
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 이제 스택에 `push`된 값을 기억해야 합니다.
     * <p>
     * '규칙 1: 작성하고 싶은 코드를 작성하게 만드는 테스트를 쓰라.'에 따라 "더 나은 해결 방법을 찾을 수밖에 없도록"
     * 이전 테스트 {@link StackTest#afterPushXThenWillPopX_11_57_success}에 테스트를 추가합니다.
     * <p>
     * 이 경우 99 외에 다른 값을 리턴해야 하는 케이스를 추가합니다.
     * 하지만 현재 {@link Stack_11_57#pop()}은 항상 `99`를 리턴하므로 테스트는 실패합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushXThenWillPopX_12_18_fail() throws Exception {
        Stack_11_57 stack = new Stack_11_57();
        stack.push(99);
        assertEquals(stack.pop(), 99);
        stack.push(88);
        // 이 시점에서는 `pop`이 항상 99를 리턴하므로 테스트는 실패해야 합니다.
        // assertEquals(stack.pop(), 88); // 다음 코드 진행 위해 주석 처리

        // Expected :88
        // Actual   :99
        //
        // StackTest > afterPushXThenWillPopX_12_18_fail FAILED
        //     java.lang.AssertionError at StackTest.java:511
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 이제 스택에 `push`된 값을 기억해야 합니다.
     * <p>
     * {@link Stack_12_50#element} 필드를 추가하고,
     * {@link Stack_12_50#push(int)} 시에 해당 필드를 업데이트하고,
     * 그리고 {@link Stack_12_50#pop()} 시에 해당 필드를 리턴합니다.
     * <p>
     * {@link StackTest#afterPushXThenWillPopX_12_18_fail}에서 발생하는 테스트 실패를 해결합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushXThenWillPopX_12_50_success() throws Exception {
        Stack_12_50 stack = new Stack_12_50();
        stack.push(99);
        assertEquals(stack.pop(), 99);
        stack.push(88);
        assertEquals(stack.pop(), 88);
    }

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * 선입 후출 동작을 테스트합니다.
     * 하지만 {@link Stack_12_50}은 마지막 요소만 기억하므로 테스트는 실패합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushXThenYAndWillPopYThenX_13_36_fail() throws Exception {
        Stack_12_50 stack = new Stack_12_50();
        stack.push(99);
        stack.push(88);
        assertEquals(stack.pop(), 88);
        // 이 시점에서는 `pop`이 항상 마지막에 푸시한 88을 리턴하므로 테스트는 실패해야 합니다.
        // assertEquals(stack.pop(), 99); // 다음 코드 진행 위해 주석 처리

        // Expected :99
        // Actual   :88
        //
        // StackTest > afterPushXThenYAndWillPopYThenX_13_36_fail FAILED
        //     java.lang.AssertionError at StackTest.java:558
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link Stack_12_50#element} 필드를 {@link Stack_14_13#elements} 배열로 변경합니다.
     * {@link Stack_14_13#push(int)} 메서드에서는 요소가 추가하고 {@link Stack_14_13#size}를 증가시킵니다.
     * {@link Stack_14_13#pop()} 메서드는 {@link Stack_14_13#size}를 감소시키고 마지막 요소를 리턴합니다.
     * <p>
     * {@link StackTest#afterPushXThenYAndWillPopYThenX_13_36_fail}에서 발생하는 테스트 실패를 해결합니다.
     *
     * @throws Exception
     */
    @Test
    public void afterPushXThenYAndWillPopYThenX_14_13_success() throws Exception {
        Stack_14_13 stack = new Stack_14_13();
        stack.push(99);
        stack.push(88);
        assertEquals(stack.pop(), 88);
        assertEquals(stack.pop(), 99);
    }
}

