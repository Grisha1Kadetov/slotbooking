package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/config"
	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/db"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Grisha1Kadetov/slotbooking/internal/handler/common"
	roomHandler "github.com/Grisha1Kadetov/slotbooking/internal/handler/room"
	roomRepo "github.com/Grisha1Kadetov/slotbooking/internal/repository/room"
	roomService "github.com/Grisha1Kadetov/slotbooking/internal/service/room"

	authHandler "github.com/Grisha1Kadetov/slotbooking/internal/handler/auth"
	authMiddle "github.com/Grisha1Kadetov/slotbooking/internal/middleware/auth"
	userRepo "github.com/Grisha1Kadetov/slotbooking/internal/repository/user"
	authService "github.com/Grisha1Kadetov/slotbooking/internal/service/auth"

	shedHandler "github.com/Grisha1Kadetov/slotbooking/internal/handler/schedule"
	shedRepo "github.com/Grisha1Kadetov/slotbooking/internal/repository/schedule"
	shedService "github.com/Grisha1Kadetov/slotbooking/internal/service/schedule"

	slotHandler "github.com/Grisha1Kadetov/slotbooking/internal/handler/slot"
	slotRepo "github.com/Grisha1Kadetov/slotbooking/internal/repository/slot"
	slotService "github.com/Grisha1Kadetov/slotbooking/internal/service/slot"

	bookingHandler "github.com/Grisha1Kadetov/slotbooking/internal/handler/booking"
	bookingRepo "github.com/Grisha1Kadetov/slotbooking/internal/repository/booking"
	bookingService "github.com/Grisha1Kadetov/slotbooking/internal/service/booking"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger := log.NewZapLogger()
	defer logger.Close()

	conf := config.GetConfig(logger)

	pool, err := db.NewPool(ctx, conf.PostgresDB.GetDSN())
	if err != nil {
		logger.Panic("failed to connect to database", log.Err(err))
	}
	defer pool.Close()

	r := NewRouter(ctx, pool, conf, logger)

	server := http.Server{
		Addr:    ":" + conf.Port,
		Handler: r,
	}
	go func() {
		logger.Info("starting server", log.Pair("port", conf.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			stop()
			logger.Panic("failed to start server", log.Err(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Panic("failed to shutdown server", log.Err(err))
	}
}

func NewRouter(ctx context.Context, pool *pgxpool.Pool, conf config.Config, logger log.Logger) http.Handler {
	roomRepo := roomRepo.New(pool)
	roomService := roomService.New(roomRepo)
	roomHandler := roomHandler.New(roomService, logger)

	userRepo := userRepo.New(pool)
	authService := authService.New([]byte(conf.JWTSecret), userRepo)
	authUserOnly := authMiddle.New(authService, user.RoleUser)
	authAdminOnly := authMiddle.New(authService, user.RoleAdmin)
	authAny := authMiddle.New(authService, user.RoleUser, user.RoleAdmin)
	authHandler := authHandler.New(authService, logger)

	slotRepo := slotRepo.New(pool)
	shedRepo := shedRepo.New(pool)
	slotServ := slotService.New(slotRepo, shedRepo, roomRepo, conf.SlotDur, conf.PregenDay)
	slotHandler := slotHandler.New(slotServ, logger)
	pregenerator := slotService.NewPregenerator(roomRepo, slotServ, time.Now)
	go func() {
		err := pregenerator.Run(ctx)
		if err != nil {
			logger.Error("error while running pregenerator", log.Err(err))
		}
	}()

	shedService := shedService.New(shedRepo, slotServ)
	shedHandler := shedHandler.New(shedService, logger)

	bookingRepo := bookingRepo.New(pool)
	bookingService := bookingService.New(bookingRepo, slotServ)
	bookingHandler := bookingHandler.New(bookingService, logger)

	r := chi.NewRouter()

	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)
	r.Post("/dummyLogin", authHandler.DummyLogin)

	r.Get("/_info", common.StatusOk)

	r.Group(func(rr chi.Router) {
		rr.Use(authAdminOnly.Middleware())
		rr.Post("/rooms/create", roomHandler.CreateRoom)
		rr.Post("/rooms/{roomId}/schedule/create", shedHandler.Create)
		rr.Get("/bookings/list", bookingHandler.GetListAll)
	})

	r.Group(func(rr chi.Router) {
		rr.Use(authUserOnly.Middleware())
		rr.Post("/bookings/create", bookingHandler.Create)
		rr.Post("/bookings/{bookingId}/cancel", bookingHandler.Cancel)
		rr.Get("/bookings/my", bookingHandler.GetMy)
	})

	r.Group(func(rr chi.Router) {
		rr.Use(authAny.Middleware())
		rr.Get("/rooms/list", roomHandler.GetAllRooms)
		rr.Get("/rooms/{roomId}/slots/list", slotHandler.GetAvailableList)
	})

	return r
}
