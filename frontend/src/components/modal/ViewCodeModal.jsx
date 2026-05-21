import React, { useState, useEffect } from 'react';
import { X, FileCode, Copy, Check } from 'lucide-react';
import { useGlobalStore } from '../../store/useGlobalStore';

const ViewCodeModal = ({ isOpen, onClose, selectedNode }) => {
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  const graphData = useGlobalStore((state) => state.graphData);

  // Determine file extension based on node language metadata
  const getFileName = (node) => {
    if (!node) return 'main.ext';
    const lang = node.properties?.language?.toLowerCase() || '';
    if (lang === 'java') return 'Main.java';
    if (lang === 'javascript' || lang === 'js') return 'main.js';
    if (lang === 'python' || lang === 'py') return 'main.py';
    return 'main.src';
  };

  useEffect(() => {
    if (!isOpen || !selectedNode) return;

    const fetchSourceCode = async () => {
      setLoading(true);
      try {
        const API_BASE = import.meta.env.VITE_ARCHITECTURAL_URL;
        const fileName = getFileName(selectedNode);
        
        const params = new URLSearchParams({graph: JSON.stringify(graphData), service: selectedNode.label});
        // Fetching the simulated source code from your Go backend orchestrator
        const response = await fetch(`${API_BASE}/getSourceCode?${params.toString()}`);
        if (!response.ok) throw new Error("Source file not found");
        
        const data = await response.text();
        setCode(data);
      } catch (err) {
        console.error(err);
        setCode(`// Error loading source code.\n// File ${getFileName(selectedNode)} could not be recovered from emulation layers.`);
      } finally {
        setLoading(false);
      }
    };

    fetchSourceCode();
  }, [isOpen, selectedNode]);

  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (!isOpen || !selectedNode) return null;

  return (
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="bg-slate-900 border border-slate-700 rounded-lg shadow-2xl w-full max-w-3xl flex flex-col h-[80vh]">
        
        {/* Modal Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-700 bg-slate-950 rounded-t-lg">
          <div className="flex items-center gap-2 text-white">
            <FileCode size={18} className="text-blue-400" />
            <div>
              <h2 className="text-sm font-bold truncate max-w-xs sm:max-w-md">
                {selectedNode.label || selectedNode.id}
              </h2>
              <p className="text-[10px] text-slate-400 font-mono mt-0.5">
                Inspecting: {getFileName(selectedNode)}
              </p>
            </div>
          </div>
          
          <div className="flex items-center gap-2">
            <button 
              onClick={handleCopy}
              className="p-1.5 text-slate-400 hover:text-white rounded hover:bg-slate-800 transition-colors"
              title="Copy Code"
            >
              {copied ? <Check size={16} className="text-green-400" /> : <Copy size={16} />}
            </button>
            <button onClick={onClose} className="p-1.5 text-slate-400 hover:text-white rounded hover:bg-slate-800 transition-colors">
              <X size={18} />
            </button>
          </div>
        </div>

        {/* Modal Content / Code Area */}
        <div className="flex-1 overflow-auto p-4 bg-slate-950 font-mono text-xs text-slate-300 leading-relaxed custom-scrollbar selection:bg-blue-500/30">
          {loading ? (
            <div className="h-full flex flex-col items-center justify-center gap-2 text-slate-400 italic">
              <div className="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
              <span>Recovering source template...</span>
            </div>
          ) : (
            <pre className="whitespace-pre-wrap break-all select-text">{code}</pre>
          )}
        </div>
      </div>
    </div>
  );
};

export default ViewCodeModal;