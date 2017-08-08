package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/subtlepseudonym/go-prompt"
)

// *semver.Version objects can't be const, which is lame (but understandable)
var defaultVersion *semver.Version = semver.MustParse("v0.1.0")

const versionFileName string = "ver.json"
const defaultProjectName string = "GoVer Project"
const defaultVersionString string = "canteloupe"
const defaultBuild int = 0

type GoVersion struct {
	ProjectName   string          `json:"name"`
	Version       *semver.Version `json:"version"`
	VersionString string          `json:"versionString"`
	Build         int             `json:"build"`
}

func initialize() *GoVersion {
	// Check to make sure that project is not already versioned by gover
	_, err := os.OpenFile(versionFileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		fmt.Println("This project is already versioned with gover")
		fmt.Printf("Do you have a %s file in your root directory for another reason?\n", versionFileName)
		os.Exit(2)
	}

	newVersion := GoVersion{}
	newVersion.ProjectName = prompt.StringRequired("Project name (required)")

	startingVersion := prompt.String("Current version (default=0.1.0)")
	if startingVersion == "" {
		newVersion.Version = semver.MustParse("v0.1.0")
	} else {
		var err error // need to declare because we can't redeclare newVersion.Version
		newVersion.Version, err = semver.NewVersion(startingVersion)
		if err != nil {
			fmt.Println("There was an error parsing the version you provided")
			os.Exit(1)
		}
	}
	newVersion.VersionString = prompt.StringRequired("Version name (required)")

	buildNumStr := prompt.String("Current build number (default=0)")
	if buildNumStr == "" {
		newVersion.Build = 0
	} else {
		var err error
		newVersion.Build, err = strconv.Atoi(buildNumStr)
		if err != nil {
			// keep calm and carry on
			fmt.Println("There was an error parsing the build number you provided")
			os.Exit(1)
		}
	}

	if !prompt.ConfirmWithDefault("Are these the correct? (Y/n)", true) {
		fmt.Println("Aborted")
		os.Exit(0)
	}

	return &newVersion
}

// Prints current version object to ver.json
func printToFile(v *GoVersion) {
	versionBytes, err := json.MarshalIndent(*v, "", "  ")
	if err != nil {
		fmt.Println("ERROR: Unable to marshal version object")
		fmt.Println(err)
		os.Exit(1)
	}

	err = os.Rename(versionFileName, versionFileName+".bak")
	if !os.IsNotExist(err) && err != nil {
		fmt.Println("ERROR: Unable to create backup version file, aborting")
		fmt.Printf("Is there already a %s.bak file in your root directory?", versionFileName)
		fmt.Println(err)
		os.Exit(1)
	}

	verFile, err := os.Create(versionFileName)
	if err != nil {
		fmt.Println("ERROR: Unable to create new version file")
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = verFile.Write(versionBytes)
	if err != nil {
		fmt.Println("ERROR: There was an error writing to the version file, restoring from backup")
		fmt.Println(err)

		mvErr := os.Rename(versionFileName+".bak", versionFileName)
		if mvErr != nil {
			fmt.Printf("ERROR: Could not restore backup. Does %s.bak still exist?\n", versionFileName)
			fmt.Println(err)
		}
		os.Exit(1)
	}

	err = os.Remove(versionFileName + ".bak")
	if !os.IsNotExist(err) && err != nil {
		fmt.Println("ERROR: Unable to remove temporary backup")
		fmt.Println(err)
		os.Exit(1)
	}
}

func incrementMajorVersion(v *GoVersion) *GoVersion {
	newV := v.Version.IncMajor()
	v.Version = &newV
	return v
}

func incrementMinorVersion(v *GoVersion) *GoVersion {
	newV := v.Version.IncMinor()
	v.Version = &newV
	return v
}

func incrementPatchVersion(v *GoVersion) *GoVersion {
	newV := v.Version.IncPatch()
	v.Version = &newV
	return v
}

func printVersionInfo(v *GoVersion) {
	fmt.Printf("%s - %s v%s build %d\n", v.ProjectName, v.VersionString, v.Version.String(), v.Build)
}

func loadVersionInfo() *GoVersion {
	verFile, err := os.Open(versionFileName)
	if err != nil {
		fmt.Printf("ERROR: Could not find %s file\n", versionFileName)
		fmt.Println("\nHave you run `gover init` ?\n")
		os.Exit(1)
	}

	var version GoVersion
	decoder := json.NewDecoder(verFile)
	err = decoder.Decode(&version)
	if err != nil {
		fmt.Printf("ERROR: Unable to parse %s file\n", versionFileName)
		fmt.Println(err)
		os.Exit(1)
	}

	return &version
}

func main() {
	flag.Parse()
	args := os.Args[1:] // cutting off binary call

	if len(args) == 0 {
		v := loadVersionInfo()
		printVersionInfo(v)
		os.Exit(0)
	}

	if args[0] == "init" {
		v := initialize()
		printToFile(v)
		return
	}

	v := loadVersionInfo()
	switch args[0] {
	case "major":
		v = incrementMajorVersion(v)
	case "minor":
		v = incrementMinorVersion(v)
	case "patch":
		v = incrementPatchVersion(v)
	default:
		fmt.Printf("Unknown command '%s'", args[0])
		os.Exit(2)
	}
	printToFile(v)
	printVersionInfo(v)
}
