package craftsmanship.f;

public class SineTaylorCalculator_1 {
    public double sin(double r) {
        double sin = 0;
        for (int i = 1; i < 10; i++) {
            sin += calculateTerm(r, i);
        }

        return sin;
    }

    public double calculateTerm(double r, int n) {
        int sign = calcSign(n - 1);
        double power = calcPower(r, 2 * n - 1);
        double factorial = calcFactorial(2 * n - 1);

        return sign * power / factorial;
    }

    protected int calcSign(int n) {
        int sign = 1;
        for (int i = 0; i < n; i++) {
            sign *= -1;
        }
        return sign;
    }

    protected double calcPower(double r, int n) {
        double power = 1;
        for (int i = 0; i < n; i++) {
            power *= r;
        }
        return power;
    }

    protected double calcFactorial(int n) {
        double fac = 1;
        for (int i = 1; i <= n; i++) {
            fac *= i;
        }
        return fac;
    }
}
