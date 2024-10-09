package primitive;

import java.util.Arrays;

public class Main {
    public static void main(String[] args) {
        Integer[] integers = new Integer[5];
        System.out.println(Arrays.toString(integers)); // [null, null, null, null, null]
        int[] ints = new int[5];
        System.out.println(Arrays.toString(ints)); // [0, 0, 0, 0, 0]

        int[] ints2 = {0, 0, 0, 0, 0};
        System.out.println(Arrays.toString(ints2)); // [0, 0, 0, 0, 0]

        int[] ints3 = new int[]{0, 0, 0, 0, 0};
        System.out.println(Arrays.toString(ints3)); // [0, 0, 0, 0, 0]

        int[][] d2ints = new int[][]{};
        System.out.println(Arrays.toString(d2ints)); // []
        // d2ints[0] = new int[3];  Index 0 out of bounds for length 0
        int[][] d2ints2 = new int[1][];
        System.out.println(Arrays.deepToString(d2ints2)); // [null]
        d2ints2[0] = new int[5];
        System.out.println(Arrays.deepToString(d2ints2)); // [[0, 0, 0, 0, 0]]

        // d2ints[0] = new int[3];  Index 0 out of bounds for length 0
        int[][] d2ints3 = new int[3][4];
        System.out.println(Arrays.deepToString(d2ints3)); // [[0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]]
        d2ints3[0][0] = 1;
        System.out.println(Arrays.deepToString(d2ints3)); // [[1, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]]

        String[][][] d3String = new String[1][1][1];
        System.out.println(Arrays.deepToString(d3String)); // [[[null]]]
        String[][][] d3String2 = new String[1][1][];
        System.out.println(Arrays.deepToString(d3String2)); // [[null]]
        d3String2[0][0] = new String[5];
        System.out.println(Arrays.deepToString(d3String2)); // [[[null, null, null, null, null]]]
    }
}