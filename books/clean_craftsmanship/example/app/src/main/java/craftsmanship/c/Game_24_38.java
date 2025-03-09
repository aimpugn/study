package craftsmanship.c;

public class Game_24_38 {
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

        for (int i = 0; i < rolls.length; i++) {
            score += rolls[i];
        }

        return score;
    }
}
