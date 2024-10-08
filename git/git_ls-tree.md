# git ls-tree

## ls-tree 설명

```sh
git ls-tree 
    [-d] # Show only the named tree entry itself, not its children
    [-r] # Recurse into sub-trees.
    [-t] # Show tree entries even when going to recurse them. 
         # Has no effect if `-r` was not passed. `-d` implies `-t`.
    [-l] # Show object size of blob (file) entries.
    [-z] # \0 line termination on output and do not quote filenames. 
         # See OUTPUT FORMAT below for more information.
    [--name-only] # equivalent to specifying --format='%(objectname)'
    [--name-status] 
    [--object-only] 
    [--full-name] 
    [--full-tree] 
    [--abbrev[=<n>]] 
    [--format=<format>]
    <tree-ish> 
    [<path>...]
```

Lists the contents of a given tree object, like what `/bin/ls -a` does in the current working directory. Note that:

- the behaviour is slightly different from that of "/bin/ls" in that the <path> denotes just a list of patterns to match, e.g. so
    specifying directory name (without -r) will behave differently, and order of the arguments does not matter.

- the behaviour is similar to that of "/bin/ls" in that the <path> is taken as relative to the current working directory.
    E.g. when you are in a directory sub that has a directory dir, you can run git ls-tree -r HEAD dir to list the contents of the tree (that is sub/dir
    in HEAD). You don’t want to give a tree that is not at the root level (e.g.  git ls-tree -r HEAD:sub dir) in this case, as that would
    result in asking for sub/sub/dir in the HEAD commit. However, the current working directory can be ignored by passing --full-tree
    option.

## format

- 기본 포맷

    ```sh
    %(objectmode) %(objecttype) %(objectname)%x09%(path)
    ```
