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
 * pc@2018/12/19
 */

package main

import (
	"flag"
	"os"

	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"net/http"

	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var port string

func init() {
	flag.StringVar(&port, "port", "8007", "listen to the given port.")
	os.Setenv("APP_RUN_ENV", "dev")
	os.Setenv("DOCKER_API_VERSION", "1.37")
}

func main() {
	flag.Parse()
	http.HandleFunc("/", index)
	http.HandleFunc("/svc", handlerServices)

	log.Println("Listening on port *:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

/*
 * curl 127.0.0.1/index
 *
 */

//Project A Project
type Project struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

//Projects A list of Projects
type Projects struct {
	Env  string    `json:"env"`
	Data []Project `json:"data"`
}

func index(w http.ResponseWriter, r *http.Request) {
	projectsData := []Project{
		Project{
			"demoproject",
			"1",
		},
		Project{
			"demo1",
			"0",
		},
	}

	projects := Projects{
		"dev",
		projectsData,
	}

	if r, err := json.Marshal(projects); err == nil {
		log.Println("[query-projects] response:", string(r))
	}
	json.NewEncoder(w).Encode(projects)
}

/*
 * curl 127.0.0.1/svc
 *
 */

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

func handlerServices(w http.ResponseWriter, r *http.Request) {
	var svcs Services

	env := os.Getenv("APP_RUN_ENV")
	projectName := "demoproject"

	servicePrefix := env + "-" + projectName
	log.Println("[filter-service] name =", servicePrefix)

	svcs.Env = env
	svcs.ProjectName = projectName

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
