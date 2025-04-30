package util

import (
	"os"
	"path/filepath"
)

type AppPaths struct {
	SchemaDir    string
	DatabaseFile string
}

func executableDir() (string, error) {
	myExecutable, err := os.Executable()
	if err != nil {
		return "", err
	}

	myExecutable, err = filepath.EvalSymlinks(myExecutable)
	if err != nil {
		return "", err
	}
	return filepath.Dir(myExecutable), nil
}

func isSystemInstall(myDir string) (bool, error) {
	if myDir != "/bin" && myDir != "/usr/bin" {
		return false, nil
	}
	return true, nil
}

func isSystemLocalInstall(myDir string) (bool, error) {
	if myDir != "/usr/local/bin" {
		return false, nil
	}
	return true, nil
}

func isUserLocalInstall(myDir string) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	if myDir != filepath.Join(home, ".local/bin") {
		return false, nil
	}
	return true, nil
}

func findDefaultDirectories() (*AppPaths, error) {
	myDir, err := executableDir()
	if err != nil {
		return nil, err
	}

	if system, err := isSystemInstall(myDir); err != nil {
		return nil, err
	} else if system {
		return &AppPaths{
			SchemaDir:    "/usr/share/linkshare/schema",
			DatabaseFile: "/var/lib/linkshare/linkshare.db",
		}, nil
	}

	if local, err := isSystemLocalInstall(myDir); err != nil {
		return nil, err
	} else if local {
		return &AppPaths{
			SchemaDir:    "/usr/local/share/linkshare/schema",
			DatabaseFile: "/var/lib/linkshare/linkshare.db",
		}, nil
	}

	if user, err := isUserLocalInstall(myDir); err != nil {
		return nil, err
	} else if user {
		return &AppPaths{
			SchemaDir:    filepath.Join(myDir, "../share/linkshare/"),
			DatabaseFile: filepath.Join(myDir, "../var/lib/linkshare/linkshare.db"),
		}, nil
	}

	return &AppPaths{
		SchemaDir:    filepath.Join(myDir, "../schema"),
		DatabaseFile: filepath.Join(myDir, "../linkshare.db"),
	}, nil
}

// FindDirectories will find all relevant directories.
// It checks where the binary is installed and based on that it will decide
// where to look for other static data.
//
// Parameters:
//   - dbPath: the explicit db path passed on the command line. If this contains
//     a non-empty string the db returned database file will always be this one.
//     If this contains an empty string a default database file will be picked
//     based on where the binary is installed.
func FindDirectories(dbPath string) (*AppPaths, error) {
	paths, err := findDefaultDirectories()
	if err != nil {
		return nil, err
	}
	if dbPath != "" {
		paths.DatabaseFile = dbPath
	}
	return paths, nil
}

// CreateDirectories ensures all application managed directories are created
func CreateDirectories(paths *AppPaths) error {
	err := os.MkdirAll(filepath.Dir(paths.DatabaseFile), 0750)
	if err != nil {
		return err
	}
	return nil
}
