import { create } from 'zustand'

export const useGlobalStore = create((set) => ({
    architectureURL: null,
    // Read graph data from json file and store it in the global state
    graphData: {nodes:[],edges:[]},
    graphJSONRaw : null,
    isFetching: false,
    isArchitectureRunning: false,
    isEmulating: false,

    // Refactoring flags
    refactoringOfNonAPIVersioned: false,
    refactoringOfCyclicDependency: false,
    refactoringOfEsbUsage: false,
    refactoringOfHardcodedEndpoints: false,
    refactoringOfInnapropriateServiceIntimacity: false,
    refactoringOfMicroserviceGreedy: false,
    refactoringOfSharedLibraries: false,
    refactoringOfSharedPersistency: false,
    refactoringOfWrongCuts: false,
    refactoringOfTooManyStandards: false,
    refactoringOfNoAPIGateway: false,

    // Refactoring data.
    refactoringOfNonAPIVersionedJSON: null,
    refactoringOfCyclicDependencyJSON: null,
    refactoringOfEsbUsageJSON: null,
    refactoringOfHardcodedEndpointsJSON: null,
    refactoringOfInnapropriateServiceIntimacityJSON: null,
    refactoringOfMicroserviceGreedyJSON: null,
    refactoringOfSharedLibrariesJSON: null,
    refactoringOfSharedPersistencyJSON: null,
    refactoringOfWrongCutsJSON: null,
    refactoringOfTooManyStandardsJSON: null,
    refactoringOfNoAPIGatewayJSON: null,

    // This is the function you can call from ANY file
    fetchGraphData: async () => {
        set({ isFetching: true });
        try {
            const repoUrl = useGlobalStore.getState().architectureURL;
            const directory_name = repoUrl.substring(repoUrl.lastIndexOf('/') + 1).replace('.git', '');
            
            const API_BASE_GRAPH = import.meta.env.VITE_CODE2DFD_URL;
            const params = new URLSearchParams({ url: `downloads/${directory_name}` });
            
            const response = await fetch(`${API_BASE_GRAPH}/dfd_local?${params.toString()}`);
            const result = await response.json();

            // Parsing logic inside the store
            const nodes = Object.keys(result.traceability_file.nodes).map(key => ({
                id: key,
                label: key
            }));

            const edges = Object.keys(result.traceability_file.edges).map(key => ({
                id: key.trim(),
                source: key.trim().split('->')[0].trim(),
                target: key.trim().split('->')[1].trim(),
                label: key.trim()
            }));

            set({ graphData: { nodes, edges }, isFetching: false });
            console.log("Graph data updated in store:", { nodes, edges });
        } catch (error) {
            console.error("Fetch failed", error);
            set({ isFetching: false });
        }
    }
}))