package main

import (
	"strings"
	"testing"
)

func TestDetectDistro(t *testing.T) {
	for _, c := range []struct {
		id, like, want string
	}{
		{"arch", "", "arch"},
		{"cachyos", "arch", "arch"},
		{"endeavouros", "arch", "arch"},
		{"debian", "", "debian"},
		{"ubuntu", "debian", "debian"},
		{"linuxmint", "ubuntu debian", "debian"},
		{"fedora", "", "fedora"},
		{"nobara", "fedora", "fedora"},
		{"bazzite", "fedora", "fedora"},
		{"void", "", ""},
	} {
		d := detectDistro(c.id, c.like)
		got := ""
		if d != nil {
			got = d.id
		}
		if got != c.want {
			t.Errorf("detectDistro(%q,%q) = %q, want %q", c.id, c.like, got, c.want)
		}
	}
}

func TestLocalAllRenamesAndDrops(t *testing.T) {
	in := []string{"git", "networkmanager", "fd", "matugen", "limine", "kitty"}
	got := debianLinux.localAll(in)
	want := []string{"git", "network-manager", "fd-find", "kitty"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("localAll (debian) = %v, want %v", got, want)
	}
	gotFedora := fedoraLinux.localAll(in)
	wantFedora := []string{"git", "NetworkManager", "fd-find", "kitty"}
	if strings.Join(gotFedora, ",") != strings.Join(wantFedora, ",") {
		t.Fatalf("localAll (fedora) = %v, want %v", gotFedora, wantFedora)
	}
	if archLinux.local("networkmanager") != "networkmanager" {
		t.Error("arch must pass base.packages names through unchanged")
	}
}

// The Arch step list is the contract that must not drift; the Debian and Fedora
// ones swap the pacman-only steps for the source build.
func TestStepsPerDistro(t *testing.T) {
	ids := func(f *facts) []string {
		e := newEngine(f, &plan{}, true, "", "")
		var out []string
		for _, s := range e.steps {
			out = append(out, s.id)
		}
		return out
	}

	arch := strings.Join(ids(&facts{distro: archLinux}), " ")
	wantArch := "legacy sysupgrade tools payload backup repo conflicts packages drivers session configs aur shell doctor verify"
	if arch != wantArch {
		t.Errorf("arch steps = %q, want %q", arch, wantArch)
	}

	deb := strings.Join(ids(&facts{distro: debianLinux}), " ")
	wantDeb := "sysupgrade tools payload backup conflicts packages build session configs shell doctor verify"
	if deb != wantDeb {
		t.Errorf("debian steps = %q, want %q", deb, wantDeb)
	}

	fed := strings.Join(ids(&facts{distro: fedoraLinux}), " ")
	wantFed := "sysupgrade tools payload backup repo conflicts packages build session configs shell doctor verify"
	if fed != wantFed {
		t.Errorf("fedora steps = %q, want %q", fed, wantFed)
	}
}

func TestInstallArgs(t *testing.T) {
	got := strings.Join(archLinux.installArgs([]string{"git"}), " ")
	if got != "pacman -Syu --needed --noconfirm git" {
		t.Errorf("arch installArgs = %q", got)
	}
	got = strings.Join(debianLinux.installArgs([]string{"git"}), " ")
	if got != "apt-get -y install git" {
		t.Errorf("debian installArgs = %q", got)
	}
	got = strings.Join(debianLinux.removeArgs([]string{"dunst"}), " ")
	if got != "apt-get -y remove dunst" {
		t.Errorf("debian removeArgs = %q", got)
	}
	got = strings.Join(fedoraLinux.installArgs([]string{"git"}), " ")
	if got != "dnf -y install git" {
		t.Errorf("fedora installArgs = %q", got)
	}
	got = strings.Join(fedoraLinux.removeArgs([]string{"dunst"}), " ")
	if got != "dnf -y remove dunst" {
		t.Errorf("fedora removeArgs = %q", got)
	}
}
