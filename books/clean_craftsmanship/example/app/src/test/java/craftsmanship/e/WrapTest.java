package craftsmanship.e;

import org.testng.annotations.Test;

import java.util.function.BiFunction;

import static org.assertj.core.api.Assertions.assertThat;

/**
 * '줄 바꿈 문제'
 * <p>
 * 줄 바꿈이 없는 글이 문자열로 주어졌을 때, N 글자 너비의 화면에 맞도록 줄 바꿈 문자를 적절히 삽입하는 문제입니다. 되도록 단어 사이에만 줄 바꿈을 넣습니다.
 * <pre>
 * {@code
 * Four score and seven years ago our fathers brought forth upon this continent a new nation conceived in liberty and dedicated to the proposition that all men are created equal
 * }
 * </pre>
 * <p>
 * 목표 너비가 30이라면 다음과 같이 출력되어야 합니다.
 * <pre>
 * {@code
 * ====:====:====:====:====:====:
 * Four score and seven years ago
 * Our fathers brought forth upon
 * This continent a new nation
 * Conceived in liberty and
 * Dedicated to the proposition
 * That all men are created equal
 * ====:====:====:====:====:====:
 * }
 * </pre>
 */
public class WrapTest {
    /**
     * 아무 동작도 하지 않는 테스트부터 시작하여 컴파일이 성공하는지 확인합니다.
     *
     * @throws Exception
     */
    @Test
    public void nothing() throws Exception {
    }

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_0_fail() throws Exception {
        // assertThat(wrap_0_fail("Four", 7)).isEqualTo("Four"); // 다음 코드 진행 위해 주석 처리
        //Expected :"Four"
        //Actual   :null
    }

    private String wrap_0_fail(String s, int width) {
        return null;
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성하다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_0_success() throws Exception {
        assertThat(wrap_0("Four", 7)).isEqualTo("Four");
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param s 줄 바꿈 대상 문자열
     * @param width 목표 너비
     *
     * @return 줄 바꿈된 문자열
     */
    private String wrap_0(String s, int width) {
        return "Four";
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_1_fail() throws Exception {
        assertThat(wrap_0("Four", 7)).isEqualTo("Four");
        // 현재 "Four"만 리턴하도록 되어 있어서 반드시 실패합니다.
        // assertThat(wrap_0("Four score", 7)).isEqualTo("Four\nscore"); // 다음 코드 진행 위해 주석 처리
        //expected:
        //  "Four
        //  score"
        // but was:
        //  "Four"
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_1_success() throws Exception {
        assertWrapped(
            this::wrap_1,
            "Four", 7,
            "Four"
        );
        assertWrapped(
            this::wrap_1,
            "Four score", 7,
            "Four\nscore"
        );
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 모든 공백을 줄 바꿈으로 대체합니다.
     *
     * @param s 줄 바꿈 대상 문자열
     * @param width 목표 너비
     *
     * @return 줄 바꿈 된 문자열
     */
    private String wrap_1(String s, int width) {
        return s.replace(" ", "\n");
    }

    // 중복되는 단정문을 추출하여 좀더 테스트하기 편하게 만듭니다.
    private void assertWrapped(
        BiFunction<String, Integer, String> wrap,
        String s,
        int width,
        String expected
    ) {
        assertThat(wrap.apply(s, width)).isEqualTo(expected);
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_2_fail() throws Exception {
        assertWrapped(
            this::wrap_1,
            "Four", 7,
            "Four"
        );
        assertWrapped(
            this::wrap_1,
            "Four score", 7,
            "Four\nscore"
        );
        // "ago our"는 줄 바꿈이 없어야 하는데, 현재 공백마다 무조건 줄 바꿈을 추가하므로 실패합니다.
        // assertWrapped(
        //     this::wrap_1,
        //     "Four score and seven years ago our", 7,
        //     "Four\nscore\nand\nseven\nyears\nago our"
        // );
        //expected:
        //  "Four
        //  score
        //  and
        //  seven
        //  years
        //  ago our"
        // but was:
        //  "Four
        //  score
        //  and
        //  seven
        //  years
        //  ago
        //  our"
    }

    // 모든 공백을 줄 바꿈으로 대체하면 안 됩니다. 그러면 여러 생각이 떠오르기 시작할 겁니다.
    // - 어떤 경우에 줄 바꿈을 넣어야 할까?
    // - 아니면 일단 모든 공백을 줄 바꿈으로 바꾼 다음에 어떤 줄 바꿈을 다시 없앨지 찾아야 할까?
    //
    // 쉬운 해결책은 떠오르지 않을 것이고, 이는 '막다른 길'에 들어섰다는 것을 의미합니다.
    //
    // `wrap("Four score and seven years ago our", 7)` 테스트를 통과하려면
    // 줄 바꿈 알고리즘의 '많은 부분을 한 번에 수정'할 수밖에 없어 보입니다.
    // 이때 규칙 8 "현재 테스트를 통과시키기 위해 너무 많은 구현을 해야 한다면, 테스트를 지우고 더 쉽게 통과할 수 있는 더 단순한 테스트를 작성하라."를 따릅니다.

    /**
     * 규칙 8 "현재 테스트를 통과시키기 위해 너무 많은 구현을 해야 한다면, 테스트를 지우고 더 쉽게 통과할 수 있는 더 단순한 테스트를 작성하라."를 따릅니다.
     * <p>
     * {@link WrapTest#wrap_2_fail()}에서 '막다른 길'에 들어섰고, 이때의 해결책은
     * '하나 또는 더 많은 테스트를 지우고, 단계적으로 통과시킬 수 있는 더 단순한 테스트를 추가"하는 것입니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_3_success() throws Exception {
        assertWrapped(
            this::wrap_3,
            "", 1,
            ""
        );
    }

    /**
     * 규칙 8 "현재 테스트를 통과시키기 위해 너무 많은 구현을 해야 한다면, 테스트를 지우고 더 쉽게 통과할 수 있는 더 단순한 테스트를 작성하라."를 따릅니다.
     *
     * @param s 줄 바꿈 대상 문자열
     * @param width 목표 너비
     *
     * @return 줄 바꿈된 문자열
     */
    private String wrap_3(String s, int width) {
        return "";
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_4_fail() throws Exception {
        assertWrapped(
            this::wrap_3,
            "", 1,
            ""
        );
        // assertWrapped(
        //     this::wrap_3,
        //     "x", 1,
        //     "x"
        // ); // 다음 코드 진행 위해 주석 처리
        //Expected :"x"
        //Actual   :""
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_4_success() throws Exception {
        assertWrapped(
            this::wrap_4,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_4,
            "x", 1,
            "x"
        );
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 첫 번째 테스트는 퇴화한 상수를 반환하여 통과시키고, 두 번째 테스트는 입력을 그대로 반환하여 통과시킵니다.
     * <p>
     * 이는 앞서 {@link craftsmanship.a.StackTest}, {@link craftsmanship.b.PrimeFactorsTest},
     * {@link craftsmanship.d.SortTest} 등에서 봤던 패턴입니다.
     *
     * @param s 줄 바꿈 대상 문자열
     * @param width 목표 너비
     *
     * @return 줄 바꿈된 문자열
     */
    private String wrap_4(String s, int width) {
        return s;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_5_fail() throws Exception {
        assertWrapped(
            this::wrap_4,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_4,
            "x", 1,
            "x"
        );
        // assertWrapped(
        //     this::wrap_4,
        //     "xx", 1,
        //     "x\nx"
        // ); // 다음 코드 진행 위해 주석 처리
        //expected:
        //  "x
        //  x"
        // but was:
        //  "xx"
    }

    /**
     * 두 번째 법칙 "실패는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_5_success() throws Exception {
        assertWrapped(
            this::wrap_5,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_5,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_5,
            "xx", 1,
            "x\nx"
        );
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param s 줄 바꿈 대상 문자열
     * @param width 목표 너비
     *
     * @return 줄 바꿈된 문자열
     */
    private String wrap_5(String s, int width) {
        if (s.length() <= width) {
            return s;
        }

        // 바로 모든 공백 문자를 개행 문자로 바꾸는 대신, 더 퇴화된 수준의 코드를 먼저 작성합니다.
        // 아직은 일일이 처리할 수 있으므로, 아직은 '모든 공백을 개행문자로 치환'하는 일반화를 하지 않습니다.
        return s.substring(0, width) + "\n" + s.substring(width);
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_6_fail() throws Exception {
        assertWrapped(
            this::wrap_5,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_5,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_5,
            "xx", 1,
            "x\nx"
        );
        // 두 글자인 경우만 처리하고 있으므로, 실패해야 합니다.
        // assertWrapped(
        //     this::wrap_5,
        //     "xxx", 1,
        //     "x\nx\nx"
        // ); // 다음 코드 진행 위해 주석 처리
        //expected:
        //  "x
        //  x
        //  x"
        // but was:
        //  "x
        //  xx"
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_6_success() throws Exception {
        assertWrapped(
            this::wrap_6,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_6,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_6,
            "xx", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_6,
            "xxx", 1,
            "x\nx\nx"
        );
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param s 줄 바꿈 대상 문자열
     * @param width 목표 너비
     *
     * @return 줄 바꿈된 문자열
     */
    private String wrap_6(String s, int width) {
        if (s.length() <= width) {
            return s;
        }

        // 반복문을 사용할 수 있지만, 여기서는 재귀가 더 쉬운 방법으로 보입니다.
        return s.substring(0, width) + "\n" + wrap_6(s.substring(width), width);
    }

    // 퇴화된, 아주 단순한 패턴의 테스트를 추가하고 있습니다.
    // 단어나 공백은 아직 하나도 없고, 'x'로만 이뤄진 문자열뿐입니다.

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_7_fail() throws Exception {
        assertWrapped(
            this::wrap_6,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_6,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_6,
            "xx", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_6,
            "xxx", 1,
            "x\nx\nx"
        );
        // 규칙 7 "더 복잡한 경우로 넘어가기 전에, 현재의 더 단순한 경우를 철저히 테스트하라."를 따릅니다.
        // 목표 너비를 늘려 봅니다. 아직까지는 정상 동작합니다.
        assertWrapped(
            this::wrap_6,
            "xxx", 2,
            "xx\nx"
        );
        assertWrapped(
            this::wrap_6,
            "xxx", 3,
            "xxx"
        );

        // 공백 없는 문자열에서 목표 너비를 증가시키는 테스트는 충분한 것으로 보입니다.
        // 이제 공백을 추가해 보면, 실패합니다.
        // assertWrapped(
        //     this::wrap_6,
        //     "x x", 1,
        //     "x\nx"
        // ); // 다음 코드 진행 위해 주석 처리
        //expected:
        //  "x
        //  x"
        // but was:
        //  "x
        //
        //  x"
        //
        // 실패하는 이유는 다음과 같습니다:
        // "x x"
        // -> "x\n" + " x"
        // -> "x\n" + " \n" + "x"
        // -> "x\n \nx"가 되어 실패해야 합니다.
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_7_success() throws Exception {
        assertWrapped(
            this::wrap_7,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_7,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_7,
            "xx", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_7,
            "xxx", 1,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_7,
            "xxx", 2,
            "xx\nx"
        );
        assertWrapped(
            this::wrap_7,
            "xxx", 3,
            "xxx"
        );
        assertWrapped(
            this::wrap_7,
            "x x", 1,
            "x\nx"
        );
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link WrapTest#wrap_7_fail()}이 실패하는 이유는 `" x"` 부분 문자열 앞에 공백 문자가 같이
     * 다음 재귀 호출로 전달되기 때문으로 보입니다.
     * {@link String#trim()}을 사용하여 공백 문자를 제거하면 테스트를 통과시킬 수 있습니다.
     *
     * @param s 줄 바꿈 대상 문자열
     * @param width 목표 너비
     *
     * @return 줄 바꿈된 문자열
     */
    private String wrap_7(String s, int width) {
        if (s.length() <= width) {
            return s;
        }

        // 재귀 호출 전에 문자열 앞의 공백을 제거합니다
        return s.substring(0, width) + "\n" + wrap_7(s.substring(width).trim(), width);
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_8_fail() throws Exception {
        assertWrapped(
            this::wrap_7,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_7,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_7,
            "xx", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_7,
            "xxx", 1,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_7,
            "xxx", 2,
            "xx\nx"
        );
        assertWrapped(
            this::wrap_7,
            "xxx", 3,
            "xxx"
        );
        assertWrapped(
            this::wrap_7,
            "x x", 1,
            "x\nx"
        );
        // 규칙 7 "더 복잡한 경우로 넘어가기 전에, 현재의 더 단순한 경우를 철저히 테스트하라."를 따릅니다.
        // 이전 테스트에서 목표 너비를 증가시켜 봅니다.
        // assertWrapped(
        //     this::wrap_7,
        //     "x x", 2,
        //     "x\nx"
        // ); // 다음 코드 진행 위해 주석 처리
        //expected:
        //  "x
        //  x"
        // but was:
        //  "x
        //  x"
        //
        // 콘솔 출력상으로는 같아 보이지만, 실패하는 이유는 다음과 같습니다.
        // "x " + "\n" + "x"
        //   ^
        //   첫 번째 부분 문자열 뒤에 공백이 붙어 있습니다.
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_8_success() throws Exception {
        assertWrapped(
            this::wrap_8,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_8,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_8,
            "xx", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_8,
            "xxx", 1,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_8,
            "xxx", 2,
            "xx\nx"
        );
        assertWrapped(
            this::wrap_8,
            "xxx", 3,
            "xxx"
        );
        assertWrapped(
            this::wrap_8,
            "x x", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_8,
            "x x", 2,
            "x\nx"
        );
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link WrapTest#wrap_8_fail()}이 실패하는 이유는 앞 부분 문자열 뒤에 공백이 붙어있기 때문입니다.
     * `"x \nx" != "x\nx"`
     * <p>
     * {@link WrapTest#wrap_7(String, int)}과 마찬가지로 앞 부분 문자열에 대해 {@link String#trim()}을 호출하면
     * 테스트를 통과시킬 수 있습니다.
     *
     * @param s 줄 바꿈 대상 문자열
     * @param width 목표 너비
     *
     * @return 줄 바꿈된 문자열
     */
    private String wrap_8(String s, int width) {
        if (s.length() <= width) {
            return s;
        }

        return s.substring(0, width).trim()
            + "\n"
            + wrap_8(s.substring(width).trim(), width);
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_9_fail() throws Exception {
        assertWrapped(
            this::wrap_8,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_8,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_8,
            "xx", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_8,
            "xxx", 1,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_8,
            "xxx", 2,
            "xx\nx"
        );
        assertWrapped(
            this::wrap_8,
            "xxx", 3,
            "xxx"
        );
        assertWrapped(
            this::wrap_8,
            "x x", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_8,
            "x x", 2,
            "x\nx"
        );
        // 규칙 7 "더 복잡한 경우로 넘어가기 전에, 현재의 더 단순한 경우를 철저히 테스트하라."를 따릅니다.
        // 이전 테스트에서 목표 너비를 증가시켜 봅니다.
        assertWrapped(
            this::wrap_8,
            "x x", 3,
            "x x"
        );
        // 조금 더 복잡한 테스트를 추가해 봅니다.
        assertWrapped(
            this::wrap_8,
            "x x x", 1,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_8,
            "x x x", 2,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_8,
            "x x x", 3,
            "x x\nx"
        );
        assertWrapped(
            this::wrap_8,
            "x x x", 4,
            "x x\nx"
        );
        assertWrapped(
            this::wrap_8,
            "x x x", 5,
            "x x x"
        );
        // 모든 테스트가 통과하므로, 여기에 `"x"`를 더 추가해도 의미가 없을 것으로 보입니다.
        // 그러면 이제 단일 문자가 아닌, 여러 문자인 경우를 추가합니다.
        assertWrapped(
            this::wrap_8,
            "xx xx", 1,
            "x\nx\nx\nx"
        );
        assertWrapped(
            this::wrap_8,
            "xx xx", 2,
            "xx\nxx"
        );
        assertWrapped(
            this::wrap_8,
            "xx xx", 3,
            "xx\nxx"
        );
        // "문자" 기준으로 처리하고 있으므로, "단어" 기준으로 줄 바꿈하길 기대하는 테스트는 실패해야 정상입니다.
        // assertWrapped(
        //     this::wrap_8,
        //     "xx xx", 4,
        //     "xx\nxx"
        // ); // 다음 코드 진행 위해 주석 처리
        //expected:
        //  "xx
        //  xx"
        // but was:
        //  "xx x
        //  x"
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_9_success() throws Exception {
        assertWrapped(
            this::wrap_9,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_9,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_9,
            "xx", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "xxx", 1,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "xxx", 2,
            "xx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "xxx", 3,
            "xxx"
        );
        assertWrapped(
            this::wrap_9,
            "x x", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x", 2,
            "x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x", 3,
            "x x"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 1,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 2,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 3,
            "x x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 4,
            "x x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 5,
            "x x x"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx", 1,
            "x\nx\nx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx", 2,
            "xx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx", 3,
            "xx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx", 4,
            "xx\nxx"
        );
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 앞서 {@link WrapTest#wrap_9_fail()}이 실패하는 이유는 '단어 단위'가 아닌 '문자 단위'로 줄 바꿈을 하기 때문입니다.
     *
     * <pre>
     * {@code
     * wrap("xx xx", 4)
     * // "xx\nxx"를 리턴해야 하지만,
     * // 실제로는 "xx x\nx" 리턴합니다.
     * }
     * </pre>
     * <p>
     * 따라서 '단어 단위'로 줄 바꿈을 하도록 단어 사이의 공백을 찾아서 줄 바꿈하도록 수정해야 합니다.
     * <p>
     * {@link String#lastIndexOf(String, int)}를 사용하면, 주어진 인덱스로부터 "역으로"(backward) 탐색하여
     * 찾으려는 문자열이 등장하는 인덱스를 반환합니다. 존재하지 않으면 `-1`을 리턴합니다.
     * <p>
     * 주어진 문자열에 대해 `width` 범위 내에서 가장 뒤에 있는 공백을 찾고,
     * 해당 공백을 기준으로 줄 바꿈을 수행한다면 '단어 단위'의 줄 바꿈이 가능합니다.
     *
     * <pre>
     * {@code
     * "xx xx", 4
     *    ^
     *    └ 2를 리턴합니다.
     *
     * 그리고 해당 인덱스로 다시 문자를 분할한다면 다음과 같이 분할이 되고,
     * "xx" + "\n" + wrap("xx", 4)
     *
     * 최종적으로 "xx\nxx"가 리턴됩니다.
     * }
     * </pre>
     *
     * @param s 줄 바꿈 대상 문자열
     * @param width 목표 너비
     *
     * @return 줄 바꿈된 문자열
     */
    private String wrap_9(String s, int width) {
        if (s.length() <= width) return s;

        var newLineAt = s.lastIndexOf(" ", width);
        if (newLineAt == -1) {
            newLineAt = width;
        }

        return s.substring(0, newLineAt).trim()
            + "\n"
            + wrap_9(s.substring(newLineAt).trim(), width);
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void wrap_10_success() throws Exception {
        assertWrapped(
            this::wrap_9,
            "", 1,
            ""
        );
        assertWrapped(
            this::wrap_9,
            "x", 1,
            "x"
        );
        assertWrapped(
            this::wrap_9,
            "xx", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "xxx", 1,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "xxx", 2,
            "xx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "xxx", 3,
            "xxx"
        );
        assertWrapped(
            this::wrap_9,
            "x x", 1,
            "x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x", 2,
            "x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x", 3,
            "x x"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 1,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 2,
            "x\nx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 3,
            "x x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 4,
            "x x\nx"
        );
        assertWrapped(
            this::wrap_9,
            "x x x", 5,
            "x x x"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx", 1,
            "x\nx\nx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx", 2,
            "xx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx", 3,
            "xx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx", 4,
            "xx\nxx"
        );
        // 테스트를 더 추가해 봅니다.
        assertWrapped(
            this::wrap_9,
            "xx xx", 5,
            "xx xx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx xx", 1,
            "x\nx\nx\nx\nx\nx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx xx", 2,
            "xx\nxx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx xx", 3,
            "xx\nxx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx xx", 4,
            "xx\nxx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx xx", 5,
            "xx xx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx xx", 6,
            "xx xx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx xx", 7,
            "xx xx\nxx"
        );
        assertWrapped(
            this::wrap_9,
            "xx xx xx", 8,
            "xx xx xx"
        );
        assertWrapped(
            this::wrap_9,
            "Four score and seven years ago our fathers brought forth upon this continent a new nation conceived in liberty and dedicated to the proposition that all men are created equal", 30,
            """
                Four score and seven years ago
                our fathers brought forth upon
                this continent a new nation
                conceived in liberty and
                dedicated to the proposition
                that all men are created equal"""
        );
    }
}




