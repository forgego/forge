// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer').themes.github;
const darkCodeTheme = require('prism-react-renderer').themes.dracula;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Forge',
  tagline: 'The batteries-included web framework for Go.',
  favicon: 'favicon.svg',

  url: 'https://forgego.github.io',
  baseUrl: '/forge/',

  organizationName: 'forgego',
  projectName: 'forge',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'throw',
    },
  },
  onBrokenLinks: 'throw',

  trailingSlash: true,

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/forgego/forge/tree/master/docs-site/',
          showLastUpdateAuthor: false,
          showLastUpdateTime: false,
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
        gtag: undefined,
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      image: 'social-card.png',
      metadata: [
        {name: 'keywords', content: 'go, golang, web framework, orm, migrations, admin panel, rest api, code generation, postgresql, sqlite'},
        {property: 'og:type', content: 'website'},
        {property: 'og:site_name', content: 'Forge'},
        {name: 'twitter:card', content: 'summary_large_image'},
      ],
      navbar: {
        title: 'Forge',
        logo: {
          alt: 'Forge',
          src: 'logo.svg',
          srcDark: 'logo-dark.svg',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docs',
            position: 'left',
            label: 'Docs',
          },
          {
            to: '/docs/features',
            label: 'Features',
            position: 'left',
          },
          {
            to: '/docs/models',
            label: 'Models & ORM',
            position: 'left',
          },
          {
            to: '/docs/admin/overview',
            label: 'Admin',
            position: 'left',
          },
          {
            to: '/docs/api/overview',
            label: 'REST API',
            position: 'left',
          },
          {
            to: '/docs/changelog',
            label: 'Changelog',
            position: 'left',
          },
          {
            href: 'https://github.com/forgego/forge',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'light',
        links: [
          {
            title: 'Start Here',
            items: [
              {label: 'Introduction', to: '/docs/introduction'},
              {label: 'Quick Start', to: '/docs/quickstart'},
              {label: 'Installation', to: '/docs/installation'},
              {label: 'Features Matrix', to: '/docs/features'},
            ],
          },
          {
            title: 'Build',
            items: [
              {label: 'Models & Schema DSL', to: '/docs/models'},
              {label: 'ORM & QuerySet', to: '/docs/orm'},
              {label: 'Migrations & Recovery', to: '/docs/migrations'},
              {label: 'Admin Console (SPA)', to: '/docs/admin/overview'},
              {label: 'REST API & OpenAPI', to: '/docs/api/overview'},
            ],
          },
          {
            title: 'Project & Community',
            items: [
              {label: 'Changelog', to: '/docs/changelog'},
              {label: 'Community', to: '/docs/community'},
              {label: 'Security', to: '/docs/security'},
              {label: 'GitHub Repository', href: 'https://github.com/forgego/forge'},
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} The Forge authors. MIT License.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ['go', 'bash', 'yaml', 'sql'],
      },
      colorMode: {
        defaultMode: 'light',
        disableSwitch: false,
        respectPrefersColorScheme: true,
      },
    }),

  plugins: [],

  staticDirectories: ['static'],
};

module.exports = config;
