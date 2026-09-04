package doctor

import (
	"os"
	"path/filepath"
	"strings"

	"ryoku-cli/internal/sys"
)

// Ryotunes ships as a [ryoku] package (a ryoku-desktop depend). Two things
// keep an updated box opening the retired Chromium YouTube Music window
// instead: a wrapper or a locally built copy in ~/.local/bin, which shadows
// /usr/bin/ryotunes on PATH (a dev deploy laid both before the package
// existed), and a box whose channel switch never installed the package.
func reconcileRyotunes(checkOnly bool) recResult {
	var problems, fixes []string

	bin := filepath.Join(sys.Home(), ".local", "bin", "ryotunes")
	stale := staleUserRyotunes(bin)
	if stale != "" {
		problems = append(problems, stale+" in ~/.local/bin shadows the packaged app")
		fixes = append(fixes, "rm -f ~/.local/bin/ryotunes ~/.local/share/applications/ryotunes.desktop")
	}
	if sys.ResolveRepo() == "" && sys.PkgInstalled("ryoku-desktop") && !sys.PkgInstalled("ryotunes") {
		problems = append(problems, "the ryotunes package is not installed")
		fixes = append(fixes, "sudo pacman -S --needed ryotunes")
	}
	if len(problems) == 0 {
		if _, err := sys.RunOut("pacman", "-Qoq", "/usr/bin/ryotunes"); err == nil {
			return okRes("ryotunes is the packaged app")
		}
		if sys.Exists("/usr/bin/ryotunes") {
			return warnRes("/usr/bin/ryotunes is not owned by the ryotunes package").
				withFix("sudo pacman -S --overwrite /usr/bin/ryotunes ryotunes")
		}
		return okRes("ryotunes is the packaged app")
	}
	if checkOnly {
		return wouldRes("%s", strings.Join(problems, "; ")).withFix(strings.Join(fixes, " && "))
	}
	if stale != "" {
		appshare := sys.Xdg("XDG_DATA_HOME", ".local/share")
		for _, p := range []string{
			bin,
			filepath.Join(appshare, "applications", "ryotunes.desktop"),
			filepath.Join(appshare, "ryoku", "ryotunes.commit"),
			filepath.Join(appshare, "icons", "hicolor", "scalable", "apps", "ryotunes.svg"),
		} {
			_ = os.Remove(p)
		}
		icons, _ := filepath.Glob(filepath.Join(appshare, "icons", "hicolor", "*", "apps", "ryotunes.png"))
		for _, p := range icons {
			_ = os.Remove(p)
		}
	}
	if sys.ResolveRepo() == "" && sys.PkgInstalled("ryoku-desktop") && !sys.PkgInstalled("ryotunes") {
		if err := sys.Sudo("pacman", "-S", "--needed", "--noconfirm", "ryotunes"); err != nil {
			return failRes("could not install ryotunes: %v", err).withFix("sudo pacman -S --needed ryotunes")
		}
	}
	return fixedRes("ryotunes opens the packaged app (%s)", strings.Join(problems, "; "))
}

// staleUserRyotunes names what ~/.local/bin/ryotunes is when it is not the
// user's own program: the Chromium wrapper (a script that opens
// music.youtube.com) or a build the dev deploy recorded a commit for.
func staleUserRyotunes(bin string) string {
	st, err := os.Stat(bin)
	if err != nil || st.IsDir() {
		return ""
	}
	head := make([]byte, 4096)
	if f, err := os.Open(bin); err == nil {
		n, _ := f.Read(head)
		f.Close()
		head = head[:n]
	}
	if strings.HasPrefix(string(head), "#!") && strings.Contains(string(head), "music.youtube.com") {
		return "the Chromium YouTube Music wrapper"
	}
	if sys.Exists(filepath.Join(sys.Xdg("XDG_DATA_HOME", ".local/share"), "ryoku", "ryotunes.commit")) {
		return "a locally built ryotunes"
	}
	return ""
}
