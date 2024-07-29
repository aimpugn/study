# ssh

- [ssh](#ssh)
    - [ssh 실행 명령어 예시](#ssh-실행-명령어-예시)
    - [포트 변경하기](#포트-변경하기)
        - [Ubuntu 24.04](#ubuntu-2404)
    - [RDS 포트 포워딩](#rds-포트-포워딩)

## ssh 실행 명령어 예시

```sh
/usr/sbin/sshd -D \
    -o AuthorizedKeysCommand /usr/share/ec2-instance-connect/eic_run_authorized_keys %u %f \
    -o AuthorizedKeysCommandUser ec2-instance-connect [listener] 0 of 10-100 startups
```

## 포트 변경하기

### Ubuntu 24.04

- 기존 `sshd_config` 파일을 백업합니다.

    ```sh
    cp /etc/ssh/sshd_config{,.$(date +'%Y%m%d_%H%M')}
    ```

- `/etc/ssh/sshd_config` Port 부분을 원하는 포트로 변경합니다.

- ssh 서비스를 재시작합니다.

    [Ubuntu 22.04 이후부터는 소켓 방식으로 변경](https://askubuntu.com/a/1439482) 됐다고 합니다.

    > [SSHd now uses socket-based activation (Ubuntu 22.10 and later)](https://discourse.ubuntu.com/t/sshd-now-uses-socket-based-activation-ubuntu-22-10-and-later/30189)
    >
    > As of version 1:9.0p1-1ubuntu1 of openssh-server in Kinetic Kudu (Ubuntu 22.10),
    > OpenSSH in Ubuntu is configured by *default to use systemd socket activation*.
    > This means that *`sshd` will not be started until an incoming connection request is received*.
    > This has been done to reduce the memory consumed by Ubuntu Server instances by default,
    > which is of particular interest with Ubuntu running in VMs or LXD containers: by not running `sshd` when it is not used, we save at least 3MiB of memory in each instance, representing a savings of roughly 5% on an idle, pristine kinetic container.
    >
    > *kinetic = Ubuntu 22.10(Kinetic Kudu)

    따라서 별도 요청이 들어오기 전에 바로 반영하려면 데몬 리로드하고 ssh 서비스를 재시작합니다.

    ```sh
    # Reload systemd manager configuration
    sudo systemctl daemon-reload
    sudo systemctl restart ssh
    ```

그리고 실제로 잘 적용됐는지 확인합니다.

```log
~$ sudo systemctl status ssh

● ssh.service - OpenBSD Secure Shell server
     Loaded: loaded (/usr/lib/systemd/system/ssh.service; disabled; preset: enabled)
    Drop-In: /usr/lib/systemd/system/ssh.service.d
             └─ec2-instance-connect.conf
     Active: active (running) since Fri 2024-08-30 21:31:51 KST; 10min ago
TriggeredBy: ● ssh.socket
       Docs: man:sshd(8)
             man:sshd_config(5)
    Process: 2394 ExecStartPre=/usr/sbin/sshd -t (code=exited, status=0/SUCCESS)
   Main PID: 2395 (sshd)
      Tasks: 1 (limit: 4676)
     Memory: 2.1M (peak: 2.8M)
        CPU: 40ms
     CGroup: /system.slice/ssh.service
             └─2395 "sshd: /usr/sbin/sshd -D -o AuthorizedKeysCommand /usr/share/ec2-instance-connect/eic_run_authorized_keys %u %f ->

Aug 30 21:31:51 ip-172-31-14-198 systemd[1]: Starting ssh.service - OpenBSD Secure Shell server...
Aug 30 21:31:51 ip-172-31-14-198 sshd[2395]: Server listening on 0.0.0.0 port 22022.
Aug 30 21:31:51 ip-172-31-14-198 sshd[2395]: Server listening on :: port 22022.
... 생략 ...
```

- [](https://www.lesstif.com/lpt/ssh-22-20776114.html)

## [RDS 포트 포워딩](https://aws.amazon.com/ko/premiumsupport/knowledge-center/rds-connect-using-bastion-host-linux/)

```shell
# Syntax 1:
ssh -i <identity_file> -f \
    -l <bastion-host-username> \
    -L <local-port-you-connect-to>:<rds-endpoint>:<rds:listening-port> \
    <bastion-host-public-ip> -v

# Example Command:
ssh -i "private_key.pem" -f -l ec2-user -L 5432:172.31.39.62:5432  3.133.141.189 -v
```

```shell
# Syntax 2:
ssh -i "Private_key.pem" -f -N \
    -L 5433:RDS_Instance_Endpoint:5432 \
    ec2-user@EC2-Instance_Endpoint -v

# Example Command:
ssh -i "private.pem" -f -N \
    -L 5433:pg115.xxxx.us-east-2.rds.amazonaws.com:5432 \
    ec2-user@ec2-xxxx-xxx9.us-east-2.compute.amazonaws.com -v
```
