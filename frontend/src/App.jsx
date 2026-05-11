import React, { useState } from 'react';
import Sidebar from './components/Sidebar';
import Graph from './components/Graph';
import ArchitectureSmellsBox from './components/ArchitectureSmellsBox';
import AddGraphElement from './components/graph_manipulation/AddGraphElement';
import UpdateArchModal from './components/modal/UpdateArchModal';
import { useGlobalStore } from './store/useGlobalStore';
import { motion, AnimatePresence } from "motion/react";
import { Loader2, Activity, PowerOff, ShieldCheck } from "lucide-react";

/**
 * FIXED: Reactive Indicator Component
 */
const UpdatingIndicator = () => {
  // This hook ensures the component re-renders when the value changes
  const isUpdating = useGlobalStore((state) => state.updatingArchitecture);

  return (
    <AnimatePresence>
      {isUpdating && (
        <motion.div
          initial={{ y: -20, opacity: 0, x: "-50%" }}
          animate={{ y: 0, opacity: 1, x: "-50%" }}
          exit={{ y: -20, opacity: 0, x: "-50%" }}
          className="fixed top-6 left-1/2 z-[100] flex items-center gap-3 px-4 py-2 
                     bg-slate-900/80 backdrop-blur-md border border-blue-500/30 
                     rounded-full shadow-lg shadow-blue-500/20 text-white"
        >
          <div className="relative flex items-center justify-center">
            <div className="absolute inset-0 bg-blue-500/20 rounded-full blur-md animate-pulse" />
            <Loader2 size={16} className="text-blue-400 animate-spin" />
          </div>

          <div className="flex flex-col">
            <span className="text-[11px] font-bold uppercase tracking-widest text-blue-400 leading-none">
              Synchronizing
            </span>
            <span className="text-[10px] text-slate-300 font-medium">
              Updating Emulation...
            </span>
          </div>

          <span className="flex h-2 w-2 ml-1">
            <span className="animate-ping absolute inline-flex h-2 w-2 rounded-full bg-blue-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
          </span>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

const EmulationStatusBadge = () => {
  const isEmulating = useGlobalStore((state) => state.isEmulating);

  return (
    <div className="fixed top-6 left-76 z-[40] transition-all duration-500">
      <motion.div
        initial={false}
        animate={{
          backgroundColor: isEmulating ? "rgba(16, 185, 129, 0.1)" : "rgba(100, 116, 139, 0.1)",
          borderColor: isEmulating ? "rgba(16, 185, 129, 0.4)" : "rgba(100, 116, 139, 0.2)",
        }}
        className={`flex items-center gap-2 px-3 py-1.5 rounded-full border backdrop-blur-md shadow-sm`}
      >
        <div className="relative flex items-center justify-center">
          {isEmulating ? (
            <>
              <span className="absolute h-2 w-2 rounded-full bg-emerald-500 animate-ping opacity-75"></span>
              <Activity size={14} className="text-emerald-500 relative z-10" />
            </>
          ) : (
            <PowerOff size={14} className="text-slate-400" />
          )}
        </div>

        <span className={`text-[10px] font-bold uppercase tracking-widest ${
          isEmulating ? "text-emerald-500" : "text-slate-400"
        }`}>
          {isEmulating ? "Live Emulation" : "System Standby"}
        </span>

        {isEmulating && (
          <div className="h-3 w-[1px] bg-emerald-500/30 mx-1" />
        )}
        
        {isEmulating && (
          <span className="text-[9px] text-emerald-400/80 font-medium">
            Docker Active
          </span>
        )}
      </motion.div>
    </div>
  );
};

function App() {
  // Note: If you are moving graphData to useGlobalStore, 
  // you might eventually want to remove these local states to avoid confusion.
  const [repoData, setRepoData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  return (
    <div className="flex">
      <Sidebar />
      
      <main className="flex-1 h-screen overflow-y-auto bg-white p-4 relative">
        {/* Persistent Status Badge */}
        <EmulationStatusBadge />

        <Graph />
        
        <UpdateArchModal />
        
        {/* The fixed, reactive indicator */}
        <UpdatingIndicator />

        {/* Right Sidebar UI Stack */}
        <div className="fixed top-4 bottom-4 right-4 w-80 flex flex-col gap-4 z-50 pointer-events-none">
          {/* The Form (Add Node/Edge/System) */}
          <div className="pointer-events-auto flex-shrink-0">
            <AddGraphElement />
          </div>

          {/* The Smells Box (Now scrolls/flexes properly) */}
          <div className="pointer-events-auto flex-grow overflow-hidden flex flex-col">
            <ArchitectureSmellsBox />
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;