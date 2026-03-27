package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/taverns-red/tavern-url/internal/model"
)

// DomainService handles custom domain operations.
type DomainService struct {
	// In production, this would use a domain repository.
}

// NewDomainService creates a new DomainService.
func NewDomainService() *DomainService {
	return &DomainService{}
}

// GenerateDNSToken creates a random verification token.
func (s *DomainService) GenerateDNSToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	h := sha256.Sum256(b)
	return "tvn-verify-" + hex.EncodeToString(h[:16]), nil
}

// VerifyDNS checks for a TXT record matching the expected token.
func (s *DomainService) VerifyDNS(ctx context.Context, domain string, expectedToken string) (bool, error) {
	records, err := net.DefaultResolver.LookupTXT(ctx, domain)
	if err != nil {
		return false, fmt.Errorf("DNS lookup failed: %w", err)
	}

	for _, r := range records {
		if strings.TrimSpace(r) == expectedToken {
			return true, nil
		}
	}
	return false, nil
}

// MatchHostToOrg resolves a custom domain to its org (placeholder).
// In production, this queries the custom_domains table.
func (s *DomainService) MatchHostToOrg(ctx context.Context, host string) (*model.CustomDomain, error) {
	// This would query the database in production.
	return nil, fmt.Errorf("custom domain %q not found", host)
}
