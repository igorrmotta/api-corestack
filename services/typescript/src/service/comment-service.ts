import { InvalidInputError } from "../repository/errors.js";
import {
  CommentRepo,
  type Comment,
  type CommentList,
  type CreateCommentParams,
  type ListCommentsParams,
} from "../repository/comment-repo.js";
import type { Logger } from "pino";

const EMPTY_UUID = "00000000-0000-0000-0000-000000000000";

export class CommentService {
  constructor(
    private repo: CommentRepo,
    private logger: Logger,
  ) {}

  async create(params: CreateCommentParams): Promise<Comment> {
    if (!params.taskId || params.taskId === EMPTY_UUID) {
      throw new InvalidInputError("invalid input: task_id is required");
    }
    if (!params.authorId) {
      throw new InvalidInputError("invalid input: author_id is required");
    }
    if (!params.content) {
      throw new InvalidInputError("invalid input: content is required");
    }
    this.logger.debug({ taskId: params.taskId, authorId: params.authorId }, "creating comment");
    return this.repo.create(params);
  }

  async list(params: ListCommentsParams): Promise<CommentList> {
    return this.repo.list(params);
  }

  async delete(id: string): Promise<void> {
    return this.repo.delete(id);
  }
}
