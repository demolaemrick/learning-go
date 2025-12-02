package main

import (
	"net/http"
	"strconv"

	"github.com/demolaemrick/social/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *application) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	user, err := app.store.Users.GetByID(ctx, userId)

	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}
