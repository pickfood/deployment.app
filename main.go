package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	path := os.Getenv("PICK_DEPLOYMENT_ROOT")
	secret := os.Getenv("PICK_DEPLOYMENT_SECRET")

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		s, ok := request.Header["Authorization"]
		if !ok || len(s) == 0 {
			writer.WriteHeader(http.StatusForbidden)
			writer.Write([]byte("no auth token"))
			return
		}
		if strings.TrimSpace(s[0]) != "Bearer "+secret {
			writer.WriteHeader(http.StatusForbidden)
			writer.Write([]byte("invalid auth token"))
			return
		}
		out, err := redeploy(path)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("unable to redeploy: %s", out)))
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(out))
	})
	log.Println(http.ListenAndServe(":8031", nil))
}

func redeploy(path string) (string, error) {
	cmd := exec.Command("make", "-f", path, "all")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err.Error(), err
	}

	if err := cmd.Start(); err != nil {
		return err.Error(), err
	}

	slurp, err := ioutil.ReadAll(stderr)
	if err != nil {
		return err.Error(), err
	}

	if err := cmd.Wait(); err != nil {
		return string(slurp), err
	}
	return string(slurp), err
}
