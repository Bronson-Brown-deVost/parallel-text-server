package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Bronson-Brown-deVost/gosqljson"
	"github.com/gorilla/mux"
)

type MsID struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type InnerText struct {
	Surface    string `json:"surface"`
	Normalized string `json:"normalized"`
}

type SplitRef struct {
	ID    int `json:"id"`
	Start int `json:"start"`
	End   int `json:"end"`
}

type MsModel struct {
	MsID      MsID                                     `json:"msID"`
	InnerText map[int]InnerText                        `json:"innerText"`
	MsPos     map[string]map[int]map[string][]SplitRef `json:"msPos"`
	CanPos    map[string]map[int]map[int]map[int][]int `json:"canPos"`
}

func writeHeader(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Accept-Charset", "utf-8")
	w.Header().Set("Content-Type", "application/json")
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Encoding", "gzip")
	// Gzip data
	gz := gzip.NewWriter(w)
	json.NewEncoder(gz).Encode(data)
	gz.Close()
}

func dbDiplomaticText(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	vars := mux.Vars(r)
	ms := vars["ms"]
	checkErr(db.Ping())
	rows, err := db.Query("CALL getDiplomaticMS(?)", ms)
	checkErr(err)
	var line, col, booknum, chapter, verse int
	var page, newSurface string
	var buffer bytes.Buffer
	buffer.WriteString("<head><style media=\"screen\" type=\"text/css\">.page {padding: 10px; display: table; table-layout: fixed; width:100%; height:100px;} .page div {display: table-cell; height:100px;}</style></head><body><div><div class=\"page\"><div class=\"column\">")
	for rows.Next() {
		var lineTmp, remainder, colTmp, booknumTmp, chapterTmp, verseTmp int
		var pageTmp, surface, newSurfaceTmp string
		err = rows.Scan(&surface, &pageTmp, &colTmp, &lineTmp, &remainder, &booknumTmp, &chapterTmp, &verseTmp)
		checkErr(err)
		if page != "" && page != pageTmp {
			page = pageTmp
			col = colTmp
			line = lineTmp
			buffer.WriteString("</div></div><div class=\"page\"><div class=\"column\"><span class=\"ms_ref\">" + page + ", " + strconv.Itoa(col) + "." + strconv.Itoa(line) + "</span>")
		} else if col != 0 && col != colTmp {
			col = colTmp
			line = lineTmp
			buffer.WriteString("</div><div class=\"column\"><span class=\"ms_ref\">" + page + ", " + strconv.Itoa(col) + "." + strconv.Itoa(line) + "</span>")
		} else if line != lineTmp || newSurface != "" || remainder != 0 {
			if remainder == 0 {
				if line != 0 {
					buffer.WriteString("</br>")
				}
				page = pageTmp
				col = colTmp
				line = lineTmp
				buffer.WriteString("<span class=\"ms_ref\">" + page + ", " + strconv.Itoa(col) + "." + strconv.Itoa(line) + "</span>")
			} else {
				surfaceRunes := []rune(surface)
				remSurface := string(surfaceRunes[:len(surfaceRunes)-remainder])
				newSurfaceTmp = string(surfaceRunes[len(surfaceRunes)-remainder:])
				surface = ""
				buffer.WriteString("<span id=\"surface\">" + remSurface + "</span>")
				booknumTmp = booknum
				chapterTmp = chapter
				verseTmp = verse
				// page = pageTmp
				// col = colTmp
				// line = lineTmp
				// buffer.WriteString("</br><span class=\"ms_ref\">" + page + ", " + strconv.Itoa(col) + "." + strconv.Itoa(line) + "</span>")
			}
		}
		if newSurface != "" {
			buffer.WriteString("<span id=\"surface\">" + newSurface + "</span>")
		}
		if newSurfaceTmp != "" {
			newSurface = newSurfaceTmp
		} else {
			newSurface = ""
		}
		if verseTmp != verse {
			booknum = booknumTmp
			chapter = chapterTmp
			verse = verseTmp
			buffer.WriteString("<span class=\"can_ref\"> " + strconv.Itoa(booknum) + " " + strconv.Itoa(chapter) + ":" + strconv.Itoa(verse) + "</span>")
		}
		if surface != "" {
			buffer.WriteString("<span id=\"surface\">" + surface + "</span>")
		}
	}
	rows.Close()
	buffer.WriteString("</div></div></div></body>")
	fmt.Fprint(w, buffer.String())
}

func dbCanonText(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	vars := mux.Vars(r)
	ms := vars["ms"]
	checkErr(db.Ping())
	rows, err := db.Query("CALL getCanonMS(?)", ms)
	checkErr(err)
	var booknum, chapter, verse int
	var buffer bytes.Buffer
	buffer.WriteString("<head><style media=\"screen\" type=\"text/css\">.verse {padding: 10px; display: inline-block;} .can_ref {display: inline-block;} .surface {display: inline-block;}</style></head><body><div><div class=\"verse\">")
	for rows.Next() {
		var lineTmp, remainder, colTmp, booknumTmp, chapterTmp, verseTmp int
		var pageTmp, surface, normalized string
		err = rows.Scan(&surface, &pageTmp, &colTmp, &lineTmp, &remainder, &booknumTmp, &chapterTmp, &verseTmp, &normalized)
		checkErr(err)
		if verseTmp != verse {
			if verse != 0 {
				buffer.WriteString("</div><br/><div class=\"verse\">")
			}
			booknum = booknumTmp
			chapter = chapterTmp
			verse = verseTmp
			buffer.WriteString("<span class=\"can_ref\"> " + strconv.Itoa(booknum) + " " + strconv.Itoa(chapter) + ":" + strconv.Itoa(verse) + "</span>")
		}
		if surface != "" {
			buffer.WriteString("<span class=\"surface\">" + surface + "</span>")
		}
	}
	rows.Close()
	buffer.WriteString("</div></div></body>")
	fmt.Fprint(w, buffer.String())
}

func dbSearchComp(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT DISTINCT composition.id AS id, composition.composition_name AS composition_name, composition_reference.composition_book_number AS composition_book_number FROM composition_reference JOIN composition ON composition.id = composition_reference.composition_id")
	fmt.Fprintln(w, data)
}

func dbGetCompBooks(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	vars := mux.Vars(r)
	composition := vars["composition"]
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT DISTINCT composition_book_number FROM composition_reference WHERE composition_id=?", composition)
	fmt.Fprintln(w, data)
}

func dbGetCompChapters(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	vars := mux.Vars(r)
	composition := vars["composition"]
	book := vars["book"]
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	ending := ""
	var params []interface{}
	params = append(params, composition)
	if book != "0" {
		ending = " AND composition_book_number = ?"
		params = append(params, book)
	}
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT DISTINCT composition_chapter AS ch FROM composition_reference WHERE composition_id = ?"+ending, params...)
	fmt.Fprintln(w, data)
}

func dbGetCompVerses(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	vars := mux.Vars(r)
	composition := vars["composition"]
	book := vars["book"]
	chapter := vars["chapter"]
	theCase := ""
	ending := ""
	var params []interface{}
	params = append(params, composition)
	params = append(params, chapter)
	if book != "0" {
		ending = " AND composition_book_number = ?"
		params = append(params, book)
	}
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT DISTINCT composition_verse AS v FROM composition_reference WHERE composition_id = ? AND composition_chapter = ?"+ending, params...)
	fmt.Fprintln(w, data)
}

func dbGetMssOfComp(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	vars := mux.Vars(r)
	composition := vars["composition"]
	book := vars["book"]
	chapter := vars["chapter"]
	verse := vars["verse"]
	theCase := ""
	ending := ""
	var params []interface{}
	params = append(params, composition)
	if book != "0" {
		ending += " and composition_reference.composition_book_number = ?"
		params = append(params, book)
	}
	if chapter != "0" {
		ending += " and composition_reference.composition_chapter = ?"
		params = append(params, chapter)
	}

	if verse != "0" {
		ending += " and composition_reference.composition_verse = ?"
		params = append(params, verse)
	}

	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "select distinct manuscript.id, manuscript.manuscript_name from manuscript join manuscript_position on manuscript_position.manuscript_id = manuscript.id join vocable_in_manuscript on vocable_in_manuscript.manuscript_pos_id = manuscript_position.id join composition_reference on composition_reference.id = vocable_in_manuscript.composition_ref_id where composition_reference.composition_id = ?"+ending, params...)
	fmt.Fprintln(w, data)
}

func dbGetMsModel(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
}

func dbGetManuscripts(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT id, manuscript_name AS m FROM manuscript")
	fmt.Fprintln(w, data)
}

func dbGetManuscriptPages(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	vars := mux.Vars(r)
	manuscript := vars["manuscript"]
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT DISTINCT page AS p FROM manuscript_reference WHERE manuscript_id = ?", manuscript)
	fmt.Fprintln(w, data)
}

func dbGetManuscriptColumns(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	vars := mux.Vars(r)
	manuscript := vars["manuscript"]
	page := vars["page"]
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT DISTINCT col AS c FROM manuscript_reference WHERE manuscript_id = ? AND page = ?", manuscript, page)
	fmt.Fprintln(w, data)
}

func dbGetManuscriptLines(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	vars := mux.Vars(r)
	manuscript := vars["manuscript"]
	page := vars["page"]
	column := vars["column"]
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT DISTINCT line AS l FROM manuscript_reference WHERE manuscript_id = ? AND page = ? AND col = ?", manuscript, page, column)
	fmt.Fprintln(w, data)
}

func dbCompositionMsText(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	vars := mux.Vars(r)
	msID, _ := strconv.Atoi(vars["ms_id"])
	msName := vars["ms_name"]

	msData := MsID{
		Name: msName,
		ID:   msID,
	}

	checkErr(db.Ping())
	var innerText = map[int]InnerText{}
	rows, err := db.Query("select vocable_in_manuscript.id as id, surface_vocable.surface_transcription as surface, COALESCE(normalized_vocable.normalized_vocable, '') as normalized from vocable_in_manuscript join manuscript_position on manuscript_position.id = vocable_in_manuscript.manuscript_pos_id join surface_vocable on surface_vocable.id = vocable_in_manuscript.surface_transcription_id left join normalized_vocable on normalized_vocable.surface_vocable_id = surface_vocable.id join composition_reference on composition_reference.id = vocable_in_manuscript.composition_ref_id where manuscript_position.manuscript_id = ? order by vocable_in_manuscript.composition_ref_id ,composition_reference.composition_book_number, composition_reference.composition_chapter, composition_reference.composition_verse, vocable_in_manuscript.manuscript_pos_id", msID)
	checkErr(err)
	defer rows.Close()
	for rows.Next() {
		var id int
		var surface, normalized string
		err = rows.Scan(&id, &surface, &normalized)
		checkErr(err)
		var entry = InnerText{
			Surface:    surface,
			Normalized: normalized,
		}
		innerText[id] = entry
	}

	var msPos = map[string]map[int]map[string][]SplitRef{}
	var savedRef SplitRef
	rows, err = db.Query("select vocable_in_manuscript.id, manuscript_reference.page, manuscript_reference.col, manuscript_reference.line, surface_vocable.surface_transcription from manuscript_position join manuscript_reference on manuscript_position.manuscript_reference_id = manuscript_reference.id join vocable_in_manuscript on vocable_in_manuscript.manuscript_pos_id = manuscript_position.id join surface_vocable on surface_vocable.id = vocable_in_manuscript.surface_transcription_id where manuscript_position.manuscript_id = ? order by manuscript_reference.page, manuscript_reference.col, manuscript_reference.line, manuscript_position.manuscript_reference_id", msID)
	checkErr(err)
	for rows.Next() {
		var textID, col int
		var line float64
		var page, surface string
		err = rows.Scan(&textID, &page, &col, &line, &surface)
		var lineStr = strconv.FormatFloat(line, 'f', -1, 64)
		checkErr(err)
		if (SplitRef{}) != savedRef {
			if len(msPos[page]) == 0 {
				msPos[page] = map[int]map[string][]SplitRef{}
			}
			if len(msPos[page][col]) == 0 {
				msPos[page][col] = map[string][]SplitRef{}
			}
			if len(msPos[page][col][lineStr]) == 0 {
				msPos[page][col][lineStr] = []SplitRef{}
			}
			msPos[page][col][lineStr] = append(msPos[page][col][lineStr], savedRef)
			savedRef = SplitRef{}
		}
		var surfaceLen = utf8.RuneCountInString(surface)
		var ref SplitRef
		if line == float64(int64(line)) {
			ref = SplitRef{
				ID:    textID,
				Start: 0,
				End:   surfaceLen,
			}
		} else {
			lineFloat := strings.Split(lineStr, ".")
			lineStr = lineFloat[0]
			cutOff, _ := strconv.Atoi(lineFloat[1])
			ref = SplitRef{
				ID:    textID,
				Start: 0,
				End:   cutOff,
			}
			savedRef = SplitRef{
				ID:    textID,
				Start: cutOff,
				End:   surfaceLen,
			}
		}
		if len(msPos[page]) == 0 {
			msPos[page] = map[int]map[string][]SplitRef{}
		}
		if len(msPos[page][col]) == 0 {
			msPos[page][col] = map[string][]SplitRef{}
		}
		if len(msPos[page][col][lineStr]) == 0 {
			msPos[page][col][lineStr] = []SplitRef{}
		}
		msPos[page][col][lineStr] = append(msPos[page][col][lineStr], ref)
	}

	var canPos = map[string]map[int]map[int]map[int][]int{}
	rows, err = db.Query("select vocable_in_manuscript.id as id, composition.composition_name as composition, composition_reference.composition_book_number as book, composition_reference.composition_chapter as chapter, composition_reference.composition_verse as verse from vocable_in_manuscript join composition_reference on composition_reference.id = vocable_in_manuscript.composition_ref_id join composition on composition.id = composition_reference.composition_id join manuscript_position on manuscript_position.id = vocable_in_manuscript.manuscript_pos_id where manuscript_position.manuscript_id = ? order by composition_reference.composition_book_number, composition_reference.composition_chapter, composition_reference.composition_verse, vocable_in_manuscript.manuscript_pos_id", msID)
	checkErr(err)
	for rows.Next() {
		var textID, book, chapter, verse int
		var composition string
		err = rows.Scan(&textID, &composition, &book, &chapter, &verse)
		checkErr(err)
		if len(canPos[composition]) == 0 {
			canPos[composition] = map[int]map[int]map[int][]int{}
		}
		if len(canPos[composition][book]) == 0 {
			canPos[composition][book] = map[int]map[int][]int{}
		}
		if len(canPos[composition][book][chapter]) == 0 {
			canPos[composition][book][chapter] = map[int][]int{}
		}
		if len(canPos[composition][book][chapter][verse]) == 0 {
			canPos[composition][book][chapter][verse] = []int{}
		}
		canPos[composition][book][chapter][verse] = append(canPos[composition][book][chapter][verse], textID)
	}

	msModel := MsModel{
		MsID:      msData,
		InnerText: innerText,
		MsPos:     msPos,
		CanPos:    canPos,
	}

	// jsonData, err := json.Marshal(msModel)
	// checkErr(err)
	// fmt.Fprintln(w, string(jsonData))
	sendJSON(w, msModel)
}

func dbSearchMS(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT * FROM manuscript")
	fmt.Fprintln(w, data)
}

func dbSearchBookNum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	composition := vars["composition"]
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "CALL getBookNums(?)", composition)
	fmt.Fprintln(w, data)
}

func dbSearchChapter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	composition := vars["composition"]
	bookNum := vars["booknum"]
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "CALL getChapters(?,?)", composition, bookNum)
	fmt.Fprintln(w, data)
}

func dbSearchVerse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	composition := vars["composition"]
	bookNum := vars["booknum"]
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "CALL getChapters(?,?)", composition, bookNum)
	fmt.Fprintln(w, data)
}

func dbSearchVocable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	composition := vars["composition"]
	bookNum := vars["booknum"]
	chapter := vars["chapter"]
	verse := vars["verse"]
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "CALL getVerseWords(?,?,?,?)", composition, bookNum, chapter, verse)
	fmt.Fprintln(w, data)
}

func dbVocableLinks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	composition := vars["composition"]
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "CALL getVocableLinks(?)", composition)
	fmt.Fprintln(w, data)
}

func htmlFromDB(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	vars := mux.Vars(r)
	composition := vars["composition"]
	ms1 := vars["ms1"]
	primaryMsID, err := strconv.Atoi(ms1)
	checkErr(err)
	ms2 := vars["ms2"]
	checkErr(db.Ping())
	rows, err := db.Query("CALL getFullTextComp(?,?,?)", composition, ms1, ms2)
	checkErr(err)
	var booknum, chapter, verse int
	var ms string
	var buffer bytes.Buffer
	var wordNum = 100
	for rows.Next() {
		var booknumTmp, chapterTmp, verseTmp, vocID, msIDTmp int
		var msTmp, vocSurface string
		err = rows.Scan(&booknumTmp, &chapterTmp, &verseTmp, &msTmp, &vocID, &vocSurface, &msIDTmp)
		checkErr(err)
		if ms != msTmp {
			if ms != "" {
				buffer.WriteString("</span>\n<br/>\n")
			}
			if verse != verseTmp {
				if verse != 0 {
					buffer.WriteString("</div>\n")
				}
				if chapter != chapterTmp {
					if chapter != 0 {
						buffer.WriteString("</div>\n")
					}
					if booknum != booknumTmp {
						if booknum != 0 {
							buffer.WriteString("</div>\n")
						}
						booknum = booknumTmp
						chapter = chapterTmp
						verse = verseTmp
						ms = msTmp
						buffer.WriteString("<div id=\"book-")
						buffer.WriteString(strconv.Itoa(booknum))
						buffer.WriteString("\" class=\"book_div\">\n")
					}
					chapter = chapterTmp
					verse = verseTmp
					ms = msTmp
					buffer.WriteString("<div id=\"chapter-")
					buffer.WriteString(strconv.Itoa(booknum))
					buffer.WriteString("-")
					buffer.WriteString(strconv.Itoa(chapter))
					buffer.WriteString("\" class=\"chapter_div\">\n")
				}
				verse = verseTmp
				ms = msTmp
				buffer.WriteString("<div id=\"verse-")
				buffer.WriteString(strconv.Itoa(booknum))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(chapter))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(verse))
				buffer.WriteString("\" class=\"chapter_div\">\n<span id=\"verse-")
				buffer.WriteString(strconv.Itoa(booknum))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(chapter))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(verse))
				buffer.WriteString("-heading\" class=\"verse_heading ignore\">")
				buffer.WriteString(strconv.Itoa(booknum))
				buffer.WriteString(" ")
				buffer.WriteString(composition)
				buffer.WriteString(" ")
				buffer.WriteString(strconv.Itoa(chapter))
				buffer.WriteString(":")
				buffer.WriteString(strconv.Itoa(verse))
				buffer.WriteString("</span>\n<br/>\n")
			}
			ms = msTmp
			buffer.WriteString("<span class=\"word_info ignore\">")
			buffer.WriteString(ms)
			buffer.WriteString("</span>\n<span class=\"ms_verse_span\" dir=\"auto\">\n")
		} else if verse != verseTmp {
			if verse != 0 {
				buffer.WriteString("</span>\n<br/>\n")
				buffer.WriteString("</div>\n")
			}
			if chapter != chapterTmp {
				if chapter != 0 {
					buffer.WriteString("</div>\n")
				}
				if booknum != booknumTmp {
					if booknum != 0 {
						buffer.WriteString("</div>\n")
					}
					booknum = booknumTmp
					chapter = chapterTmp
					verse = verseTmp
					ms = msTmp
					buffer.WriteString("<div id=\"book-")
					buffer.WriteString(strconv.Itoa(booknum))
					buffer.WriteString("\" class=\"book_div\">\n")
				}
				chapter = chapterTmp
				verse = verseTmp
				ms = msTmp
				buffer.WriteString("<div id=\"chapter-")
				buffer.WriteString(strconv.Itoa(booknum))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(chapter))
				buffer.WriteString("\" class=\"chapter_div\">\n")
			}
			verse = verseTmp
			ms = msTmp
			buffer.WriteString("<div id=\"verse-")
			buffer.WriteString(strconv.Itoa(booknum))
			buffer.WriteString("-")
			buffer.WriteString(strconv.Itoa(chapter))
			buffer.WriteString("-")
			buffer.WriteString(strconv.Itoa(verse))
			buffer.WriteString("\" class=\"chapter_div\">\n<span id=\"verse-")
			buffer.WriteString(strconv.Itoa(booknum))
			buffer.WriteString("-")
			buffer.WriteString(strconv.Itoa(chapter))
			buffer.WriteString("-")
			buffer.WriteString(strconv.Itoa(verse))
			buffer.WriteString("-heading\" class=\"verse_heading ignore\">")
			buffer.WriteString(strconv.Itoa(booknum))
			buffer.WriteString(" ")
			buffer.WriteString(composition)
			buffer.WriteString(" ")
			buffer.WriteString(strconv.Itoa(chapter))
			buffer.WriteString(":")
			buffer.WriteString(strconv.Itoa(verse))
			buffer.WriteString("</span>\n<br/>\n")
			buffer.WriteString("<span class=\"word_info ignore\">")
			buffer.WriteString(ms)
			buffer.WriteString("</span>\n<span class=\"ms_verse_span\" dir=\"auto\">\n")
		}
		buffer.WriteString("<div id=\"word-")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("-")
		buffer.WriteString(ms)
		buffer.WriteString("\" class=\"verse_word\">\n<div id=\"connection-")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("\" class=\"verse_word_upper_info\"></div>\n<div id=\"vocable-")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("-")
		buffer.WriteString(ms)
		if msIDTmp == primaryMsID {
			buffer.WriteString("\" class=\"verse_vocable\" tabindex=\"")
			buffer.WriteString(strconv.Itoa(wordNum))
			buffer.WriteString("\" onfocus=\"set_focused(this.id)\">")
			wordNum++
		} else {
			buffer.WriteString("\" class=\"verse_vocable\">")
		}

		buffer.WriteString(vocSurface)
		buffer.WriteString("</div>\n<div id=\"word-id")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("\" class=\"verse_word_lower_info\">")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("</div>\n</div>")
	}
	buffer.WriteString("</span>\n</div>\n</div>\n</div>")
	fmt.Fprint(w, buffer.String())
}

func verseVssFromDB(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	vars := mux.Vars(r)
	composition := vars["composition"]
	booknum, err := strconv.Atoi(vars["booknum"])
	checkErr(err)
	chapter, err := strconv.Atoi(vars["chapter"])
	checkErr(err)
	verse, err := strconv.Atoi(vars["verse"])
	checkErr(err)
	checkErr(db.Ping())
	rows, err := db.Query("CALL getVssOfVerse(?,?,?,?)", composition, booknum, chapter, verse)
	checkErr(err)
	var primaryMsID = 1
	var ms string
	var buffer bytes.Buffer
	var wordNum = 100
	for rows.Next() {
		booknumTmp := booknum
		chapterTmp := chapter
		verseTmp := verse
		var vocID, msIDTmp int
		var msTmp, vocSurface string
		err = rows.Scan(&msTmp, &vocID, &vocSurface, &msIDTmp)
		checkErr(err)
		if ms != msTmp {
			if ms != "" {
				buffer.WriteString("</span>\n<br/>\n")
			}
			if verse != verseTmp {
				if verse != 0 {
					buffer.WriteString("</div>\n")
				}
				if chapter != chapterTmp {
					if chapter != 0 {
						buffer.WriteString("</div>\n")
					}
					if booknum != booknumTmp {
						if booknum != 0 {
							buffer.WriteString("</div>\n")
						}
						booknum = booknumTmp
						chapter = chapterTmp
						verse = verseTmp
						ms = msTmp
						buffer.WriteString("<div id=\"book-")
						buffer.WriteString(strconv.Itoa(booknum))
						buffer.WriteString("\" class=\"book_div\">\n")
					}
					chapter = chapterTmp
					verse = verseTmp
					ms = msTmp
					buffer.WriteString("<div id=\"chapter-")
					buffer.WriteString(strconv.Itoa(booknum))
					buffer.WriteString("-")
					buffer.WriteString(strconv.Itoa(chapter))
					buffer.WriteString("\" class=\"chapter_div\">\n")
				}
				verse = verseTmp
				ms = msTmp
				buffer.WriteString("<div id=\"verse-")
				buffer.WriteString(strconv.Itoa(booknum))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(chapter))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(verse))
				buffer.WriteString("\" class=\"chapter_div\">\n<span id=\"verse-")
				buffer.WriteString(strconv.Itoa(booknum))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(chapter))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(verse))
				buffer.WriteString("-heading\" class=\"verse_heading ignore\">")
				buffer.WriteString(strconv.Itoa(booknum))
				buffer.WriteString(" ")
				buffer.WriteString(composition)
				buffer.WriteString(" ")
				buffer.WriteString(strconv.Itoa(chapter))
				buffer.WriteString(":")
				buffer.WriteString(strconv.Itoa(verse))
				buffer.WriteString("</span>\n<br/>\n")
			}
			ms = msTmp
			buffer.WriteString("<span class=\"word_info ignore\">")
			buffer.WriteString(ms)
			buffer.WriteString("</span>\n<span class=\"ms_verse_span\" dir=\"auto\">\n")
		} else if verse != verseTmp {
			if verse != 0 {
				buffer.WriteString("</span>\n<br/>\n")
				buffer.WriteString("</div>\n")
			}
			if chapter != chapterTmp {
				if chapter != 0 {
					buffer.WriteString("</div>\n")
				}
				if booknum != booknumTmp {
					if booknum != 0 {
						buffer.WriteString("</div>\n")
					}
					booknum = booknumTmp
					chapter = chapterTmp
					verse = verseTmp
					ms = msTmp
					buffer.WriteString("<div id=\"book-")
					buffer.WriteString(strconv.Itoa(booknum))
					buffer.WriteString("\" class=\"book_div\">\n")
				}
				chapter = chapterTmp
				verse = verseTmp
				ms = msTmp
				buffer.WriteString("<div id=\"chapter-")
				buffer.WriteString(strconv.Itoa(booknum))
				buffer.WriteString("-")
				buffer.WriteString(strconv.Itoa(chapter))
				buffer.WriteString("\" class=\"chapter_div\">\n")
			}
			verse = verseTmp
			ms = msTmp
			buffer.WriteString("<div id=\"verse-")
			buffer.WriteString(strconv.Itoa(booknum))
			buffer.WriteString("-")
			buffer.WriteString(strconv.Itoa(chapter))
			buffer.WriteString("-")
			buffer.WriteString(strconv.Itoa(verse))
			buffer.WriteString("\" class=\"chapter_div\">\n<span id=\"verse-")
			buffer.WriteString(strconv.Itoa(booknum))
			buffer.WriteString("-")
			buffer.WriteString(strconv.Itoa(chapter))
			buffer.WriteString("-")
			buffer.WriteString(strconv.Itoa(verse))
			buffer.WriteString("-heading\" class=\"verse_heading ignore\">")
			buffer.WriteString(strconv.Itoa(booknum))
			buffer.WriteString(" ")
			buffer.WriteString(composition)
			buffer.WriteString(" ")
			buffer.WriteString(strconv.Itoa(chapter))
			buffer.WriteString(":")
			buffer.WriteString(strconv.Itoa(verse))
			buffer.WriteString("</span>\n<br/>\n")
			buffer.WriteString("<span class=\"word_info ignore\">")
			buffer.WriteString(ms)
			buffer.WriteString("</span>\n<span class=\"ms_verse_span\" dir=\"auto\">\n")
		}
		buffer.WriteString("<div id=\"word-")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("-")
		buffer.WriteString(ms)
		buffer.WriteString("\" class=\"verse_word\">\n<div id=\"connection-")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("\" class=\"verse_word_upper_info\"></div>\n<div id=\"vocable-")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("-")
		buffer.WriteString(ms)
		if msIDTmp == primaryMsID {
			buffer.WriteString("\" class=\"verse_vocable\">")
			wordNum++
		} else {
			buffer.WriteString("\" class=\"verse_vocable\">")
		}

		buffer.WriteString(vocSurface)
		buffer.WriteString("</div>\n<div id=\"word-id")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("\" class=\"verse_word_lower_info\">")
		buffer.WriteString(strconv.Itoa(vocID))
		buffer.WriteString("</div>\n</div>")
	}
	buffer.WriteString("</span>\n</div>\n</div>\n</div>")
	fmt.Fprint(w, buffer.String())
}

func dbSynopticText(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	vars := mux.Vars(r)
	composition := vars["composition"]
	type verseRef struct {
		book  int
		chap  int
		verse int
	}
	var verseArr []verseRef
	checkErr(db.Ping())
	rows, err := db.Query("CALL getCompRefs(?)", composition)
	checkErr(err)
	for rows.Next() {
		var bookNumTmp, chapTmp, verseTmp int
		err = rows.Scan(&bookNumTmp, &chapTmp, &verseTmp)
		checkErr(err)
		var ref = verseRef{bookNumTmp, chapTmp, verseTmp}
		verseArr = append(verseArr, ref)
	}
	rows.Close()

	var buffer bytes.Buffer
	buffer.WriteString("<body>")
	for _, verse := range verseArr {
		buffer.WriteString("<div><span class=\"verse-listing\">" + strconv.Itoa(verse.book) + " " + strconv.Itoa(verse.chap) + ":" + strconv.Itoa(verse.verse) + "</span>")
		checkErr(db.Ping())
		rows2, err := db.Query("CALL getVssOfVerse(?,?,?,?)", composition, verse.book, verse.chap, verse.verse)
		checkErr(err)
		var ms string
		buffer.WriteString("<div class=\"verse\">")
		for rows2.Next() {
			var vocID, msIDTmp int
			var msTmp, vocSurface string
			err = rows2.Scan(&msTmp, &vocID, &vocSurface, &msIDTmp)
			checkErr(err)
			if ms != msTmp {
				if ms != "" {
					buffer.WriteString("</div>")
				}
				ms = msTmp
				buffer.WriteString("<div class=\"ms\"><span class=\"ms-designation\">" + ms + "</span>")
			}
			buffer.WriteString("<span class=\"word\">" + vocSurface + "</span>")
		}
		rows2.Close()
		buffer.WriteString("</div></div>")
	}

	buffer.WriteString("</body>")
	fmt.Fprint(w, buffer.String())
}

func linkVocables(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	voc1 := vars["voc1"]
	voc2 := vars["voc2"]
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, err := gosqljson.QueryDbToMapJSON(db, theCase, "CALL joinVocables(?,?)", voc1, voc2)
	if err != nil {
		fmt.Fprintln(w, "Failure")
	}
	fmt.Fprintln(w, data)
}

func unlinkVocables(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	voc1 := vars["voc1"]
	voc2 := vars["voc2"]
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, err := gosqljson.QueryDbToMapJSON(db, theCase, "CALL unlinkVocables(?,?)", voc1, voc2)
	if err != nil {
		fmt.Fprintln(w, "Failure")
	}
	fmt.Fprintln(w, data)
}
