import React, { useState } from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import ThemedImage from '@theme/ThemedImage';
import useBaseUrl from '@docusaurus/useBaseUrl';
import SEOHead from '@site/src/components/SEOHead';
import styles from './index.module.css';

// Samples follow the code in examples/ecommerce so they stay honest.
const CODE_TABS = [
  {
    id: 'schema',
    label: 'Model',
    tagline: 'Describe fields, indexes and relations in plain Go. Everything else is generated from this.',
    code: `package catalog

import "github.com/forgego/forge/schema"

type Product struct {
    schema.BaseSchema
}

func (Product) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("name", schema.Required(), schema.MaxLength(200)),
        schema.StringField("sku", schema.Required(), schema.Unique()),
        schema.Float64Field("price", schema.Required()),
        schema.BoolField("is_active", schema.Default(true)),
        schema.TimeField("created_at", schema.AutoNowAdd()),
    }
}

func (Product) Meta() schema.Meta {
    return schema.Meta{TableName: "products", OrderBy: []string{"name"}}
}`,
  },
  {
    id: 'orm',
    label: 'Query',
    tagline: 'forge generate writes a manager and typed fields for each model. A misspelled column is a compile error.',
    code: `p := catalog.ProductFieldsInstance

qs, err := catalog.ProductObjects.Filter(orm.And(
    p.IsActive.Eq(true),
    p.Price.Lte(150),
))
if err != nil {
    return err
}

products, err := qs.
    OrderBy(p.Price.Desc()).
    Limit(20).
    All(ctx)`,
  },
  {
    id: 'admin',
    label: 'Admin',
    tagline: 'Register a model to get list, search, filter and edit screens in the built-in admin.',
    code: `p := catalog.ProductFieldsInstance

admin.Register(&admin.Config[catalog.Product]{
    ListDisplay:    []admin.Field{p.Name, p.Sku, p.Price, p.IsActive},
    ListFilter:     []admin.Field{p.IsActive},
    SearchFields:   []admin.Field{p.Name, p.Sku},
    ReadOnlyFields: []string{"created_at"},
})`,
  },
  {
    id: 'api',
    label: 'REST API',
    tagline: 'Expose a model as a REST resource with a ViewSet, in the style of Django REST Framework.',
    code: `router.Register("products", &api.ViewSetConfig{
    Model:      &catalog.Product{},
    Queryset:   catalog.ProductObjects,
    Serializer: api.NewBaseSerializer(nil),
})`,
  },
];

const PILLARS = [
  {
    title: 'Schema-first models',
    description:
      'Fields, indexes, relations and hooks live in one Go definition per model. The generator, migrations, admin and API all read from it.',
    link: '/docs/models',
  },
  {
    title: 'Typed queries',
    description:
      'Generated managers and field accessors give you chainable, lazy QuerySets with typed comparisons, ordering, relations and aggregates.',
    link: '/docs/orm',
  },
  {
    title: 'Migrations from your models',
    description:
      'makemigrations --auto compares your models with the migration history and writes the SQL. Checksums catch migration files edited after they ran.',
    link: '/docs/migrations',
  },
  {
    title: 'Built-in admin',
    description:
      'A React admin with search, filters, bulk actions, inline relations, change history and per-object permission hooks.',
    link: '/docs/admin/overview',
  },
  {
    title: 'REST API layer',
    description:
      'ViewSets, serializers, pagination, throttling, versioning and an OpenAPI document, modelled on Django REST Framework.',
    link: '/docs/api/overview',
  },
  {
    title: 'Secure defaults',
    description:
      'Password hashing, sessions and tokens, CSRF protection, secure cookies, CORS checks and rate limiting are on from the start.',
    link: '/docs/server/security',
  },
];

export default function Home() {
  const [activeTab, setActiveTab] = useState(CODE_TABS[0].id);
  const selectedCode = CODE_TABS.find((t) => t.id === activeTab) || CODE_TABS[0];
  const adminLight = useBaseUrl('/img/admin-products.png');
  const adminDark = useBaseUrl('/img/admin-products-dark.png');

  return (
    <>
      <SEOHead
        title="Forge: the batteries-included web framework for Go"
        description="Define a model once. Forge generates typed queries, SQL migrations, a REST API and an admin panel for it."
        keywords={[
          'go web framework',
          'golang framework',
          'go orm',
          'go admin panel',
          'go migrations',
          'django for go',
        ]}
        url="/"
      />
      <Layout
        title="The batteries-included web framework for Go"
        description="Define a model once. Forge generates typed queries, SQL migrations, a REST API and an admin panel for it.">
        <header className={styles.hero}>
          <div className={styles.heroGlowLeft} />
          <div className={styles.heroGlowRight} />
          <div className={styles.heroContent}>
            <div className={styles.badge}>
              <span className={styles.badgeDot} />
              Open source, MIT licensed, pre-1.0
            </div>
            <h1 className={styles.heroTitle}>
              Define the model.
              <br />
              <span className={styles.gradientText}>Forge builds the rest.</span>
            </h1>
            <p className={styles.heroSubtitle}>
              Forge is a batteries-included web framework for Go. Describe your data once, and
              Forge generates typed queries, SQL migrations, a REST API and an admin panel for it.
            </p>
            <div className={styles.heroActions}>
              <Link className={styles.btnPrimary} to="/docs/quickstart">
                Get started
              </Link>
              <Link className={styles.btnSecondary} to="/docs/introduction">
                Read the introduction
              </Link>
              <Link className={styles.btnGhost} href="https://github.com/forgego/forge">
                View on GitHub
              </Link>
            </div>
            <div className={styles.heroCommandWrapper}>
              <div className={styles.heroCommand}>
                <span className={styles.commandPrompt}>$</span>
                <code>go install github.com/forgego/forge/cmd/forge@latest</code>
              </div>
            </div>
          </div>
        </header>

        <main>
          <section className={styles.codeShowcase}>
            <div className={styles.container}>
              <div className={styles.sectionHeader}>
                <span className={styles.sectionSub}>How it fits together</span>
                <h2>One definition, four outputs</h2>
                <p>Write the model. Forge reads it to build queries, the admin and the API.</p>
              </div>

              <div className={styles.interactiveWindow}>
                <div className={styles.windowHeader}>
                  <div className={styles.windowDots}>
                    <span className={styles.dotRed} />
                    <span className={styles.dotYellow} />
                    <span className={styles.dotGreen} />
                  </div>
                  <div className={styles.windowTabs} role="tablist">
                    {CODE_TABS.map((tab) => (
                      <button
                        key={tab.id}
                        type="button"
                        role="tab"
                        aria-selected={activeTab === tab.id}
                        className={`${styles.tabBtn} ${activeTab === tab.id ? styles.tabBtnActive : ''}`}
                        onClick={() => setActiveTab(tab.id)}>
                        {tab.label}
                      </button>
                    ))}
                  </div>
                  <div className={styles.windowLang}>Go 1.26+</div>
                </div>

                <div className={styles.windowDescription}>
                  <span>{selectedCode.tagline}</span>
                </div>

                <div className={styles.windowContent}>
                  <pre className={styles.codePre}>
                    <code>{selectedCode.code}</code>
                  </pre>
                </div>
              </div>
            </div>
          </section>

          <section className={styles.adminShowcase}>
            <div className={styles.container}>
              <div className={styles.sectionHeader}>
                <span className={styles.sectionSub}>The admin</span>
                <h2>An admin for every model you register</h2>
                <p>Search, filters, saved views, bulk actions and export. Shown here with the ecommerce example.</p>
              </div>
              <ThemedImage
                className={styles.adminShot}
                alt="Forge admin showing the product list from the ecommerce example"
                sources={{light: adminLight, dark: adminDark}}
                width={1800}
                height={1075}
                loading="lazy"
              />
            </div>
          </section>

          <section className={styles.pillarsSection}>
            <div className={styles.container}>
              <div className={styles.sectionHeader}>
                <span className={styles.sectionSub}>What is included</span>
                <h2>The parts most web backends need</h2>
                <p>Each part works on its own, and they are designed to work together.</p>
              </div>

              <div className={styles.pillarsGrid}>
                {PILLARS.map((pillar) => (
                  <Link key={pillar.title} to={pillar.link} className={styles.pillarCard}>
                    <h3 className={styles.pillarTitle}>{pillar.title}</h3>
                    <p className={styles.pillarDescription}>{pillar.description}</p>
                    <span className={styles.pillarLearnMore}>Read the docs</span>
                  </Link>
                ))}
              </div>
            </div>
          </section>

          <section className={styles.quickStart}>
            <div className={styles.container}>
              <div className={styles.quickStartContent}>
                <div className={styles.sectionHeader}>
                  <span className={styles.sectionSub}>Try it</span>
                  <h2>See a complete Forge app running</h2>
                  <p>
                    The ecommerce example has 57 models across 10 apps, with the admin and REST API
                    wired up. Docker runs it with PostgreSQL.
                  </p>
                </div>
                <div className={styles.quickStartSteps}>
                  <div className={styles.step}>
                    <span className={styles.stepNumber}>1</span>
                    <div className={styles.stepInfo}>
                      <h4>Clone the repository</h4>
                      <p>The example lives in examples/ecommerce.</p>
                    </div>
                    <code>git clone https://github.com/forgego/forge.git</code>
                  </div>
                  <div className={styles.step}>
                    <span className={styles.stepNumber}>2</span>
                    <div className={styles.stepInfo}>
                      <h4>Start it with Docker</h4>
                      <p>Builds the app and the admin UI, and starts PostgreSQL.</p>
                    </div>
                    <code>cd forge/examples/ecommerce &amp;&amp; docker compose up --build</code>
                  </div>
                  <div className={styles.step}>
                    <span className={styles.stepNumber}>3</span>
                    <div className={styles.stepInfo}>
                      <h4>Open the admin</h4>
                      <p>Browse the generated admin for every model in the example.</p>
                    </div>
                    <code>http://localhost:8020/admin/</code>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section className={styles.cta}>
            <div className={styles.container}>
              <div className={styles.ctaContent}>
                <h2>Know what you are adopting</h2>
                <p>
                  Forge is pre-1.0. PostgreSQL is the primary tested database. The capability status
                  page lists what is verified in CI and what is still partial.
                </p>
                <div className={styles.ctaActions}>
                  <Link className={styles.btnPrimary} to="/docs/status">
                    Capability status
                  </Link>
                  <Link className={styles.btnOutline} href="https://github.com/forgego/forge">
                    Source on GitHub
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
