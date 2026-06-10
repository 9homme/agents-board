import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useCreateProject } from '../../hooks/useCreateProject';

/** Derive the basename from a path string (no trailing slash, no Node.js path module). */
function basename(pathStr: string): string {
  return pathStr.split('/').filter(Boolean).pop() ?? '';
}

interface AddProjectDialogProps {
  /** Whether the dialog is currently open. */
  open: boolean;
  /** Called when the dialog should close (Cancel or Escape). */
  onClose: () => void;
  /** Called after a successful project creation. */
  onSuccess: () => void;
  /**
   * Ref to the trigger button that opened this dialog.
   * When provided, focus is returned to it on close.
   */
  triggerRef?: React.RefObject<HTMLButtonElement>;
}

/**
 * AddProjectDialog — modal dialog for creating a new project.
 *
 * Features:
 * - Plain text path input (no autocomplete / suggestions per D-005).
 * - Name auto-fills from path basename on each keystroke.
 * - Auto-fill is "sticky-off" once the user manually edits the name field.
 * - Submit is disabled until both name and path are non-blank.
 * - Shows in-flight loading state; prevents double-submit.
 * - Surfaces 400 VALIDATION_ERROR and 409 DUPLICATE_PATH errors inline.
 * - Preserves inputs on server error (does not reset form).
 * - Returns focus to `triggerRef` on close.
 */
export const AddProjectDialog: React.FC<AddProjectDialogProps> = ({
  open,
  onClose,
  onSuccess,
  triggerRef,
}) => {
  const [path, setPath] = useState('');
  const [name, setName] = useState('');
  // Once the user manually edits the name field, auto-fill is disabled.
  // Stored in a ref (not state) since this control flag never drives rendering.
  const autoFillEnabledRef = useRef(true);

  const { createProject, status, error, reset } = useCreateProject();

  const firstFocusableRef = useRef<HTMLInputElement>(null);

  const isSubmitting = status === 'submitting';
  const canSubmit = name.trim().length > 0 && path.trim().length > 0 && !isSubmitting;

  /** Reset form state and notify parent to close. */
  const handleClose = useCallback(() => {
    reset();
    setPath('');
    setName('');
    autoFillEnabledRef.current = true;
    onClose();
  }, [reset, onClose]);

  // Focus the first input when dialog opens; return focus to trigger on close.
  useEffect(() => {
    if (open) {
      firstFocusableRef.current?.focus();
    } else {
      triggerRef?.current?.focus();
    }
  }, [open, triggerRef]);

  // Close on Escape key.
  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        handleClose();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [open, handleClose]);

  function handlePathChange(e: React.ChangeEvent<HTMLInputElement>) {
    const newPath = e.target.value;
    setPath(newPath);
    if (autoFillEnabledRef.current) {
      setName(basename(newPath));
    }
  }

  function handleNameChange(e: React.ChangeEvent<HTMLInputElement>) {
    autoFillEnabledRef.current = false; // sticky-off: user took manual control of name
    setName(e.target.value);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;

    try {
      await createProject({ name: name.trim(), path: path.trim() });
      handleClose();
      onSuccess();
    } catch {
      // Error is already surfaced via `error` from the hook; form stays open.
    }
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <dialog
        open
        aria-labelledby="add-project-dialog-title"
        className="bg-white rounded-lg shadow-xl w-full max-w-md mx-4 p-6 m-0"
      >
        <h2
          id="add-project-dialog-title"
          className="text-xl font-semibold text-gray-900 mb-5"
        >
          Add Project
        </h2>

        <form onSubmit={handleSubmit} noValidate>
          {/* Path field */}
          <div className="mb-4">
            <label
              htmlFor="add-project-path"
              className="block text-sm font-medium text-gray-700 mb-1"
            >
              Path
            </label>
            <input
              id="add-project-path"
              ref={firstFocusableRef}
              type="text"
              value={path}
              onChange={handlePathChange}
              placeholder="/Users/you/workspace/my-project"
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              aria-describedby={error ? 'add-project-error' : undefined}
              disabled={isSubmitting}
              autoComplete="off"
            />
          </div>

          {/* Name field */}
          <div className="mb-5">
            <label
              htmlFor="add-project-name"
              className="block text-sm font-medium text-gray-700 mb-1"
            >
              Name
            </label>
            <input
              id="add-project-name"
              type="text"
              value={name}
              onChange={handleNameChange}
              placeholder="my-project"
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={isSubmitting}
              autoComplete="off"
            />
          </div>

          {/* Inline error */}
          {error && (
            <div
              id="add-project-error"
              role="alert"
              aria-live="assertive"
              className="mb-4 rounded-md bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700"
            >
              {error.message}
            </div>
          )}

          {/* Action buttons */}
          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!canSubmit}
              aria-busy={isSubmitting}
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isSubmitting ? 'Creating...' : 'Create Project'}
            </button>
          </div>
        </form>
      </dialog>
    </div>
  );
};
