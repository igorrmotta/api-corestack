import { ConnectError, Code, type Interceptor } from "@connectrpc/connect";
import type { Logger } from "pino";

export function recoveryInterceptor(logger: Logger): Interceptor {
  return (next) => async (req) => {
    try {
      return await next(req);
    } catch (err) {
      if (err instanceof ConnectError) {
        throw err;
      }
      logger.error({ procedure: req.method.name, error: err }, "panic recovered");
      throw new ConnectError("internal error", Code.Internal);
    }
  };
}
