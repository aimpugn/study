package com.example.portable.cli;

import com.example.portable.FullyPortableApplication;
import org.springframework.boot.SpringApplication;
import picocli.CommandLine.Command;
import picocli.CommandLine.Option;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.Callable;

@Command(
        name = "boot",
        aliases = {"start"},
        description = "Start the Spring Boot HTTP server and binary protocol listener"
)
public class StartCommand implements Callable<Integer> {

    @Option(
            names = {"-p", "--port"},
            description = "HTTP server port (default: ${DEFAULT-VALUE})",
            defaultValue = "8080"
    )
    private int httpPort;

    @Option(
            names = {"-b", "--binary-port"},
            description = "Binary protocol TCP port (default: ${DEFAULT-VALUE})",
            defaultValue = "9090"
    )
    private int binaryPort;

    @Override
    public Integer call() {
        Map<String, Object> properties = new HashMap<>();
        properties.put("server.port", httpPort);
        properties.put("binary.server.port", binaryPort);

        SpringApplication application = new SpringApplication(FullyPortableApplication.class);
        application.setMainApplicationClass(FullyPortableApplication.class);
        application.setDefaultProperties(properties);
        application.run();

        // Spring's non-daemon threads keep the JVM alive.
        return 0;
    }
}
