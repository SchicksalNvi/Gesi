import React from 'react';

interface GesiLogoProps {
  size?: number;
  collapsed?: boolean;
  textColor?: string;
  centered?: boolean;
}

export const GesiLogo: React.FC<GesiLogoProps> = ({ 
  size = 36, 
  collapsed = false,
  textColor,
  centered = false
}) => {
  if (collapsed) {
    // 折叠状态：只显示简化的 G 图标
    return (
      <svg
        width={size}
        height={size}
        viewBox="0 0 100 100"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <linearGradient id="gradient-collapsed" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#667eea" />
            <stop offset="100%" stopColor="#764ba2" />
          </linearGradient>
        </defs>
        
        {/* 简化的 G 字母 */}
        <path
          d="M50 20 C32 20 20 32 20 50 C20 68 32 80 50 80 C63 80 73 72 77 61 L62 61 C59 65 55 68 50 68 C39 68 32 61 32 50 C32 39 39 32 50 32 C55 32 59 34 62 37 L72 27 C66 22 58 20 50 20 Z M62 50 L80 50 L80 61 L62 61 Z"
          fill="url(#gradient-collapsed)"
        />
      </svg>
    );
  }

  // 展开状态：图标 + 文字
  return (
    <div style={{ 
      display: 'flex', 
      alignItems: 'center', 
      gap: 8,
      width: centered ? 'auto' : '100%',
      justifyContent: centered ? 'center' : 'flex-start',
      paddingLeft: centered ? 0 : 4
    }}>
      <svg
        width={size}
        height={size}
        viewBox="0 0 100 100"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        style={{ flexShrink: 0 }}
      >
        <defs>
          <linearGradient id="gradient-full" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#667eea" />
            <stop offset="100%" stopColor="#764ba2" />
          </linearGradient>
        </defs>
        
        {/* G 字母 */}
        <path
          d="M50 20 C32 20 20 32 20 50 C20 68 32 80 50 80 C63 80 73 72 77 61 L62 61 C59 65 55 68 50 68 C39 68 32 61 32 50 C32 39 39 32 50 32 C55 32 59 34 62 37 L72 27 C66 22 58 20 50 20 Z M62 50 L80 50 L80 61 L62 61 Z"
          fill="url(#gradient-full)"
        />
      </svg>
      
      <span style={{ 
        fontSize: 22,
        fontWeight: 600,
        color: textColor || '#667eea',
        letterSpacing: '0.5px',
        lineHeight: 1
      }}>
        Gesi
      </span>
    </div>
  );
};

export default GesiLogo;
