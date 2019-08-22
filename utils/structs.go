package utils

import (
	"math"
	"time"
)

// AlbumItems to hold an array of Album
type SimplifiedAlbum struct {
	AlbumType            string             `json:"album_type"`
	Artists              []SimplifiedArtist `json:"artists"`
	AvailableMarkets     []string           `json:"available_markets"`
	ExternalUrls         ExternalUrl        `json:"external_urls"`
	Href                 string             `json:"href"`
	Id                   string             `json:"id"`
	Images               []AlbumArt         `json:"images"`
	Name                 string             `json:"name"`
	ReleaseDate          string             `json:"release_date"`
	ReleaseDatePrecision string             `json:"release_date_precision"`
	Restrictions         string             `json:"restrictions"`
	Type                 string             `json:"type"`
	Uri                  string             `json:"uri"`
}

// AlbumArt to hold images
type AlbumArt struct {
	Height int    `json:"height"`
	Url    string `json:"url"`
	Width  int    `json:"width"`
}

// ExternalUrl struct for storing external urls to spotify
type ExternalUrl struct {
	Spotify string `json:"spotify"`
}

// ExternalId struct for storing external urls to spotify
type ExternalId struct {
	Isrc string `json:"isrc"`
	Ean  string `json:"ean"`
	Upc  string `json:"upc"`
}

// SimplifiedArtist to hold an artist's info
type SimplifiedArtist struct {
	Href         string      `json:"href"`
	Id           string      `json:"id"`
	Name         string      `json:"name"`
	Type         string      `json:"artist"`
	Uri          string      `json:"uri"`
	ExternalUrls ExternalUrl `json:"external_urls"`
}

// Soundtrack for working with sound tracks obtained from albums
type SimplifiedSoundtrack struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// FullSoundtrack for working with storing full soundtrack object
type FullSoundtrack struct {
	Album            SimplifiedAlbum    `json:"album"`
	Artists          []SimplifiedArtist `json:"artists"`
	AvailableMarkets []string           `json:"available_markets"`
	DiscNumber       int                `json:"disc_number"`
	DurationMs       int                `json:"duration_ms"`
	Explicit         bool               `json:"explicit"`
	ExternalIds      ExternalId         `json:"external_ids"`
	ExternalUrls     ExternalUrl        `json:"external_urls"`
	Href             string             `json:"href"`
	Id               string             `json:"id"`
	IsPlayable       bool               `json:"is_playable"`
	LinkedFrom       string             `json:"linked_from"`
	Restrictions     string             `json:"restrictions"`
	Name             string             `json:"name"`
	Popularity       int                `json:"popularity"`
	PreviewUrl       string             `json:"preview_url"`
	TrackNumber      int                `json:"track_number"`
	Type             string             `json:"type"`
	Uri              string             `json:"uri"`
	IsLocal          bool               `json:"is_local"`
}

// RetryRequest is a mechanism to retry http requests after sometime
type RetryRequest struct {
	Attempt  int
	Max      int
	Min      int
	Duration time.Duration
}

// Backoff sets a time.Duration after which one should retry sending an http request
func (r *RetryRequest) Backoff(cooldownTime int) {
	multiplier := math.Pow(2, float64(r.Attempt)) * float64(r.Min)
	sleep := time.Duration(multiplier)

	if float64(sleep) != multiplier || float64(sleep) > float64(r.Max) {
		sleep = time.Duration(r.Max)
	}
	r.Duration = sleep + time.Duration(cooldownTime)
}
