# errno

- [errno](#errno)
    - [네트워크 에러와 리눅스 errno의 관계](#네트워크-에러와-리눅스-errno의-관계)
    - [errno 목록](#errno-목록)
        - [linux errno 리스팅 by Python](#linux-errno-리스팅-by-python)
        - [ubuntu 16.04.6 LTS (Xenial Xerus)](#ubuntu-16046-lts-xenial-xerus)
        - [MacOS](#macos)
    - [`104`, `ECONNRESET`](#104-econnreset)
    - [참고 문서](#참고-문서)

## 네트워크 에러와 리눅스 errno의 관계

1. **시스템 호출과 오류 코드**:

    리눅스에서 네트워크 통신은 시스템 호출을 통해 이루어집니다.

    예를 들어, `connect()`, `send()`, `recv()`와 같은 함수들이 사용됩니다.

    이러한 *시스템 호출이 실패할 경우*, 커널은 표준 오류 코드인 `errno`를 설정하여 실패 원인을 나타냅니다.
    `errno`는 표준 C 라이브러리의 일부로, 다양한 시스템 호출의 실패 원인을 나타내는 데 사용됩니다.

2. **ECONNRESET (errno 104)**:

    `errno 104`는 `ECONNRESET`로 정의되며, 이는 "Connection reset by peer"를 의미합니다.
    이는 클라이언트와 서버 간의 연결이 원격 서버에 의해 강제로 종료되었음을 나타냅니다.

    네트워크 연결이 원격 서버에 의해 강제로 종료될 때, 관련 시스템 호출 (`recv()` 등)이 실패하고, 커널은 `ECONNRESET` 오류를 설정합니다.

3. **방화벽과 네트워크 연결**:

    방화벽이나 호스트 접근 제어 설정으로 인해 네트워크 연결이 차단될 경우, 이는 네트워크 통신 실패로 이어질 수 있습니다.
    클라이언트는 연결 시도 중 연결이 갑자기 종료되었음을 감지하게 되고, 이에 따라 `ECONNRESET` 오류가 발생합니다.

    이러한 오류는 네트워크 연결 실패로 인한 것이므로, 리눅스 커널은 표준 네트워크 오류 코드인 `ECONNRESET`을 설정하여 이를 나타냅니다.

네트워크 연결 문제가 발생할 때, 이는 시스템 호출의 실패로 이어지며, 커널은 `errno`를 통해 오류 코드를 설정합니다.
방화벽이나 호스트 접근 제어로 인해 네트워크 연결이 차단될 때, 클라이언트는 이를 네트워크 연결 실패로 인식하며, 리눅스 커널은 적절한 `errno` 코드를 반환합니다.
이는 네트워크 오류가 리눅스 `errno` 코드로 반환되는 이유입니다.

## errno 목록

> WARN: 리눅스 버전이나 MacOS 사용 여부 등에 따라 코드가 다를 수 있습니다.

### linux errno 리스팅 by Python

```py
import errno; print("\n".join(["{}: {}".format(err_code, errno.errorcode[err_code]) for err_code in sorted(errno.errorcode)]))
```

### ubuntu 16.04.6 LTS (Xenial Xerus)

```bash
1: EPERM
2: ENOENT
3: ESRCH
4: EINTR
5: EIO
6: ENXIO
7: E2BIG
8: ENOEXEC
9: EBADF
10: ECHILD
11: EAGAIN
12: ENOMEM
13: EACCES
14: EFAULT
15: ENOTBLK
16: EBUSY
17: EEXIST
18: EXDEV
19: ENODEV
20: ENOTDIR
21: EISDIR
22: EINVAL
23: ENFILE
24: EMFILE
25: ENOTTY
26: ETXTBSY
27: EFBIG
28: ENOSPC
29: ESPIPE
30: EROFS
31: EMLINK
32: EPIPE
33: EDOM
34: ERANGE
35: EDEADLOCK
36: ENAMETOOLONG
37: ENOLCK
38: ENOSYS
39: ENOTEMPTY
40: ELOOP
42: ENOMSG
43: EIDRM
44: ECHRNG
45: EL2NSYNC
46: EL3HLT
47: EL3RST
48: ELNRNG
49: EUNATCH
50: ENOCSI
51: EL2HLT
52: EBADE
53: EBADR
54: EXFULL
55: ENOANO
56: EBADRQC
57: EBADSLT
59: EBFONT
60: ENOSTR
61: ENODATA
62: ETIME
63: ENOSR
64: ENONET
65: ENOPKG
66: EREMOTE
67: ENOLINK
68: EADV
69: ESRMNT
70: ECOMM
71: EPROTO
72: EMULTIHOP
73: EDOTDOT
74: EBADMSG
75: EOVERFLOW
76: ENOTUNIQ
77: EBADFD
78: EREMCHG
79: ELIBACC
80: ELIBBAD
81: ELIBSCN
82: ELIBMAX
83: ELIBEXEC
84: EILSEQ
85: ERESTART
86: ESTRPIPE
87: EUSERS
88: ENOTSOCK
89: EDESTADDRREQ
90: EMSGSIZE
91: EPROTOTYPE
92: ENOPROTOOPT
93: EPROTONOSUPPORT
94: ESOCKTNOSUPPORT
95: ENOTSUP
96: EPFNOSUPPORT
97: EAFNOSUPPORT
98: EADDRINUSE
99: EADDRNOTAVAIL
100: ENETDOWN
101: ENETUNREACH
102: ENETRESET
103: ECONNABORTED
104: ECONNRESET
105: ENOBUFS
106: EISCONN
107: ENOTCONN
108: ESHUTDOWN
109: ETOOMANYREFS
110: ETIMEDOUT
111: ECONNREFUSED
112: EHOSTDOWN
113: EHOSTUNREACH
114: EALREADY
115: EINPROGRESS
116: ESTALE
117: EUCLEAN
118: ENOTNAM
119: ENAVAIL
120: EISNAM
121: EREMOTEIO
122: EDQUOT
```

### MacOS

```bash
1: EPERM
2: ENOENT
3: ESRCH
4: EINTR
5: EIO
6: ENXIO
7: E2BIG
8: ENOEXEC
9: EBADF
10: ECHILD
11: EDEADLK
12: ENOMEM
13: EACCES
14: EFAULT
15: ENOTBLK
16: EBUSY
17: EEXIST
18: EXDEV
19: ENODEV
20: ENOTDIR
21: EISDIR
22: EINVAL
23: ENFILE
24: EMFILE
25: ENOTTY
26: ETXTBSY
27: EFBIG
28: ENOSPC
29: ESPIPE
30: EROFS
31: EMLINK
32: EPIPE
33: EDOM
34: ERANGE
35: EAGAIN
36: EINPROGRESS
37: EALREADY
38: ENOTSOCK
39: EDESTADDRREQ
40: EMSGSIZE
41: EPROTOTYPE
42: ENOPROTOOPT
43: EPROTONOSUPPORT
44: ESOCKTNOSUPPORT
45: ENOTSUP
46: EPFNOSUPPORT
47: EAFNOSUPPORT
48: EADDRINUSE
49: EADDRNOTAVAIL
50: ENETDOWN
51: ENETUNREACH
52: ENETRESET
53: ECONNABORTED
54: ECONNRESET
55: ENOBUFS
56: EISCONN
57: ENOTCONN
58: ESHUTDOWN
59: ETOOMANYREFS
60: ETIMEDOUT
61: ECONNREFUSED
62: ELOOP
63: ENAMETOOLONG
64: EHOSTDOWN
65: EHOSTUNREACH
66: ENOTEMPTY
67: EPROCLIM
68: EUSERS
69: EDQUOT
70: ESTALE
71: EREMOTE
72: EBADRPC
73: ERPCMISMATCH
74: EPROGUNAVAIL
75: EPROGMISMATCH
76: EPROCUNAVAIL
77: ENOLCK
78: ENOSYS
79: EFTYPE
80: EAUTH
81: ENEEDAUTH
82: EPWROFF
83: EDEVERR
84: EOVERFLOW
85: EBADEXEC
86: EBADARCH
87: ESHLIBVERS
88: EBADMACHO
89: ECANCELED
90: EIDRM
91: ENOMSG
92: EILSEQ
93: ENOATTR
94: EBADMSG
95: EMULTIHOP
96: ENODATA
97: ENOLINK
98: ENOSR
99: ENOSTR
100: EPROTO
101: ETIME
102: EOPNOTSUPP
103: ENOPOLICY
104: ENOTRECOVERABLE
105: EOWNERDEAD
106: EQFULL
```

## `104`, `ECONNRESET`

`errno 104`는 일반적으로 "Connection reset by peer"를 의미하며, 이는 네트워크 연결이 원격 서버에 의해 강제로 종료되었음을 나타냅니다.

`errno 104`는 `ECONNRESET`로도 알려져 있으며, 이는 클라이언트와 서버 간의 네트워크 연결이 원격 서버에 의해 강제로 종료되었음을 의미합니다. 일반적으로 "Connection reset by peer"를 의미합니다. 이는 다양한 이유로 발생할 수 있습니다.
- 네트워크 문제
- 서버의 과부하
- 서버 애플리케이션의 충돌
- 방화벽 설정 문제
- 서버 측에서 의도적으로 연결을 종료하는 경우

리눅스의 `errno` 코드와 관련된 일반적인 정보는 `errno`와 `strerror`를 다루는 매뉴얼 페이지에 나와 있습니다.
하지만 각 오류 코드의 세부적인 설명은 주로 관련 시스템 호출의 매뉴얼 페이지에서 확인할 수 있습니다.

예를 들어, `recv` 시스템 호출은 네트워크 소켓에서 데이터를 읽을 때 사용되며, `errno` 코드의 한 예로 `ECONNRESET`을 다룹니다.

1. **recv(2) 매뉴얼 페이지**:
   - 이 매뉴얼 페이지는 `recv` 시스템 호출에 대한 설명을 제공하며, 네트워크 연결과 관련된 `errno` 코드도 설명합니다.
   - [recv(2) - Linux manual page - man7.org](https://man7.org/linux/man-pages/man2/recv.2.html)

2. **connect(2) 매뉴얼 페이지**:
   - 이 매뉴얼 페이지는 `connect` 시스템 호출에 대한 설명을 제공하며, 네트워크 연결 오류에 대한 정보를 다룹니다.
   - [connect(2) - Linux manual page - man7.org](https://man7.org/linux/man-pages/man2/connect.2.html)

## 참고 문서

- [errno(3) - Linux manual page - man7.org](https://man7.org/linux/man-pages/man3/errno.3.html): errno 코드에 대한 일반적인 설명.
- [recv(2) - Linux manual page - man7.org](https://man7.org/linux/man-pages/man2/recv.2.html): 네트워크에서 데이터를 수신할 때 발생할 수 있는 오류 코드 설명.
- [connect(2) - Linux manual page - man7.org](https://man7.org/linux/man-pages/man2/connect.2.html): 네트워크 연결을 설정할 때 발생할 수 있는 오류 코드 설명.
- [Odd cURL Error - errno 104 (Code 56) In WHMCS](https://stackoverflow.com/questions/8401956/odd-curl-error-errno-104-code-56-in-whmcs)
