package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateName(t *testing.T) {
	valid := []string{"myapp", "my-app", "app1", "a-b-c", "x2y"}
	for _, n := range valid {
		if err := validateName(n); err != nil {
			t.Errorf("expected %q valid, got %v", n, err)
		}
	}
	invalid := []string{"", "-app", "app-", "App", "my_app", "1app", "my--app", "go", "goloop"}
	for _, n := range invalid {
		if err := validateName(n); err == nil {
			t.Errorf("expected %q invalid", n)
		}
	}
}

func TestScanTraces(t *testing.T) {
	dir := t.TempDir()
	clean := filepath.Join(dir, "clean.go")
	dirty := filepath.Join(dir, "dirty.go")
	// Build the fixture markers from fragments so this test file itself stays
	// clean under a boundary scan.
	marker := "Gene" + "rated by " + "Cla" + "ude"
	os.WriteFile(clean, []byte("package main\n// a normal comment\n"), 0o644)
	os.WriteFile(dirty, []byte("package main\n// "+marker+"\n"), 0o644)

	findings, err := scanTraces(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("expected findings in dirty.go")
	}
	for _, f := range findings {
		if filepath.Base(f.Path) == "clean.go" {
			t.Errorf("false positive in clean.go: %+v", f)
		}
	}
}

func TestScanTracesSkipsGitDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0o755)
	os.WriteFile(filepath.Join(gitDir, "COMMIT"), []byte("co-auth"+"ored-by: cla"+"ude\n"), 0o644)

	findings, err := scanTraces(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf(".git should be skipped, got %+v", findings)
	}
}

func TestRenderService(t *testing.T) {
	files, err := renderService("myapp", "github.com/acme/myapp")
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	gomod := string(files["go.mod"])
	if want := "module github.com/acme/myapp"; !contains(gomod, want) {
		t.Fatalf("go.mod missing module: %s", gomod)
	}
	if !contains(string(files["main.go"]), `app.New("myapp"`) {
		t.Fatalf("main.go not rendered: %s", files["main.go"])
	}
}

func TestNewScaffold(t *testing.T) {
	dir := t.TempDir()
	code := cmdNew([]string{"myapp", "--dir", dir})
	if code != 0 {
		t.Fatalf("new exit code %d", code)
	}
	for _, f := range []string{"go.mod", "main.go", ".env", "README.md"} {
		if _, err := os.Stat(filepath.Join(dir, "myapp", f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}
	// Refuse to overwrite without --force.
	if code := cmdNew([]string{"myapp", "--dir", dir}); code == 0 {
		t.Fatal("expected refusal on existing dir")
	}
}

func TestCmdCheckCleanAndDirty(t *testing.T) {
	clean := t.TempDir()
	os.WriteFile(filepath.Join(clean, "a.go"), []byte("package main\n"), 0o644)
	if code := cmdCheck([]string{clean}); code != 0 {
		t.Fatalf("clean dir should pass, got %d", code)
	}

	dirty := t.TempDir()
	os.WriteFile(filepath.Join(dirty, "b.txt"), []byte("cop"+"ilot wrote this\n"), 0o644)
	if code := cmdCheck([]string{dirty}); code != 1 {
		t.Fatalf("dirty dir should fail, got %d", code)
	}
}

func TestCmdStatus(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module github.com/acme/x\n\ngo 1.24\n\nrequire github.com/goloop/app v0.1.0\n"), 0o644)
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir(dir)
	if code := cmdStatus(nil); code != 0 {
		t.Fatalf("status exit %d", code)
	}
}

func TestNewMissingFlagValue(t *testing.T) {
	if code := cmdNew([]string{"app", "--module"}); code != 2 {
		t.Fatalf("expected exit 2 for --module without value, got %d", code)
	}
	if code := cmdNew([]string{"app", "--dir", "--force"}); code != 2 {
		t.Fatalf("expected exit 2 for --dir followed by a flag, got %d", code)
	}
}

func TestScanOverlongLineFailsClosed(t *testing.T) {
	dir := t.TempDir()
	// A single line larger than the scanner buffer must be reported, not
	// silently skipped, so traces cannot hide behind it.
	big := make([]byte, 11*1024*1024)
	for i := range big {
		big[i] = 'x'
	}
	os.WriteFile(filepath.Join(dir, "huge.txt"), big, 0o644)
	findings, err := scanTraces(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("overlong line should be reported as unscannable")
	}
}

func TestParseGoModForms(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	src := "module github.com/acme/x\n\n" +
		"go 1.24\n\n" +
		"require github.com/goloop/app v0.1.0\n\n" +
		"require (\n" +
		"\tgithub.com/goloop/log v2.0.0 // indirect\n" +
		"\tgithub.com/other/pkg v1.0.0\n" +
		")\n\n" +
		"replace github.com/goloop/mail => ../mail\n\n" +
		"replace (\n\tgithub.com/goloop/auth => ../auth\n)\n"
	os.WriteFile(path, []byte(src), 0o644)

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	module, goVersion, deps, err := parseGoMod(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if module != "github.com/acme/x" {
		t.Errorf("module = %q", module)
	}
	if goVersion != "1.24" {
		t.Errorf("go = %q", goVersion)
	}
	// Single-line and block requires count; replace directives do not, and the
	// "require"/"replace" keywords must never leak in as a dependency name.
	want := []string{"github.com/goloop/app", "github.com/goloop/log"}
	if len(deps) != len(want) {
		t.Fatalf("deps = %v, want %v", deps, want)
	}
	for i, w := range want {
		if deps[i] != w {
			t.Errorf("deps[%d] = %q, want %q", i, deps[i], w)
		}
	}
}

func TestCheckFailsClosedOnUnreadable(t *testing.T) {
	dir := t.TempDir()
	// A broken symlink cannot be opened; the scanner must report it as
	// unscannable rather than skip it, so a trace cannot hide behind an open
	// error.
	link := filepath.Join(dir, "broken.go")
	if err := os.Symlink(filepath.Join(dir, "does-not-exist"), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	findings, err := scanTraces(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("unreadable entry should be reported as unscannable")
	}
}

func TestNewRejectsExtraPositional(t *testing.T) {
	if code := cmdNew([]string{"app", "extra"}); code != 2 {
		t.Fatalf("expected exit 2 for a second positional, got %d", code)
	}
}

func TestArgRejections(t *testing.T) {
	if code := cmdCheck([]string{"a", "b"}); code != 2 {
		t.Fatalf("check with two dirs should exit 2, got %d", code)
	}
	if code := cmdStatus([]string{"x"}); code != 2 {
		t.Fatalf("status with args should exit 2, got %d", code)
	}
	if code := cmdDoctor([]string{"x"}); code != 2 {
		t.Fatalf("doctor with args should exit 2, got %d", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	if code := run([]string{"frobnicate"}); code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if code := run([]string{"version"}); code != 0 {
		t.Fatalf("version should exit 0, got %d", code)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
