package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	MailTrapURL   = os.Getenv("MAILTRAP_URL")
	MailTrapToken = os.Getenv("MAILTRAP_TOKEN")
	SenderEmail   = os.Getenv("SENDER_EMAIL")
)

type MailTrapSuccessResponse struct {
	Success    bool     `json:"success"`
	MessageIDs []string `json:"message_ids"`
}

type MailTrapErrorResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors"`
}

type EmailPayload struct {
	From struct {
		Email string `json:"email"`
	} `json:"from"`
	To []struct {
		Email string `json:"email"`
	} `json:"to"`
	Subject string `json:"subject"`
	Text    string `json:"text,omitempty"`
	HTML    string `json:"html,omitempty"`
}

type EmailService struct {
	smtpURL     string
	accessToken string
	senderEmail string
	httpClient  *http.Client
}

func NewEmailService() *EmailService {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	emailService := &EmailService{
		MailTrapURL,
		MailTrapToken,
		SenderEmail,
		httpClient,
	}

	return emailService
}

func (es *EmailService) SendEmail(to []string, subject, bodyText, bodyHTML string) error {
	var payload EmailPayload

	payload.From.Email = es.senderEmail

	for _, addr := range to {
		payload.To = append(payload.To, struct {
			Email string `json:"email"`
		}{Email: addr})
	}

	payload.Subject = subject
	payload.Text = bodyText
	payload.HTML = bodyHTML

	message, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, es.smtpURL, bytes.NewBuffer(message))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+es.accessToken)

	res, err := es.httpClient.Do(request)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer (func() {
		_ = res.Body.Close()
	})()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	switch res.StatusCode {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError:
		errorResponse := MailTrapErrorResponse{}
		err = json.Unmarshal(body, &errorResponse)
		if nil != err {
			return err
		}

		return fmt.Errorf("email sending failed: %+v", errorResponse.Errors)
	}

	return nil
}
