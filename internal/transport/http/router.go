package httpapi

import "net/http"

func NewRouter(authHandler *AuthHandler, peopleHandler *PeopleHandler) http.Handler {
	mux := http.NewServeMux()

	authHandler.RegisterPublicRoutes(mux)

	fileServer := http.FileServer(http.Dir("web"))
	mux.Handle("/styles.css", fileServer)
	mux.Handle("/login.js", fileServer)
	mux.Handle("/admin.js", fileServer)
	mux.Handle("/entrance.js", fileServer)

	mux.Handle("/", authHandler.RequireAuth(http.HandlerFunc(authHandler.HandleHome)))
	mux.Handle("/admin", authHandler.RequireAuth(http.HandlerFunc(authHandler.HandleAdminPage)))
	mux.Handle("/entrance", authHandler.RequireAuth(http.HandlerFunc(authHandler.HandleEntrancePage)))

	mux.Handle("/people/check-in", authHandler.RequireAuth(http.HandlerFunc(peopleHandler.CheckInHandler())))
	mux.Handle("/people/check-in/undo", authHandler.RequireAuth(http.HandlerFunc(peopleHandler.UndoCheckInHandler())))
	mux.Handle("/people/suggest", authHandler.RequireAuth(http.HandlerFunc(peopleHandler.SuggestValuesHandler())))
	mux.Handle("/people", authHandler.RequireAdmin(http.HandlerFunc(peopleHandler.CreatePersonHandler())))
	mux.Handle("/people/delete", authHandler.RequireAdmin(http.HandlerFunc(peopleHandler.DeletePersonHandler())))
	mux.Handle("/people/students/delete-all", authHandler.RequireAdmin(http.HandlerFunc(peopleHandler.DeleteStudentsHandler())))
	mux.Handle("/people/import/students", authHandler.RequireAdmin(http.HandlerFunc(peopleHandler.ImportStudentsHandler())))
	mux.Handle("/people/import/teachers", authHandler.RequireAdmin(http.HandlerFunc(peopleHandler.ImportTeachersHandler())))
	mux.Handle("/people/list", authHandler.RequireAdmin(http.HandlerFunc(peopleHandler.ListPeopleHandler())))
	mux.Handle("/people/stats/programs", authHandler.RequireAdmin(http.HandlerFunc(peopleHandler.ProgramStatsHandler())))

	return mux
}
