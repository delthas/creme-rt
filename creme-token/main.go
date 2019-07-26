package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/dghubble/oauth1/twitter"

	"github.com/dghubble/oauth1"
)

func main() {
	consumerKey := flag.String("key", "", "oauth consumer API key")
	consumerSecret := flag.String("secret", "", "oauth consumer API secret")
	callbackUrl := flag.String("callback", "", "oauth callback URL")
	flag.Parse()

	oauthConfig := oauth1.NewConfig(*consumerKey, *consumerSecret)
	oauthConfig.CallbackURL = *callbackUrl
	oauthConfig.Endpoint = twitter.AuthenticateEndpoint
	token, secret, err := oauthConfig.RequestToken()
	if err != nil {
		panic(err)
	}
	u, err := oauthConfig.AuthorizationURL(token)
	if err != nil {
		panic(err)
	}
	fmt.Println("Open this URL in your browser, accept the dialog and write the redirect URL back in the program:")
	fmt.Println(u)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	u, err = url.Parse(strings.TrimSpace(text))
	if err != nil {
		panic(err)
	}
	t, s, err := oauthConfig.AccessToken(u.Query().Get("oauth_token"), secret, u.Query().Get("oauth_verifier"))
	if err != nil {
		panic(err)
	}
	fmt.Println("access token:")
	fmt.Println(t)
	fmt.Println("access token secret:")
	fmt.Println(s)
}
