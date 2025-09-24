import React, { useState, useEffect } from 'react';
import './App.css';
import FileUpload from './components/FileUpload';
import HostsList from './components/HostsList';
import TimestampsView from './components/TimestampsView';
import FileViewer from './components/FileViewer';
import DiffViewer from './components/DiffViewer';

const API_BASE_URL = 'http://localhost:8080/api';

function App() {
  const [hosts, setHosts] = useState([]);
  const [selectedHost, setSelectedHost] = useState(null);
  const [timestamps, setTimestamps] = useState([]);
  const [selectedTimestamp, setSelectedTimestamp] = useState(null);
  const [fileContent, setFileContent] = useState(null);
  const [showDiff, setShowDiff] = useState(false);
  const [selectedTimestamp2, setSelectedTimestamp2] = useState(null);
  const [diffContent, setDiffContent] = useState(null);

  // Fetch all hosts when component mounts
  useEffect(() => {
    fetchHosts();
  }, []);

  const fetchHosts = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/host/all`);
      if (response.ok) {
        const hostsData = await response.json();
        setHosts(hostsData);
      }
    } catch (error) {
      console.error('Error fetching hosts:', error);
    }
  };

  const handleFileUpload = () => {
    // Refresh hosts list after upload
    fetchHosts();
  };

  const handleHostSelect = async (host) => {
    setSelectedHost(host);
    setSelectedTimestamp(null);
    setFileContent(null);
    setShowDiff(false);
    setSelectedTimestamp2(null);
    setDiffContent(null);
    
    try {
      const response = await fetch(`${API_BASE_URL}/host?ip=${host}`);
      if (response.ok) {
        const timestampsData = await response.json();
        setTimestamps(timestampsData);
      }
    } catch (error) {
      console.error('Error fetching timestamps:', error);
    }
  };

  const handleTimestampSelect = async (timestamp) => {
    setSelectedTimestamp(timestamp);
    setShowDiff(false);
    setSelectedTimestamp2(null);
    setDiffContent(null);
    
    try {
      const response = await fetch(`${API_BASE_URL}/snapshot?ip=${selectedHost}&at=${timestamp}`);
      if (response.ok) {
        const content = await response.json();
        setFileContent(content);
      }
    } catch (error) {
      console.error('Error fetching file content:', error);
    }
  };

  const handleShowDiff = () => {
    setShowDiff(true);
  };

  const handleTimestamp2Select = async (timestamp2) => {
    setSelectedTimestamp2(timestamp2);
    
    if (selectedTimestamp && timestamp2) {
      try {
        const response = await fetch(
          `${API_BASE_URL}/snapshot/diff?ip=${selectedHost}&t1=${selectedTimestamp}&t2=${timestamp2}`
        );
        if (response.ok) {
          const diffData = await response.json();
          setDiffContent(diffData);
        }
      } catch (error) {
        console.error('Error fetching diff:', error);
      }
    }
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Host Snapshot Manager</h1>
      </header>
      
      <main className="App-main">
        <div className="upload-section">
          <FileUpload onUpload={handleFileUpload} />
        </div>
        
        <div className="content-section">
          <div className="left-panel">
            <HostsList 
              hosts={hosts} 
              selectedHost={selectedHost}
              onHostSelect={handleHostSelect} 
            />
            
            {selectedHost && (
              <TimestampsView 
                timestamps={timestamps}
                selectedTimestamp={selectedTimestamp}
                onTimestampSelect={handleTimestampSelect}
                onShowDiff={handleShowDiff}
                showDiff={showDiff}
              />
            )}
          </div>
          
          <div className="right-panel">
            {fileContent && (
              <FileViewer 
                content={fileContent}
                host={selectedHost}
                timestamp={selectedTimestamp}
              />
            )}
            
            {showDiff && selectedHost && (
              <DiffViewer 
                timestamps={timestamps}
                selectedTimestamp={selectedTimestamp}
                selectedTimestamp2={selectedTimestamp2}
                onTimestamp2Select={handleTimestamp2Select}
                diffContent={diffContent}
              />
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;