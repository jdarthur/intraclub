// Package mailer sends email directly to recipient MX servers, signed with DKIM.
//
// Intended for low-volume transactional mail (notifications, password resets)
// from a single domain. Call Send from inside the app; there is no SMTP server
// component. For high volume or strong deliverability guarantees, route through
// a transactional provider instead.
package mailer

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/emersion/go-msgauth/dkim"
)

type Config struct {
	FromDomain   string // e.g. "rcintra.club" — must match the From address domain
	Hostname     string // HELO/EHLO name, ideally matches rDNS of the sending IP
	DKIMSelector string // DNS selector, e.g. "default" for default._domainkey.rcintra.club
	DKIMKeyPath  string // PEM-encoded RSA private key (PKCS1 or PKCS8)
	DialTimeout  time.Duration
}

type Mailer struct {
	cfg    Config
	signer crypto.Signer
}

func New(cfg Config) (*Mailer, error) {
	if cfg.FromDomain == "" || cfg.Hostname == "" || cfg.DKIMSelector == "" {
		return nil, errors.New("mailer: FromDomain, Hostname, DKIMSelector are required")
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 30 * time.Second
	}
	signer, err := loadPrivateKey(cfg.DKIMKeyPath)
	if err != nil {
		return nil, fmt.Errorf("mailer: dkim key: %w", err)
	}
	return &Mailer{cfg: cfg, signer: signer}, nil
}

type Message struct {
	From    string // e.g. "noreply@rcintra.club"
	To      []string
	Subject string
	Text    string // plain UTF-8 body
}

// Send delivers msg to every recipient. Recipients are grouped by domain so
// each MX server sees one SMTP transaction. A failure for one domain does not
// prevent delivery to others; all errors are joined into the return value.
func (m *Mailer) Send(ctx context.Context, msg Message) error {
	if _, err := mail.ParseAddress(msg.From); err != nil {
		return fmt.Errorf("invalid From: %w", err)
	}
	if len(msg.To) == 0 {
		return errors.New("no recipients")
	}

	raw, err := m.buildAndSign(msg)
	if err != nil {
		return err
	}

	byDomain, err := groupByDomain(msg.To)
	if err != nil {
		return err
	}

	var errs []error
	for domain, rcpts := range byDomain {
		if err := m.deliver(ctx, domain, msg.From, rcpts, raw); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", domain, err))
		}
	}
	return errors.Join(errs...)
}

func (m *Mailer) buildAndSign(msg Message) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", msg.From)
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(msg.To, ", "))
	fmt.Fprintf(&buf, "Subject: %s\r\n", msg.Subject)
	fmt.Fprintf(&buf, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	fmt.Fprintf(&buf, "Message-ID: <%s@%s>\r\n", newMessageID(), m.cfg.FromDomain)
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(strings.ReplaceAll(strings.ReplaceAll(msg.Text, "\r\n", "\n"), "\n", "\r\n"))

	opts := &dkim.SignOptions{
		Domain:   m.cfg.FromDomain,
		Selector: m.cfg.DKIMSelector,
		Signer:   m.signer,
		Hash:     crypto.SHA256,
		HeaderKeys: []string{
			"From", "To", "Subject", "Date", "Message-ID",
			"MIME-Version", "Content-Type", "Content-Transfer-Encoding",
		},
	}
	var signed bytes.Buffer
	if err := dkim.Sign(&signed, &buf, opts); err != nil {
		return nil, fmt.Errorf("dkim sign: %w", err)
	}
	return signed.Bytes(), nil
}

func (m *Mailer) deliver(ctx context.Context, domain, from string, to []string, msg []byte) error {
	hosts, err := lookupMX(ctx, domain)
	if err != nil {
		return err
	}

	var lastErr error
	for _, host := range hosts {
		if err := m.deliverTo(ctx, host, from, to, msg); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return fmt.Errorf("all MX hosts failed: %w", lastErr)
}

func (m *Mailer) deliverTo(ctx context.Context, host, from string, to []string, msg []byte) error {
	d := net.Dialer{Timeout: m.cfg.DialTimeout}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, "25"))
	if err != nil {
		return err
	}

	deadline := time.Now().Add(2 * time.Minute)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	if err := c.Hello(m.cfg.Hostname); err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, r := range to {
		if err := c.Rcpt(r); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func lookupMX(ctx context.Context, domain string) ([]string, error) {
	var r net.Resolver
	mx, err := r.LookupMX(ctx, domain)
	if err != nil {
		// RFC 5321 §5.1: if no MX record, fall back to the A/AAAA of the domain itself.
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			return []string{domain}, nil
		}
		return nil, fmt.Errorf("mx lookup: %w", err)
	}
	if len(mx) == 0 {
		return []string{domain}, nil
	}
	sort.SliceStable(mx, func(i, j int) bool { return mx[i].Pref < mx[j].Pref })
	hosts := make([]string, len(mx))
	for i, rec := range mx {
		hosts[i] = strings.TrimSuffix(rec.Host, ".")
	}
	return hosts, nil
}

func groupByDomain(addrs []string) (map[string][]string, error) {
	out := map[string][]string{}
	for _, a := range addrs {
		parsed, err := mail.ParseAddress(a)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient %q: %w", a, err)
		}
		i := strings.LastIndex(parsed.Address, "@")
		if i < 0 {
			return nil, fmt.Errorf("recipient %q missing domain", a)
		}
		domain := strings.ToLower(parsed.Address[i+1:])
		out[domain] = append(out[domain], parsed.Address)
	}
	return out, nil
}

func loadPrivateKey(path string) (crypto.Signer, error) {
	if path == "" {
		return nil, errors.New("DKIMKeyPath is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("no PEM block in key file")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		signer, ok := k.(crypto.Signer)
		if !ok {
			return nil, errors.New("PKCS8 key is not a crypto.Signer")
		}
		return signer, nil
	}
	return nil, errors.New("unsupported key format (need PKCS1 or PKCS8 RSA/Ed25519)")
}

func newMessageID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
