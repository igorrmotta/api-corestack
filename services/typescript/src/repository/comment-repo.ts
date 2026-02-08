import type postgres from "postgres";
import { NotFoundError } from "./errors.js";

export interface Comment {
  id: string;
  taskId: string;
  authorId: string;
  content: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateCommentParams {
  taskId: string;
  authorId: string;
  content: string;
}

export interface ListCommentsParams {
  taskId: string;
  pageSize: number;
  pageToken: string;
}

export interface CommentList {
  comments: Comment[];
  nextPageToken: string;
  totalCount: number;
}

export class CommentRepo {
  constructor(private sql: postgres.Sql) {}

  async create(params: CreateCommentParams): Promise<Comment> {
    const [row] = await this.sql<Comment[]>`
      INSERT INTO task_comments (id, task_id, author_id, content, created_at, updated_at)
      VALUES (gen_random_uuid(), ${params.taskId}, ${params.authorId}, ${params.content}, NOW(), NOW())
      RETURNING id, task_id AS "taskId", author_id AS "authorId", content,
                created_at AS "createdAt", updated_at AS "updatedAt"
    `;
    return row;
  }

  async list(params: ListCommentsParams): Promise<CommentList> {
    let pageSize = params.pageSize;
    if (pageSize <= 0 || pageSize > 100) {
      pageSize = 20;
    }

    const [{ count }] = await this.sql<[{ count: number }]>`
      SELECT COUNT(*)::int AS count FROM task_comments WHERE task_id = ${params.taskId}
    `;

    let comments: Comment[];
    if (params.pageToken) {
      comments = await this.sql<Comment[]>`
        SELECT id, task_id AS "taskId", author_id AS "authorId", content,
               created_at AS "createdAt", updated_at AS "updatedAt"
        FROM task_comments WHERE task_id = ${params.taskId} AND id < ${params.pageToken}
        ORDER BY created_at DESC, id DESC LIMIT ${pageSize + 1}
      `;
    } else {
      comments = await this.sql<Comment[]>`
        SELECT id, task_id AS "taskId", author_id AS "authorId", content,
               created_at AS "createdAt", updated_at AS "updatedAt"
        FROM task_comments WHERE task_id = ${params.taskId}
        ORDER BY created_at DESC, id DESC LIMIT ${pageSize + 1}
      `;
    }

    let nextPageToken = "";
    if (comments.length > pageSize) {
      nextPageToken = comments[pageSize].id;
      comments = comments.slice(0, pageSize);
    }

    return { comments, nextPageToken, totalCount: count };
  }

  async delete(id: string): Promise<void> {
    const result = await this.sql`
      DELETE FROM task_comments WHERE id = ${id}
    `;
    if (result.count === 0) {
      throw new NotFoundError();
    }
  }
}
