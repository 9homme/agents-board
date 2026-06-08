import React from 'react';

interface UserStoryCardListProps {
  projectId: string;
  onSelect: (storyId: string) => void;
}

/**
 * Stub — will be replaced with full loading/error/empty/success implementation.
 */
export const UserStoryCardList: React.FC<UserStoryCardListProps> = () => {
  return <div>stub</div>;
};
