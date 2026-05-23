---
name: go-i18n
description: >
  MANDATORY for any production code in this project that emits text a
  CLI user reads as prose, or returns an error a CLI user could trigger.
  Defines the complete, non-deferrable workflow: add entries to the
  Underliers map in underlying-templ-data.go, run lingo
  (go run ./cmd/lingo/...) to regenerate auto files, then write call
  sites using generated types. fmt.Sprintf / fmt.Errorf / errors.New
  are permitted for formatting, internal errors, fixed-token content,
  test code, and demo programs - section 2 of this skill defines all
  exemptions precisely. When the situation is genuinely ambiguous, skip
  i18n and use fmt directly.
---

# SKILL: Adding i18n Messages via lingo

## Mandatory Pre-Check

Before writing any string or error, apply the decision test in section 2.
If i18n is required, complete the four-step workflow in section 6 in full
before writing any call site. No step may be skipped or deferred.

---

## 1. What i18n Covers

i18n is required when **all** of the following are true:

1. The string is read by a CLI user as natural-language prose - a
   message, status line, warning, or error description.
2. The string could reasonably differ between languages.
3. The code is production code (not a test file, not a demo program).

If any condition is false, i18n is not required. Use `fmt.Sprintf`,
`fmt.Errorf`, `errors.New`, etc. freely and move on.

---

## 2. Exemptions - When `fmt` and `errors` Are Correct

The following categories are **exempt from i18n**. Using `fmt.*` or
`errors.*` directly is correct in each case. Categories marked with *
require an exemption comment; the others are self-evident from context.
When genuinely unsure, skip i18n and use `fmt` directly - over-applying
i18n to internal strings creates noise without benefit.

### 2a. Pure Formatting - no comment needed

`fmt.Sprintf` assembling a computed value (timestamp, ID, duration,
path fragment) with no natural-language prose. The result is read as
data, not as a message.

```go
// Fine - computed value, not a translatable message.
func (g *Sequential) Generate() string {
    g.id++
    return fmt.Sprintf(g.Format, g.id)
}
```

### 2b. Fixed-Token Content - no comment needed

Strings that name code artefacts - flag names, enum values, package
paths - do not translate because they are fixed in code. `--dry-run` is
always `--dry-run` regardless of locale. Use `fmt.Errorf` freely for
errors whose entire content is fixed tokens. If a fixed token appears
inside a broader translatable message, embed it as a literal in the
`Other` string and use `Description` to advise human translators.

```go
// Fine - flag name is a code artefact, not translatable prose.
fmt.Errorf("flag --%s requires a value", flagName)
```

### 2c. Internal / Structural Errors *

Errors describing internal state that only a developer reads - from
logs, panics, or inter-package sentinel chains. Must carry an
`internalError:` comment confirming they are not surfaced to the CLI
user.

```go
// internalError: consumed by the coordinator log, never shown to the
// CLI user as actionable output.
return fmt.Errorf("dispatch: unexpected nil node at depth %d", depth)
```

### 2d. Programming Errors *

Errors that can only result from a developer mistake during wiring, and
can never be triggered by end-user CLI input. Must carry a
`programmingError:` comment.

```go
// programmingError: duplicate mode registration can only be caused by
// incorrect wiring in the composition root, never by end-user input.
// exempt from i18n - English only.
return fmt.Errorf("ui: display mode %q is already registered", name)
```

### 2e. Test Code - no comment needed

Files ending in `_test.go` have no i18n obligation. Use `fmt.*`,
`errors.*`, and string literals freely.

### 2f. Demo and Example Programs - no comment needed

Programs under `examples/`, `demo/`, or similar are not production code
and have no i18n obligation.

---

## 3. Applying i18n - the Decision Test

Ask these questions in order, stopping at the first No:

1. Is this production code (not `_test.go`, not a demo program)?
2. Is the string natural-language prose a CLI user reads as a message?
3. Could it reasonably differ between languages?
4. Is it outside all the exemptions in section 2?

All four Yes -> i18n required; complete the workflow in section 6.
Any No -> use `fmt` / `errors` directly; apply the appropriate
exemption comment from section 2 if one is required.

---

## 4. The Two Key Types

### `UnderlyingTemplData` - one entry per message

```go
type UnderlyingTemplData struct {
    // unique go-i18n message ID
    MessageID   string
    // PascalCase base name for all generated identifiers
    Seed        string
    // controls what code is generated
    TypeName    enums.UnderlyingType
    // short human summary; used as struct doc comment
    Description string
    // longer narrative; word-wrapped at 80 chars in banner
    Story       string
    // go-i18n translation string; may contain {{.Token}}
    Other       string
    // variable fields for dynamic messages; empty for static
    Fields      []UnderlyingField
    // optional output file prefix; leave empty for defaults
    File        string
}
```

### `UnderlyingField` - one entry per variable token

```go
type UnderlyingField struct {
    // must match a {{.Note}} token in Other exactly
    Note   string
    // valid Go type: "string", "int", "uint", "error"
    GoType string
    // doc comment for the generated field; emits 🔥 TODO if empty
    Tale   string
}
```

---

## 5. The `UnderlyingType` Enum - Full Reference

Choose the `TypeName` that matches what you need:

| TypeName | Use when |
| --- | --- |
| `UnderlyingTypeStaticCobra` | Static cobra command/flag description, no variables |
| `UnderlyingTypeDynamicCobra` | Cobra description with `{{.Token}}` variables |
| `UnderlyingTypeStaticGeneral` | Static non-error user-facing message |
| `UnderlyingTypeDynamicGeneral` | Non-error message with `{{.Token}}` variables |
| `UnderlyingTypeStaticError` | Static error, no variables, no wrapping |
| `UnderlyingTypeSentinelError` | Static sentinel error to be wrapped by outer errors; generates `ErrXxx` |
| `UnderlyingTypeStaticErrorWrapper` | Static error wrapping another for the chain only; wrapped message does NOT appear in translated output |
| `UnderlyingTypeStaticErrorWrapperMsg` | Static error wrapping another AND showing `{{.Wrapped}}` in translated output |
| `UnderlyingTypeDynamicError` | Dynamic error with variables, no wrapping |
| `UnderlyingTypeDynamicErrorWrapper` | Dynamic error with variables that also wraps another error |

Import path:

```go
import "github.com/snivilised/li18ngo/locale/enums"
```

---

## 6. MessageID and Seed Conventions

- `MessageID` must be unique across the entire `Underliers` map. Scan
  the map before inserting - lingo rejects duplicates but catching them
  early avoids a round-trip.
- For non-errors: use a `kebab-case` slug, e.g. `"using-config-file"`.
- For errors: append the kind suffix, e.g.
  `"path-not-found.dynamic-error"` or `"config-missing.static-error"`.
- `Seed` must also be unique across the entire map. lingo derives every
  generated identifier from `Seed`; a duplicate produces duplicate Go
  declarations and the build fails. Scan the map before inserting.
- `Seed` is `PascalCase`, e.g. `"PathNotFound"`. Generated identifiers:
  `PathNotFoundTemplData`, `PathNotFoundError`, `ErrPathNotFound`,
  `NewPathNotFoundError`.

---

## 7. Fields Rules

- `Fields` must be **empty** for all static `TypeName` values.
- `Fields` must be **non-empty** for all dynamic `TypeName` values.
- Every `Note` in `Fields` must have a matching `{{.Note}}` token in
  `Other` and vice versa - lingo validates this and refuses to generate
  on any mismatch.
- For wrapper types, exactly one `Fields` entry must have
  `GoType: "error"` and `Note: "Wrapped"`. No other entry may use
  `GoType: "error"`.
- `GoType` must be a valid native Go type: `"string"`, `"int"`,
  `"uint"`, `"error"`.
- Always populate `Tale`. An empty `Tale` causes lingo to emit a
  🔥 TODO in the generated doc comment.

---

## 8. The `File` Field

Leave `File` empty in almost all cases. lingo writes to the default
output file for the message kind:

- Cobra kinds -> `messages-cobra-auto.go`
- General kinds -> `messages-general-auto.go`
- Error kinds -> `messages-errors-auto.go`

Set `File` only when output must go to a custom file, e.g.
`File: "system-automation"` produces `system-automation-errors-auto.go`.

---

## 9. Agent Workflow - Adding a New Message or Error

This workflow is mandatory whenever i18n is required (section 3).
Complete all four steps before writing any call site. No step may be
skipped, reordered, or deferred.

### Step 1 - Add entries to `Underliers`

Open `underlying-templ-data.go` and add one `UnderlyingTemplData` entry
per new message or error. Use sections 4-8 to determine correct field
values. Before inserting, scan the existing map and confirm that both
`MessageID` and `Seed` are unique across the entire map. See section 11
for complete worked examples of each entry kind.

### Step 2 - Delete stale auto files

Before running lingo, delete every `*-auto.go` file that lingo owns.
This prevents stale generated identifiers surviving into the new output
and avoids duplicate-declaration compile errors.

**Deletion scope - absolute and non-negotiable:** only files inside the
`locale/` directory are ever candidates for deletion. No file outside
`locale/` may be deleted under any circumstances.

List the directory first to confirm which auto files are present, then
delete only those that appear in the listing:

```bash
ls locale/*-auto.go

# Default auto files - delete whichever are present
rm -f locale/messages-cobra-auto.go
rm -f locale/messages-general-auto.go
rm -f locale/messages-errors-auto.go

# Custom-file auto files - only if an entry uses File: "..."
# e.g. rm -f locale/system-automation-errors-auto.go
```

### Step 3 - Run lingo

```bash
go run ./cmd/lingo/...
```

Lingo regenerates all `*-auto.go` files from the current `Underliers`
map. If lingo exits with an error (duplicate `MessageID`, mismatched
`{{.Token}}` vs `Fields`, empty `Tale`, etc.), fix the entry in
`underlying-templ-data.go` and re-run. Do not proceed to step 4 until
lingo completes without error.

### Step 4 - Write call sites using generated types

Only after lingo has succeeded may call sites be written. Generated
identifiers follow directly from `Seed`:

| What you need | Generated identifier |
| --- | --- |
| Template data struct | `locale.<Seed>TemplData{}` |
| Error constructor | `locale.New<Seed>Error(...)` |
| Sentinel error value | `locale.Err<Seed>` |

**User-facing text** - static struct or with fields populated:

```go
li18ngo.Text(locale.TraversalCompleteTemplData{})
li18ngo.Text(locale.UsingConfigFileTemplData{ConfigFileName: cfgPath})
```

**Returning an error:**

```go
return locale.NewPathNotFoundError(err, "Config", path)
```

**Checking or wrapping a sentinel:**

```go
if errors.Is(err, locale.ErrSomeSentinel) { ... }
```

### Workflow rules

- Complete all four steps before writing any call site. Never reference
  a generated type that does not yet exist on disk.
- If lingo fails, fix `underlying-templ-data.go` and re-run. Never
  work around a failure by hand-editing an auto file.
- No placeholder comments are left behind. The agent performs the full
  cycle; there is nothing left for the developer to do.

---

## 10. Lists of Valid Values in Error Messages

When an error message must include a list of valid values, go-i18n
represents this as a single `string` field. The caller must pre-format
the list before passing it to the generated constructor:

```go
// ValidModes must be pre-formatted before passing to the constructor.
strings.Join(registeredModes(), ", ")
```

The `Fields` entry uses `GoType: "string"` and `Tale` should state that
the caller is responsible for formatting.

Note: lingo does not yet handle `[]string` fields natively. Until it
does, always pre-format with `strings.Join` and pass a `string`.

---

## 11. Worked Examples

### Static non-error message

```go
"localisation.test": {
    MessageID:   "localisation.test",
    Seed:        "Localisation",
    TypeName:    enums.UnderlyingTypeStaticGeneral,
    Description: "Localisation",
    Story:       "A test message for localisation.",
    Other:       "localisation",
},
```

### Dynamic non-error message

```go
"using-config-file": {
    MessageID:   "using-config-file",
    Seed:        "UsingConfigFile",
    TypeName:    enums.UnderlyingTypeDynamicGeneral,
    Description: "Message to indicate which config is being used",
    Story: "UsingConfigFile is printed on startup to indicate" +
        " which configuration file has been loaded.",
    Other: "Using config file: '{{.ConfigFileName}}'",
    Fields: []UnderlyingField{
        {
            Note:   "ConfigFileName",
            GoType: "string",
            Tale:   "is the name of the config file being used",
        },
    },
},
```

### Dynamic error, no wrapping

```go
"path-not-found.dynamic-error": {
    MessageID:   "path-not-found.dynamic-error",
    Seed:        "PathNotFound",
    TypeName:    enums.UnderlyingTypeDynamicError,
    Description: "Directory or file path does not exist",
    Story: "PathNotFoundError is used when a path does not exist.",
    Other: "{{.Name}} path not found ({{.Path}})",
    Fields: []UnderlyingField{
        {
            Note:   "Name",
            GoType: "string",
            Tale:   "is the label for the missing path (e.g. 'Config')",
        },
        {
            Note:   "Path",
            GoType: "string",
            Tale:   "is the actual path that was not found",
        },
    },
},
```

### Static error wrapping another (message shows wrapped text)

```go
"third-party.error-wrapper-msg": {
    MessageID:   "third-party.error-wrapper-msg",
    Seed:        "ThirdPartyWrapper",
    TypeName:    enums.UnderlyingTypeStaticErrorWrapperMsg,
    Description: "Wrapper for third-party errors",
    Story: "ThirdPartyErrorWrapper wraps errors from third-party " +
        "libraries.",
    Other: "Third party error occurred: '{{.Wrapped}}'",
    Fields: []UnderlyingField{
        {
            Note:   "Wrapped",
            GoType: "error",
            Tale: "is the original error from the third-party " +
                "library that is being wrapped",
        },
    },
},
```

---

## 12. Common Mistakes to Avoid

| Mistake | Correct behaviour |
| --- | --- |
| Applying i18n to pure formatting (Sprintf of computed values) | Pure formatting is exempt - see section 2a |
| Applying i18n to fixed-token content (flag names, identifiers) | Fixed tokens are exempt; embed them as literals in `Other` - see section 2b |
| Applying i18n in test files | Test code is fully exempt - see section 2e |
| Applying i18n in demo or example programs | Demo programs are exempt - see section 2f |
| Using `fmt.Errorf` for user-facing errors without an exemption | Complete the full lingo cycle (steps 1-4 in section 9) |
| Using `fmt.Println` for user-facing text without an exemption | Complete the full lingo cycle (steps 1-4 in section 9) |
| Omitting `internalError:` on an internal error | Always comment internal errors to confirm they are not user-facing |
| Omitting `programmingError:` on a programming error | Always document why the error is exempt - see section 2d |
| Hand-writing generated structs or error types | Only edit `underlying-templ-data.go`; run lingo to generate |
| `Fields` non-empty on a static `TypeName` | Static types must have empty `Fields` |
| `Fields` empty on a dynamic `TypeName` | Dynamic types must have non-empty `Fields` |
| `Note` in `Fields` not matching a `{{.Token}}` in `Other` | Every `Note` must have a matching token and vice versa |
| More than one `Fields` entry with `GoType: "error"` | Exactly one error field permitted, named `"Wrapped"` |
| `GoType: "error"` field not named `"Wrapped"` | The wrapped error field must always be named `"Wrapped"` |
| Omitting `Tale` | Always provide `Tale`; empty emits 🔥 TODO in generated code |
| Duplicate `MessageID` across the map | lingo refuses to generate on duplicate IDs |
| Duplicate `Seed` across the map | Produces duplicate Go declarations; the build fails |
| Writing a call site before lingo has run successfully | Always run lingo and confirm it exits cleanly first |
| Leaving placeholder comments for the developer to resolve | The agent completes the full cycle; no placeholders are left behind |
| Skipping the auto-file deletion before re-running lingo | Always delete `locale/*-auto.go` before re-running |
| Deleting any file outside the `locale/` directory | Deletion scope is `locale/` only - absolute and non-negotiable |
| Hand-editing an auto file to work around a lingo error | Fix `underlying-templ-data.go` and re-run; never touch auto files |
| Passing a `[]string` directly to a generated constructor | Pre-format with `strings.Join`; lingo generates a `string` field |
