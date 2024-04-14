package docs

const DocTemplate = `{
	"schemes": {{ marshal .Schemes }},
    "openapi": "3.0.0",
  	"info": {
		"description": "{{escape .Description}}",
    	"title": "{{.Title}}",
    	"contact": {},
        "version": "{{.Version}}"
  	},
	"host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
  	"paths": {
    "/register": {
      "post": {
        "description": "Запрос для регистрации в сервисе и получения токена (JWT), если передать header admin=true - будет создана запись админа, если нет - обычного пользователя",
        "tags": [
          "Authorization"
        ],
        "summary": "Запрос регистрации в сервисе",
        "parameters": [
          {
            "in": "header",
            "name": "admin",
            "description": "Флаг админа",
            "schema": {
              "type": "boolean",
              "example": true
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "login": {
                    "type": "string",
                    "description": "Логин пользователя/админа",
                    "example": "admin99"
                  },
                  "password": {
                    "type": "string",
                    "description": "Пароль",
                    "example": "simplepassword"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Регистрация прошла успешно, выдан токен на 1 день",
            "content": {
              "application/json": {
                "schema": {
                  "description": "JWT для использования в сервисе баннеров (действителен 1 день)",
                  "type": "object",
                  "additionalProperties": true,
                  "example": "{\"token\": \"jwt.for.auth\"}"
                }
              }
            }
          },
          "400": {
            "description": "Некорректные данные",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "409": {
            "description": "Пользователь с таким логином уже зарегистрирован в системе"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/acquire-token": {
      "post": {
        "description": "Запрос для получения токена (JWT) по логину и паролю",
        "tags": [
          "Authorization"
        ],
        "summary": "Запрос получения токена по логину и паролю",
        "parameters": [],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "login": {
                    "type": "string",
                    "description": "Логин пользователя/админа",
                    "example": "admin99"
                  },
                  "password": {
                    "type": "string",
                    "description": "Пароль",
                    "example": "simplepassword"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Успешный вход, выдан токен на 1 день",
            "content": {
              "application/json": {
                "schema": {
                  "description": "JWT для использования в сервисе баннеров",
                  "type": "object",
                  "additionalProperties": true,
                  "example": "{\"token\": \"jwt.for.auth\"}"
                }
              }
            }
          },
          "400": {
            "description": "Некорректные данные",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Неправильный пароль",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "404": {
            "description": "Пользователь с таким логином не найден"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/ping": {
      "get": {
        "description": "Просто пинг БД",
        "tags": [
          "Banners"
        ],
        "summary": "Пинг",
        "parameters": [
          {
            "in": "header",
            "name": "token",
            "description": "Токен админа",
            "schema": {
              "type": "string",
              "example": "admin_token"
            }
          }
        ],
        "responses": {
          "204": {
            "description": "Все в порядке"
          },
          "401": {
            "description": "Пользователь не авторизован"
          },
          "403": {
            "description": "Пользователь не имеет доступа (не админ)"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "text/plain": {
                "schema": {
                  "type": "string"
                }
              }
            }
          }
        }
      }
    },
    "/user_banner": {
      "get": {
        "description": "Запрос получения активного баннера пользователем (если токен админа - то выдаст даже неактивный) с применением кэширования, если не передан use_last_revision=true в query",
        "tags": [
          "Banners"
        ],
        "summary": "Получение баннера для пользователя",
        "parameters": [
          {
            "in": "query",
            "name": "tag_id",
            "required": true,
            "schema": {
              "type": "integer",
              "description": "Тэг пользователя"
            }
          },
          {
            "in": "query",
            "name": "feature_id",
            "required": true,
            "schema": {
              "type": "integer",
              "description": "Идентификатор фичи"
            }
          },
          {
            "in": "query",
            "name": "use_last_revision",
            "required": false,
            "schema": {
              "type": "boolean",
              "default": false,
              "description": "Получать актуальную информацию"
            }
          },
          {
            "in": "header",
            "name": "token",
            "description": "Токен пользователя/админа",
            "schema": {
              "type": "string",
              "example": "user_token"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Баннер пользователя",
            "content": {
              "application/json": {
                "schema": {
                  "description": "JSON-отображение баннера",
                  "type": "object",
                  "additionalProperties": true,
                  "example": "{\"title\": \"some_title\", \"text\": \"some_text\", \"url\": \"some_url\"}"
                }
              }
            }
          },
          "400": {
            "description": "Некорректные данные",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Пользователь не авторизован"
          },
          "403": {
            "description": "Пользователь не имеет доступа"
          },
          "404": {
            "description": "Баннер для данной пары tag-feature не найден"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/banner": {
      "get": {
        "description": "Запрос для получения всех баннеров c фильтрацией по фиче и/или тегу, если не указано ни то не другое - выдаются все баннеры, лимит по умолчанию 15, оффсет - 0, максимальный лимит - 100",
        "tags": [
          "Banners"
        ],
        "summary": "Получение всех баннеров c фильтрацией по фиче и/или тегу",
        "parameters": [
          {
            "in": "header",
            "name": "token",
            "description": "Токен админа",
            "schema": {
              "type": "string",
              "example": "admin_token"
            }
          },
          {
            "in": "query",
            "name": "feature_id",
            "required": false,
            "schema": {
              "type": "integer",
              "description": "Идентификатор фичи"
            }
          },
          {
            "in": "query",
            "name": "tag_id",
            "required": false,
            "schema": {
              "type": "integer",
              "description": "Идентификатор тега"
            }
          },
          {
            "in": "query",
            "name": "limit",
            "required": false,
            "schema": {
              "type": "integer",
              "description": "Лимит"
            }
          },
          {
            "in": "query",
            "name": "offset",
            "required": false,
            "schema": {
              "type": "integer",
              "description": "Оффсет"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "properties": {
                      "banner_id": {
                        "type": "integer",
                        "description": "Идентификатор баннера"
                      },
                      "tag_ids": {
                        "type": "array",
                        "description": "Идентификаторы тэгов",
                        "items": {
                          "type": "integer"
                        }
                      },
                      "feature_id": {
                        "type": "integer",
                        "description": "Идентификатор фичи"
                      },
                      "content": {
                        "type": "object",
                        "description": "Содержимое баннера",
                        "additionalProperties": true,
                        "example": "{\"title\": \"some_title\", \"text\": \"some_text\", \"url\": \"some_url\"}"
                      },
                      "is_active": {
                        "type": "boolean",
                        "description": "Флаг активности баннера"
                      },
                      "created_at": {
                        "type": "string",
                        "format": "date-time",
                        "description": "Дата создания баннера"
                      },
                      "updated_at": {
                        "type": "string",
                        "format": "date-time",
                        "description": "Дата обновления баннера"
                      }
                    }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Пользователь не авторизован"
          },
          "403": {
            "description": "Пользователь не имеет доступа (не админ)"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      },
      "post": {
        "description": "Создание нового баннера с получением ID",
        "tags": [
          "Banners"
        ],
        "summary": "Создание нового баннера",
        "parameters": [
          {
            "in": "header",
            "name": "token",
            "description": "Токен админа",
            "schema": {
              "type": "string",
              "example": "admin_token"
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "tag_ids": {
                    "type": "array",
                    "description": "Идентификаторы тэгов",
                    "items": {
                      "type": "integer"
                    }
                  },
                  "feature_id": {
                    "type": "integer",
                    "description": "Идентификатор фичи"
                  },
                  "content": {
                    "type": "object",
                    "description": "Содержимое баннера",
                    "additionalProperties": true,
                    "example": "{\"title\": \"some_title\", \"text\": \"some_text\", \"url\": \"some_url\"}"
                  },
                  "is_active": {
                    "type": "boolean",
                    "description": "Флаг активности баннера"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Created",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "banner_id": {
                      "type": "integer",
                      "description": "Идентификатор созданного баннера"
                    }
                  }
                }
              }
            }
          },
          "400": {
            "description": "Некорректные данные/нарушение уникальности пары тег-фича",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Пользователь не авторизован"
          },
          "403": {
            "description": "Пользователь не имеет доступа (не админ)"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      },
      "delete": {
        "description": "Запрос удаления баннеров (и всех их версий) по тегу, фиче или паре тег-фича, необходимо указать хотя бы что-то одно. Удаление производится в отдельной горутине (одновременно это может делать ограниченное число горутин (остальные ждут), можно задать в конфигурации)",
        "tags": [
          "Banners"
        ],
        "summary": "Удаление баннеров по тегу/фиче",
        "parameters": [
          {
            "in": "query",
            "name": "feature_id",
            "required": false,
            "schema": {
              "type": "integer",
              "description": "Идентификатор фичи"
            }
          },
          {
            "in": "query",
            "name": "tag_id",
            "required": false,
            "schema": {
              "type": "integer",
              "description": "Идентификатор тега"
            }
          },
          {
            "in": "header",
            "name": "token",
            "description": "Токен админа",
            "schema": {
              "type": "string",
              "example": "admin_token"
            }
          }
        ],
        "responses": {
          "202": {
            "description": "Запрос на удаление получен, удаление может выполниться не мнгновенно"
          },
          "400": {
            "description": "Некорректные данные",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Пользователь не авторизован"
          },
          "403": {
            "description": "Пользователь не имеет доступа (не админ)"
          },
          "404": {
            "description": "Ни одного баннера с указанными параметрами не найдено"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/banner/{id}": {
      "patch": {
        "description": "Частичное/полное обновление баннера (создает новую версию баннера и делает ее выбранной)",
        "tags": [
          "Banners"
        ],
        "summary": "Обновление содержимого баннера",
        "parameters": [
          {
            "in": "path",
            "name": "id",
            "required": true,
            "schema": {
              "type": "integer",
              "description": "Идентификатор баннера"
            }
          },
          {
            "in": "header",
            "name": "token",
            "description": "Токен админа",
            "schema": {
              "type": "string",
              "example": "admin_token"
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "tag_ids": {
                    "nullable": true,
                    "type": "array",
                    "description": "Идентификаторы тэгов",
                    "items": {
                      "type": "integer"
                    }
                  },
                  "feature_id": {
                    "nullable": true,
                    "type": "integer",
                    "description": "Идентификатор фичи"
                  },
                  "content": {
                    "nullable": true,
                    "type": "object",
                    "description": "Содержимое баннера",
                    "additionalProperties": true,
                    "example": "{\"title\": \"some_title\", \"text\": \"some_text\", \"url\": \"some_url\"}"
                  },
                  "is_active": {
                    "nullable": true,
                    "type": "boolean",
                    "description": "Флаг активности баннера"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Обновлено успешно"
          },
          "400": {
            "description": "Некорректные данные/запрос вызывает нарушение уникальности пары тег-фича",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Пользователь не авторизован"
          },
          "403": {
            "description": "Пользователь не имеет доступа (не админ)"
          },
          "404": {
            "description": "Баннер/версия не найден(-а)"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      },
      "delete": {
        "description": "Запрос удаления баннера (и всех его версий) по идентификатору",
        "tags": [
          "Banners"
        ],
        "summary": "Удаление баннера по идентификатору",
        "parameters": [
          {
            "in": "path",
            "name": "id",
            "required": true,
            "schema": {
              "type": "integer",
              "description": "Идентификатор баннера"
            }
          },
          {
            "in": "header",
            "name": "token",
            "description": "Токен админа",
            "schema": {
              "type": "string",
              "example": "admin_token"
            }
          }
        ],
        "responses": {
          "204": {
            "description": "Баннер успешно удален"
          },
          "400": {
            "description": "Некорректные данные",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Пользователь не авторизован"
          },
          "403": {
            "description": "Пользователь не имеет доступа (не админ)"
          },
          "404": {
            "description": "Баннер с таким ID не найден"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/banner_versions/{id}": {
        "get": {
        "description": "Запрос для получения всех версий баннера по его id, опционально с лимитом (по умолчанию - 3, максимум - 100) и оффсетом (по умолчанию - 0), отсортированные по невозрастанию времени обновления",
        "tags": [
          "Versions"
        ],
        "summary": "Получение всех версий баннера c лимитом и оффсетом",
        "parameters": [
          {
              "in": "path",
              "name": "id",
              "required": true,
              "schema": {
                "type": "integer",
                "description": "Идентификатор баннера"
              }
          },
          {
            "in": "header",
            "name": "token",
            "description": "Токен админа",
            "schema": {
              "type": "string",
              "example": "admin_token"
            }
          },
          {
            "in": "query",
            "name": "limit",
            "required": false,
            "schema": {
              "type": "integer",
              "description": "Лимит"
            }
          },
          {
            "in": "query",
            "name": "offset",
            "required": false,
            "schema": {
              "type": "integer",
              "description": "Оффсет"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "properties": {
                      "version_id": {
                        "type": "integer",
                        "description": "Идентификатор версии (уникальный по отношению ко всем существующим версиям всех баннеров)"
                      },
                      "tag_ids": {
                        "type": "array",
                        "description": "Идентификаторы тегов этой версии",
                        "items": {
                          "type": "integer"
                        }
                      },
                      "feature_id": {
                        "type": "integer",
                        "description": "Идентификатор фичи этой версии"
                      },
                      "content": {
                        "type": "object",
                        "description": "Содержимое баннера этой версии",
                        "additionalProperties": true,
                        "example": "{\"title\": \"some_title\", \"text\": \"some_text\", \"url\": \"some_url\"}"
                      },
                      "is_active": {
                        "type": "boolean",
                        "description": "Флаг активности баннера этой версии"
                      },
                      "created_at": {
                        "type": "string",
                        "format": "date-time",
                        "description": "Дата создания баннера"
                      },
                      "updated_at": {
                        "type": "string",
                        "format": "date-time",
                        "description": "Дата обновления баннера (создания этой версии)"
                      },
                      "is_chosen": {
                        "type": "boolean",
                        "description": "Флаг выбранной версии"
                      }
                    }
                  }
                }
              }
            }
          },
          "400": {
            "description": "Некорректные данные",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          },
          "404": {
            "description": "Баннер с указанным ID не найден"
          },
          "401": {
            "description": "Пользователь не авторизован"
          },
          "403": {
            "description": "Пользователь не имеет доступа (не админ)"
          },
          "500": {
            "description": "Внутренняя ошибка сервера",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": {
                      "type": "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/banner_versions/choose/{id}": {
      "patch": {
      "description": "Запрос для выбора версии баннера по его ID и ID версии (ID версии можно посмотреть при запросе списка версий)",
      "tags": [
        "Versions"
      ],
      "summary": "Выбор версии баннера",
      "parameters": [
        {
            "in": "path",
            "name": "id",
            "required": true,
            "schema": {
              "type": "integer",
              "description": "Идентификатор баннера"
            }
        },
        {
          "in": "header",
          "name": "token",
          "description": "Токен админа",
          "schema": {
            "type": "string",
            "example": "admin_token"
          }
        },
        {
          "in": "query",
          "name": "version_id",
          "required": true,
          "schema": {
            "type": "integer",
            "description": "ID версии"
          }
        },
      ],
      "responses": {
        "204": {
          "description": "Версия успешно задана"
        },
        "400": {
          "description": "Некорректные данные",
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "error": {
                    "type": "string"
                  }
                }
              }
            }
          }
        },
        "409": {
          "description": "Выполнение запроса нарушило бы требование уникальности пары тег-фича"
        },
        "404": {
          "description": "Баннер/версия с указанным ID не найден"
        },
        "401": {
          "description": "Пользователь не авторизован"
        },
        "403": {
          "description": "Пользователь не имеет доступа (не админ)"
        },
        "500": {
          "description": "Внутренняя ошибка сервера",
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "error": {
                    "type": "string"
                  }
                }
              }
            }
          }
        }
      }
    }
  }
  }
}`
