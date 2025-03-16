package craftsmanship.f;

import org.testng.annotations.Test;

import java.util.function.BiFunction;

import static org.testng.Assert.assertEquals;

public class SineTest {
    /**
     * 부동소수점 숫자 비교 시 `==` 비교를 사용하는 것은 위험합니다.
     * IEEE 754 표준에서 부동소수점 값을 표현할 때 정확한 값이 아니라 근사값을 표현하기 때문입니다.
     * `float`, `double`은 정밀도의 한계로 인해 연산 과정에서 오차가 발생할 수밖에 없습니다.
     * <p>
     * 따라서 두 실수 값을 비교할 때는 일정 오차 범위(epsilon)를 정하고 그 범위 내에 들어오는지 여부를 확인합니다.
     */
    private static final double EPSILON = 0.0001;
    // 1e-9
    private static final double EPSILON2 = 0.000_000_001;

    @Test
    public void nothing() throws Exception {
    }

    /**
     * 컴파일하고 실행할 수 있음을 증명하기 위해 "아무 일도 하지 않는 테스트로 시작"합니다.
     * 테스트가 통과하면 삭제합니다.
     *
     * @throws Exception
     */
    @Test
    public void sines_0_fail() throws Exception {
        // assertThat(sin_0(0)).isEqualTo(null); // 다음 코드 진행 위해 주석 처리
        //Expected :null
        //Actual   :0.0
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 가장 퇴화한 경우로 시작합니다.
     * <p>
     * 다만, 책에서는 이 테스트 경우에 대해 앞서 다른 테스트들과 다른 점이 있다고 설명합니다.
     * <p>
     * 앞서 소인수 분해({@link craftsmanship.b.PrimeFactorsTest})와 볼링({@link craftsmanship.c.BowlingTest}) 예제를 보면,
     * 테스트 묶음이 커질수록 제약사항이 늘어가고, 각각의 테스트가 가능한 해답을 점점 좁혀 나가다 보면 최종 해답이 밝혀집니다.
     * <p>
     * 하지만, `sin(radian)` 함수는 그런 식으로 동작하지 않는 것 같다고 말합니다.
     * `sin(0) == 0` 테스트는 맞지만, 그 이외에는 해답을 좁히지 못하는 거 같다고 합니다.
     *
     * @throws Exception
     */
    @Test
    public void sines_0_success() throws Exception {
        assertEquals(sin_0(0), 0, EPSILON);
    }

    double sin_0(double radians) {
        return 0.0;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * 하지만, 이 테스트는 해답에 아무런 보탬이 되지 않습니다.
     * 즉, 문제를 어떻게 풀어야 하는지 힌트를 전혀 주지 않습니다.
     *
     * @throws Exception
     */
    @Test
    public void sines_1_fail() throws Exception {
        assertEquals(sin_0(0), 0, EPSILON);
        assertEquals(sin_0(Math.PI), 0, EPSILON);
        // 이 테스트는 실패하지만, 어떻게 통과시킬 수 있을지 여전히 힌트를 주지 않습니다.
        // assertEquals(sin_0(Math.PI / 2), 1, EPSILON); // 다음 코드 진행 위해 주석 처리
        //Expected :1.0
        //Actual   :0.0
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sines_1_success() throws Exception {
        assertEquals(sin_1(0), 0, EPSILON);
        assertEquals(sin_1(Math.PI), 0, EPSILON);
        assertEquals(sin_1(Math.PI / 2), 1, EPSILON);
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 코드만 작성한다."를 따릅니다.
     * <p>
     * 주어진 라디안 값의 사인 값을 근사하여 반환합니다.
     * 사인 함수의 테일러 급수(Taylor series)를 사용합니다.
     *
     * <pre>
     * {@code
     * 라디안(radian)?
     *
     * 각도를 나타내는 표준 단위입니다.
     * '원의 반지름의 길이'와 '동일한 호(arc)의 길이'를 이루는 각도를 1 라디안이라고 합니다.
     * }
     * </pre>
     * <p>
     * 테일러 급수는 '복잡한 함수들을 무한히 많은 다항식의 합으로 근사하는 방법',
     * 또는 함수의 모양을 '간단한 다항식의 조합으로 비슷하게 만드는 방법'입니다.
     * 즉, 복잡한 함수(사인, 코사인, 지수 함수 등)를 간단한 덧셈, 뺄셈, 곱셈, 나눗셈으로 근사하는 방법입니다.
     * <p>
     * 예를 들어, 종이 위에 진짜 원을 정확히 그리는 것은 힘듭니다.
     * 하지만, 여러 개의 짧은 직선을 그려서 직선들을 원 모양에 최대한 비슷하게 그릴 경우,
     * 직선의 수가 많아질수록 원에 가까워집니다.
     * 이와 마찬가지로, 테일러 급수도 '아주 작은 조각(항)을 더 많이 추가할수록 원래의 함수에 더 가까운 값'을 갖게 됩니다.
     *
     * <pre>
     * {@code
     * x - (x^3 / 3!) + (x^5 / 5!) - (x^7 / 7!) + ...
     * }
     * </pre>
     * <p>
     * 그리고 다음과 같이 계산됩니다.
     *
     * <pre>
     * {@code
     * // 실제로 sin(π)의 정확한 값은 수학적으로 0입니다.
     * // 하지만 테일러 급수로 근사하는 방법을 사용했기 때문에 미세한 오차가 생깁니다.
     * System.out.println(sin_1(Math.PI));
     * // 2.1142567558399565E-5
     * //
     * // 이를 지수 표기법(a * 10^b)로 표현하면 다음과 같습니다.
     * // = 2.1142567558399565 * 10^−5
     * // = 0.000021142567558399565
     *
     * System.out.printf("sin(π) ≈ %.10f\n", sin_1(Math.PI));
     * // sin(π) ≈ 0.0000211426
     * }
     * </pre>
     *
     * @param radians 사인 값을 구하고자 하는 각도(radians) 단위의 값(ex: {@link Math#PI}, {@link Math#PI}/2 등)입니다.
     * 라디안은 원의 반지름 길이에 대한 호의 길이 비율을 나타내는 각도 단위입니다.
     * 예를 들어, 180도는 {@link Math#PI} 라디안, 90도는 {@link Math#PI}/2 라디안으로 표현됩니다.
     *
     * @return 주어진 라디안 각도의 사인(sin) 근사값. 테일러 급수의 근사 특성상 아주 작은 오차를 가질 수 있습니다.
     * @see <a href="https://ko.wikipedia.org/wiki/%ED%85%8C%EC%9D%BC%EB%9F%AC_%EA%B8%89%EC%88%98">위키 - 테일러 급수</a>
     * @see <a href="https://namu.wiki/w/%ED%85%8C%EC%9D%BC%EB%9F%AC%20%EA%B8%89%EC%88%98">나무위키 - 테일러 급수</a>
     * @see <a href="https://darkpgmr.tistory.com/59">테일러 급수의 이해와 활용 (Taylor series)</a>
     * @see <a href="https://ko.wikipedia.org/wiki/%EB%9D%BC%EB%94%94%EC%95%88">라디안(radian)</a>
     */
    double sin_1(double radians) {
        // x^2
        double r2 = radians * radians;
        // x^3
        double r3 = r2 * radians;
        // x^5
        double r5 = r3 * r2;
        // x^7
        double r7 = r5 * r2;
        // x^9
        double r9 = r7 * r2;
        // x^11
        double r11 = r9 * r2;
        // x^13
        double r13 = r11 * r2;

        return (radians
            - (r3 / 6)
            + (r5 / 120)
            - (r7 / 5_040)
            + (r9 / 362_880)
            - (r11 / 39_916_800.0)
            + (r13 / 6_227_020_800.0));
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sines_1_refactoring() throws Exception {
        assertEquals(sin_1_refactored(0), 0, EPSILON);
        assertEquals(sin_1_refactored(Math.PI), 0, EPSILON);
        assertEquals(sin_1_refactored(Math.PI / 2), 1, EPSILON);
        assertEquals(sin_1_refactored(Math.PI / 3), 0.8660, EPSILON);
        assertEquals(sin_1_refactored(Math.PI / 4), 0.7071, EPSILON);
        assertEquals(sin_1_refactored(Math.PI / 5), 0.5877, EPSILON);
        // 하지만 현재 `sin_2`은 정확도에 한계가 있기에 개선이 필요합니다.
        // 그래서 책에서는 정확도를 높여서 체크하고 항(term)을 늘리는 방식의 코드(`sin_2`)를 제안합니다.
        assertEquals(sin_1_refactored(Math.PI / 3), 0.8660254038, EPSILON2);
        assertEquals(sin_1_refactored(Math.PI / 4), 0.7071067812, EPSILON2);
        assertEquals(sin_1_refactored(Math.PI / 5), 0.5877852523, EPSILON2);
        // 하지만, 정확도를 높인 경우에 대해 `sin_1`도 통과는 합니다.
        assertEquals(sin_1(Math.PI / 3), 0.8660254038, EPSILON2);
        assertEquals(sin_1(Math.PI / 4), 0.7071067812, EPSILON2);
        assertEquals(sin_1(Math.PI / 5), 0.5877852523, EPSILON2);
    }

    /**
     * 책에서는 원하는 정확도 범위 안으로 수렴할 때까지 테일러 급수의 항(term) 개수를 계속 늘려야 한다면서
     * 이 함수를 제안합니다.
     * <p>
     * 하지만 TDD는 사라졌고, 이 알고리즘이 정말로 잘 동작하는지는 확인할 방법이 떠오르지 않습니다.
     *
     * @param radians 사인 값을 구하고자 하는 각도(radians) 단위의 값(ex: {@link Math#PI}, {@link Math#PI}/2 등)입니다.
     * 라디안은 원의 반지름 길이에 대한 호의 길이 비율을 나타내는 각도 단위입니다.
     * 예를 들어, 180도는 {@link Math#PI} 라디안, 90도는 {@link Math#PI}/2 라디안으로 표현됩니다.
     *
     * @return 주어진 라디안 각도의 사인(sin) 근사값. 테일러 급수의 근사 특성상 아주 작은 오차를 가질 수 있습니다.
     */
    double sin_1_refactored(double radians) {
        double result = radians;
        double lastResult = 2;
        double m1 = -1;
        double sign = 1;
        double power = radians;
        double fac = 1;
        double r2 = radians * radians;
        int n = 1;
        BiFunction<Double, Double, Boolean> close = (a, b) -> Math.abs(a - b) < .000_0001;

        while (!close.apply(result, lastResult)) {
            lastResult = result;
            power *= r2;
            fac *= (n + 1) * (n + 2);
            n += 2;
            sign *= m1;
            double term = sign * power / fac;
            result += term;
        }

        return result;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sines_2_fail() throws Exception {
        assertEquals(sin_1_refactored(0), 0, EPSILON);
        assertEquals(sin_1_refactored(Math.PI), 0, EPSILON);
        assertEquals(sin_1_refactored(Math.PI / 2), 1, EPSILON);
        assertEquals(sin_1_refactored(Math.PI / 3), 0.8660, EPSILON);
        assertEquals(sin_1_refactored(Math.PI / 4), 0.7071, EPSILON);
        assertEquals(sin_1_refactored(Math.PI / 5), 0.5877, EPSILON);
        // 하지만 현재 `sin_2`은 정확도에 한계가 있기에 개선이 필요합니다.
        // 그래서 책에서는 정확도를 높여서 체크하고 항(term)을 늘리는 방식의 코드(`sin_2`)를 제안합니다.
        assertEquals(sin_1_refactored(Math.PI / 3), 0.8660254038, EPSILON2);
        assertEquals(sin_1_refactored(Math.PI / 4), 0.7071067812, EPSILON2);
        assertEquals(sin_1_refactored(Math.PI / 5), 0.5877852523, EPSILON2);
        // 테스트를 몇 개 더 작성해볼 수 있습니다.
        assertEquals(sin_1_refactored(Math.PI * 2), 0, EPSILON2);
        // 이 테스트는 매우 근소한 차이(4.613017243966015E-9)로 실패합니다.
        // `close` 함수에서 비교하는 값을 줄여서 고칠 수도 있겠지만,
        // 입력값이 100 PI, 1000 PI가 되면 다시 문제가 될 수 있습니다.
        // assertEquals(sin_2(Math.PI * 3), 0, EPSILON2); // 다음 코드 진행 위해 주석 처리
        //Expected :0.0
        //Actual   :4.613017243966015E-9
    }

    /**
     * @throws Exception
     */
    @Test
    public void sines_2_success() throws Exception {
        assertEquals(sin_2(0), 0, EPSILON);
        assertEquals(sin_2(Math.PI), 0, EPSILON);
        assertEquals(sin_2(Math.PI / 2), 1, EPSILON);
        assertEquals(sin_2(Math.PI / 3), 0.8660, EPSILON);
        assertEquals(sin_2(Math.PI / 4), 0.7071, EPSILON);
        assertEquals(sin_2(Math.PI / 5), 0.5877, EPSILON);
        // 하지만 현재 `sin_2`은 정확도에 한계가 있기에 개선이 필요합니다.
        // 그래서 책에서는 정확도를 높여서 체크하고 항(term)을 늘리는 방식의 코드(`sin_2`)를 제안합니다.
        assertEquals(sin_2(Math.PI / 3), 0.8660254038, EPSILON2);
        assertEquals(sin_2(Math.PI / 4), 0.7071067812, EPSILON2);
        assertEquals(sin_2(Math.PI / 5), 0.5877852523, EPSILON2);
        assertEquals(sin_2(Math.PI * 2), 0, EPSILON2);
        assertEquals(sin_2(Math.PI * 3), 0, EPSILON2);
    }

    /**
     * @param radians 사인 값을 구하고자 하는 각도(radians) 단위의 값(ex: {@link Math#PI}, {@link Math#PI}/2 등)입니다.
     * 라디안은 원의 반지름 길이에 대한 호의 길이 비율을 나타내는 각도 단위입니다.
     * 예를 들어, 180도는 {@link Math#PI} 라디안, 90도는 {@link Math#PI}/2 라디안으로 표현됩니다.
     * <p>
     * {@link SineTest#sines_2_fail()}의 실패 케이스를 통과시키기 위해 각도를 0에서 2 PI 사이가 되도록 줄입니다.
     *
     * @return 주어진 라디안 각도의 사인(sin) 근사값. 테일러 급수의 근사 특성상 아주 작은 오차를 가질 수 있습니다.
     */
    double sin_2(double radians) {
        radians %= 2 * Math.PI;
        double result = radians;
        double lastResult = 2;
        double m1 = -1;
        double sign = 1;
        double power = radians;
        double fac = 1;
        double r2 = radians * radians;
        int n = 1;
        BiFunction<Double, Double, Boolean> close = (a, b) -> Math.abs(a - b) < .000_0001;

        while (!close.apply(result, lastResult)) {
            lastResult = result;
            power *= r2;
            fac *= (n + 1) * (n + 2);
            n += 2;
            sign *= m1;
            double term = sign * power / fac;
            result += term;
        }

        return result;
    }

    /**
     * @throws Exception
     */
    @Test
    public void sines_3_success() throws Exception {
        assertEquals(sin_2(0), 0, EPSILON);
        assertEquals(sin_2(Math.PI), 0, EPSILON);
        assertEquals(sin_2(Math.PI / 2), 1, EPSILON);
        assertEquals(sin_2(Math.PI / 3), 0.8660, EPSILON);
        assertEquals(sin_2(Math.PI / 4), 0.7071, EPSILON);
        assertEquals(sin_2(Math.PI / 5), 0.5877, EPSILON);
        // 하지만 현재 `sin_2`은 정확도에 한계가 있기에 개선이 필요합니다.
        // 그래서 책에서는 정확도를 높여서 체크하고 항(term)을 늘리는 방식의 코드(`sin_2`)를 제안합니다.
        assertEquals(sin_2(Math.PI / 3), 0.8660254038, EPSILON2);
        assertEquals(sin_2(Math.PI / 4), 0.7071067812, EPSILON2);
        assertEquals(sin_2(Math.PI / 5), 0.5877852523, EPSILON2);
        assertEquals(sin_2(Math.PI * 2), 0, EPSILON2);
        assertEquals(sin_2(Math.PI * 3), 0, EPSILON2);
        // 음수 케이스도 추가해 봅니다.
        assertEquals(sin_2(-Math.PI), 0, EPSILON2);
        assertEquals(sin_2(-Math.PI / 2), -1, EPSILON2);
        assertEquals(sin_2(-3 * Math.PI / 2), 1, EPSILON2);
        assertEquals(sin_2(-1000 * Math.PI), 0, EPSILON2);
    }
}
