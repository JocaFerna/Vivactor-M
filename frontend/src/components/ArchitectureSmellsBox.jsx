import React, { useEffect, useState, useCallback, use } from 'react';
import { useGlobalStore } from '../store/useGlobalStore';
import { ChevronDown, ChevronUp, Info } from 'lucide-react'; // Optional: Use lucide-react for icons

const ArchitectureSmellsBox = () => {
  const isArchitectureRunning = useGlobalStore((state) => state.isArchitectureRunning);
  const setIsRunning = (val) => useGlobalStore.setState({ isArchitectureRunning: val });

  const [hasError, setHasError] = useState(false);
  const [expandedSmells, setExpandedSmells] = useState({}); // Track expanded states

  const [smells, setSmells] = useState({ 
    nonAPIVersioned: "N/A",
    cyclicDependency: "N/A",
    esbUsage: "N/A",
    hardcodedEndpoints: "N/A",
    innapropriateServiceIntimacity: "N/A",
    microserviceGreedy: "N/A",
    sharedLibraries: "N/A",
    sharedPersistency: "N/A",
    wrongCuts: "N/A",
    tooManyStandards: "N/A",
    noAPIGateway: "N/A"
  });

  const [smellsData, setSmellsData] = useState({});

  const toggleInfo = (key) => {
    setExpandedSmells(prev => ({
      ...prev,
      [key]: !prev[key]
    }));
  };

  const fetchSmells = useCallback(async () => {
    try {
      const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL || "http://localhost:8080";
      const graphData = useGlobalStore.getState().graphData;
      if (!graphData) return;

      const params = new URLSearchParams({ graph: JSON.stringify(graphData) });
      const fullUrl = `${API_BASE}/smells/report?${params.toString()}`;
      
      const response = await fetch(fullUrl);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const result = await response.json();

      // Mapping Logic (Condensed for readability)
      const updates = {};
      const dataUpdates = {};

      const processSmell = (key, rawData, isBoolean = false) => {
        if (isBoolean) {
          updates[key] = rawData ? "Detected" : "Not Detected";
          useGlobalStore.setState({ [`refactoringOf${key.charAt(0).toUpperCase() + key.slice(1)}`]: !!rawData });
        } else {
          const count = rawData ? (Array.isArray(rawData) ? rawData.length : Object.keys(rawData).length) : 0;
          updates[key] = count > 0 ? count : "Not Detected";
          dataUpdates[key] = count > 0 ? rawData : null;
          useGlobalStore.setState({ [`refactoringOf${key.charAt(0).toUpperCase() + key.slice(1)}JSON`]: rawData });
          useGlobalStore.setState({ [`refactoringOf${key.charAt(0).toUpperCase() + key.slice(1)}`]: !!rawData && (isBoolean ? rawData === true : count > 0) });
        }
        
      };

      processSmell('nonAPIVersioned', result.smells?.apiNonVersioned);
      processSmell('cyclicDependency', result.smells?.cyclicDependency);
      processSmell('esbUsage', result.smells?.esbUsage, true);
      processSmell('hardcodedEndpoints', result.smells?.hardcodedEndpoints);
      processSmell('innapropriateServiceIntimacity', result.smells?.innapropriateServiceIntimacity);
      processSmell('microserviceGreedy', result.smells?.microserviceGreedy);
      processSmell('sharedLibraries', result.smells?.sharedLibraries);
      processSmell('sharedPersistency', result.smells?.sharedPersistency);
      processSmell('wrongCuts', result.smells?.wrongCuts, true);
      processSmell('tooManyStandards', result.smells?.tooManyStandards);
      processSmell('noAPIGateway', result.smells?.noAPIGateway, true);

      setSmells(updates);
      setSmellsData(dataUpdates);
      setHasError(false);

      if (result.smells) {
        setIsRunning(false);
      }
    } catch (error) {
      console.error("Failed to fetch report:", error);
      setHasError(true);
    }
  }, [setIsRunning]);

  useEffect(() => {
    let interval = null;
    if (isArchitectureRunning) {
      fetchSmells();
      interval = setInterval(fetchSmells, 5000);
    }
    return () => interval && clearInterval(interval);
  }, [isArchitectureRunning, fetchSmells]);

  // Helper to render formatted info
  const renderInfoContent = (data) => {
    if (!data) return null;
    if (Array.isArray(data)) {
      return (
        <ul className="list-disc list-inside">
          {data.map((item, i) => (
            <li key={i} className="truncate" title={typeof item === 'string' ? item : JSON.stringify(item)}>
              {typeof item === 'string' ? item : JSON.stringify(item)}
            </li>
          ))}
        </ul>
      );
    }
    if (typeof data === 'object') {
      return Object.entries(data).map(([k, v]) => (
        <div key={k} className="mb-1">
          <span className="text-blue-400">{k}:</span> {JSON.stringify(v)}
        </div>
      ));
    }
    return String(data);
  };

  return (
    <div className={`
      fixed bottom-4 right-4 left-4 sm:left-auto sm:bottom-6 sm:right-4
      w-80 
      max-h-[500px] flex flex-col
      bg-slate-900 border ${hasError ? 'border-red-500' : 'border-slate-700'} 
      rounded-lg shadow-2xl text-white z-50 transition-all duration-300
    `}>  
      {/* Header - Stays fixed at top of box */}
      <div className="flex items-center justify-between p-4 border-b border-slate-700 bg-slate-900 rounded-t-lg">
        <h3 className={`text-sm font-bold uppercase tracking-wider ${hasError ? 'text-red-500' : 'text-red-400'}`}>
          Architecture Smells
        </h3>
        <div className="flex items-center gap-2">
          {isArchitectureRunning && <div className="h-2 w-2 bg-red-500 rounded-full animate-pulse"></div>}
          <div className={`h-2 w-2 rounded-full ${hasError ? 'bg-red-500' : 'bg-gray-500'}`}></div>
        </div>
      </div>
      
      {/* Scrollable Content Area */}
      <div className="overflow-y-auto p-4 space-y-4 custom-scrollbar">
        {Object.entries(smells).map(([key, value]) => (
          <div key={key} className="border-b border-slate-800 pb-2 last:border-0">
            <div className="flex justify-between items-center text-xs">
              <div className="flex items-center gap-2">
                <span className="capitalize text-slate-400">
                  {key.replace(/([A-Z])/g, ' $1').replace("A P I", "API")}:
                </span>
                {smellsData[key] && (
                  <button 
                    onClick={() => toggleInfo(key)}
                    className="text-slate-500 hover:text-blue-400 transition-colors"
                  >
                    <Info size={14} />
                  </button>
                )}
              </div>
              
              <span className={`font-mono font-medium ${
                value === "Detected" || (typeof value === 'number' && value > 0) ? 'text-red-400' : 
                value === "Not Detected" || value === "Not Detected or Not Applicable" ? 'text-green-400' : 'text-yellow-400'
              }`}>
                {hasError ? "ERROR" : value}
              </span>
            </div>

            {/* Expandable Info Section */}
            {expandedSmells[key] && smellsData[key] && (
              <div className="mt-2 text-[10px] text-slate-300 bg-slate-800/50 p-2 rounded border border-slate-700 font-mono overflow-x-auto">
                {renderInfoContent(smellsData[key])}
              </div>
            )}
          </div>
        ))}
      </div>

      <style jsx>{`
        .custom-scrollbar::-webkit-scrollbar { width: 6px; }
        .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
        .custom-scrollbar::-webkit-scrollbar-thumb { background: #334155; border-radius: 10px; }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover { background: #475569; }
      `}</style>
    </div>
  );
};

export default ArchitectureSmellsBox;