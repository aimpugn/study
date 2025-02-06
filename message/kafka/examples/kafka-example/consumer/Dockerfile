# 멀티 스테이지 빌드: https://docs.docker.com/build/building/multi-stage/
FROM gradle:8.11-jdk21 AS builder
WORKDIR /consumer
COPY . .
RUN gradle clean :app:jar --no-daemon
RUN pwd && ls -la ./build/ && ls -la ./app/build

FROM openjdk:21-jdk-slim
WORKDIR /consumer
COPY --from=builder /consumer/app/build/libs/*.jar app.jar
ENTRYPOINT ["java", "-jar", "app.jar"]
