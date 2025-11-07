import React, { useState, useEffect } from 'react';
import './FilesList.css';

interface UploadedFile {
  id: string;
  original_name: string;
  file_size: number;
  status: string;
  uploaded_at: string;
}

export const FilesList: React.FC = () => {
  const [files, setFiles] = useState<UploadedFile[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 20;

  useEffect(() => {
    fetchFiles();

    // Listen for file upload events to refresh the list
    const handleFileUploaded = () => {
      fetchFiles();
    };
    window.addEventListener('fileUploaded', handleFileUploaded);

    return () => {
      window.removeEventListener('fileUploaded', handleFileUploaded);
    };
  }, [page]);

  const fetchFiles = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/files?page=${page}&page_size=${pageSize}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch files');
      }

      const data = await response.json();
      setFiles(data.files || []);
      setTotal(data.total || 0);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load files');
    } finally {
      setLoading(false);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getStatusBadgeClass = (status: string): string => {
    switch (status.toLowerCase()) {
      case 'pending':
        return 'status-badge status-pending';
      case 'processing':
        return 'status-badge status-processing';
      case 'completed':
        return 'status-badge status-completed';
      case 'failed':
        return 'status-badge status-failed';
      default:
        return 'status-badge';
    }
  };

  const totalPages = Math.ceil(total / pageSize);

  if (loading && files.length === 0) {
    return (
      <div className="files-list-container">
        <div className="files-list-header">
          <h2>Uploaded Files</h2>
        </div>
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Loading files...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="files-list-container">
      <div className="files-list-header">
        <h2>Uploaded Files</h2>
        <button onClick={fetchFiles} className="refresh-btn" disabled={loading}>
          🔄 Refresh
        </button>
      </div>

      {error && (
        <div className="error-message">
          <span className="error-icon">⚠️</span>
          <span>{error}</span>
        </div>
      )}

      {files.length === 0 && !loading && (
        <div className="empty-state">
          <div className="empty-icon">📁</div>
          <p className="empty-text">No files uploaded yet</p>
          <p className="empty-hint">Upload your first audio or video file above</p>
        </div>
      )}

      {files.length > 0 && (
        <>
          <div className="files-table-wrapper">
            <table className="files-table">
              <thead>
                <tr>
                  <th>Filename</th>
                  <th>Size</th>
                  <th>Status</th>
                  <th>Uploaded</th>
                </tr>
              </thead>
              <tbody>
                {files.map((file) => (
                  <tr key={file.id}>
                    <td className="filename-cell">
                      <span className="file-icon">📄</span>
                      <span className="filename">{file.original_name}</span>
                    </td>
                    <td>{formatFileSize(file.file_size)}</td>
                    <td>
                      <span className={getStatusBadgeClass(file.status)}>
                        {file.status}
                      </span>
                    </td>
                    <td>{formatDate(file.uploaded_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="pagination">
              <button
                onClick={() => setPage(page - 1)}
                disabled={page === 1}
                className="pagination-btn"
              >
                Previous
              </button>
              <span className="pagination-info">
                Page {page} of {totalPages} ({total} total files)
              </span>
              <button
                onClick={() => setPage(page + 1)}
                disabled={page >= totalPages}
                className="pagination-btn"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default FilesList;
