package support;

import java.util.ArrayDeque;
import java.util.Queue;

/**
 * LeetCode 이진 트리 문제 공용 노드 + 레벨 순서 빌더.
 * <p>
 * LeetCode 본 사이트의 TreeNode와 필드/생성자는 같고, 로컬 하니스를 위해
 * {@link #of(Integer...)}(레벨 순서 배열 -> 트리)만 더했습니다. LeetCode가 트리를
 * {@code [3,9,20,null,null,15,7]}처럼 레벨 순서(빈 자식은 null)로 주는 방식 그대로입니다.
 * 제출 시에는 import와 {@code support.} 접두어를 빼면 본 사이트 TreeNode로 그대로 동작합니다.
 */
public class TreeNode {
    public int val;
    public TreeNode left;
    public TreeNode right;

    public TreeNode() {}

    public TreeNode(int val) {
        this.val = val;
    }

    public TreeNode(int val, TreeNode left, TreeNode right) {
        this.val = val;
        this.left = left;
        this.right = right;
    }

    /**
     * of(3, 9, 20, null, null, 15, 7) -> 레벨 순서로 트리를 만든다. null은 빈 자식이고,
     * of()나 of(null)은 빈 트리(null)입니다.
     */
    public static TreeNode of(Integer... vals) {
        if (vals.length == 0 || vals[0] == null) {
            return null;
        }
        TreeNode root = new TreeNode(vals[0]);
        Queue<TreeNode> queue = new ArrayDeque<>();
        queue.add(root);
        int i = 1;
        while (i < vals.length) {
            TreeNode node = queue.poll();
            if (i < vals.length) {
                if (vals[i] != null) {
                    node.left = new TreeNode(vals[i]);
                    queue.add(node.left);
                }
                i++;
            }
            if (i < vals.length) {
                if (vals[i] != null) {
                    node.right = new TreeNode(vals[i]);
                    queue.add(node.right);
                }
                i++;
            }
        }
        return root;
    }
}
