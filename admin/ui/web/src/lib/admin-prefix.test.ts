import { describe, expect, it, beforeEach, afterEach } from "vitest";
import {
  DEFAULT_PREFIX,
  normalizePrefix,
  getAdminPrefix,
  stripAdminPrefix,
  withAdminPrefix,
  assetUrl,
} from "./admin-prefix";

describe("admin-prefix", () => {
  describe("normalizePrefix", () => {
    it("normalizes undefined to DEFAULT_PREFIX", () => {
      expect(normalizePrefix(undefined)).toBe("/admin");
    });

    it("normalizes null to DEFAULT_PREFIX", () => {
      expect(normalizePrefix(null)).toBe("/admin");
    });

    it('normalizes "" to ""', () => {
      expect(normalizePrefix("")).toBe("");
    });

    it('normalizes "/" to ""', () => {
      expect(normalizePrefix("/")).toBe("");
    });

    it('normalizes "backoffice" to "/backoffice"', () => {
      expect(normalizePrefix("backoffice")).toBe("/backoffice");
    });

    it('normalizes "/x/" to "/x"', () => {
      expect(normalizePrefix("/x/")).toBe("/x");
    });

    it('normalizes whitespace around values', () => {
      expect(normalizePrefix("  /custom  ")).toBe("/custom");
    });
  });

  describe("getAdminPrefix", () => {
    let originalHeadHtml: string;

    beforeEach(() => {
      originalHeadHtml = document.head.innerHTML;
      document.head.innerHTML = "";
    });

    afterEach(() => {
      document.head.innerHTML = originalHeadHtml;
    });

    it('returns "/admin" when no meta tag is present', () => {
      expect(getAdminPrefix(document)).toBe(DEFAULT_PREFIX);
      expect(getAdminPrefix()).toBe(DEFAULT_PREFIX);
    });

    it('returns "/backoffice" when meta tag is "/backoffice"', () => {
      const meta = document.createElement("meta");
      meta.name = "forge-admin-prefix";
      meta.content = "/backoffice";
      document.head.appendChild(meta);

      expect(getAdminPrefix(document)).toBe("/backoffice");
      expect(getAdminPrefix()).toBe("/backoffice");
    });

    it('returns "" when meta tag is empty', () => {
      const meta = document.createElement("meta");
      meta.name = "forge-admin-prefix";
      meta.content = "";
      document.head.appendChild(meta);

      expect(getAdminPrefix(document)).toBe("");
      expect(getAdminPrefix()).toBe("");
    });
  });

  describe("stripAdminPrefix", () => {
    it('strips prefix from "/admin/categories" -> "/categories"', () => {
      expect(stripAdminPrefix("/admin/categories", "/admin")).toBe("/categories");
    });

    it('returns "/" when path matches prefix exactly', () => {
      expect(stripAdminPrefix("/admin", "/admin")).toBe("/");
    });

    it('does not strip prefix if path does not have slash boundary', () => {
      expect(stripAdminPrefix("/administrator", "/admin")).toBe("/administrator");
    });

    it('returns path as-is when prefix is ""', () => {
      expect(stripAdminPrefix("/x", "")).toBe("/x");
    });

    it('handles trailing slash on prefix path', () => {
      expect(stripAdminPrefix("/admin/", "/admin")).toBe("/");
    });
  });

  describe("withAdminPrefix", () => {
    it('joins ("/api", "/admin") -> "/admin/api"', () => {
      expect(withAdminPrefix("/api", "/admin")).toBe("/admin/api");
    });

    it('joins ("/api", "") -> "/api"', () => {
      expect(withAdminPrefix("/api", "")).toBe("/api");
    });

    it('joins ("api", "/b") -> "/b/api"', () => {
      expect(withAdminPrefix("api", "/b")).toBe("/b/api");
    });

    it('joins ("api", "") -> "/api"', () => {
      expect(withAdminPrefix("api", "")).toBe("/api");
    });

    it('uses default prefix when prefix argument is omitted', () => {
      expect(withAdminPrefix("/api")).toBe("/admin/api");
    });
  });

  describe("assetUrl and window registration", () => {
    it("constructs asset URLs with the admin prefix", () => {
      expect(assetUrl("assets/app.js")).toBe(withAdminPrefix("/assets/app.js"));
      expect(assetUrl("/assets/app.js")).toBe(withAdminPrefix("/assets/app.js"));
    });

    it("registers __forgeAssetUrl on window", () => {
      expect(window.__forgeAssetUrl).toBeDefined();
      expect(window.__forgeAssetUrl?.("assets/test.js")).toBe(assetUrl("assets/test.js"));
    });
  });
});
