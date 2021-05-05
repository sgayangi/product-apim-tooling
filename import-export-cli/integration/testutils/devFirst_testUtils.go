/*
*  Copyright (c) WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
*
*  WSO2 Inc. licenses this file to you under the Apache License,
*  Version 2.0 (the "License"); you may not use this file except
*  in compliance with the License.
*  You may obtain a copy of the License at
*
*    http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied.  See the License for the
* specific language governing permissions and limitations
* under the License.
 */

package testutils

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wso2/product-apim-tooling/import-export-cli/integration/apim"
	"github.com/wso2/product-apim-tooling/import-export-cli/integration/base"
)

func AwsInitProject(t *testing.T, args *AWSInitTestArgs) (string, error) {
	//Setup Environment and login to it.
	base.SetupEnvWithoutTokenFlag(t, args.SrcAPIM.GetEnvName(), args.SrcAPIM.GetApimURL())
	base.Login(t, args.SrcAPIM.GetEnvName(), args.CtlUser.Username, args.CtlUser.Password)

	output, err := base.Execute(t, "aws", "init", "-n", args.ApiNameFlag, "-s", args.ApiStageNameFlag)
	return output, err
}

func ValidateAWSInitProject(t *testing.T, args *AWSInitTestArgs) {
	t.Helper()

	output, err := AwsInitProject(t, args)
	if err != nil {
		log.Fatal(err)
	}
	//Project initialized
	assert.Nil(t, err, "Error testing aws init command")
	assert.Contains(t, output, "Project initialized", "Error while executing aws init command")

	//Remove Created project and logout
	t.Cleanup(func() {
		base.RemoveDir(args.ApiNameFlag)
	})
	return
}

func InitProject(t *testing.T, args *InitTestArgs) (string, error) {
	//Setup Environment and login to it.
	base.SetupEnvWithoutTokenFlag(t, args.SrcAPIM.GetEnvName(), args.SrcAPIM.GetApimURL())
	base.Login(t, args.SrcAPIM.GetEnvName(), args.CtlUser.Username, args.CtlUser.Password)

	output, err := base.Execute(t, "init", args.InitFlag)
	return output, err
}

func InitProjectWithDefinitionFlag(t *testing.T, args *InitTestArgs) (string, error) {
	//Setup Environment and login to it.
	base.SetupEnvWithoutTokenFlag(t, args.SrcAPIM.GetEnvName(), args.SrcAPIM.GetApimURL())
	base.Login(t, args.SrcAPIM.GetEnvName(), args.CtlUser.Username, args.CtlUser.Password)

	output, err := base.Execute(t, "init", args.InitFlag, "--definition", args.DefinitionFlag)
	return output, err
}

func importApiFromProject(t *testing.T, projectName, apiName, paramsPath string, client *apim.Client, credentials *Credentials,
	isCleanup, isPreserveProvider bool) (string, error) {
	projectPath, _ := filepath.Abs(projectName)

	params := []string{"import", "api", "-f", projectPath, "-e", client.GetEnvName(), "-k",
		"--verbose", "--preserve-provider=" + strconv.FormatBool(isPreserveProvider)}

	if paramsPath != "" {
		params = append(params, "--params", paramsPath)
	}

	output, err := base.Execute(t, params...)

	base.WaitForIndexing()

	if isCleanup {
		t.Cleanup(func() {
			username, password := apim.RetrieveAdminCredentialsInsteadCreator(credentials.Username, credentials.Password)
			client.Login(username, password)
			err := client.DeleteAPIByName(apiName)

			if err != nil {
				t.Fatal(err)
			}
			base.WaitForIndexing()
		})
	}

	return output, err
}

func importApiFromProjectWithUpdate(t *testing.T, projectName string, client *apim.Client, apiName string, credentials *Credentials, isCleanup bool) (string, error) {
	projectPath, _ := filepath.Abs(projectName)
	output, err := base.Execute(t, "import", "api", "-f", projectPath, "-e", client.GetEnvName(), "-k", "--update", "--verbose")

	base.WaitForIndexing()

	if isCleanup {
		t.Cleanup(func() {
			username, password := apim.RetrieveAdminCredentialsInsteadCreator(credentials.Username, credentials.Password)
			client.Login(username, password)
			err := client.DeleteAPIByName(apiName)

			if err != nil {
				t.Fatal(err)
			}
			base.WaitForIndexing()
		})
	}

	return output, err
}

func ValidateInitializeProject(t *testing.T, args *InitTestArgs) {
	t.Helper()

	output, err := InitProject(t, args)
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err, "Error while generating Project")
	assert.Contains(t, output, "Project initialized", "Project initialization Failed")

	//Remove Created project and logout
	t.Cleanup(func() {
		base.RemoveDir(args.InitFlag)
	})
}

//Function to initialize a project using API definition
func ValidateInitializeProjectWithOASFlag(t *testing.T, args *InitTestArgs) {
	t.Helper()

	output, err := InitProjectWithOasFlag(t, args)
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err, "Error while generating Project")
	assert.Containsf(t, output, "Project initialized", "Test initialization Failed with --oas flag")

	//Remove Created project and logout

	t.Cleanup(func() {
		base.RemoveDir(args.InitFlag)
	})
}

//Function to initialize a project using API definition
func ValidateInitializeProjectWithOASFlagWithoutCleaning(t *testing.T, args *InitTestArgs) {
	t.Helper()

	output, err := InitProjectWithOasFlag(t, args)
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err, "Error while generating Project")
	assert.Containsf(t, output, "Project initialized", "Test initialization Failed with --oas flag")

}

//Function to initialize a project using API definition using --definition flag
func ValidateInitializeProjectWithDefinitionFlag(t *testing.T, args *InitTestArgs) {
	t.Helper()

	output, err := InitProjectWithDefinitionFlag(t, args)
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err, "Error while generating Project")
	assert.Containsf(t, output, "Project initialized", "Test initialization Failed with --oas flag")

	//Remove created project
	t.Cleanup(func() {
		base.RemoveDir(args.InitFlag)
	})
}

func ValidateImportProject(t *testing.T, args *InitTestArgs, paramsPath string) *apim.API {
	t.Helper()
	//Initialize a project with API definition
	ValidateInitializeProjectWithOASFlag(t, args)

	result, error := importApiFromProject(t, args.InitFlag, args.APIName, paramsPath, args.SrcAPIM, &args.CtlUser, true, true)

	assert.Nil(t, error, "Error while importing Project")
	assert.Contains(t, result, "Successfully imported API", "Error while importing Project")

	// Get App from env 2
	importedAPI := GetAPI(t, args.SrcAPIM, args.APIName, args.CtlUser.Username, args.CtlUser.Password)

	base.WaitForIndexing()

	//Remove Created project and logout
	t.Cleanup(func() {
		base.RemoveDir(args.InitFlag)
	})

	return importedAPI
}

func ValidateAWSProjectImport(t *testing.T, args *AWSInitTestArgs, isPreserveProvider bool) {
	t.Helper()

	result, error := importApiFromProject(t, args.ApiNameFlag, args.ApiNameFlag, "", args.SrcAPIM, &args.CtlUser, true, isPreserveProvider)

	assert.Nil(t, error, "Error while importing Project")
	assert.Contains(t, result, "Successfully imported API", "Error while importing Project")

	base.WaitForIndexing()

	//Remove Created project and logout
	t.Cleanup(func() {
		base.RemoveDir(args.InitFlag)
	})
}

func ValidateImportProjectFailed(t *testing.T, args *InitTestArgs, paramsPath string) {
	t.Helper()

	result, _ := importApiFromProject(t, args.InitFlag, args.APIName, paramsPath, args.SrcAPIM, &args.CtlUser, false, true)

	assert.Contains(t, result, "409", "Test failed because API is imported successfully")

	base.WaitForIndexing()

	//Remove Created project and logout
	t.Cleanup(func() {
		base.RemoveDir(args.InitFlag)
	})
}

func ValidateImportUpdateProject(t *testing.T, args *InitTestArgs) {
	t.Helper()

	result, error := importApiFromProjectWithUpdate(t, args.InitFlag, args.SrcAPIM, args.APIName, &args.CtlUser, false)

	assert.Nil(t, error, "Error while generating Project")
	assert.Contains(t, result, "Successfully imported API", "Test InitializeProjectWithDefinitionFlag Failed")

	//Remove Created project and logout
	t.Cleanup(func() {
		base.RemoveDir(args.InitFlag)
	})
}

func ValidateImportUpdateProjectNotAlreadyImported(t *testing.T, args *InitTestArgs) {
	t.Helper()

	result, error := importApiFromProjectWithUpdate(t, args.InitFlag, args.SrcAPIM, args.APIName, &args.CtlUser, true)

	assert.Nil(t, error, "Error while generating Project")
	assert.Contains(t, result, "Successfully imported API", "Test InitializeProjectWithDefinitionFlag Failed")

	//Remove Created project and logout
	t.Cleanup(func() {
		base.RemoveDir(args.InitFlag)
	})
}

func ValidateExportImportedAPI(t *testing.T, args *InitTestArgs, DevFirstDefaultAPIName string, DevFirstDefaultAPIVersion string) string {
	expOutput, expError := exportApiImportedFromProject(t, DevFirstDefaultAPIName, DevFirstDefaultAPIVersion, args.SrcAPIM.GetEnvName())
	//Check whether api is exported or not
	assert.Nil(t, expError, "Error while Exporting API")
	assert.Contains(t, expOutput, "Successfully exported API!", "Error while exporting API")
	return expOutput
}

func ValidateAPIWithDocIsExported(t *testing.T, args *InitTestArgs, DevFirstDefaultAPIName, DevFirstDefaultAPIVersion, TestCaseDestPathSuffix string) {
	expOutput := ValidateExportImportedAPI(t, args, DevFirstDefaultAPIName, DevFirstDefaultAPIVersion)

	//Unzip exported API and check whether the imported doc is in there
	exportedPath := base.GetExportedPathFromOutput(expOutput)
	relativePath := strings.ReplaceAll(exportedPath, ".zip", "")
	base.Unzip(relativePath, exportedPath)

	docPathOfExportedApi := relativePath + TestDefaultExtractedFileName + TestCaseDestPathSuffix

	//Check whether the file is available
	isDocExported := base.IsFileAvailable(t, docPathOfExportedApi)
	base.Log("Doc is Exported", isDocExported)
	assert.Equal(t, true, isDocExported, "Error while exporting API with document")

	t.Cleanup(func() {
		//Remove Created project and logout
		base.RemoveDir(args.InitFlag)
		base.RemoveDir(exportedPath)
		base.RemoveDir(relativePath)
	})
}

func ValidateAPIWithIconIsExported(t *testing.T, args *InitTestArgs, DevFirstDefaultAPIName string, DevFirstDefaultAPIVersion string) {
	expOutput := ValidateExportImportedAPI(t, args, DevFirstDefaultAPIName, DevFirstDefaultAPIVersion)

	//Unzip exported API and check whether the imported image(.png) is in there
	exportedPath := base.GetExportedPathFromOutput(expOutput)
	relativePath := strings.ReplaceAll(exportedPath, ".zip", "")
	base.Unzip(relativePath, exportedPath)

	iconPathOfExportedApi := relativePath + TestDefaultExtractedFileName + TestCase2DestPngPathSuffix

	isIconExported := base.IsFileAvailable(t, iconPathOfExportedApi)
	base.Log("Icon is Exported", isIconExported)
	assert.Equal(t, true, isIconExported, "Error while exporting API with icon")

	t.Cleanup(func() {
		//Remove Created project and logout
		base.RemoveDir(args.InitFlag)
		base.RemoveDir(exportedPath)
		base.RemoveDir(relativePath)
	})
}

func ValidateAPIWithImageIsExported(t *testing.T, args *InitTestArgs, DevFirstDefaultAPIName string, DevFirstDefaultAPIVersion string) {
	expOutput := ValidateExportImportedAPI(t, args, DevFirstDefaultAPIName, DevFirstDefaultAPIVersion)

	//Unzip exported API and check whethers the imported image(.png) is in there
	exportedPath := base.GetExportedPathFromOutput(expOutput)
	relativePath := strings.ReplaceAll(exportedPath, ".zip", "")
	base.Unzip(relativePath, exportedPath)

	imagePathOfExportedApi := relativePath + TestDefaultExtractedFileName + TestCase2DestJpegPathSuffix
	isIconExported := base.IsFileAvailable(t, imagePathOfExportedApi)
	base.Log("Image is Exported", isIconExported)
	assert.Equal(t, true, isIconExported, "Error while exporting API with icon")

	t.Cleanup(func() {
		//Remove Created project and logout
		base.RemoveDir(args.InitFlag)
		base.RemoveDir(exportedPath)
		base.RemoveDir(relativePath)
	})
}
