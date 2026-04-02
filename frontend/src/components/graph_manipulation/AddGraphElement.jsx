import React, { useState, useEffect } from 'react';
import { useGlobalStore } from '../../store/useGlobalStore';

const AddGraphElement = () => {
    const [activeTab, setActiveTab] = useState('node');
    const { graphData } = useGlobalStore();

    // Node Form State
    const [nodeForm, setNodeForm] = useState({ 
        id: '', 
        label: '', 
        type: 'BasicNode', 
        language: 'java',
        orderOfMagnitudeOfFiles: '' 
    });

    // Edge Form State
    const [edgeForm, setEdgeForm] = useState({ 
        source: '', 
        target: '', 
        endpoint: '', 
        callDefinitionInSource: '',
        method: 'GET'
    });

    // System Form State
    const [systemForm, setSystemForm] = useState({
        name: '',
        description: '',
        selfManagedLibraries: [],
        servicesSeparatedByBusinessDomain: false
    });

    // --- Sync Logic ---
    // This updates the local form whenever graphData changes in the global store
    useEffect(() => {
        if (graphData?.systemContext) {
            setSystemForm({
                name: graphData.systemContext.name || '',
                description: graphData.systemContext.description || '',
                selfManagedLibraries: graphData.systemContext.selfManagedLibraries || [],
                servicesSeparatedByBusinessDomain: !!graphData.systemContext.servicesSeparatedByBusinessDomain
            });
        }
    }, [graphData]); 

    // --- Handlers ---

    const handleAddNode = (e) => {
        e.preventDefault();
        var newNode = {}

        if (nodeForm.type !== "DatabaseNode") {
            newNode = {
                id: nodeForm.id || `node-${Date.now()}`,
                label: nodeForm.label,
                type: nodeForm.type,
                properties: { 
                    language: nodeForm.language, 
                    orderOfMagnitudeOfFiles: nodeForm.orderOfMagnitudeOfFiles 
                }
            };
        }
        else {
            newNode = {
                id: nodeForm.id || `node-${Date.now()}`,
                label: nodeForm.label,
                type: nodeForm.type,
        
            }
        }
        useGlobalStore.setState({
            graphData: { ...graphData, nodes: [...graphData.nodes, newNode] }
        });
        setNodeForm({ id: '', label: '', type: 'BasicNode', language: 'java', orderOfMagnitudeOfFiles: '' });
    };

    const handleAddEdge = (e) => {
        e.preventDefault();
        const newEdge = {
            source: edgeForm.source,
            target: edgeForm.target,
            endpoint: edgeForm.endpoint,
            properties: { callDefinitionInSource: edgeForm.callDefinitionInSource, method: edgeForm.method }
        };

        useGlobalStore.setState({
            graphData: { ...graphData, edges: [...graphData.edges, newEdge] }
        });
        setEdgeForm({ source: '', target: '', endpoint: '', callDefinitionInSource: '', method: 'GET' });
    };

    const handleEditSystem = (e) => {
        e.preventDefault();
        useGlobalStore.setState({
            graphData: { 
                ...graphData, 
                systemContext: { ...systemForm } 
            }
        });
        alert("System configuration updated successfully!");
    };

    // --- System Context Helpers ---

    const addLibraryField = () => {
        setSystemForm(prev => ({
            ...prev,
            selfManagedLibraries: [...prev.selfManagedLibraries, { name: '', servicesUsingLibrary: [] }]
        }));
    };
    const removeLibraryField = (data) => {
        setSystemForm(prev => ({
            ...prev,
            selfManagedLibraries: [...prev.selfManagedLibraries, { name: '', servicesUsingLibrary: [] }]
        }));
    };


    const updateLibraryName = (index, name) => {
        const updatedLibs = [...systemForm.selfManagedLibraries];
        updatedLibs[index].name = name;
        setSystemForm({ ...systemForm, selfManagedLibraries: updatedLibs });
    };

    const addServiceToLibrary = (libIndex, serviceId) => {
        if (!serviceId) return;
        const updatedLibs = [...systemForm.selfManagedLibraries];
        if (!updatedLibs[libIndex].servicesUsingLibrary.includes(serviceId)) {
            updatedLibs[libIndex].servicesUsingLibrary.push(serviceId);
            setSystemForm({ ...systemForm, selfManagedLibraries: updatedLibs });
        }
    };

    const removeServiceFromLibrary = (libIndex, serviceId) => {
        const updatedLibs = [...systemForm.selfManagedLibraries];
        updatedLibs[libIndex].servicesUsingLibrary = updatedLibs[libIndex].servicesUsingLibrary.filter(id => id !== serviceId);
        setSystemForm({ ...systemForm, selfManagedLibraries: updatedLibs });
    };

    const removeLibrary = (index) => {
        const updatedLibs = systemForm.selfManagedLibraries.filter((_, i) => i !== index);
        setSystemForm({ ...systemForm, selfManagedLibraries: updatedLibs });
    };

    return (
        <div className="absolute top-4 right-4 z-50 p-4 bg-slate-800/95 backdrop-blur-sm border border-slate-700 rounded-lg shadow-xl w-80 text-white max-h-[90vh] overflow-y-auto">
            {/* Tabs Navigation */}
            <div className="flex mb-4 border-b border-slate-700">
                {['node', 'edge', 'system'].map((tab) => (
                    <button 
                        key={tab}
                        className={`pb-2 flex-1 capitalize text-sm transition-colors ${activeTab === tab ? 'border-b-2 border-blue-500 font-bold text-blue-400' : 'text-slate-400 hover:text-slate-200'}`}
                        onClick={() => setActiveTab(tab)}
                    >
                        {tab === 'system' ? 'System' : `+ ${tab}`}
                    </button>
                ))}
            </div>

            {/* Node Form */}
            {activeTab === 'node' && (
                <form onSubmit={handleAddNode} className="space-y-3 text-black">
                    <input 
                        className="w-full p-2 border rounded bg-white" 
                        placeholder="Label (Service Name)" 
                        value={nodeForm.label}
                        onChange={e => setNodeForm({...nodeForm, label: e.target.value, id: e.target.value.toLowerCase().replace(/\s+/g, '-')})}
                        required 
                    />
                    <div>
                        <label className="block text-xs font-medium text-slate-400 mb-1">Type</label>
                        <select className="w-full p-2 border rounded bg-white" value={nodeForm.type} onChange={e => setNodeForm({...nodeForm, type: e.target.value})}>
                            <option value="BasicNode">Basic Node</option>
                            <option value="DatabaseNode">Database</option>
                            <option value="ESB">ESB</option>
                            <option value="APIGateway">API Gateway</option>
                        </select>
                    </div>
                    {nodeForm.type !== "DatabaseNode" && 
                    <div>
                        <label className="block text-xs font-medium text-slate-400 mb-1">Language</label>
                        <select className="w-full p-2 border rounded bg-white" value={nodeForm.language} onChange={e => setNodeForm({...nodeForm, language: e.target.value})}>
                            <option value="java">Java</option>
                            <option value="python">Python</option>
                            <option value="javascript">JavaScript</option>
                            <option value="html">HTML</option>
                        </select>
                    </div>}
                    {nodeForm.type !== "DatabaseNode" &&
                        <input className="w-full p-2 border rounded bg-white" placeholder="Magnitude (10^0)" value={nodeForm.orderOfMagnitudeOfFiles} onChange={e => setNodeForm({...nodeForm, orderOfMagnitudeOfFiles: e.target.value})} pattern="10\^[0-9]+" required />
                    }
                    <button type="submit" className="w-full py-2 text-white bg-blue-600 rounded hover:bg-blue-700 font-medium transition-colors">Add Node</button>
                </form>
            )}

            {/* Edge Form */}
            {activeTab === 'edge' && (
                <form onSubmit={handleAddEdge} className="space-y-3 text-black">
                    <select className="w-full p-2 border rounded bg-white" value={edgeForm.source} onChange={e => setEdgeForm({...edgeForm, source: e.target.value})} required>
                        <option value="">Select Source</option>
                        {graphData?.nodes?.map(n => <option key={n.id} value={n.id}>{n.label}</option>)}
                    </select>
                    <select className="w-full p-2 border rounded bg-white" value={edgeForm.target} onChange={e => setEdgeForm({...edgeForm, target: e.target.value})} required>
                        <option value="">Select Target</option>
                        {graphData?.nodes?.map(n => <option key={n.id} value={n.id}>{n.label}</option>)}
                    </select>
                    <input className="w-full p-2 border rounded bg-white" placeholder="Endpoint (/api/v1)" value={edgeForm.endpoint} onChange={e => setEdgeForm({...edgeForm, endpoint: e.target.value})} required />
                    <input className="w-full p-2 border rounded bg-white" placeholder="URL of Call" value={edgeForm.callDefinitionInSource} onChange={e => setEdgeForm({...edgeForm, callDefinitionInSource: e.target.value})} required />
                    <select className="w-full p-2 border rounded bg-white" value={edgeForm.method} onChange={e => setEdgeForm({...edgeForm, method: e.target.value})}>
                        <option value="GET">GET</option>
                        <option value="POST">POST</option>
                        <option value="PUT">PUT</option>
                        <option value="DELETE">DELETE</option>
                        <option value="PATCH">PATCH</option>
                    </select>
                    <button type="submit" className="w-full py-2 text-white bg-green-600 rounded hover:bg-green-700 font-medium transition-colors">Connect Services</button>
                </form>
            )}

            {/* System Context Form */}
            {activeTab === 'system' && (
                <form onSubmit={handleEditSystem} className="space-y-4">
                    <div>
                        <label className="block text-xs font-medium text-slate-400 mb-1">System's Name</label>
                        <input className="w-full p-2 border rounded text-black bg-white" value={systemForm.name} onChange={e => setSystemForm({...systemForm, name: e.target.value})} required />
                    </div>
                    <div>
                        <label className="block text-xs font-medium text-slate-400 mb-1">Description</label>
                        <textarea className="w-full p-2 border rounded text-black bg-white text-sm" value={systemForm.description} onChange={e => setSystemForm({...systemForm, description: e.target.value})} rows={2} required />
                    </div>

                    <div className="flex items-center gap-2 p-2 bg-slate-700/30 rounded border border-slate-600">
                        <input type="checkbox" id="domainSeparated" className="w-4 h-4 accent-blue-500" checked={systemForm.servicesSeparatedByBusinessDomain} onChange={e => setSystemForm({...systemForm, servicesSeparatedByBusinessDomain: e.target.checked})} />
                        <label htmlFor="domainSeparated" className="text-xs font-medium text-slate-300 cursor-pointer">Separated by Business Domain?</label>
                    </div>

                    <div className="border-t border-slate-700 pt-3">
                        <div className="flex justify-between items-center mb-2">
                            <label className="text-xs font-bold text-blue-400 uppercase tracking-wider">Self-Managed Libraries</label>
                            <button type="button" onClick={addLibraryField} className="text-[10px] bg-blue-600/20 text-blue-400 border border-blue-400/30 px-2 py-0.5 rounded hover:bg-blue-600/40">+ Add</button>
                        </div>
                        
                        <div className="space-y-4">
                            {systemForm.selfManagedLibraries.map((lib, idx) => (
                                <div key={idx} className="relative p-2 bg-slate-900/50 rounded border border-slate-700 group">
                                    <button type="button" onClick={() => removeLibrary(idx)} className="absolute -top-1 -right-1 bg-red-500 text-white rounded-full w-4 h-4 text-[10px] flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">×</button>
                                    
                                    <input 
                                        placeholder="Library Name"
                                        className="w-full p-1 mb-2 text-xs rounded text-black bg-white"
                                        value={lib.name}
                                        onChange={e => updateLibraryName(idx, e.target.value)}
                                    />

                                    <div className="flex flex-wrap gap-1 mb-2">
                                        {lib.servicesUsingLibrary.map(serviceId => (
                                            <span key={serviceId} className="flex items-center gap-1 px-1.5 py-0.5 bg-blue-900/40 text-blue-300 border border-blue-800 rounded text-[9px]">
                                                {graphData?.nodes?.find(n => n.id === serviceId)?.label || serviceId}
                                                <button type="button" onClick={() => removeServiceFromLibrary(idx, serviceId)} className="text-blue-500 hover:text-red-400 font-bold">×</button>
                                            </span>
                                        ))}
                                    </div>

                                    <select 
                                        className="w-full p-1 text-[10px] rounded text-black bg-slate-100 border-none cursor-pointer"
                                        value=""
                                        onChange={(e) => addServiceToLibrary(idx, e.target.value)}
                                    >
                                        <option value="" disabled>+ Add Service to Library</option>
                                        {graphData?.nodes
                                            ?.filter(node => !lib.servicesUsingLibrary.includes(node.id))
                                            .map(node => (
                                                <option key={node.id} value={node.id}>{node.label}</option>
                                            ))
                                        }
                                    </select>
                                </div>
                            ))}
                            {systemForm.selfManagedLibraries.length === 0 && (
                                <p className="text-[10px] text-slate-500 italic text-center">No libraries added yet.</p>
                            )}
                        </div>
                    </div>

                    <button type="submit" className="w-full py-2 text-white bg-green-600 rounded hover:bg-green-700 font-medium shadow-lg transition-transform active:scale-95">
                        Save System Changes
                    </button>
                </form>
            )}
        </div>
    );
};

export default AddGraphElement;