import React, { useState, useRef } from 'react';
import { X, CheckCircle2, Loader2, Upload, FileCode } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';
import { motion } from "framer-motion";

const LoadArchJsonModal = ({ isOpen, onClose }) => {
  const [status, setStatus] = useState('idle'); // idle | loading | success | failed
  const fileInputRef = useRef(null);

  if (!isOpen) return null;

  // --- 1. Fetching the static example file ---
  const handleFetchExample = async () => {
    setStatus('loading');
    try {
      const response = await fetch('/IR_Graph_Example/ir_example.json');
      if (!response.ok) throw new Error("Example file not found");
      const data = await response.json();
      
      useGlobalStore.setState({ graphData: data });
      setTimeout(() => setStatus('success'), 500);
    } catch (error) {
      console.error(error);
      setStatus('failed');
    }
  };

  // --- 2. Handling manual File Upload ---
  const handleFileUpload = (e) => {
    const file = e.target.files[0];
    if (!file) return;

    setStatus('loading');
    const reader = new FileReader();
    
    reader.onload = (event) => {
      try {
        const json = JSON.parse(event.target.result);
        useGlobalStore.setState({ graphData: json });
        setTimeout(() => setStatus('success'), 600);
      } catch (err) {
        console.error("Invalid JSON format", err);
        setStatus('failed');
      }
    };

    reader.onerror = () => setStatus('failed');
    reader.readAsText(file);
  };

  const handleClose = () => {
    setStatus('idle');
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-white rounded-xl shadow-2xl w-full max-w-md p-6 relative overflow-hidden border border-slate-200">
        
        {/* Close Button */}
        <button onClick={handleClose} className="absolute right-4 top-4 text-slate-400 hover:text-slate-600 transition-colors">
          <X size={20} />
        </button>

        {status === 'success' ? (
          <div className="text-center py-6 animate-in zoom-in duration-300">
            <div className="bg-green-100 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4">
                <CheckCircle2 size={32} className="text-green-600" />
            </div>
            <h2 className="text-xl font-bold text-slate-800">Architecture Mapped</h2>
            <p className="text-slate-600 mt-2 text-sm">The graph data has been loaded into the workspace.</p>
            <button onClick={handleClose} className="mt-8 w-full bg-blue-600 text-white py-2.5 rounded-lg hover:bg-blue-700 font-semibold transition-all shadow-md shadow-blue-200">
              Open Graph
            </button>
          </div>
        ) : status === 'failed' ? (
          <div className="text-center py-6">
            <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }} className="bg-red-100 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4">
                <X size={32} className="text-red-600" />
            </motion.div>
            <h2 className="text-xl font-bold text-slate-800">Load Error</h2>
            <p className="text-slate-600 mt-2 text-sm">Could not parse the JSON file. Please ensure it follows the correct schema.</p>
            <button onClick={() => setStatus('idle')} className="mt-8 w-full bg-slate-900 text-white py-2.5 rounded-lg hover:bg-slate-800 transition-all">
              Try Again
            </button>
          </div>
        ) : (
          <>
            <div className="mb-6">
                <h2 className="text-xl font-bold text-slate-800">Load Architecture JSON</h2>
                <p className="text-slate-500 text-sm">Select a method to populate your graph nodes and edges.</p>
            </div>
            
            <div className="space-y-4">
              {/* Option: Upload from Computer */}
              <div 
                onClick={() => fileInputRef.current?.click()}
                className="group cursor-pointer border-2 border-dashed border-slate-200 rounded-xl p-8 flex flex-col items-center justify-center gap-3 hover:border-blue-500 hover:bg-blue-50/50 transition-all"
              >
                <div className="p-3 bg-slate-50 rounded-full text-slate-400 group-hover:text-blue-500 group-hover:bg-blue-100 transition-all">
                    {status === 'loading' ? <Loader2 className="animate-spin" /> : <Upload size={28} />}
                </div>
                <div className="text-center">
                    <p className="text-sm font-bold text-slate-700">Upload JSON file</p>
                    <p className="text-xs text-slate-400 mt-1">Drag and drop or click to browse</p>
                </div>
                <input 
                    type="file" 
                    ref={fileInputRef} 
                    className="hidden" 
                    accept=".json" 
                    onChange={handleFileUpload} 
                />
              </div>

              <div className="flex items-center gap-3">
                <div className="flex-1 border-t border-slate-100"></div>
                <span className="text-[10px] font-bold text-slate-300 uppercase tracking-widest">or</span>
                <div className="flex-1 border-t border-slate-100"></div>
              </div>

              {/* Option: Quick-Load Example */}
              <button
                onClick={handleFetchExample}
                disabled={status === 'loading'}
                className="w-full flex items-center justify-between p-4 border border-slate-200 rounded-xl hover:border-slate-400 hover:bg-slate-50 transition-all group"
              >
                <div className="flex items-center gap-3 text-left">
                    <div className="p-2 bg-blue-50 text-blue-600 rounded-lg group-hover:bg-blue-600 group-hover:text-white transition-all">
                        <FileCode size={20} />
                    </div>
                    <div>
                        <p className="text-sm font-bold text-slate-700">Load Example Graph</p>
                        <p className="text-[10px] text-slate-400">/IR_Graph_Example/ir_example.json</p>
                    </div>
                </div>
                <span className="text-slate-300 group-hover:text-slate-600 transition-colors">→</span>
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default LoadArchJsonModal;