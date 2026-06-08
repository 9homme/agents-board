import React from 'react';
import { UserStoryCardList } from './UserStoryCardList';

interface UserStoriesTabProps {
  projectId: string;
}

/**
 * UserStoriesTab component.
 * Renders the list of user stories for a project.
 */
export const UserStoriesTab: React.FC<UserStoriesTabProps> = ({ projectId }) => {
  return (
    <div
      role="tabpanel"
      id="tabpanel-user-stories"
      aria-labelledby="tab-user-stories"
      className="p-4"
    >
      {/* onSelect is a no-op stub; the detail drawer will be wired in US005 */}
      <UserStoryCardList projectId={projectId} onSelect={() => {}} />
    </div>
  );
};
