package actions

import (
	"html"
	"os"
	"regexp"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/gobuffalo/suite/v4"
)

var csrfFieldPattern = regexp.MustCompile(`name="authenticity_token" value="([^"]+)"`)

type ActionSuite struct {
	*suite.Action
}

// authed prepara una petición que cambia estado (POST, PUT, DELETE) con el token CSRF
// de la sesión de prueba, igual que lo envía un formulario de la aplicación.
func (as *ActionSuite) authed(u string, args ...interface{}) *httptest.Request {
	req := as.HTML(u, args...)
	req.Headers["X-CSRF-Token"] = as.csrfToken()
	return req
}

// csrfToken lee el token CSRF enmascarado de un formulario renderizado por la aplicación.
func (as *ActionSuite) csrfToken() string {
	res := as.HTML("/").Get()
	match := csrfFieldPattern.FindStringSubmatch(res.Body.String())
	as.Len(match, 2)
	return html.UnescapeString(match[1])
}

func Test_ActionSuite(t *testing.T) {
	action, err := suite.NewActionWithFixtures(App(), os.DirFS("../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}
