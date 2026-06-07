# 스캔본 도서 PDF OCR: 오픈소스 파이프라인과 최적의 워크플로우

- [스캔본 도서 PDF OCR: 오픈소스 파이프라인과 최적의 워크플로우](#스캔본-도서-pdf-ocr-오픈소스-파이프라인과-최적의-워크플로우)
    - [스캔 PDF의 본질과 OCR의 핵심 과제](#스캔-pdf의-본질과-ocr의-핵심-과제)
    - [OCR 파이프라인 아키텍처 (4단계)](#ocr-파이프라인-아키텍처-4단계)
    - [오픈소스 OCR 도구 비교분석](#오픈소스-ocr-도구-비교분석)
        - [PaddleOCR (PP-OCRv4)](#paddleocr-pp-ocrv4)
        - [Docling (IBM Research)](#docling-ibm-research)
        - [Marker](#marker)
        - [Tesseract (Classic LSTM)](#tesseract-classic-lstm)
    - [OCR 전처리(Preprocessing)의 수학적·엔지니어링 원리](#ocr-전처리preprocessing의-수학적엔지니어링-원리)
        - [이진화 (Binarization - Otsu's Method)](#이진화-binarization---otsus-method)
        - [기울기 보정 (Deskewing - Hough Transform)](#기울기-보정-deskewing---hough-transform)
    - [실무 구현 예제: OpenCV 전처리 + PaddleOCR 파이프라인](#실무-구현-예제-opencv-전처리--paddleocr-파이프라인)
        - [의존성 라이브러리 설치](#의존성-라이브러리-설치)
        - [파이썬(Python) 파이프라인 스크립트](#파이썬python-파이프라인-스크립트)
    - [한국어 스캔 도서 복원을 위한 최적의 워크플로우 제안](#한국어-스캔-도서-복원을-위한-최적의-워크플로우-제안)
    - [실패 모드와 대처법 (Troubleshooting)](#실패-모드와-대처법-troubleshooting)
    - [자가 검증 질문 및 실험 (Active Recall)](#자가-검증-질문-및-실험-active-recall)

---

## 스캔 PDF의 본질과 OCR의 핵심 과제

스캔된 책 PDF(Scanned PDF)는 내부에 텍스트 정보가 포함되지 않은 **단순 이미지(Raster Image)의 연속적인 묶음**입니다. 이 상태의 PDF는 검색, 복사, LLM 기반의 RAG(Retrieval-Augmented Generation) 시스템 입력이 불가능합니다. 

이러한 raw 상태의 스캔 PDF를 사람이 읽기 좋은 디지털 텍스트나 구조화된 문서(Markdown/Structured JSON)로 복원하기 위해서는 단순히 글자를 읽는 것 이상의 과정이 필요합니다. 책의 레이아웃(다단 편집, 헤더, 푸터, 그림 캡션, 표)을 이해하지 못하고 무작위로 글자만 읽어 들인다면(Raw OCR), 문맥이 완전히 꼬여버린 '글자 파편의 무덤'을 얻게 되기 때문입니다.

따라서 최선의 OCR 결과를 얻기 위해서는 **이미지 전처리 -> 문서 레이아웃 분석(Document Layout Analysis) -> 텍스트 영역 인식(Text Recognition) -> 구조화 복원 및 교정(Post-processing)**으로 이어지는 유기적인 파이프라인을 구축해야 합니다.

---

## OCR 파이프라인 아키텍처 (4단계)

오픈소스 도구를 결합하여 스캔본 PDF를 디지털 자산으로 전환하는 전체 처리 흐름은 다음 4단계로 구성됩니다.

```plaintext
+-------------------------------------------------------------------------------------------------+
|                                    1. 이미지 추출 및 전처리 (Preprocessing)                      |
|  [스캔본 PDF] ---> [고해상도 이미지 추출] ---> [기울기 보정(Deskew)] ---> [이진화(Binarization)]   |
+-------------------------------------------------------------------------------------------------+
                                                                  |
                                                                  v
+-------------------------------------------------------------------------------------------------+
|                                    2. 문서 레이아웃 분석 (DLA)                                   |
|  [텍스트 영역 감지] ---> [다단 레이아웃 분할] ---> [헤더/푸터/노이즈 제거] ---> [읽기 순서 결정]   |
+-------------------------------------------------------------------------------------------------+
                                                                  |
                                                                  v
+-------------------------------------------------------------------------------------------------+
|                                    3. 텍스트 인식 (Text Recognition)                            |
|  [각 텍스트 블록 이미지] ---> [OCR 모델 실행 (Ko/En)] ---> [텍스트 문자열 변환 및 신뢰도 산출]     |
+-------------------------------------------------------------------------------------------------+
                                                                  |
                                                                  v
+-------------------------------------------------------------------------------------------------+
|                                    4. 후처리 및 구조화 (Post-processing)                         |
|  [행 합치기 및 줄바꿈 보정] ---> [마크다운/포맷 변환] ---> [LLM 기반 오탈자 교정] ---> [최종본]    |
+-------------------------------------------------------------------------------------------------+
```

1. **이미지 추출 및 전처리 (Preprocessing)**: PDF 내부의 비트맵 이미지를 높은 해상도(300 DPI 이상)로 추출한 후, 스캔 시 발생한 기울어짐을 바로잡고 글자와 배경의 경계를 명확히 하는 이진화 처리를 수행합니다.
2. **문서 레이아웃 분석 (Document Layout Analysis, DLA)**: 본문, 제목, 그림, 표, 머리말/꼬리말(Header/Footer), 쪽 번호 등의 논리적 블록을 구분하고, 책의 읽기 순서(Reading Order)에 맞게 영역을 정렬합니다.
3. **텍스트 인식 (Text Recognition)**: 분할된 텍스트 영역 이미지에 대해 OCR 엔진을 실행하여 실제 글자를 디지털 텍스트로 인식합니다.
4. **후처리 및 구조화 (Post-processing)**: 잘려진 문장 행들을 하나의 단락으로 합치고, 하이픈이 들어간 끊긴 단어를 연결하며, 필요에 따라 LLM(대형 언어 모델)을 사용해 오탈자를 기하학적·문맥적으로 교정합니다.

---

## 오픈소스 OCR 도구 비교분석

스캔 도서 OCR 작업에 사용할 수 있는 현대 오픈소스 도구들의 핵심 특징과 장단점은 다음과 같습니다.

### PaddleOCR (PP-OCRv4)
중국 PaddlePaddle 생태계의 대표적인 고성능 실용주의 OCR 라이브러리입니다.
*   **특징**: 텍스트 감지(DBNet)와 텍스트 인식(SVTR) 모듈이 고도로 경량화되어 CPU 환경에서도 압도적인 속도를 냅니다.
*   **장점**: 한국어를 포함한 다국어(80개 이상) 모델의 인식 품질이 오픈소스 중 최고 수준이며, 표 구조 분석용 `PP-Structure` 모듈을 제공합니다.
*   **단점**: 프레임워크가 무겁고(PaddlePaddle 종속), 서적 단위의 마크다운(Markdown) 자동 복원 기능은 기본 API로 제공되지 않아 상위 스크립트 작성이 필요합니다.

### Docling (IBM Research)
Generative AI와 RAG 파이프라인에 문서를 공급하기 위해 IBM Research가 2024~2025년에 출시한 최신 문서 파싱 프레임워크입니다.
*   **특징**: 단순 글자 인식을 넘어 문서 레이아웃 자체를 시각적으로 이해(DocLayNet 데이터셋 기반 모델 사용)하여 구조를 분석합니다.
*   **장점**: 표(Table), 리스트, 수식(LaTeX), 제목 계층을 완벽히 보존하며 최종 결과물을 마크다운이나 JSON으로 변환해 줍니다. OCR 엔진으로 EasyOCR, Tesseract, RapidOCR 등을 백엔드로 플러그인할 수 있습니다.
*   **단점**: CPU 환경에서 레이아웃 분석 딥러닝 모델 구동 시 다소 속도가 느리며 GPU 하드웨어가 권장됩니다.

### Marker
PDF 데이터를 마크다운 파일로 빠르게 변환하기 위해 개발된 특화 도구입니다.
*   **특징**: 레이아웃 감지 모델, 헤더/푸터 탐지기, 수식 및 코드 검출 모델을 파이프라인으로 결합하여 책 전체를 하나의 깔끔한 디지털 문서로 재구성합니다.
*   **장점**: 수식이 많은 수학/과학 서적이나 논문 형태의 스캔본을 LaTeX 표현식이 삽입된 마크다운으로 환상적으로 번환해 줍니다.
*   **단점**: 컴퓨팅 자원 요구량이 매우 크며, 한국어 전용 튜닝이 일부 부족할 경우 다국어 OCR 백엔드 설정에 신경을 써야 합니다.

### Tesseract (Classic LSTM)
HP와 구글이 개발해 온 오랜 역사의 전통적인 OCR 엔진입니다.
*   **특징**: 문자 분할 및 LSTM 기반 문자열 인식을 사용하며 매우 가볍고 안정적입니다.
*   **장점**: 이미 리눅스 패키지 매니저에 내장되어 있을 정도로 가볍고 CPU 자원 소모가 매우 적습니다.
*   **단점**: 다단 구성(Multi-column), 표, 그림이 뒤섞인 복잡한 레이아웃에서 텍스트 읽기 순서가 완전히 뒤엉킵니다. 한국어 스캔본 특유의 뭉개진 글씨나 노이즈가 낀 이미지에서는 인식 정확도가 크게 떨어집니다.

---

## OCR 전처리(Preprocessing)의 수학적·엔지니어링 원리

스캔 서적의 OCR 인식률을 결정짓는 가장 중요한 요소는 **원천 이미지의 품질**입니다. 노이즈가 낀 이미지를 그대로 OCR 엔진에 입력하면 오인식율(Character Error Rate, CER)이 급증합니다.

### 이진화 (Binarization - Otsu's Method)
스캔본의 종이 배경색(미색, 누런색), 그림자, 인쇄 비침 노이즈를 제거하기 위해 이미지를 흑백(0과 255)의 명확한 상태로 만드는 연산입니다. 
오츠 이진화(Otsu's Binarization) 알고리즘은 이미지의 픽셀 히스토그램에서 **배경(Background)과 전경(Foreground) 클래스 간의 분산(Intra-class Variance)을 최소화하는 최적의 임계치(Threshold, $t$)**를 수학적으로 계산해 냅니다.

$$\sigma_w^2(t) = \omega_0(t)\sigma_0^2(t) + \omega_1(t)\sigma_1^2(t)$$

여기서 $\omega_0$와 $\omega_1$은 임계치 $t$에 의해 분리된 두 클래스의 확률이며, $\sigma_0^2$와 $\sigma_1^2$은 각 클래스의 분산입니다. 이 값들을 최소화하는 $t$ 값을 찾아 동적으로 픽셀을 이진화하므로, 조명이 불균일하게 스캔된 페이지에서도 훌륭하게 글자만을 분리해 냅니다.

### 기울기 보정 (Deskewing - Hough Transform)
책을 북스캐너로 스캔할 때 필연적으로 미세한 회전 기울어짐(Skew)이 발생합니다. 글자선이 수평을 이루지 않으면 글자 영역(Bounding Box)을 잡는 딥러닝 모델의 오차가 커집니다.
기울기 보정은 허프 변환(Hough Transform)을 사용하여 이미지 내 직선 성분을 추출하고, 텍스트 라인들이 이루는 지배적인 각도(Dominant Angle)를 측정하여 이미지를 역회전시킵니다.

$$r = x \cos\theta + y \sin\theta$$

이미지 공간의 점 $(x, y)$들을 극좌표 매개변수 공간 $(r, \theta)$로 누적(Accumulate)시켜 교차점이 극대가 되는 $\theta$ 각도를 찾아냄으로써, 텍스트 줄들의 평균 기울기를 산출하고 이를 보정합니다.

---

## 실무 구현 예제: OpenCV 전처리 + PaddleOCR 파이프라인

스캔 PDF에서 이미지를 추출한 뒤, OpenCV를 통해 기울기 및 이진화 전처리를 수행하고 PaddleOCR로 한국어와 영어를 추출하는 완전한 파이썬 스크립트 예제입니다.

### 의존성 라이브러리 설치
macOS 또는 Linux 환경의 터미널에서 다음 명령을 실행하여 필수 라이브러리를 설치합니다. (PDF에서 이미지를 렌더링하기 위해 시스템 레벨의 `poppler` 패키지가 필요합니다.)

```bash
# macOS 기준 poppler 설치
brew install poppler

# 파이썬 의존성 설치
pip install pdf2image opencv-python numpy paddlepaddle paddleocr
```

### 파이썬(Python) 파이프라인 스크립트

```python
import os
import cv2
import numpy as np
from pdf2image import convert_from_path
from paddleocr import PaddleOCR

# 1. OCR 엔진 초기화 (한국어 'ko', 영어 'en' 동시 지원 모델)
# use_gpu=False 설정으로 CPU 환경에서도 원활하게 작동하도록 유도
ocr = PaddleOCR(use_angle_cls=True, lang='ko', use_gpu=False)

def preprocess_image(image_path):
    """
    OpenCV를 사용한 이미지 전처리 함수
    - 그레이스케일 변환, 오츠 이진화, 노이즈 제거 진행
    """
    # 이미지 로드
    src = cv2.imread(image_path)
    if src is None:
        raise FileNotFoundError(f"이미지를 찾을 수 없습니다: {image_path}")
        
    # A. 그레이스케일(Grayscale) 변환 (채널 수를 1개로 축소)
    gray = cv2.cvtColor(src, cv2.COLOR_BGR2GRAY)
    
    # B. 노이즈 제거를 위한 블러링 (Gaussian Blur)
    # 글자의 에지가 너무 손상되지 않도록 3x3 커널 사용
    blurred = cv2.GaussianBlur(gray, (3, 3), 0)
    
    # C. 오츠 알고리즘 기반 이진화 (Otsu's Thresholding)
    # 배경의 얼룩과 글씨 뒤편 비침을 제거하여 명확한 흑백 상태로 만듦
    _, thresh = cv2.threshold(blurred, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
    
    return thresh

def deskew_image(image_binary):
    """
    허프 변환을 사용한 텍스트 기울기 각도 측정 및 역회전 보정
    """
    coords = np.column_stack(np.where(image_binary == 0))
    angle = cv2.minAreaRect(coords)[-1]
    
    # OpenCV 각도 정의에 따른 보정 임계값 처리
    if angle < -45:
        angle = -(90 + angle)
    else:
        angle = -angle
        
    (h, w) = image_binary.shape[:2]
    center = (w // 2, h // 2)
    
    # 회전 행렬 생성 및 이미지 워핑(Warping)
    M = cv2.getRotationMatrix2D(center, angle, 1.0)
    rotated = cv2.warpAffine(
        image_binary, M, (w, h), 
        flags=cv2.INTER_CUBIC, 
        borderMode=cv2.BORDER_CONSTANT, 
        borderValue=255
    )
    return rotated, angle

def process_pdf_ocr(pdf_path, output_dir="ocr_results"):
    """
    스캔 PDF를 이미지로 변환하고 전처리 후 PaddleOCR을 실행하는 전체 제어 함수
    """
    os.makedirs(output_dir, exist_ok=True)
    
    # A. PDF -> 고해상도 Image 렌더링 (300 DPI 권장)
    print(f"[단계 1] PDF 페이지 이미지 렌더링 중: {pdf_path}")
    pages = convert_from_path(pdf_path, dpi=300)
    
    for i, page in enumerate(pages):
        page_num = i + 1
        temp_img_path = os.path.join(output_dir, f"page_{page_num}_temp.png")
        page.save(temp_img_path, "PNG")
        
        # B. 이미지 전처리 수행
        print(f"[단계 2] 페이지 {page_num} 이미지 전처리 및 보정 진행")
        binary_img = preprocess_image(temp_img_path)
        corrected_img, angle = deskew_image(binary_img)
        
        # 보정된 이미지를 다시 저장하여 OCR 엔진에 입력
        processed_img_path = os.path.join(output_dir, f"page_{page_num}_processed.png")
        cv2.imwrite(processed_img_path, corrected_img)
        
        # C. PaddleOCR 텍스트 추출 실행
        print(f"[단계 3] 페이지 {page_num} OCR 인식 중 (기울기 보정 각도: {angle:.2f}도)")
        result = ocr.ocr(processed_img_path, cls=True)
        
        # D. 결과 파일 저장
        txt_output_path = os.path.join(output_dir, f"page_{page_num}_text.txt")
        with open(txt_output_path, "w", encoding="utf-8") as f:
            for line in result[0]:
                box = line[0]        # 텍스트 영역의 4점 좌표 [[x1,y1], [x2,y2], [x3,y3], [x4,y4]]
                text, confidence = line[1]  # 인식된 텍스트와 신뢰도 점수 (0.0 ~ 1.0)
                
                # 가독성을 위해 신뢰도 80% 이상인 텍스트만 출력하거나 로그로 필터링 가능
                f.write(f"[{box}] -> {text} (신뢰도: {confidence:.2f})\n")
                
        # 임시 원본 이미지 정리
        if os.path.exists(temp_img_path):
            os.remove(temp_img_path)
            
        print(f"[완료] 페이지 {page_num} 추출 완료 -> {txt_output_path}")

if __name__ == "__main__":
    # 실행 테스트 (작업 공간 내 임의의 pdf 파일을 지정하여 사용 가능)
    # sample_pdf_path = "sample_book.pdf"
    # process_pdf_ocr(sample_pdf_path)
    print("OCR 파이프라인 스크립트가 준비되었습니다.")
```

---

## 한국어 스캔 도서 복원을 위한 최적의 워크플로우 제안

실무에서 한국어 스캔본 도서를 일반 텍스트나 단순 스냅샷 텍스트가 아닌, **재구성된 마크다운 서적**으로 완벽히 복원하고자 할 때 적용할 수 있는 가장 균형 잡힌 최선의 아키텍처 워크플로우는 다음과 같습니다.

```plaintext
                   [ 스캔 PDF ]
                        |
                        v
          +---------------------------+
          |  1. Docling (IBM) 실행    |  <--- 페이지 레이아웃 및 논리적 구조 분석
          +---------------------------+
                        |
      +-----------------+-----------------+
      |                                   |
      v                                   v
[ 표(Table) / 수식 영역 ]           [ 본문 텍스트 영역 ]
      |                                   |
      v                                   v
[ Docling 자체 Table Parser ]     [ PaddleOCR (Ko/En) ]
      |                                   |
      +-----------------+-----------------+
                        |
                        v
          +---------------------------+
          | 2. 마크다운(MD) 구조 조립  |  <--- 텍스트와 표 포맷의 병합
          +---------------------------+
                        |
                        v
          +---------------------------+
          | 3. LLM API 오탈자 교정    |  <--- 문맥(Context) 기반 최종 오차 수정
          +---------------------------+
                        |
                        v
                 [ 최종 마크다운 ]
```

1.  **DLA 프레임워크로 Docling(IBM) 사용**:
    Docling은 딥러닝 기반 모델을 이용해 문서 전체의 제목 계층, 다단 흐름, 표의 테두리와 내용을 파악하는 능력이 매우 뛰어납니다. 따라서 전체 문서의 구조 분석 및 최종 마크다운 뼈대 조립은 Docling에 맡깁니다.
2.  **한국어 OCR 백엔드로 PaddleOCR 결합**:
    Docling의 기본 OCR인 EasyOCR은 한국어 복잡 문장에서 오인식률이 다소 높고, Tesseract는 낡은 문자열 분할 모델을 씁니다. 따라서 Docling 내부 설정에서 OCR 백엔드를 PaddleOCR의 추출 모듈로 교체하여 이미지 블록 단위의 한국어 텍스트를 고정밀도로 수급합니다.
3.  **LLM 문맥 교정 파이프라인 추가**:
    아무리 우수한 OCR이라도 한국어의 조사(은/는/이/가)나 띄어쓰기, 깨진 한글 폰트를 100% 인식하지 못합니다. 추출된 마크다운 본문을 단락 단위로 묶어 GPT-4o나 Claude 3.5 Sonnet 같은 LLM API에 다음과 같은 시스템 프롬프트를 부여해 통과시키는 최종 교정기를 붙입니다.
    
    > **시스템 프롬프트 예시**:
    > "너는 스캔 서적 OCR 본문의 의미적 오류와 오인식을 보정하는 교정 전문가이다. 입력 데이터는 OCR 엔진이 추출한 한국어 마크다운 텍스트이다. 아래 규칙을 엄격히 준수하라.
    > 1. 원문의 의미나 내용을 임의로 요약하거나 왜곡하지 말고, 원래의 단락 구조를 유지하라.
    > 2. OCR 인식 결함으로 생긴 깨진 한글 글자(예: '햔국' -> '한국', '책욜' -> '책을')와 잘못된 조사 결합을 전후 문맥에 맞게 수정하라.
    > 3. 행이 나뉘면서 강제로 분절된 단어들을 자연스럽게 이어라."

---

## 실패 모드와 대처법 (Troubleshooting)

스캔 서적 OCR 작업 중 흔히 마주치는 실패 패턴들과 해결법입니다.

*   **다단(Multi-column) 편집 서적인데 문장이 가로로 겹쳐서 읽히는 현상**
    *   *원인*: 레이아웃 분석기(DLA)가 작동하지 않고 OCR 엔진이 페이지의 전체 가로줄을 왼쪽에서 오른쪽으로 그냥 읽었기 때문입니다.
    *   *해결*: 단순 이미지 OCR을 돌리기 전에 `Docling` 또는 `PP-Structure`를 사용하여 단(Column) 영역의 바운딩 박스를 먼저 분할한 후, 각 단 영역 내부만 독립적으로 잘라서 OCR을 수행해야 합니다.
*   **스캔 품질 저하로 글자가 깨지거나 점선처럼 뭉개지는 현상**
    *   *원인*: 렌더링 DPI가 너무 낮거나 이진화 Threshold 상수가 부적절해 글자의 획이 끊어졌기 때문입니다.
    *   *해결*: PDF 렌더링 시 DPI를 `300` 이상, 필요 시 `450`으로 높이고, OpenCV에서 `cv2.adaptiveThreshold`를 사용하여 국소 영역별 가변 임계치를 적용하십시오.
*   **표(Table) 안의 텍스트가 무작위로 섞여 나오는 현상**
    *   *원인*: 셀 구분이 인식되지 않고 표가 단순 텍스트 줄로 처리되었기 때문입니다.
    *   *해결*: 이 경우에는 수동 정렬 코드를 직접 구현하는 것보다 IBM Docling의 `TableFormer` 모듈이나 PaddleOCR의 `PP-Structure` 테이블 인지 모듈을 활성화하여 표 영역을 XML/HTML 포맷으로 변환한 뒤 마크다운 표 구조(`|---|`)로 컴파일하는 방식이 최선입니다.

---

## 자가 검증 질문 및 실험 (Active Recall)

이 설명의 내용을 실제 스캔 PDF OCR 환경에 적용해 보고 기억에 남기기 위해 다음 질문들을 검토해 보십시오.

1.  스캔본 PDF에서 일반 OCR을 바로 실행할 때, 머리말(Header)과 쪽 번호(Page Number)가 본문 문맥 사이에 끼어드는 문제를 코드로 어떻게 우회할 수 있을까요? (힌트: 레이아웃 바운딩 박스의 Y축 좌표 필터링)
2.  오츠 이진화(Otsu's Thresholding) 알고리즘이 이미지 조명이 불균일하게 그림자가 져 있는 모서리 영역에서도 잘 작동할까요? 작동하지 않는다면 어떤 방식의 이진화로 대체해야 할까요?
3.  직접 변환하려는 PDF 책의 1페이지를 골라, 렌더링 DPI를 `72`, `150`, `300`으로 각각 달리하여 추출된 이미지의 픽셀 조밀도를 비교해 보고 PaddleOCR의 인식 정확도(Confidence) 점수가 어떻게 변화하는지 측정해 보십시오.
