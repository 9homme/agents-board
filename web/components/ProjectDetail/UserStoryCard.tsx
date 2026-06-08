import React from 'react';
import { UserStoryListItem } from '../../lib/api/types';

interface UserStoryCardProps {
  story: UserStoryListItem;
  onSelect: (id: string) => void;
}

/**
 * Stub — will be replaced with full accessible card implementation.
 */
export const UserStoryCard: React.FC<UserStoryCardProps> = ({ story }) => {
  return (
    <div>
      <h3>{story.title}</h3>
    </div>
  );
};
