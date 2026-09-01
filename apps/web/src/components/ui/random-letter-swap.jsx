import React, { useState, useEffect } from 'react';

const CHARS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+';

export const RandomLetterSwap = ({ label, href, className }) => {
  const [text, setText] = useState(label);
  const [isHovered, setIsHovered] = useState(false);

  useEffect(() => {
    if (!isHovered) {
      setText(label);
      return;
    }

    let iterations = 0;
    const interval = setInterval(() => {
      setText((prev) =>
        prev
          .split('')
          .map((letter, index) => {
            // Keep spaces as spaces
            if (label[index] === ' ') return ' ';
            if (index < iterations) return label[index];
            return CHARS[Math.floor(Math.random() * CHARS.length)];
          })
          .join('')
      );

      if (iterations >= label.length) {
        clearInterval(interval);
      }

      iterations += 1 / 3;
    }, 30);

    return () => clearInterval(interval);
  }, [isHovered, label]);

  return (
    <a
      href={href}
      className={className}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {text}
    </a>
  );
};
