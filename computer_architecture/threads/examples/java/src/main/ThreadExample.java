package main;

public class ThreadExample {
    public static void main(String[] args) {
        for (int i = 0; i < 10; i++) {
            Thread t = new Thread(new SampleRunnable(i));
            t.start();
        }

        System.out.println("main function ends");
    }
}


class SampleRunnable implements Runnable {
    int seq;

    public SampleRunnable(int seq) {
        this.seq = seq;
    }

    public void run() {
        System.out.println(
                String.format("[%d] %s run", this.seq, this.getClass())
        );
        int max = 100;
        while (max-- > 0) {
            try {
                Thread.sleep(1000);
                System.out.println(
                        String.format(
                                "[%d] %s, max is %d",
                                this.seq,
                                this.getClass(),
                                max
                        )
                );
            } catch (Exception e) {
            }
        }

        System.out.println(
                String.format("[%d] %s end", this.seq, this.getClass())
        );
    }
}


