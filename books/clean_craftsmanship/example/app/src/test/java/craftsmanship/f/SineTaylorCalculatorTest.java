package craftsmanship.f;

import org.testng.annotations.Test;

import java.util.stream.IntStream;

import static org.assertj.core.api.Assertions.assertThat;
import static org.testng.AssertJUnit.assertEquals;
import static org.testng.AssertJUnit.assertTrue;

public class SineTaylorCalculatorTest {
    private static final double EPSILON = 0.000_000_001;

    @Test
    public void taylorTerms() throws Exception {
        SineTaylorCalculatorSpy spy = new SineTaylorCalculatorSpy();
        // 무작위 숫자와 적절한 n 값들을 사용하기 때문에, 특정한 값을 사용하지 않아도 됩니다.
        double r = Math.random() * Math.PI;
        for (int n = 1; n <= 10; n++) {
            spy.calculateTerm(r, n);
            // 여기서 실제 계산 함수들을 테스트하지는 않습니다.
            // 실제 계산 함수들의 코드는 꽤 간단하고, 어쩌면 테스트가 필요 없을 수도 있습니다.
            assertThat(spy.getSignPower()).isEqualTo(n - 1);
            assertEquals(r, spy.getR(), EPSILON);
            assertEquals(2 * n - 1, spy.getRPower());
            assertEquals(2 * n - 1, spy.getFac());
        }
    }

    // taylorTerms 테스트가 통과하면, 남은 급수의 합을 구하는 코드 `SineTaylorCalculator_1#sign`을 작성합니다.
    // 하지만 다음과 같은 의문들이 들 수 있습니다:
    // - 제대로 동작한다고 믿는지
    // - `calculatorTerm` 함수가 정확한 테일러 항을 적절하게 계산하지는지
    // - `sin` 함수가 각 항의 합을 잘 구하는지
    // - 충분한 개수의 항을 계산하는지
    // - 맨 처음에 작성한 값 기반 테스트 없이 어떻게 검증할 수 있는지 등

    /**
     * 모든 `sin(r)`의 값은 -1보다 크도 1보다 작아야 합니다.
     * <p>
     * `PI / 2` 또는 `-PI / 2`처럼 `sin`값이 -1이나 1이 되는 경우도 있지만,
     * 근사치 계산이므로 딱 떨어지지는 않을 것입니다.
     *
     * @throws Exception
     */
    @Test
    public void testSineInRange() throws Exception {
        SineTaylorCalculator_1 c = new SineTaylorCalculator_1();
        IntStream.range(0, 100)
            .forEach((idx) -> {
                var r = (Math.random() * 4 * Math.PI) - (2 * Math.PI);
                var sin_r = c.sin(r);
                assertTrue(sin_r < 1 && sin_r > -1);
            });
    }

    /**
     * 이 테스트는 실패하다가 테일러 급수 항(term)의 수를 20개까지 늘리면 통과합니다.
     *
     * @throws Exception
     */
    @Test
    public void PythagoreanIdentity() throws Exception {
        var c = new SineTaylorCalculator_2();
        IntStream.range(0, 100).forEach((idx) -> {
            var r = (Math.random() * 4 * Math.PI) - (2 * Math.PI);
            var sin_r = c.sin(r);
            var cos_r = c.cos(r);

            assertEquals(1.0, sin_r * sin_r + cos_r * cos_r, EPSILON);
            // `SineTaylorCalculator_2#sin`에서 10개의 항을 계산하는 경우
            //Expected :1.0
            //Actual   :0.999999856381723
            // 하지만, `SineTaylorCalculator_2#sin`에서 20개의 항을 계산하는 경우 성공합니다.
        });
    }

    // 이렇게 테스트한 후에 얼마나 확신이 들까?
    // 저자는 꽤 확신이 든다고 합니다.
    // 하지만 항의 개수를 10개에서 20개로 늘려야 테스트가 통과한다면,
    // 20개에서 100개로 늘려야 통과하는 극단적인 경우도 있을 텐데, 정말 괜찮은 걸까? 라는 생각이 듭니다.
}

class SineTaylorCalculatorSpy extends SineTaylorCalculator_0 {
    private int fac_n;
    private double power_r;
    private int power_n;
    private int sing_n;

    public double getR() {
        return power_r;
    }

    public int getRPower() {
        return power_n;
    }

    public int getFac() {
        return fac_n;
    }

    public int getSignPower() {
        return sing_n;
    }

    @Override
    protected double calcFactorial(int n) {
        fac_n = n;
        return super.calcFactorial(n);
    }

    @Override
    protected double calcPower(double r, int n) {
        power_r = r;
        power_n = n;
        return super.calcPower(r, n);
    }

    @Override
    protected int calcSign(int n) {
        sing_n = n;
        return super.calcSign(n);
    }

    @Override
    public double calculateTerm(double r, int n) {
        return super.calculateTerm(r, n);
    }
}
