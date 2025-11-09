FROM python:3.11-slim

WORKDIR /app

# Install system dependencies for aiortc
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    libopus-dev \
    libvpx-dev \
    libavdevice-dev \
    libavfilter-dev \
    libavformat-dev \
    libavcodec-dev \
    libswresample-dev \
    libswscale-dev \
    libsrtp2-dev \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY unified_recorder.py .

RUN mkdir -p /recordings
VOLUME /recordings

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["python3", "unified_recorder.py"]
