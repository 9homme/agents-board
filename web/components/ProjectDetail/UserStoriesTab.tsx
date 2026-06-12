import React, { useState, useRef } from 'react';
import { UserStoryCardList } from './UserStoryCardList';
import { UserStoryDrawer } from './UserStoryDrawer';
import { useRequirementUserStories } from '../../hooks/useRequirementUserStories';
import { UserStoryCard } from './UserStoryCard';

interface UserStoriesTabProps {
  /** The project id whose user stories are being browsed. */
  projectId: string;
  /**
   * The requirement id that scopes user-stories via the §6 canonical path.
   * When provided, fetches from /requirements/:rid/user-stories.
   * When absent, falls back to the project-scoped UserStoryCardList (legacy).
   */
  requirementId?: string;
}

/**
 * UserStoriesTab component.
 *
 * Orchestrates the user story card list and the right-side detail drawer.
 *
 * When `requirementId` is provided, fetches from the canonical §6 endpoint:
 *   GET /api/v1/projects/{projectId}/requirements/{requirementId}/user-stories
 *
 * When `requirementId` is absent, delegates to `UserStoryCardList` (project-scoped,
 * legacy behaviour — will be removed in US048 once flat routes are gone).
 *
 * State strategy (architecture D-005):
 * - `selectedStoryId` is the source of truth for which story is open in the drawer.
 * - Cards list is always rendered; `UserStoryDrawer` renders only when `selectedStoryId !== null`.
 * - Selecting a different card sets a new id (drawer stays mounted, re-fetches).
 * - Closing sets id to null (drawer unmounts; list stays mounted).
 * - No URL/router involvement for stories (D-006 — in-tab state only, no deep-link).
 *
 * Focus management:
 * - On open: focus moves into the drawer (close button via useEffect in UserStoryDrawer).
 * - On close: focus returns to the triggering card via triggerRef.
 */
export const UserStoriesTab: React.FC<UserStoriesTabProps> = ({ projectId, requirementId }) => {
  const [selectedStoryId, setSelectedStoryId] = useState<string | null>(null);
  // Ref to the most recently clicked card button so we can return focus on close
  const triggerRef = useRef<HTMLButtonElement | null>(null);

  const handleSelect = (storyId: string, cardElement?: HTMLButtonElement) => {
    if (cardElement) {
      triggerRef.current = cardElement;
    }
    setSelectedStoryId(storyId);
  };

  const handleClose = () => {
    setSelectedStoryId(null);
    // Return focus to the card that triggered the drawer open
    triggerRef.current?.focus();
    triggerRef.current = null;
  };

  return (
    <div
      role="tabpanel"
      id="tabpanel-user-stories"
      aria-labelledby="tab-user-stories"
      className="relative flex h-full"
    >
      {/* Card list — always mounted; flex-1 ensures it shrinks to make room for the in-flow drawer */}
      <div className="flex-1 p-4 overflow-y-auto">
        {requirementId !== undefined ? (
          <RequirementScopedStoryList
            projectId={projectId}
            requirementId={requirementId}
            onSelect={(storyId, cardEl) => handleSelect(storyId, cardEl)}
          />
        ) : (
          <UserStoryCardList
            projectId={projectId}
            onSelect={(storyId, cardEl) => handleSelect(storyId, cardEl)}
          />
        )}
      </div>

      {/* Detail drawer — mounted only when a story is selected and requirement is known */}
      {selectedStoryId !== null && requirementId !== undefined && (
        <UserStoryDrawer
          projectId={projectId}
          requirementId={requirementId}
          storyId={selectedStoryId}
          onClose={handleClose}
        />
      )}
    </div>
  );
};

/** Internal component: renders requirement-scoped stories from §6 endpoint. */
interface RequirementScopedStoryListProps {
  projectId: string;
  requirementId: string;
  onSelect: (storyId: string, cardEl: HTMLButtonElement) => void;
}

const RequirementScopedStoryList: React.FC<RequirementScopedStoryListProps> = ({
  projectId,
  requirementId,
  onSelect,
}) => {
  const { stories, loading, error } = useRequirementUserStories(projectId, requirementId);

  if (loading) {
    return (
      <div className="flex justify-center items-center p-8" aria-busy="true">
        <div className="text-gray-500 italic">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 border border-red-200 bg-red-50 text-red-700 rounded">
        Couldn&apos;t load user stories. Please try again later.
      </div>
    );
  }

  if (stories.length === 0) {
    return (
      <div className="p-8 text-center text-gray-500 border-2 border-dashed rounded">
        No user stories yet for this requirement.
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {stories.map(story => (
        <div
          key={story.id}
          {...(story.requirementId !== undefined
            ? { 'data-requirement-id': story.requirementId }
            : {})}
        >
          <UserStoryCard
            story={story}
            onSelect={onSelect}
          />
        </div>
      ))}
    </div>
  );
};
