package regex;

import util.RunExample;

import java.util.regex.Pattern;

/**
 * 정규 표현식은 백트래킹(backtracking) 방식으로 동작합니다.
 * 특히 탐욕적 표현식(`.*`)을 잘못 사용할 경우 성능이 크게 저하될 수 있습니다.
 * <p>
 * 정규 표현식을 최적화하거나, 필요에 따라 DFA(Deterministic Finite Automata) 기반의 매칭을 사용하는 라이브러리(예: Google RE2)를 사용하는 게 좋다고 합니다.s
 */
public class RegExpExample {

    @RunExample
    public void simpleRegExp() {
        var text = "User123: email@example.com";
        var patternString = "[a-zA-Z]+\\d+"; // 알파벳 뒤 숫자 패턴 예) User123

        var pattern = Pattern.compile(patternString);
        var matcher = pattern.matcher(text);

        while (matcher.find()) {
            System.out.println("matcher.group: " + matcher.group());
            //matcher.group: User123
        }
    }

    /**
     * ## 전방탐색(Lookahead)
     * `?=`처럼 사용합니다.
     * 예를 들어, `X(?=Y)` 경우 'X 뒤에 Y가 오는 경우만 매칭하되, Y는 결과에 포함하지 않습니다.'를
     * 의미합니다.
     * <p>
     * ## 후방탐색(Lookbehind)
     * `?<=`처럼 사용합니다.
     * 예를 들어, `(?<=Y)X` 경우 'X 앞에 Y가 오는 경우만 매칭하되, Y는 결과에 포함하지 말라.'를
     * 의미합니다.
     */
    @RunExample
    public void lookaheadAndLookbehind() {
        var input = "Java regex8 pattern Java regex9";
        var regex = "(?<=Java )regex(?=8)";
        var matcher = Pattern.compile(regex).matcher(input);

        while (matcher.find()) {
            System.out.println("1. 찾은 위치: " + matcher.start());
            System.out.println("1. 찾은 문자 " + matcher.group());
            //1. 찾은 위치: 5
            //1. 찾은 문자 regex
        }

        var pattern2 = Pattern.compile("hello(?=123)");
        var matcher2 = pattern2.matcher("hello123 hello456");

        while (matcher2.find()) {
            System.out.println("2. 찾은 위치: " + matcher2.start());
            System.out.println("2. 찾은 문자 " + matcher2.group());
            //2. 찾은 위치: 0
            //2. 찾은 문자 hello
        }
    }

    @RunExample
    public void splitBySpace() {
        var data = "Java Kotlin Python";
        var split = data.split("\\s+");

        for (String s : split) {
            System.out.println(s);
            //Java
            //Kotlin
            //Python
        }
    }

    @RunExample
    public void tryRemoveHtmlTag() {
        var html = "<div>Hello <b>Java</b></div>";
        var textOnly = html.replaceAll("<[^>]+>", "");
        System.out.println(textOnly);
        //Hello Java
    }

    @RunExample
    public void grouping() {
        var log = "Error: File not found (404)";
        var p = Pattern.compile("Error: (.+) \\((\\d{3})\\)");
        var m = p.matcher(log);

        if (m.find()) {
            System.out.println("에러 메시지: " + m.group(1));
            System.out.println("오류 코드: " + m.group(2));
            //에러 메시지: File not found
            //에러 코드: 404
        }
    }

    @RunExample
    public void greedyAndLazy() {
        var input = "<p>Text1</p><p>Text2</p>";
        // Greedy
        System.out.println(
            // 가능한 한 가장 긴 문자열을 선택합니다.
            input.replaceAll("<.*>", "")
        ); // 출력 없음 (전부 지워짐)

        // Lazy
        System.out.println(
            // 가능한 한 가장 짧은 문자열을 선택합니다.
            input.replaceAll("<.*?>", "")
        ); // Text1Text2
    }

    @RunExample
    public void multiLineAndDotAll() {
        var multiline = "first line\nsecond line";
        var pattern = Pattern.compile("first.*line", Pattern.DOTALL);
        var matcher = pattern.matcher(multiline);
        if (matcher.find()) {
            System.out.println(matcher.group());
            //first line
            //second line
        }

        // `MULTILINE`을 사용하면 `^`는 각 줄의 맨 앞, `$`는 각 줄의 끝에 매칭됩니다.
        // 여기서 `.*`는 줄의 시작에서 끝까지를 나타내고, 줄마다 각각 별도의 문자열로 인식됩니다.
        var input = "hello\njava\nregex";
        var pattern2 = Pattern.compile("^.*$", Pattern.MULTILINE);
        var matcher2 = pattern2.matcher(input);

        while (matcher2.find()) {
            System.out.println("found: " + matcher2.group());
            //found: hello
            //found: java
            //found: regex
        }
    }

    @RunExample
    public void checkPasswordValid() {
        var password = "Pass123!";
        var regex = "^(?=.*[A-Z])(?=.*\\d)(?=.*[!@#$%^&*]).{8,}$";
        System.out.println("유효한 비밀번호? " + password.matches(regex));
        //유효한 비밀번호? true
    }

    /**
     * `(?!대상산어)`는 바로 뒤에 '대상단어'가 오지 않아야 매칭합니다.
     *
     * <pre>
     * {@code
     * ^((?!대상단어).)*$
     *
     * - 문자열 시작(^)에서 끝($)까지 검사
     * - 매 위치에서 바로 뒤에 '대상단어'가 오는지 검사
     * - 이 조건을 만족하면서 임의 문자(.)가 0번 이상 반복(`*`)돼서 끝(`$`)까지 도달
     * }
     * </pre>
     */
    @RunExample
    public void negativeLookahead() {
        var input = "first_test\nrealreal\nlast_test";
        var pattern = Pattern.compile("^((?!realreal).)*$", Pattern.MULTILINE);
        var matcher = pattern.matcher(input);
        while (matcher.find()) {
            System.out.println("found: " + matcher.group());
            //found: first_test
            //found: last_test
        }
    }
}
