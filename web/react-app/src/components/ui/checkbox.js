import React from 'react';
import { Form } from 'react-bootstrap';

export const Checkbox = ({ checked, onCheckedChange, className, children, ...props }) => {
  return (
    <Form.Check 
      type="checkbox"
      checked={checked}
      onChange={(e) => onCheckedChange && onCheckedChange(e.target.checked)}
      className={className}
      label={children}
      {...props}
    />
  );
};