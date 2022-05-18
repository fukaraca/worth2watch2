package admin

import (
	"encoding/json"
	"fmt"
	"github.com/fukaraca/worth2watch/model"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

//findIMDBWithID finds and returns IMDB ID for given TMDB ID
func findIMDBWithID(id int) (string, error) {
	getUrl := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/external_ids?api_key=%s", id, API_KEY)
	respExtID, err := http.Get(getUrl)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer respExtID.Body.Close()

	read, err := io.ReadAll(respExtID.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	_, after, ok := strings.Cut(string(read), "\"imdb_id\":\"")
	if !ok {
		return "", fmt.Errorf("imdb_id not found")
	}
	before, _, ok := strings.Cut(after, "\",\"")
	if !ok {
		return "", fmt.Errorf("imdb_id not found")
	}

	return before, nil
}

func GetSeries(id int) *model.Series {
	getUrl := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d?api_key=%s", id, API_KEY)
	resp, err := http.Get(getUrl)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	read, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	seriesFromAPI := new(SeriesAPI)
	err = json.Unmarshal(read, &seriesFromAPI)
	if err != nil {
		log.Println("unmarshall failed:", err)
		return nil
	}

	//get cast and crew
	getUrl = fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/aggregate_credits?api_key=%s", id, API_KEY)
	respCast, err := http.Get(getUrl)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer respCast.Body.Close()

	read, err = io.ReadAll(respCast.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	castFromAPI := new(SeriesCastAPI)
	err = json.Unmarshal(read, &castFromAPI)
	if err != nil {
		log.Println("unmarshall failed:", err)
		return nil
	}

	//construct the series struct
	ret := model.Series{}
	//Title
	ret.Title = &seriesFromAPI.Name
	//Description
	ret.Description = &seriesFromAPI.Overview
	//Rating
	ret.Rating = seriesFromAPI.VoteAverage
	//Released Date
	parsed, err := time.Parse("2006-01-02", seriesFromAPI.FirstAirDate)
	if err != nil {
		{
			parsed, err = time.Parse("2006/01/02", seriesFromAPI.FirstAirDate)
			if err != nil {
				parsed, err = time.Parse("2006.01.02", seriesFromAPI.FirstAirDate)
				if err != nil {
					log.Println("release date couldn't be parsed for:", seriesFromAPI.Name, "given time format was:", seriesFromAPI.FirstAirDate, "err:", err)
				}
			}
		}
	}
	err = ret.ReleaseDate.Set(parsed)
	if err != nil {
		log.Println("release date couldn't be set for pgtype", err)
	}
	//Directors
	directors := castFromAPI.getDirectors("Director")
	for director := range directors {
		ret.Directors = append(ret.Directors, director)
	}
	//Writers
	writers := castFromAPI.getWriters("Writer")
	for writer := range writers {

		ret.Writers = append(ret.Writers, writer)
	}
	//Stars
	stars := castFromAPI.getStars(5, 6, seriesFromAPI.NumberOfEpisodes)
	for star := range stars {
		ret.Stars = append(ret.Stars, star)
	}
	//Duration
	ret.Duration = seriesFromAPI.EpisodeRunTime[0]
	//IMDB ID

	if imdb, err := findIMDBWithID(id); err != nil {
		log.Println(err)
	} else {
		ret.IMDBid = &imdb
	}
	//Year
	ret.Year = ret.ReleaseDate.Time.Year()
	//Genres
	for _, genre := range seriesFromAPI.Genres {

		ret.Genres = append(ret.Genres, genre.Name)
	}
	//seasons
	ret.Seasons = seriesFromAPI.NumberOfSeasons
	return &ret
}

//getDirectors is a helper func for GetSeries
func (crew *SeriesCastAPI) getDirectors(jobTitle string) map[string]struct{} {
	ret := make(map[string]struct{})
	var empty struct{}
	for _, s := range crew.Crew {
		for _, job := range s.Jobs {
			if job.Job == jobTitle {
				ret[s.Name] = empty
				break
			}
		}
	}
	return ret
}

//getWriters is a helper func for GetSeries
func (crew *SeriesCastAPI) getWriters(jobTitle string) map[string]struct{} {
	ret := make(map[string]struct{})
	var empty struct{}
	for _, s := range crew.Crew {
		for _, job := range s.Jobs {
			if job.Job == jobTitle {
				ret[s.Name] = empty
				break
			}
		}
	}
	return ret
}

//getStars is helper func for GetSeries.For given popularity and amount, it looks up for it among the cast and retuns that match
func (cast *SeriesCastAPI) getStars(popularity float64, count, totalEpisodesOfSerie int) map[string]struct{} {
	ret := make(map[string]struct{})
	type temp struct {
		name   string
		rating float64
	}
	populers := []temp{}
	var empty struct{}
	for _, s := range cast.Cast {
		if s.Popularity > popularity && s.KnownForDepartment == "Acting" && s.TotalEpisodeCount > (totalEpisodesOfSerie/2) {
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
