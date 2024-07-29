<?php

class BenchMarkString
{
    public function main()
    {
        // 패턴과 반복 횟수 설정
        $pattern = 'abcd-efgh-ij-';
        $patternLength = strlen($pattern);
        $iterations = 100 * 1000; // 반복 횟수

        // 벤치마크 실행
        $this->benchmark($iterations, $pattern, $patternLength);
    }

    // 벤치마크 테스트 함수
    public function benchmark($iterations, $pattern, $patternLength)
    {
        $strncmpTime = 0;
        $substrTime = 0;

        for ($i = 0; $i < $iterations; $i++) {
            $testString = $this->generateRandomString(rand(20, 100)) . $pattern; // 패턴이 확실히 포함되도록 함

            // strncmp 테스트
            $start = microtime(true);
            $result = strncmp($testString, $pattern, $patternLength) === 0;
            $strncmpTime += microtime(true) - $start;

            // substr 테스트
            $start = microtime(true);
            $result = substr($testString, 0, $patternLength) === $pattern;
            $substrTime += microtime(true) - $start;
        }

        // 7.6461553573608 × 10^-7 seconds, 0.00000076461553573608 seconds, 0.765 마이크로초
        echo 'Average time for strncmp: ' . ($strncmpTime / $iterations) . " seconds\n";
        // 7.7724695205688 × 10^-7 seconds, 0.00000077724695205688 seconds, 0.777 마이크로초
        echo 'Average time for substr: ' . ($substrTime / $iterations) . " seconds\n";
    }

    // 문자열 생성 함수
    public function generateRandomString($length)
    {
        $characters = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
        $charactersLength = strlen($characters);
        $randomString = '';
        for ($i = 0; $i < $length; $i++) {
            $randomString .= $characters[rand(0, $charactersLength - 1)];
        }
        return $randomString;
    }
}
