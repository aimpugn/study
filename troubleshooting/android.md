# Android Troubleshooting

- [Android Troubleshooting](#android-troubleshooting)
    - [현대카드만 결제가 안됨](#현대카드만-결제가-안됨)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)

## 현대카드만 결제가 안됨

### 문제

- 현대카드만 앱카드 이동 안되는것 같음

```xml
<!-- AndroidManifest.xml -->

<package android:name="com.hyundaicard.appcard" /> <!--현대 앱카드-->
```

- "앱 설치가 필요하신가요" 누르면 플레이스토어로 전환 된다
- 현대카드는 다른카드사와 달리 앱 실행하면 pass마냥 모달로 떠서 그런것 같음

adb 로그

```log
Unable to find resource: https://ansimclick.hyundaicard.com/mobile3/MBITFX501.jsp;jsessionid=X44t2rskPlUTKuQQtwJ1ayINP4a4P58z1RMIbWZALhI6aBSWnztGwCidI0q6k5kz.dpacap12_servlet_re-xacs21

showWebPage(intent:hdcardappcardansimclick://appcard?acctid=202311071640019940067226177131#Intent;package=com.hyundaicard.appcard;end;, true, false, HashMap)

11-07 16:40:48.737 11793 11793 E CordovaWebViewImpl: Error loading url intent:hdcardappcardansimclick://appcard?acctid=202311071640019940067226177131#Intent;package=com.hyundaicard.appcard;end;
11-07 16:40:48.737 11793 11793 E CordovaWebViewImpl: android.content.ActivityNotFoundException: No Activity found to handle Intent { act=android.intent.action.VIEW cat=[android.intent.category.BROWSABLE] dat=intent: }
```

콜도바 코드 보니까
`intent:hdcardappcardansimclick://appcard?acctid=202311071640019940067226177131#Intent;package=com.hyundaicard.appcard;end;` 라고 쓰면 파싱을 못하고
정확히
`intent://hdcardappcardansimclick://appcard?acctid=202311071640019940067226177131#Intent;package=com.hyundaicard.appcard;end;` 라고 써야만 파싱되게 되어있는 거 같음 (맨앞부분 :// 참고)

### 원인

```xml
<allow-intent href="intent:*"/>
```

이렇게만 되어 있으면 실행이 안 된다. `allow-intent` 설정을 아예 안한 것으로 보임

### 해결

아래 내용을 빼고

```xml
<allow-navigation href="*" />
```

아래 내용을 추가하면 현대카드 실행된다

```xml
<!-- 이 부분을 추가하면 정상 작동 -->
<allow-intent href="intent:hdcardappcardansimclick:*"/>

<!-- 아래 부분은 앱에 잘 설정 되어 있을 것 -->
<allow-navigation href="http://*" />
<allow-navigation href="https://*" />
```
