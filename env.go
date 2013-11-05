package goat

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

// FindGoatFile returns the directory name of the parent that contains the
// Goatfile
func FindGoatfile(dir string) (string, error) {

	if IsProjRoot(dir) {
		return dir, nil
	}

	parent := filepath.Dir(dir)
	if dir == parent {
		return "", errors.New("Goatfile not found")
	}

	return FindGoatfile(parent)
}

// IsProjRoot returns whether or not a particular directory is the project
// root for a goat project (aka, whether or not it has a goat file)
func IsProjRoot(dir string) bool {
  gofile := Gofile(dir)

  if gofile == "" {
    return false
  }

  return true
}

// Checks a directory for a Go file. If the directory has a go file,
// it returns the name of the go file. Otherwise empty stringis returned.
//
// This allows us to support the old "Goatfile" standard as we transifition
// to the new ".go" file.
func Gofile(dir string) string {
	gofile := filepath.Join(dir, GOFILE)

	if _, err := os.Stat(gofile); os.IsNotExist(err) {
	  goatfile := filepath.Join(dir, GOATFILE)

	  if _, err := os.Stat(goatfile); os.IsNotExist(err) {
      return ""
    } else {
		  return goatfile
	  }
  } else {
    return gofile
  }

	return ""
}

// NewGoatEnv returns a new GoatEnv struct based on the directory passed in
func SetupGoatEnv(projroot string) (*GoatEnv, error) {

	goatfile := Gofile(projroot)

	projrootlib := filepath.Join(projroot, GODIR)

	genv := GoatEnv{ProjRoot: projroot,
		ProjRootLib: projrootlib,
		Goatfile:    goatfile}

	genvraw, err := ioutil.ReadFile(goatfile)
	if err != nil {
		return nil, err
	}

	err = UnmarshalGoat(genvraw, &genv)
	return &genv, err
}

// ChrootEnv changes the root directories of a given environment. Useful if you
// want to make the dependencies download somewhere else
func ChrootEnv(genv *GoatEnv, newroot string) {
	newrootlib := filepath.Join(newroot, GODIR)
	genv.ProjRoot = newroot
	genv.ProjRootLib = newrootlib
}

func envPrepend(dir string) error {
	gopath, _ := syscall.Getenv("GOPATH")
	return syscall.Setenv("GOPATH", dir+":"+gopath)
}

// EnvPrependProj prepends a goat project's root and lib directories to the GOPATH
func EnvPrependProj(genv *GoatEnv) error {
	err := envPrepend(genv.ProjRoot)
	if err != nil {
		return err
	}

	return envPrepend(genv.ProjRootLib)
}

func ActualGo() (string, bool) {
	bin, ok := syscall.Getenv("GOAT_ACTUALGO")
	return bin, ok
}
