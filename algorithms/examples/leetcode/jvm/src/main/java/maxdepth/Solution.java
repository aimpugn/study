package maxdepth;

import support.Judge;
import support.TreeNode;

/**
 * <a href="https://leetcode.com/problems/maximum-depth-of-binary-tree/">Maximum Depth of Binary Tree</a>
 * <p>
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md에 작성합니다.
 */
class Solution {
    /**
     *
     * @param root 이진 트리의 루트
     * @return 최대 깊이
     */
    public int maxDepth(TreeNode root) {
        // 첨삭: 정답입니다(통과). 마음에 안 드는 정체는 "과한 구조"입니다. 트리 깊이의 가장 자연스러운
        //   도구인 재귀를 안 쓰고, 레벨을 전부 deque에 쌓아 세는 반복 BFS로 풀었습니다. 결과는 맞지만
        //   코드가 길고(while(true)+break, 빈 레벨을 한 번 더 쌓고 size-1), 공간도 모든 노드를 동시에
        //   들고 있어 O(n)입니다.
        //
        //   바로 아래 "curr=curr.next처럼 하면 백트래킹이 안 된다"는 주석이 간극의 핵심 단서입니다. 그
        //   관찰 자체는 정확합니다. 연결 리스트는 한 줄이라 포인터만 밀면 됐지만, 트리는 갈래가 있어
        //   한쪽을 다 본 뒤 갈림길로 되돌아와야 하니까요. 한 걸음 어긋난 곳은 그 "되돌아오기"를 직접
        //   자료구조로 관리하려 한 지점입니다. 되돌아오기는 재귀 호출 스택이 대신합니다. 왼쪽
        //   서브트리를 끝내고 함수가 반환되면, 제어가 자동으로 그 갈림길로 돌아와 오른쪽을 잇습니다.
        //
        //   씨앗 한 문장: 트리는 "노드 + 왼쪽 서브트리 + 오른쪽 서브트리"로 자기 자신으로 정의됩니다.
        //   정의가 재귀적이면 풀이도 재귀가 기본입니다. 깊이도 마찬가지여서, 한 노드의 깊이는
        //   1 + (두 서브트리 깊이 중 큰 쪽)입니다. 직전 연결 리스트(206, 21)의 "반복 + 포인터" 감각이
        //   남아 트리에도 그대로 끌어온 게 이번 미스이고, 다음에 잡을 신호가 여기 있습니다.
        //
        // > 카드: 자료구조가 자기 자신으로 정의되면(트리=노드+서브트리들) 반복+스택 직접 관리 말고 재귀부터. 백트래킹(되돌아오기)은 콜스택이 떠안는다.
        // 최대 깊이를 알려면 끝까지 타고 들어갈 수 있어야 합니다.
        // 그런데 어떤 노드를 통해서 끝까지 갈 수 있는지는 모두 가봐야 알 수 있습니다.
        // curr = curr.next 처럼 하면 백트래킹이 안 됩니다.
        // 결국에는 반복문, 조건, 자료구조.

        // 현재 뎁스를 알아야 합니다.
        // 그리고 인덱스로 알 수 있습니다.
        // 첨삭: 이 deque 사용에 헷갈림이 그대로 드러납니다. push와 add, getLast를 섞어 썼습니다.
        //   ArrayDeque의 양 끝 규약은 이렇습니다(헷갈리던 그 지점):
        //     push / pop / peek    : 머리(head, 앞)에서 동작. 스택(LIFO)으로 쓸 때.
        //     offer / add          : 꼬리(tail)에 넣음.  poll / remove : 머리에서 뺌. 큐(FIFO)로 쓸 때.
        //     getFirst / peekFirst : 머리.  getLast / peekLast : 꼬리.
        //   여기선 첫 레벨을 push(머리)로 넣고 이후 레벨은 add(꼬리) + getLast(꼬리)로 봅니다. 머리에
        //   하나, 꼬리에 여럿이 되는데 getLast가 늘 "가장 최근"을 잡아 우연히 맞게 돕니다. 동작은 하지만
        //   한 컨테이너에서 스택 끝과 큐 끝을 섞은 게 "마음에 안 듦"의 한 원인입니다. 한 가지로 통일하면
        //   (아래 베스트의 큐: offer로 넣고 poll로 빼기) 헷갈림이 사라집니다.
        //
        //   그리고 레벨을 전부 쌓을 필요가 없습니다. 깊이만 세면 되니 "지금 레벨"과 카운터 하나면
        //   충분합니다. size-1을 하는 이유도 빈 레벨을 한 번 더 쌓고 break하기 때문인데, 카운터 방식엔
        //   그 군더더기가 없습니다.
        /*
        if (root == null) return 0;

        var stack = new ArrayDeque<List<TreeNode>>();
        stack.push(List.of(root));

        while (true) {
            var curr = stack.getLast();
            var list = new ArrayList<TreeNode>();

            stack.add(list);
            // System.out.println("stack: " + stack);

            for (var node : curr) {
                if (node.left != null) {
                    list.add(node.left);
                }
                if (node.right != null) {
                    list.add(node.right);
                }
            }

            if (list.isEmpty()) {
                break;
            }
        }*/
        // System.out.println("queue size: " + stack.size());

        if (root == null) return 0;                                     // 빈 트리는 0 (재귀 바닥)
        return 1 + Math.max(maxDepth(root.left), maxDepth(root.right)); // 나 1 + 더 깊은 자식
    }

    // 학습자 접근을 살린 버전(반복 BFS, 제대로): deque를 "큐" 한 가지로만 쓰고(offer로 꼬리에 넣고
    //   poll로 머리에서 뺌), 레벨을 다 저장하는 대신 "이번 레벨 노드 수"만큼만 처리해 깊이를 한 칸
    //   올립니다. 레벨을 통째로 들고 다니지 않아 공간이 O(너비)로 줄고, push/add 혼용도 사라집니다.
    //
    // public int maxDepth(TreeNode root) {
    //     if (root == null) return 0;
    //     var queue = new ArrayDeque<TreeNode>();
    //     queue.offer(root);                       // 큐: 꼬리에 넣는다
    //     int depth = 0;
    //     while (!queue.isEmpty()) {
    //         depth++;                             // 새 레벨에 진입할 때마다 깊이 +1
    //         int size = queue.size();             // 지금 큐에 있는 게 딱 이번 레벨 (스냅샷)
    //         for (int i = 0; i < size; i++) {
    //             var node = queue.poll();         // 머리에서 뺀다
    //             if (node.left != null) queue.offer(node.left);
    //             if (node.right != null) queue.offer(node.right);
    //         }
    //     }
    //     return depth;
    // }
    //
    //   레벨로 보면 ([3,9,20,null,null,15,7]):
    //     depth 1: 큐=[3]      -> 3 빼고 9,20 넣음
    //     depth 2: 큐=[9,20]   -> 9는 자식 없음, 20이 15,7 넣음
    //     depth 3: 큐=[15,7]   -> 둘 다 자식 없음, 큐 빔
    //   큐가 빈 순간 depth=3입니다. size 스냅샷이 핵심입니다. 루프 도는 중에 다음 레벨이 큐에 섞여
    //   들어오니, 들어오기 전 크기를 미리 박아 "이번 레벨만" 처리합니다.

    // 베스트 프랙티스(재귀): 트리 깊이의 정의를 그대로 코드로 옮긴 것입니다. 한 노드의 깊이는
    //   1 + (왼쪽 깊이, 오른쪽 깊이 중 큰 쪽), 빈 트리는 0. 정의가 곧 코드라 두 줄입니다.
    //
    // public int maxDepth(TreeNode root) {
    //     if (root == null) return 0;                                     // 빈 트리는 0 (재귀 바닥)
    //     return 1 + Math.max(maxDepth(root.left), maxDepth(root.right)); // 나 1 + 더 깊은 자식
    // }
    //
    //   어떻게 도달하나 (외우지 말고 이 사고를 복제하세요):
    //     씨앗 — "깊이는 자기 자신으로 정의된다: 노드의 깊이 = 1 + 더 깊은 서브트리의 깊이."
    //     - 서브트리 깊이도 같은 함수로 구하면 된다       -> maxDepth(left), maxDepth(right) 재귀 호출.
    //     - 더 깊은 쪽이 내 깊이를 결정한다              -> Math.max(...).
    //     - 나 자신 한 층을 더한다                       -> 1 + ...
    //     - 노드가 없으면 셀 게 없다                      -> root == null이면 0 (바닥).
    //   24줄에서 걱정한 "되돌아오기"는 여기 안 보입니다. maxDepth(left)가 반환되는 순간 호출 스택이
    //   자동으로 그 갈림길로 돌아와 maxDepth(right)를 부르기 때문입니다. 그게 재귀가 트리에 맞는 이유입니다.
    //
    //   값이 바닥에서 올라오는 그림 ([3,9,20,null,null,15,7]):
    //         3            maxDepth(3) = 1 + max(1, 2) = 3
    //        / \
    //       9   20         maxDepth(9) = 1,  maxDepth(20) = 1 + max(1, 1) = 2
    //          /  \
    //        15    7       maxDepth(15) = 1,  maxDepth(7) = 1
    //   리프(9, 15, 7)가 1을 반환하고, 그 값이 부모로 올라가며 1씩 더해집니다. 잎에서 뿌리 방향입니다.
    //
    //   왜 더 나은가: 시간은 셋 다 O(n)(노드를 한 번씩 본다)이지만 코드 길이와 공간이 갈립니다.
    //     재귀는 두 줄에 공간 O(h)(h=높이, 호출 스택. 균형 트리면 O(log n), 한쪽으로 치우치면 O(n)),
    //     원래 풀이는 모든 레벨을 동시에 저장해 O(n)입니다. 트레이드오프: 트리가 한쪽으로 만(10,000)
    //     개까지 깊어지면 재귀는 호출 스택이 그만큼 쌓여 넘칠(StackOverflow) 위험이 있어, 그때는 위
    //     반복 BFS가 안전합니다. 실측: 공식 2개 + 빈/단일/좌치우침/우치우침에서 두 구현 모두 답 일치.
    //
    // > 불변식: maxDepth(node)는 그 노드를 뿌리로 보는 서브트리의 깊이를 반환한다 (자식 결과를 합쳐 올림).
    // > 카드: 트리 문제의 기본 골격은 "서브트리 결과를 받아 내 결과로 합쳐 올린다". 124(경로 합)도 같은 재귀에 "올리는 값 vs 갱신하는 global"만 더한 것.

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(s.maxDepth(TreeNode.of(3, 9, 20, null, null, 15, 7)), 3);
        Judge.check(s.maxDepth(TreeNode.of(1, null, 2)), 2);
        // 반례를 여기에 추가하세요:
    }
}
