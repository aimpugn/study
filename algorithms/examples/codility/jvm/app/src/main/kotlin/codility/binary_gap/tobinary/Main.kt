package codility.binary_gap.tobinary

fun Int.toBinaryArray32(): BooleanArray {
    // 32비트 크기의 BooleanArray 생성
    val bits = BooleanArray(32)

    for (i in 31 downTo 0) { // MSB에서 LSB로 진행
        // mask 값
        // i=31, shl  0 [(0000 0000) (0000 0000) (0000 0000) (0000 0001)]
        // i=30, shl  1 [(0000 0000) (0000 0000) (0000 0000) (0000 0010)]
        // i=29, shl  2 [(0000 0000) (0000 0000) (0000 0000) (0000 0100)]
        //                       .......
        // i= 3, shl 28 [(0001 0000) (0000 0000) (0000 0000) (0000 0000)]
        // i= 2, shl 29 [(0010 0000) (0000 0000) (0000 0000) (0000 0000)]
        // i= 1, shl 30 [(0100 0000) (0000 0000) (0000 0000) (0000 0000)]
        // i= 0, shl 31 [(1000 0000) (0000 0000) (0000 0000) (0000 0000)]
        val mask = (1 shl (31 - i))
        bits[i] = (this and mask) != 0
    }

    return bits
}

fun Int.toBinaryArray16(): BooleanArray {
    // 16비트 크기의 BooleanArray 생성
    val bits = BooleanArray(16)

    for (i in 15 downTo 0) { // MSB에서 LSB로 진행
        // mask 값
        // i=16, shl  0 [(0000 0000) (0000 0001)]
        // i=15, shl  1 [(0000 0000) (0000 0010)]
        // i=14, shl  2 [(0000 0000) (0000 0100)]
        //            .......
        // i= 3, shl 13 [(0001 0000) (0000 0000)]
        // i= 2, shl 14 [(0010 0000) (0000 0000)]
        // i= 1, shl 15 [(0100 0000) (0000 0000)]
        // i= 0, shl 16 [(1000 0000) (0000 0000)]
        val mask = (1 shl (15 - i))

        // -8 경우
        //     [(1111 1111) (1111 1000)]
        // and [(0000 0000) (0000 0001)] => [(0000 0000) (0000 0000)] => No bit = 0
        // and [(0000 0000) (0000 0010)] => [(0000 0000) (0000 0000)] => No bit = 0
        // and [(0000 0000) (0000 0100)] => [(0000 0000) (0000 0000)] => No bit = 0
        // and [(0000 0000) (0000 1000)] => [(0000 0000) (0000 1000)] => 2^3 = 8
        // and [(0000 0000) (0001 0000)] => [(0000 0000) (0001 0000)] => 2^4 = 16
        // and          ....
        // and [(0010 0000) (0000 0000)] => [(0010 0000) (0000 0000)] => 2^13 = 8192
        // and [(0100 0000) (0000 0000)] => [(0100 0000) (0000 0000)] => 2^14 = 16384
        // and [(1000 0000) (0000 0000)] => [(1000 0000) (0000 0000)] => 2^15 = 32768
        bits[i] = (this and mask) != 0 // 마스킹한 결과 0이 아니면 해당 비트가 존재한다는 의미입니다.
    }

    return bits
}


fun main() {
    // 16비트 경우
    listOf(
        Pair(
            -8, booleanArrayOf(
                // 1111 1111 1111 1000
                //
                // 이는 `8.inv() + 1`(2의 보수)입니다:
                //   ~0000 0000 0000 1000
                // => 1111 1111 1111 0111
                //   +                  1
                //    1111 1111 1111 1000
                true, true, true, true,    // 16
                true, true, true, true,    // 12
                true, true, true, true,    // 8
                true, false, false, false, // 4
            )
        ),
        Pair(
            8, booleanArrayOf(
                // 0000 0000 0000 1000
                false, false, false, false, // 16
                false, false, false, false, // 12
                false, false, false, false, // 8
                true, false, false, false,  // 4
            )
        ),
        Pair(
            0, booleanArrayOf(
                // 0000 0000 0000 0000
                false, false, false, false, // 16
                false, false, false, false, // 12
                false, false, false, false, // 8
                false, false, false, false, // 4
            )
        ),
        Pair(
            -0, booleanArrayOf(
                // 0000 0000 0000 0000
                //
                // 이는 `0.inv() + 1`(2의 보수)입니다:
                //   ~0000 0000 0000 0000
                // => 1111 1111 1111 1111
                //   +                  1
                //    0000 0000 0000 0000
                false, false, false, false, // 16
                false, false, false, false, // 12
                false, false, false, false, // 8
                false, false, false, false, // 4
            )
        ),
    ).forEach { tc ->
        println(tc.first.toBinaryArray16().contentEquals(tc.second))

    }

    // 32비트 경우
    listOf(
        Pair(
            -8, booleanArrayOf(
                // 1111 1111 1111 1111 1111 1111 1111 1000
                //
                // 이는 `8.inv() + 1`(2의 보수)입니다:
                //   ~0000 0000 0000 0000 0000 0000 0000 1000
                // => 1111 1111 1111 1111 1111 1111 1111 0111
                //   +                                      1
                //    1111 1111 1111 1111 1111 1111 1111 1000
                true, true, true, true,    // 32
                true, true, true, true,    // 28
                true, true, true, true,    // 24
                true, true, true, true,    // 20
                true, true, true, true,    // 16
                true, true, true, true,    // 12
                true, true, true, true,    // 8
                true, false, false, false, // 4
            )
        ),
    ).forEach { tc ->
        println(tc.first.toBinaryArray32().contentEquals(tc.second))
    }
}