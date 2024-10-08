package main.java.p42839;

import java.util.*;

public class Main {

    public static void main(String[] args) {
        Main main = new Main();
        System.out.println(main.solution("17"));
        System.out.println(main.solution("011"));
        System.out.println(main.solution("111"));
        System.out.println(main.isPrime(1213));
        System.out.println(main.isPrime(1231));
    }

    /**
     * <pre>
     * 주어진 숫자들로 만들 수 있는 소수의 개수를 계산합니다.
     * 1. 입력 문자열을 개별 숫자로 분리
     * 2. DFS를 사용하여 모든 가능한 숫자 조합 생성
     * 3. 생성된 각 숫자에 대해 소수 여부 확인
     * 4. 중복을 제거한 소수의 개수 반환
     * </pre>
     *
     * @param numbers 숫자들이 붙어있는 문자열 (예: "17", "011")
     * @return 만들 수 있는 소수의 개수
     */
    public int solution(String numbers) {
        Set<Integer> primes = new HashSet<>();
        boolean[] used = new boolean[numbers.length()];
        dfs("", numbers, used, primes);
        return primes.size();
    }

    /**
     * <pre>
     * DFS를 사용하여 모든 가능한 숫자 조합을 생성합니다.
     *
     * 재귀적으로 호출되는 각 단계에서:
     * 1. 현재까지 만들어진 숫자가 소수인지 확인
     * 2. 사용되지 않은 각 숫자를 현재 조합에 추가
     * 3. 다음 단계로 재귀 호출
     * 4. 백트래킹을 위해 사용 표시 제거
     *
     * </pre>
     *
     *
     * @param current 현재까지 만들어진 숫자 문자열
     * @param numbers 원본 숫자 문자열
     * @param used 각 자리 숫자의 사용 여부를 나타내는 배열
     * @param primes 발견된 소수들을 저장하는 Set
     */
    private void dfs(String current, String numbers, boolean[] used,
            Set<Integer> primes) {
        // 현재 만들어진 숫자가 비어있지 않으면 소수인지 확인
        if (!current.isEmpty()) {
            int num = Integer.parseInt(current);
            if (!primes.contains(num) && isPrime(num)) {
                primes.add(num);
            }
        }

        if (current.length() == numbers.length()) {
            return;
        }

        // 사용되지 않은 각 숫자에 대해 재귀 호출
        for (int i = 0; i < numbers.length(); i++) {
            if (!used[i]) {
                used[i] = true; // 현재 숫자 사용 표시
                // System.out.println("before used:\t" + Arrays.toString(used)
                //         + "\t primes: " + primes.toString());
                dfs(current + numbers.charAt(i), numbers, used, primes);
                // System.out.println("after used:\t" + Arrays.toString(used)
                //         + "\t primes: " + primes.toString());
                used[i] = false; // 백트래킹: 사용 표시 제거
            }
        }
    }

    /**
     * 소수(prime number): 1보다 큰 자연수 중 1과 자기 자신만을 약수로 가지는 수 1 또는 자기자신 외에 다른 수로 나누지
     * 않는 소수인지 여부 판단합니다.
     *
     * @param number
     * @return
     *
     * @see <a href=
     *      "https://ko.wikipedia.org/wiki/%EC%86%8C%EC%88%98_(%EC%88%98%EB%A1%A0)">소수</a>
     */
    private boolean isPrime(int number) {
        if (number < 2) {
            return false;
        }

        for (int i = 2; i <= Math.sqrt(number); i++) {
            if (number % i == 0) {
                return false;
            }
        }
        return true;
    }
}
