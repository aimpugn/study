# vm_stat

## description

> show Mach virtual memory statistics

```bash
vm_stat [[-c count] interval]
```

`vm_stat` displays Mach virtual memory statistics.  

If the optional `interval` is specified, then `vm_stat` will display the statistics every `interval` seconds.  

In this case, each line of output displays the change in each statistic (an `interval` count of 1 displays the values per second).  

However, the first line of output following each banner displays the system-wide totals for each statistic.  If a `count` is provided, the command will terminate after `count` intervals.  

## displayed values

     Pages free
             the total number of free pages in the system.

     Pages active
             the total number of pages currently in use and pageable.

     Pages inactive
             the total number of pages on the inactive list.

     Pages speculative
             the total number of pages on the speculative list.

     Pages throttled
             the total number of pages on the throttled list (not wired but
             not pageable).

     Pages wired down
             the total number of pages wired down.  That is, pages that cannot
             be paged out.

     Pages purgeable
             the total number of purgeable pages.

     Translation faults
             the number of times the "vm_fault" routine has been called.

     Pages copy-on-write
             the number of faults that caused a page to be copied (generally
             caused by copy-on-write faults).

     Pages zero filled
             the total number of pages that have been zero-filled on demand.

     Pages reactivated
             the total number of pages that have been moved from the inactive
             list to the active list (reactivated).

     Pages purged
             the total number of pages that have been purged.

     File-backed pages
             the total number of pages that are file-backed (non-swap)

     Anonymous pages
             the total number of pages that are anonymous

     Uncompressed pages
             the total number of pages (uncompressed) held within the
             compressor

     Pages used by VM compressor:
             the number of pages used to store compressed VM pages.

     Pages decompressed
             the total number of pages that have been decompressed by the VM
             compressor.

     Pages compressed
             the total number of pages that have been compressed by the VM
             compressor.

     Pageins
             the total number of requests for pages from a pager (such as the
             inode pager).

     Pageouts
             the total number of pages that have been paged out.

     Swapins
             the total number of compressed pages that have been swapped out
             to disk.

     Swapouts
             the total number of compressed pages that have been swapped back
             in from disk.

     If `interval` is not specified, then `vm_stat` displays all accumulated
     statistics along with the page size.

macOS                           August 13, 1997                          macOS

## examples

```log
❯ vm_stat
Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                                2568.
Pages active:                            228881.
Pages inactive:                          226655.
Pages speculative:                          302.
Pages throttled:                              0.
Pages wired down:                        131951.
Pages purgeable:                           1703.
"Translation faults":                3082430124.
Pages copy-on-write:                   73861952.
Pages zero filled:                    639453098.
Pages reactivated:                   1030927185.
Pages purged:                          54213219.
File-backed pages:                       159405.
Anonymous pages:                         296433.
Pages stored in compressor:             1475064.
Pages occupied by compressor:            417096.
Decompressions:                      1189979911.
Compressions:                        1259322627.
Pageins:                               44127772.
Pageouts:                               1241536.
Swapins:                                4114064.
Swapouts:                               6446877.
```

- **Pages free**: 현재 사용 가능한 페이지 수입니다. 이 경우 2,568페이지입니다.
- **Pages active/inactive/speculative**: 현재 사용 중이거나 곧 사용될 것으로 예상되는 페이지 수입니다.
- **Pages wired down**: 시스템에 의해 고정되어 해제될 수 없는 페이지 수입니다.
- **Pages purgeable**: 시스템에 의해 쉽게 제거될 수 있는 페이지 수입니다.
- **"Translation faults"**: 가상 주소를 실제 주소로 변환할 때 발생한 오류 횟수입니다.
- **File-backed pages**: 파일에 의해 백업되는 페이지 수입니다.
- **Anonymous pages**: 이름이 없는 페이지(주로 힙이나 스택에 사용)의 수입니다.
- **Pages stored in compressor**: 압축을 통해 저장된 페이지 수입니다.
- **Pageins/Pageouts**: 디스크에서 메모리로 가져온 페이지 수와 메모리에서 디스크로 옮겨진 페이지 수입니다.
- **Swapins/Swapouts**: 스왑 영역(디스크의 일부)에서 메모리로 가져온 페이지 수와 메모리에서 스왑 영역으로 옮겨진 페이지 수입니다.
