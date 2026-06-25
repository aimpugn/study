package support;

import java.util.ArrayList;

/**
 * LeetCode 연결 리스트 문제 공용 노드 + 배열 변환 헬퍼.
 * <p>
 * LeetCode 본 사이트의 ListNode와 필드/생성자는 같고, 로컬 하니스를 위해
 * {@link #of(int...)}(배열 -> 리스트)와 {@link #toArray(ListNode)}(리스트 -> 배열)만 더했습니다.
 * 비교는 {@code Judge.check(ListNode.toArray(actual), new int[]{...})}로 합니다.
 * 제출 시에는 import와 {@code support.} 접두어를 빼면 본 사이트 ListNode로 그대로 동작합니다.
 */
public class ListNode {
    public int val;
    public ListNode next;

    public ListNode() {}

    public ListNode(int val) {
        this.val = val;
    }

    public ListNode(int val, ListNode next) {
        this.val = val;
        this.next = next;
    }

    /** of(1, 2, 3) -> 1 -> 2 -> 3. 빈 호출 of()는 빈 리스트(null)입니다. */
    public static ListNode of(int... vals) {
        ListNode dummy = new ListNode();
        ListNode cur = dummy;
        for (int v : vals) {
            cur.next = new ListNode(v);
            cur = cur.next;
        }
        return dummy.next;
    }

    /** 연결 리스트를 값 배열로 풀어 비교·출력에 씁니다. null이면 빈 배열입니다. */
    public static int[] toArray(ListNode head) {
        var values = new ArrayList<Integer>();
        for (ListNode node = head; node != null; node = node.next) {
            values.add(node.val);
        }
        int[] array = new int[values.size()];
        for (int i = 0; i < array.length; i++) {
            array[i] = values.get(i);
        }
        return array;
    }
}
