import React from 'react';

const TimestampsView = ({ 
  timestamps, 
  selectedTimestamp, 
  onTimestampSelect, 
  onShowDiff, 
  showDiff 
}) => {
  const formatTimestamp = (timestamp) => {
    try {
      // Convert the timestamp format from the backend to a more readable format
      const date = new Date(timestamp);
      return date.toLocaleString();
    } catch (error) {
      return timestamp;
    }
  };

  if (!timestamps || timestamps.length === 0) {
    return (
      <div className="panel">
        <div className="panel-header">
          Timestamps
        </div>
        <div className="panel-content">
          <div className="loading">
            No timestamps available for this host.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="panel">
      <div className="panel-header">
        Timestamps ({timestamps.length})
      </div>
      <div className="panel-content" style={{ padding: 0 }}>
        {timestamps.map((timestamp, index) => (
          <div
            key={index}
            className={`list-item ${selectedTimestamp === timestamp ? 'selected' : ''}`}
            onClick={() => onTimestampSelect(timestamp)}
          >
            <div style={{ fontWeight: 'bold' }}>
              {formatTimestamp(timestamp)}
            </div>
            <div style={{ fontSize: '0.9rem', opacity: 0.7 }}>
              {timestamp}
            </div>
          </div>
        ))}
        
        {selectedTimestamp && (
          <div style={{ padding: '15px', borderTop: '1px solid #eee' }}>
            <button
              className={`button ${showDiff ? 'secondary' : ''}`}
              onClick={onShowDiff}
            >
              {showDiff ? 'Hide Diff' : 'Show Diff'}
            </button>
            <div style={{ fontSize: '0.9rem', marginTop: '8px', opacity: 0.7 }}>
              {showDiff ? 'Select another timestamp to compare' : 'Click to compare with other timestamps'}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default TimestampsView;
