package craftsmanship.c;

public class Game_24_55 {
    private final int[] rolls = new int[21];
    private int currentRoll;

    public void roll(int pins) {
        // 기존에 `roll` 함수에서 수행하던 점수 계산을 `score` 함수로 이동시킵니다.
        // 투구 결과를 다른 함수에서 계산하려면, 투구 결과를 저장해둬야 하고,
        // 이를 위해 `rolls` 배열을 사용합니다.
        rolls[currentRoll++] = pins;
    }

    public int score() {
        int score = 0;

        int i = 0;
        for (int frame = 0; frame < 10; frame++) {
            // 이 코드는 `i`가 짝수일 때만 올바릅니다.
            // > 스페어: 공 두 개를 다 써서 10개의 핀을 모두 쓰러뜨리는 경우
            //
            // 하지만 여전히 뭔가 좋지 않다는 느낌이 듭니다.
            // 이 경우 규칙 6 "코드가 틀렸다고 느껴지면 잠시 멈춰서 설계를 고치라."를 따릅니다.
            // ```
            // if (i % 2 == 0 && rolls[i] + rolls[i + 1] == 10) {
            //     // 스페어 처리?
            // }
            // ```

            // 원래 UML 설계를 보면, `Game`은 10개의 `Frame` 인스턴스를 갖습니다.
            // 그렇다면 `rolls.length`(여기서 21) 만큼 반복문을 순회하는 것이 옳지 않아 보입니다.
            // 볼링 경기는 10개의 프레임으로 이뤄집니다.
            // 그러니 코드 어디에도 "10"이라는 숫자가 안 보이는 것에서 이상함을 느껴야 합니다.
            // 최초의 설계대로라면 반복문에서 '한 번에 한 프레임씩 `rolls` 배열을 처리'해야 합니다.
            //
            // 하지만 이렇게 한 번에 배열에서 투구 기록을 두 개씩 반복해서 읽는 방법은 틀렸습니다.
            // 스트라이크는 한 프레임에 투구가 한 번밖에 없고,
            // 10 프레임 경우에는 투구가 세 번 있을 수도 있기 때문입니다.
            // 하지만 일단 "프레임당 투구 두 번 전략"은 잘 동작합니다.
            score += rolls[i] + rolls[i + 1];

            i += 2;
        }

        return score;
    }
}
