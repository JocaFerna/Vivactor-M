import React, { useState } from 'react';
import { motion } from "motion/react";
import { X, CheckCircle2, Loader2 } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';

// 2. Define the logic (Keep it inside or move to a separate service file)
const startSoftware = async (command, packageList) => {
    try {
        const repoUrl = useGlobalStore.getState().architectureURL;
        const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL
        // 1. Build the URL with the correct port
        const params = new URLSearchParams({ url: repoUrl, command: command, packages: packageList });
        const fullUrl = `${API_BASE}/startArchitecture?${params.toString()}`;

        // 2. Make the call
        const response = await fetch(fullUrl);
        
        if (!response.ok){
          throw new Error(`Server error: ${response.status}`);
        } 
        
        useGlobalStore.setState({ architectureURL: repoUrl }); // Store the URL in global state for later use
        useGlobalStore.setState({ isArchitectureRunning: true }); // Set the architecture as running in global state

        const result = await response.json();
      } catch (error) {
        console.error("Connection to :8000 failed:", error);
        throw error; // Rethrow to be caught in the component
      }
};

const StartArchModal = ({ isOpen, onClose }) => {
  const [command, setCommand] = useState('');
  const [status, setStatus] = useState('idle'); // idle | loading | success
  const [packageList, setPackageList] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setStatus('loading');

    

    startSoftware(command,packageList)
      .then(() => setStatus('success'))
      .catch(() => {
        setStatus('failed');
      });
  };

  const handleClose = () => {
    setStatus('idle');
    setCommand('');
    setPackageList('');
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
      {/* Added min-h-[500px] and flex flex-col to the container */}
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md min-h-[500px] p-6 relative flex flex-col">
        <button onClick={handleClose} className="absolute right-4 top-4 text-slate-400 hover:text-slate-600">
          <X size={20} />
        </button>

        {status === 'success' ? (
          // Added flex-grow and justify-center to keep the success message centered in the taller modal
          <div className="text-center py-4 animate-in zoom-in duration-300 flex-grow flex flex-col justify-center">
            <CheckCircle2 size={48} className="text-green-500 mx-auto mb-4" />
            <h2 className="text-xl font-bold text-slate-800">Architecture Started</h2>
            <p className="text-slate-600 mt-2">Your microservices map is running.</p>
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
            <h2 className="text-xl font-bold text-slate-800">Failed to Start</h2>
            <p className="text-slate-600 mt-2">Please check your instructions and try again.</p>
            <button onClick={handleClose} className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md hover:bg-slate-800">
              Close
            </button>
          </div>
        ) : (
          <>
            <h2 className="text-xl font-bold text-slate-800 mb-4">Start Architecture</h2>
            {/* Added flex-grow to the form so it fills the available space */}
            <form onSubmit={handleSubmit} className="space-y-4 flex-grow flex flex-col">
              {/* Labels for the textareas */}
              <label className="text-sm font-medium text-slate-700">Packages and Dependencies</label>
              <textarea
                placeholder="Enter The Needed Packages and Dependencies to BUILD the architecture, separated by ENTER... (maven, python, npm, etc.) Docker is not needed."
                /* flex-grow here makes the text area take up all remaining space in the modal */
                className="w-full px-3 py-3 border border-slate-300 rounded-md focus:ring-2 focus:ring-blue-500 outline-none text-sm text-black flex-grow resize-none"
                value={packageList}
                onChange={(e) => setPackageList(e.target.value)}
              />
              <label className="text-sm font-medium text-slate-700">Custom Start-up Instructions</label>
              <textarea
                placeholder="Enter Custom Start-up Instructions, separated by ENTER... (docker commands, shell commands, etc.)"
                /* flex-grow here makes the text area take up all remaining space in the modal */
                className="w-full px-3 py-3 border border-slate-300 rounded-md focus:ring-2 focus:ring-blue-500 outline-none text-sm text-black flex-grow resize-none"
                value={command}
                onChange={(e) => setCommand(e.target.value)}
              />
              <button
                disabled={status === 'loading'}
                className="w-full bg-blue-600 text-white py-3 rounded-md flex items-center justify-center gap-2 hover:bg-blue-700 disabled:bg-blue-400 mt-auto"
              >
                {status === 'loading' && <Loader2 size={18} className="animate-spin" />}
                {status === 'loading' ? "Processing..." : "Confirm Start"}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  );
};

export default StartArchModal;