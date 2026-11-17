# ---- build stage ----
FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /app
COPY pom.xml ./
RUN mvn -q -B dependency:go-offline
COPY src ./src
ARG GIT_COMMIT=unknown
RUN mvn -q -B -DskipTests=false clean package -Dapp.git-commit=${GIT_COMMIT}

# ---- final stage ----
FROM eclipse-temurin:21-jre-alpine AS final
ARG GIT_COMMIT=unknown
ENV GIT_COMMIT=$GIT_COMMIT
RUN apk add --no-cache redis && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /app/target/kitchen-worker-0.0.1.jar app.jar
COPY healthcheck.sh ./healthcheck.sh
RUN chmod +x healthcheck.sh && chown -R app:app /app
USER app
HEALTHCHECK --interval=10s --timeout=5s --retries=3 CMD ["./healthcheck.sh"]
ENTRYPOINT ["java", "-jar", "app.jar"]
