package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

//Route defines a struct for addressing paths to handlers
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

//Routes will need to be an array
type Routes []Route

//NewRouter returns a router for the various handlers
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}

var routes = Routes{
	Route{
		"dbSearchText",
		"GET",
		"/getMsText/{ms}",
		dbDiplomaticText,
	},
	Route{
		"dbSearchText",
		"GET",
		"/getMsCanText/{ms}",
		dbCanonText,
	},
	Route{
		"dbSearchText",
		"GET",
		"/getSynopText/{composition}",
		dbSynopticText,
	},
	Route{
		"dbSearchComp",
		"GET",
		"/getCompositions",
		dbSearchComp,
	},
	Route{
		"dbSearchBooks",
		"GET",
		"/getCompBooks/{composition}",
		dbGetCompBooks,
	},
	Route{
		"dbGetCompChapters",
		"GET",
		"/getCompChapters/{composition}/{book}",
		dbGetCompChapters,
	},
	Route{
		"dbGetCompVerses",
		"GET",
		"/getCompVerses/{composition}/{book}/{chapter}",
		dbGetCompVerses,
	},
	Route{
		"dbGetMssOfComp",
		"GET",
		"/getMssOfComp/{composition}/{book}/{chapter}/{verse}",
		dbGetMssOfComp,
	},
	Route{
		"dbGetManuscripts",
		"GET",
		"/getManuscripts",
		dbGetManuscripts,
	},
	Route{
		"dbGetManuscriptPages",
		"GET",
		"/getManuscriptPages/{manuscript}",
		dbGetManuscriptPages,
	},
	Route{
		"dbGetManuscriptColumns",
		"GET",
		"/getManuscriptColumns/{manuscript}/{page}",
		dbGetManuscriptColumns,
	},
	Route{
		"dbGetManuscriptLines",
		"GET",
		"/getManuscriptLines/{manuscript}/{page}/{column}",
		dbGetManuscriptLines,
	},
	Route{
		"dbCompositionMsText",
		"GET",
		"/getCompositionMSText/{ms_id}/{ms_name}",
		dbCompositionMsText,
	},
	Route{
		"dbSearchMS",
		"GET",
		"/getMss",
		dbSearchMS,
	},
	Route{
		"dbSearchBookNum",
		"GET",
		"/getBookNums/{composition}",
		dbSearchBookNum,
	},
	Route{
		"dbSearchChapter",
		"GET",
		"/getChapters/{composition}/{booknum}",
		dbSearchChapter,
	},
	Route{
		"dbSearchVerse",
		"GET",
		"/getVerses/{composition}/{booknum}/{chapter}",
		dbSearchVerse,
	},
	Route{
		"dbSearchVocable",
		"GET",
		"/getVocables/{composition}/{booknum}/{chapter}/{verse}",
		dbSearchVocable,
	},
	Route{
		"dbVocableLinks",
		"GET",
		"/getVocableLinks/{composition}",
		dbVocableLinks,
	},
	Route{
		"htmlFromDB",
		"GET",
		"/htmlFromDB/{composition}/{ms1}/{ms2}",
		htmlFromDB,
	},
	Route{
		"verseVssFromDB",
		"GET",
		"/verseVssFromDB/{composition}/{booknum}/{chapter}/{verse}",
		verseVssFromDB,
	},
	Route{
		"linkVocables",
		"GET",
		"/linkVocables/{voc1}/{voc2}",
		linkVocables,
	},
	Route{
		"unlinkVocables",
		"GET",
		"/unlinkVocables/{voc1}/{voc2}",
		unlinkVocables,
	},
}
