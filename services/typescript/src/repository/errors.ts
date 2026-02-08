export class RepositoryError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "RepositoryError";
  }
}

export class NotFoundError extends RepositoryError {
  constructor(message = "not found") {
    super(message);
    this.name = "NotFoundError";
  }
}

export class AlreadyExistsError extends RepositoryError {
  constructor(message = "already exists") {
    super(message);
    this.name = "AlreadyExistsError";
  }
}

export class InvalidInputError extends RepositoryError {
  constructor(message = "invalid input") {
    super(message);
    this.name = "InvalidInputError";
  }
}

export class ConflictError extends RepositoryError {
  constructor(message = "conflict") {
    super(message);
    this.name = "ConflictError";
  }
}

export function isPgUniqueViolation(err: unknown): boolean {
  return (
    typeof err === "object" &&
    err !== null &&
    "code" in err &&
    (err as { code: string }).code === "23505"
  );
}
