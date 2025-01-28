package spring

import org.springframework.core.task.TaskExecutor

class TrackingCustomTaskExecutor(
    private val delegator: TaskExecutor,
) : TaskExecutor {
    override fun execute(task: Runnable) {
        delegator.execute {
            TrackingExecutorTracker.currentExecutor = this
            task.run()
        }
    }
}