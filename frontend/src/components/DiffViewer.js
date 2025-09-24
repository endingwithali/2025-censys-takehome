import React from 'react';

const DiffViewer = ({ 
  timestamps, 
  selectedTimestamp, 
  selectedTimestamp2, 
  onTimestamp2Select, 
  diffContent 
}) => {
  const formatTimestamp = (timestamp) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleString();
    } catch (error) {
      return timestamp;
    }
  };

  const getAvailableTimestamps = () => {
    return timestamps.filter(ts => ts !== selectedTimestamp);
  };

  const parseAnsiToHtml = (text) => {
    if (!text) return '';
    
    // ANSI color code mappings - optimized for dark background
    const ansiColors = {
      '[0;30m': { color: '#666666' }, // Black -> Gray
      '[0;31m': { color: '#ff6b6b' }, // Red -> Light Red
      '[0;32m': { color: '#51cf66' }, // Green -> Light Green
      '[0;33m': { color: '#ffd43b' }, // Yellow -> Light Yellow
      '[0;34m': { color: '#74c0fc' }, // Blue -> Light Blue
      '[0;35m': { color: '#da77f2' }, // Magenta -> Light Magenta
      '[0;36m': { color: '#20c997' }, // Cyan -> Light Cyan
      '[0;37m': { color: '#ffffff' }, // White
      '[1;30m': { color: '#868e96' }, // Bright Black -> Light Gray
      '[1;31m': { color: '#ff8787' }, // Bright Red -> Brighter Red
      '[1;32m': { color: '#69db7c' }, // Bright Green -> Brighter Green
      '[1;33m': { color: '#ffec99' }, // Bright Yellow -> Brighter Yellow
      '[1;34m': { color: '#91d5ff' }, // Bright Blue -> Brighter Blue
      '[1;35m': { color: '#e599f7' }, // Bright Magenta -> Brighter Magenta
      '[1;36m': { color: '#63e6be' }, // Bright Cyan -> Brighter Cyan
      '[1;37m': { color: '#ffffff' }, // Bright White
      '[0m': { color: 'inherit' }, // Reset
    };

    // Enhanced regex to catch more ANSI codes
    const ansiRegex = /(\[\d+(?:;\d+)*m)/g;
    const parts = text.split(ansiRegex);
    const elements = [];
    let currentStyle = {};

    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      
      if (ansiColors[part]) {
        // Update current style
        if (part === '[0m') {
          currentStyle = {};
        } else {
          currentStyle = { ...currentStyle, ...ansiColors[part] };
        }
      } else if (part && part.trim()) {
        // Create span with current style
        const style = Object.keys(currentStyle).length > 0 ? currentStyle : {};
        elements.push(
          <span key={i} style={style}>
            {part}
          </span>
        );
      } else if (part) {
        // Include whitespace and newlines
        elements.push(
          <span key={i}>
            {part}
          </span>
        );
      }
    }

    return elements.length > 0 ? elements : text;
  };

  return (
    <div className="panel">
      <div className="panel-header">
        Compare Snapshots
      </div>
      <div className="panel-content">
        <div style={{ marginBottom: '15px' }}>
          <div style={{ padding: '10px', background: '#f8f9fa', borderRadius: '4px', marginBottom: '10px' }}>
            <div><strong>Base Timestamp:</strong> {formatTimestamp(selectedTimestamp)}</div>
            <div style={{ fontSize: '0.9rem', opacity: 0.7 }}>{selectedTimestamp}</div>
          </div>
          
          <div style={{ marginBottom: '15px' }}>
            <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
              Select timestamp to compare:
            </label>
            <select
              value={selectedTimestamp2 || ''}
              onChange={(e) => onTimestamp2Select(e.target.value)}
              style={{
                width: '100%',
                padding: '10px',
                border: '1px solid #ccc',
                borderRadius: '4px',
                fontSize: '16px'
              }}
            >
              <option value="">Choose a timestamp...</option>
              {getAvailableTimestamps().map((timestamp, index) => (
                <option key={index} value={timestamp}>
                  {formatTimestamp(timestamp)}
                </option>
              ))}
            </select>
          </div>
        </div>

        {diffContent && (
          <div>
            <div className={`diff-status ${diffContent.DiffStatus === 'identical' ? 'identical' : 'different'}`}>
              Status: {diffContent.DiffStatus}
            </div>
            
            <div className="diff-viewer">
              {diffContent.Differences ? parseAnsiToHtml(diffContent.Differences) : 'No differences found'}
            </div>
          </div>
        )}

        {selectedTimestamp2 && !diffContent && (
          <div className="loading">
            Loading differences...
          </div>
        )}

        {!selectedTimestamp2 && (
          <div style={{ textAlign: 'center', color: '#6c757d', padding: '20px' }}>
            Select a timestamp to compare and view differences
          </div>
        )}
      </div>
    </div>
  );
};

export default DiffViewer;
