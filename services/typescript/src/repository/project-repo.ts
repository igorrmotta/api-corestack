import type postgres from "postgres";
import { NotFoundError } from "./errors.js";

export interface Project {
  id: string;
  workspaceId: string;
  name: string;
  description: string;
  status: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;
}

export interface CreateProjectParams {
  workspaceId: string;
  name: string;
  description: string;
}

export interface UpdateProjectParams {
  id: string;
  name: string;
  description: string;
  status: string;
}

export interface ListProjectsParams {
  workspaceId: string;
  pageSize: number;
  pageToken: string;
}

export interface ProjectList {
  projects: Project[];
  nextPageToken: string;
  totalCount: number;
}

export class ProjectRepo {
  constructor(private sql: postgres.Sql) {}

  async create(params: CreateProjectParams): Promise<Project> {
    const [row] = await this.sql<Project[]>`
      INSERT INTO projects (id, workspace_id, name, description, status, created_at, updated_at)
      VALUES (gen_random_uuid(), ${params.workspaceId}, ${params.name}, ${params.description}, 'active', NOW(), NOW())
      RETURNING id, workspace_id AS "workspaceId", name, description, status,
                created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
    `;
    return row;
  }

  async getById(id: string): Promise<Project> {
    const [row] = await this.sql<Project[]>`
      SELECT id, workspace_id AS "workspaceId", name, description, status,
             created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
      FROM projects WHERE id = ${id} AND deleted_at IS NULL
    `;
    if (!row) {
      throw new NotFoundError();
    }
    return row;
  }

  async list(params: ListProjectsParams): Promise<ProjectList> {
    let pageSize = params.pageSize;
    if (pageSize <= 0 || pageSize > 100) {
      pageSize = 20;
    }

    const [{ count }] = await this.sql<[{ count: number }]>`
      SELECT COUNT(*)::int AS count FROM projects
      WHERE workspace_id = ${params.workspaceId} AND deleted_at IS NULL
    `;

    let projects: Project[];
    if (params.pageToken) {
      projects = await this.sql<Project[]>`
        SELECT id, workspace_id AS "workspaceId", name, description, status,
               created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
        FROM projects WHERE workspace_id = ${params.workspaceId} AND deleted_at IS NULL AND id < ${params.pageToken}
        ORDER BY created_at DESC, id DESC LIMIT ${pageSize + 1}
      `;
    } else {
      projects = await this.sql<Project[]>`
        SELECT id, workspace_id AS "workspaceId", name, description, status,
               created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
        FROM projects WHERE workspace_id = ${params.workspaceId} AND deleted_at IS NULL
        ORDER BY created_at DESC, id DESC LIMIT ${pageSize + 1}
      `;
    }

    let nextPageToken = "";
    if (projects.length > pageSize) {
      nextPageToken = projects[pageSize].id;
      projects = projects.slice(0, pageSize);
    }

    return { projects, nextPageToken, totalCount: count };
  }

  async update(params: UpdateProjectParams): Promise<Project> {
    const [row] = await this.sql<Project[]>`
      UPDATE projects SET name = ${params.name}, description = ${params.description},
             status = ${params.status}, updated_at = NOW()
      WHERE id = ${params.id} AND deleted_at IS NULL
      RETURNING id, workspace_id AS "workspaceId", name, description, status,
                created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
    `;
    if (!row) {
      throw new NotFoundError();
    }
    return row;
  }

  async delete(id: string): Promise<void> {
    const result = await this.sql`
      UPDATE projects SET deleted_at = NOW(), updated_at = NOW()
      WHERE id = ${id} AND deleted_at IS NULL
    `;
    if (result.count === 0) {
      throw new NotFoundError();
    }
  }
}
