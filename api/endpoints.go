package api

import "github.com/gin-gonic/gin"

func endpoints(r *gin.Engine, h *service) *gin.Engine {
	//naming convention for URL paths are designated for readability. So, it's not bug, it's a feature...

	//public
	r.GET("/movies/:id", h.getThisMovie)
	r.GET("/movies/list", h.getMoviesWithPage)
	r.GET("/search", h.searchContent)
	r.GET("/series/:seriesid", h.getThisSeries)
	r.GET("/series/list", h.getSeriesWithPage)
	r.GET("/series/:seriesid/:season", h.getEpisodesForaSeason)
	r.GET("/similars", h.getSimilarContent)
	//user accessed
	r.POST("/favorites", h.auth(h.addToFavorites))
	r.GET("/favorites", h.auth(h.getFavorites))
	r.GET("/searchFavorites", h.auth(h.searchFavorites))
	//content management
	r.POST("/contentByID", h.auth(h.addContentByID))
	r.POST("/contentWithJSON", h.auth(h.addContentWithJSON))
	r.DELETE("/contentByID", h.auth(h.deleteContentByID))
	//account management
	r.GET("/user/:username", h.auth(h.getUserInfo))
	r.POST("/register", h.checkRegistration)
	r.POST("/login", h.login)
	r.PATCH("/user", h.auth(h.updateUser))
	r.POST("/logout", h.auth(h.logout))
	return r
}
