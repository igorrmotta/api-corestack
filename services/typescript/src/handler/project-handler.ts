import type { ConnectRouter } from "@connectrpc/connect";
import { ConnectError, Code } from "@connectrpc/connect";
import { ProjectService as ProjectServiceDef } from "../../gen/project/v1/project_pb.js";
import { ProjectService } from "../service/project-service.js";
import { toConnectError } from "./error-mapping.js";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";

export function registerProjectHandler(router: ConnectRouter, svc: ProjectService) {
  router.service(ProjectServiceDef, {
    async createProject(req) {
      if (!req.workspaceId) {
        throw new ConnectError("workspace_id is required", Code.InvalidArgument);
      }
      try {
        const p = await svc.create({
          workspaceId: req.workspaceId,
          name: req.name,
          description: req.description,
        });
        return {
          project: {
            id: p.id,
            workspaceId: p.workspaceId,
            name: p.name,
            description: p.description,
            status: p.status,
            createdAt: timestampFromDate(p.createdAt),
            updatedAt: timestampFromDate(p.updatedAt),
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async getProject(req) {
      if (!req.id) {
        throw new ConnectError("id is required", Code.InvalidArgument);
      }
      try {
        const p = await svc.getById(req.id);
        return {
          project: {
            id: p.id,
            workspaceId: p.workspaceId,
            name: p.name,
            description: p.description,
            status: p.status,
            createdAt: timestampFromDate(p.createdAt),
            updatedAt: timestampFromDate(p.updatedAt),
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async listProjects(req) {
      if (!req.workspaceId) {
        throw new ConnectError("workspace_id is required", Code.InvalidArgument);
      }
      try {
        const list = await svc.list({
          workspaceId: req.workspaceId,
          pageSize: req.pagination?.pageSize ?? 0,
          pageToken: req.pagination?.pageToken ?? "",
        });
        return {
          projects: list.projects.map((p) => ({
            id: p.id,
            workspaceId: p.workspaceId,
            name: p.name,
            description: p.description,
            status: p.status,
            createdAt: timestampFromDate(p.createdAt),
            updatedAt: timestampFromDate(p.updatedAt),
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

    async updateProject(req) {
      if (!req.id) {
        throw new ConnectError("id is required", Code.InvalidArgument);
      }
      try {
        const p = await svc.update({
          id: req.id,
          name: req.name,
          description: req.description,
          status: req.status,
        });
        return {
          project: {
            id: p.id,
            workspaceId: p.workspaceId,
            name: p.name,
            description: p.description,
            status: p.status,
            createdAt: timestampFromDate(p.createdAt),
            updatedAt: timestampFromDate(p.updatedAt),
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async deleteProject(req) {
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
