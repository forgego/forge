/**
 * Sidebar configuration for forge documentation
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docs: [
    {
      type: 'category',
      label: 'Start Here',
      items: ['index', 'introduction', 'quickstart', 'installation'],
    },
    {
      type: 'category',
      label: 'Core Build',
      items: ['models', 'orm', 'migrations'],
    },
    {
      type: 'category',
      label: 'Admin',
      items: ['admin/overview', 'admin/config', 'admin/actions', 'admin/filters', 'admin/ui'],
    },
    {
      type: 'category',
      label: 'API',
      items: [
        'api/overview',
        'api/serializers',
        'api/viewsets',
        'api/authentication',
        'api/permissions',
        'api/throttling',
        'api/pagination',
        'api/versioning',
        'api/renderers-parsers',
        'api/errors',
        'api/openapi',
      ],
    },
    {
      type: 'category',
      label: 'Platform',
      items: ['filters', 'identity', 'server/overview', 'server/middleware', 'server/security', 'server/health'],
    },
    {
      type: 'category',
      label: 'Config',
      items: ['config/overview', 'config/app', 'config/server', 'config/database', 'config/security', 'config/logging', 'config/errors'],
    },
    {
      type: 'category',
      label: 'Reference',
      items: ['features', 'api-reference', 'status'],
    },
    {
      type: 'category',
      label: 'Project',
      items: ['changelog', 'community', 'security'],
    },
  ],
};

module.exports = sidebars;
