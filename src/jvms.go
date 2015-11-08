package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"./jvms/arch"
	"./jvms/file"
	"./jvms/node"
	"./jvms/jdk"
	"./jvms/web"
	//  "./ansi"
)

const (
  JvmsVersion = "0.0.1"
)

type Environment struct {
	settings        string
	root            string
	symlink         string
	arch            string
	proxy           string
	originalpath    string
	originalversion string
	currentVersion  string
}

var env = &Environment{
	settings:        os.Getenv("JVMS_HOME") + "\\settings.txt",
	root:            "",
	symlink:         os.Getenv("JAVA_HOME"),
	arch:            os.Getenv("PROCESSOR_ARCHITECTURE"),
	proxy:           "none",
	originalpath:    "",
	originalversion: "",
	currentVersion:  "",
}

func main() {
	args := os.Args
	detail := ""
	procarch := arch.Validate(env.arch)
	Setup()

	// Capture any additional arguments
	if len(args) > 2 {
		detail = strings.ToLower(args[2])
	}
	if len(args) > 3 {
		procarch = args[3]
	}
	if len(args) < 2 {
		help()
		return
	}

	// Run the appropriate method
	switch args[1] {
	case "install":
		install(detail, procarch)
	case "uninstall":
		uninstall(detail,procarch)
	case "use":
		use(detail, procarch)
	case "list":
		list(detail)
	case "ls":
		list(detail)
	case "ls-remote":
		listRemote(detail)
	case "on":
		enable()
	case "off":
		disable()
	case "root":
		if len(args) == 3 {
			updateRootDir(args[2])
		} else {
			fmt.Println("\nCurrent Root: " + env.root)
		}
	case "version":
		fmt.Println(JvmsVersion)
	case "v":
		fmt.Println(JvmsVersion)
	case "arch":
		if strings.Trim(detail, " \r\n") != "" {
			detail = strings.Trim(detail, " \r\n")
			if detail != "32" && detail != "64" {
				fmt.Println("\"" + detail + "\" is an invalid architecture. Use 32 or 64.")
				return
			}
			env.arch = detail
			saveSettings()
			fmt.Println("Default architecture set to " + detail + "-bit.")
			return
		}
		inuse := env.currentVersion
		var inusecpu string
		if strings.Contains(inuse,"64"){
			inusecpu = "64"
		}else {
			inusecpu = "32"
		}
		fmt.Println("System Default: " + env.arch + "-bit.")
		fmt.Println("Currently Configured: " + inusecpu + "-bit.")
	case "proxy":
		if detail == "" {
			fmt.Println("Current proxy: " + env.proxy)
		} else {
			env.proxy = detail
			saveSettings()
		}
	case "update":
		update()
	default:
		help()
	}
}

func update() {
	//  cmd := exec.Command("cmd", "/d", "echo", "testing")
	//  var output bytes.Buffer
	//  var _stderr bytes.Buffer
	//  cmd.Stdout = &output
	//  cmd.Stderr = &_stderr
	//  perr := cmd.Run()
	//  if perr != nil {
	//      fmt.Println(fmt.Sprint(perr) + ": " + _stderr.String())
	//      return
	//  }
}

func install(version string, cpuarch string) {
	if version == "" {
		fmt.Println("\nInvalid version.")
		fmt.Println(" ")
		help()
		return
	}

	cpuarch = strings.ToLower(cpuarch)

	if cpuarch != "" {
		if cpuarch != "32" && cpuarch != "64" && cpuarch != "all" {
			fmt.Println("\"" + cpuarch + "\" is not a valid CPU architecture. Must be 32 or 64.")
			return
		}
	} else {
		cpuarch = env.arch
	}

	if cpuarch != "all" {
		cpuarch = arch.Validate(cpuarch)
	}

	if cpuarch == "64" && env.arch=="32" {
		fmt.Println("JDK v" + version + " is only available in 32-bit.")
		return
	}

	// Check to see if the version is already installed
	if !jdk.IsVersionInstalled(env.root, version, cpuarch) {

		// Make the output directories
		os.Mkdir(env.root+"\\download", os.ModeDir)

		jdkdownloadURL := getJDKDownloadURL(version,cpuarch)

		if jdkdownloadURL=="" {
			fmt.Println("Version " + version + " is not available. If you are attempting to download a \"just released\" version,")
			fmt.Println("it may not be recognized by the jvms service yet (updated hourly). If you feel this is in error and")
			fmt.Println("you know the version exists, please visit http://github.com/ystyle/jvms and submit a PR.")
			return
		}

		// Download node
		if (cpuarch == "32" || cpuarch == "all") && !node.IsVersionInstalled(env.root, version, "32") {
			success := web.GetJDK(env.root, version, jdkdownloadURL, "32")
			if !success {
				os.RemoveAll(env.root+"\\download")
				fmt.Println("Could not download JDK v" + version + " 32-bit executable.")
				return
			}else {
				fmt.Printf("Installing JDK v"+version+"...")
				// new temp directory under the nvm root
				tempDir := env.root + "\\temp"

				// Extract npm to the temp directory
				file.Unzip(env.root+"\\download\\"+file.GenJDKZipFileName(version,"32"),tempDir+"\\jdk")

				// Copy the npm and npm.cmd files to the installation directory
				os.Rename(tempDir+"\\jdk",env.root+"\\v"+version+"_x86")

				// Remove the temp directory
				// may consider keep the temp files here
				os.RemoveAll(tempDir)

				fmt.Println("\n\nInstallation complete. If you want to use this version, type\n\njvms use "+version)
			}
		}
		if (cpuarch == "64" || cpuarch == "all") && !node.IsVersionInstalled(env.root, version, "64") {
			success := web.GetJDK(env.root, version, jdkdownloadURL, "64")
			if !success {
				os.RemoveAll(env.root+"\\download")
				fmt.Println("Could not download JDK v" + version + " 64-bit executable.")
				return
			}else {
				fmt.Printf("Installing JDK v"+version+"...")
				// new temp directory under the nvm root
				tempDir := env.root + "\\temp"

				// Extract npm to the temp directory
				file.Unzip(env.root+"\\download\\"+file.GenJDKZipFileName(version,"64"),tempDir+"\\jdk")

				// Copy the npm and npm.cmd files to the installation directory
				os.Rename(tempDir+"\\jdk",env.root+"\\v"+version+"_x64")

				// Remove the temp directory
				// may consider keep the temp files here
				os.RemoveAll(tempDir)

				fmt.Println("\n\nInstallation complete. If you want to use this version, type\n\njvms use "+version)
			}
		}

		return
	} else {
		fmt.Println("Version " + version + " is already installed.")
		return
	}

}

func uninstall(version string,a string) {
	// Make sure a version is specified
	if len(version) == 0 {
		fmt.Println("Provide the version you want to uninstall.")
		help()
		return
	}

	a = arch.Validate(a)

	// Determine if the version exists and skip if it doesn't
	if jdk.IsVersionInstalled(env.root, version, "32") || jdk.IsVersionInstalled(env.root, version, "64") {
		fmt.Printf("Uninstalling JDK v" + version + "...")
		v, _ := node.GetCurrentVersion()
		if v == version {
			cmd := exec.Command(env.root+"\\elevate.cmd", "cmd", "/C", "rmdir", env.symlink)
			cmd.Run()
		}
		if a == "32" && jdk.IsVersionInstalled(env.root, version, "32") {
			file32 := file.GenJDKFileName(version,"32")
			e := os.RemoveAll(env.root + "\\v" + file32)
			if e != nil {
				fmt.Println("Error removing jdk v" + version + " 32-Bit")
				fmt.Println("Manually remove " + env.root + "\\v" + file32 + ".")
			} else {
				fmt.Printf(" done")
			}
		}

		if a=="64" && jdk.IsVersionInstalled(env.root, version, "64") {
			file64 := file.GenJDKFileName(version,"64")
			e := os.RemoveAll(env.root + "\\v" + file64)
			if e != nil {
				fmt.Println("Error removing jdk v" + version + " 64-Bit")
				fmt.Println("Manually remove " + env.root + "\\v" + file64 + ".")
			} else {
				fmt.Printf(" done")
			}
		}

	} else {
		fmt.Println("jdk v" + version + " is not installed. Type \"jvms list\" to see what is installed.")
	}
	return
}

func use(version string, cpuarch string) {

	if version == "32" || version == "64" {
		cpuarch = version
		v, _ := jdk.GetCurrentVersion()
		version = v
	}

	cpuarch = arch.Validate(cpuarch)

	// Make sure the version is installed. If not, warn.
	if !jdk.IsVersionInstalled(env.root, version, cpuarch) {
		fmt.Println("jdk v" + version + " (" + cpuarch + "-bit) is not installed.")
		if cpuarch == "32" {
			if jdk.IsVersionInstalled(env.root, version, "64") {
				fmt.Println("\nDid you mean jdk v" + version + " (64-bit)?\nIf so, type \"jvms use " + version + " 64\" to use it.")
			}
		}
		if cpuarch == "64" {
			if jdk.IsVersionInstalled(env.root, version, "64") {
				fmt.Println("\nDid you mean jdk v" + version + " (64-bit)?\nIf so, type \"jvms use " + version + " 64\" to use it.")
			}
		}
		return
	}

	// Create or update the symlink
	sym, _ := os.Stat(env.symlink)

	if sym != nil {
		cmd := exec.Command(env.root+"\\elevate.cmd", "cmd", "/C", "rmdir", env.symlink)
		var output bytes.Buffer
		var _stderr bytes.Buffer
		cmd.Stdout = &output
		cmd.Stderr = &_stderr
		perr := cmd.Run()
		if perr != nil {
			fmt.Println(fmt.Sprint(perr) + ": " + _stderr.String())
			return
		}
	}
	fileName := file.GenJDKFileName(version,cpuarch)
	c := exec.Command(env.root+"\\elevate.cmd", "cmd", "/C", "mklink", "/D", env.symlink, env.root+"\\v"+fileName)
	var out bytes.Buffer
	var stderr bytes.Buffer
	c.Stdout = &out
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	env.currentVersion = fileName
	saveSettings()
	fmt.Println("Now using JDK v" + version + " (" + cpuarch + "-bit)")
}

func useArchitecture(a string) {
	if strings.ContainsAny("32", os.Getenv("PROCESSOR_ARCHITECTURE")) {
		fmt.Println("This computer only supports 32-bit processing.")
		return
	}
	if a == "32" || a == "64" {
		env.arch = a
		saveSettings()
		fmt.Println("Set to " + a + "-bit mode")
	} else {
		fmt.Println("Cannot set architecture to " + a + ". Must be 32 or 64 are acceptable values.")
	}
}

func list(listtype string) {
	if listtype == "" {
		listtype = "installed"
	}
	if listtype != "installed" && listtype != "available" {
		fmt.Println("\nInvalid list option.\n\nPlease use on of the following\n  - jvms list\n  - jvms list installed\n  - jvms list available")
		help()
		return
	}

	if listtype == "installed" {
		fmt.Println("")
		inuse := env.currentVersion
		var inusecpu string
		if strings.Contains(inuse,"64"){
			inusecpu = "64"
		}else {
			inusecpu = "32"
		}
		v := jdk.GetInstalled(env.root)
		for i := 0; i < len(v); i++ {
			version := v[i]
			isnode, _ := regexp.MatchString("v", version)
			str := ""
			if isnode {
				if "v"+inuse == version {
					str = str + "  * "
				} else {
					str = str + "    "
				}
				str = str + regexp.MustCompile("v").ReplaceAllString(version, "")
				if "v"+inuse == version {
					str = str + " (Currently using " + inusecpu + "-bit executable)"
					//            str = ansi.Color(str,"green:black")
				}
				fmt.Printf(str + "\n")
			}
		}
		if len(v) == 0 {
			fmt.Println("No installations recognized.")
		}
	} else {
		listRemote(listtype)
	}
}

func listRemote(detail string) {
	// Get raw text
	text := web.GetRemoteTextFile("https://raw.githubusercontent.com/ystyle/jvms/master/jdkversions.json")
	// Parse
	var data interface{}
	json.Unmarshal([]byte(text), &data)
	body := data.(map[string]interface{})
	fmt.Println("Remote Version List :")
	for key,_:=range body{
		fmt.Println("\t"+key)
	}
	fmt.Println("\nFor a complete list, visit https://raw.githubusercontent.com/ystyle/jvms/master/jdkversions.json")
}

func enable() {
	dir := ""
	files, _ := ioutil.ReadDir(env.root)
	for _, f := range files {
		if f.IsDir() {
			isnode, _ := regexp.MatchString("v", f.Name())
			if isnode {
				dir = f.Name()
			}
		}
	}
	fmt.Println("jvms enabled")
	if dir != "" {
		use(strings.Trim(regexp.MustCompile("v").ReplaceAllString(dir, ""), " \n\r"), env.arch)
	} else {
		fmt.Println("No versions of node.js found. Try installing the latest by typing jvms install latest")
	}
}

func disable() {
	cmd := exec.Command(env.root+"\\elevate.cmd", "cmd", "/C", "rmdir", env.symlink)
	cmd.Run()
	fmt.Println("jvms disabled")
}

func help() {
	fmt.Println("\nRunning version " + JvmsVersion + ".")
	fmt.Println("\nUsage:")
	fmt.Println(" ")
	fmt.Println("  jvms arch                     : Show if node is running in 32 or 64 bit mode.")
	fmt.Println("  jvms install <version> [arch] : The version can be a JDK version or \"latest\" for the latest stable version.")
	fmt.Println("                                 Optionally specify whether to install the 32 or 64 bit version (defaults to system arch).")
	fmt.Println("                                 Set [arch] to \"all\" to install 32 AND 64 bit versions.")
	fmt.Println("  jvms list [available]         : List the JDK installations. Type \"available\" at the end to see what can be installed. Aliased as ls.")
	fmt.Println("  jvms ls-remote                : List the JDK remote.")
	fmt.Println("  jvms on                       : Enable JDK version management.")
	fmt.Println("  jvms off                      : Disable JDK version management.")
	fmt.Println("  jvms proxy [url]              : Set a proxy to use for downloads. Leave [url] blank to see the current proxy.")
	fmt.Println("                                 Set [url] to \"none\" to remove the proxy.")
	fmt.Println("  jvms uninstall <version> <arch> : The version must be a specific version.")
	//  fmt.Println("  jvms update                   : Automatically update jvms to the latest version.")
	fmt.Println("  jvms use [version] [arch]     : Switch to use the specified version. Optionally specify 32/64bit architecture.")
	fmt.Println("                                 jvms use <arch> will continue using the selected version, but switch to 32/64 bit mode.")
	fmt.Println("  jvms root [path]              : Set the directory where jvms should store different versions of JDK.")
	fmt.Println("                                 If <path> is not set, the current root will be displayed.")
	fmt.Println("  jvms version                  : Displays the current running version of jvms for Windows. Aliased as v.")
	fmt.Println(" ")
}

// Given a jdk version, returns the associated jdk download url
func getJDKDownloadURL(jdkversion string,a string) string {
	// Get raw text
	text := web.GetRemoteTextFile("https://raw.githubusercontent.com/ystyle/jvms/master/jdkversions.json")
	// Parse
	var data interface{}
	json.Unmarshal([]byte(text), &data)
	body := data.(map[string]interface{})
	v := file.GenJDKFileName(jdkversion,a)
	return body[v].(string)
}

func updateRootDir(path string) {
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println(path + " does not exist or could not be found.")
		return
	}

	env.root = path
	saveSettings()
	fmt.Println("\nRoot has been set to " + path)
}

func saveSettings() {
	content := "root: " + strings.Trim(env.root, " \n\r") + "\r\narch: " + strings.Trim(env.arch, " \n\r") + "\r\nproxy: " + strings.Trim(env.proxy, " \n\r") + "\r\noriginalpath: " + strings.Trim(env.originalpath, " \n\r") + "\r\noriginalversion: " + strings.Trim(env.originalversion, " \n\r")+ "\r\ncurrentVersion: " + strings.Trim(env.currentVersion, " \n\r")
	ioutil.WriteFile(env.settings, []byte(content), 0644)
}

func Setup() {
	lines, err := file.ReadLines(env.settings)
	if err != nil {
		fmt.Println("\nERROR", err)
		os.Exit(1)
	}

	// Process each line and extract the value
	for _, line := range lines {
		if strings.Contains(line, "root:") {
			env.root = strings.Trim(regexp.MustCompile("root:").ReplaceAllString(line, ""), " \r\n")
		} else if strings.Contains(line, "originalpath:") {
			env.originalpath = strings.Trim(regexp.MustCompile("originalpath:").ReplaceAllString(line, ""), " \r\n")
		} else if strings.Contains(line, "originalversion:") {
			env.originalversion = strings.Trim(regexp.MustCompile("originalversion:").ReplaceAllString(line, ""), " \r\n")
		} else if strings.Contains(line, "arch:") {
			env.arch = strings.Trim(regexp.MustCompile("arch:").ReplaceAllString(line, ""), " \r\n")
		} else if strings.Contains(line, "currentVersion:") {
			env.currentVersion = strings.Trim(regexp.MustCompile("currentVersion:").ReplaceAllString(line, ""), " \r\n")
		} else if strings.Contains(line, "proxy:") {
			env.proxy = strings.Trim(regexp.MustCompile("proxy:").ReplaceAllString(line, ""), " \r\n")
			if env.proxy != "none" && env.proxy != "" {
				if strings.ToLower(env.proxy[0:4]) != "http" {
					env.proxy = "http://" + env.proxy
				}
				web.SetProxy(env.proxy)
			}
		}
	}

	env.arch = arch.Validate(env.arch)

	// Make sure the directories exist
	_, e := os.Stat(env.root)
	if e != nil {
		fmt.Println(env.root + " could not be found or does not exist. Exiting.")
		return
	}
}
