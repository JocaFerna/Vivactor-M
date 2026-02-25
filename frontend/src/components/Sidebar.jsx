import React, { useState } from 'react';
import { ChevronLeft, Upload, Settings, FileText, Users, Play } from 'lucide-react';
import LoadArchModal from './modal/LoadArchModal';
import { useGlobalStore } from '../store/useGlobalStore';
import StartArchModal from './modal/StartArchModal';

const NavItem = ({ icon, title, onClick, open, active, gap }) => (
  <li
    onClick={onClick}
    className={`
      flex rounded-md p-2 cursor-pointer items-center gap-x-4 
      duration-200 text-slate-300 text-sm hover:bg-slate-800
      ${gap ? "mt-9" : "mt-2"}
      ${active ? "bg-slate-800 text-white" : ""}
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
  const [isLoadModalOpen, setIsLoadModalOpen] = useState(false);
  const [isStartModalOpen, setIsStartModalOpen] = useState(false);

  return (
    <div className="flex">
      <div className={`bg-slate-900 h-screen p-5 pt-8 relative duration-300 ${open ? "w-72" : "w-20"}`}>
        {/* Toggle */}
        <button
          className="absolute -right-3 top-9 w-7 h-7 bg-white border-slate-900 border-2 rounded-full flex items-center justify-center"
          onClick={() => setOpen(!open)}
        >
          <ChevronLeft size={18} className={`${!open && "rotate-180"}`} />
        </button>

        {/* Logo */}
        <div className="flex gap-x-4 items-center">
          <div className="bg-blue-600 p-2 rounded shrink-0">
            {/* Image of tool */}
            <img src="/public/images/logo_white.png" alt="Vivactor-M Logo" className="w-6 h-6" />
          </div>
          <h1 className={`text-white font-medium text-xl duration-200 ${!open && "scale-0"}`}>
            Vivactor-M
          </h1>
        </div>

        {/* Menu Items */}
        <ul className="pt-6">
          <NavItem 
            icon={<Upload size={20} />} 
            title="Load Architecture" 
            open={open} 
            active={true}
            onClick={() => setIsLoadModalOpen(true)} 
          />

          {/* Only active when useGlobalStore().architectureURL is set, otherwise it should be disabled or hidden */}
          {useGlobalStore().architectureURL && (
            <NavItem 
              icon={<Play size={20} />} 
              title="Start Architecture" 
              open={open} 
              onClick={() => setIsStartModalOpen(true)} 
            />
          )}

          {/* You can easily add more items here with their own specific onClick functions */}
          <NavItem 
            icon={<FileText size={20} />} 
            title="Thesis Docs" 
            open={open} 
            onClick={() => console.log("Navigate to Docs")} 
          />
        </ul>
      </div>

      {/* Modals */}
      <LoadArchModal 
        isOpen={isLoadModalOpen} 
        onClose={() => setIsLoadModalOpen(false)} 
      />
      <StartArchModal 
        isOpen={isStartModalOpen} 
        onClose={() => setIsStartModalOpen(false)} 
      />
    </div>
  );
};

export default Sidebar;