package main

import (
	"errors"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/kataras/iris"
	"github.com/valyala/fasthttp"
)

type Oauth2KeyExists struct {
	*Middleware
}

//Important staff, iris middleware must implement the iris.Handler interface which is:
func (m Oauth2KeyExists) ProcessRequest(req fasthttp.Request, resp fasthttp.Response, c *iris.Context) (error, int) {
	fields := log.Fields{
		"path":   c.PathString(),
		"origin": c.RemoteAddr(),
	}

	// We're using OAuth, start checking for access keys
	authHeaderValue := string(req.Header.Peek("Authorization"))
	parts := strings.Split(authHeaderValue, " ")
	if len(parts) < 2 {
		log.WithFields(fields).Info("Attempted access with malformed header, no auth header found.")

		return errors.New("Authorization field missing"), fasthttp.StatusBadRequest
	}

	if strings.ToLower(parts[0]) != "bearer" {
		log.WithFields(fields).Info("Bearer token malformed")

		return errors.New("Bearer token malformed"), fasthttp.StatusBadRequest
	}

	accessToken := parts[1]
	thisSessionState, keyExists := m.CheckSessionAndIdentityForValidKey(accessToken)

	if !keyExists {
		log.WithFields(log.Fields{
			"path":   c.PathString(),
			"origin": c.RemoteAddr(),
			"key":    accessToken,
		}).Info("Attempted access with non-existent key.")

		return errors.New("Key not authorised"), fasthttp.StatusForbidden
	}

	c.Set(SessionData, thisSessionState)
	c.Set(AuthHeaderValue, accessToken)

	return nil, fasthttp.StatusOK
}