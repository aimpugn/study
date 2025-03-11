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

    // 정렬 2
    // 이번에는 살짝 다른 경로를 택해 볼 것이라고 합니다.

    /**
     * 가능한 한 가장 퇴화한 테스트와 테스트를 통과시키는 코드로 시작합니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_0_success() throws Exception {
        assertThat(sort2_0(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_0(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_0(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
    }

    /**
     * 우선 {@link SortTest#sort2_0_success()}를 통과시키는 데 필요한 만큼의 코드만 작성합니다.
     *
     * @param list 정렬 대상인 리스트입니다.
     * @return 정렬된 리스트입니다.
     */
    private List<Integer> sort2_0(List<Integer> list) {
        return list;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     */
    @Test
    public void sort2_1_fail() {
        assertThat(sort2_0(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_0(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_0(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        // 이 테스트는 실패합니다.
        // assertThat(sort2_0(mutableList(2, 1))).isEqualTo(mutableList(1, 2)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1, 2]
        //Actual   :[2, 1]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_1_success() throws Exception {
        assertThat(sort2_1(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_1(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_1(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_1(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 앞서 {@link SortTest#sort_6(List)}은 in-place 알고리즘으로 입력 자료구조에 대해 직접 수정({@link Collections#swap(List, int, int)})이
     * 이뤄졌습니다.
     * 이번에는 다르게 접근하여 '입력 리스트 안에서 비교 후 바꾸기' 대신,
     * '비교 후 올바른 순서로 추가된 새로운 리스트 만들기'를 수행합니다.
     * <p>
     * '비교 후 바꾸기'가 아닌 새로운 리스트를 만드는 방식으로도 실패하는 {@link SortTest#sort2_1_fail} 테스트를 통과시킬 수 있습니다.
     * 실패하는 테스트를 통과하게 만드는 해법은 다양하게 존재할 수 있습니다.
     *
     * @param list 정렬 대상 리스트
     * @return 정렬된 리스트
     * @see <a href="https://en.wikipedia.org/wiki/In-place_algorithm">In-place algorithm</a>
     */
    private List<Integer> sort2_1(List<Integer> list) {
        if (list.size() <= 1) {
            return list;
        }

        var first = list.get(0);
        var second = list.get(1);
        if (first > second) {
            return mutableList(second, first);
        }

        return mutableList(first, second);
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_2_fail() throws Exception {
        assertThat(sort2_1(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_1(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_1(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_1(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        // `sort2_1` 코드상 원소가 셋 이상인 리스트르르 반환하는 분기가 없으므로 실패할 수밖에 없습니다.
        // assertThat(sort2_1(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1, 2, 3]
        //Actual   :[1, 2]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_2_success() throws Exception {
        assertThat(sort2_2(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_2(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_2(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_2(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_2(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
    }

    private List<Integer> sort2_2(List<Integer> list) {
        if (list.size() <= 1) {
            return list;
        }

        if (list.size() == 2) {
            var first = list.getFirst();
            var second = list.get(1);
            if (first > second) {
                return mutableList(second, first);
            }
            return mutableList(first, second);
        }

        // 원소가 세 개 이상인 경우(e.g. [1, 2, 3]) 경우를 처리합니다.
        return list;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_3_fail() throws Exception {
        assertThat(sort2_2(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_2(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_2(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_2(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_2(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        // 길이가 2인 경우에는 두 가지 경우(2 x 1)의 리스트가 가능해서 일일이 나열할 수 있었습니다.
        // 하지만 길이가 3인 경우에는 여섯 가지 경우(3 x 2 x 1)의 리스트가 가능합니다.
        // assertThat(sort2_2(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1, 2, 3]
        //Actual   :[2, 1, 3]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * 규칙 7 "더 복잡한 경우로 넘어가기 전에, 현재의 더 단순한 경우를 철저히 테스트하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_3_success() throws Exception {
        assertThat(sort2_3(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_3(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_3(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_3(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_3(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_3(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_3(mutableList(1, 3, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_3(mutableList(3, 2, 1))).isEqualTo(mutableList(1, 2, 3));
        // 원소 네 개짜리 테스트로 넘어가기 전에 원소 세 개짜리 리스트에 대한 모든 경우의 수를 시도해봐야 합니다.
        assertThat(sort2_3(mutableList(2, 3, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_3(mutableList(3, 1, 2))).isEqualTo(mutableList(1, 2, 3));
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @param list 정렬 대상 리스트
     * @return 오름차순으로 정렬된 리스트
     */
    private List<Integer> sort2_3(List<Integer> list) {
        if (list.size() <= 1) {
            return list;
        }

        if (list.size() == 2) {
            var first = list.getFirst();
            var second = list.get(1);
            if (first > second) {
                return mutableList(second, first);
            }
            return mutableList(first, second);
        }

        // 길이가 2인 경우에는 두 가지 경우(2 x 1)의 리스트가 가능해서 일일이 나열할 수 있었습니다.
        // 하지만 길이가 3인 경우에는 여섯 가지 경우(3 x 2 x 1)의 리스트가 가능합니다.
        // 만약 리스트 길이가 증가하면 (만약 중복 없다면) 경우의 수는 `n!`만큼 존재합니다.
        // 책에서는 '삼분법(law of trichotomy)'을 제안합니다.
        // 삼분법은 어떤 두 수 A와 B에 대해, '두 수의 관계'는 항상 다음 셋 중 하나라는 것입니다.
        // - A < B
        // - A = B
        // - A > B
        // 리스트에서 임의의 원소 하나를 골라서 다른 원소들과의 관계가 셋 중 무엇인지 판별합니다.

        // NOTE:
        // - 리스트에 `middle`과 같은 값은 없다고 가정한다는 점에 주의합니다.
        // - 1번 인덱스의 원소를 `middle`로 할 이유는 없습니다. 만약 첫 번째 원소를 `middle`로 설정하면
        //   `sort2_3_success`에서 [1, 3, 2] 케이스는 실패합니다.
        //   ```
        //   var first = list.get(1);
        //   var middle = list.get(0);
        //   var last = list.get(2);
        //   ```
        //   [1, 3, 2] 경우 `first=3`, `middle=1`, `last=2`가 됩니다.
        //   - `lessThanMiddle=[]`
        //   - `greaterThanMiddle=[3, 2]`
        //   순차적으로 결과를 만들면 `result=[1, 3, 2]`가 됩니다.
        //
        // - 원소가 네 개 이상인 경우 처리하지 못합니다.
        //
        // 굳이 1번 인덱스가 `middle`이어야 할 이유가 있을까?
        // ```
        // var first = list.get(0);
        // var middle = list.get(1);
        // var last = list.get(2);
        // ```
        // `middle`의 순서를 바꾸면 테스트가 실패할 수 있으므로,
        // 이를 확인하기 위해 0번 인덱스의 값을 `middle`로 설정해 봅니다.
        var first = list.get(1);
        var middle = list.get(0);
        var last = list.get(2);

        List<Integer> lessThanMiddle = new ArrayList<>();
        List<Integer> greaterThanMiddle = new ArrayList<>();

        // 임의의 원소를 `middle`로 잡고, `middle`을 기준으로 비교합니다.
        if (first < middle) lessThanMiddle.add(first);
        if (last < middle) lessThanMiddle.add(last);

        if (first > middle) greaterThanMiddle.add(first);
        if (last > middle) greaterThanMiddle.add(last);

        // `middle`보다 작은 값들, `middle`, `middle`보다 큰 값들을 순차적으로 추가합니다.
        // `lessThanMiddle=[2, 1]`처럼 역순인 경우가 있을 수 있으므로 다음 라인은 주석 처리합니다.
        // List<Integer> result = new ArrayList<>(lessThanMiddle);
        // 마침 앞서 `list.size() == 2`인 경우를 처리하고 있으므로 해당 로직을 재활용합니다.
        List<Integer> result = new ArrayList<>(sort2_3(lessThanMiddle));
        result.add(middle);
        // `greaterThanMiddle=[3, 2]`처럼 역순인 경우가 있을 수 있으므로 다음 라인은 주석 처리합니다.
        // result.addAll(greaterThanMiddle);
        // 마침 앞서 `list.size() == 2`인 경우를 처리하고 있으므로 해당 로직을 재활용합니다.
        result.addAll(sort2_3(greaterThanMiddle));

        return result;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_4_fail() throws Exception {
        assertThat(sort2_3(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_3(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_3(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_3(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_3(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_3(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_3(mutableList(1, 3, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_3(mutableList(3, 2, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_3(mutableList(2, 3, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_3(mutableList(3, 1, 2))).isEqualTo(mutableList(1, 2, 3));
        // 원소가 네 개인 경우에 대한 처리가 없으므로 실패해야 합니다.
        // assertThat(sort2_3(mutableList(1, 2, 3, 4)))
        //         .isEqualTo(mutableList(1, 2, 3, 4)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1, 2, 3, 4]
        //Actual   :[1, 2, 3]
        // assertThat(sort2_3(mutableList(2, 1, 3, 4)))
        //         .isEqualTo(mutableList(1, 2, 3, 4)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1, 2, 3, 4]
        //Actual   :[1, 2, 3]
        // assertThat(sort2_3(mutableList(4, 3, 2, 1)))
        //         .isEqualTo(mutableList(1, 2, 3, 4)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1, 2, 3, 4]
        //Actual   :[2, 3, 4]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_4_success() throws Exception {
        assertThat(sort2_4(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_4(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_4(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_4(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_4(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(1, 3, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(3, 2, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(2, 3, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(3, 1, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(1, 2, 3, 4)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_4(mutableList(2, 1, 3, 4)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_4(mutableList(4, 3, 2, 1)))
                .isEqualTo(mutableList(1, 2, 3, 4));
    }

    private List<Integer> sort2_4(List<Integer> list) {
        if (list.size() <= 1) return list;

        if (list.size() == 2) {
            var first = list.getFirst();
            var second = list.get(1);
            if (first > second) return mutableList(second, first);

            return mutableList(first, second);
        }

        var middle = list.get(0);
        var lessThanMiddle = list.stream()
                .filter(x -> x < middle)
                // 수정 가능한 리스트(mutable list)를 반환합니다.
                // 또한 병렬 스트림(parallelStream)에서도 안전하게 동작한다고 합니다.
                // .collect(Collectors.toList());
                //
                // `toList()`는 Java 16에서 도입되었습니다.
                // `Collectors.toUnmodifiableList()`를 호출하여 불변 리스트(immutable list)를 반환합니다.
                .toList();
        var greaterThanMiddle = list.stream()
                .filter(x -> x > middle)
                .toList();
        var result = new ArrayList<>(sort2_4(lessThanMiddle));
        result.add(middle);
        result.addAll(sort2_4(greaterThanMiddle));

        return result;
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_5_fail() throws Exception {
        assertThat(sort2_4(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_4(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_4(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_4(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_4(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(1, 3, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(3, 2, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(2, 3, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(3, 1, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_4(mutableList(1, 2, 3, 4)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_4(mutableList(2, 1, 3, 4)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_4(mutableList(4, 3, 2, 1)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        // 저자는 여기서 `middle = list.get(0)`을 의심합니다.
        // 앞서 "리스트에 `middle`과 같은 값은 없다고 가정한다는 점에 주의합니다."라고 했습니다.
        // 하지만 `middle`에 해당하는 값은 얼마든지 여럿 존재할 수 있습니다.
        // `sort2_4`는 `x > middle` 그리고 `x < middle` 두 가지 경우만 처리하고,
        // `x == middle`인 경우를 처리하고 있지 않기 때문에 다음 테스트 케이스는 실패합니다.
        // assertThat(sort2_4(mutableList(1, 3, 1, 2)))
        //         .isEqualTo(mutableList(1, 2, 3, 4)); // 다음 코드 진행 위해 주석 처리
        //Expected :[1, 2, 3, 4]
        //Actual   :[1, 2, 3]
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_5_success() throws Exception {
        assertThat(sort2_5(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_5(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_5(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_5(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_5(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5(mutableList(1, 3, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5(mutableList(3, 2, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5(mutableList(2, 3, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5(mutableList(3, 1, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5(mutableList(1, 2, 3, 4)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_5(mutableList(2, 1, 3, 4)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_5(mutableList(4, 3, 2, 1)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_5(mutableList(1, 3, 1, 2)))
                .isEqualTo(mutableList(1, 1, 2, 3));
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 이제 `middle`을 특별취급하지 않아야 합니다.
     *
     * @param list 정렬 대상 리스트
     * @return 오름차순 정렬된 리스트
     */
    private List<Integer> sort2_5(List<Integer> list) {
        if (list.size() <= 1) return list;

        if (list.size() == 2) {
            var first = list.getFirst();
            var second = list.get(1);
            if (first > second) return mutableList(second, first);

            return mutableList(first, second);
        }

        var middle = list.getFirst();
        var lessThanMiddle = list.stream().filter(x -> x < middle).toList();
        var middles = list.stream().filter(x -> x == middle).toList();
        var greaterThanMiddle = list.stream().filter(x -> x > middle).toList();

        var result = new ArrayList<>(sort2_5(lessThanMiddle));
        result.addAll(middles);
        result.addAll(sort2_5(greaterThanMiddle));

        return result;
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void sort2_5_refactoring_success() throws Exception {
        assertThat(sort2_5_refactored(mutableList())).isEqualTo(mutableList());
        assertThat(sort2_5_refactored(mutableList(1))).isEqualTo(mutableList(1));
        assertThat(sort2_5_refactored(mutableList(1, 2))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_5_refactored(mutableList(2, 1))).isEqualTo(mutableList(1, 2));
        assertThat(sort2_5_refactored(mutableList(1, 2, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5_refactored(mutableList(2, 1, 3))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5_refactored(mutableList(1, 3, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5_refactored(mutableList(3, 2, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5_refactored(mutableList(2, 3, 1))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5_refactored(mutableList(3, 1, 2))).isEqualTo(mutableList(1, 2, 3));
        assertThat(sort2_5_refactored(mutableList(1, 2, 3, 4)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_5_refactored(mutableList(2, 1, 3, 4)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_5_refactored(mutableList(4, 3, 2, 1)))
                .isEqualTo(mutableList(1, 2, 3, 4));
        assertThat(sort2_5_refactored(mutableList(1, 3, 1, 2)))
                .isEqualTo(mutableList(1, 1, 2, 3));
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     * <p>
     * 이렇게 완성된 정렬 알고리즘을 '퀵(quick) 정렬'이라고 합니다.
     * <p>
     * 위치를 서로 스왑하는 경우 버블 정렬({@link SortTest#sort_6})을 구현하게 되었고,
     * 비교 후 올바른 순서로 추가된 새로운 리스트 만드는 경우 퀵 정렬({@link SortTest#sort2_5_refactored})을 구현하게 됐습니다.
     * <p>
     * "길에서 갈림길을 알아보고 올바른 길을 찾는 것"이 중요할 수 있다는 것을 의미합니다.
     * 그렇다면 "우리가 어떻게 갈림길을 알아보고, 갈림길 중 어떤 길이 더 나은지 판단할 수 있을까?"
     *
     * @param list 정렬 대상 리스트
     * @return 오름차순으로 정렬된 리스트
     */
    private List<Integer> sort2_5_refactored(List<Integer> list) {
        // 책에서는 다음과 같이 주어진 리스트가 비어있을 때 새로운 리스트를 반환하는 것을 더 개선됐다고 합니다.
        // 이를 통해 얻을 수 있는 이점은 다음과 같습니다.
        // - 우선 항상 새로운 리스트를 생성하여 리턴하므로 원본의 불변성(immutability)을 유지할 수 있을 것으로 보입니다.
        // - 정렬된 리스트는 새로운 객체를 만들고, 원본 리스트를 변경하지 않는다는 로직이 더 명확해집니다.
        // - 또한 리스트가 비어 있지 않다면 세 개의 부분으로 분할하는 로직을 일관되게 사용할 수 있습니다.
        // - "주어진 리스트를 수정하지 않고 항상 새로운 리스트를 만들어 반환"하는 함수형 스타일을 따른다고 할 때,
        //   빈 리스트를 명확하게 처리하고, 나머지는 동일한 로직을 적용하는 편이 단순화된 구조를 가질 수 있습니다.
        // - 크기가 1인 경우에 대해 추가적인 재귀 호출이 있지만, 성능에 큰 영향은 없을 것으로 보입니다.
        //
        // 다만, 다음과 같은 단점들도 예상됩니다.
        // - 비어있더라도 어쨌든 새로운 객체를 생성합니다.
        // - 크기가 1인 경우 굳이 더 이상 정렬을 할 필요가 없음에도, 재귀 호출이 한 번 더 발생합니다.
        //
        // 따라서 기존 if 문은 제거합니다:
        // if (list.size() <= 1) return list;
        if (list.isEmpty()) return new ArrayList<>();

        // 이 부분을 지워도 테스트는 통과합니다.
        // ```
        // if (list.size() == 2) {
        //     var first = list.getFirst();
        //     var second = list.get(1);
        //     if (first > second) return mutableList(second, first);
        //     return mutableList(first, second);
        // }
        // ```
        // 이 정렬 알고리즘의 핵심은 더 이상 `first`와 `second`로 새로운 리스트를 만드는 게 아니기 때문입니다.
        // 대신 재귀적으로 '주어진 리스트를 임의의 `middle`값을 기준으로 세 부분으로 나누고 그 순서대로 담기'가 정렬의 핵심 로직이 되었습니다.
        // - `middle`보다 작은 값들
        // - `middle`과 같은 값들
        // - `middle`보다 큰 값들
        // - 그리고 `middle`보다 작은 값들, `middle`과 같은 값들, `middle`보다 큰 값들 순서대로 결과에 담기
        // 따라서 더 이상 필요 없어진 로직 부분은 제거해도 정렬은 정상적으로 동작합니다.

        int middle = list.getFirst();
        var lessThanMiddle = list.stream().filter(x -> x < middle).toList();
        var middles = list.stream().filter(x -> x == middle).toList();
        var greaterThanMiddle = list.stream().filter(x -> x > middle).toList();

        // 이 부분을 통해 오름차순으로 정렬이 됩니다.
        var result = new ArrayList<>(sort2_5_refactored(lessThanMiddle));
        result.addAll(middles);
        result.addAll(sort2_5_refactored(greaterThanMiddle));
        // 만약 순서를 다음과 같이 큰 값들, 중간 값들, 작은 값들 순서로 담으면 내림차순 정렬이 됩니다.
        // ```
        // var result = new ArrayList<>(sort2_5_refactored(greaterThanMiddle));
        // result.addAll(middles);
        // result.addAll(sort2_5_refactored(lessThanMiddle));
        // ```

        return result;
    }
}
