import React from 'react';
import { Form } from 'react-bootstrap';

export const DatePicker = ({ value, onChange, ...props }) => {
  return (
    <Form.Control
      type="date"
      value={value}
      onChange={(e) => onChange && onChange(e.target.value)}
      {...props}
    />
  );
};

export default DatePicker;