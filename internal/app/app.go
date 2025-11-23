package app

import (
	"github.com/111zxc/pr-review-service/internal/config"
	"github.com/111zxc/pr-review-service/internal/handler"
	"github.com/111zxc/pr-review-service/internal/logger"
	"github.com/111zxc/pr-review-service/internal/repository/postgres"
	"github.com/111zxc/pr-review-service/internal/service"
)

func Run() {
	cfg := config.Load()

	if err := logger.Init(logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	}); err != nil {
		panic(err)
	}

	db, err := postgres.NewDB(cfg)
	if err != nil {
		logger.Error("failed to connect to DB", logger.WithError(err))
	}

	tx := postgres.NewTxManager(db)

	userRepo := postgres.NewUserRepository(db)
	teamRepo := postgres.NewTeamRepository(db, tx)
	prRepo := postgres.NewPullRequestRepository(db, tx)
	prStatusRepo := postgres.NewPRStatusRepository(db)
	statsRepo := postgres.NewStatsRepository(db)
	eventRepo := postgres.NewEventsRepository(db)

	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo)
	prService := service.NewPullRequestService(prRepo, userRepo, teamRepo, prStatusRepo, eventRepo)
	statsService := service.NewStatsService(statsRepo)

	h := handler.New(
		teamService,
		userService,
		prService,
		statsService,
	)

	router := NewRouter(h)
	srv := NewServer(cfg, router)

	srv.Start()
}
