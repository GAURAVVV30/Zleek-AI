const fs = require('fs');
const path = require('path');

const srcDir = path.join(__dirname, 'src');

const replacements = [
  // Backgrounds
  { regex: /\bbg-white\b/g, replace: 'bg-black/40 backdrop-blur-xl' },
  { regex: /bg-\[#f8faff\]/g, replace: 'bg-black/20 backdrop-blur-sm' },
  { regex: /bg-\[#edf4fe\]/g, replace: 'bg-black/20 backdrop-blur-sm' },
  { regex: /\bbg-slate-50(?!\/)\b/g, replace: 'bg-black/30 backdrop-blur-md' },
  { regex: /\bbg-slate-50\/70\b/g, replace: 'bg-black/20 backdrop-blur-md' },
  { regex: /\bbg-slate-50\/80\b/g, replace: 'bg-black/30 backdrop-blur-md' },
  { regex: /\bbg-blue-50\b/g, replace: 'bg-indigo-900/40 backdrop-blur-sm' },
  { regex: /\bbg-blue-100\b/g, replace: 'bg-indigo-900/60 backdrop-blur-md' },
  { regex: /\bbg-blue-100\/60\b/g, replace: 'bg-indigo-900/40 backdrop-blur-md' },
  
  // Borders
  { regex: /\bborder-slate-200\/80\b/g, replace: 'border-white/10' },
  { regex: /\bborder-slate-200\/60\b/g, replace: 'border-white/10' },
  { regex: /\bborder-slate-200\b/g, replace: 'border-white/10' },
  { regex: /\bborder-slate-100\b/g, replace: 'border-white/5' },
  
  // Text Colors
  { regex: /\btext-slate-900\b/g, replace: 'text-white' },
  { regex: /\btext-slate-800\b/g, replace: 'text-white' },
  { regex: /\btext-slate-700\b/g, replace: 'text-white' },
  { regex: /\btext-slate-600\b/g, replace: 'text-slate-300' },
  { regex: /\btext-slate-500\b/g, replace: 'text-slate-400' },
  
  // Accents
  { regex: /\btext-blue-600\b/g, replace: 'text-indigo-400' },
  { regex: /\btext-blue-700\b/g, replace: 'text-indigo-400' },
  { regex: /\btext-blue-800\b/g, replace: 'text-indigo-300' },
  { regex: /\bbg-blue-600\b/g, replace: 'bg-indigo-600' },
  { regex: /\bbg-blue-700\b/g, replace: 'bg-indigo-700' },
  
  // Shadows
  { regex: /\bshadow-md\b/g, replace: 'shadow-[0_0_15px_rgba(79,70,229,0.2)]' },
  { regex: /\bshadow-sm\b/g, replace: 'shadow-[0_0_10px_rgba(79,70,229,0.1)]' },
  { regex: /\bshadow-card\b/g, replace: 'shadow-[0_0_20px_rgba(79,70,229,0.15)]' },
];

function walk(dir) {
  let results = [];
  const list = fs.readdirSync(dir);
  list.forEach(file => {
    file = path.join(dir, file);
    const stat = fs.statSync(file);
    if (stat && stat.isDirectory()) {
      results = results.concat(walk(file));
    } else {
      if (file.endsWith('.jsx') || file.endsWith('.js')) {
        results.push(file);
      }
    }
  });
  return results;
}

const files = walk(srcDir);
let filesModified = 0;

files.forEach(file => {
  const content = fs.readFileSync(file, 'utf8');
  let newContent = content;
  
  replacements.forEach(({ regex, replace }) => {
    newContent = newContent.replace(regex, replace);
  });
  
  if (content !== newContent) {
    fs.writeFileSync(file, newContent, 'utf8');
    console.log(`Updated ${path.relative(srcDir, file)}`);
    filesModified++;
  }
});

console.log(`Refactored ${filesModified} files.`);
