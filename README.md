# PescaApp Backend

Backend RESTful para aplicación de pesca colaborativa, implementado como **monolito modular** en Go con Gin, Firebase/Firestore y arquitectura DDD.

## Arquitectura

```
cmd/api/              → Punto de entrada (main.go)
internal/
├── pescaapp/         → Dominio PescaApp (bounded contexts)
│   ├── auth/         → Autenticación (Google OAuth2 + JWT)
│   ├── users/        → Gestión de usuarios y favoritos
│   ├── species/      → Catálogo de especies de peces
│   ├── spots/        → Lugares de pesca (CRUD + geolocalización)
│   ├── comments/     → Comentarios y likes
│   ├── ratings/      → Valoraciones (1-5 estrellas)
│   ├── search/       → Búsqueda global unificada
│   └── statistics/   → Estadísticas y analítica
├── platform/         → Infraestructura transversal
│   ├── cache/        → Cache en memoria (go-cache)
│   ├── config/       → Configuración desde env vars
│   ├── firebase/     → Cliente Firebase (Firestore + Auth)
│   ├── logger/       → Logging estructurado (logrus)
│   └── middleware/    → Auth JWT, CORS, Rate Limiting, Recovery
└── shared/           → Paquetes compartidos
    ├── errors/       → Errores de aplicación estandarizados
    ├── pagination/   → Parser de paginación
    └── response/     → Helpers de respuesta HTTP
```

Cada módulo de dominio sigue la estructura DDD:
- **domain/** → Entidades, value objects y puertos (interfaces)
- **application/** → Casos de uso / servicios de aplicación
- **infrastructure/** → Adaptadores (repositorios Firestore)
- **interfaces/** → Handlers HTTP y rutas Gin

### Desacoplamiento para microservicios

Los módulos se comunican entre sí únicamente a través de **interfaces** (puertos), no de implementaciones concretas. Esto permite extraer cualquier módulo a un microservicio independiente reemplazando las interfaces por llamadas HTTP/gRPC.

## Tecnologías

| Componente | Tecnología |
|---|---|
| Lenguaje | Go 1.22+ |
| Framework HTTP | [Gin](https://github.com/gin-gonic/gin) |
| Base de datos | Firebase Firestore |
| Autenticación | Firebase Auth + JWT propio |
| Cache | [go-cache](https://github.com/patrickmn/go-cache) (in-memory) |
| Logging | [logrus](https://github.com/sirupsen/logrus) |
| Documentación API | OpenAPI 3.0 (`openapi.yaml`) |

## Requisitos

- Go 1.22+
- Proyecto Firebase con Firestore habilitado
- Credenciales de service account de Firebase (JSON)

## Configuración

Copia `.env.example` y configura las variables:

```bash
cp .env.example .env
```

| Variable | Descripción | Default |
|---|---|---|
| `PORT` | Puerto del servidor | `8080` |
| `ENVIRONMENT` | `development` / `production` | `development` |
| `FIREBASE_PROJECT_ID` | ID del proyecto Firebase | — |
| `FIREBASE_CREDENTIALS_JSON` | JSON completo del service account | — |
| `JWT_SECRET` | Secret para firmar tokens JWT | — |
| `JWT_EXPIRY_MINUTES` | Expiración del access token (min) | `60` |
| `RATE_LIMIT_PER_MIN` | Requests por minuto por IP | `100` |
| `ALLOWED_ORIGINS` | Orígenes CORS (separados por coma) | `*` |
| `LOG_LEVEL` | Nivel de log (`debug`, `info`, `warn`, `error`) | `info` |

## Ejecución

```bash
# Desarrollo
go run cmd/api/main.go

# Build
go build -o bin/pescaapp cmd/api/main.go
./bin/pescaapp

# Tests
go test ./internal/... -v

# Tests con cobertura
go test ./internal/... -cover

# Cobertura detallada (HTML)
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Endpoints

### Auth
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| POST | `/v1/auth/login` | Login con Google | No |
| POST | `/v1/auth/logout` | Cerrar sesión | Sí |

### Users
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| GET | `/v1/users` | Listar usuarios | Sí |
| GET | `/v1/users/:id` | Obtener usuario | Sí |
| PUT | `/v1/users/:id` | Actualizar perfil | Sí |

### Favorites
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| GET | `/v1/users/:id/favorites` | Listar favoritos | Sí |
| POST | `/v1/users/:id/favorites/:spotId` | Agregar favorito | Sí |
| DELETE | `/v1/users/:id/favorites/:spotId` | Eliminar favorito | Sí |

### Species
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| GET | `/v1/species` | Listar especies | No |
| GET | `/v1/species/:id` | Detalle especie | No |
| POST | `/v1/species` | Crear especie | Admin |
| PUT | `/v1/species/:id` | Actualizar especie | Admin |
| DELETE | `/v1/species/:id` | Eliminar especie | Admin |

### Spots
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| GET | `/v1/spots` | Listar spots (con filtros) | No |
| GET | `/v1/spots/:id` | Detalle spot | No |
| POST | `/v1/spots` | Crear spot | Sí |
| PUT | `/v1/spots/:id` | Actualizar spot | Creador/Admin |
| DELETE | `/v1/spots/:id` | Eliminar spot | Creador/Admin |
| GET | `/v1/spots/:id/nearby` | Spots cercanos | No |

### Comments
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| GET | `/v1/spots/:id/comments` | Listar comentarios | No |
| POST | `/v1/spots/:id/comments` | Crear comentario | Sí |
| PUT | `/v1/comments/:commentId` | Editar comentario | Autor |
| DELETE | `/v1/comments/:commentId` | Eliminar comentario | Autor/Admin |
| POST | `/v1/comments/:commentId/like` | Like | Sí |
| DELETE | `/v1/comments/:commentId/like` | Unlike | Sí |

### Ratings
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| GET | `/v1/spots/:id/ratings` | Listar valoraciones | No |
| POST | `/v1/spots/:id/ratings` | Crear/actualizar valoración | Sí |
| DELETE | `/v1/spots/:id/ratings` | Eliminar valoración | Sí |

### Search
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| GET | `/v1/search?q=&type=all` | Búsqueda global | No |

### Statistics
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| GET | `/v1/spots/:id/stats` | Stats de spot | No |
| GET | `/v1/user/:userId/stats` | Stats de usuario | No |
| GET | `/v1/statistics/popular-spots` | Spots populares | No |

### Health
| Método | Ruta | Descripción |
|---|---|---|
| GET | `/health` | Health check |

## Base de datos

Se utiliza **Firebase Firestore** (base de datos orientada a documentos). Las colecciones necesarias están documentadas en [`FIRESTORE_COLLECTIONS.md`](FIRESTORE_COLLECTIONS.md).

## Documentación API

La especificación completa está disponible en formato OpenAPI 3.0:
- [`openapi.yaml`](openapi.yaml) — Importar en Swagger UI, Postman o similar
- [`BACKEND_SPECIFICATION.md`](BACKEND_SPECIFICATION.md) — Especificación detallada en Markdown

## Estructura del proyecto

```
.
├── cmd/api/main.go                    # Punto de entrada
├── internal/
│   ├── pescaapp/                      # Módulos de dominio
│   │   ├── auth/                      # 🔐 Autenticación
│   │   ├── users/                     # 👤 Usuarios + Favoritos
│   │   ├── species/                   # 🐟 Especies
│   │   ├── spots/                     # 📍 Lugares de pesca
│   │   ├── comments/                  # 💬 Comentarios + Likes
│   │   ├── ratings/                   # ⭐ Valoraciones
│   │   ├── search/                    # 🔍 Búsqueda global
│   │   └── statistics/                # 📊 Estadísticas
│   ├── platform/                      # Infraestructura
│   └── shared/                        # Paquetes compartidos
├── openapi.yaml                       # Especificación OpenAPI 3.0
├── FIRESTORE_COLLECTIONS.md           # Esquema de colecciones
├── BACKEND_SPECIFICATION.md           # Especificación funcional
└── go.mod
```

## Licencia

Proyecto privado.

