package craftsmanship.c;

import org.testng.annotations.BeforeTest;
import org.testng.annotations.Test;

import java.util.stream.IntStream;

import static org.assertj.core.api.Assertions.assertThat;

public class BowlingTest {

    /**
     * 컴파일하고 실행할 수 있음을 증명하기 위해 "아무 일도 하지 않는 테스트로 시작"합니다.
     * <p>
     * 테스트가 통과하면 삭제합니다.
     * <p>
     * 계단(stairstep) 테스트이므로 삭제 대상입니다.
     *
     * @throws Exception
     */
    @Test
    public void nothing() throws Exception {
    }

    /**
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * "작성하고 싶은 코드를 작성하게 만드는 테스트를 쓰라."
     * <p>
     * 계단(stairstep) 테스트이므로 삭제 대상입니다.
     *
     * @throws Exception
     */
    @Test
    public void canCreateGame_23_22_compile_error() throws Exception {
        // Game game = new Game(); // 다음 코드 진행 위해 주석 처리

        // `Game` 클래스는 아직 구현되지 않았으므로, 다음과 같은 컴파일 타임 에러가 발생합니다.
        // /path/to/app/src/test/java/craftsmanship/c/BowlingTest.java:25: error: cannot find symbol
        //        Game game = new Game();
        //        ^
        //  symbol:   class Game
        //  location: class BowlingTest
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void canCreateGame_23_27_compile_success() throws Exception {
        Game_23_27 game = new Game_23_27();
    }

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * <p>
     * 계단(stairstep) 테스트이므로 삭제 대상입니다.
     *
     * @throws Exception
     */
    @Test
    public void canRoll_23_31_compile_error() throws Exception {
        Game_23_27 game = new Game_23_27();
        // game.roll(0); // 다음 코드 진행 위해 주석 처리

        // /path/to/app/src/test/java/craftsmanship/c/BowlingTest.java:55: error: cannot find symbol
        //        game.roll(0);
        //            ^
        //  symbol:   method roll(int)
        //  location: variable game of type Game_23_27
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 계단(stairstep) 테스트이므로 삭제 대상입니다.
     *
     * @throws Exception
     */
    @Test
    public void canRoll_23_31_compile_success() throws Exception {
        Game_23_33 game = new Game_23_33();
        game.roll(0);
    }

    /**
     * `Game` 인스턴스 생성이 중복될 것으로 보여 `setUp` 메서드로 분리할 수 있습니다.
     * 하지만 여기서는 {@link craftsmanship.a.StackTest}처럼 시간순으로 리팩토링 과정을 정리할 것이므로,
     * `setUp` 메서드는 사용하지 않습니다.
     */
    @BeforeTest
    public void setUp() {
    }

    // 다음으로 '게임 점수를 계산할 수 있는지' 확인하고 싶은데,
    // `score` 함수는 모든 투구를 마친 후에 호출할 수 있다고 정했으므로,
    // 일단 공을 굴려서 게임을 마쳐야 합니다.

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * 이 경우 실패하도록 {@link Game_23_46#score()}가 항상 -1을 반환하도록 합니다.
     *
     * @throws Exception
     */
    @Test
    public void gutterGame_23_46_fail() throws Exception {
        Game_23_46 game = new Game_23_46();

        for (int i = 0; i < 20; i++) {
            game.roll(0);
        }

        // assertThat(game.score()).isEqualTo(0); // 다음 코드 진행 위해 주석 처리
        //Expected :0
        //Actual   :-1
    }

    /**
     * 규칙 4 "실패하는 가장 간단하고, 가장 구체적이며, 가장 퇴화한 테스트 작성하라."를 따릅니다.
     * <p>
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * 이 경우 성공하도록 {@link Game_23_51#score()}가 항상 0을 반환하도록 합니다.
     *
     * @throws Exception
     */
    @Test
    public void gutterGame_23_51_success() throws Exception {
        Game_23_51 game = new Game_23_51();

        for (int i = 0; i < 20; i++) {
            game.roll(0);
        }

        assertThat(game.score()).isEqualTo(0);
    }

    // 그 다음으로 떠오르는 단순한 테스트는, 모두 1점인 경우입니다.

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void allOnes_23_51_fail() throws Exception {
        Game_23_51 game = new Game_23_51();

        for (int i = 0; i < 20; i++) {
            game.roll(1);
        }

        // 항상 0을 반환하므로 항상 실패합니다.
        // assertThat(game.score()).isEqualTo(20); // 다음 코드 진행 위해 주석 처리
        //Expected :20
        //Actual   :0
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link BowlingTest#allOnes_23_51_fail} 테스트는 모두 1점인 경우 실패합니다.
     * 이 테스트를 통과할 수 있도록 {@link Game_23_57#roll(int)}에서 각 공의 점수를 모두 더합니다.
     * <p>
     * 저자는 아직 이 알고리즘이 어떻게 볼링 점수 계산 규칙으로 진화할 수 있을지 잘 모르겠다고 합니다.
     *
     * @throws Exception
     */
    @Test
    public void allOnes_23_57_success() throws Exception {
        Game_23_57 game = new Game_23_57();

        for (int i = 0; i < 20; i++) {
            game.roll(1);
        }

        assertThat(game.score()).isEqualTo(20);
    }

    // 이때쯤 리팩토링 과정을 거치는데, `Game` 클래스가 아닌, 테스트를 조금 더 편하게 하기 위한 리팩토링입니다.
    // 대략 코드는 다음과 같습니다.
    // ```
    // private void rollMany(int times, int pins) {
    //     Game_23_57 game = new Game_23_57();
    //     for (int i = 0; i < times; i++) {
    //         game.roll(pins);
    //     }
    // }
    // ```

    // 다음으로 퇴화된 테스트를 조금 더 진행시켜 봅니다.
    // 일반 점수 합치는 것을 통과했으니, 이제 스페어를 고려해봅니다.
    // 스페어 하나, 그 다음 투구 하나, 그리고 나머지는 모두 도랑에 빠진 경우를 계산해 봅니다.

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.'
     * <p>
     * 이 테스트는 실패합니다.
     *
     * @throws Exception
     */
    @Test
    public void oneSpare_24_06_fail() throws Exception {
        Game_23_57 game = new Game_23_57();

        // 두 번 투구하여 10개의 핀을 쓰러뜨리므로, 스페어가 됩니다.
        for (int i = 0; i < 2; i++) {
            game.roll(5);
        }
        // 그 다음 투구 하나: 7개의 핀을 쓰러뜨립니다.
        game.roll(7);

        for (int i = 0; i < 17; i++) {
            game.roll(0);
        }

        // 1. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '2.1'에서 7점 + 10점 = 17점(총 17점)
        //    1. 5개의 핀 쓰러뜨림
        //    2. 5개의 핀 쓰러뜨림(스페어)
        // 2. 프레임: 7점(총 24점)
        //    1. 7개의 핀 쓰러뜨림
        //    2. 0개의 핀 쓰러뜨림
        // 나머지 모두 0점이므로, 총 24점이 예상됩니다.
        // assertThat(game.score()).isEqualTo(24); // 다음 코드 진행 위해 주석 처리
        //Expected :24
        //Actual   :17
    }

    // `Game_24_38`에서 `score` 함수가 점수를 계산하도록 리팩토링을 수행했습니다.
    // 이에 따라 gutterGame과 allOnes 테스트를 다시 테스트합니다.

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void gutterGame_24_38_success() throws Exception {
        Game_24_38 game = new Game_24_38();

        for (int i = 0; i < 20; i++) {
            game.roll(0);
        }

        assertThat(game.score()).isEqualTo(0);
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void allOnes_24_38_success() throws Exception {
        Game_24_38 game = new Game_24_38();

        for (int i = 0; i < 20; i++) {
            game.roll(1);
        }

        assertThat(game.score()).isEqualTo(20);
    }

    // 하지만 `oneSpare_24_06_fail`에서처럼 아직 스페어에 대한 처리가 없으므로,
    // 스페어 테스트는 여전히 실패합니다.

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void oneSpare_24_38_fail() throws Exception {
        Game_24_38 game = new Game_24_38();

        // 두 번 투구하여 10개의 핀을 쓰러뜨리므로, 스페어가 됩니다.
        for (int i = 0; i < 2; i++) {
            game.roll(5);
        }
        // 그 다음 투구 하나: 7개의 핀을 쓰러뜨립니다.
        game.roll(7);

        for (int i = 0; i < 17; i++) {
            game.roll(0);
        }

        // 1. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '2.1'에서 7점 + 10점 = 17점(총 17점)
        //    1. 5개의 핀 쓰러뜨림
        //    2. 5개의 핀 쓰러뜨림(스페어)
        // 2. 프레임: 7점(총 24점)
        //    1. 7개의 핀 쓰러뜨림
        //    2. 0개의 핀 쓰러뜨림
        // 나머지 모두 0점이므로, 총 24점이 예상됩니다.
        // assertThat(game.score()).isEqualTo(24); // 다음 코드 진행 위해 주석 처리
        //Expected :24
        //Actual   :17
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void gutterGame_24_55_success() throws Exception {
        Game_24_55 game = new Game_24_55();

        for (int i = 0; i < 20; i++) {
            game.roll(0);
        }

        assertThat(game.score()).isEqualTo(0);
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void allOnes_24_55_success() throws Exception {
        Game_24_55 game = new Game_24_55();

        for (int i = 0; i < 20; i++) {
            game.roll(1);
        }

        assertThat(game.score()).isEqualTo(20);
    }

    // 하지만 `oneSpare_24_06_fail`, `oneSpare_24_38_fail`에서처럼 아직 스페어에 대한 처리가 없으므로,
    // 스페어 테스트는 여전히 실패합니다.

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void oneSpare_24_55_fail() throws Exception {
        Game_24_55 game = new Game_24_55();

        // 두 번 투구하여 10개의 핀을 쓰러뜨리므로, 스페어가 됩니다.
        for (int i = 0; i < 2; i++) {
            game.roll(5);
        }
        // 그 다음 투구 하나: 7개의 핀을 쓰러뜨립니다.
        game.roll(7);

        for (int i = 0; i < 17; i++) {
            game.roll(0);
        }

        // 1. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '2.1'에서 7점 + 10점 = 17점(총 17점)
        //    1. 5개의 핀 쓰러뜨림
        //    2. 5개의 핀 쓰러뜨림(스페어)
        // 2. 프레임: 7점(총 24점)
        //    1. 7개의 핀 쓰러뜨림
        //    2. 0개의 핀 쓰러뜨림
        // 나머지 모두 0점이므로, 총 24점이 예상됩니다.
        // assertThat(game.score()).isEqualTo(24); // 다음 코드 진행 위해 주석 처리
        //Expected :24
        //Actual   :17
    }

    // `oneSpare_24_06_fail` 이후 두 번의 리팩토링을 거쳐서 `Game_24_38`, `Game_24_55`로 변경되었습니다.
    // 하지만 `gutterGame`(성공), `allOnes`(성공), `oneSpare`(실패) 테스트는 모두 계속 같은 결과가 나옵니다.
    // 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
    // 이러한 점에서 '진정한 리팩토링'이라고 볼 수 있습니다.


    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void gutterGame_25_13_success() throws Exception {
        Game_25_13 game = new Game_25_13();

        for (int i = 0; i < 20; i++) {
            game.roll(0);
        }

        assertThat(game.score()).isEqualTo(0);
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void allOnes_25_13_success() throws Exception {
        Game_25_13 game = new Game_25_13();

        for (int i = 0; i < 20; i++) {
            game.roll(1);
        }

        assertThat(game.score()).isEqualTo(20);
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     * <p>
     * {@link Game_25_13#score()}에서 스페어를 처리하는 코드를 추가합니다.
     * 이제 {@link BowlingTest#oneSpare_24_06_fail}에서 처음 마주쳤던 스페어 테스트가 통과합니다.
     *
     * @throws Exception
     */
    @Test
    public void oneSpare_25_13_success() throws Exception {
        Game_25_13 game = new Game_25_13();

        // 두 번 투구하여 10개의 핀을 쓰러뜨리므로, 스페어가 됩니다.
        for (int i = 0; i < 2; i++) {
            game.roll(5);
        }
        // 그 다음 투구 하나: 7개의 핀을 쓰러뜨립니다.
        game.roll(7);

        for (int i = 0; i < 17; i++) {
            game.roll(0);
        }

        // 1. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '2.1'에서 7점 + 10점 = 17점(총 17점)
        //    1. 5개의 핀 쓰러뜨림
        //    2. 5개의 핀 쓰러뜨림(스페어)
        // 2. 프레임: 7점(총 24점)
        //    1. 7개의 핀 쓰러뜨림
        //    2. 0개의 핀 쓰러뜨림
        // 나머지 모두 0점이므로, 총 24점이 예상됩니다.
        assertThat(game.score()).isEqualTo(24);
    }

    // 간단한 스페어 테스트가 끝났으므로, 그 다음에는 간단한 스트라이크 테스트를 추가합니다.

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     * 첫 번째 법칙 "프로덕션 코드 작성 전, 실패하는 테스트를 작성하라."를 따릅니다.
     * <p>
     * 이 테스트는 실패합니다.
     *
     * @throws Exception
     */
    @Test
    public void oneStrike_25_32_fail() throws Exception {
        Game_24_55 game = new Game_24_55();
        // 스트라이크
        game.roll(10);
        game.roll(2);
        game.roll(3);
        IntStream.range(0, 16)
                .forEach(num -> game.roll(0));

        // 1. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '2.1'과 '2.2'에서 5점 + 10점 = 15점(총 15점)
        //    1. 10개의 핀 쓰러뜨림(스트라이크)
        // 2. 프레임: 5점(총 20점)
        //    1. 2개의 핀 쓰러뜨림
        //    2. 3개의 핀 쓰러뜨림
        // 나머지는 모두 도랑에 빠집니다.
        // assertThat(game.score()).isEqualTo(20); // 다음 코드 진행 위해 주석 처리
        //Expected :20
        //Actual   :15
    }

    /**
     * 두 번째 법칙 "실패하는 테스트가 통과하는 데 필요한 최소한의 프로덕션 코드만 작성한다."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void oneStrike_26_02_success() throws Exception {
        Game_26_02 game = new Game_26_02();
        game.roll(10);
        game.roll(2);
        game.roll(3);
        IntStream.range(0, 16).forEach(num -> game.roll(0));

        // 1. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '2.1'과 '2.2'에서 5점 + 10점 = 15점(총 15점)
        //    1. 10개의 핀 쓰러뜨림(스트라이크)
        // 2. 프레임: 5점(총 20점)
        //    1. 2개의 핀 쓰러뜨림
        //    2. 3개의 핀 쓰러뜨림
        // 나머지는 모두 도랑에 빠집니다.
        assertThat(game.score()).isEqualTo(20);
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void gutterGame_26_15_refactoring_success() throws Exception {
        Game_26_15 game = new Game_26_15();
        IntStream.range(0, 20).forEach(idx -> game.roll(0));
        assertThat(game.score()).isEqualTo(0);
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void allOnes_26_15_refactoring_success() throws Exception {
        Game_26_15 game = new Game_26_15();
        IntStream.range(0, 20).forEach(idx -> game.roll(1));
        assertThat(game.score()).isEqualTo(20);
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     *
     * @throws Exception
     */
    @Test
    public void oneSpare_26_15_refactoring_success() throws Exception {
        Game_26_15 game = new Game_26_15();

        // 두 번 투구하여 10개의 핀을 쓰러뜨리므로, 스페어가 됩니다.
        IntStream.range(0, 2).forEach(idx -> game.roll(5));
        // 그 다음 투구 하나: 7개의 핀을 쓰러뜨립니다.
        game.roll(7);
        IntStream.range(0, 17).forEach(idx -> game.roll(0));

        // 1. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '2.1'에서 7점 + 10점 = 17점(총 17점)
        //    1. 5개의 핀 쓰러뜨림
        //    2. 5개의 핀 쓰러뜨림(스페어)
        // 2. 프레임: 7점(총 24점)
        //    1. 7개의 핀 쓰러뜨림
        //    2. 0개의 핀 쓰러뜨림
        // 나머지 모두 0점이므로, 총 24점이 예상됩니다.
        assertThat(game.score()).isEqualTo(24);
    }

    /**
     * 네 번째 법칙 리팩토링, "먼저 돌아가게 만들라. 그 다음 제대로 만들라."를 따릅니다.
     * 리팩토링은 "동작에는 영향이 없으면서 더 나은 구조의 코드로 변경하는 것"입니다.
     */
    @Test
    public void oneStrike_26_15_refactoring_success() {
        Game_26_15 game = new Game_26_15();
        game.roll(10);
        game.roll(2);
        game.roll(3);
        IntStream.range(0, 16).forEach(num -> game.roll(0));

        // 1. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '2.1'과 '2.2'에서 5점 + 10점 = 15점(총 15점)
        //    1. 10개의 핀 쓰러뜨림(스트라이크)
        // 2. 프레임: 5점(총 20점)
        //    1. 2개의 핀 쓰러뜨림
        //    2. 3개의 핀 쓰러뜨림
        // 나머지는 모두 도랑에 빠집니다.
        assertThat(game.score()).isEqualTo(20);
    }

    // 아직 10 프레임을 테스트하지 않았으므로, 이제 10 프레임 테스트를 시도합니다.

    /**
     * 세 번째 법칙 "테스트가 통과하면 테스트를 더 추가하라."를 따릅니다.
     *
     * @throws Exception
     */
    @Test
    public void perfectGame_26_35() throws Exception {
        Game_26_15 game = new Game_26_15();

        // 1. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '2.1'와 '3.1'에서 20점 + 10점 = 30점(총 30점)
        //   1. 10개의 핀 쓰러뜨림
        // 2. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '3.1'과 '4.1'에서 20점 + 10점 = 30점(총 60점)
        //   1. 10개의 핀 쓰러뜨림
        // 3. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '4.1'과 '5.1'에서 20점 + 10점 = 30점(총 90점)
        //   1. 10개의 핀 쓰러뜨림
        // 4. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '5.1'과 '6.1'에서 20점 + 10점 = 30점(총 120점)
        //   1. 10개의 핀 쓰러뜨림
        // ...
        // 10. 프레임: 마지막 두 번의 투구 진행 전에는 계산 불가 -> '10.2'와 '10.3'에서 20점 + 10점 = 30점(총 300점)
        //   1. 10개의 핀 쓰러뜨림
        //   2. 10개의 핀 쓰러뜨림
        //   3. 10개의 핀 쓰러뜨림
        IntStream.range(0, 12).forEach(idx -> game.roll(10));

        assertThat(game.score()).isEqualTo(300);
    }

    /**
     * 책에 소개된 일반적인 볼링 게임 시나리오에 대해 테스트합니다.
     *
     * <pre>
     * {@code
     * 1. 프레임: 5점
     *     1. 1개의 핀 쓰러뜨림
     *     2. 4개의 핀 쓰러뜨림
     * 2. 프레임: 9점(총 14점)
     *     1. 4개의 핀 쓰러뜨림
     *     2. 5개의 핀 쓰러뜨림
     * 3. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '4.1'에서 5점 + 10점 = 15점(총 29점)
     *     1. 6개의 핀 쓰러뜨림
     *     2. 4개의 핀 쓰러뜨림(스페어)
     * 4. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '5.1'에서 10점 + 10점 = 20점(총 49점)
     *     1. 5개의 핀 쓰러뜨림
     *     2. 5개의 핀 쓰러뜨림(스페어)
     * 5. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '6.1'과 '6.2'에서 1점 + 10점 = 11점(총 60점)
     *     1. 10개의 핀 쓰러뜨림(스트라이크)
     * 6. 프레임: 1점(총 61점)
     *     1. 0개의 핀 쓰러뜨림
     *     2. 1개의 핀 쓰러뜨림
     * 7. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '8.1'에서 6점 + 10점 = 16점(총 77점)
     *     1. 7개의 핀 쓰러뜨림
     *     2. 3개의 핀 쓰러뜨림(스페어)
     * 8. 프레임: 다음 프레임 진행 전에는 계산 불가 -> '9.1'에서 10점 + 10점 = 20점(총 97점)
     *     1. 6개의 핀 쓰러뜨림
     *     2. 4개의 핀 쓰러뜨림(스페어)
     * 9. 프레임: 다음 프레임 진행 저에는 계산 불가 -> '10.1'과 '10.2'에서 10점 + 10점 = 20점(총 117점)
     *     1. 10개의 핀 쓰러뜨림(스트라이크)
     * 10. 프레임: 다음 투구 진행 저에는 계산 불가 -> '10.3'에서 6점 + 10점 = 16점(총 133점)
     *     1. 2개의 핀 쓰러뜨림
     *     2. 8개의 핀 쓰러뜨림(스페어)
     *     3. (스페어 점수 계산 위해 투구) 6개의 핀 쓰러뜨림
     * }
     * </pre>
     */
    @Test
    public void typicalGame() {
        Game_26_15 game = new Game_26_15();
        // 1. 프레임
        game.roll(1);
        game.roll(4);
        // 2. 프레임
        game.roll(4);
        game.roll(5);
        // 3. 프레임(스페어)
        game.roll(6);
        game.roll(4);
        // 4. 프레임(스페어)
        game.roll(5);
        game.roll(5);
        // 5. 프레임(스트라이크)
        game.roll(10);
        // 6. 프레임
        game.roll(0);
        game.roll(1);
        // 7. 프레임(스페어)
        game.roll(7);
        game.roll(3);
        // 8. 프레임(스페어)
        game.roll(6);
        game.roll(4);
        // 9. 프레임(스트라이크)
        game.roll(10);
        // 10. 프레임(스페어)
        game.roll(2);
        game.roll(8);
        game.roll(6);

        assertThat(game.score()).isEqualTo(133);
    }
}
