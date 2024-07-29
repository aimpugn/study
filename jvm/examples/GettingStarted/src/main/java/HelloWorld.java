package main.java;

/**
 * <pre>
 * ❯ javac HelloWorld.java
 * ❯ ll
 * total 16
 * -rw-r--r--@ 1 rody  staff   426B  9  2 23:44 HelloWorld.class
 * -rw-r--r--@ 1 rody  staff   179B  9  2 23:45 HelloWorld.java
 *
 * ❯ java HelloWorld
 * Hello World!
 * </pre>
 *
 * 하지만 [JEP 330: Launch Single-File Source-Code Programs](https://openjdk.org/jeps/330)부터
 * java 명령어로 바로 실행할 수 있습니다.
 * <pre>
 * ❯ java HelloWorld.java
 * Hello World!
 * </pre>
 */
class HelloWorld {
    public static void main(String[] args) {
        System.out.println("Hello World!");
    }
}
