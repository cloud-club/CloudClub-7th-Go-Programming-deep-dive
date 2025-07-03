package handler

import (
	"feather/service"
	"feather/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(us service.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var req types.CreateUserReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response(ctx, http.StatusUnprocessableEntity, err.Error())
	} else if err := h.userService.CreateUser(req.Email, req.Password, req.Nickname); err != nil {
		response(ctx, http.StatusInternalServerError, err.Error())
	} else {
		response(ctx, http.StatusOK, "Success")
	}
}

func (h *UserHandler) User(ctx *gin.Context) {
	userId := ctx.Param("id")
	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		response(ctx, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if res, err := h.userService.User(id); err != nil {
		response(ctx, http.StatusInternalServerError, err.Error())
	} else {
		response(ctx, http.StatusOK, res)
	}
}
