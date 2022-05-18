package db

import (
	"github.com/fukaraca/worth2watch/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"log"
)

//AddContentToFavorites adds related content to favorites for specific user
func (dbi *dbImp) AddContentToFavorites(c *gin.Context, IMDB, cType string) error {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	username, _ := c.Cookie("uid")

	if cType == "movie" {
		_, err := conn.Exec(ctx, "INSERT INTO favorite_movies (favorite_id, user_id, movie_id) VALUES (nextval('favorite_movies_favorite_id_seq'),(SELECT user_id FROM users WHERE username=$1),(SELECT movie_id FROM movies WHERE imdb_id=$2));", username, IMDB)
		if err != nil {
			return err
		}
		return nil
	}
	_, err := conn.Exec(ctx, "INSERT INTO favorite_series (favorite_id, user_id, serie_id) VALUES (nextval('favorite_series_favorite_id_seq'),(SELECT user_id FROM users WHERE username=$1),(SELECT serie_id FROM series WHERE imdb_id=$2));", username, IMDB)
	if err != nil {
		return err
	}
	return nil
}

//GetFavoriteContents queires for favorites for certain user as per pagination code
func (dbi *dbImp) GetFavoriteContents(c *gin.Context, page, items int) (*[]model.Movie, *[]model.Series, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	offset := (page - 1) * items
	username, _ := c.Cookie("uid")

	rows, err := conn.Query(ctx, "SELECT title,description,rating,release_date,imdb_id,genres FROM movies LEFT JOIN favorite_movies ON movies.movie_id = favorite_movies.movie_id WHERE user_id=(SELECT user_id FROM users WHERE username=$1) ORDER BY movies.movie_id DESC OFFSET $2 LIMIT $3;", username, offset, items)
	defer rows.Close()
	tempMovies := []model.Movie{}
	if err != nil {
		tempMovies = nil
		goto serieLabel3
	}

	for rows.Next() {
		temp := model.Movie{}
		err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
		if err != nil {
			log.Println("a movie couldn't be scanned")
			continue
		}
		tempMovies = append(tempMovies, temp)
	}
serieLabel3:
	//then search for series
	rows, err = conn.Query(ctx, "SELECT title,description,rating,release_date,imdb_id,genres FROM series LEFT JOIN favorite_series ON series.serie_id = favorite_series.serie_id WHERE user_id=(SELECT user_id FROM users WHERE username=$1) ORDER BY series.serie_id DESC OFFSET $2 LIMIT $3;", username, offset, items)
	defer rows.Close()
	tempSeries := []model.Series{}
	if err != nil {
		tempSeries = nil
		goto returnLabel3
	}

	for rows.Next() {
		temp := model.Series{}
		err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
		if err != nil {
			log.Println("a series couldn't be scanned")
			continue
		}
		tempSeries = append(tempSeries, temp)
	}
returnLabel3:
	return &tempMovies, &tempSeries, nil
}

//SearchFavorites search favorites for both movie and series with case insensitive regexp and full match with genres
func (dbi *dbImp) SearchFavorites(c *gin.Context, name string, genres []string, page, items int) (*[]model.Movie, *[]model.Series, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	offset := (page - 1) * items
	username, _ := c.Cookie("uid")

	if name != "" && len(genres) > 0 {
		//both name and genres are requested (OR Conditional)
		//first search for movie
		rows, err := conn.Query(ctx, `
SELECT * FROM (SELECT title,description,rating,release_date,imdb_id,genres FROM movies
    LEFT JOIN favorite_movies ON movies.movie_id = favorite_movies.movie_id
WHERE user_id=(SELECT user_id FROM users WHERE username=$1)) AS SQ
WHERE title ~* $2 OR SQ.genres && $3 ORDER BY rating DESC OFFSET $4 LIMIT $5;
`, username, name, genres, offset, items)
		defer rows.Close()
		tempMovies := []model.Movie{}
		if err != nil {
			tempMovies = nil
			goto seriesLabel1
		}

		for rows.Next() {
			temp := model.Movie{}
			err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
			if err != nil {
				log.Println("a movie couldn't be scanned")
				continue
			}
			tempMovies = append(tempMovies, temp)
		}
	seriesLabel1:
		//then search for series
		rows, err = conn.Query(ctx, `
SELECT * FROM (SELECT title,description,rating,release_date,imdb_id,genres FROM series
    LEFT JOIN favorite_series ON series.serie_id = favorite_series.serie_id
WHERE user_id=(SELECT user_id FROM users WHERE username=$1)) AS SQ
WHERE title ~* $2 OR SQ.genres && $3 ORDER BY rating DESC OFFSET $4 LIMIT $5;
`, username, name, genres, offset, items)
		defer rows.Close()
		tempSeries := []model.Series{}
		if err != nil {
			tempSeries = nil
			goto returnLabel1
		}

		for rows.Next() {
			temp := model.Series{}
			err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
			if err != nil {
				log.Println("a series couldn't be scanned")
				continue
			}
			tempSeries = append(tempSeries, temp)
		}
	returnLabel1:
		return &tempMovies, &tempSeries, nil
	} else if name == "" && len(genres) > 0 {
		//search only for genres
		//first search for movie
		rows, err := conn.Query(ctx, `
SELECT * FROM (SELECT title,description,rating,release_date,imdb_id,genres FROM movies
    LEFT JOIN favorite_movies ON movies.movie_id = favorite_movies.movie_id
WHERE user_id=(SELECT user_id FROM users WHERE username=$1)) AS SQ
WHERE SQ.genres && $2 ORDER BY rating DESC OFFSET $3 LIMIT $4;
`, username, genres, offset, items)
		defer rows.Close()
		tempMovies := []model.Movie{}
		if err != nil {
			tempMovies = nil
			goto seriesLabel2
		}

		for rows.Next() {
			temp := model.Movie{}

			err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
			if err != nil {
				log.Println("a movie couldn't be scanned")
				continue
			}
			tempMovies = append(tempMovies, temp)
		}
	seriesLabel2:
		//then search for series
		rows, err = conn.Query(ctx, `
SELECT * FROM (SELECT title,description,rating,release_date,imdb_id,genres FROM series
    LEFT JOIN favorite_series ON series.serie_id = favorite_series.serie_id
WHERE user_id=(SELECT user_id FROM users WHERE username=$1)) AS SQ
WHERE SQ.genres && $2 ORDER BY rating DESC OFFSET $3 LIMIT $4;
`, username, genres, offset, items)
		defer rows.Close()
		tempSeries := []model.Series{}
		if err != nil {
			tempSeries = nil
			goto returnLabel2
		}

		for rows.Next() {
			temp := model.Series{}
			err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
			if err != nil {
				log.Println("a series couldn't be scanned")
				continue
			}

			tempSeries = append(tempSeries, temp)
		}
	returnLabel2:
		return &tempMovies, &tempSeries, nil
	} else if name != "" && len(genres) == 0 {
		//search only with name
		//first search for movie
		rows, err := conn.Query(ctx, `
SELECT * FROM (SELECT title,description,rating,release_date,imdb_id,genres FROM movies
    LEFT JOIN favorite_movies ON movies.movie_id = favorite_movies.movie_id
WHERE user_id=(SELECT user_id FROM users WHERE username=$1)) AS SQ
WHERE title ~* $2 ORDER BY rating DESC OFFSET $3 LIMIT $4;
`, username, name, offset, items)
		defer rows.Close()
		tempMovies := []model.Movie{}
		if err != nil {
			tempMovies = nil
			goto serieLabel3
		}

		for rows.Next() {
			temp := model.Movie{}
			err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
			if err != nil {
				log.Println("a movie couldn't be scanned")
				continue
			}
			tempMovies = append(tempMovies, temp)
		}
	serieLabel3:
		//then search for series
		rows, err = conn.Query(ctx, `
SELECT * FROM (SELECT title,description,rating,release_date,imdb_id,genres FROM series
    LEFT JOIN favorite_series ON series.serie_id = favorite_series.serie_id
WHERE user_id=(SELECT user_id FROM users WHERE username=$1)) AS SQ
WHERE title ~* $2 ORDER BY rating DESC OFFSET $3 LIMIT $4;
`, username, name, offset, items)
		defer rows.Close()
		tempSeries := []model.Series{}
		if err != nil {
			tempSeries = nil
			goto returnLabel3
		}

		for rows.Next() {
			temp := model.Series{}
			err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
			if err != nil {
				log.Println("a series couldn't be scanned")
				continue
			}
			tempSeries = append(tempSeries, temp)
		}
	returnLabel3:
		return &tempMovies, &tempSeries, nil
	}

	//there is no filter. So just get lists
	tempMovies, tempSeries, err := dbi.GetFavoriteContents(c, page, items)
	if err != nil {
		return nil, nil, err
	}

	return tempMovies, tempSeries, nil
}
