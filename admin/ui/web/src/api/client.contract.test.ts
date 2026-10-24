import { describe, it, expect, vi, beforeEach } from "vitest";
import { AdminAPIClient, ADMIN_CONTRACT_VERSION } from "./client";
import type { Metadata } from "./types";

describe("AdminAPIClient contract version", () => {
  let client: AdminAPIClient;

  beforeEach(() => {
    client = new AdminAPIClient();
  });

  it("succeeds when contract_version matches ADMIN_CONTRACT_VERSION", async () => {
    const mockMeta: Metadata = {
      name: "products",
      verbose_name: "Product",
      verbose_name_plural: "Products",
      fields: [],
      permissions: { add: true, change: true, delete: true, view: true },
      actions: [],
      filters: [],
      list_display: [],
      pagination: { page_size: 25, max_page_size: 100 },
      contract_version: ADMIN_CONTRACT_VERSION,
    };

    vi.spyOn(
      (client as unknown as { client: { get: (...args: unknown[]) => Promise<unknown> } }).client,
      "get"
    ).mockResolvedValue({ data: mockMeta });

    const result = await client.getModelMetadata("products");
    expect(result).toEqual(mockMeta);
    expect(result.contract_version).toBe(ADMIN_CONTRACT_VERSION);
  });

  it("throws when contract_version mismatches ADMIN_CONTRACT_VERSION", async () => {
    const serverVersion = 99;
    const mockMeta: Metadata = {
      name: "products",
      verbose_name: "Product",
      verbose_name_plural: "Products",
      fields: [],
      permissions: { add: true, change: true, delete: true, view: true },
      actions: [],
      filters: [],
      list_display: [],
      pagination: { page_size: 25, max_page_size: 100 },
      contract_version: serverVersion,
    };

    vi.spyOn(
      (client as unknown as { client: { get: (...args: unknown[]) => Promise<unknown> } }).client,
      "get"
    ).mockResolvedValue({ data: mockMeta });

    await expect(client.getModelMetadata("products")).rejects.toThrowError(
      new RegExp(`^Admin UI is incompatible with this server.*${serverVersion}.*${ADMIN_CONTRACT_VERSION}`)
    );
  });

  it("succeeds without error when contract_version is missing", async () => {
    const mockMeta: Metadata = {
      name: "products",
      verbose_name: "Product",
      verbose_name_plural: "Products",
      fields: [],
      permissions: { add: true, change: true, delete: true, view: true },
      actions: [],
      filters: [],
      list_display: [],
      pagination: { page_size: 25, max_page_size: 100 },
    };

    vi.spyOn(
      (client as unknown as { client: { get: (...args: unknown[]) => Promise<unknown> } }).client,
      "get"
    ).mockResolvedValue({ data: mockMeta });

    const result = await client.getModelMetadata("products");
    expect(result).toEqual(mockMeta);
    expect(result.contract_version).toBeUndefined();
  });
});
