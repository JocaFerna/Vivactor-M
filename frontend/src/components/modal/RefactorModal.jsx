import React, { useState, useEffect, useMemo } from 'react';
import { X, CheckCircle2, Loader2, AlertCircle, Hammer } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';
import { motion } from "motion/react";

/**
 * Logic to handle the API call. 
 */
const refactorSoftware = async (repoUrl, refactorType, selectedApis) => {
    try {
        const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL || "http://localhost:8080";
        const currentState = useGlobalStore.getState();
        const params = new URLSearchParams({ graph: JSON.stringify(currentState.graphData) });

        const getSelectedKeys = () => Object.keys(selectedApis).filter(api => selectedApis[api]);

        let endpoint = "";
        let paramKey = "";

        switch(refactorType) {
            case "nonAPIVersioned": endpoint = "mitigateNonAPIVersionedSmells"; paramKey = "apis"; break;
            case "hardcodedEndpoints": endpoint = "mitigateHardcodedEndpointsSmells"; paramKey = "endpoints"; break;
            case "microserviceGreedy": endpoint = "mitigateMicroserviceGreedySmells"; paramKey = "greedyMicroservices"; break;
            case "sharedLibraries": endpoint = "mitigateSharedLibrariesSmells"; paramKey = "sharedLibrariesSmells"; break;
            case "sharedPersistency": endpoint = "mitigateSharedPersistencySmells"; paramKey = "sharedPersistencySmells"; break;
            case "wrongCuts": endpoint = "mitigateWrongCutsSmells"; break;
            case "tooManyStandards": endpoint = "mitigateTooManyStandardsSmells"; paramKey = "tooManyStandardsSmells"; break;
            case "noAPIGateway": endpoint = "mitigateNonAPIGatewaySmells"; paramKey = "noAPIGatewaySmells"; break;
            default: return;
        }

        if (paramKey) params.append(paramKey, JSON.stringify(getSelectedKeys()));

        const fullUrl = `${API_BASE}/refactor/${endpoint}?${params.toString()}`;
        const response = await fetch(fullUrl);
        if (!response.ok) throw new Error(`Server error: ${response.status}`);
        
        const result = await response.json();
        
        useGlobalStore.setState({ 
            [`refactoringOf${refactorType.charAt(0).toUpperCase() + refactorType.slice(1)}`]: false, 
            [`refactoringOf${refactorType.charAt(0).toUpperCase() + refactorType.slice(1)}JSON`]: null, 
            isArchitectureRunning: true,           
            graphData: result.graph || currentState.graphData 
        });
    } catch (error) { throw error; }
};

const REFACTOR_KEY_MAP = {
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

const RefactorModal = ({ isOpen, onClose, typeOfRefactor }) => {
    const [status, setStatus] = useState('idle');
    const [selectedRefactors, setSelectedRefactors] = useState({});

    const smellData = useGlobalStore((state) => state[REFACTOR_KEY_MAP[typeOfRefactor]]);
    const nodes = useGlobalStore((state) => state.graphData?.nodes);

    // --- Derived List: Handles Label vs ID ---
    const list = useMemo(() => {
        if (typeOfRefactor === "noAPIGateway") {
            // Map to objects containing both ID and Label
            return nodes?.map(node => ({
                id: node.id,
                display: node.label || node.id 
            })) || [];
        }
        
        const rawList = Array.isArray(smellData) ? smellData : (smellData ? Object.keys(smellData) : []);
        return rawList.map(item => ({ id: item, display: item }));
    }, [smellData, nodes, typeOfRefactor]);

    const selectedCount = useMemo(() => 
        Object.values(selectedRefactors).filter(Boolean).length
    , [selectedRefactors]);

    useEffect(() => {
        if (isOpen && list.length > 0) {
            const initialState = {};
            list.forEach(item => { initialState[item.id] = false; });
            setSelectedRefactors(initialState);
            setStatus('idle');
        }
    }, [isOpen, list, typeOfRefactor]);

    if (!isOpen) return null;

    const handleToggle = (id) => {
        setSelectedRefactors(prev => ({ ...prev, [id]: !prev[id] }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setStatus('loading');
        try {
            await refactorSoftware('', typeOfRefactor, selectedRefactors);
            setStatus('success');
        } catch (err) { setStatus('failed'); }
    };

    const getTitle = () => {
        const titles = {
            nonAPIVersioned: "API Versioning",
            hardcodedEndpoints: "Endpoint Hardcoding",
            microserviceGreedy: "Service Greediness",
            sharedLibraries: "Library Coupling",
            sharedPersistency: "Database Sharing",
            noAPIGateway: "API Gateway Integration"
        };
        return titles[typeOfRefactor] || "Architecture Refactor";
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
            <motion.div 
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="bg-white rounded-lg shadow-xl w-full max-w-md p-6 relative"
            >
                <button onClick={onClose} className="absolute right-4 top-4 text-slate-400 hover:text-slate-600 transition-colors">
                    <X size={20} />
                </button>

                {status === 'success' ? (
                    <div className="text-center py-4">
                        <CheckCircle2 size={48} className="text-green-500 mx-auto mb-4" />
                        <h2 className="text-xl font-bold text-slate-800">Refactoring Successful</h2>
                        <p className="text-slate-600 mt-2">The architecture has been updated.</p>
                        <button onClick={onClose} className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md hover:bg-slate-800 transition-colors">Close</button>
                    </div>
                ) : status === 'failed' ? (
                    <div className="text-center py-4">
                        <AlertCircle size={60} className="text-red-500 mb-4 mx-auto" />
                        <h2 className="text-xl font-bold text-slate-800">Process Failed</h2>
                        <button onClick={() => setStatus('idle')} className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md transition-colors">Try Again</button>
                    </div>
                ) : (
                    <>
                        {/* --- NEW TITLE AND HEADER SECTION --- */}
                        <div className="flex items-center gap-2 mb-1">
                            <Hammer size={18} className="text-blue-600" />
                            <span className="text-xs font-bold uppercase tracking-wider text-blue-600">Refactoring Tool</span>
                        </div>
                        <h2 className="text-2xl font-black text-slate-900 mb-1">{getTitle()}</h2>
                        <p className="text-sm text-slate-500 mb-6 border-b pb-4">Select the components you wish to apply the architectural fix to.</p>

                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div className="max-h-60 overflow-y-auto p-2 border rounded-lg bg-slate-50 space-y-1 custom-scrollbar">
                                {list.length > 0 ? list.map((item) => (
                                    <label 
                                        key={item.id} 
                                        className="flex items-center gap-3 p-3 hover:bg-white rounded-md cursor-pointer transition-all border border-transparent hover:border-slate-200 group"
                                    >
                                        <input
                                            type="checkbox"
                                            className="w-4 h-4 accent-blue-600 cursor-pointer"
                                            checked={!!selectedRefactors[item.id]}
                                            onChange={() => handleToggle(item.id)}
                                        />
                                        <div className="flex flex-col truncate">
                                            <span className="text-sm font-semibold text-slate-800 group-hover:text-blue-700 transition-colors">
                                                {item.display}
                                            </span>
                                            {/* Show the ID as a small sub-label if it's different from display */}
                                            {item.display !== item.id && (
                                                <span className="text-[10px] text-slate-400 font-mono">ID: {item.id}</span>
                                            )}
                                        </div>
                                    </label>
                                )) : <p className="text-sm text-slate-500 italic p-4 text-center">No components detected for refactoring.</p>}
                            </div>
                            
                            <div className="flex items-center justify-between px-1">
                                <span className="text-xs text-slate-400">{list.length} components found</span>
                                <span className="text-xs font-bold text-blue-600">{selectedCount} selected</span>
                            </div>

                            <button
                                type="submit"
                                disabled={status === 'loading' || (list.length > 0 && selectedCount === 0)}
                                className="w-full bg-blue-600 text-white py-3 rounded-md flex items-center justify-center gap-2 hover:bg-blue-700 disabled:bg-slate-200 disabled:text-slate-400 transition-all font-bold shadow-lg shadow-blue-200"
                            >
                                {status === 'loading' ? (
                                    <>
                                        <Loader2 size={18} className="animate-spin" />
                                        <span>Applying Changes...</span>
                                    </>
                                ) : (
                                    "Apply Architectural Refactor"
                                )}
                            </button>
                        </form>
                    </>
                )}
            </motion.div>
        </div>
    );
};

export default RefactorModal;