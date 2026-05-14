package p42748.attempt211212;

import java.util.Arrays;

public class Main {
}

class Solution {
    public int[] solution(int[] array, int[][] commands) {
        int arraySize = commands.length;
        int[] answer = new int[arraySize];
        int idx = 0;
        for (int[] command : commands) {
            int from = command[0] - 1;
            int to = command[1];
            int targetIdx = command[2];
//            Print.byFormat("from: {0}, to:{1} targetIdx:{2}", from, to, targetIdx);
            // 원소가 한 개인 경우
            if ((to - from) == 1) {
                answer[idx] = array[from];
                idx++;
                continue;
            }

            int[] partialArray = Arrays.copyOfRange(array, from, to);
            Arrays.sort(partialArray);
//            Print.asArray(partialArray);
            answer[idx] = partialArray[targetIdx - 1];
            idx++;
        }

        return answer;
    }
}
