import React, { useState, useEffect } from 'react';
import { X, CheckCircle2, Loader2 } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';
import { motion } from "motion/react";

/**
 * Logic to handle the API call.
 */
const refactorSoftware = async (repoUrl, refactorType, selectedApis) => {
    try {
        const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL;

        if (refactorType === "nonAPIVersioned") {
            console.log("Initiating refactor for non-API versioned smells...");

            // Filter local state to get only the names of checked APIs
            const apisToRefactor = Object.keys(selectedApis).filter(api => selectedApis[api]);
            
            // Get architecture URL from global store
            const architectureURL = useGlobalStore.getState().architectureURL;

            // 1. Build the URL
            const params = new URLSearchParams({ 
                url: architectureURL, 
                data: JSON.stringify(apisToRefactor) 
            });
            const fullUrl = `${API_BASE}/refactor/mitigateSharedLibrarySmells?${params.toString()}`;

            // 2. Make the call
            const response = await fetch(fullUrl);
            
            if (!response.ok) throw new Error(`Server error: ${response.status}`);
            
            const result = await response.json();
            console.log("Repository refactored successfully:", result);

            // 3. Update Global Store flags
            useGlobalStore.setState({ 
                refactoringOfNonAPIVersioned: false, // Reset flag
                isArchitectureRunning: true           // Set running state
            });

            // 4. Refresh the graph data
            const fetchGraph = useGlobalStore.getState().fetchGraphData;
            if (fetchGraph) {
                await fetchGraph();
            }
        }
    } catch (error) {
        console.error("Refactor failed:", error);
        throw error;
    }
};

const RefactorModal = ({ isOpen, onClose, typeOfRefactor }) => {
    const [status, setStatus] = useState('idle'); // idle | loading | success | failed
    const [selectedApis, setSelectedApis] = useState({});

    // 1. Selector: Get raw data. Fallback to empty array if null.
    const apiData = useGlobalStore((state) => state.refactoringOfNonAPIVersionedJSON);
    
    // Determine if data is an array or object keys
    const apiList = Array.isArray(apiData) 
        ? apiData 
        : (apiData ? Object.keys(apiData) : []);

    // 2. Initialization Effect: Runs only when modal OPENS
    // This breaks the "Maximum update depth exceeded" loop
    useEffect(() => {
        if (isOpen && apiList.length > 0) {
            const initialState = {};
            apiList.forEach(apiName => {
                initialState[apiName] = false;
            });
            setSelectedApis(initialState);
        }
    }, [isOpen]); // apiList is intentionally excluded to prevent re-initialization loops

    if (!isOpen) return null;

    const handleToggle = (apiName) => {
        setSelectedApis((prev) => ({
            ...prev,
            [apiName]: !prev[apiName],
        }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        
        const hasSelection = Object.values(selectedApis).some(val => val === true);
        if (typeOfRefactor === "nonAPIVersioned" && !hasSelection) {
            alert("Please select at least one API to refactor.");
            return;
        }

        setStatus('loading');

        try {
            // We use '' for url as repoUrl is pulled from store inside the helper
            await refactorSoftware('', typeOfRefactor, selectedApis);
            setStatus('success');
        } catch (err) {
            setStatus('failed');
        }
    };

    const handleClose = () => {
        setStatus('idle');
        setSelectedApis({});
        onClose();
    };

    const selectedCount = Object.values(selectedApis).filter(Boolean).length;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
            <div className="bg-white rounded-lg shadow-xl w-full max-w-md p-6 relative">
                <button 
                    onClick={handleClose} 
                    className="absolute right-4 top-4 text-slate-400 hover:text-slate-600 transition-colors"
                >
                    <X size={20} />
                </button>

                {status === 'success' ? (
                    <div className="text-center py-4 animate-in zoom-in duration-300">
                        <CheckCircle2 size={48} className="text-green-500 mx-auto mb-4" />
                        <h2 className="text-xl font-bold text-slate-800">Architecture Refactored</h2>
                        <p className="text-slate-600 mt-2">Your microservices have been refactored successfully.</p>
                        <button 
                            onClick={handleClose} 
                            className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md hover:bg-slate-800 transition-colors"
                        >
                            Got it
                        </button>
                    </div>
                ) : status === 'failed' ? (
                    <div className="text-center py-4 flex flex-col items-center">
                        <motion.div
                            initial={{ scale: 0 }}
                            animate={{ scale: 1 }}
                            transition={{ type: "spring", stiffness: 260, damping: 20 }}
                        >
                            <X size={80} className="text-red-500 mb-6" />
                        </motion.div>
                        <h2 className="text-xl font-bold text-slate-800">Failed to Refactor</h2>
                        <p className="text-slate-600 mt-2">Oopsie... Something went wrong.</p>
                        <button 
                            onClick={() => setStatus('idle')} 
                            className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md hover:bg-slate-800 transition-colors"
                        >
                            Try Again
                        </button>
                    </div>
                ) : (
                    <>
                        <h2 className="text-xl font-bold text-slate-800 mb-4">Refactor Architecture</h2>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            
                            {typeOfRefactor === "nonAPIVersioned" && (
                                <div className="flex flex-col gap-2 p-4 border rounded-lg bg-slate-50">
                                    <h4 className="text-sm font-bold text-slate-700 border-b border-slate-200 pb-2 mb-2">
                                        Select APIs to Refactor
                                    </h4>
                                    
                                    <div className="max-h-60 overflow-y-auto space-y-1 pr-2 custom-scrollbar">
                                        {apiList.length > 0 ? (
                                            apiList.map((apiName) => (
                                                <label 
                                                    key={apiName} 
                                                    className="flex items-center gap-3 p-2 hover:bg-white hover:shadow-sm rounded-md cursor-pointer transition-all border border-transparent hover:border-slate-200"
                                                >
                                                    <input
                                                        type="checkbox"
                                                        className="w-4 h-4 accent-blue-600 cursor-pointer"
                                                        checked={!!selectedApis[apiName]}
                                                        onChange={() => handleToggle(apiName)}
                                                    />
                                                    <span className="text-sm text-slate-700 truncate">
                                                        {apiName}
                                                    </span>
                                                </label>
                                            ))
                                        ) : (
                                            <p className="text-xs text-slate-500 italic">No APIs detected.</p>
                                        )}
                                    </div>
                                    
                                    <div className="mt-2 pt-2 border-t border-slate-200 text-xs font-medium text-slate-500 flex justify-between">
                                        <span>Total available: {apiList.length}</span>
                                        <span className={selectedCount > 0 ? "text-blue-600" : ""}>
                                            {selectedCount} selected
                                        </span>
                                    </div>
                                </div>
                            )}

                            <button
                                type="submit"
                                disabled={status === 'loading' || (typeOfRefactor === "nonAPIVersioned" && selectedCount === 0)}
                                className="w-full bg-blue-600 text-white py-2.5 rounded-md flex items-center justify-center gap-2 hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed transition-colors font-medium"
                            >
                                {status === 'loading' && <Loader2 size={18} className="animate-spin" />}
                                {status === 'loading' ? "Processing..." : "Confirm Refactor"}
                            </button>
                        </form>
                    </>
                )}
            </div>
        </div>
    );
};

export default RefactorModal;