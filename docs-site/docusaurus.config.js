// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer').themes.github;
const darkCodeTheme = require('prism-react-renderer').themes.dracula;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'forge - Django-like Go Framework',
  tagline: 'Django-like Go framework with type safety. Build web applications in Go with Django\'s developer experience and Go\'s performance.',
  favicon: 'favicon.svg',

  // Set the production url of your site here
  url: 'https://hamidrabedi.github.io',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployments, it is often '/<projectName>/'
  baseUrl: '/foreit/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'hamidrabedi', // Usually your GitHub org/user name.
  projectName: 'foreit', // Usually your repo name.

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },
  onBrokenLinks: 'warn',

  // SEO configuration
  trailingSlash: true,
  
  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is in Chinese, you may
  // want to replace "en" with "zh-Hans".
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
          showLastUpdateAuthor: false, // Disable for faster builds
          showLastUpdateTime: false, // Disable for faster builds
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
        sitemap: {
          changefreq: 'weekly',
          priority: 0.5,
          ignorePatterns: ['/tags/**'],
          filename: 'sitemap.xml',
        },
        gtag: undefined, // Disable analytics for faster builds
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // SEO: Replace with your project's social card
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
            sidebarId: 'gettingStarted',
            position: 'left',
            label: 'Getting Started',
          },
          {
            type: 'docSidebar',
            sidebarId: 'fullGuides',
            position: 'left',
            label: 'Full Guides',
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
            title: 'Getting Started',
            items: [
              {
                label: 'Introduction',
                to: '/docs/introduction',
              },
              {
                label: 'Installation',
                to: '/docs/getting-started/installation',
              },
              {
                label: 'Quick Start',
                to: '/docs/getting-started/quickstart',
              },
            ],
          },
          {
            title: 'Full Guides',
            items: [
              {
                label: 'Guides',
                to: '/docs/guides/models',
              },
              {
                label: 'API Reference',
                to: '/docs/api-reference/schema',
              },
              {
                label: 'Examples',
                to: '/docs/examples/blog',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/hamidrabedi/foreit',
              },
              {
                label: 'Issues',
                href: 'https://github.com/hamidrabedi/foreit/issues',
              },
              {
                label: 'Discussions',
                href: 'https://github.com/hamidrabedi/foreit/discussions',
              },
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
        content: '🎉 forge v1.0.0 is now available!',
        backgroundColor: '#667eea',
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

  plugins: [
    // Disabled ideal-image plugin to avoid sharp native dependency issues
    // Uncomment if you need image optimization (requires sharp to be built)
    // [
    //   '@docusaurus/plugin-ideal-image',
    //   {
    //     quality: 70,
    //     max: 1030,
    //     min: 640,
    //     steps: 2,
    //     disableInDev: true,
    //   },
    // ],
  ],

  // Build performance optimizations
  staticDirectories: ['static'],
};

module.exports = config;
