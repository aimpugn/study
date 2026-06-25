package reverselinkedlist;

import support.Judge;
import support.ListNode;

import java.util.ArrayDeque;

/**
 * <a href="https://leetcode.com/problems/reverse-linked-list/">Reverse Linked List</a>
 * <p>
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md에 작성합니다.
 */
class Solution {
    /**
     *
     * @param head 단일 연결 리스트의 헤드
     * @return
     */
    public ListNode reverseList(ListNode head) {
        // 첨삭: 이 풀이는 정답입니다. 값을 스택에 쌓으면 LIFO(나중에 넣은 게 먼저 나옴)라, 꺼내는 순서가 곧 뒤집힌
        //   순서입니다. 포인터를 직접 뒤집는 어려움을 값 복사로 피해 갔고(대신 스택 + 새 노드로 O(n) 추가 공간),
        //   이게 아래 베스트(in-place)와의 차이입니다.
        //
        //   변수가 reversed와 currNode 둘로 나뉘는 이유가 헷갈리기 쉬운 지점입니다. 새 노드는 항상 "끝"에 붙이니
        //   (currNode.next = ...) 끝을 가리키는 손이 필요한데, 이 손은 붙일 때마다 뒤로 옮겨 갑니다. 정작 함수가
        //   돌려줄 건 "머리"라, 머리에 가만히 둘 손(reversed)과 끝으로 걷는 손(currNode)으로 나눕니다 —
        //   왜 하나로는 안 되는지는 아래 그림에서 드러납니다.
        //
        //   (입력 [1,2,3] -> 스택 pop은 3,2,1. '.'은 next 칸, '^'는 그 노드를 가리키는 변수)
        //
        //     reversed = new(3);  currNode = reversed;
        //       A[3|.]
        //       ^
        //       reversed, currNode              (둘 다 A를 가리킴)
        //
        //     currNode.next = new(2);  currNode = currNode.next;     // 끝에 B를 붙이고, currNode만 B로
        //       A[3|.] -> B[2|.]
        //       ^         ^
        //       reversed  currNode
        //
        //     currNode.next = new(1);  currNode = currNode.next;     // 끝에 C를 붙이고, currNode만 C로
        //       A[3|.] -> B[2|.] -> C[1|.]
        //       ^                   ^
        //       reversed            currNode
        //
        //     return reversed;     // A부터 = [3,2,1]
        //
        //   reversed는 머리 A에 붙박이고, currNode만 A -> B -> C로 걷습니다. currNode 하나만
        //   썼다면 마지막에 끝(C)에 서서 A를 잃었을 거예요. 그게 변수가 둘인 이유입니다.
        //
        //   "힙 참조"의 핵심도 이 그림 안에 있습니다. 변수는 노드 자체가 아니라, 힙에 있는 노드를 가리키는 화살표입니다.
        //     currNode = currNode.next : currNode 화살표만 다음 노드로 옮긴다 (reversed 화살표는 그대로).
        //     currNode.next = X        : currNode가 가리키는 노드의 next 칸을 X로 바꾼다 (그 노드 자체를 수정).
        //   화살표를 옮기는 것과, 화살표가 가리키는 노드를 고치는 것은 다릅니다.
        //
        // > 카드: 리스트를 만들 땐 손 둘 — 머리는 가만히(반환용), 끝을 가리키는 손만 전진(붙이기용).
        if (head == null || head.next == null) {
            return head;
        }

        var stack = new ArrayDeque<Integer>();

        while (head != null) {
            // 1 -> 2 -> 3... 으로 진행합니다.
            // 이걸 뒤집습니다.
            stack.push(head.val);
            head = head.next;
        }

        // System.out.println(stack);
        // 첨삭: 첫 노드만 따로(reversed = new ...) 만들어 시작한 게, 바로 아래 dummy 버전이 없애는 지점입니다.
        var reversed = new ListNode(stack.pop());
        var currNode = reversed;
        while (!stack.isEmpty()) {
            currNode.next = new ListNode(stack.pop());
            currNode = currNode.next;
        }
        // System.out.println(Arrays.toString(ListNode.toArray(reversed)));

        return reversed;
    }

    // 학습자 접근을 고친 동작 버전입니다. 원본이 머리(reversed)와 커서(currNode) 둘로 나눠 잡던 걸 dummy(가짜 머리)
    //   노드 하나, cur 하나로 줄였습니다. 진짜 머리는 항상 dummy.next라 머리를 따로 들 필요가 없고, 첫 노드 특수
    //   처리도 if 가드(빈/1개)도 사라집니다(빈 입력이면 스택이 비어 dummy.next가 null로 남음).
    //
    // public ListNode reverseList(ListNode head) {
    //     var stack = new ArrayDeque<Integer>();
    //     for (ListNode n = head; n != null; n = n.next) stack.push(n.val);
    //     ListNode dummy = new ListNode();   // 가짜 머리 — 진짜 머리는 dummy.next
    //     ListNode cur = dummy;              // 손은 cur 하나만
    //     while (!stack.isEmpty()) {
    //         cur.next = new ListNode(stack.pop());
    //         cur = cur.next;
    //     }
    //     return dummy.next;
    // }
    //
    // "리스트를 만들 땐 dummy 머리 + 커서 하나" — 위 그림에서 reversed 자리를 dummy가 대신 잡아 주는 셈입니다.
    //   두 리스트 병합(21), 재배열 등 리스트를 만드는 문제에 그대로 전이됩니다. 값을 복사하는 방식이라 공간은
    //   여전히 O(n). 예제 + 빈/1개/동값 통과 확인.

    // 베스트 프랙티스입니다. 새 리스트를 만들지 않고 기존 노드들의 next 화살표를 제자리에서 거꾸로 다시 이어
    //   공간이 O(1)입니다. 손 셋 — prev(뒤집은 부분의 머리), cur(지금 보는 노드), next(다음을 잃지 않게 미리 붙잡는 임시).
    //
    // > 불변식: cur 직전까지는 이미 다 뒤집혀 있고, prev가 그 뒤집힌 부분의 머리다.
    //
    // 어떻게 이 코드에 도달하나 (외우지 말고 이 사고를 복제하세요):
    //   씨앗 한 줄 — "뒤집기 = 각 노드가 [다음] 대신 [앞] 노드를 가리키게 한다." 나머지가 전부 여기서 따라 나옵니다.
    //   - 각 노드가 앞을 가리켜야 하니, 걸어가며 "내 앞 노드"를 들고 있어야 한다     -> prev가 필요하다.
    //   - 첫 노드는 뒤집으면 꼬리가 되고, 꼬리의 앞은 없다                         -> prev = null로 시작.
    //   - cur.next = prev로 돌리는 순간 원래 다음(나머지로 가는 길)을 잃는다        -> next = cur.next로 먼저 붙잡는다.
    //   - 다음 노드의 "앞"은 지금 cur, 다음에 볼 노드는 붙잡아 둔 next             -> prev = cur; cur = next.
    //   - cur이 null이면 끝, 그때 prev가 마지막 노드 = 새 머리                    -> return prev.
    //   초기값도 순서도 종료도 외운 게 아니라 씨앗에서 도출됩니다. 빈 화면에서 다시 쓸 땐 코드가 아니라 이 씨앗에서 출발하세요.
    //
    // public ListNode reverseList(ListNode head) {
    //     ListNode prev = null, cur = head;
    //     while (cur != null) {
    //         ListNode next = cur.next;  // (1) 다음을 먼저 붙잡는다 -- 이게 핵심
    //         cur.next = prev;           // (2) 화살표를 뒤로 돌린다
    //         prev = cur;                // (3) prev 한 칸 전진
    //         cur = next;                // (4) cur 한 칸 전진
    //     }
    //     return prev;
    // }
    //
    // 그림으로 (입력 [1,2,3]. 노드 사이 화살표 방향이 곧 next):
    //
    //   시작:    null    1 -> 2 -> 3 -> null
    //            ^       ^
    //            prev    cur
    //
    //   cur=1:   next=2 붙잡고 / 1.next=prev(null) / prev=1 / cur=2
    //            null <- 1    2 -> 3 -> null            // 1의 화살표가 뒤로 돌았다 (1 -> null)
    //                    ^    ^
    //                    prev cur
    //
    //   cur=2:   next=3 / 2.next=1 / prev=2 / cur=3
    //            null <- 1 <- 2    3 -> null
    //                         ^    ^
    //                         prev cur
    //
    //   cur=3:   next=null / 3.next=2 / prev=3 / cur=null
    //            null <- 1 <- 2 <- 3    null
    //                              ^    ^
    //                              prev cur(=null) -> 루프 끝
    //
    //   return prev;     // 3 -> 2 -> 1 = 뒤집힘
    //
    // 왜 베스트인가: 스택 방식이 값 복사(공간 O(n))로 피해 갔던 포인터 재배선을 여기서는 정면으로 해서 공간이
    //   O(1)이 됩니다. 그리고 위 유도의 "다음을 붙잡고 화살표를 돌린다"는 포인터 감각이 연결 리스트의 핵심 근육이라,
    //   cycle 검출(141), 재배열(143) 등으로 그대로 전이됩니다. (재귀로도 풀리지만 호출 스택이 O(n)이라, 면접
    //   기본값은 이 in-place입니다.) 실측: 원본/dummy/in-place 셋 모두 예제 + 빈/1개/동값에서 답 일치.
    //
    // > 카드: 노드의 next를 덮어쓰기 전에, 다음 노드를 먼저 붙잡아라 — 연결 리스트 조작의 1번 규칙.

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        // ListNode.of(...)로 입력을 만들고, ListNode.toArray(...)로 풀어 값 배열로 비교합니다.
        Judge.check(ListNode.toArray(s.reverseList(ListNode.of(1, 2, 3, 4, 5))), new int[]{5, 4, 3, 2, 1});
        Judge.check(ListNode.toArray(s.reverseList(ListNode.of(1, 2))), new int[]{2, 1});
        Judge.check(ListNode.toArray(s.reverseList(ListNode.of())), new int[]{});
        // 반례를 여기에 추가하세요(연결 리스트는 경계가 함정입니다):
        //   ListNode.of(7)            기대 [7]        (노드 1개 — 가드/불변식 경계)
        //   ListNode.of(1, 1, 2)      기대 [2, 1, 1]  (동값 — 값이 같아도 노드는 별개)
    }
}
