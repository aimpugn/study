package com.example.portable.binary;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "binary.server")
public class BinaryServerProperties {

    /**
     * TCP port on which the fixed-length binary protocol will listen.
     */
    private int port = 9090;

    /**
     * Fixed frame length in bytes for each message.
     */
    private int frameLength = 300;

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }

    public int getFrameLength() {
        return frameLength;
    }

    public void setFrameLength(int frameLength) {
        this.frameLength = frameLength;
    }
}

