package main

import (
	"net/http"

	"github.com/justinas/alice"
	"snippetbox.ren.dev/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET	/ping", ping)

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	protected := dynamic.Append(app.requireAuth)

	//{Open Routes}
	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /about", dynamic.ThenFunc(app.about))

	//{Snippet Routes}
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.snippetView))
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.postUserSignup))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.postUserLogin))

	//{Protected Routes}
	mux.Handle("GET /snippet/create", protected.ThenFunc(app.getSnippetCreate))
	mux.Handle("POST /snippet/create", protected.ThenFunc(app.postSnippetCreate))
	mux.Handle("GET /account/view", protected.ThenFunc(app.accountView))
	mux.Handle("POST /user/logout", protected.ThenFunc(app.postUserLogout))

	chain := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return chain.Then(mux)
}
