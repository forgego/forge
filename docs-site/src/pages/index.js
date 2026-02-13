import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import SEOHead from '@site/src/components/SEOHead';
import styles from './index.module.css';

const features = [
  {
    title: 'Type-Safe ORM',
    icon: '🔒',
    description: 'Write queries with full Go type safety. No more runtime surprises or string-based queries.',
  },
  {
    title: 'Auto Admin',
    icon: '⚡',
    description: 'Get a fully functional admin panel out of the box. Manage your data without writing boilerplate.',
  },
  {
    title: 'REST APIs',
    icon: '🚀',
    description: 'Build production-ready REST APIs with serializers, authentication, and pagination included.',
  },
  {
    title: 'Migrations',
    icon: '🔄',
    description: 'Track schema changes automatically. Generate and apply migrations like Django or Rails.',
  },
  {
    title: 'Code Generation',
    icon: '✨',
    description: 'Define schemas once, generate type-safe models, queries, and admin interfaces automatically.',
  },
  {
    title: 'Extensible',
    icon: '🔌',
    description: 'Plugin system and hooks let you customize everything. Works with your existing Go tools.',
  },
];

const codeExample = `package models

import "github.com/forgego/forge/schema"

type Article struct {
    schema.BaseSchema
}

func (Article) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary()),
        schema.StringField("title", schema.MaxLength(200)),
        schema.TextField("content"),
        schema.TimeField("published_at", schema.AutoNow()),
    }
}`;

const queryExample = `// Type-safe queries
articles := models.ArticleManager.
    Filter(models.ArticleExpr.Title.Contains("Go")).
    OrderBy("-published_at").
    Limit(10)

// Relations work seamlessly
article := models.ArticleManager.
    SelectRelated("author").
    Get(ctx, 1)`;

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  
  return (
    <>
      <SEOHead
        title="Forge - Go Web Framework"
        description="Build web apps in Go with Django-inspired productivity. Type-safe ORM, auto-generated admin, REST APIs, and migrations."
        keywords={[
          'go framework',
          'golang web framework',
          'go orm',
          'type-safe go',
          'go rest api',
          'forge framework',
        ]}
        url="/"
      />
      <Layout
        title="Go Web Framework"
        description="Django-inspired productivity for Go">
        
        {/* Hero Section */}
        <header className={styles.hero}>
          <div className={styles.heroContent}>
            <div className={styles.badge}>v1.0.0 • MIT License</div>
            <h1 className={styles.heroTitle}>
              Build web apps fast<br/>with Go's type safety
            </h1>
            <p className={styles.heroSubtitle}>
              Forge brings Django's rapid development to Go. Get a type-safe ORM, 
              auto-generated admin panel, REST API framework, and migrations—all in one toolkit.
            </p>
            <div className={styles.heroActions}>
              <Link className={styles.btnPrimary} to="/docs/quickstart">
                Get Started →
              </Link>
              <Link className={styles.btnSecondary} href="https://github.com/hamidrabedi/foreit" target="_blank">
                View on GitHub
              </Link>
            </div>
            <div className={styles.heroCommand}>
              <code>go install github.com/forgego/forge/cli/cmd@latest</code>
            </div>
          </div>
        </header>

        <main>
          {/* Features Grid */}
          <section className={styles.features}>
            <div className={styles.container}>
              <div className={styles.sectionHeader}>
                <h2>Everything you need to ship</h2>
                <p>From database to API, all the tools you need are included</p>
              </div>
              <div className={styles.featuresGrid}>
                {features.map((feature) => (
                  <div key={feature.title} className={styles.featureCard}>
                    <div className={styles.featureIcon}>{feature.icon}</div>
                    <h3>{feature.title}</h3>
                    <p>{feature.description}</p>
                  </div>
                ))}
              </div>
            </div>
          </section>

          {/* Code Examples */}
          <section className={styles.codeSection}>
            <div className={styles.container}>
              <div className={styles.sectionHeader}>
                <h2>Write less, build more</h2>
                <p>Define your models once, get everything else generated</p>
              </div>
              <div className={styles.codeExamples}>
                <div className={styles.codeBlock}>
                  <div className={styles.codeHeader}>
                    <span className={styles.codeLabel}>Define Schema</span>
                    <span className={styles.codeLang}>Go</span>
                  </div>
                  <pre><code>{codeExample}</code></pre>
                </div>
                <div className={styles.codeBlock}>
                  <div className={styles.codeHeader}>
                    <span className={styles.codeLabel}>Query Data</span>
                    <span className={styles.codeLang}>Go</span>
                  </div>
                  <pre><code>{queryExample}</code></pre>
                </div>
              </div>
            </div>
          </section>

          {/* Quick Start */}
          <section className={styles.quickStart}>
            <div className={styles.container}>
              <div className={styles.quickStartContent}>
                <div className={styles.quickStartText}>
                  <h2>Ready in 3 commands</h2>
                  <p>Create a new project and start building in under a minute</p>
                </div>
                <div className={styles.quickStartSteps}>
                  <div className={styles.step}>
                    <span className={styles.stepNumber}>1</span>
                    <code>forge new myapp</code>
                  </div>
                  <div className={styles.step}>
                    <span className={styles.stepNumber}>2</span>
                    <code>forge migrate</code>
                  </div>
                  <div className={styles.step}>
                    <span className={styles.stepNumber}>3</span>
                    <code>forge runserver</code>
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* Why Forge */}
          <section className={styles.why}>
            <div className={styles.container}>
              <div className={styles.sectionHeader}>
                <h2>Why teams choose Forge</h2>
              </div>
              <div className={styles.whyGrid}>
                <div className={styles.whyCard}>
                  <h3>Familiar patterns</h3>
                  <p>If you know Django or Rails, you already know Forge. Same concepts, Go's performance.</p>
                </div>
                <div className={styles.whyCard}>
                  <h3>Type safety</h3>
                  <p>Catch errors at compile time, not production. Full IDE autocomplete for queries and fields.</p>
                </div>
                <div className={styles.whyCard}>
                  <h3>Less boilerplate</h3>
                  <p>Code generation handles the repetitive work. Focus on your business logic.</p>
                </div>
                <div className={styles.whyCard}>
                  <h3>Production ready</h3>
                  <p>Built-in security, authentication, validation, and logging. Deploy with confidence.</p>
                </div>
              </div>
            </div>
          </section>

          {/* CTA Section */}
          <section className={styles.cta}>
            <div className={styles.container}>
              <div className={styles.ctaContent}>
                <h2>Start building today</h2>
                <p>Join developers shipping production apps with Forge</p>
                <div className={styles.ctaActions}>
                  <Link className={styles.btnPrimary} to="/docs/quickstart">
                    Read the Docs
                  </Link>
                  <Link className={styles.btnOutline} to="/docs/installation">
                    Installation Guide
                  </Link>
                </div>
              </div>
            </div>
          </section>
        </main>
      </Layout>
    </>
  );
}
