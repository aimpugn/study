package regex;

import util.RunExampleInvoker;
import utils.CoroutineUtil;

public class RegExpMain {
    public static void main(String[] args) {
        RunExampleInvoker.Companion.invoke("regex", CoroutineUtil.newContinuation());
    }
}
