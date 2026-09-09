import React, { useState } from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import SEOHead from '@site/src/components/SEOHead';
import styles from './index.module.css';

const CODE_TABS = [
  {
    id: 'schema',
    label: '1. Declarative Schema DSL',
    tagline: 'Define models, constraints, generated columns, and relations with zero boilerplate.',
    lang: 'Go',
    code: `package models

import "github.com/forgego/forge/schema"

type Product struct {
    schema.BaseModel
}

func (Product) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement(),
        schema.String("sku").MaxLength(64).Unique().DBIndex(),
        schema.String("name").MaxLength(255).Required(),
        schema.Decimal("price", 10, 2).Required(),
        schema.Float64("tax_rate").DBDefault("0.10"),
        schema.GeneratedColumn("price_with_tax", "price * (1 + tax_rate)", true),
        schema.Bool("in_stock").DBDefault("true"),
        schema.ForeignKey("category_id", "Category", schema.CascadeSET_NULL),
    }
}`,
  },
  {
    id: 'orm',
    label: '2. Type-Safe QuerySet',
    tagline: 'Complex queries with compile-time safety, boolean Q-expressions, and aggregations.',
    lang: 'Go',
    code: `// Expressive, type-safe queries with zero runtime reflection
products, err := ProductManager.
    Filter(
        orm.And(
            ProductExpr.InStock.Eq(true),
            ProductExpr.PriceWithTax.Lte(150.00),
            ProductExpr.Category.Slug.In("laptops", "accessories"),
        ),
    ).
    SelectRelated("Category").
    OrderBy("-PriceWithTax").
    Limit(20).
    All(ctx)

// Rich aggregations without raw SQL
stats, err := ProductManager.
    Filter(ProductExpr.InStock.Eq(true)).
    Aggregate(ctx,
        orm.Count("id", "total_products"),
        orm.Avg("price", "average_price"),
        orm.Max("price", "highest_price"),
    )`,
  },
  {
    id: 'admin',
    label: '3. Instant Admin Console',
    tagline: 'Production React 19 SPA with TanStack Router, custom actions, history, and widgets.',
    lang: 'Go',
    code: `// Register models with search, filters, saved views, and bulk actions
admin.Register[Product](site, admin.ModelConfig[Product]{
    ListDisplay:  []string{"SKU", "Name", "Category", "PriceWithTax", "InStock"},
    SearchFields: []string{"SKU", "Name"},
    ListFilter:   []string{"InStock", "Category"},
    Actions: []admin.Action{
        {
            ID:    "restock",
            Label: "Restock Inventory (+50)",
            Handler: func(ctx context.Context, ids []any) error {
                return InventoryService.Restock(ctx, ids, 50)
            },
        },
    },
})`,
  },
  {
    id: 'api',
    label: '4. REST API & OpenAPI',
    tagline: 'ViewSets, ModelSerializers, pagination, rate throttling, and OpenAPI 3.0 generation.',
    lang: 'Go',
    code: `// ViewSets provide full CRUD, filtering, pagination, and OpenAPI spec
type ProductViewSet struct {
    api.ModelViewSet[Product]
}

func RegisterAPI(router chi.Router) {
    viewset := &ProductViewSet{
        Serializer: &ProductSerializer{},
        Pagination: api.NewLimitOffsetPagination(50),
        Throttling: api.NewUserRateThrottle("100/min"),
    }
    api.RegisterViewSet(router, "/api/v1/products", viewset)
}
// Automatically serves interactive OpenAPI 3.0 documentation at /api/openapi.json`,
  },
];

const PILLARS = [
  {
    icon: '⚡',
    title: 'Type-Safe ORM & QuerySets',
    description:
      'Catch query typos at compile time. Generated expression trees, SelectRelated JOINs, PrefetchRelated batching, and orm.Q boolean logic.',
    link: '/docs/orm',
  },
  {
    icon: '🖥️',
    title: 'Instant React 19 Admin SPA',
    description:
      'Modern, responsive admin console with TanStack router, search, filter chips, saved views, audit logs, and Recharts KPI widgets.',
    link: '/docs/admin/overview',
  },
  {
    icon: '🔄',
    title: 'AST-Driven Migrations',
    description:
      'Zero-hassle schema evolution. Introspects Go AST code to detect changes, generates deterministic migrations, and verifies checksums.',
    link: '/docs/migrations',
  },
  {
    icon: '🚀',
    title: 'REST APIs & Serializers',
    description:
      'Django REST Framework-inspired ViewSets and ModelSerializers. Out-of-the-box cursor pagination, rate throttling, and OpenAPI 3.0.',
    link: '/docs/api/overview',
  },
  {
    icon: '📐',
    title: 'Declarative Schema DSL',
    description:
      'Functional & builder syntax, Generated Columns, DB default expressions, collation, custom field types, and lifecycle hooks.',
    link: '/docs/models',
  },
  {
    icon: '🛡️',
    title: 'Enterprise Identity & RBAC',
    description:
      'Secure password hashing with bcrypt, session token repositories, granular view/add/change/delete model permissions.',
    link: '/docs/identity',
  },
  {
    icon: '🔒',
    title: 'Production Security Defaults',
    description:
      'Gorilla CSRF with configurable exempt paths, HttpOnly SameSite secure cookies, CORS origin validation, and health checks.',
    link: '/docs/server/security',
  },
  {
    icon: '🛠️',
    title: 'Modern Developer CLI',
    description:
      'Full suite of CLI commands: forge new, makemigrations, migrate, generate, routes, and check for schema consistency.',
    link: '/docs/quickstart',
  },
];

const COMPARISONS = [
  {
    feature: 'Model & Schema Definition',
    stdlib: 'Manual structs + raw SQL DDL',
    others: 'GORM struct tags or SQLBoiler',
    forge: 'Declarative Schema DSL (builder + functional options)',
  },
  {
    feature: 'Query Safety',
    stdlib: 'String SQL, runtime error prone',
    others: 'Partial type-safety or strings',
    forge: '100% Type-Safe Expressions & orm.Q boolean trees',
  },
  {
    feature: 'Admin Panel',
    stdlib: 'Build from scratch (weeks of UI)',
    others: 'Third-party admin plugins or none',
    forge: 'Built-in React 19 SPA with RBAC, widgets & history',
  },
  {
    feature: 'Database Migrations',
    stdlib: 'Manual SQL scripts + migration tool',
    others: 'External CLI tools (golang-migrate)',
    forge: 'Built-in AST diffing, auto-generation & recovery',
  },
  {
    feature: 'REST API & Serialization',
    stdlib: 'Manual Chi/Gin handlers & DTOs',
    others: 'Manual boilerplate per endpoint',
    forge: 'ModelSerializers, ViewSets, Throttling & OpenAPI 3.0',
  },
  {
    feature: 'Security & CSRF',
    stdlib: 'Manual middleware wiring',
    others: 'Basic middleware or external libs',
    forge: 'Batteries-included CSRF exemptions, secure sessions & RBAC',
  },
];

export default function Home() {
  useDocusaurusContext();
  const [activeTab, setActiveTab] = useState(CODE_TABS[0].id);

  const selectedCode = CODE_TABS.find((t) => t.id === activeTab) || CODE_TABS[0];

  return (
    <>
      <SEOHead
        title="Forge - Django-like Web Framework for Go"
        description="The batteries-included Go web framework. Type-safe ORM, auto-generated React admin SPA, REST APIs, and AST migrations."
        keywords={[
          'go framework',
          'golang web framework',
          'go orm',
          'django for go',
          'type-safe orm',
          'go admin panel',
          'go migrations',
          'forge framework',
        ]}
        url="/"
      />
      <Layout
        title="The Type-Safe Django for Go"
        description="Django-inspired developer velocity with Go's raw performance and type safety.">
        
        {/* Modern Hero */}
        <header className={styles.hero}>
          <div className={styles.heroGlowLeft} />
          <div className={styles.heroGlowRight} />
          <div className={styles.heroContent}>
            <div className={styles.badge}>
              <span className={styles.badgeDot} />
              v1.0.0 • The Type-Safe Django for Go
            </div>
            <h1 className={styles.heroTitle}>
              Django Velocity.<br />
              <span className={styles.gradientText}>Go Performance.</span>
            </h1>
            <p className={styles.heroSubtitle}>
              Forge is the batteries-included web framework for Go. Declarative Schema DSL, 
              type-safe QuerySets, automatic React 19 Admin SPA, REST APIs with ModelSerializers, 
              and zero-downtime AST migrations.
            </p>
            <div className={styles.heroActions}>
              <Link className={styles.btnPrimary} to="/docs/quickstart">
                Get Started →
              </Link>
              <Link className={styles.btnSecondary} to="/docs/introduction">
                Architecture Tour
              </Link>
              <Link className={styles.btnGhost} to="/docs/features">
                Features Matrix
              </Link>
            </div>
            <div className={styles.heroCommandWrapper}>
              <div className={styles.heroCommand}>
                <span className={styles.commandPrompt}>$</span>
                <code>go install github.com/forgego/forge/cli/cmd@latest</code>
              </div>
            </div>
          </div>
        </header>

        <main>
          {/* Interactive Code Window */}
          <section className={styles.codeShowcase}>
            <div className={styles.container}>
              <div className={styles.sectionHeader}>
                <span className={styles.sectionSub}>CLEAN & DECLARATIVE</span>
                <h2>One toolkit. Infinite velocity.</h2>
                <p>Define your domain schema once. Forge generates everything else with complete type safety.</p>
              </div>

              <div className={styles.interactiveWindow}>
                {/* Window Header */}
                <div className={styles.windowHeader}>
                  <div className={styles.windowDots}>
                    <span className={styles.dotRed} />
                    <span className={styles.dotYellow} />
                    <span className={styles.dotGreen} />
                  </div>
                  <div className={styles.windowTabs}>
                    {CODE_TABS.map((tab) => (
                      <button
                        key={tab.id}
                        type="button"
                        className={`${styles.tabBtn} ${activeTab === tab.id ? styles.tabBtnActive : ''}`}
                        onClick={() => setActiveTab(tab.id)}>
                        {tab.label}
                      </button>
                    ))}
                  </div>
                  <div className={styles.windowLang}>Go 1.21+</div>
                </div>

                {/* Tab Description */}
                <div className={styles.windowDescription}>
                  <span>💡 {selectedCode.tagline}</span>
                </div>

                {/* Code Content */}
                <div className={styles.windowContent}>
                  <pre className={styles.codePre}>
                    <code>{selectedCode.code}</code>
                  </pre>
                </div>
              </div>
            </div>
          </section>

          {/* 8 Core Pillars */}
          <section className={styles.pillarsSection}>
            <div className={styles.container}>
              <div className={styles.sectionHeader}>
                <span className={styles.sectionSub}>BATTERIES INCLUDED</span>
                <h2>Built for mission-critical Go backends</h2>
                <p>Every subsystem is engineered to work seamlessly together with uncompromising type safety.</p>
              </div>

              <div className={styles.pillarsGrid}>
                {PILLARS.map((pillar) => (
                  <Link key={pillar.title} to={pillar.link} className={styles.pillarCard}>
                    <div className={styles.pillarIcon}>{pillar.icon}</div>
                    <h3 className={styles.pillarTitle}>{pillar.title}</h3>
                    <p className={styles.pillarDescription}>{pillar.description}</p>
                    <span className={styles.pillarLearnMore}>Read documentation →</span>
                  </Link>
                ))}
              </div>
            </div>
          </section>

          {/* Why Forge Comparison Table */}
          <section className={styles.matrixSection}>
            <div className={styles.container}>
              <div className={styles.sectionHeader}>
                <span className={styles.sectionSub}>BENCHMARK & PRODUCTIVITY</span>
                <h2>Why engineering teams choose Forge</h2>
                <p>Stop re-inventing admin dashboards, migration engines, and serializer plumbing.</p>
              </div>

              <div className={styles.tableWrapper}>
                <table className={styles.comparisonTable}>
                  <thead>
                    <tr>
                      <th>Framework Capability</th>
                      <th>Go Standard Library</th>
                      <th>Typical Go Stack (Chi/GORM)</th>
                      <th className={styles.forgeColHeader}>Forge Framework</th>
                    </tr>
                  </thead>
                  <tbody>
                    {COMPARISONS.map((row) => (
                      <tr key={row.feature}>
                        <td className={styles.featureCell}>{row.feature}</td>
                        <td className={styles.standardCell}>{row.stdlib}</td>
                        <td className={styles.othersCell}>{row.others}</td>
                        <td className={styles.forgeCell}>{row.forge}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </section>

          {/* 3-Step Quickstart */}
          <section className={styles.quickStart}>
            <div className={styles.container}>
              <div className={styles.quickStartContent}>
                <div className={styles.sectionHeader}>
                  <span className={styles.sectionSub}>ZERO TO PRODUCTION</span>
                  <h2>From zero to running in under 60 seconds</h2>
                  <p>Forge provides everything required to scaffold, migrate, and run your project.</p>
                </div>
                <div className={styles.quickStartSteps}>
                  <div className={styles.step}>
                    <span className={styles.stepNumber}>1</span>
                    <div className={styles.stepInfo}>
                      <h4>Create your project</h4>
                      <p>Scaffold a full project structure with SQLite or PostgreSQL.</p>
                    </div>
                    <code>forge new myapp --sqlite</code>
                  </div>
                  <div className={styles.step}>
                    <span className={styles.stepNumber}>2</span>
                    <div className={styles.stepInfo}>
                      <h4>Generate & apply migrations</h4>
                      <p>AST analysis detects your models and creates migrations automatically.</p>
                    </div>
                    <code>forge makemigrations &amp;&amp; forge migrate</code>
                  </div>
                  <div className={styles.step}>
                    <span className={styles.stepNumber}>3</span>
                    <div className={styles.stepInfo}>
                      <h4>Launch server &amp; admin</h4>
                      <p>Start Chi HTTP server, REST endpoints, and the React admin console.</p>
                    </div>
                    <code>forge runserver</code>
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* Bottom CTA */}
          <section className={styles.cta}>
            <div className={styles.container}>
              <div className={styles.ctaContent}>
                <h2>Ready to build with Forge?</h2>
                <p>Join developers building scalable, type-safe web applications with Django-grade velocity.</p>
                <div className={styles.ctaActions}>
                  <Link className={styles.btnPrimary} to="/docs/quickstart">
                    Start Quickstart Guide →
                  </Link>
                  <Link className={styles.btnOutline} href="https://github.com/hamidrabedi/foreit" target="_blank">
                    Star on GitHub
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
