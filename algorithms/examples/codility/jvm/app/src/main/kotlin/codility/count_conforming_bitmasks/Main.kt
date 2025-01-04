package codility.count_conforming_bitmasks

/**
 * unsigned 30 bits 정수를 고려합니다.
 * 0 <= 정수 < 2^30
 *
 * 정수 B의 비트중 1로 설정된 모든 위치에 대해, A가 그에 상응하는 비트가 1로 설정되어 있다면,
 * 정수 A가 정수 B를 부합(conform)한다고 말합니다.
 *
 * ```
 * 정수 A: 00 0000 1111 0111 1101 1110 0000 1111 = 16,244,239 는
 * 정수 B: 00 0000 1100 0110 1101 1110 0000 0001 = 13,032,961 에 대해 부합합니다.
 *                ^^    ^^  ^^ ^ ^^^          ^
 * ```
 *
 * ```
 * 정수 A: 11 0000 1101 0111 0000 1010 0000 0101 = 819,399,173 는
 * 정수 B: 00 0000 1001 0110 0011 0011 0000 1111 = 9,843,471 에 대해 부합하지 않습니다.
 *                ^  ^  ^^    xx   ^x      x x
 * ```
 *
 * 가령, 다음과 같이 A, B, C가 주어졌을 때,
 * ```
 * A: 11 1111 1111 1111 1111 1111 1001 1111 = 1,073,741,727
 * B: 11 1111 1111 1111 1111 1111 0011 1111 = 1,073,741,631
 * C: 11 1111 1111 1111 1111 1111 0110 1111 = 1,073,741,679
 * ```
 * - A에 부합하는 정수
 *    - 11 1111 1111 1111 1111 1111 1001 1111 = 1,073,741,727 == A 자신도 포함
 *    - 11 1111 1111 1111 1111 1111 1011 1111 = 1,073,741,759
 *    - 11 1111 1111 1111 1111 1111 1101 1111 = 1,073,741,791
 *    - 11 1111 1111 1111 1111 1111 1111 1111 = 1,073,741,823
 * - B에 부합하는 정수
 *    - 11 1111 1111 1111 1111 1111 0011 1111 = 1,073,741,631 == B
 *    - 11 1111 1111 1111 1111 1111 0111 1111 = 1,073,741,695
 *    - 11 1111 1111 1111 1111 1111 1011 1111 = 1,073,741,759 (A 경우에 중복)
 *    - 11 1111 1111 1111 1111 1111 1111 1111 = 1,073,741,823 (A, B 경우에 중복)
 * - C에 부합하는 정수
 *    - 11 1111 1111 1111 1111 1111 0110 1111 = 1,073,741,679 == C
 *    - 11 1111 1111 1111 1111 1111 1110 1111 = 1,073,741,807
 *    - 11 1111 1111 1111 1111 1111 0111 1111 = 1,073,741,695 (B 경우에 중복)
 *    - 11 1111 1111 1111 1111 1111 1111 1111 = 1,073,741,823 (A, B 경우에 중복)
 *
 * `0000`은 존재하지 않습니다. 중복 제거하고 8개가 부합하여 8을 반환합니다.
 *
 * 극단적으로 `00 0000 0000 0000 0000 0000 0000 0001` 경우 2^29 일 텐데... 모든 경우의 수별로 중복 체크하는 건 주어진 시간 내에 불가능해 보입니다.
 *
 * - A 경우 0인 부분이 2개, 총 2^2 네 가지 경우 가능
 * - B 경우 0인 부분이 2개, 총 2^2 네 가지 경우 가능
 * - C 경우 0인 부분이 2개, 총 2^2 네 가지 경우 가능
 * - A와 B 경우 1011, 1111 두 개가 겹칠 수 있음 (OR 연산 후 경우의 수)
 * - A와 C 경우 1111 한 개가 겹칠 수 있음 (OR 연산 후 경우의 수)
 * - B와 C 경우 0111, 1111 두 개가 겹칠 수 있음 (OR 연산 후 경우의 수)
 * - A&B, A&C, B&C 중복되는 경우를 빼고, A&B&C 세 개 겹치는 경우(여기서는 1111)을 더합니다.
 * (4 + 4 + 4) - (2 + 1 + 2) + 1
 *
 * @param A 0 <= A <= 1,073,741,823
 * @param B 0 <= B <= 1,073,741,823
 * @param C 0 <= C <= 1,073,741,823
 *
 * @return 적어도 A, B, C 중 하나에 부합하는 unsigned 30 bit 정수들의 개수
 */
fun solution(A: Int, B: Int, C: Int): Int {
    // A, B, C 등 어떤 정수가 주어지면 그에 부합하는 수를 만들어야 합니다.
    // 30비트중 1이 아닌 부분 체크하고, 0/1 둘 중 하나 들어가는 경우의 수를 계산하면 될 거 같습니다.
    // 4 비트 경우, 8 = 1000
    // 이에 부합하는 수는, 1은 반드시 존재해야 하고, 000 케이스는 존재하므로, 2^3 - 1 = 7개입니다.
    // - 1 001
    // - 1 010
    // - 1 100
    // - 1 011
    // - 1 101
    // - 1 110
    // - 1 111

    val possibleA = 2.pow(30 - Integer.bitCount(A))
    val possibleB = 2.pow(30 - Integer.bitCount(B))
    val possibleC = 2.pow(30 - Integer.bitCount(C))
    val possibleAB = 2.pow(30 - Integer.bitCount(A or B))
    val possibleAC = 2.pow(30 - Integer.bitCount(A or C))
    val possibleBC = 2.pow(30 - Integer.bitCount(B or C))
    val possibleABC = 2.pow(30 - Integer.bitCount(A or B or C))

    // println("($possibleA + $possibleB + $possibleC) - ($possibleAB + $possibleAC + $possibleBC) + $possibleABC")

    return (possibleA + possibleB + possibleC) - (possibleAB + possibleBC + possibleAC) + possibleABC
}

/**
 * [Integer.bitCount] 메서드의 동작 방식을 단계별로 따라가 봅니다.
 *
 * [Integer.bitCount]는 비트 병렬 처리(bit-parallel processing) 기법에 기반한 알고리즘입니다.
 *
 * References:
 * - [Counting Bits](https://www.manniwood.com/2019_08_24/counting_bits.html)
 * - [Counting bits set, in parallel](https://graphics.stanford.edu/~seander/bithacks.html#CountBitsSetParallel)
 * - [Software Optimization Guide for AMD Athlon™ 64 and AMD Opteron™ Processors - 8.6 Efficient Implementation of Population-Count Function in 32-Bit Mode](https://cr.yp.to/bib/2004/-amd-25112.pdf)
 */
fun Int.countOneBits(): Int {
    // HD, Figure 5-2
    var target = this
    // 1. i = i - ((i >>> 1) & 0x55555555); => 각 2비트 쌍에 대해 1의 개수를 계산합니다.
    //    a. `(i >>> 1)`는 `i`를 오른쪽으로 1비트 쉬프트하여, 각 비트 쌍이 1의 개수를 비교하게 만듭니다.
    //    b. 홀수 위치의 비트만 유지하는 마스크(`0x55555555`)와 `&` 현산을 수행하여 짝수 비트에 영향을 주지 않도록 합니다.
    //       ```
    //       0x55555555
    //       => [(0101 0101) (0101 0101) (0101 0101) (0101 0101)]
    //       ```
    //    c. 각 2비트 쌍의 1의 개수를 계산합니다.
    var tmp = target
    // 정수를 2비트 그룹으로 분할하여, 각 그룹의 1의 개수(population count)를 계산합니다.
    // ```
    // 새로운 값 = 기존 값 - (기존 값 >> 1)
    // v - (v >> 1)
    // 00b -> 00b => 0개
    // 01b -> 01b => 1개
    // 10b -> 01b => 1개
    // 11b -> 10b => 2개
    // ```
    // 즉, 이 방법은 `00` 비트에서 1의 개수는 0, 1, 2개가 존재할 수 있다는 것에 기반합니다.
    //
    // `0x55555555`는 `[(0101 0101) (0101 0101) (0101 0101) (0101 0101)]₂` 입니다.
    // 이 마스크는 모든 짝수 비트를 유지하고, 홀수 비트를 제거합니다.
    // 오른쪽으로 이동한 값에서 다른 그룹에 영향이 가지 않도록 같은 그룹 내에서만 동작하도록 제한합니다.
    // ```
    // w = v - ((v >> 1) & 0x55555555)
    //
    // 01b >> 1 => 00b
    //                                                00
    // & [(0101 0101) (0101 0101) (0101 0101) (0101 0101)]₂
    // => w = 1 - 0 = 1
    //
    //
    // 11b >> 1 => 01b
    //                                                01
    // & [(0101 0101) (0101 0101) (0101 0101) (0101 0101)]₂
    // => w = 3 - 1 = 2
    // ```
    //
    // 2비트 그룹을 `ab`라고 할 때
    // - `a`는 상위 비트, `b`는 하위 비트입니다.
    // - `v`의 2비트 그룹 값은 `2a + b`입니다.
    // - `(v >> 1)`의 값은 `a`입니다.
    // 따라서,
    // `(v - (v >> 1))`는
    // => `(2a + b) - a`가 되고
    // => `a + b`가 됩니다.
    // 즉, 각 2비트 그룹에서 1의 개수(`a + b`)를 계산하는 게 됩니다.
    target = target - ((target shr 1) and 0x55555555)
    println("1. target = target - ((target shr 1) and 0x55555555)")
    println("\t=>   ${tmp.toBinaryStringWithBase()} shr 1")
    println("\t=>   ${(tmp shr 1).toBinaryStringWithBase()}")
    println("\t   & ${(0x55555555).toBinaryStringWithBase()} => ${(tmp shr 1) and 0x55555555}")
    println("\t=> $tmp - ${(tmp shr 1) and 0x55555555}")
    println("\t=> $target => ${target.toBinaryStringWithBase()}")

    // 2. i = (i & 0x33333333) + ((i >>> 2) & 0x33333333); => 각 4비트 청크(chunk)에서 1의 개수를 계산합니다.
    //    a. 각 4비트 청크에서 하위 2비트를 유지하고 나머지를 0으로 설정하는 마스크(`0x33333333`)와 `&` 연산을 수행합니다.
    //       ```
    //       0x33333333
    //       => [(0011 0011) (0011 0011) (0011 0011) (0011 0011)]
    //       ```
    //    b. 2비트 오른쪽으로 쉬프트`(i >>> 2)`하여 각 2비트 그룹을 합산할 수 있게 만듭니다.
    tmp = target
    target = (target and 0x33333333) + ((target shr 2) and 0x33333333)
    println("2. target = (target and 0x33333333) + ((target shr 2) and 0x33333333)")
    println("\t=>   ${tmp.toBinaryStringWithBase()}")
    println("\t   & ${0x33333333.toBinaryStringWithBase()} => ${(tmp and 0x33333333)}")
    println("\t                            +")
    println("\t     ${(tmp shr 2).toBinaryStringWithBase()}")
    println("\t   & ${(0x33333333).toBinaryStringWithBase()} => ${(tmp shr 2) and 0x33333333}")
    println("\t=> ${(tmp and 0x33333333)} + ${(tmp shr 2) and 0x33333333} = $target")

    // 3. i = (i + (i >>> 4)) & 0x0f0f0f0f; => 각 8비트 청크 내에서 1의 개수를 계산합니다.
    //    a. 각 8비트 청크에서 하위 4비트를 유지하고 나머지를 0으로 설정하는 마스크(`0x0f0f0f0f`)와 `&` 연산을 수행합니다.
    //       ```
    //       0x0f0f0f0f
    //       => [(0000 1111) (00001111) (0000 1111) (0000 1111)]
    //       ```
    //    b. 4비트 오른쪽으로 쉬프트`(i >>> 2)`하여 각 4비트 그룹을 합산할 수 있게 만듭니다.
    tmp = target
    target = (target + (target shr 4)) and 0x0f0f0f0f
    println("3. (target + (target shr 4)) and 0x0f0f0f0f")
    println("\t=> ( $tmp + ${(tmp shr 4).toBinaryStringWithBase()} ) and 0x0f0f0f0f")
    println("\t=> ( $tmp + ${(tmp shr 4)} ) and 0x0f0f0f0f")
    println("\t=> ( ${(tmp + (tmp shr 4))} ) and 0x0f0f0f0f")
    println("\t=>   ${(tmp + (tmp shr 4)).toBinaryStringWithBase()}")
    println("\t   & ${(0x0f0f0f0f).toBinaryStringWithBase()}")
    println("\t=> $target")

    // 4. i = i + (i >>> 8); => 16비트 내의 1의 개수를 합산합니다.
    tmp = target
    target = target + (target shr 8)
    println("4. target = target + (target shr 8)")
    println("\t=> ( $tmp + ${(tmp shr 8).toBinaryStringWithBase()} )")
    println("\t=> ( $tmp + ${tmp shr 8} )")
    println("\t=> $target")

    // 5. i = i + (i >>> 16); => 32비트 내의 1의 개수를 합산합니다.
    tmp = target
    target = target + (target shr 16)
    println("5.  target = target + (target shr 16)")
    println("\t=> ( $tmp + ${(tmp shr 16).toBinaryStringWithBase()} )")
    println("\t=> ( $tmp + ${tmp shr 16} )")
    println("\t=> $target")

    // 6. return i & 0x3f; => 최종 결과에서 하위 6비트만 반환합니다.
    tmp = target
    println("return target & 0x3f;")
    println("\t=>   ${tmp.toBinaryStringWithBase()}")
    println("\t   & ${0x3f.toBinaryStringWithBase()}")
    println("\t=>   ${tmp and 0x3f}")

    return target and 0x3f;
}

/**
 * 아래첨자(subscript) "₂"와 함께 이진 표현을 리턴합니다.
 */
fun Int.toBinaryStringWithBase(): String {
    val sb = StringBuilder(Integer.toBinaryString(this))
    while (sb.length < 32)
        sb.insert(0, "0")

    var idx = 0
    var added = false
    sb.insert(idx, "(")
    idx++
    // `(0000 0000) (0000 0000) (0000 0000) (0000 `: 길이 42
    while (idx < 42) {
        //  0 1 2 3 4 5 6 7
        // [0 0 0 0 0 0 0 0]
        //         ^
        //         sb.insert(4, " ")

        idx += 4 // 4개로 묶기 위함
        // 공백이 하나 추가 됐다면,
        // "0000000000000000" => "0000 000000000000" 이렇게 변했다는 의미입니다.
        if (added) {
            // "0000 000000000000" => "0000 0000) (00000000" 이렇게 변경해줍니다.
            sb.insert(idx, ") (")
            idx += 3 // ") (" 문자 추가로 인한 인덱스 보정
            added = false
        } else {
            // "0000 0000) (00000000" => "0000 0000) (0000 0000" 이렇게 변경해줍니다.
            sb.insert(idx, " ")
            idx += 1 // " " 문자 추가로 인한 인덱스 보정
            added = true
        }
    }
    sb.append(")")

    return "[${sb}]\u2082"
}


fun Int.pow(exponent: Int): Int {
    var result = 1
    var tmp = this
    var exp = exponent

    while (exp > 0) {
        if ((exp and 1) == 1) {
            result *= tmp
        }
        tmp *= tmp
        exp = exp shr 1
    }

    return result
}

data class TestCase(val a: Int, val b: Int, val c: Int, val answer: Int)


fun main() {

    println("======= 255: v - (v >> 1) =======")
    /**
     * 255 => [(0000 0000) (0000 0000) (0000 0000) (1111 1111)]₂
     *        (11) (11) (11) (11), 2 비트 그룹으로 1이 2개씩 존재합니다
     * 127 => [(0000 0000) (0000 0000) (0000 0000) (0111 1111)]₂
     * 127 & 0x55555555 = 85
     * 85  => [(0000 0000) (0000 0000) (0000 0000) (0101 0101)]₂
     * 255 - 85 = 170
     *        [(0000 0000) (0000 0000) (0000 0000) (1111 1111)]₂
     *      - [(0000 0000) (0000 0000) (0000 0000) (0101 0101)]₂
     * 170 => [(0000 0000) (0000 0000) (0000 0000) (1010 1010)]₂
     *        (10) (10) (10) (10), 2비트 그룹으로 1이 2개씩 존재한다는 의미가 됩니다.`(v - (v >> 1)`)
     *
     * 천재다, 진짜...
     */
    println(0b11111111) // 255
    println(0b11111111 shr 1) // 127
    println((0b11111111 shr 1).toBinaryStringWithBase()) // [(0000 0000) (0000 0000) (0000 0000) (0111 1111)]₂
    println((0b11111111 shr 1) and 0x55555555) // 85
    println(((0b11111111 shr 1) and 0x55555555).toBinaryStringWithBase()) // [(0000 0000) (0000 0000) (0000 0000) (0101 0101)]₂


    println("======= 219: v - (v >> 1) =======")
    /**
     * 219 => [(0000 0000) (0000 0000) (0000 0000) (1101 1011)]₂
     *        (11) (01) (10) (11), 2 비트 그룹으로 1이 2개, 1개, 1개, 2개 존재합니다.
     * 109 => [(0000 0000) (0000 0000) (0000 0000) (0110 1101)]₂
     * 109 & 0x55555555 = 69
     * 69  => [(0000 0000) (0000 0000) (0000 0000) (0100 0101)]₂
     * 219 - 69 = 150
     *        [(0000 0000) (0000 0000) (0000 0000) (1101 1011)]₂
     *      - [(0000 0000) (0000 0000) (0000 0000) (0100 0101)]₂
     * 150 => [(0000 0000) (0000 0000) (0000 0000) (1001 0110)]₂
     *        (10) (01) (01) (10), 2비트 그룹으로 1이 2개, 1개, 1개, 2개 존재한다는 의미가 됩니다.`(v - (v >> 1)`)
     *        즉, `219 - 69 = 150` 과정을 통해 "같은 2비트 그룹 내에서 비트 간의 차이를 계산하여, 1의 개수를 합산"합니다.
     */
    println(0b11011011) // 219
    println(0b11011011 shr 1) // 109
    println((0b11011011 shr 1).toBinaryStringWithBase()) // [(0000 0000) (0000 0000) (0000 0000) (0110 1101)]₂
    println((0b11011011 shr 1) and 0x55555555) // 69
    println(((0b11011011 shr 1) and 0x55555555).toBinaryStringWithBase()) // [(0000 0000) (0000 0000) (0000 0000) (0100 0101)]₂
    println(0b11011011 - ((0b11011011 shr 1) and 0x55555555)) // 150
    println((0b11011011 - ((0b11011011 shr 1) and 0x55555555)).toBinaryStringWithBase()) // [(0000 0000) (0000 0000) (0000 0000) (1001 0110)]₂

    println("======= 255.countOneBits =======")
    println(255.countOneBits()) // 8개

    listOf(
        TestCase(
            a = 1_073_741_727,
            b = 1_073_741_631,
            c = 1_073_741_679,
            answer = 8,
        ),
        TestCase(
            // 11 1111 1111 1111 1111 1111 1111 1110
            // 11 1111 1111 1111 1111 1111 1111 1100
            // 11 1111 1111 1111 1111 1111 1111 1000
            a = 1_073_741_822,
            b = 1_073_741_820,
            c = 1_073_741_816,
            // - A 경우
            //   - 11 1111 1111 1111 1111 1111 1111 1110
            //   - 11 1111 1111 1111 1111 1111 1111 1111
            // - B 경우
            //   - 11 1111 1111 1111 1111 1111 1111 1100
            //   - 11 1111 1111 1111 1111 1111 1111 1101
            //   - 11 1111 1111 1111 1111 1111 1111 1110 (A & B 중복)
            //   - 11 1111 1111 1111 1111 1111 1111 1111 (A & B & C 중복)
            // - C 경우
            //   - 11 1111 1111 1111 1111 1111 1111 1000
            //   - 11 1111 1111 1111 1111 1111 1111 1001
            //   - 11 1111 1111 1111 1111 1111 1111 1010
            //   - 11 1111 1111 1111 1111 1111 1111 1100 (B & C 중복)
            //   - 11 1111 1111 1111 1111 1111 1111 1011
            //   - 11 1111 1111 1111 1111 1111 1111 1101 (B & C 중복)
            //   - 11 1111 1111 1111 1111 1111 1111 1110 (A & C 중복)
            //   - 11 1111 1111 1111 1111 1111 1111 1111 (A & B & C 중복)
            answer = 8,
        ),
    ).forEach { tc ->
        println(solution(tc.a, tc.b, tc.c))
    }

}