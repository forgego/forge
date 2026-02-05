import React, { useState } from 'react';
import clsx from 'clsx';

export default function Tabs({ children, defaultValue, values, groupId }) {
  const [selectedValue, setSelectedValue] = useState(defaultValue || values[0]?.value);

  const selectedTab = values.find((v) => v.value === selectedValue);

  return (
    <div className="tabs-container">
      <div className="tabs-list" role="tablist">
        {values.map(({ value, label }) => (
          <button
            key={value}
            role="tab"
            className={clsx('tabs-tab', {
              'tabs-tab--active': selectedValue === value,
            })}
            onClick={() => setSelectedValue(value)}
            aria-selected={selectedValue === value}
          >
            {label}
          </button>
        ))}
      </div>
      <div className="tabs-content">
        {React.Children.map(children, (child, index) => {
          if (child.props.value === selectedValue) {
            return <div key={index}>{child}</div>;
          }
          return null;
        })}
      </div>
    </div>
  );
}

export function TabItem({ children, value, label }) {
  return <div data-value={value} data-label={label}>{children}</div>;
}
