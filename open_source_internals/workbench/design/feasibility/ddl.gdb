set pagination off
set confirm off
set breakpoint pending off
handle SIGUSR1 SIGUSR2 SIGURG SIGALRM nostop noprint pass
break heap_insert
commands
silent
printf "HEAP_INSERT %s\n", relation->rd_rel->relname.data
continue
end
break smgrcreate
commands
silent
printf "SMGRCREATE relNumber=%u fork=%d\n", reln->smgr_rlocator.locator.relNumber, forknum
continue
end
break XLogInsert
commands
silent
continue
end
break XLogFlush
commands
silent
printf "XLOGFLUSH\n"
continue
end
break tuplesort_begin_index_btree
commands
silent
printf "TUPLESORT workMem=%d\n", workMem
continue
end
break btbuild
commands
silent
printf "BTBUILD index=%s heap=%s\n", index->rd_rel->relname.data, heap->rd_rel->relname.data
continue
end
break smgr_bulk_write
commands
silent
printf "BULK_WRITE block=%u\n", blocknum
continue
end
break heap_getnextslot
commands
silent
continue
end
break ReadyForQuery
commands
silent
info breakpoints
detach
quit
end
shell touch /tmp/ready
continue
