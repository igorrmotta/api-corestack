import { ConnectError, Code } from "@connectrpc/connect";
import {
  NotFoundError,
  InvalidInputError,
  AlreadyExistsError,
  ConflictError,
} from "../repository/errors.js";

export function toConnectError(err: unknown): ConnectError {
  if (err instanceof NotFoundError) {
    return new ConnectError(err.message, Code.NotFound);
  }
  if (err instanceof InvalidInputError) {
    return new ConnectError(err.message, Code.InvalidArgument);
  }
  if (err instanceof AlreadyExistsError || err instanceof ConflictError) {
    return new ConnectError(err.message, Code.AlreadyExists);
  }
  if (err instanceof Error) {
    return new ConnectError(err.message, Code.Internal);
  }
  return new ConnectError(String(err), Code.Internal);
}
