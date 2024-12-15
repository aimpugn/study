package p42884;

import java.util.*;

/**
 * [단속카메라](https://school.programmers.co.kr/learn/courses/30/lessons/42884?language=java)
 */
class Solution {
    /**
     * 고속도로 이용하는 모든 차량이 고속도로 이용하면서 단속용 카메라를 한 번은 만나도록 설치
     * <p>
     * 차량의 진입/진출 지점에 카메라가 설치되어 있어도 카메라를 만난 것으로 간주
     *
     * @param routes 고속도로 이용하는 차량의 경로
     *               - 1 <= routes.length <= 10,000
     *               - routes[i][0]: i번째 차량이 고속도로에 진입한 지점
     *               - routes[i][1]: i번째 차량이 고속도로에서 나간 지점
     *               - -30,000 <= routes[i][0,1] <= 30,000
     * @return 모든 차량이 한 번은 단속용 카메라 만나도록 하는 카메라 최소 대수
     */
    public int solution(int[][] routes) {
        int answer = 0;

        // 예제:
        // [진입, 진출]
        // [[-20,-15],[-14, -5],[-18,-13],[-5,-3]]
        //
        // -20 -> -15
        //    -18 --> -13
        //          -14 -> -5
        //                 -5 -> -3
        //
        // - -5에 설치: routes[1], routes[3] 두 차량 커버 가능
        // - -15 또는 -18에 설치: routes[0], routes[2] 두 차량 커버 가능
        //
        // [[-10, 10], [-5, -3], [9, 11], [12, 15]]
        // -10 ------------------------> 10
        //       -5 --> -3
        //                             9 --> 11
        //                                      12 ---> 15
        // - -10에 설치: routes[0] 차량 커버
        // - -5에 설치: routes[0], routes[1] 차량 커버
        // - -3에 설치: routes[0], routes[1] 차량 커버
        // - 10에 설치: routes[0], routes[2] 차량 커버
        // - 12에 설치: routes[3] 커버 가능
        //
        // 최소를 구하는 문제니, 진입점 또는 진출점 하나에 설치하도록 합니다.
        // 겹치는 게 많을수록 좋습니다.
        // 만약 겹치는 게 없다면 둘 중 한 곳에만 설치합니다.
        //
        // 1. routes[i][0]으로 겹치는 게 있는지 검사하고, routes[i][1]로 겹치는 게 있는지 검사해서 더 많은 쪽을 선택? => X
        // - 선형적으로 두 번 탐색해야 하고
        // - routes[i][1]에서 겹치는 게 많았다고 해도 그게 최선이라고 보장할 수 있는지?
        //
        // 2. 각 지점마다 목록을 두고, 커버되는 차량 목록을 append? => X
        // map = [
        //  -10 => [[-10, 10]],
        //  -5 => [[-10, 10], [-5, -3]],
        //  -3 => [[-10, 10], [-5, -3]],
        //  9 => [[-10, 10], [9, 11]],
        //  10 => [[-10, 10], [9, 11]],
        //  11 => [[9, 11]],
        //  12 => [[12, 15]],
        //  15 => [[12, 15]],
        // ]
        // -5, 9, 12에 설치하면 커버 가능
        // 최총 3 리턴
        //
        // 하지만,
        // [[-20,-15],[-14, -5],[-18,-13],[-5,-3]] 경우
        // ... 생략 ...
        // -13 [[-14, -5], [-18, -13]]
        // -14 [[-14, -5], [-18, -13]]
        // -5 [[-14, -5], [-5, -3]]
        // ... 생략 ...
        // 이런 경우 단순히 길이로만 확인하면 -13에 설치할 경우 -5에도 설치할 수가 있습니다.
        //
        // 3. 결국 "겹치는 구간 중 어디에 카메라를 설치해야 하는가?" 진입점 또는 진출점 둘 중 한 곳에 설치한다고 가정 => O
        // 진입점 또는 진출점 둘 중 하나에 설치해야 합니다.
        // 진입점에 설치하는 것은, 중복해서 커버할 수 없는 것으로 보여서 실익이 없어 보입니다.
        // 반면 진출점에 설치하면, 중복에 대해 처리할 수 있어 보입니다. 진출점 기준으로 정렬한 경우:
        //
        // [[-20,-15],[-14, -5],[-18,-13],[-5,-3]]
        // [[-20,-15],[-18,-13],[-14, -5],[-5,-3]]
        // -20 -----> -15
        //     -18 -------> -13
        //               -14 -------> -5
        //                            -5 ----> -3
        // -15 설치하면: [-20,-15], [-18,-13] 커버 가능
        // -5 설치하면: [-14, -5], [-5,-3] 커버 가능
        //
        // [[-10, 10], [-5, -3], [9, 11], [12, 15]]
        // [[-5, -3], [-10, 10], [9, 11], [12, 15]]
        //       -5 ----> -3
        // -10 -------------------------> 10
        //                            9 ------> 11
        //                                         12 -----> 15
        // -3 설치하면: [-10, 10], [-5, -3] 커버 가능
        // 11 설치하면: [9, 11] 커버 가능
        // 15 설치하면: [12, 15] 커버 가능 => 마지막이라고 굳이 12를 검토할 필요는 없습니다. 만약 겹쳤다면 이미 처리 됐을 것이기 때문입니다.

        // 진출점 기준으로 정렬합니다
        Arrays.sort(routes, (a, b) -> Integer.compare(a[1], b[1]));
        // System.out.println(Arrays.deepToString(routes));
        boolean[] covered = new boolean[routes.length];

        for (int i = 0; i < routes.length; i++) {
            if (covered[i]) continue;

            int exitPoint = routes[i][1];
            // System.out.println("i: " + i + ", exitPoint: " + exitPoint);

            for (int j = 0; j < covered.length; j++) {
                if (!covered[j]) {
                    // [100, -10]처럼 진입을 양수로 하고, 진출을 음수로 하는 역방향도 고려해야 하나 싶어서 복잡해졌지만,
                    // 이것도 결국 그냥 Math.{max,min}으로 뒤집기만 하면, 어차피 방향을 고려하지 않으니 [-10, 100]처럼 비교합니다.
                    int maxVal = Math.max(routes[j][0], routes[j][1]);
                    int minVal = Math.min(routes[j][0], routes[j][1]);
                    if (minVal <= exitPoint && exitPoint <= maxVal) {
                        // System.out.println("\tj: " + j);
                        covered[j] = true;
                    }
                }
            }
            answer++;
        }

        return answer;
    }

    public int solution2(int[][] routes) {
        Arrays.sort(routes, (a, b) -> Integer.compare(a[1], b[1]));

        int answer = 0;
        int lastCamera = Integer.MIN_VALUE;

        for (int[] route : routes) {
            int entryPoint = route[0];
            int exitPoint = route[1];

            // 진출점 기준으로 정렬을 했으므로, lastCamera 변수에는 가장 작은 값이 설정됩니다.
            // 진출점의 카메라와 엔트리를 비교하는 이유는, 현재 설정된 카메라가 가장 빠른 진출점이기 때문입니다.
            // 가령:
            // 1:      -5 -- -3
            // 2: -10 ------------------ 10
            // 이 경우, 1번 차량 경우에 -3에 카메라를 설치됐습니다.
            // 진출점 기준으로 정렬됐으므로, 2번 차량을 검사할 때 2번 차량의 exit은 1번 차량보다 이후입니다.
            // 따라서 2번 차량의 진입점만 현재 설치된 카메라 이전이면, 현재 설치된 카메라로 커버가 가능합니다.
            //
            // 1:      -5 -- -3
            // 2:              -2 -------- 10
            // 반면 이 경우에는 entry가 -3보다 크므로 현재 카메라로는 커버가 불가능하고, 10번에 설치합니다.
            if (lastCamera < entryPoint) {
                // [[-10, 10], [-5, -3], [9, 11], [12, 15]]
                // [[-5, -3], [-10, 10], [9, 11], [12, 15]] 로 정렬되고,
                // 1. lastCamera=-3, answer=1
                // 2. lastCamera=11, answer=2
                // 3. lastCamera=15, answer=3
                answer++;
                lastCamera = exitPoint; // 새로운 카메라 위치는 현재 차량의 진출점
                System.out.println("lastCamera: " + lastCamera);
            }
        }

        return answer;
    }
}

class TestCase {
    public int[][] routes;
    public int answer;

    TestCase(int[][] routes, int answer) {
        this.routes = routes;
        this.answer = answer;
    }
}

public class Main {
    public static void main(String[] args) {
        Solution s = new Solution();

        List<TestCase> cases = new ArrayList<>();
        cases.add(new TestCase(new int[][]{
                new int[]{-20, -15},
                new int[]{-14, -5},
                new int[]{-18, -13},
                new int[]{-5, -3},
        }, 2));
        // [[-10, 10], [-5, -3], [9, 11], [12, 15]]
        cases.add(new TestCase(new int[][]{
                new int[]{-10, 10},
                new int[]{-5, -3},
                new int[]{9, 11},
                new int[]{12, 15},
        }, 3));

        cases.forEach((tc) -> {
            System.out.println(s.solution(tc.routes) == tc.answer);
        });

    }
}
