# pull requests

- [pull requests](#pull-requests)
    - [stacked PR](#stacked-pr)
    - [이미지 붙여넣기](#이미지-붙여넣기)
        - [header](#header)
        - [payload](#payload)
        - [response](#response)
    - [highlihgt a Note, Warning, etc](#highlihgt-a-note-warning-etc)

## stacked PR

Stacked PR(스택드 풀 리퀘스트)는 큰 코드 변경을 관리하기 쉬운 작은 체인의 PR로 나누는 방법입니다. 이 방식은 개발자가 이전 브랜치의 PR들 위에 새로운 브랜치의 PR을 쌓아 올리면서 작업을 계속할 수 있게 해줍니다. 이는 코드 리뷰 과정을 기다리지 않고도 개발을 지속할 수 있게 하여 개발 흐름을 원활하게 합니다.

스택드 PR은 주로 큰 기능 개발이나 복잡한 코드 변경을 여러 단계로 나누어 관리할 때 사용됩니다. 이 방식은 코드 리뷰를 더 빠르고 쉽게 만들고, 필요한 경우 롤백을 간단하게 할 수 있는 장점이 있습니다. 또한, 한 기능의 개발이 다른 기능의 개발에 의존적인 경우에도 유용하게 사용될 수 있습니다 .

Git CLI를 사용하여 스택드 PR을 관리할 수 있습니다. 예를 들어, GitHub CLI 확장인 `gh pr stack`을 사용하면 스택드 PR을 쉽게 관리할 수 있습니다. 이 외에도 Graphite, Gerrit, Phabricator와 같은 도구들이 스택드 PR을 지원하여 Git의 전통적인 흐름을 보다 효과적으로 관리할 수 있게 해줍니다.

스택드 PR은 개발자가 여러 기능을 동시에 빠르게 개발하고, 리뷰 과정에서 발생할 수 있는 지연을 최소화하며, 코드의 의존성을 효과적으로 관리할 수 있게 해주는 강력한 도구입니다. 이를 통해 팀 전체의 생산성을 향상시킬 수 있습니다.

## 이미지 붙여넣기

맥에서 이미지를 복사하고 PR에 붙여넣기 하면 아래 payload로 요청이 간다

### header

```text
:authority:         github.com
:method:            POST
:path:              /upload/policies/assets
:scheme:            https
Accept:             application/json
Accept-Encoding:    gzip, deflate, br
Accept-Language:    ko,en-US;q=0.9,en;q=0.8
Content-Length:     649
Content-Type:       multipart/form-data; boundary=----WebKitFormBoundaryOZISgtXHV3hNGcRH
Cookie:             _octo=GH1.1.792420184.1696232539; _device_id=a52aff520289dc97b20cb159cdc17fa2; user_session=NmNbdw5vYhr2bwi7ZEzWKlU9xSKaqraa3EkNiWfXglt--jhs; __Host-user_session_same_site=NmNbdw5vYhr2bwi7ZEzWKlU9xSKaqraa3EkNiWfXglt--jhs; color_mode=%7B%22color_mode%22%3A%22light%22%2C%22light_theme%22%3A%7B%22name%22%3A%22light%22%2C%22color_mode%22%3A%22light%22%7D%2C%22dark_theme%22%3A%7B%22name%22%3A%22dark_dimmed%22%2C%22color_mode%22%3A%22dark%22%7D%7D; logged_in=yes; dotcom_user=aimpugn; tz=Asia%2FSeoul; preferred_color_mode=dark; GHCC=Required:1-Analytics:1-SocialMedia:1-Advertising:1; has_recent_activity=1; _gh_sess=LP%2Bwc14SMCoVuQrEELDOALHYthnBhhG1VHs8ck%2FmufAjluFE8FN0acIjVByxkw9HjZH6zMeVck%2BTsa1dNO2ZuLnS4z2yVDT6zZuWXmfLGFfWhZXAITIx16DLixMd2yr4q62I7FgJs19NMHNuq%2BAGT%2FBK8wOq8Z526g3RoE%2Fa8GZA18aVKqpiIAkM%2BevhAtLZnTY9h42pH3BT469EUtUn31eG8no5JLxezV1nlPYPCAnM%2B50oGC%2BCFSFQ6JSbI9y%2F%2BSU2%2BdQi4mpIxtYghhKgm2VmvpA7lfpZMbazNJRO3G9vPjoaTD%2BmHl69n6cZZkHHd%2BNBD1uIOyjeoQTsMlD8N3WMo2oHSTwd3ByT4ROe%2FZ%2BAJBUYK35JGVpJGtJlWi0eupEpftgZRYa2NzMiyy0sNSunpDxMjvMGbBGL%2FQASFL6kGSKrH28wObPLeeAy4PM2Rl2DKYKaOiWzDmhdLwAHnB8in5%2FX9a1QhABgTFUCNPjSdxJeLZbNgqhn%2BsISaxeRSPlJN0nwp2fWREEAjsNPba7WFyjjKdUQ4xvxVsnAv04B1j5PYC14mAn4z7lEQ7wenuZq1VkYsNbZAIkWes1c22X3aNTPaMPoAZYiei0QDal0EGfF0%2Fr%2F3UWQ9rZOFrbs61My%2BA%3D%3D--ZHRioecm3VchhxKl--B6UU8XykyWzFRwAXc0OWIA%3D%3D
Dnt:                1
Origin:             https://github.com
Referer:            https://github.com/some-qwerty-org.io/go/pull/204
Sec-Ch-Ua:          "Chromium";v="121", "Not 
                    A(Brand";v="99"
Sec-Ch-Ua-Mobile:   ?0
Sec-Ch-Ua-Platform: "macOS"
Sec-Fetch-Dest:     empty
Sec-Fetch-Mode:     cors
Sec-Fetch-Site:     same-origin
User-Agent:         Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36
X-Requested-With:   XMLHttpRequest
```

### payload

```text
------WebKitFormBoundaryOZISgtXHV3hNGcRH
Content-Disposition: form-data; name="name"

image.png
------WebKitFormBoundaryOZISgtXHV3hNGcRH
Content-Disposition: form-data; name="size"

214047
------WebKitFormBoundaryOZISgtXHV3hNGcRH
Content-Disposition: form-data; name="content_type"

image/png
------WebKitFormBoundaryOZISgtXHV3hNGcRH
Content-Disposition: form-data; name="authenticity_token"

Kd7lkc0mbTD_tX9dt8zKxgRwv6XL40p0DAJkRqekty2NWQceeVFXYSpExrkUOV-pmhW1Fiz0i94jOAKiUOYSSQ
------WebKitFormBoundaryOZISgtXHV3hNGcRH
Content-Disposition: form-data; name="repository_id"

613157971
------WebKitFormBoundaryOZISgtXHV3hNGcRH--
```

### response

```json
{
    "upload_url": "https://github-production-user-asset-6210df.s3.amazonaws.com",
    "header": {},
    "asset": {
        "id": 306989452,
        "name": "image.png",
        "size": 214047,
        "content_type": "image/png",
        "href": "https://github.com/some-qwerty-org.io/go/assets/28570432/e23a8d51-c2be-4040-89c5-4280f8fa4961",
        "original_name": "image.png"
    },
    "form": {
        "key": "28570432/306989452-e23a8d51-c2be-4040-89c5-4280f8fa4961.png",
        "acl": "private",
        "policy": "eyJleHBpcmF0aW9uIjoiMjAyNC0wMi0yMlQxMjo0MTo0NloiLCJjb25kaXRpb25zIjpbeyJidWNrZXQiOiJnaXRodWItcHJvZHVjdGlvbi11c2VyLWFzc2V0LTYyMTBkZiJ9LHsia2V5IjoiMjg1NzA0MzIvMzA2OTg5NDUyLWUyM2E4ZDUxLWMyYmUtNDA0MC04OWM1LTQyODBmOGZhNDk2MS5wbmcifSx7ImFjbCI6InByaXZhdGUifSxbImNvbnRlbnQtbGVuZ3RoLXJhbmdlIiwyMTQwNDcsMjE0MDQ3XSx7IngtYW16LWNyZWRlbnRpYWwiOiJBS0lBVkNPRFlMU0E1M1BRSzRaQS8yMDI0MDIyMi91cy1lYXN0LTEvczMvYXdzNF9yZXF1ZXN0In0seyJ4LWFtei1hbGdvcml0aG0iOiJBV1M0LUhNQUMtU0hBMjU2In0seyJ4LWFtei1kYXRlIjoiMjAyNDAyMjJUMDAwMDAwWiJ9LHsiQ29udGVudC1UeXBlIjoiaW1hZ2UvcG5nIn0seyJDYWNoZS1Db250cm9sIjoibWF4LWFnZT0yNTkyMDAwIn0seyJ4LWFtei1tZXRhLVN1cnJvZ2F0ZS1Db250cm9sIjoibWF4LWFnZT0zMTU1NzYwMCJ9XX0=",
        "X-Amz-Algorithm": "AWS4-HMAC-SHA256",
        "X-Amz-Credential": "AKIAVCODYLSA53PQK4ZA/20240222/us-east-1/s3/aws4_request",
        "X-Amz-Date": "20240222T000000Z",
        "X-Amz-Signature": "b2b781dbf3aeb6e5fba882625b3ac894f8f8c008a526cfee72473a89fed3d0ec",
        "Content-Type": "image/png",
        "Cache-Control": "max-age=2592000",
        "x-amz-meta-Surrogate-Control": "max-age=31557600"
    },
    "same_origin": false,
    "asset_upload_url": "/upload/assets/306989452",
    "upload_authenticity_token": "ugTiYCUH8QH-mfMmEfyfOebcnQqBqxgc11W2NaZxINgfATHETkdbZgqfhtRd1PrLvQC8oTmC1V52FT4sXMDBGw",
    "asset_upload_authenticity_token": "9r3jh2OwlS6hJPCb-WTSi9eeKAlugJFHGedY0lgUKdLMFv8t_p1W6VH9-aKadBbvvDYpPPmO0c5CWpQ1rBbBSA"
}
```

## [highlihgt a Note, Warning, etc](https://github.com/orgs/community/discussions/16925)

```md
> [!NOTE]  
> Highlights information that users should take into account, even when skimming.

> [!TIP]
> Optional information to help a user be more successful.

> [!IMPORTANT]  
> Crucial information necessary for users to succeed.

> [!WARNING]  
> Critical content demanding immediate user attention due to potential risks.

> [!CAUTION]
> Negative potential consequences of an action.
```
