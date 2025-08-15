import React from 'react';
import { Badge as BootstrapBadge } from 'react-bootstrap';

export const Badge = ({ children, className, variant = 'primary', ...props }) => {
  return (
    <BootstrapBadge bg={variant} className={className} {...props}>
      {children}
    </BootstrapBadge>
  );
};