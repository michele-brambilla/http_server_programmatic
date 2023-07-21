package main

import (
	"errors"
	"flag"
	"fmt"
	yaml "gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const keyServerAddr = "serverAddr"

var data map[string]interface{}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	hasService := r.URL.Query().Has("service")

	if hasService == false {
		w.Header().Set("x-missing-field", "service")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("Error\n")
		return
	}

	service := r.URL.Query().Get("service")
	service_config := fmt.Sprintf("service.%s", service)

	bash_script, ok := data[service_config].(string)
	if !ok {
		fmt.Println("missing service")
		return
	}

	if !fileExists(bash_script) {
		fmt.Printf("File %s does not exist\n", bash_script)
		return
	}

	out, err := exec.Command("/bin/bash", bash_script).Output()
	if err != nil {
		w.Header().Set("x-missing-field", "service")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("Error")
		return
	}

	fmt.Printf("%s: got /health request. service(%t)=%s\n",
		ctx.Value(keyServerAddr),
		hasService, service)

	reply := string(out)
	fmt.Printf("service %s: %s\n", service, string(out))
	io.WriteString(w, reply)
}

func main() {
	var configuration_file string
	var help bool

	flag.StringVar(&configuration_file, "c", "./resources/config.yaml", "Configuration file")
	flag.BoolVar(&help, "help", false, "Help")
	flag.Parse()

	if help {
		flag.PrintDefaults()
	} else {
		f, err := os.ReadFile(configuration_file)
		if err != nil {
			log.Fatal(err)
		}

		err = yaml.Unmarshal(f, &data)
		if err != nil {
			log.Fatal(err)
		}

		http.HandleFunc("/health", getHealth)

		listen_on := fmt.Sprintf(":%d", data["port"])
		err = http.ListenAndServe(listen_on, nil)
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed\n")
		} else if err != nil {
			fmt.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}
	}
}
