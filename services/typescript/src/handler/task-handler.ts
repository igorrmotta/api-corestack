import type { ConnectRouter } from "@connectrpc/connect";
import { ConnectError, Code } from "@connectrpc/connect";
import { TaskService as TaskServiceDef } from "../../gen/task/v1/task_pb.js";
import { TaskService } from "../service/task-service.js";
import { ImportService } from "../service/import-service.js";
import { toConnectError } from "./error-mapping.js";
import { timestampFromDate, timestampDate } from "@bufbuild/protobuf/wkt";
import type { JsonObject } from "@bufbuild/protobuf";
import type { Task } from "../repository/task-repo.js";

function taskToProto(t: Task) {
  const proto: Record<string, unknown> = {
    id: t.id,
    workspaceId: t.workspaceId,
    projectId: t.projectId,
    title: t.title,
    description: t.description,
    status: t.status,
    priority: t.priority,
    assignedTo: t.assignedTo,
    createdAt: timestampFromDate(t.createdAt),
    updatedAt: timestampFromDate(t.updatedAt),
  };
  if (t.dueDate) {
    proto.dueDate = timestampFromDate(t.dueDate);
  }
  if (t.metadata && Object.keys(t.metadata).length > 0) {
    proto.metadata = t.metadata as JsonObject;
  }
  return proto;
}

export function registerTaskHandler(
  router: ConnectRouter,
  svc: TaskService,
  importSvc: ImportService,
) {
  router.service(TaskServiceDef, {
    async createTask(req) {
      if (!req.workspaceId) {
        throw new ConnectError("workspace_id is required", Code.InvalidArgument);
      }
      if (!req.projectId) {
        throw new ConnectError("project_id is required", Code.InvalidArgument);
      }
      try {
        const task = await svc.create({
          workspaceId: req.workspaceId,
          projectId: req.projectId,
          title: req.title,
          description: req.description,
          priority: req.priority,
          assignedTo: req.assignedTo,
          dueDate: req.dueDate ? timestampDate(req.dueDate) : null,
          metadata: (req.metadata as Record<string, unknown>) ?? null,
        });
        return { task: taskToProto(task) };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async getTask(req) {
      if (!req.id) {
        throw new ConnectError("id is required", Code.InvalidArgument);
      }
      try {
        const task = await svc.getById(req.id);
        return { task: taskToProto(task) };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async listTasks(req) {
      if (!req.workspaceId) {
        throw new ConnectError("workspace_id is required", Code.InvalidArgument);
      }
      try {
        const list = await svc.list({
          workspaceId: req.workspaceId,
          projectId: req.projectId,
          status: req.status,
          priority: req.priority,
          assignedTo: req.assignedTo,
          pageSize: req.pagination?.pageSize ?? 0,
          pageToken: req.pagination?.pageToken ?? "",
        });
        return {
          tasks: list.tasks.map(taskToProto),
          pagination: {
            nextPageToken: list.nextPageToken,
            totalCount: list.totalCount,
          },
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async updateTask(req) {
      if (!req.id) {
        throw new ConnectError("id is required", Code.InvalidArgument);
      }
      try {
        const task = await svc.update({
          id: req.id,
          title: req.title,
          description: req.description,
          status: req.status,
          priority: req.priority,
          assignedTo: req.assignedTo,
          dueDate: req.dueDate ? timestampDate(req.dueDate) : null,
          metadata: (req.metadata as Record<string, unknown>) ?? null,
        });
        return { task: taskToProto(task) };
      } catch (err) {
        throw toConnectError(err);
      }
    },

    async deleteTask(req) {
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

    async bulkImportTasks(req) {
      if (!req.workspaceId) {
        throw new ConnectError("workspace_id is required", Code.InvalidArgument);
      }
      if (!req.projectId) {
        throw new ConnectError("project_id is required", Code.InvalidArgument);
      }
      try {
        const inputs = req.tasks.map((t) => ({
          title: t.title,
          description: t.description,
          priority: t.priority,
          assignedTo: t.assignedTo,
          dueDate: t.dueDate ? timestampDate(t.dueDate) : null,
          metadata: (t.metadata as Record<string, unknown>) ?? null,
        }));

        const result = await importSvc.bulkImport(req.workspaceId, req.projectId, inputs);

        return {
          total: result.total,
          succeeded: result.succeeded,
          failed: result.failed,
          errors: result.errors.map((e) => ({
            index: e.index,
            error: e.error,
          })),
        };
      } catch (err) {
        throw toConnectError(err);
      }
    },
  });
}
