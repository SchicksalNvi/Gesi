import React from 'react';
import { Modal } from 'react-bootstrap';

export const Dialog = ({ children, open, onOpenChange, ...props }) => {
  return (
    <Modal show={open} onHide={() => onOpenChange(false)} {...props}>
      {children}
    </Modal>
  );
};

export const DialogTrigger = ({ children, ...props }) => {
  return React.cloneElement(children, props);
};

export const DialogContent = ({ children, className, ...props }) => {
  return (
    <>
      {children}
    </>
  );
};

export const DialogHeader = ({ children, className, ...props }) => {
  return (
    <Modal.Header closeButton className={className} {...props}>
      {children}
    </Modal.Header>
  );
};

export const DialogTitle = ({ children, className, ...props }) => {
  return (
    <Modal.Title className={className} {...props}>
      {children}
    </Modal.Title>
  );
};

export const DialogBody = ({ children, className, ...props }) => {
  return (
    <Modal.Body className={className} {...props}>
      {children}
    </Modal.Body>
  );
};

export const DialogFooter = ({ children, className, ...props }) => {
  return (
    <Modal.Footer className={className} {...props}>
      {children}
    </Modal.Footer>
  );
};