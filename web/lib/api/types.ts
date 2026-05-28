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
