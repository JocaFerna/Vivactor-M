import React, { useMemo, useEffect, useCallback } from 'react';
import { GraphCanvas, lightTheme } from 'reagraph'; 
import { useGlobalStore } from '../store/useGlobalStore';

const NODE_ICONS = {
    DatabaseNode: 'https://cdn-icons-png.flaticon.com/512/9850/9850774.png',
    BasicNode: 'https://cdn-icons-png.flaticon.com/512/5968/5968267.png',
    java: 'https://cdn-icons-png.flaticon.com/512/226/226777.png',
    python: 'https://cdn-icons-png.flaticon.com/512/5968/5968350.png',
    javascript: 'https://cdn-icons-png.flaticon.com/512/5968/5968292.png',
    html: 'https://cdn-icons-png.flaticon.com/512/1051/1051277.png'
};

const Graph = () => {
    const irData = useGlobalStore((state) => state.graphData);

    /*useEffect(() => {
        if (!irData || !irData.nodes) {
            const fetchData = async () => {
                try {
                    const response = await fetch('/IR_Graph_Example/ir_example.json');
                    const data = await response.json();
                    useGlobalStore.setState({ graphData: data });
                } catch (error) {
                    console.error("Failed to load IR graph data:", error);
                }
            };
            fetchData();
        }
    }, [irData]);*/

    // --- DELETE HANDLERS ---
    const deleteNode = useCallback((nodeId) => {
        const newNodes = irData.nodes.filter(n => n.id !== nodeId);
        const newEdges = irData.edges.filter(e => e.source !== nodeId && e.target !== nodeId);
        useGlobalStore.setState({ graphData: { ...irData, nodes: newNodes, edges: newEdges } });
    }, [irData]);

    const deleteEdge = useCallback((edge) => {
        console.log(edge)
        const newEdges = irData.edges.filter(e => 
            !(e.source === edge.source && e.target === edge.target && e.endpoint === edge.label)
        );
        useGlobalStore.setState({ graphData: { ...irData, edges: newEdges } });
    }, [irData]);

    // --- DATA FORMATTING ---
    const { nodes, edges } = useMemo(() => {
        if (!irData?.nodes) return { nodes: [], edges: [] };
        const formattedNodes = irData.nodes.map((node) => ({
            id: node.id,
            label: node.label,
            icon: NODE_ICONS[node.properties?.language] || NODE_ICONS[node.type] || NODE_ICONS.BasicNode,
        }));
        const formattedEdges = irData.edges.map((edge, idx) => ({
            id: `edge-${idx}`,
            source: edge.source,
            target: edge.target,
            label: edge.endpoint,
        }));
        return { nodes: formattedNodes, edges: formattedEdges };
    }, [irData]);

    const customTheme = useMemo(() => ({
        ...lightTheme, 
        canvas: { ...lightTheme.canvas, background: '#ffffff' },
        node: { ...lightTheme.node, label: { ...lightTheme.node.label, color: '#1e293b' } },
        edge: { ...lightTheme.edge, label: { ...lightTheme.edge.label, color: '#64748b', fontSize: 6 } }
    }), []);

    return (
        <div className="relative w-full h-full min-h-[500px] rounded-lg overflow-hidden bg-[#f8fafc]">
            {nodes.length === 0 ? (
                <div className="flex items-center justify-center h-full text-gray-400 italic">
                    Please load an architecture or add some nodes!
                </div>
            ) : (
                <GraphCanvas
                    nodes={nodes}
                    edges={edges}
                    theme={customTheme}
                    imageStrategy="node"
                    labelType="all"
                    nodeLabelPosition="bottom" 
                    edgeLabelPosition="above"
                    contextMenu={({ data, onClose }) => {
                        // Determine if we are clicking a node or an edg
                        const isNode = data.source === undefined
                        
                        return (
                            <div className="bg-white shadow-xl border border-slate-200 rounded-md py-1 min-w-[160px] text-slate-800">
                                {/* Header */}
                                <div className="px-3 py-2 border-b border-slate-100 mb-1">
                                    <p className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                                        {isNode ? 'Service' : 'Connection'}
                                    </p>
                                    <p className="text-xs font-semibold truncate">
                                        {data.label || 'Unnamed Link'}
                                    </p>
                                </div>

                                {/* Actions */}
                                <button 
                                    className="w-full text-left px-3 py-2 text-xs text-red-600 hover:bg-red-50 flex items-center gap-2 transition-colors"
                                    onClick={() => {
                                        if (isNode) deleteNode(data.id);
                                        else deleteEdge(data);
                                        onClose();
                                    }}
                                >
                                    <span>🗑️</span> Delete {isNode ? 'Node' : 'Edge'}
                                </button>

                                <button 
                                    className="w-full text-left px-3 py-2 text-xs text-slate-500 hover:bg-slate-50 border-t border-slate-50 mt-1"
                                    onClick={onClose}
                                >
                                    <span>✕</span> Close Menu
                                </button>
                            </div>
                        );
                    }}
                />
            )}
        </div>
    );
};

export default Graph;