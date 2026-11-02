//go:build !windows

package rc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeProfiler(t *testing.T, name string) string {
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	script := `#!/bin/sh
if [ "$1" = "--version" ] && [ "$2" = "--json" ]; then
  [ -n "$FAKE_SCHEMAS" ] || { echo "fake 1"; exit 2; }
  echo "{\"name\":\"fake\",\"version\":\"1.0.0\",\"coreSchemas\":$FAKE_SCHEMAS}"
  exit 0
fi
echo "args: $*"
exit ${FAKE_EXIT:-0}
`
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestRootListsProfileAndReportsItsVersion(t *testing.T) {
	out, _, err := rc()
	if err != nil || !strings.Contains(out, "profile") {
		t.Fatalf("%v\n%s", err, out)
	}
	if out, _, _ = rc("--version"); out != "rc 1.2.3\n" {
		t.Fatalf("version output %q", out)
	}
}

func TestProfilePassesArgumentsToThePathExecutable(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "rc-profiler-ruby"))
	out, _, err := rc("profile", "ruby", "generate", "--name", "orders")
	if err != nil || out != "args: generate --name orders\n" {
		t.Fatalf("%v\n%q", err, out)
	}
}

func TestProfileFindsPublishedExecutableNamesAndEnvironmentOverrides(t *testing.T) {
	t.Setenv("PATH", withPath(fakeProfiler(t, "runtimeconditions-typescript-profiler"), fakeRuntime(t, "node"), fakeRuntime(t, "python3")))
	if out, _, err := rc("profile", "typescript", "--version"); err != nil || out != "args: --version\n" {
		t.Fatalf("%v\n%q", err, out)
	}
	dir := fakeProfiler(t, "custom")
	t.Setenv("RC_PYTHON_PROFILER_BIN", filepath.Join(dir, "custom"))
	if out, _, err := rc("profile", "python", "generate"); err != nil || out != "args: generate\n" {
		t.Fatalf("%v\n%q", err, out)
	}
}

func TestProfileReturnsTheProfilerExitStatus(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "rc-profiler-ruby"))
	t.Setenv("FAKE_EXIT", "3")
	var exit *ExitError
	if _, _, err := rc("profile", "ruby"); !errors.As(err, &exit) || exit.Code != 3 {
		t.Fatalf("exit status not propagated: %v", err)
	}
}

func TestProfileReportsAMissingOrInvalidProfiler(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, _, err := rc("profile", "ruby"); err == nil || !strings.Contains(err.Error(), "rc-profiler-ruby") {
		t.Fatalf("missing profiler not explained: %v", err)
	}
	if _, _, err := rc("profile", "../ruby"); err == nil || !strings.Contains(err.Error(), "invalid language") {
		t.Fatalf("path-like language accepted: %v", err)
	}
}

func TestProfileRunsWhenTheProfilerSupportsTheCoreSchema(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "rc-profiler-ruby"))
	t.Setenv("FAKE_SCHEMAS", `["0.2.0","`+CoreSchema+`"]`)
	out, errout, err := rc("profile", "ruby", "generate")
	if err != nil || out != "args: generate\n" || errout != "" {
		t.Fatalf("%v\n%q\n%q", err, out, errout)
	}
}

func TestProfileRefusesAProfilerForAnotherCoreSchema(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "rc-profiler-ruby"))
	t.Setenv("FAKE_SCHEMAS", `["0.2.0"]`)
	out, _, err := rc("profile", "ruby", "generate")
	if err == nil || !strings.Contains(err.Error(), "supports core schema 0.2.0, but this rc targets "+CoreSchema) || out != "" {
		t.Fatalf("incompatible profiler ran: %v %q", err, out)
	}
}

func TestProfileNotesAProfilerThatDoesNotReportItsSchemas(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "rc-profiler-ruby"))
	out, errout, err := rc("profile", "ruby", "generate")
	if err != nil || out != "args: generate\n" || !strings.Contains(errout, "does not report its core schemas") {
		t.Fatalf("%v\n%q\n%q", err, out, errout)
	}
}

func TestProfileDoesNotCheckVersionQueries(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "rc-profiler-ruby"))
	t.Setenv("FAKE_SCHEMAS", `["0.2.0"]`)
	if out, _, err := rc("profile", "ruby", "--version"); err != nil || out != "args: --version\n" {
		t.Fatalf("%v\n%q", err, out)
	}
}

func TestBindingsRunsInProcessWithAndWithoutALeadingBindings(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, args := range [][]string{{"bindings", "resolve", "--help"}, {"bindings", "bindings", "resolve", "--help"}} {
		if _, errout, err := rc(args...); err != nil || !strings.Contains(errout, "-packages") {
			t.Fatalf("%v: %v\n%s", args, err, errout)
		}
	}
	if _, _, err := rc("bindings", "nope"); err == nil || !strings.Contains(err.Error(), `unknown bindings command "nope"`) {
		t.Fatalf("unknown command accepted: %v", err)
	}
}

func goProject(t *testing.T) string {
	dir := t.TempDir()
	for name, body := range map[string]string{"go.mod": "module example.com/app\n\ngo 1.25\n", "main.go": "package main\n\nfunc main() {}\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestGoProfilerRunsInProcessAndKeepsItsOwnExitStatus(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var exit *ExitError
	out, _, err := rc("profile", "go", "--version", "--json")
	if err != nil || !strings.Contains(out, `"name":"go-rc-profiler"`) || !strings.Contains(out, `"`+CoreSchema+`"`) {
		t.Fatalf("%v\n%q", err, out)
	}
	if _, errout, err := rc("profile", "go", "bogus"); !errors.As(err, &exit) || exit.Code != 1 || !strings.Contains(errout, `unknown command "bogus"`) {
		t.Fatalf("%v %q", err, errout)
	}
	if _, _, err := rc("profile", "go", "generate", "-nope"); !errors.As(err, &exit) || exit.Code != 2 {
		t.Fatalf("%v", err)
	}
}

func TestGoProfilerErrorsReachTheUserUnchangedWhenGoIsMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var exit *ExitError
	out, errout, err := rc("profile", "go", "generate", "--dir", goProject(t))
	if !errors.As(err, &exit) || exit.Code != 1 || out != "" || !strings.Contains(errout, "go command required, not found") || strings.Contains(errout, "rc ") {
		t.Fatalf("%v %q %q", err, out, errout)
	}
}

func TestGoProfilerOverrideRunsAnExternalBinary(t *testing.T) {
	dir := fakeProfiler(t, "external")
	t.Setenv("PATH", t.TempDir())
	t.Setenv("RC_GO_PROFILER_BIN", filepath.Join(dir, "external"))
	if out, _, err := rc("profile", "go", "generate"); err != nil || out != "args: generate\n" {
		t.Fatalf("%v\n%q", err, out)
	}
	t.Setenv("RC_GO_PROFILER_BIN", filepath.Join(dir, "absent"))
	if _, _, err := rc("profile", "go", "generate"); err == nil || !strings.Contains(err.Error(), "RC_GO_PROFILER_BIN") {
		t.Fatalf("missing override not explained: %v", err)
	}
}

func TestDoctorListsToolsAndFlagsIncompatibleProfilers(t *testing.T) {
	dir := fakeProfiler(t, "rc-profiler-typescript")
	if err := os.Symlink(filepath.Join(dir, "rc-profiler-typescript"), filepath.Join(dir, "rc-profiler-ruby")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("FAKE_SCHEMAS", `["`+CoreSchema+`"]`)
	out, _, err := rc("doctor")
	for _, want := range []string{"bindings    built in", "go          built in", "python      not found", "typescript", "1.0.0  compatible", "ruby"} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor output lacks %q:\n%s", want, out)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("FAKE_SCHEMAS", `["0.2.0"]`)
	out, _, err = rc("doctor")
	if err == nil || !strings.Contains(out, "incompatible: supports 0.2.0") {
		t.Fatalf("incompatible profiler not flagged: %v\n%s", err, out)
	}
	t.Setenv("FAKE_SCHEMAS", "")
	if out, _, err = rc("doctor"); err != nil || !strings.Contains(out, "unchecked: core schemas not reported") || !strings.Contains(out, "args: --version") {
		t.Fatalf("legacy profiler not described: %v\n%s", err, out)
	}
}

func fakeRuntime(t *testing.T, name string) string {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func withPath(dirs ...string) string { return strings.Join(dirs, string(os.PathListSeparator)) }

func TestPluginAndRuntimeMustBothBeInstalled(t *testing.T) {
	plugin := fakeProfiler(t, "rc-profiler-typescript")
	node := fakeRuntime(t, "node")
	notInstalled := "the typescript profiler is not installed: brew install runtimeconditions/tools/rc-profiler-typescript"
	noNode := "the typescript profiler needs Node.js, but node was not found on PATH: brew install node"
	cases := []struct {
		name, path string
		want       []string
		absent     []string
	}{
		{"plugin missing", node, []string{notInstalled}, []string{"Node.js"}},
		{"runtime missing", plugin, []string{noNode}, []string{"not installed"}},
		{"both missing", t.TempDir(), []string{notInstalled, noNode}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("PATH", c.path)
			var exit *ExitError
			out, _, err := rc("profile", "typescript", "generate")
			if err == nil || errors.As(err, &exit) || out != "" {
				t.Fatalf("%v %q", err, out)
			}
			for _, want := range c.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("lacks %q: %v", want, err)
				}
			}
			for _, absent := range c.absent {
				if strings.Contains(err.Error(), absent) {
					t.Fatalf("unexpected %q: %v", absent, err)
				}
			}
		})
	}
	t.Setenv("PATH", withPath(plugin, node))
	t.Setenv("FAKE_EXIT", "3")
	var exit *ExitError
	out, _, err := rc("profile", "typescript", "generate", "--name", "orders", "-x")
	if !errors.As(err, &exit) || exit.Code != 3 || out != "args: generate --name orders -x\n" {
		t.Fatalf("%v %q", err, out)
	}
}

func TestPythonAndJavaPluginHintsAndRuntimes(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, _, err := rc("profile", "python", "generate")
	for _, want := range []string{"pip install runtimeconditions-profiler", "runtimeconditions-python-profiler", "needs Python, but python3 was not found"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("python lacks %q: %v", want, err)
		}
	}
	_, _, err = rc("profile", "java", "generate")
	for _, want := range []string{"the java profiler is not installed: the Java profiler is not released yet", "needs Java, but java was not found"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("java lacks %q: %v", want, err)
		}
	}
	t.Setenv("PATH", withPath(fakeProfiler(t, "runtimeconditions-python-profiler"), fakeRuntime(t, "python3")))
	if out, _, err := rc("profile", "python", "generate"); err != nil || out != "args: generate\n" {
		t.Fatalf("%v\n%q", err, out)
	}
}

func installPlugin(t *testing.T, language, manifest string) {
	prefix := t.TempDir()
	if err := os.MkdirAll(filepath.Join(prefix, "share", "rc", "plugins"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(prefix, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	profiler := fakeProfiler(t, "rc-profiler-"+language)
	if err := os.Symlink(filepath.Join(profiler, "rc-profiler-"+language), filepath.Join(prefix, "bin", "rc-profiler-"+language)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prefix, "share", "rc", "plugins", language+".json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", filepath.Join(prefix, "bin"))
}

func TestPluginManifestOverridesTheBuiltInRuntime(t *testing.T) {
	installPlugin(t, "typescript", `{"language":"typescript","runtime":{"command":"bun","name":"Bun","install":"brew install bun"}}`)
	base := os.Getenv("PATH")
	t.Setenv("PATH", withPath(base, fakeRuntime(t, "node")))
	_, errout, err := rc("profile", "typescript", "generate")
	if err == nil || !strings.Contains(err.Error(), "the typescript profiler needs Bun, but bun was not found on PATH: brew install bun") || errout != "" {
		t.Fatalf("%v %q", err, errout)
	}
	t.Setenv("PATH", withPath(base, fakeRuntime(t, "bun")))
	if out, _, err := rc("profile", "typescript", "generate"); err != nil || out != "args: generate\n" {
		t.Fatalf("%v %q", err, out)
	}
}

func TestPluginManifestAppliesToALanguageRcDoesNotKnow(t *testing.T) {
	installPlugin(t, "ruby", `{"language":"ruby","runtime":{"command":"ruby"}}`)
	if _, _, err := rc("profile", "ruby", "generate"); err == nil || !strings.Contains(err.Error(), "the ruby profiler needs ruby, but ruby was not found on PATH") {
		t.Fatalf("%v", err)
	}
}

func TestMalformedManifestFallsBackToTheBuiltInRuntimeWithAWarning(t *testing.T) {
	for _, manifest := range []string{
		`{not json`,
		`{"language":"python","runtime":{"command":"x"}}`,
		`{"language":"typescript","runtime":{}}`,
		`{"language":"typescript","runtime":{"command":"bun","minimum":"1.1"}}`,
	} {
		installPlugin(t, "typescript", manifest)
		_, errout, err := rc("profile", "typescript", "generate")
		if err == nil || !strings.Contains(err.Error(), "needs Node.js, but node was not found") || !strings.Contains(errout, "warning: ignoring the typescript plugin manifest") {
			t.Fatalf("%s: %v %q", manifest, err, errout)
		}
	}
}

func TestMissingManifestFallsBackSilently(t *testing.T) {
	t.Setenv("PATH", fakeProfiler(t, "rc-profiler-typescript"))
	if _, errout, err := rc("profile", "typescript", "generate"); err == nil || errout != "" {
		t.Fatalf("%v %q", err, errout)
	}
}

func TestCoreSchemaIsCheckedAfterThePluginAndRuntimeChecks(t *testing.T) {
	plugin := fakeProfiler(t, "rc-profiler-typescript")
	t.Setenv("FAKE_SCHEMAS", `["0.2.0"]`)
	t.Setenv("PATH", plugin)
	if _, _, err := rc("profile", "typescript", "generate"); err == nil || !strings.Contains(err.Error(), "needs Node.js") || strings.Contains(err.Error(), "core schema") {
		t.Fatalf("%v", err)
	}
	t.Setenv("PATH", withPath(plugin, fakeRuntime(t, "node")))
	if _, _, err := rc("profile", "typescript", "generate"); err == nil || !strings.Contains(err.Error(), "supports core schema 0.2.0") {
		t.Fatalf("%v", err)
	}
}
