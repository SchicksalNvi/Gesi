module.exports = {
  webpack: {
    configure: (webpackConfig, { env }) => {
      if (env === 'development') {
        // 使用更快的 source map
        webpackConfig.devtool = 'eval-cheap-module-source-map';
        webpackConfig.stats = 'errors-warnings';
        
        // 优化文件监听
        webpackConfig.watchOptions = {
          ignored: /node_modules/,
          aggregateTimeout: 300,
          poll: false, // 不使用轮询
        };

        // 优化模块解析
        webpackConfig.resolve.modules = ['node_modules'];
        webpackConfig.resolve.extensions = ['.js', '.jsx', '.json'];
      }

      if (env === 'production') {
        // 代码分割优化
        webpackConfig.optimization = {
          ...webpackConfig.optimization,
          splitChunks: {
            chunks: 'all',
            cacheGroups: {
              // 第三方库单独打包
              vendor: {
                test: /[\\/]node_modules[\\/]/,
                name: 'vendors',
                priority: 10,
                reuseExistingChunk: true,
              },
              // 公共代码单独打包
              common: {
                minChunks: 2,
                priority: 5,
                reuseExistingChunk: true,
                minSize: 0,
              },
            },
          },
          // 运行时代码单独打包
          runtimeChunk: {
            name: 'runtime',
          },
        };
      }

      return webpackConfig;
    },
  },
  devServer: (devServerConfig, { env }) => {
    return {
      ...devServerConfig,
      compress: false, // 开发环境不压缩
      hot: true, // 热模块替换
      liveReload: false, // 禁用完整页面重载，只使用 HMR
      client: {
        overlay: {
          errors: true,
          warnings: false, // 不显示警告
        },
        progress: true, // 显示编译进度
      },
      // 优化开发服务器性能
      devMiddleware: {
        writeToDisk: false, // 不写入磁盘，提高速度
        stats: 'errors-warnings', // 只显示错误和警告
      },
      // 静态文件服务优化
      static: {
        watch: {
          ignored: /node_modules/,
          usePolling: false, // 不使用轮询
        },
      },
      // 启用 HTTP/2
      http2: false, // 如果需要 HTTPS，可以设置为 true
      // 优化 WebSocket 连接
      webSocketServer: 'ws',
    };
  },
  eslint: {
    enable: false, // 开发时禁用 ESLint 以加快速度
  },
  // TypeScript 配置优化（如果使用 TypeScript）
  typescript: {
    enableTypeChecking: false, // 开发时禁用类型检查以加快速度
  },
  // Babel 配置优化
  babel: {
    loaderOptions: (babelLoaderOptions, { env }) => {
      if (env === 'development') {
        // 开发环境优化
        babelLoaderOptions.cacheDirectory = true;
        babelLoaderOptions.cacheCompression = false;
      }
      return babelLoaderOptions;
    },
  },
};