import React, { useEffect, useRef } from 'react';
import { useUserStory } from '../../hooks/useUserStory';
import { useUserStoryTasks } from '../../hooks/useUserStoryTasks';

interface UserStoryDrawerProps {
  /** The id of the user story to display. */
  storyId: string;
  /** Callback invoked when the drawer should close (X button or Escape key). */
  onClose: () => void;
}

/**
 * UserStoryDrawer component.
 *
 * Right-side detail drawer displaying a selected user story's full details
 * and its tasks. Issues two parallel fetches: story detail + task list.
 *
 * Accessibility:
 * - Uses native `<dialog open>` element (implicit role=dialog, aria-modal).
 * - Close button with accessible label.
 * - Escape key listener on document (supplements native dialog behaviour).
 * - Focus moves to the close button on mount; caller restores focus on close.
 */
export const UserStoryDrawer: React.FC<UserStoryDrawerProps> = ({
  storyId,
  onClose,
}) => {
  const { data: story, isLoading: storyLoading, error: storyError } = useUserStory(storyId);
  const { data: tasksData, isLoading: tasksLoading } = useUserStoryTasks(storyId);

  const closeButtonRef = useRef<HTMLButtonElement>(null);

  // Move focus to the close button when the drawer mounts
  useEffect(() => {
    closeButtonRef.current?.focus();
  }, []);

  // Escape key listener — ensures Escape closes the drawer in all environments
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [onClose]);

  const isLoading = storyLoading || tasksLoading;

  return (
    // `open` keeps the drawer visible; parent unmounts it on close.
    // aria-modal supplements the native dialog implicit role for screen readers.
    <dialog
      open
      aria-modal="true"
      aria-label={story?.title ?? 'User story detail'}
      className="static m-0 w-96 h-full max-h-none bg-white shadow-xl border-l border-gray-200 flex flex-col p-0 shrink-0"
    >
      {/* Header with close button — always rendered */}
      <div className="flex items-center justify-between p-4 border-b border-gray-200">
        <h2 className="text-lg font-semibold text-gray-800 truncate">
          {story?.title ?? 'User Story'}
        </h2>
        <button
          ref={closeButtonRef}
          type="button"
          aria-label="Close"
          onClick={onClose}
          className="ml-2 p-1 rounded hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <svg
            className="w-5 h-5 text-gray-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      {/* Body */}
      <div className="flex-1 overflow-y-auto p-4">
        {isLoading && (
          // role="status" is the correct ARIA live region role for spinners —
          // no HTML element maps to this implicit role, so the attribute is intentional.
          <div
            role="status"
            aria-label="Loading user story"
            className="flex items-center justify-center h-32"
          >
            <div className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin" />
          </div>
        )}

        {!isLoading && storyError && (
          <div className="text-center py-8">
            <p className="text-red-600 font-medium">
              Couldn&apos;t load this user story.
            </p>
            <p className="text-sm text-gray-500 mt-1">Please try again later.</p>
          </div>
        )}

        {!isLoading && !storyError && story && (
          <>
            {/* Story detail */}
            <div className="mb-6">
              <div className="flex items-center gap-2 mb-3">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  {story.status}
                </span>
              </div>
              <p className="text-sm text-gray-700 leading-relaxed">{story.description}</p>
            </div>

            {/* Tasks section */}
            <div>
              <h3 className="text-sm font-semibold text-gray-800 mb-3">
                Tasks ({tasksData?.tasks.length ?? 0})
              </h3>

              {tasksData?.tasks.length === 0 ? (
                <p className="text-sm text-gray-500">No tasks for this story.</p>
              ) : (
                <ul className="space-y-2">
                  {tasksData?.tasks.map((task) => (
                    <li
                      key={task.id}
                      className="p-3 bg-gray-50 rounded-lg border border-gray-200"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <span className="text-sm font-medium text-gray-800">
                          {task.title}
                        </span>
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600 whitespace-nowrap">
                          {task.status}
                        </span>
                      </div>
                      {task.description && (
                        <p className="text-xs text-gray-500 mt-1">{task.description}</p>
                      )}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </>
        )}
      </div>
    </dialog>
  );
};
