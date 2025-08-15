import React, { useState, createContext, useContext } from 'react';
import { Nav, Tab } from 'react-bootstrap';

const TabsContext = createContext();

export const Tabs = ({ children, defaultValue, className, ...props }) => {
  const [activeTab, setActiveTab] = useState(defaultValue);
  
  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      <Tab.Container activeKey={activeTab} onSelect={setActiveTab} className={className} {...props}>
        {children}
      </Tab.Container>
    </TabsContext.Provider>
  );
};

export const TabsList = ({ children, className, ...props }) => {
  return (
    <Nav variant="tabs" className={className} {...props}>
      {children}
    </Nav>
  );
};

export const TabsTrigger = ({ children, value, className, ...props }) => {
  return (
    <Nav.Item>
      <Nav.Link eventKey={value} className={className} {...props}>
        {children}
      </Nav.Link>
    </Nav.Item>
  );
};

export const TabsContent = ({ children, value, className, ...props }) => {
  return (
    <Tab.Content className={className} {...props}>
      <Tab.Pane eventKey={value}>
        {children}
      </Tab.Pane>
    </Tab.Content>
  );
};