import React from 'react';
import { Card as BootstrapCard } from 'react-bootstrap';

export const Card = ({ children, className, ...props }) => {
  return (
    <BootstrapCard className={className} {...props}>
      {children}
    </BootstrapCard>
  );
};

export const CardHeader = ({ children, className, ...props }) => {
  return (
    <BootstrapCard.Header className={className} {...props}>
      {children}
    </BootstrapCard.Header>
  );
};

export const CardTitle = ({ children, className, ...props }) => {
  return (
    <BootstrapCard.Title className={className} {...props}>
      {children}
    </BootstrapCard.Title>
  );
};

export const CardContent = ({ children, className, ...props }) => {
  return (
    <BootstrapCard.Body className={className} {...props}>
      {children}
    </BootstrapCard.Body>
  );
};