package db

import (
	"context"
	"fmt"
	"github.com/fukaraca/worth2watch2/api/admin"
	"github.com/fukaraca/worth2watch2/model"
	"github.com/gin-gonic/gin"

	"log"
)

//AddMovieContentWithID inserts movie to DBService..
func (dbi *dbImp) AddMovieContentWithID(imdb string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	id, err := admin.FindIDWithIMDB(imdb)
	if err != nil {
		return
	}
	movie := admin.GetMovie(id)
	err = dbi.AddMovieContentWithStruct(ctx, movie)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("movie: ", *movie.Title, " succesfully added")
	}
}

//AddSeriesContentWithID adds series to DBService with its seasons.
func (dbi *dbImp) AddSeriesContentWithID(imdb string) {

	id, err := admin.FindIDWithIMDB(imdb)
	if err != nil {
		return
	}

	series := admin.GetSeries(id)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = dbi.AddSeriesContentWithStruct(ctx, series)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(*series.Title, " saved to DB successfully")
	}

}

func (dbi *dbImp) AddMovieContentWithStruct(ctx context.Context, movie *model.Movie) error {
	ctx1, cancel1 := context.WithTimeout(ctx, model.TIMEOUT)
	defer cancel1()
	_, err := dbi.conn.Exec(ctx1, "INSERT INTO movies (movie_id,title,description,rating,release_date,directors,writers,stars,duration_min,imdb_id,year,genres,audios,subtitles) VALUES (nextval('movies_movie_id_seq'),$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13);", movie.Title, movie.Description, movie.Rating, movie.ReleaseDate, movie.Directors, movie.Writers, movie.Stars, movie.Duration, movie.IMDBid, movie.Year, movie.Genres, movie.Audios, movie.Subtitles)
	if err != nil {
		return fmt.Errorf("insert for movie %s failed: %v", *movie.Title, err)
	}

	return nil
}

func (dbi *dbImp) AddSeriesContentWithStruct(ctx context.Context, series *model.Series) error {
	ctx1, cancel1 := context.WithCancel(ctx)
	defer cancel1()
	//insert series
	_, err := dbi.conn.Exec(ctx1, "INSERT INTO series (serie_id,title,description,rating,release_date,directors,writers,stars,duration_min,imdb_id,year,genres,seasons) VALUES (nextval('series_serie_id_seq'),$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);", series.Title, series.Description, series.Rating, series.ReleaseDate, series.Directors, series.Writers, series.Stars, series.Duration, series.IMDBid, series.Year, series.Genres, series.Seasons)
	if err != nil {
		return err
	}

	row := dbi.conn.QueryRow(ctx1, "SELECT serie_id FROM series WHERE imdb_id=$1;", series.IMDBid)
	seriesId := 0
	err = row.Scan(&seriesId)
	if err != nil {
		return fmt.Errorf("series_id couldn't be scanned from db: %v", err)
	}

	//insert seasons
	for i := 1; i < series.Seasons+1; i++ {
		season, episodes := admin.GetSeason(series, i)

		ctx2, cancel2 := context.WithTimeout(ctx, model.TIMEOUT)
		defer cancel2()
		_, err = dbi.conn.Exec(ctx2, "INSERT INTO seasons (season_id,season_number,episodes,imdb_id) VALUES (nextval('seasons_season_id_seq'),$1,$2,$3);", season.SeasonNumber, season.Episodes, season.IMDBid)
		if err != nil {
			return fmt.Errorf("insert season %d for %s failed: %v", season.SeasonNumber, *series.Title, err)
		}
		//foreign key assignment for seasons
		_, err = dbi.conn.Exec(ctx2, "UPDATE seasons SET serie_id=$1 WHERE imdb_id=$2;", seriesId, season.IMDBid)
		if err != nil {
			return fmt.Errorf("update  season %d for %s failed when FK assignment: %v", season.SeasonNumber, *series.Title, err)
		}

		row := dbi.conn.QueryRow(ctx2, "SELECT season_id FROM seasons WHERE serie_id=$1 AND season_number=$2;", seriesId, season.SeasonNumber)
		seasonId := 0
		err = row.Scan(&seasonId)
		if err != nil {
			return fmt.Errorf("season_id couldn't be scanned from db : %v", err)
		}

		//insert episodes
		for _, episode := range episodes {

			ctx3, cancel3 := context.WithTimeout(ctx, model.TIMEOUT)
			defer cancel3()
			_, err = dbi.conn.Exec(ctx3, "INSERT INTO episodes (episode_id,title,description,rating,release_date,directors,writers,stars,duration_min,imdb_id,year,audios,subtitles,episode_number) VALUES (nextval('episodes_episode_id_seq'),$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13);", episode.Title, episode.Description, episode.Rating, episode.ReleaseDate, episode.Directors, episode.Writers, episode.Stars, episode.Duration, episode.IMDBid, episode.Year, episode.Audios, episode.Subtitles, episode.EpisodeNumber)
			if err != nil {
				return fmt.Errorf("insert episode %d for %s failed: %v", episode.EpisodeNumber, *series.Title, err)
			}

			//foreign key assignment for episodes
			_, err = dbi.conn.Exec(ctx3, "UPDATE episodes SET season_id=$1 WHERE imdb_id=$2;", seasonId, episode.IMDBid)
			if err != nil {
				return fmt.Errorf("update  season %d for %s failed when fk assignment: %v", episode.EpisodeNumber, *series.Title, err)
			}
		}
	}
	return nil
}

//DeleteContent deletes given content from DBService
func (dbi *dbImp) DeleteContent(c *gin.Context, id, contentType string) error {
	ctx, cancel := context.WithTimeout(c.Request.Context(), model.TIMEOUT)
	defer cancel()

	switch contentType {
	case "movie":
		_, err := dbi.conn.Exec(ctx, "DELETE FROM favorite_movies WHERE movie_id=(SELECT movie_id FROM movies WHERE imdb_id=$1);", id)
		_, err = dbi.conn.Exec(ctx, "DELETE FROM movies WHERE imdb_id=$1;", id)
		if err != nil {
			return err
		}
	case "series":
		//
		_, err := dbi.conn.Exec(ctx, "DELETE FROM episodes WHERE season_id=(SELECT season_id FROM seasons WHERE imdb_id=$1);", id)
		_, err = dbi.conn.Exec(ctx, "DELETE FROM seasons WHERE serie_id=(SELECT serie_id FROM series WHERE imdb_id=$1);", id)
		_, err = dbi.conn.Exec(ctx, "DELETE FROM favorite_series WHERE serie_id=(SELECT serie_id FROM series WHERE imdb_id=$1);", id)
		_, err = dbi.conn.Exec(ctx, "DELETE FROM series WHERE imdb_id=$1;", id)
		if err != nil {
			return err
		}
	}
	return nil
}
