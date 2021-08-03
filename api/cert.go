package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type cert struct {
	PrivKey   string
	FullChain string
}

type secretConfig struct {
	SecretName   string
	SecretRegion string
}

func getSecret(secretConfig secretConfig) (string, error) {
	//Create a Secrets Manager client
	svc := secretsmanager.New(session.New(),
		aws.NewConfig().WithRegion(secretConfig.SecretRegion))
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretConfig.SecretName),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	var secretString string
	secretString = *result.SecretString

	return secretString, nil
}

func getCertAndKey(secretConfig secretConfig) (cert, error) {
	secret, err := getSecret(secretConfig)
	if err != nil {
		return cert{}, err
	}
	var cert cert
	err = json.Unmarshal([]byte(secret), &cert)
	return cert, nil
}

func writeToFile(file string, data string) {
	err := ioutil.WriteFile(file, []byte(data), 0644)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// UpdateCert pulls in the latest cert from Secrets Manager
func (a *API) UpdateCert(w http.ResponseWriter, r *http.Request) error {
	var cert cert
	var secretConfig secretConfig

	jsonDecoder := json.NewDecoder(r.Body)
	if err := jsonDecoder.Decode(&secretConfig); err != nil {
		return sendJSON(w, http.StatusInternalServerError, err.Error())
	}

	cert, err := getCertAndKey(secretConfig)
	if err != nil {
		return sendJSON(w, http.StatusInternalServerError, err.Error())
	}
	writeToFile("/etc/kong/fullChain.pem", cert.FullChain)
	writeToFile("/etc/kong/privKey.pem", cert.PrivKey)

	// restart kong to load the new config
	// need to do command as goroutine because adminapi gets killed and can't respond
	go func() {
		cmd := exec.Command("sudo", "systemctl", "reload", "kong.service")
		_, err = cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}
	}()

	return sendJSON(w, http.StatusOK, "cert updated")
}
