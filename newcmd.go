package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// serviceFiles are the templates rendered by `goloop new`. Keys are relative
// paths; values are text/template sources.
var serviceFiles = map[string]string{
	"go.mod": `module {{.Module}}

go 1.24

require github.com/goloop/app v0.1.0
`,
	"main.go": `package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/goloop/app"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("hello from {{.Name}}\n"))
	})

	a := app.New("{{.Name}}",
		app.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),
		app.WithShutdownTimeout(10*time.Second),
	)
	a.Use(app.HTTPServer("http", &http.Server{Addr: ":8080", Handler: mux}))

	return a.Run(context.Background())
}
`,
	".env": `# {{.Name}} configuration
ADDR=:8080
`,
	".gitignore": `/bin/
/dist/
*.out
`,
	"README.md": `# {{.Name}}

A goloop-first service scaffolded with ` + "`goloop new`" + `.

## Run

` + "```bash\ngo mod tidy\ngo run .\n```" + `
`,
}

// serviceData is the template context.
type serviceData struct {
	Name   string
	Module string
}

// renderService renders all service files for the project.
func renderService(name, module string) (map[string][]byte, error) {
	data := serviceData{Name: name, Module: module}
	out := make(map[string][]byte, len(serviceFiles))
	for path, src := range serviceFiles {
		tmpl, err := template.New(path).Parse(src)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, err
		}
		out[path] = buf.Bytes()
	}
	return out, nil
}

// cmdNew scaffolds a minimal goloop-first service.
func cmdNew(args []string) int {
	var force bool
	var module, dir, name string

	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "--force":
			force = true
		case a == "--module":
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "-") {
				fmt.Fprintln(os.Stderr, "new: --module requires a value")
				return 2
			}
			module = args[i]
		case a == "--dir":
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "-") {
				fmt.Fprintln(os.Stderr, "new: --dir requires a value")
				return 2
			}
			dir = args[i]
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "new: unknown flag %q\n", a)
			return 2
		default:
			if name != "" {
				fmt.Fprintf(os.Stderr, "new: unexpected argument %q (name already set to %q)\n", a, name)
				return 2
			}
			name = a
		}
	}

	if err := validateName(name); err != nil {
		fmt.Fprintf(os.Stderr, "new: %v\n", err)
		return 2
	}
	if module == "" {
		module = "github.com/example/" + name
	}

	target := name
	if dir != "" {
		if err := validateDir(dir); err != nil {
			fmt.Fprintf(os.Stderr, "new: %v\n", err)
			return 2
		}
		target = filepath.Join(dir, name)
	}
	if _, err := os.Stat(target); err == nil && !force {
		fmt.Fprintf(os.Stderr, "new: %s already exists (use --force)\n", target)
		return 1
	}

	files, err := renderService(name, module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "new: %v\n", err)
		return 1
	}
	if err := writeFiles(target, files); err != nil {
		fmt.Fprintf(os.Stderr, "new: %v\n", err)
		return 1
	}

	// Print in a stable order.
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	fmt.Printf("created %s:\n", target)
	for _, p := range paths {
		fmt.Printf("  %s\n", p)
	}
	fmt.Printf("\nnext:\n  cd %s && go mod tidy && go run .\n", target)
	return 0
}

// writeFiles writes the rendered files under target, creating directories.
func writeFiles(target string, files map[string][]byte) error {
	for rel, content := range files {
		full := filepath.Join(target, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}
