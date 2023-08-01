package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/hashicorp/go-envparse"
)

type exampleFlag map[string]string

func (e exampleFlag) String() string {
	var str string
	for key, val := range e {
		str = fmt.Sprintf(`%s --example="%s=%s"`, str, key, val)
	}

	return str
}

func (e exampleFlag) Set(value string) (err error) {
	exampleParts := strings.Split(value, "=")
	if len(exampleParts) != 2 {
		err = errors.New("examples must be provided in KEY=VALUE format")
		return
	}

	e[exampleParts[0]] = exampleParts[1]

	return nil
}

var envFileFlag string
var sampleFileFlag string
var examplesFlag = make(exampleFlag)

func init() {
	flag.StringVar(&envFileFlag, "env-file", ".env", "-env-file=.env_file")
	flag.StringVar(&sampleFileFlag, "sample-file", "env.sample", "-sample-file=env_var.sample")
	flag.Var(examplesFlag, "example", "--example=FOO=\"my foo value\" --example=BAR=\"my bar value\"")

	flag.Usage = func() {
		cmd := os.Args[0]
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", cmd)
		fmt.Fprintf(os.Stderr, "%s [-flags] %s \n\n", cmd, "[install|sync]")
		flag.PrintDefaults()
	}

	flag.Parse()
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
		cmd := exec.Command("git", "add", filepath.Join(projectPath, sampleFileFlag))
		err := cmd.Run()
		if err != nil {
			fmt.Printf("unable to add sample file '%s' to git: %v", sampleFileFlag, err)
			os.Exit(1)
		}
	case "install":
		err := installHook(gitDirPath)
		if err != nil {
			fmt.Println("unable to install pre-commit hook:", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("env-sample-sync: unknown command '%s'\n\n", command)
		flag.Usage()
	}
}

func sync(dir string) {
	envFilePath := filepath.Join(dir, envFileFlag)
	sampleFilePath := filepath.Join(dir, sampleFileFlag)

	envFileReader, err := os.Open(envFilePath)
	if err != nil {
		fmt.Printf("env file '%s' was not found. skipping sync.\n", envFilePath)
		os.Exit(0)
	}

	envFile, err := envparse.Parse(envFileReader)
	if err != nil {
		fmt.Println("unable to parse env file:", err)
		os.Exit(1)
	}

	scrubEnvFile(envFile, examplesFlag)
	err = writeSampleFile(envFile, envFilePath, sampleFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
# Below code generated by env-sample-sync https://github.com/acaloiaro/env-sample-sync
{{else}}#!/usr/bin/env bash
# File generated by env-sample-sync https://github.com/acaloiaro/env-sample-sync

{{end}}ARGS=({{ .Args }})

exec $(dirname -- "${BASH_SOURCE[0]}")/{{ .PreCommitHooksDir }}/0-env-sample-sync "${ARGS[@]}"
`

// installHook installs the env-sample-sync git hook to 'dir'
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
			os.Exit(0)
		}
		if response == "a" {
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

	preCommitHookPath := filepath.Join(hooksScriptDirPath, "0-env-sample-sync")
	preCommitHook, err := os.OpenFile(preCommitHookPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}

	executablePath, err := exec.LookPath(os.Args[0])
	if err != nil {
		fmt.Println("unable to find 'env-sample-sync' in $PATH:", err)
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
	var buff = bytes.NewBufferString("")

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

	err = tmpl.Execute(buff, templateValues)
	if err != nil {
		return err
	}

	preCommitHookScript, err := os.OpenFile(preCommitHookScriptPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer preCommitHookScript.Close()

	_, err = buff.WriteTo(preCommitHookScript)
	if err != nil {
		return err
	}

	fmt.Println("env-sample-sync pre-commit hook installed!")
	return nil
}
