import React from 'react';
import logoSrc from '@/assets/logo.png';

interface SuperviewLogoProps {
  size?: number;
  collapsed?: boolean;
  textColor?: string;
  centered?: boolean;
}

const logoStyle = (size: number): React.CSSProperties => ({
  width: size,
  height: size,
  objectFit: 'contain',
  flexShrink: 0,
});

export const SuperviewLogo: React.FC<SuperviewLogoProps> = ({ 
  size = 36, 
  collapsed = false,
  textColor,
  centered = false
}) => {
  if (collapsed) {
    return <img src={logoSrc} alt="Superview" style={logoStyle(size)} />;
  }

  return (
    <div style={{ 
      display: 'flex', 
      alignItems: 'center', 
      gap: 8,
      width: centered ? 'auto' : '100%',
      justifyContent: centered ? 'center' : 'flex-start',
      paddingLeft: centered ? 0 : 4
    }}>
      <img src={logoSrc} alt="Superview" style={logoStyle(size)} />
      <span style={{ 
        fontSize: 22,
        fontWeight: 600,
        color: textColor || '#667eea',
        letterSpacing: '0.5px',
        lineHeight: 1
      }}>
        Superview
      </span>
    </div>
  );
};

export default SuperviewLogo;
