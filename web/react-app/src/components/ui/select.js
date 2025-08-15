import React from 'react';
import { Form } from 'react-bootstrap';

export const Select = ({ children, value, onValueChange, className, ...props }) => {
  return (
    <Form.Select 
      value={value} 
      onChange={(e) => onValueChange && onValueChange(e.target.value)}
      className={className} 
      {...props}
    >
      {children}
    </Form.Select>
  );
};

export const SelectTrigger = ({ children, className, ...props }) => {
  return children;
};

export const SelectValue = ({ placeholder, ...props }) => {
  return null; // This is handled by the Select component
};

export const SelectContent = ({ children, ...props }) => {
  return <>{children}</>;
};

export const SelectItem = ({ children, value, ...props }) => {
  return (
    <option value={value} {...props}>
      {children}
    </option>
  );
};