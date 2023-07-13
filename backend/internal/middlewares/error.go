package middlewares

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type Request struct {
	URL         string            `json:"url,omitempty"`
	Method      string            `json:"method,omitempty"`
	Data        string            `json:"data,omitempty"`
	QueryString string            `json:"query_string,omitempty"`
	Cookies     string            `json:"cookies,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}

func extractRequestCookies(request *http.Request) string {
	cookies := request.Cookies()
	var cookieStrings []string
	for _, cookie := range cookies {
		cookieStrings = append(cookieStrings, cookie.String())
	}
	return strings.Join(cookieStrings, "; ")
}

func extractRequestHeaders(request *http.Request) map[string]string {
	headers := make(map[string]string)
	for key, values := range request.Header {
		headers[key] = values[0] // Use the first value if there are multiple values
	}
	return headers
}

func extractRequestEnvironment(request *http.Request) map[string]string {
	environment := make(map[string]string)
	for key, value := range request.Header {
		environment[key] = value[0]
	}
	return environment
}

func captureErrorToSentry(c *gin.Context, message string) {
	// Create a new Sentry event
	event := sentry.NewEvent()

	// Set the event's error details
	event.Message = message
	event.Level = sentry.LevelError

	// Capture and set the request details
	request := c.Request

	// Capture and set the request body payload
	fmt.Println(request.Body)

	event.Request = (*sentry.Request)(&Request{
		URL:    request.URL.String(),
		Method: request.Method,
		// Data:        payload,
		QueryString: request.URL.RawQuery,
		Cookies:     extractRequestCookies(request),
		Headers:     extractRequestHeaders(request),
		Env:         extractRequestEnvironment(request),
	})

	// Capture the event and send it to Sentry
	sentry.CaptureEvent(event)
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, err := range c.Errors {
			captureErrorToSentry(c, err.Error())
			c.JSON(-1, err)
		}
	}
}
