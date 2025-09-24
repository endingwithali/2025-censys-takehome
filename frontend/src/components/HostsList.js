import React from 'react';

const HostsList = ({ hosts, selectedHost, onHostSelect }) => {
  if (!hosts || hosts.length === 0) {
    return (
      <div className="panel">
        <div className="panel-header">
          Available Hosts
        </div>
        <div className="panel-content">
          <div className="loading">
            No hosts available. Upload a snapshot file to get started.
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="panel">
      <div className="panel-header">
        Available Hosts ({hosts.length})
      </div>
      <div className="panel-content" style={{ padding: 0 }}>
        {hosts.map((host, index) => (
          <div
            key={index}
            className={`list-item ${selectedHost === host ? 'selected' : ''}`}
            onClick={() => onHostSelect(host)}
          >
            <div style={{ fontWeight: 'bold' }}>{host}</div>
            <div style={{ fontSize: '1rem', opacity: 0.8 }}>
              Click to view timestamps
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default HostsList;
