# git rev-parse

- [git rev-parse](#git-rev-parse)
    - [현재 디렉토리가 git 디렉토리인지](#현재-디렉토리가-git-디렉토리인지)
    - [현재 디렉토리가 git submodule인지](#현재-디렉토리가-git-submodule인지)
    - [upstream 확인하기](#upstream-확인하기)
    - [top level 디렉토리 확인하기](#top-level-디렉토리-확인하기)
    - [정리중](#정리중)

## 현재 디렉토리가 git 디렉토리인지

```shell
# https://stackoverflow.com/a/16925062
git rev-parse --is-inside-work-tree
```

```shell
is_git_project() {
  local result
  result="$(git rev-parse --is-inside-work-tree 2>&1)" # STDERR to STDOUT
  if [ "${result// }" = "true" ]; then
    echo "true"
  else
    echo "false"
  fi
}
```

## 현재 디렉토리가 git submodule인지

```shell
is_submodule() {
  local work_tree
  work_tree=$(git rev-parse --show-superproject-working-tree)
  if [ -n "${work_tree// }" ]; then
    echo "true"
  else
    echo "false"
  fi

  return  0
}
```

## upstream 확인하기

```bash
# https://stackoverflow.com/a/9753364
git rev-parse --abbrev-ref --symbolic-full-name @{u}
```

## top level 디렉토리 확인하기

> Show the (by default, absolute) path of the top-level directory of the working tree. If there is no working tree, report an error.

```bash
git rev-parse --show-toplevel
```

## 정리중

--path-format=(absolute|relative)
    Controls the behavior of certain other options. If specified as absolute, the paths printed by those options will be
    absolute and canonical. If specified as relative, the paths will be relative to the current working directory if that is
    possible. The default is option specific.

    This option may be specified multiple times and affects only the arguments that follow it on the command line, either to
    the end of the command line or the next instance of this option.

The following options are modified by --path-format:

--git-dir
    Show $GIT_DIR if defined. Otherwise show the path to the .git directory. The path shown, when relative, is relative to
    the current working directory.

    If $GIT_DIR is not defined and the current directory is not detected to lie in a Git repository or work tree print a
    message to stderr and exit with nonzero status.

--git-common-dir
    Show $GIT_COMMON_DIR if defined, else $GIT_DIR.

--resolve-git-dir <path>
    Check if <path> is a valid repository or a gitfile that points at a valid repository, and print the location of the
    repository. If <path> is a gitfile then the resolved path to the real repository is printed.

--git-path <path>
    Resolve "$GIT_DIR/<path>" and takes other path relocation variables such as $GIT_OBJECT_DIRECTORY, $GIT_INDEX_FILE...
    into account. For example, if $GIT_OBJECT_DIRECTORY is set to /foo/bar then "git rev-parse --git-path objects/abc"
    returns /foo/bar/abc.

--show-toplevel
    Show the (by default, absolute) path of the top-level directory of the working tree. If there is no working tree, report
    an error.

--show-superproject-working-tree
    Show the absolute path of the root of the superproject’s working tree (if exists) that uses the current repository as
    its submodule. Outputs nothing if the current repository is not used as a submodule by any project.

--shared-index-path
    Show the path to the shared index file in split index mode, or empty if not in split-index mode.

The following options are unaffected by --path-format:

--absolute-git-dir
    Like --git-dir, but its output is always the canonicalized absolute path.

--is-inside-git-dir
    When the current working directory is below the repository directory print "true", otherwise "false".

--is-inside-work-tree
    When the current working directory is inside the work tree of the repository print "true", otherwise "false".

--is-bare-repository
    When the repository is bare print "true", otherwise "false".

--is-shallow-repository
    When the repository is shallow print "true", otherwise "false".

--show-cdup
    When the command is invoked from a subdirectory, show the path of the top-level directory relative to the current
    directory (typically a sequence of "../", or an empty string).

--show-prefix
    When the command is invoked from a subdirectory, show the path of the current directory relative to the top-level
    directory.

--show-object-format[=(storage|input|output)]
    Show the object format (hash algorithm) used for the repository for storage inside the .git directory, input, or output.
    For input, multiple algorithms may be printed, space-separated. If not specified, the default is "storage".
