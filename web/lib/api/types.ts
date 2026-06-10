export interface Project {
  id: string;
  name: string;
  description: string;
  /** Local filesystem path for the project (required, validated server-side). */
  path: string;
  createdAt: string;
  updatedAt: string;
}

/** Request body shape for POST /api/v1/projects (§3). */
export interface CreateProjectRequest {
  name: string;
  description?: string;
  path: string;
}

export interface ProjectsResponse {
  projects: Project[];
}

export interface ErrorResponse {
  code: string;
  message: string;
}

/**
 * A project requirement (read-only from FE; created/managed via MCP).
 * Matches §4 API contract field-for-field.
 */
export interface Requirement {
  id: string;
  projectId: string;
  name: string;
  description: string;
  /** Enum: "draft" | "in_progress" | "done" */
  status: string;
  createdAt: string; // ISO-8601 UTC
  updatedAt: string; // ISO-8601 UTC
}

/** Response shape for GET /api/v1/projects/{id}/requirements (§4). */
export interface RequirementsResponse {
  requirements: Requirement[];
}

/** A document list item (metadata only — no content field). */
export interface DocumentListItem {
  id: string;
  projectId: string;
  /**
   * Present on items returned from the requirement-scoped §10 endpoint.
   * Optional for backward-compat with legacy project-scoped responses (removed in US048).
   */
  requirementId?: string;
  title: string;
  createdAt: string; // ISO-8601 UTC
  updatedAt: string; // ISO-8601 UTC
}

/** Response shape for GET /api/v1/projects/{id}/requirements/{rid}/documents (§10). */
export interface DocumentsListResponse {
  documents: DocumentListItem[];
}

/** Full document including raw markdown content. */
export interface Document {
  id: string;
  projectId: string;
  /**
   * Present on documents returned from the requirement-scoped §11 endpoint.
   * Optional for backward-compat with legacy responses (removed in US048).
   */
  requirementId?: string;
  title: string;
  content: string; // raw markdown; MAY be ""
  createdAt: string;
  updatedAt: string;
}

/** A user story list item — includes taskCount (aggregate from BE JOIN). */
export interface UserStoryListItem {
  id: string;
  projectId: string;
  /**
   * Present on items returned from the requirement-scoped §6 endpoint.
   * Optional for backward-compat with legacy project-scoped responses (removed in US048).
   */
  requirementId?: string;
  title: string;
  description: string; // MAY be ""
  status: string;
  taskCount: number; // integer ≥ 0
  createdAt: string; // ISO-8601 UTC
  updatedAt: string; // ISO-8601 UTC
}

/** Response shape for GET /api/v1/projects/{id}/requirements/{rid}/user-stories (§6). */
export interface UserStoriesListResponse {
  userStories: UserStoryListItem[];
}

/**
 * Full user story detail object (bare, no taskCount).
 * Returned by GET /api/v1/projects/{pid}/requirements/{rid}/user-stories/{id} (§7).
 * taskCount is intentionally absent — derive from tasks.length on the FE.
 */
export interface UserStory {
  id: string;
  projectId: string;
  /**
   * Present on stories returned from the requirement-scoped §7 endpoint.
   * Optional for backward-compat with legacy responses (removed in US048).
   */
  requirementId?: string;
  title: string;
  description: string; // MAY be ""
  status: string;
  createdAt: string; // ISO-8601 UTC
  updatedAt: string; // ISO-8601 UTC
}

/** A single task. */
export interface Task {
  id: string;
  userStoryId: string;
  title: string;
  description: string; // MAY be ""
  status: string;
  createdAt: string; // ISO-8601 UTC
  updatedAt: string; // ISO-8601 UTC
}

/** Response shape for GET /api/v1/user-stories/{id}/tasks. */
export interface TasksListResponse {
  tasks: Task[];
}
