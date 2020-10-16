package main

/*
 Acquire google issued OIDC token for user-based credentials for a given audience.
   To use, first download a client_secret.json for installed app flow (desktop):
   https://cloud.google.com/iap/docs/authentication-howto#authenticating_from_a_desktop_app

   specify the audience you would like
   run login flow on browser.  You refresh_token will be saved into credential_file so that you are not repeatedly running the login flows (be careful with this)

   go run oauth2oidc.go --audience=1071284184436-vu96hfaugnm9falak0pl00ur9cuvldl2.apps.googleusercontent.com --credential_file=creds.json --client_secrets_file=client_secret.json
*/

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

var (
	flCredentialFile    = flag.String("credential_file", "creds.json", "Credential file with access_token, refresh_token")
	flClientSecretsFile = flag.String("client_secrets_file", "client_secrets.json", "(required) client secrets json file")
	flAudience          = flag.String("audience", "", "(required) Audience for the token")
	flUseCache          = flag.Bool("use_cach", false, "force a new token")
)

const (
	userInfoEmailScope = "https://www.googleapis.com/auth/userinfo.email"
	tokenURL           = "https://oauth2.googleapis.com/token"
)

func GetIdToken(audience, clientId, clientSecret, refreshToken string) (idToken string, err error) {
	data := url.Values{
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
		"audience":      {audience},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", errors.New(fmt.Sprintf("Error exchaning token: %s", b))
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", nil
	}

	tokenRes := &tokenResponse{}

	if err := json.Unmarshal(body, tokenRes); err != nil {
		return "", nil
	}
	token := &oauth2.Token{
		AccessToken: tokenRes.AccessToken,
		TokenType:   tokenRes.TokenType,
	}
	raw := make(map[string]interface{})
	json.Unmarshal(body, &raw)
	token = token.WithExtra(raw)

	if secs := tokenRes.ExpiresIn; secs > 0 {
		token.Expiry = time.Now().Add(time.Duration(secs) * time.Second)
	}
	if v := tokenRes.IDToken; v != "" {
		claimSet, err := jws.Decode(v)
		if err != nil {
			return "", err
		}
		token.Expiry = time.Unix(claimSet.Exp, 0)
	}
	tokenRes.RefreshToken = refreshToken
	f, err := os.OpenFile(*flCredentialFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	json.NewEncoder(f).Encode(tokenRes)

	return tokenRes.IDToken, nil

}

func main() {

	flag.Parse()
	if *flClientSecretsFile == "" {
		log.Fatalf("specify either --client_secrets_file must be set")
	}

	if *flAudience == "" {
		log.Fatalf("--audience must be set")
	}

	b, err := ioutil.ReadFile(*flClientSecretsFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	conf, err := google.ConfigFromJSON(b, userInfoEmailScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	var refreshToken string
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
		tok := &tokenResponse{}
		err = json.NewDecoder(f).Decode(tok)
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
	r, err := GetIdToken(*flAudience, conf.ClientID, conf.ClientSecret, refreshToken)
	if err != nil {
		log.Fatalf("Could not parse saved id_tokne File %v", err)
	}
	fmt.Printf("%s\n", r)
}
