package rc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// runtimeSpec is the language runtime a plugin needs. Its JSON form is the runtime object of a plugin manifest.
type runtimeSpec struct {
	Command string `json:"command"`
	Name    string `json:"name"`
	Install string `json:"install"`
}

type pluginManifest struct {
	Language string      `json:"language"`
	Runtime  runtimeSpec `json:"runtime"`
}

type pluginInfo struct {
	install string
	runtime runtimeSpec
}

// knownPlugins are the plugins rc knows without running them; a plugin manifest overrides the runtime.
var knownPlugins = map[string]pluginInfo{
	"typescript": {"brew install runtimeconditions/tools/rc-profiler-typescript", runtimeSpec{"node", "Node.js", "brew install node or https://nodejs.org"}},
	"python":     {"pip install runtimeconditions-profiler (installs runtimeconditions-python-profiler)", runtimeSpec{"python3", "Python", "brew install python or https://www.python.org/downloads"}},
	"java":       {"the Java profiler is not released yet", runtimeSpec{"java", "Java", "brew install openjdk or https://adoptium.net"}},
}

// readManifest reads <prefix>/share/rc/plugins/<language>.json for a plugin installed at <prefix>/bin.
func readManifest(language, pluginPath string) (runtimeSpec, error) {
	dirs := []string{filepath.Dir(pluginPath)}
	if resolved, err := filepath.EvalSymlinks(pluginPath); err == nil {
		dirs = []string{filepath.Dir(resolved), dirs[0]}
	}
	var data []byte
	err := os.ErrNotExist
	for _, dir := range dirs {
		if data, err = os.ReadFile(filepath.Join(dir, "..", "share", "rc", "plugins", language+".json")); err == nil {
			break
		}
	}
	if err != nil {
		return runtimeSpec{}, err
	}
	var manifest pluginManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&manifest); err != nil {
		return runtimeSpec{}, err
	}
	if manifest.Language != language || manifest.Runtime.Command == "" {
		return runtimeSpec{}, fmt.Errorf("it must name the language %s and the runtime command", language)
	}
	if manifest.Runtime.Name == "" {
		manifest.Runtime.Name = manifest.Runtime.Command
	}
	return manifest.Runtime, nil
}

// checkPlugin reports everything that keeps the profiler for language from running: the plugin is not
// installed, or its runtime is not on PATH. It never runs the plugin, which could not start without its runtime.
func checkPlugin(language, pluginPath string, stderr io.Writer) error {
	info, known := knownPlugins[language]
	if pluginPath != "" {
		switch spec, err := readManifest(language, pluginPath); {
		case err == nil:
			info.runtime, known = spec, true
		case !errors.Is(err, os.ErrNotExist):
			fmt.Fprintf(stderr, "warning: ignoring the %s plugin manifest: %v\n", language, err)
		}
	}
	if !known {
		return nil
	}
	var errs []error
	if pluginPath == "" {
		errs = append(errs, fmt.Errorf("the %s profiler is not installed: %s", language, info.install))
	}
	if _, err := exec.LookPath(info.runtime.Command); err != nil {
		message := fmt.Sprintf("the %s profiler needs %s, but %s was not found on PATH", language, info.runtime.Name, info.runtime.Command)
		if info.runtime.Install != "" {
			message += ": " + info.runtime.Install
		}
		errs = append(errs, errors.New(message))
	}
	return errors.Join(errs...)
}
