package main

import (
	"context"
	"fmt"
	"os"

	"github.com/diuk/raiseup/config"
	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/auth"
	"github.com/diuk/raiseup/pkg/database"
	jwtutil "github.com/diuk/raiseup/pkg/jwt"
	"github.com/diuk/raiseup/pkg/logger"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		log.Error("ADMIN_EMAIL and ADMIN_PASSWORD are required to seed an administrator")
		os.Exit(1)
	}

	role := db.UserRole(cfg.AdminRole)
	if role != db.UserRoleSUPERADMIN && role != db.UserRoleADMINRW {
		log.Error("ADMIN_ROLE must be SUPER_ADMIN or ADMIN_RW", "role", cfg.AdminRole)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := auth.NewRepository(pool)
	tokens := jwtutil.NewManager(cfg.JWTSecret, cfg.JWTExpiresIn)
	service := auth.NewService(repo, tokens)

	user, err := service.SeedAdmin(ctx, cfg.AdminEmail, cfg.AdminName, cfg.AdminPassword, role)
	if err != nil {
		log.Error("failed to seed admin", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Seeded admin user\n")
	fmt.Printf("  id:    %s\n", user.ID)
	fmt.Printf("  email: %s\n", user.Email)
	fmt.Printf("  name:  %s\n", user.Name)
	fmt.Printf("  role:  %s\n", user.Role)
}
