import React, { useEffect, useState, useCallback, use } from 'react';
import { useGlobalStore } from '../store/useGlobalStore';

const ArchitectureSmellsBox = () => {
  // 1. Call hooks at the top level only
  const isArchitectureRunning = useGlobalStore((state) => state.isArchitectureRunning);
  const setIsRunning = (val) => useGlobalStore.setState({ isArchitectureRunning: val });

  const [hasError, setHasError] = useState(false);
  const [smells, setSmells] = useState({ 
    nonAPIVersioned: "N/A",
    cyclicDependency: "N/A"
  });

  // 2. Memoize the fetch function to keep the effect clean
  const fetchSmells = useCallback(async () => {
    try {
      const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL || "http://localhost:8080";

      const params = new URLSearchParams({ graph: JSON.stringify(useGlobalStore.getState().graphData) });
      if (useGlobalStore.getState().graphData == undefined){
        return
      }
      const fullUrl = `${API_BASE}/smells/cyclicDependency?${params.toString()}`;

      
      const response = await fetch(fullUrl);

      if (!response.ok) throw new Error(`HTTP ${response.status}`);

      const result = await response.json();
      console.log("Fetched report:", result);


      // The following blocks will extract the relevant data for each smell and update the state accordingly.



      // Extract the relevant data -> nonAPIVerisoned
      const nonAPIVersioned = result.smells?.nonAPIVersionedEndpoints;
      
      const countAPIVersioned = (nonAPIVersioned !== undefined) 
        ? (Object.keys(nonAPIVersioned).length > 0 ? Object.keys(nonAPIVersioned).length : "Not Detected")
        : "N/A";

      setSmells(prev => ({ ...prev, nonAPIVersioned: countAPIVersioned }));
      setHasError(false);




      // Extract the relevant data -> Cyclic Dependency
      const cyclicDependency = result.smells?.cycles;

      const countCyclicDependency = (cyclicDependency !== undefined)
        ? (cyclicDependency.length > 0 ? cyclicDependency.length : "Not Detected")
        : "N/A";

      setSmells(prev => ({ ...prev, cyclicDependency: countCyclicDependency }));
      setHasError(false);


      

      // Stop polling if we actually got data
      if (countAPIVersioned !== "N/A" && countCyclicDependency !== "N/A") {
        setIsRunning(false);
        if (countAPIVersioned > 0) {
          useGlobalStore.setState({ refactoringOfNonAPIVersioned: true });
          useGlobalStore.setState({ refactoringOfNonAPIVersionedJSON: result.smells });
        }
      }

    } catch (error) {
      console.error("Failed to fetch MSANose report:", error);
      setHasError(true);
    }
  }, []);

  // 3. Effect handles the lifecycle of the interval
  useEffect(() => {
    let interval = null;

    if (isArchitectureRunning) {
      fetchSmells(); // Initial call
      interval = setInterval(fetchSmells, 5000);
    } else {
      setHasError(false);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [isArchitectureRunning, fetchSmells]);

  return (
    <div className={`fixed bottom-6 right-6 w-72 bg-slate-900 border ${hasError ? 'border-red-500' : 'border-slate-700'} rounded-lg shadow-2xl p-4 text-white z-50 transition-colors duration-300`}>
      <div className="flex items-center justify-between mb-3 border-b border-slate-700 pb-2">
        <h3 className={`text-sm font-bold uppercase tracking-wider ${hasError ? 'text-red-500' : 'text-red-400'}`}>
          Architecture Smells
        </h3>
        <div className="flex items-center gap-2">
          {hasError && <span className="text-[10px] text-red-500 font-bold">FETCH ERROR</span>}
          {isArchitectureRunning ? (
            <div className="h-2 w-2 bg-red-500 rounded-full animate-pulse"></div>
          ) : (
            <div className="h-2 w-2 bg-gray-500 rounded-full"></div>
          )}
        </div>
      </div>
      
      <div className="space-y-3">
        {Object.entries(smells).map(([key, value]) => (
          <div key={key} className="flex justify-between items-center text-xs">
            <span className="capitalize text-slate-400">
              {key.replace(/([A-Z])/g, ' $1')}:
            </span>
            <span className={`font-mono font-medium ${hasError ? 'text-red-600' : (value > 0 ? 'text-red-400' : 'text-green-400')}`}>
              {hasError ? "ERROR" : value}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ArchitectureSmellsBox;