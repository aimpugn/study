# Java Thread

- [Java Thread](#java-thread)
    - [thread.start()](#threadstart)
    - [`JNINativeMethod`: 자바 메서드를 네이티브 코드(C/C++)와 매핑](#jninativemethod-자바-메서드를-네이티브-코드cc와-매핑)
    - [`JavaThread::JavaThread` in src/hotspot/share/runtime/javaThread.cpp](#javathreadjavathread-in-srchotspotshareruntimejavathreadcpp)
    - [`os::create_thread` in src/hotspot/os/linux/os\_linux.cp](#oscreate_thread-in-srchotspotoslinuxos_linuxcp)
    - [생성후 `Thread::start` 호출하여 시작](#생성후-threadstart-호출하여-시작)

## thread.start()

`Thread.class`를 보면 스레드 시작시 아래 메서드가 실행됩니다.

```java
// Thread.class
private native void start0();
```

## `JNINativeMethod`: 자바 메서드를 네이티브 코드(C/C++)와 매핑

OpenJDK에서 "start0"를 검색해보면 아래와 같이 JVM 내부에서 자바 메서드와 네이티브 메서드를 연결하는 부분이 있습니다.
주로 *JNI (Java Native Interface)*를 통해 자바 메서드를 네이티브 코드(C/C++)와 매핑하는 역할을 합니다.
자바의 특정 메서드가 호출될 때, 해당 메서드를 네이티브 코드(C/C++)에서 처리할 수 있도록 메서드 이름과 그 구현을 연결하는 설정입니다.

```c
#define THD "Ljava/lang/Thread;"
#define OBJ "Ljava/lang/Object;"
#define STE "Ljava/lang/StackTraceElement;"
#define STR "Ljava/lang/String;"

#define ARRAY_LENGTH(a) (sizeof(a)/sizeof(a[0]))

// https://github.com/openjdk/jdk/blob/73ebb848fdb66861e912ea747c039ddd1f7a5f48/src/java.base/share/native/libjava/Thread.c#L38C1-L58C3
static JNINativeMethod methods[] = {
    {"start0",           "()V",        (void *)&JVM_StartThread},
    {"setPriority0",     "(I)V",       (void *)&JVM_SetThreadPriority},
    {"yield0",           "()V",        (void *)&JVM_Yield},
    {"sleepNanos0",      "(J)V",       (void *)&JVM_SleepNanos},
    {"currentCarrierThread", "()" THD, (void *)&JVM_CurrentCarrierThread},
    {"currentThread",    "()" THD,     (void *)&JVM_CurrentThread},
    {"setCurrentThread", "(" THD ")V", (void *)&JVM_SetCurrentThread},
    {"interrupt0",       "()V",        (void *)&JVM_Interrupt},
    {"holdsLock",        "(" OBJ ")Z", (void *)&JVM_HoldsLock},
    {"getThreads",       "()[" THD,    (void *)&JVM_GetAllThreads},
    {"dumpThreads",      "([" THD ")[[" STE, (void *)&JVM_DumpThreads},
    {"getStackTrace0",   "()" OBJ,     (void *)&JVM_GetStackTrace},
    {"setNativeName",    "(" STR ")V", (void *)&JVM_SetNativeThreadName},
    {"scopedValueCache", "()[" OBJ,    (void *)&JVM_ScopedValueCache},
    {"setScopedValueCache", "([" OBJ ")V",(void *)&JVM_SetScopedValueCache},
    {"getNextThreadIdOffset", "()J",   (void *)&JVM_GetNextThreadIdOffset},
    {"findScopedValueBindings", "()" OBJ, (void *)&JVM_FindScopedValueBindings},
    {"ensureMaterializedForStackWalk",
                         "(" OBJ ")V", (void*)&JVM_EnsureMaterializedForStackWalk_func},
};
```

`{"자바 클래스에서 호출될 메서드 이름", "메서드 시그니처", 실행될 네이티브 함수의 포인터}` 구조입니다.
- 메서드 시그니처
    - "()V": 파라미터가 없음 & void 반환 타입
    - "(I)V": 정수(int) 파라미터 하나 & void 반환 타입
    - "(J)V": 정수(long) 파라미터 하나 & void 반환 타입
    - "()" THD: 파라미터가 없음 & *THD (Thread)*를 반환

    시그니처는 JNI에서 표준화된 방식으로 표현되며, 모든 자바 타입에 대한 대응 관계가 있습니다:

    - `V`: void
    - `I`: int
    - `J`: long
    - `Z`: boolean
    - `L...;`: 객체 (예: `Ljava/lang/String;`는 `java.lang.String`을 의미)
    - `[`는 배열을 의미 (예: `[" THD`는 `Thread[]` 배열을 의미)

- 네이티브 메서드 포인터 (C++ 함수 포인터)

    해당 자바 메서드를 처리할 네이티브 함수의 포인터를 가리킵니다.
    이 경우, `JVM_StartThread`라는 네이티브 함수가 자바의 `start0()` 메서드를 처리하게 됩니다.

    네이티브 함수는 JNI 규격에 맞춰서 작성된 C/C++ 함수입니다.
    자바 메서드 호출이 있을 때 JVM은 이 포인터를 통해 네이티브 함수를 호출하고, 네이티브 함수는 자바 메서드의 실제 작업을 처리합니다.

Java의 `start0()` 메서드를 처리하는 [`JVM_StartThread`](https://github.com/openjdk/jdk/blob/73ebb848fdb66861e912ea747c039ddd1f7a5f48/src/hotspot/share/prims/jvm.cpp#L2920-L3013)는 다음과 같습니다.
[`JavaThread` 타입에 대한 정의](https://github.com/openjdk/jdk/blob/73ebb848fdb66861e912ea747c039ddd1f7a5f48/src/hotspot/share/runtime/javaThread.hpp#L80C7-L1239)를 참고합니다.

```cpp
JVM_ENTRY(void, JVM_StartThread(JNIEnv* env, jobject jthread))
#if INCLUDE_CDS
  if (CDSConfig::is_dumping_static_archive()) {
    // During java -Xshare:dump, if we allow multiple Java threads to
    // execute in parallel, symbols and classes may be loaded in
    // random orders which will make the resulting CDS archive
    // non-deterministic.
    //
    // Lucikly, during java -Xshare:dump, it's important to run only
    // the code in the main Java thread (which is NOT started here) that
    // creates the module graph, etc. It's safe to not start the other
    // threads which are launched by class static initializers
    // (ReferenceHandler, FinalizerThread and CleanerImpl).
    if (log_is_enabled(Info, cds)) {
      ResourceMark rm;
      oop t = JNIHandles::resolve_non_null(jthread);
      log_info(cds)("JVM_StartThread() ignored: %s", t->klass()->external_name());
    }
    return;
  }
#endif
  JavaThread *native_thread = nullptr;

  // We cannot hold the Threads_lock when we throw an exception,
  // due to rank ordering issues. Example:  we might need to grab the
  // Heap_lock while we construct the exception.
  bool throw_illegal_thread_state = false;

  // We must release the Threads_lock before we can post a jvmti event
  // in Thread::start.
  {
    // Ensure that the C++ Thread and OSThread structures aren't freed before
    // we operate.
    MutexLocker mu(Threads_lock);

    // Since JDK 5 the java.lang.Thread threadStatus is used to prevent
    // re-starting an already started thread, so we should usually find
    // that the JavaThread is null. However for a JNI attached thread
    // there is a small window between the Thread object being created
    // (with its JavaThread set) and the update to its threadStatus, so we
    // have to check for this
    if (java_lang_Thread::thread(JNIHandles::resolve_non_null(jthread)) != nullptr) {
      throw_illegal_thread_state = true;
    } else {
      jlong size =
             java_lang_Thread::stackSize(JNIHandles::resolve_non_null(jthread));
      // Allocate the C++ Thread structure and create the native thread.  The
      // stack size retrieved from java is 64-bit signed, but the constructor takes
      // size_t (an unsigned type), which may be 32 or 64-bit depending on the platform.
      //  - Avoid truncating on 32-bit platforms if size is greater than UINT_MAX.
      //  - Avoid passing negative values which would result in really large stacks.
      NOT_LP64(if (size > SIZE_MAX) size = SIZE_MAX;)
      size_t sz = size > 0 ? (size_t) size : 0;
      native_thread = new JavaThread(&thread_entry, sz);

      // At this point it may be possible that no osthread was created for the
      // JavaThread due to lack of memory. Check for this situation and throw
      // an exception if necessary. Eventually we may want to change this so
      // that we only grab the lock if the thread was created successfully -
      // then we can also do this check and throw the exception in the
      // JavaThread constructor.
      if (native_thread->osthread() != nullptr) {
        // Note: the current thread is not being used within "prepare".
        native_thread->prepare(jthread);
      }
    }
  }

  if (throw_illegal_thread_state) {
    THROW(vmSymbols::java_lang_IllegalThreadStateException());
  }

  assert(native_thread != nullptr, "Starting null thread?");

  if (native_thread->osthread() == nullptr) {
    ResourceMark rm(thread);
    log_warning(os, thread)("Failed to start the native thread for java.lang.Thread \"%s\"",
                            JavaThread::name_for(JNIHandles::resolve_non_null(jthread)));
    // No one should hold a reference to the 'native_thread'.
    native_thread->smr_delete();
    if (JvmtiExport::should_post_resource_exhausted()) {
      JvmtiExport::post_resource_exhausted(
        JVMTI_RESOURCE_EXHAUSTED_OOM_ERROR | JVMTI_RESOURCE_EXHAUSTED_THREADS,
        os::native_thread_creation_failed_msg());
    }
    THROW_MSG(vmSymbols::java_lang_OutOfMemoryError(),
              os::native_thread_creation_failed_msg());
  }

  JFR_ONLY(Jfr::on_java_thread_start(thread, native_thread);)

  Thread::start(native_thread);

JVM_END
```

## `JavaThread::JavaThread` in src/hotspot/share/runtime/javaThread.cpp

```cpp
// https://github.com/openjdk/jdk/blob/ade17ecb6cb5125d048401a878b557e5afefc08c/src/hotspot/share/runtime/javaThread.cpp#L637C1-L655C2
JavaThread::JavaThread(ThreadFunction entry_point, size_t stack_sz, MemTag mem_tag) : JavaThread(mem_tag) {
  set_entry_point(entry_point);
  // Create the native thread itself.
  // %note runtime_23
  os::ThreadType thr_type = os::java_thread;
  thr_type = entry_point == &CompilerThread::thread_entry ? os::compiler_thread :
                                                            os::java_thread;
  os::create_thread(this, thr_type, stack_sz);
  // The _osthread may be null here because we ran out of memory (too many threads active).
  // We need to throw and OutOfMemoryError - however we cannot do this here because the caller
  // may hold a lock and all locks must be unlocked before throwing the exception (throwing
  // the exception consists of creating the exception object & initializing it, initialization
  // will leave the VM via a JavaCall and then all locks must be unlocked).
  //
  // The thread is still suspended when we reach here. Thread must be explicit started
  // by creator! Furthermore, the thread must also explicitly be added to the Threads list
  // by calling Threads:add. The reason why this is not done here, is because the thread
  // object must be fully initialized (take a look at JVM_Start)
}
```

`new JavaThread`로 새로운 자바 스레드 인스턴스 생성시에 이미 os 스레드를 생성합니다.

## `os::create_thread` in src/hotspot/os/linux/os_linux.cp

```cpp
// https://github.com/openjdk/jdk/blob/ade17ecb6cb5125d048401a878b557e5afefc08c/src/hotspot/os/linux/os_linux.cpp#L967C1-L1108C2
bool os::create_thread(Thread* thread, ThreadType thr_type,
                       size_t req_stack_size) {
  assert(thread->osthread() == nullptr, "caller responsible");

  // Allocate the OSThread object
  OSThread* osthread = new (std::nothrow) OSThread();
  if (osthread == nullptr) {
    return false;
  }

  // set the correct thread state
  osthread->set_thread_type(thr_type);

  // Initial state is ALLOCATED but not INITIALIZED
  osthread->set_state(ALLOCATED);

  thread->set_osthread(osthread);

  // init thread attributes
  pthread_attr_t attr;
  int rslt = pthread_attr_init(&attr);
  if (rslt != 0) {
    thread->set_osthread(nullptr);
    delete osthread;
    return false;
  }
  pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_DETACHED);

  // Calculate stack size if it's not specified by caller.
  size_t stack_size = os::Posix::get_initial_stack_size(thr_type, req_stack_size);
  size_t guard_size = os::Linux::default_guard_size(thr_type);

  // Configure glibc guard page. Must happen before calling
  // get_static_tls_area_size(), which uses the guard_size.
  pthread_attr_setguardsize(&attr, guard_size);

  // Apply stack size adjustments if needed. However, be careful not to end up
  // with a size of zero due to overflow. Don't add the adjustment in that case.
  size_t stack_adjust_size = 0;
  if (AdjustStackSizeForTLS) {
    // Adjust the stack_size for on-stack TLS - see get_static_tls_area_size().
    stack_adjust_size += get_static_tls_area_size(&attr);
  } else if (os::Linux::adjustStackSizeForGuardPages()) {
    stack_adjust_size += guard_size;
  }

  stack_adjust_size = align_up(stack_adjust_size, os::vm_page_size());
  if (stack_size <= SIZE_MAX - stack_adjust_size) {
    stack_size += stack_adjust_size;
  }
  assert(is_aligned(stack_size, os::vm_page_size()), "stack_size not aligned");

  if (THPStackMitigation) {
    // In addition to the glibc guard page that prevents inter-thread-stack hugepage
    // coalescing (see comment in os::Linux::default_guard_size()), we also make
    // sure the stack size itself is not huge-page-size aligned; that makes it much
    // more likely for thread stack boundaries to be unaligned as well and hence
    // protects thread stacks from being targeted by khugepaged.
    if (HugePages::thp_pagesize() > 0 &&
        is_aligned(stack_size, HugePages::thp_pagesize())) {
      stack_size += os::vm_page_size();
    }
  }

  int status = pthread_attr_setstacksize(&attr, stack_size);
  if (status != 0) {
    // pthread_attr_setstacksize() function can fail
    // if the stack size exceeds a system-imposed limit.
    assert_status(status == EINVAL, status, "pthread_attr_setstacksize");
    log_warning(os, thread)("The %sthread stack size specified is invalid: " SIZE_FORMAT "k",
                            (thr_type == compiler_thread) ? "compiler " : ((thr_type == java_thread) ? "" : "VM "),
                            stack_size / K);
    thread->set_osthread(nullptr);
    delete osthread;
    pthread_attr_destroy(&attr);
    return false;
  }

  ThreadState state;

  {
    ResourceMark rm;
    pthread_t tid;
    int ret = 0;
    int limit = 3;
    do {
      ret = pthread_create(&tid, &attr, (void* (*)(void*)) thread_native_entry, thread);
    } while (ret == EAGAIN && limit-- > 0);

    char buf[64];
    if (ret == 0) {
      log_info(os, thread)("Thread \"%s\" started (pthread id: " UINTX_FORMAT ", attributes: %s). ",
                           thread->name(), (uintx) tid, os::Posix::describe_pthread_attr(buf, sizeof(buf), &attr));

      // Print current timer slack if override is enabled and timer slack value is available.
      // Avoid calling prctl otherwise for extra safety.
      if (TimerSlack >= 0) {
        int slack = prctl(PR_GET_TIMERSLACK);
        if (slack >= 0) {
          log_info(os, thread)("Thread \"%s\" (pthread id: " UINTX_FORMAT ") timer slack: %dns",
                               thread->name(), (uintx) tid, slack);
        }
      }
    } else {
      log_warning(os, thread)("Failed to start thread \"%s\" - pthread_create failed (%s) for attributes: %s.",
                              thread->name(), os::errno_name(ret), os::Posix::describe_pthread_attr(buf, sizeof(buf), &attr));
      // Log some OS information which might explain why creating the thread failed.
      log_info(os, thread)("Number of threads approx. running in the VM: %d", Threads::number_of_threads());
      LogStream st(Log(os, thread)::info());
      os::Posix::print_rlimit_info(&st);
      os::print_memory_info(&st);
      os::Linux::print_proc_sys_info(&st);
      os::Linux::print_container_info(&st);
    }

    pthread_attr_destroy(&attr);

    if (ret != 0) {
      // Need to clean up stuff we've allocated so far
      thread->set_osthread(nullptr);
      delete osthread;
      return false;
    }

    // Store pthread info into the OSThread
    osthread->set_pthread_id(tid);

    // Wait until child thread is either initialized or aborted
    {
      Monitor* sync_with_child = osthread->startThread_lock();
      MutexLocker ml(sync_with_child, Mutex::_no_safepoint_check_flag);
      while ((state = osthread->get_state()) == ALLOCATED) {
        sync_with_child->wait_without_safepoint_check();
      }
    }
  }

  // The thread is returned suspended (in state INITIALIZED),
  // and is started higher up in the call chain
  assert(state == INITIALIZED, "race condition");
  return true;
}
```

`pthread_create(&tid, &attr, (void* (*)(void*)) thread_native_entry, thread)`를 통해 스레드를 생성합니다.
단, 생성된 스레드는 실제로 OS 레벨에서 생성되지만, 즉시 실행되지는 않으며 새로 생성된 스레드는 초기화 상태로 대기합니다.

생성된 스레드는 `ALLOCATED` 상태에서 시작하여 `INITIALIZED` 상태로 전환됩니다.
이 과정에서 JVM은 스레드에 필요한 리소스를 할당하고 초기 설정을 수행합니다.

별도의 `Thread::start`가 필요한 이유는 다음과 같습니다:
- 안전한 초기화: JVM이 스레드를 완전히 초기화하고 필요한 모든 설정을 마칠 때까지 스레드 실행을 지연시킵니다.
- 동기화 및 가시성 보장: `start()` 메서드 호출은 Java 메모리 모델에서 중요한 동기화 지점입니다. 이는 스레드 시작 전의 모든 작업이 새 스레드에 올바르게 보이도록 보장합니다.
- 상태 관리: JVM은 스레드의 생명주기를 세밀하게 제어하고 관리할 수 있습니다.

스레드의 생성과 시작을 분리함으로써, JVM은 스레드의 생명주기를 더 정밀하게 제어할 수 있고, Java 프로그래머에게는 예측 가능한 스레드 행동을 제공할 수 있습니다.

## 생성후 `Thread::start` 호출하여 시작

실제 스레드를 시작하는 `Thread::start`를 찾아가보면 아래와 같습니다.

```cpp
// https://github.com/openjdk/jdk/blob/73ebb848fdb66861e912ea747c039ddd1f7a5f48/src/hotspot/share/runtime/thread.cpp#L379-L391
void Thread::start(Thread* thread) {
  // Start is different from resume in that its safety is guaranteed by context or
  // being called from a Java method synchronized on the Thread object.
  if (thread->is_Java_thread()) {
    // Initialize the thread state to RUNNABLE before starting this thread.
    // Can not set it after the thread started because we do not know the
    // exact thread state at that time. It could be in MONITOR_WAIT or
    // in SLEEPING or some other state.
    java_lang_Thread::set_thread_status(JavaThread::cast(thread)->threadObj(),
                                        JavaThreadStatus::RUNNABLE);
  }
  os::start_thread(thread);
}
```

`Thread::start`가 호출되면 다음 과정이 진행됩니다:
- 스레드 상태를 `RUNNABLE`로 설정합니다.
- `os::start_thread`를 호출하여 OS 레벨에서 스레드를 실행 가능한 상태로 만듭니다.
- Linux의 경우, `pd_start_thread`에서 `sync_with_child->notify()`를 호출하고, notify 신호를 받은 스레드는 대기 상태에서 벗어나 실제로 실행을 시작합니다.

자바 쓰레드면 `set_thread_status` 함수를 호출하는데, 이는 내부적으로 `SET_FIELDHOLDER_FIELD` 매크로를 호출합니다.

```cpp
// https://github.com/openjdk/jdk/blob/73ebb848fdb66861e912ea747c039ddd1f7a5f48/src/hotspot/share/classfile/javaClasses.cpp#L1565C1-L1567C2
void java_lang_Thread_FieldHolder::set_thread_status(oop holder, JavaThreadStatus status) {
  holder->int_field_put(_thread_status_offset, static_cast<int>(status));
}

// https://github.com/openjdk/jdk/blob/73ebb848fdb66861e912ea747c039ddd1f7a5f48/src/hotspot/share/classfile/javaClasses.cpp#L1753C1-L1759C4
// We should never be trying to set a field of an attaching thread.
#define SET_FIELDHOLDER_FIELD(java_thread, field, value)        \
  {                                                             \
    oop holder = java_lang_Thread::holder(java_thread);         \
    assert(holder != nullptr, "Thread not fully initialized");  \
    java_lang_Thread_FieldHolder::set_##field(holder, value);   \
  }


// https://github.com/openjdk/jdk/blob/73ebb848fdb66861e912ea747c039ddd1f7a5f48/src/hotspot/share/classfile/javaClasses.cpp#L1805C1-L1808C2
// Write the thread status value to threadStatus field in java.lang.Thread java class.
void java_lang_Thread::set_thread_status(oop java_thread, JavaThreadStatus status) {
  SET_FIELDHOLDER_FIELD(java_thread, thread_status, status);
}
```

- `oop holder = java_lang_Thread::holder(java_thread);`:

    - `oop`는 ordinary object pointer의 약자로, JVM 내에서 자바 객체를 가리키는 포인터 타입입니다.
    - `holder` 객체는 `java.lang.Thread`의 필드로, 스레드의 내부 상태를 보관하는 객체입니다.

- `java_lang_Thread_FieldHolder::set_##field(holder, value)`:

    매크로 토큰 결합(token pasting) 기능을 사용한 것으로, `field` 매개변수의 값에 따라 `set_필드이름`이 동적으로 결정됩니다.
    예를 들어, `field`가 "priority"라면, 이 구문은`set_priority(holder, value)`로 변환됩니다.

    `java_lang_Thread_FieldHolder::set_thread_status`를 호출하게 됩니다.

```cpp
// https://github.com/openjdk/jdk/blob/73ebb848fdb66861e912ea747c039ddd1f7a5f48/src/hotspot/share/runtime/os.cpp#L853C1-L857C2
void os::start_thread(Thread* thread) {
  OSThread* osthread = thread->osthread();
  osthread->set_state(RUNNABLE);
  pd_start_thread(thread);
}
```

- [`osthread()`](https://github.com/openjdk/jdk/blob/ade17ecb6cb5125d048401a878b557e5afefc08c/src/hotspot/share/runtime/thread.hpp#L402C2-L402C73)

    ```hpp
    OSThread* osthread() const                     { return _osthread;   }
    ```

- `pd_start_thread`는 os별로 다양하게 구현되어 있습니다.

    - `pd`: 플랫폼 디펜던트(Platform Dependent)를 의미하며 플랫폼 종속적인 함수를 나타냅니다. 즉, 운영체제마다 다르게 구현됩니다.

    ```cpp
    // src/hotspot/os/bsd/os_bsd.cpp
    // src/hotspot/os/windows/os_windows.cpp
    // src/hotspot/os/linux/os_linux.cpp
    // src/hotspot/os/aix/os_aix.cpp
    void os::pd_start_thread(Thread* thread) {
        // 각 OS별 스레드 시작 로직
    }
    ```

    가령, 설치된 서버의 os가 리눅스인 경우 컴파일되는 `pd_start_thread`는 다음과 같습니다.

    ```cpp
    // src/hotspot/os/linux/os_linux.cpp
    void os::pd_start_thread(Thread* thread) {
        OSThread * osthread = thread->osthread();
        assert(osthread->get_state() != INITIALIZED, "just checking");
        Monitor* sync_with_child = osthread->startThread_lock();
        MutexLocker ml(sync_with_child, Mutex::_no_safepoint_check_flag);
        sync_with_child->notify();
    }
    ```
