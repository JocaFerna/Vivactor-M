import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useGlobalStore } from '../store/useGlobalStore';
import { Info } from 'lucide-react';

const ArchitectureSmellsBox = () => {
  // 1. Stable selectors from Zustand
  const isArchitectureRunning = useGlobalStore((state) => state.isArchitectureRunning);
  const graphData = useGlobalStore((state) => state.graphData);

  const [hasError, setHasError] = useState(false);
  const [expandedSmells, setExpandedSmells] = useState({});
  const [smellsData, setSmellsData] = useState({});
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

  // 2. A ref to prevent overlapping requests (Manual Lock)
  const isFetchingRef = useRef(false);

  const toggleInfo = (key) => {
    setExpandedSmells(prev => ({ ...prev, [key]: !prev[key] }));
  };

  const fetchSmells = useCallback(async () => {
    // Prevent overlapping calls if a request is already in flight
    if (isFetchingRef.current || !graphData) return;

    try {
      isFetchingRef.current = true;
      const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL || "http://localhost:8080";
      const params = new URLSearchParams({ graph: JSON.stringify(graphData) });
      const fullUrl = `${API_BASE}/smells/report?${params.toString()}`;

      const response = await fetch(fullUrl);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const result = await response.json();

      const updates = {};
      const dataUpdates = {};
      // Object to batch all Zustand updates into ONE single call
      const storeUpdates = {};

      const processSmell = (key, rawData, isBoolean = false) => {
        const storeKeyBase = key.charAt(0).toUpperCase() + key.slice(1);

        if (isBoolean) {
          if (key === 'wrongCuts') {
            updates[key] = rawData === "N/A" ? "Non Available" : rawData ? "Detected" : "Not Detected";
          } else {
            updates[key] = rawData ? "Detected" : "Not Detected";
          }
          if (updates[key] === "Non Available") {
            storeUpdates[`refactoringOf${storeKeyBase}`] = false;
          } else {
            storeUpdates[`refactoringOf${storeKeyBase}`] = !!rawData;
          }
        } else {
          if (key === 'tooManyStandards') {
            const count = rawData || 0;
            const isSmell = count > 3;
            updates[key] = isSmell ? `${count} Detected` : "Not Detected";
            dataUpdates[key] = isSmell ? rawData : null;
            storeUpdates[`refactoringOf${storeKeyBase}`] = isSmell;
            storeUpdates[`refactoringOf${storeKeyBase}JSON`] = rawData;
          }
           else {
            const count = rawData ? (Array.isArray(rawData) ? rawData.length : Object.keys(rawData).length) : 0;
            const isSmell = count > 0;
            updates[key] = isSmell ? count : "Not Detected";
            dataUpdates[key] = isSmell ? rawData : null;
            storeUpdates[`refactoringOf${storeKeyBase}JSON`] = rawData;
            storeUpdates[`refactoringOf${storeKeyBase}`] = isSmell;
          }
        }
      };

      if (result.smells) {
        processSmell('nonAPIVersioned', result.smells.apiNonVersioned);
        processSmell('cyclicDependency', result.smells.cyclicDependency);
        processSmell('esbUsage', result.smells.esbUsage, true);
        processSmell('hardcodedEndpoints', result.smells.hardcodedEndpoints);
        processSmell('innapropriateServiceIntimacity', result.smells.innapropriateServiceIntimacity);
        processSmell('microserviceGreedy', result.smells.microserviceGreedy);
        processSmell('sharedLibraries', result.smells.sharedLibraries);
        processSmell('sharedPersistency', result.smells.sharedPersistency);
        processSmell('wrongCuts', result.smells.wrongCuts, true);
        processSmell('tooManyStandards', result.smells.tooManyStandards);
        processSmell('noAPIGateway', result.smells.noAPIGateway, true);

        // Turn off running state if report is received
        storeUpdates.isArchitectureRunning = false;
      }

      // 3. Batch apply to Store (One re-render)
      useGlobalStore.setState(storeUpdates);
      
      // 4. Update local state
      setSmells(updates);
      setSmellsData(dataUpdates);
      setHasError(false);

    } catch (error) {
      console.error("Failed to fetch report:", error);
      setHasError(true);
    } finally {
      isFetchingRef.current = false;
    }
  }, [graphData]); // Now only recreates if graphData actually changes

  useEffect(() => {
    let interval = null;
    if (isArchitectureRunning) {
      // Execute once immediately
      fetchSmells();
      // Set interval for subsequent calls
      interval = setInterval(fetchSmells, 10000);
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [isArchitectureRunning, fetchSmells]);

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
      w-80 max-h-[500px] flex flex-col
      bg-slate-900 border ${hasError ? 'border-red-500' : 'border-slate-700'} 
      rounded-lg shadow-2xl text-white z-50 transition-all duration-300
    `}>  
      <div className="flex items-center justify-between p-4 border-b border-slate-700 bg-slate-900 rounded-t-lg">
        <h3 className={`text-sm font-bold uppercase tracking-wider ${hasError ? 'text-red-500' : 'text-red-400'}`}>
          Architecture Smells
        </h3>
        <div className="flex items-center gap-2">
          {isArchitectureRunning && <div className="h-2 w-2 bg-red-500 rounded-full animate-pulse"></div>}
          <div className={`h-2 w-2 rounded-full ${hasError ? 'bg-red-500' : 'bg-gray-500'}`}></div>
        </div>
      </div>
      
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
                value === "Not Detected" ? 'text-green-400' : 'text-yellow-400'
              }`}>
                {hasError ? "ERROR" : value}
              </span>
            </div>

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