import type { Interceptor } from "@connectrpc/connect";
import type { Logger } from "pino";

export function loggingInterceptor(logger: Logger): Interceptor {
  return (next) => async (req) => {
    const start = Date.now();
    const procedure = req.method.name;

    logger.info({ procedure }, "rpc started");

    try {
      const res = await next(req);
      const duration = Date.now() - start;
      logger.info({ procedure, duration }, "rpc completed");
      return res;
    } catch (err) {
      const duration = Date.now() - start;
      logger.error({ procedure, duration, error: err }, "rpc failed");
      throw err;
    }
  };
}
