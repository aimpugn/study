package lambda;

import util.RunExampleInvoker;
import utils.CoroutineUtil;

public class LambdaMain {
    public static void main(String[] args) {
        RunExampleInvoker.Companion.invoke(
                LambdaMain.class.getPackageName(),
                CoroutineUtil.newContinuation()
        );
    }
}
