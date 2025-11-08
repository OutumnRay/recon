FROM python:3.11-alpine

# Установка зависимостей
RUN apk add --no-cache \
    ffmpeg \
    opus-tools \
    gcc \
    musl-dev \
    libffi-dev \
    wget

# Установка Python пакетов
RUN pip install --no-cache-dir \
    boto3==1.34.* \
    aiohttp==3.9.* \
    redis==5.0.* \
    websockets==12.0

WORKDIR /app

# Копируем скрипт
COPY recorder.py /app/

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запуск
CMD ["python", "-u", "recorder.py"]