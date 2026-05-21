# PescaApp - Especificación de Servicios Backend

## 1. Descripción General

Backend RESTful para aplicación de pesca colaborativa. Gestiona usuarios, lugares de pesca, especies, comentarios y valoraciones. Implementar con Node.js/Express, Python/FastAPI o similar.

**Base URL:** `https://api.pescaapp.com/v1`

---

## 2. Autenticación

### 2.1 Google Sign-In Integration
- Token JWT desde Google OAuth2
- Header: `Authorization: Bearer <google_id_token>`
- Validar token y crear/actualizar usuario automáticamente

### 2.2 Endpoints de Autenticación

#### POST `/auth/login`
Autentica usuario con Google

**Request:**
```json
{
  "idToken": "string",
  "email": "string"
}
```

**Response:** `201 Created`
```json
{
  "user": {
    "id": "string",
    "name": "string",
    "email": "string",
    "photoUrl": "string|null",
    "createdAt": "ISO8601"
  },
  "accessToken": "JWT",
  "refreshToken": "string"
}
```

#### POST `/auth/logout`
Invalida sesión

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "message": "Logged out successfully"
}
```

---

## 3. Usuarios (Users)

### 3.1 GET `/users/:id`
Obtiene perfil de usuario

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "id": "user_1",
  "name": "Luis Rivera",
  "email": "luis66riv@gmail.com",
  "photoUrl": "https://...",
  "createdAt": "2024-03-10T00:00:00Z",
  "stats": {
    "spotsCreated": 5,
    "commentsCount": 12,
    "ratingsCount": 8,
    "averageRating": 4.5
  }
}
```

### 3.2 PUT `/users/:id`
Actualiza perfil de usuario

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "name": "string",
  "photoUrl": "string|null"
}
```

**Response:** `200 OK` (usuario actualizado)

### 3.3 GET `/users`
Lista usuarios (paginado, solo info pública)

**Query Parameters:**
- `limit=20`
- `offset=0`
- `search=string` (buscar por nombre)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "user_1",
      "name": "Luis Rivera",
      "photoUrl": "string|null",
      "stats": {
        "spotsCreated": 5
      }
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0
  }
}
```

---

## 4. Especies (Species)

### 4.1 GET `/species`
Lista todas las especies de peces

**Query Parameters:**
- `limit=50`
- `offset=0`
- `search=string` (buscar por nombre común o científico)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "sp_1",
      "commonName": "Trucha Arcoiris",
      "scientificName": "Oncorhynchus mykiss",
      "description": "string",
      "imageUrl": "https://...",
      "habitat": "Ríos y lagos de agua fría",
      "diet": "Insectos, crustáceos, peces pequeños",
      "averageSizeCm": 35,
      "averageWeightKg": 0.8,
      "maxSizeCm": 90,
      "maxWeightKg": 9.0,
      "fishingTips": ["tip1", "tip2", "tip3"]
    }
  ],
  "pagination": {
    "total": 50,
    "limit": 50,
    "offset": 0
  }
}
```

### 4.2 GET `/species/:id`
Obtiene detalles de una especie

**Response:** `200 OK` (mismo formato que arriba, un objeto)

### 4.3 POST `/species` ⭐ ADMIN
Crea nueva especie

**Headers:** `Authorization: Bearer <token>` (requiere rol admin)

**Request:**
```json
{
  "commonName": "string",
  "scientificName": "string",
  "description": "string",
  "imageUrl": "string|null",
  "habitat": "string",
  "diet": "string",
  "averageSizeCm": "number",
  "averageWeightKg": "number",
  "maxSizeCm": "number",
  "maxWeightKg": "number",
  "fishingTips": ["string"]
}
```

**Response:** `201 Created`

### 4.4 PUT `/species/:id` ⭐ ADMIN
Actualiza especie

**Response:** `200 OK`

### 4.5 DELETE `/species/:id` ⭐ ADMIN
Elimina especie

**Response:** `204 No Content`

---

## 5. Lugares de Pesca (Fishing Spots)

### 5.1 GET `/spots`
Lista lugares de pesca (con filtros)

**Query Parameters:**
- `limit=20`
- `offset=0`
- `region=string` (filtrar por región)
- `waterType=river|lake|sea|lagoon`
- `boatRequired=true|false`
- `latitude=number`
- `longitude=number`
- `radius=number` (en km, para búsqueda por proximidad)
- `sortBy=rating|recent|distance` (ordenamiento)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "spot_1",
      "name": "Río Maipo - Puente Negro",
      "description": "string",
      "latitude": -33.8892,
      "longitude": -70.5523,
      "region": "Región Metropolitana",
      "waterType": "river",
      "boatAllowed": false,
      "boatRequired": false,
      "access": "easy|moderate|difficult|boatOnly",
      "createdByUserId": "user_1",
      "createdAt": "ISO8601",
      "averageRating": 4.3,
      "totalRatings": 12,
      "totalComments": 3,
      "species": [
        {
          "speciesId": "sp_1",
          "recommendedBaits": ["string"],
          "recommendedRod": "string",
          "recommendedLine": "string",
          "recommendedHook": "string",
          "bestSeason": "string",
          "difficulty": "easy|medium|hard",
          "notes": "string"
        }
      ]
    }
  ],
  "pagination": {
    "total": 50,
    "limit": 20,
    "offset": 0
  }
}
```

### 5.2 GET `/spots/:id`
Obtiene detalles completos de un lugar

**Response:** `200 OK` (mismo formato que arriba)

### 5.3 POST `/spots`
Crea nuevo lugar de pesca

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "name": "string (required)",
  "description": "string (required)",
  "latitude": "number (required)",
  "longitude": "number (required)",
  "region": "string (required)",
  "waterType": "river|lake|sea|lagoon",
  "boatAllowed": "boolean",
  "boatRequired": "boolean",
  "access": "easy|moderate|difficult|boatOnly",
  "species": [
    {
      "speciesId": "string",
      "recommendedBaits": ["string"],
      "recommendedRod": "string",
      "recommendedLine": "string",
      "recommendedHook": "string",
      "bestSeason": "string",
      "difficulty": "easy|medium|hard",
      "notes": "string"
    }
  ]
}
```

**Response:** `201 Created` (lugar creado)

### 5.4 PUT `/spots/:id`
Actualiza lugar de pesca

**Headers:** `Authorization: Bearer <token>` (solo creador o admin)

**Request:** (mismo campo que POST, todos opcionales)

**Response:** `200 OK`

### 5.5 DELETE `/spots/:id`
Elimina lugar de pesca

**Headers:** `Authorization: Bearer <token>` (solo creador o admin)

**Response:** `204 No Content`

### 5.6 GET `/spots/:spotId/nearby`
Obtiene lugares cercanos

**Query Parameters:**
- `radiusKm=10`
- `limit=10`

**Response:** `200 OK` (array de spots)

---

## 6. Comentarios (Comments)

### 6.1 GET `/spots/:spotId/comments`
Lista comentarios de un lugar

**Query Parameters:**
- `limit=20`
- `offset=0`
- `sortBy=recent|oldest|helpful`

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "c_1_1",
      "spotId": "spot_1",
      "userId": "user_2",
      "user": {
        "id": "user_2",
        "name": "Pedro Soto",
        "photoUrl": "string|null"
      },
      "text": "Fui el domingo pasado...",
      "likes": 3,
      "liked": false,
      "createdAt": "ISO8601",
      "updatedAt": "ISO8601|null"
    }
  ],
  "pagination": {
    "total": 15,
    "limit": 20,
    "offset": 0
  }
}
```

### 6.2 POST `/spots/:spotId/comments`
Crea comentario en un lugar

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "text": "string (required, min 10 chars, max 500)"
}
```

**Response:** `201 Created`

### 6.3 PUT `/comments/:commentId`
Edita comentario

**Headers:** `Authorization: Bearer <token>` (solo autor)

**Request:**
```json
{
  "text": "string"
}
```

**Response:** `200 OK`

### 6.4 DELETE `/comments/:commentId`
Elimina comentario

**Headers:** `Authorization: Bearer <token>` (solo autor o admin)

**Response:** `204 No Content`

### 6.5 POST `/comments/:commentId/like`
Marca comentario como útil

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "likes": 4,
  "liked": true
}
```

### 6.6 DELETE `/comments/:commentId/like`
Desactiva like en comentario

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "likes": 3,
  "liked": false
}
```

---

## 7. Valoraciones (Ratings)

### 7.1 GET `/spots/:spotId/ratings`
Lista valoraciones de un lugar

**Query Parameters:**
- `limit=50`
- `offset=0`

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "r_1_1",
      "spotId": "spot_1",
      "userId": "user_2",
      "user": {
        "id": "user_2",
        "name": "Pedro Soto",
        "photoUrl": "string|null"
      },
      "stars": 4,
      "createdAt": "ISO8601"
    }
  ],
  "stats": {
    "averageRating": 4.3,
    "totalRatings": 12,
    "distribution": {
      "5": 5,
      "4": 4,
      "3": 2,
      "2": 1,
      "1": 0
    }
  },
  "pagination": {
    "total": 12,
    "limit": 50,
    "offset": 0
  }
}
```

### 7.2 POST `/spots/:spotId/ratings`
Crea o actualiza valoración

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "stars": "number (1-5, required)"
}
```

**Response:** `201 Created` o `200 OK` (si ya existía)
```json
{
  "id": "r_1_1",
  "spotId": "spot_1",
  "userId": "user_2",
  "stars": 4,
  "createdAt": "ISO8601"
}
```

### 7.3 DELETE `/spots/:spotId/ratings`
Elimina valoración del usuario

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

## 8. Búsqueda Global

### 8.1 GET `/search`
Búsqueda unificada

**Query Parameters:**
- `q=string` (término de búsqueda)
- `type=spot|species|user|all`
- `limit=20`

**Response:** `200 OK`
```json
{
  "results": {
    "spots": [
      {
        "id": "spot_1",
        "name": "Río Maipo...",
        "region": "...",
        "rating": 4.3
      }
    ],
    "species": [
      {
        "id": "sp_1",
        "commonName": "Trucha Arcoiris",
        "scientificName": "..."
      }
    ],
    "users": [
      {
        "id": "user_1",
        "name": "Luis Rivera"
      }
    ]
  }
}
```

---

## 9. Estadísticas y Análitica

### 9.1 GET `/spots/:spotId/stats`
Estadísticas de un lugar

**Response:** `200 OK`
```json
{
  "spotId": "spot_1",
  "name": "Río Maipo - Puente Negro",
  "visits": 156,
  "uniqueUsers": 42,
  "averageRating": 4.3,
  "totalRatings": 12,
  "totalComments": 3,
  "topSpecies": [
    {
      "speciesId": "sp_1",
      "name": "Trucha Arcoiris",
      "mentions": 8
    }
  ],
  "lastCommentDate": "ISO8601",
  "createdAt": "ISO8601"
}
```

### 9.2 GET `/user/:userId/stats`
Estadísticas de usuario

**Response:** `200 OK`
```json
{
  "userId": "user_1",
  "name": "Luis Rivera",
  "spotsCreated": 5,
  "commentsCount": 12,
  "ratingsCount": 8,
  "averageRating": 4.5,
  "favoriteRegions": ["Región Metropolitana", "Los Lagos"],
  "favoriteSpecies": ["sp_1", "sp_2"],
  "joiningDate": "ISO8601"
}
```

### 9.3 GET `/statistics/popular-spots`
Lugares más populares

**Query Parameters:**
- `limit=10`
- `timeRange=week|month|year|all`
- `orderBy=rating|visits|comments`

**Response:** `200 OK` (array de spots con estadísticas)

---

## 10. Favoritos (Wishlist)

### 10.1 GET `/users/:userId/favorites`
Lista lugares favoritos del usuario

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `limit=20`
- `offset=0`

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "spot_1",
      "name": "Río Maipo...",
      "region": "...",
      "rating": 4.3,
      "addedAt": "ISO8601"
    }
  ],
  "pagination": {}
}
```

### 10.2 POST `/users/:userId/favorites/:spotId`
Agrega lugar a favoritos

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`
```json
{
  "isFavorite": true,
  "addedAt": "ISO8601"
}
```

### 10.3 DELETE `/users/:userId/favorites/:spotId`
Elimina de favoritos

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

## 11. Manejo de Errores

### Standard Error Response
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": "Optional details"
  },
  "timestamp": "ISO8601"
}
```

### Códigos de Error Comunes
- `400 Bad Request` - Validación falló
- `401 Unauthorized` - Token inválido/expirado
- `403 Forbidden` - Permiso denegado
- `404 Not Found` - Recurso no existe
- `409 Conflict` - Recurso ya existe
- `429 Too Many Requests` - Rate limit excedido
- `500 Internal Server Error` - Error del servidor

---

## 12. Rate Limiting

- **Límite general:** 100 requests por minuto por IP
- **Límite autenticado:** 500 requests por minuto por usuario
- **Headers de respuesta:**
  ```
  X-RateLimit-Limit: 100
  X-RateLimit-Remaining: 87
  X-RateLimit-Reset: 1234567890
  ```

---

## 13. Paginación

Todas las listas paginadas deben seguir este formato:

```json
{
  "data": [],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0,
    "hasMore": true
  }
}
```

**Query Parameters estándar:**
- `limit` (default: 20, max: 100)
- `offset` (default: 0)

---

## 14. Base de Datos - Esquema Sugerido

### Colecciones/Tablas

#### users
```
id (PK)
email (UNIQUE)
name
photoUrl
createdAt
updatedAt
role (user|admin)
```

#### species
```
id (PK)
commonName
scientificName (UNIQUE)
description
imageUrl
habitat
diet
averageSizeCm
averageWeightKg
maxSizeCm
maxWeightKg
fishingTips (array)
createdAt
updatedAt
```

#### fishing_spots
```
id (PK)
name
description
latitude
longitude
region
createdByUserId (FK: users.id)
createdAt
updatedAt
waterType
boatAllowed
boatRequired
access
views (count)
```

#### spot_species
```
id (PK)
spotId (FK: fishing_spots.id)
speciesId (FK: species.id)
recommendedBaits (array)
recommendedRod
recommendedLine
recommendedHook
bestSeason
difficulty
notes
```

#### comments
```
id (PK)
spotId (FK: fishing_spots.id)
userId (FK: users.id)
text
likes (count)
createdAt
updatedAt
deletedAt (soft delete)
```

#### ratings
```
id (PK)
spotId (FK: fishing_spots.id)
userId (FK: users.id) + UNIQUE(spotId, userId)
stars (1-5)
createdAt
updatedAt
```

#### favorites
```
id (PK)
userId (FK: users.id)
spotId (FK: fishing_spots.id)
createdAt
UNIQUE(userId, spotId)
```

---

## 15. Requisitos No-Funcionales

- **Autenticación:** Google OAuth2 con JWT
- **Hosting:** AWS, Firebase, Heroku o similar
- **Base de datos:** MongoDB o PostgreSQL
- **Almacenamiento de archivos:** AWS S3 o Firebase Storage
- **Caché:** Redis para optimización
- **Logs:** Structured logging (ELK stack o similar)
- **Monitoreo:** Error tracking (Sentry)
- **API Documentation:** Swagger/OpenAPI
- **Testing:** 80%+ coverage
- **CI/CD:** GitHub Actions, GitLab CI o similar
- **Backup:** Daily backups
- **Seguridad:** HTTPS, CORS, SQL injection prevention, rate limiting

---

## 16. Próximas Fases (v2)

- [ ] Notificaciones push
- [ ] Seguir usuarios
- [ ] Feed personalizado
- [ ] Desafíos y logros
- [ ] Galería de fotos por spot
- [ ] Integración con mapas offline
- [ ] Pronóstico de pesca (weather API)
- [ ] Historial de capturas del usuario
- [ ] Compartir en redes sociales
- [ ] Premium features con suscripción

---

**Versión:** 1.0  
**Última actualización:** 2026-05-14  
**Autor:** Backend Specification
