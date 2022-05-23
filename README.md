# a Better IMDB is possible


![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white) ![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white) ![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white) ![Redis](https://img.shields.io/badge/redis-%23DD0031.svg?style=for-the-badge&logo=redis&logoColor=white)



**Worth2watch2** is a better version of [worth2watch](https://github.com/fukaraca/worth2watch)
project which was hard by testing and maintaining manner.

In this project, it's aimed to build a RESTFUL, functional, testable, maintainable application that manages back-end needs of a movie/series database. 

Features:
- You can handle account management
- You can manage contents with admin role by adding with IMDB-ID, raw-JSON and delete content by IMDB-ID
- Users can add/delete movie/series to their Favorites and search by genre and content name
- By public access, any guest can request movies list , a specific movie, series list, a specific series along with its seasons and episodes. 
- Additionally, the guest can search by genre and content name.
- Dockerized application, PostgreSQL and Redis by docker-compose.
- Unit tested...
- Aimed to follow clean code and SOLID principles.

## Get Started

```
git clone https://github.com/fukaraca/worth2watch2.git
```


- Insert API key to env file which's been provided as free by [TMDB](https://www.themoviedb.org).
- If you'll use provided docker-compose file(this is the default setting), after starting Docker daemon, run
 `docker-compose up -d` .



Now, PostgreSQL, Redis and application must be running. 

In case of using local Psql and Redis, You may need to create a database and name it in accordance with config.env>>DB_NAME value. 
Also you must uncomment the stated lines on config.env properly and delete the prioring keys.
This is a must because former values were for Dockerized option. Since you are using localhost as DB_HOST, it needs to be changed.  
And now, you can simply 

` go run .`

On initial run, application will create required tables automatically, you only need to register, log-in, and add-content you wish to.

## Testing

You can simply call `make testall` for Unix based OS users. 

For Windows users who don't have GNU-make `go test -v ./..` command may be used,  *but ownership constraints of `./db/volume/` folder must be awared of . 



## Endpoints


```go
package api

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

```

