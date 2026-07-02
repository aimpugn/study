package dijkstra;

import java.util.Random;

/**
 * Solution.java 복잡도 주장(O((n+m) log m), HR 제한 안)의 실측 결제용 벤치입니다.
 * 최대 제약(n=3,000, m=4,498,500 완전 그래프, 가중치 1..350,000 무작위)에서 shortestReach를 1회 돌립니다.
 * 실행: javac -d /tmp/djbench src/main/java/support/Judge.java src/main/java/dijkstra/*.java
 *       java -cp /tmp/djbench dijkstra.Bench
 * (소스 런처로는 Result가 Solution.java 안의 부클래스라 Bench에서 직접 못 찾습니다 — 명시 컴파일 필요)
 */
class Bench {
    static void main() {
        int n = 3000;
        int m = n * (n - 1) / 2;
        int[][] edges = new int[m][3];
        Random rnd = new Random(42);
        int idx = 0;
        for (int u = 1; u <= n; u++) {
            for (int v = u + 1; v <= n; v++) {
                edges[idx][0] = u;
                edges[idx][1] = v;
                edges[idx][2] = 1 + rnd.nextInt(350000);
                idx++;
            }
        }
        long t0 = System.nanoTime();
        int[] result = Result.shortestReach(n, edges, 1);
        long t1 = System.nanoTime();
        System.out.println("n=" + n + " m=" + m + " elapsed=" + (t1 - t0) / 1_000_000 + "ms, dist[2]=" + result[0]);
    }
}
