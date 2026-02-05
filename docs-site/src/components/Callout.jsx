import React from 'react';
import clsx from 'clsx';

const CalloutVariants = {
  info: {
    icon: '💡',
    title: 'Good to know',
    className: 'admonition-info',
  },
  warning: {
    icon: '⚠️',
    title: 'Warning',
    className: 'admonition-warning',
  },
  error: {
    icon: '❌',
    title: 'Error',
    className: 'admonition-danger',
  },
  tip: {
    icon: '💡',
    title: 'Tip',
    className: 'admonition-tip',
  },
};

export default function Callout({ children, type = 'info', title }) {
  const variant = CalloutVariants[type] || CalloutVariants.info;

  return (
    <div className={clsx('admonition', variant.className)}>
      <div className="admonition-heading">
        <span className="admonition-icon">{variant.icon}</span>
        {title || variant.title}
      </div>
      <div className="admonition-content">
        {children}
      </div>
    </div>
  );
}
