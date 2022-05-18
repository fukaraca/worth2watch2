package admin

import (
	"encoding/json"
	"fmt"
	"github.com/fukaraca/worth2watch/config"
	"github.com/fukaraca/worth2watch/model"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var API_KEY = config.GetEnv.GetString("API_KEY")

//FindIDWithIMDB searches and returns movie/series id at TMDB for given IMDB ID
func FindIDWithIMDB(imdbID string) (int, error) {
	getUrl := fmt.Sprintf("https://api.themoviedb.org/3/find/%s?api_key=%s&external_source=imdb_id", imdbID, API_KEY)
	resp, err := http.Get(getUrl)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	defer resp.Body.Close()

	read, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	_, after, ok := strings.Cut(string(read), "\"id\":")
	if !ok {
		return 0, fmt.Errorf("id not found", string(read))
	}
	before, _, ok := strings.Cut(after, ",")
	if !ok {
		return 0, fmt.Errorf("id not found")
	}
	ret, err := strconv.Atoi(before)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

//GetMovie returns Movie struct for given TMDB movie ID
func GetMovie(id int) *model.Movie {
	getUrl := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s", id, API_KEY)
	resp, err := http.Get(getUrl)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()

	read, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	movieFromAPI := new(MovieAPI)
	err = json.Unmarshal(read, &movieFromAPI)
	if err != nil {
		fmt.Println("unmarshall failed:", err)
		return nil
	}

	//
	//construct the movie struct
	ret := model.Movie{}

	//title
	ret.Title = &movieFromAPI.Title
	//description
	ret.Description = &movieFromAPI.Overview
	//release date
	parsed, err := time.Parse("2006-01-02", movieFromAPI.ReleaseDate)
	if err != nil {
		{
			parsed, err = time.Parse("2006/01/02", movieFromAPI.ReleaseDate)
			if err != nil {
				parsed, err = time.Parse("2006.01.02", movieFromAPI.ReleaseDate)
				if err != nil {
					log.Println("release date couldn't be parsed for:", movieFromAPI.Title, "err:", err, "time format:", movieFromAPI.ReleaseDate)
				}
			}
		}
	}
	err = ret.ReleaseDate.Set(parsed)
	if err != nil {
		log.Println("release date couldn't be set for pgtype", err)
	}
	//vote
	ret.Rating = movieFromAPI.VoteAverage
	//Cast and crew
	castFromAPI, err := getCastForMovie(id)
	if err != nil {
		log.Println("cast and crew data for movie couldn't be gotten:", err)
	} else {
		//Directors
		directors := castFromAPI.getDirectors("Director")
		for director := range directors {
			ret.Directors = append(ret.Directors, director)
		}
		//Writers
		writers := castFromAPI.getWriters("Writing", "Writer")
		for writer := range writers {
			ret.Writers = append(ret.Writers, writer)
		}
		//Stars
		stars := castFromAPI.getStars(5, 5)
		for star := range stars {
			ret.Stars = append(ret.Stars, star)
		}
	}

	//Duration
	ret.Duration = movieFromAPI.Runtime
	//IMDB ID
	ret.IMDBid = &movieFromAPI.ImdbID
	//Year
	ret.Year = ret.ReleaseDate.Time.Year()
	//Genre
	for _, genre := range movieFromAPI.Genres {
		ret.Genres = append(ret.Genres, genre.Name)
	}
	//Audio and subtitle
	translationFromAPI, err := getTranslationDataOfMovie(id)
	if err != nil {
		log.Println("translation data for movie couldn't be gotten:", err)
	} else {
		for _, translation := range translationFromAPI.Translations {
			ret.Audios = append(ret.Audios, translation.EnglishName)
			ret.Subtitles = append(ret.Subtitles, translation.EnglishName)
		}
	}

	log.Println(ret.Title, " movie has been succesfully fetched")
	return &ret
}

//getDirectors is a helper func for GetMovie
func (crew *CastAPI) getDirectors(jobTitle string) map[string]struct{} {
	ret := make(map[string]struct{})
	var empty struct{}
	for _, s := range crew.Crew {
		if s.Job == jobTitle {
			ret[s.Name] = empty
		}
	}
	return ret
}

//getWriters is a helper func for GetMovie
func (crew *CastAPI) getWriters(department, jobTitle string) map[string]struct{} {
	ret := make(map[string]struct{})
	var empty struct{}
	for _, s := range crew.Crew {
		if s.Department == department || s.Job == jobTitle {
			ret[s.Name] = empty
		}
	}
	return ret
}

//getStars is a helper func for GetMovie. For given popularity and amount, it looks up for it among the cast
func (cast *CastAPI) getStars(popularity float64, count int) map[string]struct{} {
	ret := make(map[string]struct{})
	type temp struct {
		name   string
		rating float64
	}
	populers := []temp{}
	var empty struct{}
	for _, s := range cast.Cast {
		if s.Popularity > popularity {
			populers = append(populers, temp{
				name:   s.Name,
				rating: s.Popularity,
			})
		}
	}
	sort.Slice(populers, func(i, j int) bool {
		return populers[i].rating > populers[j].rating
	})
	for i := 0; i < count && i < len(populers); i++ {
		ret[populers[i].name] = empty
	}
	return ret
}

//getTranslationDataOfMovie returns translation data for given movie
func getTranslationDataOfMovie(id int) (*TranslationAPI, error) {

	getUrl := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/translations?api_key=%s", id, API_KEY)
	respTranslate, err := http.Get(getUrl)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer respTranslate.Body.Close()

	read, err := io.ReadAll(respTranslate.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	translationFromAPI := new(TranslationAPI)
	err = json.Unmarshal(read, &translationFromAPI)
	if err != nil {
		fmt.Println("unmarshall failed:", err)
		return nil, err
	}
	return translationFromAPI, nil
}

//getCastForMovie returns cast and crew data
func getCastForMovie(id int) (*CastAPI, error) {
	//get cast and crew
	getUrl := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/credits?api_key=%s", id, API_KEY)
	respCast, err := http.Get(getUrl)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer respCast.Body.Close()

	read, err := io.ReadAll(respCast.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	castFromAPI := new(CastAPI)
	err = json.Unmarshal(read, &castFromAPI)
	if err != nil {
		fmt.Println("unmarshall failed:", err)
		return nil, err
	}
	return castFromAPI, nil
}
