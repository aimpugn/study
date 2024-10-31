package main

import (
    "fmt"
    "math"
)

func main() {
    fmt.Printf("%08b\n", 2)
    fmt.Printf("%08b\n", 2<<1)

    printFloatIn32BitsIntoIEEE754(float32(3.25))
    printFloatIn32BitsIntoIEEE754(float32(12.375))
    printFloatIn32BitsIntoIEEE754(float32(12.375) * float32(3.25))
}

// printFloatIn32BitsIntoIEEE754 함수는 주어진 float32 숫자를 IEEE 754 사양으로
// 파싱하여 출력합니다.
//
// Output:
//
//      IEEE 754:
//      - original: 12.375000
//      - bits: 01000001010001100000000000000000
//      - Sign bit: 0
//      - Exponent: 10000010 (130)
//      - Mantissa: 1.10001100000000000000000
func printFloatIn32BitsIntoIEEE754(num float32) {
    bits := math.Float32bits(num)

    fmt.Printf("\nIEEE 754:\n")
    fmt.Printf("- original: %f\n", num)
    fmt.Printf("- bits: %032b\n", bits)

    // IEEE 754 형식에서 각 부분 추출
    // 32비트 기준, 최상위 비트만 남깁니다.
    sign := bits >> 31
    // 원본 bits:
    // s(1) eeeeeeee(8) mmmmmmmmmmmmmmmmmmmmmmm(23)
    //      └─ 지수부
    //
    // >> 23 연산 후:
    // 000000000seeeeeeee
    //          └─ 지수부가 하위 8비트로 이동
    //
    // 0xFF와 AND 연산:
    // 000000000seeeeeeee
    // 000000000011111111 (0xFF)
    // -----------------
    // 000000000[8bit만 남음]
    //
    // 이 연산으로 지수부 8비트만 추출됨
    exponent := (bits >> 23) & 0xFF
    // 원본 bits:
    // s(1) eeeeeeee(8) mmmmmmmmmmmmmmmmmmmmmmm(23)
    //                  └─ 가수부
    //
    // fmt.Printf("%032b\n", 0x7FFFFF)
    // 00000000011111111111111111111111
    // 하위 23비트가 1인 값.
    //
    // AND 연산으로 하위 23비트(가수부)만 남음
    mantissa := bits & 0x7FFFFF

    fmt.Printf("- Sign bit: %d\n", sign)
    fmt.Printf("- Exponent: %08b (%d)\n", exponent, exponent)
    fmt.Printf("- Mantissa: 1.%023b\n", mantissa)
    fmt.Println()
}
