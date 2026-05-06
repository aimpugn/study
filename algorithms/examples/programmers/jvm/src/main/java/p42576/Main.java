package p42576;

import java.util.HashMap;
import java.util.LinkedList;
import java.util.Objects;

/**
 * <a href="https://school.programmers.co.kr/learn/courses/30/lessons/42576">완주하지 못한 선수</a>
 */
class Solution {
    /**
     * 단 한 명의 선수 제외하고 모든 선수가 마라톤을 완주.
     *
     * @param participant 마라톤 참여자 이름 배열
     * - 1 <= participant.length <= 100,000
     * - 1 <= participant[i].length <= 20. 알파벳.
     * - 동명이인 가능
     * @param completion 완주한 선수 이름 배열
     * - completion.lengt = participant.length - 1
     *
     * @return 완주하지 못한 선수 이름
     */
    public String solution(String[] participant, String[] completion) {
        // "동명이인이 중요한가?"라고 생각해 보면, 동명이인이면 여러 번 체크할 수 있어야 해서 중요합니다.
        // 그러면 참가자 한 명씩 완주자에서 체크하고, 존재하면 pass, 체크된 완주자는 제거.
        var completed = new HashMap<String, Runner>(participant.length);
        for (var name : completion) {
            completed.compute(name, (key, value) -> {
                if (value == null) {
                    return new Runner(key);
                }

                value.names.add(key);

                return value;
            });
        }

        for (var name : participant) {
            if (!completed.containsKey(name)) {
                return name;
            }

            var runner = completed.get(name);
            if (runner.names.isEmpty()) {
                return name;
            }

            runner.names.removeFirst();
        }

        return "";
    }

    public String solutionByCodex(String[] participant, String[] completion) {
        // 완주자 배열은 "완주권이 남아 있는 이름 목록"으로 볼 수 있습니다.
        // 동명이인이 있으므로 이름 존재 여부가 아니라 이름별 남은 개수를 세어야 합니다.
        var completedCounts = new HashMap<String, Integer>(completion.length);

        // 먼저 완주한 사람 이름을 하나씩 읽습니다.
        for (var name : completion) {
            // 이미 같은 이름이 있으면 기존 개수에 1을 더하고, 처음 보는 이름이면 0에서 시작해 1로 만듭니다.
            completedCounts.put(name, completedCounts.getOrDefault(name, 0) + 1);
        }

        // 이제 참가자 명단을 앞에서부터 보면서, 각 참가자가 완주권 하나를 소비할 수 있는지 확인합니다.
        for (var name : participant) {
            // 이 이름으로 남아 있는 완주 기록 개수를 꺼냅니다. 없으면 0으로 보고 바로 미완주 후보가 됩니다.
            int remainingCompletionCount = completedCounts.getOrDefault(name, 0);

            // 완주 기록이 0개라면, 이 참가자의 이름을 처리해 줄 완주자가 더 이상 없다는 뜻입니다.
            if (remainingCompletionCount == 0) {
                return name;
            }

            // 완주 기록이 있었다면 참가자 한 명과 매칭했으므로 남은 개수를 하나 줄입니다.
            completedCounts.put(name, remainingCompletionCount - 1);
        }

        // 문제 조건상 participant가 completion보다 정확히 한 명 많으므로 정상 입력에서는 여기까지 올 수 없습니다.
        throw new IllegalStateException("완주하지 못한 선수를 찾지 못했습니다.");
    }

    static class Runner {
        LinkedList<String> names;

        public Runner(String name) {
            this.names = new LinkedList<>();
            this.names.add(name);
        }
    }
}

public class Main {
    public static void main(String[] args) {
        Solution solution = new Solution();

        check(
            solution,
            new String[]{"leo", "kiki", "eden"},
            new String[]{"eden", "kiki"},
            "leo"
        );
        check(
            solution,
            new String[]{"marina", "josipa", "nikola", "vinko", "filipa"},
            new String[]{"josipa", "filipa", "marina", "nikola"},
            "vinko"
        );
        check(
            solution,
            new String[]{"mislav", "stanko", "mislav", "ana"},
            new String[]{"stanko", "ana", "mislav"},
            "mislav"
        );
    }

    private static void check(Solution solution, String[] participant, String[] completion, String expected) {
        String actual = solution.solution(participant, completion);
        String actualByCodex = solution.solutionByCodex(participant, completion);
        System.out.printf(
            "expected=%s actual=%s pass=%s codex=%s codexPass=%s%n",
            expected,
            actual,
            Objects.equals(expected, actual),
            actualByCodex,
            Objects.equals(expected, actualByCodex)
        );
    }
}
