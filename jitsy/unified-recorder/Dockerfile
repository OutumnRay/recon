# --- unified-recorder/Dockerfile ---

FROM python:3.11-slim

WORKDIR /app

# Устанавливаем системные зависимости для сборки aiortc
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    libopus-dev \
    libvpx-dev \
    libavdevice-dev \
    libsrtp2-dev \
    && rm -rf /var/lib/apt/lists/*

# Копируем файл с зависимостями
COPY requirements.txt .

# Устанавливаем Python зависимости
RUN pip install --no-cache-dir -r requirements.txt

# Копируем код приложения
COPY . .

# Создаем директорию для записей
RUN mkdir -p /recordings
VOLUME /recordings

# Открываем порт
EXPOSE 8080

# Проверка состояния
HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Запускаем приложение
CMD ["python3", "unified_recorder.py"]