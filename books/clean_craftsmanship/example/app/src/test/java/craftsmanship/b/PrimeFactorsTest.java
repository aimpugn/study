package craftsmanship.b;

import org.testng.annotations.Test;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;


/**
 * 소인수 분해 TDD 과정 정리
 *
 * @see <a href="https://www.jetbrains.com/help/idea/testng.html">IntelliJ IDEA - TestNG</a>
 */
@Test
public class PrimeFactorsTest {

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * 이 테스트는 실패합니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_0() throws Exception {
        // assertEquals(new ArrayList<>(), factorsOf_0(1)); // 다음 코드 진행 위해 주석 처리
    }

    /**
     * 편의를 위해 테스트 대상인 메서드를 테스트 클래스 안에 구현합니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_0(int i) {
        return null;
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_1() throws Exception {
        assertThat(factorsOf_1(1)).isEqualTo(Collections.emptyList());
    }

    /**
     * 편의를 위해 테스트 대상인 메서드를 테스트 클래스 안에 구현합니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_1(int i) {
        return new ArrayList<>();
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."을 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_2() throws Exception {
        // assertThat(factorsOf_1(1)).contains(2); // 다음 코드 진행 위해 주석 처리
        // 이 테스트는 실패합니다.
        //Expecting ArrayList:
        //  []
        //to contain:
        //  [2]
        //but could not find the following element(s):
        //  [2]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_3() throws Exception {
        assertThat(factorsOf_3(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_3(2))
                .contains(2)
                .size().isEqualTo(1);
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_3(int i) {
        // 규칙 5: 가능하면 일반화하라.
        // 1. `new ArrayList<>()`를 `factors`라는 변수로 추출합니다.
        //    원래 `factorsOf_1`에서 반환하던 상수 `new ArrayList<>()`는 매우 구체적입니다.
        //
        //    이 값을 조작이 가능한 변수에 넣음으로써 '일반화'할 수 있습니다.
        //    작은 일반화지만, 이런 작은 일반화로 충분할 때가 많습니다.
        //
        //    참고로 리팩터링에서의 일반화는 "코드의 구체성(concrete)을 제거하고 좀 더 유연하게 만드는 과정"을 의미합니다.
        //    여기서 변수로 추출한 것은 단순히 문법적 변화가 아니라, "코드를 확장 가능한 구조로 만들었다는 것"을 의미합니다.
        List<Integer> factors = new ArrayList<>();
        // 2. `if` 조건문을 추가합니다.
        if (i > 1) {
            factors.add(2);
        }

        return factors;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_4() throws Exception {
        assertThat(factorsOf_3(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_3(2))
                .contains(2)
                .size().isEqualTo(1);
        // assertThat(factorsOf_3(3)).contains(3); // 다음 코드 진행 위해 주석 처리
        //Expecting ArrayList:
        //  [2]
        //to contain:
        //  [3]
        //but could not find the following element(s):
        //  [3]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_5() throws Exception {
        assertThat(factorsOf_5(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_5(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_5(3))
                .contains(3)
                .size().isEqualTo(1);
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_5(int i) {
        List<Integer> factors = new ArrayList<>();
        if (i > 1) {
            // `factorsOf_3`에서는 상수 `2`를 사용했지만, 이를 변수 `i`로 교체하는 간단한 일반화를 합니다.
            // 이를 통해 기존 테스트는 물론 새로운 테스트 `factors_5`도 통과합니다.
            factors.add(i);
        }

        return factors;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_6() throws Exception {
        assertThat(factorsOf_5(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_5(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_5(3))
                .contains(3)
                .size().isEqualTo(1);
        // assertThat(factorsOf_5(4)).contains(2, 2); // 다음 코드 진행 위해 주석 처리
        //Expecting ArrayList:
        //  [4]
        //to contain:
        //  [2, 2]
        //but could not find the following element(s):
        //  [2]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_7() throws Exception {
        assertThat(factorsOf_7(1)).isEqualTo(Collections.emptyList());
        // `factorsOf_7` 구현의 경우 2에 대해 [1, 2] 두 개의 요소를 갖는 리스트가 반환됩니다.
        // assertThat(factorsOf_7(2))
        //         .contains(2)
        //         .size().isEqualTo(1); // 다음 코드 진행 위해 주석 처리
        assertThat(factorsOf_7(3))
                .contains(3)
                .size().isEqualTo(1);
        assertThat(factorsOf_7(4))
                .contains(2, 2)
                .size().isEqualTo(2);
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_7(int i) {
        List<Integer> factors = new ArrayList<>();
        if (i > 1) {
            // `i`가 2로 나뉘는지 검사합니다.
            if (i % 2 == 0) {
                factors.add(2);
                i /= 2;
            }
            factors.add(i);
        }

        return factors;
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_8() throws Exception {
        assertThat(factorsOf_8(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_8(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_8(3))
                .contains(3)
                .size().isEqualTo(1);
        assertThat(factorsOf_8(4))
                .contains(2, 2)
                .size().isEqualTo(2);
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_8(int i) {
        List<Integer> factors = new ArrayList<>();
        if (i > 1) {
            // `i`가 2로 나뉘는지 검사합니다.
            if (i % 2 == 0) {
                factors.add(2);
                i /= 2;
            }
            if (i > 1) {
                factors.add(i);
            }
        }

        return factors;
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_9() throws Exception {
        assertThat(factorsOf_9(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_9(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_9(3))
                .contains(3)
                .size().isEqualTo(1);
        assertThat(factorsOf_9(4))
                .contains(2, 2)
                .size().isEqualTo(2);
        // 테스트를 추가합니다.
        assertThat(factorsOf_9(5))
                .contains(5)
                .size().isEqualTo(1);
        assertThat(factorsOf_9(6))
                .contains(2, 3)
                .size().isEqualTo(2);
        assertThat(factorsOf_9(7))
                .contains(7)
                .size().isEqualTo(1);
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_9(int i) {
        List<Integer> factors = new ArrayList<>();
        // if (i > 1) { // 일반화의 단서: 동일한 두 if 조건절
        //     if (i % 2 == 0) {
        //         factors.add(2);
        //         i /= 2;
        //     }
        //     if (i > 1) { // 일반화의 단서: 동일한 두 if 조건절
        //         factors.add(i);
        //     }
        // }
        if (i > 1) {
            if (i % 2 == 0) {
                factors.add(2);
                i /= 2;
            }
        }

        // 두 번째 if 문이 첫 번째 if 블록 안에 있어야 할 이유가 없으므로,
        // 두 번째 if 문을 분리합니다.
        if (i > 1) {
            factors.add(i);
        }

        return factors;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_10() throws Exception {
        assertThat(factorsOf_9(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_9(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_9(3))
                .contains(3)
                .size().isEqualTo(1);
        assertThat(factorsOf_9(4))
                .contains(2, 2)
                .size().isEqualTo(2);
        assertThat(factorsOf_9(5))
                .contains(5)
                .size().isEqualTo(1);
        assertThat(factorsOf_9(6))
                .contains(2, 3)
                .size().isEqualTo(2);
        assertThat(factorsOf_9(7))
                .contains(7)
                .size().isEqualTo(1);
        // 실패하는 테스트
        // assertThat(factorsOf_9(8))
        //         .contains(2, 2, 2)
        //         .size().isEqualTo(3); // 다음 코드 진행 위해 주석 처리
        //Expected :3
        //Actual   :2
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_11() throws Exception {
        assertThat(factorsOf_11(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_11(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_11(3))
                .contains(3)
                .size().isEqualTo(1);
        assertThat(factorsOf_11(4))
                .contains(2, 2)
                .size().isEqualTo(2);
        assertThat(factorsOf_11(5))
                .contains(5)
                .size().isEqualTo(1);
        assertThat(factorsOf_11(6))
                .contains(2, 3)
                .size().isEqualTo(2);
        assertThat(factorsOf_11(7))
                .contains(7)
                .size().isEqualTo(1);
        assertThat(factorsOf_11(8))
                .contains(2, 2, 2)
                .size().isEqualTo(3);
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * 규칙 5 "가능하면 일반화하라."를 따릅니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_11(int i) {
        List<Integer> factors = new ArrayList<>();
        if (i > 1) {
            // 규칙 5: 가능하면 일반화하라.
            // `if` 조건문을 `while`로 바꿉니다. 이를 통해 다음 사실을 알 수 있습니다:
            // - `while`은 `if`의 일반적인 형태고,
            // - `if`는 `while`의 퇴화한(degenerate) 형태입니다.
            while (i % 2 == 0) {
                factors.add(2);
                i /= 2;
            }
        }

        if (i > 1) {
            factors.add(i);
        }

        return factors;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_12() throws Exception {
        assertThat(factorsOf_11(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_11(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_11(3))
                .contains(3)
                .size().isEqualTo(1);
        assertThat(factorsOf_11(4))
                .contains(2, 2)
                .size().isEqualTo(2);
        assertThat(factorsOf_11(5))
                .contains(5)
                .size().isEqualTo(1);
        assertThat(factorsOf_11(6))
                .contains(2, 3)
                .size().isEqualTo(2);
        assertThat(factorsOf_11(7))
                .contains(7)
                .size().isEqualTo(1);
        assertThat(factorsOf_11(8))
                .contains(2, 2, 2)
                .size().isEqualTo(3);
        // 실패하는 테스트
        // assertThat(factorsOf_10(9))
        //         .contains(3, 3)
        //         .size().isEqualTo(2); // 다음 코드 진행 위해 주석 처리
        //Expecting ArrayList:
        //  [9]
        //to contain:
        //  [3, 3]
        //but could not find the following element(s):
        //  [3]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_13() throws Exception {
        assertThat(factorsOf_13(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_13(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_13(3))
                .contains(3)
                .size().isEqualTo(1);
        assertThat(factorsOf_13(4))
                .contains(2, 2)
                .size().isEqualTo(2);
        assertThat(factorsOf_13(5))
                .contains(5)
                .size().isEqualTo(1);
        assertThat(factorsOf_13(6))
                .contains(2, 3)
                .size().isEqualTo(2);
        assertThat(factorsOf_13(7))
                .contains(7)
                .size().isEqualTo(1);
        assertThat(factorsOf_13(8))
                .contains(2, 2, 2)
                .size().isEqualTo(3);
        assertThat(factorsOf_13(9))
                .contains(3, 3)
                .size().isEqualTo(2);
    }

    /**
     * 두 번째 법칙 "테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_13(int i) {
        List<Integer> factors = new ArrayList<>();
        if (i > 1) {
            while (i % 2 == 0) {
                factors.add(2);
                i /= 2;
            }
            // 아래 테스트 케이스를 통과하려면 `3`으로 나눠야 합니다.
            // ```
            // assertThat(factorsOf_11(9))
            //                .contains(3, 3)
            //                .size().isEqualTo(2)
            // ```
            // 하지만 규칙 5 "가능하면 일반화하라."를 총체적으로 위반할 뿐만 아니라, 코드 중복도 많아집니다.
            // 이때 일반화 주문(mantra)가 등장합니다:
            // "테스트가 더 구체적으로 바뀔수록 제품 코드는 더 일반적으로 바뀐다."
            while (i % 3 == 0) {
                factors.add(3);
                i /= 3;
            }
        }

        if (i > 1) {
            factors.add(i);
        }

        return factors;
    }

    /**
     * 규칙 5 "가능하면 일반화하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_14() throws Exception {
        assertThat(factorsOf_14(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_14(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_14(3))
                .contains(3)
                .size().isEqualTo(1);
        assertThat(factorsOf_14(4))
                .contains(2, 2)
                .size().isEqualTo(2);
        assertThat(factorsOf_14(5))
                .contains(5)
                .size().isEqualTo(1);
        assertThat(factorsOf_14(6))
                .contains(2, 3)
                .size().isEqualTo(2);
        assertThat(factorsOf_14(7))
                .contains(7)
                .size().isEqualTo(1);
        assertThat(factorsOf_14(8))
                .contains(2, 2, 2)
                .size().isEqualTo(3);
        assertThat(factorsOf_14(9))
                .contains(3, 3)
                .size().isEqualTo(2);
    }

    /**
     * 규칙 5 "가능하면 일반화하라."를 따릅니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_14(int i) {
        List<Integer> factors = new ArrayList<>();
        // int divisor = 2; <--------- 3. `divisor` 변수 초기화 부분을 `if` 블록 밖으로 옮깁니다.
        // if (i > 1) { <------------- 4. `if`문을 `while`문으로 바꿉니다.
        //     while (i % 2 == 0) { <- 1. 3개의 2를 divisor 변수로 추출합니다.
        //         factors.add(2); <-- 1. 3개의 2를 divisor 변수로 추출합니다.
        //         i /= 2; <---------- 1. 3개의 2를 divisor 변수로 추출합니다.
        //     }
        //     divisor++; <----------- 2. `divisor++`를 추가합니다.
        //     // 삭제
        //     // while (i % 3 == 0) {
        //     //     factors.add(3);
        //     //     i /= 3;
        //     // }
        // }
        //
        int divisor = 2;
        while (i > 1) { // `if`문을 `while`문으로 일반화
            while (i % divisor == 0) {
                factors.add(divisor);
                i /= divisor;
            }

            divisor++;
        }
        // 위의 `while`문을 벗어나는 경우는 `i`가 1이 될 때가 유일합니다.
        // 따라서 여기서 `i`가 1보다 큰 경우는 없으므로, 이 조건문은 쓸모가 없어져서 삭제합니다.
        // if (i > 1) {
        //     factors.add(i);
        // }

        return factors;
    }

    /**
     * 규칙 5 "가능하면 일반화하라."를 따릅니다.
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void factors_15() throws Exception {
        assertThat(factorsOf_15(1)).isEqualTo(Collections.emptyList());
        assertThat(factorsOf_15(2))
                .contains(2)
                .size().isEqualTo(1);
        assertThat(factorsOf_15(3))
                .contains(3)
                .size().isEqualTo(1);
        assertThat(factorsOf_15(4))
                .contains(2, 2)
                .size().isEqualTo(2);
        assertThat(factorsOf_15(5))
                .contains(5)
                .size().isEqualTo(1);
        assertThat(factorsOf_15(6))
                .contains(2, 3)
                .size().isEqualTo(2);
        assertThat(factorsOf_15(7))
                .contains(7)
                .size().isEqualTo(1);
        assertThat(factorsOf_15(8))
                .contains(2, 2, 2)
                .size().isEqualTo(3);
        assertThat(factorsOf_15(9))
                .contains(3, 3)
                .size().isEqualTo(2);
    }

    /**
     * 규칙 5 "가능하면 일반화하라."를 따릅니다.
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     *
     * @param i 소인수 분해하려는 정수입니다.
     * @return 소인수 분해 결과 리스트입니다.
     */
    private List<Integer> factorsOf_15(int i) {
        List<Integer> factors = new ArrayList<>();
        // int divisor = 2;
        // while (i > 1) { // `if`문을 `while`문으로 일반화
        //     while (i % divisor == 0) {
        //         factors.add(divisor);
        //         i /= divisor;
        //     }
        //
        //     divisor++;
        // }

        for (int divisor = 2; i > 1; divisor++) {
            for (; i % divisor == 0; i /= divisor) {
                factors.add(divisor);
            }
        }

        return factors;
    }
}
