import React from 'react';
import { UserStoryListItem } from '../../lib/api/types';

interface UserStoryCardProps {
  story: UserStoryListItem;
  /** Called with the story id and the button element (for focus management). */
  onSelect: (id: string, cardEl: HTMLButtonElement) => void;
}

/**
 * A card representing a user story in a list.
 * Keyboard-accessible: uses a semantic <button> element so Enter/Space are
 * handled natively; tabIndex and keyboard events are therefore implicit.
 *
 * Passes its own DOM node to onSelect so the parent can return focus when
 * the drawer closes (architecture D-005 focus management requirement).
 */
export const UserStoryCard: React.FC<UserStoryCardProps> = ({ story, onSelect }) => {
  return (
    <button
      type="button"
      onClick={(e) => onSelect(story.id, e.currentTarget)}
      className="w-full text-left p-4 border rounded shadow-sm hover:shadow-md cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500"
      aria-label={`User story: ${story.title}`}
    >
      <div className="flex justify-between items-start mb-2">
        <h3 className="text-lg font-semibold">{story.title}</h3>
        <span className="px-2 py-1 text-xs font-medium rounded bg-gray-100 text-gray-800 uppercase tracking-wider">
          {story.status}
        </span>
      </div>

      <p className="text-sm text-gray-600 mb-4 line-clamp-3">
        {story.description}
      </p>

      <div className="flex items-center text-xs text-gray-500">
        <span className="font-medium">{story.taskCount} tasks</span>
      </div>
    </button>
  );
};
