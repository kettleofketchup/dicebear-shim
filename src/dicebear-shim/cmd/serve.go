package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kettleofketchup/dicebear-shim/src/dicebear-shim/internal/shim"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Gravatar-to-DICEbear proxy server",
	Long: `Start an HTTP server that translates Gravatar avatar requests
into DICEbear API calls. Incoming requests to /avatar/<hash> are
proxied to a DICEbear backend with the appropriate style and size.`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().String("listen", ":3001", "listen address")
	serveCmd.Flags().String("dicebear-url", "http://dicebear:3000", "DICEbear backend URL")
	serveCmd.Flags().String("default-style", "identicon", "default DICEbear style")
	serveCmd.Flags().String("default-size", "80", "default avatar size in pixels")
	serveCmd.Flags().Int("cache-max-age", 86400, "Cache-Control max-age in seconds")

	viper.BindPFlag("listen", serveCmd.Flags().Lookup("listen"))
	viper.BindPFlag("dicebear.url", serveCmd.Flags().Lookup("dicebear-url"))
	viper.BindPFlag("default.style", serveCmd.Flags().Lookup("default-style"))
	viper.BindPFlag("default.size", serveCmd.Flags().Lookup("default-size"))
	viper.BindPFlag("cache.maxAge", serveCmd.Flags().Lookup("cache-max-age"))
}

func runServe(cmd *cobra.Command, args []string) error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := shim.Config{
		DiceBearURL:  viper.GetString("dicebear.url"),
		DefaultStyle: viper.GetString("default.style"),
		DefaultSize:  viper.GetString("default.size"),
		CacheMaxAge:  viper.GetInt("cache.maxAge"),
	}

	handler := shim.NewHandler(cfg, log)
	mux := http.NewServeMux()
	mux.Handle("/avatar/", handler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	addr := viper.GetString("listen")
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		log.Info("starting dicebear-shim", "addr", addr, "backend", cfg.DiceBearURL, "style", cfg.DefaultStyle)
		errCh <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("shutting down", "signal", sig.String())
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
