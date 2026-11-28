/**
 * Sidebar configuration for forge documentation
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docs: [
    {
      type: 'category',
      label: 'Getting Started',
      items: ['index', 'introduction', 'quickstart', 'installation', 'status', 'upgrading'],
    },
    {
      type: 'category',
      label: 'Core Guides',
      items: ['models', 'codegen', 'orm', 'migrations', 'filters', 'testing'],
    },
    {
      type: 'category',
      label: 'Admin',
      items: ['admin/overview', 'admin/config', 'admin/actions', 'admin/filters', 'admin/ui', 'admin/extensions'],
    },
    {
      type: 'category',
      label: 'REST API',
      items: [
        'api/overview',
        'api/serializers',
        'api/viewsets',
        'api/data-access',
        'api/field-contract',
        'api/authentication',
        'api/permissions',
        'api/throttling',
        'api/pagination',
        'api/versioning',
        'api/renderers-parsers',
        'api/errors',
        'validation-errors',
        'api/openapi',
      ],
    },
    {
      type: 'category',
      label: 'Server, Identity & Ops',
      items: ['identity', 'operations', 'server/overview', 'server/middleware', 'server/security', 'server/health', 'deployment'],
    },
    {
      type: 'category',
      label: 'Configuration',
      items: ['config/overview', 'config/app', 'config/server', 'config/database', 'config/security', 'config/logging', 'config/errors'],
    },
    {
      type: 'category',
      label: 'Reference',
      items: [
        'cli',
        'features',
        'api-reference',
        'api-reference/schema',
        'api-reference/fields',
        'api-reference/relations',
        'api-reference/queryset',
        'api-reference/manager',
        'api-reference/hooks',
      ],
    },
    {
      type: 'category',
      label: 'AI & Project',
      items: [
        'ai-skills',
        'changelog',
        'community',
        'contributing/development',
        'contributing/versioning-and-support',
        'contributing/extension-contracts',
        'security',
      ],
    },
  ],
};

module.exports = sidebars;
