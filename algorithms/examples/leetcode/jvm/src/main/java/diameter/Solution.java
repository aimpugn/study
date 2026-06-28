package diameter;

import support.Judge;
import support.TreeNode;

import java.util.ArrayDeque;

/**
 * <a href="https://leetcode.com/problems/diameter-of-binary-tree/">Diameter of Binary Tree</a>
 * <p>
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md에 작성합니다.
 */
class Solution {

    /**
     * - diameter, 지름의 길이란? 트리의 어떤 두 노드 간의 가장 긴 길이
     * - 이 경로는 root를 지날 수도, 지나지 않을 수도 있습니다.
     * - 두 노드 사이의 경로의 길이는 간선의 수로 나타냅니다.
     *
     * @param root 이진 트리의 루트
     * - 1 <= 노드 수 <= 10^4
     * - -100 <= node.val <= 100
     * @return 트리의 지름(diameter) 길이
     */
    public int diameterOfBinaryTree(TreeNode root) {
        // 첨삭: 정답입니다(통과). 마음에 안 든 정체는 O(n^2)이고, 그 직감이 맞습니다.
        //   아래쪽 "8부터 2 + 9부터 2를 합치면 최대" 주석이 이 문제의 핵심 통찰입니다. 한 노드를 통과하는
        //   가장 긴 경로 = (왼쪽으로 내려가는 최대 깊이) + (오른쪽으로 내려가는 최대 깊이)이고, 지름은 그걸
        //   모든 노드에서 잰 최댓값이죠. 그 통찰을 정확히 잡았고, maxDepth 재귀도 옳게 짰습니다.
        //
        //   간극은 그 둘을 한 번에 합치지 못한 데 있습니다. 지금은 BFS로 모든 노드를 한 번 돌고, 그 노드마다
        //   maxDepth(left)+maxDepth(right)를 다시 계산합니다. 그런데 maxDepth는 그때마다 서브트리를 통째로
        //   다시 훑어, 같은 노드의 깊이가 처음부터 몇 번씩 재계산됩니다.
        //
        //   정답 사고: 한 번의 재귀 호출이 두 가지를 동시에 만듭니다. 각 노드에서 왼쪽 높이(left)와 오른쪽
        //   높이(right)를 구하고 나면 그 두 숫자로 두 결과가 나옵니다 — (1) 내 높이 1+max(left,right)는
        //   부모가 자기 높이를 잴 때 쓰니 return으로 올려보내고, (2) 나를 통과하는 지름 left+right는 아무도
        //   안 받고 기록만 하면 되니 바깥 변수에 "지금까지의 최대"로 적어둡니다. 집배원이 한 집에 들러
        //   본사에 보고할 것(높이)과 수첩에 적어둘 것(지름)을 한 번 방문에 끝내는 것과 같습니다. 둘 다 같은
        //   left/right에서 나오고, 그 높이 DFS가 이미 모든 노드를 한 번씩 방문하니 한 순회면 O(n)입니다.
        //
        //   왜 이 간극인가: "각 노드를 root로 잡고 타고 들어가"는 프레임이 노드마다 독립 계산을 부르고,
        //   직전 102에서 BFS가 잘 먹혀 여기서도 BFS를 또 꺼낸 듯합니다. 다음에 잡을 신호: 한 노드에서
        //   만들 수 있는 값이 둘인데 용도가 다르면(하나는 부모가 이어 쓰고, 하나는 기록만 한다), 따로
        //   순회하지 말고 한 재귀에서 같이 만든다.
        //
        //   실측(노드 방문 수, 한쪽으로 치우친 n-노드 경로 트리): 이 풀이는 n=1000 -> 약 50만, 2000 -> 200만,
        //   4000 -> 800만 (2배마다 4배 = n^2). fused 한 패스는 n -> n (2배마다 2배). 균형 트리면 이 풀이도
        //   O(n log n)이지만, 한쪽으로 치우치면 worst O(n^2); fused는 트리 모양과 무관하게 O(n).
        //
        // > 카드: 트리 재귀 한 번에 두 가지를 만든다 — 부모가 이어 쓸 값(높이)은 return으로 올리고, 기록만 할 값(지름=좌+우)은 바깥 변수에 최대로 적는다. 같은 left/right로 둘 다 나오니 한 순회면 O(n).
        // 1. root = [1,2,3,4,5]
        //      4 -> 2 -> 1 -> 3
        //   OR 5 -> 2 -> 1 -> 3
        //   => 3
        //   어떤 가장 깊은 곳에서부터 시작하여 길이를 셉니다.
        // 2. root = [1, 2]
        //       2 -> 1
        //    => 1
        //
        // root를 거치지 않고 가장 긴 경우는 다음과 같습니다.
        //                 1
        //        2               3
        //     4     5
        //   6         7
        // 8              9
        //
        // [8, 6, 4, 2, 5, 7, 9]가 가장 길게 됩니다.
        // [8, 6, 4, 2, 1, 3]
        // [9, 7, 5, 2, 1, 3]
        //
        // 8부터 2까지의 경로, 9부터 2까지의 경로를 합치면 최대값이 됩니다.
        // 이 경우 root가 2가 됩니다.
        //
        // 일단 깊이를 세려면 dfs를 활용해야 할 것으로 보이고,
        // 최대 경로를 찾아야 하므로, 어떤 노드 기준으로 도달 가능한 모든 경우의 수를 담아야 할 거 같습니다.
        // 1번 노드 기준으로, 2번 노드 기준으로, 3번 노드 기준으로, 4번 노드 기준으로....
        //
        // 또는, 가장 끝 단, 가령 8, 9, 3에서 시작하여 [8, 9], [8, 3], [9, 8], [9, 3]
        // 이렇게 두 노드 사이를 도달할 수 있는 경로를 카운트합니다. 근데 이러면 8, 9, 3을 지정하여 찾아가야 해서
        // 오히려 문제 풀이가 어려울 거 같습니다.

        // 각 노드를 root로 잡고 타고 들어가 봅니다.

        var maxDepth = 0;
        var queue = new ArrayDeque<TreeNode>();
        queue.offer(root);

        while (!queue.isEmpty()) {
            var curr = queue.poll();

            // System.out.println("curr: " + curr.val);
            var maxLeft = maxDepth(curr.left);
            var maxRight = maxDepth(curr.right);

            maxDepth = Math.max(maxDepth, maxLeft + maxRight);

            if (curr.left != null) queue.offer(curr.left);
            if (curr.right != null) queue.offer(curr.right);
        }

        return maxDepth;
    }

    int maxDepth(TreeNode root) {
        if (root == null) return 0;
        var left = maxDepth(root.left);
        var right = maxDepth(root.right);
        // System.out.println("left: " + left + ", right: " + right);

        return 1 + Math.max(left, right);
    }

    // 베스트 프랙티스(한 패스 fused): 높이 DFS 한 번으로 높이와 지름을 동시에 구합니다. BFS도, 노드마다
    //   재계산도 없습니다. 재귀는 높이를 반환하고, 지름은 바깥 변수에 갱신합니다.
    //
    //   어떻게 도달하나 (외우지 말고 이 사고를 복제하세요):
    //     씨앗 — "높이 재귀는 이미 모든 노드를 방문한다. 방문하는 김에 그 노드의 좌+우를 지름에 갱신하자."
    //     - 한 노드의 높이 = 1 + max(왼쪽 높이, 오른쪽 높이)        -> 부모에게 돌려줄 값
    //     - 그 노드를 통과하는 지름 후보 = 왼쪽 높이 + 오른쪽 높이   -> 바깥 변수에 max로 갱신
    //     - 두 값은 같은 left, right로 만들어진다                    -> 한 번의 재귀에서 둘 다 나온다
    //   반환값(높이)과 갱신값(지름)을 가르는 게 핵심입니다. 헷갈려서 "지름을 반환"하면 무너집니다.
    //
    // private int diameter = 0;
    // public int diameterOfBinaryTree(TreeNode root) {
    //     diameter = 0;
    //     height(root);
    //     return diameter;
    // }
    // private int height(TreeNode node) {
    //     if (node == null) return 0;
    //     int left = height(node.left);
    //     int right = height(node.right);
    //     diameter = Math.max(diameter, left + right);  // 이 노드를 통과하는 지름 후보 (좌+우)
    //     return 1 + Math.max(left, right);             // 부모에겐 높이만 돌려준다
    // }
    //
    //   값이 바닥에서 올라오는 그림 ([1,2,3,4,5], 노드 1에서 지름 3이 잡힘):
    //         1            height(1)=1+max(2,1)=3,  지름후보 좌2+우1=3  (<- 최댓값)
    //        / \
    //       2   3          height(2)=1+max(1,1)=2,  지름후보 좌1+우1=2
    //      / \
    //     4   5            height(4)=height(5)=1,   지름후보 0
    //   리프 4,5가 높이 1을 올리고, 노드 2에서 좌+우=2, 노드 1에서 좌(2)+우(1)=3. 지름=3.
    //
    //   왜 더 나은가: 사용자 풀이는 노드마다 maxDepth를 다시 불러 같은 서브트리를 반복해 훑습니다(치우친
    //     트리 n=4000에서 노드 방문 약 800만). fused는 각 노드를 정확히 한 번 방문해 4000번. 시간이
    //     O(n^2 worst) -> O(n)으로 떨어집니다. 공간은 둘 다 호출 스택 O(높이).
    //
    // > 불변식: height(node)는 그 서브트리의 높이를 반환하고, 그 호출이 끝난 순간 diameter에는 지금까지 본 (좌높이 + 우높이) 최댓값이 들어 있다.
    // > 카드: 트리 최대경로 합(124)도 똑같은 골격이다 — height 재귀에 global 갱신을 얹고, 음수 자식 가지를 0으로 끊는 한 줄만 더하면 124가 된다.

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(s.diameterOfBinaryTree(TreeNode.of(1, 2, 3, 4, 5)), 3);
        Judge.check(s.diameterOfBinaryTree(TreeNode.of(1, 2)), 1);
        Judge.check(s.diameterOfBinaryTree(TreeNode.of(1, 2, 3, 4, 5, 6, null, null, 7, 8, null, null, 9)), 6);
        // 반례를 여기에 추가하세요:
    }
}
