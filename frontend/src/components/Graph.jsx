import React, { useMemo, useCallback, useState } from 'react';

import { GraphCanvas, lightTheme } from 'reagraph'; 
import { useGlobalStore } from '../store/useGlobalStore';
import { triggerArchUpdateWithoutModal } from "./modal/UpdateArchModal"; // Ensure the update logic is imported so it can be triggered from the context menu
import ViewCodeModal from './modal/ViewCodeModal'; // Import the code inspection modal


const NODE_ICONS = {
    DatabaseNode: 'https://cdn-icons-png.flaticon.com/512/9850/9850774.png',
    BasicNode: 'https://cdn-icons-png.flaticon.com/512/5968/5968267.png',
    
    // Standard Languages
    java: 'https://cdn-icons-png.flaticon.com/512/226/226777.png',
    python: 'https://cdn-icons-png.flaticon.com/512/5968/5968350.png',
    javascript: 'https://cdn-icons-png.flaticon.com/512/5968/5968292.png',
    html: 'https://cdn-icons-png.flaticon.com/512/1051/1051277.png',

    // Specific Gateway Icons (Language + Gateway badge)
    APIGateway: 'https://cdn-icons-png.flaticon.com/512/1055/1055644.png', // Default Gateway Icon
    python_gateway: '/public/images/gateway_icons/python_api_gateway.png', // Custom Python Gateway Icon
    java_gateway: '/public/images/gateway_icons/java_api_gateway_icon.png',
    javascript_gateway: '/public/images/gateway_icons/javascript_api_gateway.png',
};

const Graph = () => {
    const irData = useGlobalStore((state) => state.graphData);
    const [isCodeModalOpen, setIsCodeModalOpen] = useState(false);
    const [inspectedNode, setInspectedNode] = useState(null);
    const isEmulating = useGlobalStore((state) => state.isEmulating);

    // FIX: Using functional updates for Zustand to avoid dependency loops with irData
    const deleteNode = useCallback((nodeId) => {
        useGlobalStore.setState((state) => {
            const newNodes = state.graphData.nodes.filter(n => n.id !== nodeId);
            const newEdges = state.graphData.edges.filter(e => e.source !== nodeId && e.target !== nodeId);
            return { graphData: { ...state.graphData, nodes: newNodes, edges: newEdges }, updateSuggestion: true };
        });
        triggerArchUpdateWithoutModal();
    }, []);

    const deleteEdge = useCallback((edge) => {
        const newEdges = irData.edges.filter(e => 
            !(e.source === edge.source && e.target === edge.target && e.endpoint === edge.label)
        );
        useGlobalStore.setState({ graphData: { ...irData, edges: newEdges }, updateSuggestion: true }); // Set update suggestion to true after deleting an edge
        triggerArchUpdateWithoutModal();
    }, [irData]);

    const { nodes, edges } = useMemo(() => {
        if (!irData?.nodes) return { nodes: [], edges: [] };
        const formattedNodes = irData.nodes.map((node) => {
            // 1. Determine the icon key
            const lang = node.properties?.language;
            const isGateway = node.type === 'APIGateway';
            
            // Look for 'java_gateway', then 'java', then the type 'APIGateway', then 'BasicNode'
            const iconKey = isGateway 
                ? (`${lang}_gateway` in NODE_ICONS ? `${lang}_gateway` : 'APIGateway')
                : (lang || node.type);

            return {
                id: node.id,
                label: node.label,
                icon: NODE_ICONS[iconKey] || NODE_ICONS.BasicNode,
                // 2. Add visual "Gateway" indicator via color/size
                color: isGateway ? '#3b82f6' : undefined, // Blue glow for Gateways
                size: isGateway ? 15 : 10, // Make Gateways slightly larger
            };
        });
        const formattedEdges = irData.edges.map((edge, idx) => ({
            // FIX: Using source-target-endpoint as a unique ID. 
            // idx-based IDs cause NaN errors when the array order shifts during refactoring!
            id: `${edge.source}-${edge.target}-${idx}`, 
            source: edge.source,
            target: edge.target,
            label: edge.endpoint || '',
            size: 5, // Increase edge width
        }));
        return { nodes: formattedNodes, edges: formattedEdges };
    }, [irData]);

    const customTheme = useMemo(() => ({
        ...lightTheme, 
        canvas: { ...lightTheme.canvas, background: '#ffffff' },
        node: { 
            ...lightTheme.node, 
            label: { 
                ...lightTheme.node.label, 
                fontFamily: 'sans-serif', // Force a standard font
                color: '#1e293b', 
                fontSize: 10,
                offset: 18 
            } 
        },
        edge: {
            // Increase width of edge
            ...lightTheme.edge,
            color: '#cbd5e1', // Optional: making it slightly darker makes it easier to see
            label: { 
                ...lightTheme.edge.label, 
                color: '#64748b', 
                fontSize: 8,
                background: { fill: '#ffffff', opacity: 0.9 } // Makes edge text readable
            } 
        }
    }), []);

    return (
        <div className="relative w-full h-full min-h-[600px] rounded-lg overflow-hidden bg-[#f8fafc]">
            {nodes.length === 0 ? (
                <div className="flex items-center justify-center h-full text-gray-400 italic">
                    Please load an architecture or add some nodes!
                </div>
            ) : (
                <GraphCanvas
                    nodes={nodes}
                    edges={edges}
                    draggable={true}
                    theme={customTheme}
                    imageStrategy="node"
                    labelType="all"
                    nodeLabelPosition="bottom" 
                    edgeLabelPosition="above"
                    edgeInterpolation="curved" // Fixes overlapping edges
                    
                    layoutOverrides={{
                        // Increase this significantly (from -150 to -800+) 
                        // to force nodes to stay away from the center
                        nodeStrength: -150, 
                        
                        // Increase link distance so the "circle" of nodes is larger
                        linkDistance: 300, 
                        
                        // Helps nodes settle faster and spread out more before stopping
                        alphaDecay: 0.05, 
                        
                        // Decreasing this allows the "repulsion" to win over the "centering" force
                        centeringStrength: 0.1 
                    }}
                    contextMenu={({ data, onClose }) => {
                        // Determine if we are clicking a node or an edg
                        const isNode = data.source === undefined
                        // Look up raw store node to get access to properties metadata
                        const rawNode = isNode ? irData.nodes.find(n => n.id === data.id) : null;
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
                                {/* NEW FEATURE: Inspect Code Option */}
                                {isNode && isEmulating && (
                                    <button
                                        className="w-full text-left px-3 py-2 text-xs text-slate-700 hover:bg-slate-50 flex items-center gap-2 transition-colors"
                                        onClick={() => {
                                            setInspectedNode(rawNode);
                                            setIsCodeModalOpen(true);
                                            onClose();
                                        }}
                                    >
                                        <span>🔍</span> Inspect Source Code
                                    </button>
                                )}
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
            {/* Render the ViewCodeModal safely attached to the Graph view root */}
            <ViewCodeModal 
                isOpen={isCodeModalOpen} 
                onClose={() => {
                    setIsCodeModalOpen(false);
                    setInspectedNode(null);
                }} 
                selectedNode={inspectedNode} 
            />
        </div>
    );
}; 

export default Graph;