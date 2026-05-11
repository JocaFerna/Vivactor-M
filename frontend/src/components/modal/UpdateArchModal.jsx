import React, { useState, useEffect } from 'react';
import { X, CheckCircle2, Loader2, RefreshCw, Activity, ZapOff } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';

// --- Global Tracking Variables (Outside the Component) ---
let updateTimeout = null;
let isExecuting = false;
let needsAnotherUpdate = false;

/**
 * Internal Executor: Handles the actual async loop
 */
const runUpdateSequence = async () => {
    // 1. If already busy, mark that we need to go again once finished and bail
    if (isExecuting) {
        needsAnotherUpdate = true;
        return;
    }

    isExecuting = true;
    needsAnotherUpdate = false;

    try {
        // Pass a no-op to updateArchitectureLogic for status updates
        await updateArchitectureLogic(() => {}); 
    } catch (error) {
        console.error("Sequential update failed:", error);
    } finally {
        isExecuting = false;
        
        // 2. If a request arrived while we were busy, trigger again immediately
        if (needsAnotherUpdate) {
            runUpdateSequence();
        } else {
            // 3. Only turn off the global indicator if no more updates are pending
            useGlobalStore.setState({ updatingArchitecture: false });
        }
    }
};

/**
 * Orchestration Logic: Sequence Kill then Emulate
 */
export const updateArchitectureLogic = async (setStatus) => {
    try {
        // Reset states and set indicator
        useGlobalStore.setState({ 
            isArchitectureRunning: false, 
            isEmulating: false,
            updatingArchitecture: true,
            // ... Reset all smells and JSON data as you have them ...
            refactoringOfNonAPIVersioned: false, refactoringOfCyclicDependency: false,
            refactoringOfEsbUsage: false, refactoringOfHardcodedEndpoints: false,
            refactoringOfInnapropriateServiceIntimacity: false, refactoringOfMicroserviceGreedy: false,
            refactoringOfSharedLibraries: false, refactoringOfSharedPersistency: false,
            refactoringOfWrongCuts: false, refactoringOfTooManyStandards: false,
            refactoringOfNoAPIGateway: false,
            refactoringOfNonAPIVersionedJSON: null, refactoringOfCyclicDependencyJSON: null,
            refactoringOfEsbUsageJSON: null, refactoringOfHardcodedEndpointsJSON: null,
            refactoringOfInnapropriateServiceIntimacityJSON: null, refactoringOfMicroserviceGreedyJSON: null,
            refactoringOfSharedLibrariesJSON: null, refactoringOfSharedPersistencyJSON: null,
            refactoringOfWrongCutsJSON: null, refactoringOfTooManyStandardsJSON: null,
            refactoringOfNoAPIGatewayJSON: null 
        });

        const graphData = useGlobalStore.getState().graphData;
        const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL;
        const params = new URLSearchParams({ graph: JSON.stringify(graphData) });

        // --- STEP 1: KILL ---
        setStatus('killing');
        const killResponse = await fetch(`${API_BASE}/killArchitecture?${params.toString()}`);
        if (!killResponse.ok) throw new Error("Failed to kill");

        await new Promise(resolve => setTimeout(resolve, 1000));

        // --- STEP 2: EMULATE ---
        setStatus('emulating');
        const emulateResponse = await fetch(`${API_BASE}/emulateArchitecture?${params.toString()}`);
        if (!emulateResponse.ok) throw new Error("Failed to emulate");

        useGlobalStore.setState({ 
            isArchitectureRunning: true, 
            isEmulating: true, 
            updateSuggestion: false, 
            updatingArchitecture: false 
        });
        setStatus('success');

    } catch (error) {
        console.error("Update sequence failed:", error);
        setStatus('failed');
        useGlobalStore.setState({ updatingArchitecture: false });
        throw error;
    }
};

/**
 * Trigger with Modal
 */
export const triggerArchUpdate = () => {
    useGlobalStore.getState().setUpdateModalOpen(true);
};

/**
 * Trigger without Modal (Optimized for rapid Graph changes)
 */
export const triggerArchUpdateWithoutModal = () => {
    // A. Debounce: Clear existing timer to wait for user to stop clicking
    if (updateTimeout) clearTimeout(updateTimeout);

    // B. Immediate Visual Feedback
    useGlobalStore.setState({ updatingArchitecture: true });

    // C. Set a 2-second "settle" period before attempting the Docker calls
    updateTimeout = setTimeout(() => {
        runUpdateSequence();
    }, 2000);
};

const UpdateArchModal = () => {
    const isOpen = useGlobalStore((state) => state.isUpdateModalOpen);
    const onClose = () => useGlobalStore.getState().setUpdateModalOpen(false);
    
    const [status, setStatus] = useState('idle'); // idle | killing | emulating | success | failed

    // Reset internal status when modal opens
    useEffect(() => {
        if (isOpen) setStatus('idle');
    }, [isOpen]);

    if (!isOpen) return null;

    const handleConfirm = () => {
        updateArchitectureLogic(setStatus);
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
            <div className="bg-white rounded-lg shadow-xl w-full max-w-md p-6 relative flex flex-col z-50 min-h-[350px]">
                <button onClick={onClose} className="absolute right-4 top-4 text-slate-400 hover:text-slate-600">
                    <X size={20} />
                </button>

                {status === 'success' ? (
                    <div className="text-center py-4 animate-in zoom-in duration-300 flex-grow flex flex-col justify-center">
                        <CheckCircle2 size={48} className="text-green-500 mx-auto mb-4" />
                        <h2 className="text-xl font-bold text-slate-800">Refresh Complete</h2>
                        <p className="text-slate-600 mt-2">Architecture was successfully killed and rebuilt.</p>
                        <button onClick={onClose} className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md hover:bg-slate-800">
                            Back to Graph
                        </button>
                    </div>
                ) : status === 'failed' ? (
                    <div className="text-center py-4 animate-in zoom-in duration-300 flex-grow flex flex-col justify-center">
                        <X size={48} className="text-red-500 mx-auto mb-4" />
                        <h2 className="text-xl font-bold text-slate-800">Update Failed</h2>
                        <p className="text-slate-600 mt-2">Could not synchronize with the Docker daemon.</p>
                        <button onClick={() => setStatus('idle')} className="mt-6 w-full bg-slate-900 text-white py-2 rounded-md">
                            Try Again
                        </button>
                    </div>
                ) : (
                    <>
                        <div className="flex items-center gap-3 mb-4">
                            <div className="p-2 bg-blue-100 rounded-lg text-blue-600">
                                <RefreshCw size={24} className={status !== 'idle' ? 'animate-spin' : ''} />
                            </div>
                            <h2 className="text-xl font-bold text-slate-800">Update Environment</h2>
                        </div>

                        <div className="flex-grow space-y-6">
                            <p className="text-sm text-slate-600">
                                This will perform a full reset: stopping all active services and re-applying the current graph state.
                            </p>

                            {/* Status Steps */}
                            <div className="space-y-3">
                                <div className={`flex items-center gap-3 text-sm ${status === 'killing' ? 'text-blue-600 font-bold' : 'text-slate-400'}`}>
                                    {status === 'killing' ? <Loader2 size={16} className="animate-spin" /> : <ZapOff size={16} />}
                                    Shutting down old services...
                                </div>
                                <div className={`flex items-center gap-3 text-sm ${status === 'emulating' ? 'text-blue-600 font-bold' : 'text-slate-400'}`}>
                                    {status === 'emulating' ? <Loader2 size={16} className="animate-spin" /> : <Activity size={16} />}
                                    Building new containers...
                                </div>
                            </div>
                        </div>

                        <button
                            onClick={handleConfirm}
                            disabled={status !== 'idle'}
                            className="w-full bg-blue-600 text-white py-3 rounded-md flex items-center justify-center gap-2 hover:bg-blue-700 disabled:bg-blue-300 mt-6 font-bold"
                        >
                            {status === 'idle' ? "Confirm Full Update" : "Processing Sequence..."}
                        </button>
                    </>
                )}
            </div>
        </div>
    );
};

export default UpdateArchModal;