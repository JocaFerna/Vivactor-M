
import React, { useEffect, useState } from 'react';
import Sidebar from './components/Sidebar';
import Graph from './components/Graph';
import ArchitectureSmellsBox from './components/ArchitectureSmellsBox';

function App() {
  // 1. Create state to hold the data and the loading status
  const [repoData, setRepoData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  return (
    <div className="flex">
      <Sidebar />
      <main className="flex-1 h-screen overflow-y-auto bg-white p-4">
        <Graph />
        <ArchitectureSmellsBox />
      </main>
    </div>
  );
}

export default App;