import { InvalidInputError } from "../repository/errors.js";
import {
  WorkspaceRepo,
  type CreateWorkspaceParams,
  type ListWorkspacesParams,
  type UpdateWorkspaceParams,
  type Workspace,
  type WorkspaceList,
} from "../repository/workspace-repo.js";
import type { Logger } from "pino";

export class WorkspaceService {
  constructor(
    private repo: WorkspaceRepo,
    private logger: Logger,
  ) {}

  async create(params: CreateWorkspaceParams): Promise<Workspace> {
    if (!params.name) {
      throw new InvalidInputError("invalid input: name is required");
    }
    if (!params.slug) {
      throw new InvalidInputError("invalid input: slug is required");
    }
    this.logger.debug({ name: params.name, slug: params.slug }, "creating workspace");
    return this.repo.create(params);
  }

  async getById(id: string): Promise<Workspace> {
    return this.repo.getById(id);
  }

  async list(params: ListWorkspacesParams): Promise<WorkspaceList> {
    return this.repo.list(params);
  }

  async update(params: UpdateWorkspaceParams): Promise<Workspace> {
    if (!params.name) {
      throw new InvalidInputError("invalid input: name is required");
    }
    return this.repo.update(params);
  }

  async delete(id: string): Promise<void> {
    return this.repo.delete(id);
  }
}
