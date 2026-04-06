package io.aimpugn.learn.netty.lesson1;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.net.Socket;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;

final class EchoTestClient {

    private static final int TIMEOUT_MILLIS = (int) Duration.ofSeconds(2).toMillis();

    private EchoTestClient() {
    }

    static String exchangeSingleLine(int port, String line) throws IOException {
        try (Socket socket = connect(port);
             BufferedReader reader = new BufferedReader(new InputStreamReader(socket.getInputStream(), StandardCharsets.UTF_8));
             BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream(), StandardCharsets.UTF_8))) {

            writer.write(line);
            writer.newLine();
            writer.flush();
            return reader.readLine();
        }
    }

    static List<String> exchangeTwoLines(int port, String first, String second) throws IOException {
        try (Socket socket = connect(port);
             BufferedReader reader = new BufferedReader(new InputStreamReader(socket.getInputStream(), StandardCharsets.UTF_8));
             BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(socket.getOutputStream(), StandardCharsets.UTF_8))) {

            List<String> responses = new ArrayList<>();
            writer.write(first);
            writer.newLine();
            writer.flush();
            responses.add(reader.readLine());

            writer.write(second);
            writer.newLine();
            writer.flush();
            responses.add(reader.readLine());
            return responses;
        }
    }

    static void connectAndCloseWithoutSending(int port) throws IOException {
        try (Socket socket = connect(port)) {
            // 연결만 열고 바로 닫습니다.
        }
    }

    private static Socket connect(int port) throws IOException {
        Socket socket = new Socket("127.0.0.1", port);
        socket.setSoTimeout(TIMEOUT_MILLIS);
        return socket;
    }
}
