package main

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	"context"
	"flag"
	"fmt"
	"golang.org/x/oauth2/google"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
	"log"
	"os"
	"strings"
)

func main() {

	prefix := flag.String("prefix", "secret:", "(optional) prefix of the secret defined in the env var")
	flag.Parse()

	loadSecret(*prefix)
}

func loadSecret(prefix string) {

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to setup client: %v", err)
	}

	dc, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		log.Fatalf("failed to find default credentials: %v", err)
	}
	p := dc.ProjectID

	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		k := kv[0]
		v := kv[1]
		if strings.HasPrefix(v, prefix) {
			v := v[len(prefix):]
			sver := strings.SplitN(v, "#", 2)
			s := sver[0]
			ver := "latest"
			if len(sver) == 2 {
				ver = sver[1]
			}
			plain, err := getSecret(client, p, s, ver)
			if err != nil {
				log.Printf("failed to access secret version: %v; Ignore it and continue", err)
				continue
			}
			os.Setenv(k, plain)
		}
	}
}

func getSecret(client *secretmanager.Client, project string, secret string, version string) (plain string, err error) {
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, secret, version),
	}
	ctx := context.Background()
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return
	}
	return string(result.Payload.Data), err
}
