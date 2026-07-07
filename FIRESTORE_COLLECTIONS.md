# Firestore Collections - PescaApp

## Colecciones a crear en Firebase Firestore

### 1. `users`
```
Document ID: auto-generated
Fields:
  - email: string (UNIQUE - verificar en código)
  - name: string
  - photoUrl: string | null
  - role: string ("user" | "admin")
  - createdAt: timestamp
  - updatedAt: timestamp
  - reputationScore: number (default: 0)
  - dailySpotLimitOverride: number | null (aplicado por penalizaciones, ver Etapa 6)
  - dailySpotLimitOverrideExpiresAt: timestamp | null
```

### 2. `species`
```
Document ID: auto-generated
Fields:
  - commonName: string
  - scientificName: string (UNIQUE - verificar en código)
  - description: string
  - imageUrl: string | null
  - habitat: string
  - diet: string
  - averageSizeCm: number
  - averageWeightKg: number
  - maxSizeCm: number
  - maxWeightKg: number
  - fishingTips: array<string>
  - createdAt: timestamp
  - updatedAt: timestamp
```

### 3. `fishing_spots`
```
Document ID: auto-generated
Fields:
  - name: string
  - description: string
  - latitude: number
  - longitude: number
  - region: string
  - waterType: string ("river" | "lake" | "sea" | "lagoon")
  - boatAllowed: boolean
  - boatRequired: boolean
  - access: string ("easy" | "moderate" | "difficult" | "boatOnly")
  - createdByUserId: string (reference to users)
  - views: number (default: 0)
  - averageRating: number (default: 0)
  - totalRatings: number (default: 0)
  - totalComments: number (default: 0)
  - status: string ("PENDING" | "VERIFIED" | "HIDDEN" | "DELETED", default: "PENDING")
  - reportCount: number (default: 0)
  - createdAt: timestamp
  - updatedAt: timestamp

Nota: documentos creados antes del sistema de moderación no tienen el campo
`status`. La aplicación los trata como "PENDING" (ver `Spot.EffectiveStatus()`
en spots/domain/entity.go) — no requiere backfill.
```

### 4. `spot_species`
```
Document ID: auto-generated
Fields:
  - spotId: string (reference to fishing_spots)
  - speciesId: string (reference to species)
  - recommendedBaits: array<string>
  - recommendedRod: string
  - recommendedLine: string
  - recommendedHook: string
  - bestSeason: string
  - difficulty: string ("easy" | "medium" | "hard")
  - notes: string
```

### 5. `comments`
```
Document ID: auto-generated
Fields:
  - spotId: string (reference to fishing_spots)
  - userId: string (reference to users)
  - text: string
  - likes: number (default: 0)
  - createdAt: timestamp
  - updatedAt: timestamp | null
  - deletedAt: timestamp | null (soft delete)
```

### 6. `comment_likes`
```
Document ID: auto-generated
Fields:
  - commentId: string (reference to comments)
  - userId: string (reference to users)
  - createdAt: timestamp

Composite Index: commentId + userId (UNIQUE - verificar en código)
```

### 7. `ratings`
```
Document ID: auto-generated
Fields:
  - spotId: string (reference to fishing_spots)
  - userId: string (reference to users)
  - stars: number (1-5)
  - createdAt: timestamp
  - updatedAt: timestamp

Composite Index: spotId + userId (UNIQUE - verificar en código)
```

### 8. `favorites`
```
Document ID: auto-generated
Fields:
  - userId: string (reference to users)
  - spotId: string (reference to fishing_spots)
  - createdAt: timestamp

Composite Index: userId + spotId (UNIQUE - verificar en código)
```

### 9. `spot_reports`
```
Document ID: auto-generated
Fields:
  - spotId: string (reference to fishing_spots)
  - reporterUserId: string (reference to users)
  - reason: string ("no_existe" | "ubicacion_incorrecta" | "informacion_falsa" | "duplicado" | "otro")
  - details: string | null (requerido cuando reason == "otro")
  - status: string ("PENDING_REVIEW" | "VALID" | "REJECTED", default: "PENDING_REVIEW")
  - createdAt: timestamp
  - reviewedAt: timestamp | null
  - reviewedByUserId: string | null

Unicidad (spotId + reporterUserId) verificada en código dentro de la misma
transacción que crea el reporte y, si corresponde, oculta el spot
(ver moderation/infrastructure/report_repository.go).

Composite Index: spotId + reporterUserId (para la verificación de unicidad)
Composite Index: spotId + createdAt DESC (para listar reportes de un spot)
```

### 10. `reputation_events`
```
Document ID: auto-generated
Fields:
  - userId: string (reference to users)
  - eventType: string ("SPOT_VERIFIED" | "SPOT_HIDDEN" | "SPOT_DELETED" | "GOOD_RATING_RECEIVED" | "REJECTED_CONTENT_PENALTY")
  - delta: number (con signo)
  - relatedSpotId: string | null
  - relatedReportId: string | null
  - reason: string
  - createdAt: timestamp

Esta colección es también el historial/auditoría de reputación del usuario —
no existe una colección de auditoría separada.

Composite Index: userId + createdAt DESC
```

### 11. `user_penalties`
```
Document ID: auto-generated
Fields:
  - userId: string (reference to users)
  - type: string ("DAILY_LIMIT_REDUCTION"; extensible — nuevos tipos no requieren cambio de esquema)
  - value: number (interpretado según type; para DAILY_LIMIT_REDUCTION = el límite diario reducido)
  - reason: string
  - appliedAt: timestamp
  - expiresAt: timestamp | null
  - revokedAt: timestamp | null (reservado para una futura apelación)
  - revokedByUserId: string | null

"Activo" se calcula en lectura (now < expiresAt && revokedAt == nil), nunca
como booleano mutable (ver Penalty.IsActive en reputation/domain/entity.go).

Composite Index: userId + type (para verificar penalización activa)
```

---

## Índices Compuestos Requeridos en Firestore

Ya definidos en `firestore.indexes.json` (raíz del repo) — desplegar con
`firebase deploy --only firestore:indexes` (requiere Firebase CLI y el
project ID correcto en `.firebaserc`). Los índices de un solo campo (marcados
como "auto" abajo) NO necesitan definirse: Firestore los crea automáticamente
para todo campo, en ambos órdenes.

1. **comments**: `spotId ASC, createdAt DESC`
2. **comments**: `spotId ASC, likes DESC`
3. **ratings**: `spotId ASC, createdAt DESC`
4. **ratings**: `spotId ASC, userId ASC` (para unicidad)
5. **favorites**: `userId ASC, createdAt DESC`
6. **favorites**: `userId ASC, spotId ASC` (para unicidad)
7. **fishing_spots**: `region ASC, averageRating DESC`
8. **fishing_spots**: `waterType ASC, averageRating DESC`
9. **fishing_spots**: `createdAt DESC` (auto, campo único)
10. **comment_likes**: `commentId ASC, userId ASC` (para unicidad)
11. **spot_species**: `spotId ASC` (auto, campo único)
12. **spot_species**: `speciesId ASC` (auto, campo único)
13. **fishing_spots**: `createdByUserId ASC, createdAt DESC` (para "mis spots" y el límite diario de creación)
14. **spot_reports**: `spotId ASC, reporterUserId ASC` (para unicidad)
15. **spot_reports**: `spotId ASC, createdAt DESC`
16. **reputation_events**: `userId ASC, createdAt DESC`
17. **user_penalties**: `userId ASC, type ASC`

## Reglas de Seguridad

Ya definidas en `firestore.rules` (raíz del repo) — deniegan todo acceso directo
desde clientes; el backend accede vía service account, que ignora las reglas.
Desplegar con `firebase deploy --only firestore:rules`, o ambas cosas juntas
con `firebase deploy --only firestore`.

```
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    match /{document=**} {
      allow read, write: if false; // Solo acceso desde backend con service account
    }
  }
}
```

