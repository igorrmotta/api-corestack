import type { ConnectRouter } from "@connectrpc/connect";
import { ConnectError, Code } from "@connectrpc/connect";
import { WorkspaceService as WorkspaceServiceDef } from "../../gen/workspace/v1/workspace_pb.js";
import { WorkspaceService } from "../service/workspace-service.js";
import { toConnectError } from "./error-mapping.js";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";

export function registerWorkspaceHandler(router: ConnectRouter, svc: WorkspaceService) {
  router.service(WorkspaceServiceDef, {
    async createWorkspace(req) {
      try {
        const w = await svc.create({ name: req.name, slug: req.slug });
        return {
          workspace: {
            id: w.id,
            name: w.name,
            slug: w.slug,
            createdAt: timestampFromDate(w.createdAt),
            updatedAt: timestampFromDate(w.updatedAt),
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async getWorkspace(req) {
      if (!req.id) {
        throw new ConnectError("id is required", Code.InvalidArgument);
      }
      try {
        const w = await svc.getById(req.id);
        return {
          workspace: {
            id: w.id,
            name: w.name,
            slug: w.slug,
            createdAt: timestampFromDate(w.createdAt),
            updatedAt: timestampFromDate(w.updatedAt),
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async listWorkspaces(req) {
      try {
        const list = await svc.list({
          pageSize: req.pagination?.pageSize ?? 0,
          pageToken: req.pagination?.pageToken ?? "",
        });
        return {
          workspaces: list.workspaces.map((w) => ({
            id: w.id,
            name: w.name,
            slug: w.slug,
            createdAt: timestampFromDate(w.createdAt),
            updatedAt: timestampFromDate(w.updatedAt),
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

    async updateWorkspace(req) {
      if (!req.id) {
        throw new ConnectError("id is required", Code.InvalidArgument);
      }
      try {
        const w = await svc.update({ id: req.id, name: req.name, slug: req.slug });
        return {
          workspace: {
            id: w.id,
            name: w.name,
            slug: w.slug,
            createdAt: timestampFromDate(w.createdAt),
            updatedAt: timestampFromDate(w.updatedAt),
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async deleteWorkspace(req) {
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
