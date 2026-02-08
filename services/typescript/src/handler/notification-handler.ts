import type { ConnectRouter } from "@connectrpc/connect";
import { ConnectError, Code } from "@connectrpc/connect";
import { NotificationService as NotificationServiceDef } from "../../gen/notification/v1/notification_pb.js";
import { NotificationRepo, type Notification } from "../repository/notification-repo.js";
import { toConnectError } from "./error-mapping.js";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import type { JsonObject } from "@bufbuild/protobuf";

function notificationToProto(n: Notification) {
  const proto: Record<string, unknown> = {
    id: BigInt(n.id),
    workspaceId: n.workspaceId,
    eventType: n.eventType,
    status: n.status,
    retryCount: n.retryCount,
    createdAt: timestampFromDate(n.createdAt),
  };
  if (n.processedAt) {
    proto.processedAt = timestampFromDate(n.processedAt);
  }
  if (n.payload && Object.keys(n.payload).length > 0) {
    proto.payload = n.payload as JsonObject;
  }
  return proto;
}

export function registerNotificationHandler(router: ConnectRouter, repo: NotificationRepo) {
  router.service(NotificationServiceDef, {
    async listNotifications(req) {
      if (!req.workspaceId) {
        throw new ConnectError("workspace_id is required", Code.InvalidArgument);
      }
      try {
        const list = await repo.list({
          workspaceId: req.workspaceId,
          status: req.status,
          pageSize: req.pagination?.pageSize ?? 0,
          pageToken: req.pagination?.pageToken ?? "",
        });
        return {
          notifications: list.notifications.map(notificationToProto),
          pagination: {
            nextPageToken: list.nextPageToken,
            totalCount: list.totalCount,
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async markNotificationRead(req) {
      try {
        const n = await repo.markProcessed(Number(req.id));
        return { notification: notificationToProto(n) };
      } catch (err) {
        throw toConnectError(err);
      }
    },
  });
}
