package craftsmanship.d;

import org.testng.annotations.Test;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;

import static org.assertj.core.api.Assertions.assertThat;

public class SortTest {

    /**
     * 아무 일도 하지 않는 테스트부터 시작합니다.
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
    public void sorted_0_fail() throws Exception {
        // assertThat(sort(Collections.emptyList())).isEqualTo(Collections.emptyList()); // 다음 코드 진행 위해 주석 처리
        //Expected :[]
        //Actual   :null
    }

    private List<Integer> sort_0(List<Integer> list) {
        return null;
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sorted_0_success() throws Exception {
        assertThat(sort_1(mutableList())).isEqualTo(mutableList());
    }

    private List<Integer> sort_1(List<Integer> list) {
        return new ArrayList<>();
    }

    /**
     * 가변 인자 배열을 {@link Arrays.ArrayList}로 변환하고,
     * 이를 다시 {@link ArrayList}로 변환합니다.
     * <p>
     * 참고로 {@link Arrays.ArrayList}는 `set`은 가능하지만, `add`는 지원하지 않습니다.
     * <p>
     * 제네릭을 가변 인자와 함께 사용하면 아래와 같은 경고가 발생합니다.
     * <pre>
     * {@code Possible heap pollution from parameterized vararg type }
     * </pre>
     * <p>
     * 제네릭(Generic) 타입 정보는 컴파일 시에만 유지되며, 런타임에는 타입 정보가 제거(type erasure) 됩니다. 이 때문에
     * - 컴파일 타임에는 타입 안정성이 보장되지만,
     * - 런타임에는 타입 정보가 손실되어 잘못된 타입이 저장될 위험이 있습니다.
     * 따라서 '제네릭 배열'을 만들거나, '가변 인자로 제네릭 타입'을 사용하면 런타임에서 잘못된 타입이 저장될 '가능성'이 있습니다.
     *
     * <pre>
     * {@code
     * static void unsafeMethod(List<String>... stringLists) {
     *     Object[] array = stringLists; // List<String>[] → Object[] 로 변환
     *     List<Integer> intList = List.of(42);
     *     array[0] = intList; // List<Integer>가 List<String>으로 저장되면서 힙 오염 발생
     *
     *     String s = stringLists[0].get(0); // ClassCastException 가능
     *     System.out.println(s);
     * }
     * }
     * </pre>
     * <p>
     * Java 컴파일러는 제네릭 타입을 가변 인자로 사용할 때 항상 경고를 발생시킵니다.
     * 따라서 힙 오염 가능성이 없는 안전한 상황에서도 경고는 발생합니다.
     * 이때 {@link SafeVarargs} 어노테이션을 사용하여 컴파일러에게 "이 가변 인자 사용은 안전하다"고 알려줍니다.
     *
     * @param elements 가변 배열이 전달됩니다.
     * @return 수정 가능한 리스트를 반환합니다.
     */
    @SafeVarargs
    private <T> List<T> mutableList(T... elements) {
        return new ArrayList<>(Arrays.asList(elements));
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sorted_1_fail() throws Exception {
        assertThat(sort_1(mutableList())).isEqualTo(mutableList());
        // 퇴화한 정도를 한 단계 낮춰서 정수가 하나 들어 있는 리스트를 넘겨 봅니다.
        // assertThat(sort_1(mutableIntegerList(1))).isEqualTo(mutableIntegerList(1)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1]
        //Actual   :[]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sorted_1_success() throws Exception {
        assertThat(sort_2(mutableList())).isEqualTo(mutableList());
        assertThat(sort_2(mutableList(1))).isEqualTo(mutableList(1));
    }

    private List<Integer> sort_2(List<Integer> list) {
        // 프로덕션 코드를 더 일반화시킵니다.
        // return new ArrayList<>();
        return list;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sorted_2_success() throws Exception {
        assertThat(sort_2(mutableList())).isEqualTo(mutableList());
        assertThat(sort_2(mutableList(1))).isEqualTo(mutableList(1));
        // 원소가 두 개이고 이미 정렬되어 있는 리스트를 테스트합니다.
        // 실패하는 테스트가 아니므로 추가하면 안 된다고 할 수 있지만, "이 테스트가 통과하는지 확인하는 편이 낫다"고 합니다.
        assertThat(sort_2(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
    }

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sorted_3_fail() throws Exception {
        assertThat(sort_2(mutableList())).isEqualTo(mutableList());
        assertThat(sort_2(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort_2(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        // 실패: 원소를 역순으로 변경하면 실패합니다.
        // assertThat(sort_2(mutableIntegerList(2, 1))).isEqualTo(mutableIntegerList(1, 2)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1, 2]
        //Actual   :[2, 1]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sorted_3_success() throws Exception {
        assertThat(sort_3(mutableList())).isEqualTo(mutableList());
        assertThat(sort_3(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort_3(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort_3(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link SortTest#sorted_3_fail()}에서의 `[2, 1]`로 전달되는 리스트를
     * `[1, 2]`로 정렬하여 리턴해야 합니다.
     *
     * @param list 두 개의 요소를 갖는 리스트가 주어집니다.
     * @return 두 개의 요소를 정렬한 리스트를 반환합니다.
     */
    private List<Integer> sort_3(List<Integer> list) {
        if (1 < list.size()) {
            if (list.get(0) > list.get(1)) { // 역순인 경우
                Collections.swap(list, 0, 1);
            }
        }

        return list;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sorted_4_success() throws Exception {
        assertThat(sort_3(mutableList())).isEqualTo(mutableList());
        assertThat(sort_3(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort_3(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort_3(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort_3(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        // 첫 번째와 두 번째 원소의 순서가 뒤집힌 경우에는 테스트를 통과합니다.
        assertThat(sort_3(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
    }

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sorted_5_fail() throws Exception {
        assertThat(sort_3(mutableList())).isEqualTo(mutableList());
        assertThat(sort_3(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort_3(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort_3(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort_3(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort_3(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
        // 하지만 첫 번째, 두 번째, 세 번째 원소가 모두 뒤집혀서 역순인 경우에는 실패합니다.
        // assertThat(sort_3(mutableList(2, 3, 1))).isEqualTo(mutableList(1, 2, 3)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1, 2, 3]
        //Actual   :[2, 3, 1]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sorted_5_success() throws Exception {
        assertThat(sort_5(mutableList())).isEqualTo(mutableList());
        assertThat(sort_5(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort_5(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort_5(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort_5(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort_5(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort_5(mutableList(2, 3, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort_5(mutableList(3, 2, 1))).isEqualTo(mutableList(1, 2, 3));
    }

    /**
     * {@link SortTest#sorted_5_fail()}를 통과시키기 위해,
     * 리스트 전체에 대해 '비교 후 바꾸기 알고리즘'을 수행하도록 수정합니다.
     *
     * @param list 정렬할 대상 리스트입니다.
     * @return 오름차순 정렬한 리스트입니다.
     */
    private List<Integer> sort_5(List<Integer> list) {
        if (1 < list.size()) {
            // 리스트 전체에 대해 비교 후 '비교 후 바꾸기 알고리즘' 수행합니다.
            // ```
            // for (int firstIdx = 0; firstIdx < list.size() - 1; firstIdx++) {
            //     int secondIdx = firstIdx + 1;
            //     if (list.get(firstIdx) > list.get(secondIdx)) {
            //         Collections.swap(list, firstIdx, secondIdx);
            //     }
            // }
            // ```
            // 하지만 위의 코드로는 테스트를 통과할 수 없습니다.
            // list: [2, 3, 1]
            // - firstIdx=0, [2, 3, 1]
            // - firstIdx=1, [2, 1, 3] <= 첫 번째와 두 번째 원소의 순서가 아직 틀린 상태입니다.
            //
            // 전체 요소에 대해 한 번 '비교 후 바꾸기'를 수행한 후에도 아직 순서가 잘못된 요소가 남아있다는 것은,
            // '비교 후 바꾸기' 반복문을 다른 반복문 안에 넣어야 함을 의미합니다.
            //
            // 한 번 모든 원소에 대해 '비교 후 바꾸기' 수행한 후 다시 한 번 순회하며 체크합니다.
            for (int i = 0; i < list.size() - 1; i++) {
                for (int firstIdx = 0; firstIdx < list.size() - 1; firstIdx++) {
                    int secondIdx = firstIdx + 1;
                    if (list.get(firstIdx) > list.get(secondIdx)) {
                        Collections.swap(list, firstIdx, secondIdx);
                    }
                }
            }
        }

        return list;
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     */
    @Test
    public void sorted_6_refactoring_success() {
        assertThat(sort_6(mutableList())).isEqualTo(mutableList());
        assertThat(sort_6(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort_6(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort_6(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort_6(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort_6(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort_6(mutableList(2, 3, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort_6(mutableList(3, 2, 1))).isEqualTo(mutableList(1, 2, 3));
        // 대규모 테스트로 마무리합니다.
        assertThat(sort_6(mutableList(3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5, 8, 9, 7, 9, 3)))
                .isEqualTo(mutableList(1, 1, 2, 3, 3, 3, 4, 5, 5, 5, 6, 7, 8, 9, 9, 9));
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     * <p>
     * 오름차순 버블 정렬의 경우 한 번 순회할 때마다 가장 큰 값이 가장 뒤로 가게 됩니다.
     * 따라서 한 번 전체 목록에 대해 '비교 후 바꾸기' 알고리즘을 수행한 후에는,
     * 마지막 요소에 대해서는 더 이상 '비교 후 바꾸기'를 수행할 필요가 없으므로 비교 대상에서 제외하도록 리팩토링합니다.
     *
     * @param list 정렬하려는 리스트입니다.
     * @return 오름차순으로 정렬된 리스트입니다.
     */
    private List<Integer> sort_6(List<Integer> list) {
        if (1 < list.size()) {
            // ```
            // for (int i = 0; i < list.size() - 1; i++) {
            //     for (int firstIdx = 0; firstIdx < list.size() - 1; firstIdx++) {
            //         int secondIdx = firstIdx + 1;
            //         if (list.get(firstIdx) > list.get(secondIdx)) {
            //             Collections.swap(list, firstIdx, secondIdx);
            //         }
            //     }
            // }
            // ```
            // list: [3, 2, 1]
            // 이를 정렬하기 위해서 다음 절차를 거칩니다.
            // - i: 0
            //    - firstIdx: 0, [3, 2, 1] => [2, 3, 1]
            //    - firstIdx: 1, [2, 3, 1] => [2, 1, 3]
            // - i: 1
            //    - firstIdx: 0, [2, 1, 3] => [1, 2, 3]
            //    - firstIdx: 1, [1, 2, 3] 유지
            //
            // `i=1`일 때 `[1, 2, 3] 유지`라는 불필요한 중복 비교 과정이 있습니다.
            //
            // 오름차순 버블 정렬의 경우 한 번 순회할 때마다 가장 큰 값이 가장 뒤로 가게 됩니다.
            // 따라서 한 번 전체 목록에 대해 '비교 후 바꾸기' 알고리즘을 수행한 후에는,
            // 마지막 요소에 대해서는 더 이상 '비교 후 바꾸기'를 수행할 필요가 없으므로 비교 대상에서 제외하도록 리팩토링합니다.
            for (int limit = list.size() - 1; 0 < limit; limit--) {
                for (int firstIdx = 0; firstIdx < limit; firstIdx++) {
                    int secondIdx = firstIdx + 1;
                    if (list.get(firstIdx) > list.get(secondIdx)) {
                        Collections.swap(list, firstIdx, secondIdx);
                    }
                }
            }
        }

        return list;
    }
}
