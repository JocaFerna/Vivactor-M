FROM eclipse-temurin:11-jdk-focal

# Install Maven
RUN apt-get update && apt-get install -y maven && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# The JAR is inside the same folder as the Dockerfile
COPY ws4j-1.0.1.jar /app/ws4j-1.0.1.jar

# Install the manual JAR
RUN mvn install:install-file -Dfile=/app/ws4j-1.0.1.jar \
    -DgroupId=com.sciss -DartifactId=ws4j -Dversion=1.0.1 -Dpackaging=jar

# Copy the pom and src (since context is ./msa-nose, these are at the root)
COPY pom.xml .
COPY src ./src

RUN mvn clean package -DskipTests

EXPOSE 8080
ENTRYPOINT ["java", "-jar", "target/msa-nose-0.0.1.jar"]