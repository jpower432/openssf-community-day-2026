# Gemara: The GRC Architecture You Didn't Know You Built

Demo repository for [OpenSSF Community Day 2026](https://openssf.org/).

## Different Audiences, Different Purposes

**OSCAL communicates upward and outward.** Its models document how an organization's policies and procedures satisfy external requirements. The audience is the assessor, the authorizing official, the regulator. SSPs, assessment results, and POA&Ms are proof artifacts -- they demonstrate compliance to an external party.

OSCAL does not communicate downward and inward. It is not designed to disseminate security guidance to the people who are subject to those policies and need to demonstrate conformance. An OSCAL SSP tells an auditor that TLS 1.3 is enforced. It does not tell an operator *why* that matters, *how* to verify it, or *what else* to consider.

**Gemara communicates inward and downward.** Its models give practitioners the security knowledge they need to act: what threats exist, what controls to implement, how to assess whether they've done it right. Threat catalogs, capability mappings, control objectives, and assessment requirements are operational artifacts -- they enable conformance.

| | **Gemara** | **OSCAL** |
|:--|:--|:--|
| **Direction** | Inward, to practitioners | Outward, to assessors |
| **Purpose** | Enable conformance | Prove compliance |
| **Audience** | Maintainer, operator, SRE | Auditor, regulator, authorizing official |
| **Captures** | Threats, capabilities, controls, evaluation | SSP, profiles, assessment results, POA&M |
| **Question** | "What should I do and why?" | "Can you prove you did it?" |

## Where They Meet

Projects like [FINOS Common Cloud Controls](https://github.com/finos/common-cloud-controls) author shared, reusable controls in Gemara. Upstream projects import those controls and specialize them with project-specific threat context and assessment requirements. Operators use that guidance to secure their deployments.

When those operators sit inside a regulated organization, the compliance team needs to prove conformance to an external party. That is where OSCAL enters. Tooling converts Gemara control catalogs and evaluation logs into OSCAL Catalogs and Assessment Results -- the last mile from operational guidance to formal evidence.

Neither replaces the other. Gemara has no SSP or POA&M. OSCAL has no threats or capabilities. They share a boundary at **control catalogs** and **assessment results**.

## What This Demo Shows

A Go CLI loads [Gemara governance artifacts](#governance-artifacts) for the [CloudNativePG](https://cloudnative-pg.io/) operator, validates cross-references between threat and control catalogs, and converts them to OSCAL using the [`go-gemara`](https://github.com/gemaraproj/go-gemara) SDK.

```
Gemara YAML ──▶ go-gemara SDK ──▶ OSCAL JSON + Markdown
(engineering)     (bridge)         (last mile)
```

A GitHub Actions workflow runs this on push to `main` and publishes the OSCAL artifacts as a GitHub Release.

### Input: Gemara Artifacts

| File | Type | Detail |
|:--|:--|:--|
| `governance/controls/controls.yaml` | ControlCatalog | 14 controls across 5 groups |
| `governance/catalogs/threat-catalog.yaml` | ThreatCatalog | Project-specific + imported threats |
| `governance/catalogs/capabilities.yaml` | CapabilityCatalog | Project-specific + imported capabilities |
| `governance/evaluation-log.yaml` | EvaluationLog | 5 control evaluations (3 Passed, 1 Failed, 1 Needs Review) |

### Output: OSCAL Artifacts

| File | OSCAL Type | Source |
|:--|:--|:--|
| `oscal-catalog.json` | Catalog | `ControlCatalog.ToOSCAL()` |
| `oscal-assessment-results.json` | Assessment Results | `EvaluationLog.ToOSCAL()` |
| `controls.md` | Markdown | `ControlCatalog.ToMarkdown()` |
| `summary.md` | Markdown | Assessment summary with pass/fail breakdown |

## Quick Start

```shell
make run
```

Builds the Go binary and runs the conversion. Output lands in `output/`.

```shell
make clean
```

Removes build and output directories.

## Governance Artifacts

The `governance/` directory contains Gemara artifacts for the CNCF CloudNativePG project. The original threat assessment was created during the [2026 Security Slam](https://github.com/cloudnative-pg/cloudnative-pg/pull/10304) and migrated to Gemara v1.0.0 using the [Gemara MCP Server](https://github.com/gemaraproj/gemara-mcp).

## Links

- [Gemara](https://gemara.openssf.org/)
- [OSCAL](https://pages.nist.gov/OSCAL/)
- [go-gemara SDK](https://github.com/gemaraproj/go-gemara)
- [CloudNativePG](https://cloudnative-pg.io/)
