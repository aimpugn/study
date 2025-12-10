package com.example.portable.binary;

import jakarta.annotation.PreDestroy;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.ServerSocket;
import java.net.Socket;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Arrays;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Very small binary protocol server that uses a fixed frame length.
 * <p>
 * Each frame is read as a fixed-size byte array and the contents are
 * interpreted as UTF-8 text for this PoC. The handler simply logs the
 * payload and returns a small ACK line.
 */
@Component
public class BinaryProtocolServer {

    private static final Logger log = LoggerFactory.getLogger(BinaryProtocolServer.class);

    private final BinaryServerProperties properties;
    private final ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor();

    private volatile boolean running;
    private ServerSocket serverSocket;

    public BinaryProtocolServer(BinaryServerProperties properties) {
        this.properties = properties;
    }

    @EventListener(ApplicationReadyEvent.class)
    public void start() {
        if (running) {
            return;
        }
        running = true;
        executor.submit(this::acceptLoop);
        log.info("Binary protocol server starting on port {} with frame length {} bytes",
                properties.getPort(), properties.getFrameLength());
    }

    @PreDestroy
    public void stop() {
        running = false;
        if (serverSocket != null && !serverSocket.isClosed()) {
            try {
                serverSocket.close();
            } catch (IOException ex) {
                log.warn("Error closing binary protocol server socket", ex);
            }
        }
        executor.shutdownNow();
        log.info("Binary protocol server stopped");
    }

    private void acceptLoop() {
        try (ServerSocket server = new ServerSocket(properties.getPort())) {
            this.serverSocket = server;
            while (running && !server.isClosed()) {
                Socket client = server.accept();
                executor.submit(() -> handleClient(client));
            }
        } catch (IOException ex) {
            if (running) {
                log.error("Binary protocol server fatal error", ex);
            } else {
                log.debug("Binary protocol server stopped: {}", ex.getMessage());
            }
        }
    }

    private void handleClient(Socket client) {
        String remote = client.getRemoteSocketAddress().toString();
        log.info("Accepted binary connection from {}", remote);
        try (Socket socket = client;
             InputStream in = socket.getInputStream();
             OutputStream out = socket.getOutputStream()) {

            byte[] buffer = new byte[properties.getFrameLength()];
            while (running && !socket.isClosed()) {
                int read = readFully(in, buffer);
                if (read < 0) {
                    break;
                }
                byte[] frame = Arrays.copyOf(buffer, read);
                String payload = new String(frame, StandardCharsets.UTF_8).trim();
                log.info("Received {} bytes from {}: '{}'", read, remote, payload);

                String response = "ACK " + Instant.now() + " len=" + read + "\n";
                out.write(response.getBytes(StandardCharsets.UTF_8));
                out.flush();
            }
        } catch (IOException ex) {
            log.warn("I/O error on binary connection {}: {}", remote, ex.getMessage());
        } finally {
            log.info("Closed binary connection from {}", remote);
        }
    }

    private int readFully(InputStream in, byte[] buffer) throws IOException {
        int offset = 0;
        int length = buffer.length;
        while (offset < length) {
            int read = in.read(buffer, offset, length - offset);
            if (read < 0) {
                return offset == 0 ? -1 : offset;
            }
            offset += read;
        }
        return offset;
    }
}
