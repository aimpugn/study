# process

method 주석에 문제를 분석하고, 바로 문제 해결을 시작했습니다.
그러다 중간에 멈추게 됐는데,
1. 30 x 70 -> 70 x 30으로 뒤집어서 더 작은 높이를 찾는 로직을 어떻게 해야 할지
2. 그리고 가로 기준으로 했을 때 세로를 뒤집어 본다면, 세로 기준으로 해서 가로도 뒤집어 봐야 하는지

이런 정제되지 않은 생각들이 떠오르기 시작했기 때문입니다.

```java
public class Main {

    public static void main(String[] args) {
        System.out.println(Main.solution(
                new int[][] {{60, 50}, {30, 70}, {60, 30}, {80, 40}}) == 400);

    }

    /**
     * <pre>
     * 목표: 모든 명함을 수납할 수 있는 가장 작은 지갑 만들기
     *
     * 지갑 크기 정하기
     * - 다양한 모양과 크기 명함 수납 가능
     * - 작아서 들고 다니기 편해야 함
     *
     * 가로 x 세로
     * 60 x 50
     * 30 x 70
     * 60 x 30
     * 80 x 40
     *
     * 가장 긴 가로 길이와 세로 길이가 각각 80, 70이므로 80 * 70 크기의 지갑을 만들면 모든 명함 수납 가능.
     * 하지만 30 x 70 경우 가로로 눕혀 수납할 수 있음.
     * 2번 케이스에 가로로 수납한다고 하면 80 x 50 크기의 지갑ㅂ으로 모든 명함 수납 가능.
     *
     * </pre>
     *
     * @param sizes 모든 명함의 가로 x 세로 길이 나타내는 배열. [{width, height},..] 형식. 1 <=
     *        sizes.length <= 10,000
     * @see <a href=
     *      "https://school.programmers.co.kr/learn/courses/30/lessons/86491">최소
     *      직사각형</a>
     * @return int 지갑의 크기
     */
    public static int solution(int[][] sizes) {
        // 가장 짧은 가로 길이,
        // 가잘 짧은 세로 길이
        int minWidth = Integer.MIN_VALUE;
        int minHeight = Integer.MIN_VALUE;

        for (int[] size : sizes) {
            int width = size[0];
            int height = size[1];

            // 가로는 유지한다
            minWidth = Math.max(minWidth, width);
            minHeight = Math.max(minHeight, height);
            // 세로는 한번 뒤집어 본다?

        }

        System.out.println(minWidth + " * " + minHeight);


        return minWidth * minHeight;
    }
}
```

돌이켜 보면 바로 문제 풀이를 시작하지 말고 추상화와 설계 단계를 거쳐야 생각을 정제하는 시간이 필요해 보입니다.
정제된 생각과 논리를 통해 올바른 문제 해결이 가능할 것으로 보이기 때문입니다.

해결해야 할 문제를 다시 생각해 보면,
'회전 가능한 직사각형들을 모두 포함할 수 있는 최소 크기의 직사각형 찾기' 입니다.
- 회전 가능한

    각 명함은 가로 또는 세로로 놓을 수 있습니다.
    따라서 각 명함의 [width, height]와 [height, width] 모두 고려해야 합니ㅏㄷ.

- width, height 라는 용어의 한계

    사실 너비와 높이, 가로와 세로라고 생각하지만, 사실 항상 그렇게 고정되어 있는 것은 아닙니다.

    ```sh
     ┌───┐   ┌──────┐
    2│   │  1│      │
     │   │   └──────┘
     └───┘      2
       1
    ```

    이렇게 보면 좌측은 가로가 짧지만 우측은 가로가 깁니다. 하지만 둘 다 크기는 2입니다.
    따라서 가로, 세로라는 개념 대신 "긴 쪽"과 "짧은 쪽"이라고 생각했으면 더 좋았을 거 같습니다.

    NOTE: 문제에 주어진 단어나 개념에 사고가 제한되는 경향이 있는데, 주의해야겠습니다.

- 모두 포함할 수 있는 최소 크기의 직사각형

    앞서 가로, 세로 대신 긴 쪽과 짧은 쪽으로 개념을 바꿔서 생각하기로 했습니다.

    따라서, "모두 포함"해야 하므로
    - ~~가로든 세로든 한 쪽을 기준으로 잡았을 때, 최대값 이상이어야 합니다~~ 라고 생각하는 대신
    - 긴 쪽은 긴 값들 중에 최대값 이상이고, 짧은 쪽은 짧은 값들 중에 최대값 이상이어야 합니다.

```java
if (sizes.length == 1) {
    return sizes[0][0] * sizes[0][1];
}

// 가장 짧은 가로 길이,
// 가잘 짧은 세로 길이
int maxLong = Integer.MIN_VALUE;
int maxShort = Integer.MIN_VALUE;

for (int[] size : sizes) {
    int longSide = Math.max(size[0], size[1]);
    int shortSide = Math.min(size[0], size[1]);

    maxLong = Math.max(maxLong, longSide);
    maxShort = Math.max(maxShort, shortSide);
}

return maxLong * maxShort;
```
