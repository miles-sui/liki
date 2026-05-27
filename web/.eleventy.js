const locale = process.env.LOCALE || 'en';

module.exports = function(eleventyConfig) {
  // Only copy shared assets once (on en/default build)
  if (locale === 'en') {
    eleventyConfig.addPassthroughCopy('js');
    eleventyConfig.addPassthroughCopy('img');
  }
  eleventyConfig.addWatchTarget('js/');
  eleventyConfig.addWatchTarget('css/');
  eleventyConfig.addWatchTarget('_includes/');

  // Proxy /api/* requests to Go API server on :8081
  eleventyConfig.setServerOptions({
    middleware: [function (req, res, next) {
      if (req.url.startsWith('/api/')) {
        var opts = { hostname: 'localhost', port: 8081, path: req.url, method: req.method, headers: req.headers };
        var proxy = require('http').request(opts, function (proxyRes) {
          res.writeHead(proxyRes.statusCode, proxyRes.headers);
          proxyRes.pipe(res);
        });
        proxy.on('error', function () { res.statusCode = 502; res.end(); });
        req.pipe(proxy);
        return;
      }
      next();
    }],
  });

  return {
    dir: {
      input: 'pages',
      output: 'dist',
      includes: '../_includes',
      layouts: '../_layouts',
      data: '../_data',
    },
    templateFormats: ['html', 'njk'],
    htmlTemplateEngine: 'njk',
    markdownTemplateEngine: false,
  };
};
