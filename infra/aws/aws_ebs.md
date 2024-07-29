# AWS EBS

- [AWS EBS](#aws-ebs)
    - [EBS](#ebs)
    - [EBS 볼륨 연결 확인](#ebs-볼륨-연결-확인)
    - [용량 확장](#용량-확장)
        - [EBS 볼륨, 파티션, 그리고 파일 시스템 확장](#ebs-볼륨-파티션-그리고-파일-시스템-확장)
        - [1. 볼륨 수정](#1-볼륨-수정)
        - [2. 파일시스템 확장](#2-파일시스템-확장)
            - [2.1. 파티션 크기를 조정](#21-파티션-크기를-조정)
            - [2.2. 파일 시스템을 확장](#22-파일-시스템을-확장)
    - [기타](#기타)

## [EBS](https://docs.aws.amazon.com/ko_kr/ebs/latest/userguide/ebs-volumes.html)

Amazon EBS 볼륨은 내구성이 있는 블록 수준 스토리지 디바이스이며 인스턴스를 연결하는 것이 가능합니다.
볼륨을 인스턴스에 연결하면 물리적 하드 드라이브처럼 사용할 수 있습니다

즉, EBS 자체는 블록 스토리지이며 이를 EC2 인스턴스에 붙여서 OS가 사용할 수 있게끔 만듦으로써 EC2의 파일 시스템으로 동작합니다.

[Amazon Elastic Block Store에서 다음과 같은 블록 스토리지 리소스를 생성하고 관리](https://docs.aws.amazon.com/ko_kr/ebs/latest/userguide/what-is-ebs.html)할 수 있습니다.
- Amazon EBS 볼륨

    Amazon EC2 인스턴스에 연결하는 스토리지 볼륨입니다.
    볼륨을 인스턴스에 연결하면 해당 볼륨을 컴퓨터에 연결된 로컬 하드 드라이브처럼 사용할 수 있습니다(예: 파일 저장 또는 애플리케이션 설치).

- Amazon EBS 스냅샷

    볼륨 자체와 관계없이 지속되는 Amazon EBS 볼륨의 특정 시점 백업입니다.
    Amazon EBS 볼륨의 데이터를 백업하는 스냅샷을 생성할 수 있습니다.
    그러면 언제든지 해당 스냅샷에서 새 볼륨을 복원할 수 있습니다.

## EBS 볼륨 연결 확인

```sh
$ ls -asl /dev/xvd*
0 brw-rw---- 1 root disk 202,  0 Aug 30 16:56 /dev/xvda
0 brw-rw---- 1 root disk 202,  1 Aug 30 16:56 /dev/xvda1
0 brw-rw---- 1 root disk 202, 14 Aug 30 16:56 /dev/xvda14
0 brw-rw---- 1 root disk 202, 15 Aug 30 16:56 /dev/xvda15
0 brw-rw---- 1 root disk 259,  0 Aug 30 16:56 /dev/xvda16

# list block devices
$ lsblk
NAME     MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
loop0      7:0    0 25.2M  1 loop /snap/amazon-ssm-agent/7993
loop1      7:1    0 55.7M  1 loop /snap/core18/2829
loop2      7:2    0 38.8M  1 loop /snap/snapd/21759
xvda     202:0    0    8G  0 disk
├─xvda1  202:1    0    7G  0 part /
├─xvda14 202:14   0    4M  0 part
├─xvda15 202:15   0  106M  0 part /boot/efi
└─xvda16 259:0    0  913M  0 part /boot
```

## 용량 확장

Amazon EBS 볼륨은 인스턴스에 연결하면 블록 디바이스로 표시됩니다.

Amazon EBS 볼륨 확장은 AWS 인프라스트럭쳐 수준에서 이뤄지는 작업이며,
블록 디바이스의 원시 스토리지 용량(raw storage capacity)를 증가시킵니다.
하지만 용량이 추가되었다고 해서 바로 사용할 수는 없으며, OS 수준에서 파일 시스템 확장하여 OS가 추가 공간을 인식하고 활용하게 해야 합니다.

즉, [*EBS 볼륨을 원하는 파일 시스템으로 포맷한 다음 마운트*해야](https://docs.aws.amazon.com/ko_kr/ebs/latest/userguide/ebs-using-volumes.html)합니다.

예를 들어, EBS 볼륨이 하나의 땅이라면, 파일 시스템을 그 땅 위에 있는 집입니다.
EBS 용량을 늘리는 것은 땅의 경계를 확장하는 것과 같습니다.
하지만 실제로 집(파일 시스템)을 확장하여 새로운 땅을 덮을 때까지는 여분의 공간을 사용할 수 없습니다.

EBS 볼륨을 늘렸지만 파일 시스템 확장하지 않은 경우:
- 운영 체제에서 여전히 이전 크기가 표시됩니다.
- 추가 공간은 할당되지 않고 사용할 수 없는 상태로 유지됩니다.
- 애플리케이션이 새 공간에 데이터를 쓸 수 없습니다.

### EBS 볼륨, 파티션, 그리고 파일 시스템 확장

1. Extending the EBS volume
2. Extending the partition
3. Extending the file system

These steps are required because they deal with different layers of storage abstraction:

1. EBS 볼륨 확장:
   - AWS 인프라 수준에서 수행됩니다.
   - 볼륨의 원시 스토리지 용량이 증가합니다.
   - 그러나 운영 체제는 이 변경 사항을 자동으로 인식하지 못합니다.

2. 파티션 확장
   - 파티션 테이블은 디스크가 분할되는 방식을 정의합니다.
   - *EBS 볼륨이 확장된 후에도 파티션은 여전히 이전 크기를 반영*합니다.
   - `growpart`는 새로운 사용 가능한 공간을 인식하도록 파티션 테이블을 업데이트합니다.
   - 파일 시스템은 원시 장치(여기서 EBS 볼륨)에 직접 생성되는 것이 아니라 파티션에 생성되기 때문에 이러한 단계가 필요합니다.

3. 파일 시스템 확장자:
   - 실제로 운영 체제가 새 공간을 사용할 수 있도록 하는 단계입니다.
   - *파티션이 확장된 후에도 파일 시스템 자체는 내부적으로 이전 크기가 그대로 기록*됩니다.
   - `resize2fs`(`ext4` 경우) 또는 `xfs_growfs`(`XFS` 경우)와 같은 도구는 새로 사용 가능한 공간을 사용하도록 파일 시스템의 내부 레코드를 업데이트합니다.

각 단계는 서로 다른 추상화 수준(AWS 인프라, OS 파티션 관리, 파일 시스템)에서 작동합니다.
이러한 분리를 통해 다양한 시스템과 구성에서 유연성과 호환성을 확보할 수 있습니다.

그리고 시스템마다 다른 파티셔닝 체계나 파일 시스템을 사용할 수 있는데,
이러한 다단계 프로세스를 통해 다양한 구성에서 호환성을 보장할 수 있습니다.

경우에 따라 볼륨을 확장하고 싶지만 단일 파티션이나 파일 시스템의 모든 공간을 즉시 사용하지 않을 수도 있으므로,
이러한 유연성을 고려한 구조라고 합니다.

```plaintext
[EBS Volume] -> [Partition] -> [File System] -> [Usable Space]
```

### 1. [볼륨 수정](https://docs.aws.amazon.com/ko_kr/ebs/latest/userguide/requesting-ebs-volume-modifications.html)

수정 전 기록:

```sh
$ lsblk
NAME     MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
loop0      7:0    0 25.2M  1 loop /snap/amazon-ssm-agent/7993
loop1      7:1    0 55.7M  1 loop /snap/core18/2829
loop2      7:2    0 38.8M  1 loop /snap/snapd/21759
xvda     202:0    0    8G  0 disk
├─xvda1  202:1    0    7G  0 part /
├─xvda14 202:14   0    4M  0 part
├─xvda15 202:15   0  106M  0 part /boot/efi
└─xvda16 259:0    0  913M  0 part /boot
```

볼륨 수정이 성공했으며 [`optimizing` 또는 `completed` 상태](https://docs.aws.amazon.com/ko_kr/ebs/latest/userguide/monitoring-volume-modifications.html)인지 확인합니다.

`optimizing(0%)`에서 요지부동이어서 [검색](https://repost.aws/knowledge-center/ebs-volume-stuck-optimizing-on-modification)해보니, 몇 분에서 몇 시간 소요될 수 있다고 합니다. 기다리면 `completed` 상태가 됩니다.

> Modifying an EBS volume can take from a few minutes to a few hours depending on the configuration changes being applied.

수정 후:

```sh
$ sudo lsblk
NAME     MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
loop0      7:0    0 25.2M  1 loop /snap/amazon-ssm-agent/7993
loop1      7:1    0 55.7M  1 loop /snap/core18/2829
loop2      7:2    0 38.8M  1 loop /snap/snapd/21759
xvda     202:0    0   20G  0 disk
                    ^^^^^ 8G에서 20G로 증가했습니다.
├─xvda1  202:1    0    7G  0 part /
├─xvda14 202:14   0    4M  0 part
├─xvda15 202:15   0  106M  0 part /boot/efi
└─xvda16 259:0    0  913M  0 part /boot
```

### 2. [파일시스템 확장](https://docs.aws.amazon.com/ko_kr/ebs/latest/userguide/recognize-expanded-volume-linux.html?icmpid=docs_ec2_console)

#### 2.1. 파티션 크기를 조정

볼륨 크기가 추가됐다면, 파티션 크기를 조정합니다. 진행에 앞서 현재 파티션 정보를 확인합니다.

```sh
$ sudo parted /dev/xvda print
Warning: Not all of the space available to /dev/xvda appears to be used, you can fix the GPT to use all of the space (an extra 25165824 blocks) or continue with the current setting?
Fix/Ignore? ^C
Model: Xen Virtual Block Device (xvd)
Disk /dev/xvda: 21.5GB
Sector size (logical/physical): 512B/512B
Partition Table: gpt
Disk Flags:

Number  Start   End     Size    File system  Name  Flags
14      1049kB  5243kB  4194kB                     bios_grub
15      5243kB  116MB   111MB   fat32              boot, esp
16      116MB   1074MB  957MB   ext4               bls_boot
 1      1075MB  8590MB  7515MB  ext4
```

> > Warning: Not all of the space available to /dev/xvda appears to be used, you can fix the GPT to use all of the space (an extra 25165824 blocks) or continue with the current setting?
>
> 이 메시지는 GPT가 실제 디스크의 전체 크기를 반영하지 않는다는 것을 나타냅니다.
> EBS 볼륨의 크기를 늘린 경우처럼 디스크가 확장되었지만 파티션 테이블이 이 변경 사항을 반영하도록 업데이트되지 않은 경우에 자주 발생합니다.
>
> 여기서 "수정"한다는 것은 디스크의 전체 크기를 인식하도록 파티션 테이블을 업데이트하겠다는 뜻입니다.
> 이렇게 해도 **실제로 데이터나 파티션 크기가 변경되는 것은 아니며**, 디스크 레이아웃을 설명하는 메타데이터만 업데이트됩니다.
>
> > **[GPT(GUID Partition Table)](https://www.thomas-krenn.com/en/wiki/GUID_Partition_Table_information)**:
> >
> > GPT는 물리적 저장 장치의 파티션 테이블 레이아웃을 위한 표준입니다.
> > 이는 UEFI(Unified Extensible Firmware Interface) 표준의 일부이며 과거 MBR(Master Boot Record) 파티셔닝 체계의 후속 버전입니다.
> > GPT는 MBR보다 더 큰 디스크 크기와 더 많은 파티션을 허용합니다.
> >
> > GPT 메타데이터는 파일로 저장되어 있지 않고 디스크의 특정 섹터에 직접 저장되어 있습니다.
> > `fdisk` 또는 `gdisk` 명령어를 사용하여 메타데이터를 조회할 수 있습니다.
> >
> > ```sh
> > $ sudo fdisk -l /dev/xvda
> > GPT PMBR size mismatch (16777215 != 41943039) will be corrected by write.
> > The backup GPT table is not on the end of the device.
> > Disk /dev/xvda: 20 GiB, 21474836480 bytes, 41943040 sectors
> > Units: sectors of 1 * 512 = 512 bytes
> > Sector size (logical/physical): 512 bytes / 512 bytes
> > I/O size (minimum/optimal): 512 bytes / 512 bytes
> > Disklabel type: gpt
> > Disk identifier: <Disk identifier>
> >
> > Device        Start      End  Sectors  Size Type
> > /dev/xvda1  2099200 16777182 14677983    7G Linux filesystem
> > /dev/xvda14    2048    10239     8192    4M BIOS boot
> > /dev/xvda15   10240   227327   217088  106M EFI System
> > /dev/xvda16  227328  2097152  1869825  913M Linux extended boot
> >
> > Partition table entries are not in disk order.
> > ```

파티션 크기가 볼륨 크기보다 작은 경우 `growpart` 명령을 사용하여 확장할 파티션을 지정합니다.

```sh
# /dev/xvda1 파티션을 확장하려면 다음 명령을 사용합니다.
$ sudo growpart /dev/xvda 1
CHANGED: partition=1 start=2099200 old: size=14677983 end=16777182 new: size=39843807 end=41943006
```

다시 `lsblk`로 확인하면 1번 파티션에 반영되었음을 확인할 수 있습니다.

```sh
$ lsblk
NAME     MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
loop0      7:0    0 25.2M  1 loop /snap/amazon-ssm-agent/7993
loop1      7:1    0 55.7M  1 loop /snap/core18/2829
loop2      7:2    0 38.8M  1 loop /snap/snapd/21759
xvda     202:0    0   20G  0 disk
├─xvda1  202:1    0   19G  0 part /
├─xvda14 202:14   0    4M  0 part
├─xvda15 202:15   0  106M  0 part /boot/efi
└─xvda16 259:0    0  913M  0 part /boot
```

#### 2.2. 파일 시스템을 확장

`df -hT`를 사용하여 확장해야 하는 파일 시스템의 이름, 크기, 유형 및 탑재 지점을 가져옵니다.
변경하려는 파일 시스템 `/dev/root`는 Xen instance `ext4` 타입입니다.

```sh
$ df -hT
Filesystem     Type   Size  Used Avail Use% Mounted on
/dev/root      ext4   6.8G  3.3G  3.5G  49% /
tmpfs          tmpfs  2.0G     0  2.0G   0% /dev/shm
tmpfs          tmpfs  783M  900K  782M   1% /run
tmpfs          tmpfs  5.0M     0  5.0M   0% /run/lock
/dev/xvda16    ext4   881M  133M  687M  17% /boot
/dev/xvda15    vfat   105M  6.1M   99M   6% /boot/efi
tmpfs          tmpfs  392M   16K  392M   1% /run/user/1000
```

파일 시스템을 확장하는 명령은 파일 시스템 유형에 따라 다르므로, 파일 시스템 유형에 따라 올바른 명령을 선택합니다.

`ext4` 타입 파일 시스템은 `resize2fs`를 사용합니다.

```sh
$ sudo resize2fs /dev/xvda1
resize2fs 1.47.0 (5-Feb-2023)
Filesystem at /dev/xvda1 is mounted on /; on-line resizing required
old_desc_blocks = 1, new_desc_blocks = 3
The filesystem on /dev/xvda1 is now 4980475 (4k) blocks long.
```

이제 추가된 볼륨이 파일 시스템의 용량으로 추가됐음을 확인할 수 있습니다.

```sh
$ df -hT
Filesystem     Type   Size  Used Avail Use% Mounted on
/dev/root      ext4    19G  3.3G   16G  18% /
tmpfs          tmpfs  2.0G     0  2.0G   0% /dev/shm
tmpfs          tmpfs  783M  904K  782M   1% /run
tmpfs          tmpfs  5.0M     0  5.0M   0% /run/lock
/dev/xvda16    ext4   881M  133M  687M  17% /boot
/dev/xvda15    vfat   105M  6.1M   99M   6% /boot/efi
tmpfs          tmpfs  392M   16K  392M   1% /run/user/1000
```

## 기타

- [Amazon Elastic Block Store란 무엇인가요?](https://docs.aws.amazon.com/ko_kr/ebs/latest/userguide/what-is-ebs.html)
- [[AWS] 📚 EBS 개념 & 사용법 💯 정리 (EBS Volume 추가하기)](https://inpa.tistory.com/entry/AWS-📚-EBS-개념-사용법-💯-정리-EBS-Volume-추가하기)
- [[AWS] EC2 인스턴스 용량 확장](https://velog.io/@harvey/AWS-EC2-%EC%9D%B8%EC%8A%A4%ED%84%B4%EC%8A%A4-%EC%9A%A9%EB%9F%89-%ED%99%95%EC%9E%A5)
- [Amazon EC2 인스턴스의 루트 볼륨을 중지하지 않고 교체](https://docs.aws.amazon.com/ko_kr/AWSEC2/latest/UserGuide/replace-root.html)
- [Amazon EBS 탄력적 볼륨을 사용하여 볼륨 수정](https://docs.aws.amazon.com/ko_kr/ebs/latest/userguide/ebs-modify-volume.html)
- [How do I extend my Linux file system after I increase my EBS volume on my EC2 instance?](https://repost.aws/knowledge-center/extend-linux-file-system)
