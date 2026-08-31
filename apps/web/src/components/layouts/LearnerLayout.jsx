import React from 'react';
import { Outlet, NavLink } from 'react-router-dom';
import { Map, BarChart3, Settings, Trophy, HelpCircle } from 'lucide-react';
import Header from './Header';
import { ErrorBoundary } from '../common/ErrorBoundary';
import LiquidMorphFloatingMenu from '../ui/liquid-morph-floating-menu';

export default function LearnerLayout() {
  const navItems = [
    { label: 'Personalized Roadmap', path: '/roadmap', icon: Map },
    { label: 'Progress & Competency', path: '/progress', icon: BarChart3 },
    { label: 'Goal Achieved', path: '/goal-achieved', icon: Trophy },
    { label: 'Profile & Settings', path: '/settings', icon: Settings },
  ];

  return (
    <div className="min-h-screen bg-transparent flex flex-col">
      <Header />
      <div className="flex-1 flex max-w-[1920px] w-full mx-auto px-6 lg:px-10 py-8 gap-8 justify-center">
        <LiquidMorphFloatingMenu navItems={navItems} />

        {/* Main Content View with Error Boundary */}
        <main className="flex-1 min-w-0">
          <ErrorBoundary>
            <Outlet />
          </ErrorBoundary>
        </main>
      </div>
    </div>
  );
}
