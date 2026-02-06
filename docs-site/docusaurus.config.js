// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer').themes.github;
const darkCodeTheme = require('prism-react-renderer').themes.dracula;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'forge - Django-like Go Framework',
  tagline: 'Django-inspired productivity for Go with type safety and performance.',
  favicon: 'favicon.svg',

  url: 'https://hamidrabedi.github.io',
  baseUrl: '/foreit/',

  organizationName: 'hamidrabedi',
  projectName: 'foreit',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },
  onBrokenLinks: 'warn',

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
          editUrl: 'https://github.com/hamidrabedi/foreit/tree/main/docs-site/',
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
      image: 'forge-social-card.svg',
      metadata: [
        {name: 'keywords', content: 'go, golang, framework, django, orm, type-safe, web framework, rest api, code generation, postgresql'},
        {name: 'author', content: 'forge Framework'},
        {property: 'og:type', content: 'website'},
        {property: 'og:site_name', content: 'forge Framework'},
        {name: 'twitter:card', content: 'summary_large_image'},
        {name: 'twitter:site', content: '@forgego'},
      ],
      navbar: {
        title: 'forge',
        logo: {
          alt: 'forge Logo',
          src: 'logo.svg',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docs',
            position: 'left',
            label: 'Docs',
          },
          {
            to: '/docs/changelog',
            label: 'Changelog',
            position: 'left',
          },
          {
            href: 'https://github.com/hamidrabedi/foreit',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Start Here',
            items: [
              {label: 'Introduction', to: '/docs/introduction'},
              {label: 'Quick Start', to: '/docs/quickstart'},
              {label: 'Installation', to: '/docs/installation'},
            ],
          },
          {
            title: 'Build',
            items: [
              {label: 'Models', to: '/docs/models'},
              {label: 'Admin UI', to: '/docs/admin'},
              {label: 'REST API', to: '/docs/rest-api'},
            ],
          },
          {
            title: 'Project',
            items: [
              {label: 'Changelog', to: '/docs/changelog'},
              {label: 'Community', to: '/docs/community'},
              {label: 'Security', to: '/docs/security'},
              {label: 'GitHub', href: 'https://github.com/hamidrabedi/foreit'},
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} forge Framework. Built with Docusaurus.`,
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
      announcementBar: {
        id: 'announcement-bar',
        content: 'forge v1.0.0 is now available.',
        backgroundColor: '#334155',
        textColor: '#ffffff',
        isCloseable: true,
      },
      algolia: {
        appId: 'YOUR_APP_ID',
        apiKey: 'YOUR_SEARCH_API_KEY',
        indexName: 'forge',
        contextualSearch: true,
      },
    }),

  plugins: [],

  staticDirectories: ['static'],
};

module.exports = config;
