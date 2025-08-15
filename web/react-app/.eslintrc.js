module.exports = {
  extends: [
    'react-app',
    'react-app/jest'
  ],
  rules: {
    // 在开发环境中放宽一些规则以加快检查速度
    'no-unused-vars': 'warn',
    'no-console': 'off',
    'react-hooks/exhaustive-deps': 'warn'
  },
  settings: {
    react: {
      version: 'detect'
    }
  },
  // 优化性能配置
  parserOptions: {
    ecmaVersion: 2020,
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true
    }
  }
};