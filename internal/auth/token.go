// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	auth "github.com/kakaoenterprise/kc-sdk-go/services/iam"
)

type TokenManager struct {
	mutex            sync.Mutex
	credentialID     string
	credentialSecret string
	currentToken     string
	expiresAt        time.Time
	identityAPI      auth.IdentityAPI
}

func NewTokenManager(identityAPI auth.IdentityAPI, credentialID, credentialSecret string) *TokenManager {
	return &TokenManager{
		credentialID:     credentialID,
		credentialSecret: credentialSecret,
		identityAPI:      identityAPI,
	}
}

func (tm *TokenManager) GetValidToken(ctx context.Context) (string, error) {
	tm.mutex.Lock()

	if tm.currentToken != "" && time.Now().Add(5*time.Minute).Before(tm.expiresAt) {
		tm.mutex.Unlock()
		return tm.currentToken, nil
	}

	if tm.currentToken != "" {
		if isValid, err := tm.validateToken(ctx, tm.currentToken); err == nil && isValid {
			tm.mutex.Unlock()
			return tm.currentToken, nil
		}
	}

	tm.mutex.Unlock()

	return tm.IssueNewToken(ctx)
}

func (tm *TokenManager) IssueNewToken(ctx context.Context) (string, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	authReq := auth.SwaggerIdPwdRequest{
		Auth: &auth.SwaggerAuth{
			Identity: &auth.SwaggerIdPwd{
				Methods: []string{"application_credential"},
				ApplicationCredential: &auth.SwaggerApplicationCredential{
					Id:     &tm.credentialID,
					Secret: &tm.credentialSecret,
				},
			},
		},
	}

	resp, httpResp, err := tm.identityAPI.IssueToken(ctx).SwaggerIdPwdRequest(authReq).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to issue token: %w", err)
	}

	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			err := httpResp.Body.Close()
			if err != nil {
				return
			}
		}
	}()

	token := httpResp.Header.Get("X-Subject-Token")
	if token == "" {
		return "", fmt.Errorf("no token found in response headers")
	}

	tm.currentToken = token
	if resp != nil && resp.Token != nil && resp.Token.ExpiresAt != nil {
		if expiresAt, err := time.Parse(time.RFC3339, *resp.Token.ExpiresAt); err == nil {
			tm.expiresAt = expiresAt
		} else {
			return "", fmt.Errorf("failed to parse token expiration time: %w", err)
		}
	} else {
		return "", fmt.Errorf("no expiration time found in token response")
	}

	return token, nil
}

func (tm *TokenManager) validateToken(ctx context.Context, token string) (bool, error) {
	_, httpResp, err := tm.identityAPI.ValidateToken(ctx).
		XAuthToken(token).
		XSubjectToken(token).
		Execute()

	if err != nil {
		return false, fmt.Errorf("token validation request failed: %w", err)
	}

	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			err := httpResp.Body.Close()
			if err != nil {
				return
			}
		}
	}()

	if httpResp.StatusCode != 200 {
		return false, fmt.Errorf("token validation failed: HTTP %d", httpResp.StatusCode)
	}

	return true, nil
}

func (tm *TokenManager) InvalidateToken() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.currentToken = ""
}

func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	authErrorPatterns := []string{
		"401",
		"unauthorized",
		"authentication",
		"invalid token",
		"token expired",
		"access denied",
		"forbidden",
	}

	for _, pattern := range authErrorPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
