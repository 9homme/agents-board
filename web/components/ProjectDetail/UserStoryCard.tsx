import React from 'react';
import { UserStoryListItem } from '../../lib/api/types';

interface UserStoryCardProps {
  story: UserStoryListItem;
  onSelect: (id: string) => void;
}

/**
 * A card representing a user story in a list.
 * Keyboard-accessible: role="button", tabIndex=0, Enter/Space triggers onSelect.
 */
export const UserStoryCard: React.FC<UserStoryCardProps> = ({ story, onSelect }) => {
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onSelect(story.id);
    }
  };

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => onSelect(story.id)}
      onKeyDown={handleKeyDown}
      className="p-4 border rounded shadow-sm hover:shadow-md cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500"
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
    </div>
  );
};
