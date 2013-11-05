package goat

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type typefunc func(*GoatEnv, *Dependency) error

var typemap = map[string]typefunc{
	"":    GoGet,
	"get": GoGet,
	"git": Git,
	"hg":  Hg,
}

func header(c string, strs ...interface{}) {
	fmt.Printf("\n")
	for i := 0; i < 80; i++ {
		fmt.Printf(c)
	}
	fmt.Printf("\n")

	fmt.Println(strs...)

	for i := 0; i < 80; i++ {
		fmt.Printf(c)
	}
	fmt.Printf("\n")
	fmt.Printf("\n")
}

// FetchDependencies goes and retrieves the dependencies for a given GoatEnv. If
// the dependencies have any Goatfile's in them, this will fetch the
// dependencies listed therein as well. All dependencies are placed in the root
// project's .lib directory, INCLUDING any sub-dependencies. This is done on
// purpose!
func FetchDependencies(genv *GoatEnv) error {
	var err error

	if len(genv.Dependencies) > 0 {
		header("#", "Downloading dependencies listed in", genv.Goatfile)

		for i := range genv.Dependencies {
			dep := &genv.Dependencies[i]

			header("=", "Retrieving dependency at:", dep.Location)

			if dep.Path == "" {
				dep.Path = dep.Location
			}

			fun, ok := typemap[dep.Type]
			if !ok {
				return errors.New("Unknown dependency type: " + dep.Type)
			}
			err = fun(genv, dep)
			if err != nil {
				return err
			}

			depprojroot := filepath.Join(genv.ProjRootLib, "src", dep.Path)

			if IsProjRoot(depprojroot) {
				header("-", "Reading", depprojroot, "'s dependencies")

				depgenv, err := SetupGoatEnv(depprojroot)
				if err != nil {
					return err
				}
				ChrootEnv(depgenv, genv.ProjRoot)
				err = FetchDependencies(depgenv)
				if err != nil {
					return err
				}
			} else {
				header("-", "No Goatfile found in", depprojroot)
			}
		}

		header("#", "Done downloading dependencies for", genv.Goatfile)
	} else {
		header("-", "No dependencies listed in", genv.Goatfile)
	}

	return nil
}

// Use `go get`.
func GoGet(genv *GoatEnv, dep *Dependency) error {
	fmt.Println("go", "get", dep.Location)
	return PipedCmd("go", "get", dep.Location)
}

// Use `git`.
func Git(genv *GoatEnv, dep *Dependency) error {
	localloc := filepath.Join(genv.ProjRootLib, "src", dep.Path)

	fmt.Println("git", "clone", dep.Location, localloc)
	err := PipedCmd("git", "clone", dep.Location, localloc)
	if err != nil {
		return err
	}

	origcwd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.Chdir(localloc)
	if err != nil {
		return err
	}
	defer os.Chdir(origcwd)

	fmt.Println("git", "fetch", "-pv", "--all")
	err = PipedCmd("git", "fetch", "-pv", "--all")
	if err != nil {
		return err
	}

	if dep.Reference == "" {
		dep.Reference = "master"
	}
	fmt.Println("git", "checkout", dep.Reference)
	err = PipedCmd("git", "checkout", dep.Reference)
	if err != nil {
		return err
	}

	fmt.Println("git", "clean", "-f", "-d")
	err = PipedCmd("git", "clean", "-f", "-d")

	return err

}

// Use `hg`.
func Hg(genv *GoatEnv, dep *Dependency) error {
	localloc := filepath.Join(genv.ProjRootLib, "src", dep.Path)

	fmt.Println("hg", "clone", dep.Location, localloc)
	err := PipedCmd("hg", "clone", dep.Location, localloc)
	if err != nil {
		return err
	}

	origcwd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.Chdir(localloc)
	if err != nil {
		return err
	}
	defer os.Chdir(origcwd)

	fmt.Println("hg", "pull")
	err = PipedCmd("hg", "pull")
	if err != nil {
		return err
	}

	if dep.Reference == "" {
		dep.Reference = "tip"
	}
	fmt.Println("hg", "update", "-C", dep.Reference)
	err = PipedCmd("hg", "update", "-C", dep.Reference)

	return err

}

