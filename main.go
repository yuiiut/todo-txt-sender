/**
this package send mail automatically.
	1. get file name
	2. search file & read file
		if there is't this file, return error
	3. make mail
	4. send mail
*/
package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"

	"golang.org/x/oauth2/google"

	"golang.org/x/oauth2"
)

const (
	reportedFile = "reported/"
)

const (
	clientID     = "ClientID"
	clientSecret = "ClientSecret"
	accessToken  = "AccessToken"
	refreshToken = "RefreshToken"
	fromAddress  = "FromAddress"
	toAddress    = "ToAddress"
)

var (
	configs configsStruct
	envs    = map[string]string{
		clientID:     "",
		clientSecret: "",
		accessToken:  "",
		refreshToken: "",
		fromAddress:  "",
		toAddress:    "",
	}
)

type configsStruct struct {
	oauthConf   oauth2.Config
	oauthToken  oauth2.Token
	fromAddress string
	toAddress   string
}

func init() {
	var env string
	configs = configsStruct{}

	for key := range envs {
		if env = os.Getenv(key); env == "" {
			log.Fatalf("no %s env", key)
		}
		envs[key] = env
	}

	configs.oauthConf = oauth2.Config{
		ClientID:     envs[clientID],
		ClientSecret: envs[clientSecret],
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{"https://mail.google.com/"},
	}

	expire, _ := time.Parse("2006-01-02", "2017-07-11")
	configs.oauthToken = oauth2.Token{
		AccessToken:  envs[accessToken],
		TokenType:    "Bearer",
		RefreshToken: envs[refreshToken],
		Expiry:       expire,
	}
}

func main() {
	var ctx = context.Background()

	// 1. get fileName
	todoFile := new(file)
	if err := todoFile.getParam(); err != nil {
		log.Fatal(err)
	}

	// 2. search file & read file
	if err := todoFile.readFile(); err != nil {
		log.Fatal(err)
	}

	// check "reported directory" & if there isn't directory, create directory.
	if err := checkDirectory(); err != nil {
		log.Fatal(err)
	}

	if err := todoFile.mailSender(ctx); err != nil {
		log.Fatal(err)
	}

	if err := todoFile.moveFile(); err != nil {
		log.Fatal(err)
	}

}

type file struct {
	fileName    string
	fileStrings string
	file        *os.File
}

func (f *file) getParam() error {
	flag.Parse()
	f.fileName = flag.Arg(0)
	if len(f.fileName) < 1 {
		return errors.New("no file name")
	}

	return nil
}

func (f *file) readFile() error {
	content, err := ioutil.ReadFile(f.fileName)
	if err != nil {
		return err
	}

	f.fileStrings = string(content)
	return nil
}

func checkDirectory() error {
	cmd := exec.Command("ls", "-F")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	files := strings.Split(string(output), "\n")
	var flag = false
	for _, f := range files {
		if f == reportedFile {
			flag = true
			break
		}
	}

	if !flag {
		newFile := fmt.Sprintf("./%s", reportedFile)
		cmd := exec.Command("mkdir", newFile)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *file) mailSender(ctx context.Context) error {
	mailer, err := gmail.New(configs.oauthConf.Client(ctx, &configs.oauthToken))
	if err != nil {
		return err
	}

	message := fmt.Sprintf(
		"From: %s\r\nTo:%s\r￿\nSubject:this is test mail\r\n\r\n%s",
		envs[fromAddress],
		envs[toAddress],
		f.fileStrings,
	)

	var msg gmail.Message
	msg.Raw = base64.StdEncoding.EncodeToString([]byte(message))

	if _, err := mailer.Users.Messages.Send("me", &msg).Do(); err != nil {
		return err
	}

	return nil
}

func (f *file) moveFile() error {
	cmd := exec.Command("mv", f.fileName, reportedFile)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
