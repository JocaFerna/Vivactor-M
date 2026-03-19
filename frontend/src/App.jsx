
import React, { useEffect, useState } from 'react';
import Sidebar from './components/Sidebar';
import Graph from './components/Graph';
import ArchitectureSmellsBox from './components/ArchitectureSmellsBox';
import AddGraphElement from './components/graph_manipulation/AddGraphElement';

function App() {
  // 1. Create state to hold the data and the loading status
  const [repoData, setRepoData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Define graph data state
  const [graphData, setGraphData] = useState(null);

  return (
    <div className="flex">
      <Sidebar />
      <main className="flex-1 h-screen overflow-y-auto bg-white p-4">
        <Graph />
        <ArchitectureSmellsBox />
        <AddGraphElement />
        
      </main>
    </div>
  );
}

export default App;