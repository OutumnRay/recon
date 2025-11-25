# Base image: NVIDIA PyTorch with CUDA support
FROM nvcr.io/nvidia/pytorch:25.04-py3

# Set working directory
WORKDIR /app

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements file
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Create directory for downloaded models
RUN mkdir -p /root/.cache/huggingface

# Pre-download Whisper large-v3 model to embed it in the image
# This ensures the model is available immediately when container starts
RUN python -c "from faster_whisper import WhisperModel; \
    print('Downloading Whisper large-v3 model...'); \
    model = WhisperModel('large-v3', device='cpu', compute_type='int8'); \
    print('Model downloaded successfully!')"

# Expose port if needed (for health checks or API)
# EXPOSE 8000

# Health check (optional - adjust according to your service)
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD python -c "import sys; sys.exit(0)" || exit 1

# Run the application
CMD ["python", "-u", "main.py"]
