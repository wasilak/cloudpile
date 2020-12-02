package main

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "index.html",
		FileModTime: time.Unix(1606398790, 0),

		Content: string("<!doctype html>\n<html lang=\"en\">\n<head>\n    <!-- Required meta tags -->\n    <meta charset=\"utf-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1, shrink-to-fit=no\">\n\n    <!-- Bootstrap CSS -->\n    <link rel=\"stylesheet\" href=\"https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css\" integrity=\"sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm\" crossorigin=\"anonymous\">\n</head>\n<body>\n<div class=\"container-fluid\">\n            <div class=\"row\">\n                <div class=\"col-sm\">\n                    <form id=\"searchForm\" action=\"/\">\n                        <div class=\"form-group\">\n                            <label for=\"id\">ID(s)</label>\n                            <input type=\"text\" class=\"form-control\" id=\"id\" name=\"id\" value=\"{{.IDs}}\" data-role=\"tagsinput\">\n                        </div>\n                        <button type=\"submit\" class=\"btn btn-primary\">Submit</button>\n                    </form>\n                    <hr />\n                </div>\n            </div>\n            <div class=\"row\">\n                <div class=\"col-sm\">\n                    <table class=\"table table-striped table-bordered table-dark table-hover\">\n                    <thead>\n                        <tr>\n                            <th scope=\"col\">#</th>\n                            <th scope=\"col\">ID</th>\n                            <th scope=\"col\">Type</th>\n                            <th scope=\"col\">Account</th>\n                            <th scope=\"col\">Tags</th>\n                        </tr>\n                    </thead>\n                    <tbody>\n                        {{range $index, $item := .Items}}\n                            \n                                <tr>\n                                    <th scope=\"row\">{{$index}}</th>\n                                    <td>{{$item.ID}}</td>\n                                    <td>{{$item.Type}}</td>\n                                    <td>{{if $item.AccountAlias}}{{$item.AccountAlias}} - {{end}}{{$item.Account}}</td>\n                                    <td>\n                                        <ul>\n                                        {{range $tag := $item.Tags}}\n                                            <li>{{$tag.Key}} = {{$tag.Value}}</li>\n                                        {{end}}\n                                        </ul>\n                                    </td>\n                                </tr>\n                        {{ end }}\n                    </tbody>\n                    </table>\n                </div>\n            </div>\n            {{/* <div class=\"row\">\n                <div class=\"col-sm\">\n                    <pre>\n                        <code>\n                        {{.Reservations}}\n                        </code>\n                    </pre>\n                </div>\n            </div> */}}\n        </div>\n    \n    <!-- Optional JavaScript -->\n    <!-- jQuery first, then Popper.js, then Bootstrap JS -->\n    <script src=\"https://code.jquery.com/jquery-3.5.1.min.js\" crossorigin=\"anonymous\"></script>\n    <script src=\"https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js\" integrity=\"sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q\" crossorigin=\"anonymous\"></script>\n    <script src=\"https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js\" integrity=\"sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl\" crossorigin=\"anonymous\"></script>\n\n</body>\n</html>\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1606217657, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "index.html"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`views`, &embedded.EmbeddedBox{
		Name: `views`,
		Time: time.Unix(1606217657, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"index.html": file2,
		},
	})
}
