package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

const keyServerAddr = "serverAddr"

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
	bash_script := fmt.Sprintf("./resources/%s.sh", service)
 	if !fileExists(bash_script) {
      fmt.Printf("File %s does not exist\n", bash_script)
      return
   	}

	fmt.Printf("./resources/%s.sh\n", service)
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

	// var reply string
	reply := string(out)
	fmt.Printf("service %s: %s\n", service, string(out))
	io.WriteString(w, reply)
}

func main() {
	http.HandleFunc("/health", getHealth)

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
