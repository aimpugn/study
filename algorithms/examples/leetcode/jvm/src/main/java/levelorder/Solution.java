package levelorder;

import support.Judge;
import support.TreeNode;

import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.List;

/**
 * <a href="https://leetcode.com/problems/binary-tree-level-order-traversal/">Binary Tree Level Order Traversal</a>
 * <p>
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md에 작성합니다.
 */
class Solution {
    /**
     * @param root 이진 트리 root
     * @return the level order traversal of its nodes' values
     * 즉, 레벨 순서대로 순회한 결과(왼쪽에서 오른쪽, 레벨별로)
     */
    public List<List<Integer>> levelOrder(TreeNode root) {
        // 첨삭: 정답입니다 — BFS 레벨 순회의 교과서 구현이라 더 손댈 게 없습니다. 그리고 방금 104에서
        //   배운 두 가지가 여기서 그대로 옳게 쓰였습니다.
        //
        //   첫째, size 스냅샷(while 진입 직후의 var size = queue.size())이 "while 한 바퀴 = 정확히 한
        //   레벨"을 보장합니다. 그 바퀴 동안 모은 list가 곧 한 레벨이라 answer에 레벨이 하나씩 쌓입니다.
        //   104에서는 같은 한 바퀴로 깊이만 셌고, 여기서는 그 바퀴 안에서 값을 list로 모아 answer에 넣는
        //   한 가지만 더한 것입니다.
        //
        //   둘째, deque를 큐로 일관되게 썼습니다. offer로 꼬리에 넣고 poll로 머리에서 빼니(FIFO) 먼저
        //   들어온 레벨이 먼저 나오고, 한 레벨 안에서도 왼쪽 자식을 오른쪽보다 먼저 넣어 좌->우 순서가
        //   지켜집니다. 104에서 헷갈리던 양 끝(push는 머리, offer/poll은 큐)을 여기선 큐 한 가지로만 써서
        //   그 혼란이 사라졌습니다.
        //
        //   list와 answer가 자라는 그림 ([3,9,20,null,null,15,7]):
        //     바퀴1: size=1, list=[3]     -> answer=[[3]],                큐에 9,20
        //     바퀴2: size=2, list=[9,20]  -> answer=[[3],[9,20]],         큐에 15,7
        //     바퀴3: size=2, list=[15,7]  -> answer=[[3],[9,20],[15,7]],  큐 빔
        //
        // > 카드: "레벨 경계"가 필요하면 BFS while 안에서 size를 스냅샷하고 그만큼만 for 돈다. 이 한 틀이 레벨 순회(102), 우측면도(199=각 레벨 마지막), 지그재그(103=홀짝 레벨 뒤집기), 레벨 최댓값(515)으로 그대로 전이한다.
        // > 대안: DFS(전위 + 레벨 인덱스)도 같은 레벨 순회. 보조 공간이 큐(너비) 대신 호출 스택(높이)이라 균형 트리면 O(log n)로 더 작고 치우치면 O(n)로 더 큼. 102는 출력이 어차피 O(n)이라 보통 읽기 쉬운 BFS가 기본이고, 보조 공간이 빠듯할 때 DFS를 고른다.
        // 레벨별로, 그리고 left right 순서로 값을 모아야 합니다.
        // bfs가 적절해 보입니다.
        if (root == null) return new ArrayList<>();

        var queue = new ArrayDeque<TreeNode>();
        queue.offer(root);
        var answer = new ArrayList<List<Integer>>();

        while (!queue.isEmpty()) {
            var list = new ArrayList<Integer>();
            answer.add(list);

            // 큐에 노드가 계속 추가되므로, 루프를 순회해야 하는 사이즈는 이 순간에 캡처합니다.
            var size = queue.size();
            for (var i = 0; i < size; i++) {
                var curr = queue.poll();
                list.add(curr.val);
                if (curr.left != null) queue.offer(curr.left);
                if (curr.right != null) queue.offer(curr.right);
            }
        }

        return answer;
    }

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(s.levelOrder(TreeNode.of(3, 9, 20, null, null, 15, 7)), List.of(List.of(3), List.of(9, 20), List.of(15, 7)));
        Judge.check(s.levelOrder(TreeNode.of(1)), List.of(List.of(1)));
        Judge.check(s.levelOrder(TreeNode.of()), List.of());
        // 반례를 여기에 추가하세요:
    }
}
