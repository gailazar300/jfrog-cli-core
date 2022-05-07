package envsetup

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/access/services"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"

	"github.com/google/uuid"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	ioUtils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	myJfrogEndPoint      = "https://myjfrog-api.jfrog.com/api/v1/activation/cloud/cli/getStatus/"
	syncSleepInterval    = 5 * time.Second  // 5 seconds
	maxWaitMinutes       = 30 * time.Minute // 30 minutes
	nonExpiredTokenValue = 0                // Access Tokens with 0 expiration value are actually generated by Access with 1 year expiration.
)

var trueValue = true

type EnvSetupCommand struct {
	registrationURL string
	// In case encodedConnectionDetails were provided - we have a registered user that was invited to the platform.
	encodedConnectionDetails string
	id                       uuid.UUID
	serverDetails            *config.ServerDetails
	progress                 ioUtils.ProgressMgr
}

func (ftc *EnvSetupCommand) SetRegistrationURL(registrationURL string) *EnvSetupCommand {
	ftc.registrationURL = registrationURL
	return ftc
}

func (ftc *EnvSetupCommand) SetEncodedConnectionDetails(encodedConnectionDetails string) *EnvSetupCommand {
	ftc.encodedConnectionDetails = encodedConnectionDetails
	return ftc
}

func (ftc *EnvSetupCommand) ServerDetails() (*config.ServerDetails, error) {
	return nil, nil
}

func (ftc *EnvSetupCommand) SetProgress(progress ioUtils.ProgressMgr) {
	ftc.progress = progress
}

func NewEnvSetupCommand() *EnvSetupCommand {
	return &EnvSetupCommand{
		id: uuid.New(),
	}
}

// This function is a wrapper around the 'ftc.progress.SetHeadlineMsg(msg)' API,
// to make sure that ftc.progress isn't nil. It can be nil in case the CI environment variable is set.
// In case ftc.progress is nil, the message sent will be prompted to the screen
// without the progress indication.
func (ftc *EnvSetupCommand) setHeadlineMsg(msg string) {
	if ftc.progress != nil {
		ftc.progress.SetHeadlineMsg(msg)
	} else {
		log.Output(msg + "...")
	}
}

// This function is a wrapper around the 'ftc.progress.clearHeadlineMsg()' API,
// to make sure that ftc.progress isn't nil before clearing it.
// It can be nil in case the CI environment variable is set.
func (ftc *EnvSetupCommand) clearHeadlineMsg() {
	if ftc.progress != nil {
		ftc.progress.ClearHeadlineMsg()
	}
}

// This function is a wrapper around the 'ftc.progress.Quit()' API,
// to make sure that ftc.progress isn't nil before clearing it.
// It can be nil in case the CI environment variable is set.
func (ftc *EnvSetupCommand) quitProgress() error {
	if ftc.progress != nil {
		return ftc.progress.Quit()
	}
	return nil
}

func (ftc *EnvSetupCommand) Run() (err error) {
	err = ftc.SetupAndConfigServer()
	if err != nil {
		return err
	}
	// Closes the progress manger and reset the log prints.
	err = ftc.quitProgress()
	if err != nil {
		return err
	}
	log.Output()
	log.Output(coreutils.PrintBold("Congrats! You're all set"))
	message :=
		coreutils.PrintTitle("So what's next?") + "\n" +
			"1. 'cd' into your code project directory\n" +
			"2. Run \"jf project init\"\n" +
			"3. Read more about how to get started at -\n" +
			coreutils.PrintLink(coreutils.GettingStartedGuideUrl) + "\n" +
			"4. We've just sent you an email message. Please use it to verify your email address"

	err = coreutils.PrintTable("", "", message, false)
	return
}

func (ftc *EnvSetupCommand) SetupAndConfigServer() (err error) {
	var server *config.ServerDetails
	// If credentials were provided, this means that the user was invited to join an existing JFrog environment.
	// Otherwise, this is a brand-new user, that needs to register and setup a new JFrog environment.
	if ftc.encodedConnectionDetails == "" {
		server, err = ftc.setupNewUser()
	} else {
		server, err = ftc.setupInvitedUser()
	}
	if err != nil {
		return
	}
	err = configServer(server)
	return
}

func (ftc *EnvSetupCommand) setupNewUser() (*config.ServerDetails, error) {
	ftc.setHeadlineMsg("Just fill out its details in your browser 📝")
	time.Sleep(8 * time.Second)
	err := browser.OpenURL(ftc.registrationURL + "?id=" + ftc.id.String())
	if err != nil {
		return nil, err
	}
	return ftc.getNewServerDetails()
}

func (ftc *EnvSetupCommand) setupInvitedUser() (server *config.ServerDetails, err error) {
	server, err = ftc.decodeConnectionDetails()
	if err != nil {
		return
	}
	err = GenerateNewLongTermRefreshableAccessToken(server)
	return
}

// Take the short-lived token and generate a long term (1 year expiry) refreshable accessToken.
func GenerateNewLongTermRefreshableAccessToken(server *config.ServerDetails) (err error) {
	accessManager, err := utils.CreateAccessServiceManager(server, false)
	if err != nil {
		return
	}
	// Create refreshable accessToken with 1 year expiry from the given short expiry token.
	params := createLongExpirationRefreshableTokenParams()
	token, err := accessManager.CreateAccessToken(*params)
	if err != nil {
		return
	}
	server.AccessToken = token.AccessToken
	server.RefreshToken = token.RefreshToken
	return
}

func createLongExpirationRefreshableTokenParams() *services.CreateTokenParams {
	params := services.CreateTokenParams{}
	params.ExpiresIn = nonExpiredTokenValue
	params.Refreshable = &trueValue
	params.Audience = "*@*"
	return &params
}

func (ftc *EnvSetupCommand) decodeConnectionDetails() (server *config.ServerDetails, err error) {
	rawDecodedText, err := base64.StdEncoding.DecodeString(ftc.encodedConnectionDetails)
	if errorutils.CheckError(err) != nil {
		return
	}
	err = json.Unmarshal(rawDecodedText, &server)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	return
}

func (ftc *EnvSetupCommand) CommandName() string {
	return "setup"
}

// Returns the new server deatailes from My-JFrog
func (ftc *EnvSetupCommand) getNewServerDetails() (serverDetails *config.ServerDetails, err error) {
	requestBody := &myJfrogGetStatusRequest{CliRegistrationId: ftc.id.String()}
	requestContent, err := json.Marshal(requestBody)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	httpClientDetails := httputils.HttpClientDetails{
		Headers: map[string]string{"Content-Type": "application/json"},
	}
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return nil, err
	}

	// Define the MyJFrog polling logic.
	pollingMessage := fmt.Sprintf("Sync: Get MyJFrog status report. Request ID:%s...", ftc.id)
	pollingErrorMessage := "Sync: Get MyJFrog status request failed. Attempt: %d. Error: %s"
	// The max consecutive polling errors allowed, before completely failing the setup action.
	const maxConsecutiveErrors = 6
	errorsCount := 0
	readyMessageDisplayed := false
	pollingAction := func() (shouldStop bool, responseBody []byte, err error) {
		log.Debug(pollingMessage)
		// Send request to MyJFrog.
		resp, body, err := client.SendPost(myJfrogEndPoint, requestContent, httpClientDetails, "")
		// If an HTTP error occured.
		if err != nil {
			errorsCount++
			log.Debug(fmt.Sprintf(pollingErrorMessage, errorsCount, err.Error()))
			if errorsCount == maxConsecutiveErrors {
				return true, nil, err
			}
			return false, nil, nil
		}
		// If the response is not the expected 200 or 404.
		if err = errorutils.CheckResponseStatus(resp, http.StatusOK, http.StatusNotFound); err != nil {
			err = errorutils.CheckError(errorutils.GenerateResponseError(resp.Status, clientutils.IndentJson(body)))
			errorsCount++
			log.Debug(fmt.Sprintf(pollingErrorMessage, errorsCount, err.Error()))
			if errorsCount == maxConsecutiveErrors {
				return true, nil, err
			}
			return false, nil, nil
		}
		errorsCount = 0

		// Wait for 'ready=true' response from MyJFrog
		if resp.StatusCode == http.StatusOK {
			if !readyMessageDisplayed {
				ftc.clearHeadlineMsg()
				ftc.setHeadlineMsg("Almost done! Please hang on while JFrog CLI completes the setup 🛠")
				readyMessageDisplayed = true
			}
			statusResponse := myJfrogGetStatusResponse{}
			if err = json.Unmarshal(body, &statusResponse); err != nil {
				return true, nil, err
			}
			// Got the new server details
			if statusResponse.Ready {
				return true, body, nil
			}
		}
		// The expected 404 response or 200 response without 'Ready'
		return false, nil, nil
	}

	pollingExecutor := &httputils.PollingExecutor{
		Timeout:         maxWaitMinutes,
		PollingInterval: syncSleepInterval,
		PollingAction:   pollingAction,
	}

	body, err := pollingExecutor.Execute()
	if err != nil {
		return nil, err
	}
	statusResponse := myJfrogGetStatusResponse{}
	if err = json.Unmarshal(body, &statusResponse); err != nil {
		return nil, errorutils.CheckError(err)
	}
	ftc.clearHeadlineMsg()
	serverDetails = &config.ServerDetails{
		Url:         statusResponse.PlatformUrl,
		AccessToken: statusResponse.AccessToken,
	}
	ftc.serverDetails = serverDetails
	return serverDetails, nil
}

// Add the given server details to the cli's config by running a 'jf config' command
func configServer(server *config.ServerDetails) error {
	u, err := url.Parse(server.Url)
	if errorutils.CheckError(err) != nil {
		return err
	}
	// Take the server name from host name: https://myjfrog.jfrog.com/ -> myjfrog
	serverId := strings.Split(u.Host, ".")[0]
	configCmd := commands.NewConfigCommand().SetInteractive(false).SetServerId(serverId).SetDetails(server)
	if err = configCmd.Config(); err != nil {
		return err
	}
	return commands.Use(serverId)
}

type myJfrogGetStatusRequest struct {
	CliRegistrationId string `json:"cliRegistrationId,omitempty"`
}

type myJfrogGetStatusResponse struct {
	CliRegistrationId string `json:"cliRegistrationId,omitempty"`
	Ready             bool   `json:"ready,omitempty"`
	AccessToken       string `json:"accessToken,omitempty"`
	PlatformUrl       string `json:"platformUrl,omitempty"`
}
