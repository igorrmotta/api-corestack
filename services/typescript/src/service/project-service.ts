import { InvalidInputError } from "../repository/errors.js";
import {
  ProjectRepo,
  type CreateProjectParams,
  type ListProjectsParams,
  type Project,
  type ProjectList,
  type UpdateProjectParams,
} from "../repository/project-repo.js";
import type { Logger } from "pino";

export class ProjectService {
  constructor(
    private repo: ProjectRepo,
    private logger: Logger,
  ) {}

  async create(params: CreateProjectParams): Promise<Project> {
    if (!params.name) {
      throw new InvalidInputError("invalid input: name is required");
    }
    if (!params.workspaceId) {
      throw new InvalidInputError("invalid input: workspace_id is required");
    }
    this.logger.debug({ name: params.name, workspaceId: params.workspaceId }, "creating project");
    return this.repo.create(params);
  }

  async getById(id: string): Promise<Project> {
    return this.repo.getById(id);
  }

  async list(params: ListProjectsParams): Promise<ProjectList> {
    return this.repo.list(params);
  }

  async update(params: UpdateProjectParams): Promise<Project> {
    if (!params.name) {
      throw new InvalidInputError("invalid input: name is required");
    }
    const validStatuses: Record<string, boolean> = { active: true, archived: true };
    if (params.status && !validStatuses[params.status]) {
      throw new InvalidInputError(`invalid input: invalid status: ${params.status}`);
    }
    return this.repo.update(params);
  }

  async delete(id: string): Promise<void> {
    return this.repo.delete(id);
  }
}
