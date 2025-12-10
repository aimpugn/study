package com.example.portable;

import com.example.portable.cli.ConfigCommand;
import com.example.portable.cli.PortableCommand;
import com.example.portable.cli.StartCommand;
import com.example.portable.cli.StatCommand;
import picocli.CommandLine;
import picocli.CommandLine.HelpCommand;

/**
 * Entry point of the native binary.
 *
 * The process behaves like a Picocli-based CLI. The {@code start}
 * sub-command bootstraps the Spring Boot HTTP server and the binary
 * protocol listener.
 */
public class ApplicationMain {

    public static void main(String[] args) {
        // Native images built via Spring Boot enable AOT mode by default.
        // For this CLI-based launcher we want to run the application in
        // regular (non-AOT) mode so that Spring does not expect an
        // ApplicationMain__ApplicationContextInitializer.
        System.setProperty("spring.aot.enabled", "false");

        PortableCommand portableCommand = new PortableCommand();
        CommandLine.IFactory factory = new CommandLine.IFactory() {
            @Override
            @SuppressWarnings("unchecked")
            public <K> K create(Class<K> cls) throws Exception {
                if (cls == StartCommand.class) {
                    return (K) new StartCommand();
                }
                if (cls == StatCommand.class) {
                    return (K) new StatCommand();
                }
                if (cls == ConfigCommand.class) {
                    return (K) new ConfigCommand();
                }
                if (cls == HelpCommand.class) {
                    return (K) new HelpCommand();
                }
                return CommandLine.defaultFactory().create(cls);
            }
        };

        CommandLine commandLine = new CommandLine(portableCommand, factory);

        commandLine.setExecutionExceptionHandler((ex, cmd, parseResult) -> {
            cmd.getErr().println("ERROR: " + ex.getMessage());
            return 1;
        });
        commandLine.execute(args);
        // We intentionally do NOT call System.exit() here so that
        // non-daemon threads created by Spring (HTTP server, TCP listener)
        // can keep the process alive when "start" is used.
    }
}
