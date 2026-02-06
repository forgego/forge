import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import SEOHead from '@site/src/components/SEOHead';
import styles from './index.module.css';

const quickStart = [
  'go install github.com/forgego/forge/cli/cmd@latest',
  'forge new myapp',
  'forge runserver',
];

const coreLinks = [
  {
    title: 'Quick Start',
    description: 'Get a working app in minutes.',
    to: '/docs/quickstart',
  },
  {
    title: 'Installation',
    description: 'Prerequisites and setup.',
    to: '/docs/installation',
  },
  {
    title: 'Models',
    description: 'Define schemas and query data.',
    to: '/docs/models',
  },
  {
    title: 'Admin UI',
    description: 'Configure the auto-generated admin.',
    to: '/docs/admin',
  },
  {
    title: 'REST API',
    description: 'Build APIs with serializers and auth.',
    to: '/docs/rest-api',
  },
  {
    title: 'Features',
    description: 'Full feature list by area.',
    to: '/docs/features',
  },
];

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <>
      <SEOHead
        title="forge - Django-like Go Framework"
        description="Build web applications in Go with Django's developer experience. Type-safe ORM, auto-generated admin, REST API framework, and code generation."
        keywords={[
          'go framework',
          'golang framework',
          'django go',
          'type-safe orm',
          'go web framework',
          'forge framework',
        ]}
        url="/"
      />
      <Layout
        title={`${siteConfig.title}`}
        description="Django-inspired productivity for Go with type safety.">
        <header className={styles.hero}>
          <div className={styles.container}>
            <p className={styles.eyebrow}>forge v1.0.0</p>
            <h1 className={styles.title}>Django-inspired productivity for Go</h1>
            <p className={styles.subtitle}>
              forge gives Go teams a type-safe ORM, admin UI, REST API layer, migrations, and code generation
              in a single workflow.
            </p>
            <div className={styles.actions}>
              <Link className={styles.primaryButton} to="/docs/quickstart">
                Get Started
              </Link>
              <Link className={styles.secondaryButton} href="https://github.com/hamidrabedi/foreit" target="_blank">
                GitHub
              </Link>
            </div>
          </div>
        </header>

        <main>
          <section className={styles.section}>
            <div className={styles.container}>
              <h2>Quick start</h2>
              <p className={styles.sectionIntro}>Three commands to a running app.</p>
              <div className={styles.codeBlock}>
                {quickStart.map((line) => (
                  <div key={line} className={styles.codeLine}>
                    <span className={styles.codePrompt}>$</span> {line}
                  </div>
                ))}
              </div>
            </div>
          </section>

          <section className={styles.sectionAlt}>
            <div className={styles.container}>
              <h2>Core documentation</h2>
              <p className={styles.sectionIntro}>Start here, then follow the build docs.</p>
              <div className={styles.cardGrid}>
                {coreLinks.map((item) => (
                  <Link key={item.title} className={styles.card} to={item.to}>
                    <h3>{item.title}</h3>
                    <p>{item.description}</p>
                  </Link>
                ))}
              </div>
            </div>
          </section>

          <section className={styles.section}>
            <div className={styles.container}>
              <h2>What you get</h2>
              <ul className={styles.list}>
                <li>Type-safe schema and ORM with generated field expressions.</li>
                <li>Admin UI with lists, filters, and CRUD workflows.</li>
                <li>REST API layer with serializers and authentication.</li>
                <li>CLI-driven migrations and code generation.</li>
              </ul>
            </div>
          </section>
        </main>
      </Layout>
    </>
  );
}
