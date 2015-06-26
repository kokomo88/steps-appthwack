package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	//"path/filepath"
	"time"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

type Device struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Os_version string `json:"os_version"`
}

type DevicePool struct {
	Id       int      `json:"id"`
	Name     string   `json:"name"`
	Readonly string   `json:"readonly"`
	Devices  []Device `json:"devices"`
}

type Project struct {
	Id              int    `json:"id"`
	Url             string `json:"url"`
	Name            string `json:"name"`
	Project_type_id int    `json:"project_type_id"`
}

type UploadAppParams struct {
	Name string `json:"name"`
	File string `json:"file"`
	Save string `json:"save"`
	Type string `json:type`
}

type UploadedFile struct {
	Id int `json:"file_id"`
}

type runId struct {
	Id int `json:"run_id"`
}
type status struct {
	Status string `json:"status"`
}

type DevicePoolList []DevicePool
type ProjectList []Project

//type UploadedFileList []UploadedFile

func createUploadAppPayloadParam(name string, fileName string, save string, fileType string) (UploadAppParams, error) {
	uploadParams := UploadAppParams{
		Name: name,
		File: fileName,
		Save: save,
		Type: fileType,
	}

	return uploadParams, nil
}

func sendGetRequest(requestURL string, apiKey string) string {
	req, err := http.NewRequest("GET", requestURL, nil)
	req.SetBasicAuth(apiKey, "")
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		//TODO
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//TODO
		log.Fatal(err)

	}

	resp.Body.Close()
	bodyStr := string(body)
	if resp.StatusCode != 200 {
		var msg ErrorMessage
		json.Unmarshal([]byte(bodyStr), &msg)
		fmt.Println(msg.Message)
		fmt.Printf("REQUEST ERROR: %d", resp.StatusCode)
		switch resp.StatusCode {
		case 400:
			log.Fatal(" BAD REQUEST")
		case 401:
			log.Fatal(" UNAUTHORIZED")
		case 402:
			log.Fatal(" REQUEST FAILED")
		case 404:
			log.Fatal(" NOT FOUND")
		case 500:
			log.Fatal(" SERVER ERROR")
		case 501:
			log.Fatal(" NOT IMPLEMENTED")

		}

	}
	return bodyStr
}

func sendPostRequestWithFileUpload(apiKey, requestURL string, params UploadAppParams) (str string, err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add file
	f, err := os.Open(params.File)
	if err != nil {
		log.Fatal("Failed to open file", params.File, err)
	}
	fw, err := w.CreateFormFile("file", params.File)
	if err != nil {
		log.Fatal("Failed to create form", err)
	}
	if _, err = io.Copy(fw, f); err != nil {
		return "", err
	}
	// Add the other fields
	if fw, err = w.CreateFormField("name"); err != nil {
		log.Fatal("Failed to create form", err)
	}
	if _, err = fw.Write([]byte(params.Name)); err != nil {
		log.Fatal("Failed to create form", err)
	}
	if fw, err = w.CreateFormField("type"); err != nil {
		log.Fatal("Failed to create form", err)
	}
	if _, err = fw.Write([]byte(params.Type)); err != nil {
		log.Fatal("Failed to create form", err)
	}
	if fw, err = w.CreateFormField("save"); err != nil {
		log.Fatal("Failed to create form", err)
	}
	if _, err = fw.Write([]byte(params.Save)); err != nil {
		log.Fatal("Failed to create form", err)
	}

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", requestURL, &b)
	if err != nil {
		log.Fatal("Failed to create request", err)
	}
	//It's needed for http authentication.
	req.SetBasicAuth(apiKey, "")
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send the request", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//TODO
		log.Fatal(err)
	}
	res.Body.Close()
	bodyStr := string(body)

	if res.StatusCode != 200 {
		var msg ErrorMessage
		json.Unmarshal([]byte(bodyStr), &msg)
		fmt.Println(msg.Message)
		fmt.Printf("REQUEST ERROR: %d", res.StatusCode)
		switch res.StatusCode {
		case 400:
			log.Fatal(" BAD REQUEST")
		case 401:
			log.Fatal(" UNAUTHORIZED")
		case 402:
			log.Fatal(" REQUEST FAILED")
		case 404:
			log.Fatal(" NOT FOUND")
		case 500:
			log.Fatal(" SERVER ERROR")
		case 501:
			log.Fatal(" NOT IMPLEMENTED")

		}
	}
	return bodyStr, err
}

func searchProjectIdByName(l ProjectList, name string) (id int) {
	for _, v := range l {
		if v.Name == name {
			return v.Id
		}
	}
	return -1
}

func searchPoolIdByName(l DevicePoolList, name string) (id int) {
	for _, v := range l {
		if v.Name == name {
			return v.Id
		}
	}
	return -1
}

func scheduleTest(apiKey string, projectId int, projectName string, fileId int, poolId int, requestURL string, testType string, testFileId int) (str string, err error) {
	// project, name, app, pool
	// s := []string{`"project":`, strconv.Itoa(projectId), `,"name":"`, projectName, `","app":`, strconv.Itoa(fileId), `,"pool":`, strconv.Itoa(poolId)}
	data := url.Values{}
	data.Set("project", strconv.Itoa(projectId))
	data.Set("name", projectName)
	data.Set("app", strconv.Itoa(fileId))
	data.Set("pool", strconv.Itoa(poolId))
	if testType != "built-in" {
		if testType == "kif" {
			data.Set(testType, "")
		} else {
			data.Set(testType, strconv.Itoa(testFileId))
		}
	}
	req, err := http.NewRequest("POST", requestURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.SetBasicAuth(apiKey, "")
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal("Failed to send the request", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//TODO
		log.Fatal(err)

	}
	resp.Body.Close()
	bodyStr := string(body)
	if resp.StatusCode != 200 {
		var msg ErrorMessage
		json.Unmarshal([]byte(bodyStr), &msg)
		fmt.Println(msg.Message)
		fmt.Printf("REQUEST ERROR: %d", resp.StatusCode)
		switch resp.StatusCode {
		case 400:
			log.Fatal(" BAD REQUEST")
		case 401:
			log.Fatal(" UNAUTHORIZED")
		case 402:
			log.Fatal(" REQUEST FAILED")
		case 404:
			log.Fatal(" NOT FOUND")
		case 500:
			log.Fatal(" SERVER ERROR")
		case 501:
			log.Fatal(" NOT IMPLEMENTED")

		}
	}
	return bodyStr, err

}

func retrieveResults(apiKey string, pid int, rid int, done chan string) {
	requestURL := "https://appthwack.com/api/run/" + strconv.Itoa(pid) + "/" + strconv.Itoa(rid) + "/status"
	stringifiedJSON := sendGetRequest(requestURL, apiKey)

	var stat status

	json.Unmarshal([]byte(stringifiedJSON), &stat)
	if stat.Status == "completed" {
		done <- stat.Status
	}

}

func main() {

	apiKey := os.Getenv("APPTHWACK_API_KEY")
	if apiKey == "" {
		log.Fatal("$APPTHWACK_API_KEY is not provided!")
	}

	projectName := os.Getenv("APPTHWACK_PROJECT_NAME")
	devicePoolName := os.Getenv("APPTHWACK_DEVICE_POOL_NAME")
	if projectName == "" {
		log.Fatal("$APPTHWACK_PROJECT_NAME is not provided!")
	}
	if devicePoolName == "" {
		log.Fatal("$APPTHWACK_DEVICE_POOL_NAME is not provided!")
	}

	uploadName := os.Getenv("APPTHWACK_UPLOAD_NAME")
	uploadFile := os.Getenv("APPTHWACK_UPLOAD_FILE")
	uploadSave := os.Getenv("APPTHWACK_UPLOAD_SAVE")
	uploadType := os.Getenv("APPTHWACK_UPLOAD_TYPE")
	if uploadName == "" {
		log.Fatal("$APPTHWACK_UPLOAD_NAME is not provided!")
	}
	if uploadFile == "" {
		log.Fatal("$APPTHWACK_UPLOAD_FILE is not provided!")
	}
	if uploadSave == "" {
		log.Fatal("$APPTHWACK_UPLOAD_SAVE is not provided!")
	}
	if uploadType == "" {
		log.Fatal("$APPTHWACK_UPLOAD_TYPE is not provided!")
	}

	testUploadName := ""
	testUploadFile := ""
	testUploadSave := ""
	testUploadType := os.Getenv("APPTHWACK_TEST_UPLOAD_TYPE")
	if uploadName == "" {
		log.Fatal("$APPTHWACK_TEST_UPLOAD_TYPE is not provided!")
	}

	if testUploadType != "built-in" {
		testUploadName = os.Getenv("APPTHWACK_TEST_UPLOAD_NAME")
		testUploadFile = os.Getenv("APPTHWACK_TEST_UPLOAD_FILE")
		testUploadSave = os.Getenv("APPTHWACK_TEST_UPLOAD_SAVE")

		if uploadName == "" {
			log.Fatal("$APPTHWACK_TEST_UPLOAD_NAME is not provided!")

		}
		if uploadName == "" {
			log.Fatal("$APPTHWACK_TEST_UPLOAD_FILE is not provided!")

		}
		if uploadName == "" {
			log.Fatal("$APPTHWACK_TEST_UPLOAD_SAVE is not provided!")

		}
	}

	urlDevice := "https://appthwack.com/api/devicepool"
	urlProject := "https://appthwack.com/api/project"
	urlUpload := "https://appthwack.com/api/file"
	urlSchedule := "https://appthwack.com/api/run"

	stringifiedJSON := sendGetRequest(urlProject, apiKey)
	//getting projects
	var resProjList ProjectList

	json.Unmarshal([]byte(stringifiedJSON), &resProjList)
	//fmt.Printf("%#v\n", resProjList)

	//getting device pools
	stringifiedJSON = sendGetRequest(urlDevice, apiKey)

	var resDev DevicePoolList

	json.Unmarshal([]byte(stringifiedJSON), &resDev)
	//fmt.Printf("%#v\n", resDev)

	//getting Id-s
	projectId := (searchProjectIdByName(resProjList, projectName))
	if projectId == -1 {
		//TODO
		log.Fatal("Invalid project name:", projectName)
	}
	poolId := (searchPoolIdByName(resDev, devicePoolName))
	if poolId == -1 {
		//TODO
		log.Fatal("Invalid device pool name:", devicePoolName)
	}

	//uploading app
	fileUpload, err := createUploadAppPayloadParam(uploadName, uploadFile, uploadSave, uploadType)
	if err != nil {
		//TODO
		log.Fatal(err)
	}
	stringifiedJSON, err = sendPostRequestWithFileUpload(apiKey, urlUpload, fileUpload)

	var uploadedFiles UploadedFile

	json.Unmarshal([]byte(stringifiedJSON), &uploadedFiles)

	var uploadedTestFiles UploadedFile

	if testUploadType != "built-in" {
		fileUpload, err := createUploadAppPayloadParam(testUploadName, testUploadFile, testUploadSave, testUploadType)
		if err != nil {
			//TODO
			log.Fatal(err)
		}
		stringifiedJSON, err = sendPostRequestWithFileUpload(apiKey, urlUpload, fileUpload)

		json.Unmarshal([]byte(stringifiedJSON), &uploadedTestFiles)
	}

	//schedule Test

	stringifiedJSON, err = scheduleTest(apiKey, projectId, projectName, uploadedFiles.Id, poolId, urlSchedule, testUploadType, uploadedTestFiles.Id)

	var runIds runId

	json.Unmarshal([]byte(stringifiedJSON), &runIds)

	done := make(chan string, 1)
	ticker := time.NewTicker(time.Millisecond * 2000).C
	fmt.Printf("Waiting for test results")
	counter := 0
	for {
		select {
		case <-ticker:
			go retrieveResults(apiKey, projectId, runIds.Id, done)
			fmt.Printf(".")
			if counter == 9 {
				counter = 0
				fmt.Printf("\n")

			}
		case msg := <-done:
			fmt.Println("\n", "status:", msg)
			return
		}
	}

}
