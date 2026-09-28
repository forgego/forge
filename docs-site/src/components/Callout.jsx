import React from 'react';
import clsx from 'clsx';

const CalloutVariants = {
  info: {
    title: 'Good to know',
    className: 'admonition-info',
  },
  warning: {
    title: 'Warning',
    className: 'admonition-warning',
  },
  error: {
    title: 'Error',
    className: 'admonition-danger',
  },
  tip: {
    title: 'Tip',
    className: 'admonition-tip',
  },
};

export default function Callout({ children, type = 'info', title }) {
  const variant = CalloutVariants[type] || CalloutVariants.info;

  return (
    <div className={clsx('admonition', variant.className)}>
      <div className="admonition-heading">
        {title || variant.title}
      </div>
      <div className="admonition-content">
        {children}
      </div>
    </div>
  );
}
