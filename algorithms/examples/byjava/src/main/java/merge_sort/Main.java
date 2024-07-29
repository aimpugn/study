package main.java.merge_sort;

import java.util.*;

public class Main {
    public static void main(String[] args) {
        int[] arr = {64, 34, 25, 12, 22, 11, 90};

        System.out.println("Original array: " + Arrays.toString(arr));

        MergeSort.mergeSort(arr, 0, arr.length - 1);

        System.out.println("Sorted array: " + Arrays.toString(arr));
    }
}


class MergeSort {
    /**
     * 주어진 배열을 병합 정렬로 정렬합니다.
     *
     * @param arr 정렬할 배열
     * @param left 정렬할 부분의 시작 인덱스
     * @param right 정렬할 부분의 끝 인덱스
     */
    public static void mergeSort(int[] arr, int left, int right) {
        if (left < right) {
            // 중간 지점을 찾습니다
            int mid = left + (right - left) / 2;

            // 왼쪽 부분 배열을 정렬합니다
            mergeSort(arr, left, mid);
            // 오른쪽 부분 배열을 정렬합니다
            mergeSort(arr, mid + 1, right);

            // 정렬된 두 부분을 병합합니다
            merge(arr, left, mid, right);
        }
    }

    /**
     * 두 개의 정렬된 부분 배열을 병합합니다.
     *
     * @param arr 원본 배열
     * @param left 왼쪽 부분 배열의 시작 인덱스
     * @param mid 왼쪽 부분 배열의 끝 인덱스
     * @param right 오른쪽 부분 배열의 끝 인덱스
     */
    public static void merge(int[] arr, int left, int mid, int right) {
        // 두 부분 배열의 크기를 계산합니다
        int n1 = mid - left + 1;
        int n2 = right - mid;

        // 임시 배열을 생성합니다
        int[] L = new int[n1];
        int[] R = new int[n2];

        // 데이터를 임시 배열에 복사합니다
        for (int i = 0; i < n1; ++i)
            L[i] = arr[left + i];
        for (int j = 0; j < n2; ++j)
            R[j] = arr[mid + 1 + j];

        // 두 임시 배열을 병합합니다
        int i = 0, j = 0;
        int k = left;
        while (i < n1 && j < n2) {
            if (L[i] <= R[j]) {
                arr[k] = L[i];
                i++;
            } else {
                arr[k] = R[j];
                j++;
            }
            k++;
        }

        // L[]의 남은 요소를 복사합니다
        while (i < n1) {
            arr[k] = L[i];
            i++;
            k++;
        }

        // R[]의 남은 요소를 복사합니다
        while (j < n2) {
            arr[k] = R[j];
            j++;
            k++;
        }
    }
}
