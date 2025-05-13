package cmd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use: "serve",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		cfg := &struct {
			Server struct {
				Port string `env:"SERVER_PORT" env-required:"true"`
				Tls  struct {
					Cert string `env:"SERVER_TLS_CERT"`
					Key  string `env:"SERVER_TLS_KEY"`
				}

				Response struct {
					StatusCode int    `env:"SERVER_RESPONSE_STATUS_CODE" env-required:"true"`
					JSONBody   string `env:"SERVER_RESPONSE_JSON_BODY"`
					RawBody    string `env:"SERVER_RESPONSE_RAW_BODY"`
				}
			}
		}{}

		configFile := ".env"
		if len(args) > 0 {
			configFile = args[0]
		}

		if err := cleanenv.ReadConfig(configFile, cfg); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("cleanenv.ReadConfig: %w", err)
			}

			if err := cleanenv.ReadEnv(cfg); err != nil {
				return fmt.Errorf("cleanenv.ReadEnv: %w", err)
			}
		}

		var contentType = "text/plain"
		var body = cfg.Server.Response.RawBody
		if cfg.Server.Response.JSONBody != "" {
			contentType = "json"
			body = cfg.Server.Response.JSONBody
		}

		h := buildResponseHandler(cfg.Server.Response.StatusCode, contentType, body)

		srv := http.Server{
			Addr:    ":" + cfg.Server.Port,
			Handler: h,
		}

		needTLS := cfg.Server.Tls.Cert != "" && cfg.Server.Tls.Key != ""

		if needTLS {
			certPem := []byte(cfg.Server.Tls.Cert)
			keyPem := []byte(cfg.Server.Tls.Key)

			cer, err := tls.X509KeyPair(certPem, keyPem)
			if err != nil {
				return fmt.Errorf("load tls key pair error: %w", err)
			}

			srv.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cer},
			}
		}

		errCh := make(chan error)

		go func() {
			if needTLS {
				errCh <- srv.ListenAndServeTLS("", "")

			} else {
				errCh <- srv.ListenAndServe()
			}
		}()

		select {
		case err := <-errCh:
			return fmt.Errorf("error while starting server: %w", err)

		case <-ctx.Done():
			cmd.Printf("start gracefull shutdown\n")
		}

		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}

		cmd.Printf("server stopped\n")

		return nil
	},
}

func buildResponseHandler(status int, contentType, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}
