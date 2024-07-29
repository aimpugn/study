# timedatectl

## `timedatectl`이란?

시스템의 시간과 날짜를 제어합니다.
- 시스템 시계와 그 설정을 조회 및 변경
- 시간 동기화 서비스를 활성화 또는 비활성화

마운트된(하지만 부팅되지 않은) 시스템 이미지의 시스템 표준 시간대를 초기화하려면 `systemd-firstboot`(1)를 사용합니다.

`timedatectl`을 사용하여 시간 동기화 서비스의 현재 상태를 표시할 수 있습니다(예: `systemd-timesyncd.service`(8)).

## 타임존 목록 확인하기

```sh
timedatectl list-timezones
```

## 현재 타임존 보기

```sh
ubuntu@ip-172-31-14-198:~$ timedatectl
               Local time: Fri 2024-08-30 12:38:28 UTC
           Universal time: Fri 2024-08-30 12:38:28 UTC
                 RTC time: Fri 2024-08-30 12:38:28
                Time zone: Etc/UTC (UTC, +0000)
System clock synchronized: yes
              NTP service: active
          RTC in local TZ: no
```

```sh
ubuntu@ip-172-31-14-198:~$ timedatectl show
Timezone=Etc/UTC
LocalRTC=no
CanNTP=yes
NTP=yes
NTPSynchronized=yes
TimeUSec=Fri 2024-08-30 12:38:43 UTC
RTCTimeUSec=Fri 2024-08-30 12:38:42 UTC
```

## 타임존 변경

```sh
sudo timedatectl set-timezone Asia/Seoul
```
