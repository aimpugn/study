package mergetwosortedlists;

import support.Judge;
import support.ListNode;

/**
 * <a href="https://leetcode.com/problems/merge-two-sorted-lists/">Merge Two Sorted Lists</a>
 * <p>
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md에 작성합니다.
 */
class Solution {
    /**
     * 정렬된 두 리스트를 하나의 정렬된 리스트로 병합합니다.
     *
     * @param list1
     * @param list2
     * @return
     */
    public ListNode mergeTwoLists(ListNode list1, ListNode list2) {
        // 첨삭: 정답입니다. dummy 머리에 curr 커서를 두고 두 리스트를 훑어 작은 쪽부터 잇고, 한쪽이
        //   끝나면 남은 쪽을 쏟는 골격이 정확합니다. reverseList 첨삭의 "리스트를 만들 땐 dummy 머리 +
        //   커서 하나" 카드가 이 풀이로 잘 전이됐고, <= 비교 덕에 같은 값에서 list1이 먼저 이어지는
        //   안정성도 맞습니다. 그게 이 문제의 절반입니다.
        //
        //   다만 카드 하나가 함께 전이되지 않았습니다. reverseList 베스트가 값을 복사하지 않고 기존
        //   노드의 next를 제자리에서 다시 이어 공간을 O(1)로 줄였던 그 수입니다. 병합에서도 노드를 새로
        //   "만들" 필요가 없습니다. 두 리스트의 노드는 이미 존재하니, 작은 쪽 노드를 통째로 curr.next에
        //   이어 붙이면(splice) 됩니다. 지금 골격은 curr.next = new ListNode(list1.val)로 매 단계 새
        //   노드를 찍어, 입력 길이만큼(n+m) 추가 공간을 씁니다. 결과는 같지만 그게 아래 베스트와의 유일한
        //   차이입니다. 간극의 정체는 "새 리스트를 만든다"는 프레임이 입력 노드를 재사용한다는 선택지를
        //   가린 것이고, 스택으로 값을 옮겨 뒤집던 그 습관과 같은 뿌리입니다.
        //
        // > 카드: 리스트를 합치거나 재배열할 때 노드가 이미 있으면, 새로 만들지 말고 next만 다시 이어라 (복사 X, splice O).
        var merged = new ListNode();
        var curr = merged;
        while (list1 != null && list2 != null) {
            // 첨삭: 이 디버그 출력은 통과 뒤 지웁니다. 채점/제출 환경에서 불필요한 IO이고, 회귀 확인은
            //   아래 main의 Judge.check가 맡습니다. 메서드 끝에 주석 처리된 toArray() 한 줄도 인자가
            //   없어 되살리면 컴파일이 깨지니, 함께 지우는 편이 깔끔합니다.
            System.out.println("list1.val: " + list1.val + " list2.val: " + list2.val);
            if (list1.val <= list2.val) {
                curr.next = new ListNode(list1.val);
                curr = curr.next;
                list1 = list1.next;
            } else {
                curr.next = new ListNode(list2.val);
                curr = curr.next;
                list2 = list2.next;
            }
        }

        // 첨삭: 이 두 drain 루프는 정확합니다. 다만 splice로 바꾸면 한 줄로 줄어듭니다. 메인 루프를
        //   빠져나온 순간 list1, list2 중 최대 하나만 남고, 그 남은 쪽은 이미 정렬된 채 이어져 있습니다.
        //   그러니 노드를 하나씩 복사할 게 아니라 남은 머리를 통째로 curr.next에 붙이면 끝입니다 (아래
        //   베스트의 마지막 줄).
        while (list1 != null) {
            curr.next = new ListNode(list1.val);
            list1 = list1.next;
            curr = curr.next;
        }
        while (list2 != null) {
            curr.next = new ListNode(list2.val);
            list2 = list2.next;
            curr = curr.next;
        }

        // System.out.println(Arrays.toString(ListNode.toArray()));

        return merged.next;
    }

    // 베스트 프랙티스: 새 노드를 0개 만들고 기존 노드의 next만 제자리에서 다시 이어, 추가 공간이 O(1)
    //   입니다. dummy 머리 + curr 커서는 그대로입니다. 바뀌는 건 단 하나, curr.next = new ListNode(x) 대신
    //   curr.next = (고른 노드 자체)로 잇는 것입니다.
    //
    //   어떻게 도달하나 (외우지 말고 이 사고를 복제하세요):
    //     씨앗 한 줄: "합칠 노드는 이미 다 있다. 새로 만들지 말고 작은 쪽을 골라 꼬리에 잇자."
    //     - 합친 리스트의 꼬리를 가리키는 손이 필요하다              -> curr (dummy에서 시작, 진짜 머리는 dummy.next).
    //     - 매 단계 두 머리 중 작은 쪽을 꼬리에 잇고 그쪽만 전진       -> curr.next = 작은 노드; 그 리스트 = 그.next.
    //     - 한쪽이 비면 남은 쪽은 통째로 잇는다 (이미 정렬+연결됨)     -> curr.next = 남은 머리.
    //   초기값(dummy), 잇는 규칙, 종료(한쪽 빔)가 외운 게 아니라 이 씨앗에서 따라 나옵니다.
    //
    // public ListNode mergeTwoLists(ListNode list1, ListNode list2) {
    //     ListNode dummy = new ListNode();              // 가짜 머리 - 진짜 머리는 dummy.next
    //     ListNode curr = dummy;
    //     while (list1 != null && list2 != null) {
    //         if (list1.val <= list2.val) {             // <= 라야 같은 값에서 list1 먼저 (안정성)
    //             curr.next = list1;                    // 새 노드 X, list1 노드 자체를 잇는다
    //             list1 = list1.next;
    //         } else {
    //             curr.next = list2;
    //             list2 = list2.next;
    //         }
    //         curr = curr.next;
    //     }
    //     curr.next = (list1 != null) ? list1 : list2;  // 남은 쪽 통째로 (drain 두 루프가 이 한 줄로)
    //     return dummy.next;
    // }
    //
    //   매 단계 두 리스트의 맨 앞을 비교해 작은 쪽(같으면 list1)을 결과 끝에 붙입니다.
    //   (list1 = 1, 2, 4 / list2 = 1, 3, 4)
    //
    //     step | l1앞  l2앞 | 고른 것    | 결과
    //     -----+------------+-----------+------------------
    //     1    |  1     1   | list1의 1  | 1
    //     2    |  2     1   | list2의 1  | 1 1
    //     3    |  2     3   | list1의 2  | 1 1 2
    //     4    |  4     3   | list2의 3  | 1 1 2 3
    //     5    |  4     4   | list1의 4  | 1 1 2 3 4
    //     6    |  -     4   | list2의 4  | 1 1 2 3 4 4   (l1이 비어 list2의 나머지를 통째로)
    //
    //   "고른 것"은 값을 복사한 새 노드가 아니라 그 노드 자체입니다. curr.next = list1 한 줄이
    //   list1의 현재 노드를 결과 끝에 그대로 답니다. 6단계처럼 한쪽이 비면 남은 리스트를 통째로 잇습니다.
    //   반환은 dummy.next = 1 1 2 3 4 4.
    //
    //   왜 더 나은가: 시간은 두 버전 다 O(n+m)입니다 (각 노드를 한 번씩 본다, n과 m은 두 리스트 길이).
    //     갈리는 건 공간입니다. 복사 버전은 결과 길이만큼 새 노드를 만들어 O(n+m), splice는 기존 노드를
    //     재사용해 O(1)입니다(반환 리스트 제외, 새 노드 0개). reverseList 베스트가 값 복사 대신 포인터
    //     재배선으로 O(1)을 만든 것과 똑같은 수입니다. 실측: 공식 3개 + 빈쪽/겹침없음/역관계/동값/단일교차
    //     5개에서 답 일치.
    //     트레이드오프: splice는 입력 리스트의 next를 바꿔 원본을 소비합니다. 호출자가 원본을 그대로
    //     유지해야 하면 복사 버전이 맞습니다. "더 낫다"는 공간 제약과 원본 보존 요구에 달려 있습니다.
    //
    // > 불변식: curr는 늘 합쳐진 부분의 꼬리다. curr.next부터는 아직 안 본 두 리스트의 나머지만 남는다.
    // > 카드: 정렬된 두 리스트 병합 = dummy 꼬리에 작은 머리를 splice, 한쪽 비면 나머지 통째로.

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        // ListNode.of(...)로 입력을 만들고, ListNode.toArray(...)로 풀어 값 배열로 비교합니다.
        Judge.check(ListNode.toArray(s.mergeTwoLists(ListNode.of(1, 2, 4), ListNode.of(1, 3, 4))), new int[]{1, 1, 2, 3, 4, 4});
        Judge.check(ListNode.toArray(s.mergeTwoLists(ListNode.of(), ListNode.of())), new int[]{});
        Judge.check(ListNode.toArray(s.mergeTwoLists(ListNode.of(), ListNode.of(0))), new int[]{0});
        // 반례를 여기에 추가하세요:
    }
}
