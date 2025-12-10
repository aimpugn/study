package com.example.portable.cli;

import picocli.CommandLine.Command;
import picocli.CommandLine.Option;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.concurrent.Callable;

@Command(
        name = "stat",
        description = "Check status of HTTP and binary servers"
)
public class StatCommand implements Callable<Integer> {

    @Option(
            names = {"--host"},
            description = "Target host (default: ${DEFAULT-VALUE})",
            defaultValue = "localhost"
    )
    private String host;

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
    public Integer call() throws Exception {
        System.out.println("fpsb status against " + host + " (http=" + httpPort + ", binary=" + binaryPort + ")");

        boolean httpUp = checkHttpHealth();
        boolean binaryUp = checkBinary();

        System.out.println(" HTTP   : " + (httpUp ? "UP" : "DOWN"));
        System.out.println(" Binary : " + (binaryUp ? "UP" : "DOWN"));

        return (httpUp && binaryUp) ? 0 : 1;
    }

    private boolean checkHttpHealth() {
        String url = "http://" + host + ":" + httpPort + "/actuator/health";
        try {
            HttpClient client = HttpClient.newBuilder()
                    .connectTimeout(Duration.ofSeconds(1))
                    .build();
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(url))
                    .timeout(Duration.ofSeconds(2))
                    .GET()
                    .build();
            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
            return response.statusCode() == 200 && response.body().contains("\"status\":\"UP\"");
        } catch (Exception ex) {
            return false;
        }
    }

    private boolean checkBinary() {
        try (Socket socket = new Socket()) {
            socket.connect(new InetSocketAddress(host, binaryPort), 1000);
            return socket.isConnected();
        } catch (IOException ex) {
            return false;
        }
    }
}

