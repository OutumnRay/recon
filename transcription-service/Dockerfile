FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        gcc \
        g++ \
        libpq-dev \
        ffmpeg \
        wget \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .

# Install Python dependencies
# Install PyTorch with CPU support by default (GPU version will be used if CUDA is available)
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Create directory for Whisper models
RUN mkdir -p /app/models

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV WHISPER_CACHE_DIR=/app/models

# Pre-download Whisper model during build
# This downloads the model once during image build instead of on every container start
RUN python -c "from faster_whisper import WhisperModel; \
    import os; \
    model_name = os.getenv('WHISPER_MODEL', 'medium'); \
    print(f'Downloading Whisper model: {model_name}'); \
    model = WhisperModel(model_name, device='cpu', compute_type='int8'); \
    print(f'Model {model_name} downloaded successfully')"

# Run the worker
CMD ["python", "main.py"]
