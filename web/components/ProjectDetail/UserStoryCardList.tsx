import React from 'react';
import { useProjectUserStories } from '../../hooks/useProjectUserStories';
import { UserStoryCard } from './UserStoryCard';

interface UserStoryCardListProps {
  projectId: string;
  /** Called with the story id and the card button element (for focus management). */
  onSelect: (storyId: string, cardEl?: HTMLButtonElement) => void;
}

/**
 * Renders a list of user story cards for a project.
 * Shows loading, error, empty, and success states.
 */
export const UserStoryCardList: React.FC<UserStoryCardListProps> = ({ projectId, onSelect }) => {
  const { stories, loading, error } = useProjectUserStories(projectId);

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
        No user stories yet for this project.
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {stories.map(story => (
        <UserStoryCard
          key={story.id}
          story={story}
          onSelect={onSelect}
        />
      ))}
    </div>
  );
};
