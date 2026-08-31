import React, { useMemo } from 'react';
import { subDays, format, parseISO } from 'date-fns';

const THEME_COLORS = [
  "#15132A", // Level 0 (no activity)
  "#33266B", // Level 1 (light)
  "#5941C8", // Level 2 (moderate)
  "#7657F5", // Level 3 (strong)
  "#A78BFA"  // Level 4+ (highly productive)
];

const getLevel = (count) => {
  if (count === 0) return 0;
  if (count <= 1) return 1;
  if (count <= 2) return 2;
  if (count <= 3) return 3;
  return 4;
};

// Generates an array of past days up to today
const generateCalendarData = (daysCount, data) => {
  const dataMap = new Map(data.map(d => [d.date, d.count]));
  
  const today = new Date();
  const calendarData = [];

  for (let i = daysCount - 1; i >= 0; i--) {
    const dateObj = subDays(today, i);
    const dateStr = format(dateObj, 'yyyy-MM-dd');
    const count = dataMap.get(dateStr) || 0;
    
    calendarData.push({
      date: dateStr,
      count,
      level: getLevel(count)
    });
  }

  return calendarData;
};

export default function GitHubCalendar({ data = [], daysCount = 90 }) {
  const calendarData = useMemo(() => generateCalendarData(daysCount, data), [data, daysCount]);

  // Group into weeks (columns of 7)
  const weeks = [];
  let currentWeek = [];

  calendarData.forEach((day, i) => {
    currentWeek.push(day);
    if (currentWeek.length === 7 || i === calendarData.length - 1) {
      weeks.push(currentWeek);
      currentWeek = [];
    }
  });

  return (
    <div className="flex flex-col gap-2 overflow-x-auto [&::-webkit-scrollbar]:hidden p-1">
      <div className="flex gap-1.5 shrink-0">
        {weeks.map((week, wIndex) => (
          <div key={wIndex} className="flex flex-col gap-1.5">
            {week.map((day, dIndex) => (
              <div
                key={day.date}
                className="w-3.5 h-3.5 rounded-sm transition-transform hover:scale-125 cursor-pointer relative group"
                style={{ backgroundColor: THEME_COLORS[day.level] }}
              >
                {/* Tooltip */}
                <div className="opacity-0 group-hover:opacity-100 transition-opacity absolute bottom-full mb-1.5 left-1/2 -translate-x-1/2 bg-slate-800 text-white text-[10px] py-1 px-2 rounded whitespace-nowrap z-50 pointer-events-none shadow-lg border border-white/10">
                  {format(parseISO(day.date), 'MMM d, yyyy')} — {day.count} learning activities
                </div>
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}
