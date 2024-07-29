# singleton

## `_internal()` 생성자를 통한 싱글톤 생성

```dart
class HttpUtil {
    // 클래스의 유일한 인스턴스를 저장합니다. 이 변수는 클래스가 처음 로드될 때 초기화됩니다.
    // `HttpUtil` 클래스가 처음으로 사용될 때, 즉 `HttpUtil` 클래스의 어떤 멤버나 메서드가 처음으로 접근될 때,
    // Dart는 클래스의 정적 변수들을 초기화합니다.
    // 이 초기화는 클래스가 처음 로드될 때 한 번만 실행됩니다.
    static final HttpUtil _instance = HttpUtil._internal();
    String contentType = 'application/json; charset=utf-8';

    // 항상 `_instance`를 반환합니다. 이로 인해 클래스의 인스턴스가 하나만 존재하게 됩니다.
    // `_instance` 초기화 후 `HttpUtil()` 생성자가 호출될 때마다 동일한 `_instance`를 반환합니다.
    factory HttpUtil() => _instance;

    late Dio dio;
    CancelToken cancelToken = CancelToken();

    // 프라이빗 생성자로, 외부에서 직접 호출할 수 없습니다. 이 생성자는 클래스 내부에서만 호출될 수 있습니다.
    HttpUtil._internal() {
        BaseOptions options = BaseOptions(
            baseUrl: SERVER_API_URL,
            connectTimeout: const Duration(seconds: 30),
            receiveTimeout: const Duration(seconds: 60),
            headers: {},
            contentType: contentType,
            responseType: ResponseType.json,
        );

        dio = Dio(options);

        dio.httpClientAdapter = IOHttpClientAdapter(
            createHttpClient: () {
                final HttpClient client =
                    HttpClient(context: SecurityContext(withTrustedRoots: false));
                client.badCertificateCallback =
                    ((X509Certificate cert, String host, int port) => true);
                return client;
            },
        );

        CookieJar cookieJar = CookieJar();
        dio.interceptors.add(CookieManager(cookieJar));

        dio.interceptors.addAll([
        if (!isRealServer)
            LogInterceptor(
                request: false,
                requestHeader: false,
                requestBody: false,
                responseHeader: false,
                responseBody: false,
                error: true,
            ),
        ]);
    }
}
```
