import React from 'react';
import { Form } from 'react-bootstrap';

export const Input = ({ className, ...props }) => {
  return (
    <Form.Control className={className} {...props} />
  );
};

export const Textarea = ({ className, ...props }) => {
  return (
    <Form.Control as="textarea" className={className} {...props} />
  );
};