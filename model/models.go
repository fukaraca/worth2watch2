package model

import (
	"github.com/fukaraca/worth2watch2/config"
	"github.com/jackc/pgtype"
)

var TIMEOUT = config.GetEnv.GetDuration("TIMEOUT")
var ServerHost = config.GetEnv.GetString("SERVER_HOST")
var ServerPort = config.GetEnv.GetString("SERVER_PORT")

type User struct {
	UserID    int                `db.Conn:"user_id" json:"userID,omitempty"`
	Username  string             `db.Conn:"username" json:"username"`
	Password  string             `db.Conn:"password" json:"password"`
	Email     *string            `db.Conn:"email" json:"email"`
	Name      *string            `db.Conn:"name" json:"name"`
	Lastname  *string            `db.Conn:"lastname" json:"lastname"`
	CreatedOn pgtype.Timestamptz `db.Conn:"createdon" json:"createdOn"`
	LastLogin pgtype.Timestamptz `db.Conn:"lastlogin" json:"lastLogin"`
	Isadmin   bool               `db.Conn:"isadmin" json:"isAdmin"`
}

type Movie struct {
	MovieID     int         `db.Conn:"movie_id" json:"movieID,omitempty"`
	Title       *string     `db.Conn:"title" json:"title"`
	Description *string     `db.Conn:"description" json:"description"`
	Rating      float64     `db.Conn:"rating" json:"rating"`
	ReleaseDate pgtype.Date `db.Conn:"release_date" json:"releaseDate,omitempty"`
	Directors   []string    `db.Conn:"directors" json:"director,omitempty"`
	Writers     []string    `db.Conn:"writers" json:"writer,omitempty"`
	Stars       []string    `db.Conn:"stars" json:"stars,omitempty"`
	Duration    int         `db.Conn:"duration_min" json:"duration,omitempty"`
	IMDBid      *string     `db.Conn:"imdb_id" json:"IMDBid"`
	Year        int         `db.Conn:"year" json:"year,omitempty"`
	Genres      []string    `db.Conn:"genres" json:"genre,omitempty"`
	Audios      []string    `db.Conn:"audios" json:"audio,omitempty"`
	Subtitles   []string    `db.Conn:"subtitles" json:"subtitles,omitempty"`
}

type Series struct {
	SerieID     int          `db.Conn:"serie_id" json:"serieID,omitempty"`
	Title       *string      `db.Conn:"title" json:"title"`
	Description *string      `db.Conn:"description" json:"description"`
	Rating      float64      `db.Conn:"rating" json:"rating"`
	ReleaseDate *pgtype.Date `db.Conn:"release_date" json:"releaseDate"`
	Directors   []string     `db.Conn:"directors" json:"director,omitempty"`
	Writers     []string     `db.Conn:"writers" json:"writer,omitempty"`
	Stars       []string     `db.Conn:"stars" json:"stars,omitempty"`
	Duration    int          `db.Conn:"duration_min" json:"duration,omitempty"`
	IMDBid      *string      `db.Conn:"imdb_id" json:"IMDBid"`
	Year        int          `db.Conn:"year" json:"year,omitempty"`
	Genres      []string     `db.Conn:"genres" json:"genre,omitempty"`
	Seasons     int          `db.Conn:"seasons" json:"seasons"`
}

type Seasons struct {
	SeasonID     int     `db.Conn:"season_id" json:"seasonID,omitempty"`
	SeasonNumber int     `db.Conn:"season_number" json:"seasonNumber,omitempty"`
	IMDBid       *string `db.Conn:"imdb_id" json:"IMDBid"`
	Episodes     int     `db.Conn:"episodes" json:"episodes,omitempty"`
	SerieID      int     `db.Conn:"serie_id" json:"serieID,omitempty" json:"serieID,omitempty"`
}

type Episodes struct {
	EpisodeID     int          `db.Conn:"episode_id" json:"episodeID,omitempty"`
	Title         *string      `db.Conn:"title" json:"title"`
	Description   *string      `db.Conn:"description" json:"description"`
	Rating        float64      `db.Conn:"rating" json:"rating"`
	ReleaseDate   *pgtype.Date `db.Conn:"release_date" json:"releaseDate"`
	Directors     []string     `db.Conn:"directors" json:"director,omitempty"`
	Writers       []string     `db.Conn:"writers" json:"writer,omitempty"`
	Stars         []string     `db.Conn:"stars" json:"stars,omitempty"`
	Duration      int          `db.Conn:"duration_min" json:"duration,omitempty"`
	IMDBid        *string      `db.Conn:"imdb_id" json:"IMDBid"`
	Year          int          `db.Conn:"year" json:"year,omitempty"`
	Audios        []string     `db.Conn:"audios" json:"audio,omitempty"`
	Subtitles     []string     `db.Conn:"subtitles" json:"subtitles,omitempty"`
	SeasonID      int          `db.Conn:"season_id" json:"seasonID"`
	EpisodeNumber int          `db.Conn:"episode_number" json:"episodeNumber,omitempty"`
}
