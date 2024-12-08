// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/songs/add": {
            "post": {
                "description": "Adds a new song with details retrieved from an external API.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "songs"
                ],
                "summary": "Add a new song",
                "parameters": [
                    {
                        "description": "Song data to add",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/types.SongAddPayload"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Song added successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid input",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to add song",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/songs/delete": {
            "delete": {
                "description": "Deletes a song based on its name and group.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "songs"
                ],
                "summary": "Delete a song",
                "parameters": [
                    {
                        "description": "delete the song based on ID",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/types.SongDeletePayload"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Song deleted successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid input",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to delete song",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/songs/get": {
            "get": {
                "description": "Retrieves songs matching specified criteria through query parameters.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "songs"
                ],
                "summary": "Retrieve songs",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID of the song",
                        "name": "id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Name of the song",
                        "name": "song",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Group name",
                        "name": "group",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Link to the song",
                        "name": "link",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Time (RFC3339 format)",
                        "name": "time",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Lyrics as a JSON array",
                        "name": "lyrics",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Maximum number of results to return",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Offset for pagination",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Songs retrieved successfully",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/types.Song"
                            }
                        }
                    },
                    "400": {
                        "description": "Invalid query parameter",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "No songs found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to fetch songs",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/songs/update": {
            "put": {
                "description": "Updates existing song details.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "songs"
                ],
                "summary": "Update song",
                "parameters": [
                    {
                        "description": "update the song",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/types.Song"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Song updated successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid input",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Failed to update song",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "types.Song": {
            "type": "object",
            "properties": {
                "group": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "link": {
                    "type": "string"
                },
                "published": {
                    "type": "string"
                },
                "song": {
                    "type": "string"
                },
                "songLyrics": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "types.SongAddPayload": {
            "type": "object",
            "properties": {
                "group": {
                    "type": "string"
                },
                "song": {
                    "type": "string"
                }
            }
        },
        "types.SongDeletePayload": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/",
	Schemes:          []string{},
	Title:            "Muse_Library App API",
	Description:      "API Server for Music Library App",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
