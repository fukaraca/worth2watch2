package db

import (
	"context"
	"fmt"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/fukaraca/worth2watch2/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"log"
	"strconv"
)

//GetThisMovieFromDB queries for given IMDB_id
func (dbi *dbImp) GetThisMovieFromDB(c *gin.Context, id string) (*model.Movie, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	row := dbi.conn.QueryRow(ctx, "SELECT * FROM movies WHERE imdb_id=$1", id)
	temp := new(model.Movie)

	err := row.Scan(&temp.MovieID, &temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.Directors, &temp.Writers, &temp.Stars, &temp.Duration, &temp.IMDBid, &temp.Year, &temp.Genres, &temp.Audios, &temp.Subtitles)

	if err != nil {
		return nil, err
	}
	return temp, nil
}

//GetThisSeriesFromDB queries for given series(IMDB_id) and its seasons.
//For querying episodes, another query must be conducted.
func (dbi *dbImp) GetThisSeriesFromDB(c *gin.Context, id string) (*model.Series, *[]model.Seasons, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	//get series
	row := dbi.conn.QueryRow(ctx, "SELECT * FROM series WHERE imdb_id=$1", id)
	tempSeries := new(model.Series)

	err := row.Scan(&tempSeries.SerieID, &tempSeries.Title, &tempSeries.Description, &tempSeries.Rating, &tempSeries.ReleaseDate, &tempSeries.Directors, &tempSeries.Writers, &tempSeries.Stars, &tempSeries.Duration, &tempSeries.IMDBid, &tempSeries.Year, &tempSeries.Genres, &tempSeries.Seasons)
	fmt.Println(tempSeries, id)
	if err != nil {
		return nil, nil, err
	}
	//get seasons
	rows, err := dbi.conn.Query(ctx, "SELECT * FROM seasons WHERE imdb_id=$1", id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return tempSeries, nil, fmt.Errorf("there is no season for given series")
		} else {
			return nil, nil, err
		}
	}
	defer rows.Close()
	tempSeasons := []model.Seasons{}
	for rows.Next() {
		tempSeason := model.Seasons{}
		err = rows.Scan(&tempSeason.SeasonID, &tempSeason.IMDBid, &tempSeason.SeasonNumber, &tempSeason.Episodes, &tempSeason.SerieID)
		if err != nil {
			log.Println("season couldn't be get from db", err)
			continue
		}
		tempSeasons = append(tempSeasons, tempSeason)
	}

	return tempSeries, &tempSeasons, nil

}

//GetEpisodesForaSeasonFromDB queries for episode of a certain season of given series(IMDB_id).
func (dbi *dbImp) GetEpisodesForaSeasonFromDB(c *gin.Context, seriesID, sN string) (*[]model.Episodes, error) {
	seasonNumber, err := strconv.Atoi(sN)
	if err != nil {
		log.Println("invalid season number", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()

	row := dbi.conn.QueryRow(ctx, "SELECT season_id FROM seasons WHERE imdb_id=$1 AND season_number=$2", seriesID, seasonNumber)
	var seasonID int
	err = row.Scan(&seasonID)
	if err != nil {
		log.Println("season id couldn't be get from db", err)
		return nil, err
	}
	rows, err := dbi.conn.Query(ctx, "SELECT * FROM episodes WHERE season_id=$1", seasonID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	tempSeason := []model.Episodes{}
	for rows.Next() {
		ep := model.Episodes{}
		err = rows.Scan(&ep.EpisodeID, &ep.Title, &ep.Description, &ep.Rating, &ep.ReleaseDate, &ep.Directors, &ep.Writers, &ep.Stars, &ep.Duration, &ep.IMDBid, &ep.Year, &ep.Audios, &ep.Subtitles, &ep.SeasonID, &ep.EpisodeNumber)
		if err != nil {
			log.Printf("episode %s couldn't be get from db for serie%s\n", sN, seriesID)
			continue
		}
		tempSeason = append(tempSeason, ep)
	}

	return &tempSeason, nil

}

//GetMoviesListWithPage queries for given page and amount of items as ordered by rating
func (dbi *dbImp) GetMoviesListWithPage(c *gin.Context, page, items int) (*[]model.Movie, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	offset := (page - 1) * items
	rows, err := dbi.conn.Query(ctx, "SELECT * FROM movies ORDER BY rating DESC OFFSET $1 LIMIT $2;", offset, items)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	tempMovies := []model.Movie{}
	for rows.Next() {
		temp := model.Movie{}
		err = rows.Scan(&temp.MovieID, &temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.Directors, &temp.Writers, &temp.Stars, &temp.Duration, &temp.IMDBid, &temp.Year, &temp.Genres, &temp.Audios, &temp.Subtitles)
		if err != nil {
			log.Println("a movie couldn't be scanned")
			continue
		}
		tempMovies = append(tempMovies, temp)

	}
	return &tempMovies, nil
}

//GetSeriesListWithPage queries for given page and amount of items as ordered by rating
func (dbi *dbImp) GetSeriesListWithPage(c *gin.Context, page, items int) (*[]model.Series, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	offset := (page - 1) * items
	rows, err := dbi.conn.Query(ctx, "SELECT * FROM series ORDER BY rating DESC OFFSET $1 LIMIT $2;", offset, items)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	tempSeries := []model.Series{}
	for rows.Next() {
		temp := model.Series{}
		err = rows.Scan(&temp.SerieID, &temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.Directors, &temp.Writers, &temp.Stars, &temp.Duration, &temp.IMDBid, &temp.Year, &temp.Genres, &temp.Seasons)
		if err != nil {
			log.Println("a serie couldn't be scanned")
			continue
		}
		tempSeries = append(tempSeries, temp)

	}
	return &tempSeries, nil
}

//SearchContent search db for both movie and series with case insensitive regexp and full match with genres
func (dbi *dbImp) SearchContent(c *gin.Context, name string, genres []string, page, items int) (*[]model.Movie, *[]model.Series, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	offset := (page - 1) * items
	name = *util.Striper(name)
	if name != "" && len(genres) > 0 {
		//both name and genres are requested (OR Conditional)
		//first search for movie
		rows, err := dbi.conn.Query(ctx, "SELECT DISTINCT title,description,rating,release_date, imdb_id,genres FROM movies WHERE title ~* $1 OR movies.genres && $2 ORDER BY rating DESC OFFSET $3 LIMIT $4 ;", name, genres, offset, items)
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
		rows, err = dbi.conn.Query(ctx, "SELECT DISTINCT title,description,rating,release_date, imdb_id,genres FROM series WHERE title ~* $1 OR series.genres && $2 ORDER BY rating DESC OFFSET $3 LIMIT $4;", name, genres, offset, items)
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
		rows, err := dbi.conn.Query(ctx, "SELECT DISTINCT title,description,rating,release_date, imdb_id,genres FROM movies WHERE genres && $1 ORDER BY rating DESC OFFSET $2 LIMIT $3;", genres, offset, items)
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
		rows, err = dbi.conn.Query(ctx, "SELECT DISTINCT title,description,rating,release_date, imdb_id,genres FROM series WHERE genres && $1 ORDER BY rating DESC OFFSET $2 LIMIT $3;", genres, offset, items)
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
		rows, err := dbi.conn.Query(ctx, "SELECT DISTINCT title,description,rating,release_date, imdb_id,genres FROM movies WHERE title ~* $1 ORDER BY rating DESC OFFSET $2 LIMIT $3;", name, offset, items)
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
		rows, err = dbi.conn.Query(ctx, "SELECT DISTINCT title,description,rating, release_date, imdb_id,genres FROM series WHERE title ~* $1 ORDER BY rating DESC OFFSET $2 LIMIT $3;", name, offset, items)
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
	tempMovies, _ := dbi.GetMoviesListWithPage(c, page, items)
	tempSeries, _ := dbi.GetSeriesListWithPage(c, page, items)

	return tempMovies, tempSeries, nil
}

//FindSimilarContent queries for similar content and returns a specific amount of item from DBService.
//Similarity is determined according to popularity
func (dbi *dbImp) FindSimilarContent(c *gin.Context, id, cType string) (*[]model.Movie, *[]model.Series, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()
	amount := 2
	if cType == "movie" {
		//requested content type is series
		rows, err := dbi.conn.Query(ctx, "SELECT DISTINCT title,description,rating,release_date, imdb_id,genres FROM movies WHERE genres && (SELECT genres FROM movies WHERE imdb_id=$1) AND imdb_id!=$1 ORDER BY rating DESC  LIMIT $2 ;", id, amount)
		defer rows.Close()
		if err != nil {
			return nil, nil, err
		}
		tempMovies := []model.Movie{}
		for rows.Next() {
			temp := model.Movie{}
			err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
			if err != nil {
				log.Println("a movie couldn't be scanned")
				continue
			}
			tempMovies = append(tempMovies, temp)
		}
		return &tempMovies, nil, nil
	}
	//requested content type is series
	rows, err := dbi.conn.Query(ctx, "SELECT DISTINCT title,description,rating,release_date, imdb_id,genres FROM series WHERE genres && (SELECT genres FROM series WHERE imdb_id=$1) AND imdb_id!=$1 ORDER BY rating DESC  LIMIT $2 ;", id, amount)
	defer rows.Close()
	if err != nil {
		return nil, nil, err
	}
	tempSeries := []model.Series{}
	for rows.Next() {
		temp := model.Series{}
		err = rows.Scan(&temp.Title, &temp.Description, &temp.Rating, &temp.ReleaseDate, &temp.IMDBid, &temp.Genres)
		if err != nil {
			log.Println("a serie couldn't be scanned")
			continue
		}
		tempSeries = append(tempSeries, temp)
	}
	return nil, &tempSeries, nil

}
