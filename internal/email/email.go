package email

import (
    "crypto/tls"
    "fmt"
    "net/smtp"
)

// Send sends an email using given SMTP credentials.
func Send(host string, port int, username, password, to, subject, body string) error {
    addr := fmt.Sprintf("%s:%d", host, port)
    auth := smtp.PlainAuth("", username, password, host)

    msg := []byte("To: " + to + "\r\n" +
        "Subject: " + subject + "\r\n" +
        "MIME-Version: 1.0\r\n" +
        "Content-Type: text/plain; charset=\"utf-8\"\r\n" +
        "\r\n" + body + "\r\n")

    // TLS configuration (skip verify for simplicity; user should trust)
    tlsconfig := &tls.Config{InsecureSkipVerify: true, ServerName: host}
    conn, err := tls.Dial("tcp", addr, tlsconfig)
    if err != nil {
        return err
    }
    c, err := smtp.NewClient(conn, host)
    if err != nil {
        return err
    }
    if err = c.Auth(auth); err != nil {
        return err
    }
    if err = c.Mail(username); err != nil {
        return err
    }
    if err = c.Rcpt(to); err != nil {
        return err
    }
    w, err := c.Data()
    if err != nil {
        return err
    }
    _, err = w.Write(msg)
    if err != nil {
        return err
    }
    err = w.Close()
    if err != nil {
        return err
    }
    return c.Quit()
}
