import React, { useState } from 'react';
import { motion } from "motion/react";
import { X, CheckCircle2, Loader2 } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';

// 2. Define the logic (Keep it inside or move to a separate service file)
const killSoftware = async () => {
    try {
        const graphData = useGlobalStore.getState().graphData
        const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL
        // 1. Build the URL with the correct port
        const params = new URLSearchParams({graph: JSON.stringify(graphData)});
        const fullUrl = `${API_BASE}/killArchitecture?${params.toString()}`;

        // 2. Make the call
        const response = await fetch(fullUrl);
        
        if (!response.ok){
          throw new Error(`Server error: ${response.status}`);
        } 
        
        //useGlobalStore.setState({ architectureURL: repoUrl }); // Store the URL in global state for later use
        useGlobalStore.setState({ isArchitectureRunning: false }); // Set the architecture as running in global state
        useGlobalStore.setState({ isEmulating: false }); // Set the emulation status in global state

        // Set all refactoring flags to false when killing the architecture
        useGlobalStore.setState({ refactoringOfNonAPIVersioned: false });
        useGlobalStore.setState({ refactoringOfCyclicDependency: false });
        useGlobalStore.setState({ refactoringOfEsbUsage: false });
        useGlobalStore.setState({ refactoringOfHardcodedEndpoints: false });
        useGlobalStore.setState({ refactoringOfInnapropriateServiceIntimacity: false });
        useGlobalStore.setState({ refactoringOfMicroserviceGreedy: false });
        useGlobalStore.setState({ refactoringOfSharedLibraries: false });
        useGlobalStore.setState({ refactoringOfSharedPersistency: false });
        useGlobalStore.setState({ refactoringOfWrongCuts: false });
        useGlobalStore.setState({ refactoringOfTooManyStandards: false });
        useGlobalStore.setState({ refactoringOfNoAPIGateway: false });

        // Set data related to refactorings to null as well
        useGlobalStore.setState({ refactoringOfNonAPIVersionedJSON: null });
        useGlobalStore.setState({ refactoringOfCyclicDependencyJSON: null });
        useGlobalStore.setState({ refactoringOfEsbUsageJSON: null });
        useGlobalStore.setState({ refactoringOfHardcodedEndpointsJSON: null });
        useGlobalStore.setState({ refactoringOfInnapropriateServiceIntimacityJSON: null });
        useGlobalStore.setState({ refactoringOfMicroserviceGreedyJSON: null });
        useGlobalStore.setState({ refactoringOfSharedLibrariesJSON: null });
        useGlobalStore.setState({ refactoringOfSharedPersistencyJSON: null });
        useGlobalStore.setState({ refactoringOfWrongCutsJSON: null });
        useGlobalStore.setState({ refactoringOfTooManyStandardsJSON: null });
        useGlobalStore.setState({ refactoringOfNoAPIGatewayJSON: null }); 

        const result = await response.json();
      } catch (error) {
        console.error("Connection to :8000 failed:", error);
        throw error; // Rethrow to be caught in the component
      }
};

const KillArchModal = ({ isOpen, onClose }) => {
  const [status, setStatus] = useState('idle'); // idle | loading | success

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setStatus('loading');

    

    killSoftware()
      .then(() => setStatus('success'))
      .catch(() => {
        setStatus('failed');
      });
  };

  const handleClose = () => {
    setStatus('idle');
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
      {/* Added min-h-[500px] and flex flex-col to the container */}
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md  p-6 relative flex flex-col z-50">
        <button onClick={handleClose} className="absolute right-4 top-4 text-slate-400 hover:text-slate-600">
          <X size={20} />
        </button>

        {status === 'success' ? (
          // Added flex-grow and justify-center to keep the success message centered in the taller modal
          <div className="text-center py-4 animate-in zoom-in duration-300 flex-grow flex flex-col justify-center">
            <CheckCircle2 size={48} className="text-green-500 mx-auto mb-4" />
            <h2 className="text-xl font-bold text-slate-800">Architecture Killed</h2>
            <p className="text-slate-600 mt-2">Your microservices map has been stopped.</p>
            <button onClick={handleClose} className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md hover:bg-slate-800">
              Got it
            </button>
          </div>
        ) : status === 'failed' ? (
          <div className="text-center py-4 animate-in zoom-in duration-300 flex-grow flex flex-col justify-center">
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{
                  type: "spring",
                  stiffness: 260,
                  damping: 20
                }}
              >
                <X size={200} className="text-red-500 mx-auto mb-8" />
              </motion.div>
            <h2 className="text-xl font-bold text-slate-800">Failed to Kill</h2>
            <p className="text-slate-600 mt-2">Please check your graph.</p>
            <button onClick={handleClose} className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md hover:bg-slate-800">
              Close
            </button>
          </div>
        ) : (
          <>
            <h2 className="text-xl font-bold text-slate-800 mb-4">Kill Architecture</h2>
            {/* Added flex-grow to the form so it fills the available space */}
            <form onSubmit={handleSubmit} className="space-y-4 flex-grow flex flex-col">
              {/* Labels for the textareas */}
              <label className="text-sm font-medium text-slate-700">This will kill the architecture, stopping all running services!</label>
              <button
                disabled={status === 'loading'}
                className="w-full bg-blue-600 text-white py-3 rounded-md flex items-center justify-center gap-2 hover:bg-blue-700 disabled:bg-blue-400 mt-auto"
              >
                {status === 'loading' && <Loader2 size={18} className="animate-spin" />}
                {status === 'loading' ? "Processing..." : "Confirm Kill"}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  );
};

export default KillArchModal;