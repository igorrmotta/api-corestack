import type postgres from "postgres";
import { NotFoundError } from "./errors.js";

export interface Notification {
  id: number;
  workspaceId: string;
  eventType: string;
  payload: Record<string, unknown>;
  status: string;
  retryCount: number;
  maxRetries: number;
  nextRetryAt: Date | null;
  lastError: string;
  createdAt: Date;
  processedAt: Date | null;
}

export interface CreateNotificationParams {
  workspaceId: string;
  eventType: string;
  payload: Record<string, unknown>;
}

export interface ListNotificationsParams {
  workspaceId: string;
  status: string;
  pageSize: number;
  pageToken: string;
}

export interface NotificationList {
  notifications: Notification[];
  nextPageToken: string;
  totalCount: number;
}

export class NotificationRepo {
  constructor(private sql: postgres.Sql) {}

  async create(params: CreateNotificationParams): Promise<Notification> {
    const [row] = await this.sql<Notification[]>`
      INSERT INTO notification_queue (workspace_id, event_type, payload, status, created_at)
      VALUES (${params.workspaceId}, ${params.eventType}, ${this.sql.json(params.payload as never)}, 'pending', NOW())
      RETURNING id, workspace_id AS "workspaceId", event_type AS "eventType", payload,
                status, retry_count AS "retryCount", max_retries AS "maxRetries",
                next_retry_at AS "nextRetryAt", COALESCE(last_error, '') AS "lastError",
                created_at AS "createdAt", processed_at AS "processedAt"
    `;
    return row;
  }

  async list(params: ListNotificationsParams): Promise<NotificationList> {
    let pageSize = params.pageSize;
    if (pageSize <= 0 || pageSize > 100) {
      pageSize = 20;
    }

    // Parse offset from page token (offset-based pagination for notifications)
    let offset = 0;
    if (params.pageToken) {
      offset = parseInt(params.pageToken, 10);
      if (isNaN(offset)) {
        throw new Error(`invalid page token: ${params.pageToken}`);
      }
    }

    let totalCount: number;
    if (params.status) {
      const [{ count }] = await this.sql<[{ count: number }]>`
        SELECT COUNT(*)::int AS count FROM notification_queue
        WHERE workspace_id = ${params.workspaceId} AND status = ${params.status}
      `;
      totalCount = count;
    } else {
      const [{ count }] = await this.sql<[{ count: number }]>`
        SELECT COUNT(*)::int AS count FROM notification_queue
        WHERE workspace_id = ${params.workspaceId}
      `;
      totalCount = count;
    }

    let notifications: Notification[];
    if (params.status) {
      notifications = await this.sql<Notification[]>`
        SELECT id, workspace_id AS "workspaceId", event_type AS "eventType", payload,
               status, retry_count AS "retryCount", max_retries AS "maxRetries",
               next_retry_at AS "nextRetryAt", COALESCE(last_error, '') AS "lastError",
               created_at AS "createdAt", processed_at AS "processedAt"
        FROM notification_queue
        WHERE workspace_id = ${params.workspaceId} AND status = ${params.status}
        ORDER BY created_at DESC
        LIMIT ${pageSize} OFFSET ${offset}
      `;
    } else {
      notifications = await this.sql<Notification[]>`
        SELECT id, workspace_id AS "workspaceId", event_type AS "eventType", payload,
               status, retry_count AS "retryCount", max_retries AS "maxRetries",
               next_retry_at AS "nextRetryAt", COALESCE(last_error, '') AS "lastError",
               created_at AS "createdAt", processed_at AS "processedAt"
        FROM notification_queue
        WHERE workspace_id = ${params.workspaceId}
        ORDER BY created_at DESC
        LIMIT ${pageSize} OFFSET ${offset}
      `;
    }

    let nextPageToken = "";
    const nextOffset = offset + pageSize;
    if (nextOffset < totalCount) {
      nextPageToken = String(nextOffset);
    }

    return { notifications, nextPageToken, totalCount };
  }

  async markProcessed(id: number): Promise<Notification> {
    const [row] = await this.sql<Notification[]>`
      UPDATE notification_queue
      SET status = 'processed', processed_at = NOW()
      WHERE id = ${id}
      RETURNING id, workspace_id AS "workspaceId", event_type AS "eventType", payload,
                status, retry_count AS "retryCount", max_retries AS "maxRetries",
                next_retry_at AS "nextRetryAt", COALESCE(last_error, '') AS "lastError",
                created_at AS "createdAt", processed_at AS "processedAt"
    `;
    if (!row) {
      throw new NotFoundError();
    }
    return row;
  }

  async fetchPending(limit: number): Promise<Notification[]> {
    return this.sql<Notification[]>`
      SELECT id, workspace_id AS "workspaceId", event_type AS "eventType", payload,
             status, retry_count AS "retryCount", max_retries AS "maxRetries",
             next_retry_at AS "nextRetryAt", COALESCE(last_error, '') AS "lastError",
             created_at AS "createdAt", processed_at AS "processedAt"
      FROM notification_queue
      WHERE status IN ('pending', 'failed')
        AND (next_retry_at IS NULL OR next_retry_at <= NOW())
      ORDER BY created_at ASC
      LIMIT ${limit}
      FOR UPDATE SKIP LOCKED
    `;
  }

  async markFailed(id: number, errMsg: string): Promise<void> {
    const result = await this.sql`
      UPDATE notification_queue
      SET status = 'failed',
          last_error = ${errMsg},
          retry_count = retry_count + 1,
          next_retry_at = NOW() + (INTERVAL '1 second' * POWER(2, retry_count + 1))
      WHERE id = ${id}
    `;
    if (result.count === 0) {
      throw new NotFoundError();
    }
  }
}
