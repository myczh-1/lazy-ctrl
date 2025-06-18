import { MidwayConfig } from '@midwayjs/core';

export default {
  koa: {
    port: 7001,
  },
  cors: {
    origin: '*',
    credentials: true,
  },
} as MidwayConfig;