//go:build s390x
//+build s390x

package processmoduleconfig

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
)

// createProcessModuleConfigFile creates an empty file at location `destPath` and calls
// populateProcessModuleConfigFile() to populate the config file for the s390x architecture.
// This is to fix a bug with the oneagent zip download for s390x where ruxitagentproc.conf is missing from the archive.
func createProcessModuleConfigFile(fs afero.Fs, destPath string) error {
	log.Info("generating new ruxitagentproc.conf file")
	newProcessModuleConfigFile, err := fs.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	err = populateProcessModuleConfigFile(newProcessModuleConfigFile, destPath)
	if err != nil {
		return err
	}

	return newProcessModuleConfigFile.Close()
}

// populateProcessModuleConfigFile populates the empty config file on s390x
func populateProcessModuleConfigFile(cfg afero.File, destPath string) error {
	cfgStat, err := cfg.Stat()
	if err != nil {
		log.Error(err, "failed to fetch file info", "filePath", destPath)
		return err
	}

	if cfgStat.Size() > 0 {
		err := fmt.Errorf("cannot write to non-empty file")
		log.Error(err, "error writing to config file", "filePath", destPath)
		return err
	}

	log.Info("writing configuration data to file", "filePath", destPath)
	for hdr, content := range sections {
		err := writeFile(cfg, hdr, content)
		if err != nil {
			log.Error(err, "error writing to config file", "filePath", destPath, "section", hdr)
			return err
		}
	}

	return nil
}

func writeFile(cfg afero.File, header string, contents map[string]string) error {
	_, err := fmt.Fprintf(cfg, "[%s]\n", header)
	if err != nil {
		return err
	}

	for k, v := range contents {
		_, err = fmt.Fprintf(cfg, "%s %s\n", k, v)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(cfg, "\n")

	return err
}
