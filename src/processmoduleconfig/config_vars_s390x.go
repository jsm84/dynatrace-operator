//go:build s390x
//+build s390x

package processmoduleconfig

import (
	"strings"
)

// Agent ConnectionInfo (TenantUUID, Token and Endpoints) is not defined here, in favor of existing routines which fetch and add
// this info to an in-memory ProcessModuleConfig.
var (
	defaultExecFilter = []string{`"cupsd"`, `"skype"`,
		`"google-chrome"`, `"firefox"`, `"mozilla-firefox"`,
		`"sshd"`, `"rpcbind"`, `"rsync"`, `"smbd"`, `"portmap"`,
		`"docker"`, `"docker.io"`, `"docker-proxy"`, `"jstack"`,
		`"jstat"`, `"jvisualvm"`, `"jps"`}
	defaultInjectionRules = []string{"::EXCLUDE:CONTAINS,PHP_CLI_SCRIPT_PATH,",
		"::EXCLUDE:EQUALS,EXE_NAME,php-cgi", "::INCLUDE:CONTAINS,ASPNETCORE_APPL_PATH,",
		"::INCLUDE:EQUALS,EXE_NAME,w3wp.exe", "::EXCLUDE:ENDS,NODEJS_APP_BASE_DIR,/node_modules/prebuild-install",
		"::EXCLUDE:ENDS,NODEJS_APP_BASE_DIR,/node_modules/npm", "::EXCLUDE:ENDS,NODEJS_APP_BASE_DIR,/node_modules/grunt",
		"::EXCLUDE:ENDS,NODEJS_APP_BASE_DIR,/node_modules/typescript", "::EXCLUDE:EQUALS,NODEJS_APP_NAME,yarn",
		"::EXCLUDE:ENDS,NODEJS_APP_BASE_DIR,/node_modules/node-pre-gyp", "::EXCLUDE:ENDS,NODEJS_APP_BASE_DIR,/node_modules/node-gyp",
		"::EXCLUDE:ENDS,NODEJS_APP_BASE_DIR,/node_modules/gulp-cli", "::EXCLUDE:EQUALS,NODEJS_SCRIPT_NAME,bin/pm2"}
	agentType = map[string]string{
		"php":    "off",
		"java":   "on",
		"nodejs": "off",
		"apache": "off",
		"nginx":  "off",
		"iis":    "off",
		"dotnet": "off",
		"sdk":    "off",
	}
	blockList = map[string]string{
		"ApplicationFilter": "",
		"ExecutableFilter":  strings.Join(defaultExecFilter, ","),
	}
	specializedAgent = map[string]string{
		"libraryPath64": `"../lib64"`,
	}
	general = map[string]string{
		"revision":                        "1",
		"logDir":                          `"../../log"`,
		"dataStorageDir":                  `"../.."`,
		"stripIdsFromKubernetesNamespace": "on",
		"removeIdsFromPaths":              "on",
		"websphereClusterNameInPG":        "on",
		"removeContainerIDfromPGI":        "on",
		"standalone":                      "on",
		"injectionRules":                  strings.Join(defaultInjectionRules, ";"),
	}
	sections = map[string]map[string]string{
		"agentType":        agentType,
		"blocklist":        blockList,
		"SpecializedAgent": specializedAgent,
		"general":          general,
	}
)
