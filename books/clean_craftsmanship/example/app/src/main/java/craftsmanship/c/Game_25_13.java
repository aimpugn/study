package craftsmanship.c;

import java.util.function.Predicate;

public class Game_25_13 {
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

        Predicate<Integer> isSpare = frameIdx -> rolls[frameIdx] + rolls[frameIdx + 1] == 10;

        int frameIdx = 0;
        for (int frame = 0; frame < 10; frame++) {
            // 아직까지는 "프레임당 투구 두 번 전략"에 따라 구현합니다.
            // if (rolls[frameIdx] + rolls[frameIdx + 1] == 10) { // 스페어에 대한 처리
            if (isSpare.test(frameIdx)) { // 스페어 여부를 테스트하는 `Predicate`으로 추출하고 "스페어에 대한 처리" 주석을 삭제합니다.
                score += 10 + rolls[frameIdx + 2];
            } else {
                score += rolls[frameIdx] + rolls[frameIdx + 1];
            }

            frameIdx += 2;
        }

        return score;
    }
}
