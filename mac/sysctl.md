# sysctl

- [sysctl](#sysctl)
    - [논리적 프로세서(Core) 수 확인](#논리적-프로세서core-수-확인)
    - [물리적 프로세서(Core) 수를 확인](#물리적-프로세서core-수를-확인)
    - [최대 사용 가능한 프로세서의 수를 확인](#최대-사용-가능한-프로세서의-수를-확인)
    - [시스템의 전체 CPU 정보를 확인](#시스템의-전체-cpu-정보를-확인)
    - [cpu만 보기](#cpu만-보기)

## 논리적 프로세서(Core) 수 확인

```shell
sysctl -n hw.logicalcpu
```

## 물리적 프로세서(Core) 수를 확인

```shell
sysctl -n hw.physicalcpu
```

## 최대 사용 가능한 프로세서의 수를 확인

```shell
sysctl -n hw.ncpusysctl -n hw.ncpu
```

## 시스템의 전체 CPU 정보를 확인

```shell
sysctl hw.cpu
```

## cpu만 보기

```shell
╰─ sysctl hw | rg cpu
hw.ncpu: 10
hw.activecpu: 10
hw.perflevel0.physicalcpu: 8
hw.perflevel0.physicalcpu_max: 8
hw.perflevel0.logicalcpu: 8
hw.perflevel0.logicalcpu_max: 8
hw.perflevel0.cpusperl2: 4
hw.perflevel1.physicalcpu: 2
hw.perflevel1.physicalcpu_max: 2
hw.perflevel1.logicalcpu: 2
hw.perflevel1.logicalcpu_max: 2
hw.perflevel1.cpusperl2: 2
hw.physicalcpu: 10
hw.physicalcpu_max: 10
hw.logicalcpu: 10
hw.logicalcpu_max: 10
hw.cputype: 16777228
hw.cpusubtype: 2
hw.cpu64bit_capable: 1
hw.cpufamily: 458787763
hw.cpusubfamily: 4
```
