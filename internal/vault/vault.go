package vault

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	vaultApi "github.com/hashicorp/vault/api"
)

type Vault struct {
	*vaultApi.Client
}

func NewClient(address, secretId, roleId string, timeout time.Duration) *Vault {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client, err := vaultApi.NewClient(&vaultApi.Config{
		Address: address,
		HttpClient: &http.Client{
			Transport: transport,
		},
	})

	if err != nil {
		panic(fmt.Sprintf("cannot init client. %w", err))
	}

	token := getToken(secretId, roleId, client)

	client.SetToken(token)

	return &Vault{
		client,
	}
}

func getToken(secretId, roleId string, client *vaultApi.Client) string {

	payload := map[string]interface{}{
		"role_id":   roleId,
		"secret_id": secretId,
	}

	secret, err := client.Logical().Write("auth/approle/login", payload)
	if err != nil {
		panic(fmt.Errorf("cannot make request to login. err: %w", err))
	}

	if strings.TrimSpace(secret.Auth.ClientToken) == "" {
		panic("client token is empty, ping admin")
	}

	return secret.Auth.ClientToken
}

func (v *Vault) GetSecrets(path string) (map[string]interface{}, error) {
	secret, err := v.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, fmt.Errorf("secret is empty")
	}

	var secrets map[string]interface{}

	if data, ok := secret.Data["data"]; ok {
		secrets = data.(map[string]interface{})
	} else {
		secrets = secret.Data
	}

	return secrets, nil
}
