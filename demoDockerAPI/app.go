/*
 * docker api and sdk exp
 * api ref: https://docs.docker.com/engine/api/v1.37/
 * sdk go: https://godoc.org/github.com/docker/docker/client
 *
 * Server:
 *  Engine:
 *   Version:      18.03.1-ce
 *   API version:  1.37 (minimum version 1.12)
 *
 * [howto]
 * # curl -s --unix-socket /var/run/docker.sock http:/v1.37/services |jq . |more
 *
 * pc@2018/12/20
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"net/http"

	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

//Project desc a project
type Project struct {
	Icon   string `json:"icon"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

//Projects desc a list of projects
type Projects struct {
	Env  string    `json:"env"`
	Data []Project `json:"data"`
}

//ParseToken save payload from http.Request
type ParseToken struct {
	AccessToken string `json:"accessToken"`
}

//ParseProject save payload from http.Request
type ParseProject struct {
	ProjectName string `json:"projectName"`
	AccessToken string `json:"accessToken"`
}

//Service A docker swarm service
type Service struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Replicas string `json:"replicas"`
	Image    string `json:"image"`
}

//Services A list of docker swarm services
type Services struct {
	Env         string    `json:"env"`
	ProjectName string    `json:"projectName"`
	Data        []Service `json:"data"`
}

var activeAccessToken = map[string]bool{
	"xxx": true,
	"yyy": true,
}

//😇
var activeProjects = Projects{
	"dev",
	[]Project{
		Project{
			"😇",
			"demo1",
			"1",
		},
		Project{
			"😇",
			"demoproject",
			"1",
		},
	},
}

// curl 127.0.0.1
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "EMPTY\n")
}

// curl -s -X POST -H "Content-Type:application/json" -d '{"accessToken":"xxx"}' 127.0.0.1/project
func projectHandler(w http.ResponseWriter, r *http.Request) {
	var data ParseToken
	payload, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Println("[json] failed to Unmarshal the payload!")
		return
	}
	if _, ok := activeAccessToken[data.AccessToken]; !ok {
		log.Println("[AccessToken] invalid Token!")
		return
	}
	if r, err := json.Marshal(activeProjects); err == nil {
		log.Println("[query-projects] response:", string(r))
	}
	json.NewEncoder(w).Encode(activeProjects)
}

// curl -s -X POST -H "Content-Type:application/json" -d '{"accessToken":"xxx","projectName":"demoproject"}' 127.0.0.1/service
func serviceHandler(w http.ResponseWriter, r *http.Request) {
	var svcs Services
	var data ParseProject

	env := os.Getenv("APP_RUN_ENV")

	payload, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Println("[json] failed to Unmarshal the payload!")
		return
	}
	if _, ok := activeAccessToken[data.AccessToken]; !ok {
		log.Println("[AccessToken] invalid Token!")
		return
	}

	servicePrefix := env + "-" + data.ProjectName

	svcs.Env = env
	svcs.ProjectName = data.ProjectName

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	//------ list service with filter
	fSvc := filters.NewArgs()
	fSvc.Add("name", servicePrefix)
	services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{Filters: fSvc})
	if err != nil {
		panic(err)
	}

	for _, s := range services {
		svc := Service{}

		//------ list task with filter
		fTask := filters.NewArgs()
		fTask.Add("service", s.ID)
		fTask.Add("desired-state", "running")
		tasks, err := cli.TaskList(context.Background(), types.TaskListOptions{Filters: fTask})
		if err != nil {
			panic(err)
		}

		svc.ID = s.ID[:10]
		svc.Name = s.Spec.Name
		svc.Replicas = fmt.Sprintf("%d/%s", len(tasks), strconv.FormatUint(*s.Spec.Mode.Replicated.Replicas, 10))
		image := strings.Split(strings.Split(s.Spec.TaskTemplate.ContainerSpec.Image, "@")[0], "/")
		svc.Image = image[len(image)-1]
		svcs.Data = append(svcs.Data, svc)
	}

	if r, err := json.Marshal(svcs); err == nil {
		log.Println("[query-services] response:", string(r))
	}
	json.NewEncoder(w).Encode(svcs)
}

var port string

func init() {
	flag.StringVar(&port, "port", "80", "listen to the given port.")
	os.Setenv("APP_RUN_ENV", "dev")
	os.Setenv("DOCKER_API_VERSION", "1.37")
}

func main() {
	flag.Parse()
	http.HandleFunc("/", index)
	http.HandleFunc("/project", projectHandler)
	http.HandleFunc("/service", serviceHandler)

	log.Println("Listening on port *:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
