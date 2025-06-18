module.exports = {
  env: {
    NODE_ENV: '"production"'
  },
  defineConstants: {
  },
  mini: {},
  h5: {
    /**
     * WebpackChain 插件配置
     * @docs https://github.com/neutrinojs/webpack-chain
     */
    // webpackChain (chain, webpack) {
    //   /**
    //    * 如果 h5 端编译后体积过大，可以使用 webpack-bundle-analyzer 插件对打包体积进行分析。
    //    * @docs https://github.com/webpack-contrib/webpack-bundle-analyzer
    //    */
    //   chain.plugin('analyzer')
    //     .use(require('webpack-bundle-analyzer').BundleAnalyzerPlugin, [])
    //   /**
    //    * 如果 h5 端首屏加载时间过长，可以使用 preload-webpack-plugin 等插件进行优化。
    //    * @docs https://github.com/GoogleChromeLabs/preload-webpack-plugin
    //    */
    //   chain.plugin('preload')
    //     .use(require('preload-webpack-plugin'), [{
    //       rel: 'preload',
    //       as (entry) {
    //         if (/\.css$/.test(entry)) return 'style'
    //         if (/\.woff$/.test(entry)) return 'font'
    //         if (/\.png$/.test(entry)) return 'image'
    //         return 'script'
    //       },
    //       include: 'initial'
    //     }])
    // }
  }
}