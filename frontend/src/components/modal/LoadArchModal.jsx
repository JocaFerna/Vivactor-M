import React, { useState } from 'react';
import { X, CheckCircle2, Loader2 } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';
import { motion } from "motion/react";


// 2. Define the logic (Keep it inside or move to a separate service file)
  const fetchSoftware = async (repoUrl) => {
    try {
        const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL
        // 1. Build the URL with the correct port
        const params = new URLSearchParams({ url: repoUrl });
        const fullUrl = `${API_BASE}/cloneRepository?${params.toString()}`;

        // 2. Make the call
        const response = await fetch(fullUrl);
        
        if (!response.ok) throw new Error(`Server error: ${response.status}`);
        
        useGlobalStore.setState({ architectureURL: repoUrl }); // Store the URL in global state for later use
        console.log("Repository cloning initiated, waiting for response...");
        const result = await response.json();
        console.log("Repository cloned successfully:", result);

        // 2. FIX: Access the function via getState() instead of a Hook
        const fetchGraph = useGlobalStore.getState().fetchGraphData;
        
        if (fetchGraph) {
            await fetchGraph();
        } else {
            console.error("fetchGraphData is not defined in the store");
        }
      } catch (error) {
        console.error("Connection to :8000 failed:", error);
        throw error; // Rethrow to be caught in the component
      }
};

const LoadArchModal = ({ isOpen, onClose }) => {
  const [url, setUrl] = useState('');
  const [status, setStatus] = useState('idle'); // idle | loading | success

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setStatus('loading');

    

    fetchSoftware(url)
      .then(() => setStatus('success'))
      .catch(() => {
        setStatus('failed');
      });
  };

  const handleClose = () => {
    setStatus('idle');
    setUrl('');
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md p-6 relative">
        <button onClick={handleClose} className="absolute right-4 top-4 text-slate-400 hover:text-slate-600">
          <X size={20} />
        </button>

        {status === 'success' ? (
          <div className="text-center py-4 animate-in zoom-in duration-300">
            <CheckCircle2 size={48} className="text-green-500 mx-auto mb-4" />
            <h2 className="text-xl font-bold text-slate-800">Architecture Loaded</h2>
            <p className="text-slate-600 mt-2">Your microservices map is ready for refactoring.</p>
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
            <h2 className="text-xl font-bold text-slate-800">Failed to Load</h2>
            <p className="text-slate-600 mt-2">Please check your URL and try again.</p>
            <button onClick={handleClose} className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md hover:bg-slate-800">
              Close
            </button>
          </div>
          ) : (
          <>
            <h2 className="text-xl font-bold text-slate-800 mb-4">Load Architecture</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <input
                required
                type="url"
                placeholder="Enter Architecture URL..."
                className="w-full px-3 py-2 border border-slate-300 rounded-md focus:ring-2 focus:ring-blue-500 outline-none text-sm text-black"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
              />
              <button
                disabled={status === 'loading'}
                className="w-full bg-blue-600 text-white py-2 rounded-md flex items-center justify-center gap-2 hover:bg-blue-700 disabled:bg-blue-400"
              >
                {status === 'loading' && <Loader2 size={18} className="animate-spin" />}
                {status === 'loading' ? "Processing..." : "Confirm Load"}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  );
};

export default LoadArchModal;