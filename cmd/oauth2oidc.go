package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/salrashid123/oauth2oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	flCredentialFile    = flag.String("credential_file", "creds.json", "Credential file with access_token, refresh_token")
	flClientSecretsFile = flag.String("client_secrets_file", "", "(required) client secrets json file")
	flAudience          = flag.String("audience", "", "(required) Audience for the token")
	flUseCache          = flag.Bool("use_cache", false, "force a new token")
)

func main() {

	flag.Parse()
	if *flClientSecretsFile == "" {
		flag.PrintDefaults()
		log.Fatalf("--client_secrets_file must be set")
	}

	if *flAudience == "" {
		flag.PrintDefaults()
		log.Fatalf("--audience must be set")
	}

	b, err := ioutil.ReadFile(*flClientSecretsFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	conf, err := google.ConfigFromJSON(b, oauth2oidc.UserInfoEmailScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	var refreshToken string
	var tok oauth2oidc.TokenResponse
	_, err = os.Stat(*flCredentialFile)
	if *flCredentialFile == "" || os.IsNotExist(err) {
		lurl := conf.AuthCodeURL("code")
		fmt.Printf("\nVisit the URL for the auth dialog and enter the authorization code  \n\n%s\n", lurl)
		fmt.Printf("\nEnter code:  ")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()

		tok, err := conf.Exchange(oauth2.NoContext, input.Text())
		if err != nil {
			log.Fatalf("Cloud not exchange TOken %v", err)
		}
		refreshToken = tok.RefreshToken
	} else {
		f, err := os.Open(*flCredentialFile)
		if err != nil {
			log.Fatalf("Could not open credential File %v", err)
		}
		defer f.Close()

		err = json.NewDecoder(f).Decode(&tok)
		if err != nil {
			log.Fatalf("Could not parse credential File %v", err)
		}
		refreshToken = tok.RefreshToken

		if !*flUseCache {
			var parser *jwt.Parser
			parser = new(jwt.Parser)
			tt, _, err := parser.ParseUnverified(tok.IDToken, &jwt.StandardClaims{})
			if err != nil {
				log.Fatalf("Could not parse saved id_tokne File %v", err)
			}

			c, ok := tt.Claims.(*jwt.StandardClaims)
			err = tt.Claims.Valid()
			if ok && err == nil {
				if c.Audience == *flAudience {
					fmt.Printf("%s\n", tt.Raw)
					return
				}
			}
		}

	}
	r, err := oauth2oidc.GetIdToken(*flAudience, conf.ClientID, conf.ClientSecret, refreshToken)
	if err != nil {
		log.Fatalf("Could not parse saved id_tokne File %v", err)
	}

	tok.IDToken = r

	f, err := os.OpenFile(*flCredentialFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Could not parse saved id_tokne File %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(tok)

	fmt.Printf("%s\n", r)
}
