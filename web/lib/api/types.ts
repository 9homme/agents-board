export interface Project {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectsResponse {
  projects: Project[];
}

export interface ErrorResponse {
  code: string;
  message: string;
}

/** A document list item (metadata only — no content field). */
export interface DocumentListItem {
  id: string;
  projectId: string;
  title: string;
  createdAt: string; // ISO-8601 UTC
  updatedAt: string; // ISO-8601 UTC
}

/** Response shape for GET /api/v1/projects/{id}/documents. */
export interface DocumentsListResponse {
  documents: DocumentListItem[];
}

/** Full document including raw markdown content. */
export interface Document {
  id: string;
  projectId: string;
  title: string;
  content: string; // raw markdown; MAY be ""
  createdAt: string;
  updatedAt: string;
}

/** A user story list item — includes taskCount (aggregate from BE JOIN). */
export interface UserStoryListItem {
  id: string;
  projectId: string;
  title: string;
  description: string; // MAY be ""
  status: string;
  taskCount: number; // integer ≥ 0
  createdAt: string; // ISO-8601 UTC
  updatedAt: string; // ISO-8601 UTC
}

/** Response shape for GET /api/v1/projects/{id}/user-stories. */
export interface UserStoriesListResponse {
  userStories: UserStoryListItem[];
}

/**
 * Full user story detail object (bare, no taskCount).
 * Returned by GET /api/v1/user-stories/{id}.
 * taskCount is intentionally absent — derive from tasks.length on the FE.
 */
export interface UserStory {
  id: string;
  projectId: string;
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
