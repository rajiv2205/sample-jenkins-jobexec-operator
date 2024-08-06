package jenkinsutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	//"os"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

/*
func Printname(name string) string {
	fmt.Println("show the name", name)
	return name
}
*/

type BuildInfo struct {
	Result  string `json:"result"`
	QueueId int    `json:"queueId"`
	ID      string `json:"id"`
}

func TriggerJobWithAndWithoutParams(jenkinsURL, jobName string, parameters map[string]string, username, apiToken string) (int, error) {
	buildURL := fmt.Sprintf("%s/job/%s/buildWithParameters", jenkinsURL, jobName)
	var req *http.Request
	var err error
	if len(parameters) == 0 {
		log.Log.Info("This is non param build job, handling accordingly")
		//fmt.Println("This is non param build job, handling accordingly")
		buildURL := fmt.Sprintf("%s/job/%s/build", jenkinsURL, jobName)
		req, err = http.NewRequest("POST", buildURL, nil)
		if err != nil {
			log.Log.Info("Error creating request for non param job:", err)
		}

	} else {
		log.Log.Info("Going to execute Parameterized job")
		params := url.Values{}
		for key, value := range parameters {
			params.Add(key, value)
		}
		//fmt.Println(params)
		req, err = http.NewRequest("POST", buildURL, strings.NewReader(params.Encode()))
		if err != nil {
			return 0, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	req.SetBasicAuth(username, apiToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf(resp.Status)
		fmt.Println("Error sending request:", err)

		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		log.Log.Info("Jenkins queue is full: %d", resp.StatusCode)

		return 0, fmt.Errorf("Queue is full")
	}
	//fmt.Printf("Executed the job, status code: %d", resp.StatusCode)
	log.Log.Info(fmt.Sprintf("Executed the job, status code: %d", resp.StatusCode))

	//fmt.Println()

	location := resp.Header.Get("Location")
	//fmt.Println(location)
	if location == "" {
		return 0, fmt.Errorf("Location header not found")
	}

	to_split := strings.Split(location, "/")
	build_id := to_split[len(to_split)-2]
	build_id_int, _ := strconv.Atoi(build_id)

	return build_id_int, nil
}

func PollQueueBuild(jenkinsURL, jobName string, buildID int, username, apiToken string) (string, error) {
	// polling jenkins queue to get the actual job url and build number of the job from the queue
	
	statusURL := fmt.Sprintf("%s/queue/item/%d/api/json", jenkinsURL, buildID)
	client := &http.Client{}
	//fmt.Println(statusURL)
	for {
		req, err := http.NewRequest("GET", statusURL, nil)

		if err != nil {
			return "", err
		}
		req.SetBasicAuth(username, apiToken)

		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		//fmt.Println(resp)
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to get build status: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		var result map[string]interface{}

		json.Unmarshal(body, &result)


		executable_interface_var := result["executable"]
		if executable_interface_var == nil {
			cancelled_interface_var := result["cancelled"]
			if cancelled_interface_var != nil {
				_, ok := cancelled_interface_var.(bool)
				if ok {
					log.Log.Info("Job cancelled by user")
					return "CANCELLED", errors.New("type assertion to bool failed")

				}
			}
			//fmt.Println("Job is already in progress, cancel the job or wait")
			log.Log.Info("Job is already in progress, cancel the job or wait")
			time.Sleep(10 * time.Second)
			continue
		}

		innerJobMap := executable_interface_var.(map[string]interface{})
		urlByte, _ := json.Marshal(innerJobMap["url"])

		if string(urlByte) != "" {
			return string(urlByte), nil
		}

	}
}

func PollBuildStatus(jobURL string, username, apiToken string) (string, string, int, error) {

	statusURL := fmt.Sprintf("%sapi/json", jobURL)

	client := &http.Client{}

	for {
		req, err := http.NewRequest("GET", statusURL, nil)

		if err != nil {
			return "", "", 0, err
		}
		req.SetBasicAuth(username, apiToken)

		resp, err := client.Do(req)
		if err != nil {
			return "", "", 0, err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return "", "", 0, fmt.Errorf("failed to get build status: %s", resp.Status)
		}

		var buildInfo BuildInfo
		if err := json.NewDecoder(resp.Body).Decode(&buildInfo); err != nil {
			return "", "", 0, err
		}

		if buildInfo.Result != "" {
			//log.Log.Info(fmt.Sprintf("Multiple values from build-id: %s and queueID: %d", buildInfo.ID, buildInfo.QueueId))
			return buildInfo.Result, buildInfo.ID, buildInfo.QueueId, nil
		}

		log.Log.Info("Build still is in progress mode, waiting for its completion")
		time.Sleep(10 * time.Second)
	}
}
