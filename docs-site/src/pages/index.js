import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import SEOHead from '@site/src/components/SEOHead';
import styles from './index.module.css';

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
        description="Django-like Go framework with type safety">
        
        {/* Hero Section */}
        <div className={styles.hero}>
          <div className={styles.heroContainer}>
            <div className={styles.heroContent}>
              <div className={styles.heroBadge}>
                <span className={styles.badgeText}>⚡ Production Ready</span>
              </div>
              <h1 className={styles.heroTitle}>
                Django for Go,<br />Type-Safe from the Start
              </h1>
              <p className={styles.heroSubtitle}>
                Build web applications in Go with Django's developer experience. 
                Type-safe ORM, auto-generated admin, REST API framework, and code generation—all with Go's performance.
              </p>
              <div className={styles.heroButtons}>
                <Link
                  className={styles.buttonPrimary}
                  to="/docs/getting-started/installation">
                  Get Started →
                </Link>
                <Link
                  className={styles.buttonSecondary}
                  to="/docs/introduction">
                  Learn More
                </Link>
                <Link
                  className={styles.buttonTertiary}
                  href="https://github.com/hamidrabedi/foreit"
                  target="_blank">
                  View on GitHub
                </Link>
              </div>
              <div className={styles.heroStats}>
                <div className={styles.statItem}>
                  <div className={styles.statValue}>15+</div>
                  <div className={styles.statLabel}>Core Features</div>
                </div>
                <div className={styles.statItem}>
                  <div className={styles.statValue}>100%</div>
                  <div className={styles.statLabel}>Type Safe</div>
                </div>
                <div className={styles.statItem}>
                  <div className={styles.statValue}>MVP</div>
                  <div className={styles.statLabel}>Complete</div>
                </div>
              </div>
            </div>
            <div className={styles.heroCode}>
              <div className={styles.codeBlock}>
                <div className={styles.codeHeader}>
                  <span className={styles.codeDot}></span>
                  <span className={styles.codeDot}></span>
                  <span className={styles.codeDot}></span>
                  <span className={styles.codeTitle}>models/post.go</span>
                </div>
                <pre className={styles.codeContent}>
{`type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").
            Primary().AutoIncrement().Build(),
        schema.String("title").
            Required().MaxLength(200).Build(),
        schema.Text("content").Required().Build(),
        schema.Bool("published").
            Default(false).Build(),
        schema.Time("created_at").
            AutoNowAdd().Build(),
    }
}`}
                </pre>
              </div>
              <div className={styles.codeBlock} style={{marginTop: '1rem'}}>
                <div className={styles.codeHeader}>
                  <span className={styles.codeDot}></span>
                  <span className={styles.codeDot}></span>
                  <span className={styles.codeDot}></span>
                  <span className={styles.codeTitle}>query.go</span>
                </div>
                <pre className={styles.codeContent}>
{`posts, err := Post.Objects.
    Filter(Post.Fields.Published.Equals(true)).
    OrderBy("-created_at").
    Limit(10).
    All(ctx)`}
                </pre>
              </div>
            </div>
          </div>
        </div>

        {/* Features Section */}
        <div className={styles.features}>
          <div className={styles.container}>
            <div className={styles.sectionHeader}>
              <h2 className={styles.sectionTitle}>Everything You Need</h2>
              <p className={styles.sectionSubtitle}>
                15 complete features built with type safety, performance, and developer experience in mind
              </p>
            </div>
            <div className={styles.featuresGrid}>
              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>📋</div>
                <h3 className={styles.featureTitle}>Schema System</h3>
                <p className={styles.featureDescription}>
                  Declarative model definitions with full Django field options, relationships, metadata, and lifecycle hooks.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>
              
              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>⚙️</div>
                <h3 className={styles.featureTitle}>Code Generation</h3>
                <p className={styles.featureDescription}>
                  AST-based code generation for models, managers, querysets, and field expressions. Generate type-safe code automatically.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>🔒</div>
                <h3 className={styles.featureTitle}>Type-Safe ORM</h3>
                <p className={styles.featureDescription}>
                  Complete Django-like ORM with QuerySet API, Manager CRUD operations, field expressions, and SQL builder.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>🎛️</div>
                <h3 className={styles.featureTitle}>Admin System</h3>
                <p className={styles.featureDescription}>
                  Django-style admin interface with type-safe configuration, CRUD operations, filters, widgets, and export.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>🌐</div>
                <h3 className={styles.featureTitle}>REST API Framework</h3>
                <p className={styles.featureDescription}>
                  DRF-like REST API framework with serializers, viewsets, authentication, permissions, throttling, and more.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>🔍</div>
                <h3 className={styles.featureTitle}>Filter System</h3>
                <p className={styles.featureDescription}>
                  Advanced filtering system with AST support, query parsing, expression conversion, and security validation.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>👤</div>
                <h3 className={styles.featureTitle}>Identity System</h3>
                <p className={styles.featureDescription}>
                  Complete user management and authentication system with repositories, services, backends, and permissions.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>💾</div>
                <h3 className={styles.featureTitle}>Database Layer</h3>
                <p className={styles.featureDescription}>
                  Database connection management, transactions, and migration integration with connection pooling.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>🔄</div>
                <h3 className={styles.featureTitle}>Migration System</h3>
                <p className={styles.featureDescription}>
                  Database schema migrations with detection, diff generation, execution, and rollback support.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>🚀</div>
                <h3 className={styles.featureTitle}>HTTP & Server</h3>
                <p className={styles.featureDescription}>
                  HTTP server with routing, middleware stack, security, static files, and health checks.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>📊</div>
                <h3 className={styles.featureTitle}>Logging System</h3>
                <p className={styles.featureDescription}>
                  Structured logging with multiple outputs, formats, levels, and exporters for production-ready logging.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>⚙️</div>
                <h3 className={styles.featureTitle}>Configuration</h3>
                <p className={styles.featureDescription}>
                  Application configuration with YAML, JSON, and environment variable support with hierarchical overrides.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>✓</div>
                <h3 className={styles.featureTitle}>Validation</h3>
                <p className={styles.featureDescription}>
                  Data validation with go-playground/validator integration and schema support for robust input validation.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>🛡️</div>
                <h3 className={styles.featureTitle}>Security</h3>
                <p className={styles.featureDescription}>
                  Security features including CSRF protection, XSS prevention, and SQL injection prevention built-in.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>

              <div className={styles.featureCard}>
                <div className={styles.featureIcon}>🖥️</div>
                <h3 className={styles.featureTitle}>CLI Tools</h3>
                <p className={styles.featureDescription}>
                  Command-line interface for project creation, code generation, migrations, and server management.
                </p>
                <div className={styles.featureStatus}>✅ Complete</div>
              </div>
            </div>
          </div>
        </div>

        {/* Design Principles Section */}
        <div className={styles.principles}>
          <div className={styles.container}>
            <div className={styles.sectionHeader}>
              <h2 className={styles.sectionTitle}>Built with Purpose</h2>
              <p className={styles.sectionSubtitle}>
                Core design principles that guide forge's development
              </p>
            </div>
            <div className={styles.principlesGrid}>
              <div className={styles.principleCard}>
                <div className={styles.principleIcon}>🔒</div>
                <h3 className={styles.principleTitle}>Type-Safe First</h3>
                <p className={styles.principleDescription}>
                  Primary API uses Go generics for compile-time type checking. All queries, models, and operations are type-safe.
                </p>
              </div>
              <div className={styles.principleCard}>
                <div className={styles.principleIcon}>⚡</div>
                <h3 className={styles.principleTitle}>Convention over Configuration</h3>
                <p className={styles.principleDescription}>
                  Sensible defaults everywhere. Django-like patterns and naming. Minimal boilerplate with auto-generated code.
                </p>
              </div>
              <div className={styles.principleCard}>
                <div className={styles.principleIcon}>🔌</div>
                <h3 className={styles.principleTitle}>Fully Extensible</h3>
                <p className={styles.principleDescription}>
                  Everything can be extended or overridden. Plugin architecture, hook system, and custom validators, filters, widgets.
                </p>
              </div>
              <div className={styles.principleCard}>
                <div className={styles.principleIcon}>🛡️</div>
                <h3 className={styles.principleTitle}>Security by Default</h3>
                <p className={styles.principleDescription}>
                  Built-in CSRF protection, XSS prevention, SQL injection prevention via parameter binding. Security features enabled by default.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Quick Start Section */}
        <div className={styles.quickStart}>
          <div className={styles.container}>
            <div className={styles.quickStartContent}>
              <h2 className={styles.sectionTitle}>Get Started in 5 Minutes</h2>
              <p className={styles.sectionSubtitle}>
                Create your first forge application and see it in action
              </p>
              <div className={styles.quickStartSteps}>
                <div className={styles.step}>
                  <div className={styles.stepNumber}>1</div>
                  <div className={styles.stepContent}>
                    <h3>Create Project</h3>
                    <code className={styles.stepCode}>forge new myapp</code>
                    <p className={styles.stepDescription}>Initialize a new forge project</p>
                  </div>
                </div>
                <div className={styles.step}>
                  <div className={styles.stepNumber}>2</div>
                  <div className={styles.stepContent}>
                    <h3>Define Models</h3>
                    <code className={styles.stepCode}>Define your schema</code>
                    <p className={styles.stepDescription}>Declarative model definitions</p>
                  </div>
                </div>
                <div className={styles.step}>
                  <div className={styles.stepNumber}>3</div>
                  <div className={styles.stepContent}>
                    <h3>Generate Code</h3>
                    <code className={styles.stepCode}>forge generate</code>
                    <p className={styles.stepDescription}>Generate type-safe code</p>
                  </div>
                </div>
                <div className={styles.step}>
                  <div className={styles.stepNumber}>4</div>
                  <div className={styles.stepContent}>
                    <h3>Run Server</h3>
                    <code className={styles.stepCode}>forge runserver</code>
                    <p className={styles.stepDescription}>Start your application</p>
                  </div>
                </div>
              </div>
              <div className={styles.quickStartButton}>
                <Link
                  className={styles.buttonPrimary}
                  to="/docs/getting-started/installation">
                  Start Building →
                </Link>
              </div>
            </div>
          </div>
        </div>

        {/* Architecture Highlights */}
        <div className={styles.architecture}>
          <div className={styles.container}>
            <div className={styles.architectureGrid}>
              <div className={styles.architectureContent}>
                <h2 className={styles.sectionTitle}>Built for Scale</h2>
                <p className={styles.sectionDescription}>
                  forge follows a layered architecture designed for maintainability and performance.
                </p>
                <div className={styles.architectureLayers}>
                  <div className={styles.layer}>
                    <div className={styles.layerNumber}>1</div>
                    <div className={styles.layerContent}>
                      <h4>Application Layer</h4>
                      <p>User models, views, controllers, routes</p>
                    </div>
                  </div>
                  <div className={styles.layer}>
                    <div className={styles.layerNumber}>2</div>
                    <div className={styles.layerContent}>
                      <h4>Code Generation Layer</h4>
                      <p>AST Parser → Schema Analysis → Code Generator</p>
                    </div>
                  </div>
                  <div className={styles.layer}>
                    <div className={styles.layerNumber}>3</div>
                    <div className={styles.layerContent}>
                      <h4>Framework API Layer</h4>
                      <p>Admin, API, ORM, Identity, Filter</p>
                    </div>
                  </div>
                  <div className={styles.layer}>
                    <div className={styles.layerNumber}>4</div>
                    <div className={styles.layerContent}>
                      <h4>Database Layer</h4>
                      <p>SQL Builder, Query Execution, Transactions, Migrations</p>
                    </div>
                  </div>
                  <div className={styles.layer}>
                    <div className={styles.layerNumber}>5</div>
                    <div className={styles.layerContent}>
                      <h4>Infrastructure Layer</h4>
                      <p>HTTP Server, Security, Logging, Config, Server</p>
                    </div>
                  </div>
                </div>
                <div className={styles.architectureButton}>
                  <Link
                    className={styles.buttonSecondary}
                    to="/docs/learn/architecture">
                    Learn About Architecture →
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* CTA Section */}
        <div className={styles.cta}>
          <div className={styles.container}>
            <div className={styles.ctaContent}>
              <h2 className={styles.ctaTitle}>Ready to Build?</h2>
              <p className={styles.ctaSubtitle}>
                Join the community and start building amazing web applications with forge
              </p>
              <div className={styles.ctaButtons}>
                <Link
                  className={styles.buttonPrimary}
                  to="/docs/getting-started/installation">
                  Get Started
                </Link>
                <Link
                  className={styles.buttonSecondary}
                  href="https://github.com/hamidrabedi/foreit"
                  target="_blank">
                  View on GitHub
                </Link>
              </div>
            </div>
          </div>
        </div>
      </Layout>
    </>
  );
}
