FROM python:3.11-slim

WORKDIR /app

# Установка системных зависимостей (минимальный набор, без recommends)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ffmpeg \
        libopus0 \
        wget \
        gcc \
        g++ \
        pkg-config \
        libavformat-dev \
        libavcodec-dev \
        libavutil-dev \
        libswscale-dev \
        libopus-dev && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Копируем requirements.txt
COPY requirements.txt /app/

# Установка Python пакетов (поштучно для экономии памяти)
RUN pip install --no-cache-dir websockets==12.0 && \
    pip install --no-cache-dir aiohttp==3.9.* && \
    pip install --no-cache-dir av==12.0.* && \
    pip install --no-cache-dir aiortc==1.9.* && \
    pip install --no-cache-dir boto3==1.34.* && \
    pip install --no-cache-dir redis==5.0.* && \
    pip install --no-cache-dir python-dotenv==1.0.*

# Копируем все Python скрипты
COPY *.py /app/

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запуск
CMD ["python", "-u", "jitsi_webrtc_bot.py"]