module.exports = function (config) {
  config.devServer.port = 3000;
  config.devServer.proxy = [
    {
      path: "/api/**",
      target: "http://127.0.0.1:8080",
      changeOrigin: true,
      changeHost: true,
    },
  ];
};
