package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/acaloiaro/go-envparse"
)

const VERSION = "2.14.1"
const BUILD_DATE = "2024-12-12T19:42:19+00:00"

type exampleFlag map[string]string

func (e exampleFlag) String() string {
	var str string
	for key, val := range e {
		str = fmt.Sprintf(`%s --example="%s=%s"`, str, key, val)
	}

	return str
}

func (e exampleFlag) Set(value string) (err error) {
	key, value, found := strings.Cut(value, "=")
	if !found {
		err = errors.New("examples must be provided in KEY=VALUE format")
		return
	}

	e[key] = value

	return nil
}

var (
	examplesFlag   = make(exampleFlag)
	debugFlag      bool
	skipGitAddFlag bool
	envFileFlag    string
	sampleFileFlag string
	versionFlag    bool
)

func init() {
	flag.StringVar(&envFileFlag, "env-file", ".env", "set the path to your env file: ess -env-file=.env_file [sync|install]")
	flag.StringVar(&sampleFileFlag, "sample-file", "env.sample", "set the path to your sample file: ess -sample-file=env_var.sample [sync|install]")
	flag.BoolVar(&debugFlag, "debug", false, "print debug logs: ess --debug [sync|install]")
	flag.BoolVar(&skipGitAddFlag, "skip-git-add", false, "skip doing 'git add' on generated sample file after sync")
	flag.Var(examplesFlag, "example", "set example values for samples: ess --example=BAR=\"my bar value\" [sync|install]")
	flag.BoolVar(&versionFlag, "version", false, "print the current ess version: ess --version")

	flag.Usage = func() {
		cmd := os.Args[0]
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", cmd)
		fmt.Fprintf(os.Stderr, "%s [-flags] %s \n\n", cmd, "[install|sync]")
		flag.PrintDefaults()
	}

	flag.Parse()

	if versionFlag {
		fmt.Fprintf(os.Stdout, "ess version: %s built at: %s\n", VERSION, BUILD_DATE)
		os.Exit(0)
	}

	if debugFlag {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

}

func main() {
	command := "sync"
	args := flag.Args()
	if len(args) > 0 {
		command = args[0]
	}

	projectPath, gitDirPath := gitDirPaths()

	switch command {
	case "sync":
		sync(projectPath)

		if skipGitAddFlag {
			return
		}

		cmd := exec.Command("git", "add", filepath.Join(projectPath, sampleFileFlag))
		slog.Debug("running git command", "args", cmd.Args)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("unable to add sample file '%s' to git: %v", sampleFileFlag, err)
			os.Exit(1)
		}
	case "install":
		slog.Debug("installing hook to", "path", gitDirPath)
		err := installHook(gitDirPath)
		if err != nil {
			fmt.Println("unable to install pre-commit hook:", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("ess: unknown command '%s'\n\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func sync(dir string) {
	envFilePath := filepath.Join(dir, envFileFlag)
	sampleFilePath := filepath.Join(dir, sampleFileFlag)

	slog.Debug("syncing env file with sample", "env_file", envFilePath, "sample_file", sampleFilePath)

	envFileReader, err := os.Open(envFilePath)
	if err != nil {
		fmt.Printf("env file '%s' was not found. skipping sync.\n", envFilePath)
		os.Exit(0)
	}

	envFile, err := envparse.ParsePermissive(envFileReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to parse env file (%s): %v", envFilePath, err)
		os.Exit(1)
	}

	scrubEnvFile(envFile, examplesFlag)
	err = writeSampleFile(envFile, envFilePath, sampleFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	slog.Debug("sample file written", "sample_file", sampleFilePath)
}

func writeSampleFile(sampleFileContent map[string]string, envFilePath, sampleFilePath string) (err error) {
	sampleFile, err := os.Create(sampleFilePath)
	if err != nil {
		err = fmt.Errorf("unable to create sample file: %w", err)
		return
	}
	defer sampleFile.Close()

	sampleWriter := bufio.NewWriter(sampleFile)
	envFile, err := os.Open(envFilePath)
	if err != nil {
		err = fmt.Errorf("unable to read env file: %w", err)
		return
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)

	// interate through the env file line-by-line, searching for environment variables
	// any line that starts with a key from `sampleFileContents` followed by an equal sign `=` is considered
	// a line that contains a secret
	for scanner.Scan() {
		envFileLine := scanner.Text()
		scrubbedLine := replaceSecrets(envFileLine, sampleFileContent)
		sampleWriter.WriteString(fmt.Sprintf("%s\n", scrubbedLine))
	}

	sampleWriter.Flush()

	return
}

func scrubEnvFile(envFile map[string]string, examples map[string]string) {
	slog.Debug("scrubbing env file of secrets")
	for envFileKey := range envFile {
		exampleVal, ok := examples[envFileKey]
		if ok {
			envFile[envFileKey] = exampleVal
		} else {
			envFile[envFileKey] = fmt.Sprintf("<%s>", envFileKey)
		}
	}
}

// replaceSecrets replaces the content of env file entries (lines) with a new line that is scrubbed of secrets
//
// any line containing a variable name followed by and equal sign is considered to contain a secret
func replaceSecrets(envFileEntry string, sampleFileContent map[string]string) (newLine string) {
	var r *regexp.Regexp
	newLine = envFileEntry

	for secretKey, secretPlaceholder := range sampleFileContent {
		r = regexp.MustCompile(fmt.Sprintf("(%s.*=.*)", secretKey))
		if r.MatchString(envFileEntry) {
			slog.Debug("replacing secrets with sample values", "secret_key", secretKey, "placeholder", secretPlaceholder)
			newLine = r.ReplaceAllString(envFileEntry, fmt.Sprintf("%s=%s", secretKey, secretPlaceholder))
		}
	}

	return
}

// gitDirPath returns the path the the GIT_DIR
func gitDirPaths() (projectPath, gitDir string) {
	var outb, errb bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		fmt.Println("unable to find project directory:", err)
		os.Exit(1)
	}

	projectPath = strings.TrimRight(outb.String(), "\r\n")

	outb.Reset()
	errb.Reset()

	cmd = exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		fmt.Println("unable to find GIT_DIR:", err)
		os.Exit(1)
	}

	gitDir = strings.TrimRight(outb.String(), "\r\n")
	gitDir = filepath.Join(projectPath, gitDir)

	return
}

var preCommitScriptTemplate = `{{if .Appending}}{{ .OldPreCommitScript }}
# Below code generated by ess https://github.com/acaloiaro/ess
{{else}}#!/usr/bin/env bash
# File generated by ess https://github.com/acaloiaro/ess

{{end}}ARGS=({{ .Args }})

exec $(dirname -- "${BASH_SOURCE[0]}")/{{ .PreCommitHooksDir }}/0-ess "${ARGS[@]}"
`

// installHook installs the ess git hook to 'dir'
func installHook(gitDirPath string) (err error) {
	hooksScriptDirName := "pre-commit-hooks.d"
	hooksDirName := filepath.Join(gitDirPath, "hooks")
	hooksScriptDirPath := filepath.Join(hooksDirName, hooksScriptDirName)
	preCommitHookScriptPath := filepath.Join(hooksDirName, "pre-commit")
	appendFlag := false
	var oldPreCommitScript []byte

	_, err = os.Stat(preCommitHookScriptPath)
	if !os.IsNotExist(err) {
		fmt.Printf("A pre-commit hook already exists. Would you like to cancel [c], overwrite [o], or append [a] the existing pre-commit hook script? [a, c, o]: ")
		var response string
		fmt.Scanln(&response)
		if response == "c" || (response != "o" && response != "a") {
			slog.Debug("user declined to overwrite the existing pre-commit hook")
			os.Exit(0)
		}
		if response == "a" {
			slog.Debug("will append pre-commit hook script to existing script", "script_path", preCommitHookScriptPath)
			// read the existing content of the pre-commit file
			oldPreCommitScript, err = os.ReadFile(preCommitHookScriptPath)
			if err != nil {
				return err
			}

			appendFlag = true
		}
	}

	// Create a directory to place pre-commit-hook executable within
	if _, err = os.Stat(hooksScriptDirPath); os.IsNotExist(err) {
		if err = os.MkdirAll(hooksScriptDirPath, os.ModePerm); err != nil {
			return
		}
	}

	slog.Debug("hook will be installed in", "dir_path", hooksScriptDirPath)
	preCommitHookPath := filepath.Join(hooksScriptDirPath, "0-ess")
	preCommitHook, err := os.OpenFile(preCommitHookPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return
	}

	executablePath, err := exec.LookPath(os.Args[0])
	if err != nil {
		fmt.Println("unable to find 'ess' in $PATH:", err)
		os.Exit(1)
	}

	executable, err := os.Open(executablePath)
	if err != nil {
		return
	}

	// copy the currently running executable wholsale into the pre commit hooks directory
	io.Copy(preCommitHook, executable)

	tmpl, err := template.New("pre-commit-script").Parse(preCommitScriptTemplate)
	if err != nil {
		return err
	}
	args := fmt.Sprintf("--env-file=%s --sample-file=%s %s", envFileFlag, sampleFileFlag, examplesFlag)
	buff := bytes.NewBufferString("")

	templateValues := struct {
		Args               string
		PreCommitHooksDir  string
		Appending          bool
		OldPreCommitScript string
	}{
		Args:               args,
		PreCommitHooksDir:  hooksScriptDirName,
		Appending:          appendFlag,
		OldPreCommitScript: string(oldPreCommitScript),
	}

	slog.Debug("hook metadata", "data", templateValues)
	err = tmpl.Execute(buff, templateValues)
	if err != nil {
		return err
	}

	preCommitHookScript, err := os.OpenFile(preCommitHookScriptPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer preCommitHookScript.Close()

	_, err = buff.WriteTo(preCommitHookScript)
	if err != nil {
		return err
	}

	fmt.Println("ess pre-commit hook installed!")
	return nil
}
