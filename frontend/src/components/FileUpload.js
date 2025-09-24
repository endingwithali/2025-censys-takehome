import React, { useState } from 'react';
import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const FileUpload = ({ onUpload }) => {
  const [file, setFile] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [message, setMessage] = useState('');
  const [dragActive, setDragActive] = useState(false);

  const handleFileChange = (e) => {
    const selectedFile = e.target.files[0];
    if (selectedFile) {
      setFile(selectedFile);
      setMessage('');
    }
  };

  const handleDrag = (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  };

  const handleDrop = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const droppedFile = e.dataTransfer.files[0];
      if (droppedFile.name.endsWith('.json')) {
        setFile(droppedFile);
        setMessage('');
      } else {
        setMessage('Please select a JSON file');
      }
    }
  };

  const handleUpload = async () => {
    if (!file) {
      setMessage('Please select a file first');
      return;
    }

    // Validate filename format
    const filenameRegex = /^host_((?:\d{1,3}\.){3}\d{1,3})_(\d{4}-\d{2}-\d{2})T(\d{2})-(\d{2})-(\d{2})(\.\d+)?(Z|[+-]\d{2}-\d{2})\.json$/;
    if (!filenameRegex.test(file.name)) {
      setMessage('Invalid filename format. Expected: host_<ip>_<YYYY-MM-DD>T<HH-MM-SS>[.fraction](Z|±HH-MM).json');
      return;
    }

    setUploading(true);
    setMessage('');

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await axios.post(`${API_BASE_URL}/snapshot`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      if (response.status === 200) {
        setMessage('File uploaded successfully!');
        setFile(null);
        // Reset file input
        const fileInput = document.getElementById('file-input');
        if (fileInput) fileInput.value = '';
        
        // Notify parent component
        if (onUpload) {
          onUpload();
        }
      }
    } catch (error) {
      if (error.response) {
        setMessage(`Upload failed: ${error.response.data}`);
      } else {
        setMessage('Upload failed: Network error');
      }
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="panel">
      <div className="panel-header">
        Upload Host Snapshot
      </div>
      <div className="panel-content">
        <div
          className={`file-upload ${dragActive ? 'dragover' : ''}`}
          onDragEnter={handleDrag}
          onDragLeave={handleDrag}
          onDragOver={handleDrag}
          onDrop={handleDrop}
          onClick={() => document.getElementById('file-input').click()}
        >
          <input
            id="file-input"
            type="file"
            accept=".json"
            onChange={handleFileChange}
            className="file-input"
          />
          
          {file ? (
            <div>
              <p><strong>Selected file:</strong> {file.name}</p>
              <p>Size: {(file.size / 1024).toFixed(2)} KB</p>
            </div>
          ) : (
            <div>
              <p>Click to select a JSON file or drag and drop</p>
              <p>Expected format: host_&lt;ip&gt;_&lt;timestamp&gt;.json</p>
            </div>
          )}
        </div>

        <div className="upload-button">
          <button
            className="button"
            onClick={handleUpload}
            disabled={!file || uploading}
          >
            {uploading ? 'Uploading...' : 'Upload File'}
          </button>
        </div>

        {message && (
          <div className={message.includes('successfully') ? 'success' : 'error'}>
            {message}
          </div>
        )}
      </div>
    </div>
  );
};

export default FileUpload;
