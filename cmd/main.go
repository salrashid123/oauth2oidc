package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/salrashid123/oauth2oidc"
)

var (
	flADCFile  = flag.String("adc_file", "", "(required) file to application default credentials file")
	flAudience = flag.String("audience", "", "(required) Audience for the token")
)

type ADC struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	Type         string `json:"type"`
}

func main() {

	flag.Parse()
	if *flADCFile == "" {
		flag.PrintDefaults()
		log.Fatalf("--client_secrets_file must be set")
	}

	if *flAudience == "" {
		flag.PrintDefaults()
		log.Fatalf("--audience must be set")
	}

	var tok ADC

	t, err := os.Open(*flADCFile)
	if err != nil {
		log.Fatalf("Could not open credential File %v", err)
	}
	defer t.Close()

	err = json.NewDecoder(t).Decode(&tok)
	if err != nil {
		log.Fatalf("Could not parse credential File %v", err)
	}

	r, err := oauth2oidc.GetIdToken(*flAudience, tok.ClientID, tok.ClientSecret, tok.RefreshToken)
	if err != nil {
		log.Fatalf("Could not acquire id_token.  Verify the client_id and audience client_id are in the same GCP project --\n%v", err)
		return
	}

	fmt.Printf("%s\n", r.IDToken)

	// optionally use the token to call and endpoint (eg an application behind IAP)
	// url := "https://core-eso.uc.r.appspot.com/"

	// client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
	// 	AccessToken: r.IDToken,
	// 	TokenType:   "Bearer",
	// },
	// ))
	// if err != nil {
	// 	log.Fatalf("Could not generate NewClient: %v", err)
	// }

	// req, err := http.NewRequest(http.MethodGet, url, nil)
	// if err != nil {
	// 	log.Fatalf("Error Creating HTTP Request: %v", err)
	// }
	// resp, err := client.Do(req)
	// if err != nil {
	// 	log.Fatalf("Error making authenticated call: %v", err)
	// }
	// bodyBytes, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatalf("Error Reading response body: %v", err)
	// }
	// bodyString := string(bodyBytes)
	// log.Printf("Authenticated Response: %v", bodyString)
}
