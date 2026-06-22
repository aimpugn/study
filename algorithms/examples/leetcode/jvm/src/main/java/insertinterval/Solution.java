package insertinterval;

import support.Judge;

import java.util.ArrayList;

/**
 * <a href="https://leetcode.com/problems/insert-interval/">Insert Interval</a>
 * <p>
 * 회고·복습 카드는 풀이 완료 후 같은 폴더 PROCESS.md에 작성합니다.
 */
class Solution {
    /**
     * newInterval 배열을 intervals 배열에 삽입합니다.
     * - intervals 배열은 여전히 start_i 기준으로 오름차순 정렬되어 있어야 합니다.
     * - 여전히 서로 겹치지 않아야 합니다.
     * - 제자리(in-place)로 intervals 배열을 수정할 필요 없습니다.
     *
     * @param intervals 서로 겹치지 않는 인터벌 배열
     * - intervals[i] = [start_i, end_i]
     * - 인터벌 배열은 start_i 기준으로 오름차순 정렬
     * @param newInterval 새로운 인터벌
     * - newInterval = [start, end]
     * @return newInterval 이 삽입된 새로운 인터벌 배열
     */
    /* 잘못된 풀이 방식
    public int[][] insert(int[][] intervals, int[] newInterval) {
        // 1. intervals = [[1,3],[6,9]], newInterval = [2,5]
        //    => [[1,5],[6,9]]
        //    1 < 2 < 3 < 5
        //
        // 2. intervals = [[1,2],[3,5],[6,7],[8,10],[12,16]], newInterval = [4,8]
        //    => 3 < 4 < 5 < 6 < 7 < 8 < 10
        //    => [[1,2],[3,10],[12,16]]
        //
        // "서로 겹치지 않는다"는 조건이 있으므로 intervals 배열의 start_i 는 start_{i-1}, end_{i_1}보다 반드시 큽니다.
        //
        // 병합된 배열을 모을 정답지가 필요하고,
        // 인터벌을 순회하면서 병합 대상인지 봐야 하고,
        // 새로운 인터벌은 끼워 넣어야 합니다.
        // 그러면 기존 인터벌과 새로운 인터벌을 비교하되, 합쳐지는 곳은 정답지여야 합니다.
        // 병합 대상이 아니면 answer 리스트의 새로운 요소가 되어야 합니다.
        //
        // 첨삭: 분해는 거의 정답입니다. 위에 적은 "새로운 인터벌은 끼워 넣어야 합니다"가 이 문제의 핵심이에요.
        //   그런데 아래 구현에는 newInterval을 answer에 넣는 줄이 없습니다. newInterval을 기존 칸의 끝만
        //   늘려 주는 보조 값으로 다뤘기 때문입니다. 정답 사고에서 newInterval은 스스로 자리를 잡는 주인공입니다.
        //
        //   이 차이가 두 가지 간극을 만듭니다.
        //     (a) newInterval이 answer에 자기 자신으로 들어가는 길이 없다.
        //     (b) 끝(end)만 늘리고 시작(start)은 줄이지 않는다.
        //   둘은 한 뿌리에서 나옵니다. answer를 intervals[0]로 미리 채우고 i=1부터 도는 골격이
        //   newInterval을 종속 인자로 묶었기 때문입니다.
        //
        //   다음에는 같은 신호를 잡으세요. "정렬된 구간에 새 구간을 끼워 병합"이 보이면, 새 구간을
        //   주인공으로 두고 세 구간으로 가릅니다. 고친 풀이 둘이 이 메서드 아래에 있습니다 —
        //   하나는 이 풀이의 골격을 살린 버전, 하나는 베스트 프랙티스입니다.
        //
        // > 카드: "구간 끼워 병합" 신호 -> 새 구간을 손에 들고 [왼쪽 통과 / 겹침 흡수해 양끝 키우기 / 키운 새 구간 넣고 오른쪽 통과] 3구간으로.
        if (intervals.length == 0) return new int[][]{newInterval};

        var answer = new ArrayList<int[]>();
        answer.add(intervals[0]);
        // 첨삭: 여기가 구조적 뿌리입니다. 병합 로직이 전부 루프 안에 있어, 길이 1 배열은 루프가 안 돌아 병합이 일어나지 않습니다.
        //   그래서 예제2 insert([[1,5]], [2,7])가 [[1,7]]이 아니라 [[1,5]]로 나옵니다(실측). 게다가 add(intervals[0])은
        //   배열 참조를 넣어, 아래 curr[1]=... 한 줄이 입력 intervals까지 바꿉니다(입력 변형).

        var newStart = newInterval[0];
        var newEnd = newInterval[1];

        // 인터벌 배열을 돌면서 새로운 인터벌이 속해야 하는 곳을 찾습니다.
        for (var i = 1; i < intervals.length; i++) {
            var curr = answer.get(answer.size() - 1);
            var currStart = curr[0];
            var currEnd = curr[1];
            // 겹치는 경우를 찾으므로, 현재 끝이 새로운 시작보다 큰 경우
            // 첨삭: 디버그 출력입니다. 제출 전에 지웁니다.
            System.out.println("currEnd: " + currEnd + ", newStart: " + newStart + ", newEnd: " + newEnd + ", intervals[i][0]: " + intervals[i][0] + ", intervals[i][1]: " + intervals[i][1]);
            // 첨삭: 끝만 늘리고 시작은 그대로 두는 자리입니다. curr[1]만 바꾸고 curr[0]은 건드리지 않습니다.
            //   그래서 newInterval이 왼쪽에서 겹치면 그 시작이 사라집니다. insert([[3,5],[8,9]], [1,4])가
            //   [[1,5],[8,9]] 대신 [[3,5],[8,9]]로 나옵니다(실측). 바로 위에서 뽑은 currStart를 한 번도
            //   안 쓰는 것이 빠진 연산(start=min)의 흔적입니다.
            if (newStart <= currEnd) {
                curr[1] = Math.max(newEnd, currEnd);
            }
            // 그리고 현재 인터벌 시작이 새로운 끝과 같은 경우?
            // 큰 경우는 겹치지 않으므로 무시
            // 첨삭: 이 특수 케이스는 깨진 모델을 떠받치는 임시 패치입니다. newEnd가 다음 칸 시작과 딱 맞는 경계 접촉만 따로 처리해요.
            //   덕분에 경계 접촉 입력 insert([[1,3],[6,9]], [3,6])는 원본도 [[1,9]]로 맞힙니다 — 패치가 노린 케이스는
            //   잡지만 일반 겹침은 못 잡는다는 증거입니다. 올바른 흡수 조건 "칸의 시작 <= newEnd"가 "<="로 경계 접촉까지
            //   품으니, 제대로 세우면 이 if는 필요 없습니다.
            if (newEnd == intervals[i][0]) {
                curr[1] = Math.max(newEnd, intervals[i][1]);
            }
            // 현재 끝보다 새로운 시작이 크거나
            // 새로운 끝이 현재 인터벌 시작보다 작으면 새로운 배열을 넣습니다.
            // 첨삭: newInterval이 answer에 자기 자신으로 들어갈 수 있는 유일한 자리입니다. 그런데 넣는 것은 늘 intervals[i]뿐이에요.
            //   그래서 겹치지 않는 빈틈 입력에서 newInterval이 사라집니다. insert([[1,2],[5,6]], [3,4])가
            //   [[1,2],[3,4],[5,6]] 대신 [[1,2],[5,6]]로 나옵니다(실측).
            if (currEnd < newStart || newEnd < intervals[i][0]) {
                answer.add(intervals[i]);
            }
        }

        // System.out.println(Arrays.deepToString(answer.toArray(new int[0][])));

        return answer.toArray(new int[0][]);
    }*/

    // 학습자 접근을 고친 동작 버전입니다. 원본 골격(answer에 모으고, 마지막 칸과 겹치면 끝을 늘리고, 안 겹치면
    //   새 요소로 추가)을 그대로 살렸습니다. 딱 하나 바꿨어요 — newInterval을 순회 밖에서 따로 만지는 대신,
    //   리스트에 넣고 start로 정렬한 뒤 똑같이 훑습니다. 그러면 newInterval도 그냥 한 칸이 되어 골격이 맞아 돌아갑니다.
    //
    public int[][] insert2(int[][] intervals, int[] newInterval) {
        var all = new ArrayList<int[]>(java.util.Arrays.asList(intervals));
        all.add(newInterval); // 합칠 대상으로 newInterval 을 넣어야 합니다.
        all.sort((x, y) -> Integer.compare(x[0], y[0]));

        // System.out.println("all: " + all.stream().map(Arrays::toString).toList());

        var merged = new ArrayList<int[]>();
        for (var interval : all) {
            var mergedLast = merged.isEmpty() ? null : merged.getLast();
            if (mergedLast == null || mergedLast[1] < interval[0]) { // '병합 목록의 가장 마지막 요소의 끝'보다 '현재 인터벌 시작'이 크다면?
                // mergedLast: [1, 2], interval: [3, 4]
                // - mergedLast[1]: 2
                // - interval[0]: 3
                // 겹치지 않으므로 바로 추가
                merged.add(new int[]{interval[0], interval[1]});  // 안 겹침 -> 새 요소
            } else {
                // mergedLast: [1, 3], interval: [3, 4]
                // - mergedLast[1]: 3
                // - interval[0]: 3
                // - interval[1]: 4
                mergedLast[1] = Math.max(mergedLast[1], interval[1]);   // 겹침 -> 끝을 늘림
                // mergedLast: [1, 4]
            }
        }

        return merged.toArray(new int[0][]);
    }
    //
    // 무엇이 부족했고 무엇을 잘못 생각했나 (고친 코드와 비교):
    //   1) newInterval을 순회 대상에 넣었습니다. 원본은 newInterval을 옆에 두고 기존 칸의 끝만 늘렸기에
    //      빈틈, 전부-왼쪽, 전부-오른쪽에서 newInterval이 통째로 사라졌습니다. 이제 한 칸으로 합류해 안 사라집니다.
    //   2) start로 정렬했습니다. 정렬하면 시작이 작은 칸이 먼저 들어가므로 "시작을 줄인다"는 연산 자체가 필요 없습니다.
    //      원본의 시작 증발 버그는 정렬 없이 작은 시작의 newInterval을 기존 칸에 옆치기로 합쳐서 생긴 것이었어요.
    //   3) 길이 0/1 특수 케이스가 사라졌습니다. answer.isEmpty() 검사 하나로 빈 입력까지 골격이 자연히 처리합니다.
    //   잘못 생각한 핵심: newInterval을 "기존 칸을 수정하는 도구"로 본 것. 사실은 "정렬에 합류하는 한 칸"입니다.
    //   위 8개 입력(예제 3 + 실패 5)을 실제로 돌려 전부 통과 확인했습니다.
    //   복잡도: 정렬 때문에 시간 O(n log n), 공간 O(n). 이게 곧 Merge Intervals(LC56)의 풀이입니다.

    // 베스트 프랙티스입니다. 이 문제는 입력이 이미 start 오름차순 + 서로 안 겹침으로 주어집니다. 그 전제를 쓰면
    //   정렬을 건너뛰고 한 번만 훑어 O(n)에 끝낼 수 있어요. newInterval을 손에 들고 세 구간으로 가릅니다.
    //
    // > 불변식: answer에는 newInterval보다 확실히 왼쪽인 칸만 들어간다. newInterval은 겹치는 동안 손에 들고
    // >   start=min, end=max로 키우다가, 안 겹치는 첫 칸을 만나면 한 번만 넣고 나머지를 흘려보낸다.
    //
    public int[][] insert3(int[][] intervals, int[] newInterval) {
        var merged = new ArrayList<int[]>();
        int idx = 0;
        int intervalsLength = intervals.length;
        int newStart = newInterval[0], newEnd = newInterval[1];

        // 1) newInterval보다 확실히 왼쪽인 칸: 그대로 통과
        //    [1, 2]와 [5, 7] 경우 interval[1] = 2 < 5 이므로,
        //    [1, 2]를 병합되는 리스트에 추가하고 인덱스를 증가시킵니다.
        //    다음 인터벌이 [3, 6]인 경우 6 < 5 는 거짓이므로 while 루프를 벗어납니다.
        while (idx < intervalsLength && intervals[idx][1] < newStart) {
            merged.add(intervals[idx]);
            idx++;
        }

        // 2) 여기 왔는데 `idx < intervalsLength`라는 것은 최소 newStart <= interval[1] 에 대항하여 겹친다는 의미입니다.
        //    겹치는 칸: newInterval에 흡수하며 양끝을 키웁니다. (시작은 min, 끝은 max)
        //    [3, 6]와 [5, 7] 경우 interval[0] = 3 <= 7
        //    - min(newStart: 5, interval[0]: 3) = 3 = newStart
        //    - max(newStart: 7, interval[1]: 6) = 7 = newEnd
        //
        //    [6, 9]와 [5, 7] 경우 interval[0] = 6 <= 7
        //    - min(newStart: 3, interval[0]: 6) = 3 = newStart
        //    - max(newStart: 7, interval[1]: 9) = 9 = newEnd
        while (idx < intervalsLength && intervals[idx][0] <= newEnd) {
            newStart = Math.min(newStart, intervals[idx][0]);
            newEnd = Math.max(newEnd, intervals[idx][1]);
            idx++;
        }
        merged.add(new int[]{newStart, newEnd});  // 다 키운 newInterval을 한 번 넣는다

        // 3) 완전히 오른쪽인 나머지: 그대로 통과
        while (idx < intervalsLength) {
            merged.add(intervals[idx]);
            idx++;
        }

        return merged.toArray(new int[0][]);
    }
    //
    // 고친 학습자 버전보다 왜 더 나은가:
    //   입력이 이미 정렬돼 있으니 다시 정렬할 필요가 없습니다. 정렬을 빼면 O(n log n) -> O(n)이 됩니다.
    //   "이미 정렬+비겹침"이라는 전제를 쓴 대가로 일반성은 줄지만(정렬 안 된 입력엔 못 씀), 이 문제에선 가장 빠릅니다.
    //   이것도 위 8개 입력을 실제로 돌려 전부 통과 확인했습니다.

    static void main() {
        var s = new Solution();
        // 풀기 전에는 빨간 상태(AssertionError)가 정상입니다. 풀면 초록이 됩니다.
        Judge.check(s.insert2(new int[][]{{1, 3}, {6, 9}}, new int[]{2, 5}), new int[][]{{1, 5}, {6, 9}});
        Judge.check(s.insert2(new int[][]{{1, 5}}, new int[]{2, 7}), new int[][]{{1, 7}});
        Judge.check(s.insert2(new int[][]{{1, 2}, {3, 5}, {6, 7}, {8, 10}, {12, 16}}, new int[]{4, 8}), new int[][]{{1, 2}, {3, 10}, {12, 16}});
        Judge.check(s.insert2(new int[][]{{1, 2}, {3, 6}, {6, 9}}, new int[]{5, 7}), new int[][]{{1, 2}, {3, 9}});

        Judge.check(s.insert3(new int[][]{{1, 2}, {3, 6}, {6, 9}}, new int[]{5, 7}), new int[][]{{1, 2}, {3, 9}});


        // 반례를 여기에 추가하세요. 아래는 실패 모드별로 분류한 입력입니다(원본 실측 -> 기대). 직접 Judge.check로 넣어 빨갛게 만든 뒤 고치세요.
        //   [길이 1, 루프 안 돎]    insert([[1,5]], [2,7])                     원본 [[1,5]]        -> 기대 [[1,7]]
        //   [빈틈, new 증발]        insert([[1,2],[5,6]], [3,4])               원본 [[1,2],[5,6]]  -> 기대 [[1,2],[3,4],[5,6]]
        //   [왼쪽 겹침, 시작 증발]   insert([[3,5],[8,9]], [1,4])               원본 [[3,5],[8,9]]  -> 기대 [[1,5],[8,9]]
        //   [전부보다 왼쪽]         insert([[3,5]], [1,2])                     원본 [[3,5]]        -> 기대 [[1,2],[3,5]]
        //   [전부보다 오른쪽]       insert([[1,3]], [6,9])                     원본 [[1,3]]        -> 기대 [[1,3],[6,9]]
        //   [여러 칸 삼킴]          insert([[1,2],[3,5],[6,7],[8,10]], [4,9])  원본 [[1,2],[3,9]]  -> 기대 [[1,2],[3,10]]
        //   엣지(원본도 통과): 경계 접촉 insert([[1,3],[6,9]], [3,6]) = [[1,9]] / 빈 입력 insert([], [2,5]) = [[2,5]]
    }
}
