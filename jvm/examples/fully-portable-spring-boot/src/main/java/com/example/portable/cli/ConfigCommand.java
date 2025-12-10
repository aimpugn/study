package com.example.portable.cli;

import picocli.CommandLine.Command;

import java.util.concurrent.Callable;

@Command(
        name = "config",
        description = "Explain how configuration is applied"
)
public class ConfigCommand implements Callable<Integer> {

    @Override
    public Integer call() {
        System.out.println("Configuration sources (in priority order):");
        System.out.println(" 1) CLI flags: start --port / --binary-port");
        System.out.println(" 2) Environment variables: SERVER_PORT, BINARY_SERVER_PORT");
        System.out.println(" 3) application.yml: server.port, binary.server.port, binary.server.frame-length");
        System.out.println();
        System.out.println("Example:");
        System.out.println("  portable-app start --port 8081 --binary-port 9100");
        return 0;
    }
}

