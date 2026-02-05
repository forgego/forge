/**
 * Sidebar configuration for forge documentation
 * Multi-persona navigation structure for different user types
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  // Getting Started sidebar - For beginners
  gettingStarted: [
    {
      type: 'doc',
      id: 'introduction',
      label: 'Introduction',
    },
    {
      type: 'category',
      label: 'Concepts',
      items: [
        'learn/what-is-forge',
        'learn/architecture',
        'learn/request-lifecycle',
        'learn/code-generation',
      ],
    },
    {
      type: 'category',
      label: 'Quick Start',
      items: [
        'getting-started/installation',
        'getting-started/hello-world',
        'getting-started/first-api',
        'getting-started/first-app',
        'getting-started/project-structure',
        'getting-started/quickstart',
      ],
    },
    {
      type: 'category',
      label: 'Core Logic',
      items: [
        'learn/extending-forge',
      ],
    },
  ],

  // Full Guides sidebar - For practitioners
  fullGuides: [
    {
      type: 'category',
      label: 'Tutorials',
      items: [
        'tutorials/overview',
        'tutorials/getting-started',
        'tutorials/admin-interface',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/models',
        'guides/queries',
        'guides/admin',
        'guides/rest-api',
        'guides/migrations',
        'guides/security',
        'guides/deployment',
        'guides/common-patterns',
        'guides/best-practices',
      ],
    },
    {
      type: 'category',
      label: 'System Features',
      items: [
        'features/admin-system',
        'features/api-framework',
        'features/code-generation',
        'features/database-layer',
        'features/filter-system',
        'features/identity-system',
        'features/migration-system',
        'features/orm-system',
        'features/schema-system',
      ],
    },
    {
      type: 'category',
      label: 'API Reference',
      items: [
        'api-reference/schema',
        'api-reference/queryset',
        'api-reference/manager',
        'api-reference/fields',
        'api-reference/relations',
        'api-reference/hooks',
      ],
    },
    {
      type: 'category',
      label: 'Examples',
      items: [
        'examples/blog',
        'examples/ecommerce',
        'examples/ecommerce-website',
        'examples/library',
      ],
    },
    {
      type: 'category',
      label: 'Advanced',
      items: [
        'advanced/performance',
        'advanced/plugins',
        'advanced/custom-fields',
        'advanced/code-generation',
      ],
    },
    {
      type: 'category',
      label: 'Deep Dives',
      items: [
        'deep-dives/architecture',
        'deep-dives/design-principles',
        'deep-dives/features-overview',
        'deep-dives/schema-system',
      ],
    },
    {
      type: 'category',
      label: 'Contributing',
      items: [
        'contributing/architecture',
        'contributing/development',
        'contributing/implementations',
        'contributing/changelog',
        'contributing/roadmap',
        'contributing/todos',
      ],
    },
  ],
};

module.exports = sidebars;
