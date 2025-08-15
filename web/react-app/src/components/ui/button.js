import React from 'react';
import { Button as BootstrapButton } from 'react-bootstrap';

export const Button = ({ children, className, variant = 'primary', ...props }) => {
  return (
    <BootstrapButton variant={variant} className={className} {...props}>
      {children}
    </BootstrapButton>
  );
};