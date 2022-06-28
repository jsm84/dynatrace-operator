//go:build !s390x
//+build !s390x

package processmoduleconfig

import (
	"os"

	"github.com/spf13/afero"
)

// createProcessModuleConfigFile simply creates an empty file at location `destPath`.
// The file should get populated by the existing backup, compare & merge against JSON config data obtained from the live environment.
// This was originally implemented to fix a bug with the oneagent zip download for s390 where ruxitagentproc.conf is missing.
// It is retained for unit test compatibility between x86/arm/ppc64le and s390x.
func createProcessModuleConfigFile(fs afero.Fs, destPath string) error {
	log.Info("generating new ruxitagentproc.conf file")
	newProcessModuleConfigFile, err := fs.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	return newProcessModuleConfigFile.Close()
}
