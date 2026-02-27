import React from 'react';
import { GraphCanvas } from 'reagraph';
import { useGlobalStore } from '../store/useGlobalStore';
import { use, useEffect, useState } from 'react';


const Graph = () => {
    // This component automatically re-renders when graphData in the store changes
    const graphData = useGlobalStore((state) => state.graphData);

    return (
        <div className="relative w-full h-full min-h-[400px]">
            {graphData.nodes.length === 0 ? (
                <div className="flex items-center justify-center h-full text-gray-500">No graph data available. Please load an architecture.</div>
            ) : (
                <GraphCanvas
                    nodes={graphData.nodes}
                    edges={graphData.edges}
                />
            )}
        </div>
    );
};

export default Graph;