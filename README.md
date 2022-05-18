# a Better IMDB is possible

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white) ![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white) ![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white) ![Redis](https://img.shields.io/badge/redis-%23DD0031.svg?style=for-the-badge&logo=redis&logoColor=white)
 
In this project, a functional application that manages back-end needs of a  movie/series database.

Features:
- You can make account management additionally with administration role
- You can manage contents with admin role by adding with IMDB-ID, raw-JSON and delete content by IMDB-ID
- Users can add/delete movie/series to their Favorites and search by genre and content name
- On public access, any guest can request movies list , a specific movie, series list, a specific series along with its seasons and episodes. Additionally, the guest can search by genre and content name.
- Dockerized application, PostgreSQL and Redis by docker-compose.

## Get Started

```
git clone https://github.com/fukaraca/worth2watch.git
```


- Insert API key to env file which's been provided by [TMDB](https://www.themoviedb.org).
- If you use provided docker-compose file(this is the default setting), after starting Docker daemon, run
 `docker-compose up -d` .

Now, PostgreSQL, Redis and application must be running. 

In case using local Psql and Redis, You need to create a database and name it in accordance with config.env>>DB_NAME value. 
Also you must uncomment the stated lines on config.env properly and delete the prioring keys.
This is a must because former values were for Dockerized option. Since you are using localhost, it needs to be changed.  
And now, you can simply 

` go run .`

On initial run, application will create required tables automatically, you only need to register, log-in, and add-content you wish to.

## Endpoints


```go
package api


var R *gin.Engine

func Endpoints() {
 //naming convention for URL paths are designated for readability

 //public
 R.GET("/movies/:id", GetThisMovie)
 R.GET("/movies/list", GetMoviesWithPage)
 R.GET("/searchContent", SearchContent)
 R.GET("/series/:seriesid", GetThisSeries)
 R.GET("/series/list", GetSeriesWithPage)
 R.GET("/series/:seriesid/:season", GetEpisodesForaSeason)
 R.GET("/getSimilarContent", GetSimilarContent)
 
 //user accessed
 R.POST("/addFavorites", Auth(AddToFavorites))
 R.GET("/getFavorites", Auth(GetFavorites))
 R.GET("/searchFavorites", Auth(SearchFavorites))
 
 //content management
 R.POST("/addContentByID", Auth(AddContentByID))
 R.POST("/addContentWithJSON", Auth(AddContentWithJSON))
 R.DELETE("/deleteMovieByID", Auth(DeleteContentByID))
 
 //account management
 R.GET("/user/:username", Auth(GetUserInfo))
 R.POST("/register", CheckRegistration)
 R.POST("/login", Login)
 R.PATCH("/updateUser", Auth(UpdateUser))
 R.POST("/logout", Auth(Logout))
}
```

