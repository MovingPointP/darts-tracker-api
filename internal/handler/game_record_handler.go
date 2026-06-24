package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/MovingPointP/darts-tracker-api/internal/domain/entity"
	"github.com/MovingPointP/darts-tracker-api/internal/domain/repository"
	"github.com/MovingPointP/darts-tracker-api/internal/usecase"
	"github.com/gin-gonic/gin"
)

const defaultPageSize = 20

type GameRecordHandler struct {
	gameRecordUsecase usecase.GameRecordUsecase
}

func NewGameRecordHandler(gameRecordUsecase usecase.GameRecordUsecase) *GameRecordHandler {
	return &GameRecordHandler{gameRecordUsecase: gameRecordUsecase}
}

type CreateGameRecordRequest struct {
	GameType string    `json:"game_type" binding:"required,oneof=01game cricket countup"`
	Value    float64   `json:"value" binding:"gte=0"`
	PlayedAt time.Time `json:"played_at" binding:"required"`
}

type UpdateGameRecordRequest struct {
	Value    float64   `json:"value" binding:"gte=0"`
	PlayedAt time.Time `json:"played_at" binding:"required"`
}

func getUserID(ctx *gin.Context) string {
	return ctx.MustGet("UserID").(string)
}

// @Summary     記録作成
// @Description 新しいゲーム記録を作成する。01game/cricketはレーティングを自動算出する
// @Tags        records
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body CreateGameRecordRequest true "記録情報"
// @Success     201 {object} entity.GameRecord
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /records [post]
func (h *GameRecordHandler) CreateGameRecord(ctx *gin.Context) {
	var req CreateGameRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.gameRecordUsecase.Create(getUserID(ctx), entity.GameType(req.GameType), req.Value, req.PlayedAt)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidGameType) || errors.Is(err, entity.ErrValueOutOfRange) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create game record"})
		return
	}

	ctx.JSON(http.StatusCreated, record)
}

// @Summary     記録一覧取得
// @Description ログインユーザーの記録をフィルタ・ページネーション付きで取得する
// @Tags        records
// @Security    BearerAuth
// @Produce     json
// @Param       game_type query string false "種目フィルタ(01game/cricket/countup)"
// @Param       from      query string false "開始日(YYYY-MM-DD)"
// @Param       to        query string false "終了日(YYYY-MM-DD)"
// @Param       limit     query int    false "1ページあたりの件数(デフォルト20)"
// @Param       offset    query int    false "スキップする件数(デフォルト0)"
// @Success     200 {object} repository.PagedRecords
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /records [get]
func (h *GameRecordHandler) GetGameRecords(ctx *gin.Context) {
	filter := repository.RecordsFilter{
		Limit: defaultPageSize,
	}

	if q := ctx.Query("game_type"); q != "" {
		gt := entity.GameType(q)
		if !gt.Valid() {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": entity.ErrInvalidGameType.Error()})
			return
		}
		filter.GameType = &gt
	}

	if q := ctx.Query("from"); q != "" {
		t, err := time.Parse("2006-01-02", q)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date, expected YYYY-MM-DD"})
			return
		}
		filter.From = &t
	}

	if q := ctx.Query("to"); q != "" {
		t, err := time.Parse("2006-01-02", q)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date, expected YYYY-MM-DD"})
			return
		}
		filter.To = &t
	}

	if q := ctx.Query("limit"); q != "" {
		v, err := strconv.Atoi(q)
		if err != nil || v <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		}
		filter.Limit = v
	}

	if q := ctx.Query("offset"); q != "" {
		v, err := strconv.Atoi(q)
		if err != nil || v < 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
		filter.Offset = v
	}

	result, err := h.gameRecordUsecase.GetWithFilter(getUserID(ctx), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get game records"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// @Summary     日別平均レーティング取得
// @Description 指定種目の日別平均レーティングを時系列順で返す(グラフ表示用)
// @Tags        stats
// @Security    BearerAuth
// @Produce     json
// @Param       game_type query string true "種目(01game/cricket)"
// @Success     200 {array} repository.DailyRating
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /stats/ratings [get]
func (h *GameRecordHandler) GetDailyRatings(ctx *gin.Context) {
	q := ctx.Query("game_type")
	if q == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "game_type is required"})
		return
	}
	gameType := entity.GameType(q)

	ratings, err := h.gameRecordUsecase.GetDailyRatings(getUserID(ctx), gameType)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidGameType) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get daily ratings"})
		return
	}

	ctx.JSON(http.StatusOK, ratings)
}

// @Summary     記録更新
// @Description 指定IDの記録を更新する(値変更時はレーティングを再計算)
// @Tags        records
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path int                     true "記録ID"
// @Param       body body UpdateGameRecordRequest true "更新情報"
// @Success     200 {object} entity.GameRecord
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /records/{id} [put]
func (h *GameRecordHandler) UpdateGameRecord(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid record id"})
		return
	}

	var req UpdateGameRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.gameRecordUsecase.Update(uint(id), getUserID(ctx), req.Value, req.PlayedAt)
	if err != nil {
		if errors.Is(err, entity.ErrGameRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, entity.ErrValueOutOfRange) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update game record"})
		return
	}

	ctx.JSON(http.StatusOK, record)
}

// @Summary     記録削除
// @Description 指定IDの記録を削除する
// @Tags        records
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "記録ID"
// @Success     204
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /records/{id} [delete]
func (h *GameRecordHandler) DeleteGameRecord(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid record id"})
		return
	}

	if err := h.gameRecordUsecase.Delete(uint(id), getUserID(ctx)); err != nil {
		if errors.Is(err, entity.ErrGameRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete game record"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
