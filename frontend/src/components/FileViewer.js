import React from 'react';

const FileViewer = ({ content, host, timestamp }) => {
  const formatTimestamp = (timestamp) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleString();
    } catch (error) {
      return timestamp;
    }
  };

  if (!content) {
    return (
      <div className="panel">
        <div className="panel-header">
          File Content
        </div>
        <div className="panel-content">
          <div className="loading">
            Select a timestamp to view file content.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="panel">
      <div className="panel-header">
        File Content
      </div>
      <div className="panel-content">
        <div style={{ marginBottom: '15px', padding: '10px', background: '#f8f9fa', borderRadius: '4px' }}>
          <div><strong>Host:</strong> {host}</div>
          <div><strong>Timestamp:</strong> {formatTimestamp(timestamp)}</div>
          <div style={{ fontSize: '0.9rem', opacity: 0.7 }}>{timestamp}</div>
        </div>
        
        <div className="json-viewer">
          {JSON.stringify(content, null, 2)}
        </div>
      </div>
    </div>
  );
};

export default FileViewer;
