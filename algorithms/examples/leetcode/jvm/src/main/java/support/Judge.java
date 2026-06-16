package support;

import java.util.Arrays;
import java.util.Objects;

/**
 * LeetCode 풀이의 공용 검증 유틸입니다. 각 Solution.java가 비교/출력 헬퍼를 복제하지 않고 이 한 곳을 재사용합니다.
 * JDK 소스 런처(java src/main/java/&lt;slug&gt;/Solution.java)가 import 시 이 파일을 자동으로 함께 컴파일합니다.
 */
public final class Judge {
    private Judge() {
    }

    /** 실제값과 기대값을 비교해, 다르면 AssertionError를 던지고, 같으면 PASS를 출력합니다. */
    public static void check(Object actual, Object expected) {
        if (!equal(expected, actual)) {
            throw new AssertionError("expected=" + stringify(expected) + ", actual=" + stringify(actual));
        }
        System.out.println("PASS " + stringify(actual));
    }

    private static boolean equal(Object expected, Object actual) {
        if (expected instanceof int[] e && actual instanceof int[] a) {
            return Arrays.equals(e, a);
        }
        if (expected instanceof long[] e && actual instanceof long[] a) {
            return Arrays.equals(e, a);
        }
        if (expected instanceof Object[] e && actual instanceof Object[] a) {
            return Arrays.deepEquals(e, a);
        }
        return Objects.equals(expected, actual);
    }

    private static String stringify(Object value) {
        if (value instanceof int[] v) {
            return Arrays.toString(v);
        }
        if (value instanceof long[] v) {
            return Arrays.toString(v);
        }
        if (value instanceof Object[] v) {
            return Arrays.deepToString(v);
        }
        return String.valueOf(value);
    }
}
