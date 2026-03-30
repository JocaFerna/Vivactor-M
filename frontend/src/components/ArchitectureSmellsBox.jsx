import React, { useEffect, useState, useCallback, use } from 'react';
import { useGlobalStore } from '../store/useGlobalStore';

const ArchitectureSmellsBox = () => {
  // 1. Call hooks at the top level only
  const isArchitectureRunning = useGlobalStore((state) => state.isArchitectureRunning);
  const setIsRunning = (val) => useGlobalStore.setState({ isArchitectureRunning: val });

  const [hasError, setHasError] = useState(false);
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

  // 2. Memoize the fetch function to keep the effect clean
  const fetchSmells = useCallback(async () => {
    try {
      const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL || "http://localhost:8080";

      const params = new URLSearchParams({ graph: JSON.stringify(useGlobalStore.getState().graphData) });
      if (useGlobalStore.getState().graphData == undefined){
        return
      }
      const fullUrl = `${API_BASE}/smells/report?${params.toString()}`;

      
      const response = await fetch(fullUrl);

      if (!response.ok) throw new Error(`HTTP ${response.status}`);

      const result = await response.json();
      console.log("Fetched report:", result);


      // The following blocks will extract the relevant data for each smell and update the state accordingly.



      // Extract the relevant data -> nonAPIVerisoned
      const nonAPIVersioned = result.smells?.apiNonVersioned;
      
      const countAPIVersioned = (nonAPIVersioned !== undefined) 
        ? (Object.keys(nonAPIVersioned).length > 0 ? Object.keys(nonAPIVersioned).length : "Not Detected")
        : "N/A";

      setSmells(prev => ({ ...prev, nonAPIVersioned: countAPIVersioned }));
      setHasError(false);




      // Extract the relevant data -> Cyclic Dependency
      const cyclicDependency = result.smells?.cyclicDependency;

      const countCyclicDependency = (cyclicDependency !== undefined)
        ? (cyclicDependency.length > 0 ? cyclicDependency.length : "Not Detected")
        : "N/A";

      setSmells(prev => ({ ...prev, cyclicDependency: countCyclicDependency }));
      setHasError(false);


      // Extract the relevant data -> ESB Usage
      const esbUsage = result.smells?.esbUsage ? "Detected" : "Not Detected";
      
      setSmells(prev => ({ ...prev, esbUsage: esbUsage }));
      setHasError(false);


      // Extract the relevant data -> Hardcoded Endpoints
      const hardcodedEndpoints = result.smells?.hardcodedEndpoints;
      const countHardcodedEndpoints = (hardcodedEndpoints !== undefined)
        ? (hardcodedEndpoints.length > 0 ? hardcodedEndpoints.length : "Not Detected")
        : "N/A";
        
      setSmells(prev => ({ ...prev, hardcodedEndpoints: countHardcodedEndpoints }));
      setHasError(false);


      // Extract the relevant data -> Innapropriate Service Intimacity
      const inapropriateServiceIntimacity = result.smells?.innapropriateServiceIntimacity;
      const countInnapropriateServiceIntimacity = (inapropriateServiceIntimacity !== undefined)
        ? (inapropriateServiceIntimacity.length > 0 ? inapropriateServiceIntimacity.length : "Not Detected")
        : "N/A";
        
      setSmells(prev => ({ ...prev, innapropriateServiceIntimacity: countInnapropriateServiceIntimacity }));
      setHasError(false);

      // Extract the relevant data -> Microservice Greedy
      const microserviceGreedy = result.smells?.microserviceGreedy
      const countMicroserviceGreedy = (microserviceGreedy !== undefined)
        ? (microserviceGreedy.length > 0 ? microserviceGreedy.length : "Not Detected")
        : "N/A";
        
      setSmells(prev => ({ ...prev, microserviceGreedy: countMicroserviceGreedy }));
      setHasError(false);

      // Extract the relevant data -> Shared Libraries
      const sharedLibraries = result.smells?.sharedLibraries
      const countSharedLibraries = (sharedLibraries !== undefined)
        ? (sharedLibraries.length > 0 ? sharedLibraries.length : "Not Detected")
        : "N/A";
        
      setSmells(prev => ({ ...prev, sharedLibraries: countSharedLibraries }));
      setHasError(false);

      // Extract the relevant data -> Shared Persistency
      const sharedPersistency = result.smells?.sharedPersistency
      const countSharedPersistency = (sharedPersistency !== undefined)
        ? (sharedPersistency.length > 0 ? sharedPersistency.length : "Not Detected")
        : "N/A";
        
      setSmells(prev => ({ ...prev, sharedPersistency: countSharedPersistency }));
      setHasError(false);

      // Extract the relevant data -> Wrong Cuts
      const wrongCuts = Boolean(result.smells?.wrongCuts) ? "Detected" : "Not Detected";
      setSmells(prev => ({ ...prev, wrongCuts: wrongCuts }));
      setHasError(false);

      // Extract the relevant data -> Too Many Standards
      const tooManyStandards = result.smells?.tooManyStandards
      setSmells(prev => ({ ...prev, tooManyStandards: tooManyStandards }));
      setHasError(false);
      

      // Extract the relevant data -> No API Gateway
      const noAPIGateway = Boolean(result.smells?.noAPIGateway) ? "Detected" : "Not Detected or Not Applicable";
      setSmells(prev => ({ ...prev, noAPIGateway: noAPIGateway }));
      setHasError(false);

      // Check if any smells were detected to update the global state for refactoring
      if (result.smells) {
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
    <div className={`
      /* Positioning: Fixed to bottom-right, but centered/full-width on tiny screens */
      fixed bottom-4 right-4 left-4 
      sm:left-auto sm:bottom-6 sm:right-6 
      
      /* Sizing: Adjusts based on screen width */
      w-auto sm:w-72 md:w-80 
      max-h-[85vh] overflow-y-auto 
      
      /* Styling */
      bg-slate-900 border ${hasError ? 'border-red-500' : 'border-slate-700'} 
      rounded-lg shadow-2xl p-4 text-white z-50 transition-all duration-300
    `}>      
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
        // Changed items-center to items-start to better handle multi-line wrapping
        // Added gap-2 to keep spacing between label and value
        <div key={key} className="flex justify-between items-start text-xs gap-2">
          <span className="capitalize text-slate-400 shrink-0">
            {key.replace(/([A-Z])/g, ' $1').replace("A P I", "API")}:
          </span>
          
          <span className={`font-mono font-medium text-right ${hasError ? 'text-red-600' : 
            (value == "Detected" ? 'text-red-400' : 
            (value == "Non Detected" ? 'text-green-400' :
            (value == "Not Detected or Not Applicable" ? 'text-yellow-400' :
            (value > 0 ? 'text-red-400' : 'text-green-400'))))}`}>
            {hasError ? "ERROR" : value}
          </span>
        </div>
      ))}
    </div>
    </div>
  );
};

export default ArchitectureSmellsBox;