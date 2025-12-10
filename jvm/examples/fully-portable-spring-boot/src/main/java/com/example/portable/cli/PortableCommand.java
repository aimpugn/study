package com.example.portable.cli;

import picocli.CommandLine.Command;
import picocli.CommandLine.HelpCommand;
import picocli.CommandLine.Option;

@Command(
        name = "fpsb",
        version = "0.0.1-SNAPSHOT",
        description = "Fully portable Spring Boot native image PoC",
        subcommands = {
                StartCommand.class,
                StatCommand.class,
                ConfigCommand.class,
                HelpCommand.class
        }
)
public class PortableCommand implements Runnable {

    @Option(
            names = {"-h", "--help"},
            usageHelp = true,
            description = "Show this help message and exit."
    )
    boolean helpRequested;

    @Option(
            names = {"-V", "--version"},
            versionHelp = true,
            description = "Print version information and exit."
    )
    boolean versionRequested;

    @Override
    public void run() {
        System.out.println("Use one of the sub-commands: boot, stat, config");
    }
}
