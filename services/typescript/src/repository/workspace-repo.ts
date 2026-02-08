import type postgres from "postgres";
import { ConflictError, NotFoundError, isPgUniqueViolation } from "./errors.js";

export interface Workspace {
  id: string;
  name: string;
  slug: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;
}

export interface CreateWorkspaceParams {
  name: string;
  slug: string;
}

export interface UpdateWorkspaceParams {
  id: string;
  name: string;
  slug: string;
}

export interface ListWorkspacesParams {
  pageSize: number;
  pageToken: string;
}

export interface WorkspaceList {
  workspaces: Workspace[];
  nextPageToken: string;
  totalCount: number;
}

export class WorkspaceRepo {
  constructor(private sql: postgres.Sql) {}

  async create(params: CreateWorkspaceParams): Promise<Workspace> {
    try {
      const [row] = await this.sql<Workspace[]>`
        INSERT INTO workspaces (id, name, slug, created_at, updated_at)
        VALUES (gen_random_uuid(), ${params.name}, ${params.slug}, NOW(), NOW())
        RETURNING id, name, slug, created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
      `;
      return row;
    } catch (err) {
      if (isPgUniqueViolation(err)) {
        throw new ConflictError("create workspace: conflict");
      }
      throw err;
    }
  }

  async getById(id: string): Promise<Workspace> {
    const [row] = await this.sql<Workspace[]>`
      SELECT id, name, slug, created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
      FROM workspaces WHERE id = ${id} AND deleted_at IS NULL
    `;
    if (!row) {
      throw new NotFoundError();
    }
    return row;
  }

  async list(params: ListWorkspacesParams): Promise<WorkspaceList> {
    let pageSize = params.pageSize;
    if (pageSize <= 0 || pageSize > 100) {
      pageSize = 20;
    }

    const [{ count }] = await this.sql<[{ count: number }]>`
      SELECT COUNT(*)::int AS count FROM workspaces WHERE deleted_at IS NULL
    `;

    let workspaces: Workspace[];
    if (params.pageToken) {
      workspaces = await this.sql<Workspace[]>`
        SELECT id, name, slug, created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
        FROM workspaces WHERE deleted_at IS NULL AND id < ${params.pageToken}
        ORDER BY created_at DESC, id DESC LIMIT ${pageSize + 1}
      `;
    } else {
      workspaces = await this.sql<Workspace[]>`
        SELECT id, name, slug, created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
        FROM workspaces WHERE deleted_at IS NULL
        ORDER BY created_at DESC, id DESC LIMIT ${pageSize + 1}
      `;
    }

    let nextPageToken = "";
    if (workspaces.length > pageSize) {
      nextPageToken = workspaces[pageSize].id;
      workspaces = workspaces.slice(0, pageSize);
    }

    return { workspaces, nextPageToken, totalCount: count };
  }

  async update(params: UpdateWorkspaceParams): Promise<Workspace> {
    try {
      const [row] = await this.sql<Workspace[]>`
        UPDATE workspaces SET name = ${params.name}, slug = ${params.slug}, updated_at = NOW()
        WHERE id = ${params.id} AND deleted_at IS NULL
        RETURNING id, name, slug, created_at AS "createdAt", updated_at AS "updatedAt", deleted_at AS "deletedAt"
      `;
      if (!row) {
        throw new NotFoundError();
      }
      return row;
    } catch (err) {
      if (err instanceof NotFoundError) throw err;
      if (isPgUniqueViolation(err)) {
        throw new ConflictError("update workspace: conflict");
      }
      throw err;
    }
  }

  async delete(id: string): Promise<void> {
    const result = await this.sql`
      UPDATE workspaces SET deleted_at = NOW(), updated_at = NOW()
      WHERE id = ${id} AND deleted_at IS NULL
    `;
    if (result.count === 0) {
      throw new NotFoundError();
    }
  }
}
