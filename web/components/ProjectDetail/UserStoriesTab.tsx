import React from 'react';

/**
 * UserStoriesTab component.
 * Renders a verbatim placeholder copy for the User Stories tab.
 * No network calls — this is a static placeholder for a future release.
 */
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
      <UserStoryCardList projectId={projectId} onSelect={(_id) => {}} />
    </div>
  );
};
