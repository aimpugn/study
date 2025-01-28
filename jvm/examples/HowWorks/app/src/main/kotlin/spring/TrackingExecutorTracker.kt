package spring

import org.springframework.core.task.TaskExecutor
import java.util.concurrent.atomic.AtomicReference

class TrackingExecutorTracker {
    companion object {
        private val _currentExecutor = AtomicReference<TaskExecutor?>()
        var currentExecutor: TaskExecutor?
            get() = _currentExecutor.get()
            set(value) {
                _currentExecutor.set(value)
            }
    }
}