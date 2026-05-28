import React from 'react';

/**
 * UserStoriesTab component.
 * Renders a verbatim placeholder copy for the User Stories tab.
 * No network calls — this is a static placeholder for a future release.
 */
export const UserStoriesTab: React.FC = () => {
  return (
    <div
      role="tabpanel"
      id="tabpanel-user-stories"
      aria-labelledby="tab-user-stories"
      className="p-4"
    >
      <p>Coming soon — user stories will appear here in a future release.</p>
    </div>
  );
};
