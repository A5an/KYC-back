package notifier

import (
	"fmt"
	"strings"

	"github.com/resendlabs/resend-go"

	"github.com/Sinbad-HQ/kyc/config"
)

func SendEmailNotification(recipients []string, subject string, body string, isPlainText bool) error {
	resendConfig := config.GetResendConfig()
	params := &resend.SendEmailRequest{
		From:    resendConfig.EmailFrom,
		To:      recipients,
		Subject: subject,
	}

	if isPlainText {
		params.Text = body
	} else {
		params.Html = body
	}

	resp, err := resend.NewClient(resendConfig.ApiKey).Emails.Send(params)
	if err != nil {
		return err
	}

	if resp.Id == "" {
		return fmt.Errorf("failed to send email to %s", strings.Join(recipients, ""))
	}

	return nil
}
