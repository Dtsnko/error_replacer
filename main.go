package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	directory *string
)

func main() {
	DefineFlags()
	flag.Parse()
	// files, err := ioutil.ReadDir(*directory)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	err := filepath.Walk(*directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process only regular files, skip directories
		if !info.Mode().IsRegular() {
			return nil
		}

		// Perform the error constructor replacements
		err = replaceErrorConstructors(path)
		if err != nil {
			fmt.Printf("Error replacing error constructors in file %s: %s\n", path, err)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

}

func DefineFlags() {
	directory = flag.String("directory", "", "directory")
}

func replaceErrorConstructors(file string) error {
	// Read the file content
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	// Define the regular expression pattern to match the old error constructors
	pattern := `(model\.)?NewAppError\(\s?(.*?),\s+(.*?),\s+(.*?),\s+(.*?),\s+((?:http.*?|extractCodeFromErr\(.*?\)))\)`

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Find all matches of the pattern in the file content
	matches := regex.FindAllStringSubmatch(string(content), -1)

	// Iterate through the matches and perform the replacements
	for _, match := range matches {
		oldConstructor := match[0]
		model := match[1]
		where := match[2]
		id := match[3]
		params := match[4]
		details := match[5]
		httpStatus := match[6]

		newConstructor := constructNewErrorConstructor(model, where, id, params, details, httpStatus)

		// Replace the old constructor with the new constructor in the file content
		content = []byte(strings.Replace(string(content), oldConstructor, newConstructor, -1))
	}

	// Write the modified content back to the file
	err = ioutil.WriteFile(file, content, 0644)
	if err != nil {
		return err
	}

	return nil
}

func constructNewErrorConstructor(model, where, id, params, details, httpStatus string) string {
	var res string
	switch httpStatus {
	case "http.StatusNotFound":
		res = fmt.Sprintf("%sNewNotFoundError(%s, %s)", model, id, details)
	case "http.StatusBadRequest":
		res = fmt.Sprintf("%sNewBadRequestError(%s, %s)", model, id, details)
	case "http.StatusForbidden":
		res = fmt.Sprintf("%sNewForbiddenError(%s, %s)", model, id, details)
	case "http.Unauthorized":
		res = fmt.Sprintf("%sNewUnauthorizedError(%s, %s)", model, id, details)
	default:
		if strings.Contains(httpStatus, "extractCodeFromErr") {
			res = fmt.Sprintf("%sNewCustomCodeError(%s, %s, %s)", model, id, details, httpStatus)
		} else {
			res = fmt.Sprintf("%sNewInternalError(%s, %s)", model, id, details)
		}
	}
	if params != "" && params != "nil" {
		res += fmt.Sprintf(".SetTranslationParams(%s)", params)
	}

	return res
}
