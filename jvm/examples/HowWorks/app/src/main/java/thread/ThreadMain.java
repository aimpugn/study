package thread;

import util.RunExampleInvoker;
import utils.CoroutineUtil;

public class ThreadMain {
    public static void main(String[] args) {
        // Continuation 객체 생성
        RunExampleInvoker.Companion.invoke(
                ThreadMain.class.getPackageName(),
                CoroutineUtil.newContinuation()
        );
    }
}
