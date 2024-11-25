# IntelliJ

- [IntelliJ](#intellij)
    - [Cannot find declaration to go to](#cannot-find-declaration-to-go-to)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [Mouse drag selection behaving oddly in IntelliJ IDEA on macOS](#mouse-drag-selection-behaving-oddly-in-intellij-idea-on-macos)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)
    - [unnecessary\_acquire(Node const\*)+0x24](#unnecessary_acquirenode-const0x24)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-2)
    - [IntelliJ IDEA has failed to load the environment from '/bin/zsh'](#intellij-idea-has-failed-to-load-the-environment-from-binzsh)
        - [문제](#문제-3)
        - [원인](#원인-3)
        - [해결](#해결-3)
    - [Receiver class com.aurimasniekis.phppsr4namespacedetector.PhpSourceDirectoryConfigurator does not define or inherit an implementation of the resolved method](#receiver-class-comaurimasniekisphppsr4namespacedetectorphpsourcedirectoryconfigurator-does-not-define-or-inherit-an-implementation-of-the-resolved-method)
        - [문제](#문제-4)
        - [원인](#원인-4)
    - [There was a problem with the editor](#there-was-a-problem-with-the-editor)
        - [문제](#문제-5)
        - [원인](#원인-5)
        - [해결](#해결-4)
    - [IntelliJ 터미널의 PHP 버전과 iTERM2에서의 PHP 버전이 상이함](#intellij-터미널의-php-버전과-iterm2에서의-php-버전이-상이함)
        - [문제](#문제-6)
        - [원인](#원인-6)
        - [해결](#해결-5)
    - [터미널에서 amazon q cli 작동하지 않음](#터미널에서-amazon-q-cli-작동하지-않음)
        - [문제](#문제-7)
        - [원인](#원인-7)
        - [해결](#해결-6)
    - [The plugin com.jetbrains.php failed to save settings](#the-plugin-comjetbrainsphp-failed-to-save-settings)
        - [문제](#문제-8)
        - [원인](#원인-8)
        - [해결](#해결-7)
    - [PHPdoc 작성시 하이라이트가 깜막이며 타이핑 딜레이](#phpdoc-작성시-하이라이트가-깜막이며-타이핑-딜레이)
        - [문제](#문제-9)
        - [원인](#원인-9)
        - [해결](#해결-8)

## Cannot find declaration to go to

### 문제

- 파일 오염이 발생했다는 에러가 발생하며 갑자기 코드 네비게이션이 제대로 작동하지 않음

```log
2022-11-29 16:25:32,977 [107239630] SEVERE - #c.i.i.p.PluginManager - com.intellij.util.io.PersistentEnumeratorBase$CorruptedException: PersistentEnumerator storage corrupted /Users/rody/Library/Caches/JetBrains/IntelliJIdea2022.2/caches/contentHashes.dat
java.lang.RuntimeException: com.intellij.util.io.PersistentEnumeratorBase$CorruptedException: PersistentEnumerator storage corrupted /Users/rody/Library/Caches/JetBrains/IntelliJIdea2022.2/caches/contentHashes.dat
 at com.intellij.util.ExceptionUtil.rethrow(ExceptionUtil.java:132)
 at com.intellij.openapi.vfs.newvfs.persistent.PersistentFSConnection.handleError(PersistentFSConnection.java:311)
 at com.intellij.openapi.vfs.newvfs.persistent.FSRecords.handleError(FSRecords.java:812)
 at com.intellij.openapi.vfs.newvfs.persistent.FSRecords.writeAndHandleErrors(FSRecords.java:323)
 at com.intellij.openapi.vfs.newvfs.persistent.FSRecords$1.close(FSRecords.java:756)
 at com.intellij.openapi.vfs.newvfs.persistent.PersistentFSImpl$3.close(PersistentFSImpl.java:793)
 at com.intellij.openapi.fileEditor.impl.LoadTextUtil.write(LoadTextUtil.java:450)
 at com.intellij.openapi.fileEditor.impl.FileDocumentManagerImpl.lambda$doSaveDocumentInWriteAction$4(FileDocumentManagerImpl.java:417)
 at com.intellij.pom.core.impl.PomModelImpl.guardPsiModificationsIn(PomModelImpl.java:326)
 at com.intellij.openapi.fileEditor.impl.FileDocumentManagerImpl.doSaveDocumentInWriteAction(FileDocumentManagerImpl.java:406)
 at com.intellij.openapi.fileEditor.impl.FileDocumentManagerImpl.lambda$doSaveDocument$2(FileDocumentManagerImpl.java:367)
 at com.intellij.openapi.application.WriteAction.lambda$run$1(WriteAction.java:86)
 at com.intellij.openapi.application.impl.ApplicationImpl.runWriteActionWithClass(ApplicationImpl.java:1011)
 at com.intellij.openapi.application.impl.ApplicationImpl.runWriteAction(ApplicationImpl.java:1037)
 at com.intellij.openapi.application.WriteAction.run(WriteAction.java:85)
 at com.intellij.openapi.fileEditor.impl.FileDocumentManagerImpl.doSaveDocument(FileDocumentManagerImpl.java:367)
 at com.intellij.openapi.fileEditor.impl.FileDocumentManagerImpl.saveDocuments(FileDocumentManagerImpl.java:287)
 at com.intellij.openapi.fileEditor.impl.FileDocumentManagerImpl.saveAllDocuments(FileDocumentManagerImpl.java:262)
 at com.intellij.configurationStore.SaveAndSyncHandlerImpl$addListeners$3.onFrameDeactivated(SaveAndSyncHandlerImpl.kt:153)
 at com.intellij.util.messages.impl.MessageBusImplKt.invokeMethod(MessageBusImpl.kt:646)
 at com.intellij.util.messages.impl.MessageBusImplKt.invokeListener(MessageBusImpl.kt:625)
 at com.intellij.util.messages.impl.MessageBusImplKt.deliverMessage(MessageBusImpl.kt:399)
 at com.intellij.util.messages.impl.MessageBusImplKt.pumpWaiting(MessageBusImpl.kt:378)
 at com.intellij.util.messages.impl.MessageBusImplKt.access$pumpWaiting(MessageBusImpl.kt:1)
 at com.intellij.util.messages.impl.MessagePublisher.invoke(MessageBusImpl.kt:437)
 at jdk.proxy2/jdk.proxy2.$Proxy43.onFrameDeactivated(Unknown Source)
 at com.intellij.ide.FrameStateManagerAppListener.applicationDeactivated(FrameStateManagerAppListener.java:33)
 at com.intellij.util.messages.impl.MessageBusImplKt.invokeMethod(MessageBusImpl.kt:649)
 at com.intellij.util.messages.impl.MessageBusImplKt.invokeListener(MessageBusImpl.kt:625)
 at com.intellij.util.messages.impl.MessageBusImplKt.deliverMessage(MessageBusImpl.kt:399)
 at com.intellij.util.messages.impl.MessageBusImplKt.pumpWaiting(MessageBusImpl.kt:378)
 at com.intellij.util.messages.impl.MessageBusImplKt.access$pumpWaiting(MessageBusImpl.kt:1)
 at com.intellij.util.messages.impl.MessagePublisher.invoke(MessageBusImpl.kt:437)
 at jdk.proxy2/jdk.proxy2.$Proxy42.applicationDeactivated(Unknown Source)
 at com.intellij.ide.ApplicationActivationStateManager.updateState(ApplicationActivationStateManager.java:61)
 at com.intellij.ide.IdeEventQueue.processAppActivationEvent(IdeEventQueue.java:878)
 at com.intellij.ide.IdeEventQueue._dispatchEvent(IdeEventQueue.java:738)
 at com.intellij.ide.IdeEventQueue.lambda$dispatchEvent$6(IdeEventQueue.java:450)
 at com.intellij.openapi.progress.impl.CoreProgressManager.computePrioritized(CoreProgressManager.java:791)
 at com.intellij.ide.IdeEventQueue.lambda$dispatchEvent$7(IdeEventQueue.java:449)
 at com.intellij.openapi.application.TransactionGuardImpl.performActivity(TransactionGuardImpl.java:105)
 at com.intellij.ide.IdeEventQueue.performActivity(IdeEventQueue.java:624)
 at com.intellij.ide.IdeEventQueue.lambda$dispatchEvent$8(IdeEventQueue.java:447)
 at com.intellij.openapi.application.impl.ApplicationImpl.runIntendedWriteActionOnCurrentThread(ApplicationImpl.java:881)
 at com.intellij.ide.IdeEventQueue.dispatchEvent(IdeEventQueue.java:493)
 at java.desktop/java.awt.SentEvent.dispatch(SentEvent.java:75)
 at java.desktop/java.awt.DefaultKeyboardFocusManager$DefaultKeyboardFocusManagerSentEvent.dispatch(DefaultKeyboardFocusManager.java:262)
 at java.desktop/java.awt.DefaultKeyboardFocusManager.sendMessage(DefaultKeyboardFocusManager.java:289)
 at java.desktop/java.awt.DefaultKeyboardFocusManager.dispatchEvent(DefaultKeyboardFocusManager.java:833)
 at com.intellij.ide.IdeKeyboardFocusManager.lambda$dispatchEvent$0(IdeKeyboardFocusManager.java:48)
 at com.intellij.openapi.application.TransactionGuardImpl.performActivity(TransactionGuardImpl.java:113)
 at com.intellij.ide.IdeEventQueue.performActivity(IdeEventQueue.java:624)
 at com.intellij.ide.IdeKeyboardFocusManager.dispatchEvent(IdeKeyboardFocusManager.java:48)
 at java.desktop/java.awt.Component.dispatchEventImpl(Component.java:4903)
 at java.desktop/java.awt.Container.dispatchEventImpl(Container.java:2324)
 at java.desktop/java.awt.Window.dispatchEventImpl(Window.java:2802)
 at java.desktop/java.awt.Component.dispatchEvent(Component.java:4854)
 at java.desktop/sun.awt.SunToolkit$1.run(SunToolkit.java:516)
 at java.desktop/java.awt.event.InvocationEvent.dispatch(InvocationEvent.java:318)
 at java.desktop/java.awt.EventQueue.dispatchEventImpl(EventQueue.java:779)
 at java.desktop/java.awt.EventQueue$4.run(EventQueue.java:730)
 at java.desktop/java.awt.EventQueue$4.run(EventQueue.java:724)
 at java.base/java.security.AccessController.doPrivileged(AccessController.java:399)
 at java.base/java.security.ProtectionDomain$JavaSecurityAccessImpl.doIntersectionPrivilege(ProtectionDomain.java:86)
 at java.desktop/java.awt.EventQueue.dispatchEvent(EventQueue.java:749)
 at com.intellij.ide.IdeEventQueue.defaultDispatchEvent(IdeEventQueue.java:918)
 at com.intellij.ide.IdeEventQueue._dispatchEvent(IdeEventQueue.java:766)
 at com.intellij.ide.IdeEventQueue.lambda$dispatchEvent$6(IdeEventQueue.java:450)
 at com.intellij.openapi.progress.impl.CoreProgressManager.computePrioritized(CoreProgressManager.java:791)
 at com.intellij.ide.IdeEventQueue.lambda$dispatchEvent$7(IdeEventQueue.java:449)
 at com.intellij.openapi.application.TransactionGuardImpl.performActivity(TransactionGuardImpl.java:105)
 at com.intellij.ide.IdeEventQueue.performActivity(IdeEventQueue.java:624)
 at com.intellij.ide.IdeEventQueue.lambda$dispatchEvent$8(IdeEventQueue.java:447)
 at com.intellij.openapi.application.impl.ApplicationImpl.runIntendedWriteActionOnCurrentThread(ApplicationImpl.java:881)
 at com.intellij.ide.IdeEventQueue.dispatchEvent(IdeEventQueue.java:493)
 at java.desktop/java.awt.EventDispatchThread.pumpOneEventForFilters(EventDispatchThread.java:207)
 at java.desktop/java.awt.EventDispatchThread.pumpEventsForFilter(EventDispatchThread.java:128)
 at java.desktop/java.awt.EventDispatchThread.pumpEventsForHierarchy(EventDispatchThread.java:117)
 at java.desktop/java.awt.EventDispatchThread.pumpEvents(EventDispatchThread.java:113)
 at java.desktop/java.awt.EventDispatchThread.pumpEvents(EventDispatchThread.java:105)
 at java.desktop/java.awt.EventDispatchThread.run(EventDispatchThread.java:92)
Caused by: com.intellij.util.io.PersistentEnumeratorBase$CorruptedException: PersistentEnumerator storage corrupted /Users/rody/Library/Caches/JetBrains/IntelliJIdea2022.2/caches/contentHashes.dat
 at com.intellij.util.io.PersistentEnumeratorBase.catchCorruption(PersistentEnumeratorBase.java:618)
 at com.intellij.util.io.PersistentEnumeratorBase.doEnumerate(PersistentEnumeratorBase.java:272)
 at com.intellij.util.io.PersistentEnumeratorBase.enumerate(PersistentEnumeratorBase.java:285)
 at com.intellij.util.hash.ContentHashEnumerator.enumerate(ContentHashEnumerator.java:47)
 at com.intellij.openapi.vfs.newvfs.persistent.PersistentFSContentAccessor.findOrCreateContentRecord(PersistentFSContentAccessor.java:155)
 at com.intellij.openapi.vfs.newvfs.persistent.PersistentFSContentAccessor.writeContent(PersistentFSContentAccessor.java:76)
 at com.intellij.openapi.vfs.newvfs.persistent.PersistentFSContentAccessor$ContentOutputStream.close(PersistentFSContentAccessor.java:220)
 at java.base/java.io.FilterOutputStream.close(FilterOutputStream.java:188)
 at com.intellij.openapi.vfs.newvfs.persistent.FSRecords$1.lambda$close$0(FSRecords.java:757)
 at com.intellij.openapi.vfs.newvfs.persistent.FSRecords.writeAndHandleErrors(FSRecords.java:320)
 ... 77 more

```

### 원인

- 이유는 알 수 없었음

### 해결

- Invalidate Cache & restart -> 이것만으로는 제대로 안 됨
- 계속 인덱싱이 잘 되지 않아서 프로젝트 루트의 `.idea` 디렉토리를 통째로 삭제하고 다시 그래들 프로젝트 로드

## Mouse drag selection behaving oddly in IntelliJ IDEA on macOS

### 문제

- mac에서 IntelliJ를 쓰다보면 마우스 드래그가 이상하게 잡히는 현상이 있음
- 가령 더블 클릭해서 한 단어의 변수를 선택하거나 아니면 주욱 드래그 해서 일정 범위를 잡으려고 할 때 단어 선택이 안 되거나 원하는 범위가 아닌 이상한 범위를 잡는 이상한 현상이 있음

### 원인

Due to various factors such as
1. IDE settings,
2. system configurations,
3. or potentially bugs

### 해결

1. Restarting IntelliJ IDEA:
One user reported a strange selection issue that was resolved by closing and reopening IntelliJ IDEA​1​.

2. Block Selection Mode:
It might be possible that the Block Selection Mode is enabled accidentally. In IntelliJ IDEA, there's a feature called "Block Selection Mode" or "Column Selection Mode" that can be toggled on/off using the shortcut Alt + Shift + Insert on Windows/Linux or ⌘ + Shift + 8 on macOS. This mode allows you to select text block-wise instead of the usual line-wise selection. One user experienced a similar issue where dragging the mouse started doing block selects instead of the normal line select mode​1​.

3. Mouse or Trackpad Settings:
The issue might also be related to mouse or trackpad settings either within IntelliJ IDEA or at the system level on macOS.

4. [Keymap or Shortcut Conflicts](https://stackoverflow.com/questions/21136515/intellij-what-is-this-strange-selection-and-how-do-i-turn-it-off):
There could be conflicts with keymap settings or other shortcut keys. For example, one user had an issue with multiline-select not working on their iMac and it was related to the "Add or remove caret" action assigned to a specific shortcut​​.

5. [Text Over-selection Issue](https://intellij-support.jetbrains.com/hc/en-us/community/posts/360010393260-Click-drag-select-is-over-selecting-text-in-file-tab):
In another scenario reported on JetBrains support forum, a user experienced over-selection of text when using click-drag to select. The behavior was inconsistent and tended to over-select areas of text on the line​​.

6. Checking for IDE Updates or Bug Fixes:
Ensure that your IntelliJ IDEA is updated to the latest version as sometimes bugs related to UI or selection might be fixed in newer releases.

7. Custom IDE Settings or Plugins:
If you have custom settings or plugins installed, they might interfere with the normal operation of mouse selection. It might be worth checking if the issue persists with default settings or in a new/clean installation of IntelliJ IDEA.

Reporting the Issue to JetBrains:
If none of the above solutions work, it might be a good idea to report the issue to JetBrains with detailed information about the problem, your system configuration, and IntelliJ IDEA version.

## unnecessary_acquire(Node const*)+0x24

### 문제

```log
#
# A fatal error has been detected by the Java Runtime Environment:
#
#  SIGSEGV (0xb) at pc=0x0000000103904b00, pid=4408, tid=23555
#
# JRE version: OpenJDK Runtime Environment JBR-17.0.8.1+7-1000.32-jcef (17.0.8.1+7) (build 17.0.8.1+7-b1000.32)
# Java VM: OpenJDK 64-Bit Server VM JBR-17.0.8.1+7-1000.32-jcef (17.0.8.1+7-b1000.32, mixed mode, tiered, compressed oops, compressed class ptrs, shenandoah gc, bsd-aarch64)
# Problematic frame:
# V  [libjvm.dylib+0x4b00]  unnecessary_acquire(Node const*)+0x24
#
# No core dump will be written. Core dumps have been disabled. To enable core dumping, try "ulimit -c unlimited" before starting Java again
#
# If you would like to submit a bug report, please visit:
#   https://youtrack.jetbrains.com/issues/JBR
#

---------------  S U M M A R Y ------------

Command Line: 너무 길어서 생략

Host: "MacBookPro18,1" arm64, 10 cores, 16G, Darwin 23.1.0, macOS 14.1 (23B74)
Time: Fri Nov  3 14:55:03 2023 KST elapsed time: 47.832389 seconds (0d 0h 0m 47s)

---------------  T H R E A D  ---------------

Current thread (0x0000000105013000):  JavaThread "C2 CompilerThread0" daemon [_thread_in_native, id=23555, stack(0x0000000171570000,0x0000000171773000)]
```

### 원인

JVM이 해당 함수를 호출하는 동안 메모리 접근 위반으로 인해 충돌이 발생했음을 의미
`+0x24`는 함수의 시작 주소로부터 36바이트(16진수 24는 10진수로 36) 떨어진 지점에서 문제가 발생했음을 나타낸다

### 해결

## IntelliJ IDEA has failed to load the environment from '/bin/zsh'

### 문제

DataGrip을 실행했더니 경고 알람이 뜸

```log
IntelliJ IDEA has failed to load the environment from '/bin/zsh'
```

### 원인

모르겠다... [YouTrack 이슈](https://youtrack.jetbrains.com/issue/IDEA-274419)에서 [이 링크](https://youtrack.jetbrains.com/articles/IDEA-A-19/Shell-Environment-Loading)를 공유해주는데, 내용이 안 보임

### 해결

For me, the issue was that I load `tmux` automatically using `.zshrc`.

I followed the solution [here](https://youtrack.jetbrains.com/articles/IDEA-A-19/Shell-Environment-Loading)

Wrapped just the export `ZSH_TMUX_AUTOSTART=true` like this:

```shell
if [ -z "$INTELLIJ_ENVIRONMENT_READER" ]; then
  export ZSH_TMUX_AUTOSTART=true
fi
```

또는

> I solved it by opening Intellij using the terminal. You need to create Command Line Launcher if one doesn't exist.

1. Go to `Tools -> Create Command-line Launcher`
2. Then choose the location you prefer
3. Go to your terminal, and use that launcher command: `idea`

This will open Intellij and start your app, this should be able to access your system environment properties.

## Receiver class com.aurimasniekis.phppsr4namespacedetector.PhpSourceDirectoryConfigurator does not define or inherit an implementation of the resolved method

### 문제

```log
java.lang.AbstractMethodError: Receiver class com.aurimasniekis.phppsr4namespacedetector.PhpSourceDirectoryConfigurator does not define or inherit an implementation of the resolved method 'abstract void configureProject(com.intellij.openapi.project.Project, com.intellij.openapi.vfs.VirtualFile, com.intellij.openapi.util.Ref, boolean)' of interface com.intellij.platform.DirectoryProjectConfigurator.
    at com.intellij.platform.PlatformProjectOpenProcessor$Companion$runDirectoryProjectConfigurators$3$1.invoke(PlatformProjectOpenProcessor.kt:299)
    at com.intellij.platform.PlatformProjectOpenProcessor$Companion$runDirectoryProjectConfigurators$3$1.invoke(PlatformProjectOpenProcessor.kt:298)
    at com.intellij.openapi.progress.CoroutinesKt.blockingContext(coroutines.kt:248)
    at com.intellij.openapi.progress.CoroutinesKt.blockingContext(coroutines.kt:199)
    at com.intellij.platform.PlatformProjectOpenProcessor$Companion$runDirectoryProjectConfigurators$3.invokeSuspend(PlatformProjectOpenProcessor.kt:298)
    at kotlin.coroutines.jvm.internal.BaseContinuationImpl.resumeWith(ContinuationImpl.kt:33)
    at kotlinx.coroutines.DispatchedTask.run(DispatchedTask.kt:106)
    at com.intellij.openapi.application.impl.DispatchedRunnable.run(DispatchedRunnable.kt:43)
    at com.intellij.openapi.application.TransactionGuardImpl.runWithWritingAllowed(TransactionGuardImpl.java:208)
    at com.intellij.openapi.application.TransactionGuardImpl.access$100(TransactionGuardImpl.java:21)
    at com.intellij.openapi.application.TransactionGuardImpl$1.run(TransactionGuardImpl.java:190)
    at com.intellij.openapi.application.impl.ApplicationImpl.runIntendedWriteActionOnCurrentThread(ApplicationImpl.java:861)
    at com.intellij.openapi.application.impl.ApplicationImpl$4.run(ApplicationImpl.java:478)
    at com.intellij.openapi.application.impl.FlushQueue.doRun(FlushQueue.java:79)
    at com.intellij.openapi.application.impl.FlushQueue.runNextEvent(FlushQueue.java:121)
    at com.intellij.openapi.application.impl.FlushQueue.flushNow(FlushQueue.java:41)
    at java.desktop/java.awt.event.InvocationEvent.dispatch(InvocationEvent.java:318)
    at java.desktop/java.awt.EventQueue.dispatchEventImpl(EventQueue.java:792)
    at java.desktop/java.awt.EventQueue$3.run(EventQueue.java:739)
    at java.desktop/java.awt.EventQueue$3.run(EventQueue.java:733)
    at java.base/java.security.AccessController.doPrivileged(AccessController.java:399)
    at java.base/java.security.ProtectionDomain$JavaSecurityAccessImpl.doIntersectionPrivilege(ProtectionDomain.java:86)
    at java.desktop/java.awt.EventQueue.dispatchEvent(EventQueue.java:761)
    at com.intellij.ide.IdeEventQueue.defaultDispatchEvent(IdeEventQueue.kt:690)
    at com.intellij.ide.IdeEventQueue._dispatchEvent$lambda$10(IdeEventQueue.kt:593)
    at com.intellij.openapi.application.impl.ApplicationImpl.runWithoutImplicitRead(ApplicationImpl.java:1485)
    at com.intellij.ide.IdeEventQueue._dispatchEvent(IdeEventQueue.kt:593)
    at com.intellij.ide.IdeEventQueue.access$_dispatchEvent(IdeEventQueue.kt:67)
    at com.intellij.ide.IdeEventQueue$dispatchEvent$processEventRunnable$1$1$1.compute(IdeEventQueue.kt:369)
    at com.intellij.ide.IdeEventQueue$dispatchEvent$processEventRunnable$1$1$1.compute(IdeEventQueue.kt:368)
    at com.intellij.openapi.progress.impl.CoreProgressManager.computePrioritized(CoreProgressManager.java:787)
    at com.intellij.ide.IdeEventQueue$dispatchEvent$processEventRunnable$1$1.invoke(IdeEventQueue.kt:368)
    at com.intellij.ide.IdeEventQueue$dispatchEvent$processEventRunnable$1$1.invoke(IdeEventQueue.kt:363)
    at com.intellij.ide.IdeEventQueueKt.performActivity$lambda$1(IdeEventQueue.kt:997)
    at com.intellij.openapi.application.TransactionGuardImpl.performActivity(TransactionGuardImpl.java:105)
    at com.intellij.ide.IdeEventQueueKt.performActivity(IdeEventQueue.kt:997)
    at com.intellij.ide.IdeEventQueue.dispatchEvent$lambda$7(IdeEventQueue.kt:363)
    at com.intellij.openapi.application.impl.ApplicationImpl.runIntendedWriteActionOnCurrentThread(ApplicationImpl.java:861)
    at com.intellij.ide.IdeEventQueue.dispatchEvent(IdeEventQueue.kt:405)
    at java.desktop/java.awt.EventDispatchThread.pumpOneEventForFilters(EventDispatchThread.java:207)
    at java.desktop/java.awt.EventDispatchThread.pumpEventsForFilter(EventDispatchThread.java:128)
    at java.desktop/java.awt.EventDispatchThread.pumpEventsForHierarchy(EventDispatchThread.java:117)
    at java.desktop/java.awt.EventDispatchThread.pumpEvents(EventDispatchThread.java:113)
    at java.desktop/java.awt.EventDispatchThread.pumpEvents(EventDispatchThread.java:105)
    at java.desktop/java.awt.EventDispatchThread.run(EventDispatchThread.java:92)

```

### 원인

## There was a problem with the editor

### 문제

```log
There was a problem with the editor '/Users/rody/Library/Caches/JetBrains/IntelliJIdea2023.3/tmp/intellij-git-editor-local.sh'.  Local changes were shelved before rebase.
```

리베이스 하려면 꽤 자주 위와 같은 에러가 발생

### 원인

`.gitconfig`에서 `rebase.instructionFormat` 옵션을 커스텀한 값으로 설정했기 때문

### 해결

## IntelliJ 터미널의 PHP 버전과 iTERM2에서의 PHP 버전이 상이함

### 문제

IntelliJ 터미널에서 `php --version` 실행하면 5.6, iTERM2에서 실행하면 8.1이 출력됨

### 원인

`Settings > Languages & Frameworks > PHP`에서 CLI interpreter가 5.6으로 설정되어 있었음

### 해결

`Settings > Languages & Frameworks > PHP`에서 CLI interpreter 버전을 8.1로 수정

## 터미널에서 amazon q cli 작동하지 않음

### 문제

```bash
❯ q doctor

✘ Input Method: Not installed

  Run q integrations install input-method to enable it

✘ Doctor found errors. Please fix them and try again.

If you are not sure how to fix it, please open an issue with q issue to let us know!
```

### 원인

integration이 잘 안 된 거 같음

### 해결

```bash
❯ q integrations install input-method
2024-06-28 14:30:41.322 q[73059:658006] TISFileInterrogator updateSystemInputSources false but old data invalid: currentCacheHeaderPtr nonNULL? 0, ->cacheFormatVersion 0, ->magicCookie 00000000, inputSourceTableCountSys 0
Keyboard Layouts: duplicate keyboard layout identifier -17410.
Keyboard Layouts: keyboard layout identifier -17410 has been replaced with -28673.
Keyboard Layouts: duplicate keyboard layout identifier -30769.
Keyboard Layouts: keyboard layout identifier -30769 has been replaced with -28674.
Keyboard Layouts: duplicate keyboard layout identifier -14934.
Keyboard Layouts: keyboard layout identifier -14934 has been replaced with -28675.
Installed!
You must restart your terminal to finish installing the input method.
```

그리고 재시작

## The plugin com.jetbrains.php failed to save settings

### 문제

```bash
Unable to save plugin settings
The plugin com.jetbrains.php failed to save settings. Please restart IntelliJ IDEA
```

### 원인

[참고 링크](https://youtrack.jetbrains.com/issue/WI-64671/Unable-to-save-settings#focus=Comments-27-5719414.0-0)에 의하면 composer.json 경로를 확인해보거나 아니면 repair IDE 해볼 것

### 해결

composer.json 경로 다시 지정하니까 되는 거 같다

## PHPdoc 작성시 하이라이트가 깜막이며 타이핑 딜레이

### 문제

PHPdoc 작성시 하이라이트가 깜막이며 타이핑 딜레이

### 원인

- Editor > General > Smart Keys > PHP > Auto-insert closing HTML tag in PHPDoc blocks 옵션 해제
    - 여전히 같음
- inspections 에서 PHP 다 꺼보기
    - 그래도 똑같음
- psalm 비활성화
    - 마찬가지
- grazie 비활성화
    - 마찬가지
- better highlight 삭제
    - 마찬가지

[`full-line-inference`가 CPU를 많이 잡아 먹어서](https://youtrack.jetbrains.com/issue/WEB-66566/Highlightings-flash-each-time-key-pressed) 그렇다는 얘기도 있는데...

[댓글에 달린 이슈](https://youtrack.jetbrains.com/issue/IJPL-148285/Highlighting-sometimes-blinks-during-typing)가 가장 비슷해 보인다.

> Looks like The problem is in the lack of incremental PSI reparse.
> Every time you press enter inside the method here:

증분 재분석(incremental PSI reparse)?이 부족한 거 같다?

> Invalidating too much PSI is bad for many reasons:
> - too many caches are cleared,
> - too many expensive parsing are performed,
> - too much text is blinking (because all references to this method should become red and then black again).
>
> I think this effect of blinking was always there, although less visible.

1. PSI (Program Structure Interface):
   - PSI는 JetBrains IDE에서 사용하는 추상 구문 트리(AST)의 JetBrains 버전입니다.
   - 소스 코드의 구조를 표현하는 트리 형태의 데이터 구조입니다.
   - 코드의 문법적 구조, 의미, 관계 등을 분석하고 표현합니다.

2. 증분 PSI 재분석 (Incremental PSI Reparse):
   - 코드가 변경될 때마다 전체 PSI 트리를 다시 만드는 대신, 변경된 부분만 업데이트하는 기법입니다.
   - 목적: 성능 향상과 리소스 사용 최소화

3. "증분 PSI 재분석 부족" 의미:
   - 작은 코드 변경에도 불필요하게 큰 부분의 PSI를 다시 분석하고 있다는 뜻입니다.
   - 결과: 성능 저하, UI 깜빡임, 캐시 무효화 등의 문제 발생

4. 문제 해결 방안:
   a) IDE 업데이트:
      - 최신 버전의 IDE로 업데이트하세요. JetBrains는 지속적으로 이러한 문제들을 개선하고 있습니다.

   b) 플러그인 검토:
      - 설치된 플러그인들을 검토하고, 불필요하거나 문제를 일으킬 수 있는 플러그인을 비활성화하거나 제거하세요.

   c) 캐시 및 인덱스 재생성:
      - File > Invalidate Caches and Restart 옵션을 사용하여 IDE 캐시를 초기화하세요.

   d) 프로젝트 설정 최적화:
      - 프로젝트의 빌드 설정, 인스펙션 설정 등을 검토하고 최적화하세요.

   e) 하드웨어 리소스 확인:
      - CPU, 메모리 사용량을 모니터링하고, 필요하다면 하드웨어 업그레이드를 고려하세요.

   f) JVM 옵션 조정:
      - IDE의 JVM 옵션을 조정하여 메모리 할당을 늘리거나 성능을 최적화할 수 있습니다.

   g) 문제 보고:
      - 지속적인 문제라면 JetBrains 이슈 트래커에 상세한 보고를 해주세요. 재현 가능한 최소한의 예제를 포함하면 도움이 됩니다.

5. 개발자로서 할 수 있는 일:
   - 대규모 메서드나 클래스를 피하고 코드를 모듈화하세요.
   - 복잡한 중첩 구조를 단순화하세요.
   - 코드 스타일 가이드라인을 따르세요.

6. JetBrains의 대응:
   - JetBrains 팀은 이러한 문제를 인식하고 있으며, 지속적으로 PSI 파싱 및 증분 업데이트 알고리즘을 개선하고 있습니다.
   - 향후 IDE 버전에서는 이러한 문제들이 점진적으로 해결될 것으로 예상됩니다.

결론적으로, 이 문제는 IDE의 내부 작동 방식과 관련된 복잡한 이슈입니다. 사용자 입장에서는 위의 해결 방안들을 시도해볼 수 있지만, 근본적인 해결은 JetBrains 팀의 지속적인 개선에 달려있습니다. 최신 버전의 IDE를 사용하고, 문제가 지속된다면 JetBrains에 상세한 피드백을 제공하는 것이 도움이 될 수 있습니다.

---

그냥 설정 초기화...

PHP 플러그인 설치하니 발생. 플러그인 수정될 때까지 어쩔 수 없을 듯?

### 해결

업데이트될 때까지 기다리자...
