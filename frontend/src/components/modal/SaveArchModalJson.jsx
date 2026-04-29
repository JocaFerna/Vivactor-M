import React, { useState } from 'react';
import { X, CheckCircle2, Loader2, Download, FileJson, AlertTriangle } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';
import { motion } from "framer-motion";

const SaveArchModalJson = ({ isOpen, onClose }) => {
  const [status, setStatus] = useState('idle'); // idle | loading | success | failed
  const [fileName, setFileName] = useState('architecture_snapshot');

  if (!isOpen) return null;

  const handleSaveJson = () => {
    setStatus('loading');
    
    try {
      // 1. Grab the current graph data from the store
      const graphData = useGlobalStore.getState().graphData;

      if (!graphData || (Array.isArray(graphData.nodes) && graphData.nodes.length === 0)) {
        throw new Error("No data to save");
      }

      // 2. Create the JSON blob
      const jsonString = JSON.stringify(graphData, null, 2);
      const blob = new Blob([jsonString], { type: "application/json" });
      
      // 3. Create a temporary download link
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `${fileName || 'architecture'}.json`;
      
      // 4. Trigger the download
      document.body.appendChild(link);
      link.click();
      
      // 5. Cleanup
      document.body.removeChild(link);
      URL.revokeObjectURL(url);

      setTimeout(() => setStatus('success'), 600);
    } catch (error) {
      console.error("Save failed:", error);
      setStatus('failed');
    }
  };

  const handleClose = () => {
    setStatus('idle');
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="bg-white rounded-xl shadow-2xl w-full max-w-md p-6 relative overflow-hidden border border-slate-200 z-50">
        
        {/* Close Button */}
        <button onClick={handleClose} className="absolute right-4 top-4 text-slate-400 hover:text-slate-600 transition-colors">
          <X size={20} />
        </button>

        {status === 'success' ? (
          <div className="text-center py-6 animate-in zoom-in duration-300">
            <div className="bg-green-100 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4">
                <CheckCircle2 size={32} className="text-green-600" />
            </div>
            <h2 className="text-xl font-bold text-slate-800">File Exported</h2>
            <p className="text-slate-600 mt-2 text-sm">Your architecture has been saved to your computer.</p>
            <button onClick={handleClose} className="mt-8 w-full bg-blue-600 text-white py-2.5 rounded-lg hover:bg-blue-700 font-semibold transition-all">
              Return to Workspace
            </button>
          </div>
        ) : status === 'failed' ? (
          <div className="text-center py-6">
            <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }} className="bg-amber-100 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4">
                <AlertTriangle size={32} className="text-amber-600" />
            </motion.div>
            <h2 className="text-xl font-bold text-slate-800">Export Failed</h2>
            <p className="text-slate-600 mt-2 text-sm">There was no active architecture data found to export.</p>
            <button onClick={() => setStatus('idle')} className="mt-8 w-full bg-slate-900 text-white py-2.5 rounded-lg hover:bg-slate-800 transition-all">
              Go Back
            </button>
          </div>
        ) : (
          <>
            <div className="mb-6">
                <h2 className="text-xl font-bold text-slate-800">Export Architecture</h2>
                <p className="text-slate-500 text-sm">Save your current graph nodes and edges to a JSON file.</p>
            </div>
            
            <div className="space-y-5">
              {/* Filename Input */}
              <div>
                <label className="text-[10px] font-bold text-slate-400 uppercase tracking-widest ml-1">File Name</label>
                <div className="mt-1 relative">
                  <input 
                    type="text" 
                    value={fileName}
                    onChange={(e) => setFileName(e.target.value)}
                    placeholder="architecture_snapshot"
                    className="w-full text-slate-700 bg-slate-50 border border-slate-200 rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                  />
                  <span className="absolute right-3 top-2.5 text-slate-400 text-sm">.json</span>
                </div>
              </div>

              {/* Action: Save Button */}
              <div 
                onClick={handleSaveJson}
                className="group cursor-pointer border-2 border-dashed border-slate-200 rounded-xl p-8 flex flex-col items-center justify-center gap-3 hover:border-blue-500 hover:bg-blue-50/50 transition-all"
              >
                <div className="p-3 bg-blue-50 rounded-full text-blue-500 group-hover:bg-blue-600 group-hover:text-white transition-all">
                    {status === 'loading' ? <Loader2 className="animate-spin" /> : <Download size={28} />}
                </div>
                <div className="text-center">
                    <p className="text-sm font-bold text-slate-700">Download JSON</p>
                    <p className="text-xs text-slate-400 mt-1">Export graph state as a local file</p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <div className="flex-1 border-t border-slate-100"></div>
                <span className="text-[10px] font-bold text-slate-300 uppercase tracking-widest">Format Details</span>
                <div className="flex-1 border-t border-slate-100"></div>
              </div>

              {/* Format Info */}
              <div className="flex items-center gap-3 p-4 bg-slate-50 rounded-xl border border-slate-100">
                  <div className="p-2 bg-white text-slate-400 rounded-lg shadow-sm">
                      <FileJson size={18} />
                  </div>
                  <div>
                      <p className="text-[11px] font-bold text-slate-700">Standard Schema</p>
                      <p className="text-[10px] text-slate-400">Includes all nodes, edges, and positions.</p>
                  </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default SaveArchModalJson;