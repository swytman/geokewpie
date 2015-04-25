package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

var routes = Routes{
	Route{
		"get locations",
		"GET",
		"/locations",
		getLocationsHandler,
	},
	Route{
		"post locations",
		"POST",
		"/locations",
		postLocationsHandler,
	},
	Route{
		"new user",
		"POST",
		"/users",
		createUserHandler,
	},
	Route{
		"user check",
		"GET",
		"/users/check",
		checkUserHandler,
	},
	Route{
		"find user by login",
		"GET",
		"/users/find/{letters}",
		findUserByLettersHandler,
	},
	Route{
		"new session",
		"POST",
		"/session",
		postSessionHandler,
	},
	Route{
		"get session",
		"GET",
		"/session",
		getSessionHandler,
	},
	Route{
		"add following",
		"POST",
		"/followings/{login}",
		postFollowingsHandler,
	},
	Route{
		"delete following",
		"DELETE",
		"/followings/{login}",
		deleteFollowingsHandler,
	},
	Route{
		"ask locations",
		"GET",
		"/followings/update_locations",
		askFollowingsLocationsHandler,
	},
	Route{
		"get followings",
		"GET",
		"/followings",
		getFollowingsHandler,
	},
	Route{
		"add follower",
		"POST",
		"/followers/{login}",
		postFollowersHandler,
	},
	Route{
		"delete follower",
		"DELETE",
		"/followers/{login}",
		deleteFollowersHandler,
	},
	Route{
		"get followers",
		"GET",
		"/followers",
		getFollowersHandler,
	},
	Route{
		"get devices",
		"GET",
		"/devices",
		getDevicesHandler,
	},
	Route{
		"post devices",
		"POST",
		"/devices",
		postDevicesHandler,
	},
	Route{
		"delete devices",
		"DELETE",
		"/devices",
		deleteDevicesHandler,
	},

	Route{
		"basic logs",
		"GET",
		"/logs",
		getLogsHandler,
	},
	Route{
		"gcm logs",
		"GET",
		"/gcmlogs",
		getGcmLogsHandler,
	},
}
