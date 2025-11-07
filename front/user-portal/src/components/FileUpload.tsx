import React, { useState, useRef } from 'react';
import './FileUpload.css';

const MAX_FILE_SIZE = 500 * 1024 * 1024; // 500MB
const ALLOWED_TYPES = [
  'audio/mpeg',
  'audio/wav',
  'audio/m4a',
  'audio/mp4',
  'audio/x-m4a',
  'video/mp4',
  'video/mpeg',
  'video/quicktime'
];

export const FileUpload: React.FC = () => {
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [dragActive, setDragActive] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const validateFile = (file: File): string | null => {
    if (file.size > MAX_FILE_SIZE) {
      return 'File size exceeds 500MB limit';
    }
    if (!ALLOWED_TYPES.includes(file.type)) {
      return 'Invalid file type. Please upload audio or video files only';
    }
    return null;
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const droppedFile = e.dataTransfer.files[0];
      const validationError = validateFile(droppedFile);

      if (validationError) {
        setError(validationError);
        setFile(null);
      } else {
        setFile(droppedFile);
        setError(null);
      }
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const selectedFile = e.target.files[0];
      const validationError = validateFile(selectedFile);

      if (validationError) {
        setError(validationError);
        setFile(null);
      } else {
        setFile(selectedFile);
        setError(null);
      }
    }
  };

  const handleUpload = async () => {
    if (!file) return;

    setUploading(true);
    setProgress(0);
    setError(null);
    setSuccess(null);

    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const formData = new FormData();
      formData.append('file', file);

      const xhr = new XMLHttpRequest();

      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          const percentComplete = (e.loaded / e.total) * 100;
          setProgress(Math.round(percentComplete));
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status === 200) {
          setSuccess('File uploaded successfully!');
          setFile(null);
          setProgress(0);
          if (fileInputRef.current) {
            fileInputRef.current.value = '';
          }
          // Trigger a custom event to refresh the file list
          window.dispatchEvent(new Event('fileUploaded'));
        } else {
          const response = JSON.parse(xhr.responseText);
          setError(response.error || 'Failed to upload file');
        }
        setUploading(false);
      });

      xhr.addEventListener('error', () => {
        setError('Network error occurred during upload');
        setUploading(false);
      });

      xhr.open('POST', '/api/v1/files/upload');
      xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      xhr.send(formData);

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to upload file');
      setUploading(false);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  return (
    <div className="file-upload-container">
      <h2>Upload Audio/Video File</h2>
      <p className="upload-description">
        Upload audio or video files for transcription. Maximum file size: 500MB
      </p>

      <div
        className={`drop-zone ${dragActive ? 'drag-active' : ''} ${file ? 'has-file' : ''}`}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
      >
        <input
          ref={fileInputRef}
          type="file"
          onChange={handleFileChange}
          accept="audio/*,video/*"
          style={{ display: 'none' }}
        />

        {file ? (
          <div className="file-info">
            <div className="file-icon">📄</div>
            <div className="file-details">
              <p className="file-name">{file.name}</p>
              <p className="file-size">{formatFileSize(file.size)}</p>
            </div>
          </div>
        ) : (
          <div className="drop-zone-content">
            <div className="upload-icon">⬆️</div>
            <p className="drop-zone-text">
              <strong>Click to upload</strong> or drag and drop
            </p>
            <p className="drop-zone-hint">
              Audio: MP3, WAV, M4A | Video: MP4, MOV
            </p>
          </div>
        )}
      </div>

      {uploading && (
        <div className="upload-progress">
          <div className="progress-bar">
            <div
              className="progress-fill"
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="progress-text">{progress}% uploaded</p>
        </div>
      )}

      {error && (
        <div className="upload-message error">
          <span className="message-icon">⚠️</span>
          <span>{error}</span>
        </div>
      )}

      {success && (
        <div className="upload-message success">
          <span className="message-icon">✅</span>
          <span>{success}</span>
        </div>
      )}

      <div className="upload-actions">
        {file && !uploading && (
          <>
            <button
              onClick={() => {
                setFile(null);
                setError(null);
                setSuccess(null);
                if (fileInputRef.current) {
                  fileInputRef.current.value = '';
                }
              }}
              className="btn btn-secondary"
            >
              Clear
            </button>
            <button
              onClick={handleUpload}
              className="btn btn-primary"
            >
              Upload File
            </button>
          </>
        )}
      </div>
    </div>
  );
};

export default FileUpload;
