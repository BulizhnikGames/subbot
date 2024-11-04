package api

import (
	"github.com/BulizhnikGames/subbot/fetcher/internal/fetcher"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type channelData struct {
	Username   string `json:"username"`
	ChannelID  int64  `json:"channel_id"`
	AccessHash int64  `json:"access_hash"`
}

type Api struct {
	server *http.Server
	f      *fetcher.Fetcher
}

func Init(f *fetcher.Fetcher, port string) *Api {
	api := &Api{f: f}
	router := chi.NewRouter()
	router.Post("/{channelName}", api.HandleSubscribe)
	api.server = &http.Server{
		Handler: router,
		Addr:    ":" + port,
	}
	return api
}

func (api *Api) Run() error {
	return api.server.ListenAndServe()
}

func (api *Api) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channelName")
	channelID, accessHash, err := api.f.SubscribeToChannel(r.Context(), channelName)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	responseWithJSON(w, http.StatusOK, channelData{
		Username:   channelName,
		ChannelID:  channelID,
		AccessHash: accessHash,
	})
}
