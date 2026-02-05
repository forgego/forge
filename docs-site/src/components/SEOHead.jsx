import React from 'react';
import Head from '@docusaurus/Head';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

/**
 * SEO Head component for adding structured data and enhanced meta tags
 */
export default function SEOHead({
  title,
  description,
  keywords = [],
  image = '/forge-social-card.svg',
  type = 'website',
  url,
  author = 'forge Framework',
}) {
  const {siteConfig} = useDocusaurusContext();
  const siteBaseUrl = `${siteConfig.url}${siteConfig.baseUrl}`;
  const normalizedBaseUrl = siteBaseUrl.endsWith('/')
    ? siteBaseUrl.slice(0, -1)
    : siteBaseUrl;
  const fullUrl = url ? `${normalizedBaseUrl}${url}` : normalizedBaseUrl;
  const imageUrl = image.startsWith('http')
    ? image
    : `${normalizedBaseUrl}${image}`;
  const keywordsString = Array.isArray(keywords) ? keywords.join(', ') : keywords;

  // Structured data (JSON-LD)
  const structuredData = {
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    name: 'forge',
    applicationCategory: 'WebApplication',
    operatingSystem: 'Any',
    description: description || 'Django-like Go framework with type safety',
    url: fullUrl,
    author: {
      '@type': 'Organization',
      name: author,
    },
    offers: {
      '@type': 'Offer',
      price: '0',
      priceCurrency: 'USD',
    },
    aggregateRating: {
      '@type': 'AggregateRating',
      ratingValue: '5',
      ratingCount: '1',
    },
  };

  return (
    <Head>
      {/* Primary Meta Tags */}
      {title && <meta property="og:title" content={title} />}
      {description && (
        <>
          <meta name="description" content={description} />
          <meta property="og:description" content={description} />
          <meta name="twitter:description" content={description} />
        </>
      )}
      {keywordsString && <meta name="keywords" content={keywordsString} />}

      {/* Open Graph / Facebook */}
      <meta property="og:type" content={type} />
      <meta property="og:url" content={fullUrl} />
      <meta property="og:image" content={imageUrl} />
      <meta property="og:site_name" content="forge Framework" />

      {/* Twitter */}
      <meta name="twitter:card" content="summary_large_image" />
      <meta name="twitter:url" content={fullUrl} />
      <meta name="twitter:title" content={title} />
      <meta name="twitter:image" content={imageUrl} />
      <meta name="twitter:site" content="@forgego" />
      <meta name="twitter:creator" content="@forgego" />

      {/* Canonical URL */}
      <link rel="canonical" href={fullUrl} />

      {/* Structured Data */}
      <script type="application/ld+json">
        {JSON.stringify(structuredData)}
      </script>
    </Head>
  );
}
