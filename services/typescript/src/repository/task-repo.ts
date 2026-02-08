import type postgres from "postgres";
import { NotFoundError } from "./errors.js";

export interface Task {
  id: string;
  workspaceId: string;
  projectId: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  assignedTo: string;
  dueDate: Date | null;
  metadata: Record<string, unknown>;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;
}

export interface CreateTaskParams {
  workspaceId: string;
  projectId: string;
  title: string;
  description: string;
  priority: string;
  assignedTo: string;
  dueDate: Date | null;
  metadata: Record<string, unknown> | null;
}

export interface UpdateTaskParams {
  id: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  assignedTo: string;
  dueDate: Date | null;
  metadata: Record<string, unknown> | null;
}

export interface ListTasksParams {
  workspaceId: string;
  projectId: string;
  status: string;
  priority: string;
  assignedTo: string;
  pageSize: number;
  pageToken: string;
}

export interface TaskList {
  tasks: Task[];
  nextPageToken: string;
  totalCount: number;
}

export interface TaskInput {
  title: string;
  description: string;
  priority: string;
  assignedTo: string;
  dueDate: Date | null;
  metadata: Record<string, unknown> | null;
}

export interface ImportResult {
  total: number;
  succeeded: number;
  failed: number;
  errors: ImportError[];
}

export interface ImportError {
  index: number;
  error: string;
}

const EMPTY_UUID = "00000000-0000-0000-0000-000000000000";

export class TaskRepo {
  constructor(private sql: postgres.Sql) {}

  async create(params: CreateTaskParams): Promise<Task> {
    const assignedTo = params.assignedTo || null;
    const metadata = params.metadata ?? {};

    const [row] = await this.sql<Task[]>`
      INSERT INTO tasks (id, workspace_id, project_id, title, description, status, priority,
                         assigned_to, due_date, metadata, created_at, updated_at)
      VALUES (gen_random_uuid(), ${params.workspaceId}, ${params.projectId}, ${params.title},
              ${params.description}, COALESCE(NULLIF('', ''), 'todo'),
              COALESCE(NULLIF(${params.priority}, ''), 'medium'),
              ${assignedTo}, ${params.dueDate},
              ${this.sql.json(metadata as never)}, NOW(), NOW())
      RETURNING id, workspace_id AS "workspaceId", project_id AS "projectId", title, description,
                status, priority, COALESCE(assigned_to, '') AS "assignedTo", due_date AS "dueDate",
                metadata, created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
    `;
    return row;
  }

  async getById(id: string): Promise<Task> {
    const [row] = await this.sql<Task[]>`
      SELECT id, workspace_id AS "workspaceId", project_id AS "projectId", title, description,
             status, priority, COALESCE(assigned_to, '') AS "assignedTo", due_date AS "dueDate",
             metadata, created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
      FROM tasks WHERE id = ${id} AND deleted_at IS NULL
    `;
    if (!row) {
      throw new NotFoundError();
    }
    return row;
  }

  async list(params: ListTasksParams): Promise<TaskList> {
    let pageSize = params.pageSize;
    if (pageSize <= 0 || pageSize > 100) {
      pageSize = 20;
    }

    // Build dynamic WHERE conditions using postgres.js fragments
    const conditions = [
      this.sql`workspace_id = ${params.workspaceId}`,
      this.sql`deleted_at IS NULL`,
    ];

    if (params.projectId && params.projectId !== EMPTY_UUID) {
      conditions.push(this.sql`project_id = ${params.projectId}`);
    }
    if (params.status) {
      conditions.push(this.sql`status = ${params.status}`);
    }
    if (params.priority) {
      conditions.push(this.sql`priority = ${params.priority}`);
    }
    if (params.assignedTo) {
      conditions.push(this.sql`assigned_to = ${params.assignedTo}`);
    }

    const whereClause = conditions.reduce(
      (acc, cond) => this.sql`${acc} AND ${cond}`,
    );

    // Count query
    const [{ count }] = await this.sql<[{ count: number }]>`
      SELECT COUNT(*)::int AS count FROM tasks WHERE ${whereClause}
    `;

    // Add cursor pagination
    if (params.pageToken) {
      conditions.push(this.sql`id < ${params.pageToken}`);
    }
    const whereClauseWithCursor = conditions.reduce(
      (acc, cond) => this.sql`${acc} AND ${cond}`,
    );

    const tasks = await this.sql<Task[]>`
      SELECT id, workspace_id AS "workspaceId", project_id AS "projectId", title, description,
             status, priority, COALESCE(assigned_to, '') AS "assignedTo", due_date AS "dueDate",
             metadata, created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
      FROM tasks WHERE ${whereClauseWithCursor}
      ORDER BY created_at DESC, id DESC LIMIT ${pageSize + 1}
    `;

    let nextPageToken = "";
    const result = [...tasks];
    if (result.length > pageSize) {
      nextPageToken = result[pageSize].id;
      result.length = pageSize;
    }

    return { tasks: result, nextPageToken, totalCount: count };
  }

  async update(params: UpdateTaskParams): Promise<Task> {
    const assignedTo = params.assignedTo || null;
    const metadata = params.metadata ?? {};

    const [row] = await this.sql<Task[]>`
      UPDATE tasks SET title = ${params.title}, description = ${params.description},
             status = ${params.status}, priority = ${params.priority},
             assigned_to = ${assignedTo}, due_date = ${params.dueDate},
             metadata = ${this.sql.json(metadata as never)}, updated_at = NOW()
      WHERE id = ${params.id} AND deleted_at IS NULL
      RETURNING id, workspace_id AS "workspaceId", project_id AS "projectId", title, description,
                status, priority, COALESCE(assigned_to, '') AS "assignedTo", due_date AS "dueDate",
                metadata, created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
    `;
    if (!row) {
      throw new NotFoundError();
    }
    return row;
  }

  async delete(id: string): Promise<void> {
    const result = await this.sql`
      UPDATE tasks SET deleted_at = NOW(), updated_at = NOW()
      WHERE id = ${id} AND deleted_at IS NULL
    `;
    if (result.count === 0) {
      throw new NotFoundError();
    }
  }
}
