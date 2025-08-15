import React from 'react';
import { ProgressBar } from 'react-bootstrap';

export const Progress = ({ value, max = 100, className, ...props }) => {
  const percentage = (value / max) * 100;
  
  return (
    <ProgressBar
      now={percentage}
      className={className}
      {...props}
    />
  );
};

export default Progress;