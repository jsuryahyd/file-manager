package fileops

import "github.com/spf13/afero"

// AppFs is the filesystem to use. It is a global variable for convenience.
// In production, it will be afero.NewOsFs().
// In tests, it can be replaced with afero.NewMemMapFs().
var AppFs = afero.NewOsFs()
