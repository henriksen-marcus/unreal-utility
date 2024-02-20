package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"golang.org/x/sys/windows/registry"
)

func getUprojectFile() (string, error) {
	rootPath := "."
	var uprojectPath string

	// We expect an error to be stored here because we use it to stop the walk function
	walkError := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".uproject" {
			// The file is the uproject file, stop the walk
			uprojectPath, err = filepath.Abs(path)
			if err != nil {
				return err
			}
			return filepath.SkipDir
		}

		return nil
	})

	if walkError != nil {
		return "", fmt.Errorf("Error while searching for uproject file: %v", walkError)
	}

	if uprojectPath == "" {
		fullPath, _ := filepath.Abs(rootPath)
		return "", fmt.Errorf("No uproject file found in directory: %s\nAre you sure this exe file is located in an unreal engine project?", fullPath)
	}

	return uprojectPath, nil
}

func getProjectName(uprojectPath string) (string, error) {
	fileContent, err := os.ReadFile(uprojectPath)
	if err != nil {
		return "", errors.New("Error reading file: " + uprojectPath)
	}

	var jsonData map[string]interface{}

	err = json.Unmarshal(fileContent, &jsonData)
	if err != nil {
		return "", err
	}

	// Extract project name | interface{} is a generic type, like dynamic in C#
	if modules, ok := jsonData["Modules"].([]interface{}); ok {
		// "Modules" is an array, we can iterate through it
		for _, module := range modules {
			if moduleMap, ok := module.(map[string]interface{}); ok {
				// Check if "Name" exists in each module
				if name, ok := moduleMap["Name"].(string); ok {
					return name, nil
				} else {
					fmt.Println("Name key not found in the module.")
				}
			}
		}
	} else {
		fmt.Println("Modules key not found in JSON.")
	}

	return "", errors.New("No project name found.\n")
}

func deleteFiles(projectName string) error {

	// Folders to delete
	folderList := []string{
		".vs",
		"Binaries",
		"Build",
		"Intermediate",
		"DerivedDataCache",
	}

	// Files to delete
	fileList := []string{
		projectName + ".sln",
	}

	// wwiseFolderList := []string{
	// 	"Plugins\\Wwise\\Binaries",
	// 	"Plugins\\Wwise\\Intermediate",
	// }

	// Accumulate errors during deletion
	var errorList []error

	for _, path := range folderList {
		err := os.RemoveAll(path)
		errorList = append(errorList, err)
	}

	for _, path := range fileList {
		// Check if it exists before trying to delete it
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				// File doesn't exist, return without error
				continue
			}
		}
		err = os.Remove(path)
		errorList = append(errorList, err)
	}

	// Sum errors in a string and return them
	var errorString string
	for _, err := range errorList {
		if err != nil {
			errorString += err.Error() + "\n"
		}
	}

	if errorString != "" {
		return errors.New(errorString)
	}

	return nil
}

func isPathValid(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		// File or directory exists
		return true
	}

	if os.IsNotExist(err) {
		// File or directory does not exist
		return false
	}

	// An error occurred while checking the file or directory existence
	return false
}

func getUnrealInstallationPath() (string, error) {
	// Open the registry key
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\\EpicGames\\Unreal Engine`, registry.READ)
	if err != nil {
		return "", fmt.Errorf("Error opening registry key: %v", err)
	}
	defer key.Close()

	// Get the names of all subkeys
	subkeyNames, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return "", fmt.Errorf("Error reading subkey names: %v", err)
	}

	lastSubkey := subkeyNames[len(subkeyNames)-1]

	// Open the last subkey
	lastKey, err := registry.OpenKey(key, lastSubkey, registry.READ)
	if err != nil {
		return "", fmt.Errorf("Error opening last subkey: %v", err)
	}
	defer lastKey.Close()

	// Get the value of "InstallDirectory"
	installDir, _, err := lastKey.GetStringValue("InstalledDirectory")
	if err != nil {
		return "", fmt.Errorf("Error reading InstalledDirectory value: %v", err)
	}

	return installDir, nil
}

func findFile(root, name string) (string, error) {
	var filePath string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil // Skip directories
		}
		if info.Name() == name {
			filePath = path
			return filepath.SkipDir // Stop searching after finding the file
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("Error while searching for file: %v", err)
	}

	if filePath == "" {
		return "", fmt.Errorf("File '%s' not found in directory '%s'", name, root)
	}

	return filePath, nil
}

func getUBTPath() (string, error) {
	installPath, err := getUnrealInstallationPath()

	if err != nil {
		return "", err
	}

	if !isPathValid(installPath) {
		return "", fmt.Errorf("Invalid path: %s", installPath)
	}

	ubtPath := filepath.Join(installPath, "Engine", "Binaries", "DotNET", "UnrealBuildTool", "UnrealBuildTool.exe")

	if isPathValid(ubtPath) {
		return ubtPath, nil
	}

	fmt.Print(color.YellowString("UnrealBuildTool.exe not found in default path. Searching for it..."))

	ubtPath, err = findFile(installPath, "UnrealBuildTool.exe")

	if err != nil {
		fmt.Println(color.RedString(" not found."))
		return "", err
	}

	fmt.Println(color.GreenString(" found."))

	return ubtPath, nil
}

func main() {
	titleFont := color.New(color.FgWhite, color.Bold)

	titleFont.Println("Unreal Utility v0.1\n")

	uprojectPath, err := getUprojectFile()

	if err != nil {
		color.Red(err.Error())
		return
	}

	//fmt.Println("Found uproject file: " + uprojectPath)

	projectName, err := getProjectName(uprojectPath)
	if err != nil {
		color.Red(err.Error())
	}

	fmt.Printf("Project "+color.CyanString("%v")+" was found.\n", projectName)

	fmt.Print("Deleting temporary files...")

	err = deleteFiles(projectName)
	if err != nil {
		color.Red("Error deleting files: " + err.Error())
		return
	}

	fmt.Println(" done.")

	ubtPath, err := getUBTPath()
	if err != nil {
		color.Red(err.Error())
		return
	}

	fmt.Println("Found UBT path: " + ubtPath)

	// Run UBT to generate project files
	fmt.Println("Generating project files...")
	command := exec.Command(ubtPath, uprojectPath, "-Game", "-CurrentPlatform", "-ProjectFiles")
	err = command.Run()
	if err != nil {
		color.Red("Error generating project files: " + err.Error())
		return
	}
	color.Green("Finished generating project files.")

	// Run UBT to compile the project
	fmt.Println("Compiling project...")
	command = exec.Command(ubtPath, uprojectPath, projectName+"Editor", "Development", "Win64", "-WaitMutex")

	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		color.Red("Error creating stdout pipe: %v", err)
	}

	err = command.Start()
	if err != nil {
		color.Red("Error starting compile command: %v", err)
	}

	scanner := bufio.NewScanner(stdoutPipe)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	// Check if there was an error reading from the pipe
	if err := scanner.Err(); err != nil {
		color.Red("Error reading from pipe: %v", err)
	}

	// Wait for the command to finish
	err = command.Wait()
	if err != nil {
		color.Red("Compile error: %v", err)
		return
	}

	color.Green("Finished compiling project.")
	fmt.Println("\n" + color.CyanString(projectName) + " was successfully rebuilt.")

	time.Sleep(2 * time.Second)
}
