# profiling

## Profiling?

> 메모리 사용량, 함수별 CPU 점유 시간, tracing 등 어플리케이션을 동적으로 분석하는 것

프로파일링이란 어플리케이션을 동적으로 분석하는 것을 말한다. 프로파일링을 통해 어플리케이션의 메모리 사용량, 함수별 CPU 점유 시간, tracing 등을 알 수 있다. 프로파일링을 통해 어플리케이션의 성능을 향상시킬 수 있다.

## Origin of Profiling

The term "profiling" in the context of software development originated from the need to analyze and optimize computer program performance.
As applications grew in complexity, it became essential to identify specific areas that hindered their efficiency.
Profiling emerged as a systematic approach to examining how resources are used during a program's execution, allowing developers to fine-tune the code for better performance.

## Why is Profiling Important?

1. Performance Optimization: By identifying slow or resource-heavy parts of the code, profiling helps in optimizing them for better performance.
2. Resource Management: Profiling provides insights into the usage of critical resources like CPU and memory, enabling better management and allocation.
3. Bottleneck Identification: It helps in pinpointing specific areas where performance bottlenecks occur, which might not be evident during standard testing.

## Relation to Sampling

Profiling often involves "sampling," a method where **snapshots of application performance metrics are taken at regular intervals**. Sampling is less resource-intensive than continuously monitoring every function or process in an application.

### What is Sampling?

> Sampling is a technique used in profiling to periodically record information about a program's operation, rather than continuously logging data.

Sampling involves periodically taking snapshots of a program’s state. These snapshots can include information about CPU usage, memory allocation, and other performance metrics. By analyzing these samples, developers can get a representative overview of the application's performance without the overhead of continuous monitoring.

## Additional Knowledge

- Types of Profiling: Profiling can be either *instrumentation-based*, where specific code is added to measure performance, or *sampling-based*, which involves taking periodic snapshots of the application's state.
- Tools and Utilities: Various tools exist for profiling, like gProfiler, VisualVM, and Instruments, each with its own set of features and focus areas.
- Best Practices: Effective profiling requires a balance between the depth of analysis and the performance overhead caused by the profiling process itself.
