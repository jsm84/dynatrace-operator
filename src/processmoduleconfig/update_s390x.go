//go:build s390x
//+build s390x

package processmoduleconfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Dynatrace/dynatrace-operator/src/dtclient"
	"github.com/spf13/afero"
)

var (
	ruxitAgentProcPath       = filepath.Join("agent", "conf", "ruxitagentproc.conf")
	sourceRuxitAgentProcPath = filepath.Join("agent", "conf", "_ruxitagentproc.conf")
)

func UpdateProcessModuleConfig(fs afero.Fs, targetDir string, processModuleConfig *dtclient.ProcessModuleConfig) error {
	if processModuleConfig != nil {
		log.Info("updating ruxitagentproc.conf", "targetDir", targetDir)
		usedProcessModuleConfigPath := filepath.Join(targetDir, ruxitAgentProcPath)
		sourceProcessModuleConfigPath := filepath.Join(targetDir, sourceRuxitAgentProcPath)
		if err := checkProcessModuleConfigCopy(fs, sourceProcessModuleConfigPath, usedProcessModuleConfigPath); err != nil {
			return err
		}
		return Update(fs, sourceProcessModuleConfigPath, usedProcessModuleConfigPath, processModuleConfig.ToMap())
	}
	log.Info("no changes to ruxitagentproc.conf, skipping update")
	return nil
}

// checkProcessModuleConfigCopy checks if we already made a copy of the original ruxitagentproc.conf file.
// After the initial installation of a version we copy the ruxitagentproc.conf to _ruxitagentproc.conf, and we use the _ruxitagentproc.conf + the api response to re-create the ruxitagentproc.conf
// so it`s easier to update
func checkProcessModuleConfigCopy(fs afero.Fs, sourcePath, destPath string) error {
	if _, err := fs.Open(sourcePath); os.IsNotExist(err) {
		usedProcessModuleConfigFile, err := fs.Open(destPath)
		if os.IsNotExist(err) {
			log.Info("original ruxitagentproc.conf not found")
			if err = createProcessModuleConfigFile(fs, destPath); err == nil {
				usedProcessModuleConfigFile, err = fs.Open(destPath)
			}
		}
		// catch any leftover error from fs.Open or file creation above
		if err != nil {
			return err
		}

		log.Info("saving original ruxitagentproc.conf to _ruxitagentproc.conf")
		fileInfo, err := fs.Stat(destPath)
		if err != nil {
			return err
		}

		sourceProcessModuleConfigFile, err := fs.OpenFile(sourcePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(sourceProcessModuleConfigFile, usedProcessModuleConfigFile)
		if err != nil {
			if err := sourceProcessModuleConfigFile.Close(); err != nil {
				log.Error(err, "failed to close sourceProcessModuleConfigFile")
			}
			if err := usedProcessModuleConfigFile.Close(); err != nil {
				log.Error(err, "failed to close usedProcessModuleConfigFile")
			}
			return err
		}
		if err = sourceProcessModuleConfigFile.Close(); err != nil {
			return err
		}
		return usedProcessModuleConfigFile.Close()
	}
	return nil
}

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
