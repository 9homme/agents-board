import React, { useState, useRef } from 'react';
import { UserStoryCardList } from './UserStoryCardList';
import { UserStoryDrawer } from './UserStoryDrawer';

interface UserStoriesTabProps {
  /** The project id whose user stories are being browsed. */
  projectId: string;
}

/**
 * UserStoriesTab component.
 *
 * Orchestrates the user story card list and the right-side detail drawer.
 *
 * State strategy (architecture D-005):
 * - `selectedStoryId` is the source of truth for which story is open in the drawer.
 * - `UserStoryCardList` is always rendered; `UserStoryDrawer` renders only when
 *   `selectedStoryId !== null`.
 * - Selecting a different card sets a new id (drawer stays mounted, re-fetches).
 * - Closing sets id to null (drawer unmounts; list stays mounted).
 * - No URL/router involvement (D-006 — in-tab state only, no deep-link).
 *
 * Focus management:
 * - On open: focus moves into the drawer (close button via useEffect in UserStoryDrawer).
 * - On close: focus returns to the triggering card via triggerRef.
 */
export const UserStoriesTab: React.FC<UserStoriesTabProps> = ({ projectId }) => {
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
        <UserStoryCardList
          projectId={projectId}
          onSelect={(storyId, cardEl) => handleSelect(storyId, cardEl)}
        />
      </div>

      {/* Detail drawer — mounted only when a story is selected */}
      {selectedStoryId !== null && (
        <UserStoryDrawer storyId={selectedStoryId} onClose={handleClose} />
      )}
    </div>
  );
};
