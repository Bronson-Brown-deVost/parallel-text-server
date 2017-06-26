package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Bronson-Brown-deVost/gosqljson"
	"github.com/gorilla/mux"
)

func writeHeader(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

func dbSearchText(w http.ResponseWriter, r *http.Request) {
}

func dbSearchComp(w http.ResponseWriter, r *http.Request) {
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "SELECT id, composition_name FROM composition")
	fmt.Fprintln(w, data)
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
	chapter := vars["chapter"]
	writeHeader(w)
	if err := db.Ping(); err != nil {
		checkErr(err)
	}
	theCase := ""
	data, _ := gosqljson.QueryDbToMapJSON(db, theCase, "CALL getVerses(?,?,?)", composition, bookNum, chapter)
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
