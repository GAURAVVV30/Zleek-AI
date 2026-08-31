import React, { useMemo } from 'react';
import { format, parseISO } from 'date-fns';

/**
 * Maps a count to an intensity level from 0 to 4.
 * This determines the color of the cell.
 */
function getIntensityLevel(count) {
  if (count === 0) return 0;
  if (count === 1) return 1;
  if (count === 2) return 2;
  if (count === 3) return 3;
  return 4;
}

/**
 * Returns the CSS class for the cell based on intensity level.
 * Replaces green with the requested Milky Way purple palette.
 */
function getColorClass(level) {
  switch (level) {
    case 1:
      return 'bg-indigo-900 border-indigo-800/30';
    case 2:
      return 'bg-indigo-700 border-indigo-600/40';
    case 3:
      return 'bg-indigo-500 border-indigo-400/50 shadow-[0_0_8px_rgba(99,102,241,0.5)]';
    case 4:
      return 'bg-indigo-400 border-indigo-300/60 shadow-[0_0_12px_rgba(129,140,248,0.7)]';
    case 0:
    default:
      // Level 0: very dark / almost invisible purple
      return 'bg-white/5 border-white/5';
  }
}

export default function LearningActivityHeatmap({ data = [] }) {
  // Pre-process data to a map for O(1) lookups if the data prop format is just an array.
  // Assuming data is an array of { date: 'YYYY-MM-DD', count: number }
  
  // We want to render a grid with 7 rows (days of the week) 
  // flowing in columns (weeks). 
  // CSS grid auto-flow: column handles this beautifully.

  const maxColumns = Math.ceil(data.length / 7);

  return (
    <div className="w-full">
      <h2 
        className="text-[13px] font-bold text-white uppercase tracking-[0.15em] mb-4"
      >
        Learning Activity
      </h2>
      
      <div className="w-full bg-slate-950/60 backdrop-blur-xl border border-white/10 rounded-2xl shadow-[0_0_20px_rgba(0,0,0,0.5)] p-5 overflow-hidden">
        {/* Horizontal scroll container for smaller screens */}
        <div className="w-full overflow-x-auto pb-2 scrollbar-thin scrollbar-thumb-white/10 scrollbar-track-transparent">
          <div 
            className="grid gap-[3px] sm:gap-[4px] min-w-max" 
            style={{ 
              gridTemplateRows: 'repeat(7, minmax(0, 1fr))',
              gridAutoFlow: 'column',
              gridAutoColumns: '12px' 
            }}
          >
            {data.map((day, i) => {
              const level = getIntensityLevel(day.count);
              const colorClass = getColorClass(level);
              
              // Safely parse date for tooltip
              let formattedDate = day.date;
              try {
                const parsedDate = parseISO(day.date);
                formattedDate = format(parsedDate, 'MMM d, yyyy');
              } catch (e) {
                // fallback to raw string if parsing fails
              }

              const tooltipText = `${formattedDate}: ${day.count} learning activit${day.count === 1 ? 'y' : 'ies'}`;

              return (
                <div
                  key={day.date || i}
                  title={tooltipText}
                  aria-label={tooltipText}
                  className={`
                    w-3 h-3 sm:w-3.5 sm:h-3.5 rounded-[3px] border
                    transition-all duration-200 ease-out
                    hover:scale-[1.2] hover:z-10 hover:ring-2 hover:ring-indigo-300
                    cursor-crosshair
                    ${colorClass}
                  `}
                />
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
