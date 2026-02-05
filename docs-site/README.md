# forge Documentation Site

Documentation website for the forge framework, built with Docusaurus and optimized for fast builds.

## Quick Start

### Using pnpm (Recommended - Fastest) ⚡

```bash
# Install dependencies (fastest! - uses npx, no global install needed)
npm install
# or
npm run install

# Start development server
npm start

# Build for production
npm run build

# Serve production build
npm run serve
```

**Note**: pnpm is used automatically via npx (no global install needed). If you have pnpm installed globally, you can use `pnpm` commands directly.

### Using Yarn (Alternative)

```bash
# Install dependencies
yarn install

# Start development server
yarn start

# Build for production
yarn build
```

### Using npm (Fallback)

```bash
# Install dependencies
npm ci

# Start development server
npm start

# Build for production
npm run build
```

## Performance

This site is optimized for **fastest possible builds and installs**:

- **pnpm** - Fastest package manager (2-3x faster than npm)
- **Build optimizations** - Disabled unnecessary features
- **Smart caching** - Faster subsequent builds
- **Optimized config** - Minimal overhead

### Performance Benchmarks

**Installation**:
- pnpm: ~15-30s (first), ~3-5s (cached) ⚡
- yarn: ~30-60s (first), ~5-10s (cached)
- npm: ~60-120s (first), ~10-20s (cached)

**Build**:
- First build: ~30-45s
- Cached build: ~15-25s
- Dev server: ~3-5s startup

## Development

```bash
# Fast build (no minification, for testing)
pnpm build:fast

# Clear cache and rebuild
pnpm clear && pnpm build
```

## Deployment

The site is configured for GitHub Pages deployment:

```bash
pnpm deploy
```

## Alternative: Astro Starlight

If you need **even faster builds** (2-3x faster), consider migrating to **Astro Starlight**. See `PERFORMANCE.md` for details.

**Astro Starlight benefits**:
- Build time: ~10-15s (vs 30-45s)
- Smaller bundle sizes
- Faster dev server

**Trade-offs**:
- Migration effort (3-4 hours)
- More manual setup
- Less built-in features than Docusaurus

## Why pnpm?

- **Fastest installs** - 2-3x faster than npm
- **Less disk usage** - ~50% less space
- **Better caching** - Smarter dependency resolution
- **Industry standard** - Used by major projects

See `PERFORMANCE.md` for detailed benchmarks and optimization tips.
