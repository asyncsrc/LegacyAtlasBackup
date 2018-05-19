package main

import (
	"github.com/akamensky/argparse"
	"os"
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"log"
	"time"
)

var (
	atlasSessionToken = ""
	sessionDownloadPath = ""
)

func getJSONFromRequest(req *http.Request) (map[string]interface{}, error) {
	client := &http.Client{}
	var response map[string]interface{}

	req.Header.Set("Cookie", fmt.Sprintf("_atlas_session_data=%s", atlasSessionToken))
	resp, err := client.Do(req)
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return response, err
	}

	if err = json.Unmarshal(data, &response); err != nil {
		return response, err
	}

	return response, nil
}

func getEnvironmentsForPage(pageNumber int) (map[string]int, error) {
	environments := make(map[string]int)

	req, err := http.NewRequest("GET", fmt.Sprintf("https://app.terraform.io/ui/environments?enterprise_tool=terraform&page=%d&username=saj", pageNumber), nil)

	if err != nil {
		return map[string]int{}, nil
	}

	jsonResponse, err := getJSONFromRequest(req)

	fmt.Println("JSON Response: ", jsonResponse)
	for _,v := range jsonResponse["environments"].([]interface{}) {
		environments[v.(map[string]interface{})["name"].(string)] = int(v.(map[string]interface{})["current_state_id"].(float64))
	}

	return environments, nil
}

func getLatestStateVersionForEnvironment(environmentName string, id int) (int, error) {
	requestString := fmt.Sprintf("https://app.terraform.io/ui//saj/environments/%s/states/%d/state-versions?page=1", environmentName, int(id))
	req, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		return -1, err
	}

	jsonResponse, err := getJSONFromRequest(req)

	fmt.Println("JSON response for latest state version: ", jsonResponse)
	if err != nil {
		return -1, nil
	}

	if (len(jsonResponse["state_versions"].([]interface{})) > 0) {
		latestStateVersion := jsonResponse["state_versions"].([]interface{})[0].(map[string]interface{})["version"]
		return int(latestStateVersion.(float64)), nil
	}

	return -1, nil
}

func downloadState(environmentName string, id int, stateVersion int) error {

	if stateVersion == -1 {
		ioutil.WriteFile(fmt.Sprintf("%s/states/%s-sessionState-%d", sessionDownloadPath, environmentName, stateVersion), []byte("No state found"), 0644)
		return nil
	}

	client := http.Client{}
	requestString := fmt.Sprintf("https://app.terraform.io/ui/saj/environments/%s/states/%d/state-versions/%d/raw", environmentName, id, stateVersion)
	req, err := http.NewRequest("GET", requestString, nil)

	if err != nil {
		return err
	}

	req.Header.Set("Cookie", fmt.Sprintf("_atlas_session_data=%s", atlasSessionToken))
	resp, err := client.Do(req)
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/states/%s-sessionState-%d", sessionDownloadPath, environmentName, stateVersion), data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	pageNumber := 1
	parser := argparse.NewParser("backup_atlas", "performs backup of all SAJ legacy states")
	c := parser.String("c", "cookie", &argparse.Options{Required: true, Help: "Cookie payload generated after authenticating with Atlas via web"})
	p := parser.String("p", "path", &argparse.Options{Required: true, Help: "Path to save session state files"})

	err := parser.Parse(os.Args)
	atlasSessionToken = *c
	sessionDownloadPath = *p

	if err != nil {
		fmt.Print(parser.Usage(err))
	}
	for {
		fmt.Println("Getting environments for page: ", pageNumber)
		environments, err := getEnvironmentsForPage(pageNumber)
		if err != nil {
			fmt.Printf("Error during inf loop.  %v\n", err)
			break
		}

		for name, id := range environments {
			fmt.Println("Getting latest state number for env: ", name, " with id: ", id)
			latestState, err := getLatestStateVersionForEnvironment(name, id)

			if err != nil {
				log.Printf("** Issue getting latest state for: %v.  Error: %v\n", latestState, err)
				continue
			}

			if err := downloadState(name, id, latestState); err != nil {
				log.Printf("** Issue downloading latest state.  %v", err)
				continue
			}
		}

		pageNumber += 1
		time.Sleep(2 * time.Second)
	}
}