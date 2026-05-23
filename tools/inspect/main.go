// inspect: Extract exported declarations from Go modules in the module cache.
//
// Usage:
//
//	inspect [flags] <module-path>[@version]
//
// Examples:
//
//	inspect github.com/spf13/cobra@v1.8.0
//	inspect --funcs --types github.com/spf13/cobra
//	inspect --interfaces golang.org/x/net/html
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// --- flags ---

type config struct {
	showVar       bool
	showInterface bool
	showType      bool
	showConst     bool
	showFunc      bool
}

func (c config) showAll() bool {
	return !c.showVar && !c.showInterface && !c.showType && !c.showConst && !c.showFunc
}
func (c config) wantVar() bool       { return c.showAll() || c.showVar }
func (c config) wantInterface() bool { return c.showAll() || c.showInterface }
func (c config) wantType() bool      { return c.showAll() || c.showType }
func (c config) wantConst() bool     { return c.showAll() || c.showConst }
func (c config) wantFunc() bool      { return c.showAll() || c.showFunc }

// --- entry point ---

func main() {
	cfg := config{}
	flag.BoolVar(&cfg.showVar, "vars", false, "show exported var declarations")
	flag.BoolVar(&cfg.showInterface, "interfaces", false, "show exported interface declarations")
	flag.BoolVar(&cfg.showType, "types", false, "show exported type declarations")
	flag.BoolVar(&cfg.showConst, "consts", false, "show exported const declarations")
	flag.BoolVar(&cfg.showFunc, "funcs", false, "show exported func/method signatures")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}

	moduleArg := flag.Arg(0)
	modPath, version, hasVersion := strings.Cut(moduleArg, "@")

	// Resolve the cache directory for the module.
	cacheDir, err := resolveModuleDir(modPath, version, hasVersion)
	if err != nil {
		fatalf("error resolving module: %v", err)
	}

	// Walk all .go files (non-test) and collect declarations.
	if err := walkAndPrint(cacheDir, modPath, cfg); err != nil {
		fatalf("error processing module: %v", err)
	}
}

// --- module cache resolution ---

// resolveModuleDir locates the module directory inside GOMODCACHE.
// If no version is supplied it picks the highest available version.
func resolveModuleDir(modPath, version string, hasVersion bool) (string, error) {
	gomodcache, err := goModCache()
	if err != nil {
		return "", fmt.Errorf("cannot determine GOMODCACHE: %w", err)
	}
	if gomodcache == "" {
		return "", fmt.Errorf("GOMODCACHE is empty")
	}

	// Module paths inside the cache use a lower-cased, escaped encoding
	// (e.g. capital letters → '!'+lowercase).  We replicate that here.
	escapedMod := escapePath(modPath)

	if hasVersion && version != "" {
		escapedVer := escapePath(version)
		dir := filepath.Join(gomodcache, escapedMod+"@"+escapedVer)
		info, statErr := os.Stat(dir)
		if statErr == nil && info.IsDir() {
			return dir, nil
		}
		return "", fmt.Errorf("module %s@%s not found in cache (%s)", modPath, version, dir)
	}

	// No version: find all available versions in the cache.
	prefix := filepath.Join(gomodcache, escapedMod+"@")
	parent := filepath.Dir(prefix)
	base := filepath.Base(escapedMod) + "@"

	entries, err := os.ReadDir(parent)
	if err != nil {
		return "", fmt.Errorf("cannot read cache dir %s: %w", parent, err)
	}

	var matches []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), base) {
			matches = append(matches, filepath.Join(parent, e.Name()))
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no cached versions of %q found under %s", modPath, gomodcache)
	}
	sort.Strings(matches)
	chosen := matches[len(matches)-1]
	fmt.Fprintf(os.Stderr, "# using %s\n", filepath.Base(chosen))
	return chosen, nil
}

// goModCache returns the GOMODCACHE path. The subcommand args are literals
// (not user-supplied) so the gosec G204 subprocess warning does not apply.
func goModCache() (string, error) {
	out, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// escapePath converts a module path to the cache's filesystem encoding.
// Capital letters become '!' followed by the lowercase letter.
func escapePath(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			b.WriteByte('!')
			b.WriteRune(r + 32)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// --- AST walking ---

func walkAndPrint(root, modPath string, cfg config) error {
	fset := token.NewFileSet()

	// Group files by package directory so we parse each package once.
	type pkgDir struct {
		importPath string
		dir        string
	}
	var dirs []pkgDir

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() {
			// Derive the import path from the directory relative to the module root.
			rel, _ := filepath.Rel(root, path)
			if rel == "." {
				dirs = append(dirs, pkgDir{modPath, path})
			} else {
				dirs = append(dirs, pkgDir{modPath + "/" + filepath.ToSlash(rel), path})
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, pd := range dirs {
		// ParseDir is deprecated in favour of golang.org/x/tools/go/packages, but
		// that would add an external dependency. For read-only inspection of cached
		// modules without build-tag filtering, ParseDir is sufficient.
		//nolint:staticcheck
		pkgs, err := parser.ParseDir(fset, pd.dir, func(fi os.FileInfo) bool {
			// Skip test files and files that start with '_' or '.'.
			name := fi.Name()
			if strings.HasSuffix(name, "_test.go") {
				return false
			}
			if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
				return false
			}
			return true
		}, parser.SkipObjectResolution)
		if err != nil {
			// Non-Go or broken directory; just skip.
			continue
		}

		for _, pkg := range pkgs {
			printPackage(fset, pd.importPath, pkg, cfg)
		}
	}
	return nil
}

// printPackage formats and prints all matching exported declarations from pkg.
//
//nolint:staticcheck // ast.Package deprecated alongside parser.ParseDir; no external deps introduced
func printPackage(fset *token.FileSet, importPath string, pkg *ast.Package, cfg config) {
	// Collect all declarations across files.
	var out strings.Builder

	// Sort files for deterministic output.
	filenames := make([]string, 0, len(pkg.Files))
	for name := range pkg.Files {
		filenames = append(filenames, name)
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		file := pkg.Files[filename]
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				handleGenDecl(fset, d, cfg, &out)
			case *ast.FuncDecl:
				if cfg.wantFunc() && ast.IsExported(d.Name.Name) {
					out.WriteString(funcSignature(fset, d))
					out.WriteByte('\n')
				}
			}
		}
	}

	if out.Len() > 0 {
		fmt.Printf("// package %s (%s)\n", pkg.Name, importPath)
		fmt.Print(out.String())
		fmt.Println()
	}
}

// --- declaration formatters ---

func handleGenDecl(fset *token.FileSet, d *ast.GenDecl, cfg config, out *strings.Builder) {
	switch d.Tok { //nolint:exhaustive // only VAR, CONST, TYPE are relevant declaration tokens
	case token.VAR:
		if !cfg.wantVar() {
			return
		}
		for _, spec := range d.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if !ast.IsExported(name.Name) {
					continue
				}
				var typeStr string
				if vs.Type != nil {
					typeStr = exprString(fset, vs.Type)
				} else if vs.Values != nil && i < len(vs.Values) {
					// Infer type from value via a best-effort stringification.
					typeStr = inferredType(vs.Values[i])
				}
				if typeStr != "" {
					fmt.Fprintf(out, "var %s %s\n", name.Name, typeStr)
				} else {
					fmt.Fprintf(out, "var %s\n", name.Name)
				}
			}
		}

	case token.CONST:
		if !cfg.wantConst() {
			return
		}
		for _, spec := range d.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if !ast.IsExported(name.Name) {
					continue
				}
				var parts []string
				parts = append(parts, name.Name)
				if vs.Type != nil {
					parts = append(parts, exprString(fset, vs.Type))
				}
				if i < len(vs.Values) {
					parts = append(parts, "= "+exprString(fset, vs.Values[i]))
				}
				fmt.Fprintf(out, "const %s\n", strings.Join(parts, " "))
			}
		}

	case token.TYPE:
		for _, spec := range d.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || !ast.IsExported(ts.Name.Name) {
				continue
			}
			switch t := ts.Type.(type) {
			case *ast.InterfaceType:
				if cfg.wantInterface() {
					fmt.Fprintf(out, "%s\n", interfaceDecl(fset, ts.Name.Name, t))
				}
			default:
				if cfg.wantType() {
					if ts.TypeParams != nil {
						fmt.Fprintf(out, "type %s%s %s\n",
							ts.Name.Name,
							typeParamsString(fset, ts.TypeParams),
							exprString(fset, ts.Type))
					} else {
						fmt.Fprintf(out, "type %s %s\n", ts.Name.Name, exprString(fset, ts.Type))
					}
				}
			}
		}
	default:
		// token.IMPORT, token.PACKAGE, and all non-decl tokens are irrelevant.
	}
}

func funcSignature(fset *token.FileSet, d *ast.FuncDecl) string {
	var sb strings.Builder
	sb.WriteString("func ")
	if d.Recv != nil && len(d.Recv.List) > 0 {
		sb.WriteString("(")
		sb.WriteString(fieldListString(fset, d.Recv))
		sb.WriteString(") ")
	}
	sb.WriteString(d.Name.Name)
	if d.Type.TypeParams != nil {
		sb.WriteString(typeParamsString(fset, d.Type.TypeParams))
	}
	sb.WriteString("(")
	if d.Type.Params != nil {
		sb.WriteString(fieldListString(fset, d.Type.Params))
	}
	sb.WriteString(")")
	if d.Type.Results != nil && len(d.Type.Results.List) > 0 {
		results := fieldListString(fset, d.Type.Results)
		if len(d.Type.Results.List) > 1 || (len(d.Type.Results.List) == 1 && len(d.Type.Results.List[0].Names) > 0) {
			sb.WriteString(" (")
			sb.WriteString(results)
			sb.WriteString(")")
		} else {
			sb.WriteString(" ")
			sb.WriteString(results)
		}
	}
	return sb.String()
}

func interfaceDecl(fset *token.FileSet, name string, iface *ast.InterfaceType) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "type %s interface {\n", name)
	if iface.Methods != nil {
		for _, method := range iface.Methods.List {
			switch mt := method.Type.(type) {
			case *ast.FuncType:
				// Named method.
				for _, mname := range method.Names {
					sb.WriteString("\t")
					sb.WriteString(mname.Name)
					sb.WriteString("(")
					if mt.Params != nil {
						sb.WriteString(fieldListString(fset, mt.Params))
					}
					sb.WriteString(")")
					if mt.Results != nil && len(mt.Results.List) > 0 {
						results := fieldListString(fset, mt.Results)
						if len(mt.Results.List) > 1 || (len(mt.Results.List) == 1 && len(mt.Results.List[0].Names) > 0) {
							sb.WriteString(" (")
							sb.WriteString(results)
							sb.WriteString(")")
						} else {
							sb.WriteString(" ")
							sb.WriteString(results)
						}
					}
					sb.WriteByte('\n')
				}
			default:
				// Embedded type or type constraint.
				sb.WriteString("\t")
				sb.WriteString(exprString(fset, method.Type))
				sb.WriteByte('\n')
			}
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// --- expression stringifiers ---

func exprString(fset *token.FileSet, expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprString(fset, e.X)
	case *ast.SelectorExpr:
		return exprString(fset, e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		if e.Len == nil {
			return "[]" + exprString(fset, e.Elt)
		}
		return "[" + exprString(fset, e.Len) + "]" + exprString(fset, e.Elt)
	case *ast.MapType:
		return "map[" + exprString(fset, e.Key) + "]" + exprString(fset, e.Value)
	case *ast.ChanType:
		switch e.Dir {
		case ast.SEND:
			return "chan<- " + exprString(fset, e.Value)
		case ast.RECV:
			return "<-chan " + exprString(fset, e.Value)
		default:
			return "chan " + exprString(fset, e.Value)
		}
	case *ast.FuncType:
		var sb strings.Builder
		sb.WriteString("func(")
		if e.Params != nil {
			sb.WriteString(fieldListString(fset, e.Params))
		}
		sb.WriteString(")")
		if e.Results != nil && len(e.Results.List) > 0 {
			sb.WriteString(" ")
			sb.WriteString(fieldListString(fset, e.Results))
		}
		return sb.String()
	case *ast.InterfaceType:
		if e.Methods == nil || len(e.Methods.List) == 0 {
			return "any"
		}
		return "interface{ ... }"
	case *ast.StructType:
		return "struct{ ... }"
	case *ast.Ellipsis:
		return "..." + exprString(fset, e.Elt)
	case *ast.IndexExpr:
		return exprString(fset, e.X) + "[" + exprString(fset, e.Index) + "]"
	case *ast.IndexListExpr:
		// Generic instantiation with multiple type parameters.
		indices := make([]string, len(e.Indices))
		for i, idx := range e.Indices {
			indices[i] = exprString(fset, idx)
		}
		return exprString(fset, e.X) + "[" + strings.Join(indices, ", ") + "]"
	case *ast.BinaryExpr:
		return exprString(fset, e.X) + " " + e.Op.String() + " " + exprString(fset, e.Y)
	case *ast.UnaryExpr:
		return e.Op.String() + exprString(fset, e.X)
	case *ast.BasicLit:
		return e.Value
	case *ast.ParenExpr:
		return "(" + exprString(fset, e.X) + ")"
	case *ast.CompositeLit:
		return exprString(fset, e.Type) + "{...}"
	case *ast.CallExpr:
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = exprString(fset, a)
		}
		return exprString(fset, e.Fun) + "(" + strings.Join(args, ", ") + ")"
	default:
		return fmt.Sprintf("<%T>", expr)
	}
}

func fieldListString(fset *token.FileSet, fl *ast.FieldList) string {
	if fl == nil {
		return ""
	}
	parts := make([]string, 0, len(fl.List))
	for _, field := range fl.List {
		typeStr := exprString(fset, field.Type)
		if len(field.Names) == 0 {
			parts = append(parts, typeStr)
		} else {
			names := make([]string, len(field.Names))
			for i, n := range field.Names {
				names[i] = n.Name
			}
			parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
		}
	}
	return strings.Join(parts, ", ")
}

func typeParamsString(fset *token.FileSet, tparams *ast.FieldList) string {
	return "[" + fieldListString(fset, tparams) + "]"
}

// inferredType attempts to derive a human-readable type from a value expression.
// It's intentionally shallow - just enough for common const/var patterns.
func inferredType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind { //nolint:exhaustive // only INT, FLOAT, STRING, CHAR map to useful type names
		case token.INT:
			return "int"
		case token.FLOAT:
			return "float64"
		case token.STRING:
			return "string"
		case token.CHAR:
			return "rune"
		default:
			// token.IMAG and others - not useful for type inference.
		}
	case *ast.Ident:
		switch e.Name {
		case "true", "false":
			return "bool"
		case "nil":
			return ""
		}
	case *ast.CallExpr:
		if id, ok := e.Fun.(*ast.Ident); ok {
			// e.g. errors.New(...) - just return the func name as a hint.
			return id.Name + "(...)"
		}
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
			return exprString(nil, sel)
		}
	case *ast.UnaryExpr:
		return inferredType(e.X)
	}
	return ""
}

// --- helpers ---

func usage() {
	fmt.Fprintln(os.Stderr, `inspect - extract exported declarations from a cached Go module

Usage:
  inspect [flags] <module-path>[@version]

Flags:
  --vars        show exported var declarations
  --interfaces  show exported interface types (with method signatures)
  --types       show exported type declarations
  --consts      show exported const declarations
  --funcs       show exported function/method signatures

Without any flags, all declaration kinds are shown.

Examples:
  inspect github.com/spf13/cobra@v1.8.0
  inspect --funcs --types github.com/spf13/cobra
  inspect --interfaces golang.org/x/net/html`)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "inspect: "+format+"\n", args...)
	os.Exit(1)
}
