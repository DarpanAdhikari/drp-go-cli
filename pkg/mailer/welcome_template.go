package mailer

import (
	"bytes"
	"text/template"
)

var welcomeTextTmpl = template.Must(template.New("welcome").Parse(`Hi {{.Name}},

Welcome! Your account has been created successfully.

If you didn't create this account, you can safely ignore this email.

Thanks,
The Team
`))

var welcomeHTMLTmpl = template.Must(template.New("welcome_html").Parse(`<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:20px">
  <h2>Welcome, {{.Name}}!</h2>
  <p>Your account has been created successfully.</p>
  <p>If you didn't create this account, you can safely ignore this email.</p>
  <p>Thanks,<br>The Team</p>
</body>
</html>
`))

type welcomeData struct {
	Name string
}

// WelcomeText renders the plain-text welcome email body for the given name.
func WelcomeText(name string) (string, error) {
	var buf bytes.Buffer
	if err := welcomeTextTmpl.Execute(&buf, welcomeData{Name: name}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// WelcomeHTML renders the HTML welcome email body for the given name.
func WelcomeHTML(name string) (string, error) {
	var buf bytes.Buffer
	if err := welcomeHTMLTmpl.Execute(&buf, welcomeData{Name: name}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SendWelcome is a convenience method that sends the welcome email to a user.
func (m *Mailer) SendWelcome(to, name string) error {
	body, err := WelcomeText(name)
	if err != nil {
		return err
	}
	return m.Send(to, "Welcome!", body)
}
