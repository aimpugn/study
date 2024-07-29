package datetime

import (
	"fmt"
	"time"
)

func DateTime() {
	now := time.Now()
	micros := now.Nanosecond() / 1000

	// 현재 날짜 및 시간을 "YYYYMMDDHHMMSS" 형식으로 포맷합니다.
	// 그리고 마이크로초를 3자리로 포맷하여 최종 "YYYYMMDDHHMMSSSSS" 형식으로 결합합니다.
	dateStr := now.Format("20060102150405") + fmt.Sprintf("%03d", micros/1000)
	dateStrCon := dateStr[:14] + dateStr[15:] //YYYYMMDDHHMMSSxSS (중간의 x값은 버림, millisecond의 첫번째 자리수)
	fmt.Println(dateStr)
	fmt.Println(dateStrCon)
}
