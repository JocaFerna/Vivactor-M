import React, { useState, useEffect } from 'react';
import { X, CheckCircle2, Loader2 } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';
import { motion } from "motion/react";

/**
 * Logic to handle the API call. 
 * Note: This is a standard function, not a hook, so useGlobalStore.getState() is safe here.
 */
const refactorSoftware = async (repoUrl, refactorType, selectedApis) => {
    try {
        const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL || "http://localhost:8080";
        const currentState = useGlobalStore.getState();
        const params = new URLSearchParams({ graph: JSON.stringify(currentState.graphData) });

        switch(refactorType) {
            case "nonAPIVersioned":
                console.log("Initiating refactor for non-API versioned smells...");
                const apisToRefactor = Object.keys(selectedApis).filter(api => selectedApis[api]);
                params.append('apis', JSON.stringify(apisToRefactor));

                var fullUrl = `${API_BASE}/refactor/mitigateNonAPIVersionedSmells?${params.toString()}`;
                var response = await fetch(fullUrl);
                
                if (!response.ok) throw new Error(`Server error: ${response.status}`);
                
                var result = await response.json();
                console.log("Repository refactored successfully:", result);

                useGlobalStore.setState({ 
                    refactoringOfNonAPIVersioned: false, 
                    refactoringOfNonAPIVersionedJSON: null, 
                    isArchitectureRunning: true,           
                    graphData: result.graph ? result.graph : currentState.graphData 
                });
                break;
            
            // Placeholder cases for future implementation
            case "cyclicDependency":
            case "esbUsage":
            case "hardcodedEndpoints":
                console.log("Initiating refactor for hardcoded endpoints...");
                const hardcodedEndpointsToRefactor = Object.keys(selectedApis).filter(api => selectedApis[api]);
                params.append('endpoints', JSON.stringify(hardcodedEndpointsToRefactor));

                fullUrl = `${API_BASE}/refactor/mitigateHardcodedEndpointsSmells?${params.toString()}`;
                response = await fetch(fullUrl);
                
                if (!response.ok) throw new Error(`Server error: ${response.status}`);
                
                result = await response.json();
                console.log("Repository refactored successfully:", result);

                useGlobalStore.setState({ 
                    refactoringOfHardcodedEndpoints: false, 
                    refactoringOfHardcodedEndpointsJSON: null, 
                    isArchitectureRunning: true,           
                    graphData: result.graph ? result.graph : currentState.graphData 
                });
                break;
            case "innapropriateServiceIntimacity":
            case "microserviceGreedy":
            case "sharedLibraries":
            case "sharedPersistency":
                console.log("Initiating refactor for shared persistency...");
                const sharedPersistencyToRefactor = Object.keys(selectedApis).filter(api => selectedApis[api]);
                params.append('sharedPersistencySmells', JSON.stringify(sharedPersistencyToRefactor));

                fullUrl = `${API_BASE}/refactor/mitigateSharedPersistencySmells?${params.toString()}`;
                response = await fetch(fullUrl);
                
                if (!response.ok) throw new Error(`Server error: ${response.status}`);
                
                result = await response.json();
                console.log("Repository refactored successfully:", result);
                
                useGlobalStore.setState({
                    refactoringOfSharedPersistency: false,
                    refactoringOfSharedPersistencyJSON: null,
                    isArchitectureRunning: true,
                    graphData: result.graph ? result.graph : currentState.graphData
                });
                break;
            case "wrongCuts":
            case "tooManyStandards":
            case "noAPIGateway":
                console.log(`Refactor for ${refactorType} is not implemented yet.`);
                break;
            default:
                console.warn("Unknown refactor type:", refactorType);
        }
        
    } catch (error) {
        console.error("Refactor failed:", error);
        throw error;
    }
};

const RefactorModal = ({ isOpen, onClose, typeOfRefactor }) => {
    // --- 1. HOOKS (Must always be at the top level) ---
    const [status, setStatus] = useState('idle'); // idle | loading | success | failed
    const [selectedRefactors, setSelectedRefactors] = useState({});

    // Dynamic state selection using a single stable hook call
    const data = useGlobalStore((state) => {
        const keyMap = {
            nonAPIVersioned: 'refactoringOfNonAPIVersionedJSON',
            cyclicDependency: 'refactoringOfCyclicDependencyJSON',
            esbUsage: 'refactoringOfEsbUsageJSON',
            hardcodedEndpoints: 'refactoringOfHardcodedEndpointsJSON',
            innapropriateServiceIntimacity: 'refactoringOfInnapropriateServiceIntimacityJSON',
            microserviceGreedy: 'refactoringOfMicroserviceGreedyJSON',
            sharedLibraries: 'refactoringOfSharedLibrariesJSON',
            sharedPersistency: 'refactoringOfSharedPersistencyJSON',
            wrongCuts: 'refactoringOfWrongCutsJSON',
            tooManyStandards: 'refactoringOfTooManyStandardsJSON',
            noAPIGateway: 'refactoringOfNoAPIGatewayJSON'
        };
        return state[keyMap[typeOfRefactor]];
    });

    // --- 2. DERIVED STATE (Logic, not hooks) ---
    const list = Array.isArray(data) 
        ? data 
        : (data ? Object.keys(data) : []);

    const selectedCount = Object.values(selectedRefactors).filter(Boolean).length;

    // --- 3. EFFECTS ---
    useEffect(() => {
        if (isOpen) {
            const initialState = {};
            list.forEach(item => {
                initialState[item] = false;
            });
            setSelectedRefactors(initialState);
            setStatus('idle');
        }
    }, [isOpen, typeOfRefactor]); // list excluded to avoid infinite loops, typeOfRefactor included to reset on type change

    // --- 4. EARLY RETURN (Must be AFTER all hooks) ---
    if (!isOpen) return null;

    // --- 5. HANDLERS ---
    const handleToggle = (name) => {
        setSelectedRefactors((prev) => ({
            ...prev,
            [name]: !prev[name],
        }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        
        const hasSelection = Object.values(selectedRefactors).some(val => val === true);
        if (list.length > 0 && !hasSelection) {
            alert("Please select at least one object to refactor.");
            return;
        }

        setStatus('loading');
        try {
            await refactorSoftware('', typeOfRefactor, selectedRefactors);
            setStatus('success');
        } catch (err) {
            setStatus('failed');
        }
    };

    const handleClose = () => {
        setStatus('idle');
        setSelectedRefactors({});
        onClose();
    };

    // Helper to render dynamic titles
    const getTitle = () => {
        const titles = {
            nonAPIVersioned: "Select APIs to Refactor",
            cyclicDependency: "Select Cycles to Refactor",
            esbUsage: "Refactor ESB Services",
            hardcodedEndpoints: "Select Hardcoded Endpoints to Refactor",
            innapropriateServiceIntimacity: "Select Services with Inappropriate Intimacy",
            microserviceGreedy: "Select Greedy Microservices to Refactor",
            sharedLibraries: "Select Shared Libraries to Refactor",
            sharedPersistency: "Select Shared Persistency to Refactor",
            wrongCuts: "Refactor Wrong Service Cuts",
            tooManyStandards: "Refactor Too Many Standards",
            noAPIGateway: "Refactor No API Gateway"
        };
        return titles[typeOfRefactor] || "Refactor Architecture";
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
            <div className="bg-white rounded-lg shadow-xl w-full max-w-md p-6 relative ">
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
                            
                            <div className="flex flex-col gap-2 p-4 border rounded-lg bg-slate-50">
                                <h4 className="text-sm font-bold text-slate-700 border-b border-slate-200 pb-2 mb-2">
                                    {getTitle()}
                                </h4>

                                <div className="max-h-60 overflow-y-auto space-y-1 pr-2 custom-scrollbar">
                                    {list.length > 0 ? (
                                        list.map((name) => (
                                            <label 
                                                key={name} 
                                                className="flex items-center gap-3 p-2 hover:bg-white hover:shadow-sm rounded-md cursor-pointer transition-all border border-transparent hover:border-slate-200"
                                            >
                                                <input
                                                    type="checkbox"
                                                    className="w-4 h-4 accent-blue-600 cursor-pointer"
                                                    checked={!!selectedRefactors[name]}
                                                    onChange={() => handleToggle(name)}
                                                />
                                                <span className="text-sm text-slate-700 truncate">
                                                    {name}
                                                </span>
                                            </label>
                                        ))
                                    ) : (
                                        <p className="text-xs text-slate-500 italic">No objects detected for this smell.</p>
                                    )}
                                </div>
                                
                                {list.length > 0 && (
                                    <div className="mt-2 pt-2 border-t border-slate-200 text-xs font-medium text-slate-500 flex justify-between">
                                        <span>Total available: {list.length}</span>
                                        <span className={selectedCount > 0 ? "text-blue-600" : ""}>
                                            {selectedCount} selected
                                        </span>
                                    </div>
                                )}
                            </div>

                            <button
                                type="submit"
                                disabled={status === 'loading' || (list.length > 0 && selectedCount === 0)}
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