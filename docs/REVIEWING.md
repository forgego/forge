# Reviewing Forge as a system

Review changed behavior in the surrounding system, not only changed lines.
Read the PR requirements, non-goals, [design](DESIGN.md), relevant subsystem
guidance, and [documentation index](README.md). Separate intended contracts
from verified implementation; neither a roadmap checkbox nor a passing process
exit proves a capability works.

## Review scope

For a large or cross-package change, request an explicit system audit and split
the work by affected flows. Record the base and head revisions, files inspected,
tests executed, and unreviewed areas. Do not claim every line was reviewed unless
that coverage was actually recorded. Review is read-only unless fixes are requested.

Hosted Codex uses applicable AGENTS.md guidance and focuses on P0/P1 findings,
according to [OpenAI's GitHub review documentation](https://learn.chatgpt.com/docs/third-party/github)
(checked 2026-09-14). That is not a guarantee of a full architecture audit.
Local reviewer model configuration does not establish the hosted model identity.
Instructions supplement tests and human judgment; they do not replace them.

## Seven passes

1. **Change map.** Identify changed subsystems, entry points, exported APIs,
   configuration, schema/migrations, dependencies, deleted behavior, and affected
   callers, including unchanged code. State the user need and explicit non-goals.
2. **End-to-end flows.** Trace each actual path, citing source locations at every
   relevant boundary. For runtime writes, follow route or command, principal,
   authorization and object scope, parsing/validation, business operation,
   transaction, ORM/driver/database constraints, error/response, and client UI.
   Do not assume every entry point shares the same service. Separately trace
   schema definition, parsing/normalization, generation, consumer compilation,
   migration diff/SQL review, application, status, and recovery. Ordinary requests
   do not pass through migration generation.
3. **Architecture and feature necessity.** Check responsibility placement,
   dependency direction, package coupling, global state ownership and isolation,
   duplicated abstractions, and extension/escape routes. Keep shared field facts
   distinct from admin/API audience policy. Ask whether application-specific
   behavior belongs in the framework. Identify the concrete requirement served,
   the simpler alternative, and the maintenance cost of each new abstraction.
4. **Names and public contracts.** Examine package/file names, exported types and
   functions, routes, fields, tables, columns, CLI flags, and configuration keys.
   Check terminology, singular/plural consistency, ambiguous booleans, accidental
   exports, and ownership. Explain compatibility and migration costs of renames;
   do not present preference alone as a correctness blocker.
5. **Adversarial behavior.** Exercise happy paths, invalid and boundary inputs,
   unauthenticated/unauthorized access, retries and duplicates, concurrent writes,
   cancellation, database failures, partial effects, rollback, and old data.
   Examine unbounded queries, N+1 behavior, swallowed errors, safe disclosure,
   logging/correlation, and whether external effects occur before commit.
   Compare API, admin, CLI, and worker paths where they exist; mark absent paths.
6. **Requirement-to-test evidence.** Map each affected requirement to test symbols,
   scenarios, command/environment, and results. Include failure paths, supported
   database differences, migration recovery, and independent generated consumers.
   Report pass/fail/skip counts. Missing required database infrastructure is a
   coverage gap, not a successful database test. Distinguish inspection from execution.
7. **Synthesis.** Present substantiated merge blockers first, then non-blocking
   risks and architectural proposals. Include naming/placement changes, unnecessary
   features, missing requirements, and alternatives. Recommend merge, split,
   redesign, defer, or remove with reasons and explicit remaining uncertainty.

## Acceptance review

Implementation and acceptance are separate roles. The author of a capability,
whether a person or a model, writes tests but does not accept the capability.
Agreement between an implementation, its own tests, and its own documentation
does not show that the user-facing contract is right.

- The acceptance reviewer starts from the requirement and the resulting diff,
  not from the author's summary.
- The reviewer exercises the capability through public interfaces: HTTP
  routes, CLI commands, generated code, or exported APIs.
- For each capability, the reviewer records at least one attempted
  counterexample, such as an undeclared field in a response, a leaked error
  detail, an unauthorized write, or an interrupted migration, together with the
  observed outcome.
- A capability is complete only when that record exists, alongside the
  requirement-to-test evidence from pass 6.

## Evidence templates

For each flow, record:

- Requirement, actor, entry point, and expected observable outcome.
- Actual call chain with file/line or symbol references, including unchanged code.
- Authorization scope, business invariant, transaction owner, and side effects.
- Happy/failure/retry/concurrency/rollback/legacy-data scenarios checked.
- Response and UI behavior, diagnostic evidence, and compatibility implications.
- Tests run, results, inspection-only conclusions, and unverified boundaries.

| Requirement / flow | Test symbol and location | Command / environment | Result / evidence | Missing scenario and test level |
| --- | --- | --- | --- | --- |
| State the contract | Existing test or none | Exact command and prerequisites | Pass, fail, skip, or not run | Concrete follow-up |

A finding needs a trigger, exact execution path, violated contract, consequence,
and evidence. An architecture proposal needs the current responsibility, why its
placement or necessity is questionable, a feasible alternative, and trade-offs.
Keep these categories separate. A green test matrix does not certify untested flows.
