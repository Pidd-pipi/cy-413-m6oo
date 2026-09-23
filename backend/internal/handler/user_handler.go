package handler

import (
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/middleware"
	"github.com/blueship581/mindgarden/backend/internal/service"
	"github.com/blueship581/mindgarden/backend/internal/util"
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

type UserHandler struct {
	users          *service.UserService
	assessments    *service.AssessmentService
	recaps         *service.RecapService
	logger         *slog.Logger
	secret, issuer string
}

func NewUserHandler(u *service.UserService, a *service.AssessmentService, r *service.RecapService, l *slog.Logger, secret, issuer string) *UserHandler {
	return &UserHandler{u, a, r, l, secret, issuer}
}
func (h *UserHandler) Register(c *gin.Context) {
	var r dto.RegisterRequest
	if !bind(c, &r) {
		return
	}
	u, e := h.users.Register(r)
	if e != nil {
		c.Error(e)
		return
	}
	t, e := util.CreateToken(u.ID, u.Role, h.secret, h.issuer)
	if e != nil {
		c.Error(e)
		return
	}
	h.logger.Info(constants.LogAuthRegister, "user_id", u.ID)
	created(c, gin.H{"token": t, "user": u})
}
func (h *UserHandler) Login(c *gin.Context) {
	var r dto.LoginRequest
	if !bind(c, &r) {
		return
	}
	u, e := h.users.Login(r)
	if e != nil {
		c.Error(e)
		return
	}
	t, e := util.CreateToken(u.ID, u.Role, h.secret, h.issuer)
	if e != nil {
		c.Error(e)
		return
	}
	h.logger.Info(constants.LogAuthLogin, "user_id", u.ID)
	ok(c, gin.H{"token": t, "user": u})
}
func (h *UserHandler) Me(c *gin.Context) {
	u, e := h.users.Get(middleware.UserID(c))
	if e != nil {
		c.Error(e)
		return
	}
	h.logger.Info(constants.LogUserProfileRead, "user_id", u.ID)
	ok(c, u)
}
func (h *UserHandler) Update(c *gin.Context) {
	var r dto.UpdateProfileRequest
	if !bind(c, &r) {
		return
	}
	u, e := h.users.Update(middleware.UserID(c), r)
	if e != nil {
		c.Error(e)
		return
	}
	ok(c, u)
}
func (h *UserHandler) Report(c *gin.Context) {
	v, e := h.assessments.Report(middleware.UserID(c))
	if e != nil {
		c.Error(e)
		return
	}
	ok(c, v)
}
func (h *UserHandler) WeeklyRecap(c *gin.Context) {
	v, e := h.recaps.WeeklyRecap(middleware.UserID(c), time.Now())
	if e != nil {
		c.Error(e)
		return
	}
	ok(c, v)
}
