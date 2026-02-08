import { InvalidInputError } from "../repository/errors.js";
import { NotificationRepo } from "../repository/notification-repo.js";
import {
  TaskRepo,
  type CreateTaskParams,
  type ListTasksParams,
  type Task,
  type TaskList,
  type UpdateTaskParams,
} from "../repository/task-repo.js";
import type { Logger } from "pino";

const EMPTY_UUID = "00000000-0000-0000-0000-000000000000";

export class TaskService {
  constructor(
    private repo: TaskRepo,
    private notifRepo: NotificationRepo | null,
    private logger: Logger,
  ) {}

  async create(params: CreateTaskParams): Promise<Task> {
    if (!params.title) {
      throw new InvalidInputError("invalid input: title is required");
    }
    if (!params.workspaceId || params.workspaceId === EMPTY_UUID) {
      throw new InvalidInputError("invalid input: workspace_id is required");
    }
    if (!params.projectId || params.projectId === EMPTY_UUID) {
      throw new InvalidInputError("invalid input: project_id is required");
    }

    const validPriorities: Record<string, boolean> = {
      low: true,
      medium: true,
      high: true,
      critical: true,
      "": true,
    };
    if (!validPriorities[params.priority]) {
      throw new InvalidInputError(`invalid input: invalid priority: ${params.priority}`);
    }

    this.logger.debug({ title: params.title, projectId: params.projectId }, "creating task");
    const task = await this.repo.create(params);

    // Enqueue notification (fire-and-forget)
    if (this.notifRepo) {
      try {
        await this.notifRepo.create({
          workspaceId: task.workspaceId,
          eventType: "task.created",
          payload: { task_id: task.id, title: task.title },
        });
      } catch {
        // fire-and-forget
      }
    }

    return task;
  }

  async getById(id: string): Promise<Task> {
    return this.repo.getById(id);
  }

  async list(params: ListTasksParams): Promise<TaskList> {
    const validStatuses: Record<string, boolean> = {
      todo: true,
      in_progress: true,
      review: true,
      done: true,
      "": true,
    };
    if (!validStatuses[params.status]) {
      throw new InvalidInputError(`invalid input: invalid status filter: ${params.status}`);
    }
    return this.repo.list(params);
  }

  async update(params: UpdateTaskParams): Promise<Task> {
    if (!params.title) {
      throw new InvalidInputError("invalid input: title is required");
    }

    const validStatuses: Record<string, boolean> = {
      todo: true,
      in_progress: true,
      review: true,
      done: true,
      "": true,
    };
    if (!validStatuses[params.status]) {
      throw new InvalidInputError(`invalid input: invalid status: ${params.status}`);
    }

    return this.repo.update(params);
  }

  async delete(id: string): Promise<void> {
    return this.repo.delete(id);
  }
}
