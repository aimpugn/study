# Bankers Rounding

## 1. 정의와 원리

Bankers Rounding은 반올림 시 정확히 중간값(.5)일 경우 가장 가까운 짝수로 반올림하는 방식입니다.
이 방식은 "round-to-even" 또는 "round-half-to-even"이라고도 불립니다.

## 2. 작동 방식

- 일반적인 반올림: 3.4 → 3, 3.6 → 4
- 중간값(.5)의 처리:
    - 짝수로 끝나는 경우: 2.5 → 2, 4.5 → 4
    - 홀수로 끝나는 경우: 3.5 → 4, 5.5 → 6

## 3. 수학적 특성

- 통계적으로 중립적: 대량의 데이터를 처리할 때, 상향 반올림과 하향 반올림의 빈도가 거의 동일해집니다.
- 기대값 보존: 반올림 전후의 평균값 변화를 최소화합니다.

## 4. 장점

- 편향 감소: 항상 올리거나 내리는 방식에 비해 전체적인 편향을 줄입니다.
- 누적 오차 최소화: 대량의 계산에서 발생할 수 있는 누적 오차를 줄입니다.
- 공정성: 금융 거래에서 은행과 고객 모두에게 공정한 결과를 제공합니다.

## 5. 실제 적용 사례

- 은행 이자 계산: 일일 이자를 계산하고 합산할 때 사용됩니다.
- 회계 소프트웨어: 재무제표 작성 시 여러 항목의 반올림에 사용됩니다.
- 과학적 계산: 측정값의 정확도를 유지하면서 반올림할 때 사용됩니다.
- 통계 분석: 대규모 데이터셋의 처리 시 편향을 줄이는 데 활용됩니다.

## 6. 구체적인 예시

   은행 이자 계산의 경우:
- 일일 이자가 $3.245인 경우: 3.25로 반올림
- 일일 이자가 $3.255인 경우: 3.26으로 반올림
   이렇게 하면 장기적으로 은행과 고객 모두에게 공정한 결과를 제공합니다.

## 7. 다른 반올림 방식과의 비교

- 일반적인 반올림 (half-up): 항상 .5를 올림하므로 약간의 상향 편향이 있습니다.
- Half-down: 항상 .5를 내림하므로 약간의 하향 편향이 있습니다.
- Bankers Rounding: .5를 짝수로 반올림하여 편향을 최소화합니다.

## 8. 주의점

- 단일 계산에서는 차이가 미미할 수 있지만, 대량의 거래나 장기간의 계산에서 중요한 영향을 미칠 수 있습니다.
- 모든 상황에 적합한 것은 아닙니다. 예를 들어, 항상 고객에게 유리한 방향으로 반올림해야 하는 경우에는 적합하지 않을 수 있습니다.

## 9. 프로그래밍 구현

PHP에서는 `round()` 함수의 mode 매개변수로 `PHP_ROUND_HALF_EVEN`을 지정하여 사용할 수 있습니다.

```php
$value = 2.5;
echo round($value, 0, PHP_ROUND_HALF_EVEN); // 출력: 2

$value = 3.5;
echo round($value, 0, PHP_ROUND_HALF_EVEN); // 출력: 4
```

Bankers Rounding은 특히 금융 및 회계 분야에서 중요하게 사용되는 방식입니다. 이 방식은 대량의 거래를 처리할 때 공정성을 유지하고 누적 오차를 최소화하는 데 효과적이므로, 많은 금융 기관과 회계 시스템에서 채택하고 있습니다.

## Is "banker's rounding" really more numerically stable?

Yes! It really is more numerically stable.

For the case that you're looking at, the numbers `[0.0, 0.1, ..., 0.9]`, note that under round-ties-to-away,
- only *four* of those numbers are rounding down (0.1 through 0.4),
- five are rounded up,
- and one (`0.0`) is unchanged by the rounding operation,
- and then of course that pattern repeats for `1.0` through `1.9`, `2.0` through `2.9`, etc.

So on average, more values are rounded away from zero than towards it.
But under `round-ties-to-even`, we'd get:

- `[0.0, 0.9]`:
    - five values rounding down
    - and four rounding up in
- `[1.0, 1.9]`
    - four values rounding down
    - and five rounding up in

and so on.

On average, we get the same number of values rounding up as rounding down.
More importantly, the expected error introduced by the rounding is (under suitable assumptions on the distribution of the inputs) closer to zero.

Here's a quick demonstration using Python.
To avoid difficulties due to Python 2 / Python 3 differences in the builtin `round` function, we give two Python-version-agnostic rounding functions:

```py
def round_ties_to_even(x):
    """
    Round a float x to the nearest integer, rounding ties to even.
    """
    if x < 0:
        return -round_ties_to_even(-x)  # use symmetry
    int_part, frac_part = divmod(x, 1)
    return int(int_part) + (
        frac_part > 0.5
        or (frac_part == 0.5 and int_part % 2.0 == 1.0))

def round_ties_away_from_zero(x):
    """
    Round a float x to the nearest integer, rounding ties away from zero.
    """
    if x < 0:
        return -round_ties_away_from_zero(-x)  # use symmetry
    int_part, frac_part = divmod(x, 1)
    return int(int_part) + (frac_part >= 0.5)
```

```py
test_values = [n / 10.0 for n in range(500, 1001)]
errors_even = [round_ties_to_even(value) - value for value in test_values]
errors_away = [round_ties_away_from_zero(value) - value for value in test_values]

import statistics
statistics.mean(errors_even), statistics.stdev(errors_even)
(0.0, 0.2915475947422656)
statistics.mean(errors_away), statistics.stdev(errors_away)
(0.0499001996007984, 0.28723681870533313)
```

The key point here is that `errors_even` has zero mean: the average error is zero.
But `errors_away` has positive mean: the average error is biased away from zero.

### A more realistic example

Here's a semi-realistic example that demonstrates the bias from `round-ties-away-from-zero` in a numerical algorithm. We're going to compute the sum of a list of floating-point numbers, using the [pairwise summation](https://en.wikipedia.org/wiki/Pairwise_summation) algorithm. This algorithm breaks the sum to be computed into two roughly equal parts, recursively sums those two parts, then adds the results. It's substantially more accurate than a naive sum, but typically not as good as more sophisticated algorithms like [Kahan summation](https://en.wikipedia.org/wiki/Kahan_summation_algorithm). It's the algorithm that's used by NumPy's sum function. Here's a simple Python implementation.

```py
import operator

def pairwise_sum(xs, i, j, add=operator.add):
    """
    Return the sum of floats xs[i:j] (0 <= i <= j <= len(xs)),
    using pairwise summation.
    """
    count = j - i
    if count >= 2:
        k = (i + j) // 2
        return add(pairwise_sum(xs, i, k, add),
                   pairwise_sum(xs, k, j, add))
    elif count == 1:
        return xs[i]
    else:  # count == 0
        return 0.0
```

We've included a parameter `add` to the function above, representing the operation to be used for addition.
By default, it uses Python's normal addition algorithm, which on a typical machine will resolve to the standard *IEEE 754* addition, using `round-ties-to-even` rounding mode.

We want to look at the expected error from the `pairwise_sum` function, using both standard addition and using a `round-ties-away-from-zero` version of addition.
Our first problem is that we don't have an easy and portable way to change the hardware's rounding mode from within Python, and a software implementation of binary floating-point would be large and slow.
Fortunately, there's a trick we can use to get `round-ties-away-from-zero` while still using the hardware floating-point.
For the first part of that trick, we can employ Knuth's "2Sum" algorithm to add two floats and obtain the correctly-rounded sum together with the exact error in that sum:

```py
def exact_add(a, b):
    """
    Add floats a and b, giving a correctly rounded sum and exact error.

    Mathematically, a + b is exactly equal to sum + error.
    """
    # This is Knuth's 2Sum algorithm. See section 4.3.2 of the Handbook
    # of Floating-Point Arithmetic for exposition and proof.
    sum = a + b
    bv = sum - a
    error = (a - (sum - bv)) + (b - bv)
    return sum, error
```

With this in hand, we can easily use the error term to determine when the exact sum is a tie. We have a tie if and only if `error` is nonzero and `sum + 2 * error` is exactly representable, and in that case `sum` and `sum + 2 * error` are the two floats nearest that tie.
Using this idea, here's a function that adds two numbers and gives a correctly rounded result, but rounds ties *away* from zero.

```py
def add_ties_away(a, b):
    """
    Return the sum of a and b. Ties are rounded away from zero.
    """
    sum, error = exact_add(a, b)
    sum2, error2 = exact_add(sum, 2.0*error)
    if error2 or not error:
        # Not a tie.
        return sum
    else:
        # Tie. Choose the larger of sum and sum2 in absolute value.
        return max([sum, sum2], key=abs)
```

Now we can compare results. `sample_sum_errors` is a function that generates a list of floats in the range `[1, 2]`, adds them using both normal `round-ties-to-even` addition and our custom `round-ties-away-from-zero` version, compares with the exact sum and returns the errors for both versions, measured in units in the last place.

```py
import fractions
import random

def sample_sum_errors(sample_size=1024):
    """
    Generate `sample_size` floats in the range [1.0, 2.0], sum
    using both addition methods, and return the two errors in ulps.
    """
    xs = [random.uniform(1.0, 2.0) for _ in range(sample_size)]
    to_even_sum = pairwise_sum(xs, 0, len(xs))
    to_away_sum = pairwise_sum(xs, 0, len(xs), add=add_ties_away)

    # Assuming IEEE 754, each value in xs becomes an integer when
    # scaled by 2**52; use this to compute an exact sum as a Fraction.
    common_denominator = 2**52
    exact_sum = fractions.Fraction(
        sum(int(m*common_denominator) for m in xs),
        common_denominator)

    # Result will be in [1024, 2048]; 1 ulp in this range is 2**-44.
    ulp = 2**-44
    to_even_error = (fractions.Fraction(to_even_sum) - exact_sum) / ulp
    to_away_error = (fractions.Fraction(to_away_sum) - exact_sum) / ulp

    return to_even_error, to_away_error
```

Here's an example run:

```py
>>> sample_sum_errors()
(1.6015625, 9.6015625)
```

So that's an error of 1.6 ulps using the standard addition, and an error of 9.6 ulps when rounding ties away from zero. It certainly *looks* as though the `ties-away-from-zero` method is worse, but a single run isn't particularly convincing. Let's do this 10000 times, with a different random sample each time, and plot the errors we get. Here's the code:

```py
import statistics
import numpy as np
import matplotlib.pyplot as plt

def show_error_distributions():
    errors = [sample_sum_errors() for _ in range(10000)]
    to_even_errors, to_away_errors = zip(*errors)
    print("Errors from ties-to-even: "
          "mean {:.2f} ulps, stdev {:.2f} ulps".format(
              statistics.mean(to_even_errors),
              statistics.stdev(to_even_errors)))
    print("Errors from ties-away-from-zero: "
          "mean {:.2f} ulps, stdev {:.2f} ulps".format(
              statistics.mean(to_away_errors),
              statistics.stdev(to_away_errors)))

    ax1 = plt.subplot(2, 1, 1)
    plt.hist(to_even_errors, bins=np.arange(-7, 17, 0.5))
    ax2 = plt.subplot(2, 1, 2)
    plt.hist(to_away_errors, bins=np.arange(-7, 17, 0.5))
    ax1.set_title("Errors from ties-to-even (ulps)")
    ax2.set_title("Errors from ties-away-from-zero (ulps)")
    ax1.xaxis.set_visible(False)
    plt.show()
```

When I run the above function on my machine, I see:

```bash
Errors from ties-to-even: mean 0.00 ulps, stdev 1.81 ulps
Errors from ties-away-from-zero: mean 9.76 ulps, stdev 1.40 ulps
```

I was planning to go one step further and perform a statistical test for bias on the two samples, but the bias from the `ties-away-from-zero` method is so marked that that looks unnecessary.
Interestingly, while the `ties-away-from-zero` method gives poorer result, it does give a smaller spread of errors.

## [은행가의 반올림은 무죄인가?](https://www.officetutor.co.kr/board/faq_lib/frm_vba_content.asp?page=2&idx=20)

반올림문제가 그리 큰 문제가 아닌 것처럼 보이지만 수 백억 원이 오고 가는 금융권에서는 이것으로 인해 적지 않은 금액의 계산이 달라집니다.

대개의 프로그래밍 언어들은 반올림의 방법으로서 “Banker’s rounding”를 사용합니다.

많은 사람들이 이 반올림 방법이 상식적으로 틀린 결과를 돌려주기 때문에 싫어하지만 아이러니컬하게도 이것은 가장 정확한 라운딩(rounding)방법으로서 개발된 것입니다.

- 0.5이하는 버리고 0.5이상이면 올립니다.
- 그리고 정확하게 0.5이면 가장 가까운 짝수로 올립니다.

가령 12.5에서 0.5는 버려지고 12로 만들지만 13.5는 0.5를 더하여 14가 됩니다.

Bankers rounding은 Gauss법을 사용하는 것으로 이는 0.5인 경우 2로 나누어질 수 있는 가장 가까운 수로 반올림 한다는 것입니다.

1.5 is rounded to 2
2.5 is rounded to 2
3.5 is rounded to 4
4.5 is rounded to 4

종종 이러한 방법이 일상적인 반올림(Standard Rounding) 상식과 어긋나지만 이 방법은 다음과 같은 정당성을 가집니다.

가령 12.0부 터 13.0사이를 0.1씩의 간격으로 나누면 9개의 값이 들어가고, 이 값들은 반올림의 대상이 됩니다.

```plaintext
12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7, 12.8, 12.9
```

상식적인 반올림이라면 9개의 숫자 중 5개는 올리고 4개는 버리게 됩니다.
- round up: 12.5, 12.6, 12.7, 12.8, 12.9
- round down: 12.1, 12.2, 12.3, 12.4

그러나 이 방법은 공평하지 않다.

1/9만큼 한쪽은 더 가지고 한쪽은 부족하게 됩니다.

그러나 0.5에서 가장 가까운 짝수로 옮기도록 하게 되면 어떻게 될까요?

12.0 부터 14.0까지 18개의 반올림 대상이 생기고 버리는 쪽이나 올리는 쪽 모두 9개의 숫자를 나누어 갖게 됩니다.

```plaintext
12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7, 12.8, 12.9
13.1, 13.2, 13.3, 13.4, 13.5, 13.6, 13.7, 13.8, 13.9
```

- 12로 옮기는 경우: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7, 12.8, 12.9
- 14로 옮기는 경우: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6, 13.7, 13.8, 13.9

따라서 한쪽에 치우치지 않는 공평한 셈이 됩니다.

## 기타

- [Is "banker's rounding" really more numerically stable?](https://stackoverflow.com/a/45245802)
