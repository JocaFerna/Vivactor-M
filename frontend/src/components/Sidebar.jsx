import React, { useState, useMemo } from 'react';
import { 
  ChevronLeft, Upload, FileText, Play, Waypoints, 
  Link, PackageSearch, Boxes, Repeat, Zap 
} from 'lucide-react';
import { useGlobalStore } from '../store/useGlobalStore';

// Modals
import LoadArchModal from './modal/LoadArchModal';
import StartArchModal from './modal/StartArchModal';
import RefactorModal from './modal/RefactorModal';
import LoadArchJsonModal from './modal/LoadArchJsonModal';
import EmulateArchModal from './modal/EmulateArchModal';

const NavItem = ({ icon, title, onClick, open, active, gap, variant = "default" }) => (
  <li
    onClick={onClick}
    className={`
      flex rounded-md p-2 cursor-pointer items-center gap-x-4 
      duration-200 text-sm hover:bg-slate-800
      ${gap ? "mt-9" : "mt-2"}
      ${active ? "bg-slate-800 text-white" : "text-slate-300"}
      ${variant === "refactor" ? "border border-dashed border-slate-700 hover:border-blue-500" : ""}
    `}
  >
    {icon}
    <span className={`${!open && "hidden"} origin-left duration-200`}>
      {title}
    </span>
  </li>
);

const Sidebar = () => {
  const [open, setOpen] = useState(true);
  
  // Modal States
  const [isLoadModalOpen, setIsLoadModalOpen] = useState(false);
  const [isLoadModalJsonOpen, setIsLoadModalJsonOpen] = useState(false);
  const [isEmulateModalOpen, setIsEmulateModalOpen] = useState(false);
  const [isStartModalOpen, setIsStartModalOpen] = useState(false);
  
  // Combined Refactor Modal State
  const [refactorConfig, setRefactorConfig] = useState({ isOpen: false, type: "" });

  // Store Subscriptions
  const architectureURL = useGlobalStore((state) => state.architectureURL);
  const graphData = useGlobalStore((state) => state.graphData);
  
  // Smell Flags from Store (Add more here as you implement them)
  const smellsDetected = {
    nonAPIVersioned: useGlobalStore((state) => state.refactoringOfNonAPIVersioned),
    cyclicDependency: useGlobalStore((state) => state.refactoringOfCyclicDependency), 
    esbUsage: useGlobalStore((state) => state.refactoringOfEsbUsage), 
    hardcodedEndpoints: useGlobalStore((state) => state.refactoringOfHardcodedEndpoints),     
    innapropriateServiceIntimacity: useGlobalStore((state) => state.refactoringOfInnapropriateServiceIntimacity), 
    microserviceGreedy: useGlobalStore((state) => state.refactoringOfMicroserviceGreedy),
    sharedLibraries: useGlobalStore((state) => state.refactoringOfSharedLibraries),
    sharedPersistency: useGlobalStore((state) => state.refactoringOfSharedPersistency),
    wrongCuts: useGlobalStore((state) => state.refactoringOfWrongCuts),
    tooManyStandards: useGlobalStore((state) => state.refactoringOfTooManyStandards),
    noAPIGateway: useGlobalStore((state) => state.refactoringOfNoAPIGateway),
  };

  // --- REFACTORING LOGIC CONFIGURATION ---
  const refactorOptions = useMemo(() => [
    {
      id: "nonAPIVersioned",
      title: "Fix API Versioning",
      icon: <Waypoints size={20} className="text-blue-400" />,
      enabled: smellsDetected.nonAPIVersioned,
    },
    {
      id: "cyclicDependency",
      title: "Break Cyclic Dependencies",
      icon: <Repeat size={20} className="text-purple-400" />,
      enabled: smellsDetected.cyclicDependency,
    },
    {
      id: "hardcodedEndpoints",
      title: "Resolve Hardcoded Endpoints",
      icon: <Zap size={20} className="text-yellow-400" />,
      enabled: smellsDetected.hardcodedEndpoints,
    },
    {
      id: "innapropriateServiceIntimacity",
      title: "Address Service Intimacy",
      icon: <Boxes size={20} className="text-green-400" />,
      enabled: smellsDetected.innapropriateServiceIntimacity,
    },
    {
      id: "microserviceGreedy",
      title: "Split Greedy Microservices",
      icon: <PackageSearch size={20} className="text-red-400" />,
      enabled: smellsDetected.microserviceGreedy,
    },
     {
      id: "sharedLibraries",
      title: "Decouple Shared Libraries",
      icon: <Link size={20} className="text-pink-400" />,
      enabled:  smellsDetected.sharedLibraries,
    },
     {
      id: "sharedPersistency",
      title: "Separate Shared Persistency",
      icon: <FileText size={20} className="text   -indigo-400" />,
      enabled:  smellsDetected.sharedPersistency,
    },
      {
        id: "wrongCuts",
        title: "Reevaluate Service Cuts",
        icon: <Waypoints size={20} className="text-cyan-400" />,
        enabled: smellsDetected.wrongCuts,
      },
      {
        id: "tooManyStandards",
        title: "Unify Standards",
        icon: <Boxes size={20} className="text-orange-400" />,
        enabled: smellsDetected.tooManyStandards,
      },
      {
        id: "noAPIGateway",
        title: "Introduce API Gateway",
        icon: <Link size={20} className="text-gray-400" />,
        enabled: smellsDetected.noAPIGateway,
      },
  ], [smellsDetected]);

  const activeRefactors = refactorOptions.filter(opt => opt.enabled);

  return (
    <div className="flex">
      <div className={`bg-slate-900 h-screen p-5 pt-8 relative duration-300 ${open ? "w-72" : "w-20"}`}>
        <button
          className="absolute -right-3 top-9 w-7 h-7 bg-white border-slate-900 border-2 rounded-full flex items-center justify-center z-50"
          onClick={() => setOpen(!open)}
        >
          <ChevronLeft size={18} className={`${!open && "rotate-180"}`} />
        </button>

        <div className="flex gap-x-4 items-center">
          <div className="bg-blue-600 p-2 rounded shrink-0">
            <img src="/public/images/logo_white.png" alt="Vivactor-M Logo" className="w-6 h-6" />
          </div>
          <h1 className={`text-white font-medium text-xl duration-200 ${!open && "scale-0"}`}>
            Vivactor-M
          </h1>
        </div>

        <ul className="pt-6 overflow-y-auto h-[calc(100vh-120px)] custom-scrollbar">
          {/* Section: Project Management */}
          <NavItem 
            icon={<Link size={20} />} 
            title="Load via GITHUB (Legacy)" 
            open={open}
            onClick={() => setIsLoadModalOpen(true)} 
          />
          <NavItem 
            icon={<Upload size={20} />} 
            title="Load via JSON" 
            open={open} 
            onClick={() => setIsLoadModalJsonOpen(true)} 
          />

          {/* Section: Execution */}
          {architectureURL && (
            <NavItem 
              icon={<Play size={20} className="text-green-500" />} 
              title="Start Architecture" 
              open={open} 
              onClick={() => setIsStartModalOpen(true)} 
            />
          )}

          {graphData?.nodes?.length > 0 && (
            <NavItem 
              icon={<PackageSearch size={20} />} 
              title="Emulate Architecture" 
              open={open} 
              onClick={() => setIsEmulateModalOpen(true)} 
            />
          )}

          {/* DYNAMIC REFACTORING SECTION */}
          {activeRefactors.length > 0 && (
            <>
              <div className={`mt-10 mb-2 ml-2 transition-opacity duration-200 ${!open ? "opacity-0" : "opacity-100"}`}>
                <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                  Refactorings Detected
                </p>
              </div>
              {activeRefactors.map((option) => (
                <NavItem 
                  key={option.id}
                  icon={option.icon} 
                  title={option.title} 
                  open={open} 
                  variant="refactor"
                  onClick={() => setRefactorConfig({ isOpen: true, type: option.id })} 
                />
              ))}
            </>
          )}

          {/* Secondary Items */}
          <NavItem 
            gap={activeRefactors.length > 0 ? true : false}
            icon={<FileText size={20} />} 
            title="Thesis Docs" 
            open={open} 
            onClick={() => window.open('https://your-docs-link.com', '_blank')} 
          />
        </ul>
      </div>

      {/* Modals */}
      <LoadArchModal isOpen={isLoadModalOpen} onClose={() => setIsLoadModalOpen(false)} />
      <LoadArchJsonModal isOpen={isLoadModalJsonOpen} onClose={() => setIsLoadModalJsonOpen(false)} />
      <StartArchModal isOpen={isStartModalOpen} onClose={() => setIsStartModalOpen(false)} />
      <EmulateArchModal isOpen={isEmulateModalOpen} onClose={() => setIsEmulateModalOpen(false)} />
      
      {/* Universal Refactor Modal */}
      <RefactorModal 
        isOpen={refactorConfig.isOpen} 
        onClose={() => setRefactorConfig({ ...refactorConfig, isOpen: false })} 
        typeOfRefactor={refactorConfig.type}
      />
    </div>
  );
};

export default Sidebar;