use std::ffi::CString;
use std::ptr;
use std::time::Instant;

/// CGContext 타입 구조체를 가리키는 포인터
type CGContextRef = *mut libc::c_void;

type CGFloat = std::ffi::c_double;

// Objective-C 런타임 함수 선언
//
// Rust FFI(Foreign Function Interface) 통해 MacOS의 시스템 프레임워크를 사용할 경우,
// 링커에게 해당 프레임워크를 링크하도록 명시적으로 지시해야 합니다.
// ```
// #[link(name = "FrameworkName", kind = "framework")]
// ```
//
// 이를 링크하지 않으면 아래와 같은 에러가 발생합니다.
// ```
// error: linking with `cc` failed: exit status: 1
// ```
//
// MacOS에서 Objective-C 런타임 함수들은 `libobjc` 라이브러리에 들어있습니다.
#[link(name = "objc", kind = "dylib")]
extern "C" {
    /// 지정된 클래스의 클래스 정의를 리턴합니다.
    ///
    /// References:
    /// - https://developer.apple.com/documentation/objectivec/1418952-objc_getclass?language=objc
    fn objc_getClass(name: *const i8) -> CGContextRef;

    /// Objective C 런타임 시스템에 메서드를 등록하고, 메서드를 등록된 메서드를 셀렉터와 매핑하고, 셀렉터 값을 반환합니다.
    /// 반환되는 [`SEL`](https://developer.apple.com/documentation/objectivec/sel?language=objc) 타입의 포인터는 명명된 메서드의 셀렉터를 가리킵니다.
    /// `SEL` 타입은 메서드 셀렉터를 나타내는 불투명(opaque)한 타입입니다. 실행할 메서드의 이름을 나타내는 데 사용됩니다.
    ///
    /// References:
    /// - https://developer.apple.com/documentation/objectivec/1418557-sel_registername?language=objc
    fn sel_registerName(name: *const i8) -> CGContextRef;

    /// 클래스 인스턴스에 메시지를 보내고 반환 값을 반환합니다.
    /// ```
    /// [view setHidden:YES];
    /// // 아래와 같이 변환됩니다.
    /// // objc_msgSend(view, @selector(setHidden:), YES);
    ///
    /// BOOL isHidden = [view isHidden];
    /// // 아래와 같이 변환됩니다.
    /// // BOOL isHidden = (BOOL)objc_msgSend(view, @selector(isHidden));
    /// ```
    ///
    /// References:
    /// - https://developer.apple.com/documentation/objectivec/1456712-objc_msgsend
    fn objc_msgSend(id: CGContextRef, sel: CGContextRef, ...) -> CGContextRef;

    /// 클래스 인스턴스의 부모 클래스에 메시지를 보내고 반환 값을 반환합니다.
    /// 상속 관계 구조에서 상위 클래스 초기화를 호출해야 하는 경우 사용됩니다.
    ///
    /// References:
    /// - https://developer.apple.com/documentation/objectivec/1456716-objc_msgsendsuper/
    fn objc_msgSendSuper(sup: *mut SuperObject, sel: CGContextRef, ...) -> CGContextRef;

    /// 새로운 클래스와 메타 클래스를 생성합니다.
    /// 클래스 생성 후 [`class_addMethod`](https://developer.apple.com/documentation/objectivec/1418901-class_addmethod?language=objc), 그리고
    /// [`class_addIvar`](https://developer.apple.com/documentation/objectivec/1418756-class_addivar?language=objc) 같은 함수들을 사용하여
    /// 클래스의 속성을 설정합니다.
    ///
    /// References:
    /// - https://developer.apple.com/documentation/objectivec/1418559-objc_allocateclasspair?language=objc
    fn objc_allocateClassPair(
        superclass: CGContextRef,
        name: *const i8,
        extra_bytes: usize,
    ) -> CGContextRef;

    ///
    /// References:
    /// - https://developer.apple.com/documentation/objectivec/1418901-class_addmethod?language=objc
    fn class_addMethod(
        cls: CGContextRef,
        name: CGContextRef,
        imp: extern "C" fn(_this: CGContextRef, _cmd: CGContextRef, _rect: NSRect),
        types: *const i8,
    ) -> bool;
    fn objc_registerClassPair(cls: CGContextRef);
}

// Rust에서 MacOS의 Core Graphic 프레임워크 함수를 FFI로 사용하기 위해 링크합니다.
// - `AppKit`:
//   MacOS 애플리케이션 개발 위한 고수준 UI 프레임워크로 창, 버튼, 텍스트 필드 등의 요소들을 관리합니다.
//   내부적으로 `CoreGraphics`를 사용하여 그래픽 렌더링과 같은 작업을 처리합니다.
//
// - `CoreGraphics`
//   저수준 그래픽 API로, 벡터 기반의 그래픽 작업을 제공합니다.
//   선, 사각형, 텍스트 렌더링, PDF 등 그래픽 작업을 처리합니다.
//   창에 선을 그리거나 색상을 설정하는 등, UI 컨트롤이 아닌 그래픽 렌더링 작업을 처리하려면 `CoreGraphics`가 필요합니다.
#[link(name = "AppKit", kind = "framework")]
#[link(name = "CoreGraphics", kind = "framework")]
extern "C" {
    /// 현재 스트로크의 색상을 DeviceRGB 색상 공간의 값으로 설정합니다.
    ///
    /// - `red`: 0.0 (zero intensity)부터 1.0 (full intensity)까지의 값
    /// - `green`: 0.0 (zero intensity)부터 1.0 (full intensity)까지의 값
    /// - `blue`: 0.0 (zero intensity)부터 1.0 (full intensity)까지의 값
    ///
    /// References:
    /// - https://developer.apple.com/documentation/coregraphics/1456378-cgcontextsetrgbstrokecolor/
    fn CGContextSetRGBStrokeColor(
        context: CGContextRef,
        red: CGFloat,
        green: CGFloat,
        blue: CGFloat,
        alpha: CGFloat,
    );

    /// 지정한 지점에서 새 하위 경로를 시작합니다.
    /// 즉, 선이 시작되는 시작점(`x`, `y`)을 지정합니다.
    ///
    /// References:
    /// - https://developer.apple.com/documentation/coregraphics/1454738-cgcontextmovetopoint/
    fn CGContextMoveToPoint(context: CGContextRef, x: CGFloat, y: CGFloat);

    /// 현재 점부터 지정된 점까지 직선을 추가합니다.
    ///
    /// References:
    /// - https://developer.apple.com/documentation/coregraphics/1455213-cgcontextaddlinetopoint
    fn CGContextAddLineToPoint(context: CGContextRef, x: CGFloat, y: CGFloat);

    /// 현재 경로를 따라 선을 그립니다.
    ///
    /// References:
    /// - https://developer.apple.com/documentation/coregraphics/1454490-cgcontextstrokepath
    fn CGContextStrokePath(context: CGContextRef);

    /// 현재 경로에 원의 호(arc of circle)를 추가합니다.
    ///
    /// References:
    /// - https://developer.apple.com/documentation/coregraphics/1455756-cgcontextaddarc
    fn CGContextAddArc(
        context: CGContextRef,
        x: CGFloat,
        y: CGFloat,
        radius: CGFloat,
        start_angle: CGFloat,
        end_angle: CGFloat,
        clockwise: libc::c_int,
    );
}

/// 다음과 같은 오류가 발생하여 확인해 보니, 단순히 [objc_msgSend]를 호출하는 것은 위험하고 문제가 발생할 가능성이 높습니다.
///
/// ```
/// 2024-12-11 17:50:05.688 draw_lines_macos[78947:64669842] NSWindow does not support nonactivating panel styleMask 0x80
/// 2024-12-11 17:50:05.688 draw_lines_macos[78947:64669842] NSUnscaledWindowMask is deprecated and will be ignored.
/// 2024-12-11 17:50:05.688 draw_lines_macos[78947:64669842] NSWindow does not support HUD styleMask 0x2000; use NSPanel instead.
/// ```
///
/// 일단 Rust FFI(Foreign Function Interface)는 함수 인자 전달과 반환 값 처리 방식을 정의하는 C ABI(Application Binary Interface)와 호환됩니다.
///
/// `#[link(name = "objc", kind = "dylib")]`를 통해 Objective-C 런타임을 연결하면, 런타임의 `objc_msgSend`를 호출할 수 있습니다.
/// 이때 Rust는 [`cdecl` 호출 규약](https://en.wikipedia.org/wiki/X86_calling_conventions#cdecl)을 사용하여 호출합니다.
///
/// 그런데 문제는 [objc_msgSend]는 가변 인자 `...`를 처리하지만 Rust의 FFI에서 `...`는 기본적으로 지원하지 않습니다.
/// Rust는 `...`를 정적으로 해석하며, 런타임 시의 메서드 시그니처를 참조하지 않습니다.
///
/// 가변 인자를 직접 호출할 때 Rust는 전달된 인자를 C ABI 규칙에 따라 배치하지만,
/// Objective-C 런타임에서는 메서드 시그니처를 기반으로 인자를 런타임에 동적으로 처리합니다.
/// Rust는 이를 알지 못하고 정확한 시그니처 정보가 없으므로, 컴파일러는 C ABI 규칙에 따라 매개변수를 스택이나 레지스터에 배치합니다.
/// C 가변 인자 함수에서는 다음과 같은 규칙이 적용됩니다:
/// - 작은 정수 타입(`char`, `short`)은 `int`로 프로모션됩니다.
/// - `float`는 `double`로 프로모션됩니다.
/// - 매개변수의 개수와 타입은 호출자가 컴파일 타임에 결정하며, 함수는 이를 정적으로 처리합니다.
///
/// Rust는 C ABI 규칙을 따르므로, 작은 정수(`u8`, `i8`, `u16`, `i16`)가 자동으로 `int`로 프로모션될 수 있습니다.
/// 이로 인해 잘못된 레지스터나 스택 위치에 인자를 배치할 가능성이 있고,
/// 인자 타입이나 개수, 정렬 규약에 따라 레지스터나 스택 할당이 어긋나면 세그멘테이션 폴트가 발생할 수 있습니다.
///
/// 하지만 함수 포인터 [InitWindowFn]라는 타입을 정의할 경우 Rust는 인자의 타입, 개수, 순서를 명확히 알 수 있고,
/// 이를 통해 이는 Objective-C 메서드의 호출 규약과 일치하게 만들 수 있다고 합니다.
///
/// 따라서 [objc_msgSend]를 아래와 같이 그대로 사용하는 대신
/// ```
/// let window = objc_msgSend(window, init_sel, rect, style_mask, backing, defer_bool);
/// ```
///
/// 정확한 함수 시그니처를 가진 함수 포인터로 [std::mem::transmute] 한 뒤 호출하는 것이 좋다고 합니다.
/// ```
/// let init_window_fn: InitWindowFn = std::mem::transmute(objc_msgSend as CGContextRef);
/// let window = init_window_fn(window, init_sel, rect, style_mask, backing, defer_bool);
/// ```
///
/// 이렇게 하면 Rust 컴파일러가 해당 함수 포인터 시그니처에 맞게 인자를 배치하므로 ABI 불일치를 방지할 수 있습니다.
///
/// References:
/// - [Rust Book - FFI](https://doc.rust-lang.org/nomicon/ffi.html)
/// - [Rust Book - FFI - Foreign Calling conventions](https://doc.rust-lang.org/nomicon/ffi.html#foreign-calling-conventions)
/// - [cpp - Default argument promotions](https://en.cppreference.com/w/c/language/conversion#Default_argument_promotions)
type InitWindowFn = unsafe extern "C" fn(
    CGContextRef, // window (id)
    CGContextRef, // _cmd (SEL)
    NSRect,
    u64, // styleMask (NSWindowStyleMask는 NSUInteger)
    u64, // backing (NSBackingStoreType도 NSUInteger)
    u8,  // defer (BOOL는 signed char, u8 사용 가능)
) -> CGContextRef;

type SetContentViewFn =
    unsafe extern "C" fn(id: CGContextRef, sel: CGContextRef, view: CGContextRef);

// NSRect 등 구조체 정의
/// 직사각형 구조체.
///
/// [`repr`](https://doc.rust-lang.org/nomicon/repr-rust.html)은 메모리 표형 방식(representation)을 의미합니다.
/// Rust는 기본적으로 메모리 정렬(alignment)을 최적화하기 위해 구조체 필드의 순서나 패딩을 재배치할 수 있지만,
/// `#[repr(C)]`는 이러한 최적화를 방지하고, 필드 순서를 정의된 대로 유지하며, C의 레이아웃 규칙을 강제합니다.
///
/// 가령 'C 코드에서 정의된 구조체 또는 열거형'과 'Rust에서 정의된 타입'이 동일한 메모리 레이아웃을 가지게 하여,
/// 포인터로 데이터를 교환하거나 함수 호출 시 타입 간의 호환성을 보장합니다.
///
/// ```
/// struct A {
///     a: u8,
///     b: u32,
///     c: u16,
/// }
/// // Rust 컴파일러는 성능 최적화를 위해 필드 순서를 재배치할 수 있습니다.
/// // 예: b -> c -> a (정확한 순서는 컴파일러에 따라 다를 수 있음)
/// ```
/// 기본 타입을 각각의 크기에 맞게 정렬하는 대상에 대해 32 비트로 정렬됩니다.
///
/// ```
/// struct A {
///     a: u8,
///     _pad1: [u8; 3], // to align `b`
///     // 여기까지 32 비트
///     b: u32,
///     // 여기까지 32 비트
///     c: u16,
///     _pad2: [u8; 2], // to make overall size multiple of 4
///     // 여기까지 32 비트
/// }
/// ```
/// 이렇게 C 메모리 정렬 규칙에 따라 필드 배치를 수정합니다.
///
/// References:
/// - https://developer.apple.com/documentation/foundation/nsrect?language=objc
/// - https://github.com/servo/core-foundation-rs/blob/8c71d0f34f7586a049f02b3ffcb7a6bc20a9d9d4/cocoa-foundation/src/foundation.rs#L92-L97
/// - `repr` 관련
///   - https://developer.apple.com/documentation/coregraphics/cgcontextref?language=objc
///   - https://doc.rust-lang.org/nomicon/other-reprs.html#reprc
#[repr(C)]
#[derive(Copy, Clone)]
struct NSRect {
    pub origin: NSPoint,
    pub size: NSSize,
}

/// - https://developer.apple.com/documentation/foundation/nspoint?language=objc
/// - https://github.com/servo/core-foundation-rs/blob/8c71d0f34f7586a049f02b3ffcb7a6bc20a9d9d4/cocoa-foundation/src/foundation.rs#L67-L72
#[repr(C)]
#[derive(Copy, Clone)]
struct NSPoint {
    pub x: CGFloat,
    pub y: CGFloat,
}

/// - https://developer.apple.com/documentation/foundation/nssize?language=objc
/// - https://github.com/servo/core-foundation-rs/blob/8c71d0f34f7586a049f02b3ffcb7a6bc20a9d9d4/cocoa-foundation/src/foundation.rs#L74-L79
#[repr(C)]
#[derive(Copy, Clone)]
struct NSSize {
    pub width: CGFloat,
    pub height: CGFloat,
}

#[repr(C)]
#[derive(Copy, Clone)]
struct SuperObject {
    receiver: CGContextRef,
    super_class: CGContextRef,
}

/// [CString::into_raw]를 통해 [CString]의 소유권을 포기하고, 내부 버퍼의 포인터를 반환합니다.
/// 메모리는 여전히 Rust 힙에 존재하지만, 소유권을 포기했기 때문에 Rust의 자동 메모리 해제 시스템에 의해 관리되지 않습니다.
///
/// 따라서 호출자가 [CString::from_raw]를 사용하여 메모리 해제를 명시적으로 해제해야 합니다.
/// 그렇지 않으면 메모리 누수가 발생할 수 있습니다.
///
/// 이 포인터는 FFI를 통해 Objective-C 코드에서 사용될 수 있습니다.
fn c_string_ptr(s: &str) -> *mut std::ffi::c_char {
    let cstring = CString::new(s).expect("CString creation failed");
    let raw_ptr = cstring.into_raw();
    raw_ptr
}

/// 직사각형에 선을 그리기를 시작합니다.
#[allow(dead_code)]
fn start_draw_line_in_rect() {
    unsafe {
        // 애플리케이션 오브젝트 생성:
        // 앱의 메인 이벤트 루프와 앱의 모든 오브젝트가 사용하는 자원을 관리합니다.
        // - https://developer.apple.com/documentation/appkit/nsapplication?language=objc
        let ns_application_class = objc_getClass(c_string_ptr("NSApplication"));
        let shared_app_sel = sel_registerName(c_string_ptr("sharedApplication"));
        let app = objc_msgSend(ns_application_class, shared_app_sel);

        // 화면에 앱이 보여주는 윈도우:
        // https://developer.apple.com/documentation/appkit/nswindow?language=objc
        let ns_window_class = objc_getClass(c_string_ptr("NSWindow"));
        // 주어진 클래스의 새로운 인스턴스를 반환합니다.
        // 부모인 `NSObject`를 통해 사용할 수 있습니다.
        // https://developer.apple.com/documentation/objectivec/nsobject/1571958-alloc?language=objc
        let alloc_sel = sel_registerName(c_string_ptr("alloc"));

        // Objective C에서 객체 생성 시 `alloc`으로 객체를 할당하여 메모리 확보하고, `init*`으로 초기화하여 사용할 준비를 합니다.
        //
        // ```
        // NSWindow *window = [NSWindow alloc];
        // window = [window initWithContentRect:styleMask:backing:defer:];
        // ```
        // 여기서는 `alloc`으로 윈도우 객체에 대한 메모리를 확보합니다.
        let window = objc_msgSend(ns_window_class, alloc_sel);
        let rect = NSRect {
            origin: NSPoint { x: 100.0, y: 100.0 },
            size: NSSize {
                width: 400.0,
                height: 300.0,
            },
        };
        // [styleMask](https://developer.apple.com/documentation/appkit/nswindow/stylemask-swift.struct?language=objc)
        // - NSWindowStyleMaskTitled (0x1): 제목 표시줄
        // - NSWindowStyleMaskClosable (0x2): 닫기 버튼
        // - NSWindowStyleMaskResizable (0x8): 창 크기 조정 가능
        // 참고
        // - https://github.com/servo/core-foundation-rs/blob/8c71d0f34f7586a049f02b3ffcb7a6bc20a9d9d4/cocoa/src/appkit.rs#L226-L238
        let style_mask = 1u64 << 0  // .titled
            | 1u64 << 1            // .closable
            | 1u64 << 3; // .resizable

        // Objective C 런타임에서 `objc_msgSend`는 C 함수로 정의되어 있고,
        // 이를 Rust에서 호출하기 위해 FFI 통해 적절한 함수 시그니처로 캐스팅합니다.
        let init_window_fn: InitWindowFn = std::mem::transmute(objc_msgSend as CGContextRef);
        // https://github.com/servo/core-foundation-rs/blob/8c71d0f34f7586a049f02b3ffcb7a6bc20a9d9d4/cocoa/src/appkit.rs#L1370-L1381
        let window = init_window_fn(
            window,
            sel_registerName(c_string_ptr("initWithContentRect:styleMask:backing:defer:")), // https://developer.apple.com/documentation/appkit/nswindow/init(contentrect:stylemask:backing:defer:)?language=objc
            rect,
            style_mask,
            2,   // NSBackingStoreBuffered
            0u8, // defer: No
        );

        // 커스텀 뷰 클래스 생성
        // 클래스 등록 전이나 등록 후의 메서드 추가 시퀀스가 잘못되면 문제가 생길 수 있습니다.
        // 1. `objc_allocateClassPair`로 클래스 생성
        // 2. `class_addMethod`로 메서드 추가
        // 3. 마지막에 `objc_registerClassPair` 호출
        let ns_view_class = objc_getClass(c_string_ptr("NSView"));
        let my_view_class = objc_allocateClassPair(ns_view_class, c_string_ptr("MyCustomView"), 0);

        // drawRect: 추가
        // 타입 인코딩 확인: drawRect:(NSRect) -> v@:{NSRect={NSPoint=dd}{NSSize=dd}}
        class_addMethod(
            my_view_class,
            sel_registerName(c_string_ptr("drawRect:")),
            draw_line_in_rect,
            c_string_ptr("v@:{NSRect={NSPoint=dd}{NSSize=dd}}"),
        );
        objc_registerClassPair(my_view_class);

        // 뷰 인스턴스 생성 (alloc)
        let mut my_view_obj = objc_msgSend(my_view_class, alloc_sel) as CGContextRef;

        // super initWithFrame:
        // initWithFrame:(NSRect) -> @:@{NSRect={NSPoint=dd}{NSSize=dd}}
        let init_with_frame_sel = sel_registerName(c_string_ptr("initWithFrame:"));

        let mut sup = SuperObject {
            receiver: my_view_obj,
            super_class: ns_view_class,
        };

        my_view_obj = objc_msgSendSuper(&mut sup, init_with_frame_sel, rect);

        // 윈도우에 뷰 설정
        let set_content_view_fn: SetContentViewFn =
            std::mem::transmute(objc_msgSend as CGContextRef);
        set_content_view_fn(
            window,
            sel_registerName(c_string_ptr("setContentView:")),
            my_view_obj,
        );

        // 윈도우 표시
        let sel_make_key_and_order_front = sel_registerName(c_string_ptr("makeKeyAndOrderFront:"));
        objc_msgSend(
            window,
            sel_make_key_and_order_front,
            ptr::null::<libc::c_void>(),
        );

        // 앱 실행
        let run_sel = sel_registerName(c_string_ptr("run"));
        objc_msgSend(app, run_sel);
    }
}

/// `drawRect:`를 구현합니다.
///
/// `_this`는 현재 `MyCustomView` 객체를 나타내며, 해당 객체의 데이터나 상태를 담고 있습니다.
///
/// `rect`는 뷰의 경계(bounds) 내에서 다시 그려야(업데이트) 할 영역을 정의합니다.
/// 뷰가 처음으로 그려질 때 이 직사각형은 일반적으로 뷰의 보이는 전체 경계를 지정합니다.
/// 그러나 이후의 그리기 작업에서는 이 직사각형이 뷰의 일부만 지정할 수도 있습니다.
///
/// References:
/// - https://developer.apple.com/documentation/uikit/uiview/draw(_:)?language=objc
extern "C" fn draw_line_in_rect(_this: CGContextRef, _cmd: CGContextRef, rect: NSRect) {
    unsafe {
        // 실제로 그리기 위해서는 NSGraphicsContext 컨텍스트를 사용해야 합니다.
        let current_context = objc_msgSend(
            objc_getClass(c_string_ptr("NSGraphicsContext")),
            sel_registerName(c_string_ptr("currentContext")),
        );

        let cg_context = objc_msgSend(
            current_context,
            sel_registerName(c_string_ptr("graphicsPort")),
        ) as CGContextRef;

        let length = 200.0;
        let length_half = length / 2.0;
        // 400 / 2 = 200 - 100
        let start_x = rect.size.width / 2.0 - length_half;
        let start_y = rect.size.height / 2.0;

        // 선 색상 설정
        CGContextSetRGBStrokeColor(cg_context, 0.7, 0.7, 0.7, 1.0);
        // 선이 시작할 지점 설정
        CGContextMoveToPoint(cg_context, start_x, start_y);
        // 어떤 점까지 선을 이을 것인지 설정
        CGContextAddLineToPoint(cg_context, start_x + length, start_y);
        // 선 그리기
        CGContextStrokePath(cg_context);
    }
}

/// - https://developer.apple.com/documentation/foundation/nstimer/1412416-scheduledtimerwithtimeinterval
#[allow(dead_code)]
type ScheduledTimerFn = unsafe extern "C" fn(
    CGContextRef, // class (id)
    CGContextRef, // selector (_cmd)
    f64,          // NSTimeInterval: The number of seconds between firings of the timer.
    CGContextRef, // target (id)
    CGContextRef, // selector (SEL)
    CGContextRef, // userInfo (id)
    i8,           // repeats (BOOL): YES 경우 무효화될 때까지 반복적으로 다시 스케쥴 합니다.
) -> CGContextRef;

/// 직사각형에 원 그리기를 시작합니다.
#[allow(dead_code)]
fn start_draw_animated_circle_in_rect() {
    unsafe {
        // NSApplication 초기화
        let ns_application_class = objc_getClass(c_string_ptr("NSApplication"));
        let sel_shared_application = sel_registerName(c_string_ptr("sharedApplication"));
        let app = objc_msgSend(ns_application_class, sel_shared_application);

        // NSWindow 생성

        let window = objc_msgSend(
            objc_getClass(c_string_ptr("NSWindow")),
            sel_registerName(c_string_ptr("alloc")),
        );

        let rect = NSRect {
            origin: NSPoint { x: 100.0, y: 100.0 },
            size: NSSize {
                width: 400.0,
                height: 400.0,
            },
        };
        let style_mask = 1u64 << 0  // .titled
            | 1u64 << 1            // .closable
            | 1u64 << 3; // .resizable
        let init_window_fn: InitWindowFn = std::mem::transmute(objc_msgSend as CGContextRef);
        let window = init_window_fn(
            window,
            sel_registerName(c_string_ptr("initWithContentRect:styleMask:backing:defer:")),
            rect,
            style_mask,
            2,   // NSBackingStoreBuffered
            0u8, // defer: No
        );

        // 커스텀 뷰 클래스 생성
        let ns_view_class = objc_getClass(c_string_ptr("NSView"));
        let my_view_class =
            objc_allocateClassPair(ns_view_class, c_string_ptr("MyAnimatedCircleView"), 0);

        // drawRect: 메서드 추가
        class_addMethod(
            my_view_class,
            sel_registerName(c_string_ptr("drawRect:")),
            draw_animated_circle_in_rect,
            c_string_ptr("v@:{NSRect={NSPoint=dd}{NSSize=dd}}"),
        );
        objc_registerClassPair(my_view_class);

        // 뷰 인스턴스 생성
        let mut my_view_obj =
            objc_msgSend(my_view_class, sel_registerName(c_string_ptr("alloc"))) as CGContextRef;

        // 뷰 인스턴스의 부모 클래스 초기화
        let mut sup = SuperObject {
            receiver: my_view_obj,
            super_class: ns_view_class,
        };
        my_view_obj = objc_msgSendSuper(
            &mut sup,
            sel_registerName(c_string_ptr("initWithFrame:")),
            rect,
        );

        // 윈도우에 뷰 설정
        let set_content_view_fn: SetContentViewFn =
            std::mem::transmute(objc_msgSend as CGContextRef);
        set_content_view_fn(
            window,
            sel_registerName(c_string_ptr("setContentView:")),
            my_view_obj,
        );

        // NSTimer 설정하여 redraw 호출
        let scheduled_timer_fn: ScheduledTimerFn =
            std::mem::transmute(objc_msgSend as CGContextRef);

        // NSTimer를 생성하여 drawFrame:을 주기적으로 호출
        scheduled_timer_fn(
            objc_getClass(c_string_ptr("NSTimer")),
            sel_registerName(c_string_ptr(
                "scheduledTimerWithTimeInterval:target:selector:userInfo:repeats:",
            )),
            1.0 / 10.0,  // time interval: 10 FPS
            my_view_obj, // target
            // `needsDisplay` 속성은 `NSView` 오브젝트를 다시 그릴 것인지 결정하며,
            // `setNeedsDisplay`를 호출하여 뷰가 다시 그려져야 함을 시스템에 알리는 데 사용됩니다.
            // 이 메서드를 호출하면, 시스템은 해당 뷰를 "더티(dirty)" 상태로 표시하고,
            // 다음 렌더링 주기에서 drawRect: 메서드를 호출하여 전체 뷰를 다시 그립니다.
            sel_registerName(c_string_ptr("setNeedsDisplay:")),
            ptr::null::<libc::c_void>() as CGContextRef, // userInfo
            1,                                           // BOOL (true)
        );

        // 윈도우 표시
        objc_msgSend(
            window,
            sel_registerName(c_string_ptr("makeKeyAndOrderFront:")),
            ptr::null::<libc::c_void>(),
        );

        // 이벤트 루프 실행
        objc_msgSend(app, sel_registerName(c_string_ptr("run")));
    }
}

/// 원을 그리는 `drawRect:` 구현
extern "C" fn draw_animated_circle_in_rect(_this: CGContextRef, _cmd: CGContextRef, rect: NSRect) {
    // 프로그램 실행 이후 애니메이션 시작 시점을 나타냅니다.
    // 함수 호출 간 상태를 유지하도록 시작된 시점을 고정하기 위해 `static`으로 선언합니다.
    static START_TIME: std::sync::OnceLock<Instant> = std::sync::OnceLock::new();

    // 애니메이션 시작 시간을 초기화
    let start_time = START_TIME.get_or_init(Instant::now);
    let elapsed_time = start_time.elapsed().as_secs_f64();

    unsafe {
        let current_context = objc_msgSend(
            objc_getClass(c_string_ptr("NSGraphicsContext")),
            sel_registerName(c_string_ptr("currentContext")),
        );

        let cg_context = objc_msgSend(
            current_context,
            sel_registerName(c_string_ptr("graphicsPort")),
        ) as CGContextRef;

        // 원의 중심
        let center_x = rect.size.width / 2.0;
        let center_y = rect.size.height / 2.0;
        // 반지름
        let radius = 100.0;

        // 시간에 따라 변하도록 색상 계산
        // sin, cos는 서로 다른 주기로 -1 ~ 1 값을 오갑니다.
        // - https://justinparrtech.com/JustinParr-Tech/spectrum-generating-color-function-using-sine-waves/
        let elapsed_time_sin = elapsed_time.sin();
        let elapsed_time_cos = elapsed_time.cos();
        // 겹치지 않게 값을 조정합니다.
        let scaled_elapsed_time_sin_red = elapsed_time_sin * 0.5;
        let scaled_elapsed_time_cos_green = elapsed_time_cos * 0.5;
        let scaled_elapsed_time_sin_blue = -elapsed_time_sin * 0.5;

        let offset = 0.5;
        let red = offset + scaled_elapsed_time_sin_red;
        let green = offset + scaled_elapsed_time_cos_green;
        let blue = offset + scaled_elapsed_time_sin_blue;
        println!(
            "elapsed_time: {}, sin: {}, cos: {}, tan: {}, red: {}, green: {}, blue: {}",
            elapsed_time,
            elapsed_time_sin,
            elapsed_time_cos,
            elapsed_time.tan(),
            red,
            green,
            blue
        );

        // 선 색상 설정
        CGContextSetRGBStrokeColor(cg_context, red, green, blue, 1.0);

        // 원호 추가 (전체 원)
        CGContextAddArc(
            cg_context,
            center_x,
            center_y,
            radius,
            0.0,
            2.0 * std::f64::consts::PI,
            1,
        );

        // 경로를 그림
        CGContextStrokePath(cg_context);
    }
}

// 맥에서 화면을 그리는 과정이 궁금하여 선 하나 간단하게 그려보려고 했지만, 생각보다 쉽지 않습니다.
// 1. 메인 이벤트 루프를 관리할 애플리케이션 생성하기
// 2. 애플리케이션에서 윈도우 창 만들기
// 3. 그림을 그릴 뷰 인스턴스 만들고, 그 부모까지 초기화하기
// 4. 뷰의 사각형 영역을 실제로 그릴 함수 구현하기
// 5. 윈도우에 뷰 설정하기
// 6. 만약 다시 그러야 한다면 타이머 설정하기
// 7. 메인 이벤트 루프 실행하기
//
// FFI 통해서 MacOS의 CoreGraphics 프레임워크의 코드를 호출해야 하고,
// 그러기 위해서 관련된 프레임워크들을 링크하고,
// Objective-C 코드와 런타임이 어떻게 동작하는지도 알아야 하고,
// 특히 메서드 경우 기본 타입 프로모션으로 인해 세그먼트 폴트가 발생할 수 있으므로 타입에 맞게 함수 포인터를 지정해야 하는 등 여러 사항을 고려해야 합니다.
//
// 더 낮은 수준은 어떤지 궁금하지만, 결국 CoreGraphics를 사용하는 방법 외에 없어 보이고,
// 그냥 직접하기보다는 오픈소스를 사용합니다.
// - https://github.com/servo/core-foundation-rs
// - https://github.com/SSheldon/rust-objc
fn main() {
    // start_draw_line_in_rect();
    start_draw_animated_circle_in_rect();
}
