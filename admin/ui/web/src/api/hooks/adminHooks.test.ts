import { createElement, type ReactNode } from "react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useUpdateObject,
  useCreateObject,
  adminKeys,
} from "./adminHooks";
import { adminAPI } from "../client";

vi.mock("../client", () => ({
  adminAPI: {
    updateObject: vi.fn(),
    createObject: vi.fn(),
  },
}));

describe("adminHooks mutation callbacks", () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  const createWrapper = () => {
    return ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client: queryClient }, children);
  };

  it("forwards all 4 arguments to onSuccess in useUpdateObject (including onMutateResult and context)", async () => {
    const invalidateQueriesSpy = vi.spyOn(queryClient, "invalidateQueries");
    const resolvedData = { id: 42, title: "Updated Item" };
    const variables = { id: 42, data: { title: "Updated Item" } };
    const onMutate = vi.fn().mockReturnValue({ tag: "m" });
    const onSuccess = vi.fn();

    vi.mocked(adminAPI.updateObject).mockResolvedValue(resolvedData);

    const { result } = renderHook(
      () =>
        useUpdateObject("products", {
          onMutate,
          onSuccess,
        }),
      { wrapper: createWrapper() }
    );

    await act(async () => {
      await result.current.mutateAsync(variables);
    });

    // Verify onMutate was called
    expect(onMutate).toHaveBeenCalledWith(variables, expect.anything());

    // Verify onSuccess was called with 4 arguments: (data, variables, onMutateResult, context)
    expect(onSuccess).toHaveBeenCalledTimes(1);
    expect(onSuccess).toHaveBeenCalledWith(
      resolvedData,
      variables,
      { tag: "m" },
      expect.anything()
    );

    // Verify invalidateQueries ran for both detail and model list
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({
      queryKey: adminKeys.modelDetail("products", 42),
    });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({
      queryKey: adminKeys.model("products"),
    });
  });

  it("forwards all 4 arguments to onSuccess in mutations using withInvalidation", async () => {
    const invalidateQueriesSpy = vi.spyOn(queryClient, "invalidateQueries");
    const resolvedData = { id: 99, title: "Created Item" };
    const variables = { title: "Created Item" };
    const onMutate = vi.fn().mockReturnValue({ tag: "create-tag" });
    const onSuccess = vi.fn();

    vi.mocked(adminAPI.createObject).mockResolvedValue(resolvedData);

    const { result } = renderHook(
      () =>
        useCreateObject("products", {
          onMutate,
          onSuccess,
        }),
      { wrapper: createWrapper() }
    );

    await act(async () => {
      await result.current.mutateAsync(variables);
    });

    expect(onMutate).toHaveBeenCalledWith(variables, expect.anything());
    expect(onSuccess).toHaveBeenCalledTimes(1);
    expect(onSuccess).toHaveBeenCalledWith(
      resolvedData,
      variables,
      { tag: "create-tag" },
      expect.anything()
    );

    expect(invalidateQueriesSpy).toHaveBeenCalledWith({
      queryKey: adminKeys.model("products"),
    });
  });
});
