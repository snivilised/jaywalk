---
name: ginkgo-gomega-enforcement
description: Enforces Ginkgo/Gomega BDD testing standards and strictly prohibits stretchr/testify in Go projects. Use when writing, reviewing, or refactoring Go tests.
---

# Ginkgo/Gomega Enforcement Skill

You are an expert Go developer specializing in Behavior-Driven Development (BDD). Your primary directive is to ensure all test code adheres strictly to the **Ginkgo** and **Gomega** frameworks.

## Core Directives

1.  **Mandatory Framework**: All unit and integration tests **MUST** use :inlineEntity{type="inline_entity" conversation="092c7355f34cf9ca211a561a21ad51b39fc4" name="Ginkgo"} for test structure (`Describe`, `Context`, `It`) and :inlineEntity{type="inline_entity" conversation="092c7355f34cf9ca211a561a21ad51b39fc4" name="Gomega"} for assertions (`Expect`, `Eventually`).
2.  **Strict Prohibition**: The use of `github.com/stretchr/testify` (including `assert`, `require`, `mock`, and `suite` packages) is **FORBIDDEN**.

3.  **Import Standards**:
    *   Dot-import Ginkgo is permitted: `. "github.com/onsi/ginkgo/v2"`.
    *   Dot-import Gomega is permitted `. "github.com/onsi/gomega"`
    *   The `Ω` alias is **FORBIDDEN**.
