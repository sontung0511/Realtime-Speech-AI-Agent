# Realtime Speech AI Agent

## Overview

Realtime Speech AI Agent là hệ thống AI có khả năng nghe audio/speech theo thời gian thực, tự động chuyển giọng nói thành văn bản, nhận diện ngôn ngữ, hiểu ngữ cảnh hội thoại và trả lời gần như ngay lập tức.

Mục tiêu của project là xây dựng một AI agent có thể:

- Nghe giọng nói realtime từ microphone hoặc audio stream
- Detect ngôn ngữ đang nói
- Chuyển speech sang text theo từng đoạn nhỏ
- Hiểu ngữ cảnh của câu chuyện hoặc hội thoại
- Ghi chú nội dung quan trọng
- Sinh câu trả lời realtime
- Tùy chọn đọc câu trả lời bằng giọng nói

---

## Core Idea

Thay vì upload audio rồi xử lý sau, hệ thống hoạt động theo dạng streaming pipeline.

```text
Microphone / Audio Stream
  -> Voice Activity Detection
  -> Streaming Speech To Text
  -> Language Detection
  -> Context Understanding
  -> AI Response Generator
  -> Optional Text To Speech
