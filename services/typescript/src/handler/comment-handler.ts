import type { ConnectRouter } from "@connectrpc/connect";
import { ConnectError, Code } from "@connectrpc/connect";
import { CommentService as CommentServiceDef } from "../../gen/comment/v1/comment_pb.js";
import { CommentService } from "../service/comment-service.js";
import { toConnectError } from "./error-mapping.js";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";

export function registerCommentHandler(router: ConnectRouter, svc: CommentService) {
  router.service(CommentServiceDef, {
    async createComment(req) {
      if (!req.taskId) {
        throw new ConnectError("task_id is required", Code.InvalidArgument);
      }
      try {
        const c = await svc.create({
          taskId: req.taskId,
          authorId: req.authorId,
          content: req.content,
        });
        return {
          comment: {
            id: c.id,
            taskId: c.taskId,
            authorId: c.authorId,
            content: c.content,
            createdAt: timestampFromDate(c.createdAt),
            updatedAt: timestampFromDate(c.updatedAt),
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async listComments(req) {
      if (!req.taskId) {
        throw new ConnectError("task_id is required", Code.InvalidArgument);
      }
      try {
        const list = await svc.list({
          taskId: req.taskId,
          pageSize: req.pagination?.pageSize ?? 0,
          pageToken: req.pagination?.pageToken ?? "",
        });
        return {
          comments: list.comments.map((c) => ({
            id: c.id,
            taskId: c.taskId,
            authorId: c.authorId,
            content: c.content,
            createdAt: timestampFromDate(c.createdAt),
            updatedAt: timestampFromDate(c.updatedAt),
          })),
          pagination: {
            nextPageToken: list.nextPageToken,
            totalCount: list.totalCount,
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async deleteComment(req) {
      if (!req.id) {
        throw new ConnectError("id is required", Code.InvalidArgument);
      }
      try {
        await svc.delete(req.id);
        return {};
      } catch (err) {
        throw toConnectError(err);
      }
    },
  });
}
