import React from 'react';
import clsx from 'clsx';

export default function VersionBadge({ version, children }) {
  return (
    <span className={clsx('badge', 'badge--primary')}>
      {children || `Available since ${version}`}
    </span>
  );
}
