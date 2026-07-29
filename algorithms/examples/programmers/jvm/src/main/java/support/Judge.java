package support;

import java.util.Arrays;
import java.util.Objects;

/** Programmers Java/Kotlin 스캐폴드가 함께 쓰는 비교·출력 유틸입니다. */
public final class Judge {
    private Judge() {
    }

    public static void check(Object actual, Object expected) {
        if (!equal(expected, actual)) {
            throw new AssertionError("expected=" + stringify(expected) + ", actual=" + stringify(actual));
        }
        System.out.println("PASS " + stringify(actual));
    }

    private static boolean equal(Object expected, Object actual) {
        if (expected instanceof boolean[] expectedArray && actual instanceof boolean[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof byte[] expectedArray && actual instanceof byte[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof short[] expectedArray && actual instanceof short[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof char[] expectedArray && actual instanceof char[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof int[] expectedArray && actual instanceof int[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof long[] expectedArray && actual instanceof long[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof float[] expectedArray && actual instanceof float[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof double[] expectedArray && actual instanceof double[] actualArray) {
            return Arrays.equals(expectedArray, actualArray);
        }
        if (expected instanceof Object[] expectedArray && actual instanceof Object[] actualArray) {
            return Arrays.deepEquals(expectedArray, actualArray);
        }
        return Objects.equals(expected, actual);
    }

    private static String stringify(Object value) {
        if (value instanceof boolean[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof byte[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof short[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof char[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof int[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof long[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof float[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof double[] values) {
            return Arrays.toString(values);
        }
        if (value instanceof Object[] values) {
            return Arrays.deepToString(values);
        }
        return String.valueOf(value);
    }
}
