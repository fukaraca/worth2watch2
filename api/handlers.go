package api

import (
	"encoding/json"
	"fmt"
	"github.com/fukaraca/worth2watch/model"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"log"
	"net/http"
	"strconv"
	"time"
)

//Auth is the authentication middleware
func (s *service) auth(fn gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Req ID:", requestid.Get(c))
		if !s.CheckSession(c) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"notification": "you must login",
			})
			c.Abort()
			return
		}

		fn(c)
	}
}

//CheckRegistration is a func for registering a new user or admin.
//Data must be POSTed as "form data".
//Leading or trailing whitespaces will be handled by frontend
func (s *service) checkRegistration(c *gin.Context) {
	if s.CheckSession(c) {

		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "already logged in",
		})
		return
	}

	newUser := new(model.User)
	err := c.BindJSON(newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": err.Error(),
		})
		log.Println("new user info as JSON couldn't be binded:", err)
		return
	}
	if newUser.Username == "" || *newUser.Email == "" || newUser.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "username or email or password cannot be empty",
		})
		return
	}

	newUser.Username = *s.Striper(newUser.Username)
	newUser.Email = s.Striper(*newUser.Email)
	pass, err := s.HashPassword(*s.Striper(newUser.Password))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": err.Error(),
		})
		log.Println("new user's password couldn't be hashed:", err)
		return
	}
	newUser.Password = pass
	err = newUser.CreatedOn.Set(time.Now())
	newUser.LastLogin.Set(time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": err.Error(),
		})
		log.Println("new user info creation time couldn't be assigned:", err)
		return
	}

	err = s.CreateNewUser(c, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"notification": "account created successfully",
	})
}

//Login is handler function for login process.
//It requires form-data for 'logUsername' and 'logPassword' keys.
func (s *service) login(c *gin.Context) {
	if s.CheckSession(c) {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "already logged in",
		})
		return
	}
	logUsername := c.PostForm("logUsername")
	logPassword := c.PostForm("logPassword")

	hashedPass, err := s.QueryLogin(c, logUsername)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": err.Error(),
		})
		return
	}
	if !s.CheckPasswordHash(logPassword, hashedPass) {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "password or username is incorrect"})
		return
	}
	lastLoginTime := time.Now()

	err = s.UpdateLastLogin(c, lastLoginTime, logUsername)
	if err != nil {
		log.Println("update login time failed:", err)
	}
	s.CreateSession(logUsername, c)
	log.Printf("%s has logged in:\n", logUsername)
	c.JSON(http.StatusOK, gin.H{
		"notification": "user " + logUsername + " successfully logged in",
	})
}

//Logout handler
func (s *service) logout(c *gin.Context) {
	_, err := s.DeleteSession(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"notification": "logged out successfully",
	})
}

//UpdateUser is handler for updating user/admins informations. Changing admin/user role was not implemented.
func (s *service) updateUser(c *gin.Context) {
	firstname := *s.Striper(c.PostForm("firstname"))
	lastname := *s.Striper(c.PostForm("lastname"))

	username, _ := c.Cookie("uid")

	err := s.UpdateUserInfo(c, firstname, lastname, username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "user information update failed",
		})
		log.Println("user information update has failed:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"notification": "user informations updated succesfully",
	})
}

//GetUserInfo handles GET request for user infos. Only admin and the user can get the info
func (s *service) getUserInfo(c *gin.Context) {
	username, _ := c.Cookie("uid")
	usernameP := c.Param("username")
	//only user and admin may peek user info
	if usernameP != username || !s.CheckAdminForLoggedIn(c, username) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"notification": "you are not allowed to see another users info",
		})
		return
	}
	user, err := s.QueryUserInfo(c, usernameP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "user information get failed",
		})
		log.Println("user information get has failed:", err)
		return
	}
	c.JSON(http.StatusOK, user)
}

//AddContentByID is handler func for content adding by IMDB ID and contents type.
//It requires "movie" or "series" for content-type key.
//Insertion is being maintained asynchronously due to expensive amount of time that was required to be got all data related to series.
func (s *service) addContentByID(c *gin.Context) {
	username, _ := c.Cookie("uid")
	if !s.CheckAdminForLoggedIn(c, username) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"notification": "unauthorized attempt",
		})
		log.Println("unauthorized attempt by user:", username)
		return
	}
	IMDBID := c.PostForm("imdb-id")
	contentType := c.PostForm("content-type")
	switch contentType {
	case "movie":
		go s.AddMovieContentWithID(IMDBID)

	case "series":
		go s.AddSeriesContentWithID(IMDBID)

	default:
		log.Println("invalid content-type:", contentType)
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "invalid content-type: " + contentType,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"notification": "content will be inserted shortly after",
	})
}

//AddContentWithJSON handles adding new content with JSON format.
//Content-type "movie" or "series" must be provided.
//This is not practical at all.
//However, for an internal movie database that includes content that is not contained in IMDB, it can be useful with an additional struct field exposes e.g. BetterIMDB_ID..
func (s *service) addContentWithJSON(c *gin.Context) {
	username, _ := c.Cookie("uid")
	if !s.CheckAdminForLoggedIn(c, username) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"notification": "unauthorized attempt",
		})
		log.Println("log: unauthorized attempt by user:", username)
		return
	}

	contentType := c.PostForm("content-type")
	inputJSON := c.PostForm("content-raw-data")

	switch contentType {
	case "movie":
		movie := new(model.Movie)
		err := json.Unmarshal([]byte(inputJSON), movie)
		if err != nil {
			log.Println("movie content couldn't be added: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "movie content couldn't be added. error: " + err.Error(),
			})
			return
		}
		err = s.AddMovieContentWithStruct(c, movie)
		if err != nil {
			log.Println("movie content couldn't be added: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"notification": "movie content couldn't be added. error: " + err.Error(),
			})
			return
		}
	case "series":
		series := new(model.Series)
		err := json.Unmarshal([]byte(inputJSON), series)
		if err != nil {
			log.Println("series content couldn't be added: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "series content couldn't be added. error: " + err.Error(),
			})
			return
		}
		err = s.AddSeriesContentWithStruct(c, series)
		if err != nil {
			log.Println("series content couldn't be added: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"notification": "series content couldn't be added. error: " + err.Error(),
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "invalid content-type: " + contentType,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"notification": "content has been created successfully",
	})
}

//DeleteContentByID deletes content for given IMDB id. Content-type must be provided
func (s *service) deleteContentByID(c *gin.Context) {
	username, _ := c.Cookie("uid")
	if !s.CheckAdminForLoggedIn(c, username) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"notification": "unauthorized attempt",
		})
		log.Println("unauthorized attempt by user:", username)
		return
	}
	IMDBID := c.PostForm("imdb-id")
	contentType := c.PostForm("content-type")

	if !(contentType == "movie" || contentType == "series") {
		log.Println("invalid content-type:", contentType)
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "invalid content-type: " + contentType,
		})
		return
	}

	err := s.DeleteContent(c, IMDBID, contentType)
	if err != nil {
		log.Println("content ", IMDBID, " couldn't be deleted: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": "content " + IMDBID + " couldn't be deleted. error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notification": "content has been deleted successfully",
	})
}

//GetThisMovie is a handler function for responsing a specific movie details
func (s *service) getThisMovie(c *gin.Context) {
	id := c.Param("id")
	movie, err := s.GetThisMovieFromDB(c, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"notification": "no such movie",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"notification": err,
			})
			return
		}
	}
	fmt.Println()
	c.JSON(http.StatusOK, movie)
}

//GetThisSeries is a handler function for responsing a specific serie with its seasons
func (s *service) getThisSeries(c *gin.Context) {
	id := c.Param("seriesid")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "missing serie id",
		})
		return
	}
	series, seasons, err := s.GetThisSeriesFromDB(c, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"notification": "no such serie",
			})
			return
		} else if err.Error() == "there is no season for given series" {
			c.JSON(http.StatusOK, gin.H{
				"series":       series,
				"notification": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"notification": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"series":  series,
		"seasons": seasons,
	})
}

//GetEpisodesForaSeason is a handle function for responsing episodes for a certain season of a series
func (s *service) getEpisodesForaSeason(c *gin.Context) {
	seriesid := c.Param("seriesid")
	seasonNumber := c.Param("season")

	season, err := s.GetEpisodesForaSeasonFromDB(c, seriesid, seasonNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"notification": "no such season",
			})
			return
		} else {
			log.Println("get this season failed for serie: ", seriesid, " season:", seasonNumber, " err: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"notification": err,
			})
			return
		}
	}

	c.JSON(http.StatusOK, season)
}

//GetMoviesWithPage handles request for given amount of item and page
func (s *service) getMoviesWithPage(c *gin.Context) {
	var err error
	q := c.Request.URL.Query()
	page := 1
	items := 10

	if q.Has("page") {
		page, err = strconv.Atoi(q.Get("page"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid page format",
			})
			return
		}
	}
	if q.Has("items") {
		items, err = strconv.Atoi(q.Get("items"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid items format",
			})
			return
		}
	}
	movies, err := s.GetMoviesListWithPage(c, page, items)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{
				"notification": "end of the list",
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"notification": err,
			})
			return
		}
	}
	c.JSON(http.StatusOK, movies)

}

//GetSeriesWithPage handles request for given amount of item and page
func (s *service) getSeriesWithPage(c *gin.Context) {
	var err error
	q := c.Request.URL.Query()
	page := 1
	items := 10

	if q.Has("page") {
		page, err = strconv.Atoi(q.Get("page"))
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid page format",
			})
			return
		}
	}
	if q.Has("items") {
		items, err = strconv.Atoi(q.Get("items"))
		if err != nil || items < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid items format",
			})
			return
		}
	}

	series, err := s.GetSeriesListWithPage(c, page, items)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{
				"notification": "end of the list",
			})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"notification": err,
			})
			return
		}
	}
	c.JSON(http.StatusOK, series)

}

//SearchContent is handler function for searching movies/series by name and genres.
//Page and item amount for movie or series on a page must be provided
func (s *service) searchContent(c *gin.Context) {
	var err error
	q := c.Request.URL.Query()
	name := ""
	if q.Has("name") {
		if len(q["name"]) > 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "only one name can be accepted",
			})
			return
		}
		name = q.Get("name")
	}
	genres := []string{}
	if q.Has("genre") {

		genres = q["genre"]
	}
	page := 1
	items := 10

	if q.Has("page") {
		page, err = strconv.Atoi(q.Get("page"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid page format",
			})
			return
		}
	}
	if q.Has("items") {
		items, err = strconv.Atoi(q.Get("items"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid items format",
			})
			return
		}
	}
	//db query not a handler
	movies, series, err := s.SearchContent(c, name, genres, page, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": "en error occurred",
		})
		return
	}
	if movies == nil && series == nil {
		c.JSON(http.StatusOK, gin.H{
			"notification": "no content found for given filter",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"movies": movies,
		"series": series,
	})
}

//GetSimilarContent handles for similar content request. Similarity evaluation is made by genre tags and sorted descending of rating
func (s *service) getSimilarContent(c *gin.Context) {
	IMDBID := ""
	contentType := ""
	var err error
	q := c.Request.URL.Query()

	if q.Has("imdb-id") && q.Has("content-type") {
		IMDBID = q.Get("imdb-id")
		contentType = q.Get("content-type")

	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "invalid imdb-id and content-type format",
		})
		return
	}

	if !(contentType == "movie" || contentType == "series") {
		log.Println("invalid content-type:", contentType)
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "invalid content-type: " + contentType,
		})
		return
	}
	movies, series, err := s.FindSimilarContent(c, IMDBID, contentType)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{
				"notification": "no match of content",
			})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"notification": err,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"movies": movies,
		"series": series,
	})
}

//AddToFavorites adds content to users favorites.
func (s *service) addToFavorites(c *gin.Context) {
	id := c.PostForm("imdb-id")
	contentType := c.PostForm("content-type")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "missing imdb-id",
		})
		return
	}
	if !(contentType == "movie" || contentType == "series") {
		c.JSON(http.StatusBadRequest, gin.H{
			"notification": "invalid content-type: " + contentType,
		})
		return
	}
	err := s.AddContentToFavorites(c, id, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"notification": "content has been added successfully",
	})
}

//GetFavorites ...
func (s *service) getFavorites(c *gin.Context) {
	var err error
	q := c.Request.URL.Query()
	page := 1
	items := 10

	if q.Has("page") {
		page, err = strconv.Atoi(q.Get("page"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid page format",
			})
			return
		}
	}
	if q.Has("items") {
		items, err = strconv.Atoi(q.Get("items"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid items format",
			})
			return
		}
	}

	movies, series, err := s.GetFavoriteContents(c, page, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": "en error occurred",
		})
		return
	}
	if movies == nil && series == nil {
		c.JSON(http.StatusOK, gin.H{
			"notification": "there is no favorite item",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"movies": movies,
		"series": series,
	})
}

//SearchFavorites is handler function for searching on favorites
func (s *service) searchFavorites(c *gin.Context) {
	var err error
	q := c.Request.URL.Query()
	name := ""
	if q.Has("name") {
		if len(q["name"]) > 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "only one name can be accepted",
			})
			return
		}
		name = q.Get("name")
	}
	genres := []string{}
	if q.Has("genre") {

		genres = q["genre"]
	}
	page := 1
	items := 10

	if q.Has("page") {
		page, err = strconv.Atoi(q.Get("page"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid page format",
			})
			return
		}
	}
	if q.Has("items") {
		items, err = strconv.Atoi(q.Get("items"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"notification": "invalid items format",
			})
			return
		}
	}
	//db query not a handler
	movies, series, err := s.SearchFavorites(c, name, genres, page, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"notification": "en error occurred",
		})
		return
	}
	if movies == nil && series == nil {
		c.JSON(http.StatusOK, gin.H{
			"notification": "no content found for given filter",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"movies": movies,
		"series": series,
	})
}
