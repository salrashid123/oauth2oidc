package oauth2oidc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

const (
	UserInfoEmailScope = "https://www.googleapis.com/auth/userinfo.email"
	tokenURL           = "https://oauth2.googleapis.com/token"
)

func GetIdToken(audience, clientId, clientSecret, refreshToken string) (tresp *TokenResponse, err error) {
	data := url.Values{
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
		"audience":      {audience},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return &TokenResponse{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return &TokenResponse{}, errors.New(fmt.Sprintf("Error exchange token: %s", b))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return &TokenResponse{}, err
	}

	tokenRes := &TokenResponse{}

	if err := json.Unmarshal(body, tokenRes); err != nil {
		return &TokenResponse{}, err
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
			return &TokenResponse{}, err
		}
		token.Expiry = time.Unix(claimSet.Exp, 0)
	}
	tokenRes.RefreshToken = refreshToken

	return tokenRes, nil
}
