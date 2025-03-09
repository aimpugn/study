package craftsmanship.c;

public class Game_24_14 {
    private int score;

    public void roll(int pins) {
        // 점수는 roll 함수에서 계산되므로, 이 함수가 스페어를 고려하도록 수정합니다.

        // `lastPins` 벼수는 `Game` 클래스의 멤버 변수로 선언하여, 이전 투구 점수를 기억하게 합니다.
        // 만약 '이전 투구'와 '현재 투구'의 점수가 합해서 10이면 스페어입니다.
        // ```
        // if (pins + lastPins == 10) {
        //     // 아무도 이해 못할 코드
        // }
        // ```
        // 아마 이때 "이건 완전히 틀렸다"라는 기분이 들 것이라 하고, 실제로 그런 느낌이 듭니다.
        //
        // 저자는 그 느낌이 맞으며, "설계에 결함이 있다"고 합니다.
        // 이름에 따르면 점수를 계산해야 하는 함수는 `score()` 함수입니다.
        // 하지만 실제로 점수를 계산하는 것은 `roll()` 함수입니다.
        // 이것을 '책임 배치 오류(misplaced responsibility)'라고 합니다.
        score += pins;
    }

    public int score() {
        return score;
    }
}
