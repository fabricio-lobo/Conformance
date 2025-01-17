package utils

import (
	"fmt"
	"strings"

	"github.com/br-openinsurance/Conformance/tree/main/conformance_table_generator/models"
)

func GenerateTable(apisList []string, phase string, version string) {
	// import files
	repositoryUrl := "https://api.github.com/repos/br-openinsurance/Conformance/git/trees/main?recursive=1"
	submissionFiles := importSubmittedFiles(repositoryUrl)

	// filter files by chosen apis and version
	filteredFiles := filterFilesByApisAndVersion(submissionFiles, apisList)

	// create table and dump
	tableHeaders := []string {"Organização", "Deployment"}
	tableHeaders = append(tableHeaders, apiTableHeaders(apisList)...)
	table := [][]string {tableHeaders}

	dumpHeaders := []string {"Id da Organização", "Deployment", "API", "Version", "Data"}
	dump := [][]string {dumpHeaders}

	organisationsMap := makeOrganisationsMap(false)

	for _, file := range filteredFiles {
		fileSplit := strings.Split(file, "/")
		fileName := fileSplit[len(fileSplit) - 1]
		fileNameSplit := strings.Split(fileName, "_")
		orgId := fileNameSplit[0]
		orgName := organisationsMap[orgId]
		deploymentName := fileNameSplit[1]
		api := fileNameSplit[2]
		version := fileNameSplit[3]
		if len(version) == 2 {
			version += ".0"
		}
		date := strings.Split(fileNameSplit[4], ".")[0]

		zipUrl := strings.Replace(repositoryUrl, "api.github.com/repos", "github.com", 1)
		zipUrl = strings.Replace(zipUrl, "git/trees/main?recursive=1", "blob/main/" + file, 1)
		zipUrl = strings.Replace(zipUrl, " ", "%20", -1)

		dump = append(dump, []string {
			orgId,
			deploymentName,
			api,
			version,
			date,
		})

		apiIndex := findApiIndex(apisList, translateNameFromFileNameToApisList(api, version))
		if ind := searchFileInTable(table, orgName, deploymentName); ind == -1 {
			newRow := make([]string, len(tableHeaders))
			newRow[0] = orgName
			newRow[1] = deploymentName
			newRow[apiIndex + 2] = fmt.Sprintf("[%s](%s)", date, zipUrl)
			
			table = append(table, newRow)
		} else {
			table[ind][apiIndex + 2] = fmt.Sprintf("[%s](%s)", date, zipUrl)
		}
	}

	dumpFileName := fmt.Sprintf("../results/%s/%s/%s-%s-conformance-dump.csv", phase, version, phase, version)
	exportTable(dump, dumpFileName)

	tableFileName := fmt.Sprintf("../results/%s/%s/%s-%s-conformance-table.csv", phase, version, phase, version)
	exportTable(table, tableFileName)
}

func filterFilesByApisAndVersion(submissionFiles models.GithubTree, apisList []string) []string {
	var filteredFiles []string

	for _, file := range submissionFiles {
		filePath := file.Path
		fileSplit := strings.Split(filePath, "/")
		fileApi := fileSplit[2]
		fileVersion := fileSplit[3]

		apiIndex := findApiIndex(apisList, translateNameFromFileToApisList(fileApi, fileVersion))
		
		if apiIndex != -1 {
			listVersion := strings.Split(apisList[apiIndex], "_")[1][1:]
			isVersionEqual := fileVersion == listVersion + ".0"
			isOldVersion := strings.HasSuffix(listVersion, "-old")
			isOldVersionEqual := strings.Replace(fileVersion, ".0", "", 1) == listVersion

			if isOldVersion {
				fmt.Println(fileVersion, listVersion, isOldVersionEqual)
			}

			if isOldVersion && isOldVersionEqual || isVersionEqual {
				filteredFiles = append(filteredFiles, filePath)
			}
		}
	}

	return filteredFiles
}

func findApiIndex(apis []string, api string) int {
	for i, element := range apis {
		if element == api || element == api + "-old" {
			return i
		}
	}
	return -1
}

func apiTableHeaders(apisList []string) []string {
	var tableHeaders []string
	for _, apiElement := range apisList {
		apiSplit := strings.Split(apiElement, "_")
		api := apiSplit[0]
		version := apiSplit[1]

		header := strings.ReplaceAll(api, "-", " ") + " " + version
		tableHeaders = append(tableHeaders, header)
	}
	return tableHeaders
}

func translateNameFromFileToApisList(api string, version string) string {
	return api + "_v" + strings.Replace(version, ".0", "", 1)
}

func translateNameFromFileNameToApisList(api string, version string) string {
	if len(strings.Split(version, "-")[0]) == 2 {
		version = "v1.0"
	}
	return api + "_" + strings.Split(version, "-")[0]
}
