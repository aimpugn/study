<!-- @format -->

# tmux

- [tmux](#tmux)
    - [개요](#개요)
    - [commands](#commands)

## 개요

tmux is a terminal multiplexer:

it enables a number of terminals to be created, accessed, and controlled from a single screen.

tmux may be detached from a screen and continue running in the background, then later reattached.

When tmux is started, it creates a new session with a single window and displays it on screen. A status line at the bottom of the screen shows information on the current session and is used to enter interactive commands.

A session is a single collection of pseudo terminals under the management of tmux. Each session has one or more windows linked to it.

A window occupies the entire screen and may be split into rectangular panes, each of which is a separate pseudo terminal (the pty(4) manual page documents the technical details of pseudo terminals).

Any number of tmux instances may connect to the same session, and any number of windows may be present in the same session. Once all sessions are killed, tmux exits.

Each session is persistent and will survive accidental disconnection (such as ssh(1) connection timeout) or intentional detaching (with the `C-b d` key strokes). tmux may be reattached using:

    tmux attach

In tmux, a session is displayed on screen by a client and all sessions are managed by a single server. The server and each client are separate
processes which communicate through a socket in /tmp.

## commands
