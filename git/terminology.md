# Git Terminology

- [Git Terminology](#git-terminology)
    - [`fast-forward`(`ff`)](#fast-forwardff)
    - [`source branch`](#source-branch)
    - [스테이징](#스테이징)
        - [Git에서 Staging이란?](#git에서-staging이란)
        - [Staging의 목적](#staging의-목적)
        - [Staging의 단계](#staging의-단계)

## `fast-forward`(`ff`)

`fast-forward`? to describe any operation that can be performed quickly and easily, without requiring complex merge operations or manual conflict resolution

"Fast-forward" is a term used to describe a type of Git merge operation that occurs when the target branch (in this case, `develop`) has no new commits since the source branch (in this case, `temporary-branch`) was created. In other words, there is a linear path of commits from the develop branch to the temporary-branch.

In a fast-forward merge, Git simply moves the pointer of the target branch to point to the head of the source branch, effectively integrating the changes from the source branch into the target branch. This operation is "fast" because Git does not need to perform any complex merge operations or create a new merge commit. It is "forward" because the branch pointer is moving in the direction of the new changes.

Fast-forward merges are typically used to integrate changes from a feature branch into a main branch (such as develop or master) when the feature branch has been kept up-to-date with the main branch, and there are no conflicting changes. This allows for a clean and simple integration of changes, without the need for a new merge commit.

In contrast, if there are conflicting changes between the source and target branches, Git will perform a regular merge, which creates a new merge commit to reconcile the changes. This can be more complex, and requires manual resolution of conflicts.

Fast-forward merges are important because they allow for a clean and simple integration of changes, without the need for a new merge commit. This can help keep the Git history clean and easy to read, and can simplify the process of integrating changes from different branches. However, it's important to note that fast-forward merges are not always possible or appropriate, and should be used with care.

## `source branch`

In Git, the `source branch` refers to the branch that contains the changes you want to integrate into another branch.

When you perform a Git operation like a *merge* or a *rebase*, you typically specify the `source branch` and the **target branch**, where the target branch is the branch you want to integrate the changes into.

For example, let's say you have a Git repository with two branches: *develop* and *feature*.

1. Working on *feature* branch, and integrate into the *develop* branch
2. In this case, the *feature* branch would be the `source branch`, and the *develop* branch would be the target branch.

When you perform a merge or a rebase operation to integrate changes from the source branch into the target branch, Git will apply the changes from the source branch onto the target branch, either by creating a new merge commit or by rewriting the history of the source branch on top of the target branch.

In general, it's important to keep the source branch up-to-date with the target branch, to ensure that the changes can be integrated cleanly and with minimal conflicts. This can be done using Git operations like git merge or git rebase, or through other workflows like pull requests on Git hosting services like GitHub.

## 스테이징

### Git에서 Staging이란?

Staging은 Git의 중요한 개념 중 하나로, 파일 변경 사항을 준비하는 단계입니다.
변경된 파일들을 커밋(commit)하기 전에 스테이징(staging) 영역에 추가해야 합니다.
이 영역은 Index 또는 Cache라고도 불립니다.

### Staging의 목적

- 커밋 준비: 파일 변경 사항을 선택적으로 커밋할 수 있도록 준비하는 단계입니다.
- 부분 커밋: 일부 파일만 선택적으로 커밋할 수 있습니다.
- 변경 사항 검토: 커밋하기 전에 어떤 변경 사항이 있는지 검토할 수 있습니다.

### Staging의 단계

1. Working Directory:

    실제 파일들이 존재하고 작업이 이루어지는 디렉토리입니다.
    여기서 파일을 수정하거나 새로운 파일을 추가하거나 삭제할 수 있습니다.

2. Staging Area:

    커밋할 변경 사항을 임시로 저장하는 영역입니다.
    파일을 스테이징하면, 해당 파일의 현재 상태가 스냅샷으로 저장됩니다.
    이는 커밋할 준비가 된 파일들만 포함됩니다.

3. Repository:

    커밋된 변경 사항이 저장되는 곳입니다.
    커밋하면 스테이징된 모든 변경 사항이 로컬 리포지토리에 저장됩니다.
