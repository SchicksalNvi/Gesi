module.exports = {
  webpack: {
    configure: (webpackConfig, { env }) => {
      if (env === 'development') {
        // 简化的优化配置
        webpackConfig.devtool = 'eval-cheap-module-source-map';
        webpackConfig.stats = 'errors-warnings';
        
        // 优化文件监听
        webpackConfig.watchOptions = {
          ignored: /node_modules/,
          aggregateTimeout: 300
        };
      }
      return webpackConfig;
    }
  },
  devServer: {
    compress: false,
    hot: true,
    client: {
      overlay: {
        errors: true,
        warnings: false
      }
    }
  },
  eslint: {
    enable: false
  }
};