package main

import (
	"net/http"
	"github.com/demolaemrick/social/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	fq := store.Pagination{
		Limit:  10,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.ParsePagination(r)

	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	feed, err := app.store.Posts.GetUserFeed(r.Context(), int64(80), fq)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
