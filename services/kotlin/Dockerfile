FROM gradle:8-jdk21-alpine AS builder

WORKDIR /app
COPY build.gradle.kts settings.gradle.kts gradle.properties ./
COPY gradle ./gradle
COPY gradlew ./
COPY src ./src
COPY gen ./gen

RUN ./gradlew shadowJar --no-daemon

FROM eclipse-temurin:21-jre-alpine

WORKDIR /app
COPY --from=builder /app/build/libs/app.jar /app/app.jar

EXPOSE 8083

CMD ["java", "-jar", "/app/app.jar"]
